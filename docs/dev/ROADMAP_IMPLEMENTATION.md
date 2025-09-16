# Roadmap Implementation Guide

This document outlines the technical implementation requirements for the "Full MVP Loop and Persistence" release based on roadmap planning meeting outcomes from Issue #109.

## Overview

**Release Theme**: "Full MVP Loop and Persistence"  
**Timeline**: 10 weeks (Sept 2025 - Dec 2025)  
**Core Milestones**: M5 (Inventory & Equipment), M6 (Targeting & Skills), M7 (Persistence DB)

## M5: Inventory & Equipment MVP (Weeks 1-3)

### Required Codebase Changes

#### 1. Backend Data Structures (`backend/internal/inventory/`)
```go
// New package: backend/internal/inventory/
type ItemTemplate struct {
    TemplateID   string             `json:"template_id"`
    DisplayName  string             `json:"display_name"`
    SlotMask     EquipmentSlotMask  `json:"slot_mask"`
    Weight       float64            `json:"weight"`
    Bulk         int                `json:"bulk"`
    DamageType   DamageType         `json:"damage_type"`
    SkillReq     map[string]int     `json:"skill_req"`
    StanzaHooks  map[string]any     `json:"stanza_hooks"`
}

type PlayerInventory struct {
    PlayerID     string                    `json:"player_id"`
    BagCapacity  InventoryCapacity        `json:"bag_capacity"`
    Items        []InventoryItem          `json:"items"`
    Equipment    map[SlotID]EquippedItem  `json:"equipment"`
}
```

#### 2. Equipment System (`backend/internal/equipment/`)
```go
// New package: backend/internal/equipment/
type EquipmentSlot struct {
    SlotID       SlotID    `json:"slot_id"`
    InstanceID   string    `json:"instance_id,omitempty"`
    CooldownUntil time.Time `json:"cooldown_until"`
}

func (e *Equipment) EquipItem(player *Player, item ItemInstance, slot SlotID) error {
    // Validate slot compatibility, skill requirements, cooldown
    // Update equipment state, apply stat modifiers
}
```

#### 3. Integration Points
- **State Management**: Extend `backend/internal/state/player.go` to include inventory
- **WebSocket Protocol**: Add inventory sync messages to WS transport
- **Persistence**: Prepare for M7 database schema requirements

### Testing Requirements
- Unit tests for inventory operations (add/remove/encumbrance)
- Equipment system tests (equip/unequip/cooldowns/skill gates)
- Integration tests with existing player state
- WebSocket message format validation

## M6: Targeting & Skills MVP (Weeks 4-6)

### Required Codebase Changes

#### 1. Targeting System (`backend/internal/targeting/`)
```go
// New package: backend/internal/targeting/
type Target struct {
    EntityID   string    `json:"entity_id"`
    TargetType TargetType `json:"target_type"`
    Position   spatial.Vec2 `json:"position"`
    LockTime   time.Time `json:"lock_time"`
}

type TargetingSystem struct {
    // Tab-target cycling, soft lock, range validation
}

func (ts *TargetingSystem) AcquireTarget(player *Player, direction spatial.Vec2) (*Target, error)
func (ts *TargetingSystem) CycleTarget(player *Player, direction CycleDirection) (*Target, error)
```

#### 2. Skills System (`backend/internal/skills/`)
```go
// New package: backend/internal/skills/
type SkillTree struct {
    SkillID     string           `json:"skill_id"`
    CurrentXP   int              `json:"current_xp"`
    Level       int              `json:"level"`
    Abilities   []UnlockedAbility `json:"abilities"`
}

type AbilityUse struct {
    AbilityID   string    `json:"ability_id"`
    TargetID    string    `json:"target_id"`
    Timestamp   time.Time `json:"timestamp"`
    XPGained    int       `json:"xp_gained"`
}
```

#### 3. Integration Points
- **Simulation Loop**: Integrate targeting into tick processing
- **Equipment Integration**: Skills affect equipment requirements
- **WebSocket Protocol**: Add targeting and skill progression messages

### Testing Requirements
- Targeting system tests (acquire/cycle/range validation)
- Skills progression tests (XP gain/level up/ability unlock)
- Equipment + skills integration tests
- Combat loop integration tests

## M7: Persistence DB Integration (Weeks 7-9)

### Required Codebase Changes

#### 1. Database Layer (`backend/internal/persistence/`)
```go
// New package: backend/internal/persistence/
type DatabaseConfig struct {
    PostgresURL string
    RedisURL    string
    MaxConns    int
    Timeout     time.Duration
}

type PlayerRepository interface {
    SavePlayer(ctx context.Context, player *Player) error
    LoadPlayer(ctx context.Context, playerID string) (*Player, error)
    SaveInventory(ctx context.Context, inv *PlayerInventory) error
    LoadInventory(ctx context.Context, playerID string) (*PlayerInventory, error)
}
```

#### 2. Schema Migrations (`backend/migrations/`)
```sql
-- Create initial tables for player state, inventory, equipment, skills
-- Migration versioning and rollback support
-- Indexes for performance (player_id, template_id, etc.)
```

#### 3. Reconnect Logic (`backend/internal/session/`)
```go
type ReconnectHandler struct {
    playerRepo PlayerRepository
    timeout    time.Duration // <2s target
}

func (rh *ReconnectHandler) RestorePlayerState(playerID string) (*Player, error) {
    // Load from database with <2s budget
    // Restore position, inventory, equipment, skills
    // Re-insert into simulation
}
```

