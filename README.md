# Go Backend Boilerplate Generator

A powerful CLI tool that generates production-ready Go backend projects with clean architecture, dual protocol support (REST + gRPC), and optional feature toggles.

## Features

- 🏗️ **Clean Architecture** - Well-organized project structure following best practices
- 🔄 **Dual Protocol Support** - REST API and gRPC out of the box
- 🔐 **Authentication** - Optional JWT-based authentication system
- 📊 **Observability** - Built-in metrics, logging, and monitoring
- 🐳 **Docker Ready** - Docker and docker-compose configuration
- 🗄️ **Database Integration** - PostgreSQL with migrations using sqlc
- 🛡️ **Security Hardened** - Input validation, rate limiting, and security best practices
- ⚡ **Fast Development** - Generate complete projects in seconds

## What Gets Generated

The generator creates a complete backend project with:

- RESTful API with versioning (`/api/v1/`)
- gRPC server with protocol buffers
- Clean layered architecture (handlers, services, repositories)
- Database migrations and models
- Health check endpoints
- Makefile for common tasks
- Comprehensive testing setup
- Docker and docker-compose files (optional)
- Authentication middleware (optional)
- Prometheus metrics (optional)

## Prerequisites

- Go 1.21 or higher
- Make (for running Makefile commands)
- Docker and docker-compose (optional, for containerization)
- Protocol Buffers compiler (for gRPC development)

## Installation

### Option 1: Download Binary

```bash
# Clone the repository
git clone https://github.com/ThinhDangDev/go-boilerplate.git
cd go-boilerplate

# Build the binary
go build -o go-boilerplate .
```

### Option 2: Install Globally

```bash
go install github.com/ThinhDangDev/go-boilerplate@latest
```

## Usage

### Interactive Mode (Recommended for First-Time Users)

Simply run the command and follow the prompts:

```bash
./go-boilerplate init
```

You'll be asked to provide:
- **Project name**: The name of your project (e.g., `my-api`)
- **Module name**: Your Go module path (e.g., `github.com/username/my-api`)
- **Features**: Which optional features to include

### Non-Interactive Mode

Provide all configuration via command-line flags:

```bash
./go-boilerplate init \
  --name=my-api \
  --module=github.com/username/my-api \
  --features=auth,observability,docker
```

### Command-Line Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--name` | Project name | `--name=user-service` |
| `--module` | Go module name | `--module=github.com/myorg/user-service` |
| `--features` | Comma-separated feature list | `--features=auth,observability,docker` |

### Available Features

| Feature | Description |
|---------|-------------|
| `auth` | JWT-based authentication system with middleware |
| `observability` | Prometheus metrics, structured logging, tracing |
| `docker` | Dockerfile, docker-compose.yml with PostgreSQL |

## Examples

### Basic Project (REST + gRPC only)

```bash
./go-boilerplate init \
  --name=simple-api \
  --module=github.com/mycompany/simple-api
```

### Full-Featured Project

```bash
./go-boilerplate init \
  --name=user-service \
  --module=github.com/mycompany/user-service \
  --features=auth,observability,docker
```

### E-commerce Backend

```bash
./go-boilerplate init \
  --name=ecommerce-api \
  --module=github.com/myshop/ecommerce-api \
  --features=auth,observability,docker
```

## After Generation

Once your project is generated, navigate to it and follow these steps:

```bash
# Navigate to your new project
cd my-api

# Generate code from protocol buffers
make proto

# Download Go dependencies
go mod download

# Start PostgreSQL and other services (if Docker feature enabled)
docker-compose up -d

# Run database migrations
make migrate-up

# Start the development server
make run
```

### Your Server Endpoints

After running `make run`, your server will be available at:

- **REST API**: `http://localhost:8080/api/v1/`
- **gRPC**: `localhost:9090`
- **Health Check**: `http://localhost:8080/api/v1/health`
- **Metrics** (if observability enabled): `http://localhost:8080/metrics`

## Generated Project Structure

```
my-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── domain/                  # Domain models and interfaces
│   ├── handler/                 # HTTP and gRPC handlers
│   ├── middleware/              # HTTP middleware (auth, logging, etc.)
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic layer
│   └── server/                  # Server setup (REST + gRPC)
├── proto/                       # Protocol buffer definitions
├── migrations/                  # Database migration files
├── docker-compose.yml           # Docker services (optional)
├── Dockerfile                   # Application container (optional)
├── Makefile                     # Common development tasks
├── go.mod                       # Go module definition
└── README.md                    # Project documentation
```

## Common Makefile Commands

Generated projects include a Makefile with useful commands:

```bash
make run           # Run the application
make test          # Run tests
make test-coverage # Run tests with coverage report
make proto         # Generate code from .proto files
make migrate-up    # Apply database migrations
make migrate-down  # Rollback database migrations
make docker-build  # Build Docker image
make docker-run    # Run application in Docker
make lint          # Run linters
make clean         # Clean build artifacts
```

## Development Workflow

1. **Generate your project**
   ```bash
   ./go-boilerplate init --name=my-api --module=github.com/me/my-api --features=auth,docker
   ```

2. **Set up the environment**
   ```bash
   cd my-api
   cp .env.example .env  # Configure environment variables
   docker-compose up -d   # Start dependencies
   ```

3. **Run migrations**
   ```bash
   make migrate-up
   ```

4. **Start developing**
   ```bash
   make run
   ```

5. **Test your endpoints**
   ```bash
   curl http://localhost:8080/api/v1/health
   ```

## Testing

Generated projects include a comprehensive testing setup:

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
go test ./tests/integration/...

# Run unit tests only
go test ./internal/...
```

## Configuration

Generated projects use environment variables for configuration. Create a `.env` file:

```env
# Server
SERVER_PORT=8080
GRPC_PORT=9090

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=myapp

# JWT (if auth feature enabled)
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Observability (if observability feature enabled)
LOG_LEVEL=info
METRICS_ENABLED=true
```

## Troubleshooting

### Port Already in Use

If you get an error about ports being in use, either stop the conflicting service or change the port in your `.env` file.

### Protocol Buffer Generation Fails

Make sure you have `protoc` and required plugins installed:

```bash
# Install protoc
# macOS: brew install protobuf
# Linux: apt-get install protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Database Connection Fails

Ensure PostgreSQL is running and environment variables are correct:

```bash
docker-compose ps  # Check if PostgreSQL is running
docker-compose logs postgres  # Check PostgreSQL logs
```

## Architecture Decisions

### Why Clean Architecture?

- **Testability**: Each layer can be tested independently
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations (e.g., different databases)
- **Scalability**: Well-organized code scales better with team size

### Why REST + gRPC?

- **REST**: Human-readable, great for web clients and public APIs
- **gRPC**: High performance, strong typing, ideal for microservices
- **Best of Both**: Use the right protocol for each use case

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Version History

- **v1.0.0** - Initial release
  - Core CLI generator
  - REST + gRPC support
  - Feature toggles (auth, observability, docker)
  - Security hardening
  - Clean architecture

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

If you have questions or run into issues:

- Open an issue on GitHub
- Check existing issues for solutions
- Review the generated project's README for project-specific help

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Protocol Buffers](https://protobuf.dev/) - gRPC definitions
- [sqlc](https://sqlc.dev/) - Type-safe SQL
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations

---

**Made with ❤️ for the Go community**
