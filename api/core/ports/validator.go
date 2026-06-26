package ports

import "context"

// Validator define o contrato genérico para validação de dados na aplicação.
type Validator interface {
	Validate(ctx context.Context, input any) error
}
