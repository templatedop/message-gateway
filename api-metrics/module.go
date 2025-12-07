package fxmetrics

import (
	config "MgApplication/api-config"
	log "MgApplication/api-log"

	"github.com/VictoriaMetrics/metrics"
	"go.uber.org/fx"
)

type FxMetricsRegistryParam struct {
	fx.In
	Factory MetricsRegistryFactory
	Config  *config.Config
}

func NewFxMetricsRegistry(p FxMetricsRegistryParam) (*metrics.Set, error) {
	set, err := p.Factory.Create()
	if err != nil {
		log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Msg("failed to create metrics set")
		return nil, err
	}

	// VictoriaMetrics automatically includes process metrics (memory, CPU, goroutines, etc.)
	// No need to manually register collectors like Prometheus
	// Build info, process, and Go metrics are included by default

	log.GetBaseLoggerInstance().ToZerolog().Debug().Msg("VictoriaMetrics set created successfully")

	return set, nil
}
