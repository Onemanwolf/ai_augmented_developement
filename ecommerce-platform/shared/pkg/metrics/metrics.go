// Package metrics provides Prometheus metrics instrumentation for services.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for a service.
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Business metrics
	EventsPublished *prometheus.CounterVec
	EventsFailed    *prometheus.CounterVec

	// Custom counters and gauges
	customCounters map[string]*prometheus.CounterVec
	customGauges   map[string]*prometheus.GaugeVec
}

// Config holds metrics configuration.
type Config struct {
	ServiceName string
	Namespace   string
	Subsystem   string
}

// New creates a new Metrics instance with standard metrics.
func New(cfg Config) *Metrics {
	namespace := cfg.Namespace
	if namespace == "" {
		namespace = "ecommerce"
	}

	subsystem := cfg.Subsystem
	if subsystem == "" {
		subsystem = cfg.ServiceName
	}

	m := &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status"},
		),

		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),

		EventsPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "events_published_total",
				Help:      "Total number of domain events published",
			},
			[]string{"event_type"},
		),

		EventsFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "events_failed_total",
				Help:      "Total number of domain events that failed to publish",
			},
			[]string{"event_type"},
		),

		customCounters: make(map[string]*prometheus.CounterVec),
		customGauges:   make(map[string]*prometheus.GaugeVec),
	}

	return m
}

// Handler returns the Prometheus HTTP handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTPRequest records an HTTP request.
func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusStr := strconv.Itoa(status)
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
}

// RecordEventPublished records a successfully published event.
func (m *Metrics) RecordEventPublished(eventType string) {
	m.EventsPublished.WithLabelValues(eventType).Inc()
}

// RecordEventFailed records a failed event publication.
func (m *Metrics) RecordEventFailed(eventType string) {
	m.EventsFailed.WithLabelValues(eventType).Inc()
}

// IncrementInFlight increments the in-flight request counter.
func (m *Metrics) IncrementInFlight() {
	m.HTTPRequestsInFlight.Inc()
}

// DecrementInFlight decrements the in-flight request counter.
func (m *Metrics) DecrementInFlight() {
	m.HTTPRequestsInFlight.Dec()
}

// RegisterCounter registers a custom counter metric.
func (m *Metrics) RegisterCounter(name, help string, labels []string) *prometheus.CounterVec {
	counter := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.customCounters[name] = counter
	return counter
}

// RegisterGauge registers a custom gauge metric.
func (m *Metrics) RegisterGauge(name, help string, labels []string) *prometheus.GaugeVec {
	gauge := promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	m.customGauges[name] = gauge
	return gauge
}

// HTTPMiddleware returns middleware that records HTTP metrics.
func (m *Metrics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.IncrementInFlight()
		defer m.DecrementInFlight()

		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		m.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
