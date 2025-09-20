#!/usr/bin/env zsh
set -euo pipefail

# rename-milestones.zsh
# Interactively rename GitHub milestones to outcome-focused, purpose-driven names.
# Requirements: GitHub CLI (gh) authenticated with repo scope; git remote set to target repo.

command -v gh >/dev/null 2>&1 || {
  echo "Error: GitHub CLI (gh) is required. See https://cli.github.com/" >&2
  exit 1
}
command -v jq >/dev/null 2>&1 || {
  echo "Error: jq is required. Install jq (e.g., sudo pacman -S jq or brew install jq)." >&2
  exit 1
}

# Derive owner/repo
if ! repo_json=$(gh repo view --json nameWithOwner 2>/dev/null); then
  echo "Error: Unable to view repo via gh. Ensure you are in the repository directory and authenticated." >&2
  exit 1
fi
owner_repo=$(echo "$repo_json" | jq -r .nameWithOwner)
echo "Target repository: $owner_repo"

echo "Fetching milestones..."
milestones=$(gh api repos/$owner_repo/milestones --paginate)
count=$(echo "$milestones" | jq 'length')
if [[ "$count" -eq 0 ]]; then
  echo "No milestones found. You can create one with: gh api -X POST repos/$owner_repo/milestones -f title='Release: <Outcome>' -f description='...' -f due_on='2025-10-31T00:00:00Z'" 
  exit 0
fi

echo
echo "Current milestones:" 
echo "$milestones" | jq -r '.[] | "#\(.number) | \(.title) | state=\(.state) | due_on=\(.due_on // "-")"'
echo
echo "Tip: Use descriptive names focused on the outcome/problem being addressed (e.g., 'Reduce matchmaking latency', 'Seamless cross-region handover')."
echo

while true; do
  vared -p "Enter milestone number to rename (or press Enter to finish): " -c num
  [[ -z "${num:-}" ]] && break
  if ! echo "$milestones" | jq -e ".[] | select(.number == ${num})" >/dev/null 2>&1; then
    echo "Milestone #$num not found. Try again." >&2
    continue
  fi
  current_title=$(echo "$milestones" | jq -r ".[] | select(.number == ${num}) | .title")
  vared -p "New title for '#$num: $current_title': " -c new_title
  [[ -z "${new_title:-}" ]] && { echo "Title cannot be empty. Skipping."; continue }
  vared -p "Optional new description (Enter to skip): " -c new_desc
  vared -p "Optional new due date ISO8601 (e.g., 2025-10-31T00:00:00Z) (Enter to skip): " -c new_due

  args=( -X PATCH repos/$owner_repo/milestones/$num -f title="$new_title" )
  [[ -n "${new_desc:-}" ]] && args+=( -f description="$new_desc" )
  [[ -n "${new_due:-}" ]] && args+=( -f due_on="$new_due" )

  echo "Renaming milestone #$num → '$new_title'..."
  gh api $args
  echo "Done."

  # Refresh list for subsequent iterations
  milestones=$(gh api repos/$owner_repo/milestones --paginate)
done

echo "All done. Updated milestones:"
echo "$milestones" | jq -r '.[] | "#\(.number) | \(.title) | state=\(.state) | due_on=\(.due_on // "-")"'
