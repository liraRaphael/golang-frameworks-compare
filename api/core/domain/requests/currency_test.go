package requests

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyConversionRequest_CurrencyQuote(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		to        string
		value     float64
		wantQuote float64
		wantRate  float64
	}{
		{name: "USD to BRL", from: "USD", to: "BRL", value: 10, wantQuote: 63.1, wantRate: 6.31},
		{name: "BRL to USD", from: "BRL", to: "USD", value: 63.1, wantQuote: 10, wantRate: 1 / 6.31},
		{name: "same currency rounded", from: "USD", to: "USD", value: 1.125, wantQuote: 1.13, wantRate: 1},
		{name: "unsupported conversion", from: "EUR", to: "BRL", value: 10, wantQuote: 0, wantRate: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := CurrencyConversionRequest{From: tt.from, To: tt.to, Value: &tt.value}

			assert.InDelta(t, tt.wantQuote, request.CurrencyQuote(), 1e-9)
			assert.InDelta(t, tt.wantRate, request.CurrencyRate(), 1e-9)
		})
	}
}

func TestCurrencyConversionRequest_CurrencyQuotePreservesLargeFiniteValues(t *testing.T) {
	value := math.MaxFloat64
	request := CurrencyConversionRequest{From: "USD", To: "USD", Value: &value}

	assert.Equal(t, value, request.CurrencyQuote())
}

func TestCurrencyConversionRequest_CurrencyQuoteReturnsZeroWithoutValue(t *testing.T) {
	request := CurrencyConversionRequest{From: "USD", To: "BRL"}

	assert.Zero(t, request.CurrencyQuote())
}
