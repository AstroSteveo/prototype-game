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
	"prototype-game/backend/internal/spatial"
)

type fakeAuthAOI struct{}

func (fakeAuthAOI) Validate(ctx context.Context, token string) (string, string, bool) {
	if token == "tok" {
		return "p1", "Alice", true
	}
	return "", "", false
}

// Ensures that after a handover, AOI results include cross-border neighbors within radius
// without duplicates and within the next snapshot.
func TestWS_AOIContinuity_AcrossHandover(t *testing.T) {
	eng := sim.NewEngine(sim.Config{CellSize: 2, AOIRadius: 2, TickHz: 60, SnapshotHz: 30, HandoverHysteresisM: 0.25})
	eng.Start()
	defer eng.Stop(context.Background())

	mux := http.NewServeMux()
	Register(mux, "/ws", fakeAuthAOI{}, eng)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Pre-create a neighbor entity across the eastern border that will be within radius after crossing.
	// Neighbor at x=2.5 from origin; once client crosses to ~x>2.25, distance <= 0.25.
	_ = eng.DevSpawn("nb", "Neighbor", spatial.Vec2{X: 2.5, Z: 0})

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

	// Move east to trigger handover beyond hysteresis
	if err := wsjson.Write(ctx, c, map[string]any{
		"type":   "input",
		"seq":    1,
		"dt":     0.05,
		"intent": map[string]float64{"x": 1, "z": 0},
	}); err != nil {
		t.Fatalf("write input: %v", err)
	}

	sawHO := false
	checkedStateAfterHO := false
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		var msg json.RawMessage
		rctx, cancelR := context.WithTimeout(context.Background(), 250*time.Millisecond)
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
		switch env.Type {
		case "handover":
			sawHO = true
		case "state":
			if !sawHO {
				continue
			}
			// First state after handover should include neighbor and not have duplicates
			var body struct {
				Ack      int             `json:"ack"`
				Player   json.RawMessage `json:"player"`
				Entities []struct {
					ID string `json:"id"`
				} `json:"entities"`
			}
			if err := json.Unmarshal(env.Data, &body); err != nil {
				continue
			}
			// collect IDs and check for neighbor and duplicates
			seen := map[string]bool{}
			dup := false
			hasNeighbor := false
			for _, e := range body.Entities {
				if seen[e.ID] {
					dup = true
				}
				seen[e.ID] = true
				if e.ID == "nb" {
					hasNeighbor = true
				}
			}
			if !hasNeighbor {
				t.Fatalf("expected neighbor to be visible in first state after handover")
			}
			if dup {
				t.Fatalf("expected no duplicate entity IDs in entities list")
			}
			checkedStateAfterHO = true
			return
		}
	}
	if !sawHO {
		t.Fatalf("did not receive handover event in time")
	}
	if !checkedStateAfterHO {
		t.Fatalf("did not receive state after handover with expected AOI contents")
	}
}
