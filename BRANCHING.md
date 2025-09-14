# Branch Policy and Governance

This document establishes branch governance for the Prototype Game repository, ensuring clean history and efficient collaboration.

## Core Policy: Main-Only Persistence

**Main branch (`main`) is the only persistent branch.** All feature, hotfix, and maintenance branches are short-lived and deleted after merge.

### Key Principles
- **Single source of truth**: `main` branch contains the authoritative codebase
- **Linear history**: Prefer rebase/fast-forward merges for clean commit history
- **Short-lived branches**: Feature and fix branches exist only during active development
- **Automatic cleanup**: Branches are auto-deleted on merge to prevent accumulation

## Branch Naming Conventions

Use consistent naming patterns to improve organization and automation:

### Format Patterns
- **Feature branches**: `feat/<id><slug>`
- **Bug fixes**: `fix/<id><slug>`
- **Maintenance/chores**: `chore/<slug>`

### Examples
```
feat/us-201-aoi-visibility
feat/67-add-branching-docs
fix/handover-hysteresis
fix/23-websocket-auth-error
chore/update-dependencies
chore/refactor-spatial-tests
```

### Naming Guidelines
- Use lowercase with hyphens for readability
- Include issue/story ID when applicable (`us-201`, `67`, `23`)
- Keep slugs concise but descriptive
- Avoid special characters except hyphens

## Branch Lifecycle Rules

### 1. Creation and Development
- **One branch per story/task**: Keep scope focused and reviewable
- **Branch from main**: Always create branches from latest `main`
- **Early PR creation**: Open Draft PR within 24 hours for CI validation

### 2. Pull Request Requirements
- **Mandatory CI**: All checks must pass (fmt, vet, test, test-ws)
- **Focused scope**: Aim for <300 lines changed when practical
- **Link issues**: Reference GitHub issues or project board items
- **Update documentation**: Include relevant doc updates for behavior changes

### 3. Merge and Cleanup
- **Auto-delete on merge**: Branches are automatically deleted after successful merge
- **No long-lived branches**: Feature branches should not exist longer than a few days
- **Rebase preferred**: Use rebase/fast-forward to maintain linear history

## Issue Linking and PR Title Examples

### Pull Request Titles
Follow the pattern: `<area>: <imperative description>`

```
sim: fix handover hysteresis in AOI boundaries
gateway: add player validation endpoint
docs: update TDD with spatial cell design
chore: upgrade Go dependencies to 1.23
```

### Issue References in PRs
Reference issues in PR descriptions and commit messages:

**In PR Description:**
```markdown
## Intent
Fixes #67 - Add comprehensive branch governance documentation

## Summary of Changes
- Created BRANCHING.md with main-only persistence policy
- Documented naming conventions and lifecycle rules
- Added examples for issue linking and PR titles
```

**In Commit Messages:**
```
docs: add BRANCHING.md with governance policy

- Documents main-only persistence and short-lived branches
- Establishes naming conventions (feat/fix/chore patterns)
- Defines 24h PR rule and auto-delete on merge
- Includes examples for issue linking

Fixes #67
```

### Issue Labels for Automation
Use consistent labels to trigger project automation:

- `story` - Feature development work
- `bug` - Bug fixes and corrections  
- `task` - Maintenance and administrative work

These labels automatically add items to the project board via [add-to-project.yml](/.github/workflows/add-to-project.yml).

## Project Automation References

### Workflow Integration
- **CI Pipeline**: [ci.yml](/.github/workflows/ci.yml) - Validates all PRs with comprehensive testing
- **Project Board**: [add-to-project.yml](/.github/workflows/add-to-project.yml) - Auto-adds labeled issues/PRs to project

### Documentation Links
- **Agent Instructions**: [AGENTS.md](/AGENTS.md) - Authoritative guidelines for automation
- **Contributing Guide**: [.github/CONTRIBUTING.md](/.github/CONTRIBUTING.md) - Developer workflow overview
- **Development Guide**: [docs/dev/DEV.md](/docs/dev/DEV.md) - Daily development commands and tips

## Enforcement and Compliance

### Automated Checks
- **Branch protection**: Direct pushes to `main` are blocked
- **Required CI**: All status checks must pass before merge
- **Auto-deletion**: Merged branches are automatically removed

### Manual Review Points
- PR scope and focus (avoid large, unfocused changes)
- Documentation updates for behavior changes
- Proper issue linking and project board updates
- Adherence to naming conventions

## Examples Walkthrough

### Complete Feature Development Flow

1. **Create Feature Branch**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b feat/67-add-branching-docs
   ```

2. **Open Draft PR Early**
   - Create Draft PR within 24 hours
   - Include initial intent and scope
   - Link relevant issues: "Addresses #67"

3. **Development and Testing**
   ```bash
   # Make changes, test locally
   make fmt vet test test-ws
   
   # Commit with clear messages
   git commit -m "docs: add BRANCHING.md structure

   - Created initial framework for branch governance
   - Added main-only persistence policy
   
   Part of #67"
   ```

4. **Final PR and Merge**
   - Mark PR as ready for review
   - Ensure all CI checks pass
   - Merge using rebase/fast-forward
   - Branch auto-deleted on merge

This workflow ensures clean history, proper documentation, and efficient collaboration while maintaining the main-only persistence policy.