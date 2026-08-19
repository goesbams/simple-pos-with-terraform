package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
)

var (
	ErrCartNotFound = errors.New("cart not found")
)

type CartRepository interface {
	GetOrCreate(ctx context.Context, cartID string) (*model.Cart, error)
	AddItem(ctx context.Context, cartID string, product *model.Product, quantity int) error
	RemoveItem(ctx context.Context, cartID string, productID string) error
	ClearCart(ctx context.Context, cartID string) error
}

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreate(ctx context.Context, cartID string) (*model.Cart, error) {
	if cartID == "" {
		cartID = "default_cart"
	}

	// 1. Ensure cart exists
	insertQuery := `INSERT INTO carts (id, created_at, updated_at) 
	                VALUES ($1, $2, $2) 
	                ON CONFLICT (id) DO NOTHING`
	now := time.Now()
	if _, err := r.db.ExecContext(ctx, insertQuery, cartID, now); err != nil {
		return nil, err
	}

	// 2. Fetch cart items with product details
	query := `SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.subtotal, ci.created_at, ci.updated_at,
	                 p.name, p.price, p.stock, p.category
	          FROM cart_items ci
	          JOIN products p ON ci.product_id = p.id
	          WHERE ci.cart_id = $1 AND p.deleted_at IS NULL`

	rows, err := r.db.QueryContext(ctx, query, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.CartItem
	var total float64

	for rows.Next() {
		var item model.CartItem
		var p model.Product
		if err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.Subtotal, &item.CreatedAt, &item.UpdatedAt,
			&p.Name, &p.Price, &p.Stock, &p.Category,
		); err != nil {
			return nil, err
		}
		p.ID = item.ProductID
		item.Product = &p
		total += item.Subtotal
		items = append(items, item)
	}

	if items == nil {
		items = []model.CartItem{}
	}

	return &model.Cart{
		ID:        cartID,
		Items:     items,
		Total:     total,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *cartRepository) AddItem(ctx context.Context, cartID string, product *model.Product, quantity int) error {
	subtotal := product.Price * float64(quantity)
	now := time.Now()

	// Check if item exists in cart
	var existingQty int
	var itemID string
	checkQuery := `SELECT id, quantity FROM cart_items WHERE cart_id = $1 AND product_id = $2`
	err := r.db.QueryRowContext(ctx, checkQuery, cartID, product.ID).Scan(&itemID, &existingQty)

	if errors.Is(err, sql.ErrNoRows) {
		// Insert new item
		insertQuery := `INSERT INTO cart_items (id, cart_id, product_id, quantity, subtotal, created_at, updated_at) 
		                VALUES ($1, $2, $3, $4, $5, $6, $7)`
		newItemID := "CI-" + time.Now().Format("20060102150405999999")
		_, err := r.db.ExecContext(ctx, insertQuery, newItemID, cartID, product.ID, quantity, subtotal, now, now)
		return err
	} else if err != nil {
		return err
	}

	// Update existing item quantity
	newQty := existingQty + quantity
	newSubtotal := product.Price * float64(newQty)
	updateQuery := `UPDATE cart_items SET quantity = $1, subtotal = $2, updated_at = $3 WHERE id = $4`
	_, err = r.db.ExecContext(ctx, updateQuery, newQty, newSubtotal, now, itemID)
	return err
}

func (r *cartRepository) RemoveItem(ctx context.Context, cartID string, productID string) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`
	_, err := r.db.ExecContext(ctx, query, cartID, productID)
	return err
}

func (r *cartRepository) ClearCart(ctx context.Context, cartID string) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err := r.db.ExecContext(ctx, query, cartID)
	return err
}
