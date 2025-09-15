package sim

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"prototype-game/backend/internal/metrics"
	"prototype-game/backend/internal/spatial"
)

type Engine struct {
	cfg          Config
	mu           sync.RWMutex
	cells        map[spatial.CellKey]*CellInstance
	players      map[string]*Player // id -> player
	bots         map[string]*botState
	rng          *rand.Rand
	stopCh       chan struct{}
	stoppedCh    chan struct{}
	nodeRegistry *NodeRegistry
	crossNodeSvc CrossNodeHandoverService
	// lifecycle guards
	startOnce sync.Once
	stopOnce  sync.Once
	// state flags
	started atomic.Bool
	stopped atomic.Bool
	// control accumulators
	densityAcc time.Duration
	// ids
	botSeq int64
	// metrics (atomic)
	met struct {
		handovers   int64 // count of player handovers
		aoiQueries  int64 // number of AOI queries executed
		aoiEntities int64 // total entities returned across AOI queries
	}
}

func NewEngine(cfg Config) *Engine {
	return &Engine{
		cfg:          cfg,
		cells:        make(map[spatial.CellKey]*CellInstance),
		players:      make(map[string]*Player),
		bots:         make(map[string]*botState),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		stopCh:       make(chan struct{}),
		stoppedCh:    make(chan struct{}),
		nodeRegistry: NewNodeRegistry(cfg.NodeID),
	}
}

func (e *Engine) Start() {
	e.startOnce.Do(func() {
		e.started.Store(true)
		go e.loop()
	})
}

func (e *Engine) Stop(ctx context.Context) {
	e.stopOnce.Do(func() { close(e.stopCh) })
	if !e.started.Load() {
		return
	}
	select {
	case <-e.stoppedCh:
	case <-ctx.Done():
	}
}

func (e *Engine) loop() {
	defer func() {
		e.stopped.Store(true)
		close(e.stoppedCh)
	}()
	tickDur := time.Second / time.Duration(max(1, e.cfg.TickHz))
	snapDur := time.Second / time.Duration(max(1, e.cfg.SnapshotHz))
	ticker := time.NewTicker(tickDur)
	defer ticker.Stop()
	lastSnap := time.Now()
	for {
		select {
		case <-e.stopCh:
			return
		case t := <-ticker.C:
			start := time.Now()
			e.tick(tickDur)
			metrics.ObserveTickDuration(time.Since(start))
			if t.Sub(lastSnap) >= snapDur {
				e.snapshot()
				lastSnap = t
			}
		}
	}
}

func (e *Engine) tick(dt time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Integrate very simple kinematics.
	for _, p := range e.players {
		p.Pos.X += p.Vel.X * dt.Seconds()
		p.Pos.Z += p.Vel.Z * dt.Seconds()
	}
	// Update bots: steer and integrate
	for ck, cell := range e.cells {
		for _, ent := range cell.Entities {
			if ent.Kind != KindBot {
				continue
			}
			st, ok := e.bots[ent.ID]
			if !ok {
				// Initialize missing state defensively
				st = &botState{OwnedCell: ck}
				e.bots[ent.ID] = st
			}
			e.updateBot(ent, dt, st)
			ent.Pos.X += ent.Vel.X * dt.Seconds()
			ent.Pos.Z += ent.Vel.Z * dt.Seconds()
			// Constrain bots to their owned cell (bounce at borders)
			e.constrainBotWithinCell(ent, st)
		}
	}
	// Check handovers.
	for _, p := range e.players {
		e.checkAndHandoverLocked(p)
	}
	// Density maintenance at 1Hz
	e.densityAcc += dt
	for e.densityAcc >= time.Second {
		e.maintainBotDensityLocked()
		e.densityAcc -= time.Second
	}
}

