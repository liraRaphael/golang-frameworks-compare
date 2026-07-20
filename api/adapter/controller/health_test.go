package controller

import (
	"context"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/stretchr/testify/assert"
)

func TestHealthController_WrapperExecute(t *testing.T) {
	ctrl := NewHealthController()
	req := requests.NewRequestFromParams[any, any](nil, nil, nil, nil, nil)

	res, err := ctrl.WrapperExecute(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, res)
}
