# Phase 2: gRPC-REST Integration

**Duration**: Week 2-3
**Goal**: Implement dual-protocol API support with grpc-gateway, buf tooling, and PostgreSQL persistence
**Success Criteria**: Generated projects expose REST and gRPC APIs simultaneously with working database queries

---

## Objectives

1. Set up buf configuration for proto linting and code generation
2. Create proto file templates with API versioning (v1)
3. Implement grpc-gateway mux for REST-gRPC bridging
4. Generate PostgreSQL adapter with pgx + sqlc
5. Configure golang-migrate for database migrations
6. Implement example health check endpoint in both protocols
7. Validate dual API functionality

---

## Implementation Details

### 2.1 Extended Template Structure

```
templates/
├── base/
│   ├── api/
│   │   ├── proto/v1/
│   │   │   ├── health.proto.tmpl
│   │   │   └── common.proto.tmpl
│   │   ├── buf.yaml.tmpl
│   │   └── buf.gen.yaml.tmpl
│   ├── internal/
│   │   ├── ports/
│   │   │   ├── grpc/
│   │   │   │   ├── server.go.tmpl
│   │   │   │   └── health.go.tmpl
│   │   │   └── http/
│   │   │       └── gateway.go.tmpl      # grpc-gateway mux
│   │   └── adapters/
│   │       └── postgres/
│   │           ├── queries.sql.tmpl
│   │           ├── schema.sql.tmpl
│   │           └── sqlc.yaml.tmpl
│   ├── migrations/
│   │   └── 000001_initial_schema.up.sql.tmpl
│   │   └── 000001_initial_schema.down.sql.tmpl
│   └── scripts/
│       ├── generate-proto.sh.tmpl
│       └── migrate.sh.tmpl
```

### 2.2 Buf Configuration

**File**: `templates/base/api/buf.yaml.tmpl`

```yaml
version: v2
modules:
  - path: proto
lint:
  use:
    - STANDARD
    - COMMENTS
    - UNARY_RPC
  except:
    - PACKAGE_VERSION_SUFFIX  # Allow v1 suffix
breaking:
  use:
    - FILE
```

**File**: `templates/base/api/buf.gen.yaml.tmpl`

```yaml
version: v2
managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: {{ .ModulePath }}/internal/gen
plugins:
  # Go gRPC code generation
  - remote: buf.build/protocolbuffers/go
    out: internal/gen
    opt:
      - paths=source_relative

  # gRPC service stubs
  - remote: buf.build/grpc/go
    out: internal/gen
    opt:
      - paths=source_relative

  # gRPC-Gateway HTTP/JSON proxy
  - remote: buf.build/grpc-ecosystem/gateway
    out: internal/gen
    opt:
      - paths=source_relative
      - generate_unbound_methods=true

  # OpenAPI v2 documentation
  - remote: buf.build/grpc-ecosystem/openapiv2
    out: api/openapi
    opt:
      - allow_merge=true
      - merge_file_name=api
```

### 2.3 Proto File Templates

**File**: `templates/base/api/proto/v1/common.proto.tmpl`

```protobuf
syntax = "proto3";

package api.v1;

option go_package = "{{ .ModulePath }}/internal/gen/api/v1;apiv1";

import "google/protobuf/timestamp.proto";

// Common error response
message Error {
  string code = 1;
  string message = 2;
  map<string, string> details = 3;
}

// Pagination request
message PaginationRequest {
  int32 page = 1;      // Page number (1-indexed)
  int32 page_size = 2; // Items per page (max 100)
}

// Pagination response
message PaginationResponse {
  int32 page = 1;
  int32 page_size = 2;
  int32 total_pages = 3;
  int64 total_items = 4;
}
```

**File**: `templates/base/api/proto/v1/health.proto.tmpl`

```protobuf
syntax = "proto3";

package api.v1;

option go_package = "{{ .ModulePath }}/internal/gen/api/v1;apiv1";

import "google/api/annotations.proto";

// Health check service
service HealthService {
  // Check service health
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse) {
    option (google.api.http) = {
      get: "/api/v1/health"
    };
  }

  // Check readiness (database connectivity)
  rpc Ready(ReadyRequest) returns (ReadyResponse) {
    option (google.api.http) = {
      get: "/api/v1/ready"
    };
  }
}

message HealthCheckRequest {}

message HealthCheckResponse {
  enum Status {
    UNKNOWN = 0;
    HEALTHY = 1;
    UNHEALTHY = 2;
  }

  Status status = 1;
  string version = 2;
  int64 uptime_seconds = 3;
}

message ReadyRequest {}

message ReadyResponse {
  bool ready = 1;
  map<string, string> checks = 2; // Component -> status
}
```

### 2.4 gRPC Server Implementation

**File**: `templates/base/internal/ports/grpc/server.go.tmpl`

