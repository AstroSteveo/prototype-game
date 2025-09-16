# Wiki Migration Scripts

This directory contains automation scripts to support the migration of documentation from the repository to the GitHub wiki.

## Scripts Overview

### 1. `generate-wiki-structure.sh`
**Purpose**: Creates the proposed wiki page structure and local content directories for preparation.

**Usage**:
```bash
./generate-wiki-structure.sh
```

**Output**:
- Displays the complete wiki structure (33 pages)
- Creates local `wiki-content/` directory with organized subdirectories
- Creates placeholder files for content preparation

### 2. `extract-content.sh`
**Purpose**: Extracts content from existing repository markdown files and processes it for wiki migration.

**Usage**:
```bash
./extract-content.sh
```

**Features**:
- Extracts sections and full files from existing documentation
- Organizes content by wiki sections (getting-started, development, architecture, etc.)
- Creates processed content ready for wiki upload
- Generates landing pages for major sections

**Output Files**:
- `wiki-content/home-processed.md` - Complete home page for wiki
- `wiki-content/getting-started.md` - Landing page for Getting Started section
- Section-specific content in organized subdirectories

### 3. `update-links.sh`
**Purpose**: Updates repository files to link to wiki pages instead of local documentation files.

**Usage**:
```bash
./update-links.sh
```

**Features**:
- Creates simplified README.md with wiki links
- Updates AGENTS.md with wiki references
- Updates internal links in repository files
- Creates backup files before making changes

**Safety**:
- All modified files are backed up with `.backup` extension
- Changes can be reverted if needed

## Migration Process

### Phase 1: Preparation
```bash
# 1. Generate wiki structure
./generate-wiki-structure.sh

# 2. Extract and process content
./extract-content.sh

# 3. Review extracted content
ls -la wiki-content/
```

### Phase 2: Wiki Creation
1. Manually create wiki pages using GitHub's wiki interface
2. Copy processed content from `wiki-content/` to appropriate wiki pages
3. Set up navigation and cross-references

### Phase 3: Repository Updates
```bash
# Update repository links to point to wiki
./update-links.sh

# Review changes
git diff
```

## Content Organization

### Generated Directory Structure
```
wiki-content/
├── getting-started/
│   ├── quick-start.md
│   ├── prerequisites.md
│   └── ...
├── development/
│   ├── dev-guide-full.md
│   ├── build-system.md
│   └── ...
├── architecture/
│   ├── game-design-document.md
│   ├── technical-architecture.md
│   └── ...
├── contributing/
│   ├── contributing-full.md
│   ├── workflow.md
│   └── ...
├── automation/
│   ├── agent-instructions.md
│   ├── copilot-setup.md
│   └── ...
├── api/
│   ├── websocket-protocol.md
│   └── ...
├── home-processed.md
└── getting-started.md
```

## Wiki Structure

### Proposed 33-Page Structure
- **Core**: Home
- **Getting Started** (5 pages): Landing, Welcome, Quick-Start, Prerequisites, Installation
- **Development Guide** (6 pages): Landing, Developer-Setup, Build-System, Testing, Daily-Workflows, Troubleshooting  
- **Architecture & Design** (6 pages): Landing, Game-Design-Document, Technical-Architecture, Spatial-Systems, Network-Protocol, Sharding-Strategy
- **Contributing** (5 pages): Landing, How-to-Contribute, Workflow, Code-Standards, Pull-Request-Template
- **Agent & Automation** (5 pages): Landing, Agent-Instructions, Copilot-Setup, Automation-Workflows, Testing-Automation
- **API Reference** (5 pages): Landing, Gateway-API, Simulation-API, WebSocket-Protocol, Metrics-API

## Quality Assurance

### Pre-Migration Checklist
- [ ] All source files identified and accessible
- [ ] Backup strategy in place
- [ ] Local testing of extraction scripts
- [ ] Review of proposed wiki structure

### Post-Migration Checklist
- [ ] All wiki pages created and accessible
- [ ] Navigation flows work correctly
- [ ] Repository links updated and tested
- [ ] Original documentation backed up
- [ ] Team trained on new structure

## Troubleshooting

### Common Issues

**Script Permission Errors**:
```bash
chmod +x *.sh
```

**Missing Source Files**:
- Scripts will warn about missing files
- Review the source file paths in extract-content.sh

**Link Update Issues**:
- Check backup files (*.backup) if restoration needed
- Verify wiki URLs are accessible before running update-links.sh

### Rollback Process

1. Restore from backup files:
```bash
# Restore README
mv README.md.pre-wiki-backup README.md

# Restore AGENTS.md  
mv AGENTS.md.backup AGENTS.md
```

2. Remove generated content:
```bash
rm -rf wiki-content/
```

## Support

For issues with these scripts or the migration process:
1. Check the backup files for safe restoration
2. Review the current documentation governance in `docs/.llm/AGENTS.md`
3. File an issue in the repository with details about the problem

---

**Maintenance Note**: These scripts were authored for the previous wiki migration effort. Consult `docs/.llm/AGENTS.md` and the roadmap handbook for up-to-date documentation policies before running them.