# Contributing to Prototype Game

Thanks for your interest in improving Prototype Game! This repository contains a Go-based multiplayer backend with two primary services:
- Gateway (auth/player/session) – default: :8080
- Simulation (game loop + WebSocket transport) – default: :8081

This guide reflects the current architecture, tooling, and checks.

## Ways to contribute
- Code: features, bug fixes, performance and scalability, observability
- Tests: unit, WebSocket integration, end-to-end using Make targets
- Docs: developer workflow, design/process documentation, troubleshooting
- Proposals: AOI/handovers, protocol/transport behavior, gameplay rules

## Prerequisites
- Go 1.23+
- make
- curl
- Python 3 (used by some Makefile targets)

Tip: Prefer using the Makefile targets to keep commands consistent.

## Getting started (local dev)
- Build binaries
  ```bash
  make build
  # outputs: backend/bin/{gateway,sim,wsprobe}
  ```

- Format and vet (should be clean before committing)
  ```bash
  make fmt vet
  ```

- Run tests
  ```bash
  # Unit tests
  make test

  # WebSocket-inclusive tests (requires -tags ws)
  make test-ws

  # Optional race-detection variants
  make test-race
  make test-ws-race
  ```

- Run services
  ```bash
  make run
  # Logs: backend/logs/{gateway.log,sim.log}
  # PIDs: backend/.pids/{gateway.pid,sim.pid}
  ```

- Health checks
  ```bash
  curl http://localhost:8080/healthz
  curl http://localhost:8081/healthz
  ```

WebSocket-specific code/tests must be behind the build tag:
```go
//go:build ws
```

## Validations to try before review
- Get a dev token and validate
  ```bash
  make run
  TOKEN=$(make login)
  curl "http://localhost:8080/validate?token=$TOKEN"
  ```

- Probe WebSocket join
  ```bash
  make wsprobe TOKEN="$TOKEN"
  # Expect a join_ack with player_id, position, cell, config
  ```

- Probe movement
  ```bash
  make wsprobe TOKEN="$TOKEN" MOVE_X=1 MOVE_Z=0
  # Expect updated position/velocity in subsequent state
  ```

- End-to-end helpers
  ```bash
  make e2e-join
  make e2e-move
  ```

## Branching model
- One branch per task/story. Examples:
  - feat/aoi-visibility-improvements
  - fix/handover-hysteresis
  - docs/update-contributing
- Do not push directly to main.
- Open a Draft PR early to exercise checks and get feedback.

## Commit conventions
- Small, focused commits; keep tests green.
- Imperative subject with a scoped prefix. Examples:
  - sim: tighten AOI boundary clamp
  - gateway: validate token expiry on join
  - wsprobe: add reconnect backoff
  - docs: clarify local dev workflow
- Reference story/issue IDs when applicable (e.g., #123, US-201).

## Tests and checks
- Local commands (mirrored by CI):
  ```bash
  make fmt vet test test-ws
  # Optionally: make test-race test-ws-race
  ```
- WS-only code and tests must be guarded with `//go:build ws`.
- PRs that fail formatting, vetting, or tests will not be merged.

## Running services manually
- Gateway
  ```bash
  cd backend && go run ./cmd/gateway -port 8080 -sim localhost:8081
  ```

- Simulation (with WebSocket support)
  ```bash
  cd backend && go run -tags ws ./cmd/sim -port 8081
  ```

## Code style and engineering guidelines
- Use `make fmt vet` before committing; formatting and vetting are enforced.
- Follow Go defaults (gofmt, short lowercase package names, minimal exports).
- Prefer clear interfaces for collaborators/mocks in tests.
- Errors: wrap with context; return early on failure.
- Logging: structured, context-aware; avoid `panic` in libraries.
- Tests: deterministic and fast; prefer table-driven tests and subtests.

## Documentation and process
- Repo-wide guidance: see [AGENTS.md](../AGENTS.md)
- Backend conventions: see [backend/AGENTS.md](../backend/AGENTS.md)
- Docs conventions: see [docs/AGENTS.md](../docs/AGENTS.md)
- Process/design documentation lives under [docs/](../docs/)

## Pull request checklist
- [ ] Branch created (no direct commits to main)
- [ ] `make fmt vet test test-ws` passes locally
- [ ] WS-only code/tests behind `//go:build ws`
- [ ] Validation commands run (include outputs or notes in PR)
- [ ] Relevant docs updated (when behavior or status changes)
- [ ] No secrets, logs, or build artifacts committed
- [ ] Clear commit messages with appropriate prefixes
- [ ] Linked issues/story IDs as applicable

## House rules
- Keep main linear; prefer rebase/fast-forward merges.
- Do not commit secrets or build artifacts; respect `.gitignore`.
- Use Makefile targets where available.
- Open Draft PRs early to surface issues quickly.

## Key references
- Make targets: [Makefile](../Makefile)
- CI workflow: [workflows/ci.yml](workflows/ci.yml)
- Repo-wide guidance: [AGENTS.md](../AGENTS.md)
- Backend guidance: [backend/AGENTS.md](../backend/AGENTS.md)
- Docs guidance: [docs/AGENTS.md](../docs/AGENTS.md)
