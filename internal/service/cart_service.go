package service

import (
	"context"
	"fmt"

	"github.com/goesbams/simple-pos-with-terraform/internal/model"
	"github.com/goesbams/simple-pos-with-terraform/internal/repository"
)

type CartService interface {
	GetCart(ctx context.Context, cartID string) (*model.Cart, error)
	AddToCart(ctx context.Context, cartID string, req model.AddToCartRequest) (*model.Cart, error)
	RemoveFromCart(ctx context.Context, cartID string, productID string) (*model.Cart, error)
	ClearCart(ctx context.Context, cartID string) error
}

type cartService struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartService(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartService {
	return &cartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *cartService) GetCart(ctx context.Context, cartID string) (*model.Cart, error) {
	return s.cartRepo.GetOrCreate(ctx, cartID)
}

func (s *cartService) AddToCart(ctx context.Context, cartID string, req model.AddToCartRequest) (*model.Cart, error) {
	p, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	if p.Stock < req.Quantity {
		return nil, fmt.Errorf("insufficient stock for product %s (available: %d)", p.Name, p.Stock)
	}

	if err := s.cartRepo.AddItem(ctx, cartID, p, req.Quantity); err != nil {
		return nil, err
	}

	return s.cartRepo.GetOrCreate(ctx, cartID)
}

func (s *cartService) RemoveFromCart(ctx context.Context, cartID string, productID string) (*model.Cart, error) {
	if err := s.cartRepo.RemoveItem(ctx, cartID, productID); err != nil {
		return nil, err
	}
	return s.cartRepo.GetOrCreate(ctx, cartID)
}

func (s *cartService) ClearCart(ctx context.Context, cartID string) error {
	return s.cartRepo.ClearCart(ctx, cartID)
}
