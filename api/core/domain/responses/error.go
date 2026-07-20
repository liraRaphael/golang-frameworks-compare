package responses

type ErrorResponse struct {
	Code    string        `json:"codigo" title:"Código" description:"Código identificador do erro" example:"internal_error"`
	Message string        `json:"mensagem" title:"Mensagem" description:"Mensagem explicativa do erro" example:"Erro interno do servidor"`
	Details []ErrorDetail `json:"detalhes,omitempty" title:"Detalhes" description:"Lista de detalhes do erro"`
}

type ErrorDetail struct {
	Field string `json:"campo,omitempty" title:"Campo" description:"Campo que originou o erro" example:"name"`
	Cause string `json:"causa,omitempty" title:"Causa" description:"Causa detalhada do erro" example:"required"`
	Value any    `json:"valor,omitempty" title:"Valor" description:"Valor enviado que causou o erro"`
}

func NewErrorResponse(code string, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

func (e *ErrorResponse) WithDetails(field string, cause string, value any) *ErrorResponse {
	if e.Details == nil {
		e.Details = []ErrorDetail{}
	}
	e.Details = append(e.Details, NewErrorDetail(field, cause, value))

	return e
}

func NewErrorDetail(field string, cause string, value any) ErrorDetail {
	return ErrorDetail{
		Field: field,
		Cause: cause,
		Value: value,
	}
}
