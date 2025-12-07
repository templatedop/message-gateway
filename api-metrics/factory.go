package fxmetrics

import (
	"github.com/VictoriaMetrics/metrics"
)

type MetricsRegistryFactory interface {
	Create() (*metrics.Set, error)
}

type DefaultMetricsRegistryFactory struct{}

func NewDefaultMetricsRegistryFactory() MetricsRegistryFactory {
	return &DefaultMetricsRegistryFactory{}
}

func (f *DefaultMetricsRegistryFactory) Create() (*metrics.Set, error) {
	return metrics.NewSet(), nil
}
