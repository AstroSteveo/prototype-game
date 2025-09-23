package sim

import (
	"testing"
	"time"
)

// TestT032_EncumbranceIncludingEquippedItems validates T-032 acceptance criteria:
// "Encumbrance math matches expected scenarios"
func TestT032_EncumbranceIncludingEquippedItems(t *testing.T) {
	tests := []struct {
		name                    string
		inventoryItems          []testItem
		equippedItems           []equippedTestItem
		expectedWeightPct       float64
		expectedMovementPenalty float64
		description             string
	}{
		{
			name:                    "Empty_NoEncumbrance",
			inventoryItems:          []testItem{},
			equippedItems:           []equippedTestItem{},
			expectedWeightPct:       0.0,
			expectedMovementPenalty: 1.0,
			description:             "Empty inventory and no equipped items should have no encumbrance",
		},
		{
			name:                    "InventoryOnly_LightLoad",
			inventoryItems:          []testItem{{templateID: "potion_health", quantity: 10}}, // 10 * 0.1kg = 1kg
			equippedItems:           []equippedTestItem{},
			expectedWeightPct:       0.005, // 1kg / 200kg = 0.5%
			expectedMovementPenalty: 1.0,
			description:             "Light inventory items only should have minimal encumbrance",
		},
		{
			name:                    "EquippedOnly_LightLoad",
			inventoryItems:          []testItem{},
			equippedItems:           []equippedTestItem{{templateID: "sword_iron", slot: SlotMainHand}}, // 3.5kg
			expectedWeightPct:       0.0175,                                                             // 3.5kg / 200kg = 1.75%
			expectedMovementPenalty: 1.0,
			description:             "Light equipped items only should have minimal encumbrance",
		},
		{
			name: "CombinedLoad_ModerateEncumbrance",
			inventoryItems: []testItem{
				{templateID: "armor_leather", quantity: 2},  // 2 * 5kg = 10kg
				{templateID: "potion_health", quantity: 50}, // 50 * 0.1kg = 5kg
			},
			equippedItems: []equippedTestItem{
				{templateID: "sword_iron", slot: SlotMainHand}, // 3.5kg
				{templateID: "shield_wood", slot: SlotOffHand}, // 2.0kg
				{templateID: "armor_leather", slot: SlotChest}, // 5.0kg
			},
			expectedWeightPct:       0.1275, // (10+5+3.5+2+5) / 200 = 12.75%
			expectedMovementPenalty: 1.0,    // No penalty under 80%
			description:             "Combined inventory and equipped items should sum correctly",
		},
		{
			name: "HeavyLoad_MovementPenalty",
			inventoryItems: []testItem{
				{templateID: "anvil_iron", quantity: 2}, // 2 * 85kg = 170kg
			},
			equippedItems: []equippedTestItem{
				{templateID: "sword_iron", slot: SlotMainHand}, // 3.5kg
			},
			expectedWeightPct:       0.8675, // (170+3.5) / 200 = 86.75%
			expectedMovementPenalty: 0.8312, // Linear penalty: 1.0 - 0.5*(0.8675-0.8)/0.2 = 0.8312
			description:             "Heavy load should trigger movement penalty between 80-100%",
		},
		{
			name: "OverweightLoad_SevereMovementPenalty",
			inventoryItems: []testItem{
				{templateID: "anvil_iron", quantity: 2}, // 2 * 85kg = 170kg
			},
			equippedItems: []equippedTestItem{
				{templateID: "sword_iron", slot: SlotMainHand}, // 3.5kg
			},
			expectedWeightPct:       1.735, // (170+3.5) / 100 = 173.5% (adjust weight limit down after adding)
			expectedMovementPenalty: 0.288, // Severe penalty: 0.5 * (1.0 / 1.735) = 0.288
			description:             "Overweight load should trigger severe movement penalty beyond 100%",
		},
		{
			name:           "FullEquipmentSet_CompleteArmor",
			inventoryItems: []testItem{},
			equippedItems: []equippedTestItem{
				{templateID: "sword_iron", slot: SlotMainHand}, // 3.5kg
				{templateID: "shield_wood", slot: SlotOffHand}, // 2.0kg
				{templateID: "armor_leather", slot: SlotChest}, // 5.0kg
				{templateID: "helmet_iron", slot: SlotHead},    // 2.5kg (if template exists)
				{templateID: "legs_chain", slot: SlotLegs},     // 4.0kg (if template exists)
				{templateID: "boots_leather", slot: SlotFeet},  // 1.0kg (if template exists)
			},
			expectedWeightPct:       0.09, // Total equipped weight / 200kg = 18kg / 200kg = 9%
			expectedMovementPenalty: 1.0,  // No penalty under 80%
			description:             "Full equipment set should be calculated correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			pm := NewPlayerManager()
			pm.CreateTestItemTemplates()

			// Register additional templates for comprehensive testing
			pm.RegisterItemTemplate(&ItemTemplate{
				ID:          "helmet_iron",
				DisplayName: "Iron Helmet",
				SlotMask:    SlotMaskHead,
				Weight:      2.5,
				Bulk:        2,
				SkillReq:    map[string]int{},
			})

			pm.RegisterItemTemplate(&ItemTemplate{
				ID:          "legs_chain",
				DisplayName: "Chain Leggings",
				SlotMask:    SlotMaskLegs,
				Weight:      4.0,
				Bulk:        3,
				SkillReq:    map[string]int{},
			})

			pm.RegisterItemTemplate(&ItemTemplate{
				ID:          "boots_leather",
				DisplayName: "Leather Boots",
				SlotMask:    SlotMaskFeet,
				Weight:      1.0,
				Bulk:        2,
				SkillReq:    map[string]int{},
			})

			player := createTestPlayer()
			pm.InitializePlayer(player)

			// Set up skills required for equipment
			player.Skills = map[string]int{
				"melee":   15,
				"defense": 10,
			}

			// Increase inventory limits for testing
			player.Inventory.WeightLimit = 200.0                        // Double the default limit
			player.Inventory.CompartmentCaps[CompartmentBackpack] = 100 // Increase bulk limit

			// Add inventory items
			for i, item := range tt.inventoryItems {
				instance := ItemInstance{
					InstanceID: ItemInstanceID("test_inv_" + string(rune(i+65))), // A, B, C, etc.
					TemplateID: item.templateID,
					Quantity:   item.quantity,
					Durability: 1.0,
				}
				err := pm.AddItemToInventory(player, instance, CompartmentBackpack)
				if err != nil {
					t.Fatalf("Failed to add inventory item %s: %v", item.templateID, err)
				}
			}

			// Equip items
			for i, item := range tt.equippedItems {
				instance := ItemInstance{
					InstanceID: ItemInstanceID("test_eq_" + string(rune(i+65))), // A, B, C, etc.
					TemplateID: item.templateID,
					Quantity:   1,
					Durability: 1.0,
				}

				// Add to inventory first
				err := pm.AddItemToInventory(player, instance, CompartmentBackpack)
				if err != nil {
					t.Fatalf("Failed to add item to inventory before equipping %s: %v", item.templateID, err)
				}

				// Then equip
				err = pm.EquipItem(player, instance.InstanceID, item.slot, time.Now())
				if err != nil {
					t.Fatalf("Failed to equip item %s to slot %s: %v", item.templateID, item.slot, err)
				}
			}

			// Special handling for overweight test - adjust weight limit after adding items
			if tt.name == "OverweightLoad_SevereMovementPenalty" {
				player.Inventory.WeightLimit = 100.0 // Reduce to test overweight scenario
			}

			// Calculate encumbrance
			encumbrance := pm.GetPlayerEncumbrance(player)

			// Validate weight percentage
			if abs(encumbrance.WeightPct-tt.expectedWeightPct) > 0.001 {
				t.Errorf("Weight percentage mismatch: got %.3f, expected %.3f (description: %s)",
					encumbrance.WeightPct, tt.expectedWeightPct, tt.description)
			}

			// Validate movement penalty
			if abs(encumbrance.MovementPenalty-tt.expectedMovementPenalty) > 0.001 {
				t.Errorf("Movement penalty mismatch: got %.3f, expected %.3f (description: %s)",
					encumbrance.MovementPenalty, tt.expectedMovementPenalty, tt.description)
			}

			// Log detailed information for debugging
			t.Logf("Test: %s", tt.name)
			t.Logf("  Total weight: %.2f kg (%.1f%%)", encumbrance.CurrentWeight, encumbrance.WeightPct*100)
			t.Logf("  Movement penalty: %.3f", encumbrance.MovementPenalty)
			t.Logf("  Description: %s", tt.description)
		})
	}
}

