package config

import Channels "learningGolang/internal/config/channels"

type AppConfig struct {
	Queue      QueueConfig             `yaml:"queue"`
	Dispatcher DispatcherConfig        `yaml:"dispatcher"`
	Storage    StorageConfig           `yaml:"storage"`
	Retry      RetryConfig             `yaml:"retry"`
	Channels   Channels.ChannelsConfig `yaml:"channels"`
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Queue:      DefaultQueueConfig(),
		Dispatcher: DefaultDispatcherConfig(),
		Storage:    DefaultStorageConfig(),
		Retry:      DefaultRetryConfig(),
		Channels:   Channels.DefaultChannelsConfig(),
	}
}
