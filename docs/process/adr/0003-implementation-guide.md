# ADR 0003: Simplified Implementation Guide

This guide provides practical implementation steps for the approved ADR 0003: Distributed Cell Assignment Strategy.

## Overview

The approved ADR implements basic consistent hashing for Phase B cross-node sharding. This is a simplified, production-ready approach that avoids the complexity of dynamic rebalancing and locality optimization.

## Implementation Steps

### Phase 1: Core Consistent Hashing (Week 1-2)

1. **Add Consistent Hashing Library**
```bash
cd backend && go get github.com/stathat/consistent
```

2. **Implement CellAssignment**
Create `backend/internal/assignment/assignment.go`:

```go
package assignment

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "sync"
    "time"
    
    "github.com/stathat/consistent"
    "prototype-game/backend/internal/spatial"
)

type CellAssignment struct {
    ring     *consistent.Consistent
    nodes    map[string]*SimNode
    replicas int
    mu       sync.RWMutex
}

type SimNode struct {
    ID       string    `json:"id"`
    Status   string    `json:"status"`
    Endpoint string    `json:"endpoint"`
    LastSeen time.Time `json:"last_seen"`
}

func NewCellAssignment(replicas int) *CellAssignment {
    if replicas == 0 {
        replicas = 150 // Industry standard
    }
    
    c := consistent.New()
    c.NumberOfReplicas = replicas
    
    return &CellAssignment{
        ring:     c,
        nodes:    make(map[string]*SimNode),
        replicas: replicas,
    }
}

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
    
    node, exists := ca.nodes[nodeID]
    if !exists {
        return "", fmt.Errorf("assigned node %s not found", nodeID)
    }
    
    if node.Status != "healthy" {
        return "", fmt.Errorf("assigned node %s is not healthy (status: %s)", nodeID, node.Status)
    }
    
    return nodeID, nil
}

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

3. **Update Engine Configuration**
Modify `backend/internal/sim/types.go`:

```go
type Config struct {
    // Existing fields...
    CellAssignment *assignment.CellAssignment `json:"-"`
    NodeID         string                     `json:"node_id"`
    EnableDistributed bool                   `json:"enable_distributed"`
}
```

4. **Update Engine Cell Creation**
Modify `engine.go`:

```go
func (e *Engine) getOrCreateCellLocked(key spatial.CellKey) (*CellInstance, error) {
    // Check if this node owns the cell (if distributed assignment is enabled)
    if e.cfg.EnableDistributed && e.cfg.CellAssignment != nil {
        ownerNode, err := e.cfg.CellAssignment.GetOwnerNode(key)
        if err != nil {
            return nil, fmt.Errorf("failed to determine cell owner: %w", err)
        }
        
        if ownerNode != e.cfg.NodeID {
            return nil, fmt.Errorf("cell %v is owned by node %s, not %s", 
                key, ownerNode, e.cfg.NodeID)
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
```

### Phase 1b: Testing Infrastructure (Week 2)

1. **Create Assignment Tests**
Create `backend/internal/assignment/assignment_test.go`:

```go
package assignment

import (
    "testing"
    "time"
    
    "prototype-game/backend/internal/spatial"
)

func TestBasicAssignment(t *testing.T) {
    ca := NewCellAssignment(10)
    
    // Add test nodes
    node1 := &SimNode{ID: "node1", Status: "healthy", Endpoint: "http://node1:8081", LastSeen: time.Now()}
    node2 := &SimNode{ID: "node2", Status: "healthy", Endpoint: "http://node2:8081", LastSeen: time.Now()}
    
    if err := ca.AddNode(node1); err != nil {
        t.Fatalf("Failed to add node1: %v", err)
    }
    if err := ca.AddNode(node2); err != nil {
        t.Fatalf("Failed to add node2: %v", err)
    }
    
    // Test cell assignment
    cellKey := spatial.CellKey{Cx: 0, Cz: 0}
    owner, err := ca.GetOwnerNode(cellKey)
    if err != nil {
        t.Fatalf("Failed to get owner: %v", err)
    }
    
    if owner != "node1" && owner != "node2" {
        t.Fatalf("Invalid owner: %s", owner)
    }
    
    // Test deterministic assignment
    owner2, err := ca.GetOwnerNode(cellKey)
    if err != nil {
        t.Fatalf("Failed to get owner second time: %v", err)
    }
    
    if owner != owner2 {
        t.Fatalf("Assignment not deterministic: %s != %s", owner, owner2)
    }
}

func TestNodeFailure(t *testing.T) {
    ca := NewCellAssignment(10)
    
    node1 := &SimNode{ID: "node1", Status: "healthy", Endpoint: "http://node1:8081", LastSeen: time.Now()}
    node2 := &SimNode{ID: "node2", Status: "unhealthy", Endpoint: "http://node2:8081", LastSeen: time.Now()}
    
    ca.AddNode(node1)
    ca.AddNode(node2)
    
    cellKey := spatial.CellKey{Cx: 0, Cz: 0}
    
    // Keep trying until we get a cell assigned to the unhealthy node
    for i := 0; i < 100; i++ {
        testKey := spatial.CellKey{Cx: i, Cz: 0}
        owner, err := ca.GetOwnerNode(testKey)
        
        if err != nil && owner == "node2" {
            // Expected: should fail for unhealthy node
            return
        }
    }
    
    t.Fatalf("Expected to find a cell assigned to unhealthy node that fails")
}
```

### Phase 2: Gateway Integration (Week 3)

1. **Add Gateway Router**
Create `backend/internal/gateway/router.go`:

```go
package gateway

import (
    "fmt"
    "sync"
    
    "prototype-game/backend/internal/assignment"
    "prototype-game/backend/internal/spatial"
)

type GatewayRouter struct {
    assignment *assignment.CellAssignment
    nodes      map[string]*NodeInfo
    mu         sync.RWMutex
}

type NodeInfo struct {
    ID       string `json:"id"`
    Endpoint string `json:"endpoint"`
    Status   string `json:"status"`
}

func NewGatewayRouter(assignment *assignment.CellAssignment) *GatewayRouter {
    return &GatewayRouter{
        assignment: assignment,
        nodes:      make(map[string]*NodeInfo),
    }
}

func (gr *GatewayRouter) RouteNewPlayer(pos spatial.Vec2) (*NodeInfo, error) {
    cx, cz := spatial.WorldToCell(pos.X, pos.Z, 100.0) // CELL_SIZE = 100.0
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
```

### Phase 3: File Storage (Week 4)

1. **Implement File Storage**
Create `backend/internal/assignment/storage.go`:

```go
package assignment

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
)

type FileStorage struct {
    filePath string
    mu       sync.RWMutex
}

type StoredState struct {
    Nodes map[string]*SimNode `json:"nodes"`
    Config struct {
        Replicas int `json:"replicas"`
    } `json:"config"`
}

func NewFileStorage(filePath string) *FileStorage {
    return &FileStorage{filePath: filePath}
}

func (fs *FileStorage) Save(ca *CellAssignment) error {
    fs.mu.Lock()
    defer fs.mu.Unlock()
    
    ca.mu.RLock()
    state := StoredState{
        Nodes: make(map[string]*SimNode),
    }
    for id, node := range ca.nodes {
        state.Nodes[id] = node
    }
    state.Config.Replicas = ca.replicas
    ca.mu.RUnlock()
    
    data, err := json.MarshalIndent(state, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal state: %w", err)
    }
    
    return os.WriteFile(fs.filePath, data, 0644)
}

func (fs *FileStorage) Load() (*CellAssignment, error) {
    fs.mu.RLock()
    defer fs.mu.RUnlock()
    
    data, err := os.ReadFile(fs.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            // Return empty assignment for first startup
            return NewCellAssignment(150), nil
        }
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    var state StoredState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("failed to unmarshal state: %w", err)
    }
    
    ca := NewCellAssignment(state.Config.Replicas)
    for _, node := range state.Nodes {
        if err := ca.AddNode(node); err != nil {
            return nil, fmt.Errorf("failed to add node %s: %w", node.ID, err)
        }
    }
    
    return ca, nil
}
```

## Testing Strategy

1. **Unit Tests**: Test assignment algorithms in isolation
2. **Integration Tests**: Test with multiple engine instances
3. **End-to-End Tests**: Test complete player routing scenarios

## Configuration Example

```json
{
  "distributed": {
    "enable": true,
    "node_id": "sim-node-1",
    "storage_file": "cluster_state.json",
    "initial_nodes": ["sim-node-1", "sim-node-2"],
    "virtual_replicas": 150
  }
}
```

## Deployment Steps

1. **Development**: Single node with distributed assignment disabled
2. **Testing**: Multi-node with file storage
3. **Production**: Multi-node with Redis/etcd storage (Phase 2)

This implementation provides a solid foundation for distributed cell assignment while maintaining simplicity and avoiding premature optimization.