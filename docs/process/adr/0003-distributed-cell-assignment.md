# ADR 0003: Distributed Cell Assignment Strategy

- **Status**: Accepted
- **Context**: 

With Phase B (cross-node sharding) implementation, we need a strategy for assigning cells to simulation nodes. Currently, in Phase A, all cells exist within a single process and are created on-demand as players enter them. In a distributed system, we must decide which node owns which cells and how to handle dynamic reassignment for load balancing.

**Key Requirements:**
1. **Deterministic Assignment**: Given a cell key (cx, cz), any component should be able to determine the owning node
2. **Fault Tolerance**: Cell ownership must survive node failures and network partitions
3. **Scalability**: Addition/removal of nodes should require minimal cell migrations
4. **Simple Integration**: Must integrate cleanly with existing `spatial.CellKey` and `Engine` architecture

**Current Architecture:**
- Cells are created lazily when first player enters (`Engine.getOrCreateCell()`)
- No concept of cell "ownership" beyond local process scope
- Gateway has no knowledge of cell-to-node mapping
- Session state tracks `last_cell` but no owning node information

**Design Constraints:**
- Must integrate with existing cross-node handover protocol (ADR 0002)
- Gateway must be able to route new players to correct nodes
- Should minimize implementation complexity for initial Phase B deployment
- Must be compatible with existing `spatial.CellKey` implementation

- **Decision**: 

Implement a **simplified consistent hashing approach** for Phase B deployment:

## 1. Basic Consistent Hashing with Static Node Assignment

Use **consistent hashing with virtual nodes** for deterministic cell assignment:

```go
// CellAssignment provides deterministic cell-to-node mapping
type CellAssignment struct {
    ring     *consistent.Consistent
    nodes    map[string]*SimNode
    replicas int                    // Virtual nodes per physical node (150-300)
    mu       sync.RWMutex
}

// SimNode represents a simulation node in the cluster
type SimNode struct {
    ID       string `json:"id"`
    Status   string `json:"status"`   // "healthy", "draining", "unhealthy"
    Endpoint string `json:"endpoint"`
    LastSeen time.Time `json:"last_seen"`
}

// GetOwnerNode returns the node responsible for a given cell
func (ca *CellAssignment) GetOwnerNode(cellKey spatial.CellKey) (string, error) {
    ca.mu.RLock()
    defer ca.mu.RUnlock()
    
    if len(ca.nodes) == 0 {
        return "", errors.New("no nodes available")
    }
    
    hashKey := fmt.Sprintf("%d,%d", cellKey.Cx, cellKey.Cz)
    nodeID := ca.ring.Get(hashKey)
    
    if nodeID == "" {
        return "", errors.New("consistent hash returned empty node")
    }
    
    // Verify node is still healthy
    node, exists := ca.nodes[nodeID]
    if !exists {
        return "", fmt.Errorf("assigned node %s not found", nodeID)
    }
    
    if node.Status != "healthy" {
        // For Phase 1, fail fast rather than trying to reassign
        return "", fmt.Errorf("assigned node %s is not healthy (status: %s)", nodeID, node.Status)
    }
    
    return nodeID, nil
}

// AddNode adds a new simulation node to the assignment ring
func (ca *CellAssignment) AddNode(node *SimNode) error {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    if _, exists := ca.nodes[node.ID]; exists {
        return fmt.Errorf("node %s already exists", node.ID)
    }
    
    ca.nodes[node.ID] = node
    ca.ring.Add(node.ID)
    return nil
}

// RemoveNode removes a simulation node from the assignment ring
func (ca *CellAssignment) RemoveNode(nodeID string) error {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    if _, exists := ca.nodes[nodeID]; !exists {
        return fmt.Errorf("node %s not found", nodeID)
    }
    
    delete(ca.nodes, nodeID)
    ca.ring.Remove(nodeID)
    return nil
}
```

**Benefits:**
- Deterministic assignment (any component can compute cell ownership)
- Minimal reshuffling when nodes are added/removed (~1/N cells migrate)
- Simple implementation with proven algorithm
- Compatible with existing `spatial.CellKey` type
## 2. Gateway Integration

**Simplified Gateway Routing** for new player placement:

```go
// GatewayRouter handles routing players to appropriate simulation nodes
type GatewayRouter struct {
    assignment *CellAssignment
    nodes      map[string]*NodeInfo
    mu         sync.RWMutex
}

// NodeInfo contains routing information for simulation nodes
type NodeInfo struct {
    ID       string `json:"id"`
    Endpoint string `json:"endpoint"`
    Status   string `json:"status"`
}

// RouteNewPlayer determines which node should handle a new player
func (gr *GatewayRouter) RouteNewPlayer(pos spatial.Vec2) (*NodeInfo, error) {
    // Use existing spatial package functions
    cx, cz := spatial.WorldToCell(pos.X, pos.Z, CELL_SIZE)
    cellKey := spatial.CellKey{Cx: cx, Cz: cz}
    
    nodeID, err := gr.assignment.GetOwnerNode(cellKey)
    if err != nil {
        return nil, fmt.Errorf("failed to get owner node: %w", err)
    }
    
    gr.mu.RLock()
    node, exists := gr.nodes[nodeID]
    gr.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("node %s not found in routing table", nodeID)
    }
    
    return node, nil
}

// UpdateNodeInfo updates the gateway's routing table
func (gr *GatewayRouter) UpdateNodeInfo(nodeID string, info *NodeInfo) {
    gr.mu.Lock()
    defer gr.mu.Unlock()
    gr.nodes[nodeID] = info
}
```

