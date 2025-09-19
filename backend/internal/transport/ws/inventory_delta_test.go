//go:build ws

package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/join"
	"prototype-game/backend/internal/sim"
)

type fakeAuthDelta struct{}

func (fakeAuthDelta) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "TestPlayer", true
	}
	return "", "", false
}

func TestInventoryDeltaBroadcast(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthDelta{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "test done")

	// Send hello message
	hello := join.Hello{Token: "tok"}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	// Read join_ack
	var joinResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &joinResponse); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	if joinResponse["type"] != "join_ack" {
		t.Errorf("Expected join_ack, got %v", joinResponse["type"])
	}

	joinData := joinResponse["data"].(map[string]interface{})
	playerID := joinData["player_id"].(string)

	// Add an item to player's inventory using the dev API
	err = eng.DevAddItemToPlayer(playerID, "potion_health", 3, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add item via dev API: %v", err)
	}

	// Wait for the next state update (should happen within ~100ms at 10Hz)
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	var foundInventoryUpdate bool
	for !foundInventoryUpdate {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for inventory delta in state update")
		default:
			var response map[string]interface{}
			readCtx, cancelRead := context.WithTimeout(ctx, 500*time.Millisecond)
			err := wsjson.Read(readCtx, c, &response)
			cancelRead()
			if err != nil {
				continue // Keep trying
			}

			// Check if this is a state message with inventory delta
			if response["type"] == "state" {
				stateData := response["data"].(map[string]interface{})
				if inventory, hasInventory := stateData["inventory"].(map[string]interface{}); hasInventory {
					// Verify inventory delta structure
					items, hasItems := inventory["items"].([]interface{})
					if !hasItems {
						t.Error("Inventory delta should include items array")
						continue
					}

					// Check that we have the potion we added
					foundPotion := false
					for _, item := range items {
						itemObj := item.(map[string]interface{})
						instance := itemObj["instance"].(map[string]interface{})
						if instance["template_id"] == "potion_health" {
							var qtyInt int
							switch qty := instance["quantity"].(type) {
							case float64:
								qtyInt = int(qty)
							case int:
								qtyInt = qty
							case int32:
								qtyInt = int(qty)
							case int64:
								qtyInt = int(qty)
							default:
								continue // skip if type is unexpected
							}
							if qtyInt == 3 {
								foundPotion = true
								break
							}
						}
					}

					if !foundPotion {
						t.Error("Expected to find health potion in inventory delta")
						continue
					}

					// Verify encumbrance is included
					encumbrance, hasEncumbrance := inventory["encumbrance"].(map[string]interface{})
					if !hasEncumbrance {
						t.Error("Inventory delta should include encumbrance")
						continue
					}

					// Verify encumbrance fields
					expectedFields := []string{"current_weight", "max_weight", "current_bulk", "max_bulk", "weight_pct", "bulk_pct", "movement_penalty"}
					for _, field := range expectedFields {
						if _, exists := encumbrance[field]; !exists {
							t.Errorf("Encumbrance should include %s", field)
						}
					}

					foundInventoryUpdate = true
					t.Logf("Successfully received inventory delta with %d items", len(items))
				}
			}
		}
	}
}

func TestEncumbranceWarningBroadcast(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthDelta{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "test done")

	// Send hello message
	hello := join.Hello{Token: "tok"}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	// Read join_ack
	var joinResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &joinResponse); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	joinData := joinResponse["data"].(map[string]interface{})
	playerID := joinData["player_id"].(string)

	// Add a heavy item to cause encumbrance warning
	err = eng.DevAddItemToPlayer(playerID, "anvil_iron", 1, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add heavy item: %v", err)
	}

	// Wait for state update with encumbrance warning
	timeout := time.NewTimer(3 * time.Second)
	defer timeout.Stop()

	var foundEncumbranceWarning bool
	for !foundEncumbranceWarning {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for encumbrance warning in state update")
		default:
			var response map[string]interface{}
			readCtx, cancelRead := context.WithTimeout(ctx, 500*time.Millisecond)
			err := wsjson.Read(readCtx, c, &response)
			cancelRead()
			if err != nil {
				continue
			}

			if response["type"] == "state" {
				stateData := response["data"].(map[string]interface{})
				if inventory, hasInventory := stateData["inventory"].(map[string]interface{}); hasInventory {
					encumbrance := inventory["encumbrance"].(map[string]interface{})

					// Check for movement penalty (should be < 1.0 due to heavy anvil)
					movementPenalty := encumbrance["movement_penalty"].(float64)
					weightPct := encumbrance["weight_pct"].(float64)

					if weightPct > 0.8 && movementPenalty < 1.0 {
						foundEncumbranceWarning = true
						t.Logf("Encumbrance warning detected: weight_pct=%f, movement_penalty=%f", weightPct, movementPenalty)
					}
				}
			}
		}
	}
}
