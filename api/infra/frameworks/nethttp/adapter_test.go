package nethttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func (d *dummyHandler) RegisterByMessage(message string, fn ports.ErrorResponseFuncType) {}
func (d *dummyHandler) RegisterByType(err error, fn ports.ErrorResponseFuncType)         {}

func (d *dummyHandler) ResolveError(ctx ports.Context, err error) responses.Response[any, any] {
	return responses.NewResponse[any, any](err.Error(), nil, 400)
}

func TestNetHttpAdapter(t *testing.T) {
	dh := &dummyHandler{}
	adapter := NewAdapter(dh).(*adapter)

	t.Run("successful GET request", func(t *testing.T) {
		ctrl := &dummyController{
			fn: func(ctx context.Context, req requests.Request[any, any]) (any, error) {
				return map[string]string{"message": "success"}, nil
			},
		}

		adapter.RegisterRoute("GET", "/test-get", ctrl)

		req := httptest.NewRequest("GET", "/test-get", nil)
		w := httptest.NewRecorder()

		adapter.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Body       map[string]string `json:"Body"`
			StatusCode int               `json:"StatusCode"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "success", resp.Body["message"])
	})

	t.Run("method not allowed", func(t *testing.T) {
		ctrl := &dummyController{
			fn: func(ctx context.Context, req requests.Request[any, any]) (any, error) {
				return "ok", nil
			},
		}

		adapter.RegisterRoute("POST", "/test-post", ctrl)

		req := httptest.NewRequest("GET", "/test-post", nil)
		w := httptest.NewRecorder()

		adapter.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("controller error handling", func(t *testing.T) {
		ctrl := &dummyController{
			fn: func(ctx context.Context, req requests.Request[any, any]) (any, error) {
				return nil, errors.New("something went wrong")
			},
		}

		adapter.RegisterRoute("GET", "/test-error", ctrl)

		req := httptest.NewRequest("GET", "/test-error", nil)
		w := httptest.NewRecorder()

		adapter.mux.ServeHTTP(w, req)

		var resp struct {
			Body       string `json:"Body"`
			StatusCode int    `json:"StatusCode"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
		assert.Equal(t, "something went wrong", resp.Body)
	})

	t.Run("with JSON body parsing", func(t *testing.T) {
		ctrl := &dummyController{
			fn: func(ctx context.Context, req requests.Request[any, any]) (any, error) {
				bodyMap, ok := req.Body().(map[string]any)
				if !ok {
					return nil, errors.New("invalid body format")
				}
				name := bodyMap["name"].(string)
				return map[string]string{"name": name}, nil
			},
		}

		adapter.RegisterRoute("POST", "/test-body", ctrl)

		bodyBytes, _ := json.Marshal(map[string]string{"name": "test-user"})
		req := httptest.NewRequest("POST", "/test-body", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		adapter.mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Body       map[string]string `json:"Body"`
			StatusCode int               `json:"StatusCode"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "test-user", resp.Body["name"])
	})

	t.Run("invalid JSON body parsing", func(t *testing.T) {
		ctrl := &dummyController{
			fn: func(ctx context.Context, req requests.Request[any, any]) (any, error) {
				return "ok", nil
			},
		}

		adapter.RegisterRoute("POST", "/test-invalid-body", ctrl)

		req := httptest.NewRequest("POST", "/test-invalid-body", bytes.NewReader([]byte("{invalid-json")))
		w := httptest.NewRecorder()

		adapter.mux.ServeHTTP(w, req)

		var resp struct {
			Body       string `json:"Body"`
			StatusCode int    `json:"StatusCode"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, 400, resp.StatusCode)
		assert.Contains(t, resp.Body, "invalid character")
	})
}
