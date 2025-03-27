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
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/currency"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCartHandler_GetCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cart1 := fakeCart()
	cart1.Items = nil
	cart1DTO := mapper.CartToDTO(cart1)

	cart2 := fakeCart()
	cart2DTO := mapper.CartToDTO(cart2)

	ownerID := gofakeit.UUID()

	tests := []struct {
		name       string
		ownerID    string
		mockSetup  func(service *service.MockCartService)
		wantStatus int
		wantCart   dto.Cart
	}{
		{
			name:    "okay, no items",
			ownerID: cart1.OwnerID,
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, cart1.OwnerID).
					Return(cart1, nil)
			},
			wantStatus: http.StatusOK,
			wantCart:   cart1DTO,
		},
		{
			name:    "okay, with one item",
			ownerID: cart2.OwnerID,
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, cart2.OwnerID).
					Return(cart2, nil)
			},
			wantStatus: http.StatusOK,
			wantCart:   cart2DTO,
		},
		{
			name:    "service error",
			ownerID: ownerID,
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
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler, err := rest.NewCart(mockService)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(recorder)
			c.Params = gin.Params{gin.Param{Key: "owner_id", Value: tt.ownerID}}

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

func fakeCart() domain.Cart {
	var items []domain.CartItem

	for range gofakeit.Number(1, 5) {
		items = append(items, fakeCartItem())
	}

	return domain.Cart{
		OwnerID: gofakeit.UUID(),
		Items:   items,
	}
}

func TestCartHandler_GetCart2(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cart1 := fakeCart()
	cart1.Items = nil
	cart1DTO := mapper.CartToDTO(cart1)

	cart2 := fakeCart()
	cart2DTO := mapper.CartToDTO(cart2)

	ownerID := gofakeit.UUID()

	tests := []struct {
		name       string
		ownerID    string
		mockSetup  func(service *service.MockCartService)
		wantStatus int
		wantCart   dto.Cart
	}{
		{
			name:    "okay, no items",
			ownerID: cart1.OwnerID,
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, cart1.OwnerID).
					Return(cart1, nil)
			},
			wantStatus: http.StatusOK,
			wantCart:   cart1DTO,
		},
		{
			name:    "okay, with one item",
			ownerID: cart2.OwnerID,
			mockSetup: func(service *service.MockCartService) {
				service.On("GetCart", mock.Anything, cart2.OwnerID).
					Return(cart2, nil)
			},
			wantStatus: http.StatusOK,
			wantCart:   cart2DTO,
		},
		{
			name:    "service error",
			ownerID: ownerID,
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

			apitest.New().
				HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c, _ := gin.CreateTestContext(w)
					c.Request = r
					c.Params = gin.Params{gin.Param{Key: "owner_id", Value: tt.ownerID}}

					handler.GetCart(c)
				}).
				Get("/").
				Expect(t).
				Status(tt.wantStatus).
				Assert(func(res *http.Response, r *http.Request) error {
					// Read the body
					body, err := ioutil.ReadAll(res.Body)
					assert.NoError(t, err)

					var actualCart dto.Cart
					err = json.Unmarshal(body, &actualCart)
					require.NoError(t, err)

					assertEqualCart(t, tt.wantCart, actualCart)

					return nil
				}).
				End()

			mockService.AssertExpectations(t)
		})
	}
}
