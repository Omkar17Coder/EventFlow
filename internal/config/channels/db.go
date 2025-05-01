package config

import "time"

type DBConfig struct {
	BufferSize  int           `yaml:"buffer_size"`
	WorkerCount int           `yaml:"worker_count"`
	Timeout     time.Duration `yaml:"timeout"`
	DSN         string        `yaml:"dsn"`
}

func DefaultDBConfig() DBConfig {
	return DBConfig{
		BufferSize:  200,
		WorkerCount: 2,
		Timeout:     3 * time.Second,
		DSN:         "user:pass@tcp(localhost:3306)/db",
	}
}
