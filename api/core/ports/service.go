package ports

import "github.com/liraraphael/go-framework-bench/api/core/domain/responses"

type CepService interface {
	Find(ctx Context, cep string) (responses.CepAddress, error)
}
