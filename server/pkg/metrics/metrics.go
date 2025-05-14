package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path"},
	)

	SuccessRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_success_total",
			Help: "Successful HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	ErrorRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_error_total",
			Help: "Failed HTTP requests",
		},
		[]string{"method", "path"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Request handling duration",
			Buckets: []float64{0.1, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)

	QPS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_qps",
			Help: "Requests per second",
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(
		TotalRequests,
		SuccessRequests,
		ErrorRequests,
		RequestDuration,
		QPS,
	)
}
