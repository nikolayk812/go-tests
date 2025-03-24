package service

import "errors"

var (
	ErrCartDuplicateItem = errors.New("duplicate cart item")
	ErrCartItemNotFound  = errors.New("cart item not found")
)
