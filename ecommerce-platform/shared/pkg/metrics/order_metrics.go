// Package metrics provides Prometheus metrics instrumentation for services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OrderMetrics holds metrics specific to the Order service.
type OrderMetrics struct {
	*Metrics

	// Order-specific metrics
	OrdersCreated   prometheus.Counter
	OrdersCompleted prometheus.Counter
	OrdersCancelled prometheus.Counter
	OrdersFailed    prometheus.Counter

	OrdersByStatus      *prometheus.GaugeVec
	OrderProcessingTime prometheus.Histogram
	OrderValue          prometheus.Histogram
}

// NewOrderMetrics creates metrics for the Order service.
func NewOrderMetrics() *OrderMetrics {
	base := New(Config{
		ServiceName: "order",
		Namespace:   "ecommerce",
		Subsystem:   "order",
	})

	return &OrderMetrics{
		Metrics: base,

		OrdersCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "order",
			Name:      "orders_created_total",
			Help:      "Total number of orders created",
		}),

		OrdersCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "order",
			Name:      "orders_completed_total",
			Help:      "Total number of orders completed",
		}),

		OrdersCancelled: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "order",
			Name:      "orders_cancelled_total",
			Help:      "Total number of orders cancelled",
		}),

		OrdersFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "order",
			Name:      "orders_failed_total",
			Help:      "Total number of orders that failed processing",
		}),

		OrdersByStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ecommerce",
				Subsystem: "order",
				Name:      "orders_by_status",
				Help:      "Current number of orders by status",
			},
			[]string{"status"},
		),

		OrderProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "order",
			Name:      "order_processing_duration_seconds",
			Help:      "Time spent processing orders",
			Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		}),

		OrderValue: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "order",
			Name:      "order_value_dollars",
			Help:      "Distribution of order values in dollars",
			Buckets:   []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		}),
	}
}

// RecordOrderCreated records a new order creation.
func (m *OrderMetrics) RecordOrderCreated(valueInCents int64) {
	m.OrdersCreated.Inc()
	m.OrderValue.Observe(float64(valueInCents) / 100)
}

// RecordOrderCompleted records an order completion.
func (m *OrderMetrics) RecordOrderCompleted() {
	m.OrdersCompleted.Inc()
}

// RecordOrderCancelled records an order cancellation.
func (m *OrderMetrics) RecordOrderCancelled() {
	m.OrdersCancelled.Inc()
}

// RecordOrderFailed records an order failure.
func (m *OrderMetrics) RecordOrderFailed() {
	m.OrdersFailed.Inc()
}

// SetOrdersByStatus sets the current count of orders by status.
func (m *OrderMetrics) SetOrdersByStatus(status string, count float64) {
	m.OrdersByStatus.WithLabelValues(status).Set(count)
}

// RecordProcessingTime records the time taken to process an order.
func (m *OrderMetrics) RecordProcessingTime(seconds float64) {
	m.OrderProcessingTime.Observe(seconds)
}
