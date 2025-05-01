package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type APIMetrics struct {
	MessageProcessed prometheus.Counter
	APIQueueSize     prometheus.Gauge
	WorkerCount      prometheus.Gauge
}

func NewAPIMetrics(reg prometheus.Registerer) *APIMetrics {
	return &APIMetrics{
		MessageProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "messaged_processed_total",
			Help: "number  of messages processed",
		}),
		APIQueueSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "api_queue_size",
			Help: "current_queue_size",
		}),
		WorkerCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "worker_count_total",
			Help: "number of worker count",
		}),
	}
}
func (m *APIMetrics) Register() error {
	return prometheus.Register(m)
}
func (m *APIMetrics) Collect(ch chan<- prometheus.Metric) {
	m.APIQueueSize.Collect(ch)
	m.MessageProcessed.Collect(ch)
	m.WorkerCount.Collect(ch)
}
func (m *APIMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.APIQueueSize.Describe(ch)
	m.MessageProcessed.Describe(ch)
	m.WorkerCount.Describe(ch)
}
