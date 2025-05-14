package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTPRequestsTotal 总请求数（包括成功和失败）
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, path and status",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration distribution",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 1.5, 2},
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
	)
}
