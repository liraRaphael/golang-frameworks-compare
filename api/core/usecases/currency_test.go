package usecases

import (
	"math"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyUseCase_Execute(t *testing.T) {
	useCase := NewCurrencyUseCase()
	tests := []struct {
		name      string
		from      string
		to        string
		value     float64
		wantValue float64
		wantQuote float64
	}{
		{name: "USD to BRL", from: "USD", to: "BRL", value: 10, wantValue: 63.1, wantQuote: 6.31},
		{name: "BRL to USD", from: "BRL", to: "USD", value: 63.1, wantValue: 10, wantQuote: 1 / 6.31},
		{name: "same currency", from: "USD", to: "USD", value: 10, wantValue: 10, wantQuote: 1},
		{name: "zero", from: "BRL", to: "BRL", value: 0, wantValue: 0, wantQuote: 1},
		{name: "round half away from zero", from: "BRL", to: "BRL", value: 1.125, wantValue: 1.13, wantQuote: 1},
		{name: "more than two decimal places", from: "USD", to: "USD", value: 1.2345, wantValue: 1.23, wantQuote: 1},
		{name: "large finite same-currency result", from: "USD", to: "USD", value: math.MaxFloat64, wantValue: math.MaxFloat64, wantQuote: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := useCase.Execute(requests.CurrencyConversionRequest{
				From:  tt.from,
				To:    tt.to,
				Value: &tt.value,
			})
			require.NoError(t, err)
			assert.InDelta(t, tt.wantValue, got.Value, 1e-9)
			assert.InDelta(t, tt.wantQuote, got.Quote, 1e-9)
			assert.Equal(t, tt.from, got.From)
			assert.Equal(t, tt.to, got.To)
		})
	}
}
