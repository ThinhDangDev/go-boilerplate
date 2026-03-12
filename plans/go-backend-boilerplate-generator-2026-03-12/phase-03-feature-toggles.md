---
phase: 3
status: completed
completed_date: 2026-03-12
---

# Phase 3: Feature Toggles Implementation

**Duration**: Week 3-4
**Goal**: Implement conditional code generation for auth, observability, and Docker features
**Success Criteria**: All 8 feature combinations generate valid, tested code

---

## Objectives

1. Implement authentication feature (JWT + OAuth2 + Redis + RBAC)
2. Implement observability feature (Prometheus + OpenTelemetry + slog)
3. Implement Docker feature (Dockerfile + docker-compose.yml)
4. Create conditional template rendering logic
5. Test all feature combinations (2^3 = 8 permutations)
6. Ensure no dependency conflicts between features

---

## Feature Combination Matrix

| # | Auth | Observability | Docker | Description |
|---|------|---------------|--------|-------------|
| 1 | ❌ | ❌ | ❌ | Base only |
| 2 | ✅ | ❌ | ❌ | Auth only |
| 3 | ❌ | ✅ | ❌ | Observability only |
| 4 | ❌ | ❌ | ✅ | Docker only |
| 5 | ✅ | ✅ | ❌ | Auth + Observability |
| 6 | ✅ | ❌ | ✅ | Auth + Docker |
| 7 | ❌ | ✅ | ✅ | Observability + Docker |
| 8 | ✅ | ✅ | ✅ | All features |

---

## Implementation Details

### 3.1 Extended Template Structure

```
templates/
├── base/                       # Core templates (always included)
└── features/
    ├── auth/
    │   ├── internal/
    │   │   ├── domain/
    │   │   │   └── user.go.tmpl
    │   │   ├── application/
    │   │   │   └── auth_service.go.tmpl
    │   │   ├── ports/
    │   │   │   ├── http/
    │   │   │   │   ├── auth_middleware.go.tmpl
    │   │   │   │   └── auth_handler.go.tmpl
    │   │   │   └── grpc/
    │   │   │       └── auth_interceptor.go.tmpl
    │   │   └── adapters/
    │   │       ├── redis/
    │   │       │   └── session_store.go.tmpl
    │   │       └── oauth/
    │   │           ├── google.go.tmpl
    │   │           └── github.go.tmpl
    │   ├── api/proto/v1/
    │   │   └── auth.proto.tmpl
    │   ├── migrations/
    │   │   └── 000002_create_users.up.sql.tmpl
    │   │   └── 000002_create_users.down.sql.tmpl
    │   └── pkg/
    │       └── jwt/
    │           └── jwt.go.tmpl
    ├── observability/
    │   ├── internal/
    │   │   └── telemetry/
    │   │       ├── metrics.go.tmpl
    │   │       ├── tracing.go.tmpl
    │   │       └── logging.go.tmpl
    │   └── pkg/
    │       └── logger/
    │           └── logger.go.tmpl
    └── docker/
        ├── Dockerfile.tmpl
        ├── docker-compose.yml.tmpl
        └── .dockerignore.tmpl
```

### 3.2 Authentication Feature

#### 3.2.1 User Domain Model

**File**: `templates/features/auth/internal/domain/user.go.tmpl`

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
    FullName     string
    Role         Role
    Provider     AuthProvider
    ProviderID   string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type Role string

const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
)

type AuthProvider string

const (
    ProviderLocal  AuthProvider = "local"
    ProviderGoogle AuthProvider = "google"
    ProviderGitHub AuthProvider = "github"
)

func (u *User) HasRole(role Role) bool {
    return u.Role == role
}

func (u *User) IsAdmin() bool {
    return u.Role == RoleAdmin
}
```

#### 3.2.2 JWT Package

**File**: `templates/features/auth/pkg/jwt/jwt.go.tmpl`

```go
package jwt

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    Role   string    `json:"role"`
    jwt.RegisteredClaims
}

type Manager struct {
    secretKey     []byte
    accessExpiry  time.Duration
    refreshExpiry time.Duration
}

func NewManager(secretKey string, accessExpiry, refreshExpiry time.Duration) *Manager {
    return &Manager{
        secretKey:     []byte(secretKey),
        accessExpiry:  accessExpiry,
        refreshExpiry: refreshExpiry,
    }
}

func (m *Manager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secretKey)
}

func (m *Manager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
    claims := jwt.RegisteredClaims{
        Subject:   userID.String(),
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.refreshExpiry)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secretKey)
}

func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return m.secretKey, nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    claims, ok := token.Claims.(*Claims)
    if !ok {
        return nil, fmt.Errorf("invalid claims type")
    }

    return claims, nil
}
```

#### 3.2.3 HTTP Auth Middleware

**File**: `templates/features/auth/internal/ports/http/auth_middleware.go.tmpl`

```go
package http

