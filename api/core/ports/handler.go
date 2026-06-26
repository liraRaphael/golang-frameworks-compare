package ports

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type FrameworkAdapter interface {
	RegisterRoute(method string, path string, ctrl Controller)
	Use(middleware Middleware)
	Start(addr string) error
}

type Router interface {
	AddRoute(method string, path string, handler Controller)
	Use(middleware Middleware)
	Start(addr string) error
}

type Handler interface {
	Handle(ctx context.Context, successStatusCode int, output any, headers domain.HttpParam) *responses.Response
	RegisterByMessage(message string, fn enums.ErrorResponseFuncType)
	RegisterByType(err error, fn enums.ErrorResponseFuncType)
	ResolveError(ctx context.Context, err error) *responses.Response
}
