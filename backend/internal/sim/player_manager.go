package sim

import (
	"fmt"
	"time"
)

// EquipCooldown is the default cooldown duration for equipment changes
const EquipCooldown = 2 * time.Second

// PlayerManager handles inventory and equipment operations for players
type PlayerManager struct {
	itemTemplates map[ItemTemplateID]*ItemTemplate
}

// NewPlayerManager creates a new player manager with item templates
func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		itemTemplates: make(map[ItemTemplateID]*ItemTemplate),
	}
}

// RegisterItemTemplate adds an item template to the manager
func (pm *PlayerManager) RegisterItemTemplate(template *ItemTemplate) {
	pm.itemTemplates[template.ID] = template
}

// GetItemTemplate returns an item template by ID
func (pm *PlayerManager) GetItemTemplate(id ItemTemplateID) (*ItemTemplate, bool) {
	template, exists := pm.itemTemplates[id]
	return template, exists
}

// GetAllItemTemplates returns a copy of all registered item templates
func (pm *PlayerManager) GetAllItemTemplates() map[ItemTemplateID]*ItemTemplate {
	copied := make(map[ItemTemplateID]*ItemTemplate, len(pm.itemTemplates))
	for k, v := range pm.itemTemplates {
		copied[k] = v
	}
	return copied
}

// InitializePlayer sets up a new player with default inventory and equipment
func (pm *PlayerManager) InitializePlayer(player *Player) {
	if player.Inventory == nil {
		player.Inventory = NewInventory()
	}
	player.Inventory.SetTemplateCatalog(pm.itemTemplates)
	if player.Equipment == nil {
		player.Equipment = NewEquipment()
	}
	if player.Skills == nil {
		player.Skills = make(map[string]int)
	}
}

// CheckSkillRequirements verifies if a player meets the skill requirements for an item
func (pm *PlayerManager) CheckSkillRequirements(player *Player, template *ItemTemplate) bool {
	for skill, requiredLevel := range template.SkillReq {
		playerLevel, exists := player.Skills[skill]
		if !exists || playerLevel < requiredLevel {
			return false
		}
	}
	return true
}

// AddItemToInventory adds an item to a player's inventory
func (pm *PlayerManager) AddItemToInventory(player *Player, instance ItemInstance, compartment CompartmentType) error {
	template, exists := pm.GetItemTemplate(instance.TemplateID)
	if !exists {
		return fmt.Errorf("unknown item template: %s", instance.TemplateID)
	}

	err := player.Inventory.AddItem(instance, compartment, template)
	if err == nil {
		player.InventoryVersion++
	}
	return err
}

// RemoveItemFromInventory removes an item from a player's inventory
func (pm *PlayerManager) RemoveItemFromInventory(player *Player, instanceID ItemInstanceID) error {
	err := player.Inventory.RemoveItem(instanceID)
	if err == nil {
		player.InventoryVersion++
	}
	return err
}

// EquipItem equips an item from inventory to an equipment slot
func (pm *PlayerManager) EquipItem(player *Player, instanceID ItemInstanceID, slot SlotID, now time.Time) error {
	// Find the item in inventory
	idx := player.Inventory.FindItem(instanceID)
	if idx < 0 {
		return ErrItemNotFound
	}

	item := player.Inventory.Items[idx]
	template, exists := pm.GetItemTemplate(item.Instance.TemplateID)
	if !exists {
		return fmt.Errorf("unknown item template: %s", item.Instance.TemplateID)
	}

	// Validate slot compatibility
	if !template.Allows(slot) {
		return ErrIllegalSlot
	}

	// Check skill requirements
	if !pm.CheckSkillRequirements(player, template) {
		return ErrSkillGate
	}

	// Check cooldown
	if player.Equipment.IsSlotOnCooldown(slot, now) {
		return ErrEquipLocked
	}

	// If there's already an item in the slot, move it to inventory first
	if !player.Equipment.IsSlotEmpty(slot) {
		oldItem := player.Equipment.GetSlot(slot)
		if oldItem != nil {
			oldTemplate, exists := pm.GetItemTemplate(oldItem.Instance.TemplateID)
			if !exists {
				return fmt.Errorf("unknown item template: %s", oldItem.Instance.TemplateID)
			}
			// Add old item back to inventory (use same compartment as current item)
			err := player.Inventory.AddItem(oldItem.Instance, item.Compartment, oldTemplate)
			if err != nil {
				return fmt.Errorf("cannot unequip old item: %w", err)
			}
		}
	}

	// Remove item from inventory
	if err := player.Inventory.RemoveItem(instanceID); err != nil {
		return err
	}

	// Equip the new item
	player.Equipment.SetSlot(slot, item.Instance, EquipCooldown, now)

	// Update both inventory and equipment versions
	player.InventoryVersion++
	player.EquipmentVersion++

	return nil
}

