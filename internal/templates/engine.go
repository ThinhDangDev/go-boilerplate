package templates

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/yourusername/go-boilerplate/internal/config"
)

// Engine handles template rendering
type Engine struct {
	templatesDir string
}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	// Try multiple paths to find the templates directory
	possiblePaths := []string{
		"templates",                    // From project root
		"../../templates",              // From internal/generator (tests)
		"../../../templates",           // From internal/generator/subdir
		filepath.Join(".", "templates"), // Current directory
	}

	// Get the executable directory as fallback
	execPath, err := os.Executable()
	if err == nil {
		baseDir := filepath.Dir(execPath)
		possiblePaths = append(possiblePaths, filepath.Join(baseDir, "templates"))
	}

	// Find the first existing templates directory
	var templatesDir string
	for _, path := range possiblePaths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			// Verify it contains actual templates
			if hasTemplates(absPath) {
				templatesDir = absPath
				break
			}
		}
	}

	// If no templates directory found, use relative path as fallback
	if templatesDir == "" {
		templatesDir = "templates"
	}

	return &Engine{
		templatesDir: templatesDir,
	}
}

// hasTemplates checks if directory contains template files
func hasTemplates(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "base"))
	return err == nil
}

// Render renders a template with the given data
func (e *Engine) Render(templateName string, data *config.Config) ([]byte, error) {
	// Validate template path to prevent path traversal
	if err := validateTemplatePath(e.templatesDir, templateName); err != nil {
		return nil, fmt.Errorf("invalid template path: %w", err)
	}

	templatePath := filepath.Join(e.templatesDir, templateName)

	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("template %s not found at %s", templateName, templatePath)
	}

	// Read template file
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New(filepath.Base(templateName)).Funcs(customFuncs()).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// ListTemplates returns a list of all available templates
func (e *Engine) ListTemplates(pattern string) ([]string, error) {
	var templates []string

	err := filepath.WalkDir(e.templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			relPath, err := filepath.Rel(e.templatesDir, path)
			if err != nil {
				return err
			}
			templates = append(templates, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}

	return templates, nil
}

// customFuncs returns custom template functions
func customFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower":   strings.ToLower,
		"toUpper":   strings.ToUpper,
		"replace":   strings.ReplaceAll,
		"trimSpace": strings.TrimSpace,
		"join":      strings.Join,
	}
}
