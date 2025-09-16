# Project Board Automation Setup

Follow these steps to configure GitHub Project board automations that pair with the `project-sync` workflow.

## Repository Configuration
1. Navigate to **Settings → Secrets and variables → Actions → Variables** and add:
   - `PROJECT_URL` — URL for the roadmap board (user or organization project).
2. Navigate to **Settings → Secrets and variables → Actions → Secrets** and add:
   - `PROJECTS_TOKEN` — Fine-grained personal access token with repository access to this repo and **Projects: Write** permission for the target board.

## Project Field Configuration
Ensure the following custom fields exist on the project board. Adjust names only if you update the workflow to match.
- **Status** (Single select): Backlog, Ready, In Progress, In Review, Blocked, Done.
- **Estimate** (Number): Accepts numeric story points.
- **Milestone** (Milestone): Links to GitHub milestones when applicable.
- **Sprint** (Iteration): Configure the active and upcoming iteration windows.

## Project Workflows (Board Automations)
Configure Project Workflows to complement the Action automation:
1. **Item added → Set Status to Backlog**
2. **Item assigned → Set Status to In Progress**
3. **Issue closed → Set Status to Done**
4. **PR merged → Set Status to Done**
5. **Auto-archive** items with Status `Done` for more than 14 days.

## GitHub Action Behavior Summary
The `.github/workflows/project-sync.yml` job:
- Adds issues/PRs labeled `story`, `bug`, or `task` to the project defined by `PROJECT_URL`.
- Updates status using labels (`ready`, `blocked`, `in-progress`) and PR state transitions.
- Parses estimates from labels (`estimate:3`, `size:m`) or trailing `[3]` tokens in titles.
- Mirrors milestone and iteration fields when present.
- Retries project lookup using exponential backoff; tune with `PROJECT_SYNC_MAX_RETRIES` and `PROJECT_SYNC_BASE_DELAY_MS` environment variables if needed.

## Validation Checklist
1. Create an issue using the Task template; ensure it appears on the board with Status `Backlog` and correct estimate.
2. Apply `ready`/`blocked` labels and confirm Status updates.
3. Close the issue; Status should transition to `Done` and board workflows should archive after the grace period.
4. Label a PR and walk it through draft → review → merge to verify status changes.

## Troubleshooting Tips
- Verify `PROJECT_URL` is correct and the token has not expired.
- Confirm field names on the board match the workflow configuration.
- Use the workflow run logs for detailed error messages, including parsed scope and item IDs.
- Re-run the workflow with "Re-run jobs" after updating secrets or variables.

## Maintenance
- Review iteration schedules and board column policies each planning cycle.
- Rotate PATs before they expire and update the secret.
- Keep this document in sync when automation logic changes.
