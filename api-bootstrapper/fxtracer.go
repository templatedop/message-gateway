package bootstrapper

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	log "MgApplication/api-log"
	trace "MgApplication/api-trace"

	config "MgApplication/api-config"
)

const (
	ModuleName = "trace"
)

var (
	hostname, _ = os.Hostname()
	version     = "0.0.1"
)

var fxTrace = fx.Module(
	ModuleName,
	fx.Provide(
		trace.NewDefaultTracerProviderFactory,
		NewFxTracerProvider,
		fx.Annotate(
			NewFxTracerProvider,
			fx.As(new(oteltrace.TracerProvider)),
		),
	),
)

type FxTraceParam struct {
	fx.In
	LifeCycle fx.Lifecycle
	Factory   trace.TracerProviderFactory
	Config    *config.Config
}

func NewFxTracerProvider(param FxTraceParam) (*otelsdktrace.TracerProvider, error) {

	ctx := context.Background()

	//resources here...
	resource, err := createResource(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("cannot create tracer provider resource: %w", err)
	}
	var processer otelsdktrace.SpanProcessor
	processer = trace.NewNoopSpanProcessor()
	if !param.Config.GetBool("trace.enabled") {
		processer = trace.NewNoopSpanProcessor()
	} else {
		//exporters here ...
		processer, err = createSpanProcessor(ctx, param)
		if err != nil {
			processer = trace.NewNoopSpanProcessor()
		}
	}

	//samplers here..
	sampler := createSampler(param)

	//providers here...
	tracerProvider, err := param.Factory.Create(
		trace.WithResource(resource),
		trace.WithSpanProcessor(processer),
		trace.WithSampler(sampler),
	)
	if err != nil {
		return nil, err
	}

	// Set as global trace provider to ensure context propagation works
	// This is CRITICAL for distributed tracing:
	// - All libraries using otel.GetTracerProvider() will get this provider
	// - Ensures parent-child span relationships across boundaries (HTTP -> DB)
	// - Enables trace context propagation through pgx database operations
	otel.SetTracerProvider(tracerProvider)

	param.LifeCycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {

			cctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err = tracerProvider.ForceFlush(cctx); err != nil {
				return err
			}

			if err = tracerProvider.Shutdown(cctx); err != nil {
				return err
			}
			return nil
		},
	})

	return tracerProvider, nil
}

// exporters here...
func fetchSpanProcessorType(p FxTraceParam) trace.SpanProcessor {
	processtype := p.Config.GetString("trace.processor.type")

	return trace.FetchSpanProcessor(processtype)

}

// Resources here.....
func createResource(ctx context.Context, p FxTraceParam) (*resource.Resource, error) {
	if p.Config.AppName() == "" {
		log.GetBaseLoggerInstance().ToZerolog().Info().Msg("App Name doesn't exist. Trace cannot be enabled")
	}
	var servicename string
	if p.Config.AppName() == "" {
		servicename = "Default"

	} else {
		servicename = p.Config.GetString("appname")
	}

	// if p.Config.Exists("appname") {

	// } else {
	// 	log.Info(nil, "AppName is nil . Trace cannot be enabled")
	// }

	if p.Config.Exists("info.version") {
		version = p.Config.GetString("info.version")
	}

	res, err := resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			attribute.String("service.name", servicename),
			//attribute.String("telemetry.sdk.language", "go"),

			semconv.TelemetrySDKLanguageGo.Key.String("go"),
			//semconv.ServiceNameKey.String(servicename),
			semconv.HostName(hostname),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// exporters here ...
func createSpanProcessor(ctx context.Context, p FxTraceParam) (otelsdktrace.SpanProcessor, error) {

	pp := fetchSpanProcessorType(p)
	switch pp {

	case trace.StdoutSpanProcessor:
		var opts []stdouttrace.Option
		if p.Config.GetBool("trace.processor.options.pretty") {
			opts = append(opts, stdouttrace.WithPrettyPrint())
		}
		log.Info(nil, "stdout trace enabled")
		return trace.NewStdoutSpanProcessor(opts...), nil

	case trace.OtlpGrpcSpanProcessor:
		conn, err := trace.NewOtlpGrpcClientConnection(ctx, p.Config.GetString("trace.processor.options.host"))
		if err != nil {
			log.Error(ctx, "cannot create otlp grpc client connection", err)
			return trace.NewNoopSpanProcessor(), err
		}
		log.Info(nil, "otlpgrc trace enabled")
		//batch exporter here..
		return trace.NewOtlpGrpcSpanProcessor(ctx, conn)
	default:
		log.Info(nil, "default noop trace enabled")
		return trace.NewNoopSpanProcessor(), nil
	}
}

func createSampler(p FxTraceParam) otelsdktrace.Sampler {
	var ratio float64
	ratio = 0.1
	samplertype := p.Config.GetString("trace.sampler.type")
	sampler := trace.FetchSampler(samplertype)
	if p.Config.Exists("trace.sampler.options.ratio") {
		ratio = p.Config.GetFloat64("trace.sampler.options.ratio")
	} else {
		ratio = 0.1
	}

	if os.Getenv("APP_ENV") == "prod" {
		if samplertype == "always-on" {
			samplertype = "always-off"
			sampler = trace.FetchSampler(samplertype)
		}
		ratio = 0.1

		// if p.Config.Exists("trace.sampler.options.ratio") {
		// 	ratio = 0.1
		// }

	}

	switch sampler {
	case trace.ParentBasedAlwaysOffSampler:
		return trace.NewParentBasedAlwaysOffSampler()
	case trace.ParentBasedTraceIdRatioSampler:
		return trace.NewParentBasedTraceIdRatioSampler(ratio)
	case trace.AlwaysOnSampler:
		return trace.NewAlwaysOnSampler()
	case trace.AlwaysOffSampler:
		return trace.NewAlwaysOffSampler()
	case trace.TraceIdRatioSampler:
		return trace.NewTraceIdRatioSampler(ratio)
	default:
		return trace.NewParentBasedAlwaysOnSampler()
	}
}
