package controller

import (
	"net/http"

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
	return responses.NewResponse[any, any](map[string]string{
		"status": "ok",
	}, nil, http.StatusOK), nil
}
