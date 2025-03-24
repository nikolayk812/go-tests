package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nikolayk812/go-tests/internal/domain"
	"golang.org/x/text/currency"
)

func (r *repo) GetCart(ctx context.Context, ownerID string) (domain.Cart, error) {
	var c domain.Cart

	rows, err := r.pool.Query(ctx, "SELECT * FROM cart_items WHERE owner_id = $1", ownerID)
	if err != nil {
		return c, fmt.Errorf("pool.Query: %w", err)
	}

	var scannedOwnerID string

	cartItems, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.CartItem, error) {
		var (
			item        domain.CartItem
			currencyStr string
		)

		if err := row.Scan(&scannedOwnerID, &item.ProductID, &item.Price.Amount, &currencyStr, &item.CreatedAt); err != nil {
			return domain.CartItem{}, fmt.Errorf("row.Scan: %w", err)
		}

		currencyUnit, err := currency.ParseISO(currencyStr)
		if err != nil {
			return domain.CartItem{}, fmt.Errorf("currency.ParseISO[%s]: %w", currencyStr, err)
		}

		item.Price.Currency = currencyUnit

		return item, nil
	})
	if err != nil {
		return c, fmt.Errorf("pgx.CollectRows: %w", err)
	}

	return domain.Cart{
		OwnerID: ownerID,
		Items:   cartItems,
	}, nil
}

func (r *repo) AddItem(ctx context.Context, ownerID string, item domain.CartItem) error {

	_, err := r.pool.Exec(ctx, `
			INSERT INTO cart_items (owner_id, product_id, price_amount, price_currency) 
			VALUES ($1, $2, $3, $4)`,
		ownerID, item.ProductID, item.Price.Amount, item.Price.Currency)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrCartDuplicateItem
		}
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *repo) DeleteItem(ctx context.Context, ownerID string, productID uuid.UUID) (bool, error) {
	cmdTag, err := r.pool.Exec(ctx, "DELETE FROM cart_items WHERE owner_id = $1 AND product_id = $2", ownerID, productID)
	if err != nil {
		return false, fmt.Errorf("pool.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}
