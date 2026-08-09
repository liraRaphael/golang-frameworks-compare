package context

import (
	"context"

	"github.com/liraraphael/go-framework-bench/api/core/ports"
	"github.com/liraraphael/go-framework-bench/api/infra/observability/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type contextWrapper struct {
	context.Context
}

func Background() ports.Context {
	return NewContext(context.Background())
}

func TODO() ports.Context {
	return NewContext(context.TODO())
}

func NewContext(ctx context.Context) ports.Context {
	return &contextWrapper{Context: ctx}
}

func (c *contextWrapper) Logger() logger.Logger {
	return logger.FromContext(c)
}

func (c *contextWrapper) WithLogger(logger.Logger) ports.Context {
	return c
}

func (c *contextWrapper) TraceID() string {
	if spanCtx := trace.SpanContextFromContext(c.Context); spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func (c *contextWrapper) NewSpan(name, description string) (ports.Context, trace.Span) {
	_ = description
	ctx, span := otel.Tracer("ports.context").Start(c.Context, name)
	return NewContext(ctx), span
}
