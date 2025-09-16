# Developer Guide

This document captures day-to-day commands for building, running, and testing the project locally.
For proposing new features or making architecture decisions, see [docs/process/FEATURE_PROPOSAL.md](../process/FEATURE_PROPOSAL.md).

## Prerequisites
- Go 1.23++
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

## CI and Testing
The project uses GitHub Actions for continuous integration. The CI workflow:

- **Triggers**: Runs on pushes and PRs to `main`, but skips docs-only changes
- **Test Matrix**: Tests both unit tests and WebSocket integration tests, with and without race detection
- **Make Targets**: CI uses Makefile targets for consistency:
  - `make fmt` - Format Go code with gofmt
  - `make fmt-check` - Check if Go code is formatted (non-mutating, used in CI)
  - `make vet` - Run go vet
  - `make test` - Run unit tests  
  - `make test-ws` - Run WebSocket integration tests
  - `make test-race` - Run unit tests with race detection
  - `make test-ws-race` - Run WebSocket tests with race detection

**Before committing**, always run: `make fmt vet test test-ws`

**CI Format Checking**: The CI uses a non-mutating format check that lists unformatted files without modifying them. If you see CI failures related to formatting, run `make fmt` locally to fix formatting issues.

## Reconnect / Resume (WS)
- On successful join, the server includes a `resume` token in `join_ack`.
- Clients can reconnect by sending `{token, resume, last_seq}` so the server continues input ACKs from `last_seq`.
- Resume tokens are in-memory with a short TTL (dev only).

## Player Persistence (Dev)
- The sim wires an in-memory store to remember each player's last position and a simple `logins` counter.
- On join, the server loads the saved position (if any) and increments `logins`.
- On WebSocket disconnect, the last known position is saved.
- Probe movement + state (sends one input and prints a state):
  - `make wsprobe TOKEN=<value> MOVE_X=1 MOVE_Z=0`
  - Note: while moving, the server may emit `handover` events when the player crosses cell boundaries (see Protocol Notes below).
- One-shot E2E (spins services up, tests, then stops):
  - Join: `make e2e-join`
  - Move: `make e2e-move`
- Build binaries:
  - `make build` (outputs to `backend/bin/`)
- Tests:
  - Unit: `make test`
  - WS integration (requires ws build tag): `make test-ws`
  - Unit with race detection: `make test-race`
  - WS integration with race detection: `make test-ws-race`

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
- Get token and WebSocket URL:
  - With jq: `TOKEN=$(curl -s "http://localhost:8080/login?name=Dev" | jq -r .token)`
  - With Python: `TOKEN=$(curl -s "http://localhost:8080/login?name=Dev" | python3 -c 'import sys,json; print(json.load(sys.stdin)["token"])')`
  - Get WS URL from login response: `WS_URL=$(curl -s "http://localhost:8080/login?name=Dev" | python3 -c 'import sys,json; print(json.load(sys.stdin)["sim"]["address"])')`
- WS probe CLI:
  - Join only: `cd backend && go run ./cmd/wsprobe -url ws://localhost:8081/ws -token "$TOKEN"`
  - Using URL from login: `cd backend && go run ./cmd/wsprobe -url "$WS_URL" -token "$TOKEN"`
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
- Unit tests: `make test` or `cd backend && go test ./...`
- WS integration tests (require ws tag): `make test-ws` or `cd backend && go test -tags ws ./...`
- Race detection: `make test-race` or `make test-ws-race`
  - Includes tests for movement, telemetry, cadence, and handover events.

## Useful Endpoints
- Gateway:
  - `GET /healthz`
  - `GET /login?name=YourName` → `{ token, player_id, sim: { address: "ws://host:port/ws", protocol: "ws-json", version: "1" } }`
  - `GET /validate?token=...` (used by sim for auth)
- Sim:
  - `GET /healthz`
  - `GET /config`
  - `GET /metrics` (Prometheus text)
  - `GET /metrics.json` → `{ handovers, aoi_queries, aoi_entities_total, aoi_avg_entities }`
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

## Protocol Notes
- Messages from server:
  - `join_ack { player_id, pos, cell, config }`
  - `state { ack, player, entities[] }` at ~`SnapshotHz`
  - `telemetry { tick_rate, rtt_ms }` at ~1 Hz
  - `handover { from: {cx,cz}, to: {cx,cz} }` whenever the player’s owned cell changes (hysteresis applied)

## Project Layout (Quick Reference)
- Gateway service: `backend/cmd/gateway`
- Sim service: `backend/cmd/sim` (WebSocket handler behind `-tags ws`)
- Join/auth logic: `backend/internal/join`
- Engine: `backend/internal/sim`
- Spatial math: `backend/internal/spatial`
- WebSocket transport: `backend/internal/transport/ws`
- CLI probe: `backend/cmd/wsprobe`
- Backlog/TDD: GitHub issues, `docs/design/TDD.md`

## Story Checklist (Copy into PR description)
- [ ] Acceptance criteria implemented and exercised
- [ ] Unit tests added/updated and passing (`make test`)
- [ ] Integration tests added/updated if applicable (`make test-ws`)
- [ ] GitHub issue updated with status and tests/evidence
- [ ] Developer docs updated if commands changed (`docs/dev/DEV.md`)
- [ ] Code formatted (`make fmt`), vet clean (`make vet`), and all tests pass