func (e *Engine) snapshot() {
	// For MVP skeleton, just log entity counts per cell.
	e.mu.RLock()
	defer e.mu.RUnlock()
	total := len(e.players)
	if total == 0 {
		return
	}
	// Only log when debug mode is enabled to avoid log spam
	if e.cfg.DebugSnapshot {
		counts := 0
		for _, c := range e.cells {
			counts += len(c.Entities)
		}
		log.Printf("sim: snapshot players=%d entities=%d cells=%d", total, counts, len(e.cells))
	}
}

// maintainBotDensityLocked attempts to keep actors per cell within ±20% of target
// while respecting a global MaxBots cap. e.mu must be held by caller.
func (e *Engine) maintainBotDensityLocked() {
	target := e.cfg.TargetDensityPerCell
	if target <= 0 || e.cfg.MaxBots < 0 {
		return
	}
	low := int(math.Floor(float64(target) * 0.8))
	if low < 0 {
		low = 0
	}
	high := int(math.Ceil(float64(target) * 1.2))
	if high < low {
		high = low
	}
	ramp := int(math.Ceil(float64(max(1, target)) / 10.0)) // reach target in ~10s
	if ramp < 1 {
		ramp = 1
	}

	// Helper: count bots globally using the state map
	totalBots := len(e.bots)

	// Iterate cells deterministically (by key order); collect keys first
	keys := make([]spatial.CellKey, 0, len(e.cells))
	for k := range e.cells {
		keys = append(keys, k)
	}
	// No need to sort strictly for correctness; stable enough for tests

	for _, k := range keys {
		cell := e.cells[k]
		players, bots := 0, 0
		for _, ent := range cell.Entities {
			switch ent.Kind {
			case KindPlayer:
				players++
			case KindBot:
				bots++
			}
		}
		active := players + bots
		if active < low {
			need := low - active
			spawn := min3(need, ramp, max(0, e.cfg.MaxBots-totalBots))
			for i := 0; i < spawn; i++ {
				if e.spawnBotInCellLocked(k) {
					totalBots++
				}
			}
		} else if active > high {
			excess := active - high
			remove := min(excess, ramp)
			for i := 0; i < remove; i++ {
				if e.removeOneBotFromCellLocked(k) {
					totalBots--
				} else {
					break
				}
			}
		}
	}
}

func (e *Engine) spawnBotInCellLocked(k spatial.CellKey) bool {
	c := e.getOrCreateCellLocked(k)
	if e.cfg.MaxBots > 0 && len(e.bots) >= e.cfg.MaxBots {
		return false
	}
	id := fmt.Sprintf("bot-%d", atomic.AddInt64(&e.botSeq, 1))
	// random position inside cell bounds
	x0 := float64(k.Cx) * e.cfg.CellSize
	z0 := float64(k.Cz) * e.cfg.CellSize
	pos := spatial.Vec2{X: x0 + e.rng.Float64()*e.cfg.CellSize, Z: z0 + e.rng.Float64()*e.cfg.CellSize}
	ent := &Entity{ID: id, Kind: KindBot, Pos: pos, Name: id}
	c.Entities[id] = ent
	// initial state
	st := &botState{OwnedCell: k}
	// choose initial dir/retarget to avoid stationary
	e.updateBot(ent, 0, st)
	e.bots[id] = st
	return true
}

func (e *Engine) removeOneBotFromCellLocked(k spatial.CellKey) bool {
	c, ok := e.cells[k]
	if !ok {
		return false
	}
	for id, ent := range c.Entities {
		if ent.Kind == KindBot {
			delete(c.Entities, id)
			delete(e.bots, id)
			return true
		}
	}
	return false
}

func min3(a, b, c int) int { return min(min(a, b), c) }

