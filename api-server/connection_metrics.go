package router

import (
	"github.com/VictoriaMetrics/metrics"
)

var (
	// activeConnectionsGauge tracks the current number of active connections
	activeConnectionsGauge *metrics.Gauge

	// rejectedConnectionsCounter tracks total number of rejected connections
	rejectedConnectionsCounter *metrics.Counter

	// maxConnectionsGauge tracks the configured max connections limit
	maxConnectionsGauge *metrics.Gauge
)

// InitConnectionMetrics initializes connection tracking metrics
func InitConnectionMetrics(set *metrics.Set) {
	if set == nil {
		set = metrics.NewSet()
	}

	activeConnectionsGauge = set.NewGauge("http_server_active_connections", func() float64 { return 0 })
	rejectedConnectionsCounter = set.NewCounter("http_server_rejected_connections_total")
	maxConnectionsGauge = set.NewGauge("http_server_max_connections", func() float64 { return 0 })
}

// SetMaxConnections sets the max connections gauge value
func SetMaxConnections(max int64) {
	if maxConnectionsGauge != nil {
		maxConnectionsGauge.Set(uint64(max))
	}
}

// IncActiveConnections increments the active connections gauge
func IncActiveConnections() {
	if activeConnectionsGauge != nil {
		activeConnectionsGauge.Add(1)
	}
}

// DecActiveConnections decrements the active connections gauge
func DecActiveConnections() {
	if activeConnectionsGauge != nil {
		activeConnectionsGauge.Add(-1)
	}
}

// IncRejectedConnections increments the rejected connections counter
func IncRejectedConnections() {
	if rejectedConnectionsCounter != nil {
		rejectedConnectionsCounter.Inc()
	}
}
