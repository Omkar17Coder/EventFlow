package config

import "time"

type DispatcherConfig struct {
	PollInterval              time.Duration `yaml:"poll_interval"`
	WorkerCount               int           `yaml:"worker_count"`
	ProcessorRetryCount       int           `yaml:"processor_retry_count"`
	ProcessRetryMinDuration   time.Duration `yaml:"processor_retry_min_duration"`
	ProcessorRetryMaxDuration time.Duration `yaml:"processor_retry_max_duration"`
}

func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		PollInterval:              20 * time.Millisecond,
		WorkerCount:               20,
		ProcessorRetryCount:       3,
		ProcessRetryMinDuration:   50 * time.Millisecond,
		ProcessorRetryMaxDuration: 200 * time.Millisecond,
	}
}
