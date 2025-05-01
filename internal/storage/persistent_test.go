package storage_test

import (
	"encoding/json"
	"fmt"
	"learningGolang/internal/config"
	"learningGolang/internal/storage"
	"learningGolang/pkg/message"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStorage(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := config.StorageConfig{
			FolderPath:         tempDir,
			FailedMessagesPath: "testfile.json",
			MaxFileSize:        1024,
			RotationInterval:   2,
		}
		storage, err := storage.NewFileStorage(cfg)
		require.NoError(t, err)
		defer storage.Close()
		assert.NotNil(t, storage)
		assert.Equal(t, filepath.Join(tempDir, "testfile.json"), storage.GetFilePath())
		_, MaxSize := storage.SizeDetials()
		assert.Equal(t, 1024, MaxSize)
		_, errorFileInfo := os.Stat(storage.GetFilePath())
		assert.NoError(t, errorFileInfo)

	})

	t.Run("invalid max file size", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := config.StorageConfig{
			FolderPath:         tempDir,
			FailedMessagesPath: "testfile.json",
			MaxFileSize:        0,
			RotationInterval:   2,
		}
		storage, err := storage.NewFileStorage(cfg)
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.Equal(t, err.Error(), "File Size should not be less than 1")

	})

	t.Run("invalid file path", func(t *testing.T) {
		cfg := config.StorageConfig{
			FailedMessagesPath: "../../invalid/path/testfile.json",
			MaxFileSize:        1024,
		}

		storage, err := storage.NewFileStorage(cfg)
		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.Contains(t, err.Error(), "Error file validitiy")
	})
}

func TestPersistMessage(t *testing.T) {
	t.Run("successful message persistence", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := config.StorageConfig{
			FolderPath:         tempDir,
			FailedMessagesPath: "test_persitence.json",
			MaxFileSize:        1024,
		}
		storage, err := storage.NewFileStorage(cfg)
		require.NoError(t, err)
		defer storage.Close()
		messagePayload := map[string]interface{}{"Inserted_At": "12:00", "Message": "Test Log Message"}
		TestMessage, err := message.NewMessage("001", "API", messagePayload)
		assert.NoError(t, err)
		err = storage.PersistMessage(TestMessage)
		assert.NoError(t, err)

		fileContents, err := os.ReadFile(storage.GetFilePath())
		require.NoError(t, err)

		var decodedMessage map[string]interface{}
		err = json.Unmarshal(fileContents, &decodedMessage)
		assert.NoError(t, err)
		assert.Equal(t, decodedMessage["payload"], messagePayload)

	})
	t.Run("exceeds max file size", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := config.StorageConfig{
			FolderPath:         tempDir,
			FailedMessagesPath: "test_persitence.json",
			MaxFileSize:        2,
		}
		storage, err := storage.NewFileStorage(cfg)
		require.NoError(t, err)
		defer storage.Close()
		messagePayload := map[string]interface{}{"Inserted_At": "12:00", "Message": "Test Log Message"}
		TestMessage, err := message.NewMessage("001", "API", messagePayload)
		assert.NoError(t, err)
		err = storage.PersistMessage(TestMessage)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "max file size exceeded")

	})

	t.Run("invalid message fails to marshal", func(t *testing.T) {
		tempDir := t.TempDir()
		cfg := config.StorageConfig{
			FolderPath:         tempDir,
			FailedMessagesPath: "test_persitence.json",
			MaxFileSize:        1024,
		}
		storage, err := storage.NewFileStorage(cfg)
		require.NoError(t, err)
		defer storage.Close()
		messagePayload := make(chan int)
		err = storage.PersistMessage(messagePayload)
		assert.Error(t, err)
		fmt.Println(err.Error())
		assert.True(t, strings.Contains(err.Error(), "failed to marshal"))

	})

	
}
