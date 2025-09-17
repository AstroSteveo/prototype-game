# ADR 0001: Local Sharding for World Cells

- **Status**: Accepted (2024-01-15)
- **Decision Date**: Early prototype phase
- **Related ADRs**: [0002-cross-node-handover-protocol.md](0002-cross-node-handover-protocol.md), [0003-distributed-cell-assignment.md](0003-distributed-cell-assignment.md)
- **Context**: 

Early prototypes needed a simple way to scale the world and limit broadcast scope.

**Problem Statement:**
- Single global state requires broadcasting all player updates to all connected clients
- No spatial optimization for area-of-interest (AOI) queries results in O(n²) complexity
- Need foundation for future multi-node scaling without premature complexity
- Must support initial MVP requirements for 10-50 concurrent players

**Requirements:**
- Support 10-50 concurrent players in MVP phase
- Maintain <16ms simulation tick rate under target load
- Limit AOI broadcast radius to spatially relevant players
- Establish patterns for future transition to distributed sharding (Phase B)
- Enable deterministic simulation behavior for consistent gameplay

**Constraints:**
- Must be implementable within prototype timeline
- Should not require external dependencies or distributed systems complexity
- Must integrate with existing WebSocket transport layer
- Should support configurable parameters for tuning

- **Decision**: 

Implement local sharding by partitioning the world into fixed-size grid cells managed within a single simulation process.

**Key Components:**
1. **Spatial Partitioning**: 256m × 256m cells (configurable via `--cell` flag)
2. **Area of Interest**: 3×3 cell grid centered on player's current cell
3. **Local Handovers**: In-memory transfers when players cross cell boundaries
4. **State Management**: Each cell maintains independent entity lists and spatial queries

**Design Rationale:**
- **Cell Size (256m)**: Balances granularity vs. overhead - typical walking speed crosses cell in ~3-4 minutes, reducing handover frequency while maintaining spatial efficiency
- **3×3 AOI**: Ensures players see entities up to ~384m radius (adequate for planned game mechanics and visual range)
- **Local-only handovers**: Minimize complexity while establishing sharding patterns for future distributed implementation
- **Fixed grid**: Simplifies spatial queries and enables predictable future distributed partitioning

**Alternatives Considered:**
- **Global simulation**: Rejected due to O(n²) update complexity at target scale
- **Dynamic spatial partitioning**: Rejected as too complex for MVP timeline  
- **Smaller cells (64m-128m)**: Rejected due to increased handover frequency
- **Larger cells (512m+)**: Rejected due to reduced spatial query efficiency
- **Hierarchical cell structures**: Deferred to future distributed scaling phase

**Implementation:**
- Core logic: `backend/internal/spatial/spatial.go` (cell math, 3×3 neighborhoods)
- Engine integration: `backend/internal/sim/engine.go` (handovers, AOI queries)
- Configuration: `backend/cmd/sim/main.go` (--cell, --aoi flags)
- Handover logic: `backend/internal/sim/handovers.go`

- **Consequences**: 

**Positive:**
- **Performance**: Reduces AOI complexity from O(n²) to O(k) where k is entities per cell neighborhood
- **Scalability**: Successfully supports 10-50 players with <16ms tick latency in testing
- **Isolation**: Cell-based state reduces inter-player interference and enables parallel processing potential
- **Foundation**: Establishes consistent patterns for future distributed sharding (see ADR 0002)
- **Operational Simplicity**: No network protocols, consensus, or distributed state management required
- **Configuration Flexibility**: Cell size tunable via command-line flags for different deployment scenarios

**Negative:**
- **Single Point of Failure**: All cells fail if simulation process crashes (mitigated by process monitoring)
- **Memory Constraints**: Limited by single-node memory capacity (~1000 players estimated ceiling)
- **Future Migration Debt**: Requires significant infrastructure changes for multi-node scaling
- **Cell Boundary Effects**: Players near edges may experience AOI discontinuities (addressed by 3×3 overlap)
- **Handover Latency**: While fast (~1ms), becomes critical bottleneck for future cross-node migrations

**Measured Impact:**
- AOI query performance: ~0.1ms average per query (vs. ~10ms projected for global)
- Handover frequency: ~0.5 handovers per player per minute in typical gameplay
- Memory efficiency: ~50KB per active cell
- Support validated: 50 concurrent players with 12ms average tick latency

**Future Considerations:**
- Cell size may need regional adjustment based on observed player density patterns
- AOI algorithm could benefit from distance-based filtering within cell neighborhoods  
- Handover latency monitoring becomes critical for cross-node migration planning
- State serialization format should be designed for future network transport compatibility

**Migration Path:**
This local sharding implementation directly enables the cross-node handover protocol (ADR 0002) and distributed cell assignment strategy (ADR 0003) by:
- Establishing cell-based state management patterns
- Defining handover mechanisms that can be extended to network boundaries
- Creating spatial partitioning that supports consistent hashing distribution
- Implementing AOI queries that work across process boundaries

