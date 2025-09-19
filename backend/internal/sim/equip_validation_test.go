package sim

import (
	"testing"
	"time"
)

func TestPlayerManagerEquipValidation(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	// Create test player
	player := &Player{
		Entity:    Entity{ID: "test_player", Name: "Test Player"},
		Inventory: NewInventory(),
		Equipment: NewEquipment(),
		Skills:    make(map[string]int),
	}
	pm.InitializePlayer(player)

	// Add sword to inventory
	swordInstance := ItemInstance{
		InstanceID: "sword_001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}

	err := pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword to inventory: %v", err)
	}

	now := time.Now()

	// Test 1: Equip without skill requirement (should fail)
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != ErrSkillGate {
		t.Errorf("Expected ErrSkillGate, got: %v", err)
	}

	// Test 2: Give skill and equip (should succeed)
	player.Skills["melee"] = 10
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword with skill: %v", err)
	}

	// Test 3: Try to equip to wrong slot (should fail)
	// Add another sword first
	sword2Instance := ItemInstance{
		InstanceID: "sword_002",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}
	err = pm.AddItemToInventory(player, sword2Instance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add second sword: %v", err)
	}

	err = pm.EquipItem(player, sword2Instance.InstanceID, SlotChest, now)
	if err != ErrIllegalSlot {
		t.Errorf("Expected ErrIllegalSlot, got: %v", err)
	}

	// Test 4: Try to unequip during cooldown (should fail)
	err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, now)
	if err != ErrEquipLocked {
		t.Errorf("Expected ErrEquipLocked, got: %v", err)
	}

	// Test 5: Unequip after cooldown (should succeed)
	futureTime := now.Add(EquipCooldown + time.Second)
	err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, futureTime)
	if err != nil {
		t.Errorf("Failed to unequip after cooldown: %v", err)
	}

	// Verify sword is back in inventory
	if !player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Sword should be back in inventory after unequipping")
	}
}

func TestEquipmentSlotValidation(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	// Test different item types and their allowed slots
	testCases := []struct {
		templateID   ItemTemplateID
		validSlots   []SlotID
		invalidSlots []SlotID
	}{
		{
			templateID:   "sword_iron",
			validSlots:   []SlotID{SlotMainHand},
			invalidSlots: []SlotID{SlotOffHand, SlotChest, SlotLegs, SlotFeet, SlotHead},
		},
		{
			templateID:   "shield_wood",
			validSlots:   []SlotID{SlotOffHand},
			invalidSlots: []SlotID{SlotMainHand, SlotChest, SlotLegs, SlotFeet, SlotHead},
		},
		{
			templateID:   "armor_leather",
			validSlots:   []SlotID{SlotChest},
			invalidSlots: []SlotID{SlotMainHand, SlotOffHand, SlotLegs, SlotFeet, SlotHead},
		},
	}

	for _, tc := range testCases {
		template, exists := pm.GetItemTemplate(tc.templateID)
		if !exists {
			t.Fatalf("Template %s should exist", tc.templateID)
		}

		// Test valid slots
		for _, slot := range tc.validSlots {
			if !template.Allows(slot) {
				t.Errorf("Template %s should allow slot %s", tc.templateID, slot)
			}
		}

		// Test invalid slots
		for _, slot := range tc.invalidSlots {
			if template.Allows(slot) {
				t.Errorf("Template %s should not allow slot %s", tc.templateID, slot)
			}
		}
	}
}

func TestEquipmentCooldownSystem(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	player := &Player{
		Entity:    Entity{ID: "test_player", Name: "Test Player"},
		Inventory: NewInventory(),
		Equipment: NewEquipment(),
		Skills:    map[string]int{"melee": 10}, // Has required skill
	}
	pm.InitializePlayer(player)

	// Add sword to inventory
	swordInstance := ItemInstance{
		InstanceID: "sword_001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}
	err := pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword: %v", err)
	}

	now := time.Now()

	// Equip sword
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword: %v", err)
	}

	// Verify cooldown is active
	if !player.Equipment.IsSlotOnCooldown(SlotMainHand, now) {
		t.Error("Slot should be on cooldown immediately after equipping")
	}

	// Verify cooldown expires
	futureTime := now.Add(EquipCooldown + time.Millisecond)
	if player.Equipment.IsSlotOnCooldown(SlotMainHand, futureTime) {
		t.Error("Slot should not be on cooldown after cooldown period")
	}

	// Test cooldown duration
	equippedItem := player.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil {
		t.Fatal("Should have equipped item")
	}

	expectedCooldownEnd := now.Add(EquipCooldown)
	if !equippedItem.CooldownUntil.Equal(expectedCooldownEnd) {
		t.Errorf("Cooldown should end at %v, got %v", expectedCooldownEnd, equippedItem.CooldownUntil)
	}
}

