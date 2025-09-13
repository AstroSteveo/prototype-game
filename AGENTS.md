# Repository Guidelines

## Instruction Scope
- AGENTS.md is the authoritative instruction set for agents and automation.
- Documents under `docs/` are human-facing references and do not override AGENTS.md.
- Per-subtree AGENTS.md files may exist and take precedence within their subtree.

## Project Structure & Module Organization
- **Root:** High‑level docs in `docs/` organized as:
  - `docs/design/` (GDD, TDD)
  - `docs/process/` (BACKLOG, PROGRESS)
  - `docs/dev/` (DEV guide)
- **Backend (Go):** `backend/` with service entries under `backend/cmd/` (`gateway`, `sim`) and libraries in `backend/internal/` (`sim`, `spatial`, `join`, `transport/ws`).
- **Client:** `client/` (placeholder for the game client; see `client/README.md`).

## Build, Test, and Development Commands
- **Makefile:** run `make help` for common targets (`run`, `stop`, `login`, `wsprobe`, `test`, `test-ws`, `build`). Prefer Makefile targets over raw commands to avoid drift.
- **Dev guide:** see `docs/dev/DEV.md` for day-to-day workflows and tips.
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
- **Frequency:** make small, focused commits with green tests. Reference story IDs in messages where relevant (e.g., `US-201`).
- **Branching:** one feature branch per story/task. Use names like `feat/us-201-aoi-visibility`, `fix/handover-hysteresis`, or `docs/progress-snapshot`. Don’t push to `main` directly; open a Draft PR early to get CI.
- **Scope:** keep PRs narrowly scoped (aim <300 lines changed when practical). Include tests for new logic and any doc updates.
- **CI gate:** run `make fmt vet test` and `make test-ws` locally before pushing; all checks must pass in CI.
- **Docs updates:** when story status or behavior changes, update `docs/process/BACKLOG.md` and `docs/process/PROGRESS.md` and reference those updates in the PR description.
- **History:** prefer rebase/fast-forward onto `main` for a linear history.
- **Build tags:** guard WS-only code/tests with `//go:build ws`; include them with `make test-ws`.

## Security & Configuration Tips
- **Ports:** gateway on `:8080`; sim on `:8081` (configurable).
- **Auth:** gateway issues dev tokens; sim validates via `-gateway` URL.
- **Local WS:** build with `-tags ws` to enable `/ws`; otherwise `/ws` returns 501.
- **Secrets/artifacts:** do not commit secrets, logs, or build artifacts; ensure they’re ignored locally.
