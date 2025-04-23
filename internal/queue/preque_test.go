package queue_test

import (
	"errors"
	"learningGolang/internal/dispatcher/config"
	"learningGolang/internal/queue"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldGiveInvalidQueueSizeWhenInvalidSizePassed(t *testing.T) {
	metrics := metrics.NewQueueMetrics()
	config := queue.QueueConfig(config.DefaultConfig().Queue)
	config.MaxSize = -1
	_, err := queue.NewQueue(config, *metrics)
	ErrorMessage := errors.New("invalid queue size")
	assert.Equal(t, ErrorMessage, err)
}
func TestQueueBasicOperation(t *testing.T) {
	metrics := metrics.NewQueueMetrics()
	queue, err := queue.NewQueue(queue.QueueConfig(config.DefaultConfig().Queue), *metrics)
	assert.NoError(t, err)
	defer queue.Close()

	msg := message.NewMessage("1", "API", "Message 1 to API")
	err = queue.Enqueue(msg)
	assert.NoError(t, err)
	assert.Equal(t, 1, queue.Size())

	dequedMessage, ok := queue.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, msg.ID, dequedMessage.ID)
	assert.Equal(t, 0, queue.Size())
}

func TestRaiseErrorWhenEnquedOnClosedChannel(t *testing.T) {
	metrics := metrics.NewQueueMetrics()
	queue, err := queue.NewQueue(queue.QueueConfig(config.DefaultConfig().Queue), *metrics)
	assert.NoError(t, err)
	ErrorMessage := errors.New("queue is closed")
	queue.Close()
	err = queue.Enqueue(message.NewMessage("2", "API", "DEMO"))
	assert.Equal(t, ErrorMessage, err)
}

func TestShouldReturnFalseWhenNothingToDeque(t *testing.T) {
	metrics := metrics.NewQueueMetrics()
	queue, err := queue.NewQueue(queue.QueueConfig(config.DefaultConfig().Queue), *metrics)
	assert.NoError(t, err)
	_, result := queue.Dequeue()
	assert.False(t, result)

}
func TestQueueFullAndNotClearedBeforeMaxRetry(t *testing.T) {
	metrics := metrics.NewQueueMetrics()
	config := queue.QueueConfig(config.DefaultConfig().Queue)
	config.MaxSize = 1
	queue, err := queue.NewQueue(config, *metrics)
	assert.NoError(t, err)
	defer queue.Close()
	msg1 := message.NewMessage("1", "API", "payload1")
	msg2 := message.NewMessage("2", "DB", "payload2")

	err = queue.Enqueue(msg1)
	assert.NoError(t, err)

	// The Second messgae will fail.
	err = queue.Enqueue(msg2)
	ErrMaxRetry := errors.New("max retry attempts reached")
	assert.Equal(t, ErrMaxRetry, err)
}

// func TestConcurrenctLoadWithMetrics(t *testing.T) {
// 	config := cmd.Config{
// 		Queue: cmd.QueueConfig{
// 			MaxSize:          10,
// 			InitialBackoff:   2 * time.Millisecond,
// 			MaxBackoff:       50 * time.Millisecond,
// 			BackoffFactor:    2,
// 			MaxRetryAttempts: 5,
// 		},
// 		Dispatcher: cmd.DispatcherConfig{
// 			PollInterval: 10 * time.Millisecond,
// 			WorkerCount:  4,
// 		},
// 	}
// 	prometheus.DefaultRegisterer=prometheus.NewRegistry()
// 	metrics:=metrics.NewQueueMetrics()
// 	metrics.Register()
// 	queue,err:=queue.NewQueue(queue.QueueConfig(config.Queue),*metrics)
// 	assert.NoError(t,err)
// 	defer queue.Close()

// 	dispatcher:=dispatcher.NewDispatcher(queue,dispatcher.DispatcherConfig(config.Dispatcher),*metrics)
// 	dispatcher.Start()
// 	defer dispatcher.Stop()

// 	var(
// 		wg 	sync.WaitGroup
// 		totalMessage  =500
// 		producers  =2,
// 	)

// }
