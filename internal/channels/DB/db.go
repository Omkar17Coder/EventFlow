package db

import (
	"fmt"
	"log"
	"sync"
	"time"

	channelsConfig "learningGolang/internal/config/channels"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
)

type DBProcessor struct {
	DBChannel         chan message.Message
	OutputChannel     chan message.Message
	config            channelsConfig.DBConfig
	wg                sync.WaitGroup
	stopChan          chan struct{}
	closeOnce         sync.Once
	workerCount       int
	desiredWorkerCount int
	queueMetrics      metrics.QueueMetrics
	globalStop        chan struct{}
}

func NewDBProcessor(cfg channelsConfig.DBConfig, outputChan chan message.Message, m metrics.QueueMetrics) *DBProcessor {
	return &DBProcessor{
		DBChannel:         make(chan message.Message, cfg.BufferSize),
		OutputChannel:     outputChan,
		config:            cfg,
		stopChan:          make(chan struct{}),
		workerCount:       0,
		desiredWorkerCount: cfg.WorkerCount,
		queueMetrics:      m,
		globalStop:        make(chan struct{}),
	}
}

func (p *DBProcessor) monitorAndScaleWorkers() {
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

func (p *DBProcessor) adjustWorkers() {
	queueSize := len(p.DBChannel)
	queueCapacity := cap(p.DBChannel)
	loadPercentage := float64(queueSize) / float64(queueCapacity) * 100

	log.Printf("[DBProcessor] Queue load: %.2f%% (current: %d, capacity: %d)", loadPercentage, queueSize, queueCapacity)

	if queueSize >= queueCapacity-1 {
		p.desiredWorkerCount = min(p.desiredWorkerCount+3, p.config.WorkerCount*2)
	} else if loadPercentage > 80 {
		p.desiredWorkerCount = min(p.desiredWorkerCount+2, p.config.WorkerCount*2)
	} else if loadPercentage < 20 && p.desiredWorkerCount > p.config.WorkerCount {
		p.desiredWorkerCount = max(p.desiredWorkerCount-1, p.config.WorkerCount)
	}

	p.syncWorkers()
}

func (p *DBProcessor) syncWorkers() {
	diff := p.desiredWorkerCount - p.workerCount
	if diff > 0 {
		log.Printf("[DBProcessor] Scaling up: adding %d workers (current: %d, desired: %d)", diff, p.workerCount, p.desiredWorkerCount)
		for i := 0; i < diff; i++ {
			p.startWorker()
		}
	} else if diff < 0 {
		log.Printf("[DBProcessor] Scaling down: removing %d workers (current: %d, desired: %d)", -diff, p.workerCount, p.desiredWorkerCount)
		for i := 0; i < -diff; i++ {
			p.stopWorker()
		}
	}
}

func (p *DBProcessor) startWorker() {
	p.workerCount++
	p.wg.Add(1)
	go p.processDBRequest()
}

func (p *DBProcessor) stopWorker() {
	p.workerCount--
	select {
	case p.stopChan <- struct{}{}:
	default:
	}
}

func (db *DBProcessor) StartProcessingChannel() {
	for i := 0; i < db.config.WorkerCount; i++ {
		db.wg.Add(1)
		go db.processDBRequest()
	}
}

func (db *DBProcessor) processDBRequest() {
	defer db.wg.Done()
	for {
		select {
		case <-db.stopChan:
			return
		case message, ok := <-db.DBChannel:
			if !ok {
				return
			}
			processedMessage, err := processMessage(message)
			if err != nil {
				log.Println("failed to process message by db worker")
				continue
			}
			select {
			case db.OutputChannel <- processedMessage:
			case <-db.stopChan:
				return
			}
		}
	}
}

func (db *DBProcessor) StopProcessingChannel() {
	db.closeOnce.Do(func() {
		close(db.stopChan)
		db.wg.Wait()
		close(db.DBChannel)
	})
}

func processMessage(message message.Message) (message.Message, error) {
	if len(message.Payload) == 0 {
		return message, fmt.Errorf("message payload is empty")
	}
	time.Sleep(60 * time.Millisecond)
	message.MetaData["Processed_At"] = time.Now().UTC()
	message.MetaData["Processes_By"] = "Processed by DB Worker"
	return message, nil
}

func (p *DBProcessor) GetInputChannel() chan message.Message {
	return p.DBChannel
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
