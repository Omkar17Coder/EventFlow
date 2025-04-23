package config

import "time"

type Config struct {
	Queue      QueueConfig
	Dispatcher DispatcherConfig
}




func DefaultConfig() Config {
	return Config{
		Queue: QueueConfig{
			MaxSize:          1000,
			InitialBackoff:   10 * time.Millisecond,
			MaxBackoff:       1 * time.Second,
			BackoffFactor:    1.5,
			MaxRetryAttempts: 5,
		},
		Dispatcher: DispatcherConfig{
			PollInterval: 50 * time.Millisecond,
			WorkerCount:  10,
		},
	}
}
