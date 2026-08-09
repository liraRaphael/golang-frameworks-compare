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

func (c *helloController) Execute(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
	baseCtx, span := c.tracer.Start(ctx, "hello-controller.execute")
	defer span.End()

	ctx = context.NewContext(baseCtx)

	if err := c.validator.Validate(ctx, req); err != nil {
		return responses.HelloOutput{}, err
	}

	return c.useCase.Execute(ctx, req)
}

func (c *helloController) WrapperExecute(ctx ports.Context, req requests.Request[any, any]) (any, error) {
	var body requests.HelloRequest

	// Handle nil or empty body
	if req.Body() != nil {
		if b, ok := req.Body().(requests.HelloRequest); ok {
			body = b
		} else if m, ok := req.Body().(map[string]any); ok {
			// Basic conversion for generic input
			if name, ok := m["name"].(string); ok {
				body.Name = name
			}
		}
	}

	return c.Execute(ctx, body)
}
