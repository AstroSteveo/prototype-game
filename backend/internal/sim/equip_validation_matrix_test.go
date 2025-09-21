package sim

import (
	"testing"
	"time"
)

// TestEquipValidationMatrix provides comprehensive integration testing of the
// equipment validation system, covering all possible combinations of:
// - Slot compatibility (R14)
// - Skill requirements (R15)
// - Cooldown validation
// - Success and error scenarios
func TestEquipValidationMatrix(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	// Add more test items for comprehensive matrix testing
	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "bow_elven",
		DisplayName: "Elven Bow",
		SlotMask:    SlotMaskMainHand,
		Weight:      2.0,
		Bulk:        3,
		DamageType:  DamagePierce,
		SkillReq:    map[string]int{"archery": 15, "dexterity": 12},
	})

	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "helmet_iron",
		DisplayName: "Iron Helmet",
		SlotMask:    SlotMaskHead,
		Weight:      2.5,
		Bulk:        2,
		DamageType:  "",
		SkillReq:    map[string]int{"armor": 8},
	})

	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "boots_leather",
		DisplayName: "Leather Boots",
		SlotMask:    SlotMaskFeet,
		Weight:      1.0,
		Bulk:        2,
		DamageType:  "",
		SkillReq:    map[string]int{},
	})

	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "legs_chain",
		DisplayName: "Chain Leggings",
		SlotMask:    SlotMaskLegs,
		Weight:      4.0,
		Bulk:        3,
		DamageType:  "",
		SkillReq:    map[string]int{"armor": 10},
	})

	// Test matrix structure
	type ValidationCase struct {
		name           string
		itemTemplateID ItemTemplateID
		targetSlot     SlotID
		playerSkills   map[string]int
		expectError    error
		description    string
	}

	// Comprehensive validation matrix
	cases := []ValidationCase{
		// === SLOT COMPATIBILITY MATRIX (R14) ===
		// Main Hand items
		{
			name:           "Sword_MainHand_Valid",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    nil,
			description:    "Sword equipped to main hand - valid slot",
		},
		{
			name:           "Sword_OffHand_Invalid",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotOffHand,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    ErrIllegalSlot,
			description:    "Sword equipped to off hand - invalid slot",
		},
		{
			name:           "Sword_Chest_Invalid",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotChest,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    ErrIllegalSlot,
			description:    "Sword equipped to chest - invalid slot",
		},
		{
			name:           "Sword_Head_Invalid",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotHead,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    ErrIllegalSlot,
			description:    "Sword equipped to head - invalid slot",
		},
		{
			name:           "Sword_Legs_Invalid",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotLegs,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    ErrIllegalSlot,
			description:    "Sword equipped to legs - invalid slot",
		},
		{
			name:           "Sword_Feet_Invalid",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotFeet,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    ErrIllegalSlot,
			description:    "Sword equipped to feet - invalid slot",
		},

		// Off Hand items
		{
			name:           "Shield_OffHand_Valid",
			itemTemplateID: "shield_wood",
			targetSlot:     SlotOffHand,
			playerSkills:   map[string]int{"defense": 5},
			expectError:    nil,
			description:    "Shield equipped to off hand - valid slot",
		},
		{
			name:           "Shield_MainHand_Invalid",
			itemTemplateID: "shield_wood",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"defense": 5},
			expectError:    ErrIllegalSlot,
			description:    "Shield equipped to main hand - invalid slot",
		},

		// Chest armor
		{
			name:           "Armor_Chest_Valid",
			itemTemplateID: "armor_leather",
			targetSlot:     SlotChest,
			playerSkills:   map[string]int{},
			expectError:    nil,
			description:    "Leather armor equipped to chest - valid slot",
		},
		{
			name:           "Armor_MainHand_Invalid",
			itemTemplateID: "armor_leather",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{},
			expectError:    ErrIllegalSlot,
			description:    "Armor equipped to main hand - invalid slot",
		},

		// Head armor
		{
			name:           "Helmet_Head_Valid",
			itemTemplateID: "helmet_iron",
			targetSlot:     SlotHead,
			playerSkills:   map[string]int{"armor": 8},
			expectError:    nil,
			description:    "Helmet equipped to head - valid slot",
		},
		{
			name:           "Helmet_Chest_Invalid",
			itemTemplateID: "helmet_iron",
			targetSlot:     SlotChest,
			playerSkills:   map[string]int{"armor": 8},
			expectError:    ErrIllegalSlot,
			description:    "Helmet equipped to chest - invalid slot",
		},

		// Legs armor
		{
			name:           "Legs_Legs_Valid",
			itemTemplateID: "legs_chain",
			targetSlot:     SlotLegs,
			playerSkills:   map[string]int{"armor": 10},
			expectError:    nil,
			description:    "Chain leggings equipped to legs - valid slot",
		},
		{
			name:           "Legs_Feet_Invalid",
			itemTemplateID: "legs_chain",
			targetSlot:     SlotFeet,
			playerSkills:   map[string]int{"armor": 10},
			expectError:    ErrIllegalSlot,
			description:    "Leggings equipped to feet - invalid slot",
		},

		// Feet armor
		{
			name:           "Boots_Feet_Valid",
			itemTemplateID: "boots_leather",
			targetSlot:     SlotFeet,
			playerSkills:   map[string]int{},
			expectError:    nil,
			description:    "Boots equipped to feet - valid slot",
		},
		{
			name:           "Boots_Head_Invalid",
			itemTemplateID: "boots_leather",
			targetSlot:     SlotHead,
			playerSkills:   map[string]int{},
			expectError:    ErrIllegalSlot,
			description:    "Boots equipped to head - invalid slot",
		},

		// === SKILL REQUIREMENT MATRIX (R15) ===
		// Single skill requirement scenarios
		{
			name:           "Sword_NoSkills_Fail",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{},
			expectError:    ErrSkillGate,
			description:    "Sword with no skills - skill gate failure",
		},
		{
			name:           "Sword_InsufficientSkill_Fail",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"melee": 5},
			expectError:    ErrSkillGate,
			description:    "Sword with insufficient skill level - skill gate failure",
		},
		{
			name:           "Sword_ExactSkill_Success",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"melee": 10},
			expectError:    nil,
			description:    "Sword with exact skill requirement - success",
		},
		{
			name:           "Sword_ExcessSkill_Success",
			itemTemplateID: "sword_iron",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"melee": 20},
			expectError:    nil,
			description:    "Sword with excess skill level - success",
		},

		// Multiple skill requirement scenarios
		{
			name:           "Bow_NoSkills_Fail",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{},
			expectError:    ErrSkillGate,
			description:    "Bow with no skills - multiple skill gate failure",
		},
		{
			name:           "Bow_PartialSkills_Fail",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"archery": 20},
			expectError:    ErrSkillGate,
			description:    "Bow with only archery skill - missing dexterity requirement",
		},
		{
			name:           "Bow_PartialSkills2_Fail",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"dexterity": 15},
			expectError:    ErrSkillGate,
			description:    "Bow with only dexterity skill - missing archery requirement",
		},
		{
			name:           "Bow_BothSkillsInsufficient_Fail",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"archery": 10, "dexterity": 8},
			expectError:    ErrSkillGate,
			description:    "Bow with both skills but insufficient levels",
		},
		{
			name:           "Bow_OneSkillInsufficient_Fail",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"archery": 20, "dexterity": 8},
			expectError:    ErrSkillGate,
			description:    "Bow with sufficient archery but insufficient dexterity",
		},
		{
			name:           "Bow_ExactSkills_Success",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"archery": 15, "dexterity": 12},
			expectError:    nil,
			description:    "Bow with exact skill requirements - success",
		},
		{
			name:           "Bow_ExcessSkills_Success",
			itemTemplateID: "bow_elven",
			targetSlot:     SlotMainHand,
			playerSkills:   map[string]int{"archery": 25, "dexterity": 20, "cooking": 50},
			expectError:    nil,
			description:    "Bow with excess skills and extra skills - success",
		},

		// No skill requirement scenarios
		{
			name:           "Boots_NoRequirement_Success",
			itemTemplateID: "boots_leather",
			targetSlot:     SlotFeet,
			playerSkills:   map[string]int{},
			expectError:    nil,
			description:    "Boots with no skill requirements - success",
		},
		{
			name:           "Armor_NoRequirement_Success",
			itemTemplateID: "armor_leather",
			targetSlot:     SlotChest,
			playerSkills:   map[string]int{"melee": 100},
			expectError:    nil,
			description:    "Armor with no requirements but player has skills - success",
		},
	}

	// Execute all validation matrix cases
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh player for each test
			player := &Player{
				Entity:    Entity{ID: "test_player", Name: "Test Player"},
				Inventory: NewInventory(),
				Equipment: NewEquipment(),
				Skills:    make(map[string]int),
			}
			
			// Set player skills for this test case
			for skill, level := range tc.playerSkills {
				player.Skills[skill] = level
			}
			
			pm.InitializePlayer(player)

			// Add test item to inventory
			instance := ItemInstance{
				InstanceID: ItemInstanceID("test_" + string(tc.itemTemplateID)),
				TemplateID: tc.itemTemplateID,
				Quantity:   1,
				Durability: 1.0,
			}

			err := pm.AddItemToInventory(player, instance, CompartmentBackpack)
			if err != nil {
				t.Fatalf("Failed to add item to inventory: %v", err)
			}

			// Attempt to equip the item
			now := time.Now()
			err = pm.EquipItem(player, instance.InstanceID, tc.targetSlot, now)

			// Verify expected result
			if tc.expectError != nil {
				if err != tc.expectError {
					t.Errorf("Case '%s': Expected error %v, got %v. %s", 
						tc.name, tc.expectError, err, tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Case '%s': Expected success, got error %v. %s", 
						tc.name, err, tc.description)
				} else {
					// Verify item was actually equipped
					equippedItem := player.Equipment.GetSlot(tc.targetSlot)
					if equippedItem == nil || equippedItem.Instance.InstanceID != instance.InstanceID {
						t.Errorf("Case '%s': Item was not properly equipped. %s", 
							tc.name, tc.description)
					}
					// Verify item was removed from inventory
					if player.Inventory.HasItem(instance.InstanceID) {
						t.Errorf("Case '%s': Item should not remain in inventory when equipped. %s", 
							tc.name, tc.description)
					}
				}
			}
		})
	}
}

