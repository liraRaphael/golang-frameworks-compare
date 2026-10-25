package router

import (
	"encoding/json"
	"errors"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	domainerrors "github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type typedRequest[Body, Params any] struct {
	body   Body
	params Params
	source requests.Request[any, any]
}

type typedResponse struct {
	body       any
	params     any
	statusCode int
	headers    domain.HttpParamsType
	cookies    domain.HttpParamsType
}

func (r typedResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Body       any                   `json:"Body"`
		StatusCode int                   `json:"StatusCode"`
		Headers    domain.HttpParamsType `json:"Headers"`
	}{
		Body:       r.body,
		StatusCode: r.statusCode,
		Headers:    r.Headers(),
	})
}

func (r typedResponse) Body() any                      { return r.body }
func (r typedResponse) Params() any                    { return r.params }
func (r typedResponse) Headers() domain.HttpParamsType { return r.headers }
func (r typedResponse) Cookies() domain.HttpParamsType { return r.cookies }
func (r typedResponse) StatusCode() int                { return r.statusCode }

func (r typedRequest[Body, Params]) Body() Body                     { return r.body }
func (r typedRequest[Body, Params]) Params() Params                 { return r.params }
func (r typedRequest[Body, Params]) Headers() domain.HttpParamsType { return r.source.Headers() }
func (r typedRequest[Body, Params]) QueryParams() domain.HttpParamsType {
	return r.source.QueryParams()
}
func (r typedRequest[Body, Params]) PathParams() domain.HttpParamsType { return r.source.PathParams() }
func (r typedRequest[Body, Params]) Cookies() domain.HttpParamsType    { return r.source.Cookies() }

type typedControllerAdapter[InputBody, InputParams, OutputBody, OutputParams any] struct {
	controller ports.Controller[InputBody, InputParams, OutputBody, OutputParams]
	validator  ports.Validator
}

// AddTypedRoute adapts a strongly typed controller to the framework's
// non-generic route interface.
func AddTypedRoute[InputBody, InputParams, OutputBody, OutputParams any](
	r ports.Router,
	method string,
	path string,
	controller ports.Controller[InputBody, InputParams, OutputBody, OutputParams],
	validator ports.Validator,
	meta any,
) {
	r.AddRoute(method, path, typedControllerAdapter[InputBody, InputParams, OutputBody, OutputParams]{
		controller: controller,
		validator:  validator,
	}, meta)
}

func (a typedControllerAdapter[InputBody, InputParams, OutputBody, OutputParams]) Execute(
	ctx ports.Context,
	req requests.Request[any, any],
) (responses.Response[any, any], error) {
	if a.validator == nil {
		return nil, errors.New("typed route validator is not configured")
	}

	body, err := convertValue[InputBody](req.Body())
	if err != nil {
		return nil, domainerrors.NewRequestError(err)
	}

	params, err := convertValue[InputParams](req.Params())
	if err != nil {
		return nil, domainerrors.NewRequestError(err)
	}

	typedReq := typedRequest[InputBody, InputParams]{
		body:   body,
		params: params,
		source: req,
	}

	if any(body) != nil {
		if err := a.validator.Validate(ctx, body); err != nil {
			return nil, err
		}
	}
	if any(params) != nil {
		if err := a.validator.Validate(ctx, params); err != nil {
			return nil, err
		}
	}

	result, err := a.controller.Execute(ctx, typedReq)
	if err != nil {
		return nil, err
	}

	return typedResponse{
		body:       result.Body(),
		params:     result.Params(),
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
