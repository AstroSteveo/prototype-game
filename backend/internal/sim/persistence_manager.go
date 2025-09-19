package sim

import (
	"context"
	"log"
	"sync"
	"time"

	"prototype-game/backend/internal/state"
)

// PersistenceManager handles periodic persistence and disconnect handling
type PersistenceManager struct {
	store       state.Store
	checkpoints chan PlayerPersistRequest
	disconnects chan PlayerPersistRequest
	done        chan struct{}
	wg          sync.WaitGroup
	engine      *Engine

	// Configuration
	checkpointInterval time.Duration
	batchSize          int

	// Metrics
	persistAttempts    int64
	persistSuccesses   int64
	persistFailures    int64
	lastPersistTime    time.Time
	avgPersistDuration time.Duration
}

// PlayerPersistRequest represents a request to persist player data
type PlayerPersistRequest struct {
	PlayerID string
	Priority PersistPriority
	Context  context.Context
}

// PersistPriority determines the urgency of persistence operations
type PersistPriority int

const (
	PriorityCheckpoint PersistPriority = iota // Regular periodic save
	PriorityDisconnect                        // Player disconnected
	PriorityShutdown                          // Server shutdown
)

// NewPersistenceManager creates a new persistence manager
func NewPersistenceManager(store state.Store, engine *Engine) *PersistenceManager {
	return &PersistenceManager{
		store:              store,
		checkpoints:        make(chan PlayerPersistRequest, 1000),
		disconnects:        make(chan PlayerPersistRequest, 100),
		done:               make(chan struct{}),
		engine:             engine,
		checkpointInterval: 30 * time.Second, // Configurable checkpoint interval
		batchSize:          10,
	}
}

// Start begins the persistence manager background workers
func (pm *PersistenceManager) Start(ctx context.Context) {
	pm.wg.Add(3)

	// Worker for high-priority disconnect persistence
	go pm.disconnectWorker(ctx)

	// Worker for lower-priority checkpoint persistence
	go pm.checkpointWorker(ctx)

	// Periodic checkpoint scheduler
	go pm.checkpointScheduler(ctx)
}

// Stop gracefully shuts down the persistence manager
func (pm *PersistenceManager) Stop() {
	close(pm.done)
	pm.wg.Wait()
}

// RequestCheckpoint queues a checkpoint save for a player
func (pm *PersistenceManager) RequestCheckpoint(ctx context.Context, playerID string) {
	select {
	case pm.checkpoints <- PlayerPersistRequest{
		PlayerID: playerID,
		Priority: PriorityCheckpoint,
		Context:  ctx,
	}:
	default:
		// Channel full, skip this checkpoint (best-effort)
		log.Printf("PersistenceManager: checkpoint queue full, skipping player %s", playerID)
	}
}

// RequestDisconnectPersist immediately queues a disconnect save for a player
func (pm *PersistenceManager) RequestDisconnectPersist(ctx context.Context, playerID string) {
	select {
	case pm.disconnects <- PlayerPersistRequest{
		PlayerID: playerID,
		Priority: PriorityDisconnect,
		Context:  ctx,
	}:
	case <-time.After(1 * time.Second):
		// If we can't queue in 1 second, something is very wrong
		log.Printf("PersistenceManager: disconnect queue blocked, forcing sync save for player %s", playerID)
		pm.persistPlayerSync(ctx, playerID)
	}
}

// disconnectWorker handles high-priority disconnect persistence
func (pm *PersistenceManager) disconnectWorker(ctx context.Context) {
	defer pm.wg.Done()

	for {
		select {
		case req := <-pm.disconnects:
			pm.persistPlayerSync(req.Context, req.PlayerID)
		case <-pm.done:
			// Drain remaining disconnect requests before shutdown
			for {
				select {
				case req := <-pm.disconnects:
					pm.persistPlayerSync(req.Context, req.PlayerID)
				default:
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// checkpointWorker handles lower-priority checkpoint persistence
func (pm *PersistenceManager) checkpointWorker(ctx context.Context) {
	defer pm.wg.Done()

	batch := make([]PlayerPersistRequest, 0, pm.batchSize)
	ticker := time.NewTicker(5 * time.Second) // Process batches every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case req := <-pm.checkpoints:
			batch = append(batch, req)
			if len(batch) >= pm.batchSize {
				pm.processBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				pm.processBatch(batch)
				batch = batch[:0]
			}
		case <-pm.done:
			// Process remaining batch before shutdown
			if len(batch) > 0 {
				pm.processBatch(batch)
			}
			return
		case <-ctx.Done():
			return
		}
	}
}

// checkpointScheduler periodically schedules checkpoints for all connected players
func (pm *PersistenceManager) checkpointScheduler(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.checkpointInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Schedule checkpoints for all connected players
			pm.scheduleAllPlayerCheckpoints(ctx)
		case <-pm.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// scheduleAllPlayerCheckpoints requests checkpoints for all connected players
func (pm *PersistenceManager) scheduleAllPlayerCheckpoints(ctx context.Context) {
	// This would need to be implemented to get all connected players from the engine
	// For now, we'll implement this as a stub
	// TODO: Add engine method to get all connected player IDs
}

// processBatch processes a batch of checkpoint requests
func (pm *PersistenceManager) processBatch(batch []PlayerPersistRequest) {
	start := time.Now()
	successes := 0

	for _, req := range batch {
		if pm.persistPlayerSync(req.Context, req.PlayerID) {
			successes++
		}
	}

	duration := time.Since(start)
	pm.persistAttempts += int64(len(batch))
	pm.persistSuccesses += int64(successes)
	pm.persistFailures += int64(len(batch) - successes)
	pm.lastPersistTime = time.Now()
	pm.avgPersistDuration = duration / time.Duration(len(batch))

	if successes < len(batch) {
		log.Printf("PersistenceManager: batch processed %d/%d successfully in %v",
			successes, len(batch), duration)
	}
}

// persistPlayerSync synchronously persists a single player's data
func (pm *PersistenceManager) persistPlayerSync(ctx context.Context, playerID string) bool {
	if pm.store == nil {
		return false
	}

	// Get player data from engine
	player, exists := pm.engine.GetPlayer(playerID)
	if !exists {
		log.Printf("PersistenceManager: player %s not found for persistence", playerID)
		return false
	}

	// Serialize player data
	persistState, err := SerializePlayerData(&player)
	if err != nil {
		log.Printf("PersistenceManager: failed to serialize player %s: %v", playerID, err)
		return false
	}

	// Load current state to get version for optimistic locking
	if currentState, exists, err := pm.store.Load(ctx, playerID); exists && err == nil {
		persistState.Version = currentState.Version
		persistState.Logins = currentState.Logins // Preserve login count
	}

	// Save to store
	err = pm.store.Save(ctx, playerID, persistState)
	if err != nil {
		if err == state.ErrOptimisticLock {
			log.Printf("PersistenceManager: optimistic lock conflict for player %s", playerID)
		} else {
			log.Printf("PersistenceManager: failed to save player %s: %v", playerID, err)
		}
		return false
	}

	return true
}

// GetMetrics returns persistence metrics for telemetry
func (pm *PersistenceManager) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"persist_attempts":     pm.persistAttempts,
		"persist_successes":    pm.persistSuccesses,
		"persist_failures":     pm.persistFailures,
		"last_persist_time":    pm.lastPersistTime,
		"avg_persist_duration": pm.avgPersistDuration.Milliseconds(),
		"checkpoint_queue_len": len(pm.checkpoints),
		"disconnect_queue_len": len(pm.disconnects),
	}
}
