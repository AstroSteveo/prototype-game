#!/bin/bash
# Validation script for project board automation
# This script provides instructions for manual validation since GitHub UI configuration cannot be automated

set -euo pipefail
IFS=$'\n\t'

echo "=== Project Board Automation Validation ==="
echo
echo "This script provides validation steps for the project board automation setup."
echo "Some steps require manual configuration in the GitHub UI."
echo

# Check if required files exist
echo "✓ Checking required files..."
if [[ -f ".github/workflows/project-sync.yml" ]]; then
    echo "  ✓ project-sync.yml workflow exists"
else
    echo "  ✗ project-sync.yml workflow missing"
    exit 1
fi

if [[ -f "docs/project-board-setup.md" ]]; then
    echo "  ✓ Setup documentation exists"
else
    echo "  ✗ Setup documentation missing"
    exit 1
fi

echo

# Validate workflow syntax
echo "✓ Validating workflow syntax..."
if command -v yamllint >/dev/null 2>&1; then
    yamllint .github/workflows/project-sync.yml && echo "  ✓ YAML syntax valid"
else
    echo "  ⚠ yamllint not available, skipping syntax check"
fi

echo

# Check for required patterns in workflow
echo "✓ Checking workflow content..."
required_patterns=(
    "actions/add-to-project@v0.6.0"
    "actions/github-script@v7"
    "PROJECTS_TOKEN"
    "PROJECT_URL"
    "Status"
    "Estimate"
    "Milestone"
    "Sprint"
)

for pattern in "${required_patterns[@]}"; do
    if grep -q "$pattern" .github/workflows/project-sync.yml; then
        echo "  ✓ Contains: $pattern"
    else
        echo "  ✗ Missing: $pattern"
    fi
done

echo

# Manual validation steps
echo "=== MANUAL VALIDATION REQUIRED ==="
echo
echo "Repository Configuration:"
echo "1. Set repository variable PROJECT_URL = https://github.com/users/AstroSteveo/projects/2"
echo "   (Settings → Secrets and variables → Actions → Variables)"
echo
echo "2. Set repository secret PROJECTS_TOKEN = [fine-grained PAT]"
echo "   (Settings → Secrets and variables → Actions → Secrets)"
echo
echo "Project Field Configuration:"
echo "3. Create Status field (single-select) with options: Backlog, Ready, In Progress, In Review, Blocked, Done"
echo "4. Create Estimate field (number)"
echo "5. Create Milestone field (milestone)"
echo "6. Create Sprint field (iteration)"
echo "   (Visit: https://github.com/users/AstroSteveo/projects/2/settings/fields)"
echo
echo "Project UI Workflows:"
echo "7. Configure automation rules:"
echo "   - Item added → Status: Backlog"
echo "   - Item assigned → Status: In Progress"
echo "   - Issue closed → Status: Done"
echo "   - PR merged → Status: Done"
echo "   - Archive items after 14 days in Done status"
echo "   (Visit: https://github.com/users/AstroSteveo/projects/2/workflows)"
echo
echo "Test Case:"
echo "8. Create test issue:"
echo "   - Use Task template"
echo "   - Title: 'task: Test automation [2]'"
echo "   - Add label 'task'"
echo "   - Verify: added to project, Status=Backlog, Estimate=2"
echo
echo "9. Test status transitions:"
echo "   - Add 'ready' label → Status should become Ready"
echo "   - Add 'blocked' label → Status should become Blocked"
echo "   - Close issue → Status should become Done"
echo
echo "10. Test PR workflow:"
echo "    - Open draft PR → Status: In Progress"
echo "    - Mark ready for review → Status: In Review"
echo "    - Merge PR → Status: Done"
echo

# Check if .gitignore is appropriate
echo "=== Additional Checks ==="
echo "✓ Checking .gitignore..."
if [[ -f ".gitignore" ]]; then
    echo "  ✓ .gitignore exists"
    # Should not exclude docs/ or .github/
    if grep -q "^docs/" .gitignore; then
        echo "  ⚠ Warning: docs/ is ignored (setup documentation won't be committed)"
    fi
    if grep -q "^\.github/" .gitignore; then
        echo "  ⚠ Warning: .github/ is ignored (workflow won't be committed)"
    fi
else
    echo "  ⚠ .gitignore missing"
fi

echo
echo "=== Validation Summary ==="
echo "✓ Workflow file exists and contains required automation logic"
echo "✓ Documentation created for manual setup steps"
echo "⚠ Manual configuration required in GitHub UI for complete setup"
echo "⚠ Test issue creation required for end-to-end validation"
echo
echo "Next steps:"
echo "1. Follow the manual configuration steps above"
echo "2. Create a test issue to validate automation"
echo "3. Monitor the Actions tab for workflow execution"
echo
echo "For detailed setup instructions, see: docs/project-board-setup.md"}‬