```go
package grpc

import (
    "fmt"
    "net"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    {{- if .EnableObservability }}
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    {{- end }}

    apiv1 "{{ .ModulePath }}/internal/gen/api/v1"
    "{{ .ModulePath }}/internal/config"
)

type Server struct {
    config     *config.Config
    grpcServer *grpc.Server
}

func NewServer(cfg *config.Config) *Server {
    opts := []grpc.ServerOption{
        {{- if .EnableObservability }}
        grpc.StatsHandler(otelgrpc.NewServerHandler()),
        {{- end }}
    }

    grpcServer := grpc.NewServer(opts...)

    // Register services
    apiv1.RegisterHealthServiceServer(grpcServer, &HealthServiceServer{})

    // Enable reflection for grpcurl
    reflection.Register(grpcServer)

    return &Server{
        config:     cfg,
        grpcServer: grpcServer,
    }
}

func (s *Server) Start() error {
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
    if err != nil {
        return fmt.Errorf("failed to listen: %w", err)
    }

    return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
    s.grpcServer.GracefulStop()
}
```

**File**: `templates/base/internal/ports/grpc/health.go.tmpl`

```go
package grpc

import (
    "context"
    "time"

    apiv1 "{{ .ModulePath }}/internal/gen/api/v1"
)

type HealthServiceServer struct {
    apiv1.UnimplementedHealthServiceServer
    startTime time.Time
}

func NewHealthServiceServer() *HealthServiceServer {
    return &HealthServiceServer{
        startTime: time.Now(),
    }
}

func (s *HealthServiceServer) Check(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckResponse, error) {
    return &apiv1.HealthCheckResponse{
        Status:        apiv1.HealthCheckResponse_HEALTHY,
        Version:       "1.0.0", // TODO: inject from build
        UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
    }, nil
}

func (s *HealthServiceServer) Ready(ctx context.Context, req *apiv1.ReadyRequest) (*apiv1.ReadyResponse, error) {
    checks := make(map[string]string)
    checks["database"] = "ok" // TODO: implement real check

    return &apiv1.ReadyResponse{
        Ready:  true,
        Checks: checks,
    }, nil
}
```

### 2.5 gRPC-Gateway Integration

**File**: `templates/base/internal/ports/http/gateway.go.tmpl`

```go
package http

import (
    "context"
    "fmt"
    "net/http"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    apiv1 "{{ .ModulePath }}/internal/gen/api/v1"
    "{{ .ModulePath }}/internal/config"
)

type GatewayServer struct {
    config *config.Config
    mux    *runtime.ServeMux
}

func NewGatewayServer(cfg *config.Config) *GatewayServer {
    // Custom JSON marshaler options
    mux := runtime.NewServeMux(
        runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
            MarshalOptions: protojson.MarshalOptions{
                UseProtoNames:   true,  // Use snake_case field names
                EmitUnpopulated: false, // Omit zero values
            },
            UnmarshalOptions: protojson.UnmarshalOptions{
                DiscardUnknown: true, // Ignore unknown fields
            },
        }),
    )

    return &GatewayServer{
        config: cfg,
        mux:    mux,
    }
}

func (s *GatewayServer) RegisterHandlers(ctx context.Context) error {
    grpcEndpoint := fmt.Sprintf("localhost:%d", s.config.GRPCPort)
    opts := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    }

    // Register all service handlers
    if err := apiv1.RegisterHealthServiceHandlerFromEndpoint(ctx, s.mux, grpcEndpoint, opts); err != nil {
        return fmt.Errorf("failed to register health service: %w", err)
    }

    return nil
}

func (s *GatewayServer) Handler() http.Handler {
    return s.mux
}
```

**File**: `templates/base/internal/ports/http/server.go.tmpl` (updated)

```go
package http

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "{{ .ModulePath }}/internal/config"
)

type Server struct {
    config  *config.Config
    server  *http.Server
    gateway *GatewayServer
}

func NewServer(cfg *config.Config) *Server {
    gateway := NewGatewayServer(cfg)

    mux := http.NewServeMux()

    // Register gRPC-Gateway handlers
    ctx := context.Background()
    if err := gateway.RegisterHandlers(ctx); err != nil {
        panic(fmt.Sprintf("Failed to register gateway handlers: %v", err))
    }

    mux.Handle("/api/", gateway.Handler())

    // Additional HTTP-only endpoints
    mux.HandleFunc("/health", handleHealth)
    {{- if .EnableObservability }}
    mux.Handle("/metrics", promhttp.Handler())
    {{- end }}

    return &Server{
        config:  cfg,
        gateway: gateway,
        server: &http.Server{
            Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
            Handler:      mux,
            ReadTimeout:  15 * time.Second,
            WriteTimeout: 15 * time.Second,
            IdleTimeout:  60 * time.Second,
        },
    }
}

func (s *Server) Start() error {
    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.server.Shutdown(ctx)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### 2.6 PostgreSQL with sqlc

**File**: `templates/base/internal/adapters/postgres/sqlc.yaml.tmpl`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries.sql"
    schema: "schema.sql"
    gen:
      go:
        package: "postgres"
        out: "."
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
```

