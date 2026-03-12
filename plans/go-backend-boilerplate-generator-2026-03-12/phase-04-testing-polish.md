# Phase 4: Testing, Polish & Documentation

**Duration**: Week 4-5
**Goal**: Comprehensive testing, example implementations, CI/CD, and production-ready documentation
**Success Criteria**: 100% test pass rate, no lint errors, complete documentation, ready for v1.0 release

---

## Objectives

1. Implement golden file tests for all templates
2. Create integration test suite with testcontainers
3. Generate example CRUD implementation (User resource)
4. Set up CI/CD pipeline for generator (GitHub Actions)
5. Generate comprehensive README with setup instructions
6. Create documentation site structure
7. Version and release v1.0.0

---

## Implementation Details

### 4.1 Golden File Testing

#### 4.1.1 Test Structure

```
tests/
├── golden/
│   ├── testdata/
│   │   ├── base-only/              # Expected output for base-only
│   │   │   ├── cmd/
│   │   │   ├── internal/
│   │   │   ├── go.mod
│   │   │   └── Makefile
│   │   ├── all-features/           # Expected output for all features
│   │   │   └── ...
│   │   └── ...
│   └── golden_test.go
├── integration/
│   ├── api_test.go
│   ├── auth_test.go
│   └── db_test.go
└── e2e/
    └── workflow_test.go
```

#### 4.1.2 Golden File Test Implementation

**File**: `tests/golden/golden_test.go`

```go
package golden

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/sebdah/goldie/v2"
    "github.com/stretchr/testify/require"

    "github.com/yourorg/go-boilerplate/internal/config"
    "github.com/yourorg/go-boilerplate/internal/generator"
)

func TestGoldenFiles_BaseOnly(t *testing.T) {
    g := goldie.New(t,
        goldie.WithFixtureDir("testdata/base-only"),
        goldie.WithNameSuffix(".golden"),
    )

    cfg := &config.Config{
        ProjectName: "test-base",
        ModulePath:  "github.com/test/base",
        EnableAuth:  false,
        EnableObservability: false,
        EnableDocker: false,
        OutputDir:   t.TempDir(),
    }

    gen := generator.New()
    err := gen.Generate(cfg)
    require.NoError(t, err)

    // Verify each generated file against golden files
    filesToTest := []string{
        "cmd/server/main.go",
        "internal/config/config.go",
        "internal/ports/http/server.go",
        "internal/ports/grpc/server.go",
        "go.mod",
        "Makefile",
        "README.md",
    }

    for _, file := range filesToTest {
        path := filepath.Join(cfg.OutputDir, file)
        content, err := os.ReadFile(path)
        require.NoError(t, err)

        g.Assert(t, filepath.Base(file), content)
    }
}

func TestGoldenFiles_AllFeatures(t *testing.T) {
    g := goldie.New(t,
        goldie.WithFixtureDir("testdata/all-features"),
        goldie.WithNameSuffix(".golden"),
    )

    cfg := &config.Config{
        ProjectName: "test-full",
        ModulePath:  "github.com/test/full",
        EnableAuth:  true,
        EnableObservability: true,
        EnableDocker: true,
        GenerateExample: true,
        OutputDir:   t.TempDir(),
    }

    gen := generator.New()
    err := gen.Generate(cfg)
    require.NoError(t, err)

    filesToTest := []string{
        "cmd/server/main.go",
        "pkg/jwt/jwt.go",
        "internal/telemetry/metrics.go",
        "Dockerfile",
        "docker-compose.yml",
    }

    for _, file := range filesToTest {
        path := filepath.Join(cfg.OutputDir, file)
        content, err := os.ReadFile(path)
        require.NoError(t, err)

        g.Assert(t, filepath.Base(file), content)
    }
}

// Test template rendering with different data
func TestTemplateRendering(t *testing.T) {
    tests := []struct {
        name     string
        template string
        data     map[string]interface{}
        expected string
    }{
        {
            name:     "project-name-substitution",
            template: "base/go.mod.tmpl",
            data: map[string]interface{}{
                "ModulePath": "github.com/example/myapi",
            },
            expected: "module github.com/example/myapi",
        },
        {
            name:     "conditional-auth-import",
            template: "base/cmd/server/main.go.tmpl",
            data: map[string]interface{}{
                "EnableAuth": true,
                "ModulePath": "github.com/example/api",
            },
            expected: `"github.com/example/api/pkg/jwt"`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := templates.NewEngine()
            err := engine.LoadTemplates()
            require.NoError(t, err)

            content, err := engine.Render(tt.template, tt.data)
            require.NoError(t, err)
            require.Contains(t, string(content), tt.expected)
        })
    }
}
```

