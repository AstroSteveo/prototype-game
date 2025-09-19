package sim

import (
	"time"

	"prototype-game/backend/internal/spatial"
)

type EntityKind int

const (
	KindPlayer EntityKind = iota
	KindBot
)

type Entity struct {
	ID   string
	Kind EntityKind
	Pos  spatial.Vec2
	Vel  spatial.Vec2
	Yaw  float64
	Name string
}

type Player struct {
	Entity
	OwnedCell  spatial.CellKey
	PrevCell   spatial.CellKey // Previous cell for anti-thrash logic
	HandoverAt time.Time
	ConnID     string // placeholder for connection id
	LastSeq    int

	// Inventory and equipment systems
	Inventory *Inventory     `json:"inventory"`
	Equipment *Equipment     `json:"equipment"`
	Skills    map[string]int `json:"skills"` // skill_name -> level

	// Delta tracking for efficient state updates
	InventoryVersion int64 `json:"-"` // Increment when inventory changes
	EquipmentVersion int64 `json:"-"` // Increment when equipment changes
	SkillsVersion    int64 `json:"-"` // Increment when skills change
}

type Config struct {
	CellSize            float64
	AOIRadius           float64
	TickHz              int
	SnapshotHz          int
	HandoverHysteresisM float64
	// Bots & density control
	TargetDensityPerCell int // desired actors (players+bots) per cell
	MaxBots              int // global cap across all cells
	// Debug settings
	DebugSnapshot bool // enable snapshot logging
}
