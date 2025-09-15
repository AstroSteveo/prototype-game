//go:build ws

package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/join"
	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

func TestCrossNodeHandover(t *testing.T) {
	// Setup simulation engine with cross-node capability
	cfg := sim.Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
		NodeID:              "test-node1",
	}
	engine := sim.NewEngine(cfg)
	engine.Start()
	defer engine.Stop(context.Background())

	// Register a remote node and assign cell (1,0) to it
	remoteNode := &sim.NodeInfo{
		ID:      "test-node2",
		Address: "localhost",
		Port:    8082,
	}
	engine.RegisterNode(remoteNode)
	targetCell := spatial.CellKey{Cx: 1, Cz: 0}
	engine.AssignCellToNode(targetCell, "test-node2")

	// Setup cross-node handover service
	crossNodeSvc := sim.NewHTTPCrossNodeService("test-node1", 8081)
	engine.SetCrossNodeHandoverService(crossNodeSvc)

	// Setup mock gateway
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "valid-token" {
			json.NewEncoder(w).Encode(map[string]any{
				"valid":     true,
				"player_id": "test-player",
				"name":      "TestPlayer",
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer gateway.Close()

	// Setup WebSocket server
	mux := http.NewServeMux()
	auth := join.NewHTTPAuth(gateway.URL)
	store := state.NewMemStore()
	RegisterWithStore(mux, "/ws", auth, engine, store)

	server := httptest.NewServer(mux)
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + server.URL[4:] + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test complete")

	// Send hello with valid token
	hello := map[string]any{
		"token": "valid-token",
	}
	if err := wsjson.Write(ctx, conn, hello); err != nil {
		t.Fatalf("Failed to send hello: %v", err)
	}

	// Read join_ack
	var joinAck map[string]any
	if err := wsjson.Read(ctx, conn, &joinAck); err != nil {
		t.Fatalf("Failed to read join_ack: %v", err)
	}

	if joinAck["type"] != "join_ack" {
		t.Fatalf("Expected join_ack, got %v", joinAck["type"])
	}

	// Get the actual player ID from join_ack
	data, ok := joinAck["data"].(map[string]any)
	if !ok {
		t.Fatal("Invalid join_ack data")
	}
	playerID, ok := data["player_id"].(string)
	if !ok {
		t.Fatal("No player_id in join_ack")
	}

	// Position the WebSocket-created player near the border of cell (1,0)
	startPos := spatial.Vec2{X: 250, Z: 100} // Near cell boundary
	if !engine.DevSetPosition(playerID, startPos) {
		t.Fatal("Failed to set player position")
	}

	// Set velocity to move player into cell (1,0) which is owned by node2
	if !engine.DevSetVelocity(playerID, spatial.Vec2{X: 10, Z: 0}) { // Fast eastward movement
		t.Fatal("Failed to set player velocity")
	}

	// Give a moment for the player to move and handover to be processed
	time.Sleep(200 * time.Millisecond)

	// Wait for cross-node handover to be detected
	handoverDetected := false
	timeout := time.After(2 * time.Second)
	
messageLoop:
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for cross-node handover")
		default:
			// Try to read a message with short timeout
			msgCtx, msgCancel := context.WithTimeout(ctx, 100*time.Millisecond)
			var msg map[string]any
			err := wsjson.Read(msgCtx, conn, &msg)
			msgCancel()
			
			if err != nil {
				// No message available, continue waiting
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// Check if this is a cross-node handover event
			if msg["type"] == "handover_start" {
				data, ok := msg["data"].(map[string]any)
				if !ok {
					t.Fatal("Invalid handover_start data")
				}
				
				if data["reason"] == "cross_node_transfer" {
					// Verify handover details
					targetNode, ok := data["target_node"].(string)
					if !ok || targetNode != "test-node2" {
						t.Errorf("Expected target_node 'test-node2', got %v", data["target_node"])
					}
					
					handoverDetected = true
					break messageLoop
				}
			}
		}
	}

	if !handoverDetected {
		t.Error("Cross-node handover was not detected")
	}

	// Verify player is in cross-node handover state
	players := engine.DevList()
	var testPlayer *sim.Player
	for _, p := range players {
		if p.ID == playerID {
			testPlayer = &p
			break
		}
	}

	if testPlayer == nil {
		t.Fatal("Test player not found")
	}

	if testPlayer.CrossNodeHandover == nil {
		t.Error("Player should be in cross-node handover state")
	} else {
		if testPlayer.CrossNodeHandover.TargetNode != "test-node2" {
			t.Errorf("Expected target node 'test-node2', got %s", testPlayer.CrossNodeHandover.TargetNode)
		}
		if testPlayer.CrossNodeHandover.ToCell != targetCell {
			t.Errorf("Expected target cell %v, got %v", targetCell, testPlayer.CrossNodeHandover.ToCell)
		}
	}
}