package fxmetrics

import (
	config "MgApplication/api-config"
	log "MgApplication/api-log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/fx"
)

type FxMetricsRegistryParam struct {
	fx.In
	Factory    MetricsRegistryFactory
	Config     *config.Config
	Collectors []prometheus.Collector `group:"metrics-collectors"`
}

func NewFxMetricsRegistry(p FxMetricsRegistryParam) (*prometheus.Registry, error) {
	registry, err := p.Factory.Create()
	if err != nil {
		log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Msg("failed to create metrics registry")

		return nil, err
	}

	var registrableCollectors []prometheus.Collector

	if p.Config.GetBool("metrics.collect.build") {
		registrableCollectors = append(registrableCollectors, collectors.NewBuildInfoCollector())

	}

	if p.Config.GetBool("metrics.collect.process") {

		registrableCollectors = append(registrableCollectors, collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	if p.Config.GetBool("metrics.collect.go") {

		registrableCollectors = append(registrableCollectors, collectors.NewGoCollector())
	}

	registrableCollectors = append(registrableCollectors, p.Collectors...)

	for _, collector := range registrableCollectors {
		err = registry.Register(collector)
		if err != nil {
			log.GetBaseLoggerInstance().ToZerolog().Error().Err(err).Msgf("failed to register metrics collector %+T", collector)

			return nil, err
		} else {
			log.GetBaseLoggerInstance().ToZerolog().Debug().Msgf("registered metrics collector %+T", collector)
		}
	}

	return registry, err
}
