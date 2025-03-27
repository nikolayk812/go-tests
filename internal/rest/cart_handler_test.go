package rest_test

import (
	"encoding/json"
	"errors"
	"github.com/brianvoe/gofakeit"
	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/internal/rest"
	"github.com/nikolayk812/go-tests/internal/rest/mapper"
	"github.com/nikolayk812/go-tests/internal/service"
	"github.com/nikolayk812/go-tests/pkg/dto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/currency"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCartHandler_GetCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ownerID := gofakeit.UUID()
	item1 := fakeCartItem()
	item1DTO := mapper.CartItemToDTO(item1)

	tests := []struct {
		name       string
		mockSetup  func(service *service.MockCartService)
		wantStatus int
		wantCart   dto.Cart
	}{
		{
			name: "okay, no items",
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, ownerID).
					Return(domain.Cart{OwnerID: ownerID}, nil)
			},
			wantStatus: http.StatusOK,
			wantCart: dto.Cart{
				OwnerID: ownerID,
			},
		},
		{
			name: "okay, with one item",
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, ownerID).
					Return(domain.Cart{
						OwnerID: ownerID,
						Items:   []domain.CartItem{item1},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantCart: dto.Cart{
				OwnerID: ownerID,
				Items:   []dto.CartItem{item1DTO},
			},
		},
		{
			name: "service error",
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, ownerID).
					Return(domain.Cart{}, errors.New("service error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(service.MockCartService)
			handler, err := rest.NewCart(mockService)
			require.NoError(t, err)

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			recorder := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(recorder)
			c.Params = gin.Params{gin.Param{Key: "owner_id", Value: ownerID}}

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			require.NoError(t, err)
			c.Request = req

			handler.GetCart(c)

			assert.Equal(t, tt.wantStatus, recorder.Code)

			if tt.wantStatus == http.StatusOK {
				var actualCart dto.Cart
				err := json.Unmarshal(recorder.Body.Bytes(), &actualCart)
				require.NoError(t, err)
				assertEqualCart(t, tt.wantCart, actualCart)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func assertEqualCart(t *testing.T, expected, actual dto.Cart) {
	opts := cmp.Options{
		cmpopts.EquateEmpty(),
	}

	diff := cmp.Diff(expected, actual, opts)
	assert.Empty(t, diff)
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
