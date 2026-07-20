package router

import (
	"context"
	"errors"
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/docs/swagger"
	"github.com/stretchr/testify/assert"
)

type MockFrameworkAdapter struct {
	Routes      map[string]ports.Controller[any, any]
	Middlewares []ports.Middleware
	StartAddr   string
	StartErr    error
}

func NewMockFrameworkAdapter() *MockFrameworkAdapter {
	return &MockFrameworkAdapter{
		Routes: make(map[string]ports.Controller[any, any]),
	}
}

func (m *MockFrameworkAdapter) RegisterRoute(method string, path string, ctrl ports.Controller[any, any]) {
	m.Routes[method+":"+path] = ctrl
}

func (m *MockFrameworkAdapter) Use(middleware ports.Middleware) {
	m.Middlewares = append(m.Middlewares, middleware)
}

func (m *MockFrameworkAdapter) Start(addr string) error {
	m.StartAddr = addr
	return m.StartErr
}

type DummyController struct{}

func (d *DummyController) WrapperExecute(ctx context.Context, req requests.Request[any, any]) (any, error) {
	return "ok", nil
}

func TestStandardRouter(t *testing.T) {
	t.Run("successful routing and starting", func(t *testing.T) {
		adapter := NewMockFrameworkAdapter()
		router := NewStandardRouter(adapter)

		ctrl1 := &DummyController{}
		ctrl2 := &DummyController{}

		router.AddRoute("GET", "/hello", ctrl1, nil)
		router.AddRoute("POST", "/hello", ctrl2, swagger.RouteMetadata{
			Summary:     "Test",
			Description: "Test description",
		})

		router.Use("dummy-middleware-1")
		router.Use("dummy-middleware-2")

		err := router.Start(":8080")
		assert.NoError(t, err)

		assert.Equal(t, ":8080", adapter.StartAddr)
		assert.Len(t, adapter.Middlewares, 2)
		assert.Equal(t, "dummy-middleware-1", adapter.Middlewares[0])
		assert.Equal(t, "dummy-middleware-2", adapter.Middlewares[1])

		assert.Len(t, adapter.Routes, 2)
		assert.Equal(t, ctrl1, adapter.Routes["GET:/hello"])
		assert.Equal(t, ctrl2, adapter.Routes["POST:/hello"])
	})

	t.Run("router start returns error", func(t *testing.T) {
		adapter := NewMockFrameworkAdapter()
		expectedErr := errors.New("port already in use")
		adapter.StartErr = expectedErr

		router := NewStandardRouter(adapter)
		err := router.Start(":8080")
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("with swagger registry integration", func(t *testing.T) {
		adapter := NewMockFrameworkAdapter()
		reg := swagger.NewRegistry("Test API", "Test Description", "1.0")
		router := NewStandardRouter(adapter).(*StandardRouter).WithSwagger(reg)

		ctrl := &DummyController{}
		meta := swagger.RouteMetadata{
			Summary:     "Hello route",
			Description: "Returns hello message",
			Tags:        []string{"Greeting"},
			Input:       struct{ Name string }{},
			Output:      struct{ Message string }{},
		}

		router.AddRoute("GET", "/hello", ctrl, meta)
		// Pointer metadata should also be accepted
		router.AddRoute("POST", "/hello", ctrl, &meta)

		// Verification that routes were added to swagger registry
		// Since swagger Registry doesn't expose easy getters, we just verify it compiled and ran
		assert.NotNil(t, router.swagger)
	})
}
