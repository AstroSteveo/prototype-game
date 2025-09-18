# Agent Validation Checklist

Use this checklist to validate that an AI agent can successfully work with the prototype-game repository. Run through this checklist whenever onboarding a new agent or troubleshooting agent issues.

## Repository Access Validation

- [ ] **Repository cloned and accessible**
  - [ ] Can navigate to repository root
  - [ ] Can list files in repository
  - [ ] `.git` directory exists and is accessible

- [ ] **AGENTS.md hierarchy readable**
  - [ ] Root `AGENTS.md` accessible and readable
  - [ ] `docs/AGENTS.md` accessible and readable  
  - [ ] `backend/AGENTS.md` accessible and readable
  - [ ] `docs/.llm/AGENTS.md` accessible and readable

## Documentation Structure Validation

- [ ] **Core documentation exists and is accessible**
  - [ ] `README.md` - Quick start guide
  - [ ] `docs/development/developer-guide.md` - Developer workflows
  - [ ] `docs/product/vision/game-design-document.md` - Game Design Document
  - [ ] `docs/architecture/technical-design-document.md` - Technical Design Document
  - [ ] `docs/README.md` - Documentation overview

- [ ] **LLM onboarding framework complete**
  - [ ] `docs/.llm/AGENTS.md` - Operating manual
  - [ ] `docs/.llm/onboarding/quick-start.md` - Quick start guide
  - [ ] `docs/.llm/onboarding/contribution-checklist.md` - Contribution checklist
  - [ ] `docs/.llm/onboarding/copilot-playbook.md` - Copilot guidance
  - [ ] `docs/.llm/onboarding/story-template.md` - Story template

## Build and Test Infrastructure Validation

- [ ] **Make targets accessible and functional**
  - [ ] `make help` shows available targets
  - [ ] `make fmt vet` runs successfully (if Go environment available)
  - [ ] `make test test-ws` executes (if Go environment available)
  - [ ] `make build` works (if Go environment available)

- [ ] **Repository scripts executable**
  - [ ] `scripts/agents/prepare-context.sh` exists and is executable
  - [ ] Context preparation script runs without errors

## GitHub Integration Validation

- [ ] **Issue templates accessible**
  - [ ] `.github/ISSUE_TEMPLATE/` directory exists
  - [ ] All referenced templates (standup.yml, planning.yml, roadmap.yml, etc.) exist
  - [ ] `config.yml` has correct paths

- [ ] **Copilot integration configured**
  - [ ] `.github/copilot-instructions.md` exists and is comprehensive
  - [ ] All file paths in copilot instructions are correct and accessible

## Cross-Reference Validation

- [ ] **Internal links functional**
  - [ ] All markdown links within docs point to existing files
  - [ ] Cross-references between AGENTS.md files are accurate
  - [ ] Template references in session guides point to existing files

- [ ] **No broken references**
  - [ ] No references to non-existent `docs/dev/DEV.md`
  - [ ] No references to non-existent `docs/design/GDD.md` or `docs/design/TDD.md`
  - [ ] All script paths in documentation are correct

## Agent Capability Validation

- [ ] **File operations**
  - [ ] Can read files from repository
  - [ ] Can create new files (test with temporary file)
  - [ ] Can modify existing files (test with temporary file)
  - [ ] Can delete files (test with temporary file)

- [ ] **Command execution** (if environment supports it)
  - [ ] Can execute shell commands
  - [ ] Can run make targets
  - [ ] Can access git commands
  - [ ] Can run context preparation script

## Framework Completeness Check

- [ ] **Agent guidance comprehensive**
  - [ ] Role-specific instructions available for different agent types
  - [ ] Error handling and escalation procedures documented
  - [ ] Working with humans guidance provided
  - [ ] Repository conventions clearly documented

- [ ] **Integration with agents.md standards**
  - [ ] References https://agents.md/ appropriately
  - [ ] Follows agents.md file structure conventions
  - [ ] Hierarchical instruction system properly implemented

## Troubleshooting Reference

**Common Issues and Solutions:**

- **"Cannot find DEV.md"**: File was moved - use `docs/development/developer-guide.md` instead
- **"GDD.md or TDD.md not found"**: Files are at `docs/product/vision/game-design-document.md` and `docs/architecture/technical-design-document.md`
- **"Context script fails"**: Ensure you're in repository root and script has execute permissions
- **"Make targets fail"**: Verify Go 1.23+ is installed and accessible

**Escalation:**
If validation fails despite following troubleshooting steps, create an issue using the appropriate template in `.github/ISSUE_TEMPLATE/` to get human assistance.

---

**Last updated**: When modifying this checklist, update this timestamp and note what changed.