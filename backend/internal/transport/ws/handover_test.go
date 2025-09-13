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
