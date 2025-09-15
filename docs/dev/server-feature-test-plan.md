# Server Feature & Test Coverage Guide

This guide summarizes the backend feature set and maps each scenario to the automated tests that cover it. It focuses on the simulation service, persistence layer, and WebSocket transport that together implement the prototype game's server side.

## Architecture Snapshot

The simulation binary wires together the engine, persistence store, WebSocket transport, metrics, and developer HTTP endpoints, providing configurable cell sizing, snapshot rates, bot density targets, and graceful shutdown logic.【F:backend/cmd/sim/main.go†L31-L183】 Configuration validation keeps runtime flags within safe bounds and is unit tested in `TestValidateConfig`.【F:backend/cmd/sim/main.go†L200-L213】【F:backend/cmd/sim/main_test.go†L8-L120】

## Authentication & Session Lifecycle

- **Join handshake** – `HandleJoin` authenticates tokens, spawns players, echoes engine config, issues resume tokens, and persists login counters when a store is configured.【F:backend/internal/join/join.go†L45-L88】【F:backend/internal/transport/ws/register_ws.go†L93-L209】 Automated coverage includes successful joins, auth failures, bad requests, and timeout handling via `TestHandleJoin_Success`, `TestHandleJoin_AuthFailure`, `TestHandleJoin_BadRequest`, and `TestHandleJoin_ContextTimeout`.【F:backend/internal/join/join_test.go†L31-L93】
- **HTTP gateway integration** – The gateway-style auth client enforces a 3s timeout and fails closed when the upstream is unreachable, covered by `TestHTTPAuth_ClientTimeout`.【F:backend/internal/join/auth_http.go†L11-L39】【F:backend/internal/join/join_test.go†L95-L118】 
- **Spawn persistence** – When a store is present, joins restore the last saved position and increment login counts; both the in-memory and file-backed implementations are validated by `TestHandleJoin_UsesSavedPosition` and `TestHandleJoin_WorksWithFileStore`.【F:backend/internal/join/join.go†L55-L87】【F:backend/internal/join/persist_test.go†L14-L115】
- **Resume tokens** – The WebSocket layer issues and validates short-lived resume tokens so reconnects can honor the last acknowledged sequence. Token management lives in `ResumeManager`, with behavior checked in `TestWS_ReconnectAndResume` and `TestResumeManager_Validate`.【F:backend/internal/transport/ws/session.go†L12-L64】【F:backend/internal/transport/ws/reconnect_test.go†L30-L165】

## Player State Persistence Stores

- **FileStore durability** – JSON-backed persistence supports load/save, directory creation, atomic flushes, periodic syncing, and graceful shutdown; unit tests cover each pathway through `TestFileStore_BasicOperations`, `TestFileStore_Persistence`, `TestFileStore_FlushBehavior`, `TestFileStore_GracefulShutdown`, `TestFileStore_DirectoryCreation`, and `TestFileStore_PeriodicFlush`.【F:backend/internal/state/store.go†L15-L200】【F:backend/internal/state/store_test.go†L13-L351】
- **In-memory store** – A thread-safe `MemStore` backs tests and development; `HandleJoin_UsesSavedPosition` exercises its behavior when wired into the join flow.【F:backend/internal/state/store.go†L28-L48】【F:backend/internal/join/persist_test.go†L14-L38】

## Simulation Engine

### Movement & Cell Ownership
- Player placement snaps to the correct grid cell via `AddOrUpdatePlayer`, and integration uses deterministic position updates. `TestIntegratesVelocityStep` and `TestAddOrUpdatePlacesPlayerInCorrectCell` cover these basics.【F:backend/internal/sim/engine.go†L300-L353】【F:backend/internal/sim/engine_test.go†L21-L51】

### Handover & Anti-Thrash
- Cell ownership changes enforce hysteresis, double-hysteresis when re-entering the previous cell, and timestamp capture for latency metrics.【F:backend/internal/sim/handovers.go†L11-L40】 The suite checks correctness, anti-thrash, and latency budgets through `TestHandoverAfterHysteresis`, `TestCrossedBeyondHysteresis`, `TestHandoverThrashPrevention`, `TestHandoverThrashingProblem`, `TestHandoverLatencyTimestampPrecision`, `TestHandoverLatencyAccuracy`, and `TestHandoverLatencyRequirement`, plus pacing continuity in `TestNoPacingThrashAndStateContinuity`.【F:backend/internal/sim/engine_test.go†L53-L83】【F:backend/internal/sim/handovers_test.go†L10-L334】【F:backend/internal/sim/handover_latency_test.go†L10-L130】

### Area-of-Interest (AOI) Streaming
- AOI queries gather a 3×3 neighborhood with a radius filter and exclude the requester.【F:backend/internal/sim/engine.go†L410-L443】 Coverage spans boundary inclusion, cross-border visibility, duplicate prevention, and rebuild timing via `TestAOI_InclusiveBoundaryAndExclusion`, `TestAOI_CoversAcrossBorder_NoFlap`, `TestAOI3x3CellQuery`, `TestContinuousAOIAcrossBorderWithStaticNeighbors`, `TestAOIRebuildTimingRequirement`, and `TestNoDuplicateEntityIDs`.【F:backend/internal/sim/aoi_test.go†L10-L73】【F:backend/internal/sim/continuous_aoi_test.go†L11-L298】

