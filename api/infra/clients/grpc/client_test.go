package grpc

import (
	"context"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/errors"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/stretchr/testify/assert"
)

func TestGrpcClient(t *testing.T) {
	client, err := NewGrpcClient("localhost:9999")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	if closer, ok := client.(interface{ Close() error }); ok {
		defer closer.Close()
	}

	t.Run("Do returns unimplemeted ClientError", func(t *testing.T) {
		headers := domain.NewHttpParam()
		headers.Set("X-Test-Header", "Value")

		req := requests.NewRequestFromParams[any](nil, headers, nil, nil, nil)
		resp, err := client.Do(context.Background(), enums.MethodGet, "some-url", req)

		assert.Nil(t, resp)
		assert.Error(t, err)

		clientErr, ok := err.(*errors.ClientError)
		assert.True(t, ok)
		assert.Equal(t, 501, clientErr.StatusCode)
		assert.Contains(t, clientErr.Message, "generic gRPC Do not fully implemented")
	})
}
