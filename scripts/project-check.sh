#!/usr/bin/env bash
set -euo pipefail

# Verifies that the target Project has required fields: Status, Estimate, Sprint, Milestone.
# Requires: gh CLI authenticated (gh auth login) or GH_TOKEN set.
# Usage: PROJECT_URL=https://github.com/users/<login>/projects/<number> scripts/project-check.sh

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required. Install https://cli.github.com/ and run 'gh auth login'." >&2
  exit 1
fi

PROJECT_URL=${PROJECT_URL:-}
if [[ -z "$PROJECT_URL" ]]; then
  echo "Set PROJECT_URL to the Project v2 URL (users/orgs)." >&2
  exit 1
fi

parse_url() {
  local url=$1
  # Extract path parts
  local path=${url#*://*/}
  local parts
  IFS='/' read -r -a parts <<<"${url#*://github.com/}"
  if [[ ${#parts[@]} -ge 4 && ${parts[2]} == "projects" ]]; then
    echo "${parts[0]}" "${parts[1]}" "${parts[3]}"
    return 0
  fi
  return 1
}

read scope login number < <(parse_url "$PROJECT_URL") || { echo "Invalid PROJECT_URL: $PROJECT_URL" >&2; exit 1; }
if [[ -z "$scope" || -z "$login" || -z "$number" ]]; then
  echo "Invalid PROJECT_URL: $PROJECT_URL (parsed values are empty)" >&2
  exit 1
fi

read -r -d '' Q_USER <<'EOF'
query($login:String!, $number:Int!){
  user(login:$login){
    projectV2(number:$number){
      id
      fields(first:100){
        nodes{
          ... on ProjectV2Field { id name dataType }
          ... on ProjectV2SingleSelectField { id name dataType options{ id name } }
          ... on ProjectV2IterationField { id name dataType }
          ... on ProjectV2MilestoneField { id name dataType }
        }
      }
    }
  }
}
EOF

read -r -d '' Q_ORG <<'EOF'
query($login:String!, $number:Int!){
  organization(login:$login){
    projectV2(number:$number){
      id
      fields(first:100){
        nodes{
          ... on ProjectV2Field { id name dataType }
          ... on ProjectV2SingleSelectField { id name dataType options{ id name } }
          ... on ProjectV2IterationField { id name dataType }
          ... on ProjectV2MilestoneField { id name dataType }
        }
      }
    }
  }
}
EOF

JSON=$(gh api graphql -f query="$( [[ "$scope" == users ]] && echo "$Q_USER" || echo "$Q_ORG" )" -f login="$login" -F number="$number")

FIELDS=$(echo "$JSON" | jq -r '..|.fields?.nodes? // empty | .[] | .name')
echo "Project fields:" >&2
echo "$FIELDS" | sed 's/^/ - /' >&2

missing=0
for name in Status Estimate Sprint Milestone; do
  if ! echo "$FIELDS" | grep -qi "^$name$"; then
    echo "Missing field: $name" >&2
    missing=1
  fi
done

if [[ $missing -eq 0 ]]; then
  echo "All required fields present."
else
  exit 2
fi

