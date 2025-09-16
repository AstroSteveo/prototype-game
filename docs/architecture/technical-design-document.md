# Technical Design Document (TDD)

For day-to-day dev commands (run, build, tests), see `docs/development/developer-guide.md`.

## Architectural Overview
- Gateway: handles auth, session lookup, and hands out connection details to the simulation service.
- Simulation service: server‑authoritative world tick, spatial partitioning (grid cells), interest management, and handovers.
- Persistence: PostgreSQL (players, profiles, progress), Redis or in‑memory cache for session + cell ownership.
- Protocol: WebSocket (JSON for MVP), upgradeable to binary later.

### Phase A Architecture (Local Sharding)

```
┌──────────────┐       WS        ┌────────────────┐
│   Client     │  <──────────►   │   Gateway      │
│  (Unity/etc) │                 │ (auth/session) │
└──────┬───────┘                  └──────┬────────┘
       │                                   │
       │  WS (token, input, acks)          │ REST/DB cache
       ▼                                   ▼
┌────────────────────────────────────────────────────┐
│               Simulation Service (one proc)        │
│  ┌───────────────┐   ┌───────────────┐            │
│  │  Cell cx,cz   │…  │  Cell cx+1,cz │  … local   │
│  │  (instance)   │   │  (instance)   │  instances │
│  └───────────────┘   └───────────────┘            │
│      ▲      ▲             ▲      ▲                │
│      │ AOI  │             │ AOI  │                │
│  ┌───────────────┐   ┌───────────────┐            │
│  │ AOI/Replicate │   │ Handover Mgr  │            │
│  └───────────────┘   └───────────────┘            │
└────────────────────────────────────────────────────┘
```

## Signature Feature Strategy: Sharding Early
Implement sharding in two phases to de‑risk complexity while validating the experience fast:
- Phase A (MVP): single process, multi‑cell instancing. Cells are “local instances” inside one simulation node; handover is an in‑process transfer.
- Phase B (post‑MVP stretch): multi‑node. Cells are assigned to nodes, and handover performs a cross‑node handshake.

## World & Spatial Partitioning
- Coordinates: meters in a 2D plane (X, Z) with Y reserved for later.
- Grid: fixed‑size square cells (suggested 256m). Unique cell key `(cx, cz)` where `cx = floor(x/256)`, `cz = floor(z/256)`.
- Ownership: each cell belongs to a local instance; adjacent cells may be grouped logically for perf but remain distinct for handovers.
- Interest radius (AOI): 128m (configurable). Entities outside are not streamed to the client.

### Cell Math Details
- World→cell: `cx = floor(x / CELL_SIZE)`, `cz = floor(z / CELL_SIZE)`.
- Cell bounds (inclusive lower, exclusive upper): `x ∈ [cx*CELL_SIZE, (cx+1)*CELL_SIZE)`.
- Neighbor set for AOI: 3×3 cells around player’s current cell; filter by Euclidean radius.
- Hysteresis band `H`: require position to be at least `H` meters past the border into the target cell before finalizing a handover (e.g., `H=2m`).

## Interest Management
- Spatial index: grid buckets keyed by `(cx, cz)`; AOI query pulls the 3×3 neighboring buckets around player’s cell and filters by radius.
- Replication: server sends periodic state deltas for entities within AOI.
- Rate: tick at 20 Hz (50ms); replicate snapshots at 10 Hz (100ms) for MVP; delta compression optional later.

### AOI Query Pseudocode
```
function queryAOI(player):
  centerCell = cellOf(player.pos)
  candidates = []
  for each n in neighbors3x3(centerCell):
    candidates += cellBuckets[n]
  return filter(candidates, dist(e.pos, player.pos) <= AOI_RADIUS)
```

