# requirements.md

Purpose: Re-establish the source of truth using EARS notation after substantial changes. This was produced from a focused ANALYZE pass over core backend modules.

## Overview
- Generated: 2025-09-19
- Confidence Score: 83%
   - Rationale: High clarity from code/tests in `backend/internal/sim` and `backend/internal/transport/ws`. Moderate unknowns around full auth semantics (`internal/join`) and persistence schema (`internal/state`).
- File inventory summary (observed key modules):
   - `backend/internal/sim/engine.go`: tick loop, snapshotting, handovers, AOI, bots, inventory/equipment hooks, persistence hooks, metrics
   - `backend/internal/sim/player_manager.go`: templates, inventory/equipment operations, cooldowns, encumbrance, skill gates
   - `backend/internal/sim/handovers.go`: hysteresis-based cell handover + latency measurement
   - `backend/internal/transport/ws/register_ws.go`: WS handler options, input protocol, state streaming with deltas, equip/unequip commands, resume tokens, telemetry, idle timeout
   - `backend/internal/transport/ws/session.go`: resume token manager
   - `backend/internal/spatial/spatial.go`: cell math, neighbors (implied by usage)
   - `backend/internal/state/*.go`: persistence store interfaces and Postgres implementation (not deeply analyzed here)
   - `backend/internal/metrics/metrics.go`: counters, histograms for ticks, AOI, equip, snapshots, handovers, WS
   - `backend/internal/join/*.go`: auth/join handshake (not deeply analyzed here)

## Instructions
- Define each requirement using EARS: WHEN/WHILE/IF/THEN patterns.
- Each requirement is testable, unambiguous, feasible, necessary, and traceable to code/tests.

## Requirements (EARS)

Connectivity & session
1. WHEN a client connects to the WebSocket endpoint and sends a valid hello with credentials, THE SYSTEM SHALL authenticate via `join.AuthService`, initialize or attach to a player, and respond with `join_ack` including a `playerID` and a resume token.
2. IF the hello is invalid or authentication fails THEN THE SYSTEM SHALL send an error message and close the connection.
3. WHILE a client is idle longer than the configured `IdleTimeout`, THE SYSTEM SHALL disconnect the client.
4. WHEN a client reconnects with a valid resume token and `lastSeq`, THE SYSTEM SHALL accept the session and continue sequence acknowledgement from `lastSeq`; IF the token is invalid THEN THE SYSTEM SHALL ignore resume data.
5. IF a message exceeds the configured read limit (32KB) or per-message read deadline THEN THE SYSTEM SHALL terminate the read loop and close the connection gracefully.
6. WHERE `DevMode` is disabled, THE SYSTEM SHALL enforce allowed WebSocket origins (localhost and server host); WHERE `DevMode` is enabled, THE SYSTEM SHALL relax origin verification for local testing.

Movement, tick, and AOI
7. WHEN an `input` message arrives with intent vector `{x,z}`, THE SYSTEM SHALL clamp each component to [-1,1], compute player velocity as `moveSpeed*intent`, and update the authoritative player velocity.
8. WHILE the simulation is started, THE SYSTEM SHALL tick at `TickHz`, integrating player and bot positions over `dt` and recording tick duration metrics.
9. WHEN emitting periodic `state` at `SnapshotHz`, THE SYSTEM SHALL include entities within AOI radius of the player using a 3x3 cell neighborhood and exclude the querying player.
10. IF the AOI radius is non-positive THEN THE SYSTEM SHALL return an empty AOI entity list.

Cells and handovers
11. WHEN a player crosses a cell border beyond hysteresis `H` into a neighbor cell, THE SYSTEM SHALL perform a handover, update ownership, and emit a `handover` event on the subsequent snapshot while recording handover latency.
12. WHILE preventing cell thrashing, THE SYSTEM SHALL apply doubled hysteresis when the player returns to the previous cell.

Bots and density
13. WHILE the simulation runs, THE SYSTEM SHALL maintain bot population per cell within ±20% of `TargetDensityPerCell`, adjusting at approximately 1Hz and never exceeding `MaxBots` globally.

