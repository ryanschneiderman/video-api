package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "myapp_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "myapp_http_request_duration_seconds",
			Help:    "Histogram of latencies for HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	) 
	WorkerProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "worker_processing_duration_seconds",
			Help:    "Histogram of the duration to process messages",
			Buckets: prometheus.DefBuckets,
		},
	)
	WorkerErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_errors_total",
			Help: "Total number of errors encountered by the worker",
		},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(HTTPRequestTotal)
	prometheus.MustRegister(HTTPRequestDuration)
}

func InstrumentHandler(path string, handlerFunc func() error) func() error {
	return func() error {
		start := time.Now()
		err := handlerFunc()
		duration := time.Since(start).Seconds()
		HTTPRequestTotal.With(prometheus.Labels{"path": path, "method": "GET"}).Inc()
		HTTPRequestDuration.With(prometheus.Labels{"path": path, "method": "GET"}).Observe(duration)
		return err
	}
}


func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method
		status := c.Writer.Status()

		HTTPRequestTotal.WithLabelValues(path, method, formatStatus(status)).Inc()

		HTTPRequestDuration.WithLabelValues(path, method, formatStatus(status)).Observe(duration)
	}
}

func formatStatus(code int) string {
	return fmt.Sprintf("%d", code)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
