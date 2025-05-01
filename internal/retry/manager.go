package retry

import (
	"learningGolang/internal/config"
	"learningGolang/internal/storage"
	"learningGolang/pkg/message"
	"log"
	"strconv"
	"sync"
	"time"
)

type RetryHandler struct {
	config              config.RetryConfig
	retryQueue          chan message.Message
	mainChannel         chan message.Message
	stopWorkerChan      chan struct{}
	globalStopChan      chan struct{}
	storage             *storage.FileStorage
	channelType         string
	wg                  sync.WaitGroup
	mu                  sync.RWMutex
	WorkersCount        int
	desiredWorkersCount int
	workerID            int // to give each worker a unique ID
}

func NewRetryHandler(cfg config.RetryConfig, storage *storage.FileStorage, channelType string) *RetryHandler {
	return &RetryHandler{
		config:              cfg,
		retryQueue:          make(chan message.Message, cfg.RetryQueueSize),
		stopWorkerChan:      make(chan struct{}),
		globalStopChan:      make(chan struct{}),
		storage:             storage,
		channelType:         channelType,
		WorkersCount:        0,
		desiredWorkersCount: cfg.WorkersCount,
		workerID:            0,
	}
}

func (rh *RetryHandler) SetMainChannel(ch chan message.Message) {
	rh.mainChannel = ch
}

func (rh *RetryHandler) calculateBackoff(retries int) time.Duration {
	backoff := rh.config.InitialBackoff * time.Duration(1<<retries)
	if backoff > rh.config.MaxBackoff {
		return rh.config.MaxBackoff
	}
	return backoff
}

func (rh *RetryHandler) EnqueueIntoRetryQueue(msg message.Message) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case rh.retryQueue <- msg:
			return nil
		case <-ticker.C:
			// Retry after timeout
		case <-rh.globalStopChan:
			return nil // if shutting down, exit
		}
	}
}

func (rh *RetryHandler) monitorAndScaleWorkers() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rh.globalStopChan:
			return
		case <-ticker.C:
			rh.adjustWorkers()
		}
	}
}

func (rh *RetryHandler) adjustWorkers() {
	queueSize := len(rh.retryQueue)
	queueCapacity := cap(rh.retryQueue)
	loadPercentage := float64(queueSize) / float64(queueCapacity) * 100

	rh.mu.Lock()
	defer rh.mu.Unlock()

	// Smoother scaling (instead of big jumps)
	switch {
	case queueSize >= queueCapacity-1:
		rh.desiredWorkersCount = min(rh.desiredWorkersCount+1, rh.config.WorkersCount*2)
	case loadPercentage > 80:
		rh.desiredWorkersCount = min(rh.desiredWorkersCount+1, rh.config.WorkersCount*2)
	case loadPercentage < 30 && rh.desiredWorkersCount > rh.config.WorkersCount:
		rh.desiredWorkersCount = max(rh.desiredWorkersCount-1, rh.config.WorkersCount)
	}

	rh.syncWorkers()
}

func (rh *RetryHandler) syncWorkers() {
	diff := rh.desiredWorkersCount - rh.WorkersCount

	if diff > 0 {
		log.Printf("[RetryHandler] Scaling up by %d workers", diff)
		for i := 0; i < diff; i++ {
			rh.startWorker()
		}
	} else if diff < 0 {
		log.Printf("[RetryHandler] Scaling down by %d workers", -diff)
		for i := 0; i < -diff; i++ {
			rh.stopWorker()
		}
	}
}

func (rh *RetryHandler) HandleFailedMessage(msg message.Message) error {
	if msg.Retries >= rh.config.MaxRetries {
		return rh.storage.PersistMessage(msg)
	}

	msg.Retries++
	backoff := rh.calculateBackoff(msg.Retries)

	time.AfterFunc(backoff, func() {
		err := rh.EnqueueIntoRetryQueue(msg)
		if err != nil {
			log.Println("[RetryHandler] Failed to re-enqueue message:", err)
		}
	})

	return nil
}

func (rh *RetryHandler) StartRetryManager() {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	for i := 0; i < rh.desiredWorkersCount; i++ {
		rh.startWorker()
	}
	go rh.monitorAndScaleWorkers()
}

func (rh *RetryHandler) startWorker() {
	rh.workerID++
	id := rh.workerID

	rh.WorkersCount++
	rh.wg.Add(1)

	go rh.processRetries(id)
}

func (rh *RetryHandler) stopWorker() {
	select {
	case rh.stopWorkerChan <- struct{}{}:
	default:
	}
}

func (rh *RetryHandler) StopRetryManager() {
	close(rh.globalStopChan)
	rh.wg.Wait()
}

func (rh *RetryHandler) processRetries(id int) {
	defer rh.wg.Done()

	workerName := "worker-" + itoa(id)

	log.Printf("[RetryHandler] %s started", workerName)

	for {
		select {
		case <-rh.globalStopChan:
			log.Printf("[RetryHandler] %s received global stop, exiting", workerName)
			rh.mu.Lock()
			rh.WorkersCount--
			rh.mu.Unlock()
			return
		case <-rh.stopWorkerChan:
			log.Printf("[RetryHandler] %s received scale-down stop, exiting", workerName)
			rh.mu.Lock()
			rh.WorkersCount--
			rh.mu.Unlock()
			return
		case msg := <-rh.retryQueue:
			select {
			case rh.mainChannel <- msg:
				log.Printf("[RetryHandler] %s successfully retried a message", workerName)
			default:
				_ = rh.HandleFailedMessage(msg)
			}
		}
	}
}

func (rh *RetryHandler) RetryQueueSize() int {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	return len(rh.retryQueue)
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Very simple integer to string conversion without fmt
func itoa(i int) string {
	return strconv.Itoa(i)
}