**File**: `templates/base/internal/adapters/postgres/schema.sql.tmpl`

```sql
-- Health check table
CREATE TABLE IF NOT EXISTS health_checks (
    id SERIAL PRIMARY KEY,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL
);
```

**File**: `templates/base/internal/adapters/postgres/queries.sql.tmpl`

```sql
-- name: GetLatestHealthCheck :one
SELECT * FROM health_checks
ORDER BY checked_at DESC
LIMIT 1;

-- name: CreateHealthCheck :one
INSERT INTO health_checks (status)
VALUES ($1)
RETURNING *;
```

**File**: `templates/base/internal/adapters/postgres/adapter.go.tmpl`

```go
package postgres

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"

    "{{ .ModulePath }}/internal/config"
)

type Adapter struct {
    pool *pgxpool.Pool
    *Queries // Embed generated queries
}

func NewAdapter(cfg *config.Config) (*Adapter, error) {
    ctx := context.Background()

    poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse database URL: %w", err)
    }

    // Connection pool settings
    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

    // Verify connection
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &Adapter{
        pool:    pool,
        Queries: New(pool),
    }, nil
}

func (a *Adapter) Close() {
    a.pool.Close()
}

func (a *Adapter) Pool() *pgxpool.Pool {
    return a.pool
}
```

### 2.7 Database Migrations

**File**: `templates/base/migrations/000001_initial_schema.up.sql.tmpl`

```sql
-- Initial schema for {{ .ProjectName }}

CREATE TABLE IF NOT EXISTS health_checks (
    id SERIAL PRIMARY KEY,
    checked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    status VARCHAR(50) NOT NULL
);

CREATE INDEX idx_health_checks_checked_at ON health_checks(checked_at DESC);
```

**File**: `templates/base/migrations/000001_initial_schema.down.sql.tmpl`

```sql
DROP TABLE IF EXISTS health_checks;
```

**File**: `templates/base/scripts/migrate.sh.tmpl`

```bash
#!/bin/bash
set -e

# Migration script for {{ .ProjectName }}

MIGRATIONS_PATH="./migrations"
DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/{{ .ProjectName }}?sslmode=disable}"

case "$1" in
  up)
    echo "Running migrations up..."
    migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
    ;;
  down)
    echo "Running migrations down..."
    migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down
    ;;
  force)
    if [ -z "$2" ]; then
      echo "Usage: $0 force <version>"
      exit 1
    fi
    migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$2"
    ;;
  version)
    migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
    ;;
  *)
    echo "Usage: $0 {up|down|force|version}"
    exit 1
esac
```

### 2.8 Updated Configuration

**File**: `templates/base/internal/config/config.go.tmpl` (enhanced)

```go
package config

import (
    "fmt"

    "github.com/spf13/viper"
)

type Config struct {
    Environment string
    HTTPPort    int
    GRPCPort    int
    DatabaseURL string
    LogLevel    string
    {{- if .EnableAuth }}
    JWTSecret      string
    RedisURL       string
    OAuth2Google   OAuth2Config
    OAuth2GitHub   OAuth2Config
    {{- end }}
}

{{- if .EnableAuth }}
type OAuth2Config struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}
{{- end }}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./configs")
    viper.AddConfigPath(".")

    // Environment variables override
    viper.SetEnvPrefix("APP")
    viper.AutomaticEnv()

    // Defaults
    viper.SetDefault("environment", "development")
    viper.SetDefault("http_port", 8080)
    viper.SetDefault("grpc_port", 9090)
    viper.SetDefault("log_level", "info")

    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &cfg, nil
}
```

**File**: `templates/base/configs/config.dev.yaml.tmpl`

```yaml
environment: development
http_port: 8080
grpc_port: 9090
database_url: "postgres://postgres:postgres@localhost:5432/{{ .ProjectName }}?sslmode=disable"
log_level: debug

{{- if .EnableAuth }}
jwt_secret: "dev-secret-change-in-production"
redis_url: "redis://localhost:6379/0"

oauth2_google:
  client_id: "your-google-client-id"
  client_secret: "your-google-client-secret"
  redirect_url: "http://localhost:8080/api/v1/auth/google/callback"

oauth2_github:
  client_id: "your-github-client-id"
  client_secret: "your-github-client-secret"
  redirect_url: "http://localhost:8080/api/v1/auth/github/callback"
{{- end }}
```

### 2.9 Updated Makefile

**File**: `templates/base/Makefile.tmpl` (enhanced)

