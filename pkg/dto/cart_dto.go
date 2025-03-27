package dto

import (
	"github.com/google/uuid"
	"time"
)

type Cart struct {
	OwnerID string     `json:"owner_id"`
	Items   []CartItem `json:"items"`
}

type CartItem struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Price     Money     `json:"price" binding:"required"`

	CreatedAt time.Time `json:"created_at"`
}
