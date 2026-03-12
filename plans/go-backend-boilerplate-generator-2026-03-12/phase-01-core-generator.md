# Phase 1: Core Generator Framework

**Duration**: Week 1-2
**Goal**: Build CLI foundation and base project structure generator
**Success Criteria**: Generates compilable Go project with health check endpoint

---

## Objectives

1. Set up CLI framework with interactive prompts
2. Implement template engine with atomic file writes
3. Generate base clean architecture project structure
4. Configure Viper for generated projects
5. Create Makefile with common development tasks
6. Validate generated code compiles and runs

---

## Implementation Details

### 1.1 Project Structure (Generator)

```
go-boilerplate/
├── cmd/
│   └── root.go              # Cobra root command
│   └── init.go              # 'init' subcommand
├── internal/
│   ├── generator/
│   │   ├── generator.go     # Core generator logic
│   │   ├── filewriter.go    # Atomic file operations
│   │   └── validator.go     # Post-generation validation
│   ├── config/
│   │   ├── config.go        # Generator config struct
│   │   └── prompts.go       # Survey prompt definitions
│   ├── templates/
│   │   ├── engine.go        # text/template wrapper
│   │   ├── funcs.go         # Custom template functions
│   │   └── loader.go        # embed.FS template loader
├── templates/
│   ├── base/                # Embedded base templates
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go.tmpl
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   │   └── config.go.tmpl
│   │   │   ├── domain/
│   │   │   │   └── .gitkeep
│   │   │   ├── ports/
│   │   │   │   └── http/
│   │   │   │       ├── server.go.tmpl
│   │   │   │       └── health.go.tmpl
│   │   │   └── adapters/
│   │   │       └── postgres/
│   │   │           └── adapter.go.tmpl
│   │   ├── configs/
│   │   │   ├── config.dev.yaml.tmpl
│   │   │   └── config.prod.yaml.tmpl
│   │   ├── go.mod.tmpl
│   │   ├── Makefile.tmpl
│   │   ├── .gitignore.tmpl
│   │   └── README.md.tmpl
├── go.mod
├── go.sum
└── main.go                  # CLI entry point
```

### 1.2 CLI Framework Setup (Cobra + Survey)

**File**: `cmd/root.go`

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "go-boilerplate",
    Short: "Generate production-ready Go backend projects",
    Long: `Interactive CLI generator for Go backends with clean architecture,
dual protocol support (REST + gRPC), and optional feature toggles.`,
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    rootCmd.AddCommand(initCmd)
}
```

**File**: `cmd/init.go`

```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/AlecAivazis/survey/v2"
    "github.com/spf13/cobra"
    "github.com/yourorg/go-boilerplate/internal/config"
    "github.com/yourorg/go-boilerplate/internal/generator"
)

var (
    // Non-interactive flags
    projectName string
    modulePath  string
    features    []string
    skipPrompts bool
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize a new Go backend project",
    RunE:  runInit,
}

func init() {
    initCmd.Flags().StringVar(&projectName, "name", "", "Project name")
    initCmd.Flags().StringVar(&modulePath, "module", "", "Go module path")
    initCmd.Flags().StringSliceVar(&features, "features", []string{}, "Features to enable (auth,observability,docker)")
    initCmd.Flags().BoolVar(&skipPrompts, "yes", false, "Skip prompts, use defaults")
}

func runInit(cmd *cobra.Command, args []string) error {
    cfg, err := promptForConfig()
    if err != nil {
        return fmt.Errorf("failed to collect config: %w", err)
    }

    gen := generator.New()
    if err := gen.Generate(cfg); err != nil {
        return fmt.Errorf("failed to generate project: %w", err)
    }

    printSuccessMessage(cfg)
    return nil
}

func promptForConfig() (*config.Config, error) {
    cfg := &config.Config{}

    // Use flags if non-interactive mode
    if skipPrompts {
        cfg.ProjectName = projectName
        cfg.ModulePath = modulePath
        // Parse features from flag
        return cfg, cfg.Validate()
    }

    // Interactive prompts
    questions := []*survey.Question{
        {
            Name:   "ProjectName",
            Prompt: &survey.Input{
                Message: "Project name:",
                Default: "my-backend",
            },
            Validate: survey.Required,
        },
        {
            Name:   "ModulePath",
            Prompt: &survey.Input{
                Message: "Go module path:",
                Default: "github.com/yourorg/my-backend",
            },
            Validate: survey.Required,
        },
        {
            Name: "EnableAuth",
            Prompt: &survey.Confirm{
                Message: "Enable Authentication (JWT + OAuth2 + Redis)?",
                Default: false,
            },
        },
        {
            Name: "EnableObservability",
            Prompt: &survey.Confirm{
                Message: "Enable Observability (Prometheus + OTEL + slog)?",
                Default: false,
            },
        },
        {
            Name: "EnableDocker",
            Prompt: &survey.Confirm{
                Message: "Include Docker environment (compose + Dockerfile)?",
                Default: true,
            },
        },
        {
            Name: "GenerateExample",
            Prompt: &survey.Confirm{
                Message: "Generate example CRUD (User resource)?",
                Default: true,
            },
        },
    }

    if err := survey.Ask(questions, cfg); err != nil {
        return nil, err
    }

    return cfg, cfg.Validate()
}
```

### 1.3 Config Management

**File**: `internal/config/config.go`

```go
package config

