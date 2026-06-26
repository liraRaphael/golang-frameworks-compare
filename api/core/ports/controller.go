package ports

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type (
	Controller interface {
		WrapperExecute(ctx context.Context, req requests.Request[any]) (any, error)
	}

	HelloController interface {
		Controller
		Execute(ctx context.Context, req requests.HelloRequest) (responses.HelloOutput, error)
	}

	HealthController interface {
		Controller
	}
)
