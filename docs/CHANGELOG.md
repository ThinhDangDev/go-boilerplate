# Changelog

All notable changes to the Go Backend Boilerplate Generator will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [Phase 3] - 2026-03-12

### Added - Feature Toggles
- **Authentication Feature (12 templates)**
  - User domain model with roles (user, admin) and providers (local, Google, GitHub)
  - JWT token manager with access/refresh token generation
  - HTTP auth middleware with Bearer token validation and RBAC
  - gRPC auth interceptor with public method whitelist
  - Redis session store for distributed sessions
  - OAuth2 providers (Google, GitHub) with token exchange
  - Auth service with bcrypt password hashing and login/register
  - Auth HTTP handler for REST endpoints
  - Auth proto definitions with gRPC-gateway annotations
  - Database migrations for users and refresh_tokens tables

- **Observability Feature (6 templates)**
  - Prometheus metrics (HTTP, gRPC, DB counters and histograms)
  - OpenTelemetry tracing with OTLP exporter
  - HTTP and gRPC metrics middleware
  - Structured logging with slog (JSON/text formats)
  - Prometheus scrape configuration template
  - OTEL collector pipeline configuration template

- **Docker Feature (3 templates)**
  - Multi-stage Dockerfile with non-root user
  - Docker Compose with conditional services (Redis, Prometheus, Grafana, OTEL)
  - .dockerignore for optimal build context

### Changed
- Updated `internal/generator/generator.go` to support conditional feature rendering
- Enhanced `templates/base/go.mod.tmpl` with feature-specific dependencies
- Improved `internal/templates/engine.go` with multi-path template discovery for tests

### Tests
- Added 4 comprehensive test suites for feature combinations
- `TestAllFeatureCombinations` - validates all 8 permutations (2^3)
- `TestFeatureIndependence` - ensures features work independently
- `TestGoModDependencies` - verifies conditional dependencies
- `TestAuthMigrationOrdering` - checks migration file numbering
- Total: 45 tests passing (100% pass rate)

### Fixed
- Template engine now finds templates from test directories (../../templates)
- Validator tests updated to reflect Phase 2 validation logic
- Added missing Prometheus and OTEL collector config files

## [Phase 2] - Phase 2 Complete

### Added
- buf configuration for proto linting and generation
- Proto file templates with API versioning (v1)
- gRPC server with reflection and graceful shutdown
- grpc-gateway for automatic REST API generation
- PostgreSQL adapter with pgx driver and connection pooling
- sqlc configuration for type-safe SQL queries
- Database migration system (golang-migrate)
- Helper scripts (generate-proto.sh, migrate.sh)
- Dual-server architecture (HTTP + gRPC)

### Changed
- Updated main.go to start both HTTP and gRPC servers
- Enhanced Makefile with proto, sqlc, migrate tasks
- Updated README with dual-API documentation

### Security
- Added security warning for insecure gRPC credentials
- Channel-based server synchronization (replaced time.Sleep)

### Technical Details
- 25 new template files added to the boilerplate
- 30 tests achieving 100% pass rate
- Support for REST and gRPC simultaneously
- OpenAPI spec auto-generation from proto files

## [Unreleased] - Phase 1 Complete

### Added
- CLI generator with interactive prompts using Cobra framework and Survey library
- Template engine with filesystem-based loading and Go text/template integration
- Atomic file writer implementing temp + rename pattern for safe file operations
- Base project structure generator following clean architecture principles
- Docker feature templates (Dockerfile and docker-compose.yml)
- Security: Path traversal validation for output directories
- Security: Template path validation to prevent unauthorized file access
- Command execution timeouts to protect against DoS attacks
- Git repository initialization with proper .gitignore
- Comprehensive test suite with 44 tests achieving 100% pass rate
- Project configuration management with validation
- Template variable substitution system

### Security
- Path traversal protection in output directory validation
- Template path validation prevents access to files outside template directory
- Command timeout protection (30 second default) against denial-of-service
- Atomic file writes prevent partial file corruption

### Technical Details
- Generator CLI entry point: `cmd/generate/main.go`
- Template engine: `internal/template/`
- File operations: `internal/fileops/`
- Project scaffolding: `internal/scaffold/`
- Interactive prompts: `internal/prompt/`
- Configuration: `internal/config/`
- Templates: `templates/` directory structure
  - `base/`: Core project structure (always included)
  - `features/docker/`: Optional Docker environment

### Test Coverage
- Unit tests: Template engine, file operations, configuration
- Integration tests: End-to-end project generation
- Test suite: 44 tests, 100% pass rate
- Coverage areas: Generator logic, security validation, file I/O, template rendering

## [0.0.0] - 2026-03-12

### Project Initialization
- Initial project setup and architecture planning
- Clean architecture design documented in `docs/brainstorm-2026-03-12.md`
- Technology stack selection (Cobra, Survey, pgx, sqlc, gin, gRPC)
