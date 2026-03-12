package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/go-boilerplate/internal/config"
)

func TestEngine_Render(t *testing.T) {
	// Create a temporary templates directory for testing
	tmpDir := t.TempDir()

	// Create test templates
	testTemplates := map[string]string{
		"simple.tmpl": "Hello {{.ProjectName}}!",
		"complex.tmpl": `Project: {{.ProjectName}}
Module: {{.ModuleName}}
{{if .Features.Docker}}Docker enabled{{end}}`,
		"functions.tmpl": "{{toLower .ProjectName}} - {{toUpper .ModuleName}}",
	}

	for name, content := range testTemplates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test template: %v", err)
		}
	}

	engine := &Engine{templatesDir: tmpDir}

	tests := []struct {
		name         string
		templateName string
		config       *config.Config
		want         string
		wantErr      bool
	}{
		{
			name:         "simple template",
			templateName: "simple.tmpl",
			config: &config.Config{
				ProjectName: "test-project",
				ModuleName:  "github.com/user/test-project",
			},
			want:    "Hello test-project!",
			wantErr: false,
		},
		{
			name:         "complex template with features",
			templateName: "complex.tmpl",
			config: &config.Config{
				ProjectName: "my-app",
				ModuleName:  "example.com/my-app",
				Features: config.Features{
					Docker: true,
				},
			},
			want: `Project: my-app
Module: example.com/my-app
Docker enabled`,
			wantErr: false,
		},
		{
			name:         "template with custom functions",
			templateName: "functions.tmpl",
			config: &config.Config{
				ProjectName: "MyProject",
				ModuleName:  "github.com/user/project",
			},
			want:    "myproject - GITHUB.COM/USER/PROJECT",
			wantErr: false,
		},
		{
			name:         "non-existent template",
			templateName: "nonexistent.tmpl",
			config: &config.Config{
				ProjectName: "test",
				ModuleName:  "test",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.templateName, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Engine.Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && strings.TrimSpace(string(got)) != strings.TrimSpace(tt.want) {
				t.Errorf("Engine.Render() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestEngine_ListTemplates(t *testing.T) {
	// Create a temporary templates directory
	tmpDir := t.TempDir()

	// Create test template structure
	templates := []string{
		"base/main.go.tmpl",
		"base/config.go.tmpl",
		"features/auth/jwt.go.tmpl",
		"features/docker/Dockerfile.tmpl",
	}

	for _, tmpl := range templates {
		path := filepath.Join(tmpDir, tmpl)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create template: %v", err)
		}
	}

	// Create non-template files (should be ignored)
	nonTemplates := []string{
		"README.md",
		"test.go",
	}
	for _, f := range nonTemplates {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	engine := &Engine{templatesDir: tmpDir}

	t.Run("list all templates", func(t *testing.T) {
		got, err := engine.ListTemplates("")
		if err != nil {
			t.Fatalf("Engine.ListTemplates() error = %v", err)
		}

		// Verify count
		if len(got) != len(templates) {
			t.Errorf("Engine.ListTemplates() returned %d templates, want %d", len(got), len(templates))
		}

		// Verify all templates are present
		gotMap := make(map[string]bool)
		for _, tmpl := range got {
			gotMap[tmpl] = true
		}

		for _, expected := range templates {
			if !gotMap[expected] {
				t.Errorf("Engine.ListTemplates() missing template %s", expected)
			}
		}
	})
}

func TestCustomFuncs(t *testing.T) {
	funcs := customFuncs()

	tests := []struct {
		name     string
		funcName string
		verify   func(t *testing.T)
	}{
		{
			name:     "toLower function exists",
			funcName: "toLower",
			verify: func(t *testing.T) {
				if _, ok := funcs["toLower"]; !ok {
					t.Error("toLower function not found")
				}
			},
		},
		{
			name:     "toUpper function exists",
			funcName: "toUpper",
			verify: func(t *testing.T) {
				if _, ok := funcs["toUpper"]; !ok {
					t.Error("toUpper function not found")
				}
			},
		},
		{
			name:     "replace function exists",
			funcName: "replace",
			verify: func(t *testing.T) {
				if _, ok := funcs["replace"]; !ok {
					t.Error("replace function not found")
				}
			},
		},
		{
			name:     "trimSpace function exists",
			funcName: "trimSpace",
			verify: func(t *testing.T) {
				if _, ok := funcs["trimSpace"]; !ok {
					t.Error("trimSpace function not found")
				}
			},
		},
		{
			name:     "join function exists",
			funcName: "join",
			verify: func(t *testing.T) {
				if _, ok := funcs["join"]; !ok {
					t.Error("join function not found")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verify(t)
		})
	}
}

func TestEngine_NewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
	if engine.templatesDir == "" {
		t.Error("NewEngine() templates directory not set")
	}
}
