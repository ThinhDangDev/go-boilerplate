package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileWriter handles atomic file writing operations
type FileWriter struct{}

// NewFileWriter creates a new FileWriter instance
func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

// WriteFile writes content to a file atomically using a temp file + rename pattern
func (fw *FileWriter) WriteFile(path string, data []byte) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create temporary file in the same directory
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // Clean up in case of error

	// Write content to temp file
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Close temp file
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set file permissions
	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomically rename temp file to target file
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
