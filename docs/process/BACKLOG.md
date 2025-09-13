# User Stories Backlog (MVP → Stretch)

This backlog turns the GDD/TDD into testable stories. Each story has clear acceptance criteria and a testing note to guide TDD.

## Status Board (Kanban Snapshot)
- In Progress
  - US-301 — Handover with hysteresis (M3)
- In Review
  - —
- Done
  - US-101 — WebSocket handshake and join (M1)
  - US-000 — Start services and health check (M0)
  - US-102 — Spawn at default or last known location (M1)
  - US-103 — Movement input and authoritative update (M1)
  - US-201 — AOI visibility by radius (M2)
  - US-202 — Snapshot cadence and payload budget (M2)
  - US-302 — Continuous AOI across borders (M3)
  - US-NF1 — Observability (Non-Functional)
  - ENG-001 — GitHub Actions CI (Tooling)
  - ENG-002 — PR template + CODEOWNERS (Tooling)
  - ENG-003 — Branch protection on main (Tooling)
- Not Started
  - US-401 — Maintain target density per cell (M4)
  - US-402 — Simple wander + separation (M4)
  - US-501 — Save position and simple stat (M5)
  - US-502 — Reconnect flow and session resume (M5)
  - US-601 — Cross-node transfer (Stretch)
  - US-NF2 — Security base (Non-Functional)

## Summary Table
| ID      | Milestone       | Title                                      | Status       |
|---------|------------------|--------------------------------------------|--------------|
| US-000  | M0               | Start services and health check            | Done         |
| US-101  | M1               | WebSocket handshake and join               | Done         |
| US-102  | M1               | Spawn at default or last known location    | Done         |
| US-103  | M1               | Movement input and authoritative update    | Done         |
| US-104  | M1               | Telemetry                                  | Done         |
| US-201  | M2               | AOI visibility by radius                   | Done         |
| US-202  | M2               | Snapshot cadence and payload budget        | Done         |
| US-301  | M3               | Handover with hysteresis                   | In Progress  |
| US-302  | M3               | Continuous AOI across borders              | Done         |
| US-401  | M4               | Maintain target density per cell           | Not Started  |
| US-402  | M4               | Simple wander + separation                 | Not Started  |
| US-501  | M5               | Save position and simple stat              | Not Started  |
| US-502  | M5               | Reconnect flow and session resume          | Not Started  |
| US-601  | Stretch          | Cross-node transfer                        | Not Started  |
| US-NF1  | Non-Functional   | Observability                              | Done         |
| US-NF2  | Non-Functional   | Security base                              | Not Started  |

## Engineering Tasks (Tooling)
| ID       | Area     | Title                               | Status       |
|----------|----------|-------------------------------------|--------------|
| ENG-001  | CI/CD    | GitHub Actions (fmt, vet, tests)    | Done         |
| ENG-002  | Hygiene  | PR template + CODEOWNERS            | Done         |
| ENG-003  | Repo     | Branch protection on `main`         | Done         |
| ENG-004  | CI/CD    | Go 1.23 CI matrix + race tests      | Done         |

Conventions
- IDs: `US-<milestone><seq>` (e.g., `US-301` belongs to M3).
- Status: Planned | In Progress | Done.
- Tests: unit/integration/perf references and suggested cases.

## M0 — Project Skeleton (Done)
- US-000: Start services and health check [Done]
  - As a dev, I can run `gateway` and `sim` and see `200 OK` on `/healthz`.
  - Acceptance: both services start without panic; health endpoints return in <50ms locally.
  - Tests: simple HTTP smoke (optional); covered by manual for now.

## M1 — Presence & Movement (Single Cell)
- US-101: WebSocket handshake and join
  - As a client, I connect via WS and send `hello { token }` to receive `join_ack { player_id, pos, cell, config }`.
  - Acceptance: valid token joins; invalid token → `error { code: "auth" }`; latency < 100ms locally.
  - Tests: WS loopback test; auth error path; manual E2E via gateway `/login` + sim `/ws` using `wsprobe`.

- US-102: Spawn at default or last known location
  - As a client, on join I spawn at my saved position or a default spawn for first login.
  - Acceptance: first login spawns at `(0,0)`; later logins use DB position (M5 dependency noted).
  - Tests: engine spawns at default (assert `(0,0)` in join ack); persistence path added in M5.

