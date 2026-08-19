package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type ProductRepository interface {
	GetAll(ctx context.Context) ([]model.Product, error)
	GetByID(ctx context.Context, id string) (*model.Product, error)
	Create(ctx context.Context, p *model.Product) error
	Update(ctx context.Context, p *model.Product) error
	SoftDelete(ctx context.Context, id string) error
	DeductStock(ctx context.Context, id string, quantity int) error
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	query := `SELECT id, name, price, stock, category, created_at, updated_at 
	          FROM products 
	          WHERE deleted_at IS NULL 
	          ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if products == nil {
		products = []model.Product{}
	}

	return products, nil
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	query := `SELECT id, name, price, stock, category, created_at, updated_at 
	          FROM products 
	          WHERE id = $1 AND deleted_at IS NULL`

	var p model.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &p, nil
}

func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	query := `INSERT INTO products (id, name, price, stock, category, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query, p.ID, p.Name, p.Price, p.Stock, p.Category, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *productRepository) Update(ctx context.Context, p *model.Product) error {
	query := `UPDATE products 
	          SET name = $1, price = $2, stock = $3, category = $4, updated_at = $5 
	          WHERE id = $6 AND deleted_at IS NULL`

	p.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query, p.Name, p.Price, p.Stock, p.Category, p.UpdatedAt, p.ID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *productRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE products 
	          SET deleted_at = $1, updated_at = $1 
	          WHERE id = $2 AND deleted_at IS NULL`

	now := time.Now()
	res, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrProductNotFound
	}

	return nil
}

func (r *productRepository) DeductStock(ctx context.Context, id string, quantity int) error {
	query := `UPDATE products 
	          SET stock = stock - $1, updated_at = $2 
	          WHERE id = $3 AND stock >= $1 AND deleted_at IS NULL`

	now := time.Now()
	res, err := r.db.ExecContext(ctx, query, quantity, now, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInsufficientStock
	}

	return nil
}
