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
	// Get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current directory
		execPath = "."
	}
	baseDir := filepath.Dir(execPath)
	templatesDir := filepath.Join(baseDir, "templates")

	// Check if templates directory exists, if not use relative path
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		templatesDir = "templates"
	}

	return &Engine{
		templatesDir: templatesDir,
	}
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
