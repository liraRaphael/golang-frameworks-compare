package responses

import "github.com/liraraphael/go-framework-bench/api/core/domain"

type Response struct {
	Body       any              `json:"body,omitempty"`
	Headers    domain.HttpParam `json:"headers,omitempty"`
	StatusCode int              `json:"status_code,omitempty"`
}
