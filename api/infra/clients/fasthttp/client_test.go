package fasthttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/stretchr/testify/assert"
)

func TestFastHttpClient_Do(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		expectedResponse := map[string]any{"message": "success"}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewFastHttpClient(5 * time.Second)
		req := requests.NewRequestFromParams[any](nil, nil, nil, nil, nil)
		resp, err := client.Do(context.Background(), enums.MethodGet, server.URL, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, expectedResponse, resp.Body)
	})

	t.Run("successful POST request with body and headers", func(t *testing.T) {
		requestBody := map[string]any{"name": "test"}
		expectedResponse := map[string]any{"status": "created"}
		headers := domain.NewHttpParam()
		headers.Set("X-Custom-Header", "value")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "value", r.Header.Get("X-Custom-Header"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body map[string]any
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, requestBody, body)

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		client := NewFastHttpClient(5 * time.Second)
		req := requests.NewRequestFromParams[any](requestBody, headers, nil, nil, nil)
		resp, err := client.Do(context.Background(), enums.MethodPost, server.URL, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, expectedResponse, resp.Body)
	})
}
