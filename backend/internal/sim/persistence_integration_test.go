package sim

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

func TestInventoryPersistence_EndToEnd(t *testing.T) {
	// Test complete inventory persistence flow
	eng := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})

	// Set up in-memory persistence store
	store := state.NewMemStore()
	eng.SetPersistenceStore(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng.StartPersistence(ctx)
	defer eng.StopPersistence()

	playerID := "test-player-persistence"
	playerName := "TestPlayer"
	initialPos := spatial.Vec2{X: 100, Z: 200}

	// Add player to engine
	player := eng.AddOrUpdatePlayer(playerID, playerName, initialPos, spatial.Vec2{})
	playerMgr := eng.GetPlayerManager()

	// Initialize player with inventory and equipment
	playerMgr.InitializePlayer(player)

	// Add skills first so we can equip items
	player.Skills["melee"] = 15
	player.Skills["defense"] = 8

	// Add some items to inventory
	ironSword := ItemInstance{
		InstanceID: "sword-001",
		TemplateID: "sword_iron",
		Quantity:   1,
		Durability: 1.0,
	}

	err := playerMgr.AddItemToInventory(player, ironSword, CompartmentBackpack)
	if err != nil {
		t.Fatalf("Failed to add item to inventory: %v", err)
	}

	// Equip the sword
	err = playerMgr.EquipItem(player, ironSword.InstanceID, SlotMainHand, time.Now())
	if err != nil {
		t.Fatalf("Failed to equip item: %v", err)
	}

	// Move player to a different position
	newPos := spatial.Vec2{X: 150, Z: 250}
	player.Pos = newPos

	// Request immediate persistence (simulate disconnect)
	eng.RequestPlayerDisconnectPersist(ctx, playerID)

	// Give persistence manager time to process
	time.Sleep(100 * time.Millisecond)

	// Verify state was persisted
	persistedState, exists, err := store.Load(ctx, playerID)
	if err != nil {
		t.Fatalf("Failed to load persisted state: %v", err)
	}
	if !exists {
		t.Fatal("Player state should be persisted")
	}

	// Verify position was saved
	if persistedState.Pos.X != newPos.X || persistedState.Pos.Z != newPos.Z {
		t.Errorf("Position not persisted correctly: got %v, want %v", persistedState.Pos, newPos)
	}

	// Create a new engine instance (simulate server restart)
	eng2 := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})
	eng2.SetPersistenceStore(store)

	// Simulate player reconnecting - restore from persistence
	restoredPlayer := eng2.AddOrUpdatePlayer(playerID, playerName, spatial.Vec2{}, spatial.Vec2{})
	playerMgr2 := eng2.GetPlayerManager()

	// Deserialize the persisted state
	templates := playerMgr2.GetAllItemTemplates()
	err = DeserializePlayerData(persistedState, restoredPlayer, templates)
	if err != nil {
		t.Fatalf("Failed to deserialize player data: %v", err)
	}

	// Verify position restoration
	if restoredPlayer.Pos.X != newPos.X || restoredPlayer.Pos.Z != newPos.Z {
		t.Errorf("Position not restored correctly: got %v, want %v", restoredPlayer.Pos, newPos)
	}

	// Verify equipment restoration
	if restoredPlayer.Equipment == nil {
		t.Fatal("Equipment should be restored")
	}

	equippedItem := restoredPlayer.Equipment.GetSlot(SlotMainHand)
	if equippedItem == nil {
		t.Fatal("Main hand item should be restored")
	}

	if equippedItem.Instance.InstanceID != ironSword.InstanceID {
		t.Errorf("Equipped item not restored correctly: got %s, want %s",
			equippedItem.Instance.InstanceID, ironSword.InstanceID)
	}

	// Verify skills restoration
	if restoredPlayer.Skills == nil {
		t.Fatal("Skills should be restored")
	}

	if restoredPlayer.Skills["melee"] != 15 {
		t.Errorf("Melee skill not restored correctly: got %d, want %d",
			restoredPlayer.Skills["melee"], 15)
	}

	if restoredPlayer.Skills["defense"] != 8 {
		t.Errorf("Defense skill not restored correctly: got %d, want %d",
			restoredPlayer.Skills["defense"], 8)
	}

	// Verify encumbrance calculation works with restored data
	encumbrance := playerMgr2.GetPlayerEncumbrance(restoredPlayer)
	if encumbrance.CurrentWeight <= 0 {
		t.Error("Encumbrance should reflect equipped items")
	}

	t.Logf("Successfully persisted and restored player state")
	t.Logf("Position: %v", restoredPlayer.Pos)
	t.Logf("Equipped items: %d", len(restoredPlayer.Equipment.Slots))
	t.Logf("Skills: %v", restoredPlayer.Skills)
	t.Logf("Encumbrance: %.2f kg", encumbrance.CurrentWeight)
}

func TestPersistenceManager_Metrics(t *testing.T) {
	eng := NewEngine(Config{
		CellSize:             256,
		AOIRadius:            128,
		TickHz:               20,
		SnapshotHz:           10,
		HandoverHysteresisM:  2,
		TargetDensityPerCell: 3,
		MaxBots:              100,
	})

	store := state.NewMemStore()
	eng.SetPersistenceStore(store)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eng.StartPersistence(ctx)
	defer eng.StopPersistence()

	playerID := "test-metrics-player"
	player := eng.AddOrUpdatePlayer(playerID, "TestPlayer", spatial.Vec2{}, spatial.Vec2{})
	playerMgr := eng.GetPlayerManager()
	playerMgr.InitializePlayer(player)

	// Request several persistence operations
	for i := 0; i < 5; i++ {
		eng.RequestPlayerCheckpoint(ctx, playerID)
	}

	eng.RequestPlayerDisconnectPersist(ctx, playerID)

	// Give persistence manager time to process
	time.Sleep(200 * time.Millisecond)

	// Check metrics
	metrics := eng.GetPersistenceMetrics()
	if len(metrics) == 0 {
		t.Error("Should have persistence metrics")
	}

	t.Logf("Persistence metrics: %+v", metrics)

	// Verify some key metrics exist
	if _, exists := metrics["persist_attempts"]; !exists {
		t.Error("Should have persist_attempts metric")
	}

	if _, exists := metrics["persist_successes"]; !exists {
		t.Error("Should have persist_successes metric")
	}
}
