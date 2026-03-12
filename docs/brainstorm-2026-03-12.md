# Go Backend Boilerplate Generator - Architecture Brainstorm

**Date**: 2026-03-12
**Type**: CLI Generator Tool
**Target**: Production-ready Go backend foundation for REST + gRPC APIs

---

## Problem Statement

Need a CLI generator that scaffolds production-ready Go backend projects with:
- Dual protocol support (REST + gRPC)
- PostgreSQL-only data layer
- JWT + OAuth2 authentication
- Full observability stack (metrics, tracing, structured logging)
- Interactive setup with feature toggles
- API versioning built-in from day one
- Monolithic deployment architecture (easy to extract microservices later)

**Goal**: Save 2-3 weeks of initial setup, enforce best practices, avoid common Go anti-patterns.

---

## Evaluated Approaches

### Option 1: Template Repository (GitHub Template)
**Pros**:
- Simple: clone and customize
- Version control friendly
- Easy to maintain

**Cons**:
- Manual find/replace for project names
- No conditional features (all-or-nothing)
- No interactive customization
- Users must manually remove unwanted features

**Verdict**: ❌ Too rigid for feature toggles requirement

---

### Option 2: CLI Generator with Templates (cookiecutter-style)
**Pros**:
- Interactive prompts for customization
- Conditional feature generation
- Can scaffold multiple file types
- Good developer experience

**Cons**:
- Template language overhead (Go templates or similar)
- Complex template maintenance
- Testing complexity (need to test all combinations)

**Verdict**: ✅ **RECOMMENDED** - Balances flexibility with maintainability

---

### Option 3: Code Generation from Specs (.proto first)
**Pros**:
- Single source of truth (proto files)
- Auto-generates REST + gRPC simultaneously
- Strong typing guarantees

**Cons**:
- Doesn't handle full project scaffolding
- Still needs project structure generator
- Limited to API layer only

**Verdict**: ⚠️ Use as **complement** to Option 2 (generate API layer from protos)

---

## Recommended Solution: Interactive CLI Generator

### Architecture Overview

```
go-boilerplate (CLI tool)
├── cmd/generate          # Generator CLI entry point
├── internal/
│   ├── template/         # Go template engine wrapper
│   ├── prompt/           # Interactive prompts (survey/promptui)
│   ├── scaffold/         # Project structure logic
│   └── config/           # Generator configuration
└── templates/            # Project templates
    ├── base/             # Core structure (always included)
    ├── features/         # Optional features
    │   ├── auth/
    │   ├── observability/
    │   └── docker/
    └── examples/         # Reference implementations
```

### Generated Project Structure (Clean Architecture)

```
<project-name>/
├── cmd/
│   └── server/           # Main application entry
├── internal/
│   ├── domain/           # Business entities (no dependencies)
│   ├── application/      # Use cases, business logic
│   ├── ports/            # Interfaces (HTTP, gRPC handlers)
│   │   ├── http/         # REST handlers + middleware
│   │   └── grpc/         # gRPC service implementations
│   ├── adapters/         # External integrations
│   │   ├── postgres/     # Database (pgx + sqlc)
│   │   ├── redis/        # Caching, session store
│   │   └── oauth/        # OAuth2 providers
│   └── config/           # Config management (Viper)
├── pkg/                  # Reusable packages
│   ├── logger/           # Structured logging (slog)
│   ├── telemetry/        # OpenTelemetry setup
│   └── validator/        # Input validation
├── api/
│   └── proto/v1/         # Protocol buffer definitions
├── migrations/           # SQL migration files (golang-migrate)
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── configs/              # Config files (dev, prod)
├── scripts/              # Build, deployment scripts
└── tests/
    ├── integration/      # testcontainers for PostgreSQL
    └── e2e/              # End-to-end API tests
```

