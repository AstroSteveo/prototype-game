# Project Board Automation Setup

This document provides step-by-step instructions for configuring the [Game Roadmap Project](https://github.com/users/AstroSteveo/projects/2) board automation.

## Repository Configuration

### 1. Repository Variables

Navigate to **Settings → Secrets and variables → Actions → Variables** and add:

| Variable | Value |
|----------|-------|
| `PROJECT_URL` | `https://github.com/users/AstroSteveo/projects/2` |

### 2. Repository Secrets

Navigate to **Settings → Secrets and variables → Actions → Secrets** and add:

| Secret | Description |
|--------|-------------|
| `PROJECTS_TOKEN` | Fine-grained Personal Access Token with:<br/>• Repository access: This repository<br/>• Project permissions: Write access to target Project V2 |

**Creating the PAT:**
1. Go to GitHub **Settings → Developer settings → Personal access tokens → Fine-grained tokens**
2. Generate new token with:
   - **Repository access**: `AstroSteveo/prototype-game`
   - **Account permissions**: Projects (Write)
3. Copy the token value to the `PROJECTS_TOKEN` secret

## Project Field Configuration

Navigate to the [Project Settings](https://github.com/users/AstroSteveo/projects/2/settings/fields) and ensure these fields exist with exact names:

### Status Field (Single Select)
- **Name**: `Status` (case-insensitive matching)
- **Type**: Single select
- **Options**:
  - Backlog
  - Ready
  - In Progress
  - In Review
  - Blocked
  - Done

### Estimate Field (Number)
- **Name**: `Estimate`
- **Type**: Number

### Milestone Field
- **Name**: `Milestone`
- **Type**: Milestone

### Sprint Field (Iteration)
- **Name**: `Sprint`
- **Type**: Iteration
- **Configuration**: Set up current and upcoming iterations

## Project UI Workflows

Navigate to [Project Workflows](https://github.com/users/AstroSteveo/projects/2/workflows) and configure these automation rules:

### 1. Item Added → Set Status to Backlog
- **Trigger**: Item added to project
- **Action**: Set field `Status` to `Backlog`

### 2. Item Assigned → Set Status to In Progress  
- **Trigger**: Item assigned
- **Action**: Set field `Status` to `In Progress`

### 3. Issue Closed → Set Status to Done
- **Trigger**: Issue closed
- **Action**: Set field `Status` to `Done`

### 4. PR Merged → Set Status to Done
- **Trigger**: Pull request merged
- **Action**: Set field `Status` to `Done`

### 5. Auto-Archive Completed Items
- **Trigger**: Status changed to `Done` for 14 days
- **Action**: Archive item

## GitHub Action Behavior

The `.github/workflows/project-sync.yml` action handles:

### Auto-Add Items
- Issues/PRs labeled with `story`, `bug`, or `task` are automatically added to the project

### Status Management
- `ready` label → Status: Ready
- `blocked` label → Status: Blocked  
- `in-progress` label → Status: In Progress
- Issue opened → Status: Backlog
- Issue closed → Status: Done
- PR ready for review → Status: In Review
- PR converted to draft → Status: In Progress
- PR merged → Status: Done

### Estimate Parsing
Automatically extracts estimates from:
- Labels: `estimate:3`, `points:3`, `size:m`
- Title patterns: `[3]` at end of title
- Size mappings: `size:xs`=1, `size:s`=2, `size:m`=3, `size:l`=5, `size:xl`=8

### Other Fields
- **Milestone**: Set from item's milestone
- **Sprint**: New issues assigned to current iteration

### Optional Tuning (Env Vars)
- `PROJECT_SYNC_MAX_RETRIES` (default: `5`): Max attempts to find the Project item after add-to-project.
- `PROJECT_SYNC_BASE_DELAY_MS` (default: `2000`): Initial delay for exponential backoff between retries.

## Validation

### Test Case: Create Task Issue
1. Create new issue using Task template
2. Add `[2]` to title (e.g., "task: Test automation [2]")
3. Verify automation:
   - ✅ Issue added to project (labeled `task`)
   - ✅ Status set to Backlog
   - ✅ Estimate set to 2 (from `[2]` in title)
   - ✅ Sprint set to current iteration (if configured)

### Test Status Transitions
1. Add `ready` label → Status should become Ready
2. Add `blocked` label → Status should become Blocked  
3. Close issue → Status should become Done

### Test PR Workflow
1. Open draft PR → Status: In Progress
2. Mark ready for review → Status: In Review
3. Merge PR → Status: Done

## Troubleshooting

### Action Not Triggering
- Check repository variables and secrets are set correctly
- Verify PAT has required permissions
- Re-edit issue/PR to retrigger automation

### Fields Not Updating  
- Verify field names match exactly (case-insensitive)
- Check project field types are correct
- Ensure Sprint/iteration field has active iterations

### Missing Project Permissions
- Regenerate `PROJECTS_TOKEN` with Projects (Write) permission
- Verify token scope includes target repository

## Maintenance

- Review and update iteration/sprint schedules regularly
- Monitor automation logs in Actions tab
- Update PAT before expiration
- Adjust field options as project needs evolve