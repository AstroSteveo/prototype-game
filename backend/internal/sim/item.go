package sim

import (
	"time"
)

// ItemTemplateID uniquely identifies an item template
type ItemTemplateID string

// ItemInstanceID uniquely identifies a specific item instance
type ItemInstanceID string

// SlotID identifies equipment slots
type SlotID string

const (
	SlotMainHand SlotID = "main_hand"
	SlotOffHand  SlotID = "off_hand"
	SlotChest    SlotID = "chest"
	SlotLegs     SlotID = "legs"
	SlotFeet     SlotID = "feet"
	SlotHead     SlotID = "head"
)

// DamageType represents weapon damage types
type DamageType string

const (
	DamageSlash     DamageType = "slash"
	DamagePierce    DamageType = "pierce"
	DamageBlunt     DamageType = "blunt"
	DamageElemental DamageType = "elemental"
)

// SlotMask represents which slots an item can be equipped to
type SlotMask uint32

const (
	SlotMaskMainHand SlotMask = 1 << iota
	SlotMaskOffHand
	SlotMaskChest
	SlotMaskLegs
	SlotMaskFeet
	SlotMaskHead
)

// ItemTemplate defines the static properties of an item type
type ItemTemplate struct {
	ID          ItemTemplateID `json:"id"`
	DisplayName string         `json:"display_name"`
	SlotMask    SlotMask       `json:"slot_mask"`
	Weight      float64        `json:"weight"`      // For encumbrance calculation
	Bulk        int            `json:"bulk"`        // Inventory space used
	DamageType  DamageType     `json:"damage_type"` // For combat resolution
	SkillReq    map[string]int `json:"skill_req"`   // Skill requirements to equip
}

// Allows checks if this item can be equipped to the given slot
func (t *ItemTemplate) Allows(slot SlotID) bool {
	var mask SlotMask
	switch slot {
	case SlotMainHand:
		mask = SlotMaskMainHand
	case SlotOffHand:
		mask = SlotMaskOffHand
	case SlotChest:
		mask = SlotMaskChest
	case SlotLegs:
		mask = SlotMaskLegs
	case SlotFeet:
		mask = SlotMaskFeet
	case SlotHead:
		mask = SlotMaskHead
	default:
		return false
	}
	return t.SlotMask&mask != 0
}

// ItemInstance represents a specific instance of an item
type ItemInstance struct {
	InstanceID ItemInstanceID `json:"instance_id"`
	TemplateID ItemTemplateID `json:"template_id"`
	Quantity   int            `json:"quantity"`
	Durability float64        `json:"durability"` // 0.0 to 1.0
}

// EquippedItem represents an item in an equipment slot
type EquippedItem struct {
	Instance      ItemInstance `json:"instance"`
	CooldownUntil time.Time    `json:"cooldown_until"`
}

// CooldownActive returns true if the item is still on cooldown
func (e *EquippedItem) CooldownActive(now time.Time) bool {
	return now.Before(e.CooldownUntil)
}

// CompartmentType represents different inventory compartments
type CompartmentType string

const (
	CompartmentBackpack CompartmentType = "backpack"
	CompartmentBelt     CompartmentType = "belt"
	CompartmentCraftBag CompartmentType = "craft_bag"
)

// InventoryItem represents an item in inventory
type InventoryItem struct {
	Instance    ItemInstance    `json:"instance"`
	Compartment CompartmentType `json:"compartment"`
	template    *ItemTemplate
}
