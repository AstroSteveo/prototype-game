# design.md

Purpose: Capture technical architecture, data flows, interfaces, and design decisions following Spec-Driven Workflow. Reflects current backend (`backend/`) with WebSocket transport and simulation engine.

## Overview
- Generated: 2025-09-19 (DESIGN pass)
- Related requirements: See `requirements.md` (R1–R23)
- Design Decision Records: TBD in `/docs/decisions` or `DECISIONS.md`

## Architecture Summary
- Components
  - Simulation Engine (`internal/sim`): tick loop, AOI, cells/handovers, bots density, player inventory/equipment, metrics hooks.
  - WebSocket Transport (`internal/transport/ws`): connection lifecycle, auth/join handshake, message protocol, input ingestion, state snapshots/deltas, telemetry, resume tokens, idle timeout.
  - Spatial (`internal/spatial`): cell math, neighbor queries, distance.
  - Persistence (`internal/state`): Store interface and Postgres implementation; persistence manager in engine for checkpoints/disconnect saves.
  - Auth/Join (`internal/join`): validates hello credentials, constructs join acknowledgement.
  - Metrics (`internal/metrics`): Prometheus counters/histograms.

## Component Interactions (Data Flows)
1) Connect + Join
   - Client → WS `hello` → `join.AuthService` → Engine creates/attaches player → WS replies `join_ack` (includes `playerID`, `resumeToken`).

2) Input → Movement
   - Client → WS `input {seq, dt, intent{x,z}}` → Engine `DevSetVelocity` → Engine `tick` applies velocity integration at `TickHz`.

3) State Snapshot + AOI
   - Engine at `SnapshotHz` → AOI query (3x3 cells) → WS sends `state {ack, player, entities, [inventory/equipment/skills deltas]}`. Metrics record entities count and payload bytes.

4) Handovers
   - Engine detects border crossing beyond hysteresis → updates owned cell, records `HandoverAt` → next snapshot emits `handover {from,to}` and latency metric.

5) Equipment Commands
   - Client → WS `equip/unequip` with `seq` → Engine validates (slot, skills, cooldown, presence) → success updates versions; WS replies `equipment_result {operation, slot, success, code, message}`. Duplicate seq ignored.

6) Persistence
   - On disconnect (WS done) → Engine requests disconnect persist (background context, timeout-bound). On demand: checkpoint and restore via store.

7) Telemetry
   - WS `ping` around 1Hz → RTT ms in `telemetry` messages; idle timer resets on activity.

## Public Interfaces & APIs

### WebSocket Protocol (JSON)
- `hello` (client→server)
  - `{ "type": "hello", "token": string, "resume": string?, "lastSeq": number? }`
- `join_ack` (server→client)
  - `{ "type": "join_ack", "data": { "playerID": string, "cell": {cx,cz}, "resumeToken": string } }`
- `error` (server→client)
  - `{ "type": "error", "data": { "code": string, "message": string } }`
- `input` (client→server)
  - `{ "type": "input", "seq": number, "dt": number, "intent": { "x": -1..1, "z": -1..1 } }`
- `state` (server→client)
  - `{ "type": "state", "data": { "ack": number, "player": {id,pos,vel}, "entities": [ {id,pos,vel,kind,name} ], "inventory"?: {...}, "equipment"?: {...}, "skills"?: {...} } }`
- `equip` (client→server)
  - `{ "type": "equip", "seq": number, "instance_id": string, "slot": string }`
- `unequip` (client→server)
  - `{ "type": "unequip", "seq": number, "slot": string, "compartment"?: string }`
- `equipment_result` (server→client)
  - `{ "type": "equipment_result", "data": { "operation": "equip"|"unequip", "slot": string, "success": bool, "code": string, "message": string } }`
- `handover` (server→client)
  - `{ "type": "handover", "data": { "from": {cx,cz}, "to": {cx,cz} } }`
- `telemetry` (server→client)
  - `{ "type": "telemetry", "data": { "tick_rate": number, "rtt_ms": number } }`