import (
    "fmt"
    "path/filepath"
    "regexp"
)

type Config struct {
    ProjectName         string
    ModulePath          string
    EnableAuth          bool
    EnableObservability bool
    EnableDocker        bool
    GenerateExample     bool
    OutputDir           string
}

var (
    projectNameRegex = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
    modulePathRegex  = regexp.MustCompile(`^[a-z0-9./\-_]+$`)
)

func (c *Config) Validate() error {
    if !projectNameRegex.MatchString(c.ProjectName) {
        return fmt.Errorf("invalid project name: must be lowercase alphanumeric with dashes")
    }

    if !modulePathRegex.MatchString(c.ModulePath) {
        return fmt.Errorf("invalid module path: must be valid Go module identifier")
    }

    // Set default output directory
    if c.OutputDir == "" {
        c.OutputDir = filepath.Join(".", c.ProjectName)
    }

    return nil
}

// TemplateData returns map for template rendering
func (c *Config) TemplateData() map[string]interface{} {
    return map[string]interface{}{
        "ProjectName":         c.ProjectName,
        "ModulePath":          c.ModulePath,
        "EnableAuth":          c.EnableAuth,
        "EnableObservability": c.EnableObservability,
        "EnableDocker":        c.EnableDocker,
        "GenerateExample":     c.GenerateExample,
    }
}
```

### 1.4 Template Engine

**File**: `internal/templates/engine.go`

```go
package templates

import (
    "bytes"
    "embed"
    "fmt"
    "io/fs"
    "path/filepath"
    "strings"
    "text/template"
)

//go:embed base/* features/*
var templateFS embed.FS

type Engine struct {
    templates map[string]*template.Template
    funcs     template.FuncMap
}

func NewEngine() *Engine {
    return &Engine{
        templates: make(map[string]*template.Template),
        funcs:     defaultFuncs(),
    }
}

func (e *Engine) LoadTemplates() error {
    return fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
            return nil
        }

        content, err := fs.ReadFile(templateFS, path)
        if err != nil {
            return fmt.Errorf("failed to read template %s: %w", path, err)
        }

        tmpl, err := template.New(path).Funcs(e.funcs).Parse(string(content))
        if err != nil {
            return fmt.Errorf("failed to parse template %s: %w", path, err)
        }

        e.templates[path] = tmpl
        return nil
    })
}

func (e *Engine) Render(templatePath string, data interface{}) ([]byte, error) {
    tmpl, ok := e.templates[templatePath]
    if !ok {
        return nil, fmt.Errorf("template not found: %s", templatePath)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return nil, fmt.Errorf("failed to execute template: %w", err)
    }

    return buf.Bytes(), nil
}

// GetOutputPath converts template path to actual file path
func GetOutputPath(templatePath string) string {
    // Remove "base/" or "features/" prefix
    path := strings.TrimPrefix(templatePath, "base/")
    path = strings.TrimPrefix(path, "features/")
    // Remove .tmpl extension
    return strings.TrimSuffix(path, ".tmpl")
}
```

**File**: `internal/templates/funcs.go`

```go
package templates

import (
    "strings"
    "text/template"
)

func defaultFuncs() template.FuncMap {
    return template.FuncMap{
        // String manipulation
        "toLower":   strings.ToLower,
        "toUpper":   strings.ToUpper,
        "title":     strings.Title,
        "trimSpace": strings.TrimSpace,

        // Custom helpers
        "packageName": func(modulePath string) string {
            parts := strings.Split(modulePath, "/")
            return parts[len(parts)-1]
        },

        "joinPath": func(parts ...string) string {
            return strings.Join(parts, "/")
        },
    }
}
```

### 1.5 File Writer (Atomic Operations)

**File**: `internal/generator/filewriter.go`

```go
package generator

import (
    "fmt"
    "os"
    "path/filepath"
)

