package trace

import (
	"strings"
)

type SpanProcessor int

const (
	NoopSpanProcessor SpanProcessor = iota
	StdoutSpanProcessor
	TestSpanProcessor
	OtlpGrpcSpanProcessor
)

func (p SpanProcessor) String() string {
	switch p {
	case StdoutSpanProcessor:
		return Stdout

	case OtlpGrpcSpanProcessor:
		return OtlpGrpc
	default:
		return Noop
	}
}

func FetchSpanProcessor(p string) SpanProcessor {

	switch strings.ToLower(p) {
	case Stdout:
		return StdoutSpanProcessor
	case OtlpGrpc:
		return OtlpGrpcSpanProcessor
	default:
		return NoopSpanProcessor
	}
}

type Sampler int

const (
	ParentBasedAlwaysOnSampler Sampler = iota
	ParentBasedAlwaysOffSampler
	ParentBasedTraceIdRatioSampler
	AlwaysOnSampler
	AlwaysOffSampler
	TraceIdRatioSampler
)

func (s Sampler) String() string {
	switch s {
	case ParentBasedAlwaysOffSampler:
		return ParentBasedAlwaysOff
	case ParentBasedTraceIdRatioSampler:
		return ParentBasedTraceIdRatio
	case AlwaysOnSampler:
		return AlwaysOn
	case AlwaysOffSampler:
		return AlwaysOff
	case TraceIdRatioSampler:
		return TraceIdRatio
	default:
		return ParentBasedAlwaysOn
	}
}

func FetchSampler(s string) Sampler {
	switch strings.ToLower(s) {
	case ParentBasedAlwaysOff:
		return ParentBasedAlwaysOffSampler
	case ParentBasedTraceIdRatio:
		return ParentBasedTraceIdRatioSampler
	case AlwaysOn:
		return AlwaysOnSampler
	case AlwaysOff:
		return AlwaysOffSampler
	case TraceIdRatio:
		return TraceIdRatioSampler
	default:
		return ParentBasedAlwaysOnSampler
	}
}
