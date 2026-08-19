package model

import "time"

// Cart represents a cashier shopping cart
type Cart struct {
	ID        string     `json:"id" db:"id"`
	Items     []CartItem `json:"items,omitempty"`
	Total     float64    `json:"total"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// CartItem represents an item inside the cart
type CartItem struct {
	ID        string    `json:"id" db:"id"`
	CartID    string    `json:"cart_id" db:"cart_id"`
	ProductID string    `json:"product_id" db:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Subtotal  float64   `json:"subtotal" db:"subtotal"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AddToCartRequest DTO for adding items to cart
type AddToCartRequest struct {
	ProductID string `json:"product_id" validate:"required"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}