// constrainBotWithinCell clamps a bot position to its owned cell and reflects direction when hitting borders.
func (e *Engine) constrainBotWithinCell(ent *Entity, st *botState) {
	x0 := float64(st.OwnedCell.Cx) * e.cfg.CellSize
	z0 := float64(st.OwnedCell.Cz) * e.cfg.CellSize
	x1 := x0 + e.cfg.CellSize
	z1 := z0 + e.cfg.CellSize
	bounced := false
	if ent.Pos.X < x0 {
		ent.Pos.X = x0
		st.dir.X = math.Abs(st.dir.X)
		bounced = true
	} else if ent.Pos.X > x1 {
		ent.Pos.X = x1
		st.dir.X = -math.Abs(st.dir.X)
		bounced = true
	}
	if ent.Pos.Z < z0 {
		ent.Pos.Z = z0
		st.dir.Z = math.Abs(st.dir.Z)
		bounced = true
	} else if ent.Pos.Z > z1 {
		ent.Pos.Z = z1
		st.dir.Z = -math.Abs(st.dir.Z)
		bounced = true
	}
	if bounced {
		// Immediately apply new velocity after bounce
		ent.Vel = spatial.Vec2{X: st.dir.X * botSpeed, Z: st.dir.Z * botSpeed}
	}
}

// AddOrUpdatePlayer creates or updates a player entity and places it in the correct cell.
func (e *Engine) AddOrUpdatePlayer(id, name string, pos spatial.Vec2, vel spatial.Vec2) *Player {
	e.mu.Lock()
	defer e.mu.Unlock()
	cx, cz := spatial.WorldToCell(pos.X, pos.Z, e.cfg.CellSize)
	key := spatial.CellKey{Cx: cx, Cz: cz}
	cell := e.getOrCreateCellLocked(key)
	pl, ok := e.players[id]
	if !ok {
		pl = &Player{Entity: Entity{ID: id, Kind: KindPlayer, Pos: pos, Vel: vel, Name: name}, OwnedCell: key}
		e.players[id] = pl
		cell.Entities[id] = &pl.Entity
	} else {
		// update
		pl.Pos, pl.Vel, pl.Name = pos, vel, name
		if pl.OwnedCell != key {
			// immediate place correction
			e.moveEntityLocked(pl, pl.OwnedCell, key)
			pl.OwnedCell = key
		}
	}
	return pl
}

func (e *Engine) getOrCreateCellLocked(key spatial.CellKey) *CellInstance {
	cell, ok := e.cells[key]
	if !ok {
		cell = NewCellInstance(key)
		e.cells[key] = cell
	}
	return cell
}

func (e *Engine) moveEntityLocked(p *Player, from, to spatial.CellKey) {
	if from == to {
		return
	}
	if c, ok := e.cells[from]; ok {
		delete(c.Entities, p.ID)
	}
	nc := e.getOrCreateCellLocked(to)
	nc.Entities[p.ID] = &p.Entity
}

// Step advances the simulation by dt. Exposed for tests and headless driving.
func (e *Engine) Step(dt time.Duration) {
	e.tick(dt)
}

// GetPlayer returns a snapshot copy of a player by id.
func (e *Engine) GetPlayer(id string) (Player, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	p, ok := e.players[id]
	if !ok {
		return Player{}, false
	}
	return *p, true
}

// DevSpawn creates a player at a position with zero velocity (dev-only helper).
func (e *Engine) DevSpawn(id, name string, pos spatial.Vec2) *Player {
	return e.AddOrUpdatePlayer(id, name, pos, spatial.Vec2{})
}

// DevSetVelocity sets a player's velocity (dev-only helper).
func (e *Engine) DevSetVelocity(id string, vel spatial.Vec2) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.players[id]
	if !ok {
		return false
	}
	p.Vel = vel
	return true
}

// DevSetPosition sets a player's position directly (dev-only helper).
func (e *Engine) DevSetPosition(id string, pos spatial.Vec2) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.players[id]
	if !ok {
		return false
	}
	p.Pos = pos
	return true
}

// DevList returns a snapshot list of current players (dev-only helper).
func (e *Engine) DevList() []Player {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Player, 0, len(e.players))
	for _, p := range e.players {
		out = append(out, *p)
	}
	return out
}

// DevListAllEntities returns a snapshot list of all entities including bots (dev-only helper).
func (e *Engine) DevListAllEntities() []Entity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var entities []Entity
	for _, cell := range e.cells {
		for _, ent := range cell.Entities {
			entities = append(entities, *ent)
		}
	}
	return entities
}

