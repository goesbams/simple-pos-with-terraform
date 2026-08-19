package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
)

type TransactionRepository interface {
	Create(ctx context.Context, txn *model.Transaction) error
	GetByID(ctx context.Context, id string) (*model.Transaction, error)
	GetAll(ctx context.Context) ([]model.Transaction, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, txn *model.Transaction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txnQuery := `INSERT INTO transactions (id, total_amount, payment_status, payment_type, snap_token, snap_redirect_url, created_at) 
	             VALUES ($1, $2, $3, $4, $5, $6, $7)`
	txn.CreatedAt = time.Now()

	_, err = tx.ExecContext(
		ctx, txnQuery,
		txn.ID, txn.TotalAmount, txn.PaymentStatus, txn.PaymentType, txn.SnapToken, txn.SnapRedirectURL, txn.CreatedAt,
	)
	if err != nil {
		return err
	}

	itemQuery := `INSERT INTO transaction_items (id, transaction_id, product_id, product_name, product_price, quantity, subtotal) 
	              VALUES ($1, $2, $3, $4, $5, $6, $7)`

	for _, item := range txn.Items {
		itemID := "TI-" + time.Now().Format("20060102150405999999")
		_, err = tx.ExecContext(
			ctx, itemQuery,
			itemID, txn.ID, item.ProductID, item.ProductName, item.ProductPrice, item.Quantity, item.Subtotal,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	txnQuery := `SELECT id, total_amount, payment_status, payment_type, snap_token, snap_redirect_url, created_at 
	              FROM transactions 
	              WHERE id = $1`

	var txn model.Transaction
	var snapToken, snapRedirectURL sql.NullString

	err := r.db.QueryRowContext(ctx, txnQuery, id).Scan(
		&txn.ID, &txn.TotalAmount, &txn.PaymentStatus, &txn.PaymentType, &snapToken, &snapRedirectURL, &txn.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}

	if snapToken.Valid {
		txn.SnapToken = snapToken.String
	}
	if snapRedirectURL.Valid {
		txn.SnapRedirectURL = snapRedirectURL.String
	}

	// Fetch transaction items (historical snapshot)
	itemsQuery := `SELECT id, transaction_id, product_id, product_name, product_price, quantity, subtotal 
	               FROM transaction_items 
	               WHERE transaction_id = $1`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.TransactionItem
	for rows.Next() {
		var item model.TransactionItem
		if err := rows.Scan(&item.ID, &item.TransactionID, &item.ProductID, &item.ProductName, &item.ProductPrice, &item.Quantity, &item.Subtotal); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	txn.Items = items
	return &txn, nil
}

func (r *transactionRepository) GetAll(ctx context.Context) ([]model.Transaction, error) {
	query := `SELECT id, total_amount, payment_status, payment_type, snap_token, snap_redirect_url, created_at 
	          FROM transactions 
	          ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for rows.Next() {
		var txn model.Transaction
		var snapToken, snapRedirectURL sql.NullString

		if err := rows.Scan(&txn.ID, &txn.TotalAmount, &txn.PaymentStatus, &txn.PaymentType, &snapToken, &snapRedirectURL, &txn.CreatedAt); err != nil {
			return nil, err
		}
		if snapToken.Valid {
			txn.SnapToken = snapToken.String
		}
		if snapRedirectURL.Valid {
			txn.SnapRedirectURL = snapRedirectURL.String
		}
		transactions = append(transactions, txn)
	}

	if transactions == nil {
		transactions = []model.Transaction{}
	}

	return transactions, nil
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE transactions SET payment_status = $1 WHERE id = $2`
	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTransactionNotFound
	}

	return nil
}
