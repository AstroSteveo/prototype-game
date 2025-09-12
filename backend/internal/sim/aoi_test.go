package sim

import (
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
)

func TestAOI_InclusiveBoundaryAndExclusion(t *testing.T) {
	e := NewEngine(Config{CellSize: 10, AOIRadius: 5, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})
	// Anchor player at origin
	p := e.DevSpawn("p0", "Anchor", spatial.Vec2{X: 0, Z: 0})
	// Entity exactly at radius 5 (3,4) distance 5
	_ = e.AddOrUpdatePlayer("p1", "In", spatial.Vec2{X: 3, Z: 4}, spatial.Vec2{})
	// Entity just outside radius
	_ = e.AddOrUpdatePlayer("p2", "Out", spatial.Vec2{X: 5.1, Z: 0}, spatial.Vec2{})

	got := e.QueryAOI(p.Pos, e.cfg.AOIRadius, p.ID)
	has := func(id string) bool {
		for _, en := range got {
			if en.ID == id {
				return true
			}
		}
		return false
	}
	if !has("p1") {
		t.Fatalf("expected to include p1 at radius boundary")
	}
	if has("p2") {
		t.Fatalf("did not expect to include p2 beyond radius")
	}
	if has("p0") {
		t.Fatalf("did not expect to include self in AOI results")
	}
}

func TestAOI_CoversAcrossBorder_NoFlap(t *testing.T) {
	// AOI radius smaller than cell size; ensure entity across border remains visible
	e := NewEngine(Config{CellSize: 10, AOIRadius: 5, TickHz: 20, SnapshotHz: 10, HandoverHysteresisM: 2})
	// Anchor near east border of cell (0,0)
	p := e.AddOrUpdatePlayer("pA", "A", spatial.Vec2{X: 9.9, Z: 0}, spatial.Vec2{})
	// Neighbor entity just across border within radius
	_ = e.AddOrUpdatePlayer("pB", "B", spatial.Vec2{X: 10.5, Z: 0}, spatial.Vec2{})

	got := e.QueryAOI(p.Pos, e.cfg.AOIRadius, p.ID)
	seen := false
	for _, en := range got {
		if en.ID == "pB" {
			seen = true
			break
		}
	}
	if !seen {
		t.Fatalf("expected to see pB across border within radius")
	}

	// Move anchor slightly across the border; AOI should still include pB
	_ = e.DevSetVelocity("pA", spatial.Vec2{X: 1, Z: 0})
	e.Step(200 * time.Millisecond) // move to ~10.1
	snap, _ := e.GetPlayer("pA")
	got2 := e.QueryAOI(snap.Pos, e.cfg.AOIRadius, snap.ID)
	seen2 := false
	for _, en := range got2 {
		if en.ID == "pB" {
			seen2 = true
			break
		}
	}
	if !seen2 {
		t.Fatalf("expected to continue seeing pB after crossing border slightly")
	}
}
