package queue

import (
	"errors"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
	"sync"
	"time"
)

var (
	ErrQueueFull   = errors.New("queues is full")
	ErrQueueClosed = errors.New("queue is closed")
	ErrInvalidSize = errors.New("invalid queue size")
	ErrMaxRetry    = errors.New("max retry attempts reached")
)

type QueueConfig struct {
	MaxSize          int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	BackoffFactor    float64
	MaxRetryAttempts int
}

type Queue struct {
	MessagesList  chan message.Message
	config        QueueConfig
	metrics       *metrics.QueueMetrics
	mu            sync.RWMutex
	closeOnce     sync.Once
	done          chan struct{}
	wg            sync.WaitGroup
	pendingWrites int32
}

func NewQueue(config QueueConfig, metrics metrics.QueueMetrics) (*Queue, error) {
	if config.MaxSize <= 0 {
		return nil, ErrInvalidSize
	}

	return &Queue{
		MessagesList: make(chan message.Message, config.MaxSize),
		config:       config,
		metrics:      &metrics,
		done:         make(chan struct{}),
	}, nil
}

func (q *Queue) Enqueue(msg *message.Message) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	select {
	case <-q.done:
		return ErrQueueClosed
	case q.MessagesList <- *msg:
		q.metrics.MessagesQueued.Add(1)
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Dequeue() (*message.Message, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	select {
	case <-q.done:
		return nil, false
	case msg, ok := <-q.MessagesList:
		if ok {
			q.metrics.MessagesDequeued.Add(1)
		}
		return &msg, ok
	default:
		return nil, false
	}
}
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closeOnce.Do(func() {
		q.mu.Lock()
		close(q.done)
		q.mu.Unlock()
		close(q.MessagesList)
	})
}

func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.MessagesList)
}

// func (q *Queue) Capacity() int {
// 	q.mu.RLock()
// 	defer q.mu.RUnlock()
// 	return cap(q.messagesList)
// }

// func (q *Queue) IsClosed() bool {
// 	select {
// 	case <-q.done:
// 		return true
// 	default:
// 		return false
// 	}
// }
