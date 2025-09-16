# Branch Maintenance Runbook

Use this runbook to keep the repository’s branch list clean and to ensure `main` remains the authoritative source of truth.

## When to Run This Playbook
- After large releases or documentation sweeps when many short-lived branches may linger.
- Before rotating maintainers or enabling automation that assumes a clean branch list.
- Whenever `git branch -a` shows stale feature or experiment branches older than the agreed retention window.

## Preparation
1. Fetch the latest history: `git fetch --all --prune`.
2. Review open pull requests to ensure no active work relies on branches targeted for cleanup.
3. Communicate the maintenance window in team channels to avoid surprises.

## Local Branch Cleanup
1. List local branches merged into `main`:
   ```bash
   git branch --merged main
   ```
2. Delete each merged branch:
   ```bash
   git branch -d <branch>
   ```
   Use `-D` only if you have confirmed the branch is obsolete and not merged.
3. For branches that should stay (long-lived experiments), document the reason in the roadmap or issue tracker.

## Remote Branch Cleanup
1. List remote branches that are fully merged:
   ```bash
   git branch -r --merged origin/main | grep -v main
   ```
2. For each candidate, confirm the merge via:
   ```bash
   git log <branch> --not main
   ```
   If no commits appear, the branch is safe to delete.
3. Delete remote branches that are no longer needed:
   ```bash
   git push origin --delete <branch>
   ```
4. Record the deleted branches and rationale in a short maintenance note (issue or doc) for traceability.

## Verification
- Run `git status` to ensure the working tree is clean.
- Verify the GitHub branches view matches expectations.
- Ensure CI pipelines referencing deleted branches are retired or updated.

## Governance Notes
- Keep `main` protected; require PR reviews and status checks before merge.
- Avoid resurrecting deleted branches; create fresh branches from `main` to maintain a clear history.
- If automation relies on branch naming conventions, document them here and in `../.llm/AGENTS.md` so agents follow the rules.

## Follow-up
- Update stakeholders on the cleanup results (e.g., number of branches removed, outstanding follow-ups).
- Schedule the next branch audit cadence (monthly, per release, etc.).

Maintaining a tidy branch list reduces confusion for humans and automation alike, and keeps the roadmap aligned with actual work in flight.
