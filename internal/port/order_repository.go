package port

import (
	"context"
	"github.com/nikolayk812/go-tests/internal/domain"
)

type OrderRepository interface {
	GetOrder(ctx context.Context, ownerID string) (domain.Order, error)
	CreateOrder(ctx context.Context, ownerID string) error
}
