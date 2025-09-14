# Prototype Game Wiki

Welcome to the comprehensive documentation for Prototype Game, a multiplayer game backend with seamless local sharding.

## Quick Navigation

### 🚀 [Getting Started](Getting-Started)
New to the project? Start here for setup and first steps.

### 🛠️ [Development Guide](Development-Guide)
Daily development workflows, build system, and testing.

### 🏗️ [Architecture & Design](Architecture-&-Design)
Technical design, game vision, and system architecture.

### 🤝 [Contributing](Contributing)
How to contribute, workflow guidelines, and standards.

### 🤖 [Agent & Automation](Agent-&-Automation)
AI assistant setup and automation workflows.

### 📚 [API Reference](API-Reference)
Complete API documentation and protocol specifications.

## Project Overview

Prototype Game is a Go-based multiplayer game backend featuring:
- Seamless local sharding for scalable multiplayer
- WebSocket-based real-time communication
- Server-authoritative simulation with client prediction
- Spatial partitioning and Area of Interest (AOI) management

## Current Status

- ✅ Core simulation engine with spatial math
- ✅ WebSocket transport layer
- ✅ Authentication and session management
- ✅ Local sharding implementation
- 🚧 Multi-node sharding (planned)
- 🚧 Client implementation (planned)

## Quick Start

```bash
# Clone and build
git clone https://github.com/AstroSteveo/prototype-game.git
cd prototype-game
make run

# Test the setup
make login
TOKEN=$(make login) && make wsprobe TOKEN="$TOKEN"
```

For detailed setup instructions, see [Getting Started](Getting-Started).

## Community and Support

- 📁 [Repository](https://github.com/AstroSteveo/prototype-game)
- 🐛 [Issues](https://github.com/AstroSteveo/prototype-game/issues)
- 💬 [Discussions](https://github.com/AstroSteveo/prototype-game/discussions)

---
*This wiki is automatically synchronized with the repository documentation.*
