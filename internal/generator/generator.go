package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/yourusername/go-boilerplate/internal/config"
	"github.com/yourusername/go-boilerplate/internal/templates"
)

// Generator handles project generation
type Generator struct {
	templateEngine *templates.Engine
	fileWriter     *FileWriter
}

// New creates a new Generator instance
func New() *Generator {
	return &Generator{
		templateEngine: templates.NewEngine(),
		fileWriter:     NewFileWriter(),
	}
}

// Generate creates a new project based on the provided configuration
func (g *Generator) Generate(cfg *config.Config) error {
	// Validate and sanitize output directory to prevent path traversal
	if err := validateOutputPath(cfg.OutputDir); err != nil {
		return fmt.Errorf("invalid output directory: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate base project structure
	if err := g.generateBaseStructure(cfg); err != nil {
		return fmt.Errorf("failed to generate base structure: %w", err)
	}

	// Generate feature-specific files
	if err := g.generateFeatures(cfg); err != nil {
		return fmt.Errorf("failed to generate features: %w", err)
	}

	// Initialize Git repository if requested
	if cfg.InitGit {
		if err := initGit(cfg.OutputDir); err != nil {
			return fmt.Errorf("failed to initialize git: %w", err)
		}
	}

	// Validate generated project
	validator := NewValidator(cfg.OutputDir)
	if err := validator.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

func (g *Generator) generateBaseStructure(cfg *config.Config) error {
	baseFiles := []struct {
		template string
		output   string
	}{
		// Root files
		{"base/go.mod.tmpl", "go.mod"},
		{"base/.gitignore.tmpl", ".gitignore"},
		{"base/Makefile.tmpl", "Makefile"},
		{"base/README.md.tmpl", "README.md"},

		// Application entry
		{"base/cmd/server/main.go.tmpl", "cmd/server/main.go"},

		// Configuration
		{"base/internal/config/config.go.tmpl", "internal/config/config.go"},
		{"base/configs/config.dev.yaml.tmpl", "configs/config.dev.yaml"},
		{"base/configs/config.prod.yaml.tmpl", "configs/config.prod.yaml"},

		// HTTP/REST layer
		{"base/internal/ports/http/server.go.tmpl", "internal/ports/http/server.go"},
		{"base/internal/ports/http/gateway.go.tmpl", "internal/ports/http/gateway.go"},

		// gRPC layer
		{"base/internal/ports/grpc/server.go.tmpl", "internal/ports/grpc/server.go"},
		{"base/internal/ports/grpc/health.go.tmpl", "internal/ports/grpc/health.go"},

		// API/Proto definitions
		{"base/api/buf.yaml.tmpl", "api/buf.yaml"},
		{"base/api/buf.gen.yaml.tmpl", "api/buf.gen.yaml"},
		{"base/api/proto/v1/common.proto.tmpl", "api/proto/v1/common.proto"},
		{"base/api/proto/v1/health.proto.tmpl", "api/proto/v1/health.proto"},

		// PostgreSQL adapter
		{"base/internal/adapters/postgres/adapter.go.tmpl", "internal/adapters/postgres/adapter.go"},
		{"base/internal/adapters/postgres/sqlc.yaml.tmpl", "internal/adapters/postgres/sqlc.yaml"},
		{"base/internal/adapters/postgres/schema.sql.tmpl", "internal/adapters/postgres/schema.sql"},
		{"base/internal/adapters/postgres/queries.sql.tmpl", "internal/adapters/postgres/queries.sql"},

		// Database migrations
		{"base/migrations/000001_initial_schema.up.sql.tmpl", "migrations/000001_initial_schema.up.sql"},
		{"base/migrations/000001_initial_schema.down.sql.tmpl", "migrations/000001_initial_schema.down.sql"},

		// Helper scripts
		{"base/scripts/generate-proto.sh.tmpl", "scripts/generate-proto.sh"},
		{"base/scripts/migrate.sh.tmpl", "scripts/migrate.sh"},
	}

	for _, f := range baseFiles {
		content, err := g.templateEngine.Render(f.template, cfg)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", f.template, err)
		}

		outputPath := filepath.Join(cfg.OutputDir, f.output)
		if err := g.fileWriter.WriteFile(outputPath, content); err != nil {
			return fmt.Errorf("failed to write file %s: %w", outputPath, err)
		}
	}

	// Create empty directories with .gitkeep
	emptyDirs := []string{
		"internal/domain",
		"internal/application",
		"internal/gen",      // For generated proto code
		"api/openapi",       // For generated OpenAPI specs
		"pkg/logger",
		"pkg/validator",
		"tests/integration",
		"tests/e2e",
	}

	for _, dir := range emptyDirs {
		dirPath := filepath.Join(cfg.OutputDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
		gitkeepPath := filepath.Join(dirPath, ".gitkeep")
		if err := g.fileWriter.WriteFile(gitkeepPath, []byte("")); err != nil {
			return fmt.Errorf("failed to create .gitkeep in %s: %w", dirPath, err)
		}
	}

	// Make scripts executable
	scriptsToChmod := []string{
		"scripts/generate-proto.sh",
		"scripts/migrate.sh",
	}

	for _, script := range scriptsToChmod {
		scriptPath := filepath.Join(cfg.OutputDir, script)
		if err := os.Chmod(scriptPath, 0755); err != nil {
			return fmt.Errorf("failed to make %s executable: %w", script, err)
		}
	}

	return nil
}

func (g *Generator) generateFeatures(cfg *config.Config) error {
	// Auth feature
	if cfg.Features.Auth {
		authFiles := []struct {
			template string
			output   string
		}{
			// Domain
			{"features/auth/internal/domain/user.go.tmpl", "internal/domain/user.go"},

			// Application
			{"features/auth/internal/application/auth_service.go.tmpl", "internal/application/auth_service.go"},

			// Ports - HTTP
			{"features/auth/internal/ports/http/auth_middleware.go.tmpl", "internal/ports/http/auth_middleware.go"},
			{"features/auth/internal/ports/http/auth_handler.go.tmpl", "internal/ports/http/auth_handler.go"},

			// Ports - gRPC
			{"features/auth/internal/ports/grpc/auth_interceptor.go.tmpl", "internal/ports/grpc/auth_interceptor.go"},

			// Adapters
			{"features/auth/internal/adapters/redis/session_store.go.tmpl", "internal/adapters/redis/session_store.go"},
			{"features/auth/internal/adapters/oauth/google.go.tmpl", "internal/adapters/oauth/google.go"},
			{"features/auth/internal/adapters/oauth/github.go.tmpl", "internal/adapters/oauth/github.go"},

			// Packages
			{"features/auth/pkg/jwt/jwt.go.tmpl", "pkg/jwt/jwt.go"},

			// API
			{"features/auth/api/proto/v1/auth.proto.tmpl", "api/proto/v1/auth.proto"},

			// Migrations
			{"features/auth/migrations/000002_create_users.up.sql.tmpl", "migrations/000002_create_users.up.sql"},
			{"features/auth/migrations/000002_create_users.down.sql.tmpl", "migrations/000002_create_users.down.sql"},
		}
		if err := g.generateTemplateFiles(cfg, authFiles, "auth"); err != nil {
			return err
		}
	}

	// Observability feature
	if cfg.Features.Observability {
		obsFiles := []struct {
			template string
			output   string
		}{
			{"features/observability/internal/telemetry/metrics.go.tmpl", "internal/telemetry/metrics.go"},
			{"features/observability/internal/telemetry/tracing.go.tmpl", "internal/telemetry/tracing.go"},
			{"features/observability/internal/telemetry/logging.go.tmpl", "internal/telemetry/logging.go"},
			{"features/observability/pkg/logger/logger.go.tmpl", "pkg/logger/logger.go"},
			{"features/observability/configs/prometheus.yml.tmpl", "configs/prometheus.yml"},
			{"features/observability/configs/otel-collector-config.yaml.tmpl", "configs/otel-collector-config.yaml"},
		}
		if err := g.generateTemplateFiles(cfg, obsFiles, "observability"); err != nil {
			return err
		}
	}

	// Docker feature
	if cfg.Features.Docker {
		dockerFiles := []struct {
			template string
			output   string
		}{
			{"features/docker/Dockerfile.tmpl", "Dockerfile"},
			{"features/docker/docker-compose.yml.tmpl", "docker-compose.yml"},
			{"features/docker/.dockerignore.tmpl", ".dockerignore"},
		}
		if err := g.generateTemplateFiles(cfg, dockerFiles, "docker"); err != nil {
			return err
		}
	}

	return nil
}

// generateTemplateFiles renders and writes a list of template files (DRY helper)
func (g *Generator) generateTemplateFiles(cfg *config.Config, files []struct{ template, output string }, featureName string) error {
	for _, f := range files {
		content, err := g.templateEngine.Render(f.template, cfg)
		if err != nil {
			return fmt.Errorf("failed to render %s template %s: %w", featureName, f.template, err)
		}

		outputPath := filepath.Join(cfg.OutputDir, f.output)
		if err := g.fileWriter.WriteFile(outputPath, content); err != nil {
			return fmt.Errorf("failed to write %s file %s: %w", featureName, outputPath, err)
		}
	}
	return nil
}

// initGit initializes a git repository in the specified directory
func initGit(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
