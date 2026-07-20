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
	headers := domain.HttpParamsType{
		"X-Trace": []string{"abc"},
	}

	query := domain.HttpParamsType{
		"name": []string{"Ada"},
	}

	pathParams := domain.HttpParamsType{
		"id": []string{"42"},
	}

	req := NewRequestFromParams[string, *testParams]("body", headers, query, pathParams, nil)
	paramValues := req.Params()

	require.Equal(t, "Ada", paramValues.Name)
	require.Equal(t, "abc", paramValues.Trace)
	require.Equal(t, "42", paramValues.ID)
}

func TestRequestParamsWithNonPointerParamType(t *testing.T) {
	headers := domain.HttpParamsType{
		"X-Trace": []string{"abc"},
	}

	query := domain.HttpParamsType{
		"name": []string{"Ada"},
	}

	pathParams := domain.HttpParamsType{
		"id": []string{"42"},
	}

	req := NewRequestFromParams[string, testParams]("body", headers, query, pathParams, nil)
	paramValues := req.Params()

	require.Equal(t, "Ada", paramValues.Name)
	require.Equal(t, "abc", paramValues.Trace)
	require.Equal(t, "42", paramValues.ID)
}
