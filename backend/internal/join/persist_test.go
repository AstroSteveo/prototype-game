package join

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

func TestHandleJoin_UsesSavedPosition(t *testing.T) {
	eng := newTestEngine()
	st := state.NewMemStore()
	SetStore(st)
	defer SetStore(nil)

	// Seed saved state for player p2
	_ = st.Save(context.Background(), "p2", state.PlayerState{Pos: spatial.Vec2{X: 7.5, Z: -1.25}, Logins: 3, Updated: time.Now()})

	auth := fakeAuth{"tok": {"p2", "Eve"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ack, err := HandleJoin(ctx, auth, eng, Hello{Token: "tok"})
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	if diff := abs(ack.Pos.X-7.5) + abs(ack.Pos.Z-(-1.25)); diff > 1e-9 {
		t.Fatalf("expected spawn at saved pos (7.5,-1.25), got %#v", ack.Pos)
	}
	// Verify login count incremented
	saved, ok, _ := st.Load(context.Background(), "p2")
	if !ok || saved.Logins != 4 {
		t.Fatalf("expected logins incremented to 4, got %+v", saved)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
