package fxmetrics

import (
	"go.uber.org/fx"
)

// Note: VictoriaMetrics doesn't use the collector pattern like Prometheus.
// Metrics are created directly in the Set and automatically registered.
// These functions are kept for compatibility but are no-ops.

func AsMetricsCollector(collector interface{}) fx.Option {
	// VictoriaMetrics doesn't use collectors - metrics are created directly
	return fx.Options()
}

func AsMetricsCollectors(collectors ...interface{}) fx.Option {
	// VictoriaMetrics doesn't use collectors - metrics are created directly
	return fx.Options()
}
