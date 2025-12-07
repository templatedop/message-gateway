package fxmetrics

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	vmmetrics "github.com/VictoriaMetrics/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ============================================================================
// PROMETHEUS vs VICTORIAMETRICS COMPARISON BENCHMARKS
// ============================================================================

// Counter Operations Comparison
func BenchmarkPrometheus_CounterInc(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "Test counter",
	})
	registry.MustRegister(counter)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}

func BenchmarkVictoriaMetrics_CounterInc_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()
	counter := set.NewCounter("test_counter")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}

// Gauge Operations Comparison
func BenchmarkPrometheus_GaugeSet(b *testing.B) {
	registry := prometheus.NewRegistry()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_gauge",
		Help: "Test gauge",
	})
	registry.MustRegister(gauge)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gauge.Set(float64(i))
	}
}

func BenchmarkVictoriaMetrics_GaugeSet_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()
	counter := set.NewCounter("test_gauge")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.Set(uint64(i))
	}
}

// Gauge Add Operations Comparison
func BenchmarkPrometheus_GaugeAdd(b *testing.B) {
	registry := prometheus.NewRegistry()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_gauge",
		Help: "Test gauge",
	})
	registry.MustRegister(gauge)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gauge.Add(1)
	}
}

func BenchmarkVictoriaMetrics_GaugeAdd_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()
	gauge := set.NewGauge("test_gauge", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gauge.Add(1)
	}
}

// Metric Creation Comparison
func BenchmarkPrometheus_MetricCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		registry := prometheus.NewRegistry()
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "test_counter",
			Help: "Test counter",
		})
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "test_gauge",
			Help: "Test gauge",
		})
		histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "test_histogram",
			Help: "Test histogram",
		})
		registry.MustRegister(counter, gauge, histogram)
	}
}

func BenchmarkVictoriaMetrics_MetricCreation_Comparison(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		set := vmmetrics.NewSet()
		_ = set.NewCounter("test_counter")
		_ = set.NewGauge("test_gauge", nil)
		_ = set.NewSummary("test_summary")
	}
}

// Labeled Metrics Comparison
func BenchmarkPrometheus_WithLabels(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter",
			Help: "Test counter",
		},
		[]string{"method", "path", "status"},
	)
	registry.MustRegister(counter)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.WithLabelValues("GET", "/api/test", fmt.Sprintf("%d", i%5+200)).Inc()
	}
}

func BenchmarkVictoriaMetrics_WithLabels_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		metricName := fmt.Sprintf(`test_counter{method="GET",path="/api/test",status="%d"}`, i%5+200)
		set.GetOrCreateCounter(metricName).Inc()
	}
}

// Histogram vs Summary Comparison
func BenchmarkPrometheus_HistogramObserve(b *testing.B) {
	registry := prometheus.NewRegistry()
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "test_histogram",
		Help:    "Test histogram",
		Buckets: prometheus.DefBuckets,
	})
	registry.MustRegister(histogram)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		histogram.Observe(float64(i % 1000))
	}
}

func BenchmarkVictoriaMetrics_SummaryUpdate_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()
	summary := set.NewSummary("test_summary")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		summary.Update(float64(i % 1000))
	}
}

// Exposition Comparison - 100 Metrics
func BenchmarkPrometheus_Exposition_100Metrics(b *testing.B) {
	registry := prometheus.NewRegistry()

	for i := 0; i < 100; i++ {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("test_counter_%d", i),
			Help: "Test counter",
		})
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("test_gauge_%d", i),
			Help: "Test gauge",
		})
		registry.MustRegister(counter, gauge)
		counter.Inc()
		gauge.Set(float64(i))
	}

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/metrics", nil)
		handler.ServeHTTP(w, r)
	}
}

func BenchmarkVictoriaMetrics_Exposition_100Metrics_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()

	values := make([]uint64, 100)
	for i := 0; i < 100; i++ {
		counter := set.NewCounter(fmt.Sprintf("test_counter_%d", i))
		counter.Inc()
		idx := i
		values[i] = uint64(i)
		_ = set.NewGauge(fmt.Sprintf("test_gauge_%d", i), func() float64 { return float64(values[idx]) })
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		set.WritePrometheus(&buf)
	}
}

// Exposition Comparison - 1000 Metrics
func BenchmarkPrometheus_Exposition_1000Metrics(b *testing.B) {
	registry := prometheus.NewRegistry()

	for i := 0; i < 1000; i++ {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("test_counter_%d", i),
			Help: "Test counter",
		})
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("test_gauge_%d", i),
			Help: "Test gauge",
		})
		registry.MustRegister(counter, gauge)
		counter.Inc()
		gauge.Set(float64(i))
	}

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/metrics", nil)
		handler.ServeHTTP(w, r)
	}
}

func BenchmarkVictoriaMetrics_Exposition_1000Metrics_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()

	values := make([]uint64, 1000)
	for i := 0; i < 1000; i++ {
		counter := set.NewCounter(fmt.Sprintf("test_counter_%d", i))
		counter.Inc()
		idx := i
		values[i] = uint64(i)
		_ = set.NewGauge(fmt.Sprintf("test_gauge_%d", i), func() float64 { return float64(values[idx]) })
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		set.WritePrometheus(&buf)
	}
}

// Concurrent Operations Comparison
func BenchmarkPrometheus_ConcurrentCounterInc(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "Test counter",
	})
	registry.MustRegister(counter)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}

func BenchmarkVictoriaMetrics_ConcurrentCounterInc_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()
	counter := set.NewCounter("test_counter")

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}

// Concurrent Labeled Metrics Comparison
func BenchmarkPrometheus_ConcurrentWithLabels(b *testing.B) {
	registry := prometheus.NewRegistry()
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter",
			Help: "Test counter",
		},
		[]string{"method", "path", "status"},
	)
	registry.MustRegister(counter)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			counter.WithLabelValues("GET", "/api/test", fmt.Sprintf("%d", i%5+200)).Inc()
			i++
		}
	})
}

func BenchmarkVictoriaMetrics_ConcurrentWithLabels_Comparison(b *testing.B) {
	set := vmmetrics.NewSet()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			metricName := fmt.Sprintf(`test_counter{method="GET",path="/api/test",status="%d"}`, i%5+200)
			set.GetOrCreateCounter(metricName).Inc()
			i++
		}
	})
}

// Memory Usage Comparison - 1000 Metrics
func BenchmarkPrometheus_MemoryUsage_1000Metrics(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		registry := prometheus.NewRegistry()
		for j := 0; j < 1000; j++ {
			counter := prometheus.NewCounter(prometheus.CounterOpts{
				Name: fmt.Sprintf("counter_%d", j),
				Help: "Test counter",
			})
			gauge := prometheus.NewGauge(prometheus.GaugeOpts{
				Name: fmt.Sprintf("gauge_%d", j),
				Help: "Test gauge",
			})
			registry.MustRegister(counter, gauge)
			counter.Inc()
			gauge.Set(float64(j))
		}
	}
}

func BenchmarkVictoriaMetrics_MemoryUsage_1000Metrics_Comparison(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		set := vmmetrics.NewSet()
		values := make([]uint64, 1000)
		for j := 0; j < 1000; j++ {
			counter := set.NewCounter(fmt.Sprintf("counter_%d", j))
			counter.Inc()
			idx := j
			values[j] = uint64(j)
			_ = set.NewGauge(fmt.Sprintf("gauge_%d", j), func() float64 { return float64(values[idx]) })
		}
	}
}
