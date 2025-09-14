#!/bin/bash
# update-links.sh
# Updates internal links in repository files to point to wiki pages

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
WIKI_BASE_URL="https://github.com/AstroSteveo/prototype-game/wiki"

echo "=== Repository Link Updater for Wiki Migration ==="
echo "Project root: $PROJECT_ROOT"
echo "Wiki base URL: $WIKI_BASE_URL"
echo ""

# Function to update links in a file
update_repo_links() {
    local file="$1"
    local backup_file="${file}.backup"
    
    if [[ ! -f "$file" ]]; then
        echo "⚠️  Warning: File $file not found"
        return 1
    fi
    
    echo "📝 Updating links in $file..."
    
    # Create backup
    cp "$file" "$backup_file"
    
    # Update links to design docs
    sed -i "s|docs/design/GDD\.md|$WIKI_BASE_URL/Game-Design-Document|g" "$file"
    sed -i "s|docs/design/TDD\.md|$WIKI_BASE_URL/Technical-Architecture|g" "$file"
    
    # Update links to dev docs
    sed -i "s|docs/dev/DEV\.md|$WIKI_BASE_URL/Development-Guide|g" "$file"
    
    # Update links to contributing docs
    sed -i "s|\.github/CONTRIBUTING\.md|$WIKI_BASE_URL/Contributing|g" "$file"
    sed -i "s|\.github/copilot-instructions\.md|$WIKI_BASE_URL/Copilot-Setup|g" "$file"
    
    # Update relative links to absolute wiki links
    sed -i "s|](docs/design/|]($WIKI_BASE_URL/|g" "$file"
    sed -i "s|](docs/dev/|]($WIKI_BASE_URL/|g" "$file"
    sed -i "s|](\.github/|]($WIKI_BASE_URL/|g" "$file"
    
    # Update specific content references
    sed -i "s|- \`docs/design/GDD\.md\` — Game Design Document|- [$WIKI_BASE_URL/Game-Design-Document](Game Design Document)|g" "$file"
    sed -i "s|- \`docs/design/TDD\.md\` — Technical Design Document|- [$WIKI_BASE_URL/Technical-Architecture](Technical Design Document)|g" "$file"
    sed -i "s|- \`docs/dev/DEV\.md\` — Developer Guide|- [$WIKI_BASE_URL/Development-Guide](Developer Guide)|g" "$file"
    
    echo "✅ Updated $file (backup saved as $backup_file)"
}

# Function to create a simplified README that links to wiki
create_simplified_readme() {
    local readme_file="$PROJECT_ROOT/README.md"
    local backup_file="${readme_file}.pre-wiki-backup"
    
    echo "📝 Creating simplified README with wiki links..."
    
    # Create backup
    cp "$readme_file" "$backup_file"
    
    cat > "$readme_file" << EOF
# Prototype Game

A Go-based multiplayer game backend with seamless local sharding and WebSocket support.

## 🚀 Quick Start

\`\`\`bash
# Clone and build
git clone https://github.com/AstroSteveo/prototype-game.git
cd prototype-game
make run

# Test the setup
make login
TOKEN=\$(make login) && make wsprobe TOKEN="\$TOKEN"
\`\`\`

## 📚 Documentation

**Complete documentation is available in our [Wiki]($WIKI_BASE_URL)**

### Key Resources
- 🏁 [Getting Started]($WIKI_BASE_URL/Getting-Started) - Setup and first steps
- 🛠️ [Development Guide]($WIKI_BASE_URL/Development-Guide) - Build, test, and daily workflows
- 🏗️ [Architecture & Design]($WIKI_BASE_URL/Architecture-&-Design) - Technical design and game vision
- 🤝 [Contributing]($WIKI_BASE_URL/Contributing) - How to contribute and workflow guidelines
- 🤖 [Agent & Automation]($WIKI_BASE_URL/Agent-&-Automation) - AI assistant setup and automation

### Essential Local Development
- **Build**: \`make build\` - Builds all binaries
- **Test**: \`make fmt vet test test-ws\` - Complete validation
- **Run**: \`make run\` - Start services (gateway :8080, sim :8081)
- **Help**: \`make help\` - Show all available targets

## 🎯 Project Overview

Features:
- **Seamless local sharding** for scalable multiplayer
- **WebSocket real-time communication** with JSON protocol
- **Server-authoritative simulation** with spatial partitioning
- **Area of Interest (AOI)** management for efficient updates

Current Status:
- ✅ Core simulation engine with spatial math
- ✅ WebSocket transport layer
- ✅ Authentication and session management
- ✅ Local sharding implementation
- 🚧 Multi-node sharding (planned)
- 🚧 Client implementation (planned)

## 🤝 Community

- 📁 [Repository](https://github.com/AstroSteveo/prototype-game)
- 📖 [Wiki]($WIKI_BASE_URL) - Complete documentation
- 🐛 [Issues](https://github.com/AstroSteveo/prototype-game/issues)
- 💬 [Discussions](https://github.com/AstroSteveo/prototype-game/discussions)

---

For comprehensive documentation, development guides, and architecture details, visit the **[Project Wiki]($WIKI_BASE_URL)**.
EOF

    echo "✅ Created simplified README (backup saved as $backup_file)"
}

# Function to update AGENTS.md with wiki references
update_agents_md() {
    local file="$PROJECT_ROOT/AGENTS.md"
    local backup_file="${file}.backup"
    
    if [[ ! -f "$file" ]]; then
        echo "⚠️  Warning: AGENTS.md not found"
        return 1
    fi
    
    echo "📝 Updating AGENTS.md with wiki references..."
    
    # Create backup
    cp "$file" "$backup_file"
    
    # Add wiki reference to instruction scope
    sed -i '/^## Instruction Scope/a\\n- **Wiki**: Comprehensive documentation available at '"$WIKI_BASE_URL"' for detailed guides and references.' "$file"
    
    # Update project structure section
    sed -i '/^- **Root:** High‑level docs/a\  - **Wiki**: Complete documentation at '"$WIKI_BASE_URL"' with detailed guides and references' "$file"
    
    echo "✅ Updated AGENTS.md (backup saved as $backup_file)"
}

echo "Starting repository link updates..."

# Update main repository files
update_repo_links "$PROJECT_ROOT/AGENTS.md"

# Create simplified README
create_simplified_readme

# Update AGENTS.md with wiki references
update_agents_md

# Update any other files that might have documentation links
if [[ -f "$PROJECT_ROOT/.github/CONTRIBUTING.md" ]]; then
    update_repo_links "$PROJECT_ROOT/.github/CONTRIBUTING.md"
fi

echo ""
echo "✅ Repository link updates complete!"
echo ""
echo "Files updated:"
echo "- README.md (simplified with wiki links)"
echo "- AGENTS.md (added wiki references)"
echo "- .github/CONTRIBUTING.md (if present)"
echo ""
echo "Backup files created with .backup extension"
echo ""
echo "Next steps:"
echo "1. Review the updated files"
echo "2. Test all wiki links work correctly"
echo "3. Commit the changes to repository"
echo "4. Ensure wiki pages are created and accessible"