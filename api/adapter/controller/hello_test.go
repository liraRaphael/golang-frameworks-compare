package controller

import (
	"errors"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/infra/context"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/stretchr/testify/assert"
)

// MockHelloUseCase implements usecases.HelloUseCase
type MockHelloUseCase struct {
	ExecuteFunc func(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error)
}

func (m *MockHelloUseCase) Execute(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req)
	}
	return responses.HelloOutput{Message: "mocked"}, nil
}

// MockValidator implements ports.Validator
type MockValidator struct {
	ValidateFunc func(ctx ports.Context, input any) error
}

func (m *MockValidator) Validate(ctx ports.Context, input any) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx, input)
	}
	return nil
}

func TestHelloController_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		uc := &MockHelloUseCase{
			ExecuteFunc: func(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
				return responses.HelloOutput{Message: "Hello, " + req.Name}, nil
			},
		}
		val := &MockValidator{
			ValidateFunc: func(ctx ports.Context, input any) error {
				return nil
			},
		}

		ctrl := NewHelloController(uc, val)
		res, err := ctrl.Execute(context.Background(), requests.HelloRequest{Name: "Go"})

		assert.NoError(t, err)
		assert.Equal(t, responses.HelloOutput{Message: "Hello, Go"}, res)
	})

	t.Run("validation failure", func(t *testing.T) {
		uc := &MockHelloUseCase{}
		expectedErr := errors.New("validation failed")
		val := &MockValidator{
			ValidateFunc: func(ctx ports.Context, input any) error {
				return expectedErr
			},
		}

		ctrl := NewHelloController(uc, val)
		_, err := ctrl.Execute(context.Background(), requests.HelloRequest{Name: ""})

		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestHelloController_WrapperExecute(t *testing.T) {
	t.Run("with HelloRequest body", func(t *testing.T) {
		uc := &MockHelloUseCase{
			ExecuteFunc: func(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
				return responses.HelloOutput{Message: "Hello, " + req.Name}, nil
			},
		}
		val := &MockValidator{}

		ctrl := NewHelloController(uc, val)
		req := requests.NewRequestFromParams[any, any](requests.HelloRequest{Name: "Struct"}, nil, nil, nil, nil)
		res, err := ctrl.WrapperExecute(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, responses.HelloOutput{Message: "Hello, Struct"}, res)
	})

	t.Run("with map[string]any body", func(t *testing.T) {
		uc := &MockHelloUseCase{
			ExecuteFunc: func(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
				return responses.HelloOutput{Message: "Hello, " + req.Name}, nil
			},
		}
		val := &MockValidator{}

		ctrl := NewHelloController(uc, val)
		req := requests.NewRequestFromParams[any, any](map[string]any{"name": "Map"}, nil, nil, nil, nil)
		res, err := ctrl.WrapperExecute(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, responses.HelloOutput{Message: "Hello, Map"}, res)
	})

	t.Run("with nil body", func(t *testing.T) {
		uc := &MockHelloUseCase{
			ExecuteFunc: func(ctx ports.Context, req requests.HelloRequest) (responses.HelloOutput, error) {
				return responses.HelloOutput{Message: "Hello, " + req.Name}, nil
			},
		}
		val := &MockValidator{}

		ctrl := NewHelloController(uc, val)
		req := requests.NewRequestFromParams[any, any](nil, nil, nil, nil, nil)
		res, err := ctrl.WrapperExecute(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, responses.HelloOutput{Message: "Hello, "}, res)
	})
}
