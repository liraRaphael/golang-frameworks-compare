package responses

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/stretchr/testify/assert"
)

type testResponseParams struct {
	Trace string `header:"X-Trace"`
}

func TestResponse_ContractAndSerialization(t *testing.T) {
	t.Run("NewResponse initialization", func(t *testing.T) {
		body := map[string]string{"foo": "bar"}
		params := &testResponseParams{Trace: "abc"}
		resp := NewResponse(body, params, http.StatusCreated)

		assert.Equal(t, body, resp.Body())
		assert.Equal(t, params, resp.Params())
		assert.Equal(t, http.StatusCreated, resp.StatusCode())
	})

	t.Run("NewResponseFromParams with headers mapping", func(t *testing.T) {
		body := "simple body"
		headers := domain.HttpParamsType{
			"X-Trace": []string{"123"},
		}

		resp := NewResponseFromParams[string, *testResponseParams](body, http.StatusOK, headers, nil)

		assert.Equal(t, body, resp.Body())
		assert.Equal(t, "123", resp.Params().Trace)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, headers, resp.Headers())
	})

	t.Run("MarshalJSON serialization", func(t *testing.T) {
		body := "serialized"
		params := &testResponseParams{Trace: "xyz"}
		resp := NewResponse(body, params, http.StatusOK)

		b, err := json.Marshal(resp)
		assert.NoError(t, err)

		var result struct {
			Body       string              `json:"Body"`
			StatusCode int                 `json:"StatusCode"`
			Headers    domain.HttpParamsType `json:"Headers"`
		}
		err = json.Unmarshal(b, &result)
		assert.NoError(t, err)

		assert.Equal(t, "serialized", result.Body)
		assert.Equal(t, 200, result.StatusCode)
		assert.Equal(t, "xyz", result.Headers.GetValues("X-Trace")[0])
	})

	t.Run("NewResponseFromParams with non-pointer params type", func(t *testing.T) {
		body := "simple body"
		headers := domain.HttpParamsType{
			"X-Trace": []string{"123"},
		}

		resp := NewResponseFromParams[string, testResponseParams](body, http.StatusOK, headers, nil)

		assert.Equal(t, body, resp.Body())
		assert.Equal(t, "123", resp.Params().Trace)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, headers, resp.Headers())
	})
}
