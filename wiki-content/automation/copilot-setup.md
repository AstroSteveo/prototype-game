# Prototype Game - GitHub Copilot Instructions

**ALWAYS follow these instructions first and only fallback to additional search and context gathering if the information here is incomplete or found to be in error.**

## Project Overview

Prototype Game is a Go-based multiplayer game backend with WebSocket support. The project consists of:
- **Gateway service**: Authentication and player management (port 8080)
- **Sim service**: Game simulation engine with WebSocket transport (port 8081) 
- **Client**: Placeholder for game client (future implementation)

The codebase is organized with Go 1.23+ backend services and comprehensive testing infrastructure.

## Prerequisites

- Go 1.23+ (project uses Go 1.23.0 minimum)
- curl (for HTTP health checks and API testing)
- Python 3 (for JSON parsing in make targets)
- No additional SDKs or complex dependencies required

## Building and Testing

### Build Commands

**NEVER CANCEL BUILDS** - All commands below include appropriate timeout guidance.

Build all binaries (gateway, sim with WebSocket, wsprobe):
```bash
make build
```
- **Time**: ~33 seconds first build (with dependency downloads), ~1.3 seconds subsequent builds
- **Timeout**: Set 120+ seconds for first build, 60+ seconds for subsequent builds
- **Output**: Binaries created in `backend/bin/` directory

Format and validate code:
```bash
make fmt vet
```
- **Time**: ~6.7 seconds
- **Timeout**: 30+ seconds
- **CRITICAL**: Always run before committing - CI will fail without proper formatting

### Test Commands

Run unit tests:
```bash
make test
```
- **Time**: ~6.3 seconds
- **Timeout**: 60+ seconds
- **Scope**: Tests packages under `backend/internal/`

Run WebSocket integration tests:
```bash
make test-ws
```
- **Time**: ~9.2 seconds 
- **Timeout**: 60+ seconds
- **Scope**: Includes WebSocket transport tests requiring `-tags ws` build flag

Run both format, vet, and all tests (CI validation):
```bash
make fmt vet test test-ws
```
- **Time**: ~22 seconds total (or ~0.7 seconds if cached)
- **Timeout**: 120+ seconds
- **CRITICAL**: Always run this exact sequence before pushing - matches CI requirements

## Running the Application

### Quick Start (Recommended)

Start both services in background:
```bash
make run
```
- **Time**: ~0.3 seconds after build
- **Timeout**: 60+ seconds
- **Result**: Gateway on :8080, Sim on :8081
- **Logs**: `backend/logs/gateway.log`, `backend/logs/sim.log`
- **PIDs**: `backend/.pids/gateway.pid`, `backend/.pids/sim.pid`

Stop services:
```bash
make stop
```

### Manual Service Management

Run gateway manually:
```bash
cd backend && go run ./cmd/gateway -port 8080 -sim localhost:8081
```

Run sim with WebSocket support:
```bash
cd backend && go run -tags ws ./cmd/sim -port 8081
```

**CRITICAL**: Sim requires `-tags ws` build flag for WebSocket functionality, otherwise `/ws` endpoint returns 501.

### Health Validation

Verify services are running:
```bash
curl http://localhost:8080/healthz  # Should return "ok"
curl http://localhost:8081/healthz  # Should return "ok"
```

## User Scenarios and Validation

### CRITICAL: Always Test These Scenarios After Changes

After making any code changes, **ALWAYS** run through these complete validation scenarios:

#### Basic Authentication Flow
```bash
# 1. Start services
make run

# 2. Get authentication token
TOKEN=$(make login)
echo "Got token: $TOKEN"

# 3. Validate token works
curl "http://localhost:8080/validate?token=$TOKEN"
```

#### WebSocket Join Scenario
```bash
# After getting token from above
make wsprobe TOKEN="$TOKEN"
```
**Expected output**: JSON with `join_ack` message containing player_id, position, cell, and config

#### WebSocket Movement Scenario  
```bash
# Test player movement and state updates
make wsprobe TOKEN="$TOKEN" MOVE_X=1 MOVE_Z=0
```
**Expected output**: 
- `join_ack` message
- `state` message showing player movement with updated position and velocity

#### Complete E2E Scenarios
```bash
# Automated join test (builds, runs, tests, stops)
make e2e-join

# Automated movement test
make e2e-move
```
- **Time**: ~0.4-0.5 seconds each
- **Timeout**: 120+ seconds

### Monitoring and Debugging

Check simulation metrics:
```bash
curl http://localhost:8081/metrics.json
```
**Expected fields**: `handovers`, `aoi_queries`, `aoi_entities_total`, `aoi_avg_entities`

View service logs:
```bash
tail -f backend/logs/gateway.log
tail -f backend/logs/sim.log
```

## Repository Navigation

### Critical Files and Locations

