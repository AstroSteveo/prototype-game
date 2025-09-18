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

// InitializePlayer sets up a new player with default inventory and equipment
func (pm *PlayerManager) InitializePlayer(player *Player) {
	player.Inventory = NewInventory()
	player.Equipment = NewEquipment()
	player.Skills = make(map[string]int)
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

	return player.Inventory.AddItem(instance, compartment, template)
}

// RemoveItemFromInventory removes an item from a player's inventory
func (pm *PlayerManager) RemoveItemFromInventory(player *Player, instanceID ItemInstanceID) error {
	return player.Inventory.RemoveItem(instanceID)
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
			// Add old item back to inventory (use same compartment as current item)
			err := player.Inventory.AddItem(oldItem.Instance, item.Compartment, template)
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

	return nil
}

// GetPlayerEncumbrance calculates the player's current encumbrance state
func (pm *PlayerManager) GetPlayerEncumbrance(player *Player) EncumbranceState {
	return player.Inventory.ComputeEncumbrance(pm.itemTemplates)
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
}
