package requests

import (
	"reflect"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
)

type (
	Request[BodyType any, ParamType any] interface {
		Body() BodyType
		Params() ParamType
		Headers() domain.HttpParamsType
		QueryParams() domain.HttpParamsType
		PathParams() domain.HttpParamsType
		Cookies() domain.HttpParamsType
	}

	Requester[BodyType any, ParamType any] struct {
		body   BodyType  `validate:"omitempty,dive"`
		params ParamType `validate:"omitempty,dive"`

		headersParamsFields []int
		queryParamsFields   []int
		pathParamsFields    []int
		cookiesParamsFields []int
	}
)

func NewRequest[BodyType any, ParamType any](body BodyType, params ParamType) Request[BodyType, ParamType] {
	r := Requester[BodyType, ParamType]{
		body:   body,
		params: params,
	}

	r.headersParamsFields, r.queryParamsFields, r.pathParamsFields, r.cookiesParamsFields = r.extractParamsFromTag()

	return r
}

func NewRequestFromParams[BodyType any, ParamType any](body BodyType, headers, queryParams, pathParams, cookies domain.HttpParamsType) Request[BodyType, ParamType] {
	params, err := domain.PopulateParams[ParamType](headers, queryParams, pathParams, cookies)
	if err != nil {
		panic(err)
	}

	r := Requester[BodyType, ParamType]{
		body:   body,
		params: params,
	}

	r.headersParamsFields, r.queryParamsFields, r.pathParamsFields, r.cookiesParamsFields = r.extractParamsFromTag()

	return r
}

func (r Requester[BodyType, ParamType]) Body() BodyType {
	return r.body
}

func (r Requester[BodyType, ParamType]) Headers() domain.HttpParamsType {
	return r.getParamsValuesInKnowsFields(r.headersParamsFields)
}

func (r Requester[BodyType, ParamType]) Cookies() domain.HttpParamsType {
	return r.getParamsValuesInKnowsFields(r.cookiesParamsFields)
}

func (r Requester[BodyType, ParamType]) QueryParams() domain.HttpParamsType {
	return r.getParamsValuesInKnowsFields(r.queryParamsFields)
}

func (r Requester[BodyType, ParamType]) PathParams() domain.HttpParamsType {
	return r.getParamsValuesInKnowsFields(r.pathParamsFields)
}

func (r Requester[BodyType, ParamType]) Params() ParamType {
	return r.params
}

func (r Requester[BodyType, ParamType]) extractParamsFromTag() (headersFields []int, queryFields []int, pathFields []int, cookieFields []int) {
	var param ParamType
	return domain.ExtractParamsFromTag(reflect.TypeOf(param))
}

func (r Requester[BodyType, ParamType]) getParamsValuesInKnowsFields(indexs []int) domain.HttpParamsType {
	vOf := reflect.ValueOf(r.params)
	return domain.ParamsValuesToHttpParamsType(vOf, indexs)
}
