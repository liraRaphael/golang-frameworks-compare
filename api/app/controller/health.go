package controller

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type healthController struct {
}

func NewHealthController() ports.HealthController {
	return &healthController{}
}

func (c *healthController) Execute(ctx ports.Context, req requests.Request[any, any]) (responses.Response[any, any], error) {
	responser := map[string]string{
		"status": "ok",
	}
	return responses.NewResponse[any, any](responser, interface{}(nil), 200), nil
}
