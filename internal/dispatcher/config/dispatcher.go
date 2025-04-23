package config

import "time"

type DispatcherConfig struct {
	PollInterval time.Duration `yaml:"poll_interval"`
	WorkerCount  int           `yaml:"worker_count"`
}

func DefaultDispatcherConfig() DispatcherConfig {
	return DispatcherConfig{
		PollInterval: 50 * time.Millisecond,
		WorkerCount:  10,
	}
}
