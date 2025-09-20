# Getting Started with Prototype Game Backend

## Prerequisites

Before you begin developing with the Prototype Game Backend, ensure you have the following tools and knowledge:

### Required Software

- **Go 1.21 or later**: Download from [golang.org](https://golang.org/downloads/)
- **Git**: For version control and dependency management
- **Make**: Build automation (included on most Unix systems, available for Windows)
- **A modern IDE**: VS Code with Go extension, GoLand, or similar

### Required Knowledge

- **Basic Go Programming**: Understanding of Go syntax, goroutines, and channels
- **WebSocket Concepts**: Familiarity with real-time web communication
- **Command Line Operations**: Comfortable with terminal/command prompt usage

### Optional Tools

- **Docker**: For containerized development (advanced setup)
- **PostgreSQL**: If using persistent storage features
- **Wireshark**: For network protocol analysis and debugging

## Project Setup

### 1. Clone the Repository

```bash
# Clone the project
git clone <repository-url>
cd prototype-game

# Verify the project structure
ls -la
```

You should see the following key directories:
```
├── cmd/           # Executable entry points
├── internal/      # Internal packages
├── pkg/           # Public packages
├── docs/          # Documentation
├── Makefile       # Build automation
└── go.mod         # Go module definition
```

### 2. Install Dependencies

```bash
# Download and install all dependencies
go mod download

# Verify dependencies are correctly installed
go mod verify
```

### 3. Verify Go Environment

```bash
# Check Go version
go version

# Verify Go environment
go env GOPATH
go env GOROOT
```

**Troubleshooting**: If you encounter module resolution issues, ensure your `GOPROXY` is correctly configured:
```bash
go env GOPROXY
# Should show: https://proxy.golang.org,direct
```

## Building the Services

The Prototype Game Backend consists of several key services that work together. We'll build each one step by step.

### 1. Build All Services

Use the provided Makefile for streamlined building:

```bash
# Build all services at once
make build

# This creates executables in the ./bin/ directory
ls -la bin/
```

Expected output:
```
-rwxr-xr-x gateway
-rwxr-xr-x simulation
-rwxr-xr-x probe
```

### 2. Build Individual Services

If you prefer to build services individually or encounter issues:

```bash
# Build the Gateway service
go build -o bin/gateway ./cmd/gateway

# Build the Simulation Engine
go build -o bin/simulation ./cmd/simulation

# Build the Probe tool (debugging utility)
go build -o bin/probe ./cmd/probe
```

### 3. Development Build

For development with debugging symbols and without optimizations:

```bash
# Development build with race detection
make build-dev

# Or manually:
go build -race -gcflags="all=-N -l" -o bin/gateway-dev ./cmd/gateway
```

**Note**: Development builds are larger and slower but provide better debugging information.

## Running Your First Game Session

Now let's start the services and establish a basic game session.

### 1. Start the Simulation Engine

The simulation engine must be running before clients can connect:

```bash
# Start the simulation engine with default settings
./bin/simulation

# Or with custom configuration
./bin/simulation -bind-addr=:8081 -cell-size=1000
```

**Expected Output**:
```
2024/01/15 10:30:00 INFO Simulation engine starting
2024/01/15 10:30:00 INFO Spatial system initialized with cell size: 500
2024/01/15 10:30:00 INFO Server listening on :8080
2024/01/15 10:30:00 INFO Simulation engine ready
```

**Keep this terminal open** - the simulation engine needs to stay running.

### 2. Start the Gateway Service

In a **new terminal window**, start the Gateway:

```bash
# Navigate to project directory
cd prototype-game

# Start the Gateway service
./bin/gateway

# Or with custom settings
./bin/gateway -bind-addr=:9090 -sim-addr=localhost:8080
```

**Expected Output**:
```
2024/01/15 10:31:00 INFO Gateway service starting
2024/01/15 10:31:00 INFO Connected to simulation engine at localhost:8080
2024/01/15 10:31:00 INFO WebSocket server listening on :9090
2024/01/15 10:31:00 INFO Gateway ready for connections
```

### 3. Test Connection with Probe Tool

Use the built-in probe tool to verify everything is working:

```bash
# In a third terminal window
./bin/probe -gateway=localhost:9090

# Expected successful connection output:
# 2024/01/15 10:32:00 INFO Connected to gateway
# 2024/01/15 10:32:00 INFO Session established
# 2024/01/15 10:32:00 INFO Player spawned at position (0, 0)
```

**Success!** You now have a running game backend with a test client connected.

## Verifying Everything Works

Let's perform comprehensive verification to ensure all components are functioning correctly.

### 1. Service Health Checks

Check that all services are responding to health checks:

```bash
# Check simulation engine health
curl http://localhost:8080/health
# Expected: {"status":"healthy","uptime":"2m30s"}

# Check gateway health  
curl http://localhost:9090/health
# Expected: {"status":"healthy","simulation":"connected"}
```

### 2. WebSocket Connection Test

Test WebSocket connectivity using a simple script or browser developer tools:

```javascript
// In browser console or Node.js
const ws = new WebSocket('ws://localhost:9090/ws');
ws.onopen = () => console.log('Connected successfully');
ws.onmessage = (msg) => console.log('Received:', msg.data);
ws.onerror = (err) => console.error('WebSocket error:', err);
```

### 3. Multi-Client Test

Test with multiple simultaneous connections:

```bash
# Start multiple probe instances
./bin/probe -gateway=localhost:9090 -player-id=player1 &
./bin/probe -gateway=localhost:9090 -player-id=player2 &
./bin/probe -gateway=localhost:9090 -player-id=player3 &

# Wait a moment, then check connections
jobs
# Should show three running probe processes
```

### 4. Performance Verification

Verify system performance under basic load:

```bash
# Check CPU and memory usage
top -p $(pgrep -f "simulation|gateway")

# Monitor connection count
curl http://localhost:9090/metrics | grep connection_count
```

## Development Workflow

Now that you have a working setup, here's the recommended development workflow:

### 1. Code Changes

When making code changes:

```bash
# Make your changes to source files
# Then rebuild only what's needed

# Quick rebuild for testing
make build

# Or rebuild specific service
go build -o bin/gateway ./cmd/gateway
```

### 2. Testing Changes

```bash
# Stop running services (Ctrl+C)
# Restart affected services
./bin/simulation &
./bin/gateway &

# Test with probe
./bin/probe -gateway=localhost:9090
```

### 3. Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/spatial/...
```

### 4. Code Formatting and Linting

```bash
# Format code
make fmt

# Run linter
make lint

# Or manually:
gofmt -w .
golangci-lint run
```

## Common Configuration Options

### Simulation Engine Configuration

```bash
# Common flags for simulation engine
./bin/simulation \
  -bind-addr=:8080 \        # Server bind address
  -cell-size=500 \          # Spatial cell size in units
  -tick-rate=20 \           # Updates per second
  -max-entities=10000 \     # Maximum entities per shard
  -log-level=info           # Logging verbosity
```

### Gateway Configuration

```bash
# Common flags for gateway
./bin/gateway \
  -bind-addr=:9090 \        # WebSocket server address
  -sim-addr=localhost:8080 \ # Simulation engine address
  -max-connections=1000 \   # Maximum concurrent connections
  -heartbeat-interval=30s \ # Connection heartbeat interval
  -log-level=info           # Logging verbosity
```

### Environment Variables

Create a `.env` file for development settings:

```bash
# .env file
GATEWAY_BIND_ADDR=:9090
SIMULATION_BIND_ADDR=:8080
LOG_LEVEL=debug
MAX_CONNECTIONS=1000
SPATIAL_CELL_SIZE=500
```

Load with:
```bash
source .env
./bin/gateway
```

## Troubleshooting

### Common Issues and Solutions

#### "Connection Refused" Errors

**Problem**: Gateway can't connect to simulation engine
```
ERROR Failed to connect to simulation engine: connection refused
```

**Solution**:
1. Ensure simulation engine is running first
2. Check the addresses match (`-sim-addr` vs `-bind-addr`)
3. Verify no firewall is blocking the ports

#### "Port Already in Use" Errors

**Problem**: Service can't bind to port
```
ERROR bind: address already in use
```

**Solution**:
```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use different port
./bin/simulation -bind-addr=:8081
```

#### Build Failures

**Problem**: Go build fails with dependency errors
```
ERROR module not found
```

**Solution**:
```bash
# Clean module cache
go clean -modcache

# Reinstall dependencies
go mod download
go mod tidy

# Rebuild
make build
```

#### WebSocket Connection Issues

**Problem**: Probe tool can't connect
```
ERROR WebSocket connection failed
```

**Solution**:
1. Check gateway is running and listening
2. Verify correct WebSocket URL format: `ws://localhost:9090/ws`
3. Check for proxy or firewall interference

### Development Tips

1. **Use Logging**: Enable debug logging for detailed information
   ```bash
   ./bin/simulation -log-level=debug
   ```

2. **Monitor Resources**: Keep an eye on CPU and memory usage during development
   ```bash
   watch -n 1 'ps aux | grep -E "(simulation|gateway)"'
   ```

3. **Code Hot Reloading**: Use tools like `air` for automatic rebuilding:
   ```bash
   go install github.com/cosmtrek/air@latest
   air
   ```

4. **Database Integration**: When ready for persistence:
   ```bash
   # Start PostgreSQL
   docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15

   # Connect to database
   ./bin/simulation -db-url="postgres://postgres:password@localhost/gamedb"
   ```

## Next Steps

Congratulations! You now have a fully functional Prototype Game Backend. Here's what you can explore next:

### Immediate Next Steps

1. **Explore the API**: Review the API reference documentation to understand available endpoints
2. **Client Development**: Build a simple web client to interact with your game backend
3. **Custom Game Logic**: Modify the simulation engine to implement your specific game rules
4. **Spatial Features**: Experiment with entity movement and area-of-interest updates

### Advanced Topics

1. **Performance Tuning**: Optimize for higher player counts and larger worlds
2. **Persistence Integration**: Add database storage for player data and world state
3. **Clustering**: Set up multiple simulation shards for horizontal scaling
4. **Monitoring**: Implement comprehensive metrics and alerting

### Learning Resources

- **Code Examples**: Check the `examples/` directory for sample implementations
- **API Documentation**: Review the complete API reference (next section)
- **Architecture Guide**: Deep dive into system design principles
- **Community**: Join the development community for questions and contributions

You're now ready to build amazing multiplayer game experiences with the Prototype Game Backend!