# Roadmap Implementation Playbook

Use this playbook to translate roadmap milestones into concrete delivery plans. Adapt the checklists to the current release theme and track completion inside the associated GitHub issues or project board.

## Getting Started
1. **Confirm scope** — Align on the milestones captured in the [roadmap handbook](roadmap.md) and record success metrics.
2. **Assemble owners** — Assign a lead for each capability (simulation, persistence, networking, UX) plus a coordinator for cross-cutting work.
3. **Snapshot baseline** — Capture test results (`make fmt vet test test-ws`), current metrics, and any open ADRs so regressions are easy to detect.

## Capability Workstreams
The table below provides ready-made workstreams for the features currently defined in the Technical Design Document. Swap or extend rows as the roadmap evolves.

| Milestone | Objective | Implementation Starters | Validation Checklist |
|-----------|-----------|-------------------------|----------------------|
| Inventory & Equipment | Authoritative inventory limits, equip slots, cooldown enforcement | Create `backend/internal/inventory` data models, integrate equip commands, extend persistence hooks | Unit tests for encumbrance, equip gating, AOI payload updates; WebSocket payload diff coverage |
| Targeting & Skills | Player targeting pipeline plus per-skill XP progression | Add targeting package with cycle/visibility rules, implement XP queues, surface ability unlocks | Integration tests for target acquisition, XP application, and telemetry; ensure UI payload shapes stay stable |
| Persistence & Reconnect | Durable state save/restore with <2s reconnect budget | Introduce persistence interface (`PlayerRepository`), wire database/Redis config, add reconnect handler | Migration tests, reconnect benchmarks, failover drills, observability hooks for reconnect duration |

Update or replace the rows to match the active milestones. Each workstream should link back to relevant ADRs or issues.

## Delivery Rhythm
1. **Design deep dives** — Facilitate lightweight design reviews for each milestone and capture decisions in ADRs when architecture shifts.
2. **Implementation slices** — Break milestones into vertical slices that walk data from API/transport through the simulation and persistence layers.
3. **Demo cadence** — Schedule midpoint and end-of-milestone demos to surface integration risks early.
4. **Instrumentation** — Expand metrics and logs as functionality grows; record additions in [`development/server-feature-test-plan.md`](../../development/server-feature-test-plan.md).

## Testing Strategy
- **Unit**: Prioritize deterministic tests for state transitions (inventory ops, targeting selection, XP accrual).
- **Integration**: Exercise WebSocket flows under the `ws` build tag to verify payload continuity and reconnect behavior.
- **Performance**: Capture reconnect latency, snapshot sizes, and bot density stats. Document results in the roadmap update log.
- **Regression**: Update fixtures and golden data whenever message schemas change. Coordinate with client consumers when breaking changes are unavoidable.

## Cross-Cutting Concerns
### Configuration
Keep feature flags and configuration centralized. Example structure:
```go
// backend/internal/config/features.go
type FeatureConfig struct {
    InventoryEnabled bool   `env:"INVENTORY_ENABLED" default:"true"`
    TargetingEnabled bool   `env:"TARGETING_ENABLED" default:"true"`
    PersistenceMode  string `env:"PERSISTENCE_MODE" default:"memory"` // memory|postgres
    DatabaseURL      string `env:"DATABASE_URL"`
    RedisURL         string `env:"REDIS_URL"`
}
```
Document new flags in the developer guide.

### Protocol Extensions
Whenever the WebSocket contract expands, capture sample payloads and bump schema versions when necessary.
```json
{"type": "inventory_sync", "data": {"items": [...], "equipment": {...}}}
{"type": "target_acquired", "data": {"target_id": "...", "position": [x,z]}}
{"type": "skill_levelup", "data": {"skill_id": "...", "new_level": 5}}
```
List required client updates in roadmap issues so downstream teams can plan.

### Observability
Add counters and histograms for new systems, e.g.:
```go
var (
    InventoryOperations = prometheus.NewCounterVec(...)
    TargetingActions    = prometheus.NewCounterVec(...)
    ReconnectDuration   = prometheus.NewHistogram(...)
)
```
Remember to register metrics and document dashboards or alerts tied to roadmap goals.

## Exit Criteria
Before closing a milestone:
- ✅ Acceptance criteria satisfied in the [Technical Design Document](../../architecture/technical-design-document.md).
- ✅ Tests updated and passing (`make fmt vet test test-ws`).
- ✅ Metrics and dashboards reflect new functionality with agreed-upon SLOs.
- ✅ Documentation refreshed (developer guide, roadmap handbook, release readiness template if applicable).

Keep this playbook synchronized with roadmap updates so that execution plans always trace back to the current strategy.
