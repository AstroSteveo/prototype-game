# ADR 0003: Distributed Cell Assignment Strategy

- **Status**: Proposed
- **Context**: 

With Phase B (cross-node sharding) implementation, we need a strategy for assigning cells to simulation nodes. Currently, in Phase A, all cells exist within a single process and are created on-demand as players enter them. In a distributed system, we must decide which node owns which cells and how to handle dynamic reassignment for load balancing.

**Key Requirements:**
1. **Deterministic Assignment**: Given a cell key (cx, cz), any component should be able to determine the owning node
2. **Load Balancing**: Even distribution of cells and player load across available nodes
3. **Fault Tolerance**: Cell ownership must survive node failures and network partitions
4. **Scalability**: Addition/removal of nodes should require minimal cell migrations
5. **Locality**: Adjacent cells should prefer being on the same node (for AOI efficiency)

**Current Architecture:**
- Cells are created lazily when first player enters (`Engine.getOrCreateCell()`)
- No concept of cell "ownership" beyond local process scope
- Gateway has no knowledge of cell-to-node mapping
- Session state tracks `last_cell` but no owning node information

**Design Constraints:**
- Must integrate with existing cross-node handover protocol (ADR 0002)
- Should minimize AOI queries across network boundaries
- Gateway must be able to route new players to correct nodes
- Cell reassignment must not interrupt active players

- **Decision**: 

Implement a **hybrid consistent hashing with locality optimization** approach:

## 1. Primary Assignment: Consistent Hashing

Use **consistent hashing with virtual nodes** for the primary cell assignment algorithm:

```go
type CellAssignment struct {
    ring      *consistent.Consistent  // Virtual node ring
    nodes     map[string]*SimNode     // node_id -> node info
    replicas  int                     // Virtual nodes per physical node
}

func (ca *CellAssignment) GetOwnerNode(cellKey spatial.CellKey) string {
    hashKey := fmt.Sprintf("%d,%d", cellKey.Cx, cellKey.Cz)
    return ca.ring.Get(hashKey)
}
```

**Benefits:**
- Deterministic assignment (any component can compute cell ownership)
- Minimal reshuffling when nodes are added/removed (~1/N cells migrate)
- Load distribution improves as cluster size increases
- No single point of failure for assignment decisions

## 2. Locality Optimization: Region Clustering

**Override consistent hashing for adjacent cell clusters** when beneficial:

```go
type RegionCluster struct {
    NodeID    string              `json:"node_id"`
    Cells     []spatial.CellKey   `json:"cells"`
    Players   int                 `json:"players"`
    CreatedAt time.Time           `json:"created_at"`
}

// Cluster adjacent cells on same node when they have active players
func (ca *CellAssignment) OptimizeForLocality(cells []spatial.CellKey) {
    // Group cells into regions where players interact across cell boundaries
    // Assign entire regions to single nodes to minimize cross-node AOI queries
}
```

**Clustering Rules:**
- Cluster cells when > 50% of players have cross-cell AOI overlap
- Maximum cluster size: 9 cells (3x3 grid) to prevent hotspots
- Revert to hash assignment when cluster load exceeds node capacity

## 3. Dynamic Load Rebalancing

**Trigger cell migrations** based on load metrics:

```go
type LoadMetrics struct {
    PlayerCount   int     `json:"player_count"`
    TickLatency   float64 `json:"tick_latency_ms"`
    AOIQueries    int     `json:"aoi_queries_per_sec"`
    CPUUsage      float64 `json:"cpu_usage_pct"`
    MemoryUsage   float64 `json:"memory_usage_pct"`
}

// Rebalance triggers
const (
    HIGH_LOAD_THRESHOLD    = 0.85  // 85% CPU/memory usage
    HOTSPOT_THRESHOLD      = 100   // 100+ players in single cell
    LATENCY_THRESHOLD      = 40.0  // 40ms tick latency
)
```

**Migration Algorithm:**
1. **Identify Overloaded Nodes**: Nodes exceeding HIGH_LOAD_THRESHOLD
2. **Select Migration Candidates**: Cells with lowest activity that maintain locality
3. **Find Target Nodes**: Nodes with available capacity (< 70% threshold)
4. **Execute Gradual Migration**: Move 1 cell at a time, monitor impact

## 4. Gateway Integration

**Gateway Cell Routing Table** for new player placement:

```go
type CellRouting struct {
    cellToNode map[spatial.CellKey]string    // cell -> owning node
    nodeInfo   map[string]*NodeInfo          // node health and endpoints
    lastUpdate time.Time                     // routing table freshness
}

func (g *Gateway) RoutePlayer(playerPos spatial.Vec2) (*NodeInfo, error) {
    cellKey := spatial.WorldToCell(playerPos.X, playerPos.Z, CELL_SIZE)
    nodeID := g.routing.GetOwnerNode(cellKey)
    return g.routing.nodeInfo[nodeID], nil
}
```

