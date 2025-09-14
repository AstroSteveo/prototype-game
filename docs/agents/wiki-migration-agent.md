# Wiki Migration Agent

## Purpose
This agent is responsible for executing the documentation migration plan created by the project manager agent. It will organize and migrate all markdown documentation to the GitHub wiki following the defined structure and user stories.

## Wiki Structure Design

Based on the documentation analysis, here is the proposed wiki structure:

### Home Page
- Project overview and vision statement
- Quick navigation to major sections
- Getting started links
- Recent updates and announcements

### 1. Getting Started
- **Welcome**: Project introduction and core concepts
- **Quick Start**: Fast setup and first run
- **Prerequisites**: System requirements and dependencies
- **Installation**: Detailed setup instructions

### 2. Development Guide
- **Developer Setup**: Complete development environment setup
- **Build System**: Build commands and Makefile usage
- **Testing**: Unit tests, integration tests, and validation procedures
- **Daily Workflows**: Common development tasks and commands
- **Troubleshooting**: Common issues and solutions

### 3. Architecture & Design
- **Game Design Document**: Vision, pillars, and player experience
- **Technical Architecture**: System design and component overview
- **Spatial Systems**: Cell math, handovers, and AOI management
- **Network Protocol**: WebSocket implementation and message formats
- **Sharding Strategy**: Local and distributed sharding approaches

### 4. Contributing
- **How to Contribute**: Guidelines for new contributors
- **Workflow**: Branch strategy, PR process, and commit conventions
- **Code Standards**: Formatting, naming conventions, and best practices
- **Pull Request Template**: Standard PR format and requirements

### 5. Agent & Automation
- **Agent Instructions**: Guidelines for AI assistants and automation
- **Copilot Setup**: GitHub Copilot configuration and usage
- **Automation Workflows**: Automated processes and scripts
- **Testing Automation**: Automated testing strategies

### 6. API Reference
- **Gateway API**: Authentication and session management endpoints
- **Simulation API**: Game engine and state management
- **WebSocket Protocol**: Real-time communication specification
- **Metrics API**: Performance and monitoring endpoints

## Content Migration Plan

### Phase 1: Core Documentation Migration

#### 1.1 Home Page Creation
**Source**: `README.md` (condensed)
**Target**: Wiki Home
**Content**:
- Project vision and key features
- Quick start links
- Navigation to major sections
- Build status and key metrics

#### 1.2 Getting Started Section
**Sources**: 
- `README.md` (quick start)
- `docs/dev/DEV.md` (prerequisites)
- `.github/copilot-instructions.md` (build instructions)

**Pages**:
- Welcome (project intro)
- Quick Start (make commands)
- Prerequisites (Go 1.23+, tools)
- Installation (detailed setup)

#### 1.3 Development Guide Section
**Sources**:
- `docs/dev/DEV.md` (primary content)
- `.github/copilot-instructions.md` (workflows)
- `AGENTS.md` (build guidelines)

**Pages**:
- Developer Setup
- Build System
- Testing
- Daily Workflows
- Troubleshooting

### Phase 2: Design Documentation Migration

#### 2.1 Architecture & Design Section
**Sources**:
- `docs/design/GDD.md`
- `docs/design/TDD.md`

**Pages**:
- Game Design Document
- Technical Architecture
- Spatial Systems
- Network Protocol
- Sharding Strategy

### Phase 3: Process Documentation Migration

#### 3.1 Contributing Section
**Sources**:
- `.github/CONTRIBUTING.md`
- `AGENTS.md` (guidelines)
- `.github/pull_request_template.md`

**Pages**:
- How to Contribute
- Workflow
- Code Standards
- Pull Request Template

#### 3.2 Agent & Automation Section
**Sources**:
- `AGENTS.md`
- `.github/copilot-instructions.md`
- `docs/agents/project-manager-agent.md`

**Pages**:
- Agent Instructions
- Copilot Setup
- Automation Workflows
- Testing Automation

### Phase 4: API Documentation

#### 4.1 API Reference Section
**Sources**: Code analysis and existing documentation
**Pages**:
- Gateway API
- Simulation API
- WebSocket Protocol
- Metrics API

## Implementation Scripts

