# Technical Design for M6 Equipment Foundation

**Design Date**: 2025-01-21  
**Implementation Target**: M6 Milestone  
**Dependencies**: Current state management system, PostgreSQL persistence layer

---

## Architecture Overview

The equipment system extends the existing `PlayerState` persistence model and integrates with the simulation engine's tick-based processing. It leverages the established patterns for state management, optimistic locking, and WebSocket communication.

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client UI     │    │  WebSocket       │    │  Simulation     │
│  - Equipment    │◄──►│  - Equip/Unequip │◄──►│  - State Mgmt   │
│  - Inventory    │    │  - State Updates │    │  - Validation   │
│  - Stat Display │    │  - Error Msgs    │    │  - Persistence  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                 │                        │
                                 ▼                        ▼
                        ┌──────────────────┐    ┌─────────────────┐
                        │  Equipment       │    │  PostgreSQL     │
                        │  Manager         │    │  - Items Table  │
                        │  - Templates     │◄──►│  - Equipment    │
                        │  - Validation    │    │  - Templates    │
                        │  - Calculations  │    │  - Player State │
                        └──────────────────┘    └─────────────────┘
```

---

## Data Models

### Equipment Schema (PostgreSQL)

```sql
-- Item templates define equipment properties
CREATE TABLE item_templates (
    id SERIAL PRIMARY KEY,
    item_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    item_type VARCHAR(20) NOT NULL, -- 'weapon', 'armor', 'tool'
    slot_type VARCHAR(20) NOT NULL, -- 'hands', 'head', 'chest', 'legs'
    stat_bonuses JSONB DEFAULT '{}',
    requirements JSONB DEFAULT '{}',
    cooldown_ms INTEGER DEFAULT 0,
    weight INTEGER DEFAULT 0,
    bulk INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Equipment cooldown tracking
CREATE TABLE equipment_cooldowns (
    player_id VARCHAR(50) NOT NULL,
    slot_type VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (player_id, slot_type)
);

-- Index for efficient cooldown queries
CREATE INDEX idx_equipment_cooldowns_expires ON equipment_cooldowns(expires_at);
```

### Go Data Structures

```go
// Equipment represents a single equipped item
type Equipment struct {
    ItemID      string            `json:"item_id"`
    SlotType    string            `json:"slot_type"`
    StatBonuses map[string]int    `json:"stat_bonuses"`
    Cooldown    *time.Time        `json:"cooldown_expires,omitempty"`
}

// EquipmentSlots manages all equipped items for a player
type EquipmentSlots struct {
    Hands  *Equipment `json:"hands,omitempty"`
    Head   *Equipment `json:"head,omitempty"`
    Chest  *Equipment `json:"chest,omitempty"`
    Legs   *Equipment `json:"legs,omitempty"`
}

// ItemTemplate defines equipment properties
type ItemTemplate struct {
    ItemID       string            `json:"item_id"`
    Name         string            `json:"name"`
    ItemType     string            `json:"item_type"`
    SlotType     string            `json:"slot_type"`
    StatBonuses  map[string]int    `json:"stat_bonuses"`
    Requirements map[string]int    `json:"requirements"`
    CooldownMS   int               `json:"cooldown_ms"`
    Weight       int               `json:"weight"`
    Bulk         int               `json:"bulk"`
}
```

---

## Equipment Manager Interface

```go
type EquipmentManager interface {
    // LoadTemplates loads item templates from database
    LoadTemplates(ctx context.Context) error
    
    // EquipItem attempts to equip an item to a slot
    EquipItem(playerID, itemID, slotType string) (*EquipResult, error)
    
    // UnequipItem removes an item from a slot
    UnequipItem(playerID, slotType string) (*UnequipResult, error)
    
    // GetEquipment returns current equipment for a player
    GetEquipment(playerID string) (*EquipmentSlots, error)
    
    // CalculateStats computes total stat bonuses from equipment
    CalculateStats(equipment *EquipmentSlots) map[string]int
    
    // ValidateEquipment checks if equipment state is valid
    ValidateEquipment(playerID string, equipment *EquipmentSlots) error
    
    // ProcessCooldowns updates cooldown timers (called each tick)
    ProcessCooldowns(now time.Time) error
}
```

---

## WebSocket Protocol Extensions

### Equip Item Request
```json
{
    "type": "equip_item",
    "data": {
        "item_id": "sword_iron_01",
        "slot": "hands"
    }
}
```

### Equip Item Response (Success)
```json
{
    "type": "equip_result",
    "data": {
        "success": true,
        "slot": "hands",
        "item": {
            "item_id": "sword_iron_01",
            "stat_bonuses": {"damage": 10, "speed": -2},
            "cooldown_expires": "2025-01-21T15:30:45Z"
        },
        "total_stats": {"damage": 15, "defense": 5, "speed": 8}
    }
}
```

### Equip Item Response (Failure)
```json
{
    "type": "equip_result",
    "data": {
        "success": false,
        "error": "insufficient_skill",
        "message": "Requires Weapon Skill level 5",
        "requirements": {"weapon_skill": 5}
    }
}
```

---

## Integration Points

### State Management
- **Current**: `PlayerState` struct in `internal/state/store.go`
- **Extension**: Add `EquipmentData` field (already exists as `json.RawMessage`)
- **Serialization**: JSON marshaling of `EquipmentSlots` to `EquipmentData`

### Simulation Engine
- **Hook Point**: `engine.RestorePlayerState()` and `AddOrUpdatePlayer()`
- **Tick Integration**: Equipment cooldown processing in main tick loop
- **Performance**: Equipment stat calculations cached, recalculated only on changes

### Persistence Layer
- **Store Interface**: Extend existing `Store` interface for equipment operations
- **PostgreSQL**: Implement equipment queries in `postgres_store.go`
- **Optimistic Locking**: Use existing version field for conflict resolution

---

## Performance Considerations

### Caching Strategy
```go
type EquipmentCache struct {
    templates    map[string]*ItemTemplate // Item templates
    playerStats  map[string]map[string]int // Calculated stats per player
    cooldowns    map[string]map[string]time.Time // Active cooldowns
    mutex        sync.RWMutex
}
```

### Tick Budget Allocation
- **Equipment Processing**: <5ms per tick for 100 players
- **Cooldown Updates**: Batch processing, lazy evaluation
- **Stat Calculations**: On-demand with caching

### Database Optimization
- **Read Queries**: Index on player_id, item_id, slot_type
- **Write Queries**: Batch updates for multiple equipment changes
- **Connection Pooling**: Reuse existing connection management

---

## Error Handling Strategy

### Validation Errors
```go
type EquipmentError struct {
    Type    string `json:"type"`    // "invalid_item", "insufficient_skill", etc.
    Message string `json:"message"` // Human-readable error
    Code    int    `json:"code"`    // Numeric error code
}
```

### Error Codes and Types
| Code | Type | Description | Retry Recommended |
|------|------|-------------|------------------|
| 1001 | `invalid_item` | Item ID not found in templates | No |
| 1002 | `insufficient_skill` | Player lacks required skill level | No |
| 1003 | `slot_mismatch` | Item incompatible with target slot | No |
| 1004 | `cooldown_active` | Equipment still in cooldown period | No |
| 1005 | `state_corruption` | Player equipment state corrupted | Yes |
| 1006 | `database_timeout` | Database operation timed out | Yes |
| 1007 | `optimistic_conflict` | Concurrent modification detected | Yes |
| 1008 | `template_unavailable` | Item templates not loaded | Yes |
| 1009 | `network_partition` | Database connectivity lost | Yes |
| 1010 | `cache_overflow` | Equipment cache at capacity | Yes |

### Error Matrix

The following matrix defines specific error conditions, their detection criteria, response procedures, and expected outcomes:

| Error Condition | Detection | Immediate Response | Recovery Procedure | Expected Outcome | Monitoring Alert |
|----------------|-----------|-------------------|-------------------|------------------|------------------|
| **Template Loading Failure** | `LoadTemplates()` returns error | Use cached templates, log ERROR | Retry every 30s for 5 attempts, then every 5min | Templates reload or fallback maintained | `equipment_template_load_failed` |
| **Database Connection Timeout** | Query timeout >5s | Queue operation, return temp error | Exponential backoff: 1s, 2s, 4s, then circuit breaker | Operation completes or fails gracefully | `equipment_db_timeout` |
| **Player State Corruption** | Invalid JSON or missing required fields | Isolate player, default equipment state | Restore from latest backup, notify admins | Player restored or isolated until manual fix | `equipment_state_corruption` |
| **Optimistic Lock Conflict** | Version mismatch on update | Return conflict error to client | Retry with fresh state, max 3 attempts | Update succeeds or operation rejected | `equipment_lock_conflict` |
| **Invalid Item ID** | Item not in template cache | Reject operation, log WARN | No retry, await client correction | Operation rejected with clear error | None (expected user error) |
| **Insufficient Skill Level** | Requirements check fails | Reject operation, return requirements | No retry, await character progression | Operation rejected with skill requirements | None (expected game rule) |
| **Slot Type Mismatch** | Item type incompatible with slot | Reject operation, log WARN | No retry, await client correction | Operation rejected with slot info | None (expected user error) |
| **Equipment Cooldown Active** | Current time < cooldown expiry | Reject operation, return remaining time | No retry, await cooldown expiration | Operation rejected with cooldown info | None (expected game rule) |
| **Network Partition** | Multiple consecutive DB failures | Enable degraded mode, cache-only | Sync when connectivity restored | Operations resume or manual intervention | `equipment_network_partition` |
| **Memory Cache Overflow** | Cache size > 100MB or >10k players | Start cache eviction, log WARN | LRU eviction of inactive players | Memory usage stabilized | `equipment_cache_overflow` |
| **Database Schema Mismatch** | Migration version conflict | Refuse to start, log FATAL | Manual intervention required | Service starts after schema fix | `equipment_schema_mismatch` |
| **Concurrent Equipment Operations** | Race condition detected | Serialize operations per player | Queue subsequent operations | All operations complete in order | None (handled internally) |

### Recovery Procedures

**1. Template Loading Failure**
- **Immediate**: Continue with cached templates if available
- **Short-term**: Retry loading every 30 seconds for 5 attempts
- **Long-term**: Retry every 5 minutes until success or manual intervention
- **Fallback**: Minimal default templates to prevent complete failure

**2. Database Timeout**
- **Immediate**: Queue operation in memory buffer (max 1000 operations)
- **Retry**: Exponential backoff: 1s, 2s, 4s, then circuit breaker for 30s
- **Circuit Breaker**: Fail fast for 30s, then attempt single test operation
- **Fallback**: Return cached state for read operations

**3. State Corruption**
- **Immediate**: Isolate affected player, log detailed corruption info
- **Recovery**: Attempt restore from latest backup within 5 minutes
- **Escalation**: Alert administrators if backup restore fails
- **Fallback**: Default empty equipment state to maintain player session

**4. Optimistic Lock Conflicts**
- **Immediate**: Return conflict error with current state version
- **Retry**: Client should fetch fresh state and retry operation
- **Limit**: Maximum 3 automatic retries with 100ms delays
- **Escalation**: Log if conflicts exceed 5% of operations

---

## Testing Strategy

### Unit Tests
- **Template Management**: Loading, validation, caching
- **Equipment Operations**: Equip, unequip, stat calculations
- **Error Conditions**: Invalid items, skill requirements, conflicts
- **Performance**: Stat calculation efficiency, memory usage

### Integration Tests
- **Database Persistence**: Equipment state across restarts
- **WebSocket Communication**: Protocol compliance, error handling
- **Concurrency**: Multiple players, simultaneous operations

### Load Tests
- **Throughput**: 100+ concurrent equip/unequip operations
- **Memory**: Equipment cache growth under load
- **Database**: Connection pool utilization, query performance

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [ ] Database schema creation and migration
- [ ] Basic `EquipmentManager` implementation
- [ ] Item template loading and caching
- [ ] Unit tests for core functionality

### Phase 2: Integration (Week 2)
- [ ] WebSocket protocol integration
- [ ] State persistence integration
- [ ] Cooldown processing in tick loop
- [ ] Integration tests

### Phase 3: Optimization (Week 3)
- [ ] Performance testing and optimization
- [ ] Error handling hardening
- [ ] Load testing validation
- [ ] Documentation completion

---

## Success Metrics

### Performance Targets
- Equipment operations complete within 100ms
- Tick processing impact <5ms for 100 players
- Database queries complete within 50ms (95th percentile)
- Memory usage increase <50MB for equipment system

### Reliability Targets
- Zero data corruption incidents
- 99.9% uptime for equipment operations
- Graceful degradation under database outages
- Automatic recovery from transient failures

---

## Migration Strategy

### Existing Data
- Current `EquipmentData` field is JSON placeholder
- No breaking changes to existing player state
- Backward compatibility with empty equipment state

### Deployment Plan
1. Deploy schema changes during maintenance window
2. Update application with equipment system disabled
3. Enable equipment system with feature flag
4. Monitor performance and error rates
5. Remove feature flag after validation

This design leverages existing architecture patterns while introducing the equipment system incrementally. The implementation follows established coding standards and maintains the project's high-quality testing approach.