package join

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/sim"
)

type fakeAuth map[string][2]string // token -> [playerID, name]

func (f fakeAuth) Validate(ctx context.Context, token string) (string, string, bool) {
	if v, ok := f[token]; ok {
		return v[0], v[1], true
	}
	return "", "", false
}

func newTestEngine() *sim.Engine {
	return sim.NewEngine(sim.Config{
		CellSize:            10,
		AOIRadius:           5,
		TickHz:              20,
		SnapshotHz:          10,
		HandoverHysteresisM: 2,
	})
}

func TestHandleJoin_Success(t *testing.T) {
	eng := newTestEngine()
	auth := fakeAuth{"tok123": {"p1", "Alice"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ack, errMsg := HandleJoin(ctx, auth, eng, Hello{Token: "tok123"})
	if errMsg != nil {
		t.Fatalf("unexpected error: %+v", errMsg)
	}
	if ack.PlayerID != "p1" || ack.Cell.Cx != 0 || ack.Cell.Cz != 0 {
		t.Fatalf("bad ack: %#v", ack)
	}
	if ack.Config.TickHz != 20 || ack.Config.AOIRadius != 5 || ack.Config.CellSize != 10 {
		t.Fatalf("bad config in ack: %#v", ack.Config)
	}
	if ack.Pos.X != 0 || ack.Pos.Z != 0 {
		t.Fatalf("expected spawn at origin (0,0), got: %#v", ack.Pos)
	}
}

func TestHandleJoin_AuthFailure(t *testing.T) {
	eng := newTestEngine()
	auth := fakeAuth{}
	ctx := context.Background()
	_, errMsg := HandleJoin(ctx, auth, eng, Hello{Token: "bad"})
	if errMsg == nil || errMsg.Code != "auth" {
		t.Fatalf("expected auth error, got: %#v", errMsg)
	}
}

func TestHandleJoin_BadRequest(t *testing.T) {
	eng := newTestEngine()
	auth := fakeAuth{"tok": {"p", "Bob"}}
	ctx := context.Background()
	_, errMsg := HandleJoin(ctx, auth, eng, Hello{Token: ""})
	if errMsg == nil || errMsg.Code != "bad_request" {
		t.Fatalf("expected bad_request error, got: %#v", errMsg)
	}
}
