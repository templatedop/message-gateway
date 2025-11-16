package fxmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

func AsMetricsCollector(collector prometheus.Collector) fx.Option {
	return fx.Supply(
		fx.Annotate(
			collector,
			fx.As(new(prometheus.Collector)),
			fx.ResultTags(`group:"metrics-collectors"`),
		),
	)
}

func AsMetricsCollectors(collectors ...prometheus.Collector) fx.Option {
	registrations := []fx.Option{}

	for _, collector := range collectors {
		registrations = append(
			registrations,
			fx.Supply(
				fx.Annotate(
					collector,
					fx.As(new(prometheus.Collector)),
					fx.ResultTags(`group:"metrics-collectors"`),
				),
			),
		)
	}

	return fx.Options(registrations...)
}
