package metrics

import "github.com/prometheus/client_golang/prometheus"

type QueueMetrics struct {
	MessagesQueued            prometheus.Counter
	MessagesDequeued          prometheus.Counter
	MessagesProcessed         prometheus.Counter
	EnqueueRetries            prometheus.Counter
	EnqueueRetriesExhausted   prometheus.Counter
	MessageProcessingDuration prometheus.Histogram
	QueueSize                 prometheus.Gauge
}

func NewQueueMetrics() *QueueMetrics {
	return &QueueMetrics{
		MessagesQueued: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "message_queue_enqueued_total",
			Help: "Total number of messages enqueued",
		}),
		MessagesDequeued: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "message_queue_dequeued_total",
			Help: "Total number of messages dequeued",
		}),
		MessagesProcessed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "message_queue_processed_total",
			Help: "Total number of messages processed",
		}),
		EnqueueRetries: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "message_queue_enqueue_retries_total",
			Help: "Total number of enqueue retries",
		}),
		EnqueueRetriesExhausted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "message_queue_enqueue_retries_exhausted_total",
			Help: "Total number of failed enqueue attempts after retries",
		}),
		MessageProcessingDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "message_processing_duration_seconds",
			Help:    "Time taken to process a message",
			Buckets: prometheus.DefBuckets,
		}),
		QueueSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "message_queue_current_size",
			Help: "Current number of messages in the queue",
		}),
	}
}

func (m *QueueMetrics) Register() error {
	return prometheus.Register(m)
}

func (m *QueueMetrics) Collect(ch chan<- prometheus.Metric) {
	m.MessagesQueued.Collect(ch)
	m.MessagesDequeued.Collect(ch)
	m.MessagesProcessed.Collect(ch)
	m.EnqueueRetries.Collect(ch)
	m.EnqueueRetriesExhausted.Collect(ch)
	m.MessageProcessingDuration.Collect(ch)
	m.QueueSize.Collect(ch)
}

func (m *QueueMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.MessagesQueued.Describe(ch)
	m.MessagesDequeued.Describe(ch)
	m.MessagesProcessed.Describe(ch)
	m.EnqueueRetries.Describe(ch)
	m.EnqueueRetriesExhausted.Describe(ch)
	m.MessageProcessingDuration.Describe(ch)
	m.QueueSize.Describe(ch)
}
