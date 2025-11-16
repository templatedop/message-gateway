package router

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// activeConnectionsGauge tracks the current number of active connections
	activeConnectionsGauge prometheus.Gauge

	// rejectedConnectionsCounter tracks total number of rejected connections
	rejectedConnectionsCounter prometheus.Counter

	// maxConnectionsGauge tracks the configured max connections limit
	maxConnectionsGauge prometheus.Gauge
)

// InitConnectionMetrics initializes connection tracking metrics
func InitConnectionMetrics(registry prometheus.Registerer) {
	factory := promauto.With(registry)

	activeConnectionsGauge = factory.NewGauge(prometheus.GaugeOpts{
		Namespace: "http",
		Subsystem: "server",
		Name:      "active_connections",
		Help:      "Current number of active HTTP connections",
	})

	rejectedConnectionsCounter = factory.NewCounter(prometheus.CounterOpts{
		Namespace: "http",
		Subsystem: "server",
		Name:      "rejected_connections_total",
		Help:      "Total number of connections rejected due to max connections limit",
	})

	maxConnectionsGauge = factory.NewGauge(prometheus.GaugeOpts{
		Namespace: "http",
		Subsystem: "server",
		Name:      "max_connections",
		Help:      "Configured maximum number of concurrent connections",
	})
}

// SetMaxConnections sets the max connections gauge value
func SetMaxConnections(max int64) {
	if maxConnectionsGauge != nil {
		maxConnectionsGauge.Set(float64(max))
	}
}

// IncActiveConnections increments the active connections gauge
func IncActiveConnections() {
	if activeConnectionsGauge != nil {
		activeConnectionsGauge.Inc()
	}
}

// DecActiveConnections decrements the active connections gauge
func DecActiveConnections() {
	if activeConnectionsGauge != nil {
		activeConnectionsGauge.Dec()
	}
}

// IncRejectedConnections increments the rejected connections counter
func IncRejectedConnections() {
	if rejectedConnectionsCounter != nil {
		rejectedConnectionsCounter.Inc()
	}
}
