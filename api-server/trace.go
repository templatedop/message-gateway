package router

import (
	"context"

	"MgApplication/api-server/middlewares/reqid"

	"go.opentelemetry.io/otel/attribute"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const TraceSpanAttributeHttpRequestId = "Traceparent"

func AnnotateTracerProvider(base oteltrace.TracerProvider) oteltrace.TracerProvider {
	if tp, ok := base.(*otelsdktrace.TracerProvider); ok {
		tp.RegisterSpanProcessor(NewTracerProviderRequestIdAnnotator())

		return tp
	}

	return base
}

type TracerProviderRequestIdAnnotator struct{}

func NewTracerProviderRequestIdAnnotator() *TracerProviderRequestIdAnnotator {
	return &TracerProviderRequestIdAnnotator{}
}

func (a *TracerProviderRequestIdAnnotator) OnStart(ctx context.Context, s otelsdktrace.ReadWriteSpan) {
	if rid, ok := ctx.Value(reqid.CtxRequestIdKey{}).(string); ok {
		s.SetAttributes(attribute.String(TraceSpanAttributeHttpRequestId, rid))
	}
}

func (a *TracerProviderRequestIdAnnotator) Shutdown(context.Context) error {
	return nil
}

func (a *TracerProviderRequestIdAnnotator) ForceFlush(context.Context) error {
	return nil
}

func (a *TracerProviderRequestIdAnnotator) OnEnd(otelsdktrace.ReadOnlySpan) {
}
