package state

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

func TestPostgresStore_Integration(t *testing.T) {
	// Skip if PostgreSQL is not available (would need live DB)
	// This is a design test showing the expected interface
	t.Skip("Integration test requires PostgreSQL connection")

	dsn := "postgres://localhost/test_db?sslmode=disable"
	store, err := NewPostgresStore(dsn)
	if err != nil {
		t.Fatalf("Failed to create PostgreSQL store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	playerID := "test-player-123"

	// Test initial save
	initialState := PlayerState{
		Pos:     spatial.Vec2{X: 100, Z: 200},
		Logins:  1,
		Updated: time.Now(),
		Version: 1,
	}

	err = store.Save(ctx, playerID, initialState)
	if err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Test load
	loadedState, exists, err := store.Load(ctx, playerID)
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if !exists {
		t.Fatal("Player state should exist")
	}

	if loadedState.Pos.X != initialState.Pos.X || loadedState.Pos.Z != initialState.Pos.Z {
		t.Errorf("Position mismatch: got %v, want %v", loadedState.Pos, initialState.Pos)
	}

	// Test optimistic locking
	updatedState1 := loadedState
	updatedState1.Pos.X = 150
	updatedState1.Updated = time.Now()

	updatedState2 := loadedState
	updatedState2.Pos.Z = 250
	updatedState2.Updated = time.Now()

	// First update should succeed
	err = store.Save(ctx, playerID, updatedState1)
	if err != nil {
		t.Fatalf("First update should succeed: %v", err)
	}

	// Second update should fail due to optimistic locking
	err = store.Save(ctx, playerID, updatedState2)
	if err != ErrOptimisticLock {
		t.Errorf("Second update should fail with optimistic lock error, got: %v", err)
	}
}

func TestInventoryPersistence_Integration(t *testing.T) {
	store := NewMemStore()
	ctx := context.Background()
	playerID := "test-player-456"

	// This test demonstrates the expected flow for inventory persistence
	// In a real integration test, this would test with the sim engine

	testData := `{
		"items": [
			{
				"instance": {
					"instance_id": "item-123",
					"template_id": "sword_iron",
					"quantity": 1,
					"durability": 1.0
				},
				"compartment": "backpack"
			}
		],
		"compartment_caps": {
			"backpack": 50,
			"belt": 10,
			"craft_bag": 30
		},
		"weight_limit": 100.0
	}`

	equipData := `{
		"slots": {
			"main_hand": {
				"instance": {
					"instance_id": "item-123",
					"template_id": "sword_iron",
					"quantity": 1,
					"durability": 1.0
				},
				"cooldown_until": "2024-01-01T00:00:00Z"
			}
		}
	}`

	skillsData := `{
		"melee": 10,
		"defense": 5
	}`

	state := PlayerState{
		Pos:           spatial.Vec2{X: 100, Z: 200},
		Logins:        1,
		Updated:       time.Now(),
		Version:       1,
		InventoryData: []byte(testData),
		EquipmentData: []byte(equipData),
		SkillsData:    []byte(skillsData),
	}

	// Save player state
	err := store.Save(ctx, playerID, state)
	if err != nil {
		t.Fatalf("Failed to save player state: %v", err)
	}

	// Load and verify
	loaded, exists, err := store.Load(ctx, playerID)
	if err != nil {
		t.Fatalf("Failed to load player state: %v", err)
	}
	if !exists {
		t.Fatal("Player state should exist")
	}

	// Verify inventory data is preserved
	if string(loaded.InventoryData) != testData {
		t.Error("Inventory data not preserved correctly")
	}

	// Verify equipment data is preserved
	if string(loaded.EquipmentData) != equipData {
		t.Error("Equipment data not preserved correctly")
	}

	// Verify skills data is preserved
	if string(loaded.SkillsData) != skillsData {
		t.Error("Skills data not preserved correctly")
	}

	t.Logf("Successfully persisted and restored player state with inventory, equipment, and skills")
}