#### 4. Integration Points
- **Gateway Service**: Add persistence config and database connections
- **Sim Service**: Integrate save/load during handovers and disconnects  
- **Configuration**: Environment variables for database URLs
- **Health Checks**: Database connectivity monitoring

### Infrastructure Requirements
- PostgreSQL setup for development and CI
- Redis cache for session management
- Database migration tool (golang-migrate or similar)
- Connection pooling and health monitoring

### Testing Requirements
- Database integration tests (save/load/migrations)
- Reconnect performance tests (<2s target)
- Data consistency tests during handovers
- Schema migration tests (up/down)

## Cross-Cutting Concerns

### Configuration Management
```go
// backend/internal/config/roadmap.go
type RoadmapConfig struct {
    InventoryEnabled  bool `env:"INVENTORY_ENABLED" default:"true"`
    TargetingEnabled  bool `env:"TARGETING_ENABLED" default:"true"`
    PersistenceMode   string `env:"PERSISTENCE_MODE" default:"memory"` // memory|postgres
    DatabaseURL       string `env:"DATABASE_URL"`
    RedisURL          string `env:"REDIS_URL"`
}
```

### WebSocket Protocol Extensions
```json
// New message types for inventory/equipment
{"type": "inventory_sync", "data": {"items": [...], "equipment": {...}}}
{"type": "equipment_change", "data": {"slot": "main_hand", "item": {...}}}

// New message types for targeting/skills  
{"type": "target_acquired", "data": {"target_id": "...", "position": [x,z]}}
{"type": "ability_used", "data": {"ability_id": "...", "xp_gained": 50}}
{"type": "skill_levelup", "data": {"skill_id": "...", "new_level": 5}}
```

### Observability & Metrics
```go
// Extend backend/internal/metrics/ with new metrics
var (
    InventoryOperations = prometheus.NewCounterVec(...)
    EquipmentChanges    = prometheus.NewCounterVec(...)
    TargetingActions    = prometheus.NewCounterVec(...)
    SkillProgressions   = prometheus.NewCounterVec(...)
    DatabaseLatency     = prometheus.NewHistogramVec(...)
    ReconnectDuration   = prometheus.NewHistogram(...)
)
```

## Build & Test Integration

### Makefile Targets
```makefile
# Add new test targets for roadmap features
test-inventory:
	cd backend && go test ./internal/inventory/... ./internal/equipment/...

test-targeting:
	cd backend && go test ./internal/targeting/... ./internal/skills/...

test-persistence:
	cd backend && go test ./internal/persistence/...

# Integration target
test-roadmap: test-inventory test-targeting test-persistence
```

### CI/CD Extensions
- Add PostgreSQL service to GitHub Actions
- Database migration tests in CI pipeline
- Performance regression tests for <2s reconnect target
- Integration test coverage for new subsystems

## Implementation Schedule

### Week 1-2: M5 Foundation
- [ ] Create inventory and equipment packages
- [ ] Implement core data structures and operations
- [ ] Add WebSocket protocol extensions
- [ ] Unit test coverage >80%

### Week 3: M5 Integration
- [ ] Integrate with player state management
- [ ] Add equipment stat effects to simulation
- [ ] End-to-end testing with WebSocket client
- [ ] Performance validation

### Week 4-5: M6 Foundation
- [ ] Create targeting and skills packages
- [ ] Implement tab-target cycling and soft lock
- [ ] Add skill progression and ability systems
- [ ] Integration with equipment requirements

### Week 6: M6 Integration
- [ ] Combat loop integration (targeting + skills + equipment)
- [ ] WebSocket protocol for targeting/skills
- [ ] Performance testing at scale
- [ ] User acceptance testing

### Week 7-8: M7 Foundation
- [ ] Database schema design and migrations
- [ ] PostgreSQL and Redis integration
- [ ] Player repository implementation
- [ ] Connection pooling and health checks

### Week 9: M7 Integration
- [ ] Reconnect logic with <2s target
- [ ] Save/load integration with simulation
- [ ] Data consistency during handovers
- [ ] Production readiness validation

### Week 10: Polish & Stretch
- [ ] Auth hardening (rate limiting, token lifecycle)
- [ ] Observability (metrics, tracing, dashboards)
- [ ] Cross-node handover preparation (stretch goal)
- [ ] Client plan finalization

## Success Validation

Each milestone should be validated against these criteria:

**M5 Validation:**
- [ ] Player can equip/unequip items with proper restrictions
- [ ] Inventory encumbrance affects movement speed
- [ ] Equipment changes reflect in WebSocket messages
- [ ] All unit and integration tests pass

**M6 Validation:**
- [ ] Tab-targeting cycles through nearby entities
- [ ] Ability use grants XP and skill progression
- [ ] Equipment requirements enforced by skill levels
- [ ] Combat loop functions end-to-end

**M7 Validation:**
- [ ] Player disconnection and reconnection <2s
- [ ] All state (position/inventory/skills) restored correctly
- [ ] Database handles concurrent connections
- [ ] Zero data loss during handovers

## Related Documentation

- **Design Documents**: `docs/design/TDD.md` - Technical architecture
- **Development Guide**: `docs/dev/DEV.md` - Build and test procedures  
- **Roadmap Planning**: `docs/roadmap/ROADMAP.md` - Project timeline
- **Issue Tracking**: [Issue #109](https://github.com/AstroSteveo/prototype-game/issues/109) - Roadmap planning outcomes

---

**Document Owner**: Technical Architect  
**Last Updated**: September 16, 2025  
**Source**: Roadmap Planning Meeting (Issue #109)