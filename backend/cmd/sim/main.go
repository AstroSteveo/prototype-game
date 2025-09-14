package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prototype-game/backend/internal/join"
	"prototype-game/backend/internal/metrics"
	"prototype-game/backend/internal/sim"
	"prototype-game/backend/internal/spatial"
	transportws "prototype-game/backend/internal/transport/ws"
)

type httpConfig struct {
	CellSize       float64 `json:"cell_size"`
	AOIRadius      float64 `json:"aoi_radius"`
	TickHz         int     `json:"tick_hz"`
	SnapshotHz     int     `json:"snapshot_hz"`
	HandoverHyster float64 `json:"handover_hysteresis"`
}

func main() {
	var (
		port       = flag.String("port", "8081", "HTTP listen port for sim service")
		cellSize   = flag.Float64("cell", 256, "cell size in meters")
		aoiRadius  = flag.Float64("aoi", 128, "AOI radius in meters")
		tickHz     = flag.Int("tick", 20, "simulation tick rate (Hz)")
		snapshotHz = flag.Int("snap", 10, "snapshot rate (Hz)")
		hysteresis = flag.Float64("hyst", 2, "handover hysteresis in meters")
		gatewayURL = flag.String("gateway", "http://localhost:8080", "gateway base URL for token validation")
		debug      = flag.Bool("debug", false, "enable debug logging (including snapshot logs)")
		botDensity = flag.Int("bot-density", 3, "target actors (players+bots) per cell")
		maxBots    = flag.Int("max-bots", 100, "maximum total bots across all cells")
	)
	flag.Parse()

	// Validate configuration flags
	if err := validateConfig(*cellSize, *aoiRadius, *tickHz, *snapshotHz, *hysteresis); err != nil {
		log.Fatalf("sim: invalid configuration: %v", err)
	}

	// Initialize Prometheus metrics registry and collectors
	metrics.Init()

	eng := sim.NewEngine(sim.Config{
		CellSize:             *cellSize,
		AOIRadius:            *aoiRadius,
		TickHz:               *tickHz,
		SnapshotHz:           *snapshotHz,
		HandoverHysteresisM:  *hysteresis,
		TargetDensityPerCell: *botDensity,
		MaxBots:              *maxBots,
		DebugSnapshot:        *debug,
	})
	eng.Start()
	log.Printf("sim: started. tick=%dHz snap=%dHz cell=%.0fm aoi=%.0fm bot-density=%d max-bots=%d",
		*tickHz, *snapshotHz, *cellSize, *aoiRadius, *botDensity, *maxBots)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(httpConfig{
			CellSize:       *cellSize,
			AOIRadius:      *aoiRadius,
			TickHz:         *tickHz,
			SnapshotHz:     *snapshotHz,
			HandoverHyster: *hysteresis,
		})
	})
	// Simple JSON metrics for development/observability (prep for US-NF1)
	mux.HandleFunc("/metrics.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(eng.MetricsSnapshot())
	})
	// Prometheus metrics endpoint
	mux.Handle("/metrics", metrics.Handler())
	// WebSocket endpoint (stub unless built with -tags ws)
	auth := join.NewHTTPAuth(*gatewayURL)
	transportws.Register(mux, "/ws", auth, eng)
	// Dev endpoints to poke the engine without a client transport yet.
	mux.HandleFunc("/dev/spawn", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			id = shortID()
		}
		name := q.Get("name")
		if name == "" {
			name = "Player"
		}
		x := parseFloat(q.Get("x"), 0)
		z := parseFloat(q.Get("z"), 0)
		p := eng.DevSpawn(id, name, spatial.Vec2{X: x, Z: z})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(p)
	})
	mux.HandleFunc("/dev/vel", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		vx := parseFloat(q.Get("vx"), 0)
		vz := parseFloat(q.Get("vz"), 0)
		ok := eng.DevSetVelocity(id, spatial.Vec2{X: vx, Z: vz})
		if !ok {
			http.Error(w, "unknown id", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "ok")
	})
	mux.HandleFunc("/dev/players", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(eng.DevList())
	})
	mux.HandleFunc("/dev/entities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(eng.DevListAllEntities())
	})

	srv := &http.Server{Addr: ":" + *port, Handler: mux}
	go func() {
		log.Printf("sim: http listening on :%s", *port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("sim http: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Printf("sim: shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	eng.Stop(ctx)
	log.Printf("sim: stopped")
}

func shortID() string {
	// timestamp-based short id (dev only)
	return fmt.Sprintf("p%x", time.Now().UnixNano()&0xfffffff)
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	var v float64
	if _, err := fmt.Sscanf(s, "%f", &v); err != nil {
		return def
	}
	return v
}

// validateConfig validates configuration parameters to prevent divide-by-zero and other issues
func validateConfig(cellSize, aoiRadius float64, tickHz, snapshotHz int, hysteresis float64) error {
	if cellSize <= 0 {
		return fmt.Errorf("cell size must be > 0, got %.2f", cellSize)
	}
	if aoiRadius < 0 {
		return fmt.Errorf("AOI radius must be >= 0, got %.2f", aoiRadius)
	}
	if tickHz < 1 {
		return fmt.Errorf("tick rate must be >= 1 Hz, got %d", tickHz)
	}
	if snapshotHz < 1 {
		return fmt.Errorf("snapshot rate must be >= 1 Hz, got %d", snapshotHz)
	}
	if hysteresis < 0 {
		return fmt.Errorf("handover hysteresis must be >= 0, got %.2f", hysteresis)
	}
	return nil
}
