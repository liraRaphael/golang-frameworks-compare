package presenter

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type helloPresenter struct{}

func NewHelloPresenter() ports.HelloPresenter {
	return &helloPresenter{}
}

func (p *helloPresenter) Output(input requests.HelloRequest) (responses.HelloOutput, error) {
	return responses.HelloOutput{Message: "Hello, " + input.Name}, nil
}
