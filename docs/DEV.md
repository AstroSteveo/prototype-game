# Developer Guide

This document captures day-to-day commands for building, running, and testing the project locally.

## Prerequisites
- Go 1.23+
- curl (for simple HTTP checks)
- Optional: jq or Python 3 (to extract JSON fields in shell)

## Quick Start (Makefile)
The repo includes a Makefile with common workflows.

- Start services (gateway + sim with WS):
  - `make run`
- Stop services:
  - `make stop`
- Get a dev token from the gateway:
  - `make login`
- Probe WebSocket (join only):
  - `make wsprobe TOKEN=<value>`
- Probe movement + state (sends one input and prints a state):
  - `make wsprobe TOKEN=<value> MOVE_X=1 MOVE_Z=0`
- One-shot E2E (spins services up, tests, then stops):
  - Join: `make e2e-join`
  - Move: `make e2e-move`
- Build binaries:
  - `make build` (outputs to `backend/bin/`)
- Tests:
  - Unit: `make test`
  - WS integration (requires ws build tag): `make test-ws`

### Ports and Overrides
- Defaults: gateway `:8080`, sim `:8081`
- Override via variables:
  - `make run SIM_PORT=8082 GATEWAY_PORT=8080`
  - Probing uses the overridden ports automatically.

### Logs and PIDs
- Logs: `backend/logs/gateway.log`, `backend/logs/sim.log`
- PID files: `backend/.pids/gateway.pid`, `backend/.pids/sim.pid`
- Clean artifacts: `make clean`

## Running Manually (without Makefile)
From the repo root:

- Gateway:
  - `cd backend && go run ./cmd/gateway -port 8080 -sim localhost:8081`
- Sim (with WebSocket enabled via build tag):
  - `cd backend && go run -tags ws ./cmd/sim -port 8081`
- Health checks:
  - `curl http://localhost:8080/healthz`
  - `curl http://localhost:8081/healthz`
- Get token:
  - With jq: `TOKEN=$(curl -s "http://localhost:8080/login?name=Dev" | jq -r .token)`
  - With Python: `TOKEN=$(curl -s "http://localhost:8080/login?name=Dev" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')`
- WS probe CLI:
  - Join only: `cd backend && go run ./cmd/wsprobe -url ws://localhost:8081/ws -token "$TOKEN"`
  - Move + state: `cd backend && go run ./cmd/wsprobe -url ws://localhost:8081/ws -token "$TOKEN" -move_x 1`

## Building Binaries
- `cd backend && mkdir -p bin`
- Gateway: `go build -o bin/gateway ./cmd/gateway`
- Sim (WS enabled): `go build -tags ws -o bin/sim ./cmd/sim`
- WS probe: `go build -o bin/wsprobe ./cmd/wsprobe`
- Run binaries:
  - `./backend/bin/gateway -port 8080 -sim localhost:8081`
  - `./backend/bin/sim -port 8081`

## Tests
- Unit tests: `cd backend && go test ./...`
- WS integration tests (require ws tag): `cd backend && go test -tags ws ./...`

## Useful Endpoints
- Gateway:
  - `GET /healthz`
  - `GET /login?name=YourName` → `{ token, player_id, sim }`
  - `GET /validate?token=...` (used by sim for auth)
- Sim:
  - `GET /healthz`
  - `GET /config`
  - `GET /ws` (WebSocket; available only when built with `-tags ws`)
  - Dev-only helpers:
    - `GET /dev/spawn?id=p123&name=Alice&x=0&z=0`
    - `GET /dev/vel?id=p123&vx=1&vz=0`
    - `GET /dev/players`

## Troubleshooting
- `/ws` returns 501 Not Implemented:
  - Start/build sim with the `ws` build tag (Makefile does this).
- Address already in use:
  - `make stop` to stop background processes; or kill manually (`pkill -f "/cmd/gateway"` and `pkill -f "/cmd/sim"`).
- Auth errors (`error { code: "auth" }`):
  - Ensure you use a fresh `TOKEN` from `GET /login`.
- No `state` messages after `input`:
  - The sim broadcasts `state` at `SnapshotHz` (default 10 Hz); wait up to a couple of ticks.

## Project Layout (Quick Reference)
- Gateway service: `backend/cmd/gateway`
- Sim service: `backend/cmd/sim` (WebSocket handler behind `-tags ws`)
- Join/auth logic: `backend/internal/join`
- Engine: `backend/internal/sim`
- Spatial math: `backend/internal/spatial`
- WebSocket transport: `backend/internal/transport/ws`
- CLI probe: `backend/cmd/wsprobe`
- Backlog/TDD: `docs/BACKLOG.md`, `docs/TDD.md`

## Story Checklist (Copy into PR description)
- [ ] Acceptance criteria implemented and exercised
- [ ] Unit tests added/updated and passing (`make test`)
- [ ] Integration tests added/updated if applicable (`make test-ws`)
- [ ] Backlog updated (status, tests/evidence) in `docs/BACKLOG.md`
- [ ] Developer docs updated if commands changed (`docs/DEV.md`)
- [ ] `go fmt`, `go vet`, and all tests pass