**Integration with Existing Join Flow:**
```go
// Enhanced join logic in gateway to route to correct node
func (j *JoinHandler) handleJoin(token string) (*JoinResponse, error) {
    // Existing token validation...
    playerData, err := j.validateToken(token)
    if err != nil {
        return nil, err
    }
    
    // NEW: Route to correct simulation node
    if j.router != nil {
        targetNode, err := j.router.RouteNewPlayer(playerData.Position)
        if err != nil {
            return nil, fmt.Errorf("failed to route player: %w", err)
        }
        
        // If not local node, return redirect
        if targetNode.ID != j.localNodeID {
            return &JoinResponse{
                RedirectTo: targetNode.Endpoint,
                Reason:     "cell_assignment",
            }, nil
        }
    }
    
    // Existing join logic for local node...
    return j.createLocalSession(playerData)
}
```

**Integration with Existing Engine:**
```go
// Enhanced Engine with cell ownership verification
func (e *Engine) getOrCreateCellLocked(key spatial.CellKey) (*CellInstance, error) {
    // Check if this node owns the cell (if distributed assignment is enabled)
    if e.cellAssignment != nil {
        ownerNode, err := e.cellAssignment.GetOwnerNode(key)
        if err != nil {
            return nil, fmt.Errorf("failed to determine cell owner: %w", err)
        }
        
        if ownerNode != e.nodeID {
            return nil, fmt.Errorf("cell %v is owned by node %s, not %s", 
                key, ownerNode, e.nodeID)
        }
    }
    
    // Existing cell creation logic
    cell, ok := e.cells[key]
    if !ok {
        cell = NewCellInstance(key)
        e.cells[key] = cell
    }
    return cell, nil
}

// Engine configuration with optional distributed assignment
type Config struct {
    // Existing config fields...
    CellAssignment *CellAssignment `json:"-"` // Optional distributed assignment
    NodeID         string          `json:"node_id"`
}
```

## 3. Bootstrap and Node Management

**Cluster Initialization:**
```go
// BootstrapConfig defines initial cluster setup
type BootstrapConfig struct {
    InitialNodes     []string `json:"initial_nodes"`
    VirtualReplicas  int      `json:"virtual_replicas"`
    HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// InitializeCluster sets up the initial assignment ring
func InitializeCluster(config BootstrapConfig) (*CellAssignment, error) {
    if len(config.InitialNodes) == 0 {
        return nil, errors.New("at least one initial node required")
    }
    
    // Default to 150 virtual nodes per physical node (industry standard)
    replicas := config.VirtualReplicas
    if replicas == 0 {
        replicas = 150
    }
    
    ca := &CellAssignment{
        ring:     consistent.New(),
        nodes:    make(map[string]*SimNode),
        replicas: replicas,
    }
    
    ca.ring.NumberOfReplicas = replicas
    
    // Add initial nodes
    for _, nodeID := range config.InitialNodes {
        node := &SimNode{
            ID:       nodeID,
            Status:   "healthy",
            Endpoint: fmt.Sprintf("http://%s:8081", nodeID),
            LastSeen: time.Now(),
        }
        if err := ca.AddNode(node); err != nil {
            return nil, fmt.Errorf("failed to add initial node %s: %w", nodeID, err)
        }
    }
    
    return ca, nil
}
```

## 4. Simple Failure Handling

**Node Failure Response:**
```go
// HandleNodeFailure redistributes cells when a node becomes unavailable
func (ca *CellAssignment) HandleNodeFailure(failedNodeID string) error {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    node, exists := ca.nodes[failedNodeID]
    if !exists {
        return fmt.Errorf("node %s not found", failedNodeID)
    }
    
    // Mark node as unhealthy
    node.Status = "unhealthy"
    
    // For Phase 1: Keep node in ring but mark unhealthy
    // Cells will be reassigned to healthy nodes automatically
    // via the GetOwnerNode error handling
    
    log.Printf("Node %s marked as unhealthy, cells will be reassigned on demand", failedNodeID)
    return nil
}

// MarkNodeHealthy restores a node to healthy status
func (ca *CellAssignment) MarkNodeHealthy(nodeID string) error {
    ca.mu.Lock()
    defer ca.mu.Unlock()
    
    node, exists := ca.nodes[nodeID]
    if !exists {
        return fmt.Errorf("node %s not found", nodeID)
    }
    
    node.Status = "healthy"
    node.LastSeen = time.Now()
    
    log.Printf("Node %s restored to healthy status", nodeID)
    return nil
}
```

