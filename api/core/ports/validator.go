package ports

import "github.com/liraraphael/go-framework-bench/api/core/domain/responses"

type ErrorResponseFuncType func(ctx Context, err error) responses.Response[any, any]

// Validator define o contrato genérico para validação de dados na aplicação.
type Validator interface {
	Validate(ctx Context, input any) error
}
