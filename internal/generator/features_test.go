package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ThinhDangDev/go-boilerplate/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllFeatureCombinations(t *testing.T) {
	combinations := []struct {
		name          string
		auth          bool
		observability bool
		docker        bool
	}{
		{"base-only", false, false, false},
		{"auth-only", true, false, false},
		{"observability-only", false, true, false},
		{"docker-only", false, false, true},
		{"auth-observability", true, true, false},
		{"auth-docker", true, false, true},
		{"observability-docker", false, true, true},
		{"all-features", true, true, true},
	}

	for _, tc := range combinations {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				ProjectName: "test-" + tc.name,
				ModuleName:  "github.com/test/" + tc.name,
				Features: config.Features{
					Auth:          tc.auth,
					Observability: tc.observability,
					Docker:        tc.docker,
				},
				OutputDir: tmpDir,
				InitGit:   false,
			}

			gen := New()
			err := gen.Generate(cfg)
			require.NoError(t, err)

			// Verify base files always exist
			verifyBaseFiles(t, tmpDir)

			// Verify feature-specific files
			verifyFeatureFiles(t, tmpDir, cfg)
		})
	}
}

func verifyBaseFiles(t *testing.T, dir string) {
	baseFiles := []string{
		"go.mod",
		"Makefile",
		"README.md",
		".gitignore",
		"cmd/server/main.go",
		"internal/config/config.go",
		"internal/ports/http/server.go",
		"internal/ports/grpc/server.go",
		"api/proto/v1/health.proto",
	}

	for _, file := range baseFiles {
		path := filepath.Join(dir, file)
		assert.FileExists(t, path, "Base file should exist: %s", file)
	}
}

func verifyFeatureFiles(t *testing.T, dir string, cfg *config.Config) {
	if cfg.Features.Auth {
		authFiles := []string{
			"internal/domain/user.go",
			"internal/application/auth_service.go",
			"internal/ports/http/auth_middleware.go",
			"internal/ports/http/auth_handler.go",
			"internal/ports/grpc/auth_interceptor.go",
			"internal/adapters/redis/session_store.go",
			"internal/adapters/oauth/google.go",
			"internal/adapters/oauth/github.go",
			"pkg/jwt/jwt.go",
			"api/proto/v1/auth.proto",
			"migrations/000002_create_users.up.sql",
			"migrations/000002_create_users.down.sql",
		}
		for _, file := range authFiles {
			path := filepath.Join(dir, file)
			assert.FileExists(t, path, "Auth file should exist: %s", file)
		}

		// Verify go.mod contains auth dependencies
		goModContent, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		require.NoError(t, err)
		assert.Contains(t, string(goModContent), "github.com/golang-jwt/jwt/v5")
		assert.Contains(t, string(goModContent), "github.com/redis/go-redis/v9")
		assert.Contains(t, string(goModContent), "golang.org/x/oauth2")
	} else {
		// Verify auth files do NOT exist
		assert.NoFileExists(t, filepath.Join(dir, "pkg/jwt/jwt.go"))
		assert.NoFileExists(t, filepath.Join(dir, "internal/domain/user.go"))
	}

	if cfg.Features.Observability {
		obsFiles := []string{
			"internal/telemetry/metrics.go",
			"internal/telemetry/tracing.go",
			"internal/telemetry/logging.go",
			"pkg/logger/logger.go",
			"configs/prometheus.yml",
			"configs/otel-collector-config.yaml",
		}
		for _, file := range obsFiles {
			path := filepath.Join(dir, file)
			assert.FileExists(t, path, "Observability file should exist: %s", file)
		}

		// Verify go.mod contains observability dependencies
		goModContent, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		require.NoError(t, err)
		assert.Contains(t, string(goModContent), "github.com/prometheus/client_golang")
		assert.Contains(t, string(goModContent), "go.opentelemetry.io/otel")
	} else {
		// Verify observability files do NOT exist
		assert.NoFileExists(t, filepath.Join(dir, "internal/telemetry/metrics.go"))
		assert.NoFileExists(t, filepath.Join(dir, "configs/prometheus.yml"))
	}

	if cfg.Features.Docker {
		dockerFiles := []string{
			"Dockerfile",
			"docker-compose.yml",
			".dockerignore",
		}
		for _, file := range dockerFiles {
			path := filepath.Join(dir, file)
			assert.FileExists(t, path, "Docker file should exist: %s", file)
		}

		// Verify docker-compose.yml has correct services based on features
		composeContent, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
		require.NoError(t, err)
		composeStr := string(composeContent)

		assert.Contains(t, composeStr, "postgres:")
		assert.Contains(t, composeStr, "app:")

		if cfg.Features.Auth {
			assert.Contains(t, composeStr, "redis:")
		} else {
			assert.NotContains(t, composeStr, "redis:")
		}

		if cfg.Features.Observability {
			assert.Contains(t, composeStr, "prometheus:")
			assert.Contains(t, composeStr, "grafana:")
			assert.Contains(t, composeStr, "otel-collector:")
		} else {
			assert.NotContains(t, composeStr, "prometheus:")
		}
	} else {
		// Verify docker files do NOT exist
		assert.NoFileExists(t, filepath.Join(dir, "Dockerfile"))
		assert.NoFileExists(t, filepath.Join(dir, "docker-compose.yml"))
	}
}