Inventory, equipment, and skills
14. WHEN an `equip` command is received, THE SYSTEM SHALL equip the specified item if: the item exists in inventory, the target slot is compatible with the item template, the player meets skill requirements, and the slot is not on cooldown; OTHERWISE THE SYSTEM SHALL send an `equipment_result` with an appropriate error code.
15. WHEN an `unequip` command is received, THE SYSTEM SHALL move the item from the slot back into the designated compartment (defaulting to `backpack`) if the slot is not on cooldown; OTHERWISE THE SYSTEM SHALL return an `equip_locked` error code.
16. WHEN inventory or equipment changes, THE SYSTEM SHALL increment version counters and include inventory/equipment deltas on the next state message.
17. WHILE reporting load, THE SYSTEM SHALL compute and return encumbrance including equipped item weights and a movement penalty derived from weight percentage.

Persistence
18. WHEN a persistence store is configured and a client disconnects, THE SYSTEM SHALL persist the player’s latest state within a bounded timeout.
19. WHEN a checkpoint is requested for a player, THE SYSTEM SHALL enqueue a save operation via the persistence manager.
20. WHEN restoring player state from persistence, THE SYSTEM SHALL deserialize inventory, equipment, and skills against known templates and apply to the authoritative player record.

Metrics and telemetry
21. WHILE operating, THE SYSTEM SHALL record metrics for handovers, AOI queries and returned entities, snapshot payload bytes, equip operation outcomes, WebSocket connections, and tick durations.
22. WHILE connected, THE SYSTEM SHALL ping approximately once per second and, IF ping fails within the deadline, THEN THE SYSTEM SHALL close the connection.

Development helpers (optional)
23. WHERE development helpers are enabled, THE SYSTEM SHALL support `DevSpawn`, `DevAddItemToPlayer`, `DevGivePlayerSkill`, and `DevSetVelocity` for testing.

## Edge Cases & Failure Modes
- Invalid hello payload or missing credentials → send error and close.
- Invalid or expired resume token → ignore resume fields; treat as fresh session.
- Duplicate command sequences (equip/unequip) → idempotently ignore repeats using processed sequence tracking.
- Equip errors:
   - Illegal slot → code `illegal_slot`.
   - Insufficient skill → code `skill_gate`.
   - Slot on cooldown → code `equip_locked`.
   - Item not found in inventory → code `item_not_found`.
- AOI radius ≤ 0 → return empty AOI list.
- Oversized message (>32KB) or read timeout → close connection.
- Idle timeout exceeded → disconnect client.
- Ping failure within 500ms budget → close connection.
- Handover near borders with floating-point noise → apply epsilon tolerance to AOI and hysteresis checks.
- Persistence unavailable or slow → best-effort save with timeout; do not block shutdown indefinitely.
- MaxBots reached → do not spawn new bots.
- Inventory capacity/weight exceeded when unequipping → unequip fails with add-to-inventory error.

## Dependencies
- Runtime
   - Go 1.23 (`backend/go.mod`)
   - Build tag `ws` to include WebSocket transport
- External libraries
   - `nhooyr.io/websocket` for WS + `wsjson`
   - `github.com/prometheus/client_golang` for metrics
   - `github.com/lib/pq` for Postgres (via state store) [indirect]
- Internal modules
   - `internal/sim` (engine, player manager, handovers, inventory/equipment)
   - `internal/transport/ws` (WS handler, session management)
   - `internal/spatial` (cell math, neighbors, distance)
   - `internal/state` (Store interface, Postgres impl)
   - `internal/join` (auth/join handshake)
   - `internal/metrics` (counters/histograms)

## Traceability
- Code mapping
   - R1–R6 → `internal/transport/ws/register_ws.go`, `session.go`, `join/*`
   - R7–R10 → `internal/sim/engine.go` (input handling) and AOI code
   - R11–R12 → `internal/sim/handovers.go`
   - R13 → `internal/sim/engine.go` density logic
   - R14–R17 → `internal/sim/player_manager.go`, `inventory.go`, `equipment`
   - R18–R20 → `internal/sim/*persistence*`, `internal/state/*`
   - R21–R22 → `internal/metrics/metrics.go`, WS telemetry
   - R23 → `internal/sim/engine.go` dev helpers
- Design mapping (to be authored in `design.md`)
   - Architecture: Engine↔WS, Persistence, Metrics, Spatial
   - Data flow: Input→Engine→State, Engine→AOI→WS, Equip/Unequip paths
   - Interfaces: WS protocol schemas, Store API, Player/Inventory structures
- Tasks mapping (to be authored in `tasks.md`)
   - Validation tasks per requirement with unit/integration tests
   - Hardening tasks (origin policy, idle timeout tuning, persistence retries)

## Notes
- This document will drive the DESIGN phase (`design.md`) and the implementation plan (`tasks.md`).
