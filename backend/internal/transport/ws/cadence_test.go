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

type fakeAuthCad struct{}

func (fakeAuthCad) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

// US-202: Snapshot cadence (~10Hz) and payload budget (<30KB/s) locally.
func TestSnapshotCadenceAndPayloadBudget(t *testing.T) {
	// Configure SnapshotHz to 10 (100ms)
	eng := sim.NewEngine(sim.Config{CellSize: 10, AOIRadius: 5, TickHz: 50, SnapshotHz: 10, HandoverHysteresisM: 1})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthCad{}, eng)
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

	// Hello / join_ack
	if err := wsjson.Write(ctx, c, map[string]any{"token": "tok"}); err != nil {
		t.Fatalf("hello: %v", err)
	}
	var rm json.RawMessage
	if err := wsjson.Read(ctx, c, &rm); err != nil {
		t.Fatalf("join_ack: %v", err)
	}

	// Collect state messages for ~1.5s to measure cadence/payload.
	var times []time.Time
	var bytes int
	start := time.Now()
	deadline := start.Add(1500 * time.Millisecond)
	for time.Now().Before(deadline) {
		rctx, cancelR := context.WithTimeout(context.Background(), 300*time.Millisecond)
		var msg json.RawMessage
		if err := wsjson.Read(rctx, c, &msg); err == nil {
			// Filter for state messages only
			var env struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			if json.Unmarshal(msg, &env) == nil && env.Type == "state" {
				times = append(times, time.Now())
				bytes += len(msg)
			}
		}
		cancelR()
	}
	if len(times) < 8 { // expect ~15, tolerate at least 8
		t.Fatalf("insufficient state samples: %d", len(times))
	}
	// Compute average inter-arrival time
	var sum time.Duration
	for i := 1; i < len(times); i++ {
		sum += times[i].Sub(times[i-1])
	}
	avg := time.Duration(int64(sum) / int64(len(times)-1))
	if avg < 80*time.Millisecond || avg > 120*time.Millisecond {
		t.Fatalf("average cadence out of bounds: got %v, want 100ms ±20ms", avg)
	}
	// Compute average bytes/sec for state payloads
	dur := time.Since(start).Seconds()
	bps := float64(bytes) / dur
	if bps >= 30000 {
		t.Fatalf("payload budget exceeded: %.0f bytes/s", bps)
	}
}