## Inventory, Equipment, and Items
- Item templates are defined authoritatively with `weight`, `bulk`, `damage_type`, `slot_mask`, and optional `skill_req` metadata.[^1][^2][^3][^5]
- Player state persists `bag_capacity` (weight + bulk), compartment counts (backpack, belt, craft bag), and an `equipment` map keyed by slot id.
- Equip actions enforce cooldown locks so recently swapped items cannot be used for a short period.[^1]

### Item Template Schema
| Field | Type | Notes |
| --- | --- | --- |
| `template_id` | UUID | Primary key referenced by instances |
| `display_name` | text | Localized name for UI |
| `slot_mask` | bitset | Legal equipment slots (e.g., main_hand, off_hand, chest) |
| `weight` | numeric | Applies to encumbrance totals; overage slows movement before hard cap.[^2] |
| `bulk` | integer | Inventory slot usage; cannot exceed compartment limit.[^2] |
| `damage_type` | enum | `slash`, `pierce`, `blunt`, or `elemental` for resist calculations.[^5] |
| `skill_req` | jsonb | Minimum skill levels per discipline required to equip/use.[^3] |
| `stanza_hooks` | jsonb | Optional modifiers to stanza costs/effects when slotted.[^6] |

### Inventory Operations
- Bag compartments (`inventory_items`) track `player_id`, `template_id`, `instance_id`, `quantity`, and durability.
- Equip slots (`equipment_slots`) store `player_id`, `slot_id`, `instance_id`, and `cooldown_until` timestamp.[^1]
- Encumbrance is recomputed on item add/remove; crossing thresholds triggers movement penalties and server warnings.[^2]
- Tooltips fetch `skill_req` data so UI can gray out unusable items until prerequisites are met.[^3]

### Equip Flow (Server)
```go
func EquipItem(p *Player, item ItemInstance, slot SlotID) error {
    if !item.Template.Allows(slot) {
        return ErrIllegalSlot
    }
    if !p.MeetsSkillReq(item.Template.SkillReq) {
        return ErrSkillGate
    }
    if p.Equipment[slot].CooldownActive(now()) {
        return ErrEquipLocked
    }
    p.Equipment[slot] = EquippedItem{Instance: item, CooldownUntil: now().Add(equipLock)}
    p.Inventory.Remove(item.InstanceID)
    p.RecomputeStats()
    enqueueEvent(p, InventoryUpdate())
    enqueueEvent(p, EquipmentUpdate(slot))
    return nil
}
```

### Damage Types and Mitigation
- Incoming hits resolve against the defender’s equipped shield/armor mitigation profile (max vs. slash/pierce/blunt) with clamps applied per template.[^5]
- Stanza effects can inject additional damage riders or resistance buffs that are processed in the combat resolver.[^6]

## Targeting System
- Each connection tracks `current_target` and `target_history` with timestamps for audit.
- Server validates line‑of‑sight and AOI membership before accepting a target change; rejected targets emit a `target_error` event.
- Target metadata exposed to the client includes name, title, vitals, aggression flag, and region difficulty color band.[^4]

### Target Selection Flow
```go
func SetTarget(p *Player, target EntityID) {
    if target == p.CurrentTarget {
        return
    }
    if !IsVisible(p, target) {
        send(p.Conn, TargetError{"not_visible"})
        return
    }
    p.CurrentTarget = target
    send(p.Conn, TargetUpdate{ID: target, Snapshot: buildTargetSnapshot(target)})
}
```

- Difficulty color is computed from target level vs. regional band thresholds so the HUD mirrors the reference color table.[^4]
- When a target despawns or leaves AOI, the server clears `current_target` and emits a `target_clear` message.

## Handover (Phase A: Local)
1. Detect boundary crossing when player center moves into a new cell.
2. Serialize player state (id, pos/vel, orientation, emote state, simple stats).
3. Move server‑side ownership to new cell instance and update session record.
4. Send client a transparent `handover` event; keep same connection.

