package sim

import (
	"testing"
	"time"
)

func createTestPlayer() *Player {
	return &Player{
		Entity: Entity{
			ID:   "test_player",
			Kind: KindPlayer,
			Name: "TestPlayer",
		},
		Skills: make(map[string]int),
	}
}

func TestPlayerManager_InitializePlayer(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()

	pm.InitializePlayer(player)

	if player.Inventory == nil {
		t.Error("Player inventory should be initialized")
	}

	if player.Equipment == nil {
		t.Error("Player equipment should be initialized")
	}

	if player.Skills == nil {
		t.Error("Player skills should be initialized")
	}
}

func TestPlayerManager_ItemTemplates(t *testing.T) {
	pm := NewPlayerManager()

	template := &ItemTemplate{
		ID:          "test_sword",
		DisplayName: "Test Sword",
		SlotMask:    SlotMaskMainHand,
		Weight:      3.0,
		Bulk:        2,
		DamageType:  DamageSlash,
		SkillReq:    map[string]int{"melee": 10},
	}

	// Register template
	pm.RegisterItemTemplate(template)

	// Retrieve template
	retrieved, exists := pm.GetItemTemplate(template.ID)
	if !exists {
		t.Fatal("Template should exist after registration")
	}

	if retrieved.ID != template.ID {
		t.Errorf("Retrieved template ID %s != %s", retrieved.ID, template.ID)
	}

	// Test non-existent template
	_, exists = pm.GetItemTemplate("nonexistent")
	if exists {
		t.Error("Non-existent template should not be found")
	}
}

func TestPlayerManager_CheckSkillRequirements(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()
	pm.InitializePlayer(player)

	template := &ItemTemplate{
		ID:       "skill_item",
		SkillReq: map[string]int{"melee": 10, "defense": 5},
	}

	// Player has no skills initially
	if pm.CheckSkillRequirements(player, template) {
		t.Error("Player should not meet skill requirements initially")
	}

	// Give player partial skills
	player.Skills["melee"] = 10
	if pm.CheckSkillRequirements(player, template) {
		t.Error("Player should not meet all skill requirements yet")
	}

	// Give player all required skills
	player.Skills["defense"] = 5
	if !pm.CheckSkillRequirements(player, template) {
		t.Error("Player should meet all skill requirements now")
	}

	// Exceed requirements
	player.Skills["defense"] = 15
	if !pm.CheckSkillRequirements(player, template) {
		t.Error("Player should still meet requirements when exceeding them")
	}
}

func TestPlayerManager_AddItemToInventory(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()
	pm.InitializePlayer(player)

	template := &ItemTemplate{
		ID:     "test_item",
		Weight: 5.0,
		Bulk:   2,
	}
	pm.RegisterItemTemplate(template)

	instance := ItemInstance{
		InstanceID: "item1",
		TemplateID: template.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	// Test successful add
	err := pm.AddItemToInventory(player, instance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("AddItemToInventory failed: %v", err)
	}

	if !player.Inventory.HasItem(instance.InstanceID) {
		t.Error("Item should be in player inventory")
	}

	// Test adding unknown template
	unknownInstance := ItemInstance{
		InstanceID: "unknown1",
		TemplateID: "unknown_template",
		Quantity:   1,
	}

	err = pm.AddItemToInventory(player, unknownInstance, CompartmentBackpack)
	if err == nil {
		t.Error("Should fail when adding item with unknown template")
	}
}

func TestPlayerManager_EquipItem(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()
	pm.InitializePlayer(player)

	// Create and register a sword template
	swordTemplate := &ItemTemplate{
		ID:         "test_sword",
		SlotMask:   SlotMaskMainHand,
		Weight:     3.0,
		Bulk:       2,
		DamageType: DamageSlash,
		SkillReq:   map[string]int{"melee": 10},
	}
	pm.RegisterItemTemplate(swordTemplate)

	// Add sword to inventory
	swordInstance := ItemInstance{
		InstanceID: "sword1",
		TemplateID: swordTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err := pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword to inventory: %v", err)
	}

	now := time.Now()

	// Try to equip without skill (should fail)
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != ErrSkillGate {
		t.Errorf("Expected ErrSkillGate, got %v", err)
	}

	// Give player the required skill
	player.Skills["melee"] = 10

	// Equip the sword (should succeed)
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword: %v", err)
	}

	// Check that sword is equipped
	equippedItem := player.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil {
		t.Fatal("Sword should be equipped in main hand")
	}

	if equippedItem.Instance.InstanceID != swordInstance.InstanceID {
		t.Error("Wrong item equipped in main hand")
	}

	// Check that sword is no longer in inventory
	if player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Sword should not be in inventory after equipping")
	}

	// Try to equip to wrong slot (should fail)
	err = pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword back to inventory: %v", err)
	}

	err = pm.EquipItem(player, swordInstance.InstanceID, SlotChest, now)
	if err != ErrIllegalSlot {
		t.Errorf("Expected ErrIllegalSlot, got %v", err)
	}
}

