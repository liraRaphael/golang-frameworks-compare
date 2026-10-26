package errors

import "fmt"

// CEPInvalidError indica que o CEP informado está fora do formato esperado.
type CEPInvalidError struct {
	CEP string
}

func (e *CEPInvalidError) Error() string {
	return fmt.Sprintf("invalid cep: %s", e.CEP)
}

// CEPNotFoundError indica que o ViaCEP respondeu com um CEP inexistente.
type CEPNotFoundError struct {
	CEP string
}

func (e *CEPNotFoundError) Error() string {
	return fmt.Sprintf("cep not found: %s", e.CEP)
}

// CEPUpstreamError indica falha na dependência externa.
type CEPUpstreamError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *CEPUpstreamError) Error() string {
	if e == nil {
		return "cep upstream error"
	}
	if e.Err != nil {
		return fmt.Sprintf("cep upstream error: %s (status: %d): %v", e.Message, e.StatusCode, e.Err)
	}
	return fmt.Sprintf("cep upstream error: %s (status: %d)", e.Message, e.StatusCode)
}
