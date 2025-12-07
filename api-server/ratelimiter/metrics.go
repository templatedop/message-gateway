package ratelimiter

import (
	"github.com/VictoriaMetrics/metrics"
)

var (
	AllowedTotal  *metrics.Counter
	RejectedTotal *metrics.Counter
	FillLevel     *metrics.Gauge
)

func InitMetrics(bucket *LeakyBucket, set *metrics.Set) {
	if set == nil {
		set = metrics.NewSet()
	}

	AllowedTotal = set.NewCounter("leakybucket_allowed_total")
	RejectedTotal = set.NewCounter("leakybucket_rejected_total")
	FillLevel = set.NewGauge("leakybucket_fill_level", func() float64 {
		return bucket.PeekFill()
	})
}
