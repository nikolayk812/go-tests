package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/internal/port"
	"github.com/nikolayk812/go-tests/internal/repository"
)

//go:generate mockery --name=CartService --structname=MockCartService --output=. --outpkg=service --filename=cart_service_mock.go
type CartService interface {
	GetCart(ctx context.Context, ownerID string) (domain.Cart, error)
	AddItem(ctx context.Context, ownerID string, item domain.CartItem) error
	DeleteItem(ctx context.Context, ownerID string, productID uuid.UUID) error
}

type cartService struct {
	repo port.CartRepository
}

func NewCart(repo port.CartRepository) (CartService, error) {
	if repo == nil {
		return nil, errors.New("repo is nil")
	}

	return &cartService{repo: repo}, nil
}

func (cs *cartService) GetCart(ctx context.Context, ownerID string) (domain.Cart, error) {
	var cart domain.Cart

	if ownerID == "" {
		return cart, errors.New("ownerID is empty")
	}

	return cs.repo.GetCart(ctx, ownerID)
}

func (cs *cartService) AddItem(ctx context.Context, ownerID string, item domain.CartItem) error {
	if ownerID == "" {
		return errors.New("ownerID is empty")
	}

	if item.ProductID == uuid.Nil {
		return errors.New("productID is empty")
	}

	if err := cs.repo.AddItem(ctx, ownerID, item); err != nil {
		if errors.Is(err, repository.ErrCartDuplicateItem) {
			return ErrCartDuplicateItem // from service layer
		}
		return fmt.Errorf("repo.AddItem: %w", err)
	}

	return nil
}

func (cs *cartService) DeleteItem(ctx context.Context, ownerID string, productID uuid.UUID) error {
	if ownerID == "" {
		return errors.New("ownerID is empty")
	}

	if productID == uuid.Nil {
		return errors.New("productID is empty")
	}

	deleted, err := cs.repo.DeleteItem(ctx, ownerID, productID)
	if err != nil {
		return fmt.Errorf("repo.DeleteItem: %w", err)
	}

	if !deleted {
		return ErrCartItemNotFound // to return error is decision of service layer
	}

	return nil
}