// UnequipItem removes an item from an equipment slot and puts it in inventory
func (pm *PlayerManager) UnequipItem(player *Player, slot SlotID, compartment CompartmentType, now time.Time) error {
	// Check if slot has an item
	if player.Equipment.IsSlotEmpty(slot) {
		return fmt.Errorf("slot %s is empty", slot)
	}

	// Check cooldown
	if player.Equipment.IsSlotOnCooldown(slot, now) {
		return ErrEquipLocked
	}

	equippedItem := player.Equipment.GetSlot(slot)
	template, exists := pm.GetItemTemplate(equippedItem.Instance.TemplateID)
	if !exists {
		return fmt.Errorf("unknown item template: %s", equippedItem.Instance.TemplateID)
	}

	// Try to add item back to inventory
	if err := player.Inventory.AddItem(equippedItem.Instance, compartment, template); err != nil {
		return fmt.Errorf("cannot add item to inventory: %w", err)
	}

	// Clear the equipment slot
	player.Equipment.ClearSlot(slot)

	// Update both inventory and equipment versions
	player.InventoryVersion++
	player.EquipmentVersion++

	return nil
}

// GetPlayerEncumbrance calculates the player's current encumbrance state
func (pm *PlayerManager) GetPlayerEncumbrance(player *Player) EncumbranceState {
	// Start with inventory encumbrance
	encumbrance := player.Inventory.ComputeEncumbrance(pm.itemTemplates)

	// Add weight from equipped items
	equippedWeight := 0.0
	for _, equippedItem := range player.Equipment.Slots {
		if equippedItem != nil {
			if template, exists := pm.itemTemplates[equippedItem.Instance.TemplateID]; exists {
				equippedWeight += template.Weight * float64(equippedItem.Instance.Quantity)
			}
		}
	}

	// Update encumbrance with equipped weight
	totalWeight := encumbrance.CurrentWeight + equippedWeight
	weightPct := totalWeight / encumbrance.MaxWeight

	// Recalculate movement penalty with total weight
	var movementPenalty float64 = 1.0
	if weightPct > 0.8 {
		if weightPct <= 1.0 {
			// Linear penalty from 100% to 50% speed
			movementPenalty = 1.0 - 0.5*(weightPct-0.8)/0.2
		} else {
			// Severe penalty beyond 100%
			movementPenalty = 0.5 * (1.0 / weightPct)
		}
	}

	encumbrance.CurrentWeight = totalWeight
	encumbrance.WeightPct = weightPct
	encumbrance.MovementPenalty = movementPenalty

	return encumbrance
}

// GetEquippedItemStats returns combined stats from all equipped items
func (pm *PlayerManager) GetEquippedItemStats(player *Player) map[string]interface{} {
	stats := make(map[string]interface{})

	totalWeight := 0.0
	damageTypes := make(map[DamageType]bool)

	for _, equippedItem := range player.Equipment.Slots {
		if equippedItem != nil {
			template, exists := pm.GetItemTemplate(equippedItem.Instance.TemplateID)
			if exists {
				totalWeight += template.Weight
				damageTypes[template.DamageType] = true
			}
		}
	}

	stats["equipped_weight"] = totalWeight
	stats["damage_types"] = damageTypes

	return stats
}

// CreateTestItemTemplates creates some basic item templates for testing
func (pm *PlayerManager) CreateTestItemTemplates() {
	// Simple sword
	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "sword_iron",
		DisplayName: "Iron Sword",
		SlotMask:    SlotMaskMainHand,
		Weight:      3.5,
		Bulk:        2,
		DamageType:  DamageSlash,
		SkillReq:    map[string]int{"melee": 10},
	})

	// Shield
	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "shield_wood",
		DisplayName: "Wooden Shield",
		SlotMask:    SlotMaskOffHand,
		Weight:      2.0,
		Bulk:        3,
		DamageType:  DamageBlunt,
		SkillReq:    map[string]int{"defense": 5},
	})

	// Armor
	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "armor_leather",
		DisplayName: "Leather Armor",
		SlotMask:    SlotMaskChest,
		Weight:      5.0,
		Bulk:        4,
		DamageType:  "",
		SkillReq:    map[string]int{},
	})

	// Consumable item
	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "potion_health",
		DisplayName: "Health Potion",
		SlotMask:    0, // Cannot be equipped
		Weight:      0.1,
		Bulk:        1,
		DamageType:  "",
		SkillReq:    map[string]int{},
	})

	// Heavy test items for encumbrance testing
	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "anvil_iron",
		DisplayName: "Iron Anvil",
		SlotMask:    0,    // Cannot be equipped
		Weight:      85.0, // Heavy for encumbrance testing
		Bulk:        25,
		DamageType:  "",
		SkillReq:    map[string]int{},
	})

	pm.RegisterItemTemplate(&ItemTemplate{
		ID:          "rock_small",
		DisplayName: "Small Rock",
		SlotMask:    0,
		Weight:      0.5,
		Bulk:        1,
		DamageType:  "",
		SkillReq:    map[string]int{},
	})
}