### Local Handover Pseudocode
```
if cellOf(player.pos) != player.ownedCell:
  if insideTargetBeyondHysteresis(player.pos, player.ownedCell):
    state = extractPlayerState(player)
    cells[player.ownedCell].remove(player.id)
    newCell = cellOf(player.pos)
    cells[newCell].add(player.id, state)
    player.ownedCell = newCell
    send(player.conn, { type: 'handover', from: oldCell, to: newCell })
```

## Handover (Phase B: Cross‑Node, Post‑MVP)
1. Current node contacts target node with a transfer token and serialized state.
2. Target reserves the player; gateway can instruct reconnect or tunnel the stream.
3. Client receives `handover_start` then `handover_complete` with minimal hitch.

## Network Protocol (MVP, WebSocket JSON)
- Client→Server
  - `hello { token }`
  - `input { seq, dt, intent: { move: { x,z }, look: { yaw }, emote?, hotbar? } }`
  - `ack { last_seq }`
  - `target_select { entity_id }`
  - `ability_use { stanza_id, target_id?, mode }`
  - `inventory_move { instance_id, from, to }`
- Server→Client
  - `join_ack { player_id, pos, cell, config, inventory, equipment, skills }`
  - `state { tick, entities: [ { id, type, pos, vel, yaw, name?, level_band? } ], removals: [id] }`
  - `inventory_update { add: [...], remove: [...], encumbrance }`
  - `equipment_update { slot, item?, stats }`
  - `target_update { id, name, title, vitals, difficulty_color }`
  - `skill_progress { skill_id, xp, rank, unlocked }`
  - `handover { from_cell, to_cell }`
  - `telemetry { rtt, tick_rate }`
  - `error { code, message }`

### Join Handler (Transport-Agnostic)
- Implement join logic as a pure function separate from the WebSocket transport to enable unit testing without a socket.
- Interface: `AuthService.Validate(token) -> (player_id, name, ok)`.
- On success: attach/create player at default or persisted position and return `join_ack { player_id, pos, cell, config }`.
- On failure: return `error { code: "auth" }`; on malformed input: `error { code: "bad_request" }`.

### WebSocket Transport
- Implement a WS endpoint at `/ws` that reads `hello`, calls the join handler, and replies with `join_ack` or `error`.
- To keep tests fast and avoid external deps by default, provide two builds:
  - Default build: registers a stub HTTP handler returning 501 (no WS).
  - `-tags ws`: enables the real implementation using `nhooyr.io/websocket`.
- Gateway exposes `/validate?token=...` used by sim via an HTTP `AuthService` implementation.

### Tick, Prediction, and Reconciliation
- Client sends `input` with `seq` numbers. Server applies authoritative physics each tick and echoes latest `ack_seq` in `state`.
- Client predicts locally; when receiving `state`, it rewinds to the last acknowledged input, reapplies unacknowledged inputs, and corrects.

### Tick Loop Pseudocode (Server)
```
loop at 20 Hz:
  dt = timeSinceLastTick()
  // 1) ingest inputs
  for each conn in connections:
    inputs = drainInputQueue(conn)
    applyInputsToActor(conn.player, inputs)
  // 2) integrate simulation
  for each cell in cells:
    updateCell(cell, dt)
  // 3) AOI + replication @ 10 Hz
  if snapshotDue():
    for each conn in connections:
      visible = queryAOI(conn.player)
      snapshot = buildDeltas(conn, visible)
      send(conn, { type: 'state', ...snapshot })
  // 4) housekeeping (handover checks, bot spawns)
  manageHandovers()
  maintainBotDensity()
```

## Skill Progression & XP Pipeline
- Each validated action produces an `XPEvent{skill_id, amount, source}` that feeds a per‑player queue.[^7]
- The queue batches events by skill and applies rested/bonus modifiers before persisting to `skill_line` rows.
- Rank‑up detection emits `skill_progress` messages and unlock notifications for newly available stanzas.[^6]
- Skill points are banked per discipline and surface in the UI for ability purchases or loadout swaps.[^6][^7]

