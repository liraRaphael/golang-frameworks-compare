package ports

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type (
	Controller[InputBodyType, InputParamType, OutputBodyType, OutputParamType any] interface {
		Execute(ctx Context, req requests.Request[InputBodyType, InputParamType]) (responses.Response[OutputBodyType, OutputParamType], error)
	}

	HelloController interface {
		Controller[requests.HelloRequest, any, responses.HelloOutput, any]
	}

	CurrencyController interface {
		Controller[requests.CurrencyConversionRequest, any, responses.ConvertCurrencyResponse, any]
	}

	HealthController interface {
		Controller[any, any, any, any]
	}
)
