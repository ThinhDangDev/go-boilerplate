package templates

import (
	"errors"
	"path/filepath"
	"strings"
)

// validateTemplatePath ensures the template path doesn't contain path traversal attempts
func validateTemplatePath(templateName string) error {
	// Clean paths
	cleanTemplateName := filepath.Clean(templateName)

	// Check for path traversal attempts
	if strings.Contains(cleanTemplateName, "..") {
		return errors.New("path traversal detected in template name")
	}

	// Check for absolute paths
	if filepath.IsAbs(cleanTemplateName) {
		return errors.New("absolute paths not allowed in template names")
	}

	return nil
}
