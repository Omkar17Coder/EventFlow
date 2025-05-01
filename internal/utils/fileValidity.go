package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"
)

var invalidWindowsChars = []rune{'<', '>', ':', '"', '/', '\\', '|', '?', '*'}

func EnsureFolder(hardcodedFolder string) error {
	absPath, err := filepath.Abs(filepath.Clean(hardcodedFolder))
	if err != nil {
		return fmt.Errorf("failed to clean folder")
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(absPath, 0755); err != nil {
				return fmt.Errorf("Failed to create folder")
			}
		} else {
			return fmt.Errorf("Error accesssing folder")
		}
	} else if !info.IsDir() {
		return fmt.Errorf("This is not a directory")
	}
	return nil
}

// Check if a name (file or folder) contains invalid Windows chars
func hasInvalidWindowsChars(name string) bool {
	for _, ch := range name {
		for _, invalid := range invalidWindowsChars {
			if ch == invalid {
				return true
			}
		}
	}
	return false
}
func ValidateFilePath(pathElements []string) error {
	for _, elem := range pathElements {
		if elem == "" || elem == "." || elem == ".." {
			return fmt.Errorf("invalid path element")
		}
		if hasInvalidWindowsChars(elem) {
			return fmt.Errorf("path element %q contains invalid characters", elem)
		}
		if utf8.RuneCountInString(elem) > 255 {
			return fmt.Errorf("path element %q is too long (max 255 chars)", elem)
		}
	}
	return nil
}