- US-103: Movement input and authoritative update
  - As a client, I send `input { seq, dt, intent }` and see smooth authoritative motion.
  - Acceptance: 20Hz tick; position integrates with server authority; server returns `ack` sequence in `state`.
  - Tests: engine kinematics unit tests; WS integration asserting increasing `ack` and motion (build with `-tags ws`).

- US-104: Telemetry
  - As a client, I receive periodic `telemetry { rtt, tick_rate }`.
  - Acceptance: tick_rate reflects configured tick; rtt matches ping/ack within ±10ms locally.
  - Tests: time-based assertions in WS test with tolerance.

## M2 — Interest Management (AOI Streaming)
- US-201: AOI visibility by radius
  - As a client, I see entities within `AOI_RADIUS` and not beyond.
  - Acceptance: AOI inclusion policy defined (inclusive at radius); add/remove sets stable while moving.
  - Tests: table-driven AOI edge cases; moving client with static entities → no flapping.

- US-202: Snapshot cadence and payload budget
  - As a client, I receive `state` at ~10Hz with reasonable payloads.
  - Acceptance: cadence 100ms ±20ms; average payload size budget <30KB/s per client (local).
  - Tests: metric capture in integration; assert cadence and payload averages.

## M3 — Local Sharding (In-Process Handover)
- US-301: Handover with hysteresis
  - As a player, crossing a cell border triggers a handover only after `H` meters into the new cell.
  - Acceptance: no thrash when pacing along the border; handover lat <250ms; state continuity.
  - Tests: unit (hysteresis) [exists]; engine integration (paced border); later WS handover event.

- US-302: Continuous AOI across borders
  - As a player, I continue to see entities in neighboring cells if within radius during/after handover.
  - Acceptance: AOI rebuild completes within next snapshot; no duplicate entity IDs.
  - Tests: AOI query fetching 3×3 cells; movement across border with static neighbors.

## M4 — Bots & Density Targets
- US-401: Maintain target density per cell
  - As a player, if my vicinity is underpopulated, bots appear to reach a configured minimum.
  - Acceptance: cell maintains target within ±20% within 10s; global bot cap respected.
  - Tests: density controller under churn; spawn/despawn bounds.

- US-402: Simple wander + separation
  - As a player, bots wander believably without clustering too tightly.
  - Acceptance: direction retarget 3–7s; speed clamped; separation when <2m.
  - Tests: deterministic RNG seed; step bot state; assert constraints.

## M5 — Persistence
- US-501: Save position and simple stat
  - As a player, my position and a simple progression stat persist on disconnect.
  - Acceptance: reconnect in <2s restores within 1m of last position; stat increments retained.
  - Tests: DB mock/fake; save/restore path; tolerance on position.

- US-502: Reconnect flow and session resume
  - As a player, I can resume a dropped session without duplicate entities.
  - Acceptance: session reuse or clean rejoin without ghost actors; no double streaming.
  - Tests: simulate drop; ensure single entity instance on resume.

## Stretch — Cross-Node Handover
- US-601: Cross-node transfer
  - As a player, crossing into a cell owned by another node hands me off seamlessly.
  - Acceptance: reconnect or tunneled handover <500ms; state continuity; no duplicates.
  - Tests: two sim processes in CI; RPC handshake; handover timing.

## Non-Functional Stories
- US-NF1: Observability
  - Metrics for `tick_time_ms`, `snapshot_bytes`, `entities_in_aoi`, `handover_latency_ms`, `ws_connected`.
  - Tests: scrape metrics in integration; sanity thresholds.

- US-NF2: Security base
  - Token-based auth, heartbeat timeouts, reject impossible velocities.
  - Tests: auth failure path; heartbeat disconnect; velocity clamps.

## Traceability to TDD
- M1 ↔ Network Protocol (hello/input/ack, join_ack), Tick & Reconciliation.
- M2 ↔ Interest Management, AOI query details.
- M3 ↔ Handover (Phase A), hysteresis.
- M4 ↔ Bots & Density Targets.
- M5 ↔ Persistence model and reconnect behavior.

## Implementation Notes
- Prefer adding tests alongside each story before wiring the full feature.
- Keep WS JSON messages minimal; evolve with versioned fields.
- Guard borders with hysteresis to prevent thrash; document policy in config.
