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
}

type Config struct {
	CellSize            float64
	AOIRadius           float64
	TickHz              int
	SnapshotHz          int
	HandoverHysteresisM float64
}
