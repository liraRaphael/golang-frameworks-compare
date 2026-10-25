package fiber

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/app/handler"
	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/frameworks/testsupport"
	"github.com/stretchr/testify/assert"
)

type dummyController struct {
	fn func(ctx ports.Context, req requests.Request[any, any]) (any, error)
}

func (d *dummyController) Execute(ctx ports.Context, req requests.Request[any, any]) (responses.Response[any, any], error) {
	output, err := d.fn(ctx, req)
	if err != nil {
		return nil, err
	}
	return responses.NewResponse[any, any](output, nil, http.StatusOK), nil
}

type dummyHandler struct{}

func (d *dummyHandler) Handle(ctx ports.Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any] {
	return responses.NewResponse[any, any](output, nil, successStatusCode)
}
func (d *dummyHandler) RegisterByMessage(message string, fn ports.ErrorResponseFuncType) {}
func (d *dummyHandler) RegisterByType(err error, fn ports.ErrorResponseFuncType)         {}
func (d *dummyHandler) ResolveError(ctx ports.Context, err error) responses.Response[any, any] {
	return responses.NewResponse[any, any](err.Error(), nil, 500)
}

func TestFiberAdapter(t *testing.T) {
	dh := &dummyHandler{}
	adapter := NewAdapter(dh).(*adapter)

	ctrl := &dummyController{
		fn: func(ctx ports.Context, req requests.Request[any, any]) (any, error) {
			assert.Equal(t, []string{"abc"}, req.Cookies().GetValues("session"))
			values := req.PathParams().GetValues("id")
			var param string
			if len(values) > 0 {
				param = values[0]
			}

			return map[string]string{"id": param}, nil
		},
	}

	adapter.RegisterRoute("GET", "/test/:id", ctrl)

	req := httptest.NewRequest("GET", "/test/123", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc"})
	resp, err := adapter.app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var apiResp struct {
		Body       map[string]string `json:"Body"`
		StatusCode int               `json:"StatusCode"`
	}
	err = json.Unmarshal(bodyBytes, &apiResp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, apiResp.StatusCode)
	assert.Equal(t, "123", apiResp.Body["id"])
}

func TestFiberCurrencyContract(t *testing.T) {
	a := NewAdapter(handler.NewStandardHandler()).(*adapter)
	testsupport.TestCurrencyContract(t, a, func(req *http.Request) (*http.Response, error) {
		if len(req.TransferEncoding) > 0 {
			req = req.Clone(req.Context())
			req.ContentLength = 0
		}
		return a.app.Test(req, -1)
	})
}
