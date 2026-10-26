package requests

// CepPathRequest representa o parâmetro de rota do CEP.
type CepPathRequest struct {
	CEP string `json:"cep" path:"cep" validate:"required,cep" title:"CEP" description:"CEP com 8 dígitos numéricos" example:"01001000"`
}
