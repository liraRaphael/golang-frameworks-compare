package ports

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"go.opentelemetry.io/otel/trace"
)

type Context interface {
	context.Context
	Logger() logger.Logger
	WithLogger(logger.Logger) Context
	TraceID() string
	NewSpan(name, description string) (Context, trace.Span)
}
