package beego

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/stretchr/testify/assert"
)

type dummyController struct {
	fn func(ctx context.Context, req requests.Request[any, any]) (any, error)
}

func (d *dummyController) WrapperExecute(ctx context.Context, req requests.Request[any, any]) (any, error) {
	return d.fn(ctx, req)
}

type dummyHandler struct{}

func (d *dummyHandler) Handle(ctx context.Context, successStatusCode int, output any, headers domain.HttpParam) *responses.Response[any] {
	return &responses.Response[any]{StatusCode: successStatusCode, Body: output}
}
func (d *dummyHandler) RegisterByMessage(message string, fn enums.ErrorResponseFuncType) {}
func (d *dummyHandler) RegisterByType(err error, fn enums.ErrorResponseFuncType)         {}
func (d *dummyHandler) ResolveError(ctx context.Context, err error) *responses.Response[any] {
	return &responses.Response[any]{StatusCode: 500, Body: err.Error()}
}

func TestBeegoAdapter(t *testing.T) {
	dh := &dummyHandler{}
	adapter := NewAdapter(dh).(*adapter)

	ctrl := &dummyController{
		fn: func(ctx context.Context, req requests.Request[any]) (any, error) {
			param, _ := req.PathParams().GetFirst("id")
			return map[string]string{"id": param}, nil
		},
	}

	adapter.RegisterRoute("GET", "/test/:id", ctrl)

	req := httptest.NewRequest("GET", "/test/123", nil)
	w := httptest.NewRecorder()

	adapter.server.Handlers.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp responses.Response[map[string]string]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "123", resp.Body["id"])
}
