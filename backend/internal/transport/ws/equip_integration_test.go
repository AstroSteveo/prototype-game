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

type fakeAuthEquip struct{}

func (fakeAuthEquip) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "TestPlayer", true
	}
	return "", "", false
}

func TestEquipFlowWithSkillGating(t *testing.T) {
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
	Register(mux, "/ws", fakeAuthEquip{}, eng)
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

	// Add a sword to player's inventory
	err = eng.DevAddItemToPlayer(playerID, "sword_iron", 1, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword to inventory: %v", err)
	}

	// Get the item instance ID - we need to wait for the inventory update
	var swordInstanceID string
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	for swordInstanceID == "" {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for inventory update with sword")
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
					items := inventory["items"].([]interface{})
					for _, item := range items {
						itemObj := item.(map[string]interface{})
						instance := itemObj["instance"].(map[string]interface{})
						if instance["template_id"] == "sword_iron" {
							swordInstanceID = instance["instance_id"].(string)
							break
						}
					}
				}
			}
		}
	}

	// Test 1: Try to equip sword without required skill (should fail)
	equipMsg := map[string]interface{}{
		"type":        "equip",
		"seq":         1,
		"instance_id": swordInstanceID,
		"slot":        "main_hand",
	}

	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send equip command: %v", err)
	}

	// Should receive equipment_result with skill_gate error
	var equipResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &equipResponse); err != nil {
		t.Fatalf("Failed to read equip response: %v", err)
	}

	if equipResponse["type"] != "equipment_result" {
		t.Errorf("Expected equipment_result, got %v", equipResponse["type"])
	}

	resultData := equipResponse["data"].(map[string]interface{})
	if resultData["success"].(bool) {
		t.Error("Equip should fail without required skill")
	}
	if resultData["code"].(string) != "skill_gate" {
		t.Errorf("Expected skill_gate error, got %v", resultData["code"])
	}

	// Test 2: Give player required skill and try again
	err = eng.DevGivePlayerSkill(playerID, "melee", 10)
	if err != nil {
		t.Fatalf("Failed to give player skill: %v", err)
	}

	equipMsg["seq"] = 2 // New sequence number
	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send second equip command: %v", err)
	}

	// Should receive successful equipment_result
	if err := wsjson.Read(ctx, c, &equipResponse); err != nil {
		t.Fatalf("Failed to read second equip response: %v", err)
	}

	resultData = equipResponse["data"].(map[string]interface{})
	if !resultData["success"].(bool) {
		t.Errorf("Equip should succeed with required skill, got error: %v", resultData["message"])
	}
	if resultData["code"].(string) != "success" {
		t.Errorf("Expected success code, got %v", resultData["code"])
	}

	// Test 3: Try to unequip immediately (should fail due to cooldown)
	unequipMsg := map[string]interface{}{
		"type": "unequip",
		"seq":  3,
		"slot": "main_hand",
	}

	if err := wsjson.Write(ctx, c, unequipMsg); err != nil {
		t.Fatalf("Failed to send unequip command: %v", err)
	}

	var unequipResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &unequipResponse); err != nil {
		t.Fatalf("Failed to read unequip response: %v", err)
	}

	resultData = unequipResponse["data"].(map[string]interface{})
	if resultData["success"].(bool) {
		t.Error("Unequip should fail due to cooldown")
	}
	if resultData["code"].(string) != "equip_locked" {
		t.Errorf("Expected equip_locked error, got %v", resultData["code"])
	}

	// Verify equipment and inventory state updates were sent
	equipmentFound := false
	inventoryFound := false

	// Read a few more messages to get state updates
	for i := 0; i < 5 && (!equipmentFound || !inventoryFound); i++ {
		var response map[string]interface{}
		readCtx, cancelRead := context.WithTimeout(ctx, 1*time.Second)
		err := wsjson.Read(readCtx, c, &response)
		cancelRead()
		if err != nil {
			continue
		}

		if response["type"] == "state" {
			stateData := response["data"].(map[string]interface{})
			if _, hasEquipment := stateData["equipment"]; hasEquipment {
				equipmentFound = true
			}
			if _, hasInventory := stateData["inventory"]; hasInventory {
				inventoryFound = true
			}
		}
	}

	if !equipmentFound {
		t.Error("Expected equipment delta after successful equip")
	}
	if !inventoryFound {
		t.Error("Expected inventory delta after successful equip")
	}
}

