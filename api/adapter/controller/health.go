package controller

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type healthController struct {
}

func NewHealthController() ports.HealthController {
	return &healthController{}
}

func (c *healthController) WrapperExecute(ctx context.Context, req requests.Request[any, any]) (any, error) {
	return map[string]string{
		"status": "ok",
	}, nil
}
