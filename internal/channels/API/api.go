package processor

import (
	"fmt"
	"log"
	"sync"
	"time"

	Channelconfig "learningGolang/internal/config/channels"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
)

type APIProcessor struct {
	APIChannel    chan message.Message
	outputChannel chan message.Message
	config        Channelconfig.APIConfig
	wg            sync.WaitGroup
	stopChan      chan struct{}
	closeOnce     sync.Once

	// New Fields for scaling and monitoring
	workerCount     int
	desiredWorkerCount int
	queueMetrics    metrics.QueueMetrics
	globalStop      chan struct{}
}

func NewAPIProcessor(cfg Channelconfig.APIConfig, outputChan chan message.Message, m metrics.QueueMetrics) (*APIProcessor,error) {
	if cfg.WorkerCount<1{
		return nil,fmt.Errorf("worker count should not be less than 1")
	}
	if outputChan==nil{
		return nil,fmt.Errorf("output channel should not be nil")
	}

	return &APIProcessor{
		APIChannel:       make(chan message.Message, cfg.BufferSize),
		outputChannel:    outputChan,
		config:           cfg,
		stopChan:         make(chan struct{}),
		workerCount:      0,
		desiredWorkerCount: cfg.WorkerCount,
		queueMetrics:     m,
		globalStop:       make(chan struct{}),
	},nil
}

// Monitor and Scale Workers Based on Queue Size
func (p *APIProcessor) monitorAndScaleWorkers() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.globalStop:
			return
		case <-ticker.C:
			p.adjustWorkers()
		}
	}
}

// Adjust Worker Count Based on Queue Load
func (p *APIProcessor) adjustWorkers() {
	queueSize := len(p.APIChannel)        // current queue size
	queueCapacity := cap(p.APIChannel)    // max queue capacity
	loadPercentage := float64(queueSize) / float64(queueCapacity) * 100

	log.Printf("[APIProcessor] Queue load: %.2f%% (current: %d, capacity: %d)", loadPercentage, queueSize, queueCapacity)

	// Scale Workers Based on Queue Load
	if queueSize >= queueCapacity-1 {
		p.desiredWorkerCount = min(p.desiredWorkerCount+3, p.config.WorkerCount*2)
		log.Println("first:",p.desiredWorkerCount)
	} else if loadPercentage > 80 {
		p.desiredWorkerCount = min(p.desiredWorkerCount+2, p.config.WorkerCount*2)
		log.Println("second:",p.desiredWorkerCount)


	} else if loadPercentage < 20 && p.desiredWorkerCount > p.config.WorkerCount {
		p.desiredWorkerCount = max(p.desiredWorkerCount-1, p.config.WorkerCount)
		log.Println("third:",p.desiredWorkerCount)

	}

	// Sync Workers
	p.syncWorkers()
}

// Synchronize Workers by Scaling Up/Down
func (p *APIProcessor) syncWorkers() {
	log.Printf("Synch desired-%d",p.desiredWorkerCount)
	log.Printf("sych worker-%d",p.workerCount)
	diff := p.desiredWorkerCount - p.workerCount
	if diff > 0 {
		log.Printf("[APIProcessor] Scaling up: adding %d workers (current: %d, desired: %d)", diff, p.workerCount, p.desiredWorkerCount)
		for i := 0; i < diff; i++ {
			log.Printf("Starting worker count %d",i)
			p.startWorker()
		}
	} else if diff < 0 {
		log.Printf("[APIProcessor] Scaling down: removing %d workers (current: %d, desired: %d)", -diff, p.workerCount, p.desiredWorkerCount)
		for i := 0; i < -diff; i++ {
			p.stopWorker()

		}
	}else{
		log.Printf("Reduce fk")
	}
}

// Start Worker Goroutines
func (p *APIProcessor) StartProcessingChannel() {
	for i := 0; i < p.config.WorkerCount; i++ {
		p.startWorker()

	}
	go p.monitorAndScaleWorkers() // Start the monitoring and scaling goroutine
}

// Start a Worker
func (p *APIProcessor) startWorker() {
	p.workerCount++
	p.wg.Add(1)
	log.Printf("Inrease count new count %d",p.workerCount)
	go p.processAPIRequest()
}

// Stop a Worker
func (p *APIProcessor) stopWorker() {
	p.workerCount--
	select {
	case p.stopChan <- struct{}{}:
	default:
	}
}

// Process API Request
func (p *APIProcessor) processAPIRequest() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stopChan:
			return
		case msg, ok := <-p.APIChannel:
			if !ok {
				return
			}

			processedMessage, err := p.processMessage(msg)
			if err != nil {
				log.Println("Failed to process message by API Worker:", err)
				continue
			}

			// Send the processed message to outputChannel
			select {
			case p.outputChannel <- processedMessage:
			case <-p.stopChan:
				return
			}
		}
	}
}

// Stop Processing Channel
func (p *APIProcessor) StopProcessingChannel() {
	p.closeOnce.Do(func() {
		close(p.stopChan)
		p.wg.Wait()
		close(p.APIChannel)
	})
}

// Process API Request Message (Mocked API Call)
func (p *APIProcessor) processMessage(message message.Message) (message.Message, error) {
	if len(message.Payload) == 0 {
		return message, fmt.Errorf("message payload should not be empty")
	}

	// Simulate actual API call (e.g., HTTP request)
	time.Sleep(50 * time.Millisecond)

	// Add metadata to the processed message
	message.MetaData["processed_at"] = time.Now().UTC()
	message.MetaData["Processed_By"] = "Processed By API Worker"

	// Update metrics for processed message
	p.queueMetrics.MessagesProcessed.Inc()
	return message, nil
}

// Returns the channel in which the dispatcher puts the message load.
func (p *APIProcessor) GetInputChannel() chan message.Message {
	return p.APIChannel
}
func(P * APIProcessor)GetCurrentWorkerCount()int{
	return P.workerCount
}
func(P * APIProcessor)GetDesiredWorkerCount() int{
	return P.desiredWorkerCount
}

// Helper functions for min and max
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