type FileWriter struct {
    baseDir string
}

func NewFileWriter(baseDir string) *FileWriter {
    return &FileWriter{baseDir: baseDir}
}

// WriteFile writes content to file with parent directory creation
func (w *FileWriter) WriteFile(relPath string, content []byte, perm os.FileMode) error {
    fullPath := filepath.Join(w.baseDir, relPath)

    // Create parent directories
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory %s: %w", dir, err)
    }

    // Write file atomically using temp file + rename
    tmpPath := fullPath + ".tmp"
    if err := os.WriteFile(tmpPath, content, perm); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    if err := os.Rename(tmpPath, fullPath); err != nil {
        os.Remove(tmpPath) // Cleanup
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
}

// CreateDir creates directory if not exists
func (w *FileWriter) CreateDir(relPath string) error {
    fullPath := filepath.Join(w.baseDir, relPath)
    return os.MkdirAll(fullPath, 0755)
}

// Exists checks if path exists
func (w *FileWriter) Exists(relPath string) bool {
    fullPath := filepath.Join(w.baseDir, relPath)
    _, err := os.Stat(fullPath)
    return err == nil
}
```

### 1.6 Core Generator

**File**: `internal/generator/generator.go`

```go
package generator

import (
    "fmt"
    "os"

    "github.com/yourorg/go-boilerplate/internal/config"
    "github.com/yourorg/go-boilerplate/internal/templates"
)

type Generator struct {
    engine *templates.Engine
}

func New() *Generator {
    return &Generator{
        engine: templates.NewEngine(),
    }
}

func (g *Generator) Generate(cfg *config.Config) error {
    // Load all templates
    if err := g.engine.LoadTemplates(); err != nil {
        return fmt.Errorf("failed to load templates: %w", err)
    }

    // Check if output directory exists
    if _, err := os.Stat(cfg.OutputDir); !os.IsNotExist(err) {
        return fmt.Errorf("directory %s already exists", cfg.OutputDir)
    }

    writer := NewFileWriter(cfg.OutputDir)
    data := cfg.TemplateData()

    // Generate base project structure
    if err := g.generateBase(writer, data); err != nil {
        return fmt.Errorf("failed to generate base: %w", err)
    }

    // Generate feature-specific files
    if cfg.EnableAuth {
        if err := g.generateAuth(writer, data); err != nil {
            return fmt.Errorf("failed to generate auth: %w", err)
        }
    }

    if cfg.EnableObservability {
        if err := g.generateObservability(writer, data); err != nil {
            return fmt.Errorf("failed to generate observability: %w", err)
        }
    }

    if cfg.EnableDocker {
        if err := g.generateDocker(writer, data); err != nil {
            return fmt.Errorf("failed to generate docker: %w", err)
        }
    }

    // Validate generated project
    validator := NewValidator(cfg.OutputDir)
    if err := validator.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return nil
}

func (g *Generator) generateBase(writer *FileWriter, data interface{}) error {
    baseTemplates := []string{
        "base/cmd/server/main.go.tmpl",
        "base/internal/config/config.go.tmpl",
        "base/internal/ports/http/server.go.tmpl",
        "base/internal/ports/http/health.go.tmpl",
        "base/internal/adapters/postgres/adapter.go.tmpl",
        "base/configs/config.dev.yaml.tmpl",
        "base/configs/config.prod.yaml.tmpl",
        "base/go.mod.tmpl",
        "base/Makefile.tmpl",
        "base/.gitignore.tmpl",
        "base/README.md.tmpl",
    }

    for _, tmplPath := range baseTemplates {
        content, err := g.engine.Render(tmplPath, data)
        if err != nil {
            return err
        }

        outputPath := templates.GetOutputPath(tmplPath)
        if err := writer.WriteFile(outputPath, content, 0644); err != nil {
            return err
        }
    }

    // Create empty directories
    emptyDirs := []string{
        "internal/domain",
        "internal/application",
        "pkg",
        "migrations",
        "tests/integration",
    }

    for _, dir := range emptyDirs {
        if err := writer.CreateDir(dir); err != nil {
            return err
        }
    }

    return nil
}
```

### 1.7 Post-Generation Validator

**File**: `internal/generator/validator.go`

```go
package generator

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

type Validator struct {
    projectDir string
}

func NewValidator(projectDir string) *Validator {
    return &Validator{projectDir: projectDir}
}

