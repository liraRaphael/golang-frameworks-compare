package ports

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type FrameworkAdapter interface {
	RegisterRoute(method string, path string, ctrl Controller[any, any])
	Use(middleware Middleware)
	Start(addr string) error
}

type Router interface {
	AddRoute(method string, path string, handler Controller[any, any], meta any)
	Group(prefix string, middlewares ...Middleware) Router
	Use(middleware Middleware)
	Start(addr string) error
}

type Handler interface {
	Handle(ctx context.Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any]
	RegisterByMessage(message string, fn responses.ErrorResponseFuncType)
	RegisterByType(err error, fn responses.ErrorResponseFuncType)
	ResolveError(ctx context.Context, err error) responses.Response[any, any]
}
