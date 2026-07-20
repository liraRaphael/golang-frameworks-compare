package responses

import (
	"context"
	"reflect"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
)

type ErrorResponseFuncType func(ctx context.Context, err error) Response[any, any]

type (
	Response[BodyType any, ParamType any] interface {
		Body() BodyType
		Params() ParamType
		Headers() domain.HttpParamsType
		Cookies() domain.HttpParamsType
		StatusCode() int
	}

	Responser[BodyType any, ParamType any] struct {
		body   BodyType
		params ParamType

		statusCode int

		headersParamsFields []int
		cookiesParamsFields []int
	}
)

func NewResponse[BodyType any, ParamType any](body BodyType, params ParamType, statusCode int) Response[BodyType, ParamType] {
	r := Responser[BodyType, ParamType]{
		body:       body,
		params:     params,
		statusCode: statusCode,
	}

	r.headersParamsFields, _, _, r.cookiesParamsFields = r.extractParamsFromTag()

	return r
}

func NewResponseFromParams[BodyType any, ParamType any](body BodyType, statusCode int, headers, cookies domain.HttpParamsType) Response[BodyType, ParamType] {
	params, err := domain.PopulateParams[ParamType](headers, nil, nil, cookies)
	if err != nil {
		panic(err)
	}

	r := Responser[BodyType, ParamType]{
		body:   body,
		params: params,
	}

	r.headersParamsFields, _, _, r.cookiesParamsFields = r.extractParamsFromTag()

	return r
}

func (r Responser[BodyType, ParamType]) Body() BodyType {
	return r.body
}

func (r Responser[BodyType, ParamType]) Headers() domain.HttpParamsType {
	return r.getParamsValuesInKnowsFields(r.headersParamsFields)
}

func (r Responser[BodyType, ParamType]) Cookies() domain.HttpParamsType {
	return r.getParamsValuesInKnowsFields(r.cookiesParamsFields)
}

func (r Responser[BodyType, ParamType]) StatusCode() int {
	return r.statusCode
}

func (r Responser[BodyType, ParamType]) Params() ParamType {
	return r.params
}

func (r Responser[BodyType, ParamType]) extractParamsFromTag() (headersFields []int, queryFields []int, pathFields []int, cookieFields []int) {
	var param ParamType
	return domain.ExtractParamsFromTag(reflect.TypeOf(param))
}

func (r Responser[BodyType, ParamType]) getParamsValuesInKnowsFields(indexs []int) domain.HttpParamsType {
	vOf := reflect.ValueOf(r.params)
	return domain.ParamsValuesToHttpParamsType(vOf, indexs)
}
