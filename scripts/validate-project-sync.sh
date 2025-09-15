#!/bin/bash

# Project Sync Configuration Validator
# This script validates the project sync configuration and documentation

set -e

echo "🔍 Validating Project Sync Configuration..."

# Check if required files exist
echo "📁 Checking required files..."

DOCS_FILE="docs/process/PROJECTS.md"
WORKFLOW_FILE=".github/workflows/project-sync.yml"

if [[ ! -f "$DOCS_FILE" ]]; then
    echo "❌ Missing documentation: $DOCS_FILE"
    exit 1
fi
echo "✅ Found documentation: $DOCS_FILE"

if [[ ! -f "$WORKFLOW_FILE" ]]; then
    echo "❌ Missing workflow: $WORKFLOW_FILE"
    exit 1
fi
echo "✅ Found workflow: $WORKFLOW_FILE"

# Check workflow configuration
echo "⚙️  Checking workflow configuration..."

# Check for required variables and secrets
if ! grep -q "vars.PROJECT_URL" "$WORKFLOW_FILE"; then
    echo "❌ Workflow missing PROJECT_URL variable reference"
    exit 1
fi
echo "✅ Workflow references PROJECT_URL variable"

if ! grep -q "secrets.PROJECTS_TOKEN" "$WORKFLOW_FILE"; then
    echo "❌ Workflow missing PROJECTS_TOKEN secret reference"
    exit 1
fi
echo "✅ Workflow references PROJECTS_TOKEN secret"

# Check for correct labels
if ! grep -q "labeled: story, bug, task" "$WORKFLOW_FILE"; then
    echo "❌ Workflow missing or incorrect labeled configuration"
    exit 1
fi
echo "✅ Workflow configured for story, bug, task labels"

# Check issue templates
echo "🏷️  Checking issue templates..."

TEMPLATES_DIR=".github/ISSUE_TEMPLATE"

validate_template_label() {
    local template_name="$1"
    local expected_label="$2"
    local template_file="$TEMPLATES_DIR/${template_name}.yml"
    if [[ ! -f "$template_file" ]] || ! grep -q "labels: \[\"$expected_label\"\]" "$template_file"; then
        echo "❌ ${template_name^} template missing or incorrect label"
        exit 1
    fi
    echo "✅ ${template_name^} template has correct label"
}

validate_template_label "story" "story"
validate_template_label "task" "task"
validate_template_label "bug" "bug"

# Check documentation content
echo "📖 Checking documentation content..."

if ! grep -q "https://github.com/users/AstroSteveo/projects/2" "$DOCS_FILE"; then
    echo "❌ Documentation missing correct PROJECT_URL"
    exit 1
fi
echo "✅ Documentation contains correct PROJECT_URL"

if ! grep -q "scopes:.*project.*repo" "$DOCS_FILE"; then
    echo "❌ Documentation missing correct token scopes"
    exit 1
fi
echo "✅ Documentation contains correct token scopes"

# Test URL parsing logic
echo "🧪 Testing URL parsing logic..."
if command -v node >/dev/null 2>&1 && [[ -f "scripts/test-project-url-parsing.js" ]]; then
    if node scripts/test-project-url-parsing.js >/dev/null 2>&1; then
        echo "✅ URL parsing logic tests pass"
    else
        echo "❌ URL parsing logic tests failed"
        exit 1
    fi
else
    echo "⚠️  Skipping URL parsing tests (Node.js not available or test script missing)"
fi

echo ""
echo "🎉 All validation checks passed!"
echo ""
echo "📋 Next steps to complete configuration:"
echo "1. Set repository variable PROJECT_URL to: https://github.com/users/AstroSteveo/projects/2"
echo "2. Set repository secret PROJECTS_TOKEN with scopes: project, repo"
echo "3. Test with a labeled issue to verify workflow execution"
echo ""
echo "📚 See docs/process/PROJECTS.md for detailed validation checklist"