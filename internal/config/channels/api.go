package config

import "time"

type APIConfig struct {
	BufferSize  int           `yaml:"buffer_size"`
	WorkerCount int           `yaml:"worker_count"`
	Timeout     time.Duration `yaml:"timeout"`
	EndPoint    string        `yaml:"endpoint"`
	// the end point has no role in this project,
}

func DefaultAPIConfig() APIConfig {
	return APIConfig{
		BufferSize:  20,
		WorkerCount: 3,
		Timeout:     5 * time.Second,
		EndPoint:    "https://100xdev.api.com",
	}
}
