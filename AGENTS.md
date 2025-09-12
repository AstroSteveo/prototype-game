# Repository Guidelines

## Project Structure & Module Organization
- **Root:** High‑level docs in `docs/` (GDD, TDD, backlog).
- **Backend (Go):** `backend/` with service entries under `backend/cmd/` (`gateway`, `sim`) and libraries in `backend/internal/` (`sim`, `spatial`, `join`, `transport/ws`).
- **Client:** `client/` (placeholder for the game client; see `client/README.md`).

## Build, Test, and Development Commands
- **Run gateway:** `cd backend && go run ./cmd/gateway -port 8080 -sim localhost:8081`
- **Run sim (HTTP + stub WS):** `cd backend && go run ./cmd/sim -port 8081`
- **Run sim with WebSocket:** `cd backend && go run -tags ws ./cmd/sim -port 8081`
- **Unit tests:** `cd backend && go test ./...` (runs package tests under `internal/`)
- **Format & vet:** `cd backend && go fmt ./... && go vet ./...`

## Coding Style & Naming Conventions
- **Language:** Go 1.21. Use `gofmt` defaults (tabs, standard import grouping).
- **Packages:** short, lowercase (e.g., `sim`, `spatial`, `join`).
- **Files & symbols:** lowercase file names; exported types/functions only when needed; prefer clear, short identifiers.
- **Build tags:** WebSocket transport guarded by `//go:build ws` in `internal/transport/ws`.

## Testing Guidelines
- **Framework:** standard `testing` package; colocate `*_test.go` next to sources.
- **Scope:** add tests for new logic (engine stepping, handovers, spatial math).
- **Naming:** test files `*_test.go`; functions `TestXxx` with table tests where helpful.
- **Run:** `go test ./...`; aim to keep tests deterministic and fast (<1s per package).

## Commit & Pull Request Guidelines
- **Commits:** use concise, imperative subjects (e.g., "sim: fix handover hysteresis"). Group related changes; keep diffs focused.
- **PRs:** include intent, summary of changes, and testing notes. Link issues/backlog items from `docs/`. Add screenshots or CLI transcripts for behavior changes (e.g., `/dev/players` output).
- **Checks:** ensure `go fmt`, `go vet`, and tests pass; update docs (GDD/TDD) when behavior or APIs change.

## Security & Configuration Tips
- **Ports:** gateway on `:8080`; sim on `:8081` (configurable).
- **Auth:** gateway issues dev tokens; sim validates via `-gateway` URL.
- **Local WS:** build with `-tags ws` to enable `/ws`; otherwise `/ws` returns 501.