### XP Application Pseudocode
```go
func ApplyXP(p *Player, events []XPEvent) {
    grouped := aggregateBySkill(events)
    for skillID, amount := range grouped {
        line := p.Skills[skillID]
        line.XP += amount
        for line.XP >= xpForNextRank(line.Rank) {
            line.XP -= xpForNextRank(line.Rank)
            line.Rank++
            unlock := discoverNewStanzas(line.Rank)
            line.UnlockedStanzas = append(line.UnlockedStanzas, unlock...)
            p.SkillPoints.Grant(skillID, rewardForRank(line.Rank))
            enqueueEvent(p, SkillProgress(line, unlock))
        }
        persistSkillLine(p.ID, line)
    }
}
```

## Data Model (MVP)
- Player: `id, name, last_pos, last_cell, last_seen, encumbrance_wt, encumbrance_bulk, skill_points(jsonb), current_target`.[^1][^2][^4][^7]
- Session: `player_id, conn_id, cell, last_seq, last_tick` (cached).
- Inventory item: `instance_id, player_id, template_id, quantity, durability, location(compartment), created_at`.[^2]
- Equipment slot: `player_id, slot_id, instance_id, cooldown_until`.[^1]
- Item template cache: `template_id, weight, bulk, slot_mask, damage_type, skill_req, stanza_hooks`.[^2][^3][^5][^6]
- Skill line: `player_id, skill_id, rank, xp, unlocked_stanzas[]`.[^6][^7]
- Ability stanza: `stanza_id, skill_id, cost, credit, modifiers`.[^6]
- Entity: `id, kind(player|bot), pos(x,z), vel(x,z), yaw, name?, level_band`.[^4]
- Cell: `key(cx,cz), instance_id, population, load`.
- Bot template/config: `behavior, speed, density_target, loot_table`.

### Configuration (Defaults)
- `CELL_SIZE = 256` meters
- `AOI_RADIUS = 128` meters
- `TICK_HZ = 20` (50ms)
- `SNAPSHOT_HZ = 10` (100ms)
- `HANDOVER_HYSTERESIS = 2` meters

## Services & Modules (Proposed)
- `gateway`: auth, session lookup, simulation address, stateless.
- `sim/engine`: tick loop, physics lite, AOI, replication.
- `sim/cells`: cell table, ownership, handover.
- `sim/bots`: spawner, simple wander behavior.
- `net/ws`: transport, message codecs, heartbeat.
- `store`: PostgreSQL models; `cache`: Redis/in‑mem.

### Module API Sketches (Pseudo‑TS)
```
interface CellKey { cx: number; cz: number }
interface Entity { id: string; kind: 'player'|'bot'; pos: Vec2; vel: Vec2; yaw: number; name?: string }

interface CellInstance {
  key: CellKey
  entities: Map<string, Entity>
  add(e: Entity): void
  remove(id: string): void
}

interface CellManager {
  getOrCreate(key: CellKey): CellInstance
  moveEntity(id: string, from: CellKey, to: CellKey, state: Entity): void
}

interface AOIIndex {
  buckets: Map<string, Set<string>> // key = `${cx},${cz}`
  query(pos: Vec2, r: number): string[] // entity ids
}

interface HandoverManager {
  checkAndHandover(playerId: string): void
}

interface BotSpawner {
  maintainDensity(cell: CellKey, target: number): void
}
```

## MVP Milestones & Acceptance Criteria
M0: Project skeleton
- Outcome: runnable gateway and simulation services scaffolding; local run scripts.

M1: Presence & Movement (single cell)
- Outcome: connect, spawn, move; see yourself replicated and latency/RTT.
- Criteria: 20 Hz tick; movement prediction client‑side OK; server authoritative reconciliation works.

M2: Interest Management (AOI streaming)
- Outcome: other entities in radius appear/disappear; delta updates.
- Criteria: ≤ 100ms snapshot cadence; no duplicate/removal flapping in steady movements.

