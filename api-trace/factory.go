package trace

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	defaultPropagator propagation.TextMapPropagator
)

type TracerProviderFactory interface {
	Create(options ...TracerProviderOption) (*trace.TracerProvider, error)
}

type DefaultTracerProviderFactory struct{}

func NewDefaultTracerProviderFactory() TracerProviderFactory {

	return &DefaultTracerProviderFactory{}
}

func (f *DefaultTracerProviderFactory) Create(options ...TracerProviderOption) (*trace.TracerProvider, error) {
	appliedOptions := DefaultTracerProviderOptions()
	for _, opt := range options {
		opt(&appliedOptions)
	}

	//providers  here....
	tracerProvider := trace.NewTracerProvider(
		trace.WithResource(appliedOptions.Resource),
		trace.WithSampler(appliedOptions.Sampler),
	)

	if len(appliedOptions.SpanProcessors) == 0 {
		tracerProvider.RegisterSpanProcessor(NewNoopSpanProcessor())
	} else {
		for _, processor := range appliedOptions.SpanProcessors {
			tracerProvider.RegisterSpanProcessor(processor)
		}
	}

	if appliedOptions.Global {
		defaultPropagator = propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)

		otel.SetTracerProvider(tracerProvider)

		otel.SetTextMapPropagator(
			defaultPropagator,
		)
	}

	return tracerProvider, nil
}
