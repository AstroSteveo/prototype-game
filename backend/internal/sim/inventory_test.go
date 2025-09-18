package sim

import (
	"testing"
	"time"
)

func TestItemTemplate_Allows(t *testing.T) {
	template := &ItemTemplate{
		ID:       "test_weapon",
		SlotMask: SlotMaskMainHand | SlotMaskOffHand,
	}

	tests := []struct {
		slot     SlotID
		expected bool
	}{
		{SlotMainHand, true},
		{SlotOffHand, true},
		{SlotChest, false},
		{SlotLegs, false},
	}

	for _, test := range tests {
		result := template.Allows(test.slot)
		if result != test.expected {
			t.Errorf("Allows(%s) = %v, expected %v", test.slot, result, test.expected)
		}
	}
}

func TestEquippedItem_CooldownActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		cooldownUntil time.Time
		checkTime     time.Time
		expected      bool
	}{
		{"active cooldown", now.Add(10 * time.Second), now, true},
		{"expired cooldown", now.Add(-10 * time.Second), now, false},
		{"exactly expired", now, now, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := &EquippedItem{CooldownUntil: test.cooldownUntil}
			result := item.CooldownActive(test.checkTime)
			if result != test.expected {
				t.Errorf("CooldownActive() = %v, expected %v", result, test.expected)
			}
		})
	}
}

func TestNewInventory(t *testing.T) {
	inv := NewInventory()

	if inv == nil {
		t.Fatal("NewInventory() returned nil")
	}

	if len(inv.Items) != 0 {
		t.Errorf("New inventory should have 0 items, got %d", len(inv.Items))
	}

	if inv.WeightLimit <= 0 {
		t.Errorf("Weight limit should be positive, got %f", inv.WeightLimit)
	}

	expectedCompartments := []CompartmentType{
		CompartmentBackpack,
		CompartmentBelt,
		CompartmentCraftBag,
	}

	for _, comp := range expectedCompartments {
		if cap, exists := inv.CompartmentCaps[comp]; !exists || cap <= 0 {
			t.Errorf("Compartment %s should have positive capacity, got %d", comp, cap)
		}
	}
}

func TestInventory_AddItem(t *testing.T) {
	inv := NewInventory()
	template := &ItemTemplate{
		ID:     "test_item",
		Weight: 5.0,
		Bulk:   2,
	}

	instance := ItemInstance{
		InstanceID: "item1",
		TemplateID: template.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	// Test successful add
	err := inv.AddItem(instance, CompartmentBackpack, template)
	if err != nil {
		t.Fatalf("AddItem() failed: %v", err)
	}

	if len(inv.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(inv.Items))
	}

	if !inv.HasItem(instance.InstanceID) {
		t.Error("Item should be found in inventory")
	}

	// Test duplicate instance ID
	err = inv.AddItem(instance, CompartmentBackpack, template)
	if err != ErrDuplicateInstance {
		t.Errorf("Expected ErrDuplicateInstance, got %v", err)
	}
}

