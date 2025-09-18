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
	"prototype-game/backend/internal/spatial"
)

type fakeAuthInventory struct{}

func (fakeAuthInventory) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "TestPlayer", true
	}
	return "", "", false
}

func TestInventoryInJoinAck(t *testing.T) {
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
	Register(mux, "/ws", fakeAuthInventory{}, eng)
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

	// Send hello message
	hello := join.Hello{Token: "tok"}
	if err := wsjson.Write(ctx, c, hello); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	// Read join_ack
	var response map[string]interface{}
	if err := wsjson.Read(ctx, c, &response); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	// Verify response structure
	if response["type"] != "join_ack" {
		t.Errorf("Expected join_ack, got %v", response["type"])
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("join_ack data should be an object")
	}

	// Verify inventory is present and structured correctly
	inventory, ok := data["inventory"].(map[string]interface{})
	if !ok {
		t.Fatalf("inventory should be present in join_ack")
	}

	// Check inventory structure
	items, ok := inventory["items"].([]interface{})
	if !ok {
		t.Fatalf("inventory should have items array")
	}
	if len(items) != 0 {
		t.Errorf("new player should start with empty inventory, got %d items", len(items))
	}

	compartmentCaps, ok := inventory["compartment_caps"].(map[string]interface{})
	if !ok {
		t.Fatalf("inventory should have compartment_caps")
	}

	expectedCompartments := []string{"backpack", "belt", "craft_bag"}
	for _, comp := range expectedCompartments {
		if _, exists := compartmentCaps[comp]; !exists {
			t.Errorf("compartment_caps should include %s", comp)
		}
	}

	weightLimit, ok := inventory["weight_limit"].(float64)
	if !ok || weightLimit <= 0 {
		t.Errorf("inventory should have positive weight_limit, got %v", weightLimit)
	}

	// Verify equipment is present
	equipment, ok := data["equipment"].(map[string]interface{})
	if !ok {
		t.Fatalf("equipment should be present in join_ack")
	}

	slots, ok := equipment["slots"].(map[string]interface{})
	if !ok {
		t.Fatalf("equipment should have slots object")
	}
	if len(slots) != 0 {
		t.Errorf("new player should start with empty equipment slots, got %d slots", len(slots))
	}

	// Verify skills is present
	skills, ok := data["skills"].(map[string]interface{})
	if !ok {
		t.Fatalf("skills should be present in join_ack")
	}
	if len(skills) != 0 {
		t.Errorf("new player should start with empty skills, got %d skills", len(skills))
	}

	// Verify encumbrance is present and structured correctly
	encumbrance, ok := data["encumbrance"].(map[string]interface{})
	if !ok {
		t.Fatalf("encumbrance should be present in join_ack")
	}

	expectedEncumbranceFields := []string{
		"current_weight", "max_weight", "current_bulk", "max_bulk",
		"weight_pct", "bulk_pct", "movement_penalty",
	}
	for _, field := range expectedEncumbranceFields {
		if _, exists := encumbrance[field]; !exists {
			t.Errorf("encumbrance should include %s", field)
		}
	}

	// Verify movement penalty starts at 1.0 (no penalty)
	movementPenalty, ok := encumbrance["movement_penalty"].(float64)
	if !ok || movementPenalty != 1.0 {
		t.Errorf("new player should have no movement penalty (1.0), got %v", movementPenalty)
	}

	// Verify weight percentage starts at 0
	weightPct, ok := encumbrance["weight_pct"].(float64)
	if !ok || weightPct != 0.0 {
		t.Errorf("new player should have 0%% weight, got %v", weightPct)
	}
}

