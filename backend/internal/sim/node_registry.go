package sim

import (
	"sync"

	"prototype-game/backend/internal/spatial"
)

// NodeInfo represents information about a simulation node
type NodeInfo struct {
	ID      string // unique node identifier
	Address string // network address for inter-node communication
	Port    int    // port for cross-node handover API
}

// NodeRegistry manages cell ownership across multiple simulation nodes
type NodeRegistry struct {
	mu          sync.RWMutex
	localNodeID string
	nodes       map[string]*NodeInfo           // nodeID -> NodeInfo
	cellOwners  map[spatial.CellKey]string     // cellKey -> nodeID
}

// NewNodeRegistry creates a new node registry for the given local node
func NewNodeRegistry(localNodeID string) *NodeRegistry {
	return &NodeRegistry{
		localNodeID: localNodeID,
		nodes:       make(map[string]*NodeInfo),
		cellOwners:  make(map[spatial.CellKey]string),
	}
}

// RegisterNode adds a node to the registry
func (nr *NodeRegistry) RegisterNode(nodeInfo *NodeInfo) {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	nr.nodes[nodeInfo.ID] = nodeInfo
}

// UnregisterNode removes a node from the registry
func (nr *NodeRegistry) UnregisterNode(nodeID string) {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	delete(nr.nodes, nodeID)
	
	// Remove cell ownership for this node
	for cell, owner := range nr.cellOwners {
		if owner == nodeID {
			delete(nr.cellOwners, cell)
		}
	}
}

// AssignCell assigns ownership of a cell to a specific node
func (nr *NodeRegistry) AssignCell(cell spatial.CellKey, nodeID string) {
	nr.mu.Lock()
	defer nr.mu.Unlock()
	nr.cellOwners[cell] = nodeID
}

// GetCellOwner returns the node ID that owns the given cell
// Returns local node ID if no specific owner is assigned
func (nr *NodeRegistry) GetCellOwner(cell spatial.CellKey) string {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	
	if owner, exists := nr.cellOwners[cell]; exists {
		return owner
	}
	
	// Default to local node for unassigned cells
	return nr.localNodeID
}

// GetNodeInfo returns the NodeInfo for a given node ID
func (nr *NodeRegistry) GetNodeInfo(nodeID string) (*NodeInfo, bool) {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	
	info, exists := nr.nodes[nodeID]
	return info, exists
}

// IsLocalCell returns true if the cell is owned by the local node
func (nr *NodeRegistry) IsLocalCell(cell spatial.CellKey) bool {
	return nr.GetCellOwner(cell) == nr.localNodeID
}

// GetLocalNodeID returns the local node's ID
func (nr *NodeRegistry) GetLocalNodeID() string {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	return nr.localNodeID
}

// ListNodes returns all registered nodes
func (nr *NodeRegistry) ListNodes() map[string]*NodeInfo {
	nr.mu.RLock()
	defer nr.mu.RUnlock()
	
	result := make(map[string]*NodeInfo)
	for id, info := range nr.nodes {
		result[id] = info
	}
	return result
}