func TestInventory_RemoveItem(t *testing.T) {
	inv := NewInventory()
	template := &ItemTemplate{
		ID:     "test_item",
		Weight: 5.0,
		Bulk:   2,
	}

	instance := ItemInstance{
		InstanceID: "item1",
		TemplateID: template.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	// Test removing non-existent item
	err := inv.RemoveItem(instance.InstanceID)
	if err != ErrItemNotFound {
		t.Errorf("Expected ErrItemNotFound, got %v", err)
	}

	// Add item then remove it
	inv.AddItem(instance, CompartmentBackpack, template)
	err = inv.RemoveItem(instance.InstanceID)
	if err != nil {
		t.Fatalf("RemoveItem() failed: %v", err)
	}

	if len(inv.Items) != 0 {
		t.Errorf("Expected 0 items after removal, got %d", len(inv.Items))
	}

	if inv.HasItem(instance.InstanceID) {
		t.Error("Item should not be found after removal")
	}
}

func TestInventory_WeightLimits(t *testing.T) {
	inv := NewInventory()
	inv.WeightLimit = 10.0 // Small limit for testing

	heavyTemplate := &ItemTemplate{
		ID:     "heavy_item",
		Weight: 8.0,
		Bulk:   1,
	}

	lightTemplate := &ItemTemplate{
		ID:     "light_item",
		Weight: 1.0,
		Bulk:   1,
	}

	// Add heavy item first (should succeed)
	heavyInstance := ItemInstance{
		InstanceID: "heavy1",
		TemplateID: heavyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err := inv.AddItem(heavyInstance, CompartmentBackpack, heavyTemplate)
	if err != nil {
		t.Fatalf("Adding heavy item failed: %v", err)
	}

	// Try to add another heavy item (should fail)
	heavyInstance2 := ItemInstance{
		InstanceID: "heavy2",
		TemplateID: heavyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err = inv.AddItem(heavyInstance2, CompartmentBackpack, heavyTemplate)
	if err != ErrExceedsWeight {
		t.Errorf("Expected ErrExceedsWeight, got %v", err)
	}

	// Light item should still fit
	lightInstance := ItemInstance{
		InstanceID: "light1",
		TemplateID: lightTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err = inv.AddItem(lightInstance, CompartmentBackpack, lightTemplate)
	if err != nil {
		t.Fatalf("Adding light item failed: %v", err)
	}
}

func TestInventory_BulkLimits(t *testing.T) {
	inv := NewInventory()
	inv.CompartmentCaps[CompartmentBelt] = 3 // Small limit for testing

	bulkyTemplate := &ItemTemplate{
		ID:     "bulky_item",
		Weight: 1.0,
		Bulk:   2,
	}

	// Add first bulky item (should succeed)
	instance1 := ItemInstance{
		InstanceID: "bulky1",
		TemplateID: bulkyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err := inv.AddItem(instance1, CompartmentBelt, bulkyTemplate)
	if err != nil {
		t.Fatalf("Adding first bulky item failed: %v", err)
	}

	// Try to add second bulky item (should fail - 2+2 > 3)
	instance2 := ItemInstance{
		InstanceID: "bulky2",
		TemplateID: bulkyTemplate.ID,
		Quantity:   1,
		Durability: 1.0,
	}

	err = inv.AddItem(instance2, CompartmentBelt, bulkyTemplate)
	if err != ErrExceedsBulk {
		t.Errorf("Expected ErrExceedsBulk, got %v", err)
	}
}

func TestInventory_ComputeEncumbrance(t *testing.T) {
	inv := NewInventory()
	inv.WeightLimit = 100.0
	inv.CompartmentCaps[CompartmentBackpack] = 10

	templates := map[ItemTemplateID]*ItemTemplate{
		"light_item": {
			ID:     "light_item",
			Weight: 20.0, // 20% of weight limit
			Bulk:   2,    // 20% of backpack limit
		},
		"heavy_item": {
			ID:     "heavy_item",
			Weight: 70.0, // 70% of weight limit
			Bulk:   3,    // 30% of backpack limit
		},
	}

	// Test with light load (no movement penalty)
	lightInstance := ItemInstance{
		InstanceID: "light1",
		TemplateID: "light_item",
		Quantity:   1,
		Durability: 1.0,
	}

	inv.AddItem(lightInstance, CompartmentBackpack, templates["light_item"])

	encumbrance := inv.ComputeEncumbrance(templates)
	if encumbrance.WeightPct != 0.2 {
		t.Errorf("Expected weight percentage 0.2, got %f", encumbrance.WeightPct)
	}
	if encumbrance.MovementPenalty != 1.0 {
		t.Errorf("Expected no movement penalty, got %f", encumbrance.MovementPenalty)
	}

	// Add heavy item (should trigger movement penalty)
	heavyInstance := ItemInstance{
		InstanceID: "heavy1",
		TemplateID: "heavy_item",
		Quantity:   1,
		Durability: 1.0,
	}

	inv.AddItem(heavyInstance, CompartmentBackpack, templates["heavy_item"])

	encumbrance = inv.ComputeEncumbrance(templates)
	if encumbrance.WeightPct != 0.9 {
		t.Errorf("Expected weight percentage 0.9, got %f", encumbrance.WeightPct)
	}
	if encumbrance.MovementPenalty >= 1.0 {
		t.Errorf("Expected movement penalty < 1.0, got %f", encumbrance.MovementPenalty)
	}
}
