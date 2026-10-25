package controller

import (
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyController_ExecuteReturnsUseCaseOutput(t *testing.T) {
	value := 10.0
	ctrl := NewCurrencyController(usecases.NewCurrencyUseCase())
	req := requests.NewRequest[requests.CurrencyConversionRequest, any](
		requests.CurrencyConversionRequest{From: "USD", To: "BRL", Value: &value},
		nil,
	)

	result, err := ctrl.Execute(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, 200, result.StatusCode())
	assert.Equal(t, responses.ConvertCurrencyResponse{
		Value: 63.1,
		From:  "USD",
		To:    "BRL",
		Quote: 6.31,
	}, result.Body())
}

type failingCurrencyUseCase struct {
	err error
}

func (f failingCurrencyUseCase) Execute(requests.CurrencyConversionRequest) (responses.ConvertCurrencyResponse, error) {
	return responses.ConvertCurrencyResponse{}, f.err
}

func TestCurrencyController_ReturnsUseCaseError(t *testing.T) {
	expectedErr := assert.AnError
	ctrl := NewCurrencyController(failingCurrencyUseCase{err: expectedErr})
	req := requests.NewRequest[requests.CurrencyConversionRequest, any](requests.CurrencyConversionRequest{}, nil)

	result, err := ctrl.Execute(context.Background(), req)

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, result)
}

var _ ports.CurrencyController = (*currencyController)(nil)
