package usecases

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type CurrencyUseCase interface {
	Execute(req requests.CurrencyConversionRequest) (responses.ConvertCurrencyResponse, error)
}

type currencyUseCase struct{}

func NewCurrencyUseCase() CurrencyUseCase {
	return &currencyUseCase{}
}

func (u *currencyUseCase) Execute(req requests.CurrencyConversionRequest) (responses.ConvertCurrencyResponse, error) {
	return responses.ConvertCurrencyResponse{
		Value: req.CurrencyQuote(),
		From:  req.From,
		To:    req.To,
		Quote: req.CurrencyRate(),
	}, nil
}