func TestEquipIdempotency(t *testing.T) {
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
	Register(mux, "/ws", fakeAuthEquip{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "test done")

	// Setup: hello and get player
	hello := join.Hello{Token: "tok"}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	var joinResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &joinResponse); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	joinData := joinResponse["data"].(map[string]interface{})
	playerID := joinData["player_id"].(string)

	// Add sword and give skill
	err = eng.DevAddItemToPlayer(playerID, "sword_iron", 1, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword: %v", err)
	}
	err = eng.DevGivePlayerSkill(playerID, "melee", 10)
	if err != nil {
		t.Fatalf("Failed to give skill: %v", err)
	}

	// Wait for inventory update to get instance ID
	var swordInstanceID string
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	for swordInstanceID == "" {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for sword in inventory")
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
					items := inventory["items"].([]interface{})
					for _, item := range items {
						itemObj := item.(map[string]interface{})
						instance := itemObj["instance"].(map[string]interface{})
						if instance["template_id"] == "sword_iron" {
							swordInstanceID = instance["instance_id"].(string)
							break
						}
					}
				}
			}
		}
	}

	// Send the same equip command twice with the same sequence number
	equipMsg := map[string]interface{}{
		"type":        "equip",
		"seq":         1,
		"instance_id": swordInstanceID,
		"slot":        "main_hand",
	}

	// First attempt
	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send first equip command: %v", err)
	}

	var firstResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &firstResponse); err != nil {
		t.Fatalf("Failed to read first response: %v", err)
	}

	// Second attempt with same sequence number (should be ignored)
	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send duplicate equip command: %v", err)
	}

	// Should not receive a second response (ignored due to idempotency)
	// We'll wait a short time and expect timeout
	readCtx, cancelRead := context.WithTimeout(ctx, 1*time.Second)
	var duplicateResponse map[string]interface{}
	err = wsjson.Read(readCtx, c, &duplicateResponse)
	cancelRead()

	// Should either timeout or receive a different message type (like state)
	if err == nil && duplicateResponse["type"] == "equipment_result" {
		// If we got an equipment_result, check if it's genuinely duplicate
		if duplicateResponse["data"].(map[string]interface{})["operation"] == "equip" {
			t.Error("Duplicate equip command should be ignored, but got response")
		}
	}

	// Verify the first command succeeded
	firstData := firstResponse["data"].(map[string]interface{})
	if !firstData["success"].(bool) {
		t.Errorf("First equip should succeed, got: %v", firstData["message"])
	}
}

func TestEquipSlotValidation(t *testing.T) {
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
	Register(mux, "/ws", fakeAuthEquip{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "test done")

	// Setup
	hello := join.Hello{Token: "tok"}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	var joinResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &joinResponse); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	joinData := joinResponse["data"].(map[string]interface{})
	playerID := joinData["player_id"].(string)

	// Add sword (main hand only) and give skill
	err = eng.DevAddItemToPlayer(playerID, "sword_iron", 1, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword: %v", err)
	}
	err = eng.DevGivePlayerSkill(playerID, "melee", 10)
	if err != nil {
		t.Fatalf("Failed to give skill: %v", err)
	}

	// Wait for inventory update to get instance ID
	var swordInstanceID string
	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	for swordInstanceID == "" {
		select {
		case <-timeout.C:
			t.Fatal("Timeout waiting for sword in inventory")
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
					items := inventory["items"].([]interface{})
					for _, item := range items {
						itemObj := item.(map[string]interface{})
						instance := itemObj["instance"].(map[string]interface{})
						if instance["template_id"] == "sword_iron" {
							swordInstanceID = instance["instance_id"].(string)
							break
						}
					}
				}
			}
		}
	}

	// Try to equip sword to chest slot (should fail)
	equipMsg := map[string]interface{}{
		"type":        "equip",
		"seq":         1,
		"instance_id": swordInstanceID,
		"slot":        "chest", // Invalid slot for sword
	}

	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send equip command: %v", err)
	}

	var equipResponse map[string]interface{}
	if err := wsjson.Read(ctx, c, &equipResponse); err != nil {
		t.Fatalf("Failed to read equip response: %v", err)
	}

	resultData := equipResponse["data"].(map[string]interface{})
	if resultData["success"].(bool) {
		t.Error("Equip should fail for illegal slot")
	}
	if resultData["code"].(string) != "illegal_slot" {
		t.Errorf("Expected illegal_slot error, got %v", resultData["code"])
	}

	// Verify sword is still in inventory by trying to equip to valid slot
	equipMsg["slot"] = "main_hand"
	equipMsg["seq"] = 2
	if err := wsjson.Write(ctx, c, equipMsg); err != nil {
		t.Fatalf("Failed to send valid equip command: %v", err)
	}

	if err := wsjson.Read(ctx, c, &equipResponse); err != nil {
		t.Fatalf("Failed to read valid equip response: %v", err)
	}

	resultData = equipResponse["data"].(map[string]interface{})
	if !resultData["success"].(bool) {
		t.Errorf("Valid equip should succeed, got: %v", resultData["message"])
	}
}
