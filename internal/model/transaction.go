package model

import "time"

// Transaction status constants
const (
	PaymentStatusPending    = "PENDING"
	PaymentStatusSettlement = "SETTLEMENT"
	PaymentStatusExpired    = "EXPIRED"
	PaymentStatusCancel     = "CANCEL"
)

// Transaction represents a completed or pending POS sale
type Transaction struct {
	ID              string            `json:"id" db:"id"`
	TotalAmount     float64           `json:"total_amount" db:"total_amount"`
	PaymentStatus   string            `json:"payment_status" db:"payment_status"`
	PaymentType     string            `json:"payment_type" db:"payment_type"`
	SnapToken       string            `json:"snap_token,omitempty" db:"snap_token"`
	SnapRedirectURL string            `json:"snap_redirect_url,omitempty" db:"snap_redirect_url"`
	Items           []TransactionItem `json:"items,omitempty"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at"`
}

// TransactionItem represents a historical snapshot of sold items
type TransactionItem struct {
	ID            string  `json:"id" db:"id"`
	TransactionID string  `json:"transaction_id" db:"transaction_id"`
	ProductID     string  `json:"product_id" db:"product_id"`
	ProductName   string  `json:"product_name" db:"product_name"`
	ProductPrice  float64 `json:"product_price" db:"product_price"`
	Quantity      int     `json:"quantity" db:"quantity"`
	Subtotal      float64 `json:"subtotal" db:"subtotal"`
}

// CheckoutRequest DTO for initiating checkout from a Cart
type CheckoutRequest struct {
	CartID      string `json:"cart_id" validate:"required"`
	PaymentType string `json:"payment_type" validate:"required"` // e.g. qris, gopay, bank_transfer, cash
}

// MidtransNotificationPayload represents the HTTP webhook payload sent by Midtrans
type MidtransNotificationPayload struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	FraudStatus       string `json:"fraud_status"`
}
