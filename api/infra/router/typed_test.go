package router

import (
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/context"
	"github.com/stretchr/testify/require"
)

type allowValidator struct{}

func (allowValidator) Validate(ports.Context, any) error { return nil }

type typedResponseController struct{}

type typedResponseParams struct {
	Location string `header:"Location"`
}

func (typedResponseController) Execute(ports.Context, requests.Request[struct{}, any]) (responses.Response[string, typedResponseParams], error) {
	return responses.NewResponse[string, typedResponseParams]("created", typedResponseParams{Location: "/created"}, 201), nil
}

func TestTypedControllerAdapterPreservesControllerResponse(t *testing.T) {
	adapter := typedControllerAdapter[struct{}, any, string, typedResponseParams]{
		controller: typedResponseController{},
		validator:  allowValidator{},
	}
	req := requests.NewRequest[any, any](struct{}{}, nil)

	result, err := adapter.Execute(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, 201, result.StatusCode())
	require.Equal(t, "created", result.Body())
	require.Equal(t, typedResponseParams{Location: "/created"}, result.Params())
	require.Equal(t, "/created", result.Headers()["Location"][0])
}
