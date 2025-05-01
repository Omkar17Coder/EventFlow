package config

import (
	"os"
	"time"
)

type StorageConfig struct {
	FolderPath         string        `yaml:"folder_path"`
	FailedMessagesPath string        `yaml:"failed_messages_path"`
	MaxFileSize        int64         `yaml:"max_file_size"`
	File               *os.File      `yaml:"file"`
	RotationInterval   time.Duration `yaml:"rotation_interval"`
}

func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		FolderPath:         "D:/AppData/YourApp/FailedMessages",
		FailedMessagesPath: "./failed_messages",
		MaxFileSize:        100 * 1024 * 1024, // 100MB
		RotationInterval:   24 * time.Hour,
	}
}
