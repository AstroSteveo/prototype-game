package sim

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrIllegalSlot       = errors.New("item cannot be equipped to this slot")
	ErrSkillGate         = errors.New("insufficient skill level for item")
	ErrEquipLocked       = errors.New("equipment slot is on cooldown")
	ErrItemNotFound      = errors.New("item not found in inventory")
	ErrInsufficientSpace = errors.New("insufficient space in compartment")
	ErrExceedsWeight     = errors.New("item would exceed weight limit")
	ErrExceedsBulk       = errors.New("item would exceed bulk limit")
	ErrDuplicateInstance = errors.New("item instance already exists")
)

// EncumbranceState represents the player's current encumbrance
type EncumbranceState struct {
	CurrentWeight   float64 `json:"current_weight"`
	MaxWeight       float64 `json:"max_weight"`
	CurrentBulk     int     `json:"current_bulk"`
	MaxBulk         int     `json:"max_bulk"`
	WeightPct       float64 `json:"weight_pct"`       // 0.0 to 1.0+
	BulkPct         float64 `json:"bulk_pct"`         // 0.0 to 1.0+
	MovementPenalty float64 `json:"movement_penalty"` // Speed multiplier 0.0 to 1.0
}

// Equipment represents a player's equipped items
type Equipment struct {
	Slots map[SlotID]*EquippedItem `json:"slots"`
}

// NewEquipment creates a new empty equipment set
func NewEquipment() *Equipment {
	return &Equipment{
		Slots: make(map[SlotID]*EquippedItem),
	}
}

// GetSlot returns the equipped item in the given slot, or nil if empty
func (e *Equipment) GetSlot(slot SlotID) *EquippedItem {
	return e.Slots[slot]
}

// IsSlotEmpty returns true if the slot is empty or cooldown has expired
func (e *Equipment) IsSlotEmpty(slot SlotID) bool {
	item := e.Slots[slot]
	return item == nil || item.Instance.InstanceID == ""
}

// IsSlotOnCooldown returns true if the slot has an active cooldown
func (e *Equipment) IsSlotOnCooldown(slot SlotID, now time.Time) bool {
	item := e.Slots[slot]
	return item != nil && item.CooldownActive(now)
}

// SetSlot equips an item to a slot with cooldown
func (e *Equipment) SetSlot(slot SlotID, instance ItemInstance, cooldown time.Duration, now time.Time) {
	e.Slots[slot] = &EquippedItem{
		Instance:      instance,
		CooldownUntil: now.Add(cooldown),
	}
}

// ClearSlot removes an item from a slot
func (e *Equipment) ClearSlot(slot SlotID) {
	delete(e.Slots, slot)
}

// Inventory represents a player's inventory system
type Inventory struct {
	Items           []InventoryItem         `json:"items"`
	CompartmentCaps map[CompartmentType]int `json:"compartment_caps"` // Bulk limits per compartment
	WeightLimit     float64                 `json:"weight_limit"`
	itemIndex       map[ItemInstanceID]int  // Index for fast lookup
	templateCatalog map[ItemTemplateID]*ItemTemplate
}

// NewInventory creates a new inventory with default capacities
func NewInventory() *Inventory {
	inv := &Inventory{
		Items: make([]InventoryItem, 0),
		CompartmentCaps: map[CompartmentType]int{
			CompartmentBackpack: 50, // Default bulk limits
			CompartmentBelt:     10,
			CompartmentCraftBag: 30,
		},
		WeightLimit: 100.0, // Default weight limit
		itemIndex:   make(map[ItemInstanceID]int),
	}
	return inv
}

// SetTemplateCatalog wires in the authoritative template catalog for lookups.
func (inv *Inventory) SetTemplateCatalog(catalog map[ItemTemplateID]*ItemTemplate) {
	inv.templateCatalog = catalog
	if catalog == nil {
		return
	}
	for i := range inv.Items {
		if inv.Items[i].template == nil {
			if template, ok := catalog[inv.Items[i].Instance.TemplateID]; ok {
				inv.Items[i].template = template
			}
		}
	}
}

func (inv *Inventory) resolveTemplate(item *InventoryItem, templates map[ItemTemplateID]*ItemTemplate) *ItemTemplate {
	if item.template != nil {
		return item.template
	}
	if templates != nil {
		if template, ok := templates[item.Instance.TemplateID]; ok {
			item.template = template
			return template
		}
	}
	if inv.templateCatalog != nil {
		if template, ok := inv.templateCatalog[item.Instance.TemplateID]; ok {
			item.template = template
			return template
		}
	}
	return nil
}

// rebuildIndex rebuilds the item lookup index
func (inv *Inventory) rebuildIndex() {
	inv.itemIndex = make(map[ItemInstanceID]int)
	for i, item := range inv.Items {
		inv.itemIndex[item.Instance.InstanceID] = i
	}
}

