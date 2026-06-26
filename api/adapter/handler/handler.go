package handler

import (
	"context"
	"reflect"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type standardHandler struct {
	errorMessageHandlers map[string]enums.ErrorResponseFuncType
	errorTypeHandlers    map[reflect.Type]enums.ErrorResponseFuncType
}

func NewStandardHandler() ports.Handler {
	return &standardHandler{
		errorMessageHandlers: map[string]enums.ErrorResponseFuncType{},
		errorTypeHandlers:    map[reflect.Type]enums.ErrorResponseFuncType{},
	}
}

func (h *standardHandler) RegisterByMessage(message string, fn enums.ErrorResponseFuncType) {
	h.errorMessageHandlers[message] = fn
}

func (h *standardHandler) RegisterByType(err error, fn enums.ErrorResponseFuncType) {
	h.errorTypeHandlers[reflect.TypeOf(err)] = fn
}

func (h *standardHandler) ResolveError(ctx context.Context, err error) *responses.Response {
	if fn, ok := h.errorMessageHandlers[err.Error()]; ok {
		return fn(ctx, err)
	}
	if fn, ok := h.errorTypeHandlers[reflect.TypeOf(err)]; ok {
		return fn(ctx, err)
	}
	return &responses.Response{StatusCode: 500, Body: &responses.ErrorResponse{Code: "internal_error", Message: "Erro interno do servidor não mapeado"}}
}

func (h *standardHandler) Handle(ctx context.Context, successStatusCode int, output any, headers domain.HttpParam) *responses.Response {
	if err, ok := output.(error); ok {
		return h.ResolveError(ctx, err)
	}
	return &responses.Response{StatusCode: successStatusCode, Body: output}
}
