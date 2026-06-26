package requests

import "github.com/liraraphael/go-framework-bench/api/core/domain"

type (
	Request[BodyType any] interface {
		Body() BodyType
		Headers() domain.HttpParam
		QueryParams() domain.HttpParam
		PathParams() domain.HttpParam
	}

	Requester[BodyType any] struct {
		body        BodyType         `json:"body,omitempty"`
		headers     domain.HttpParam `json:"headers,omitempty"`
		queryParams domain.HttpParam `json:"query_params,omitempty"`
		pathParams  domain.HttpParam `json:"path_params,omitempty"`
	}
)

func NewRequest[BodyType any](body BodyType, headers, queryParams, pathParams domain.HttpParam) Request[BodyType] {
	return Requester[BodyType]{
		body:        body,
		headers:     headers,
		queryParams: queryParams,
		pathParams:  pathParams,
	}
}

func (r Requester[BodyType]) Body() BodyType {
	return r.body
}

func (r Requester[BodyType]) Headers() domain.HttpParam {
	return r.headers
}

func (r Requester[BodyType]) QueryParams() domain.HttpParam {
	return r.queryParams
}

func (r Requester[BodyType]) PathParams() domain.HttpParam {
	return r.pathParams
}
