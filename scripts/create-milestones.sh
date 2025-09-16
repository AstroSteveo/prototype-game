#!/usr/bin/env bash
set -euo pipefail

# Create standard milestones (M1..M5) with optional due dates.
# Requires: gh CLI authenticated (gh auth login) or GH_TOKEN set.
# Usage:
#   scripts/create-milestones.sh
#   DUE_M1=2025-10-01 DUE_M2=2025-10-15 scripts/create-milestones.sh

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required. Install https://cli.github.com/ and run 'gh auth login'." >&2
  exit 1
fi

REPO=$(git config --get remote.origin.url | sed -E 's#(git@github.com:|https://github.com/)##; s#\.git$##')
if [[ -z "$REPO" ]]; then
  echo "Unable to determine repository (remote.origin.url)." >&2
  exit 1
fi

echo "Target repo: $REPO"

create_milestone() {
  local title=$1
  local due_var=$2
  local due=${!due_var:-}
  local args=(-f title="$title" -f state="open")
  if [[ -n "$due" ]]; then
    # Normalize YYYY-MM-DD to ISO 8601 midnight Z
    if [[ "$due" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
      due="${due}T00:00:00Z"
    fi
    args+=(-f due_on="$due")
  fi
  echo "Creating milestone: $title ${due:+(due $due)}"
  gh api -X POST \
    -H "Accept: application/vnd.github+json" \
    "/repos/$REPO/milestones" \
    "${args[@]}" >/dev/null || true
}

for n in 1 2 3 4 5; do
  create_milestone "M${n}" "DUE_M${n}"
done

echo "Done. Verify milestones at: https://github.com/$REPO/milestones"

