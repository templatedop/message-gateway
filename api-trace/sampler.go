package trace

import (
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	ParentBasedAlwaysOn     = "parent-based-always-on"      // parent based always on sampling
	ParentBasedAlwaysOff    = "parent-based-always-off"     // parent based always off sampling
	ParentBasedTraceIdRatio = "parent-based-trace-id-ratio" // parent based trace id ratio sampling
	AlwaysOn                = "always-on"                   // always on sampling
	AlwaysOff               = "always-off"                  // always off sampling
	TraceIdRatio            = "trace-id-ratio"              // trace id ratio sampling
)

func NewParentBasedAlwaysOnSampler() otelsdktrace.Sampler {
	return otelsdktrace.ParentBased(otelsdktrace.AlwaysSample())
}

func NewParentBasedAlwaysOffSampler() otelsdktrace.Sampler {
	return otelsdktrace.ParentBased(otelsdktrace.NeverSample())
}

func NewParentBasedTraceIdRatioSampler(ratio float64) otelsdktrace.Sampler {
	return otelsdktrace.ParentBased(otelsdktrace.TraceIDRatioBased(ratio))
}

func NewAlwaysOnSampler() otelsdktrace.Sampler {
	return otelsdktrace.AlwaysSample()
}

func NewAlwaysOffSampler() otelsdktrace.Sampler {
	return otelsdktrace.NeverSample()
}

func NewTraceIdRatioSampler(ratio float64) otelsdktrace.Sampler {
	return otelsdktrace.TraceIDRatioBased(ratio)
}
