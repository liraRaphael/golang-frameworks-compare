package nethttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
)

type httpClient struct {
	client *http.Client
}

func NewHttpClient(timeout time.Duration) ports.Client {
	return &httpClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *httpClient) Do(ctx context.Context, method enums.HttpMethod, url string, req requests.Request[any, any]) (*responses.Response[any], error) {
	var bodyReader io.Reader
	if req.Body() != nil {
		b, err := json.Marshal(req.Body())
		if err != nil {
			return nil, &errors.ClientError{Message: "failed to marshal request body", Err: err}
		}
		bodyReader = bytes.NewReader(b)
	}

	httpReq, err := http.NewRequestWithContext(ctx, string(method), url, bodyReader)
	if err != nil {
		return nil, &errors.ClientError{Message: "failed to create request", Err: err}
	}

	if req.Headers() != nil {
		for k, v := range req.Headers().GetAll() {
			for _, val := range v {
				httpReq.Header.Add(k, val)
			}
		}
	}

	if req.Body() != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, &errors.ClientError{Message: "failed to execute request", Err: err}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &errors.ClientError{StatusCode: resp.StatusCode, Message: "failed to read response body", Err: err}
	}

	headers := domain.NewHttpParam()
	headers.SetAll(resp.Header)

	var responseBody any
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &responseBody); err != nil {
			// If it's not JSON, we keep it as raw bytes or string if possible,
			// but for this generic 'any', unmarshal is the standard attempt.
			responseBody = string(respBody)
		}
	}

	return &responses.Response[any]{
		Body:       responseBody,
		StatusCode: resp.StatusCode,
		Headers:    headers,
	}, nil
}
