package service

import (
	"context"
	"fmt"
	"time"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
	"github.com/goesbams/simple-pos-with-terraform/internal/repository"
)

type POSService interface {
	Checkout(ctx context.Context, req model.CheckoutRequest) (*model.Transaction, error)
	GetTransactionByID(ctx context.Context, id string) (*model.Transaction, error)
	GetAllTransactions(ctx context.Context) ([]model.Transaction, error)
	ProcessMidtransWebhook(ctx context.Context, payload model.MidtransNotificationPayload) error
}

type posService struct {
	cartRepo        repository.CartRepository
	productRepo     repository.ProductRepository
	transactionRepo repository.TransactionRepository
	midtransService MidtransService
}

func NewPOSService(
	cartRepo repository.CartRepository,
	productRepo repository.ProductRepository,
	transactionRepo repository.TransactionRepository,
	midtransService MidtransService,
) POSService {
	return &posService{
		cartRepo:        cartRepo,
		productRepo:     productRepo,
		transactionRepo: transactionRepo,
		midtransService: midtransService,
	}
}

func (s *posService) Checkout(ctx context.Context, req model.CheckoutRequest) (*model.Transaction, error) {
	cart, err := s.cartRepo.GetOrCreate(ctx, req.CartID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// 1. Verify stock for all items
	var txnItems []model.TransactionItem
	for _, item := range cart.Items {
		p, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if p.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for %s", p.Name)
		}

		txnItems = append(txnItems, model.TransactionItem{
			ProductID:    item.ProductID,
			ProductName:  p.Name,
			ProductPrice: p.Price,
			Quantity:     item.Quantity,
			Subtotal:     item.Subtotal,
		})
	}

	txnID := fmt.Sprintf("TXN-%d", time.Now().UnixNano()/1e6)
	txn := &model.Transaction{
		ID:            txnID,
		TotalAmount:   cart.Total,
		PaymentStatus: model.PaymentStatusPending,
		PaymentType:   req.PaymentType,
		Items:         txnItems,
	}

	// 2. Generate Midtrans Snap Token if non-cash
	if req.PaymentType != "cash" && s.midtransService != nil {
		token, redirectURL, err := s.midtransService.CreateSnapTransaction(txn.ID, int64(cart.Total), req.PaymentType)
		if err == nil {
			txn.SnapToken = token
			txn.SnapRedirectURL = redirectURL
		}
	} else if req.PaymentType == "cash" {
		txn.PaymentStatus = model.PaymentStatusSettlement
	}

	// 3. Save Transaction
	if err := s.transactionRepo.Create(ctx, txn); err != nil {
		return nil, err
	}

	// 4. Deduct stock if cash checkout (for digital payment, stock is deducted upon webhook settlement)
	if req.PaymentType == "cash" {
		for _, item := range cart.Items {
			_ = s.productRepo.DeductStock(ctx, item.ProductID, item.Quantity)
		}
	}

	// 5. Clear Cart
	_ = s.cartRepo.ClearCart(ctx, req.CartID)

	return txn, nil
}

func (s *posService) GetTransactionByID(ctx context.Context, id string) (*model.Transaction, error) {
	return s.transactionRepo.GetByID(ctx, id)
}

func (s *posService) GetAllTransactions(ctx context.Context) ([]model.Transaction, error) {
	return s.transactionRepo.GetAll(ctx)
}

func (s *posService) ProcessMidtransWebhook(ctx context.Context, payload model.MidtransNotificationPayload) error {
	// 1. Verify Midtrans Signature Key
	if s.midtransService != nil && !s.midtransService.VerifySignatureKey(payload.OrderID, payload.StatusCode, payload.GrossAmount, payload.SignatureKey) {
		return fmt.Errorf("invalid midtrans signature key")
	}

	// 2. Fetch existing transaction
	txn, err := s.transactionRepo.GetByID(ctx, payload.OrderID)
	if err != nil {
		return err
	}

	var newStatus string
	switch payload.TransactionStatus {
	case "settlement", "capture":
		newStatus = model.PaymentStatusSettlement
	case "expire":
		newStatus = model.PaymentStatusExpired
	case "cancel", "deny":
		newStatus = model.PaymentStatusCancel
	default:
		newStatus = model.PaymentStatusPending
	}

	// 3. If newly settled, deduct stock
	if txn.PaymentStatus != model.PaymentStatusSettlement && newStatus == model.PaymentStatusSettlement {
		for _, item := range txn.Items {
			_ = s.productRepo.DeductStock(ctx, item.ProductID, item.Quantity)
		}
	}

	// 4. Update transaction status
	return s.transactionRepo.UpdateStatus(ctx, payload.OrderID, newStatus)
}
