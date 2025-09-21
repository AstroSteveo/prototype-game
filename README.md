# Prototype Game Backend

A high-performance, scalable multiplayer game backend built with Go, featuring real-time WebSocket communication, spatial partitioning, and distributed simulation architecture.

## 🎮 Features

- **Real-time Multiplayer**: WebSocket-based communication for low-latency gaming
- **Spatial Optimization**: Cell-based partitioning for efficient area-of-interest management  
- **Horizontal Scaling**: Distributed simulation engine with sharding support
- **Performance Monitoring**: Built-in Prometheus metrics and health checks
- **Developer Tools**: Comprehensive testing and debugging utilities

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download here](https://golang.org/downloads/)
- **Make** - Build automation
- **Git** - Version control

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd prototype-game

# Build all services
make build

# Start the game backend
make run
```

The services will be available at:
- **Gateway**: http://localhost:8080 (WebSocket endpoint)
- **Simulation**: http://localhost:8081 (Internal service)

### Test Connection

```bash
# Get a development token
TOKEN=$(make login)

# Test WebSocket connection
make wsprobe TOKEN=$TOKEN
```

## 📁 Project Structure

```
├── backend/                 # Go backend services
│   ├── cmd/                # Service entry points
│   │   ├── gateway/        # WebSocket gateway server
│   │   ├── sim/           # Simulation engine
│   │   └── wsprobe/       # Testing utility
│   ├── internal/          # Private packages
│   │   ├── spatial/       # Spatial partitioning system
│   │   ├── sim/          # Game simulation logic
│   │   ├── join/         # Player connection handling
│   │   └── metrics/      # Performance monitoring
│   └── go.mod            # Go dependencies
├── docs/                  # Documentation
│   ├── getting-started.md # Detailed setup guide
│   ├── architecture.md   # System design
│   └── api-reference.md  # API documentation
├── Makefile              # Build automation
└── README.md            # This file
```

## 🛠️ Development

### Available Commands

```bash
# Development
make run              # Start all services
make stop             # Stop all services
make build            # Build binaries
make test             # Run tests
make test-ws          # Run WebSocket integration tests

# Code Quality
make fmt              # Format Go code
make fmt-check        # Check formatting
make vet              # Run Go vet

# Testing & Debugging
make login            # Get development token
make wsprobe TOKEN=x  # Test WebSocket connection
```

### Development Workflow

1. **Make Changes**: Edit source code in `backend/`
2. **Build**: Run `make build` to compile
3. **Test**: Use `make test` for unit tests
4. **Integration Test**: Use `make run` + `make wsprobe` for end-to-end testing

### Service Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Game Client   │────│    Gateway       │────│  Simulation     │
│   (WebSocket)   │    │   (Auth/Route)   │    │   Engine        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

- **Gateway**: Handles client connections, authentication, and message routing
- **Simulation**: Manages game state, entity updates, and spatial calculations
- **WebSocket**: Real-time bidirectional communication protocol

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Detailed setup and first steps |
| [Architecture](docs/architecture.md) | System design and principles |
| [API Reference](docs/api-reference.md) | Complete API documentation |

## 🧪 Testing

```bash
# Unit tests
make test

# WebSocket integration tests  
make test-ws

# Race condition detection
make test-race
make test-ws-race

# End-to-end testing
make e2e-join    # Test player connection
make e2e-move    # Test player movement
```

## 📊 Monitoring

### Health Checks

```bash
# Check service health
curl http://localhost:8080/healthz  # Gateway
curl http://localhost:8081/healthz  # Simulation
```

### Metrics

Prometheus metrics are available at:
- Gateway: `http://localhost:8080/metrics`
- Simulation: `http://localhost:8081/metrics`

## 🔧 Configuration

### Environment Variables

```bash
# Service ports
GATEWAY_PORT=8080
SIM_PORT=8081

# Simulation settings
SPATIAL_CELL_SIZE=500
TICK_RATE=20
MAX_ENTITIES=10000

# Logging
LOG_LEVEL=info
```

### Command Line Options

```bash
# Gateway
./bin/gateway -port 8080 -sim localhost:8081

# Simulation
./bin/sim -port 8081 -cell-size 500 -tick-rate 20
```

## 🚨 Troubleshooting

### Common Issues

**Port already in use**
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

**Connection refused**
```bash
# Ensure simulation is running first
make run-sim
# Then start gateway
make run-gateway
```

**Build failures**
```bash
# Clean and rebuild
go clean -modcache
go mod download
make build
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Development Guidelines

- Follow Go conventions and `gofmt` formatting
- Add tests for new functionality
- Update documentation for API changes
- Ensure all tests pass before submitting

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [Client SDK](client/) - JavaScript/TypeScript client library
- [Admin Tools](admin/) - Game management utilities
- [Performance Tests](benchmarks/) - Load testing suite

---

**Built with ❤️ for the multiplayer gaming community**