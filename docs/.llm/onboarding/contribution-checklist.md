# Contribution Checklist for Automation

Use this checklist before finalizing any automated change.

## Branching & Git Hygiene
- [ ] Branch created from the latest `main`.
- [ ] Branch name follows `feature/<slug>` or `docs/<slug>` conventions; delete branch after merge.
- [ ] Commits are scoped and use imperative messages.

## Change Preparation
- [ ] Verified applicable instructions in `AGENTS.md` files for all touched directories.
- [ ] Updated `docs/README.md` if new long-lived docs were added or reorganized.
- [ ] Cross-links in documentation updated to new paths.

## Validation
- [ ] Required commands executed (see task instructions; typically `make fmt vet test test-ws` for code, or note N/A for doc-only).
- [ ] Documented command output or explanation when a check is intentionally skipped.
- [ ] Performed spell check and formatting review for Markdown content.

## PR Package
- [ ] Summary references updated files with citations.
- [ ] Tests/validation listed explicitly in the PR description.
- [ ] Linked to relevant issues, roadmap updates, or ADRs.

Automations must treat this checklist as a blocking gate before requesting human review.
