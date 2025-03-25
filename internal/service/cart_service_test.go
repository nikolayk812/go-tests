package service_test

import (
	"errors"
	"github.com/brianvoe/gofakeit"
	"github.com/nikolayk812/go-tests/internal/repository"
	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
	"testing"

	"github.com/google/uuid"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/internal/port"
	"github.com/nikolayk812/go-tests/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCartService_AddItem(t *testing.T) {
	item1 := fakeCartItem()

	item2 := fakeCartItem()
	item2.ProductID = uuid.Nil

	okOwnerID := gofakeit.UUID()

	tests := []struct {
		name      string
		item      domain.CartItem
		ownerID   string
		mockSetup func(repo *port.MockCartRepository)
		wantErr   error
	}{
		{
			name:    "success",
			item:    item1,
			ownerID: okOwnerID,
			mockSetup: func(repo *port.MockCartRepository) {
				repo.On("AddItem", mock.Anything, okOwnerID, item1).
					Return(nil)
			},
		},
		{
			name:    "ownerID is empty",
			item:    item1,
			ownerID: "",
			wantErr: errors.New("ownerID is empty"),
		},
		{
			name:    "productID is empty",
			item:    item2,
			ownerID: okOwnerID,
			wantErr: errors.New("productID is empty"),
		},
		{
			name:    "duplicate item",
			item:    item1,
			ownerID: okOwnerID,
			mockSetup: func(repo *port.MockCartRepository) {
				repo.On("AddItem", mock.Anything, okOwnerID, item1).
					Return(repository.ErrCartDuplicateItem)
			},
			wantErr: service.ErrCartDuplicateItem,
		},
		{
			name:    "unexpected error from repo",
			item:    item1,
			ownerID: okOwnerID,
			mockSetup: func(repo *port.MockCartRepository) {
				repo.On("AddItem", mock.Anything, okOwnerID, item1).
					Return(errors.New("unexpected error"))
			},
			wantErr: errors.New("repo.AddItem: unexpected error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(port.MockCartRepository)

			cs, err := service.NewCartService(mockRepo)
			require.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			err = cs.AddItem(t.Context(), tt.ownerID, tt.item)
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func fakeCartItem() domain.CartItem {
	productID := uuid.MustParse(gofakeit.UUID())

	price := gofakeit.Price(1, 100)

	currencyUnit := currency.MustParseISO(gofakeit.CurrencyShort())

	return domain.CartItem{
		ProductID: productID,
		Price: domain.Money{
			Amount:   decimal.NewFromFloat(price),
			Currency: currencyUnit,
		},
	}
}
