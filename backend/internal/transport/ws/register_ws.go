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
	RegisterWithStoreAndDevMode(mux, path, auth, eng, store, false)
}

// RegisterWithStoreAndDevMode is like RegisterWithStore but allows configuring dev mode for relaxed security.
func RegisterWithStoreAndDevMode(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine, store state.Store, devMode bool) {
	RegisterWithOptions(mux, path, auth, eng, store, WSOptions{DevMode: devMode})
}

// WSOptions contains configuration options for WebSocket behavior
type WSOptions struct {
	IdleTimeout time.Duration // if zero, defaults to 30 seconds
	DevMode     bool          // if true, enables relaxed security for local testing
}

// RegisterWithOptions allows configuring WebSocket behavior for testing
func RegisterWithOptions(mux *http.ServeMux, path string, auth join.AuthService, eng *sim.Engine, store state.Store, opts WSOptions) {
	idleTimeout := opts.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = 30 * time.Second
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Configure WebSocket accept options based on dev mode
		var acceptOptions *nws.AcceptOptions
		if opts.DevMode {
			// Development mode: relaxed security for local testing
			acceptOptions = &nws.AcceptOptions{InsecureSkipVerify: true}
		} else {
			// Production mode: strict origin checking with localhost and same-origin allowlist
			originPatterns := []string{
				"localhost",
				"localhost:*",
				"127.0.0.1",
				"127.0.0.1:*",
				"[::1]",
				"[::1]:*",
			}
			// Add the server's own host (same-origin) to the allowlist
			if r.Host != "" {
				originPatterns = append(originPatterns, r.Host)
				originPatterns = append(originPatterns, r.Host+":*")
			}
			acceptOptions = &nws.AcceptOptions{
				OriginPatterns: originPatterns,
			}
		}

		c, err := nws.Accept(w, r, acceptOptions)
		if err != nil {
			log.Printf("ws accept: %v", err)
			return
		}
		defer c.Close(nws.StatusNormalClosure, "bye")

		// Set read limit to prevent oversized messages (32KB)
		c.SetReadLimit(32 << 10)

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
		// Handle join (resume is optional; token still required by AuthService)
		ack, em := join.HandleJoin(ctx, auth, eng, hello)
		if em != nil {
			_ = wsjson.Write(ctx, c, map[string]any{"type": "error", "error": em})
			return
		}
		// Issue resume token for future reconnects
		ack.ResumeToken = defaultResume.Issue(ack.PlayerID)
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

		// Equipment command messages
		type equipMsg struct {
			Type       string `json:"type"`
			Seq        int    `json:"seq"`
			InstanceID string `json:"instance_id"`
			Slot       string `json:"slot"`
		}

		type unequipMsg struct {
			Type        string `json:"type"`
			Seq         int    `json:"seq"`
			Slot        string `json:"slot"`
			Compartment string `json:"compartment,omitempty"` // defaults to backpack if empty
		}
		inputs := make(chan inputMsg, 16)
		equipCmds := make(chan equipMsg, 16)
		unequipCmds := make(chan unequipMsg, 16)
		done := make(chan struct{})
		activityCh := make(chan time.Time, 1)

		// Message sequence tracking for idempotency
		processedSeqs := make(map[int]bool)
		seqCleanupTicker := time.NewTicker(30 * time.Second) // Clean old sequences periodically
		defer seqCleanupTicker.Stop()

		go func() {
			defer close(done)
			// per-message read deadline to prevent hanging on slow/malicious clients
			for {
				readCtx, cancelRead := context.WithTimeout(r.Context(), 2*time.Second)
				var raw json.RawMessage
				err := wsjson.Read(readCtx, c, &raw)
				cancelRead()
				if err != nil {
					return
				}
				// Signal activity
				select {
				case activityCh <- time.Now():
				default:
				}
				// Try to decode as input
				var in inputMsg
				if err := json.Unmarshal(raw, &in); err == nil && in.Type == "input" {
					select {
					case inputs <- in:
					default:
						// drop if backpressured
					}
					continue
				}

				// Try to decode as equip command
				var equipCmd equipMsg
				if err := json.Unmarshal(raw, &equipCmd); err == nil && equipCmd.Type == "equip" {
					select {
					case equipCmds <- equipCmd:
					default:
						// drop if backpressured
					}
					continue
				}

				// Try to decode as unequip command
				var unequipCmd unequipMsg
				if err := json.Unmarshal(raw, &unequipCmd); err == nil && unequipCmd.Type == "unequip" {
					select {
					case unequipCmds <- unequipCmd:
					default:
						// drop if backpressured
					}
					continue
				}

				// ignore unknown message types
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
		// idle timeout: disconnect clients idle for more than configured timeout
		idleTimer := time.NewTimer(idleTimeout)
		defer idleTimer.Stop()
		lastAck := 0
		playerID := ack.PlayerID
		// Validate resume token before trusting LastSeq
		if hello.Resume != "" {
			if defaultResume.Validate(hello.Resume, playerID) {
				// Restore lastAck from hello.LastSeq when resume token is valid
				lastAck = hello.LastSeq
			}
		}
		lastCell := ack.Cell // track last known owned cell to emit handover events

		// Track last sent versions for delta updates
		var lastInventoryVersion int64 = -1 // Force initial send
		var lastEquipmentVersion int64 = -1 // Force initial send
		var lastSkillsVersion int64 = -1    // Force initial send
		// movement speed meters/sec when intent vector length is 1
		const moveSpeed = 3.0

		// writer loop
		for {
			select {
			case <-idleTimer.C:
				log.Printf("ws: disconnecting idle client %s after %v", playerID, idleTimeout)
				return
			case <-done:
				// On disconnect, persist last known state including inventory/equipment (US-006)
				if store != nil {
					// Use background context with timeout instead of request context which will be canceled
					persistCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					eng.RequestPlayerDisconnectPersist(persistCtx, playerID)
				}
				return
			case <-activityCh:
				idleTimer.Reset(idleTimeout)
			case in := <-inputs:
				// clamp intent and update velocity
				vx := clamp(in.Intent.X, -1, 1) * moveSpeed
				vz := clamp(in.Intent.Z, -1, 1) * moveSpeed
				_ = eng.DevSetVelocity(playerID, spatial.Vec2{X: vx, Z: vz})
				if in.Seq > lastAck {
					lastAck = in.Seq
				}
			case equipCmd := <-equipCmds:
				// Handle equip command with idempotency check
				if processedSeqs[equipCmd.Seq] {
					// Already processed this command, ignore duplicate
					continue
				}
				processedSeqs[equipCmd.Seq] = true

				slotID := sim.SlotID(equipCmd.Slot)
				instanceID := sim.ItemInstanceID(equipCmd.InstanceID)
				now := time.Now()

				err := eng.EquipItem(playerID, instanceID, slotID, now)
				success := err == nil

				// Record metrics
				metrics.ObserveEquipOperation("equip", success)
				if err == sim.ErrEquipLocked {
					metrics.IncEquipCooldownBlocks()
				}

				sendEquipResult(r.Context(), c, success, "equip", equipCmd.Slot, err)

				if success {
					// Force inventory and equipment delta on next state update
					lastInventoryVersion = -1
					lastEquipmentVersion = -1
				}

			case unequipCmd := <-unequipCmds:
				// Handle unequip command with idempotency check
				if processedSeqs[unequipCmd.Seq] {
					// Already processed this command, ignore duplicate
					continue
				}
				processedSeqs[unequipCmd.Seq] = true

				slotID := sim.SlotID(unequipCmd.Slot)
				compartment := sim.CompartmentType(unequipCmd.Compartment)
				if compartment == "" {
					compartment = sim.CompartmentBackpack // Default compartment
				}
				now := time.Now()

				err := eng.UnequipItem(playerID, slotID, compartment, now)
				success := err == nil

				// Record metrics
				metrics.ObserveEquipOperation("unequip", success)
				if err == sim.ErrEquipLocked {
					metrics.IncEquipCooldownBlocks()
				}

				sendEquipResult(r.Context(), c, success, "unequip", unequipCmd.Slot, err)

				if success {
					// Force inventory and equipment delta on next state update
					lastInventoryVersion = -1
					lastEquipmentVersion = -1
				}
			case <-seqCleanupTicker.C:
				// Clean up old processed sequences to prevent memory growth
				// Keep only sequences from the last 100 messages
				if len(processedSeqs) > 100 {
					// Simple cleanup: clear all and start fresh
					processedSeqs = make(map[int]bool)
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

				// Prepare state message data
				msgData := map[string]any{
					"ack":      lastAck,
					"player":   map[string]any{"id": p.ID, "pos": p.Pos, "vel": p.Vel},
					"entities": ents,
				}

				// Add inventory delta if changed
				if p.InventoryVersion != lastInventoryVersion {
					playerMgr := eng.GetPlayerManager()
					encumbrance := playerMgr.GetPlayerEncumbrance(&p)
					msgData["inventory"] = map[string]any{
						"items":            p.Inventory.Items,
						"compartment_caps": p.Inventory.CompartmentCaps,
						"weight_limit":     p.Inventory.WeightLimit,
						"encumbrance":      encumbrance,
					}
					lastInventoryVersion = p.InventoryVersion
				}

				// Add equipment delta if changed
				if p.EquipmentVersion != lastEquipmentVersion {
					msgData["equipment"] = p.Equipment
					lastEquipmentVersion = p.EquipmentVersion
				}

				// Add skills delta if changed
				if p.SkillsVersion != lastSkillsVersion {
					msgData["skills"] = p.Skills
					lastSkillsVersion = p.SkillsVersion
				}

				msg := map[string]any{
					"type": "state",
					"data": msgData,
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
					log.Printf("ws: ping failed for client %s: %v", playerID, err)
					return
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

// sendError sends an error message to the WebSocket client
func sendError(ctx context.Context, c *nws.Conn, code, message string) {
	errorMsg := map[string]any{
		"type": "error",
		"data": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	errCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = wsjson.Write(errCtx, c, errorMsg)
}

// sendEquipResult sends an equipment operation result to the client
func sendEquipResult(ctx context.Context, c *nws.Conn, success bool, operation, slot string, err error) {
	var message string
	var code string
	if success {
		message = "Equipment operation successful"
		code = "success"
	} else {
		code = "equip_failed"
		if err != nil {
			switch err {
			case sim.ErrIllegalSlot:
				code = "illegal_slot"
				message = "Item cannot be equipped to this slot"
			case sim.ErrSkillGate:
				code = "skill_gate"
				message = "Insufficient skill level to equip item"
			case sim.ErrEquipLocked:
				code = "equip_locked"
				message = "Equipment slot is on cooldown"
			case sim.ErrItemNotFound:
				code = "item_not_found"
				message = "Item not found in inventory"
			default:
				message = err.Error()
			}
		}
	}

	resultMsg := map[string]any{
		"type": "equipment_result",
		"data": map[string]any{
			"operation": operation,
			"slot":      slot,
			"success":   success,
			"code":      code,
			"message":   message,
		},
	}
	resultCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_ = wsjson.Write(resultCtx, c, resultMsg)
}