import (
    "context"
    "net/http"
    "strings"

    "{{ .ModulePath }}/pkg/jwt"
)

type contextKey string

const (
    UserIDKey contextKey = "user_id"
    EmailKey  contextKey = "email"
    RoleKey   contextKey = "role"
)

type AuthMiddleware struct {
    jwtManager *jwt.Manager
}

func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
    return &AuthMiddleware{
        jwtManager: jwtManager,
    }
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "missing authorization header", http.StatusUnauthorized)
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
            return
        }

        claims, err := m.jwtManager.ValidateToken(parts[1])
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        // Add claims to context
        ctx := r.Context()
        ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
        ctx = context.WithValue(ctx, EmailKey, claims.Email)
        ctx = context.WithValue(ctx, RoleKey, claims.Role)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userRole, ok := r.Context().Value(RoleKey).(string)
            if !ok || userRole != role {
                http.Error(w, "insufficient permissions", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

#### 3.2.4 gRPC Auth Interceptor

**File**: `templates/features/auth/internal/ports/grpc/auth_interceptor.go.tmpl`

```go
package grpc

import (
    "context"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"

    "{{ .ModulePath }}/pkg/jwt"
)

type AuthInterceptor struct {
    jwtManager *jwt.Manager
    // Map of method names that don't require authentication
    publicMethods map[string]bool
}

func NewAuthInterceptor(jwtManager *jwt.Manager) *AuthInterceptor {
    return &AuthInterceptor{
        jwtManager: jwtManager,
        publicMethods: map[string]bool{
            "/api.v1.HealthService/Check": true,
            "/api.v1.HealthService/Ready": true,
            "/api.v1.AuthService/Login":   true,
            "/api.v1.AuthService/Register": true,
        },
    }
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
    return func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        // Skip auth for public methods
        if i.publicMethods[info.FullMethod] {
            return handler(ctx, req)
        }

        claims, err := i.authorize(ctx)
        if err != nil {
            return nil, err
        }

        // Add claims to context
        ctx = context.WithValue(ctx, "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "email", claims.Email)
        ctx = context.WithValue(ctx, "role", claims.Role)

        return handler(ctx, req)
    }
}

func (i *AuthInterceptor) authorize(ctx context.Context) (*jwt.Claims, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return nil, status.Error(codes.Unauthenticated, "missing metadata")
    }

    values := md.Get("authorization")
    if len(values) == 0 {
        return nil, status.Error(codes.Unauthenticated, "missing authorization header")
    }

    authHeader := values[0]
    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) != 2 || parts[0] != "Bearer" {
        return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
    }

    claims, err := i.jwtManager.ValidateToken(parts[1])
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    return claims, nil
}
```

#### 3.2.5 Redis Session Store

**File**: `templates/features/auth/internal/adapters/redis/session_store.go.tmpl`

```go
package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type SessionStore struct {
    client *redis.Client
    ttl    time.Duration
}

func NewSessionStore(redisURL string, ttl time.Duration) (*SessionStore, error) {
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
    }

    client := redis.NewClient(opt)

    // Verify connection
    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &SessionStore{
        client: client,
        ttl:    ttl,
    }, nil
}

func (s *SessionStore) Set(ctx context.Context, key string, value string) error {
    return s.client.Set(ctx, key, value, s.ttl).Err()
}

func (s *SessionStore) Get(ctx context.Context, key string) (string, error) {
    val, err := s.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return "", fmt.Errorf("session not found")
    }
    return val, err
}

func (s *SessionStore) Delete(ctx context.Context, key string) error {
    return s.client.Del(ctx, key).Err()
}

func (s *SessionStore) Close() error {
    return s.client.Close()
}
```

#### 3.2.6 OAuth2 Providers

**File**: `templates/features/auth/internal/adapters/oauth/google.go.tmpl`

```go
package oauth

import (
    "context"
    "fmt"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

type GoogleProvider struct {
    config *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret, redirectURL string) *GoogleProvider {
    return &GoogleProvider{
        config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            RedirectURL:  redirectURL,
            Scopes:       []string{"email", "profile"},
            Endpoint:     google.Endpoint,
        },
    }
}

func (p *GoogleProvider) GetAuthURL(state string) string {
    return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    token, err := p.config.Exchange(ctx, code)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }
    return token, nil
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
    client := p.config.Client(ctx, token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    defer resp.Body.Close()

    var info UserInfo
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, fmt.Errorf("failed to decode user info: %w", err)
    }

    return &info, nil
}

type UserInfo struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Name          string `json:"name"`
    Picture       string `json:"picture"`
}
```

#### 3.2.7 Auth Migration

**File**: `templates/features/auth/migrations/000002_create_users.up.sql.tmpl`

```sql
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    provider VARCHAR(50) NOT NULL DEFAULT 'local',
    provider_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider, provider_id);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

### 3.3 Observability Feature

#### 3.3.1 Metrics Setup

**File**: `templates/features/observability/internal/telemetry/metrics.go.tmpl`

```go
package telemetry

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    // gRPC metrics
    GRPCRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests",
        },
        []string{"method", "status"},
    )

    GRPCRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "gRPC request latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method"},
    )

    // Database metrics
    DBQueriesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_queries_total",
            Help: "Total number of database queries",
        },
        []string{"query", "status"},
    )

    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query latencies in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"query"},
    )
)
```

#### 3.3.2 OpenTelemetry Tracing

**File**: `templates/features/observability/internal/telemetry/tracing.go.tmpl`

```go
package telemetry