// GetConfig returns a copy of the engine's config.
func (e *Engine) GetConfig() Config { return e.cfg }

// QueryAOI returns a snapshot list of entities within radius of the given position.
// The result excludes the entity with id == excludeID (typically the querying player).
func (e *Engine) QueryAOI(pos spatial.Vec2, radius float64, excludeID string) []Entity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if radius <= 0 {
		return nil
	}
	// Determine world cell of the position to build a 3x3 neighborhood query.
	cx, cz := spatial.WorldToCell(pos.X, pos.Z, e.cfg.CellSize)
	center := spatial.CellKey{Cx: cx, Cz: cz}
	neigh := spatial.Neighbors3x3(center)
	r2 := radius * radius
	const eps = 1e-9 // tolerance to avoid flapping from FP roundoff at the boundary
	out := make([]Entity, 0, 16)
	for _, k := range neigh {
		cell, ok := e.cells[k]
		if !ok {
			continue
		}
		for id, ent := range cell.Entities {
			if id == excludeID {
				continue
			}
			if spatial.Dist2(ent.Pos, pos) <= r2+eps {
				out = append(out, *ent)
			}
		}
	}
	// metrics: record AOI query volume and total returned entities
	atomic.AddInt64(&e.met.aoiQueries, 1)
	atomic.AddInt64(&e.met.aoiEntities, int64(len(out)))
	return out
}

// Metrics holds a snapshot of engine metrics.
type Metrics struct {
	Handovers        int64   `json:"handovers"`
	AOIQueries       int64   `json:"aoi_queries"`
	AOIEntitiesTotal int64   `json:"aoi_entities_total"`
	AOIAvgEntities   float64 `json:"aoi_avg_entities"`
}

// MetricsSnapshot returns a copy of current counters.
func (e *Engine) MetricsSnapshot() Metrics {
	q := atomic.LoadInt64(&e.met.aoiQueries)
	ent := atomic.LoadInt64(&e.met.aoiEntities)
	ho := atomic.LoadInt64(&e.met.handovers)
	avg := 0.0
	if q > 0 {
		avg = float64(ent) / float64(q)
	}
	return Metrics{Handovers: ho, AOIQueries: q, AOIEntitiesTotal: ent, AOIAvgEntities: avg}
}

// RegisterNode registers a remote node in the node registry
func (e *Engine) RegisterNode(nodeInfo *NodeInfo) {
	e.nodeRegistry.RegisterNode(nodeInfo)
}

// UnregisterNode removes a node from the registry
func (e *Engine) UnregisterNode(nodeID string) {
	e.nodeRegistry.UnregisterNode(nodeID)
}

// AssignCellToNode assigns ownership of a cell to a specific node
func (e *Engine) AssignCellToNode(cell spatial.CellKey, nodeID string) {
	e.nodeRegistry.AssignCell(cell, nodeID)
}

// GetCellOwner returns the node ID that owns the given cell
func (e *Engine) GetCellOwner(cell spatial.CellKey) string {
	return e.nodeRegistry.GetCellOwner(cell)
}

// IsLocalCell returns true if the cell is owned by the local node
func (e *Engine) IsLocalCell(cell spatial.CellKey) bool {
	return e.nodeRegistry.IsLocalCell(cell)
}

// GetNodeInfo returns information about a specific node
func (e *Engine) GetNodeInfo(nodeID string) (*NodeInfo, bool) {
	return e.nodeRegistry.GetNodeInfo(nodeID)
}

// GetLocalNodeID returns the local node's ID
func (e *Engine) GetLocalNodeID() string {
	return e.nodeRegistry.GetLocalNodeID()
}

// ListNodes returns all registered nodes
func (e *Engine) ListNodes() map[string]*NodeInfo {
	return e.nodeRegistry.ListNodes()
}

// SetCrossNodeHandoverService sets the cross-node handover service
func (e *Engine) SetCrossNodeHandoverService(svc CrossNodeHandoverService) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.crossNodeSvc = svc
}