### Engine Interfaces (selected)
- `Engine.Start()/Stop(ctx)` — lifecycle; `Step(dt)` for tests.
- `Engine.AddOrUpdatePlayer(id,name,pos,vel)` → `*Player`
- `Engine.DevSetVelocity(id, vel)` → bool
- `Engine.QueryAOI(pos, radius, excludeID)` → `[]Entity`
- `Engine.EquipItem(id, instanceID, slot, now)` → error
- `Engine.UnequipItem(id, slot, compartment, now)` → error
- `Engine.SetPersistenceStore(store state.Store)`; `StartPersistence(ctx)`, `StopPersistence()`
- `Engine.RequestPlayerDisconnectPersist(ctx, playerID)`; `RequestPlayerCheckpoint(ctx, playerID)`
- `Engine.RestorePlayerState(playerID, persisted, templates)` → error
- `Engine.MetricsSnapshot()` → metrics struct

### Persistence Store
- `state.Store` (from `internal/state`): provides methods for saving and loading player state (inventory, equipment, skills) using a PostgreSQL backend. The implementation in `postgres_store.go` stores player state in a `player_state` table, utilizing JSONB columns to persist inventory, equipment, and skills data efficiently.

## Data Models (conceptual)
- Entity: `{ id: string, kind: enum{player,bot}, pos: {x,z}, vel: {x,z}, name?: string }`
- Player (authoritative): `Entity + { OwnedCell:{cx,cz}, PrevCell:{cx,cz}, Inventory, Equipment, Skills: map[string]int, InventoryVersion:int64, EquipmentVersion:int64, SkillsVersion:int64, HandoverAt: time }`
- Inventory: `{ Items: [ { Instance:{InstanceID, TemplateID, Quantity, Durability}, Compartment } ], CompartmentCaps: {...}, WeightLimit: number }`
- Equipment: `{ Slots: map[SlotID]EquippedItem, cooldowns per slot }`
- ItemTemplate: `{ ID, DisplayName, SlotMask, Weight, Bulk, DamageType, SkillReq: map[string]int }`
- Encumbrance: `{ CurrentWeight, WeightPct, MovementPenalty }`

## Error Handling Matrix (selected)
- WebSocket
  - Invalid hello → send `error(bad_request)` and close
  - Origin not allowed (prod) → refuse accept
  - Oversized message / read timeout → close connection
  - Idle timeout → disconnect with log
  - Ping failure → close connection
- Equip/Unequip
  - Illegal slot → `equipment_result{code: illegal_slot}`
  - Insufficient skill → `equipment_result{code: skill_gate}`
  - Slot on cooldown → `equipment_result{code: equip_locked}`
  - Item not found → `equipment_result{code: item_not_found}`
  - Duplicate seq → ignored (idempotent)
- Persistence
  - Save timeout/failure on disconnect → best-effort; log; do not block shutdown
  - Deserialize mismatch on restore → return error; keep player session running with defaults
- AOI/Handover
  - Floating-point boundary → epsilon tolerance to prevent flapping

## Unit & Integration Testing Strategy
- Unit tests (fast):
  - Spatial utilities: cell math, neighbors, distance
  - Engine tick: velocity integration, density maintenance boundaries
  - Handovers: hysteresis thresholds, anti-thrash logic, latency capture
  - PlayerManager: slot validation, skill gates, cooldowns, inventory mutations, encumbrance math
- Integration tests (WS):
  - Join + resume token validation
  - Input cadence → movement → snapshots with AOI entities
  - Equip/unequip flows with idempotent sequence handling
  - Handover event emission and latency measurement
  - Idle timeout and ping/pong behavior
- Persistence (if Postgres available):
  - Save on disconnect; checkpoint; restore correctness
- Metrics assertions where feasible (counts, histograms)

## Migration Notes (docs reset)
- This design replaces prior docs. Interface names reflect current code. Public WS schemas summarized above; validate client compatibility before deployment.

## Open Questions
- Exact auth semantics and required claims in `join.AuthService`.
- Store schema guarantees and idempotency for saves/checkpoints.
- Production origin allowlist beyond localhost and same-origin.
