// Package metrics provides Prometheus metrics instrumentation for services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PaymentMetrics holds metrics specific to the Payment service.
type PaymentMetrics struct {
	*Metrics

	// Payment-specific metrics
	PaymentsTotal     *prometheus.CounterVec
	PaymentsSucceeded prometheus.Counter
	PaymentsFailed    prometheus.Counter
	RefundsProcessed  prometheus.Counter

	PaymentAmount         *prometheus.HistogramVec
	PaymentProcessingTime prometheus.Histogram
	PaymentsByMethod      *prometheus.CounterVec
	PaymentsByStatus      *prometheus.GaugeVec
}

// NewPaymentMetrics creates metrics for the Payment service.
func NewPaymentMetrics() *PaymentMetrics {
	base := New(Config{
		ServiceName: "payment",
		Namespace:   "ecommerce",
		Subsystem:   "payment",
	})

	return &PaymentMetrics{
		Metrics: base,

		PaymentsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "ecommerce",
				Subsystem: "payment",
				Name:      "payments_total",
				Help:      "Total number of payment attempts",
			},
			[]string{"method", "status"},
		),

		PaymentsSucceeded: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "payment",
			Name:      "payments_succeeded_total",
			Help:      "Total number of successful payments",
		}),

		PaymentsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "payment",
			Name:      "payments_failed_total",
			Help:      "Total number of failed payments",
		}),

		RefundsProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "payment",
			Name:      "refunds_processed_total",
			Help:      "Total number of refunds processed",
		}),

		PaymentAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "ecommerce",
				Subsystem: "payment",
				Name:      "payment_amount_dollars",
				Help:      "Distribution of payment amounts in dollars",
				Buckets:   []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
			},
			[]string{"method"},
		),

		PaymentProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "payment",
			Name:      "payment_processing_duration_seconds",
			Help:      "Time spent processing payments",
			Buckets:   []float64{.1, .25, .5, 1, 2.5, 5, 10, 30},
		}),

		PaymentsByMethod: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "ecommerce",
				Subsystem: "payment",
				Name:      "payments_by_method_total",
				Help:      "Total payments by payment method",
			},
			[]string{"method"},
		),

		PaymentsByStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ecommerce",
				Subsystem: "payment",
				Name:      "payments_by_status",
				Help:      "Current number of payments by status",
			},
			[]string{"status"},
		),
	}
}

// RecordPaymentAttempt records a payment attempt.
func (m *PaymentMetrics) RecordPaymentAttempt(method, status string, amountCents int64) {
	m.PaymentsTotal.WithLabelValues(method, status).Inc()
	m.PaymentsByMethod.WithLabelValues(method).Inc()
	m.PaymentAmount.WithLabelValues(method).Observe(float64(amountCents) / 100)
}

// RecordPaymentSucceeded records a successful payment.
func (m *PaymentMetrics) RecordPaymentSucceeded() {
	m.PaymentsSucceeded.Inc()
}

// RecordPaymentFailed records a failed payment.
func (m *PaymentMetrics) RecordPaymentFailed() {
	m.PaymentsFailed.Inc()
}

// RecordRefund records a refund.
func (m *PaymentMetrics) RecordRefund() {
	m.RefundsProcessed.Inc()
}

// SetPaymentsByStatus sets the current count of payments by status.
func (m *PaymentMetrics) SetPaymentsByStatus(status string, count float64) {
	m.PaymentsByStatus.WithLabelValues(status).Set(count)
}

// RecordProcessingTime records the time taken to process a payment.
func (m *PaymentMetrics) RecordProcessingTime(seconds float64) {
	m.PaymentProcessingTime.Observe(seconds)
}
