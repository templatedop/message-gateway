package trace

import (
	"context"
	//"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
	otelsdktracetest "go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
)

const (
	Stdout   = "stdout"
	OtlpGrpc = "otlp-grpc"
	Noop     = "noop"
)

func NewNoopSpanProcessor() trace.SpanProcessor {
	return trace.NewBatchSpanProcessor(otelsdktracetest.NewNoopExporter())
}

func NewStdoutSpanProcessor(options ...stdouttrace.Option) trace.SpanProcessor {
	exporter, _ := stdouttrace.New(options...)

	return trace.NewBatchSpanProcessor(exporter)
}

// exporters here..
func NewOtlpGrpcSpanProcessor(ctx context.Context, conn *grpc.ClientConn) (trace.SpanProcessor, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	return trace.NewBatchSpanProcessor(exporter), nil
}
