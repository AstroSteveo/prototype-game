#!/bin/bash
# Agent onboarding validation script
# This script helps validate that an AI agent can access all necessary repository resources

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

echo "🤖 Agent Onboarding Validation"
echo "==============================="
echo "Repository: $(pwd)"
echo ""

# Function to check file exists and is readable
check_file() {
    local file="$1"
    local description="$2"
    if [[ -f "$file" && -r "$file" ]]; then
        echo "✅ $description: $file"
        return 0
    else
        echo "❌ $description: $file (missing or unreadable)"
        return 1
    fi
}

# Function to check directory exists
check_dir() {
    local dir="$1"
    local description="$2"
    if [[ -d "$dir" ]]; then
        echo "✅ $description: $dir"
        return 0
    else
        echo "❌ $description: $dir (missing)"
        return 1
    fi
}

VALIDATION_FAILED=0

echo "📋 Checking AGENTS.md hierarchy..."
check_file "AGENTS.md" "Root AGENTS.md" || VALIDATION_FAILED=1
check_file "docs/AGENTS.md" "Docs AGENTS.md" || VALIDATION_FAILED=1
check_file "backend/AGENTS.md" "Backend AGENTS.md" || VALIDATION_FAILED=1
check_file "docs/.llm/AGENTS.md" "LLM AGENTS.md" || VALIDATION_FAILED=1
echo ""

echo "📚 Checking core documentation..."
check_file "README.md" "Repository README" || VALIDATION_FAILED=1
check_file "docs/README.md" "Documentation README" || VALIDATION_FAILED=1
check_file "docs/development/developer-guide.md" "Developer Guide" || VALIDATION_FAILED=1
check_file "docs/product/vision/game-design-document.md" "Game Design Document" || VALIDATION_FAILED=1
check_file "docs/architecture/technical-design-document.md" "Technical Design Document" || VALIDATION_FAILED=1
echo ""

echo "🤖 Checking LLM onboarding framework..."
check_file "docs/.llm/onboarding/quick-start.md" "Quick Start Guide" || VALIDATION_FAILED=1
check_file "docs/.llm/onboarding/contribution-checklist.md" "Contribution Checklist" || VALIDATION_FAILED=1
check_file "docs/.llm/onboarding/copilot-playbook.md" "Copilot Playbook" || VALIDATION_FAILED=1
check_file "docs/.llm/onboarding/story-template.md" "Story Template" || VALIDATION_FAILED=1
check_file "docs/.llm/onboarding/agent-validation-checklist.md" "Agent Validation Checklist" || VALIDATION_FAILED=1
check_file "docs/.llm/onboarding/file-organization-guide.md" "File Organization Guide" || VALIDATION_FAILED=1
echo ""

echo "🔧 Checking build infrastructure..."
check_file "Makefile" "Makefile" || VALIDATION_FAILED=1
check_file "scripts/agents/prepare-context.sh" "Context Preparation Script" || VALIDATION_FAILED=1
echo ""

echo "🐙 Checking GitHub integration..."
check_dir ".github/ISSUE_TEMPLATE" "Issue Templates Directory" || VALIDATION_FAILED=1
check_file ".github/copilot-instructions.md" "Copilot Instructions" || VALIDATION_FAILED=1
check_file ".github/ISSUE_TEMPLATE/config.yml" "Issue Template Config" || VALIDATION_FAILED=1
echo ""

echo "🔍 Checking for known broken references..."
BROKEN_REFS=0

# Check for outdated DEV.md references (excluding this validation script and the checklist)
if grep -r "docs/dev/DEV.md" --exclude-dir=.git --exclude-dir=wiki-content --exclude="$(basename "$0")" --exclude="agent-validation-checklist.md" . > /dev/null 2>&1; then
    echo "❌ Found references to non-existent docs/dev/DEV.md"
    BROKEN_REFS=1
    VALIDATION_FAILED=1
fi