### Core Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **HTTP Framework** | gin-gonic/gin | Fast, minimal, middleware ecosystem |
| **gRPC** | google.golang.org/grpc | Standard, with grpc-gateway for REST |
| **Database Driver** | jackc/pgx/v5 | 30-50% faster than GORM, native PostgreSQL |
| **Query Builder** | sqlc | Type-safe SQL, zero-runtime overhead |
| **Migrations** | golang-migrate | Standard, CLI + library support |
| **Auth** | golang-jwt/jwt (RS256) | Asymmetric signing, microservice-ready |
| **OAuth2** | golang.org/x/oauth2 | Official, supports all major providers |
| **Config** | spf13/viper | 12-factor compliance, multi-format |
| **Logging** | log/slog | Stdlib (Go 1.21+), structured, performant |
| **Metrics** | prometheus/client_golang | Industry standard, native histograms |
| **Tracing** | go.opentelemetry.io/otel | Vendor-neutral, auto-instrumentation |
| **Validation** | go-playground/validator | Declarative struct tags |
| **DI** | Manual (Wire for complex cases) | Start simple, add Wire if needed |
| **Testing** | testcontainers-go | Real PostgreSQL for integration tests |
| **API Docs** | grpc-gateway OpenAPIv2 | Auto-generated from proto files |

### Feature Toggle Matrix

| Feature | Default | Includes |
|---------|---------|----------|
| **Base** | Always | REST+gRPC, PostgreSQL, migrations, validation, config, logging |
| **Authentication** | Optional | JWT auth, OAuth2 (Google/GitHub), refresh tokens, Redis session store, RBAC helpers |
| **Full Observability** | Optional | Prometheus metrics, OpenTelemetry traces, slog JSON logging, /metrics endpoint, trace context propagation |
| **Docker Environment** | Optional | Dockerfile, docker-compose.yml (PostgreSQL, Redis, Prometheus, Grafana), health checks |

### API Versioning Strategy

**gRPC**:
```protobuf
// api/proto/v1/user.proto
package api.v1;
service UserService { ... }

// api/proto/v2/user.proto (future)
package api.v2;
service UserService { ... }
```

**REST** (via grpc-gateway):
```
/api/v1/users   → api.v1.UserService
/api/v2/users   → api.v2.UserService (future)
```

**Pattern**: Proto package versioning + URL path versioning, both enforced at compile-time.

---

## Generator CLI Specification

### Installation
```bash
go install github.com/yourorg/go-boilerplate@latest
```

### Usage
```bash
# Interactive mode (recommended)
go-boilerplate init

# Non-interactive (for CI/scripts)
go-boilerplate init \
  --name=my-api \
  --module=github.com/org/my-api \
  --features=auth,observability,docker
```

### Interactive Prompts Flow

1. **Project Name**: `my-backend`
2. **Go Module**: `github.com/myorg/my-backend`
3. **Enable Authentication?** (JWT + OAuth2) → Yes/No
4. **Enable Full Observability?** (Prometheus + OTEL + slog) → Yes/No
5. **Include Docker Environment?** (compose + Dockerfile) → Yes/No
6. **Generate Example CRUD?** (User resource as reference) → Yes/No
7. **Initialize Git Repository?** → Yes/No

**Post-generation**:
```bash
cd my-backend
go mod download
docker-compose up -d  # if Docker enabled
make migrate-up
make run
```

Server starts:
- REST: `http://localhost:8080/api/v1/`
- gRPC: `localhost:9090`
- Metrics: `http://localhost:8080/metrics` (if observability enabled)
- Health: `http://localhost:8080/health`

---

## Implementation Considerations

### Critical Success Factors

1. **Template Maintainability**
   - Use Go's `text/template` with clear delimiters
   - Keep templates DRY (shared partials for imports, middleware)
   - Test generated code with `go vet` + `golangci-lint`

2. **Dependency Versioning**
   - Pin dependencies in generated `go.mod`
   - Update strategy: quarterly security patches
   - Document breaking changes in CHANGELOG

3. **Testing Strategy**
   - Generator tests: golden file comparisons
   - Generated code tests: compile + integration suite
   - Test matrix: all feature combinations (8 combos total)

4. **Documentation**
   - Generate README.md with setup instructions
   - Inline code comments for customization points
   - Architecture decision records (ADRs) in `docs/adr/`

5. **Migration Path**
   - Provide upgrade guide for generated projects
   - Version generator CLI (semver)
   - Breaking changes only in major versions

