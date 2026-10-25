package requests

import "math"

type CurrencyConversionRequest struct {
	From  string   `json:"from" validate:"required,oneof=USD BRL" title:"Moeda de origem" description:"Código da moeda de origem" example:"USD"`
	To    string   `json:"to" validate:"required,oneof=USD BRL" title:"Moeda de destino" description:"Código da moeda de destino" example:"BRL"`
	Value *float64 `json:"value" validate:"required,finite,gte=0,finite_conversion" title:"Valor" description:"Valor não negativo a converter" example:"10.00"`
}

const usdToBRLQuote = 6.31

func (r CurrencyConversionRequest) CurrencyQuote() float64 {
	if r.Value == nil {
		return 0
	}

	rate := r.CurrencyRate()
	if rate == 0 {
		return 0
	}

	return roundToCents(*r.Value * rate)
}

func (r CurrencyConversionRequest) CurrencyRate() float64 {
	switch {
	case r.From == "USD" && r.To == "BRL":
		return usdToBRLQuote
	case r.From == "BRL" && r.To == "USD":
		return 1 / usdToBRLQuote
	case r.From == r.To && (r.From == "USD" || r.From == "BRL"):
		return 1
	default:
		return 0
	}
}

func roundToCents(value float64) float64 {
	if math.Abs(value) > math.MaxFloat64/100 {
		return value
	}
	return math.Round(value*100) / 100
}
