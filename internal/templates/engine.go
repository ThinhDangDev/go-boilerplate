package templates

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ThinhDangDev/go-boilerplate/internal/config"
	embeddedtemplates "github.com/ThinhDangDev/go-boilerplate/templates"
)

// Engine handles template rendering
type Engine struct {
	fs fs.FS
}

// NewEngine creates a new template engine using embedded templates
func NewEngine() *Engine {
	return &Engine{
		fs: embeddedtemplates.Templates,
	}
}

// Render renders a template with the given data
func (e *Engine) Render(templateName string, data *config.Config) ([]byte, error) {
	// Validate template path to prevent path traversal
	if err := validateTemplatePath(templateName); err != nil {
		return nil, fmt.Errorf("invalid template path: %w", err)
	}

	// Read template content from embedded FS
	content, err := fs.ReadFile(e.fs, templateName)
	if err != nil {
		return nil, fmt.Errorf("template %s not found in embedded files: %w", templateName, err)
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

	err := fs.WalkDir(e.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			templates = append(templates, path)
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