func TestFeatureIndependence(t *testing.T) {
	// Test that features can be enabled independently without conflicts
	t.Run("auth-without-redis-in-compose", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := &config.Config{
			ProjectName: "test-auth-no-docker",
			ModuleName:  "github.com/test/auth-no-docker",
			Features: config.Features{
				Auth:   true,
				Docker: false,
			},
			OutputDir: tmpDir,
			InitGit:   false,
		}

		gen := New()
		err := gen.Generate(cfg)
		require.NoError(t, err)

		// Auth files should exist
		assert.FileExists(t, filepath.Join(tmpDir, "pkg/jwt/jwt.go"))

		// Docker files should not exist
		assert.NoFileExists(t, filepath.Join(tmpDir, "docker-compose.yml"))
	})

	t.Run("docker-without-optional-services", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := &config.Config{
			ProjectName: "test-docker-only",
			ModuleName:  "github.com/test/docker-only",
			Features: config.Features{
				Docker: true,
			},
			OutputDir: tmpDir,
			InitGit:   false,
		}

		gen := New()
		err := gen.Generate(cfg)
		require.NoError(t, err)

		// Verify docker-compose.yml exists but doesn't have redis/prometheus
		composeContent, err := os.ReadFile(filepath.Join(tmpDir, "docker-compose.yml"))
		require.NoError(t, err)
		composeStr := string(composeContent)

		assert.Contains(t, composeStr, "postgres:")
		assert.NotContains(t, composeStr, "redis:")
		assert.NotContains(t, composeStr, "prometheus:")
	})
}

func TestGoModDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-all-features",
		ModuleName:  "github.com/test/all-features",
		Features: config.Features{
			Auth:          true,
			Observability: true,
			Docker:        true,
		},
		OutputDir: tmpDir,
		InitGit:   false,
	}

	gen := New()
	err := gen.Generate(cfg)
	require.NoError(t, err)

	goModContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	require.NoError(t, err)
	goModStr := string(goModContent)

	// Base dependencies
	assert.Contains(t, goModStr, "google.golang.org/grpc")
	assert.Contains(t, goModStr, "github.com/jackc/pgx/v5")

	// Auth dependencies
	assert.Contains(t, goModStr, "github.com/golang-jwt/jwt/v5")
	assert.Contains(t, goModStr, "github.com/redis/go-redis/v9")
	assert.Contains(t, goModStr, "github.com/google/uuid")
	assert.Contains(t, goModStr, "golang.org/x/crypto")
	assert.Contains(t, goModStr, "golang.org/x/oauth2")

	// Observability dependencies
	assert.Contains(t, goModStr, "github.com/prometheus/client_golang")
	assert.Contains(t, goModStr, "go.opentelemetry.io/otel")
	assert.Contains(t, goModStr, "go.opentelemetry.io/otel/sdk")
	assert.Contains(t, goModStr, "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc")
}

func TestAuthMigrationOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-auth-migrations",
		ModuleName:  "github.com/test/auth-migrations",
		Features: config.Features{
			Auth: true,
		},
		OutputDir: tmpDir,
		InitGit:   false,
	}

	gen := New()
	err := gen.Generate(cfg)
	require.NoError(t, err)

	// Verify both base and auth migrations exist
	assert.FileExists(t, filepath.Join(tmpDir, "migrations/000001_initial_schema.up.sql"))
	assert.FileExists(t, filepath.Join(tmpDir, "migrations/000001_initial_schema.down.sql"))
	assert.FileExists(t, filepath.Join(tmpDir, "migrations/000002_create_users.up.sql"))
	assert.FileExists(t, filepath.Join(tmpDir, "migrations/000002_create_users.down.sql"))
}
