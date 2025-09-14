---
title: Global Branch Prune (2025-09-14)
date: 2025-09-14
related_issue: #64
---

# Summary

Prunes all non-`main` branches locally and on the remote, with per-branch analysis and rationale.

## Environment

- Repo: AstroSteveo/prototype-game
- Policy: Only `main` should exist (local and remote) absent open PRs or active work.

## Branch Analyses

### Local: `feat/us-502-reconnect-resume`
- Status: Fully merged via PR #53 (2025-09-14T05:28:56Z) → safe to delete.
- Verification: `git merge-base --is-ancestor feat/us-502-reconnect-resume main` → true.
- Action: Delete local branch.

### Local: `code-code-us-502--reconnect`
- Status: No remote; all commits are ancestors of `main`.
- Verification: `git merge-base --is-ancestor code-code-us-502--reconnect main` → true.
- Action: Delete local branch.

### Remote: `fix/engine-idempotent-stop-17`
- Status: PR #40 merged on 2025-09-13T20:55:00Z; no unique commits vs `origin/main`.
- Diff: none (fully contained).
- Action: Delete remote branch.

### Remote: `codex/review-hardcoded-retargetmax-in-bots.go`
- Status: 1 unique commit (`a4739e3`) changing `backend/internal/sim/bots.go` (1 line).
- PRs: none. Change is cosmetic/clarifying and not on roadmap; deferring.
- Action: Delete remote branch to reduce clutter. If needed later, reintroduce via a focused PR.

## Actions Performed

1. Created this analysis document for auditability.
2. Will merge this PR and prune branches accordingly so only `main` remains.

## Rationale

Maintains a clean branch model and avoids confusion about the source of truth.

