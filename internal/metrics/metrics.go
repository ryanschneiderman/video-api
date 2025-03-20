package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type APIMetrics struct {
	RequestTotal    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewAPIMetrics() *APIMetrics {
	return &APIMetrics{
		RequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "myapp_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"path", "method", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "myapp_http_request_duration_seconds",
				Help:    "Histogram of latencies for HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"path", "method", "status"},
		),
	}
}

func (m *APIMetrics) Register(registry *prometheus.Registry) {
	registry.MustRegister(m.RequestTotal)
	registry.MustRegister(m.RequestDuration)
}

type WorkerMetrics struct {
	ProcessingDuration prometheus.Histogram
	Errors             prometheus.Counter
}

func NewWorkerMetrics() *WorkerMetrics {
	return &WorkerMetrics{
		ProcessingDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "worker_processing_duration_seconds",
				Help:    "Histogram of the duration to process messages",
				Buckets: prometheus.DefBuckets,
			},
		),
		Errors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "worker_errors_total",
				Help: "Total number of errors encountered by the worker",
			},
		),
	}
}

func (m *WorkerMetrics) Register(registry *prometheus.Registry) {
	registry.MustRegister(m.ProcessingDuration)
	registry.MustRegister(m.Errors)
}

func PrometheusMiddleware(m *APIMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method
		status := formatStatus(c.Writer.Status())

		m.RequestTotal.WithLabelValues(path, method, status).Inc()
		m.RequestDuration.WithLabelValues(path, method, status).Observe(duration)
	}
}

func formatStatus(code int) string {
	return fmt.Sprintf("%d", code)
}

func MetricsHandler(registry *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
