package responses

type ErrorResponse struct {
	Code    string        `json:"codigo"`
	Message string        `json:"mensagem"`
	Details []ErrorDetail `json:"detalhes,omitempty"`
}

type ErrorDetail struct {
	Field string `json:"campo,omitempty"`
	Cause string `json:"causa,omitempty"`
	Value any    `json:"valor,omitempty"`
}