import (
    "context"
    "fmt"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracing(serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
    ctx := context.Background()

    // Create OTLP exporter
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(), // Use TLS in production
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create exporter: %w", err)
    }

    // Create resource
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create resource: %w", err)
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)

    return tp, nil
}
```

#### 3.3.3 Structured Logging

**File**: `templates/features/observability/pkg/logger/logger.go.tmpl`

```go
package logger

import (
    "log/slog"
    "os"
)

type Logger struct {
    *slog.Logger
}

func New(level string, format string) *Logger {
    var logLevel slog.Level
    switch level {
    case "debug":
        logLevel = slog.LevelDebug
    case "info":
        logLevel = slog.LevelInfo
    case "warn":
        logLevel = slog.LevelWarn
    case "error":
        logLevel = slog.LevelError
    default:
        logLevel = slog.LevelInfo
    }

    var handler slog.Handler
    opts := &slog.HandlerOptions{
        Level: logLevel,
    }

    if format == "json" {
        handler = slog.NewJSONHandler(os.Stdout, opts)
    } else {
        handler = slog.NewTextHandler(os.Stdout, opts)
    }

    return &Logger{
        Logger: slog.New(handler),
    }
}

func (l *Logger) WithRequest(requestID string, method string, path string) *Logger {
    return &Logger{
        Logger: l.Logger.With(
            "request_id", requestID,
            "method", method,
            "path", path,
        ),
    }
}
```

### 3.4 Docker Feature

#### 3.4.1 Dockerfile

**File**: `templates/features/docker/Dockerfile.tmpl`

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app .
COPY --from=builder /build/configs ./configs
{{- if .EnableAuth }}
COPY --from=builder /build/migrations ./migrations
{{- end }}

# Create non-root user
RUN adduser -D -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

EXPOSE 8080 9090

CMD ["./app"]
```

#### 3.4.2 Docker Compose

**File**: `templates/features/docker/docker-compose.yml.tmpl`

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - APP_ENVIRONMENT=production
      - APP_DATABASE_URL=postgres://postgres:postgres@postgres:5432/{{ .ProjectName }}?sslmode=disable
      {{- if .EnableAuth }}
      - APP_REDIS_URL=redis://redis:6379/0
      {{- end }}
      {{- if .EnableObservability }}
      - APP_OTEL_ENDPOINT=otel-collector:4317
      {{- end }}
    depends_on:
      - postgres
      {{- if .EnableAuth }}
      - redis
      {{- end }}
    networks:
      - app-network

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB={{ .ProjectName }}
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - app-network

  {{- if .EnableAuth }}
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - app-network
  {{- end }}

  {{- if .EnableObservability }}
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - app-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    networks:
      - app-network

  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./configs/otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
    networks:
      - app-network
  {{- end }}

volumes:
  postgres-data:
  {{- if .EnableAuth }}
  redis-data:
  {{- end }}
  {{- if .EnableObservability }}
  prometheus-data:
  grafana-data:
  {{- end }}

networks:
  app-network:
    driver: bridge
```

### 3.5 Conditional Template Rendering

**File**: `internal/generator/generator.go` (updated methods)

```go
func (g *Generator) generateAuth(writer *FileWriter, data interface{}) error {
    authTemplates := []string{
        "features/auth/internal/domain/user.go.tmpl",
        "features/auth/internal/application/auth_service.go.tmpl",
        "features/auth/internal/ports/http/auth_middleware.go.tmpl",
        "features/auth/internal/ports/grpc/auth_interceptor.go.tmpl",
        "features/auth/internal/adapters/redis/session_store.go.tmpl",
        "features/auth/internal/adapters/oauth/google.go.tmpl",
        "features/auth/internal/adapters/oauth/github.go.tmpl",
        "features/auth/pkg/jwt/jwt.go.tmpl",
        "features/auth/api/proto/v1/auth.proto.tmpl",
        "features/auth/migrations/000002_create_users.up.sql.tmpl",
        "features/auth/migrations/000002_create_users.down.sql.tmpl",
    }

    return g.renderTemplates(authTemplates, writer, data)
}

