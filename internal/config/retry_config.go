package config

import "time"

type RetryConfig struct {
	MaxRetries     int           `yaml:"max_retries"`
	InitialBackoff time.Duration `yaml:"inital_backoff"`
	MaxBackoff     time.Duration `yaml:"max_backoff"`
	RetryQueueSize int           `yaml:"retry_queue_size"`
	WorkersCount   int           `yaml:"workers_count"`
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
		RetryQueueSize: 30,
		WorkersCount: 2,
	}
}
