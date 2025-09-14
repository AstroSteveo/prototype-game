#!/bin/bash
# generate-wiki-structure.sh
# Creates the basic wiki page structure for the prototype-game project

set -euo pipefail

# Wiki pages to create
WIKI_PAGES=(
    "Home"
    "Getting-Started"
    "Welcome"
    "Quick-Start"
    "Prerequisites"
    "Installation"
    "Development-Guide"
    "Developer-Setup"
    "Build-System"
    "Testing"
    "Daily-Workflows"
    "Troubleshooting"
    "Architecture-&-Design"
    "Game-Design-Document"
    "Technical-Architecture"
    "Spatial-Systems"
    "Network-Protocol"
    "Sharding-Strategy"
    "Contributing"
    "How-to-Contribute"
    "Workflow"
    "Code-Standards"
    "Pull-Request-Template"
    "Agent-&-Automation"
    "Agent-Instructions"
    "Copilot-Setup"
    "Automation-Workflows"
    "Testing-Automation"
    "API-Reference"
    "Gateway-API"
    "Simulation-API"
    "WebSocket-Protocol"
    "Metrics-API"
)

echo "=== Prototype Game Wiki Structure Generator ==="
echo "This script will generate the wiki page structure for documentation migration."
echo ""
echo "Proposed wiki structure with ${#WIKI_PAGES[@]} pages:"
echo ""

# Group pages by section for better visualization
echo "📖 Core Pages:"
echo "   - Home"
echo ""

echo "🚀 Getting Started:"
echo "   - Getting-Started"
echo "   - Welcome"
echo "   - Quick-Start"  
echo "   - Prerequisites"
echo "   - Installation"
echo ""

echo "🛠️ Development Guide:"
echo "   - Development-Guide"
echo "   - Developer-Setup"
echo "   - Build-System"
echo "   - Testing"
echo "   - Daily-Workflows"
echo "   - Troubleshooting"
echo ""

echo "🏗️ Architecture & Design:"
echo "   - Architecture-&-Design"
echo "   - Game-Design-Document"
echo "   - Technical-Architecture"
echo "   - Spatial-Systems"
echo "   - Network-Protocol"
echo "   - Sharding-Strategy"
echo ""

echo "🤝 Contributing:"
echo "   - Contributing"
echo "   - How-to-Contribute"
echo "   - Workflow"
echo "   - Code-Standards"
echo "   - Pull-Request-Template"
echo ""

echo "🤖 Agent & Automation:"
echo "   - Agent-&-Automation"
echo "   - Agent-Instructions"
echo "   - Copilot-Setup"
echo "   - Automation-Workflows"
echo "   - Testing-Automation"
echo ""

echo "📚 API Reference:"
echo "   - API-Reference"
echo "   - Gateway-API"
echo "   - Simulation-API"
echo "   - WebSocket-Protocol"
echo "   - Metrics-API"
echo ""

# Create directory structure for local wiki content preparation
WIKI_DIR="wiki-content"
mkdir -p "$WIKI_DIR"/{getting-started,development,architecture,contributing,automation,api}

echo "Created local wiki content directory structure:"
echo "   $WIKI_DIR/"
echo "   ├── getting-started/"
echo "   ├── development/"
echo "   ├── architecture/"
echo "   ├── contributing/"
echo "   ├── automation/"
echo "   └── api/"
echo ""

# Create placeholder files for each section
for page in "${WIKI_PAGES[@]}"; do
    case "$page" in
        Welcome|Quick-Start|Prerequisites|Installation)
            touch "$WIKI_DIR/getting-started/${page,,}.md"
            ;;
        Developer-Setup|Build-System|Testing|Daily-Workflows|Troubleshooting)
            touch "$WIKI_DIR/development/${page,,}.md"
            ;;
        Game-Design-Document|Technical-Architecture|Spatial-Systems|Network-Protocol|Sharding-Strategy)
            touch "$WIKI_DIR/architecture/${page,,}.md"
            ;;
        How-to-Contribute|Workflow|Code-Standards|Pull-Request-Template)
            touch "$WIKI_DIR/contributing/${page,,}.md"
            ;;
        Agent-Instructions|Copilot-Setup|Automation-Workflows|Testing-Automation)
            touch "$WIKI_DIR/automation/${page,,}.md"
            ;;
        Gateway-API|Simulation-API|WebSocket-Protocol|Metrics-API)
            touch "$WIKI_DIR/api/${page,,}.md"
            ;;
        *)
            touch "$WIKI_DIR/${page,,}.md"
            ;;
    esac
done

echo "✅ Wiki structure preparation complete!"
echo ""
echo "Next steps:"
echo "1. Run ./extract-content.sh to extract content from existing docs"
echo "2. Review and edit extracted content in $WIKI_DIR/"
echo "3. Use GitHub's wiki interface to create the actual pages"
echo "4. Run ./update-links.sh to update repository links"