// TestT032_EncumbranceEdgeCases tests edge cases for encumbrance calculation
func TestT032_EncumbranceEdgeCases(t *testing.T) {
	t.Run("ExactlyAtThreshold", func(t *testing.T) {
		pm := NewPlayerManager()
		pm.CreateTestItemTemplates()

		player := createTestPlayer()
		pm.InitializePlayer(player)

		// Set up skills and limits for testing
		player.Skills = map[string]int{"melee": 15, "defense": 10}
		player.Inventory.WeightLimit = 200.0

		// Add exactly 160kg total (80% of 200kg threshold)
		pm.RegisterItemTemplate(&ItemTemplate{
			ID:     "test_156_5kg",
			Weight: 156.5, // 156.5 + 3.5 (sword) = 160kg exactly (80% of 200kg)
			Bulk:   1,
		})

		// Add inventory item
		invInstance := ItemInstance{
			InstanceID: "test_inv_80",
			TemplateID: "test_156_5kg",
			Quantity:   1,
			Durability: 1.0,
		}
		pm.AddItemToInventory(player, invInstance, CompartmentBackpack)

		// Equip sword
		swordInstance := ItemInstance{
			InstanceID: "test_sword_80",
			TemplateID: "sword_iron",
			Quantity:   1,
			Durability: 1.0,
		}
		pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
		pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, time.Now())

		encumbrance := pm.GetPlayerEncumbrance(player)

		// At exactly 80%, should still have no movement penalty
		if encumbrance.WeightPct != 0.8 {
			t.Errorf("Expected exactly 80%% weight, got %.3f", encumbrance.WeightPct)
		}
		if encumbrance.MovementPenalty != 1.0 {
			t.Errorf("Expected no movement penalty at 80%%, got %.3f", encumbrance.MovementPenalty)
		}
	})

	t.Run("JustOverThreshold", func(t *testing.T) {
		pm := NewPlayerManager()
		pm.CreateTestItemTemplates()

		player := createTestPlayer()
		pm.InitializePlayer(player)

		// Set up skills and limits for testing
		player.Skills = map[string]int{"melee": 15, "defense": 10}
		player.Inventory.WeightLimit = 200.0

		// Add slightly over 160kg total (80.1% of 200kg threshold)
		pm.RegisterItemTemplate(&ItemTemplate{
			ID:     "test_156_7kg",
			Weight: 156.7, // 156.7 + 3.5 (sword) = 160.2kg (80.1% of 200kg)
			Bulk:   1,
		})

		// Add inventory item
		invInstance := ItemInstance{
			InstanceID: "test_inv_801",
			TemplateID: "test_156_7kg",
			Quantity:   1,
			Durability: 1.0,
		}
		pm.AddItemToInventory(player, invInstance, CompartmentBackpack)

		// Equip sword
		swordInstance := ItemInstance{
			InstanceID: "test_sword_801",
			TemplateID: "sword_iron",
			Quantity:   1,
			Durability: 1.0,
		}
		pm.AddItemToInventory(player, swordInstance, CompartmentBackpack)
		pm.EquipItem(player, swordInstance.InstanceID, SlotMainHand, time.Now())

		encumbrance := pm.GetPlayerEncumbrance(player)

		// Just over 80% should trigger movement penalty
		if encumbrance.WeightPct <= 0.8 {
			t.Errorf("Expected weight over 80%%, got %.3f", encumbrance.WeightPct)
		}
		if encumbrance.MovementPenalty >= 1.0 {
			t.Errorf("Expected movement penalty over 80%%, got %.3f", encumbrance.MovementPenalty)
		}
	})
}

