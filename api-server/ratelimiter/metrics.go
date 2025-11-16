package ratelimiter

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	AllowedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "leakybucket_allowed_total",
			Help: "Total number of allowed requests",
		},
	)

	RejectedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "leakybucket_rejected_total",
			Help: "Total number of rejected requests",
		},
	)
)

func InitMetrics(bucket *LeakyBucket, p prometheus.Registerer) {

	p.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "leakybucket_fill_level",
				Help: "Current fill level of the leaky bucket",
			},
			func() float64 { return bucket.PeekFill() }, // closure captures your bucket
		),
		AllowedTotal,
		RejectedTotal,
	)
}
