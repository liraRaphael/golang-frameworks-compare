package enums

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type (
	ErrorResponseFuncType func(ctx context.Context, err error) *responses.Response
)
