package handler

import (
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type standardRouter struct {
	adapter     ports.FrameworkAdapter
	routes      []routeRegistration
	middlewares []ports.Middleware
}

type routeRegistration struct {
	method  string
	path    string
	handler ports.Controller
	meta    routeMetadata
}

func NewStandardRouter(adapter ports.FrameworkAdapter) *standardRouter {
	return &standardRouter{adapter: adapter}
}

func (r *standardRouter) AddRoute(method string, path string, handler ports.Controller) {
	r.routes = append(r.routes, routeRegistration{method: method, path: path, handler: handler, meta: routeMetadata{}})
}

func (r *standardRouter) Use(middleware ports.Controller) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *standardRouter) Start(addr string) error {
	for _, middleware := range r.middlewares {
		r.adapter.Use(middleware)
	}
	for _, route := range r.routes {
		r.adapter.RegisterRoute(route.method, route.path, route.handler)
	}
	return r.adapter.Start(addr)
}
