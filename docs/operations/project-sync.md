# Project Sync Configuration and Validation

This runbook describes how to configure the GitHub Actions workflow that keeps issues and pull requests synchronized with the roadmap project board.

## Configuration Requirements
### Repository Variables
- **PROJECT_URL** — URL of the Project (classic or Project v2) to sync against.
  - Format examples:
    - User projects: `https://github.com/users/<username>/projects/<number>`
    - Organization projects: `https://github.com/orgs/<org>/projects/<number>`
  - Set via **Settings → Secrets and variables → Actions → Variables**.
  - Verify by opening the URL in a browser to ensure it resolves to the expected board.

### Repository Secrets
- **PROJECTS_TOKEN** — Fine-grained personal access token with `Project` (write) and `Repository` (read/write) scopes for the target project.
  - Configure under **Settings → Secrets and variables → Actions → Secrets**.
  - Confirm the token user can view and modify the project board.

## Workflow Behavior
The `.github/workflows/project-sync.yml` file:
- Triggers on issue and pull request events (opened, labeled, closed, reopened, merged).
- Adds items when they include labels `story`, `bug`, or `task`.
- Updates status based on labels (`ready`, `blocked`, `in-progress`) and PR state (draft, ready for review, merged).
- Parses `PROJECT_URL` to support both user and organization scopes and retries when GitHub propagation lags.

## Validation Checklist
### Prerequisites
- [ ] `PROJECT_URL` variable populated with a valid project board URL.
- [ ] `PROJECTS_TOKEN` secret exists and has access to the project.
- [ ] Issue templates apply the `story`, `bug`, or `task` labels used for automation.

### Functional Tests
1. **Story Issue**
   - Create an issue with the Story template.
   - Confirm the workflow run adds it to the project with status `Backlog`.
2. **Task Issue**
   - Repeat using the Task template; verify item placement and status updates when labels change.
3. **Bug Issue**
   - Ensure bugs sync correctly and land in the proper column.
4. **Pull Request**
   - Label a PR with `story`, `bug`, or `task`.
   - Check that status transitions follow PR state (draft → In Progress, ready for review → In Review, merged → Done).

### Guard Rails
- [ ] Workflow skips Dependabot events.
- [ ] Workflow exits early when `PROJECT_URL` or `PROJECTS_TOKEN` are missing.
- [ ] Unlabeled issues remain unsynced until a supported label is applied.

## Troubleshooting
1. Inspect the latest workflow logs for the run; debug messages print the parsed project scope and item identifiers.
2. Confirm token permissions have not expired or been revoked.
3. Validate that the project board still exists and is not archived.
4. Check label spelling against the automation configuration.
5. Re-run the workflow after updating secrets/variables to refresh the environment.

## Manual Sync (Fallback)
If automation is unavailable:
1. Open the project board at the configured `PROJECT_URL`.
2. Use **Add item → Add by URL** to search for the issue or PR.
3. Set the appropriate status, iteration, and metadata manually.

## Related Files
- `.github/workflows/project-sync.yml`
- `.github/ISSUE_TEMPLATE/story.yml`
- `.github/ISSUE_TEMPLATE/task.yml`
- `.github/ISSUE_TEMPLATE/bug.yml`

Review this runbook whenever the workflow changes or new labels/board columns are introduced.
