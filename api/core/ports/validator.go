package ports

// Validator define o contrato genérico para validação de dados na aplicação.
type Validator interface {
	Validate(ctx Context, input any) error
}
