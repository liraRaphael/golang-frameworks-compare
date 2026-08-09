package ports

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/infra/docs/swagger"
)

type FrameworkAdapter interface {
	RegisterRoute(method string, path string, ctrl Controller[any, any])
	Use(middleware Middleware)
	Start(addr string) error
}

type Router interface {
	WithSwagger(s *swagger.Registry)

	AddRoute(method string, path string, handler Controller[any, any], meta any)
	Use(middleware Middleware)
	Start(addr string) error
}

type Handler interface {
	Handle(ctx Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any]
	RegisterByMessage(message string, fn ErrorResponseFuncType)
	RegisterByType(err error, fn ErrorResponseFuncType)
	ResolveError(ctx Context, err error) responses.Response[any, any]
}
