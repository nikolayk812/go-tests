package repository

import (
	"context"
	"github.com/nikolayk812/go-tests/internal/domain"
)

func (r *repo) GetOrder(ctx context.Context, orderID string) (domain.Order, error) {
	// TODO: implement

	return domain.Order{}, nil
}

func (r *repo) CreateOrder(ctx context.Context, orderID string) error {
	// TODO: implement

	return nil
}