### Bot Population & Density Control
- Wander AI enforces separation, clamped speed, retarget timing, and deterministic RNG seeding.【F:backend/internal/sim/bots.go†L10-L127】 Tests `TestBotSeparation`, `TestBotWanderRetargetTiming`, `TestBotSpeedClamping`, `TestBotClusteringPrevention`, and `TestBotSeparationDeterministic` exercise these rules.【F:backend/internal/sim/bots_test.go†L13-L197】
- Density management keeps each cell within ±20% of its target while honoring a global cap and reacting to player churn.【F:backend/internal/sim/engine.go†L174-L235】 The scenarios in `TestDensityControllerBasicSpawn`, `TestDensityControllerBasicDespawn`, `TestDensityControllerGlobalBotCap`, `TestDensityControllerChurnScenario`, `TestDensityControllerSpawnDespawnBounds`, `TestDensityControllerTimingConvergence`, `TestDensityControllerRampingRate`, `TestDensityControllerZeroTarget`, and `TestDensityControllerNegativeMaxBots` verify those edges.【F:backend/internal/sim/density_test.go†L14-L247】【F:backend/internal/sim/density_test.go†L200-L371】

### Engine Lifecycle & Metrics
- The engine loop enforces tick/snapshot cadences, maintains AOI/handovers metrics, and exposes debug snapshots.【F:backend/internal/sim/engine.go†L70-L172】【F:backend/internal/sim/engine.go†L439-L458】 Lifecycle robustness is covered by `TestEngine_StartTwiceIsIdempotent`, `TestEngine_StopTwiceIsIdempotent`, `TestEngine_StopWithoutStartReturns`, and `TestSnapshotDebugLogging`.【F:backend/internal/sim/engine_test.go†L85-L172】

## WebSocket Transport

### Connection Setup, Auth, and Resume
- The WebSocket handler performs origin checks (with a dev-mode bypass), applies read limits, emits join acknowledgements with resume tokens, and manages idle disconnects while persisting last-known positions.【F:backend/internal/transport/ws/register_ws.go†L43-L209】 Origin handling is validated by the dev/production scenarios in `TestOriginValidation_DevMode`, `TestOriginValidation_ProductionMode`, `TestOriginValidation_ProductionMode_LocalhostAllowed`, and `TestOriginValidation_ProductionMode_SameOriginAllowed`.【F:backend/internal/transport/ws/origin_test.go†L19-L142】 Join timeout resilience is verified in `TestWS_JoinTimeout_HandlesAuthTimeout`.【F:backend/internal/transport/ws/ws_integration_test.go†L131-L195】
- Resume support is exercised by `TestWS_ReconnectAndResume`, while `TestResumeManager_Validate` checks token issuance and expiry semantics.【F:backend/internal/transport/ws/reconnect_test.go†L30-L165】

### Input/State Loop & AOI Continuity
- Client inputs are debounced through a reader goroutine and reflected in periodic state payloads that include AOI entities and handover events.【F:backend/internal/transport/ws/register_ws.go†L210-L266】 `TestWS_InputState_AckAndMotion` confirms motion acknowledgment, `TestWS_HandoverEvent_EmittedOnCellChange` ensures handover events precede ownership changes, and `TestWS_HandoverAntiThrash_WithWebSocketContinuity` covers anti-thrash behavior over the transport.【F:backend/internal/transport/ws/ws_integration_test.go†L31-L129】【F:backend/internal/transport/ws/handover_test.go†L30-L234】 AOI continuity after handovers is guarded by `TestWS_AOIContinuity_AcrossHandover`.【F:backend/internal/transport/ws/aoi_continuity_test.go†L30-L140】

### Snapshot Cadence, Telemetry, and Metrics
- Snapshot payload sizes and cadence feed Prometheus histograms, while RTT telemetry is delivered at 1 Hz.【F:backend/internal/transport/ws/register_ws.go†L217-L286】【F:backend/internal/metrics/metrics.go†L12-L132】 Tests `TestSnapshotCadenceAndPayloadBudget`, `TestTelemetry_TickAndRTT`, and `TestMetrics_WsConnectedGauge` validate cadence, telemetry content, and gauge wiring respectively.【F:backend/internal/transport/ws/cadence_test.go†L29-L99】【F:backend/internal/transport/ws/telemetry_test.go†L29-L101】【F:backend/internal/transport/ws/metrics_test.go†L32-L87】

### Security & Resource Management
- The transport enforces message size caps, per-message read deadlines, and idle timeouts.【F:backend/internal/transport/ws/register_ws.go†L83-L209】 Automated coverage appears in `TestWS_OversizedMessageRejection`, `TestWS_ReadDeadline`, and `TestWS_IdleTimeout`.【F:backend/internal/transport/ws/security_test.go†L20-L197】

## Metrics & Observability

The metrics package registers histograms and gauges for tick timing, snapshot sizes, AOI load, handover latency, and WebSocket connections, with the handler wired into the sim service.【F:backend/internal/metrics/metrics.go†L12-L132】【F:backend/cmd/sim/main.go†L85-L111】 End-to-end verification that the `ws_connected` gauge reflects active clients is provided by `TestMetrics_WsConnectedGauge`.【F:backend/internal/transport/ws/metrics_test.go†L32-L87】

## Running the Test Suites

Run the core backend suite (no build tags required):

```bash
go test ./...
```

WebSocket transport tests require the `ws` build tag:

```bash
go test -tags ws ./backend/internal/transport/ws
```

Together these commands execute every scenario referenced above.
