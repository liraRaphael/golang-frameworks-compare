package ports

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type (
	Controller[BodyType any, ParamType any] interface {
		WrapperExecute(ctx Context, req requests.Request[any, any]) (any, error)
	}

	HelloController interface {
		Controller[requests.HelloRequest, any]
		Execute(ctx Context, req requests.HelloRequest) (responses.HelloOutput, error)
	}

	HealthController interface {
		Controller[any, any]
	}
)
