package usecases

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type HelloUseCase interface {
	Execute(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error)
}

type helloUseCase struct {
	presenter ports.HelloPresenter
}

func NewHelloUseCase(presenter ports.HelloPresenter) HelloUseCase {
	return &helloUseCase{presenter: presenter}
}

func (u *helloUseCase) Execute(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
	return u.presenter.Output(req)
}
