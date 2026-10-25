package controller

import (
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type currencyController struct {
	useCase usecases.CurrencyUseCase
	tracer  tracing.Tracer
}

func NewCurrencyController(useCase usecases.CurrencyUseCase) ports.CurrencyController {
	return &currencyController{
		useCase: useCase,
		tracer:  tracing.NewTracer("currency-controller"),
	}
}

func (c *currencyController) Execute(
	ctx ports.Context,
	req requests.Request[requests.CurrencyConversionRequest, any],
) (responses.Response[responses.ConvertCurrencyResponse, any], error) {
	_, span := c.tracer.Start(ctx, "currency-controller.execute")
	defer span.End()
	output, err := c.useCase.Execute(req.Body())
	if err != nil {
		return nil, err
	}
	return responses.NewResponse[responses.ConvertCurrencyResponse, any](output, nil, http.StatusOK), nil
}
