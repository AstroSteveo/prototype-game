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

type fakeAuthT031 struct{}

func (fakeAuthT031) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "TestPlayer", true
	}
	return "", "", false
}

// TestT031_InventoryEquipmentDeltas validates that the T-031 acceptance criteria are met:
// "State contains deltas when versions change"
func TestT031_InventoryEquipmentDeltas(t *testing.T) {
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
	Register(mux, "/ws", fakeAuthT031{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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

	// Helper function to wait for specific state updates
	waitForStateUpdate := func(testName string, checkFunc func(map[string]interface{}) bool) bool {
		timeout := time.NewTimer(3 * time.Second)
		defer timeout.Stop()

		for {
			select {
			case <-timeout.C:
				t.Errorf("T-031 FAILED: Timeout waiting for %s", testName)
				return false
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
					if checkFunc(stateData) {
						return true
					}
				}
			}
		}
	}

	// Test 1: Inventory version delta
	t.Log("T-031 Test 1: Adding item should trigger inventory delta in state message")

	err = eng.DevAddItemToPlayer(playerID, "sword_iron", 1, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}

	var swordInstanceID string
	found := waitForStateUpdate("inventory delta", func(stateData map[string]interface{}) bool {
		if inventory, hasInventory := stateData["inventory"]; hasInventory {
			t.Log("T-031 SUCCESS: State message contains inventory delta when version changes")

			invMap := inventory.(map[string]interface{})
			if items, hasItems := invMap["items"]; hasItems {
				itemsArray := items.([]interface{})
				for _, item := range itemsArray {
					itemObj := item.(map[string]interface{})
					instance := itemObj["instance"].(map[string]interface{})
					if instance["template_id"] == "sword_iron" {
						swordInstanceID = instance["instance_id"].(string)
						return true
					}
				}
			}
		}
		return false
	})

	if !found {
		t.Fatal("Failed to receive inventory delta")
	}

	// Test 2: Equipment version delta
	t.Log("T-031 Test 2: Equipment action should trigger equipment delta in state message")

	// Give player required skill for equipment
	err = eng.DevGivePlayerSkill(playerID, "melee", 10)
	if err != nil {
		t.Fatalf("Failed to give skill: %v", err)
	}

	// Equip the sword
	equipMsg := map[string]interface{}{
		"type":        "equip",
		"seq":         1,
		"instance_id": swordInstanceID,
		"slot":        "main_hand",
	}

	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send equip command: %v", err)
	}

	// Read equipment result
	var equipResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &equipResponse); err != nil {
		t.Fatalf("Failed to read equip response: %v", err)
	}

	var foundEquipmentDelta bool
	var foundInventoryDeltaAfterEquip bool

	// Wait for both equipment and inventory deltas
	for i := 0; i < 10 && (!foundEquipmentDelta || !foundInventoryDeltaAfterEquip); i++ {
		waitForStateUpdate("equipment/inventory deltas", func(stateData map[string]interface{}) bool {
			if _, hasEquipment := stateData["equipment"]; hasEquipment && !foundEquipmentDelta {
				t.Log("T-031 SUCCESS: State message contains equipment delta when version changes")
				foundEquipmentDelta = true
			}

			if _, hasInventory := stateData["inventory"]; hasInventory && !foundInventoryDeltaAfterEquip {
				foundInventoryDeltaAfterEquip = true
				t.Log("T-031 SUCCESS: Inventory delta also sent when equipment changes")
			}

			return foundEquipmentDelta && foundInventoryDeltaAfterEquip
		})
	}

	// Test 3: Skills version delta
	t.Log("T-031 Test 3: Skill change should trigger skills delta in state message")

	err = eng.DevGivePlayerSkill(playerID, "archery", 5)
	if err != nil {
		t.Fatalf("Failed to give additional skill: %v", err)
	}

	found = waitForStateUpdate("skills delta", func(stateData map[string]interface{}) bool {
		if _, hasSkills := stateData["skills"]; hasSkills {
			t.Log("T-031 SUCCESS: State message contains skills delta when version changes")
			return true
		}
		return false
	})

	if !found {
		t.Error("Failed to receive skills delta")
	}

	t.Log("T-031 ACCEPTANCE CRITERIA VALIDATED: State contains deltas when versions change")
	t.Log("✓ Inventory deltas included when inventory version changes")
	t.Log("✓ Equipment deltas included when equipment version changes")
	t.Log("✓ Skills deltas included when skills version changes")
}
