package mapper

import (
	"fmt"
	"github.com/nikolayk812/go-tests/internal/domain"
	"github.com/nikolayk812/go-tests/pkg/dto"
	"golang.org/x/text/currency"
)

func MoneyToDTO(money domain.Money) dto.Money {
	return dto.Money{
		Amount:   money.Amount,
		Currency: money.Currency.String(),
	}
}

func MoneyFromDTO(money dto.Money) (domain.Money, error) {
	parsedCurrency, err := currency.ParseISO(money.Currency)
	if err != nil {
		return domain.Money{}, fmt.Errorf("c currency.ParseISO[%s]: %w", money.Currency, err)
	}

	return domain.Money{
		Amount:   money.Amount,
		Currency: parsedCurrency,
	}, nil
}
