package config

import "time"

type EmailConfig struct {
	BufferSize  int           `yaml:"buffer_size"`
	WorkerCount int           `yaml:"worker_count"`
	Timeout     time.Duration `yaml:"timeout"`
	SMTPHost    string        `yaml:"smtp_host"`
	SMTPPort    int           `yaml:"smtp_port"`
}

func DefaultEmailConfig() EmailConfig {
	return EmailConfig{
		BufferSize:  50,
		WorkerCount: 2,
		Timeout:     10 * time.Second,
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
	}
}
