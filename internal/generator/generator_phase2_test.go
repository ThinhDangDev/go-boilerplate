package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/go-boilerplate/internal/config"
)

// setupPhase2Test ensures we're in the correct directory for template access
func setupPhase2Test(t *testing.T) {
	// Change to project root if we're not there
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		// Try to find project root by looking for go.mod
		currentDir, _ := os.Getwd()
		for {
			if _, err := os.Stat("go.mod"); err == nil {
				break
			}
			parent := filepath.Dir(currentDir)
			if parent == currentDir {
				t.Fatal("Could not find project root (no templates directory)")
			}
			os.Chdir(parent)
			currentDir = parent
		}
	}
}

// TestPhase2_GenerateBaseStructure tests that all Phase 2 files are generated
func TestPhase2_GenerateBaseStructure(t *testing.T) {
	setupPhase2Test(t)

	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features: config.Features{
			Auth:          false,
			Observability: false,
			Docker:        false,
		},
		InitGit:         false,
		GenerateExample: false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test all Phase 2 base files (25+ files)
	expectedFiles := []string{
		// Root files
		"go.mod",
		".gitignore",
		"Makefile",
		"README.md",

		// Application entry
		"cmd/server/main.go",

		// Configuration
		"internal/config/config.go",
		"configs/config.dev.yaml",
		"configs/config.prod.yaml",

		// HTTP/REST layer
		"internal/ports/http/server.go",
		"internal/ports/http/gateway.go",

		// gRPC layer
		"internal/ports/grpc/server.go",
		"internal/ports/grpc/health.go",

		// API/Proto definitions
		"api/buf.yaml",
		"api/buf.gen.yaml",
		"api/proto/v1/common.proto",
		"api/proto/v1/health.proto",

		// PostgreSQL adapter
		"internal/adapters/postgres/adapter.go",
		"internal/adapters/postgres/sqlc.yaml",
		"internal/adapters/postgres/schema.sql",
		"internal/adapters/postgres/queries.sql",

		// Database migrations
		"migrations/000001_initial_schema.up.sql",
		"migrations/000001_initial_schema.down.sql",

		// Helper scripts
		"scripts/generate-proto.sh",
		"scripts/migrate.sh",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected Phase 2 file %s does not exist", file)
		}
	}
}

// TestPhase2_BaseFilesCount verifies that we have 25+ base files in the baseFiles list
func TestPhase2_BaseFilesCount(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Count actual generated files (excluding .gitkeep files)
	fileCount := 0
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasSuffix(path, ".gitkeep") {
			fileCount++
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk directory: %v", err)
	}

	// Should have at least 24 files (24 base files from the baseFiles list)
	if fileCount < 24 {
		t.Errorf("generated %d files, want at least 24 Phase 2 base files", fileCount)
	}
}

// TestPhase2_ScriptPermissions tests that generated scripts are executable
func TestPhase2_ScriptPermissions(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test script permissions
	scripts := []string{
		"scripts/generate-proto.sh",
		"scripts/migrate.sh",
	}

	for _, script := range scripts {
		path := filepath.Join(tmpDir, script)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("failed to stat script %s: %v", script, err)
			continue
		}

		// Check if executable bit is set (0755)
		if info.Mode().Perm() != 0755 {
			t.Errorf("script %s has permissions %v, want 0755", script, info.Mode().Perm())
		}
	}
}

// TestPhase2_ScriptShebang tests that scripts have correct shebang
func TestPhase2_ScriptShebang(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	scripts := []string{
		"scripts/generate-proto.sh",
		"scripts/migrate.sh",
	}

	for _, script := range scripts {
		path := filepath.Join(tmpDir, script)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read script %s: %v", script, err)
			continue
		}

		// Check for bash shebang
		if !strings.HasPrefix(string(content), "#!/bin/bash") {
			t.Errorf("script %s does not start with #!/bin/bash shebang", script)
		}
	}
}

