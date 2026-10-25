package controller

import (
	"net/http"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type helloController struct {
	useCase usecases.HelloUseCase
	tracer  tracing.Tracer
}

func NewHelloController(useCase usecases.HelloUseCase) ports.HelloController {
	return &helloController{
		useCase: useCase,
		tracer:  tracing.NewTracer("hello-controller"),
	}
}

func (c *helloController) Execute(ctx ports.Context, req requests.Request[requests.HelloRequest, any]) (responses.Response[responses.HelloOutput, any], error) {
	baseCtx, span := c.tracer.Start(ctx, "hello-controller.execute")
	defer span.End()

	output, err := c.useCase.Execute(context.NewContext(baseCtx), req.Body())
	if err != nil {
		return nil, err
	}
	return responses.NewResponse[responses.HelloOutput, any](output, nil, http.StatusOK), nil
}