func (g *Generator) generateObservability(writer *FileWriter, data interface{}) error {
    obsTemplates := []string{
        "features/observability/internal/telemetry/metrics.go.tmpl",
        "features/observability/internal/telemetry/tracing.go.tmpl",
        "features/observability/internal/telemetry/logging.go.tmpl",
        "features/observability/pkg/logger/logger.go.tmpl",
    }

    return g.renderTemplates(obsTemplates, writer, data)
}

func (g *Generator) generateDocker(writer *FileWriter, data interface{}) error {
    dockerTemplates := []string{
        "features/docker/Dockerfile.tmpl",
        "features/docker/docker-compose.yml.tmpl",
        "features/docker/.dockerignore.tmpl",
    }

    return g.renderTemplates(dockerTemplates, writer, data)
}

func (g *Generator) renderTemplates(templates []string, writer *FileWriter, data interface{}) error {
    for _, tmplPath := range templates {
        content, err := g.engine.Render(tmplPath, data)
        if err != nil {
            return fmt.Errorf("failed to render %s: %w", tmplPath, err)
        }

        outputPath := templates.GetOutputPath(tmplPath)
        if err := writer.WriteFile(outputPath, content, 0644); err != nil {
            return fmt.Errorf("failed to write %s: %w", outputPath, err)
        }
    }

    return nil
}
```

---

## Testing Strategy

### 3.6 Feature Combination Tests

```go
// internal/generator/features_test.go
func TestAllFeatureCombinations(t *testing.T) {
    combinations := []struct {
        name              string
        enableAuth        bool
        enableObservability bool
        enableDocker      bool
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
                ProjectName:         "test-" + tc.name,
                ModulePath:          "github.com/test/" + tc.name,
                EnableAuth:          tc.enableAuth,
                EnableObservability: tc.enableObservability,
                EnableDocker:        tc.enableDocker,
                OutputDir:           tmpDir,
            }

            gen := New()
            err := gen.Generate(cfg)
            require.NoError(t, err)

            // Verify compilation
            cmd := exec.Command("go", "mod", "tidy")
            cmd.Dir = tmpDir
            require.NoError(t, cmd.Run())

            cmd = exec.Command("go", "build", "./cmd/server")
            cmd.Dir = tmpDir
            require.NoError(t, cmd.Run())

            // Verify expected files exist
            verifyFeatureFiles(t, tmpDir, cfg)
        })
    }
}

func verifyFeatureFiles(t *testing.T, dir string, cfg *config.Config) {
    if cfg.EnableAuth {
        assert.FileExists(t, filepath.Join(dir, "pkg/jwt/jwt.go"))
        assert.FileExists(t, filepath.Join(dir, "internal/domain/user.go"))
    }

    if cfg.EnableObservability {
        assert.FileExists(t, filepath.Join(dir, "internal/telemetry/metrics.go"))
        assert.FileExists(t, filepath.Join(dir, "pkg/logger/logger.go"))
    }

    if cfg.EnableDocker {
        assert.FileExists(t, filepath.Join(dir, "Dockerfile"))
        assert.FileExists(t, filepath.Join(dir, "docker-compose.yml"))
    }
}
```

---

## Anti-Patterns to Avoid

1. **No Feature Coupling**: Features must be independent, no cross-dependencies
2. **No Conditional Compilation Tags**: Use template conditionals, not build tags
3. **No Hard-Coded URLs**: All endpoints configurable via environment
4. **No Default Secrets**: Force users to set JWT secrets, OAuth credentials
5. **No Metrics Cardinality Explosion**: Limit label values in Prometheus
6. **No Blocking Telemetry**: All observability calls must be async

---

## Success Validation Checklist

- [ ] All 8 feature combinations generate without errors
- [ ] All combinations compile successfully
- [ ] Auth feature includes JWT + OAuth2 + Redis
- [ ] Observability feature includes Prometheus + OTEL + slog
- [ ] Docker feature includes multi-stage Dockerfile + compose
- [ ] No dependency conflicts between features
- [ ] Generated go.mod contains correct dependencies per feature
- [ ] Docker images build successfully for Docker-enabled projects
- [ ] Metrics endpoint responds when observability enabled
- [ ] Auth middleware protects endpoints when auth enabled

---

## Deliverables

1. Complete authentication feature (JWT + OAuth2 + Redis + RBAC)
2. Complete observability feature (Prometheus + OTEL + slog)
3. Complete Docker feature (Dockerfile + docker-compose.yml)
4. Conditional template rendering logic
5. Feature combination test suite (all 8 permutations)
6. Updated go.mod templates with feature-specific dependencies
7. Updated configuration templates with feature-specific settings

**Next Phase**: Phase 4 - Testing & Polish
