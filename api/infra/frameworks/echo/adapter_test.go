package echo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/stretchr/testify/assert"
)

type dummyController struct {
	fn func(ctx context.Context, req requests.Request[any, any]) (any, error)
}

func (d *dummyController) WrapperExecute(ctx ports.Context, req requests.Request[any, any]) (any, error) {
	return d.fn(ctx, req)
}

type dummyHandler struct{}

func (d *dummyHandler) Handle(ctx ports.Context, successStatusCode int, output any, headers domain.HttpParamsType, cookies domain.HttpParamsType) responses.Response[any, any] {
	return responses.NewResponse[any, any](output, nil, successStatusCode)
}
func (d *dummyHandler) RegisterByMessage(message string, fn responses.ErrorResponseFuncType) {}
func (d *dummyHandler) RegisterByType(err error, fn responses.ErrorResponseFuncType)         {}
func (d *dummyHandler) ResolveError(ctx ports.Context, err error) responses.Response[any, any] {
	return responses.NewResponse[any, any](err.Error(), nil, 500)
}

func TestEchoAdapter(t *testing.T) {
	dh := &dummyHandler{}
	adapter := NewAdapter(dh).(*adapter)

	ctrl := &dummyController{
		fn: func(ctx context.Context, req requests.Request[any, any]) (any, error) {
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
	w := httptest.NewRecorder()

	adapter.echo.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Body       map[string]string `json:"Body"`
		StatusCode int               `json:"StatusCode"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "123", resp.Body["id"])
}
