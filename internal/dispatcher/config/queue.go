
package config

import "time"

type QueueConfig struct {
	MaxSize          int           `yaml:"max_size"`
	InitialBackoff   time.Duration `yaml:"initial_backoff"`
	MaxBackoff       time.Duration `yaml:"max_backoff"`
	BackoffFactor    float64       `yaml:"backoff_factor"`
	MaxRetryAttempts int           `yaml:"max_retry_attempts"`
}

//*
func DefaultQueueConfig() QueueConfig {
	return QueueConfig{
		MaxSize:          1000,
		InitialBackoff:   10 * time.Millisecond,
		MaxBackoff:       1 * time.Second,
		BackoffFactor:    1.5,
		MaxRetryAttempts: 5,
	}
}

