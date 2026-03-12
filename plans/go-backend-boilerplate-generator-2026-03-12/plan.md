---
status: in-progress
created: 2026-03-12
estimated_duration: 4-5 weeks
progress:
  phase_1: completed
  phase_2: completed
  phase_3: completed
  phase_4: pending
---

# Go Backend Boilerplate Generator - Implementation Plan

## Project Overview

Build an interactive CLI tool that generates production-ready Go backend projects with clean architecture, dual protocol support (REST + gRPC via grpc-gateway), PostgreSQL persistence, and optional feature toggles for authentication, observability, and Docker environments.

**Key Differentiators**:
- Clean 4-layer architecture (domain, application, ports, adapters)
- Built-in API versioning from day one
- Type-safe PostgreSQL queries (pgx + sqlc)
- Zero anti-patterns, strict linting compliance
- All feature combinations tested and validated

## Success Criteria

1. **Functional Requirements**:
   - Generates compilable Go projects in < 30 seconds
   - All 8 feature combinations (2^3) produce valid code
   - Generated projects pass golangci-lint strict preset
   - Integration tests pass with real PostgreSQL (testcontainers)

2. **Quality Gates**:
   - Generator test coverage > 80%
   - Generated code test coverage > 70%
   - Zero known security vulnerabilities (gosec clean)
   - Docker images < 100MB (multi-stage builds)

3. **Developer Experience**:
   - Time to running "Hello World" API: < 5 minutes
   - Clear error messages with actionable guidance
   - Generated README with step-by-step setup
   - Example CRUD implementation included

## Phase Summary

### Phase 1: Core Generator (Week 1-2)
Foundation CLI framework and base project structure generator. Produces minimal viable project with REST+gRPC, PostgreSQL adapter, and basic configuration.

**Deliverable**: CLI that generates compilable base project with health check endpoint.

### Phase 2: gRPC-REST Integration (Week 2-3)
Complete dual-protocol setup with buf, proto templates, grpc-gateway mux, sqlc queries, and migration system.

**Deliverable**: Generated projects expose REST and gRPC APIs simultaneously with PostgreSQL persistence.

### Phase 3: Feature Toggles (Week 3-4)
Implement conditional template rendering for auth (JWT+OAuth2+Redis), observability (Prometheus+OTEL+slog), and Docker environments.

**Deliverable**: All 8 feature combinations generate valid, tested code.

### Phase 4: Testing & Polish (Week 4-5)
Comprehensive test suite, golden file validation, example implementations, CI/CD pipeline, and documentation.

**Deliverable**: Production-ready v1.0 release with full test coverage.

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Template complexity explosion | Medium | High | Strict modular design, shared partials, DRY principles |
| Feature combination bugs | High | Medium | Automated matrix testing, golden file validation |
| Dependency version conflicts | Low | High | Pin all versions, quarterly update schedule |
| Proto-first learning curve | Medium | Low | Rich documentation, example implementations |
| Test flakiness (containers) | Medium | Medium | Retry logic, connection pooling, CI timeouts |

## Dependencies

### External Tools Required:
- Go 1.22+ (generics, slog, native context)
- buf CLI (proto linting, generation)
- golangci-lint (code quality)
- docker + docker-compose (optional feature)
- PostgreSQL 15+ (testing)

### Go Libraries (Generator):
- github.com/spf13/cobra (CLI framework)
- github.com/AlecAivazis/survey/v2 (interactive prompts)
- github.com/spf13/viper (config management)
- github.com/sebdah/goldie/v2 (golden file testing)

### Go Libraries (Generated Project):
- github.com/gin-gonic/gin (HTTP router)
- google.golang.org/grpc (gRPC server)
- github.com/grpc-ecosystem/grpc-gateway/v2 (REST-gRPC bridge)
- github.com/jackc/pgx/v5 (PostgreSQL driver)
- github.com/kyleconroy/sqlc (query code generation)
- github.com/golang-migrate/migrate/v4 (migrations)
- github.com/testcontainers/testcontainers-go (integration tests)

## Project Constraints

1. **YAGNI Enforcement**: No speculative features beyond documented requirements
2. **Go Idioms Only**: No reflection magic, no code generation at runtime
3. **PostgreSQL Only**: No multi-database abstraction layers
4. **Monolithic First**: No distributed patterns (event sourcing, CQRS, sagas)
5. **Stdlib Preference**: Use standard library when sufficient (slog, context, errors)

## Definition of Done

- [ ] All phase deliverables completed and tested
- [ ] Code passes golangci-lint with strict preset
- [ ] All integration tests pass on clean environment
- [ ] Documentation complete (README, inline comments, examples)
- [ ] Version tagged and released (v1.0.0)
- [ ] CI/CD pipeline green
- [ ] Migration guide published
