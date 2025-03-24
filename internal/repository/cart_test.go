package repository_test

import (
	"github.com/brianvoe/gofakeit"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/internal/port"
	"github.com/nikolayk812/go-tests/internal/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/goleak"
	"golang.org/x/text/currency"
	"testing"
)

type cartRepositorySuite struct {
	suite.Suite

	pool      *pgxpool.Pool
	repo      port.CartRepository
	container testcontainers.Container
}

// entry point to run the tests in the suite
func TestCartRepositorySuite(t *testing.T) {
	// Verifies no leaks after all tests in the suite run.
	defer goleak.VerifyNone(t)

	suite.Run(t, new(cartRepositorySuite))
}

// before all tests in the suite
func (suite *cartRepositorySuite) SetupSuite() {
	ctx := suite.T().Context()

	var (
		connStr string
		err     error
	)

	suite.container, connStr, err = startPostgres(ctx)
	suite.NoError(err)

	suite.pool, err = pgxpool.New(ctx, connStr)
	suite.NoError(err)

	suite.repo, err = repository.New(suite.pool)
	suite.NoError(err)
}

// after all tests in the suite
func (suite *cartRepositorySuite) TearDownSuite() {
	ctx := suite.T().Context()

	if suite.pool != nil {
		suite.pool.Close()
	}
	if suite.container != nil {
		suite.NoError(suite.container.Terminate(ctx))
	}
}

func (suite *cartRepositorySuite) TestAddItem() {
	item1 := fakeCartItem()
	item2 := fakeCartItem()

	testCases := []struct {
		name      string
		items     []domain.CartItem
		ownerID   string
		wantError error
	}{
		{
			name: "empty cart: ok",
		},
		{
			name:  "single item: ok",
			items: []domain.CartItem{item1},
		},
		{
			name:  "three items: ok",
			items: []domain.CartItem{item1, item2},
		},
		{
			name:      "duplicate item: fail",
			items:     []domain.CartItem{item1, item1},
			wantError: repository.ErrCartDuplicateItem,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			t := suite.T()
			ctx := t.Context()

			ownerID := gofakeit.UUID()

			for _, item := range tc.items {
				err := suite.repo.AddItem(ctx, ownerID, item)
				if err != nil {
					require.ErrorIs(t, err, tc.wantError)
					return
				}
			}

			cart, err := suite.repo.GetCart(ctx, ownerID)
			require.NoError(t, err)

			expectedCart := domain.Cart{
				OwnerID: ownerID,
				Items:   tc.items,
			}
			assertCart(t, expectedCart, cart)
		})
	}
}

func (suite *cartRepositorySuite) TestDeleteItem() {
	item1 := fakeCartItem()
	item2 := fakeCartItem()

	testCases := []struct {
		name     string
		items    []domain.CartItem
		deleteID uuid.UUID
		deleted  bool
	}{
		{
			name:     "empty cart",
			deleteID: item1.ProductID,
		},
		{
			name:     "existing item",
			items:    []domain.CartItem{item1},
			deleteID: item1.ProductID,
			deleted:  true,
		},
		{
			name:     "one of multiple items",
			items:    []domain.CartItem{item1, item2},
			deleteID: item1.ProductID,
			deleted:  true,
		},
		{
			name:     "non-existent item",
			items:    []domain.CartItem{item1},
			deleteID: uuid.MustParse(gofakeit.UUID()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			t := suite.T()
			ctx := t.Context()

			ownerID := gofakeit.UUID()

			for _, item := range tc.items {
				err := suite.repo.AddItem(ctx, ownerID, item)
				require.NoError(t, err)
			}

			deleted, err := suite.repo.DeleteItem(ctx, ownerID, tc.deleteID)
			require.NoError(t, err)
			assert.Equal(t, tc.deleted, deleted)

			cart, err := suite.repo.GetCart(ctx, ownerID)
			require.NoError(t, err)

			expectedItems := make([]domain.CartItem, 0)
			for _, item := range tc.items {
				if item.ProductID != tc.deleteID {
					expectedItems = append(expectedItems, item)
				}
			}

			expectedCart := domain.Cart{
				OwnerID: ownerID,
				Items:   expectedItems,
			}
			assertCart(t, expectedCart, cart)
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

func assertCart(t *testing.T, expected domain.Cart, actual domain.Cart) {
	t.Helper()

	// Custom comparer for Money.Currency fields
	comparer := cmp.Comparer(func(x, y currency.Unit) bool {
		return x.String() == y.String()
	})

	// Ignore the CreatedAt field in CartItem and
	// Treat empty slices as equal to nil
	opts := cmp.Options{
		cmpopts.IgnoreFields(domain.CartItem{}, "CreatedAt"),
		cmpopts.EquateEmpty(),
	}

	diff := cmp.Diff(expected, actual, comparer, opts)
	assert.Empty(t, diff)
}