func TestInventorySystemIntegration(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	// Get player manager to test item operations
	playerMgr := eng.GetPlayerManager()

	// Create a test player directly in the engine
	playerID := "test_player"
	playerName := "TestPlayer"
	player := eng.DevSpawn(playerID, playerName, spatial.Vec2{X: 0, Z: 0})

	// Verify player has initialized inventory and equipment
	if player.Inventory == nil {
		t.Fatal("Player should have initialized inventory")
	}
	if player.Equipment == nil {
		t.Fatal("Player should have initialized equipment")
	}
	if player.Skills == nil {
		t.Fatal("Player should have initialized skills")
	}

	// Test adding an item to inventory
	swordTemplate, exists := playerMgr.GetItemTemplate("sword_iron")
	if !exists {
		t.Fatal("Test item template sword_iron should exist")
	}

	swordInstance := sim.ItemInstance{
		InstanceID: "sword_001",
		TemplateID: swordTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err := playerMgr.AddItemToInventory(player, swordInstance, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword to inventory: %v", err)
	}

	// Verify item is in inventory
	if !player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Sword should be in player inventory")
	}

	// Test skill requirement for equipment
	err = playerMgr.EquipItem(player, swordInstance.InstanceID, sim.SlotMainHand, time.Now())
	if err != sim.ErrSkillGate {
		t.Errorf("Should fail to equip sword without skill, got: %v", err)
	}

	// Give player the required skill
	player.Skills["melee"] = 10

	// Test successful equipment
	now := time.Now()
	err = playerMgr.EquipItem(player, swordInstance.InstanceID, sim.SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword with skill: %v", err)
	}

	// Verify sword is equipped and not in inventory
	equippedItem := player.Equipment.GetSlot(sim.SlotMainHand)
	if equippedItem == nil {
		t.Fatal("Sword should be equipped in main hand")
	}
	if equippedItem.Instance.InstanceID != swordInstance.InstanceID {
		t.Error("Wrong item equipped")
	}
	if player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Sword should not be in inventory after equipping")
	}

	// Test encumbrance calculation
	encumbrance := playerMgr.GetPlayerEncumbrance(player)
	expectedWeight := swordTemplate.Weight
	if encumbrance.CurrentWeight != expectedWeight {
		t.Errorf("Expected weight %f, got %f", expectedWeight, encumbrance.CurrentWeight)
	}

	// Test cooldown protection
	err = playerMgr.UnequipItem(player, sim.SlotMainHand, sim.CompartmentBackpack, now)
	if err != sim.ErrEquipLocked {
		t.Errorf("Should fail to unequip during cooldown, got: %v", err)
	}

	// Test unequip after cooldown
	futureTime := now.Add(sim.EquipCooldown + time.Second)
	err = playerMgr.UnequipItem(player, sim.SlotMainHand, sim.CompartmentBackpack, futureTime)
	if err != nil {
		t.Fatalf("Failed to unequip after cooldown: %v", err)
	}

	// Verify sword is back in inventory and slot is empty
	if !player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Sword should be back in inventory after unequipping")
	}
	if !player.Equipment.IsSlotEmpty(sim.SlotMainHand) {
		t.Error("Main hand slot should be empty after unequipping")
	}
}

func TestEncumbranceMovementPenalty(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	playerMgr := eng.GetPlayerManager()
	player := eng.DevSpawn("heavy_test", "HeavyTester", spatial.Vec2{X: 0, Z: 0})

	// Create a very heavy item template for testing
	heavyTemplate := &sim.ItemTemplate{
		ID:          "test_heavy_armor",
		DisplayName: "Heavy Test Armor",
		SlotMask:    sim.SlotMaskChest,
		Weight:      85.0, // 85% of default weight limit
		Bulk:        5,
		DamageType:  "",
		SkillReq:    map[string]int{},
	}
	playerMgr.RegisterItemTemplate(heavyTemplate)

	// Test normal encumbrance (no penalty)
	encumbrance := playerMgr.GetPlayerEncumbrance(player)
	if encumbrance.MovementPenalty != 1.0 {
		t.Errorf("Empty inventory should have no movement penalty, got %f", encumbrance.MovementPenalty)
	}

	// Add heavy armor to inventory
	heavyInstance := sim.ItemInstance{
		InstanceID: "heavy_001",
		TemplateID: heavyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err := playerMgr.AddItemToInventory(player, heavyInstance, sim.CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add heavy armor: %v", err)
	}

	// Test high encumbrance (should cause movement penalty)
	encumbrance = playerMgr.GetPlayerEncumbrance(player)
	if encumbrance.WeightPct != 0.85 {
		t.Errorf("Expected 85%% weight, got %f", encumbrance.WeightPct)
	}
	if encumbrance.MovementPenalty >= 1.0 {
		t.Errorf("High weight should cause movement penalty < 1.0, got %f", encumbrance.MovementPenalty)
	}
}

func TestItemTemplateValidation(t *testing.T) {
	eng := sim.NewEngine(sim.Config{
		CellSize:            256,
		AOIRadius:           128,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
	eng.Start()
	defer eng.Stop(context.Background())

	playerMgr := eng.GetPlayerManager()

	// Verify test templates exist and are valid
	expectedTemplates := []sim.ItemTemplateID{
		"sword_iron",
		"shield_wood",
		"armor_leather",
		"potion_health",
	}

	for _, templateID := range expectedTemplates {
		template, exists := playerMgr.GetItemTemplate(templateID)
		if !exists {
			t.Errorf("Template %s should exist", templateID)
			continue
		}

		// Verify basic template properties
		if template.DisplayName == "" {
			t.Errorf("Template %s should have display name", templateID)
		}
		if template.Weight < 0 {
			t.Errorf("Template %s should have non-negative weight", templateID)
		}
		if template.Bulk < 0 {
			t.Errorf("Template %s should have non-negative bulk", templateID)
		}

		// Test slot mask functionality
		if templateID == "sword_iron" {
			if !template.Allows(sim.SlotMainHand) {
				t.Errorf("Sword should be equippable to main hand")
			}
			if template.Allows(sim.SlotChest) {
				t.Errorf("Sword should not be equippable to chest")
			}
		}
	}
}