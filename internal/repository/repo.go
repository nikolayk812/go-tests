package repository

import (
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikolayk812/go-tests/internal/port"
)

type Repo interface {
	port.CartRepository
	port.OrderRepository
}

type repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (Repo, error) {
	if pool == nil {
		return nil, errors.New("pool is nil")
	}

	return &repo{
		pool: pool,
	}, nil
}
