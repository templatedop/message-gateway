package fxmetrics

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/VictoriaMetrics/metrics"
)

// Benchmark VictoriaMetrics counter operations
func BenchmarkVictoriaMetrics_CounterInc(b *testing.B) {
	set := metrics.NewSet()
	counter := set.NewCounter("test_counter")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.Inc()
	}
}

// Benchmark VictoriaMetrics gauge operations
func BenchmarkVictoriaMetrics_GaugeSet(b *testing.B) {
	set := metrics.NewSet()
	counter := set.NewCounter("test_gauge")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.Set(uint64(i))
	}
}

// Benchmark VictoriaMetrics gauge Add operations
func BenchmarkVictoriaMetrics_GaugeAdd(b *testing.B) {
	set := metrics.NewSet()
	gauge := set.NewGauge("test_gauge", nil) // nil callback allows Set/Add

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gauge.Add(1)
	}
}

// Benchmark VictoriaMetrics metric creation
func BenchmarkVictoriaMetrics_MetricCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		set := metrics.NewSet()
		_ = set.NewCounter("test_counter")
		_ = set.NewGauge("test_gauge", func() float64 { return 0 })
		_ = set.NewSummary("test_summary")
	}
}

// Benchmark VictoriaMetrics with labels (dynamic metric creation)
func BenchmarkVictoriaMetrics_WithLabels_5Different(b *testing.B) {
	set := metrics.NewSet()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		metricName := fmt.Sprintf(`http_requests_total{method="GET",path="/api/test",status="%d"}`, i%5+200)
		set.GetOrCreateCounter(metricName).Inc()
	}
}

// Benchmark VictoriaMetrics with many unique labels
func BenchmarkVictoriaMetrics_WithLabels_ManyUnique(b *testing.B) {
	set := metrics.NewSet()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		metricName := fmt.Sprintf(`http_requests_total{method="GET",path="/api/test/%d",status="200"}`, i%100)
		set.GetOrCreateCounter(metricName).Inc()
	}
}

// Benchmark VictoriaMetrics summary operations
func BenchmarkVictoriaMetrics_SummaryUpdate(b *testing.B) {
	set := metrics.NewSet()
	summary := set.NewSummary("test_summary")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		summary.Update(float64(i % 1000))
	}
}

// Benchmark VictoriaMetrics exposition (writing metrics)
func BenchmarkVictoriaMetrics_Exposition_10Metrics(b *testing.B) {
	set := metrics.NewSet()

	// Create 10 metrics
	values := make([]uint64, 10)
	for i := 0; i < 10; i++ {
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

// Benchmark VictoriaMetrics exposition with 100 metrics
func BenchmarkVictoriaMetrics_Exposition_100Metrics(b *testing.B) {
	set := metrics.NewSet()

	// Create 100 metrics
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

// Benchmark VictoriaMetrics exposition with 1000 metrics
func BenchmarkVictoriaMetrics_Exposition_1000Metrics(b *testing.B) {
	set := metrics.NewSet()

	// Create 1000 metrics
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

// Benchmark concurrent counter updates
func BenchmarkVictoriaMetrics_ConcurrentCounterInc(b *testing.B) {
	set := metrics.NewSet()
	counter := set.NewCounter("test_counter")

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}

// Benchmark concurrent gauge updates
func BenchmarkVictoriaMetrics_ConcurrentGaugeAdd(b *testing.B) {
	set := metrics.NewSet()
	gauge := set.NewGauge("test_gauge", nil) // nil callback allows Set/Add

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gauge.Add(1)
		}
	})
}

// Benchmark concurrent labeled metric creation and increment
func BenchmarkVictoriaMetrics_ConcurrentWithLabels(b *testing.B) {
	set := metrics.NewSet()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			metricName := fmt.Sprintf(`http_requests_total{method="GET",path="/api/test",status="%d"}`, i%5+200)
			set.GetOrCreateCounter(metricName).Inc()
			i++
		}
	})
}

// Benchmark memory usage with small number of metrics
func BenchmarkVictoriaMetrics_MemoryUsage_100Metrics(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		set := metrics.NewSet()
		values := make([]uint64, 100)
		for j := 0; j < 100; j++ {
			counter := set.NewCounter(fmt.Sprintf("counter_%d", j))
			counter.Inc()
			idx := j
			values[j] = uint64(j)
			_ = set.NewGauge(fmt.Sprintf("gauge_%d", j), func() float64 { return float64(values[idx]) })
		}
	}
}

// Benchmark memory usage with medium number of metrics
func BenchmarkVictoriaMetrics_MemoryUsage_1000Metrics(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		set := metrics.NewSet()
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

// Benchmark realistic HTTP metrics scenario
func BenchmarkVictoriaMetrics_RealisticHTTPMetrics(b *testing.B) {
	set := metrics.NewSet()

	methods := []string{"GET", "POST", "PUT", "DELETE"}
	paths := []string{"/api/users", "/api/orders", "/api/products", "/api/auth"}
	statuses := []string{"200", "201", "400", "404", "500"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		method := methods[i%len(methods)]
		path := paths[i%len(paths)]
		status := statuses[i%len(statuses)]

		// Increment request counter
		counterName := fmt.Sprintf(`http_requests_total{method="%s",path="%s",status="%s"}`, method, path, status)
		set.GetOrCreateCounter(counterName).Inc()

		// Update duration summary
		summaryName := fmt.Sprintf(`http_request_duration_seconds{method="%s",path="%s"}`, method, path)
		set.GetOrCreateSummary(summaryName).Update(float64(i%1000) / 1000.0)
	}
}
