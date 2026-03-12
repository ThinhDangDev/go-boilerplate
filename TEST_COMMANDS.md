# Test Commands Quick Reference

## Run All Tests
```bash
cd /Users/thinhdang/pila-hcm/go-backend-boilerplate
go test ./... -v
```

## Run Unit Tests Only
```bash
# Config tests
go test ./internal/config/... -v

# Generator tests (FileWriter + Validator)
go test ./internal/generator/... -v

# Template engine tests
go test ./internal/templates/... -v

# All unit tests
go test ./internal/... -v
```

## Run Integration Tests Only
```bash
go test ./tests/integration/... -v -timeout 10m
```

## Run Tests with Coverage
```bash
# Quick coverage summary
go test -cover ./...

# Detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# HTML coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Run Specific Tests
```bash
# Run a specific test function
go test ./internal/config -run TestConfig_Validate -v

# Run tests matching a pattern
go test ./internal/generator -run TestFileWriter -v

# Run a specific sub-test
go test ./internal/config -run TestConfig_Validate/valid_config_with_all_fields -v
```

## Run Tests in Short Mode (skip integration tests)
```bash
go test ./... -short -v
```

## Run Tests with Race Detection
```bash
go test ./... -race -v
```

## Continuous Testing (watch mode)
```bash
# Using watchexec (install: brew install watchexec)
watchexec -e go -r -- go test ./... -v

# Or using entr (install: brew install entr)
find . -name '*.go' | entr -c go test ./...
```

## Benchmark Tests (when added)
```bash
go test -bench=. -benchmem ./...
```

## Test with Verbose Output and JSON Format
```bash
go test ./... -v -json > test-results.json
```

## Clean Test Cache
```bash
go clean -testcache
go test ./... -v
```

## Count Tests
```bash
# Count test functions
grep -r "^func Test" . --include="*_test.go" | wc -l

# Count sub-tests
go test ./... -v | grep "RUN" | wc -l
```

## Test Coverage by Package
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "^(github.com|total)"
```

## Parallel Testing
```bash
# Run tests in parallel (default)
go test ./... -v -parallel 4

# Run tests sequentially
go test ./... -v -parallel 1
```

## Test Timeout Configuration
```bash
# Set custom timeout for long-running tests
go test ./tests/integration/... -v -timeout 30m
```

## Generate Test Report
```bash
# Install gotestsum (go install gotest.tools/gotestsum@latest)
gotestsum --format testname ./...
gotestsum --format pkgname ./...
```

## Test with Different Build Tags
```bash
go test -tags integration ./...
go test -tags unit ./...
```
