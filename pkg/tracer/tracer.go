package tracer

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/trace"
)

// TODO (KB)Connect tracing
func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	span := trace.SpanFromContext(ctx)
	span.SetName(name)
	log.Println("tracing: ", name)

	return trace.ContextWithSpan(ctx, span), span
}
