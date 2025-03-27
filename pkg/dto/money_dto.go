package dto

import (
	"github.com/shopspring/decimal"
)

type Money struct {
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
}