# Check for outdated design doc references (excluding this validation script and the checklist)
if grep -Er "docs/design/GDD\.md|docs/design/TDD\.md" --exclude-dir=.git --exclude-dir=wiki-content --exclude="$(basename "$0")" --exclude="agent-validation-checklist.md" --exclude="file-organization-guide.md" . > /dev/null 2>&1; then
    echo "❌ Found references to non-existent docs/design/ files"
    BROKEN_REFS=1
    VALIDATION_FAILED=1
fi

if [[ $BROKEN_REFS -eq 0 ]]; then
    echo "✅ No broken references found"
fi
echo ""

echo "📁 Checking file organization..."
# Check for potential duplicate files (simplified check)
DUPLICATE_CHECK=0

# Look for files with similar names that might be duplicates
SIMILAR_FILES=$(find docs/ -name "*.md" | xargs basename -s .md | sort | uniq -d 2>/dev/null || true)
if [[ -n "$SIMILAR_FILES" ]]; then
    echo "⚠️  Found files with similar names (review for potential duplicates):"
    echo "$SIMILAR_FILES" | sed 's/^/    /'
    DUPLICATE_CHECK=1
fi

# Check for common problematic patterns
if find docs/ -name "*-new.md" -o -name "*-old.md" -o -name "*-backup.md" -o -name "*-copy.md" 2>/dev/null | grep -q .; then
    echo "⚠️  Found files with naming patterns that suggest duplicates:"
    find docs/ -name "*-new.md" -o -name "*-old.md" -o -name "*-backup.md" -o -name "*-copy.md" 2>/dev/null | sed 's/^/    /' || true
    DUPLICATE_CHECK=1
fi

if [[ $DUPLICATE_CHECK -eq 0 ]]; then
    echo "✅ No obvious file organization issues found"
fi
echo ""

# Test basic make functionality if available
echo "🔨 Testing make infrastructure..."
if command -v make >/dev/null 2>&1; then
    if make help >/dev/null 2>&1; then
        echo "✅ Make help works"
    else
        echo "❌ Make help failed"
        VALIDATION_FAILED=1
    fi
else
    echo "⚠️  Make not available (this may be expected in some environments)"
fi
echo ""

# Test context script
echo "📝 Testing context preparation..."
if [[ -x "scripts/agents/prepare-context.sh" ]]; then
    # Run script and capture both output and exit code using a temporary file
    CONTEXT_TMP=$(mktemp)
    if ./scripts/agents/prepare-context.sh >"$CONTEXT_TMP" 2>&1 && grep -q "Generated:" "$CONTEXT_TMP"; then
        echo "✅ Context preparation script works"
    elif grep -q "Generated:" "$CONTEXT_TMP"; then
        echo "⚠️  Context preparation script works but has warnings (this is expected if ripgrep is not available)"
    else
        echo "❌ Context preparation script failed"
        echo "Output:"
        cat "$CONTEXT_TMP"
        VALIDATION_FAILED=1
    fi
    rm -f "$CONTEXT_TMP"
else
    echo "❌ Context preparation script not executable"
    VALIDATION_FAILED=1
fi
echo ""

# Summary
echo "📊 Validation Summary"
echo "===================="
if [[ $VALIDATION_FAILED -eq 0 ]]; then
    echo "🎉 All validations passed! Agent is ready to work with this repository."
    echo ""
    echo "Next steps:"
    echo "1. Review docs/.llm/onboarding/quick-start.md for workflow guidance"
    echo "2. Follow docs/.llm/onboarding/contribution-checklist.md before making changes"
    echo "3. Use scripts/agents/prepare-context.sh to generate repository context"
    exit 0
else
    echo "⚠️  Some validations failed. Please review the errors above."
    echo ""
    echo "Common fixes:"
    echo "- Ensure you're in the repository root directory"
    echo "- Check file permissions and accessibility"
    echo "- Verify the repository was cloned completely"
    echo "- For build issues, ensure Go 1.23+ is available"
    exit 1
fi