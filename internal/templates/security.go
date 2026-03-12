package templates

import (
	"errors"
	"path/filepath"
	"strings"
)

// validateTemplatePath ensures the template path is within the templates directory
func validateTemplatePath(templatesDir, templateName string) error {
	// Clean paths
	cleanTemplateName := filepath.Clean(templateName)

	// Check for path traversal attempts
	if strings.Contains(cleanTemplateName, "..") {
		return errors.New("path traversal detected in template name")
	}

	// Resolve absolute paths
	templatePath := filepath.Join(templatesDir, cleanTemplateName)
	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return err
	}

	absTemplatesDir, err := filepath.Abs(templatesDir)
	if err != nil {
		return err
	}

	// Ensure the resolved path is within templates directory
	if !strings.HasPrefix(absTemplatePath, absTemplatesDir+string(filepath.Separator)) &&
		absTemplatePath != absTemplatesDir {
		return errors.New("template path is outside templates directory")
	}

	return nil
}
