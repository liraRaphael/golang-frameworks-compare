package usecases

import (
	"context"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/adapter/presenter"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/stretchr/testify/assert"
)

func TestHelloUseCaseExecuteReturnsGreeting(t *testing.T) {
	uc := NewHelloUseCase(presenter.NewHelloPresenter())

	resp, err := uc.Execute(context.Background(), requests.HelloRequest{Name: "World"})

	assert.Equal(t, "Hello, World", resp.Message)
	assert.NoError(t, err)
}
