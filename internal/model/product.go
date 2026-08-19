package model

import "time"

// Product represents a POS item entity
type Product struct {
	ID        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Price     float64    `json:"price" db:"price"`
	Stock     int        `json:"stock" db:"stock"`
	Category  string     `json:"category" db:"category"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Soft Delete
}

// CreateProductRequest DTO for creating a new product
type CreateProductRequest struct {
	Name     string  `json:"name" validate:"required"`
	Price    float64 `json:"price" validate:"required,gt=0"`
	Stock    int     `json:"stock" validate:"required,gte=0"`
	Category string  `json:"category"`
}

// UpdateProductRequest DTO for updating existing product
type UpdateProductRequest struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price" validate:"gt=0"`
	Stock    int     `json:"stock" validate:"gte=0"`
	Category string  `json:"category"`
}