// TestEquipCooldownValidationMatrix tests all cooldown-related scenarios
func TestEquipCooldownValidationMatrix(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	type CooldownCase struct {
		name           string
		equipItem      ItemInstanceID
		unequipSlot    SlotID
		cooldownOffset time.Duration
		expectError    error
		description    string
	}

	cases := []CooldownCase{
		{
			name:           "Unequip_ImmediatelyAfterEquip_Fail",
			equipItem:      "sword_001",
			unequipSlot:    SlotMainHand,
			cooldownOffset: 0,
			expectError:    ErrEquipLocked,
			description:    "Unequip immediately after equip - cooldown active",
		},
		{
			name:           "Unequip_DuringCooldown_Fail",
			equipItem:      "sword_001",
			unequipSlot:    SlotMainHand,
			cooldownOffset: EquipCooldown / 2,
			expectError:    ErrEquipLocked,
			description:    "Unequip during cooldown period - cooldown still active",
		},
		{
			name:           "Unequip_JustBeforeExpiry_Fail",
			equipItem:      "sword_001",
			unequipSlot:    SlotMainHand,
			cooldownOffset: EquipCooldown - time.Millisecond,
			expectError:    ErrEquipLocked,
			description:    "Unequip just before cooldown expiry - still on cooldown",
		},
		{
			name:           "Unequip_ExactExpiry_Success",
			equipItem:      "sword_001",
			unequipSlot:    SlotMainHand,
			cooldownOffset: EquipCooldown,
			expectError:    nil,
			description:    "Unequip exactly at cooldown expiry - success",
		},
		{
			name:           "Unequip_AfterExpiry_Success",
			equipItem:      "sword_001",
			unequipSlot:    SlotMainHand,
			cooldownOffset: EquipCooldown + time.Second,
			expectError:    nil,
			description:    "Unequip after cooldown expiry - success",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh player
			player := &Player{
				Entity:    Entity{ID: "test_player", Name: "Test Player"},
				Inventory: NewInventory(),
				Equipment: NewEquipment(),
				Skills:    map[string]int{"melee": 10}, // Sufficient skills
			}
			pm.InitializePlayer(player)

			// Add sword to inventory
			swordInstance := ItemInstance{
				InstanceID: tc.equipItem,
				TemplateID: "sword_iron",
				Quantity:   1,
				Durability: 1.0,
			}

			err := pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
			if err != nil {
				t.Fatalf("Failed to add sword to inventory: %v", err)
			}

			now := time.Now()

			// Equip the item
			err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
			if err != nil {
				t.Fatalf("Failed to equip sword: %v", err)
			}

			// Attempt unequip with cooldown offset
			unequipTime := now.Add(tc.cooldownOffset)
			err = pm.UnequipItem(player, tc.unequipSlot, CompartmentBackpack, unequipTime)

			// Verify expected result
			if tc.expectError != nil {
				if err != tc.expectError {
					t.Errorf("Case '%s': Expected error %v, got %v. %s", 
						tc.name, tc.expectError, err, tc.description)
				}
				// Verify item is still equipped when error expected
				equippedItem := player.Equipment.GetSlot(tc.unequipSlot)
				if equippedItem == nil || equippedItem.Instance.InstanceID != tc.equipItem {
					t.Errorf("Case '%s': Item should still be equipped after failed unequip. %s", 
						tc.name, tc.description)
				}
			} else {
				if err != nil {
					t.Errorf("Case '%s': Expected success, got error %v. %s", 
						tc.name, err, tc.description)
				} else {
					// Verify item was unequipped
					if !player.Equipment.IsSlotEmpty(tc.unequipSlot) {
						t.Errorf("Case '%s': Slot should be empty after successful unequip. %s", 
							tc.name, tc.description)
					}
					// Verify item returned to inventory
					if !player.Inventory.HasItem(tc.equipItem) {
						t.Errorf("Case '%s': Item should be back in inventory after unequip. %s", 
							tc.name, tc.description)
					}
				}
			}
		})
	}
}

