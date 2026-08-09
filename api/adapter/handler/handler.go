package handler

import (
	"net/http"
	"reflect"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type standardHandler struct {
	errorMessageHandlers map[string]ports.ErrorResponseFuncType
	errorTypeHandlers    map[reflect.Type]ports.ErrorResponseFuncType
}

func NewStandardHandler() ports.Handler {
	return &standardHandler{
		errorMessageHandlers: map[string]ports.ErrorResponseFuncType{},
		errorTypeHandlers:    map[reflect.Type]ports.ErrorResponseFuncType{},
	}
}

func (h *standardHandler) RegisterByMessage(message string, fn ports.ErrorResponseFuncType) {
	h.errorMessageHandlers[message] = fn
}

func (h *standardHandler) RegisterByType(err error, fn ports.ErrorResponseFuncType) {
	h.errorTypeHandlers[reflect.TypeOf(err)] = fn
}

func (h *standardHandler) ResolveError(ctx ports.Context, err error) responses.Response[any, any] {
	if fn, ok := h.errorMessageHandlers[err.Error()]; ok {
		return fn(ctx, err)
	}
	if fn, ok := h.errorTypeHandlers[reflect.TypeOf(err)]; ok {
		return fn(ctx, err)
	}

	return responses.NewResponse[any, any](responses.NewErrorResponse("internal_error", "Erro interno do servidor não mapeado"), nil, http.StatusInternalServerError)
}

func (h *standardHandler) Handle(ctx ports.Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any] {
	if err, ok := output.(error); ok {
		return h.ResolveError(ctx, err)
	}
	return responses.NewResponseFromParams[any, any](output, successStatusCode, headers, cookies)
}