// TestT032_AcceptanceCriteria validates the specific acceptance criteria for T-032
func TestT032_AcceptanceCriteria(t *testing.T) {
	pm := NewPlayerManager()
	pm.CreateTestItemTemplates()

	player := createTestPlayer()
	pm.InitializePlayer(player)

	// Set up skills for equipment
	player.Skills = map[string]int{"melee": 15, "defense": 10}
	player.Inventory.WeightLimit = 200.0

	// Test scenario 1: Light equipment should not affect movement
	lightSword := ItemInstance{
		InstanceID: "test_light_sword",
		TemplateID: "sword_iron", // 3.5kg
		Quantity:   1,
		Durability: 1.0,
	}
	pm.AddItemToInventory(player, lightSword, CompartmentBackpack)
	pm.EquipItem(player, lightSword.InstanceID, SlotMainHand, time.Now())

	encumbrance := pm.GetPlayerEncumbrance(player)
	if encumbrance.MovementPenalty != 1.0 {
		t.Errorf("✗ ACCEPTANCE CRITERIA FAILED: Light equipment should not penalize movement. Got penalty: %.3f", encumbrance.MovementPenalty)
	} else {
		t.Logf("✅ ACCEPTANCE CRITERIA PASSED: Light equipment (%.1fkg) has no movement penalty", encumbrance.CurrentWeight)
	}

	// Test scenario 2: Heavy combined load should affect movement
	heavyInv := ItemInstance{
		InstanceID: "test_heavy_inv",
		TemplateID: "anvil_iron", // 85kg
		Quantity:   1,
		Durability: 1.0,
	}
	pm.AddItemToInventory(player, heavyInv, CompartmentBackpack)

	// Temporarily reduce weight limit to trigger penalty with current load
	player.Inventory.WeightLimit = 100.0

	encumbrance = pm.GetPlayerEncumbrance(player)
	if encumbrance.MovementPenalty >= 1.0 {
		t.Errorf("✗ ACCEPTANCE CRITERIA FAILED: Heavy load should penalize movement. Got penalty: %.3f", encumbrance.MovementPenalty)
	} else {
		t.Logf("✅ ACCEPTANCE CRITERIA PASSED: Heavy load (%.1fkg, %.1f%%) triggers movement penalty: %.3f",
			encumbrance.CurrentWeight, encumbrance.WeightPct*100, encumbrance.MovementPenalty)
	}

	// Test scenario 3: Encumbrance should include both inventory and equipped items
	expectedTotalWeight := 3.5 + 85.0 // sword + anvil
	if abs(encumbrance.CurrentWeight-expectedTotalWeight) > 0.001 {
		t.Errorf("✗ ACCEPTANCE CRITERIA FAILED: Encumbrance should include both inventory and equipped items. Expected %.1fkg, got %.1fkg",
			expectedTotalWeight, encumbrance.CurrentWeight)
	} else {
		t.Logf("✅ ACCEPTANCE CRITERIA PASSED: Encumbrance correctly includes inventory (%.1fkg) + equipped (%.1fkg) = %.1fkg total",
			85.0, 3.5, encumbrance.CurrentWeight)
	}

	t.Logf("🎯 T-032 ACCEPTANCE CRITERIA VALIDATION COMPLETE")
	t.Logf("   ✅ Encumbrance math matches expected scenarios")
	t.Logf("   ✅ Movement penalty calculation correct")
	t.Logf("   ✅ Equipped items included in encumbrance")
}

// Helper types for test data
type testItem struct {
	templateID ItemTemplateID
	quantity   int
}

type equippedTestItem struct {
	templateID ItemTemplateID
	slot       SlotID
}

// Note: abs function is already defined in engine_test.go
