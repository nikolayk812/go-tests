package rest_test

import (
	"bytes"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/internal/rest/mapper"
	"golang.org/x/text/currency"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nikolayk812/go-tests/internal/rest"
	"github.com/nikolayk812/go-tests/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCartGroupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cart1 := fakeCart()
	cartItem1DTO := mapper.CartItemToDTO(cart1.Items[0])
	product1UID := cartItem1DTO.ProductID

	mockService := new(service.MockCartService)
	handler, err := rest.NewCart(mockService)
	require.NoError(t, err)

	router := rest.SetupRouter(handler)

	tests := []struct {
		name       string
		method     string
		url        string
		body       interface{}
		mockFunc   func()
		statusCode int
	}{
		{
			name:   "GetCart",
			method: http.MethodGet,
			url:    "/carts/123",
			mockFunc: func() {
				mockService.On("GetCart", mock.Anything, "123").
					Return(cart1, nil)
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "AddItem",
			method: http.MethodPost,
			url:    "/carts/123",
			body:   cartItem1DTO,
			mockFunc: func() {
				mockService.On("AddItem", mock.Anything, "123", mock.MatchedBy(func(item domain.CartItem) bool {
					equalCartItem(item, cart1.Items[0])
					return true
				})).Return(nil)
			},
			statusCode: http.StatusCreated,
		},
		{
			name:   "DeleteItem",
			method: http.MethodDelete,
			url:    "/carts/123/" + product1UID.String(),
			mockFunc: func() {
				mockService.On("DeleteItem", mock.Anything, "123", product1UID).Return(nil)
			},
			statusCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			var bodyBytes []byte
			if tt.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			mockService.AssertExpectations(t)
		})
	}
}

func equalCartItem(item1, item2 domain.CartItem) bool {

	// Custom comparer for Money.Currency fields
	comparer := cmp.Comparer(func(x, y currency.Unit) bool {
		return x.String() == y.String()
	})

	// Treat empty slices as equal to nil
	opts := cmp.Options{
		cmpopts.EquateEmpty(),
	}

	diff := cmp.Diff(item1, item2, comparer, opts)
	return diff == ""
}
