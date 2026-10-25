package responses

type ConvertCurrencyResponse struct {
	Value float64 `json:"value" title:"Valor convertido" description:"Valor arredondado para duas casas decimais" example:"63.10"`
	From  string  `json:"from" title:"Moeda de origem" example:"USD"`
	To    string  `json:"to" title:"Moeda de destino" example:"BRL"`
	Quote float64 `json:"quote" title:"Cotação aplicada" description:"Multiplicador usado na conversão" example:"6.31"`
}