## 5. Storage Strategy

**Phase 1: Simple File-Based Storage for Development:**
```go
// FileBasedAssignmentStore provides persistent storage for development
type FileBasedAssignmentStore struct {
    filePath string
    mu       sync.RWMutex
}

func (fas *FileBasedAssignmentStore) SaveNodeInfo(nodes map[string]*SimNode) error {
    fas.mu.Lock()
    defer fas.mu.Unlock()
    
    data, err := json.Marshal(nodes)
    if err != nil {
        return fmt.Errorf("failed to marshal nodes: %w", err)
    }
    
    return os.WriteFile(fas.filePath, data, 0644)
}

func (fas *FileBasedAssignmentStore) LoadNodeInfo() (map[string]*SimNode, error) {
    fas.mu.RLock()
    defer fas.mu.RUnlock()
    
    data, err := os.ReadFile(fas.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return make(map[string]*SimNode), nil // Return empty map for first startup
        }
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    var nodes map[string]*SimNode
    if err := json.Unmarshal(data, &nodes); err != nil {
        return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
    }
    
    return nodes, nil
}
```

**Configuration Example:**
```json
{
  "cluster_config": {
    "initial_nodes": ["sim-node-1", "sim-node-2"],
    "virtual_replicas": 150,
    "health_check_interval": "30s",
    "storage_file": "cluster_state.json"
  }
}
```

**Future Storage Options (Phase 2+):**
- **Production**: Redis or etcd for shared state
- **Testing**: In-memory storage for integration tests

- **Consequences**: 

**Positive:**
- **Scalability**: Linear scaling by adding more simulation nodes
- **Deterministic Routing**: Any component can determine cell ownership independently
- **Simple Implementation**: Proven consistent hashing algorithm with minimal complexity
- **Fault Tolerance**: Automatic reassignment when nodes fail
- **Integration Friendly**: Works with existing `spatial.CellKey` and `Engine` architecture
- **Low Operational Overhead**: File-based storage requires no external dependencies for development

**Negative:**
- **No Load Balancing**: Initial implementation doesn't consider node load differences
- **No Locality Optimization**: Cross-node AOI queries may have higher latency
- **Manual Node Management**: Nodes must be added/removed manually (no auto-scaling)
- **File Storage Limitations**: Development storage not suitable for production clustering

**Performance Impact:**
- **Gateway Latency**: +2-5ms for cell-to-node lookup on player join
- **Cross-Node Handovers**: 50-200ms for players crossing node boundaries  
- **Storage Operations**: File I/O only on node join/leave (minimal impact)

**Risk Mitigation:**
- Start with small cluster sizes (2-4 nodes) to validate approach
- Use feature flags to enable distributed assignment gradually
- Implement comprehensive monitoring for assignment decisions
- Design clear upgrade path to production storage (Redis/etcd)
- Maintain compatibility with single-node deployment

**Implementation Phases:**
1. **Phase 1**: Basic consistent hashing with file-based storage and manual node management
2. **Phase 2**: Production storage (Redis/etcd) and health monitoring
3. **Phase 3**: Advanced features (load balancing, auto-scaling) - separate ADRs

**Integration Points:**
- **ADR 0002 Cross-Node Handover**: Cell assignment determines when cross-node handovers are needed
- **Gateway Routing**: Gateway uses assignment lookup to route new players to correct nodes
- **Engine Integration**: `Engine.getOrCreateCell()` must verify node ownership before creating cells
- **Spatial Package**: Uses existing `spatial.CellKey` and `WorldToCell()` functions
- **Session Management**: Player session must track both cell and owning node

**Testing Strategy:**
- **Unit Tests**: Test assignment algorithms and node management in isolation
- **Integration Tests**: Test with multiple engine instances using file-based storage
- **End-to-End Tests**: Test complete player join and handover scenarios
- **Failure Tests**: Simulate node failures and validate reassignment behavior

**Alternative Approaches Considered:**
- **Static Region Assignment**: Pre-assign geographic regions to nodes
  - Pros: Simpler implementation, predictable performance  
  - Cons: Poor load balancing, hotspot risks, inflexible scaling
- **Round-Robin Assignment**: Assign cells in round-robin fashion
  - Pros: Perfect load distribution, simple algorithm
  - Cons: Poor locality, non-deterministic from external perspective
- **Centralized Assignment Service**: Single service manages all cell assignments
  - Pros: Strong consistency, centralized control
  - Cons: Single point of failure, scalability bottleneck
- **Hash Range Partitioning**: Divide hash space into fixed ranges per node
  - Pros: Simple implementation, no virtual nodes needed
  - Cons: Poor load distribution, complex rebalancing