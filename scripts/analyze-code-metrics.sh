#!/bin/bash

# Code Metrics Analysis Script for prototype-game
# This script analyzes the codebase and generates line count statistics

echo "=== PROTOTYPE GAME - CODE METRICS ANALYSIS ==="
echo "Analysis Date: $(date)"
echo "Repository: $(basename $(pwd))"
echo

# Find all source code files
echo "=== SOURCE CODE FILES ==="
find . -type f \( -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.jsx" \
    -o -name "*.tsx" -o -name "*.java" -o -name "*.c" -o -name "*.cpp" \
    -o -name "*.h" -o -name "*.hpp" -o -name "*.cs" -o -name "*.go" \
    -o -name "*.rs" -o -name "*.php" -o -name "*.rb" -o -name "*.swift" \
    -o -name "*.kt" -o -name "*.scala" -o -name "*.vue" -o -name "*.html" \
    -o -name "*.css" -o -name "*.scss" -o -name "*.sass" -o -name "*.less" \) \
    | wc -l | xargs echo "Total source code files:"

echo

# Count lines by language
echo "=== LINES BY LANGUAGE ==="
echo "Go files:"
go_lines=$(find . -name "*.go" -exec cat {} \; | wc -l)
echo "  $go_lines lines"

echo "JavaScript files:"  
js_lines=$(find . -name "*.js" -exec cat {} \; | wc -l)
echo "  $js_lines lines"

echo "Shell scripts:"
sh_lines=$(find . -name "*.sh" -exec cat {} \; | wc -l)
echo "  $sh_lines lines"

echo

# Breakdown production vs test code
echo "=== PRODUCTION vs TEST CODE ==="
prod_go_lines=$(find . -name "*.go" ! -name "*_test.go" -exec cat {} \; | wc -l)
test_go_lines=$(find . -name "*_test.go" -exec cat {} \; | wc -l)

echo "Production Go code: $prod_go_lines lines"
echo "Test Go code: $test_go_lines lines"

total_prod=$((prod_go_lines + js_lines))
echo "Total production code: $total_prod lines"
echo "Total test code: $test_go_lines lines"

echo

# Calculate totals
total_code_lines=$((go_lines + js_lines + sh_lines))
echo "=== SUMMARY ==="
echo "TOTAL CODE (Go + JS + Shell): $total_code_lines lines"

# Calculate percentages
if [ $total_code_lines -gt 0 ]; then
    prod_percent=$((total_prod * 100 / total_code_lines))
    test_percent=$((test_go_lines * 100 / total_code_lines))
    echo "Production code: $prod_percent%"
    echo "Test code: $test_percent%"
fi

echo

# File counts
echo "=== FILE COUNTS ==="
echo "Go files: $(find . -name "*.go" | wc -l)"
echo "  - Production: $(find . -name "*.go" ! -name "*_test.go" | wc -l)"
echo "  - Test: $(find . -name "*_test.go" | wc -l)"
echo "JavaScript files: $(find . -name "*.js" | wc -l)"
echo "Shell scripts: $(find . -name "*.sh" | wc -l)"

echo
echo "=== LARGEST FILES ==="
echo "Top 5 production Go files:"
find . -name "*.go" ! -name "*_test.go" -exec wc -l {} \; | sort -nr | head -5

echo
echo "Top 5 test files:"
find . -name "*_test.go" -exec wc -l {} \; | sort -nr | head -5

echo
echo "Analysis complete. For detailed report, see docs/metrics/code-metrics-report.md"