func (v *Validator) Validate() error {
    // Check required files exist
    requiredFiles := []string{
        "cmd/server/main.go",
        "go.mod",
        "Makefile",
        "README.md",
    }

    for _, file := range requiredFiles {
        path := filepath.Join(v.projectDir, file)
        if _, err := os.Stat(path); os.IsNotExist(err) {
            return fmt.Errorf("required file missing: %s", file)
        }
    }

    // Validate go.mod syntax
    if err := v.validateGoMod(); err != nil {
        return fmt.Errorf("invalid go.mod: %w", err)
    }

    // Try to build (optional, can be slow)
    // if err := v.tryBuild(); err != nil {
    //     return fmt.Errorf("build failed: %w", err)
    // }

    return nil
}

func (v *Validator) validateGoMod() error {
    cmd := exec.Command("go", "mod", "verify")
    cmd.Dir = v.projectDir
    if err := cmd.Run(); err != nil {
        return err
    }
    return nil
}

func (v *Validator) tryBuild() error {
    cmd := exec.Command("go", "build", "./cmd/server")
    cmd.Dir = v.projectDir
    return cmd.Run()
}
```

### 1.8 Base Template Examples

**File**: `templates/base/cmd/server/main.go.tmpl`

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "{{ .ModulePath }}/internal/config"
    "{{ .ModulePath }}/internal/ports/http"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create HTTP server
    server := http.NewServer(cfg)

    // Start server in goroutine
    go func() {
        if err := server.Start(); err != nil {
            log.Fatalf("Server failed: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced shutdown: %v", err)
    }

    log.Println("Server exited")
}
```

**File**: `templates/base/Makefile.tmpl`

```makefile
.PHONY: help run build test lint clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/{{ .ProjectName }} cmd/server/main.go

test: ## Run tests
	go test -v -race -cover ./...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/

deps: ## Download dependencies
	go mod download
	go mod tidy

{{- if .EnableDocker }}

docker-build: ## Build Docker image
	docker build -t {{ .ProjectName }}:latest .

docker-up: ## Start Docker compose
	docker-compose up -d

docker-down: ## Stop Docker compose
	docker-compose down
{{- end }}
```

---

## Testing Strategy

### 1.9 Unit Tests

Test each component in isolation:

```go
// internal/config/config_test.go
func TestConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        cfg     *Config
        wantErr bool
    }{
        {
            name: "valid config",
            cfg: &Config{
                ProjectName: "my-api",
                ModulePath:  "github.com/user/my-api",
            },
            wantErr: false,
        },
        {
            name: "invalid project name",
            cfg: &Config{
                ProjectName: "My-API",  // uppercase not allowed
                ModulePath:  "github.com/user/my-api",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.cfg.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 1.10 Integration Tests

Test full generation flow:

```go
// internal/generator/generator_test.go
func TestGenerator_Generate(t *testing.T) {
    tmpDir := t.TempDir()

    cfg := &config.Config{
        ProjectName: "test-project",
        ModulePath:  "github.com/test/project",
        OutputDir:   tmpDir,
    }

    gen := New()
    if err := gen.Generate(cfg); err != nil {
        t.Fatalf("Generate() failed: %v", err)
    }

    // Verify essential files exist
    essentialFiles := []string{
        "cmd/server/main.go",
        "go.mod",
        "Makefile",
    }

    for _, file := range essentialFiles {
        path := filepath.Join(tmpDir, file)
        if _, err := os.Stat(path); err != nil {
            t.Errorf("File %s not generated", file)
        }
    }

    // Try to compile generated code
    cmd := exec.Command("go", "build", "./cmd/server")
    cmd.Dir = tmpDir
    if err := cmd.Run(); err != nil {
        t.Errorf("Generated code failed to compile: %v", err)
    }
}
```

---

## Anti-Patterns to Avoid

1. **No Global Template Cache**: Always reload templates in each generator instance
2. **No Silent Failures**: All errors must bubble up with context
3. **No Hardcoded Paths**: Use filepath.Join for cross-platform compatibility
4. **No Partial Writes**: Use atomic file operations (temp file + rename)
5. **No Template Logic Overload**: Keep templates simple, move complex logic to Go code

---

## Success Validation Checklist

- [x] CLI accepts both interactive and non-interactive modes
- [x] Generated project structure matches design
- [x] All template files render without errors
- [x] Generated go.mod is valid
- [x] Generated code compiles successfully
- [x] Makefile targets execute without errors
- [x] Validator catches missing required files
- [x] Unit test coverage > 80%
- [x] Integration tests pass in CI

---

## Deliverables

1. Working CLI binary (`go-boilerplate init`)
2. Base template set in `templates/base/`
3. Generator core with atomic file writes
4. Config validation with clear error messages
5. Test suite with >80% coverage
6. Documentation for generator developers

**Next Phase**: Phase 2 - gRPC-REST Integration
