package router

import (
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/docs/swagger"
)

type StandardRouter struct {
	adapter     ports.FrameworkAdapter
	routes      []routeRegistration
	middlewares []ports.Middleware
	swagger     *swagger.Registry
}

type routeRegistration struct {
	method  string
	path    string
	handler ports.Controller[any, any]
	meta    swagger.RouteMetadata
}

func NewStandardRouter(adapter ports.FrameworkAdapter) ports.Router {
	return &StandardRouter{adapter: adapter}
}

func (r *StandardRouter) WithSwagger(s *swagger.Registry) *StandardRouter {
	r.swagger = s
	return r
}

func (r *StandardRouter) AddRoute(method string, path string, handler ports.Controller[any, any], meta any) {
	var sMeta swagger.RouteMetadata
	if m, ok := meta.(swagger.RouteMetadata); ok {
		sMeta = m
	} else if mPtr, ok := meta.(*swagger.RouteMetadata); ok && mPtr != nil {
		sMeta = *mPtr
	}

	r.routes = append(r.routes, routeRegistration{method: method, path: path, handler: handler, meta: sMeta})
	if r.swagger != nil {
		_ = r.swagger.AddRoute(method, path, sMeta.Summary, sMeta.Description, sMeta.Tags, sMeta.Input, sMeta.Output, sMeta.Params)
	}
}

func (r *StandardRouter) Use(middleware ports.Middleware) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *StandardRouter) Start(addr string) error {
	for _, middleware := range r.middlewares {
		r.adapter.Use(middleware)
	}
	for _, route := range r.routes {
		r.adapter.RegisterRoute(route.method, route.path, route.handler)
	}
	return r.adapter.Start(addr)
}
