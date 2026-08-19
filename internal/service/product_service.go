package service

import (
	"context"
	"fmt"
	"time"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
	"github.com/goesbams/simple-pos-with-terraform/internal/repository"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]model.Product, error)
	GetProductByID(ctx context.Context, id string) (*model.Product, error)
	CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error)
	UpdateProduct(ctx context.Context, id string, req model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	return s.repo.GetAll(ctx)
}

func (s *productService) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	productID := fmt.Sprintf("PRD-%d", time.Now().UnixNano()/1e6)
	p := &model.Product{
		ID:       productID,
		Name:     req.Name,
		Price:    req.Price,
		Stock:    req.Stock,
		Category: req.Category,
	}

	if p.Category == "" {
		p.Category = "General"
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id string, req model.UpdateProductRequest) (*model.Product, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Price > 0 {
		existing.Price = req.Price
	}
	if req.Stock >= 0 {
		existing.Stock = req.Stock
	}
	if req.Category != "" {
		existing.Category = req.Category
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