### 4.2 Integration Tests

#### 4.2.1 Full Stack Integration Test

**File**: `tests/integration/fullstack_test.go`

```go
//go:build integration
// +build integration

package integration

import (
    "context"
    "net/http"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

type TestEnvironment struct {
    PostgresContainer testcontainers.Container
    RedisContainer    testcontainers.Container
    AppContainer      testcontainers.Container
}

func setupTestEnvironment(t *testing.T, features map[string]bool) *TestEnvironment {
    ctx := context.Background()
    env := &TestEnvironment{}

    // Start PostgreSQL
    postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15-alpine",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_USER":     "postgres",
                "POSTGRES_PASSWORD": "postgres",
                "POSTGRES_DB":       "testdb",
            },
            WaitingFor: wait.ForLog("database system is ready").
                WithOccurrence(2).
                WithStartupTimeout(30 * time.Second),
        },
        Started: true,
    })
    require.NoError(t, err)
    env.PostgresContainer = postgresC

    // Start Redis if auth enabled
    if features["auth"] {
        redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image:        "redis:7-alpine",
                ExposedPorts: []string{"6379/tcp"},
                WaitingFor:   wait.ForLog("Ready to accept connections"),
            },
            Started: true,
        })
        require.NoError(t, err)
        env.RedisContainer = redisC
    }

    return env
}

func (e *TestEnvironment) Cleanup() {
    ctx := context.Background()
    if e.PostgresContainer != nil {
        e.PostgresContainer.Terminate(ctx)
    }
    if e.RedisContainer != nil {
        e.RedisContainer.Terminate(ctx)
    }
    if e.AppContainer != nil {
        e.AppContainer.Terminate(ctx)
    }
}

func TestGeneratedProject_BaseOnly(t *testing.T) {
    // Generate project
    tmpDir := t.TempDir()
    cfg := &config.Config{
        ProjectName: "integration-test",
        ModulePath:  "github.com/test/integration",
        OutputDir:   tmpDir,
    }

    gen := generator.New()
    err := gen.Generate(cfg)
    require.NoError(t, err)

    // Setup test environment
    env := setupTestEnvironment(t, map[string]bool{})
    defer env.Cleanup()

    // Build generated project
    buildCmd := exec.Command("go", "build", "-o", "app", "./cmd/server")
    buildCmd.Dir = tmpDir
    require.NoError(t, buildCmd.Run())

    // Run application
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    runCmd := exec.CommandContext(ctx, "./app")
    runCmd.Dir = tmpDir
    runCmd.Env = append(os.Environ(),
        fmt.Sprintf("APP_DATABASE_URL=%s", getDatabaseURL(env.PostgresContainer)),
    )

    require.NoError(t, runCmd.Start())
    defer runCmd.Process.Kill()

    // Wait for server to start
    time.Sleep(2 * time.Second)

    // Test health endpoint
    resp, err := http.Get("http://localhost:8080/api/v1/health")
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGeneratedProject_WithAuth(t *testing.T) {
    tmpDir := t.TempDir()
    cfg := &config.Config{
        ProjectName: "integration-auth",
        ModulePath:  "github.com/test/auth",
        EnableAuth:  true,
        OutputDir:   tmpDir,
    }

    gen := generator.New()
    err := gen.Generate(cfg)
    require.NoError(t, err)

    env := setupTestEnvironment(t, map[string]bool{"auth": true})
    defer env.Cleanup()

    // Build and run
    buildCmd := exec.Command("go", "build", "-o", "app", "./cmd/server")
    buildCmd.Dir = tmpDir
    require.NoError(t, buildCmd.Run())

    // Test JWT generation and validation
    // (implementation details...)
}

func getDatabaseURL(container testcontainers.Container) string {
    ctx := context.Background()
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")
    return fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", host, port.Port())
}
```

