package controller

import (
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/stretchr/testify/assert"
)

func TestHealthController_Execute(t *testing.T) {
	ctrl := NewHealthController()
	req := requests.NewRequest[any, any](nil, nil)

	res, err := ctrl.Execute(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"status": "ok"}, res.Body())
	assert.Equal(t, 200, res.StatusCode())
	assert.Nil(t, res.Params())
}
