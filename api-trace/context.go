package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const TracerName = "DOP-IT2.0"

type CtxKey struct{}

func WithContext(ctx context.Context, tp trace.TracerProvider) context.Context {
	return context.WithValue(ctx, CtxKey{}, tp)
}

func CtxTracerProvider(ctx context.Context) trace.TracerProvider {
	if tp, ok := ctx.Value(CtxKey{}).(trace.TracerProvider); ok {
		return tp
	} else {
		return otel.GetTracerProvider()
	}
}

func CtxTracer(ctx context.Context) trace.Tracer {
	return CtxTracerProvider(ctx).Tracer(TracerName)
}
