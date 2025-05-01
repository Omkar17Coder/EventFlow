

package processor_test

import (
	"context"
	
	processor "learningGolang/internal/channels/API"
	ChannelConfig "learningGolang/internal/config/channels"
	"learningGolang/metrics"
	"learningGolang/pkg/message"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert")

func TestNewApiProcessor(t *testing.T) {
	t.Run("sucessful api processor creation", func(t *testing.T) {
		testApiConfig := ChannelConfig.APIConfig{
			BufferSize:  5,
			WorkerCount: 1,
			Timeout:     1 * time.Second,
			EndPoint:    "https://100xdev.api.com",
		}

		// next output channel to put message into
		OutputChannel := make(chan message.Message)
		NewMockApiProcessor, err := processor.NewAPIProcessor(testApiConfig, OutputChannel, *metrics.NewQueueMetrics())
		assert.NoError(t, err)
		assert.NotNil(t, NewMockApiProcessor)

	})
	t.Run("nil output channel throws error", func(t *testing.T) {
		testApiConfig := ChannelConfig.APIConfig{
			BufferSize:  5,
			WorkerCount: 1,
			Timeout:     1 * time.Second,
			EndPoint:    "https://100xdev.api.com",
		}

		NewMockApiProcessor, err := processor.NewAPIProcessor(testApiConfig, nil, *metrics.NewQueueMetrics())
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "output channel should not be nil")
		assert.Nil(t, NewMockApiProcessor)
	})
	t.Run("worker count less than 1 throws error", func(t *testing.T) {
		testApiConfig := ChannelConfig.APIConfig{
			BufferSize:  5,
			WorkerCount: 0,
			Timeout:     1 * time.Second,
			EndPoint:    "https://100xdev.api.com",
		}
		NewMockApiProcessor, err := processor.NewAPIProcessor(testApiConfig, make(chan message.Message,100), *metrics.NewQueueMetrics())
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "worker count should not be less than 1")
		assert.Nil(t, NewMockApiProcessor)
	})

}


func floodQueue(ctx context.Context, ch chan message.Message, rate time.Duration, workers int, wg *sync.WaitGroup) {
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Printf("Iam going ")
					return
				default:
					msgPtr, err := message.NewMessage("002", "API", "Hello")
					if err != nil {
						log.Printf("error creating message: %v", err)
						continue
					}

					select {
					case ch <- *msgPtr:

					default:

					}
					time.Sleep(rate)

				}
			}
		}()
	}
}

func TestAPIProcessorCoreFunction(t *testing.T) {
	t.Run("start processing count and validate worker count", func(t *testing.T) {
		testApiConfig := ChannelConfig.APIConfig{
			BufferSize:  5,
			WorkerCount: 5,
			Timeout:     1 * time.Second,
			EndPoint:    "https://100xdev.api.com",
		}

		// next output channel to put message into
		OutputChannel := make(chan message.Message)
		MockApiProcessor, err := processor.NewAPIProcessor(testApiConfig, OutputChannel, *metrics.NewQueueMetrics())
		assert.NoError(t, err)
		assert.NotNil(t, MockApiProcessor)
		MockApiProcessor.StartProcessingChannel()
		currentWorkerCount := MockApiProcessor.GetCurrentWorkerCount()
		assert.Equal(t, currentWorkerCount, testApiConfig.WorkerCount)
		MockApiProcessor.StopProcessingChannel()
	})
	t.Run("Test scale-up under sustained load", func(t *testing.T) {
		testConfig := ChannelConfig.APIConfig{
			BufferSize:  10,
			WorkerCount: 2,
			Timeout:     1 * time.Second,
			EndPoint:    "https://100xdev.api.com",
		}
	
		outputCh := make(chan message.Message)
		queueMetrics := metrics.NewQueueMetrics()
		apiProcessor, err := processor.NewAPIProcessor(testConfig, outputCh, *queueMetrics)
	
		assert.NoError(t, err)
		assert.NotNil(t, apiProcessor)
	
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
	
		var wg sync.WaitGroup
	
		
		floodQueue(ctx, apiProcessor.GetInputChannel(), 3*time.Millisecond, 2, &wg)
	
		
		apiProcessor.StartProcessingChannel()
	
		
		time.Sleep(4 * time.Second)
	
		
		currentWorkers := apiProcessor.GetCurrentWorkerCount()
		t.Logf("Workers after load: %d", currentWorkers)
	
		assert.Greater(t, currentWorkers, testConfig.WorkerCount, "Expected worker count to increase under sustained load")
	
		cancel()
		wg.Wait()
		apiProcessor.StopProcessingChannel()
	})

	t.Run("Test scale-up and scale-down under varying load", func(t *testing.T) {
		testConfig := ChannelConfig.APIConfig{
			BufferSize:  10,
			WorkerCount: 2,
			Timeout:     1 * time.Second,
			EndPoint:    "https://100xdev.api.com",
		}

		outputCh := make(chan message.Message,10000)
		queueMetrics := metrics.NewQueueMetrics()

		apiProcessor, err := processor.NewAPIProcessor(testConfig, outputCh, *queueMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, apiProcessor)

		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		// Start flooding queue with fast message rate and multiple workers
		floodQueue(ctx, apiProcessor.GetInputChannel(), 2*time.Millisecond, 3, &wg)

		// Start processor
		apiProcessor.StartProcessingChannel()

		// Let autoscaler detect and scale up
		time.Sleep(4 * time.Second)
		

		currentWorkers := apiProcessor.GetCurrentWorkerCount()
		assert.Greater(t,currentWorkers,testConfig.WorkerCount,"When Load increases workers should be scalled")
		t.Logf("[Load Phase] Workers after load: %d", currentWorkers)
		
		cancel()
		wg.Wait()

		t.Log("[Cooldown Phase] Flood stopped. Waiting for scale down...")


		// Wait long enough for scale-down (based on your autoscaler logic, e.g., 12-15 seconds)
		time.Sleep(8 * time.Second)

		currentWorkers = apiProcessor.GetDesiredWorkerCount()
		t.Logf("[Cooldown Phase] Final desired workers: %d", currentWorkers)
		t.Log("Outpitload",len(outputCh))
		//Scale  back to inital load.
		assert.Equal(t, currentWorkers, testConfig.WorkerCount ,"Expected processor to scale down to min workers count")
		apiProcessor.StopProcessingChannel()
	})

	
}