### 4.3 Example CRUD Implementation

#### 4.3.1 User Resource Template

**File**: `templates/examples/internal/domain/user_example.go.tmpl`

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

// Example User entity
type User struct {
    ID        uuid.UUID
    Email     string
    FirstName string
    LastName  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Business validation
func (u *User) Validate() error {
    if u.Email == "" {
        return ErrInvalidEmail
    }
    if u.FirstName == "" {
        return ErrFirstNameRequired
    }
    return nil
}

var (
    ErrInvalidEmail       = errors.New("invalid email address")
    ErrFirstNameRequired  = errors.New("first name is required")
)
```

**File**: `templates/examples/api/proto/v1/user.proto.tmpl`

```protobuf
syntax = "proto3";

package api.v1;

option go_package = "{{ .ModulePath }}/internal/gen/api/v1;apiv1";

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "api/proto/v1/common.proto";

service UserService {
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
        option (google.api.http) = {
            post: "/api/v1/users"
            body: "*"
        };
    }

    rpc GetUser(GetUserRequest) returns (GetUserResponse) {
        option (google.api.http) = {
            get: "/api/v1/users/{id}"
        };
    }

    rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
        option (google.api.http) = {
            put: "/api/v1/users/{id}"
            body: "*"
        };
    }

    rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
        option (google.api.http) = {
            delete: "/api/v1/users/{id}"
        };
    }

    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse) {
        option (google.api.http) = {
            get: "/api/v1/users"
        };
    }
}

message User {
    string id = 1;
    string email = 2;
    string first_name = 3;
    string last_name = 4;
    google.protobuf.Timestamp created_at = 5;
    google.protobuf.Timestamp updated_at = 6;
}

message CreateUserRequest {
    string email = 1;
    string first_name = 2;
    string last_name = 3;
}

message CreateUserResponse {
    User user = 1;
}

message GetUserRequest {
    string id = 1;
}

message GetUserResponse {
    User user = 1;
}

message UpdateUserRequest {
    string id = 1;
    string email = 2;
    string first_name = 3;
    string last_name = 4;
}

message UpdateUserResponse {
    User user = 1;
}

message DeleteUserRequest {
    string id = 1;
}

message DeleteUserResponse {
    bool success = 1;
}

message ListUsersRequest {
    PaginationRequest pagination = 1;
}