### Known Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Template complexity grows** | Hard to maintain | Modular templates, shared partials, automated tests |
| **Dependency conflicts** | Generated projects break | Pin versions, quarterly updates, test matrix |
| **Over-engineering temptation** | Bloated boilerplate | Strict YAGNI enforcement, feature toggle design |
| **Proto-first learning curve** | User confusion | Generate example proto + docs, clear error messages |
| **Test flakiness (containers)** | CI failures | Use testcontainers pooling, retry logic, timeout configs |

### Anti-Patterns to Avoid

1. ❌ **Single Model**: Don't mix JSON/DB tags (separate domain/persistence models)
2. ❌ **Global State**: No global DB connections, pass via DI
3. ❌ **Goroutine Leaks**: Always use context cancellation
4. ❌ **time.Sleep() in Tests**: Use polling with timeout helpers
5. ❌ **Error Wrapping Loss**: Always `fmt.Errorf("context: %w", err)`
6. ❌ **Premature Abstraction**: Generate concrete implementations first

---

## Success Metrics

### Developer Experience
- ✅ Time to "Hello World" API: **< 5 minutes**
- ✅ Time to production-ready foundation: **< 2 hours** (vs 2-3 weeks manual)
- ✅ Zero security vulnerabilities in generated code (`gosec` clean)
- ✅ 80%+ test coverage in generated base code

### Technical Validation
- ✅ Generated project passes `golangci-lint` strict preset
- ✅ All integration tests pass with real PostgreSQL (testcontainers)
- ✅ OpenAPI spec validates with Swagger validator
- ✅ Docker image builds < 100MB (multi-stage build)
- ✅ gRPC reflection enabled for `grpcurl` testing

### Maintainability
- ✅ Clear separation of concerns (4-layer architecture)
- ✅ Easy to add new endpoints (proto → generate → implement use case)
- ✅ Database migrations are reversible
- ✅ Environment-specific configs without code changes

---

## Next Steps & Dependencies

### Phase 1: Core Generator (Week 1-2)
1. CLI framework setup (cobra + survey prompts)
2. Base template: project structure, main.go, config
3. PostgreSQL adapter (pgx + sqlc setup)
4. REST+gRPC handlers with example endpoint
5. Makefile with common tasks

### Phase 2: Feature Toggles (Week 3)
1. Auth feature: JWT middleware, OAuth2 flow, refresh logic
2. Observability feature: Prometheus, OTEL, slog integration
3. Docker feature: Compose, Dockerfile, health checks

### Phase 3: Testing & Documentation (Week 4)
1. Generator test suite (golden files)
2. Generated code integration tests
3. README generator with setup instructions
4. Example CRUD implementation (User resource)

### Phase 4: Polish & Release (Week 5)
1. CI/CD for generator itself (GitHub Actions)
2. Versioning strategy + changelog
3. Migration guide for updates
4. Public documentation site (optional)

---

## Open Questions / Decisions Needed

1. **Generator Distribution**: Go install vs binary releases vs Docker image?
   - **Recommendation**: All three (main via `go install`)

2. **Configuration Format**: YAML vs TOML vs ENV-only?
   - **Recommendation**: YAML for dev (Viper), ENV override for prod (12-factor)

3. **Error Handling Pattern**: Sentinel errors vs error wrapping vs custom types?
   - **Recommendation**: Error wrapping with `%w` + custom types for domain errors

4. **API Rate Limiting**: Built-in or external (nginx)?
   - **Recommendation**: External initially (YAGNI), provide integration guide

5. **File Upload Support**: Local vs S3-compatible?
   - **Recommendation**: Not in base, provide example implementation

---

## Conclusion

**Recommended Architecture**: Interactive CLI generator producing clean-architecture Go monoliths with:
- Dual protocol (REST + gRPC via grpc-gateway)
- PostgreSQL + pgx + sqlc
- Optional auth (JWT + OAuth2 + Redis)
- Optional observability (Prometheus + OTEL + slog)
- Optional Docker environment
- Built-in API versioning
- Production-ready defaults

**Estimated Development**: 4-5 weeks to v1.0
**Estimated Time Saved per Project**: 2-3 weeks of setup work
**Maintenance Burden**: Low (modular templates, quarterly dependency updates)

**Critical Path**: Phase 1 (base generator) must be rock-solid before adding features. All generated code must pass strict linting and have integration tests.

---

**Ready to proceed?** Next step: Create detailed implementation plan.
