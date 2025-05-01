package storage

import (
	"encoding/json"
	"fmt"
	"learningGolang/internal/config"
	"learningGolang/internal/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileStorage struct {
	filePath    string
	file        *os.File
	encoder     json.Encoder
	mu          sync.Mutex
	currentSize int
	maxSize     int
	rotationCh  chan struct{}
}

// invalid file checks

func NewFileStorage(StorageConfig config.StorageConfig) (*FileStorage, error) {
	if StorageConfig.MaxFileSize <= 0 {
		return nil, fmt.Errorf("File Size should not be less than 1")
	}
	if err := utils.EnsureFolder(StorageConfig.FolderPath); err != nil {
		return nil, err
	}
	cleanedRelPath := filepath.Clean(StorageConfig.FailedMessagesPath)
	pathElements := strings.Split(cleanedRelPath, string(os.PathSeparator))

	if err := utils.ValidateFilePath(pathElements); err != nil {
		return nil, fmt.Errorf("Error file validitiy %s", err)
	}
	fullPath := filepath.Join(StorageConfig.FolderPath, cleanedRelPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %q:%w", dir, err)
	}

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file")
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get the stat of file")
	}
	fs := FileStorage{
		filePath:    fullPath,
		file:        file,
		encoder:     *json.NewEncoder(file),
		currentSize: int(info.Size()),
		maxSize:     int(StorageConfig.MaxFileSize),
		rotationCh:  make(chan struct{}, 1),
	}
	return &fs, nil
}

func (fs *FileStorage) PersistMessage(msg interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal:%w", err)
	}
	data = append(data, '\n')
	if fs.currentSize+int(len(data)) > fs.maxSize {
		return fmt.Errorf("max file size exceeded")
	}
	n, err := fs.file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to the fie %w", err)
	}
	fs.currentSize += int(n)
	return nil

}
func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.file.Close()

}

func(fs *FileStorage)GetFilePath() string{
	return fs.filePath
}
func(fs * FileStorage)SizeDetials()(int,int ){
	return fs.currentSize,fs.maxSize
}