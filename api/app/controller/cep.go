package controller

import (
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type cepController struct {
	service ports.CepService
	tracer  tracing.Tracer
}

func NewCEPController(service ports.CepService) ports.CEPController {
	return &cepController{
		service: service,
		tracer:  tracing.NewTracer("cep-controller"),
	}
}

func (c *cepController) Execute(ctx ports.Context, req requests.Request[any, requests.CepPathRequest]) (responses.Response[responses.CepAddress, any], error) {
	_, span := c.tracer.Start(ctx, "cep-controller.execute")
	defer span.End()

	address, err := c.service.Find(ctx, req.Params().CEP)
	if err != nil {
		return nil, err
	}

	return responses.NewResponse[responses.CepAddress, any](address, nil, http.StatusOK), nil
}
