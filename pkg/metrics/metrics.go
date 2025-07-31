package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all the prometheus metrics
type Metrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	ActiveConnections prometheus.Gauge
	DatabaseQueries   *prometheus.CounterVec
	ErrorsTotal       *prometheus.CounterVec
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		ActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
		DatabaseQueries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "component"},
		),
	}

	// Register metrics with prometheus
	prometheus.MustRegister(
		m.RequestsTotal,
		m.RequestDuration,
		m.ActiveConnections,
		m.DatabaseQueries,
		m.ErrorsTotal,
	)

	return m
}

// PrometheusMiddleware creates a Gin middleware for prometheus metrics
func (m *Metrics) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Increment active connections
		m.ActiveConnections.Inc()
		defer m.ActiveConnections.Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		
		m.RequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()
		
		m.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// RecordDatabaseQuery records a database query metric
func (m *Metrics) RecordDatabaseQuery(operation, table string) {
	m.DatabaseQueries.WithLabelValues(operation, table).Inc()
}

// RecordError records an error metric
func (m *Metrics) RecordError(errorType, component string) {
	m.ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// Handler returns the prometheus metrics handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// Global metrics instance
var globalMetrics *Metrics

// InitGlobalMetrics initializes the global metrics instance
func InitGlobalMetrics() *Metrics {
	globalMetrics = NewMetrics()
	return globalMetrics
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		globalMetrics = NewMetrics()
	}
	return globalMetrics
}

// Convenience functions using global metrics
func RecordDatabaseQuery(operation, table string) {
	GetMetrics().RecordDatabaseQuery(operation, table)
}

func RecordError(errorType, component string) {
	GetMetrics().RecordError(errorType, component)
}