// FindItem returns the index of an item by instance ID, or -1 if not found
func (inv *Inventory) FindItem(instanceID ItemInstanceID) int {
	if idx, exists := inv.itemIndex[instanceID]; exists {
		return idx
	}
	return -1
}

// HasItem returns true if the inventory contains the item
func (inv *Inventory) HasItem(instanceID ItemInstanceID) bool {
	return inv.FindItem(instanceID) >= 0
}

// GetCompartmentContents returns all items in a specific compartment
func (inv *Inventory) GetCompartmentContents(compartment CompartmentType) []InventoryItem {
	var items []InventoryItem
	for _, item := range inv.Items {
		if item.Compartment == compartment {
			items = append(items, item)
		}
	}
	return items
}

// GetCompartmentBulk returns the current bulk used in a compartment
func (inv *Inventory) GetCompartmentBulk(compartment CompartmentType, templates map[ItemTemplateID]*ItemTemplate) int {
	bulk := 0
	for i := range inv.Items {
		item := &inv.Items[i]
		if item.Compartment != compartment {
			continue
		}
		if template := inv.resolveTemplate(item, templates); template != nil {
			bulk += template.Bulk * item.Instance.Quantity
		}
	}
	return bulk
}

// GetTotalWeight returns the total weight of all items
func (inv *Inventory) GetTotalWeight(templates map[ItemTemplateID]*ItemTemplate) float64 {
	weight := 0.0
	for i := range inv.Items {
		item := &inv.Items[i]
		if template := inv.resolveTemplate(item, templates); template != nil {
			weight += template.Weight * float64(item.Instance.Quantity)
		}
	}
	return weight
}

// CanAddItem checks if an item can be added to the inventory
func (inv *Inventory) CanAddItem(instance ItemInstance, compartment CompartmentType, template *ItemTemplate) error {
	// Check for duplicate instance ID
	if inv.HasItem(instance.InstanceID) {
		return ErrDuplicateInstance
	}

	if template == nil {
		if inv.templateCatalog != nil {
			template = inv.templateCatalog[instance.TemplateID]
		}
		if template == nil {
			return fmt.Errorf("unknown item template: %s", instance.TemplateID)
		}
	}

	// Check weight limit
	totalWeight := inv.GetTotalWeight(nil)
	newWeight := template.Weight * float64(instance.Quantity)
	if totalWeight+newWeight > inv.WeightLimit {
		return ErrExceedsWeight
	}

	// Check compartment bulk limit
	currentBulk := inv.GetCompartmentBulk(compartment, nil)
	newBulk := template.Bulk * instance.Quantity
	if limit, exists := inv.CompartmentCaps[compartment]; exists {
		if currentBulk+newBulk > limit {
			return ErrExceedsBulk
		}
	}

	return nil
}

// AddItem adds an item to the inventory
func (inv *Inventory) AddItem(instance ItemInstance, compartment CompartmentType, template *ItemTemplate) error {
	if err := inv.CanAddItem(instance, compartment, template); err != nil {
		return err
	}

	invItem := InventoryItem{
		Instance:    instance,
		Compartment: compartment,
		template:    template,
	}

	inv.Items = append(inv.Items, invItem)
	inv.itemIndex[instance.InstanceID] = len(inv.Items) - 1

	return nil
}

// RemoveItem removes an item from the inventory
func (inv *Inventory) RemoveItem(instanceID ItemInstanceID) error {
	idx := inv.FindItem(instanceID)
	if idx < 0 {
		return ErrItemNotFound
	}

	// Remove item by swapping with last and truncating
	lastIdx := len(inv.Items) - 1
	if idx != lastIdx {
		inv.Items[idx] = inv.Items[lastIdx]
	}
	inv.Items = inv.Items[:lastIdx]

	// Rebuild index after removal
	inv.rebuildIndex()

	return nil
}

// ComputeEncumbrance calculates the current encumbrance state
func (inv *Inventory) ComputeEncumbrance(templates map[ItemTemplateID]*ItemTemplate) EncumbranceState {
	totalWeight := inv.GetTotalWeight(templates)
	totalBulk := 0
	maxBulk := 0

	for compartment, limit := range inv.CompartmentCaps {
		totalBulk += inv.GetCompartmentBulk(compartment, templates)
		maxBulk += limit
	}

	weightPct := totalWeight / inv.WeightLimit
	bulkPct := float64(totalBulk) / float64(maxBulk)

	// Calculate movement penalty based on weight
	// 0-80%: no penalty, 80-100%: linear penalty, 100%+: severe penalty
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

	return EncumbranceState{
		CurrentWeight:   totalWeight,
		MaxWeight:       inv.WeightLimit,
		CurrentBulk:     totalBulk,
		MaxBulk:         maxBulk,
		WeightPct:       weightPct,
		BulkPct:         bulkPct,
		MovementPenalty: movementPenalty,
	}
}
