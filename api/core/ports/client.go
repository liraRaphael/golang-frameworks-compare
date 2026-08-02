package ports

import (
	"github.com/liraraphael/go-framework-bench/api/core/domain/enums"
	"github.com/liraraphael/go-framework-bench/api/core/domain/requests"
	"github.com/liraraphael/go-framework-bench/api/core/domain/responses"
)

type Client interface {
	Do(ctx Context, method enums.HttpMethod, url string, req requests.Request[any, any]) (responses.Response[any, any], error)
}