M3: Local Sharding (multi‑cell handover in one process)
- Outcome: cross cell borders with a `handover` event; state continuity.
- Criteria: handover notification < 250ms; no entity duplication or loss; AOI re‑build within next snapshot.

M4: Bots & Density Targets
- Outcome: bots spawn/despawn to maintain target density per cell; wander behavior.
- Criteria: maintain configured min entities in a cell (within ±20%).

M5: Persistence
- Outcome: position, loadout (inventory + equipment), and per-skill progress saved; reconnect restores full state.
- Criteria: reconnect in < 2s; spawn within 1m of saved position with inventory and equipment parity and no XP loss.

Stretch: Cross‑Node Handover
- Outcome: two sim processes with cell ownership split; boundary crossing re‑homes the player.
- Criteria: reconnect or tunneled handover < 500ms; no state loss.

## Testing Strategy (TDD)
- Unit
  - Cell math: `world→cell`, neighbor lookup.
  - AOI membership: inclusion/exclusion edge cases; 8‑neighbor coverage.
  - Handover decision: thresholds, oscillation guard (hysteresis).
  - Inventory math: weight/bulk accumulation, slot legality, equip cooldown timers.[^1][^2]
  - Skill gating: verifying required skill levels prevent unauthorized equips/ability use.[^3][^6]
  - Target difficulty mapping: level band to color translation and serialization.[^4]
- Integration
  - Multi‑entity AOI streaming: add/remove sets stable over 1,000 ticks.
  - Local handover: position continuity, sequence number continuity.
  - Bot density: holds target under movement churn.
  - Inventory/equipment persistence: reconnect reproduces state and encumbrance metrics.
  - Skill progression pipeline: XP events apply correctly and unlock stanzas with notifications.[^6][^7]
  - Ability execution: damage type modifiers interact with mitigation tables as expected.[^5][^6]
- Soak/Perf (local tooling)
  - 200 simulated clients; CPU budget per tick; GC pauses.
  - Replication payload size budget (target < 30KB/s per client avg in MVP scene).
  - Hotbar spam with item swaps to measure equip cooldown enforcement and inventory diff bandwidth.[^1]

### Additional Test Cases
- AOI edges: entities exactly at `AOI_RADIUS` are consistently included/excluded (define inclusive policy).
- Border straddling: player pacing along border does not thrash handover due to hysteresis.
- Snapshot cadence: jitter under heavy load remains within ±20ms.
- Disconnect/reconnect: session resumes without duplicate player entities.
- Equip cooldown integrity: item swaps respect lockouts and prevent immediate ability use until timer expires.[^1]
- Encumbrance boundaries: crossing thresholds emits warnings and clamps movement appropriately.[^2]
- Skill respec edge cases: unlocking/removing stanzas does not orphan loadout bindings.[^6]

## Performance Budgets (MVP Targets)
- Tick: 20 Hz; server tick < 25ms at 200 entities in AOI.
- Handover: < 250ms local; < 500ms cross‑node (stretch).
- Bandwidth: < 30KB/s per client avg; spikes < 100KB/s.

## Observability & Telemetry
- Metrics: `tick_time_ms`, `snapshot_bytes`, `entities_in_aoi`, `handover_latency_ms`, `bot_count`, `ws_connected`, `ws_dropped`, `inventory_weight_pct`, `equip_cooldown_active`, `skill_xp_gain`, `target_switch_rate`.
- Traces: handover span with child spans for serialize→apply→notify plus equip/XP pipelines.
- Logs: rate‑limited info on handover start/complete; warn on AOI build > 10ms and on equip attempts blocked by skill/slot gates.[^1][^3]

## Security & Reliability (MVP)
- Server authoritative simulation; reject impossible velocities.
- Token‑based auth; WS heartbeat with kick on timeout.
- Crash recovery: session table restores cell, loadout, and skill progress.
- Equip and ability requests must pass server-side slot and skill gate validation before state mutation.[^1][^3]
- Inventory diffs are signed with monotonically increasing counters to prevent replay or duplication exploits.

