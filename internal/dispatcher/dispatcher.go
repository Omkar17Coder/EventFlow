package dispatcher

import (
	"learningGolang/internal/config"
	"learningGolang/internal/queue"
	"learningGolang/internal/retry"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
	"log"
	"sync"
	"time"
)

type Dispatcher struct {
	inputQueue   *queue.Queue
	config       config.DispatcherConfig
	processors   map[string]ChannelProcessor
	retryMgrs    map[string]retry.RetryManager
	queueMetrics metrics.QueueMetrics

	globalStop   chan struct{}
	workerStop   chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	workerCount  int
	desiredCount int
	nextWorkerID int
}

type ChannelProcessor interface {
	GetInputChannel() chan message.Message
	StartProcessingChannel()
	StopProcessingChannel()
}

func NewDispatcher(q *queue.Queue, config config.DispatcherConfig, m metrics.QueueMetrics, processors map[string]ChannelProcessor, retryManagers map[string]retry.RetryManager) *Dispatcher {
	return &Dispatcher{
		inputQueue:   q,
		config:       config,
		processors:   processors,
		retryMgrs:    retryManagers,
		queueMetrics: m,
		globalStop:   make(chan struct{}),
		workerStop:   make(chan struct{}),
		desiredCount: config.WorkerCount,
	}
}

func (d *Dispatcher) Start() {
	for _, p := range d.processors {
		p.StartProcessingChannel()
	}
	for _, mgr := range d.retryMgrs {
		mgr.StartRetryManager()
	}

	for i := 0; i < d.desiredCount; i++ {
		d.startWorker()
	}

	go d.monitorAndScaleWorkers()
}

func (d *Dispatcher) monitorAndScaleWorkers() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.globalStop:
			return
		case <-ticker.C:
			d.adjustWorkers()
		}
	}
}

func (d *Dispatcher) adjustWorkers() {
	queueSize := d.inputQueue.Size()
	queueCapacity := cap(d.inputQueue.MessagesList)
	loadPercentage := float64(queueSize) / float64(queueCapacity) * 100

	d.mu.Lock()
	defer d.mu.Unlock()

	if queueSize >= queueCapacity-1 {
		d.desiredCount = min(d.desiredCount+3, d.config.WorkerCount*2)
	} else if loadPercentage > 80 {
		d.desiredCount = min(d.desiredCount+2, d.config.WorkerCount*2)
	} else if loadPercentage < 20 && d.desiredCount > d.config.WorkerCount {
		d.desiredCount = max(d.desiredCount-1, d.config.WorkerCount)
	}

	d.syncWorkers()
}

func (d *Dispatcher) syncWorkers() {
	diff := d.desiredCount - d.workerCount
	if diff > 0 {
		log.Printf("[Dispatcher] Scaling up: adding %d workers (current: %d, desired: %d)", diff, d.workerCount, d.desiredCount)
		for i := 0; i < diff; i++ {
			d.startWorker()
		}
	} else if diff < 0 {
		log.Printf("[Dispatcher] Scaling down: removing %d workers (current: %d, desired: %d)", -diff, d.workerCount, d.desiredCount)
		for i := 0; i < -diff; i++ {
			d.stopWorker()
		}
	}
}

func (d *Dispatcher) startWorker() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.workerCount++
	d.nextWorkerID++
	workerID := d.nextWorkerID

	d.wg.Add(1)
	go d.worker(workerID)
}

func (d *Dispatcher) stopWorker() {
	select {
	case d.workerStop <- struct{}{}:
	default:
	}
}

func (d *Dispatcher) Stop() {
	close(d.globalStop)
	d.wg.Wait()
}

func (d *Dispatcher) worker(workerID int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Worker-%d] Panic recovered: %v", workerID, r)
		}
		d.mu.Lock()
		d.workerCount--
		d.mu.Unlock()
		d.wg.Done()
	}()

	log.Printf("[Worker-%d] Started", workerID)

	for {
		select {
		case <-d.globalStop:
			log.Printf("[Worker-%d] Shutting down (global stop)", workerID)
			return
		case <-d.workerStop:
			log.Printf("[Worker-%d] Shutting down (scale down)", workerID)
			return
		default:
		}

		msg, ok := d.inputQueue.Dequeue()
		if !ok {
			select {
			case <-d.globalStop:
				return
			case <-d.workerStop:
				return
			case <-time.After(d.config.PollInterval):
				continue
			}
		}

		if msg == nil {
			continue
		}

		d.processMessage(msg, workerID)
	}
}

func (d *Dispatcher) processMessage(msg *message.Message, workerID int) {
	start := time.Now()

	defer func() {
		d.queueMetrics.MessageProcessingDuration.Observe(float64(time.Since(start).Seconds()))
		d.queueMetrics.MessagesProcessed.Inc()
	}()

	processor, exists := d.processors[msg.Type]
	if !exists {
		log.Printf("[Worker-%d] No processor found for type %s", workerID, msg.Type)
		return
	}

	// Try sending to processor channel
	select {
	case processor.GetInputChannel() <- *msg:
		log.Printf("[Worker-%d] Message sent to processor: %s", workerID, msg.Type)
		return
	case <-time.After(50 * time.Millisecond):
	}

	// Retry once more
	select {
	case processor.GetInputChannel() <- *msg:
		log.Printf("[Worker-%d] Message retried successfully to processor: %s", workerID, msg.Type)
		return
	case <-time.After(50 * time.Millisecond):
		log.Printf("[Worker-%d] Dropping message after retries: %s", workerID, msg.Type)
		// Optionally requeue into retry manager if available
		if retryMgr, ok := d.retryMgrs[msg.Type]; ok {
			retryMgr.EnqueueIntoRetryQueue(*msg)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
