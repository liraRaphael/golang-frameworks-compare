package router

import (
	"encoding/json"
	"fmt"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type typedRequest[Body, Params any] struct {
	body   Body
	params Params
	source requests.Request[any, any]
}

func (r typedRequest[Body, Params]) Body() Body                     { return r.body }
func (r typedRequest[Body, Params]) Params() Params                 { return r.params }
func (r typedRequest[Body, Params]) Headers() domain.HttpParamsType { return r.source.Headers() }
func (r typedRequest[Body, Params]) QueryParams() domain.HttpParamsType {
	return r.source.QueryParams()
}
func (r typedRequest[Body, Params]) PathParams() domain.HttpParamsType { return r.source.PathParams() }
func (r typedRequest[Body, Params]) Cookies() domain.HttpParamsType    { return r.source.Cookies() }

type typedResponse struct {
	body       any
	statusCode int
	headers    domain.HttpParamsType
	cookies    domain.HttpParamsType
}

func (r typedResponse) Body() any                      { return r.body }
func (r typedResponse) Params() any                    { return nil }
func (r typedResponse) Headers() domain.HttpParamsType { return r.headers }
func (r typedResponse) Cookies() domain.HttpParamsType { return r.cookies }
func (r typedResponse) StatusCode() int                { return r.statusCode }

type typedControllerAdapter[InputBody, InputParams, OutputBody, OutputParams any] struct {
	controller ports.Controller[InputBody, InputParams, OutputBody, OutputParams]
}

// AddTypedRoute adapts a strongly typed controller to the framework's
// non-generic route interface.
func AddTypedRoute[InputBody, InputParams, OutputBody, OutputParams any](
	r ports.Router,
	method string,
	path string,
	controller ports.Controller[InputBody, InputParams, OutputBody, OutputParams],
	meta any,
) {
	r.AddRoute(method, path, typedControllerAdapter[InputBody, InputParams, OutputBody, OutputParams]{
		controller: controller,
	}, meta)
}

func (a typedControllerAdapter[InputBody, InputParams, OutputBody, OutputParams]) Execute(
	ctx ports.Context,
	req requests.Request[any, any],
) (responses.Response[any, any], error) {
	body, err := convertValue[InputBody](req.Body())
	if err != nil {
		return nil, fmt.Errorf("converter corpo da requisição: %w", err)
	}

	params, err := convertValue[InputParams](req.Params())
	if err != nil {
		return nil, fmt.Errorf("converter parâmetros da requisição: %w", err)
	}

	typedReq := typedRequest[InputBody, InputParams]{
		body:   body,
		params: params,
		source: req,
	}

	result, err := a.controller.Execute(ctx, typedReq)
	if err != nil {
		return nil, err
	}

	return typedResponse{
		body:       result.Body(),
		statusCode: result.StatusCode(),
		headers:    result.Headers(),
		cookies:    result.Cookies(),
	}, nil
}

func convertValue[T any](value any) (T, error) {
	var result T

	if value == nil {
		return result, nil
	}
	if typed, ok := value.(T); ok {
		return typed, nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}

	return result, nil
}
