package ports

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type Client interface {
	Do(ctx context.Context, method enums.HttpMethod, url string, req requests.Request[any, any]) (*responses.Response[any], error)
}

func NewResponse[T, Param any](body T, statusCode int, headers domain.HttpParamsType) *responses.Response[T] {
	return &responses.Response[T]{
		Body:       body,
		StatusCode: statusCode,
		Headers:    headers,
	}
}
