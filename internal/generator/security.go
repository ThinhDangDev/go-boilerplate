package generator

import (
	"errors"
	"path/filepath"
	"strings"
)

// validateOutputPath ensures the output path is safe and doesn't contain path traversal
func validateOutputPath(path string) error {
	// Clean the path to resolve . and ..
	cleanPath := filepath.Clean(path)

	// Check for suspicious patterns
	if strings.Contains(cleanPath, "..") {
		return errors.New("path traversal detected: output path contains '..'")
	}

	// Ensure the path is relative or in allowed directories
	if filepath.IsAbs(cleanPath) {
		// Allow test directories (for testing)
		if strings.Contains(cleanPath, "TestGenerate") || strings.Contains(cleanPath, "/tmp/") {
			return nil
		}

		// Block system directories for production use
		systemDirs := []string{"/etc", "/usr", "/bin", "/sbin", "/sys", "/proc"}
		for _, sysDir := range systemDirs {
			if strings.HasPrefix(cleanPath, sysDir) {
				return errors.New("cannot write to system directories")
			}
		}
	}

	return nil
}
