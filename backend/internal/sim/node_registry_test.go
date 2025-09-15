package sim

import (
	"testing"

	"prototype-game/backend/internal/spatial"
)

func TestNodeRegistry_Basic(t *testing.T) {
	registry := NewNodeRegistry("node1")
	
	// Test initial state
	if registry.GetLocalNodeID() != "node1" {
		t.Errorf("Expected local node ID 'node1', got %s", registry.GetLocalNodeID())
	}
	
	// Test default cell ownership
	cell := spatial.CellKey{Cx: 0, Cz: 0}
	if owner := registry.GetCellOwner(cell); owner != "node1" {
		t.Errorf("Expected default cell owner 'node1', got %s", owner)
	}
	
	if !registry.IsLocalCell(cell) {
		t.Errorf("Expected cell to be local")
	}
}

func TestNodeRegistry_RegisterUnregisterNode(t *testing.T) {
	registry := NewNodeRegistry("node1")
	
	// Register a remote node
	node2 := &NodeInfo{
		ID:      "node2",
		Address: "localhost",
		Port:    8082,
	}
	registry.RegisterNode(node2)
	
	// Test node retrieval
	info, exists := registry.GetNodeInfo("node2")
	if !exists {
		t.Errorf("Expected node2 to exist")
	}
	if info.ID != "node2" || info.Address != "localhost" || info.Port != 8082 {
		t.Errorf("Node info mismatch: %+v", info)
	}
	
	// Test node listing
	nodes := registry.ListNodes()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}
	
	// Unregister node
	registry.UnregisterNode("node2")
	_, exists = registry.GetNodeInfo("node2")
	if exists {
		t.Errorf("Expected node2 to be removed")
	}
}

func TestNodeRegistry_CellAssignment(t *testing.T) {
	registry := NewNodeRegistry("node1")
	
	// Register remote node
	node2 := &NodeInfo{ID: "node2", Address: "localhost", Port: 8082}
	registry.RegisterNode(node2)
	
	// Assign cell to remote node
	cell := spatial.CellKey{Cx: 1, Cz: 1}
	registry.AssignCell(cell, "node2")
	
	// Test ownership
	if owner := registry.GetCellOwner(cell); owner != "node2" {
		t.Errorf("Expected cell owner 'node2', got %s", owner)
	}
	
	if registry.IsLocalCell(cell) {
		t.Errorf("Expected cell to be remote")
	}
	
	// Test unassigned cell still defaults to local
	unassignedCell := spatial.CellKey{Cx: 2, Cz: 2}
	if owner := registry.GetCellOwner(unassignedCell); owner != "node1" {
		t.Errorf("Expected unassigned cell owner 'node1', got %s", owner)
	}
}

func TestNodeRegistry_UnregisterNodeClearsAssignments(t *testing.T) {
	registry := NewNodeRegistry("node1")
	
	// Register and assign cells to remote node
	node2 := &NodeInfo{ID: "node2", Address: "localhost", Port: 8082}
	registry.RegisterNode(node2)
	
	cell1 := spatial.CellKey{Cx: 1, Cz: 1}
	cell2 := spatial.CellKey{Cx: 2, Cz: 2}
	registry.AssignCell(cell1, "node2")
	registry.AssignCell(cell2, "node2")
	
	// Verify assignments
	if registry.GetCellOwner(cell1) != "node2" {
		t.Errorf("Expected cell1 owned by node2")
	}
	if registry.GetCellOwner(cell2) != "node2" {
		t.Errorf("Expected cell2 owned by node2")
	}
	
	// Unregister node
	registry.UnregisterNode("node2")
	
	// Verify cell assignments are cleared (revert to local)
	if registry.GetCellOwner(cell1) != "node1" {
		t.Errorf("Expected cell1 to revert to local node")
	}
	if registry.GetCellOwner(cell2) != "node1" {
		t.Errorf("Expected cell2 to revert to local node")
	}
}