func TestPlayerManager_EquipItemCooldown(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()
	pm.InitializePlayer(player)

	// Create sword template (no skill requirement for this test)
	swordTemplate := &ItemTemplate{
		ID:       "test_sword",
		SlotMask: SlotMaskMainHand,
		Weight:   3.0,
		Bulk:     2,
	}
	pm.RegisterItemTemplate(swordTemplate)

	// Add two swords to inventory
	sword1 := ItemInstance{
		InstanceID: "sword1",
		TemplateID: swordTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	sword2 := ItemInstance{
		InstanceID: "sword2",
		TemplateID: swordTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	pm.AddItemToInventory(player, sword1, CompartmentBackpack)
	pm.AddItemToInventory(player, sword2, CompartmentBackpack)

	now := time.Now()

	// Equip first sword
	err := pm.EquipItem(player, sword1.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip first sword: %v", err)
	}

	// Try to equip second sword immediately (should fail due to cooldown)
	err = pm.EquipItem(player, sword2.InstanceID, SlotMainHand, now)
	if err != ErrEquipLocked {
		t.Errorf("Expected ErrEquipLocked due to cooldown, got %v", err)
	}

	// Fast forward past cooldown and try again
	futureTime := now.Add(EquipCooldown + time.Second)
	err = pm.EquipItem(player, sword2.InstanceID, SlotMainHand, futureTime)
	if err != nil {
		t.Fatalf("Should be able to equip after cooldown expires: %v", err)
	}

	// Check that second sword is now equipped
	equippedItem := player.Equipment.GetSlot(SlotMainHand)
	if equippedItem.Instance.InstanceID != sword2.InstanceID {
		t.Error("Second sword should be equipped")
	}

	// Check that first sword is back in inventory
	if !player.Inventory.HasItem(sword1.InstanceID) {
		t.Error("First sword should be back in inventory")
	}
}

func TestPlayerManager_UnequipItem(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()
	pm.InitializePlayer(player)

	// Create and register a sword template
	swordTemplate := &ItemTemplate{
		ID:       "test_sword",
		SlotMask: SlotMaskMainHand,
		Weight:   3.0,
		Bulk:     2,
	}
	pm.RegisterItemTemplate(swordTemplate)

	// Add and equip sword
	swordInstance := ItemInstance{
		InstanceID: "sword1",
		TemplateID: swordTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)

	now := time.Now()
	err := pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword: %v", err)
	}

	// Try to unequip immediately (should fail due to cooldown)
	err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, now)
	if err != ErrEquipLocked {
		t.Errorf("Expected ErrEquipLocked, got %v", err)
	}

	// Fast forward past cooldown and unequip
	futureTime := now.Add(EquipCooldown + time.Second)
	err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, futureTime)
	if err != nil {
		t.Fatalf("Failed to unequip sword: %v", err)
	}

	// Check that slot is empty
	if !player.Equipment.IsSlotEmpty(SlotMainHand) {
		t.Error("Main hand slot should be empty after unequipping")
	}

	// Check that sword is back in inventory
	if !player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Sword should be back in inventory")
	}
}

func TestPlayerManager_GetPlayerEncumbrance(t *testing.T) {
	pm := NewPlayerManager()
	player := createTestPlayer()
	pm.InitializePlayer(player)

	// Create heavy item template
	heavyTemplate := &ItemTemplate{
		ID:     "heavy_item",
		Weight: 50.0, // Half of default weight limit
		Bulk:   10,
	}
	pm.RegisterItemTemplate(heavyTemplate)

	// Test with empty inventory
	encumbrance := pm.GetPlayerEncumbrance(player)
	if encumbrance.WeightPct != 0.0 {
		t.Errorf("Empty inventory should have 0%% weight, got %f", encumbrance.WeightPct)
	}

	// Add heavy item
	heavyInstance := ItemInstance{
		InstanceID: "heavy1",
		TemplateID: heavyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	pm.AddItemToInventory(player, heavyInstance, CompartmentBackpack)

	encumbrance = pm.GetPlayerEncumbrance(player)
	if encumbrance.WeightPct != 0.5 {
		t.Errorf("Expected 50%% weight, got %f", encumbrance.WeightPct)
	}

	if encumbrance.MovementPenalty != 1.0 {
		t.Errorf("No movement penalty expected at 50%% weight, got %f", encumbrance.MovementPenalty)
	}

	// Add another heavy item to trigger movement penalty
	heavyInstance2 := ItemInstance{
		InstanceID: "heavy2",
		TemplateID: heavyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	pm.AddItemToInventory(player, heavyInstance2, CompartmentBackpack)

	encumbrance = pm.GetPlayerEncumbrance(player)
	if encumbrance.WeightPct != 1.0 {
		t.Errorf("Expected 100%% weight, got %f", encumbrance.WeightPct)
	}

	if encumbrance.MovementPenalty >= 1.0 {
		t.Errorf("Movement penalty expected at 100%% weight, got %f", encumbrance.MovementPenalty)
	}
}

func TestPlayerManager_CreateTestItemTemplates(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	expectedTemplates := []ItemTemplateID{
		"sword_iron",
		"shield_wood",
		"armor_leather",
		"potion_health",
	}

	for _, templateID := range expectedTemplates {
		template, exists := pm.GetItemTemplate(templateID)
		if !exists {
			t.Errorf("Expected template %s to exist", templateID)
			continue
		}

		if template.DisplayName == "" {
			t.Errorf("Template %s should have display name", templateID)
		}

		if template.Weight < 0 {
			t.Errorf("Template %s should have non-negative weight", templateID)
		}

		if template.Bulk < 0 {
			t.Errorf("Template %s should have non-negative bulk", templateID)
		}
	}
}
