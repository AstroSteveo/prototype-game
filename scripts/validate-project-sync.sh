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
if [[ ! -f "$TEMPLATES_DIR/story.yml" ]] || ! grep -q 'labels: \["story"\]' "$TEMPLATES_DIR/story.yml"; then
    echo "❌ Story template missing or incorrect label"
    exit 1
fi
echo "✅ Story template has correct label"

if [[ ! -f "$TEMPLATES_DIR/task.yml" ]] || ! grep -q 'labels: \["task"\]' "$TEMPLATES_DIR/task.yml"; then
    echo "❌ Task template missing or incorrect label"
    exit 1
fi
echo "✅ Task template has correct label"

if [[ ! -f "$TEMPLATES_DIR/bug.yml" ]] || ! grep -q 'labels: \["bug"\]' "$TEMPLATES_DIR/bug.yml"; then
    echo "❌ Bug template missing or incorrect label"
    exit 1
fi
echo "✅ Bug template has correct label"

# Check documentation content
echo "📖 Checking documentation content..."

if ! grep -q "https://github.com/users/AstroSteveo/projects/2" "$DOCS_FILE"; then
    echo "❌ Documentation missing correct PROJECT_URL"
    exit 1
fi
echo "✅ Documentation contains correct PROJECT_URL"

if ! grep -q "scopes.*project.*repo" "$DOCS_FILE"; then
    echo "❌ Documentation missing correct token scopes"
    exit 1
fi
echo "✅ Documentation contains correct token scopes"

echo ""
echo "🎉 All validation checks passed!"
echo ""
echo "📋 Next steps to complete configuration:"
echo "1. Set repository variable PROJECT_URL to: https://github.com/users/AstroSteveo/projects/2"
echo "2. Set repository secret PROJECTS_TOKEN with scopes: project, repo"
echo "3. Test with a labeled issue to verify workflow execution"
echo ""
echo "📚 See docs/process/PROJECTS.md for detailed validation checklist"