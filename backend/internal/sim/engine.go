package sim

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"prototype-game/backend/internal/metrics"
	"prototype-game/backend/internal/spatial"
)

type Engine struct {
	cfg       Config
	mu        sync.RWMutex
	cells     map[spatial.CellKey]*CellInstance
	players   map[string]*Player // id -> player
	stopCh    chan struct{}
	stoppedCh chan struct{}
	// metrics (atomic)
	met struct {
		handovers   int64 // count of player handovers
		aoiQueries  int64 // number of AOI queries executed
		aoiEntities int64 // total entities returned across AOI queries
	}
}

func NewEngine(cfg Config) *Engine {
	return &Engine{
		cfg:       cfg,
		cells:     make(map[spatial.CellKey]*CellInstance),
		players:   make(map[string]*Player),
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}
}

func (e *Engine) Start() {
	go e.loop()
}

func (e *Engine) Stop(ctx context.Context) {
	close(e.stopCh)
	select {
	case <-e.stoppedCh:
	case <-ctx.Done():
	}
}

func (e *Engine) loop() {
	defer close(e.stoppedCh)
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
	// Check handovers.
	for _, p := range e.players {
		e.checkAndHandoverLocked(p)
	}
	// TODO: update bots & AI
}

func (e *Engine) snapshot() {
	// For MVP skeleton, just log entity counts per cell.
	e.mu.RLock()
	defer e.mu.RUnlock()
	total := len(e.players)
	if total == 0 {
		return
	}
	counts := 0
	for _, c := range e.cells {
		counts += len(c.Entities)
	}
	log.Printf("sim: snapshot players=%d entities=%d cells=%d", total, counts, len(e.cells))
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
