package dispatcher

import (
	"learningGolang/internal/queue"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
	"sync"
	"time"
)

type DispatcherConfig struct {
	PollInterval time.Duration
	WorkerCount  int
}

type Dispatcher struct {
	queue        *queue.Queue
	config       DispatcherConfig
	Queuemetrics metrics.QueueMetrics
	stopChan     chan struct{}
	wg           sync.WaitGroup
	closeOnce    sync.Once
}

// nice it is important to close channel only once.

func NewDispatcher(q *queue.Queue, config DispatcherConfig, m metrics.QueueMetrics) *Dispatcher {
	return &Dispatcher{
		queue:        q,
		config:       config,
		Queuemetrics: m,
		stopChan:     make(chan struct{}),
	}
}
func (d *Dispatcher) Start() {
	d.wg.Add(d.config.WorkerCount)
	for i := 0; i < d.config.WorkerCount; i++ {
		go d.worker()
	}

}
func (d *Dispatcher) Stop() {
	d.closeOnce.Do(func() {
		close(d.stopChan)
		d.wg.Wait()
	})
}

func (d *Dispatcher) worker() {
	defer d.wg.Done()
	for {
		select {
		case <-d.stopChan:
			return
		default:
			msg, ok := d.queue.Dequeue()
			if !ok {
				select {
				case <-d.stopChan:
					return
				case <-time.After(d.config.PollInterval):
					continue
				}
			}
			d.processMessage(msg)
		}
	}
}

func (d *Dispatcher) processMessage(msg *message.Message) {
	start := time.Now()
	defer func() {
		d.Queuemetrics.MessageProcessingDuration.Observe(float64(time.Since(start).Seconds()))
		d.Queuemetrics.MessagesProcessed.Inc()
	}()
	time.Sleep(10 * time.Millisecond)
}
