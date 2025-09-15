package sim

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

func TestEngineWithCrossNodeHandover(t *testing.T) {
	t.Run("ReconnectMode", func(t *testing.T) {
		testCrossNodeHandover(t, "reconnect")
	})
	
	t.Run("TunnelMode", func(t *testing.T) {
		testCrossNodeHandover(t, "tunnel")
	})
}

func testCrossNodeHandover(t *testing.T, handoverMode string) {
	// Setup simulation engine with cross-node capability
	cfg := Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
		NodeID:              "test-node1",
		HandoverMode:        handoverMode,
	}
	engine := NewEngine(cfg)
	engine.Start()
	defer engine.Stop(context.Background())

	// Register a remote node and assign cell (1,0) to it
	remoteNode := &NodeInfo{
		ID:      "test-node2",
		Address: "localhost", 
		Port:    8082,
	}
	engine.RegisterNode(remoteNode)
	targetCell := spatial.CellKey{Cx: 1, Cz: 0}
	engine.AssignCellToNode(targetCell, "test-node2")

	// Setup cross-node handover service
	crossNodeSvc := NewHTTPCrossNodeService("test-node1", 8081)
	engine.SetCrossNodeHandoverService(crossNodeSvc)

	// Spawn player near the border of cell (1,0)
	playerID := "test-player"
	startPos := spatial.Vec2{X: 250, Z: 100}
	player := engine.DevSpawn(playerID, "TestPlayer", startPos)
	if player == nil {
		t.Fatal("Failed to spawn player")
	}

	// Verify initial state
	if player.OwnedCell != (spatial.CellKey{Cx: 0, Cz: 0}) {
		t.Errorf("Expected initial cell (0,0), got %v", player.OwnedCell)
	}

	if player.CrossNodeHandover != nil {
		t.Error("Player should not be in cross-node handover state initially")
	}

	// Set velocity to move player into cell (1,0) which is owned by node2
	if !engine.DevSetVelocity(playerID, spatial.Vec2{X: 10, Z: 0}) {
		t.Fatal("Failed to set player velocity")
	}

	// Wait for simulation to process movement and handover
	// At 10 m/s, player needs to travel 6+ meters to cross boundary with 2m hysteresis
	// This should take about 800ms at 20Hz tick rate
	time.Sleep(1 * time.Second)

	// Check updated player state
	players := engine.DevList()
	var updatedPlayer *Player
	for _, p := range players {
		if p.ID == playerID {
			updatedPlayer = &p
			break
		}
	}

	if updatedPlayer == nil {
		t.Fatal("Player not found after movement")
	}

	// Verify cross-node handover was triggered
	if updatedPlayer.CrossNodeHandover == nil {
		t.Error("Expected player to be in cross-node handover state")
	} else {
		if updatedPlayer.CrossNodeHandover.TargetNode != "test-node2" {
			t.Errorf("Expected target node 'test-node2', got %s", updatedPlayer.CrossNodeHandover.TargetNode)
		}
		if updatedPlayer.CrossNodeHandover.ToCell != targetCell {
			t.Errorf("Expected target cell %v, got %v", targetCell, updatedPlayer.CrossNodeHandover.ToCell)
		}
		if updatedPlayer.CrossNodeHandover.Status != HandoverInProgress {
			t.Errorf("Expected handover status InProgress, got %v", updatedPlayer.CrossNodeHandover.Status)
		}
	}

	// Verify player moved to the target cell
	if updatedPlayer.OwnedCell != targetCell {
		t.Errorf("Expected player in cell %v, got %v", targetCell, updatedPlayer.OwnedCell)
	}

	// Verify position moved beyond the cell boundary
	if updatedPlayer.Pos.X <= 256 { // Should be past the cell boundary at x=256
		t.Errorf("Expected player position X > 256, got %f", updatedPlayer.Pos.X)
	}
}