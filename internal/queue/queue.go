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
	messagesList  chan *message.Message
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
		messagesList: make(chan *message.Message, config.MaxSize),
		config:       config,
		metrics:      &metrics,
		done:         make(chan struct{}),
	}, nil
}

func (q *Queue) Enqueue(msg *message.Message) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	// read lock.
	select {
	case <-q.done:
		return ErrQueueClosed
	default:

	}
	backoff := q.config.MaxBackoff
	attempt := 0

	for {
		select {
		case <-q.done:
			return ErrQueueClosed

		case q.messagesList <- msg:
			q.metrics.MessagesQueued.Add(1)

			return nil
		default:
			attempt++
			if attempt > q.config.MaxRetryAttempts {
				q.metrics.EnqueueRetriesExhausted.Add(1)
				return ErrMaxRetry
			}
			q.metrics.EnqueueRetries.Inc()
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * q.config.BackoffFactor)
			if backoff > q.config.MaxBackoff {
				backoff = q.config.MaxBackoff
			}
		}
	}

}

func (q *Queue) Dequeue() (*message.Message, bool) {
	select {
	case <-q.done:
		return nil, false
	case msg, ok := <-q.messagesList:
		if ok {
			q.metrics.MessagesDequeued.Add(1)
		}
		return msg, ok
	default:
		return nil, false
	}
}
func (q *Queue) Close() {
	q.closeOnce.Do(func() {
		q.mu.Lock()
		close(q.done)
		q.mu.Unlock()
		close(q.messagesList)
	})
}

func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.messagesList)
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
