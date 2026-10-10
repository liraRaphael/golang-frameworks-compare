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

type MockValidator struct {
	ValidateFunc func(ctx ports.Context, input any) error
}

func (m *MockValidator) Validate(ctx ports.Context, input any) error {
	return m.ValidateFunc(ctx, input)
}

func TestHelloController_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		expectedBody := responses.HelloOutput{Message: "Hello, Go"}
		useCaseCalled := false
		uc := &MockHelloUseCase{
			ExecuteFunc: func(_ ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
				useCaseCalled = true
				assert.Equal(t, requests.HelloRequest{Name: "Go"}, req)
				return expectedBody, nil
			},
		}
		validatorCalled := false
		val := &MockValidator{
			ValidateFunc: func(_ ports.Context, input any) error {
				validatorCalled = true
				return nil
			},
		}
		ctrl := NewHelloController(uc, val)
		req := requests.NewRequest[requests.HelloRequest, any](requests.HelloRequest{Name: "Go"}, nil)

		res, err := ctrl.Execute(context.Background(), req)

		assert.NoError(t, err)
		assert.True(t, validatorCalled)
		assert.True(t, useCaseCalled)
		assert.Equal(t, expectedBody, res.Body())
		assert.Equal(t, 200, res.StatusCode())
		assert.Nil(t, res.Params())
	})

	t.Run("validation failure", func(t *testing.T) {
		expectedErr := errors.New("validation failed")
		useCaseCalled := false
		uc := &MockHelloUseCase{
			ExecuteFunc: func(_ ports.Context, _ requests.HelloRequest) (responses.HelloOutput, error) {
				useCaseCalled = true
				return responses.HelloOutput{}, nil
			},
		}
		val := &MockValidator{
			ValidateFunc: func(_ ports.Context, _ any) error {
				return expectedErr
			},
		}
		ctrl := NewHelloController(uc, val)
		req := requests.NewRequest[requests.HelloRequest, any](requests.HelloRequest{}, nil)

		res, err := ctrl.Execute(context.Background(), req)

		assert.ErrorIs(t, err, expectedErr)
		assert.False(t, useCaseCalled)
		assert.Equal(t, responses.HelloOutput{}, res.Body())
		assert.Equal(t, 200, res.StatusCode())
	})

	t.Run("use case failure", func(t *testing.T) {
		expectedErr := errors.New("use case failed")
		uc := &MockHelloUseCase{
			ExecuteFunc: func(_ ports.Context, _ requests.HelloRequest) (responses.HelloOutput, error) {
				return responses.HelloOutput{}, expectedErr
			},
		}
		val := &MockValidator{
			ValidateFunc: func(_ ports.Context, _ any) error {
				return nil
			},
		}
		ctrl := NewHelloController(uc, val)
		req := requests.NewRequest[requests.HelloRequest, any](requests.HelloRequest{Name: "Go"}, nil)

		res, err := ctrl.Execute(context.Background(), req)

		assert.ErrorIs(t, err, expectedErr)
		assert.Equal(t, responses.HelloOutput{}, res.Body())
		assert.Equal(t, 200, res.StatusCode())
	})
}
