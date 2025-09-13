# Progress Log

This document tracks milestone status, what’s done, and what’s next.

## Snapshot (Current Status)
- Stack: Go 1.23; Makefile added; DEV guide added; CI via GitHub Actions; branch protection on main enforced.
- Services: `cmd/sim` (WS behind build tag), `cmd/gateway` (login + validate).
- Core sim: tick loop, kinematics, local handover with hysteresis; dev HTTP endpoints.
- WS transport: join handshake, input/state loop implemented under `-tags ws`; emits `handover` events on cell change.
  - AOI continuity across borders validated by WS test (no duplicates; within next snapshot).
  - Observability: Prometheus metrics exposed at `/metrics` (see below).
- Tests: spatial, engine, handover unit tests; WS integration test behind `-tags ws` passing; CI runs fmt/vet/unit and ws-tagged suites with `-race` across a small matrix.

Done stories (M0–M3 so far): US-000, US-101, US-102, US-103, US-104, US-201, US-202, US-302.

## Milestones (from TDD)
- M0 — Project skeleton: Completed
  - go.mod; `cmd/sim` + `cmd/gateway`; health/config endpoints.
  - Engine start/stop, tick/snapshot loop; README/DEV docs in place.
- M1 — Presence & Movement (single cell): Complete
  - US-101 (join): Done.
  - US-102 (spawn default): Done.
  - US-103 (movement + authoritative state): Done.
  - US-104 (telemetry): Done.
- M2 — Interest Management (AOI streaming): Complete
  - US-201 (AOI visibility): Done.
  - US-202 (snapshot cadence/budget): Done.
- M3 — Local Sharding (in-process): In progress
  - Core: handover + hysteresis implemented and tested.
  - Update: client `handover` event surfaced over WS; AOI rebuild handled via 3×3 cell `QueryAOI`.
  - Observability baseline added: tick time, snapshot bytes, AOI entity counts, WS connection gauge.
- M4 — Bots & Density Targets: US-401 Done; US-402 Pending
- M5 — Persistence: Pending

## Test Coverage
- Unit
  - `internal/spatial/spatial_test.go`: `WorldToCell`, 3×3 neighbors.
  - `internal/sim/handovers_test.go`: hysteresis thresholds on all borders.
  - `internal/sim/engine_test.go`: placement and kinematic integration.
  - `internal/join/join_test.go`: join success, auth failure, bad request, spawn at origin.
- Integration
  - `internal/transport/ws/ws_integration_test.go` (build-tagged `ws`): join, input, expect `state` with `ack >= 1` and positive motion.
  - `internal/transport/ws/telemetry_test.go` (build-tagged `ws`): telemetry includes `tick_rate` and reasonable `rtt_ms`.
  - `internal/transport/ws/cadence_test.go` (build-tagged `ws`): state cadence ~100ms ±20ms; payload <30KB/s.

Commands:
- Unit: `cd backend && go test ./...`
- WS: `cd backend && go test -tags ws ./...`
  - Includes `internal/transport/ws/handover_test.go` to validate `handover` emission.
  - Includes `internal/transport/ws/aoi_continuity_test.go` to validate AOI across handover.
- CI (GitHub): runs on push/PR to `main` with `go fmt`, `go vet`, unit tests, and ws-tagged tests; Go 1.23 with module/build caching and race-enabled matrix.

## Tooling Updates
- ENG-001: Added GitHub Actions workflow (`.github/workflows/ci.yml`) that formats, vets, and runs unit + ws-tagged tests.
- ENG-004: CI updated to Go 1.23 with module/build cache and a test matrix that includes race-enabled runs.
- ENG-002: Added PR template and CODEOWNERS under `.github/`.
- ENG-003: Enabled branch protection on `main` with required status checks.
 - Added Prometheus metrics endpoint at sim `/metrics`; JSON snapshot now at `/metrics.json`.
   - Metrics include: `sim_tick_time_ms`, `ws_snapshot_bytes`, `sim_entities_in_aoi`, `sim_handover_latency_ms`, `ws_connected`, `sim_handovers_total`.

## How to Drive the Sim (Dev)
See `docs/dev/DEV.md` for Makefile-based workflows.
Key commands:
- `make run` → start gateway and sim (WS enabled).
- `make login` → acquire a token.
- `make wsprobe TOKEN=... [MOVE_X=1 MOVE_Z=0]` → join and optionally send input.
 - `curl localhost:8081/metrics` → scrape Prometheus metrics; `curl localhost:8081/metrics.json` for dev JSON snapshot.

## Next Up
- M2 (AOI streaming): entity sets and cadence at 10 Hz; budget checks.
- M3: handover events surfaced over WS.
- M4: bot density controller (Done) and wander behavior (Pending).
- M5: persistence for position + simple stat.

## Decisions
- Sharding strategy: Phase A (local) first; Phase B (cross-node) post-MVP.
- Protocol: JSON over WebSocket for MVP; binary later.

---

## Done Stories — Checklist Backfill

### US-000 — Start services and health check
- Acceptance: Gateway and Sim start; `/healthz` returns <50ms locally. ✅
- Tests: Manual HTTP smoke; covered by `make run` + curl. ✅
- Docs/Tooling: README quick start; DEV guide and Makefile updated. ✅
- Format/Vet/Tests: `go fmt && go vet && go test ./...` clean. ✅

### US-101 — WebSocket handshake and join
- Acceptance: `hello { token }` → `join_ack { player_id, pos, cell, config }`; invalid → `error { code:"auth" }`. ✅
- Tests: `internal/join/join_test.go` covers success/auth/bad_request; E2E via `cmd/wsprobe`; WS integration test under `-tags ws`. ✅
- Docs/Tooling: Backlog marked Done; DEV guide includes WS run/probe; Makefile targets added. ✅
- Format/Vet/Tests: `make test` and `make test-ws` green. ✅

### US-102 — Spawn at default or last known location
- Acceptance: First login spawns at `(0,0)`; persistence deferred to M5. ✅
- Tests: Assert origin in `join_test.go` success path. ✅
- Docs/Tooling: Backlog marked Done with test note. ✅
- Format/Vet/Tests: Clean. ✅

### US-103 — Movement input and authoritative update
- Acceptance: Server integrates at 20 Hz; echoes latest `ack` in `state`. ✅
- Tests: Engine kinematics test; WS integration asserts `ack >= 1` and motion. ✅
- Docs/Tooling: Backlog marked Done; DEV guide and Makefile include probe and test instructions. ✅
- Format/Vet/Tests: `make fmt vet test` and `make test-ws` green. ✅
