# Backend Guidelines

## Instruction Scope
- Applies to all files under `backend/`.

## Build & Test
- Use Go 1.23.
- Format and vet with `go fmt ./...` and `go vet ./...` or `make fmt vet`.
- Run unit tests with `go test ./...`.
- WebSocket-specific code uses `//go:build ws`; include it in `make test-ws`.

## Coding Style
- `gofmt` defaults (tabs, standard import grouping).
- Packages are short and lowercase.
- Files and identifiers are lowercase; export only when necessary.

## Testing Guidelines
- Use the standard `testing` package.
- Test files end with `_test.go`; test functions use `TestXxx`.
- Keep tests deterministic and fast (<1s per package).
