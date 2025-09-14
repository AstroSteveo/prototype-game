//go:build ws

package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	nws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"prototype-game/backend/internal/join"
	"prototype-game/backend/internal/metrics"
	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/spatial"
	"prototype-game/backend/internal/state"
)

// Register installs the websocket handler when built with the `ws` tag.
func Register(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine) {
	RegisterWithStore(mux, path, auth, eng, nil)
}

// RegisterWithStore is like Register but allows wiring a persistence store (US-501).
func RegisterWithStore(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine, store state.Store) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		c, err := nws.Accept(w, r, &nws.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			log.Printf("ws accept: %v", err)
			return
		}
		defer c.Close(nws.StatusNormalClosure, "bye")

		// metrics: track connected clients
		metrics.IncWSConnected()
		defer metrics.DecWSConnected()

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var hello join.Hello
		if err := wsjson.Read(ctx, c, &hello); err != nil {
			_ = wsjson.Write(ctx, c, map[string]any{"type": "error", "error": join.ErrorMsg{Code: "bad_request", Message: "invalid hello"}})
			return
		}
		// Handle join
		ack, em := join.HandleJoin(r.Context(), auth, eng, hello)
		if em != nil {
			_ = wsjson.Write(ctx, c, map[string]any{"type": "error", "error": em})
			return
		}
        // No resume token in US-501 branch
		if err := wsjson.Write(ctx, c, map[string]any{"type": "join_ack", "data": ack}); err != nil {
			return
		}

		// Keep connection open for input/state loop (US-103).
		// Basic protocol:
		//  - Client sends: {"type":"input", "seq":N, "dt":seconds, "intent":{"x":-1..1, "z":-1..1}}
		//  - Server sends periodic: {"type":"state", "data":{"ack":N, "player":{...}}}

		// Reader goroutine -> inputs channel
		type inputMsg struct {
			Type   string  `json:"type"`
			Seq    int     `json:"seq"`
			Dt     float64 `json:"dt"`
			Intent struct {
				X float64 `json:"x"`
				Z float64 `json:"z"`
			} `json:"intent"`
		}
		inputs := make(chan inputMsg, 16)
		done := make(chan struct{})
		go func() {
			defer close(done)
			// allow indefinite reading; use ping/pong timeouts via context deadline per message
			for {
				var raw json.RawMessage
				if err := wsjson.Read(r.Context(), c, &raw); err != nil {
					return
				}
				// Try to decode as input
				var in inputMsg
				if err := json.Unmarshal(raw, &in); err != nil {
					// ignore malformed input
					continue
				}
				if in.Type != "input" {
					// ignore unknown types for now
					continue
				}
				select {
				case inputs <- in:
				default:
					// drop if backpressured
				}
			}
		}()

		// State ticker
		cfg := eng.GetConfig()
		snapDur := time.Second / time.Duration(max(1, cfg.SnapshotHz))
		telemetryDur := time.Second // 1Hz telemetry
		ticker := time.NewTicker(snapDur)
		defer ticker.Stop()
		telemTicker := time.NewTicker(telemetryDur)
		defer telemTicker.Stop()
        lastAck := 0
		playerID := ack.PlayerID
		lastCell := ack.Cell // track last known owned cell to emit handover events
		// movement speed meters/sec when intent vector length is 1
		const moveSpeed = 3.0

		// writer loop
		for {
			select {
			case <-done:
				// On disconnect, persist last known position (US-501)
				if store != nil {
					if p, ok := eng.GetPlayer(playerID); ok {
						_ = store.Save(r.Context(), playerID, state.PlayerState{Pos: p.Pos, Logins: p.Logins, Updated: time.Now()})
					}
				}
				return
			case in := <-inputs:
				// clamp intent and update velocity
				vx := clamp(in.Intent.X, -1, 1) * moveSpeed
				vz := clamp(in.Intent.Z, -1, 1) * moveSpeed
				_ = eng.DevSetVelocity(playerID, spatial.Vec2{X: vx, Z: vz})
				if in.Seq > lastAck {
					lastAck = in.Seq
				}
			case <-ticker.C:
				// send state with AOI entities
				p, ok := eng.GetPlayer(playerID)
				if !ok {
					return
				}
				// If player's owned cell changed since last snapshot, emit a handover event first
				if p.OwnedCell != lastCell {
					metrics.ObserveHandoverLatency(time.Since(p.HandoverAt))
					hov := map[string]any{
						"type": "handover",
						"data": map[string]any{
							"from": lastCell,
							"to":   p.OwnedCell,
						},
					}
					hctx, cancelH := context.WithTimeout(r.Context(), 2*time.Second)
					_ = wsjson.Write(hctx, c, hov)
					cancelH()
					lastCell = p.OwnedCell
				}
				nearby := eng.QueryAOI(p.Pos, cfg.AOIRadius, p.ID)
				metrics.ObserveEntitiesInAOI(len(nearby))
				ents := make([]map[string]any, 0, len(nearby))
				for _, e := range nearby {
					ents = append(ents, map[string]any{
						"id":   e.ID,
						"pos":  e.Pos,
						"vel":  e.Vel,
						"kind": int(e.Kind),
						"name": e.Name,
					})
				}
				msg := map[string]any{
					"type": "state",
					"data": map[string]any{
						"ack":      lastAck,
						"player":   map[string]any{"id": p.ID, "pos": p.Pos, "vel": p.Vel},
						"entities": ents,
					},
				}
				// Observe snapshot payload size (JSON encoded)
				if bs, err := json.Marshal(msg); err == nil {
					metrics.ObserveSnapshotBytes(len(bs))
				}
				// short deadline to avoid blocking forever
				wctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
				_ = wsjson.Write(wctx, c, msg)
				cancel()
			case <-telemTicker.C:
				// measure RTT via websocket Ping/Pong
				start := time.Now()
				pingCtx, cancelPing := context.WithTimeout(r.Context(), 500*time.Millisecond)
				err := c.Ping(pingCtx)
				cancelPing()
				if err != nil {
					continue
				}
				rtt := time.Since(start).Seconds() * 1000.0 // ms
				telem := map[string]any{
					"type": "telemetry",
					"data": map[string]any{
						"tick_rate": cfg.TickHz,
						"rtt_ms":    rtt,
					},
				}
				tctx, cancelT := context.WithTimeout(r.Context(), 2*time.Second)
				_ = wsjson.Write(tctx, c, telem)
				cancelT()
			}
		}
	})
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
