---
description: 'Comprehensive Go (Golang) project instructions for best practices, conventions, and tooling.'
applyTo: ['**/*.go', 'go.mod', 'go.sum']
---

# Go Project Instructions

## Project Structure
- Organize code into `cmd/`, `internal/`, and `pkg/` directories.
- Place main application entry points in `cmd/<appname>/main.go`.
- Use `internal/` for private application code.
- Use `pkg/` for reusable packages.

## Dependency Management
- Use Go modules (`go.mod`, `go.sum`) for dependency tracking.
- Run `go mod tidy` to clean up unused dependencies.
- Pin dependencies to specific versions.

## Code Style
- Follow `gofmt` and `goimports` for formatting and imports.
- Use `golint` or `staticcheck` for linting.
- Prefer short, clear variable and function names.
- Use `err` for error variables.
- Return errors as the last return value.
- Use `context.Context` for cancellation and timeouts in APIs.

## Testing
- Place tests in the same package with `_test.go` suffix.
- Use `go test ./...` to run all tests.
- Use table-driven tests for multiple scenarios.
- Use `testing.T` and `testing.M` for test setup and teardown.

## Error Handling
- Always check and handle errors.
- Use `errors.Is` and `errors.As` for error inspection.
- Wrap errors with context using `fmt.Errorf` and `%w`.

## Documentation
- Write package-level and exported function comments.
- Use `godoc` conventions for documentation.

## Tooling
- Use `go build`, `go run`, and `go install` for building and running code.
- Use `go vet` for static analysis.
- Use `golangci-lint` for comprehensive linting.
- Use `delve` (`dlv`) for debugging.

## CI/CD
- Run tests and linters in CI pipelines.
- Use `go test -cover` for code coverage.
- Build binaries for multiple platforms using `GOOS` and `GOARCH`.

## Versioning
- Follow semantic versioning for modules.
- Tag releases in version control.

## Security
- Use `govulncheck` to scan for vulnerabilities.
- Avoid hardcoding secrets; use environment variables or secret managers.

## Performance
- Use `pprof` for profiling.
- Benchmark code with `go test -bench`.

## References
- [Go Official Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Modules Reference](https://blog.golang.org/using-go-modules)