// TestPhase2_ProtoFiles tests proto file generation
func TestPhase2_ProtoFiles(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	protoFiles := []string{
		"api/proto/v1/common.proto",
		"api/proto/v1/health.proto",
	}

	for _, protoFile := range protoFiles {
		path := filepath.Join(tmpDir, protoFile)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read proto file %s: %v", protoFile, err)
			continue
		}

		protoContent := string(content)

		// Verify proto syntax
		if !strings.Contains(protoContent, "syntax = \"proto3\";") {
			t.Errorf("proto file %s does not contain proto3 syntax declaration", protoFile)
		}

		// Verify package declaration
		if !strings.Contains(protoContent, "package api.v1;") {
			t.Errorf("proto file %s does not contain correct package declaration", protoFile)
		}

		// Verify go_package option with correct module name
		expectedGoPackage := "github.com/test/project/internal/gen/api/v1;apiv1"
		if !strings.Contains(protoContent, expectedGoPackage) {
			t.Errorf("proto file %s does not contain correct go_package option, want: %s", protoFile, expectedGoPackage)
		}
	}
}

// TestPhase2_BufConfiguration tests buf configuration files
func TestPhase2_BufConfiguration(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test buf.yaml
	bufYamlPath := filepath.Join(tmpDir, "api/buf.yaml")
	bufYamlContent, err := os.ReadFile(bufYamlPath)
	if err != nil {
		t.Fatalf("failed to read buf.yaml: %v", err)
	}

	bufYaml := string(bufYamlContent)
	if !strings.Contains(bufYaml, "version:") {
		t.Error("buf.yaml does not contain version")
	}

	// Test buf.gen.yaml
	bufGenYamlPath := filepath.Join(tmpDir, "api/buf.gen.yaml")
	bufGenYamlContent, err := os.ReadFile(bufGenYamlPath)
	if err != nil {
		t.Fatalf("failed to read buf.gen.yaml: %v", err)
	}

	bufGenYaml := string(bufGenYamlContent)
	expectedPlugins := []string{"protocolbuffers/go", "grpc/go", "grpc-ecosystem/gateway", "grpc-ecosystem/openapiv2"}
	for _, plugin := range expectedPlugins {
		if !strings.Contains(bufGenYaml, plugin) {
			t.Errorf("buf.gen.yaml does not contain plugin: %s", plugin)
		}
	}
}

// TestPhase2_PostgreSQLFiles tests PostgreSQL-related files
func TestPhase2_PostgreSQLFiles(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test sqlc.yaml
	sqlcYamlPath := filepath.Join(tmpDir, "internal/adapters/postgres/sqlc.yaml")
	sqlcContent, err := os.ReadFile(sqlcYamlPath)
	if err != nil {
		t.Fatalf("failed to read sqlc.yaml: %v", err)
	}

	sqlcYaml := string(sqlcContent)
	if !strings.Contains(sqlcYaml, "version:") {
		t.Error("sqlc.yaml does not contain version")
	}
	if !strings.Contains(sqlcYaml, "schema:") {
		t.Error("sqlc.yaml does not contain schema configuration")
	}
	if !strings.Contains(sqlcYaml, "queries:") {
		t.Error("sqlc.yaml does not contain queries configuration")
	}

	// Test schema.sql
	schemaPath := filepath.Join(tmpDir, "internal/adapters/postgres/schema.sql")
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema.sql: %v", err)
	}

	schema := string(schemaContent)
	if !strings.Contains(schema, "CREATE TABLE") {
		t.Error("schema.sql does not contain CREATE TABLE statement")
	}

	// Test queries.sql
	queriesPath := filepath.Join(tmpDir, "internal/adapters/postgres/queries.sql")
	if _, err := os.Stat(queriesPath); os.IsNotExist(err) {
		t.Error("queries.sql does not exist")
	}
}

// TestPhase2_Migrations tests database migration files
func TestPhase2_Migrations(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test migration up file
	upMigrationPath := filepath.Join(tmpDir, "migrations/000001_initial_schema.up.sql")
	upContent, err := os.ReadFile(upMigrationPath)
	if err != nil {
		t.Fatalf("failed to read up migration: %v", err)
	}

	upMigration := string(upContent)
	if !strings.Contains(upMigration, "CREATE TABLE") {
		t.Error("up migration does not contain CREATE TABLE statement")
	}

	// Test migration down file
	downMigrationPath := filepath.Join(tmpDir, "migrations/000001_initial_schema.down.sql")
	downContent, err := os.ReadFile(downMigrationPath)
	if err != nil {
		t.Fatalf("failed to read down migration: %v", err)
	}

	downMigration := string(downContent)
	if !strings.Contains(downMigration, "DROP TABLE") {
		t.Error("down migration does not contain DROP TABLE statement")
	}
}

