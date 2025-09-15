package join

import (
	"context"
	"testing"
	"time"

	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/testutil"
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

func TestHandleJoin_ContextTimeout(t *testing.T) {
	eng := newTestEngine()
	auth := testutil.SlowAuth{Delay: 2 * time.Second} // Auth will take 2 seconds

	// Create a context that times out in 500ms
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, errMsg := HandleJoin(ctx, auth, eng, Hello{Token: "slow"})
	elapsed := time.Since(start)

	// Should fail due to timeout, not take the full 2 seconds
	if errMsg == nil || errMsg.Code != "auth" {
		t.Fatalf("expected auth error due to timeout, got: %#v", errMsg)
	}

	// Should complete in less than 1 second (much less than the 2s auth delay)
	if elapsed > time.Second {
		t.Fatalf("HandleJoin took too long (%v), timeout context not working", elapsed)
	}
}

func TestHTTPAuth_ClientTimeout(t *testing.T) {
	// Create an HTTPAuth instance
	auth := NewHTTPAuth("http://localhost:9999") // non-existent server

	// Verify the client has a timeout set
	if auth.Client.Timeout != 3*time.Second {
		t.Fatalf("expected HTTPAuth client timeout to be 3s, got: %v", auth.Client.Timeout)
	}

	// Test that validation fails quickly when server is unreachable
	ctx := context.Background()
	start := time.Now()
	_, _, ok := auth.Validate(ctx, "anytoken")
	elapsed := time.Since(start)

	// Should fail (unreachable server)
	if ok {
		t.Fatalf("expected validation to fail for unreachable server")
	}

	// Should complete within timeout + some buffer (4 seconds max)
	if elapsed > 4*time.Second {
		t.Fatalf("HTTPAuth validation took too long (%v), client timeout not working", elapsed)
	}
}
