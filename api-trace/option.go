package trace

import (
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	maxQueueSize = 5000
)

type Options struct {
	Global         bool
	Resource       *resource.Resource
	Sampler        trace.Sampler
	SpanProcessors []trace.SpanProcessor
	QueueSize      int
}

func DefaultTracerProviderOptions() Options {
	return Options{
		Global:         true,
		Resource:       resource.Default(),
		Sampler:        NewParentBasedAlwaysOnSampler(),
		SpanProcessors: []trace.SpanProcessor{},
	}
}

type TracerProviderOption func(o *Options)

func WithQueueSize(size int) TracerProviderOption {
	return func(o *Options) {
		if size <= 0 || size > maxQueueSize {
			size = maxQueueSize
		}
		o.QueueSize = size

	}
}

func Global(b bool) TracerProviderOption {
	return func(o *Options) {
		o.Global = b
	}
}

func WithResource(r *resource.Resource) TracerProviderOption {
	return func(o *Options) {
		o.Resource = r
	}
}

func WithSampler(s trace.Sampler) TracerProviderOption {
	return func(o *Options) {
		o.Sampler = s
	}
}

func WithSpanProcessor(spanProcessor trace.SpanProcessor) TracerProviderOption {
	return func(o *Options) {
		o.SpanProcessors = append(o.SpanProcessors, spanProcessor)
	}
}
