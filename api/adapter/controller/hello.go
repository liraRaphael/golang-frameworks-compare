package controller

import (
	"context"
	"errors"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
)

type helloController struct {
	useCase usecases.HelloUseCase
}

func NewHelloController(useCase usecases.HelloUseCase) ports.HelloController {
	return &helloController{useCase: useCase}
}

func (c *helloController) Execute(ctx context.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
	out, err := c.useCase.Execute(ctx, req)
	return out, err
}

func (c *helloController) WrapperExecute(ctx context.Context, req requests.Request[any]) (any, error) {
	r, ok := req.(requests.Request[requests.HelloRequest])
	if !ok {
		panic(errors.ErrUnsupported)
	}
	return c.Execute(ctx, r.Body())
}
