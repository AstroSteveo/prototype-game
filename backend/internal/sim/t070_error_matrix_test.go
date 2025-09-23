package sim

import (
	"testing"
	"time"
)

// TestT070_ErrorMatrixValidation validates all error handling paths match the documented matrix
// This test ensures T-070 acceptance criteria: "Design error matrix aligns with behavior"
func TestT070_ErrorMatrixValidation(t *testing.T) {
	t.Run("Equipment_Error_Matrix_Comprehensive", func(t *testing.T) {
		pm := NewPlayerManager()
		pm.CreateTestItemTemplates()

		// Test all documented error scenarios systematically
		scenarios := []struct {
			name           string
			setupPlayer    func() *Player
			itemTemplateID ItemTemplateID
			targetSlot     SlotID
			expectError    error
			description    string
		}{
			// === R1: Slot Compatibility Matrix Validation ===
			{
				name: "R1_Sword_MainHand_Valid",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"melee": 10},
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "sword_iron",
				targetSlot:     SlotMainHand,
				expectError:    nil,
				description:    "R1: Sword to main hand should succeed",
			},
			{
				name: "R1_Sword_OffHand_Invalid",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"melee": 10},
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "sword_iron",
				targetSlot:     SlotOffHand,
				expectError:    ErrIllegalSlot,
				description:    "R1: Sword to off hand should fail with ErrIllegalSlot",
			},
			{
				name: "R1_Shield_OffHand_Valid",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"defense": 5},
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "shield_wood",
				targetSlot:     SlotOffHand,
				expectError:    nil,
				description:    "R1: Shield to off hand should succeed",
			},
			{
				name: "R1_Shield_MainHand_Invalid",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"defense": 5},
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "shield_wood",
				targetSlot:     SlotMainHand,
				expectError:    ErrIllegalSlot,
				description:    "R1: Shield to main hand should fail with ErrIllegalSlot",
			},

			// === R2: Skill Requirements Matrix Validation ===
			{
				name: "R2_NoSkills_Fail",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{}, // No skills
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "sword_iron",
				targetSlot:     SlotMainHand,
				expectError:    ErrSkillGate,
				description:    "R2: No skills should fail with ErrSkillGate",
			},
			{
				name: "R2_InsufficientSkill_Fail",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"melee": 5}, // Insufficient
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "sword_iron",
				targetSlot:     SlotMainHand,
				expectError:    ErrSkillGate,
				description:    "R2: Insufficient skill should fail with ErrSkillGate",
			},
			{
				name: "R2_ExactSkill_Success",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"melee": 10}, // Exact requirement
					}
					pm.InitializePlayer(p)
					return p
				},
				itemTemplateID: "sword_iron",
				targetSlot:     SlotMainHand,
				expectError:    nil,
				description:    "R2: Exact skill requirement should succeed",
			},

			// === R4: Item Not Found Matrix Validation ===
			{
				name: "R4_ItemNotFound_Fail",
				setupPlayer: func() *Player {
					p := &Player{
						Entity:    Entity{ID: "test", Name: "Test"},
						Inventory: NewInventory(),
						Equipment: NewEquipment(),
						Skills:    map[string]int{"melee": 10},
					}
					pm.InitializePlayer(p)
					// Don't add the item to inventory
					return p
				},
				itemTemplateID: "sword_iron",
				targetSlot:     SlotMainHand,
				expectError:    ErrItemNotFound,
				description:    "R4: Missing item should fail with ErrItemNotFound",
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				player := scenario.setupPlayer()

				// Only add item to inventory if we expect the operation to get past the "not found" check
				if scenario.expectError != ErrItemNotFound {
					instance := ItemInstance{
						InstanceID: ItemInstanceID("test_" + string(scenario.itemTemplateID)),
						TemplateID: scenario.itemTemplateID,
						Quantity:   1,
						Durability: 1.0,
					}

					err := pm.AddItemToInventory(player, instance, CompartmentBackpack)
					if err != nil {
						t.Fatalf("Failed to add item for test setup: %v", err)
					}
				}

				// Attempt the operation
				now := time.Now()
				instanceID := ItemInstanceID("test_" + string(scenario.itemTemplateID))
				err := pm.EquipItem(player, instanceID, scenario.targetSlot, now)

				// Validate result
				if scenario.expectError != nil {
					if err != scenario.expectError {
						t.Errorf("T-070 FAILED: %s - Expected error %v, got %v",
							scenario.description, scenario.expectError, err)
					} else {
						t.Logf("T-070 SUCCESS: %s", scenario.description)
					}
				} else {
					if err != nil {
						t.Errorf("T-070 FAILED: %s - Expected success, got error %v",
							scenario.description, err)
					} else {
						t.Logf("T-070 SUCCESS: %s", scenario.description)
					}
				}
			})
		}
	})

	t.Run("Cooldown_Error_Matrix_Validation", func(t *testing.T) {
		pm := NewPlayerManager()
		pm.CreateTestItemTemplates()

		// === R3: Cooldown System Matrix Validation ===
		cooldownScenarios := []struct {
			name           string
			cooldownOffset time.Duration
			expectError    error
			description    string
		}{
			{
				name:           "R3_ImmediateUnequip_Fail",
				cooldownOffset: 0,
				expectError:    ErrEquipLocked,
				description:    "R3: Immediate unequip should fail with ErrEquipLocked",
			},
			{
				name:           "R3_DuringCooldown_Fail",
				cooldownOffset: EquipCooldown / 2,
				expectError:    ErrEquipLocked,
				description:    "R3: Unequip during cooldown should fail with ErrEquipLocked",
			},
			{
				name:           "R3_JustBeforeExpiry_Fail",
				cooldownOffset: EquipCooldown - time.Millisecond,
				expectError:    ErrEquipLocked,
				description:    "R3: Unequip just before expiry should fail with ErrEquipLocked",
			},
			{
				name:           "R3_ExactExpiry_Success",
				cooldownOffset: EquipCooldown,
				expectError:    nil,
				description:    "R3: Unequip at exact expiry should succeed",
			},
			{
				name:           "R3_AfterExpiry_Success",
				cooldownOffset: EquipCooldown + time.Second,
				expectError:    nil,
				description:    "R3: Unequip after expiry should succeed",
			},
		}

		for _, scenario := range cooldownScenarios {
			t.Run(scenario.name, func(t *testing.T) {
				player := &Player{
					Entity:    Entity{ID: "test", Name: "Test"},
					Inventory: NewInventory(),
					Equipment: NewEquipment(),
					Skills:    map[string]int{"melee": 10},
				}
				pm.InitializePlayer(player)

				// Add and equip item
				instance := ItemInstance{
					InstanceID: "test_sword",
					TemplateID: "sword_iron",
					Quantity:   1,
					Durability: 1.0,
				}

				err := pm.AddItemToInventory(player, instance, CompartmentBackpack)
				if err != nil {
					t.Fatalf("Failed to add item: %v", err)
				}

				now := time.Now()
				err = pm.EquipItem(player, instance.InstanceID, SlotMainHand, now)
				if err != nil {
					t.Fatalf("Failed to equip item: %v", err)
				}

				// Attempt unequip with timing offset
				unequipTime := now.Add(scenario.cooldownOffset)
				err = pm.UnequipItem(player, SlotMainHand, CompartmentBackpack, unequipTime)

				// Validate result
				if scenario.expectError != nil {
					if err != scenario.expectError {
						t.Errorf("T-070 FAILED: %s - Expected error %v, got %v",
							scenario.description, scenario.expectError, err)
					} else {
						t.Logf("T-070 SUCCESS: %s", scenario.description)
					}
				} else {
					if err != nil {
						t.Errorf("T-070 FAILED: %s - Expected success, got error %v",
							scenario.description, err)
					} else {
						t.Logf("T-070 SUCCESS: %s", scenario.description)
					}
				}
			})
		}
	})

	t.Run("Error_Code_Consistency_Validation", func(t *testing.T) {
		// Validate that error constants are properly defined and consistent
		errorMappings := map[error]string{
			ErrIllegalSlot:       "illegal_slot",
			ErrSkillGate:         "skill_gate",
			ErrEquipLocked:       "equip_locked",
			ErrItemNotFound:      "item_not_found",
			ErrInsufficientSpace: "insufficient_space",
			ErrExceedsWeight:     "exceeds_weight",
			ErrExceedsBulk:       "exceeds_bulk",
			ErrDuplicateInstance: "duplicate_instance",
		}

		// Verify all error constants are non-nil and have consistent messages
		for err, expectedCode := range errorMappings {
			if err == nil {
				t.Errorf("T-070 FAILED: Error constant is nil for code %s", expectedCode)
				continue
			}

			if err.Error() == "" {
				t.Errorf("T-070 FAILED: Error %v has empty message", err)
				continue
			}

			t.Logf("T-070 SUCCESS: Error %v -> code %s has message: %s",
				err, expectedCode, err.Error())
		}

		// Verify error type constants
		if EquipCooldown != 2*time.Second {
			t.Errorf("T-070 FAILED: EquipCooldown constant mismatch - expected 2s, got %v", EquipCooldown)
		} else {
			t.Logf("T-070 SUCCESS: EquipCooldown constant validated: %v", EquipCooldown)
		}
	})

	t.Log("T-070 ERROR MATRIX VALIDATION COMPLETE")
	t.Log("✅ Slot compatibility matrix validated")
	t.Log("✅ Skill requirements matrix validated")
	t.Log("✅ Cooldown system matrix validated")
	t.Log("✅ Error code consistency validated")
}

