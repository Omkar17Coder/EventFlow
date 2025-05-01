package api

import (
	"encoding/json"
	"errors"
	"learningGolang/internal/queue"
	"learningGolang/pkg/message"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	
	MaxBackoffDuration = 30 * time.Second
	BackoffFactor = 1.5 
	MaxRetryAttempts = 10 
	InitialBackoff = 1 * time.Second
)

type Gateway struct {
	queue       *queue.Queue
	server      *http.Server
	mu          sync.RWMutex
	workerCount int
	maxWorkers  int
}

type MessageRequest struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewGateway(q *queue.Queue, port string, maxWorkers int) *Gateway {
	mux := http.NewServeMux()

	gw := &Gateway{
		queue:       q,
		workerCount: 1,
		maxWorkers:  maxWorkers,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		},
	}

	mux.HandleFunc("/api/messages", gw.handleMessages)
	mux.HandleFunc("/api/health", gw.handleHealth)
	mux.HandleFunc("/api/metrics", gw.handleMetrics)

	return gw
}

func (gw *Gateway) Start() error {
	go gw.monitorQueueLoad()
	return gw.server.ListenAndServe()
}

func (gw *Gateway) monitorQueueLoad() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gw.mu.Lock()

		queueSize := gw.queue.Size()
		queueCapacity := cap(gw.queue.MessagesList)
		loadPercentage := float64(queueSize) / float64(queueCapacity) * 100

		switch {
		case queueSize >= queueCapacity-1 && gw.workerCount < gw.maxWorkers:
			workersToAdd := min(3, gw.maxWorkers-gw.workerCount)
			gw.workerCount += workersToAdd
			log.Printf("[Gateway] Queue almost full! Adding %d workers (total: %d)", workersToAdd, gw.workerCount)

		case loadPercentage > 80 && gw.workerCount < gw.maxWorkers:
			workersToAdd := min(2, gw.maxWorkers-gw.workerCount)
			gw.workerCount += workersToAdd
			log.Printf("[Gateway] High load (%.2f%%)! Adding %d workers (total: %d)", loadPercentage, workersToAdd, gw.workerCount)

		case loadPercentage < 20 && gw.workerCount > 1:
			gw.workerCount--
			log.Printf("[Gateway] Low load (%.2f%%)! Reducing workers to %d", loadPercentage, gw.workerCount)
		}

		gw.mu.Unlock()
	}
}

func (gw *Gateway) handleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	msg, err := message.NewMessage("", req.Type, req.Payload)
	if err != nil {
		http.Error(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	// Retry logic for enqueueing the message with exponential backoff
	err = gw.retryEnqueue(msg)
	if err != nil {
		http.Error(w, "Failed to enqueue message", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "message accepted"})
}

func (gw *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (gw *Gateway) handleMetrics(w http.ResponseWriter, r *http.Request) {
	gw.mu.RLock()
	defer gw.mu.RUnlock()

	metrics := map[string]interface{}{
		"queue_size":     gw.queue.Size(),
		"queue_capacity": cap(gw.queue.MessagesList),
		"worker_count":   gw.workerCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// retryEnqueue handles retries with exponential backoff for the enqueue operation
func (gw *Gateway) retryEnqueue(msg *message.Message) error {
	backoff := InitialBackoff

	// Retry logic with exponential backoff and jitter
	for attempt := 1; attempt <= MaxRetryAttempts; attempt++ {
		err := gw.queue.Enqueue(msg)
		if err == nil {
			return nil
		}
		if err == queue.ErrQueueClosed {
			return err
		}

		// Add jitter: randomize the backoff to avoid synchronized retries
		jitter := time.Duration(rand.Int63n(int64(backoff))) // Randomize the delay up to `backoff` time
		backoffWithJitter := backoff + jitter

		// Apply max backoff limit
		if backoffWithJitter > MaxBackoffDuration {
			backoffWithJitter = MaxBackoffDuration
		}

		// Log retry attempt
		log.Printf("Retrying enqueue operation (attempt %d), backoff: %v", attempt, backoffWithJitter)

		// Wait before retrying
		time.Sleep(backoffWithJitter)

		// Exponential backoff with factor
		backoff = time.Duration(float64(backoff) * BackoffFactor)

		// Don't allow backoff to exceed max limit
		if backoff > MaxBackoffDuration {
			backoff = MaxBackoffDuration
		}
	}

	return errors.New("max retries exhausted")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
