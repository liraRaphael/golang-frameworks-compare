package controller

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/tracing"
)

type helloController struct {
	useCase   usecases.HelloUseCase
	validator ports.Validator
	tracer    tracing.Tracer
}

func NewHelloController(useCase usecases.HelloUseCase, validator ports.Validator) ports.HelloController {
	return &helloController{
		useCase:   useCase,
		validator: validator,
		tracer:    tracing.NewTracer("hello-controller"),
	}
}

func (c *helloController) Execute(ctx ports.Context, req requests.Request[requests.HelloRequest, any]) (responses.Response[responses.HelloOutput, any], error) {
	baseCtx, span := c.tracer.Start(ctx, "hello-controller.execute")
	defer span.End()

	responser := c.newResponser()

	ctx = context.NewContext(baseCtx)

	if err := c.validator.Validate(ctx, req); err != nil {
		return responser, err
	}

	ex, err := c.useCase.Execute(ctx, req.Body())
	if err != nil {
		return responser, err
	}

	return c.newResponserBody(ex), nil
}

func (c *helloController) newResponserBody(b responses.HelloOutput) responses.Response[responses.HelloOutput, any] {
	return responses.NewResponse(b, interface{}(nil), 200)
}

func (c *helloController) newResponser() responses.Response[responses.HelloOutput, any] {
	return c.newResponserBody(responses.HelloOutput{})
}
