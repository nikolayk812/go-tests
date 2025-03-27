package mapper

import (
	"fmt"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/pkg/dto"
)

func CartToDTO(cart domain.Cart) dto.Cart {
	items := make([]dto.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, CartItemToDTO(item))
	}

	return dto.Cart{
		OwnerID: cart.OwnerID,
		Items:   items,
	}
}

func CartItemToDTO(item domain.CartItem) dto.CartItem {
	return dto.CartItem{
		ProductID: item.ProductID,
		Price:     MoneyToDTO(item.Price),
		CreatedAt: item.CreatedAt,
	}
}

func CartItemFromDTO(item dto.CartItem) (domain.CartItem, error) {
	price, err := MoneyFromDTO(item.Price)
	if err != nil {
		return domain.CartItem{}, fmt.Errorf("MoneyFromDTO: %w", err)
	}

	return domain.CartItem{
		ProductID: item.ProductID,
		Price:     price,
		CreatedAt: item.CreatedAt,
	}, nil
}