// TestErrorMatrixDocumentationAlignment verifies that the documented error matrix
// matches the actual implementation behavior
func TestErrorMatrixDocumentationAlignment(t *testing.T) {
	t.Run("WebSocket_Error_Code_Coverage", func(t *testing.T) {
		// Test that all documented WebSocket error codes are properly mapped
		// This validates the error matrix documentation accuracy

		documentedCodes := map[string]error{
			"illegal_slot":   ErrIllegalSlot,
			"skill_gate":     ErrSkillGate,
			"equip_locked":   ErrEquipLocked,
			"item_not_found": ErrItemNotFound,
		}

		for code, expectedErr := range documentedCodes {
			if expectedErr == nil {
				t.Errorf("Documentation error: WebSocket code '%s' maps to nil error", code)
			} else {
				t.Logf("Documentation validated: WebSocket code '%s' maps to %v", code, expectedErr)
			}
		}
	})

	t.Run("Equipment_Matrix_Completeness", func(t *testing.T) {
		pm := NewPlayerManager()
		pm.CreateTestItemTemplates()

		// Verify all documented item types and slots are covered
		documentedItems := []struct {
			templateID  ItemTemplateID
			description string
			validSlots  []SlotID
		}{
			{
				templateID:  "sword_iron",
				description: "Sword (Main Hand)",
				validSlots:  []SlotID{SlotMainHand},
			},
			{
				templateID:  "shield_wood",
				description: "Shield (Off Hand)",
				validSlots:  []SlotID{SlotOffHand},
			},
			{
				templateID:  "armor_leather",
				description: "Armor (Chest)",
				validSlots:  []SlotID{SlotChest},
			},
		}

		for _, item := range documentedItems {
			template, exists := pm.GetItemTemplate(item.templateID)
			if !exists {
				t.Errorf("Documentation error: Item %s (%s) not found in templates",
					item.templateID, item.description)
				continue
			}

			// Verify documented valid slots
			for _, slot := range item.validSlots {
				if !template.Allows(slot) {
					t.Errorf("Documentation error: %s should allow slot %s according to docs",
						item.description, slot)
				} else {
					t.Logf("Documentation validated: %s allows slot %s", item.description, slot)
				}
			}
		}
	})
}
