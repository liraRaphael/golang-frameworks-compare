package requests

import (
	"testing"

	"github.com/liraraphael/go-framework-bench/api/core/domain"
	"github.com/stretchr/testify/require"
)

type testParams struct {
	Name  string `query:"name"`
	Trace string `header:"X-Trace"`
	ID    string `path:"id"`
}

func TestRequestParams(t *testing.T) {
	headers := domain.NewHttpParam()
	headers.Set("X-Trace", "abc")

	query := domain.NewHttpParam()
	query.Set("name", "Ada")

	pathParams := domain.NewHttpParam()
	pathParams.Set("id", "42")

	params := testParams{}
	req := NewRequestFromParams("body", headers, query, pathParams, &params)
	paramValues := req.Params()

	require.Equal(t, "Ada", paramValues.Name)
	require.Equal(t, "abc", paramValues.Trace)
	require.Equal(t, "42", paramValues.ID)
}
