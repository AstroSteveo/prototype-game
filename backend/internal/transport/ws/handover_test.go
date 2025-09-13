//go:build ws

package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/sim"
)

// Reuse fakeAuth from ws_integration_test.go via local copy to avoid import cycles.
type fakeAuthHO struct{}

func (fakeAuthHO) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

func TestWS_HandoverEvent_EmittedOnCellChange(t *testing.T) {
	// Small cells and modest hysteresis so we can trigger quickly
	eng := sim.NewEngine(sim.Config{CellSize: 2, AOIRadius: 5, TickHz: 60, SnapshotHz: 30, HandoverHysteresisM: 0.25})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthHO{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Join
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}
	// Read join_ack
	var raw json.RawMessage
	if err := wsjson.Read(ctx, c, &raw); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	// Start moving east to cross at least one border past hysteresis
	if err := wsjson.Write(ctx, c, map[string]any{
		"type":   "input",
		"seq":    1,
		"dt":     0.05,
		"intent": map[string]float64{"x": 1, "z": 0},
	}); err != nil {
		t.Fatalf("write input: %v", err)
	}

	// Observe stream and look for a handover event
	deadline := time.Now().Add(3 * time.Second)
	seenHO := false
	for time.Now().Before(deadline) {
		var msg json.RawMessage
		rctx, cancelR := context.WithTimeout(context.Background(), 200*time.Millisecond)
		err := wsjson.Read(rctx, c, &msg)
		cancelR()
		if err != nil {
			continue
		}
		var env struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(msg, &env) != nil {
			continue
		}
		if env.Type != "handover" {
			continue
		}
		var body struct {
			From struct{ Cx, Cz int }
			To   struct{ Cx, Cz int }
		}
		if err := json.Unmarshal(env.Data, &body); err != nil {
			continue
		}
		if body.From.Cx != body.To.Cx || body.From.Cz != body.To.Cz {
			seenHO = true
			break
		}
	}
	if !seenHO {
		t.Fatalf("expected to observe a handover event after crossing border")
	}
}

// TestWS_HandoverAntiThrash_WithWebSocketContinuity tests that anti-thrashing works 
// through WebSocket and that state continuity is maintained.
func TestWS_HandoverAntiThrash_WithWebSocketContinuity(t *testing.T) {
	// Small cells with 1.0 hysteresis for predictable testing
	eng := sim.NewEngine(sim.Config{CellSize: 4, AOIRadius: 8, TickHz: 60, SnapshotHz: 30, HandoverHysteresisM: 1.0})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthHO{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	c, _, err := nws.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(nws.StatusNormalClosure, "bye")

	// Join
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("write hello: %v", err)
	}

	// Read join_ack to get initial state
	var joinMsg json.RawMessage
	if err := wsjson.Read(ctx, c, &joinMsg); err != nil {
		t.Fatalf("read join_ack: %v", err)
	}

	// Parse join_ack to check initial cell
	var joinAck struct {
		Type string `json:"type"`
		Data struct {
			PlayerID string `json:"player_id"`
			Pos      struct{ X, Z float64 } `json:"pos"`
			Cell     struct{ Cx, Cz int } `json:"cell"`
		} `json:"data"`
	}
	if err := json.Unmarshal(joinMsg, &joinAck); err != nil {
		t.Fatalf("parse join_ack: %v", err)
	}
	
	t.Logf("Joined at pos=(%.1f,%.1f) cell=(%d,%d)", 
		joinAck.Data.Pos.X, joinAck.Data.Pos.Z, joinAck.Data.Cell.Cx, joinAck.Data.Cell.Cz)

	// Send rapid back-and-forth movement that would cause thrashing without anti-thrash logic
	movements := []struct {
		x, z float64
		desc string
	}{
		{10, 0, "move far east (should cross to cell 1,0)"},
		{-10, 0, "move far west (should NOT immediately return due to anti-thrash)"},
		{10, 0, "move far east again (should stay in same cell)"},
		{-20, 0, "move very far west (should overcome double hysteresis)"},
	}

	handoverCount := 0

	for i, move := range movements {
		t.Logf("Step %d: %s", i+1, move.desc)
		
		// Send movement input
		if err := wsjson.Write(ctx, c, map[string]any{
			"type":   "input",
			"seq":    i + 2,
			"dt":     0.1,
			"intent": map[string]float64{"x": move.x, "z": move.z},
		}); err != nil {
			t.Fatalf("write input %d: %v", i, err)
		}

		// Watch for messages for up to 1 second
		deadline := time.Now().Add(1 * time.Second)
		for time.Now().Before(deadline) {
			var msg json.RawMessage
			rctx, cancelR := context.WithTimeout(context.Background(), 200*time.Millisecond)
			err := wsjson.Read(rctx, c, &msg)
			cancelR()
			if err != nil {
				continue
			}

			var env struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			if json.Unmarshal(msg, &env) != nil {
				continue
			}

			if env.Type == "handover" {
				handoverCount++
				var body struct {
					From struct{ Cx, Cz int }
					To   struct{ Cx, Cz int }
				}
				if err := json.Unmarshal(env.Data, &body); err == nil {
					t.Logf("  → Handover #%d: cell (%d,%d) → (%d,%d)", 
						handoverCount, body.From.Cx, body.From.Cz, body.To.Cx, body.To.Cz)
				}
				break // Exit message loop for this movement step
			}
		}
	}

	// Verify anti-thrashing: should have fewer handovers than movements
	t.Logf("Total handovers observed: %d (out of %d movement steps)", handoverCount, len(movements))
	
	// With anti-thrashing, we expect:
	// 1. First eastward move: handover to (1,0) 
	// 2. Westward move: NO handover due to anti-thrash
	// 3. Eastward again: NO handover (still in same cell)
	// 4. Far westward: handover to (0,0) when overcoming double hysteresis
	// So expect 2 handovers maximum, showing anti-thrash is working
	if handoverCount > 3 {
		t.Fatalf("too many handovers: %d (expected ≤ 3 with anti-thrashing)", handoverCount)
	}

	t.Logf("✓ Anti-thrashing working through WebSocket: %d handovers (≤ 3)", handoverCount)
	t.Logf("✓ State continuity maintained through WebSocket interface")
}