// TestPhase2_GRPCServer tests gRPC server files
func TestPhase2_GRPCServer(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test gRPC server.go
	serverPath := filepath.Join(tmpDir, "internal/ports/grpc/server.go")
	serverContent, err := os.ReadFile(serverPath)
	if err != nil {
		t.Fatalf("failed to read gRPC server.go: %v", err)
	}

	server := string(serverContent)
	if !strings.Contains(server, "func New") {
		t.Error("gRPC server.go does not contain constructor function")
	}
	if !strings.Contains(server, "grpc.NewServer") {
		t.Error("gRPC server.go does not create new gRPC server")
	}

	// Test health.go
	healthPath := filepath.Join(tmpDir, "internal/ports/grpc/health.go")
	healthContent, err := os.ReadFile(healthPath)
	if err != nil {
		t.Fatalf("failed to read gRPC health.go: %v", err)
	}

	health := string(healthContent)
	if !strings.Contains(health, "HealthService") {
		t.Error("gRPC health.go does not implement HealthService")
	}
}

// TestPhase2_HTTPGateway tests HTTP gateway files
func TestPhase2_HTTPGateway(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	// Test gateway.go
	gatewayPath := filepath.Join(tmpDir, "internal/ports/http/gateway.go")
	gatewayContent, err := os.ReadFile(gatewayPath)
	if err != nil {
		t.Fatalf("failed to read gateway.go: %v", err)
	}

	gateway := string(gatewayContent)
	if !strings.Contains(gateway, "grpc-gateway") {
		t.Error("gateway.go does not reference grpc-gateway")
	}
	if !strings.Contains(gateway, "runtime.NewServeMux") {
		t.Error("gateway.go does not create ServeMux")
	}
}

// TestPhase2_TemplateRendering tests template rendering with different module names
func TestPhase2_TemplateRendering(t *testing.T) {
	setupPhase2Test(t)
	testCases := []struct {
		name       string
		moduleName string
	}{
		{"standard module", "github.com/user/project"},
		{"nested module", "github.com/org/team/project"},
		{"gitlab module", "gitlab.com/user/project"},
		{"custom domain", "example.com/project"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			cfg := &config.Config{
				ProjectName: "test-project",
				ModuleName:  tc.moduleName,
				OutputDir:   tmpDir,
				Features:    config.Features{},
				InitGit:     false,
			}

			gen := New()
			if err := gen.generateBaseStructure(cfg); err != nil {
				t.Fatalf("generateBaseStructure failed for %s: %v", tc.name, err)
			}

			// Verify go.mod contains correct module name
			goModPath := filepath.Join(tmpDir, "go.mod")
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				t.Fatalf("failed to read go.mod: %v", err)
			}

			if !strings.Contains(string(goModContent), tc.moduleName) {
				t.Errorf("go.mod does not contain module name %s", tc.moduleName)
			}

			// Verify proto files contain correct go_package
			protoPath := filepath.Join(tmpDir, "api/proto/v1/health.proto")
			protoContent, err := os.ReadFile(protoPath)
			if err != nil {
				t.Fatalf("failed to read proto file: %v", err)
			}

			expectedGoPackage := tc.moduleName + "/internal/gen/api/v1;apiv1"
			if !strings.Contains(string(protoContent), expectedGoPackage) {
				t.Errorf("proto file does not contain correct go_package: %s", expectedGoPackage)
			}
		})
	}
}

// TestPhase2_EmptyDirectories tests that empty directories are created with .gitkeep
func TestPhase2_EmptyDirectories(t *testing.T) {
	setupPhase2Test(t)
	tmpDir := t.TempDir()

	cfg := &config.Config{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/project",
		OutputDir:   tmpDir,
		Features:    config.Features{},
		InitGit:     false,
	}

	gen := New()
	if err := gen.generateBaseStructure(cfg); err != nil {
		t.Fatalf("generateBaseStructure failed: %v", err)
	}

	expectedDirs := []string{
		"internal/domain",
		"internal/application",
		"internal/gen",
		"api/openapi",
		"pkg/logger",
		"pkg/validator",
		"tests/integration",
		"tests/e2e",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(tmpDir, dir)
		gitkeepPath := filepath.Join(dirPath, ".gitkeep")

		// Check directory exists
		if info, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("expected directory %s does not exist", dir)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}

		// Check .gitkeep exists
		if _, err := os.Stat(gitkeepPath); os.IsNotExist(err) {
			t.Errorf(".gitkeep does not exist in %s", dir)
		}
	}
}
