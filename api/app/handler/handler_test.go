package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/infra/context"

	domainerrors "github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/stretchr/testify/assert"
)

type customTestError struct {
	msg string
}

func (e *customTestError) Error() string {
	return e.msg
}

func TestStandardHandler_Handle(t *testing.T) {
	h := NewStandardHandler()
	ctx := context.Background()

	t.Run("success response handling", func(t *testing.T) {
		output := map[string]string{"result": "ok"}
		resp := h.Handle(ctx, http.StatusOK, output, nil, nil)

		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, output, resp.Body())
	})

	t.Run("unmapped error returns 500 internal_error", func(t *testing.T) {
		err := errors.New("something went wrong")
		resp := h.Handle(ctx, http.StatusOK, err, nil, nil)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())

		body, ok := resp.Body().(*responses.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, "internal_error", body.Code)
		assert.Equal(t, "Erro interno do servidor não mapeado", body.Message)
	})

	t.Run("request errors are mapped to 400 without exposing internals", func(t *testing.T) {
		resp := h.Handle(ctx, http.StatusOK, domainerrors.NewRequestError(errors.New("parser details")), nil, nil)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		body, ok := resp.Body().(*responses.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, "invalid_request", body.Code)
		assert.NotContains(t, body.Message, "parser details")
	})

	t.Run("validation errors include field details and return 400", func(t *testing.T) {
		err := &domainerrors.FieldValidationError{Errors: []domainerrors.FieldError{{
			Field: "Value",
			Tag:   "gte",
			Value: -1.0,
		}}}
		resp := h.Handle(ctx, http.StatusOK, err, nil, nil)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
		body, ok := resp.Body().(*responses.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, "validation_error", body.Code)
		assert.Equal(t, "Value", body.Details[0].Field)
		assert.Equal(t, "gte", body.Details[0].Cause)
	})

	t.Run("register and resolve by error message", func(t *testing.T) {
		hCustom := NewStandardHandler()
		targetErr := errors.New("user not found")

		hCustom.RegisterByMessage("user not found", func(ctx ports.Context, err error) responses.Response[any, any] {
			return responses.NewResponse[any, any](
				&responses.ErrorResponse{Code: "not_found", Message: err.Error()},
				nil,
				http.StatusNotFound,
			)
		})

		resp := hCustom.Handle(ctx, http.StatusOK, targetErr, nil, nil)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode())
		body, ok := resp.Body().(*responses.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, "not_found", body.Code)
		assert.Equal(t, "user not found", body.Message)
	})

	t.Run("register and resolve by error type", func(t *testing.T) {
		hCustom := NewStandardHandler()
		targetErr := &customTestError{msg: "some type error"}

		hCustom.RegisterByType(&customTestError{}, func(ctx ports.Context, err error) responses.Response[any, any] {
			return responses.NewResponse[any, any](
				&responses.ErrorResponse{Code: "conflict_type", Message: err.Error()},
				nil,
				http.StatusConflict,
			)
		})

		resp := hCustom.Handle(ctx, http.StatusOK, targetErr, nil, nil)

		assert.Equal(t, http.StatusConflict, resp.StatusCode())
		body, ok := resp.Body().(*responses.ErrorResponse)
		assert.True(t, ok)
		assert.Equal(t, "conflict_type", body.Code)
		assert.Equal(t, "some type error", body.Message)
	})
}
