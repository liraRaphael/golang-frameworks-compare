package controller

import (
	"errors"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/stretchr/testify/assert"
)

type MockHelloUseCase struct {
	ExecuteFunc func(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error)
}

func (m *MockHelloUseCase) Execute(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
	return m.ExecuteFunc(ctx, req)
}

func TestHelloController_Execute(t *testing.T) {
	t.Run("successful execution returns the use case output", func(t *testing.T) {
		expected := responses.HelloOutput{Message: "Hello, Go"}
		useCase := &MockHelloUseCase{
			ExecuteFunc: func(_ ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
				assert.Equal(t, requests.HelloRequest{Name: "Go"}, req)
				return expected, nil
			},
		}
		ctrl := NewHelloController(useCase)
		req := requests.NewRequest[requests.HelloRequest, any](requests.HelloRequest{Name: "Go"}, nil)

		result, err := ctrl.Execute(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, expected, result.Body())
		assert.Equal(t, 200, result.StatusCode())
	})

	t.Run("use case failure is returned", func(t *testing.T) {
		expectedErr := errors.New("use case failed")
		useCase := &MockHelloUseCase{
			ExecuteFunc: func(_ ports.Context, _ requests.HelloRequest) (responses.HelloOutput, error) {
				return responses.HelloOutput{}, expectedErr
			},
		}
		ctrl := NewHelloController(useCase)
		req := requests.NewRequest[requests.HelloRequest, any](requests.HelloRequest{Name: "Go"}, nil)

		result, err := ctrl.Execute(context.Background(), req)

		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
	})
}
