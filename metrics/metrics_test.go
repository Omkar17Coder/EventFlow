package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestQueueMetrics(t *testing.T) {
	// Create new metrics
	metrics := NewQueueMetrics()

	// Test registration
	err := metrics.Register()
	assert.NoError(t, err)

	// Test all metrics
	metrics.MessagesQueued.Inc()
	metrics.MessagesDequeued.Inc()
	metrics.MessagesProcessed.Inc()
	metrics.EnqueueRetries.Inc()
	metrics.EnqueueRetriesExhausted.Inc()
	metrics.QueueSize.Inc()

	// Test histogram
	timer := prometheus.NewTimer(metrics.MessageProcessingDuration)
	time.Sleep(10 * time.Millisecond)
	timer.ObserveDuration()

	// Test Collect and Describe
	collectCh := make(chan prometheus.Metric, 10)
	describeCh := make(chan *prometheus.Desc, 10)
	metrics.Collect(collectCh)
	metrics.Describe(describeCh)
}