## Bot Behavior (MVP)
- Wander: pick a random direction every 3–7 seconds, clamp speed, avoid leaving cell by turning along border.
- Avoidance: simple separation—steer away if another entity within 2m.
- Density control: PID‑lite adjustment—spawn/despawn one bot at a time per tick until within ±20% of target.

### Bot Pseudocode
```
function maintainBotDensity(cell, target):
  cur = countBots(cell)
  if cur < target.min: spawnBot(cell)
  if cur > target.max: despawnBot(cell)

function updateBot(bot, dt):
  repel = separationVec(bot, botsWithin(2m))
  if repel != zero:
    bot.dir = blend(bot.dir, repel)
    bot.retargetAt = now + rand(3..7)s
  if timeToRetarget(bot):
    bot.dir = randomUnitVector()
    bot.retargetAt = now + rand(3..7)s
  bot.vel = clamp(bot.dir * BOT_SPEED, BOT_SPEED)
  if willExitCell(bot.pos, bot.vel, dt): bot.dir = turnAlongBorder(bot.dir)
```

## Non‑Goals (MVP)
- Player-to-player trading or auction houses (bag is personal only).
- Advanced combat systems beyond modular stanzas (e.g., combo breakers, physics-based hit detection).
- Matchmaking/party tools, voice chat, or large-scale guild management.

## Risks & Mitigations
- Handover hitches: keep Phase A in‑process; defer cross‑node until stable.
- Bandwidth spikes: snapshot cadence caps; coarse delta until binary codec exists.
- Load skew: bot budgets per cell; simple cap to avoid stampedes.

## Open Questions
- Client tech stack: engine/runtime decision (e.g., Unity, Godot, custom).
- Transport options: stick with JSON WS for MVP or jump to binary early?
- How much client prediction to implement before combat exists?

## Operational Notes (Local Dev)
- Single binary/service for Phase A with config flags for `CELL_SIZE`, `AOI_RADIUS`, and tick/snapshot rates.
- Hot‑reloadable config during dev; debug overlay toggles via `dev/debug` messages.
- Replay capture: record last N seconds of inputs for deterministic bug repro.

## Definition of Done (per Story)
- Implementation matches acceptance criteria and is minimal/clear.
- Tests added and passing:
  - Unit tests for core logic (e.g., engine, math, join/auth paths).
  - Integration tests where applicable (e.g., WS under `-tags ws`).
- Tooling updated as needed (Makefile/scripts) and docs updated:
  - Backlog status moved; tests/evidence noted in the corresponding GitHub issue.
  - Developer commands or runbooks reflected in `docs/development/developer-guide.md`.
- Format and vet clean: `go fmt ./... && go vet ./...` with `go test ./...` green.
- Security/safety considerations addressed (validate inputs, avoid panics, respect build tags).

## References
[^1]: Ryzom Core – “How to change the items in your hands” help page. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/how_to_changeitemsinhand_en.html
[^2]: Ryzom Core – Craft tool item info highlighting weight/bulk limits. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_info_item_craft_tool_en.html
[^3]: Ryzom Core – Abilities and items guide noting skill requirements for equipment. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/abilities_item_step4_en.html
[^4]: Ryzom Core – Target interface documentation with level color coding. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_target_en.html
[^5]: Ryzom Core – Shield item info describing damage type mitigation. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_info_item_shield_en.html
[^6]: Ryzom Core – Stanza system overview for modular ability construction. https://github.com/ryzom/ryzomcore/blob/master/ryzom/client/data/gamedev/html/help/interf_info_sbrick_en.html
[^7]: Ryzom Core – Server command definitions showing per-skill XP and skill points. https://github.com/ryzom/ryzomcore/blob/master/ryzom/server/data_shard/egs/client_commands_privileges_open.txt