```makefile
.PHONY: help proto sqlc migrate-up migrate-down run build test lint clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate code from proto files
	buf generate api

sqlc: ## Generate code from SQL queries
	sqlc generate -f internal/adapters/postgres/sqlc.yaml

generate: proto sqlc ## Run all code generation

migrate-up: ## Run database migrations up
	./scripts/migrate.sh up

migrate-down: ## Run database migrations down
	./scripts/migrate.sh down

migrate-create: ## Create new migration (usage: make migrate-create name=add_users)
	migrate create -ext sql -dir migrations -seq $(name)

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/{{ .ProjectName }} cmd/server/main.go

test: ## Run tests
	go test -v -race -cover ./...

test-integration: ## Run integration tests
	go test -v -race -tags=integration ./tests/integration/...

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf internal/gen/

deps: ## Download dependencies
	go mod download
	go mod tidy

tools: ## Install required tools
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

{{- if .EnableDocker }}

docker-build: ## Build Docker image
	docker build -t {{ .ProjectName }}:latest .

docker-up: ## Start Docker compose
	docker-compose up -d

docker-down: ## Stop Docker compose
	docker-compose down

docker-logs: ## Show Docker compose logs
	docker-compose logs -f
{{- end }}
```

---

## Testing Strategy

### 2.10 Integration Tests for Dual Protocol

**File**: `tests/integration/api_test.go` (generated)

```go
//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    apiv1 "{{ .ModulePath }}/internal/gen/api/v1"
)

func TestHealthCheck_gRPC(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    postgresC, err := startPostgres(ctx)
    require.NoError(t, err)
    defer postgresC.Terminate(ctx)

    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)
    defer conn.Close()

    client := apiv1.NewHealthServiceClient(conn)

    // Test health check
    resp, err := client.Check(ctx, &apiv1.HealthCheckRequest{})
    require.NoError(t, err)
    assert.Equal(t, apiv1.HealthCheckResponse_HEALTHY, resp.Status)
    assert.Greater(t, resp.UptimeSeconds, int64(0))
}

func TestHealthCheck_REST(t *testing.T) {
    ctx := context.Background()

    // Test REST endpoint via grpc-gateway
    resp, err := http.Get("http://localhost:8080/api/v1/health")
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)

    var health apiv1.HealthCheckResponse
    err = json.NewDecoder(resp.Body).Decode(&health)
    require.NoError(t, err)
    assert.Equal(t, "HEALTHY", health.Status)
}

func startPostgres(ctx context.Context) (testcontainers.Container, error) {
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_USER":     "postgres",
            "POSTGRES_PASSWORD": "postgres",
            "POSTGRES_DB":       "testdb",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections").
            WithOccurrence(2).
            WithStartupTimeout(30 * time.Second),
    }

    return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
}
```

### 2.11 Golden File Tests for Proto Generation

```go
// internal/generator/proto_test.go
func TestProtoGeneration(t *testing.T) {
    g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))

    cfg := &config.Config{
        ProjectName: "test-api",
        ModulePath:  "github.com/test/api",
    }

    gen := New()
    content, err := gen.engine.Render("base/api/proto/v1/health.proto.tmpl", cfg.TemplateData())
    require.NoError(t, err)

    g.Assert(t, "health.proto", content)
}
```

---

## Anti-Patterns to Avoid

1. **No Mixed Serialization**: Don't use different JSON marshalers for REST vs gRPC
2. **No Reflection in Hot Path**: Pre-register all services, no runtime discovery
3. **No Blocking in Handlers**: Always respect context deadlines
4. **No Connection Leaks**: Always defer Close() on gRPC connections
5. **No Manual Proto Compilation**: Always use buf, never protoc directly
6. **No Database in Main**: Pass adapter via DI, not global variables

---

## Success Validation Checklist

- [ ] buf.yaml and buf.gen.yaml validate successfully
- [ ] Proto files compile without errors
- [ ] gRPC server starts and accepts connections
- [ ] grpc-gateway correctly proxies REST to gRPC
- [ ] Health check works via both protocols
- [ ] sqlc generates valid Go code from queries
- [ ] Database migrations run successfully (up and down)
- [ ] PostgreSQL connection pool initializes correctly
- [ ] OpenAPI spec generates from proto files
- [ ] Integration tests pass with testcontainers

---

## Deliverables

1. Complete buf tooling setup (buf.yaml, buf.gen.yaml)
2. Proto templates with API versioning (v1)
3. Dual-protocol server (gRPC + REST via grpc-gateway)
4. PostgreSQL adapter with pgx and sqlc
5. Migration system with golang-migrate
6. Working health check endpoint (REST + gRPC)
7. Integration test suite with testcontainers
8. OpenAPI documentation generation

**Next Phase**: Phase 3 - Feature Toggles (Auth, Observability, Docker)
