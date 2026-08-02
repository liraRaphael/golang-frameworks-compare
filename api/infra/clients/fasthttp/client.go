package fasthttp

import (
	"encoding/json"
	"time"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/valyala/fasthttp"
)

type fastHttpClient struct {
	client *fasthttp.Client
}

func NewFastHttpClient(timeout time.Duration) ports.Client {
	return &fastHttpClient{
		client: &fasthttp.Client{
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},
	}
}

func (c *fastHttpClient) Do(ctx ports.Context, method enums.HttpMethod, url string, req requests.Request[any, any]) (responses.Response[any, any], error) {
	fReq := fasthttp.AcquireRequest()
	fResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(fReq)
	defer fasthttp.ReleaseResponse(fResp)

	fReq.SetRequestURI(url)
	fReq.Header.SetMethod(string(method))

	if req.Headers() != nil {
		for k, v := range req.Headers() {
			for _, val := range v {
				fReq.Header.Add(k, val)
			}
		}
	}

	if req.Body() != nil {
		b, err := json.Marshal(req.Body())
		if err != nil {
			return nil, &errors.ClientError{Message: "failed to marshal request body", Err: err}
		}
		fReq.SetBody(b)
		if fReq.Header.ContentType() == nil {
			fReq.Header.SetContentType("application/json")
		}
	}

	var err error
	if deadline, ok := ctx.Deadline(); ok {
		err = c.client.DoDeadline(fReq, fResp, deadline)
	} else {
		err = c.client.Do(fReq, fResp)
	}

	if err != nil {
		return nil, &errors.ClientError{Message: "failed to execute request", Err: err}
	}

	respBody := fResp.Body()
	statusCode := fResp.StatusCode()

	headers := domain.HttpParamsType{}
	fResp.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		headers[k] = append(headers[k], string(value))
	})

	var responseBody any
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &responseBody); err != nil {
			responseBody = string(respBody)
		}
	}

	return responses.NewResponseFromParams[any, any](responseBody, statusCode, headers, nil), nil
}