func TestEquipmentSlotSwapping(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	player := &Player{
		Entity:    Entity{ID: "test_player", Name: "Test Player"},
		Inventory: NewInventory(),
		Equipment: NewEquipment(),
		Skills:    map[string]int{"melee": 10},
	}
	pm.InitializePlayer(player)

	// Add two swords to inventory
	sword1Instance := ItemInstance{
		InstanceID: "sword_001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}
	sword2Instance := ItemInstance{
		InstanceID: "sword_002",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}

	err := pm.AddItemToInventory(player, sword1Instance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add first sword: %v", err)
	}
	err = pm.AddItemToInventory(player, sword2Instance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add second sword: %v", err)
	}

	now := time.Now()

	// Equip first sword
	err = pm.EquipItem(player, sword1Instance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip first sword: %v", err)
	}

	// Verify first sword is equipped and not in inventory
	equippedItem := player.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil || equippedItem.Instance.InstanceID != sword1Instance.InstanceID {
		t.Error("First sword should be equipped")
	}
	if player.Inventory.HasItem(sword1Instance.InstanceID) {
		t.Error("First sword should not be in inventory when equipped")
	}

	// Try to equip second sword (should swap - move first back to inventory)
	futureTime := now.Add(EquipCooldown + time.Second) // After cooldown
	err = pm.EquipItem(player, sword2Instance.InstanceID, SlotMainHand, futureTime)
	if err != nil {
		t.Fatalf("Failed to swap swords: %v", err)
	}

	// Verify swap occurred
	equippedItem = player.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil || equippedItem.Instance.InstanceID != sword2Instance.InstanceID {
		t.Error("Second sword should be equipped after swap")
	}
	if !player.Inventory.HasItem(sword1Instance.InstanceID) {
		t.Error("First sword should be back in inventory after swap")
	}
	if player.Inventory.HasItem(sword2Instance.InstanceID) {
		t.Error("Second sword should not be in inventory when equipped")
	}
}

func TestEquipmentVersionTracking(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	player := &Player{
		Entity:           Entity{ID: "test_player", Name: "Test Player"},
		Inventory:        NewInventory(),
		Equipment:        NewEquipment(),
		Skills:           map[string]int{"melee": 10},
		InventoryVersion: 0,
		EquipmentVersion: 0,
	}
	pm.InitializePlayer(player)

	// Add sword to inventory
	swordInstance := ItemInstance{
		InstanceID: "sword_001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}
	err := pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add sword: %v", err)
	}

	// Check that adding to inventory increments inventory version
	if player.InventoryVersion != 1 {
		t.Errorf("Expected inventory version 1, got %d", player.InventoryVersion)
	}

	initialInventoryVersion := player.InventoryVersion
	initialEquipmentVersion := player.EquipmentVersion

	now := time.Now()

	// Equip sword - should increment both versions
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword: %v", err)
	}

	if player.InventoryVersion <= initialInventoryVersion {
		t.Error("Equipping should increment inventory version")
	}
	if player.EquipmentVersion <= initialEquipmentVersion {
		t.Error("Equipping should increment equipment version")
	}

	preUnequipInventoryVersion := player.InventoryVersion
	preUnequipEquipmentVersion := player.EquipmentVersion

	// Unequip sword - should increment both versions again
	futureTime := now.Add(EquipCooldown + time.Second)
	err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, futureTime)
	if err != nil {
		t.Fatalf("Failed to unequip sword: %v", err)
	}

	if player.InventoryVersion <= preUnequipInventoryVersion {
		t.Error("Unequipping should increment inventory version")
	}
	if player.EquipmentVersion <= preUnequipEquipmentVersion {
		t.Error("Unequipping should increment equipment version")
	}
}

func TestSkillRequirementEdgeCases(t *testing.T) {
	pm := NewPlayerManager()

	// Create item with multiple skill requirements
	complexItem := &ItemTemplate{
		ID:          "complex_weapon",
		DisplayName: "Complex Weapon",
		SlotMask:    SlotMaskMainHand,
		Weight:      5.0,
		Bulk:        3,
		DamageType:  DamageSlash,
		SkillReq: map[string]int{
			"melee":  15,
			"arcane": 10,
			"craft":  5,
		},
	}
	pm.RegisterItemTemplate(complexItem)

	testCases := []struct {
		name         string
		playerSkills map[string]int
		shouldAllow  bool
	}{
		{
			name:         "no skills",
			playerSkills: map[string]int{},
			shouldAllow:  false,
		},
		{
			name:         "partial skills",
			playerSkills: map[string]int{"melee": 20, "arcane": 5},
			shouldAllow:  false,
		},
		{
			name:         "exact requirements",
			playerSkills: map[string]int{"melee": 15, "arcane": 10, "craft": 5},
			shouldAllow:  true,
		},
		{
			name:         "exceed requirements",
			playerSkills: map[string]int{"melee": 20, "arcane": 15, "craft": 10},
			shouldAllow:  true,
		},
		{
			name:         "extra skills don't matter",
			playerSkills: map[string]int{"melee": 15, "arcane": 10, "craft": 5, "cooking": 100},
			shouldAllow:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			player := &Player{Skills: tc.playerSkills}
			result := pm.CheckSkillRequirements(player, complexItem)
			if result != tc.shouldAllow {
				t.Errorf("Expected %v, got %v for skills %v", tc.shouldAllow, result, tc.playerSkills)
			}
		})
	}
}
