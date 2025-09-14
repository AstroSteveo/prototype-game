#!/bin/bash
# extract-content.sh
# Extracts and processes content from existing markdown files for wiki migration

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
WIKI_DIR="$PROJECT_ROOT/wiki-content"

echo "=== Content Extraction for Wiki Migration ==="
echo "Project root: $PROJECT_ROOT"
echo "Wiki content dir: $WIKI_DIR"
echo ""

# Ensure wiki content directory exists
mkdir -p "$WIKI_DIR"/{getting-started,development,architecture,contributing,automation,api}

extract_section() {
    local file="$1"
    local start_pattern="$2"
    local end_pattern="$3"
    local output_file="$4"
    
    if [[ ! -f "$file" ]]; then
        echo "⚠️  Warning: Source file $file not found"
        return 1
    fi
    
    if [[ -n "$end_pattern" ]]; then
        awk "/$start_pattern/,/$end_pattern/" "$file" > "$output_file"
    else
        awk "/$start_pattern/,EOF" "$file" > "$output_file"
    fi
    
    echo "✅ Extracted from $file -> $output_file"
}

extract_full_file() {
    local file="$1"
    local output_file="$2"
    
    if [[ ! -f "$file" ]]; then
        echo "⚠️  Warning: Source file $file not found"
        return 1
    fi
    
    cp "$file" "$output_file"
    echo "✅ Copied $file -> $output_file"
}

echo "Extracting content from existing documentation..."

# Extract Home page content (from README)
echo "📖 Extracting Home page content..."
extract_section "$PROJECT_ROOT/README.md" "^# prototype-game" "^Notes:" "$WIKI_DIR/home.md"

# Extract Getting Started content
echo "🚀 Extracting Getting Started content..."
extract_section "$PROJECT_ROOT/README.md" "^Quick start" "^WebSocket" "$WIKI_DIR/getting-started/quick-start.md"
extract_section "$PROJECT_ROOT/docs/dev/DEV.md" "^## Prerequisites" "^## Quick Start" "$WIKI_DIR/getting-started/prerequisites.md"

# Extract Development Guide content
echo "🛠️ Extracting Development Guide content..."
extract_full_file "$PROJECT_ROOT/docs/dev/DEV.md" "$WIKI_DIR/development/dev-guide-full.md"
extract_section "$PROJECT_ROOT/docs/dev/DEV.md" "^## Quick Start" "^## Reconnect" "$WIKI_DIR/development/build-system.md"
extract_section "$PROJECT_ROOT/docs/dev/DEV.md" "^### Tests" "^### Ports" "$WIKI_DIR/development/testing.md"

# Extract Architecture & Design content
echo "🏗️ Extracting Architecture & Design content..."
extract_full_file "$PROJECT_ROOT/docs/design/GDD.md" "$WIKI_DIR/architecture/game-design-document.md"
extract_full_file "$PROJECT_ROOT/docs/design/TDD.md" "$WIKI_DIR/architecture/technical-architecture.md"

# Extract Contributing content
echo "🤝 Extracting Contributing content..."
extract_full_file "$PROJECT_ROOT/.github/CONTRIBUTING.md" "$WIKI_DIR/contributing/contributing-full.md"
extract_section "$PROJECT_ROOT/AGENTS.md" "^## Commit & Pull Request Guidelines" "^## Security" "$WIKI_DIR/contributing/workflow.md"

# Extract Agent & Automation content
echo "🤖 Extracting Agent & Automation content..."
extract_full_file "$PROJECT_ROOT/AGENTS.md" "$WIKI_DIR/automation/agent-instructions.md"
extract_full_file "$PROJECT_ROOT/.github/copilot-instructions.md" "$WIKI_DIR/automation/copilot-setup.md"

# Extract API and protocol content
echo "📚 Extracting API content..."
extract_section "$PROJECT_ROOT/docs/design/TDD.md" "^## Protocol" "^## Persistence" "$WIKI_DIR/api/websocket-protocol.md" || true

# Create some additional processed content
echo ""
echo "Creating processed content..."

# Create comprehensive home page
cat > "$WIKI_DIR/home-processed.md" << 'EOF'
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
EOF

# Create section landing pages
cat > "$WIKI_DIR/getting-started.md" << 'EOF'
# Getting Started

Welcome to Prototype Game! This section will help you set up your development environment and get the project running locally.

## Pages in this Section

- **[Welcome](Welcome)** - Project introduction and core concepts
- **[Quick Start](Quick-Start)** - Fast setup and first run
- **[Prerequisites](Prerequisites)** - System requirements and dependencies  
- **[Installation](Installation)** - Detailed setup instructions

## Quick Links

- [Build and run services](Quick-Start#build-and-run)
- [Test WebSocket connection](Quick-Start#test-websocket)
- [Common troubleshooting](Troubleshooting)

## Prerequisites

- Go 1.23+
- curl (for HTTP health checks)
- Basic command line familiarity

---
📖 **Navigation**: [Home](Home) → Getting Started
EOF

echo "✅ Content extraction complete!"
echo ""
echo "Extracted content structure:"
find "$WIKI_DIR" -type f -name "*.md" | sort

echo ""
echo "Next steps:"
echo "1. Review extracted content in $WIKI_DIR/"
echo "2. Edit and format content for wiki presentation"
echo "3. Create wiki pages manually or use GitHub API"
echo "4. Run ./update-links.sh to update repository links"