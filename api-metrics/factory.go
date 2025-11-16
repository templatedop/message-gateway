package fxmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsRegistryFactory interface {
	Create() (*prometheus.Registry, error)
}

type DefaultMetricsRegistryFactory struct{}

func NewDefaultMetricsRegistryFactory() MetricsRegistryFactory {
	return &DefaultMetricsRegistryFactory{}
}

func (f *DefaultMetricsRegistryFactory) Create() (*prometheus.Registry, error) {
	return prometheus.NewRegistry(), nil
}
