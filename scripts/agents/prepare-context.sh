#!/usr/bin/env bash
# prepare-context.sh — Generate a concise repo snapshot for AI sessions
# Usage:
#   scripts/agents/prepare-context.sh [--since 2d]
# Examples:
#   scripts/agents/prepare-context.sh > docs/process/sessions/_latest_context.md

set -euo pipefail

SINCE="24 hours ago"
if [[ "${1:-}" == "--since" && -n "${2:-}" ]]; then
  SINCE="$2"
  shift 2
fi

echo "# Repo Context Snapshot" 
echo
echo "Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
echo

echo "## Git"
echo "- Branch: $(git rev-parse --abbrev-ref HEAD)"
echo "- HEAD: $(git log -1 --pretty=format:'%h %ad %s' --date=short)"
echo
echo "### Changes (uncommitted)"
git status -sb || true
echo
echo "### Recent commits (since ${SINCE})"
git log --since "${SINCE}" --oneline || true
echo

echo "## Make Targets (help)"
if grep -q "^help:" Makefile 2>/dev/null; then
  make help || true
else
  grep '^[a-zA-Z0-9_.-]*:' Makefile 2>/dev/null | cut -d: -f1 | sort -u || true
fi
echo

echo "## Backend Packages"
if [[ -d backend ]]; then
  find backend -maxdepth 2 -type d \( -name internal -o -name cmd \) -print | sed 's/^/- /'
fi
echo

echo "## ADRs"
if [[ -d docs/process/adr ]]; then
  for f in docs/process/adr/*.md; do
    [ -e "$f" ] || continue
    title=$(head -n1 "$f" | sed 's/^# //')
    status=$(rg -n "^-[[:space:]]*\*\*Status\*\*:" -N "$f" | sed 's/.*Status**: *//')
    echo "- ${title} — ${status:-unknown} (${f})"
  done
fi
echo

echo "## Open Issues (gh)"
if command -v gh >/dev/null 2>&1; then
  gh issue list --limit 20 || echo "gh issue list failed (not authenticated?)"
else
  echo "gh not installed; skipping."
fi
echo

echo "## Next Suggested Actions"
echo "- Run weekly planning: docs/process/sessions/PLANNING.md"
echo "- If a decision is pending, schedule a panel: docs/process/sessions/DECISION_PANEL.md"
echo "- For daily sync, use: docs/process/sessions/STANDUP.md"
