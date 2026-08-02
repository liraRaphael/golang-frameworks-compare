package ports

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Context interface {
	context.Context
	Logger() logger.Logger
	NewSpan(name, description string) (Context, trace.Span)
}

type contextWrapper struct {
	context.Context
}

func NewContext(ctx context.Context) Context {
	return &contextWrapper{Context: ctx}
}

func (c *contextWrapper) Logger() logger.Logger {
	return logger.FromContext(c.Context)
}

func (c *contextWrapper) NewSpan(name, description string) (Context, trace.Span) {
	_ = description
	ctx, span := otel.Tracer("ports.context").Start(c.Context, name)
	return NewContext(ctx), span
}