### Script 1: Wiki Structure Generator
```bash
#!/bin/bash
# generate-wiki-structure.sh
# Creates the basic wiki page structure

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

echo "Creating wiki structure with ${#WIKI_PAGES[@]} pages..."
for page in "${WIKI_PAGES[@]}"; do
    echo "- $page"
done
```

### Script 2: Content Extractor
```bash
#!/bin/bash
# extract-content.sh
# Extracts and processes content from existing markdown files

extract_section() {
    local file="$1"
    local start_pattern="$2"
    local end_pattern="$3"
    local output_file="$4"
    
    awk "/$start_pattern/,/$end_pattern/" "$file" > "$output_file"
}

# Extract sections from existing documentation
extract_section "README.md" "^# prototype-game" "^Notes:" "wiki-content/home-intro.md"
extract_section "docs/design/GDD.md" "^# Game Design Document" "" "wiki-content/gdd.md"
extract_section "docs/design/TDD.md" "^# Technical Design Document" "" "wiki-content/tdd.md"
extract_section "docs/dev/DEV.md" "^# Developer Guide" "" "wiki-content/dev-guide.md"
```

### Script 3: Link Updater
```bash
#!/bin/bash
# update-links.sh
# Updates internal links to point to wiki pages

update_repo_links() {
    local file="$1"
    
    # Update links to design docs
    sed -i 's|docs/design/GDD\.md|https://github.com/AstroSteveo/prototype-game/wiki/Game-Design-Document|g' "$file"
    sed -i 's|docs/design/TDD\.md|https://github.com/AstroSteveo/prototype-game/wiki/Technical-Architecture|g' "$file"
    sed -i 's|docs/dev/DEV\.md|https://github.com/AstroSteveo/prototype-game/wiki/Development-Guide|g' "$file"
    
    # Update links to contributing docs
    sed -i 's|\.github/CONTRIBUTING\.md|https://github.com/AstroSteveo/prototype-game/wiki/Contributing|g' "$file"
}

# Update repository files
update_repo_links "README.md"
update_repo_links "AGENTS.md"
```

## Content Templates

### Home Page Template
```markdown
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
Last updated: {current_date}
```

### Section Landing Page Template
```markdown
# {Section Name}

{Brief description of the section and its purpose}

## Pages in this Section

{List of pages with brief descriptions}

## Quick Links

{Most commonly accessed pages or actions}

## Prerequisites

{Any setup or knowledge required for this section}

---
📖 **Navigation**: [Home](Home) → {Section Name}
```

## Execution Checklist

### Pre-Migration Tasks
- [ ] Backup all existing documentation
- [ ] Set up wiki repository access
- [ ] Create content extraction scripts
- [ ] Test migration process on sample content

### Migration Execution
- [ ] Create wiki page structure
- [ ] Extract and process content from source files
- [ ] Upload content to wiki pages
- [ ] Update internal cross-references
- [ ] Add navigation elements

### Post-Migration Tasks
- [ ] Update repository README with wiki links
- [ ] Clean up redundant repository documentation
- [ ] Update AGENTS.md with wiki references
- [ ] Test all wiki navigation and links
- [ ] Create maintenance procedures

### Quality Assurance
- [ ] Verify all content migrated successfully
- [ ] Check formatting and markdown rendering
- [ ] Test navigation flow and user experience
- [ ] Validate external and internal links
- [ ] Ensure search functionality works

## Success Criteria

1. **Completeness**: All existing documentation content available in wiki
2. **Organization**: Logical structure with clear navigation
3. **Accessibility**: All content findable within 2 clicks from home
4. **Maintainability**: Clear process for keeping content current
5. **Usability**: Improved developer onboarding and information discovery

## Maintenance Strategy

### Content Synchronization
- Repository docs focus on essential development info
- Wiki contains comprehensive documentation
- Automated checks for content drift
- Regular review cycle for updates

### Responsibility Matrix
- **Repository maintainers**: Keep essential docs current
- **Wiki editors**: Organize and expand comprehensive content
- **Automation**: Sync processes and consistency checks
- **Contributors**: Follow guidelines for documentation updates

## Rollback Plan

If migration issues occur:
1. Keep original repository docs until wiki is stable
2. Maintain backup of all content
3. Use gradual migration approach (section by section)
4. Have rollback scripts to restore repository links

---

**Next Actions**:
1. Get approval for wiki structure design
2. Create and test migration scripts
3. Execute migration in phases
4. Update repository documentation
5. Train team on new documentation structure