**Routing Table Updates:**
- Nodes broadcast cell ownership changes to gateway
- Gateway polls nodes every 30s for health status
- Fallback to consistent hash calculation if routing table stale

## 5. Fault Tolerance and Split-Brain Prevention

**Node Failure Handling:**
```go
// When node fails, redistribute its cells using consistent hashing
func (ca *CellAssignment) HandleNodeFailure(failedNodeID string) []CellMigration {
    failedCells := ca.getCellsOwnedBy(failedNodeID)
    migrations := make([]CellMigration, 0, len(failedCells))
    
    for _, cell := range failedCells {
        // Remove failed node from ring temporarily
        ca.ring.Remove(failedNodeID)
        newOwner := ca.ring.Get(cellKeyHash(cell))
        ca.ring.Add(failedNodeID) // Re-add for other calculations
        
        migrations = append(migrations, CellMigration{
            Cell:       cell,
            FromNode:   failedNodeID,
            ToNode:     newOwner,
            Reason:     "node_failure",
            Priority:   "urgent",
        })
    }
    return migrations
}
```

**Split-Brain Prevention:**
- Use gateway as authoritative source for cell assignments
- Nodes cache assignments but defer to gateway on conflicts
- Implement lease-based ownership (nodes must renew cell ownership every 60s)

## 6. Storage and Persistence

**Assignment State Storage:**
```json
{
  "cell_assignments": {
    "0,0": {"node": "sim-node-1", "assigned_at": "2024-01-15T10:00:00Z"},
    "0,1": {"node": "sim-node-1", "assigned_at": "2024-01-15T10:00:00Z"},
    "1,0": {"node": "sim-node-2", "assigned_at": "2024-01-15T10:00:00Z"}
  },
  "node_status": {
    "sim-node-1": {"status": "healthy", "last_heartbeat": "2024-01-15T10:05:00Z"},
    "sim-node-2": {"status": "healthy", "last_heartbeat": "2024-01-15T10:05:00Z"}
  },
  "cluster_config": {
    "virtual_nodes_per_physical": 100,
    "max_cluster_size": 9,
    "rebalance_threshold": 0.85
  }
}
```

**Storage Options:**
- **Phase 1**: Redis for shared state (simple, fast, atomic operations)
- **Phase 2**: Etcd for production (stronger consistency, leader election)
- **Phase 3**: Custom Raft implementation (no external dependencies)

- **Consequences**: 

**Positive:**
- **Scalability**: Linear scaling by adding more simulation nodes
- **Load Distribution**: Automatic spreading of cells across available capacity
- **Locality Optimization**: Reduces cross-node AOI queries for clustered players
- **Fault Tolerance**: Automatic reassignment when nodes fail
- **Operational Flexibility**: Can drain nodes for maintenance by migrating cells

**Negative:**
- **Complexity**: Significantly more complex than single-node assignment
- **Consistency Challenges**: Race conditions between gateway and simulation nodes
- **Migration Overhead**: Cell reassignments temporarily increase handover latency
- **Storage Dependency**: Requires shared storage (Redis/Etcd) for cluster coordination
- **Debugging Difficulty**: Load balancing bugs hard to reproduce and diagnose

**Performance Impact:**
- **Gateway Latency**: +5-10ms for cell-to-node lookup on player join
- **Cross-Node AOI**: 10-50ms penalty for players near cell boundaries on different nodes
- **Migration Cost**: 100-500ms service disruption per cell during reassignment
- **Storage Load**: ~1KB per active cell, ~100 ops/sec for typical cluster

**Risk Mitigation:**
- Start with consistent hashing only (no locality optimization)
- Use feature flags to enable dynamic rebalancing gradually
- Implement comprehensive monitoring for assignment decisions
- Create manual override tools for operational emergencies
- Design assignment logs for post-mortem analysis

**Implementation Phases:**
1. **Phase 1**: Basic consistent hashing with static node assignments
2. **Phase 2**: Dynamic node joining/leaving with automatic cell migration
3. **Phase 3**: Load-based rebalancing and locality optimization
4. **Phase 4**: Advanced cluster management (auto-scaling, predictive migration)

**Integration Points:**
- **ADR 0002 Cross-Node Handover**: Cell assignment determines when cross-node handovers are needed
- **Gateway Routing**: Gateway uses assignment table to route new players to correct nodes
- **Session Management**: Player session must track both cell and owning node
- **Monitoring**: Cell assignment changes must be observable for debugging

**Alternative Approaches Considered:**
- **Centralized Assignment Service**: Rejected due to single point of failure
- **Node-to-Node Negotiation**: Rejected due to complexity and inconsistency risks
- **Geographic Partitioning**: Rejected due to hotspot risks and inflexibility
- **Pure Round-Robin**: Rejected due to poor locality and uneven load distribution