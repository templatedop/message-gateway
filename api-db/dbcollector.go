package db

import (
	"sync"

	"github.com/VictoriaMetrics/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Stater is a provider of the Stat() function. Implemented by pgxpool.Pool.
type Stater interface {
	Stat() *pgxpool.Stat
}

// Collector is a VictoriaMetrics collector that collects statistics from pgxpool.
type Collector struct {
	stater Stater

	// Use callback-based gauges that fetch values dynamically
	acquireCount         *metrics.Counter
	acquireDuration      *metrics.Counter
	acquiredConns        *metrics.Gauge
	canceledAcquireCount *metrics.Counter
	constructingConns    *metrics.Gauge
	emptyAcquireCount    *metrics.Counter
	idleConns            *metrics.Gauge
	maxConns             *metrics.Gauge
	totalConns           *metrics.Gauge
	newConnsCount        *metrics.Counter
	maxLifetimeDestroy   *metrics.Counter
	maxIdleDestroy       *metrics.Counter

	mu sync.Mutex
}

// NewCollector creates a new Collector to collect stats from pgxpool.
func NewCollector(stater Stater, labels map[string]string, set *metrics.Set) *Collector {
	if set == nil {
		set = metrics.NewSet()
	}

	c := &Collector{
		stater: stater,
	}

	// Create metrics with callback-based gauges for real-time values
	c.acquireCount = set.NewCounter("pgxpool_acquire_count")
	c.acquireDuration = set.NewCounter("pgxpool_acquire_duration_ns")
	c.acquiredConns = set.NewGauge("pgxpool_acquired_conns", func() float64 {
		return float64(c.stater.Stat().AcquiredConns())
	})
	c.canceledAcquireCount = set.NewCounter("pgxpool_canceled_acquire_count")
	c.constructingConns = set.NewGauge("pgxpool_constructing_conns", func() float64 {
		return float64(c.stater.Stat().ConstructingConns())
	})
	c.emptyAcquireCount = set.NewCounter("pgxpool_empty_acquire")
	c.idleConns = set.NewGauge("pgxpool_idle_conns", func() float64 {
		return float64(c.stater.Stat().IdleConns())
	})
	c.maxConns = set.NewGauge("pgxpool_max_conns", func() float64 {
		return float64(c.stater.Stat().MaxConns())
	})
	c.totalConns = set.NewGauge("pgxpool_total_conns", func() float64 {
		return float64(c.stater.Stat().TotalConns())
	})
	c.newConnsCount = set.NewCounter("pgxpool_new_conns_count")
	c.maxLifetimeDestroy = set.NewCounter("pgxpool_max_lifetime_destroy_count")
	c.maxIdleDestroy = set.NewCounter("pgxpool_max_idle_destroy_count")

	// Initialize counters with current stat values
	c.updateCounters()

	return c
}

// updateCounters updates counter values from pool stats
func (c *Collector) updateCounters() {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := c.stater.Stat()

	// Set counter values (VictoriaMetrics counters are cumulative)
	c.acquireCount.Set(uint64(stats.AcquireCount()))
	c.acquireDuration.Set(uint64(stats.AcquireDuration()))
	c.canceledAcquireCount.Set(uint64(stats.CanceledAcquireCount()))
	c.emptyAcquireCount.Set(uint64(stats.EmptyAcquireCount()))
	c.newConnsCount.Set(uint64(stats.NewConnsCount()))
	c.maxLifetimeDestroy.Set(uint64(stats.MaxLifetimeDestroyCount()))
	c.maxIdleDestroy.Set(uint64(stats.MaxIdleDestroyCount()))
}

// Update should be called periodically to refresh counter values
func (c *Collector) Update() {
	c.updateCounters()
}