// TestEquipSlotSwappingMatrix tests equipment swapping scenarios
func TestEquipSlotSwappingMatrix(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	// Create player with sufficient skills
	player := &Player{
		Entity:    Entity{ID: "test_player", Name: "Test Player"},
		Inventory: NewInventory(),
		Equipment: NewEquipment(),
		Skills:    map[string]int{"melee": 10, "defense": 5},
	}
	pm.InitializePlayer(player)

	// Add multiple items to inventory
	sword1 := ItemInstance{
		InstanceID: "sword_001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}
	sword2 := ItemInstance{
		InstanceID: "sword_002",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 0.8,
	}
	shield := ItemInstance{
		InstanceID: "shield_001",
		TemplateID: "shield_wood",
		Quantity:   1,
		Durability: 1.0,
	}

	items := []ItemInstance{sword1, sword2, shield}
	for _, item := range items {
		err := pm.AddItemToInventory(player, item, CompartmentBackpack)
		if err != nil {
			t.Fatalf("Failed to add item %s to inventory: %v", item.InstanceID, err)
		}
	}

	now := time.Now()

	// Test 1: Initial equip
	err := pm.EquipItem(player, sword1.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip first sword: %v", err)
	}

	// Verify first sword is equipped
	equippedItem := player.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil || equippedItem.Instance.InstanceID != sword1.InstanceID {
		t.Error("First sword should be equipped")
	}
	if player.Inventory.HasItem(sword1.InstanceID) {
		t.Error("First sword should not be in inventory when equipped")
	}

	// Test 2: Attempt swap during cooldown (should fail)
	err = pm.EquipItem(player, sword2.InstanceID, SlotMainHand, now)
	if err != ErrEquipLocked {
		t.Errorf("Expected ErrEquipLocked when swapping during cooldown, got: %v", err)
	}

	// Test 3: Successful swap after cooldown
	futureTime := now.Add(EquipCooldown + time.Second)
	err = pm.EquipItem(player, sword2.InstanceID, SlotMainHand, futureTime)
	if err != nil {
		t.Fatalf("Failed to swap swords after cooldown: %v", err)
	}

	// Verify swap occurred correctly
	equippedItem = player.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil || equippedItem.Instance.InstanceID != sword2.InstanceID {
		t.Error("Second sword should be equipped after swap")
	}
	if !player.Inventory.HasItem(sword1.InstanceID) {
		t.Error("First sword should be back in inventory after swap")
	}
	if player.Inventory.HasItem(sword2.InstanceID) {
		t.Error("Second sword should not be in inventory when equipped")
	}

	// Test 4: Equip to different slot (no swap needed)
	err = pm.EquipItem(player, shield.InstanceID, SlotOffHand, futureTime)
	if err != nil {
		t.Fatalf("Failed to equip shield to off hand: %v", err)
	}

	// Verify both items equipped in different slots
	mainHandItem := player.Equipment.GetSlot(SlotMainHand)
	offHandItem := player.Equipment.GetSlot(SlotOffHand)
	
	if mainHandItem == nil || mainHandItem.Instance.InstanceID != sword2.InstanceID {
		t.Error("Sword should still be in main hand")
	}
	if offHandItem == nil || offHandItem.Instance.InstanceID != shield.InstanceID {
		t.Error("Shield should be in off hand")
	}
	
	// Verify inventory state
	if !player.Inventory.HasItem(sword1.InstanceID) {
		t.Error("First sword should still be in inventory")
	}
	if player.Inventory.HasItem(sword2.InstanceID) {
		t.Error("Second sword should not be in inventory (equipped)")
	}
	if player.Inventory.HasItem(shield.InstanceID) {
		t.Error("Shield should not be in inventory (equipped)")
	}
}

// TestEquipVersionTrackingMatrix verifies version tracking for all operations
func TestEquipVersionTrackingMatrix(t *testing.T) {
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

	// Track version changes through complete equip/unequip cycle
	initialInventoryVersion := player.InventoryVersion
	initialEquipmentVersion := player.EquipmentVersion

	// Add item - should increment inventory version only
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

	if player.InventoryVersion <= initialInventoryVersion {
		t.Error("Adding item should increment inventory version")
	}
	if player.EquipmentVersion != initialEquipmentVersion {
		t.Error("Adding item should not increment equipment version")
	}

	afterAddInventoryVersion := player.InventoryVersion
	afterAddEquipmentVersion := player.EquipmentVersion

	// Equip item - should increment both versions
	now := time.Now()
	err = pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, now)
	if err != nil {
		t.Fatalf("Failed to equip sword: %v", err)
	}

	if player.InventoryVersion <= afterAddInventoryVersion {
		t.Error("Equipping should increment inventory version")
	}
	if player.EquipmentVersion <= afterAddEquipmentVersion {
		t.Error("Equipping should increment equipment version")
	}

	afterEquipInventoryVersion := player.InventoryVersion
	afterEquipEquipmentVersion := player.EquipmentVersion

	// Unequip item - should increment both versions
	futureTime := now.Add(EquipCooldown + time.Second)
	err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, futureTime)
	if err != nil {
		t.Fatalf("Failed to unequip sword: %v", err)
	}

	if player.InventoryVersion <= afterEquipInventoryVersion {
		t.Error("Unequipping should increment inventory version")
	}
	if player.EquipmentVersion <= afterEquipEquipmentVersion {
		t.Error("Unequipping should increment equipment version")
	}

	// Verify final state
	if player.Equipment.GetSlot(SlotMainHand) != nil {
		t.Error("Slot should be empty after unequip")
	}
	if !player.Inventory.HasItem(swordInstance.InstanceID) {
		t.Error("Item should be back in inventory after unequip")
	}
}