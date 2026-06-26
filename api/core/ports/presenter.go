package ports

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type (
	HelloPresenter interface {
		Output(input requests.HelloRequest) (responses.HelloOutput, error)
	}
)