**Always reference AGENTS.md first** - it contains authoritative agent instructions that override other documentation.

Key documentation (read these when onboarding):
- `AGENTS.md` - Authoritative agent/automation instructions
- `README.md` - Quick start guide
- `docs/dev/DEV.md` - Detailed developer workflows
- `docs/product/vision/game-design-document.md` - Game Design Document
- `docs/architecture/technical-design-document.md` - Technical Design Document

Build and CI:
- `Makefile` - **Use make targets instead of raw commands to avoid drift**
- `.github/workflows/ci.yml` - CI pipeline (runs fmt, vet, test, test-ws)
- `backend/go.mod` - Go dependencies

Service entry points:
- `backend/cmd/gateway/` - Authentication service
- `backend/cmd/sim/` - Game simulation service  
- `backend/cmd/wsprobe/` - WebSocket testing utility

Core libraries:
- `backend/internal/sim/` - Game engine and simulation logic
- `backend/internal/spatial/` - Spatial mathematics and cell calculations
- `backend/internal/join/` - Authentication and join logic
- `backend/internal/transport/ws/` - WebSocket transport (requires `ws` build tag)
- `backend/internal/metrics/` - Prometheus metrics

### Common File Patterns

When making changes to specific areas:

**Authentication/Join Logic**: Always check both:
- `backend/internal/join/` - Core join logic
- `backend/cmd/gateway/` - HTTP endpoints

**Game Simulation**: Always check:
- `backend/internal/sim/` - Engine logic
- `backend/internal/spatial/` - Math calculations  
- After changes, run movement validation scenarios

**WebSocket Transport**: Always check:
- `backend/internal/transport/ws/` - WebSocket handlers
- Build and test with `-tags ws` flag
- Test with `make wsprobe` scenarios

## Port Configuration

**Default ports** (configurable via variables):
- Gateway: 8080 (`GATEWAY_PORT`)
- Sim: 8081 (`SIM_PORT`)

Override ports:
```bash
make run GATEWAY_PORT=9080 SIM_PORT=9081
```

## Build Tags and Special Considerations

**WebSocket functionality** requires `ws` build tag:
- Correct: `go run -tags ws ./cmd/sim`
- Incorrect: `go run ./cmd/sim` (WebSocket endpoint returns 501)

**Test execution**:
- Unit tests: `go test ./...` (standard tests)
- WebSocket tests: `go test -tags ws ./...` (includes WebSocket integration)

## Troubleshooting Common Issues

**"Address already in use"**:
```bash
make stop  # Stop background services
```

**"/ws returns 501 Not Implemented"**:
- Ensure sim is built/run with `-tags ws` flag
- Use `make run` instead of manual `go run` commands

**Authentication errors in WebSocket**:
- Get fresh token: `TOKEN=$(make login)`
- Tokens are short-lived dev tokens

**No state messages after input**:
- Server broadcasts state at ~10Hz, wait up to 200ms
- Check logs: `tail backend/logs/sim.log`

## Development Workflow

### Making Changes
1. **Start with validation**: Run `make fmt vet test test-ws` to establish baseline
2. **Make minimal changes** to achieve your goal
3. **Test immediately**: Run relevant test suites after each change
4. **Validate scenarios**: Run appropriate user scenarios from the validation section
5. **Final check**: Run complete CI sequence: `make fmt vet test test-ws`

### Before Committing
**MANDATORY CHECKLIST**:
- [ ] `make fmt vet` passes (CI will fail otherwise)
- [ ] `make test test-ws` passes 
- [ ] At least one complete user scenario validated
- [ ] WebSocket functionality tested if transport changes made
- [ ] No build artifacts committed (check `.gitignore`)

### Commit Guidelines
- Use imperative subjects: "sim: fix handover hysteresis"
- Reference user story IDs when applicable: "US-201"
- Keep diffs focused and under 300 lines when practical
- Update relevant docs in `docs/process/` for behavior changes

## Command Reference

**Most Frequently Used Commands**:
```bash
make help                           # Show all available targets
make run                           # Start services (most common)
make stop                          # Stop services  
make login                         # Get dev token
TOKEN=$(make login) && make wsprobe TOKEN="$TOKEN"  # Quick WebSocket test
make fmt vet test test-ws          # Full CI validation
make clean                         # Remove build artifacts and logs
```

**Timing Expectations**:
- Initial build: ~33 seconds (with dependency downloads)
- Subsequent builds: ~1.3 seconds  
- Format + vet: ~6.7 seconds
- Unit tests: ~6.3 seconds
- WebSocket tests: ~9.2 seconds
- All tests: ~22 seconds (first run) or ~0.7 seconds (cached)
- Service startup: ~0.3 seconds
- E2E scenarios: ~0.4-0.5 seconds each

**NEVER CANCEL any of these operations** - they are designed to complete quickly and canceling may leave the system in an inconsistent state.