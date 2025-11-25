// Package metrics provides Prometheus metrics instrumentation for services.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// FulfillmentMetrics holds metrics specific to the Fulfillment service.
type FulfillmentMetrics struct {
	*Metrics

	// Fulfillment-specific metrics
	ShipmentsCreated   prometheus.Counter
	ShipmentsShipped   prometheus.Counter
	ShipmentsDelivered prometheus.Counter
	ShipmentsFailed    prometheus.Counter
	ShipmentsCancelled prometheus.Counter

	ShipmentsByStatus   *prometheus.GaugeVec
	ShipmentsByCarrier  *prometheus.CounterVec
	ShipmentTime        prometheus.Histogram
	DeliveryTime        prometheus.Histogram
	ShipmentItemCount   prometheus.Histogram
}

// NewFulfillmentMetrics creates metrics for the Fulfillment service.
func NewFulfillmentMetrics() *FulfillmentMetrics {
	base := New(Config{
		ServiceName: "fulfillment",
		Namespace:   "ecommerce",
		Subsystem:   "fulfillment",
	})

	return &FulfillmentMetrics{
		Metrics: base,

		ShipmentsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipments_created_total",
			Help:      "Total number of shipments created",
		}),

		ShipmentsShipped: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipments_shipped_total",
			Help:      "Total number of shipments shipped",
		}),

		ShipmentsDelivered: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipments_delivered_total",
			Help:      "Total number of shipments delivered",
		}),

		ShipmentsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipments_failed_total",
			Help:      "Total number of shipments that failed",
		}),

		ShipmentsCancelled: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipments_cancelled_total",
			Help:      "Total number of shipments cancelled",
		}),

		ShipmentsByStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "ecommerce",
				Subsystem: "fulfillment",
				Name:      "shipments_by_status",
				Help:      "Current number of shipments by status",
			},
			[]string{"status"},
		),

		ShipmentsByCarrier: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "ecommerce",
				Subsystem: "fulfillment",
				Name:      "shipments_by_carrier_total",
				Help:      "Total shipments by carrier",
			},
			[]string{"carrier"},
		),

		ShipmentTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipment_preparation_duration_seconds",
			Help:      "Time from order to shipment",
			Buckets:   []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		}),

		DeliveryTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "delivery_duration_seconds",
			Help:      "Time from shipment to delivery",
			Buckets:   []float64{3600, 7200, 14400, 28800, 86400, 172800, 259200, 432000, 604800},
		}),

		ShipmentItemCount: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: "ecommerce",
			Subsystem: "fulfillment",
			Name:      "shipment_item_count",
			Help:      "Number of items per shipment",
			Buckets:   []float64{1, 2, 3, 5, 10, 20, 50},
		}),
	}
}

// RecordShipmentCreated records a new shipment creation.
func (m *FulfillmentMetrics) RecordShipmentCreated(carrier string, itemCount int) {
	m.ShipmentsCreated.Inc()
	m.ShipmentsByCarrier.WithLabelValues(carrier).Inc()
	m.ShipmentItemCount.Observe(float64(itemCount))
}

// RecordShipmentShipped records a shipment being shipped.
func (m *FulfillmentMetrics) RecordShipmentShipped(preparationSeconds float64) {
	m.ShipmentsShipped.Inc()
	m.ShipmentTime.Observe(preparationSeconds)
}

// RecordShipmentDelivered records a shipment delivery.
func (m *FulfillmentMetrics) RecordShipmentDelivered(deliverySeconds float64) {
	m.ShipmentsDelivered.Inc()
	m.DeliveryTime.Observe(deliverySeconds)
}

// RecordShipmentFailed records a failed shipment.
func (m *FulfillmentMetrics) RecordShipmentFailed() {
	m.ShipmentsFailed.Inc()
}

// RecordShipmentCancelled records a cancelled shipment.
func (m *FulfillmentMetrics) RecordShipmentCancelled() {
	m.ShipmentsCancelled.Inc()
}

// SetShipmentsByStatus sets the current count of shipments by status.
func (m *FulfillmentMetrics) SetShipmentsByStatus(status string, count float64) {
	m.ShipmentsByStatus.WithLabelValues(status).Set(count)
}