message ListUsersResponse {
    repeated User users = 1;
    PaginationResponse pagination = 2;
}
```

**File**: `templates/examples/internal/adapters/postgres/user_queries.sql.tmpl`

```sql
-- name: CreateUser :one
INSERT INTO users (email, first_name, last_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $2, first_name = $3, last_name = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;
```

### 4.4 README Generator

**File**: `templates/base/README.md.tmpl`

```markdown
# {{ .ProjectName }}

Production-ready Go backend with clean architecture, dual protocol support (REST + gRPC), and PostgreSQL persistence.

## Features

- ✅ REST + gRPC APIs via grpc-gateway
- ✅ PostgreSQL with pgx and sqlc
- ✅ Database migrations with golang-migrate
- ✅ Configuration management with Viper
- ✅ Structured logging
{{- if .EnableAuth }}
- ✅ JWT authentication
- ✅ OAuth2 integration (Google, GitHub)
- ✅ Redis session store
- ✅ Role-based access control (RBAC)
{{- end }}
{{- if .EnableObservability }}
- ✅ Prometheus metrics
- ✅ OpenTelemetry tracing
- ✅ Structured logging with slog
{{- end }}
{{- if .EnableDocker }}
- ✅ Docker support
- ✅ Docker Compose for local development
{{- end }}

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
{{- if .EnableAuth }}
- Redis 7+
{{- end }}
- buf CLI
- sqlc
- golang-migrate

Install tools:

\`\`\`bash
make tools
\`\`\`

### Setup

1. **Clone and setup**:

\`\`\`bash
cd {{ .ProjectName }}
cp configs/config.dev.yaml configs/config.yaml
# Edit configs/config.yaml with your settings
\`\`\`

2. **Start dependencies** {{- if .EnableDocker }}(using Docker){{- end }}:

{{- if .EnableDocker }}
\`\`\`bash
docker-compose up -d
\`\`\`
{{- else }}
\`\`\`bash
# Start PostgreSQL on localhost:5432
{{- if .EnableAuth }}
# Start Redis on localhost:6379
{{- end }}
\`\`\`
{{- end }}

3. **Run migrations**:

\`\`\`bash
make migrate-up
\`\`\`

4. **Generate code**:

\`\`\`bash
make generate
\`\`\`

5. **Run application**:

\`\`\`bash
make run
\`\`\`

Server will start on:
- REST API: http://localhost:8080/api/v1/
- gRPC: localhost:9090
{{- if .EnableObservability }}
- Metrics: http://localhost:8080/metrics
{{- end }}
- Health: http://localhost:8080/api/v1/health

### Testing

Run all tests:

\`\`\`bash
make test
\`\`\`

Run integration tests:

\`\`\`bash
make test-integration
\`\`\`

### API Documentation

OpenAPI documentation is available at:
- OpenAPI v2 spec: \`api/openapi/api.swagger.json\`

Use grpcurl for gRPC exploration:

\`\`\`bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 api.v1.HealthService/Check
\`\`\`

## Project Structure

\`\`\`
{{ .ProjectName }}/
├── cmd/
│   └── server/              # Application entry point
├── internal/
│   ├── domain/              # Business entities
│   ├── application/         # Use cases
│   ├── ports/               # Interface adapters
│   │   ├── http/            # REST handlers
│   │   └── grpc/            # gRPC service implementations
│   ├── adapters/            # External integrations
│   │   └── postgres/        # Database layer
│   └── config/              # Configuration
├── pkg/                     # Reusable packages
├── api/
│   └── proto/v1/            # Protocol buffers
├── migrations/              # SQL migrations
{{- if .EnableDocker }}
├── docker/                  # Docker files
{{- end }}
└── tests/                   # Test files
\`\`\`

## Development

### Adding a new endpoint

1. Define proto file in \`api/proto/v1/\`:

\`\`\`protobuf
service MyService {
  rpc MyMethod(MyRequest) returns (MyResponse) {
    option (google.api.http) = {
      post: "/api/v1/my-endpoint"
      body: "*"
    };
  }
}
\`\`\`

2. Generate code:

\`\`\`bash
make proto
\`\`\`

3. Implement service in \`internal/ports/grpc/\`

4. Business logic goes in \`internal/application/\`

5. Database queries in \`internal/adapters/postgres/queries.sql\` + \`make sqlc\`

### Database migrations

Create new migration:

\`\`\`bash
make migrate-create name=add_new_table
\`\`\`

Apply migrations:

\`\`\`bash
make migrate-up
\`\`\`

Rollback:

\`\`\`bash
make migrate-down
\`\`\`

### Code quality

Run linter:

\`\`\`bash
make lint
\`\`\`

Format code:

\`\`\`bash
go fmt ./...
\`\`\`

## Configuration

Configuration is managed via YAML files and environment variables.

Priority (highest to lowest):
1. Environment variables (\`APP_*\`)
2. \`configs/config.yaml\`
3. Defaults

Example environment variables:

\`\`\`bash
export APP_ENVIRONMENT=production
export APP_DATABASE_URL="postgres://user:pass@localhost:5432/dbname"
{{- if .EnableAuth }}
export APP_JWT_SECRET="your-secret-key"
export APP_REDIS_URL="redis://localhost:6379/0"
{{- end }}
\`\`\`

## Deployment

{{- if .EnableDocker }}

Build Docker image:

\`\`\`bash
make docker-build
\`\`\`

Run with Docker:

\`\`\`bash
docker run -p 8080:8080 -p 9090:9090 \\
  -e APP_DATABASE_URL="..." \\
  {{ .ProjectName }}:latest
\`\`\`

{{- else }}

Build binary:

\`\`\`bash
make build
\`\`\`

Run:

\`\`\`bash
./bin/{{ .ProjectName }}
\`\`\`

{{- end }}

## Architecture Decisions

See \`docs/architecture/\` for detailed architecture decision records (ADRs).

## Contributing

1. Fork the repository
2. Create feature branch (\`git checkout -b feature/amazing-feature\`)
3. Commit changes (\`git commit -m 'Add amazing feature'\`)
4. Push to branch (\`git push origin feature/amazing-feature\`)
5. Open Pull Request

## License

MIT License - see LICENSE file for details.

---

Generated with [go-boilerplate](https://github.com/yourorg/go-boilerplate)
\`\`\`

### 4.5 CI/CD Pipeline

**File**: `.github/workflows/ci.yml` (for generator project)

```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install tools
        run: |
          go install github.com/bufbuild/buf/cmd/buf@latest
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out

      - name: Run linter
        run: golangci-lint run --timeout=5m

  integration:
    name: Integration Tests
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run integration tests
        run: go test -v -tags=integration ./tests/integration/...
        env:
          DATABASE_URL: postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379/0

  generate-matrix:
    name: Test All Feature Combinations
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Install tools
        run: |
          go install github.com/bufbuild/buf/cmd/buf@latest
          go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

      - name: Test all combinations
        run: |
          ./scripts/test-all-combinations.sh

  golangci:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m
```

**File**: `scripts/test-all-combinations.sh`

```bash
#!/bin/bash
set -e

# Test all 8 feature combinations

combinations=(
  "false,false,false"  # base
  "true,false,false"   # auth
  "false,true,false"   # observability
  "false,false,true"   # docker
  "true,true,false"    # auth+observability
  "true,false,true"    # auth+docker
  "false,true,true"    # observability+docker
  "true,true,true"     # all
)

for combo in "${combinations[@]}"; do
  IFS=',' read -r auth obs docker <<< "$combo"

  echo "Testing combination: auth=$auth, observability=$obs, docker=$docker"

  tmpdir=$(mktemp -d)

  ./bin/go-boilerplate init \
    --name="test-$auth-$obs-$docker" \
    --module="github.com/test/combo" \
    --yes \
    $([ "$auth" = "true" ] && echo "--features=auth" || echo "") \
    $([ "$obs" = "true" ] && echo "--features=observability" || echo "") \
    $([ "$docker" = "true" ] && echo "--features=docker" || echo "")

  cd "$tmpdir"

  # Try to build
  go mod download
  go build ./cmd/server

  # Run linter
  golangci-lint run

  cd -
  rm -rf "$tmpdir"

  echo "✅ Combination passed: auth=$auth, observability=$obs, docker=$docker"
done

echo "✅ All combinations passed!"
```

### 4.6 Documentation Site Structure

**File**: `docs/index.md`

```markdown
# Go Backend Boilerplate Generator

Production-ready Go backend scaffolding tool with clean architecture.

## Quick Links

- [Getting Started](getting-started.md)
- [Features](features.md)
- [Architecture](architecture/overview.md)
- [API Examples](examples/api.md)
- [Configuration](configuration.md)
- [Deployment](deployment.md)

## What is this?

An opinionated CLI generator for Go backend projects featuring:

- Clean 4-layer architecture
- Dual protocol support (REST + gRPC)
- Type-safe database queries
- Optional feature toggles
- Production-ready defaults

## Philosophy

1. **YAGNI**: Generate only what you need
2. **KISS**: Simple, readable code over clever abstractions
3. **DRY**: Reusable packages, no duplication
4. **Convention over Configuration**: Sensible defaults

## Installation

\`\`\`bash
go install github.com/yourorg/go-boilerplate@latest
\`\`\`

## Usage

\`\`\`bash
go-boilerplate init
\`\`\`

Follow the interactive prompts or use flags for non-interactive mode.
```

---

## Anti-Patterns to Avoid

1. **No Flaky Tests**: All tests must be deterministic, no random timeouts
2. **No Test Data Pollution**: Each test uses isolated temp directories
3. **No Golden File Drift**: CI must fail on golden file mismatches
4. **No Undocumented Features**: Every feature must have examples
5. **No Breaking Changes Without Migration**: Always provide upgrade path
6. **No Platform-Specific Code**: Tests must pass on Linux, macOS, Windows

---

## Success Validation Checklist

- [ ] All unit tests pass (coverage > 80%)
- [ ] All integration tests pass with real containers
- [ ] All 8 feature combinations generate and compile
- [ ] Golden file tests validate template output
- [ ] golangci-lint passes with strict preset
- [ ] Example CRUD implementation included
- [ ] README generation works for all feature combos
- [ ] CI/CD pipeline green on GitHub Actions
- [ ] Documentation complete and accessible
- [ ] Version tagged (v1.0.0)
- [ ] Release notes published
- [ ] Migration guide available

---

## Release Process

### 4.7 Versioning

Follow semantic versioning (semver):

- **MAJOR**: Breaking changes to CLI or generated code structure
- **MINOR**: New features, backward-compatible
- **PATCH**: Bug fixes, dependency updates

### 4.8 Release Checklist

1. **Pre-release**:
   - [ ] All tests pass
   - [ ] Documentation updated
   - [ ] CHANGELOG.md updated
   - [ ] Version bumped in code

2. **Release**:
   - [ ] Tag version: `git tag v1.0.0`
   - [ ] Push tag: `git push origin v1.0.0`
   - [ ] GitHub Release created with notes
   - [ ] Binaries built for Linux/macOS/Windows

3. **Post-release**:
   - [ ] Announcement in docs
   - [ ] Example projects updated
   - [ ] Migration guide published (if breaking)

### 4.9 Performance Benchmarks

```go
// tests/benchmark/generator_bench_test.go
func BenchmarkGenerate_BaseOnly(b *testing.B) {
    cfg := &config.Config{
        ProjectName: "bench-base",
        ModulePath:  "github.com/bench/base",
        OutputDir:   b.TempDir(),
    }

    gen := generator.New()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if err := gen.Generate(cfg); err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGenerate_AllFeatures(b *testing.B) {
    cfg := &config.Config{
        ProjectName:         "bench-full",
        ModulePath:          "github.com/bench/full",
        EnableAuth:          true,
        EnableObservability: true,
        EnableDocker:        true,
        GenerateExample:     true,
        OutputDir:           b.TempDir(),
    }

    gen := generator.New()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        if err := gen.Generate(cfg); err != nil {
            b.Fatal(err)
        }
    }
}
```

**Performance Targets**:
- Base generation: < 1 second
- All features generation: < 3 seconds
- Template rendering: < 100ms per file

---

## Deliverables

1. Complete golden file test suite (all templates validated)
2. Integration test suite with testcontainers (all services tested)
3. Example CRUD implementation (User resource)
4. CI/CD pipeline (GitHub Actions with all checks)
5. Comprehensive README generator (feature-aware)
6. Documentation site structure (getting started, examples, architecture)
7. Release automation scripts
8. Performance benchmarks
9. v1.0.0 release ready

---

## Success Metrics

- ✅ 100% test pass rate in CI
- ✅ Zero golangci-lint errors
- ✅ Test coverage > 80%
- ✅ All 8 feature combinations compile and run
- ✅ Golden files match generated output
- ✅ Integration tests pass with real databases
- ✅ Documentation complete
- ✅ Ready for production use

**Project Complete**: Ready for v1.0.0 release!
