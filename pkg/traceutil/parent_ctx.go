package traceutil

import (
	"context"
	"github.com/opentracing/opentracing-go"
)

// ProvideParentSpan returns parent span context option if it exists
func ProvideParentSpan(ctx context.Context) opentracing.StartSpanOption {
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		return opentracing.ChildOf(parentSpan.Context())
	}

	return nil
}
