// K6 Load Test Configuration for E-Commerce Platform

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const orderCreationTime = new Trend('order_creation_time');
const paymentProcessingTime = new Trend('payment_processing_time');
const ordersCreated = new Counter('orders_created');
const paymentsProcessed = new Counter('payments_processed');

// Test configuration
export const options = {
  scenarios: {
    // Smoke test - quick validation
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '1m',
      tags: { test_type: 'smoke' },
    },
    // Load test - normal load
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },   // Ramp up to 50 users
        { duration: '5m', target: 50 },   // Stay at 50 users
        { duration: '2m', target: 100 },  // Ramp up to 100 users
        { duration: '5m', target: 100 },  // Stay at 100 users
        { duration: '2m', target: 0 },    // Ramp down
      ],
      tags: { test_type: 'load' },
    },
    // Stress test - find breaking point
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 100 },
        { duration: '5m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '5m', target: 200 },
        { duration: '2m', target: 300 },
        { duration: '5m', target: 300 },
        { duration: '5m', target: 0 },
      ],
      tags: { test_type: 'stress' },
    },
    // Spike test - sudden traffic spike
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '10s', target: 500 },
        { duration: '3m', target: 500 },
        { duration: '10s', target: 100 },
        { duration: '3m', target: 100 },
        { duration: '10s', target: 0 },
      ],
      tags: { test_type: 'spike' },
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],  // 95% < 500ms, 99% < 1s
    http_req_failed: ['rate<0.01'],                   // Error rate < 1%
    errors: ['rate<0.05'],                            // Custom error rate < 5%
    order_creation_time: ['p(95)<1000'],              // Order creation p95 < 1s
    payment_processing_time: ['p(95)<2000'],          // Payment processing p95 < 2s
  },
};

// Environment configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ORDER_SERVICE = __ENV.ORDER_SERVICE || `${BASE_URL}`;
const PAYMENT_SERVICE = __ENV.PAYMENT_SERVICE || 'http://localhost:8081';
const FULFILLMENT_SERVICE = __ENV.FULFILLMENT_SERVICE || 'http://localhost:8082';

// Test data generators
function generateOrderData() {
  return {
    customer_id: `cust-${Math.random().toString(36).substr(2, 9)}`,
    items: [
      {
        product_id: `prod-${Math.floor(Math.random() * 1000)}`,
        name: 'Test Product',
        quantity: Math.floor(Math.random() * 5) + 1,
        unit_price: Math.floor(Math.random() * 10000) + 100,
        currency: 'USD',
      },
    ],
  };
}

function generatePaymentData(orderId, amount) {
  return {
    order_id: orderId,
    amount: amount,
    currency: 'USD',
    method: ['credit_card', 'debit_card', 'paypal'][Math.floor(Math.random() * 3)],
  };
}

// Main test function
export default function () {
  group('Order Flow', function () {
    // Create order
    group('Create Order', function () {
      const orderData = generateOrderData();
      const startTime = Date.now();

      const orderRes = http.post(
        `${ORDER_SERVICE}/api/v1/orders`,
        JSON.stringify(orderData),
        {
          headers: { 'Content-Type': 'application/json' },
          tags: { name: 'CreateOrder' },
        }
      );

      const duration = Date.now() - startTime;
      orderCreationTime.add(duration);

      const orderSuccess = check(orderRes, {
        'order created': (r) => r.status === 201,
        'order has id': (r) => r.json('id') !== undefined,
      });

      if (!orderSuccess) {
        errorRate.add(1);
      } else {
        ordersCreated.add(1);
        errorRate.add(0);
      }

      // If order created successfully, process payment
      if (orderRes.status === 201) {
        const orderId = orderRes.json('id');
        const totalAmount = orderRes.json('total_amount') || 1000;

        sleep(0.5); // Small delay between operations

        // Process payment
        group('Process Payment', function () {
          const paymentData = generatePaymentData(orderId, totalAmount);
          const paymentStart = Date.now();

          const paymentRes = http.post(
            `${PAYMENT_SERVICE}/api/v1/payments`,
            JSON.stringify(paymentData),
            {
              headers: { 'Content-Type': 'application/json' },
              tags: { name: 'ProcessPayment' },
            }
          );

          const paymentDuration = Date.now() - paymentStart;
          paymentProcessingTime.add(paymentDuration);

          const paymentSuccess = check(paymentRes, {
            'payment processed': (r) => r.status === 201 || r.status === 200,
            'payment has id': (r) => r.json('id') !== undefined,
          });

          if (!paymentSuccess) {
            errorRate.add(1);
          } else {
            paymentsProcessed.add(1);
            errorRate.add(0);
          }
        });
      }
    });

    // Get order status
    group('Get Order', function () {
      const orderRes = http.get(`${ORDER_SERVICE}/api/v1/orders/test-order-id`, {
        tags: { name: 'GetOrder' },
      });

      check(orderRes, {
        'order retrieved or not found': (r) => r.status === 200 || r.status === 404,
      });
    });
  });

  group('Health Checks', function () {
    // Order service health
    const orderHealth = http.get(`${ORDER_SERVICE}/health`, {
      tags: { name: 'OrderHealth' },
    });
    check(orderHealth, {
      'order service healthy': (r) => r.status === 200,
    });

    // Payment service health
    const paymentHealth = http.get(`${PAYMENT_SERVICE}/health`, {
      tags: { name: 'PaymentHealth' },
    });
    check(paymentHealth, {
      'payment service healthy': (r) => r.status === 200,
    });

    // Fulfillment service health
    const fulfillmentHealth = http.get(`${FULFILLMENT_SERVICE}/health`, {
      tags: { name: 'FulfillmentHealth' },
    });
    check(fulfillmentHealth, {
      'fulfillment service healthy': (r) => r.status === 200,
    });
  });

  sleep(1); // Think time between iterations
}

// Summary handler
export function handleSummary(data) {
  return {
    'reports/load-test-summary.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  // Simple text summary
  let summary = '\n========== Load Test Summary ==========\n\n';

  summary += `Total Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `Failed Requests: ${data.metrics.http_req_failed.values.passes}\n`;
  summary += `Request Duration (p95): ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `Request Duration (p99): ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n`;

  if (data.metrics.orders_created) {
    summary += `Orders Created: ${data.metrics.orders_created.values.count}\n`;
  }
  if (data.metrics.payments_processed) {
    summary += `Payments Processed: ${data.metrics.payments_processed.values.count}\n`;
  }

  summary += '\n========================================\n';

  return summary;
}
