package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type TracerConfig struct {
	ServiceName string
	Endpoint    string
}

type Tracer interface {
	Start(ctx context.Context, spanName string) (context.Context, trace.Span)
	GetTraceID(ctx context.Context) string
}

type otelTracer struct {
	tracer trace.Tracer
}

func InitTracer(ctx context.Context, cfg TracerConfig) (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp.Shutdown, nil
}

func NewTracer(name string) Tracer {
	return &otelTracer{
		tracer: otel.Tracer(name),
	}
}

func (t *otelTracer) Start(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName)
}

func (t *otelTracer) GetTraceID(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}
	return ""
}
