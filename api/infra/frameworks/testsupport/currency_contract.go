package testsupport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/app/controller"
	"github.com/liraraphael/go-framework-bench/api/app/validation"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/core/usecases"
	"github.com/liraraphael/go-framework-bench/api/infra/docs/swagger"
	"github.com/liraraphael/go-framework-bench/api/infra/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type directRouter struct {
	adapter ports.FrameworkAdapter
}

func (r directRouter) WithSwagger(*swagger.Registry) {}

func (r directRouter) AddRoute(method, path string, ctrl ports.Controller[any, any, any, any], _ any) {
	r.adapter.RegisterRoute(method, path, ctrl)
}

func (r directRouter) Use(ports.Middleware) {}

func (r directRouter) Start(string) error { return nil }

func TestCurrencyContract(t *testing.T, adapter ports.FrameworkAdapter, serve func(*http.Request) (*http.Response, error)) {
	router.AddTypedRoute(
		directRouter{adapter: adapter},
		http.MethodPost,
		"/currency/convert",
		controller.NewCurrencyController(usecases.NewCurrencyUseCase()),
		validation.NewPlaygroundValidator(),
		nil,
	)
	tests := []struct {
		name        string
		body        string
		contentType string
		status      int
		errorCode   string
		value       float64
		from        string
		to          string
		quote       float64
		chunked     bool
	}{
		{name: "USD to BRL", body: `{"from":"USD","to":"BRL","value":10}`, contentType: "application/json", status: http.StatusOK, value: 63.1, from: "USD", to: "BRL", quote: 6.31},
		{name: "chunked request body", body: `{"from":"USD","to":"BRL","value":10}`, contentType: "application/json", status: http.StatusOK, value: 63.1, from: "USD", to: "BRL", quote: 6.31, chunked: true},
		{name: "zero is valid", body: `{"from":"BRL","to":"BRL","value":0}`, contentType: "application/json", status: http.StatusOK, from: "BRL", to: "BRL", quote: 1},
		{name: "BRL to USD", body: `{"from":"BRL","to":"USD","value":63.1}`, contentType: "application/json", status: http.StatusOK, value: 10, from: "BRL", to: "USD", quote: 1 / 6.31},
		{name: "missing body", contentType: "application/json", status: http.StatusBadRequest, errorCode: "validation_error"},
		{name: "invalid JSON", body: `{invalid`, contentType: "application/json", status: http.StatusBadRequest, errorCode: "invalid_request"},
		{name: "invalid value type", body: `{"from":"USD","to":"BRL","value":"10"}`, contentType: "application/json", status: http.StatusBadRequest, errorCode: "invalid_request"},
		{name: "unsupported currency", body: `{"from":"EUR","to":"BRL","value":10}`, contentType: "application/json", status: http.StatusBadRequest, errorCode: "validation_error"},
		{name: "negative value", body: `{"from":"USD","to":"BRL","value":-1}`, contentType: "application/json", status: http.StatusBadRequest, errorCode: "validation_error"},
		{name: "missing content type", body: `{"from":"USD","to":"BRL","value":10}`, status: http.StatusBadRequest, errorCode: "invalid_request"},
		{name: "unsupported content type", body: `{"from":"USD","to":"BRL","value":10}`, contentType: "text/plain", status: http.StatusBadRequest, errorCode: "invalid_request"},
		{name: "non-finite conversion", body: `{"from":"USD","to":"BRL","value":1.7976931348623157e308}`, contentType: "application/json", status: http.StatusBadRequest, errorCode: "validation_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/currency/convert", bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			if tt.chunked {
				req.ContentLength = -1
				req.TransferEncoding = []string{"chunked"}
			}
			resp, err := serve(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.status, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			payload, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			var envelope struct {
				Body       json.RawMessage `json:"Body"`
				StatusCode int             `json:"StatusCode"`
			}
			require.NoError(t, json.Unmarshal(payload, &envelope))
			assert.Equal(t, tt.status, envelope.StatusCode)
			if tt.status == http.StatusOK {
				var result responses.ConvertCurrencyResponse
				require.NoError(t, json.Unmarshal(envelope.Body, &result))
				assert.InDelta(t, tt.value, result.Value, 1e-9)
				assert.Equal(t, tt.from, result.From)
				assert.Equal(t, tt.to, result.To)
				assert.InDelta(t, tt.quote, result.Quote, 1e-9)
				return
			}

			var result responses.ErrorResponse
			require.NoError(t, json.Unmarshal(envelope.Body, &result))
			assert.Equal(t, tt.errorCode, result.Code)
			assert.NotEmpty(t, result.Message)
		})
	}
}
