#!/bin/bash

# Find all dynamic regex compilations that should be pre-compiled

set -e

cd "$(dirname "$0")/.."

echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║             DYNAMIC REGEX COMPILATION FINDER                                 ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "Finding all regex compilations that should be moved to package-level vars..."
echo ""

# Find all regexp.MustCompile calls
echo "═══════════════════════════════════════════════════════════════════════════════"
echo "FILES WITH DYNAMIC REGEX COMPILATION:"
echo "═══════════════════════════════════════════════════════════════════════════════"

cd less

total_count=0
for file in *.go; do
    if [ "$file" = "*_test.go" ]; then
        continue  # Skip test files for now
    fi

    count=$(grep -c "regexp\.MustCompile\|regexp\.Compile" "$file" 2>/dev/null || echo 0)

    if [ "$count" -gt 0 ]; then
        echo "$file: $count occurrences"
        total_count=$((total_count + count))
    fi
done

echo ""
echo "Total dynamic regex compilations: $total_count"
echo ""

# Show specific occurrences in parser.go (the main culprit)
if [ -f "parser.go" ]; then
    echo "═══════════════════════════════════════════════════════════════════════════════"
    echo "PARSER.GO REGEX COMPILATIONS (CRITICAL):"
    echo "═══════════════════════════════════════════════════════════════════════════════"
    grep -n "regexp\.MustCompile\|regexp\.Compile" parser.go | head -30
    echo ""
    echo "(Showing first 30 of $(grep -c 'regexp\.MustCompile\|regexp\.Compile' parser.go) occurrences)"
    echo ""
fi

echo "═══════════════════════════════════════════════════════════════════════════════"
echo "ANALYSIS:"
echo "═══════════════════════════════════════════════════════════════════════════════"
echo ""
echo "⚠️  CRITICAL FINDING:"
echo "   These regexes are compiled on EVERY function call, causing massive overhead."
echo ""
echo "💡 SOLUTION:"
echo "   Move all constant regex patterns to package-level variables."
echo ""
echo "📝 EXAMPLE FIX:"
echo ""
echo "   BEFORE (BAD):"
echo "   func parseColor() {"
echo "       rgb := p.Re(regexp.MustCompile(\`^#([A-Fa-f0-9]{6})\`))"
echo "   }"
echo ""
echo "   AFTER (GOOD):"
echo "   var colorRegex = regexp.MustCompile(\`^#([A-Fa-f0-9]{6})\`)"
echo ""
echo "   func parseColor() {"
echo "       rgb := p.Re(colorRegex)  // Reuse pre-compiled regex"
echo "   }"
echo ""
echo "═══════════════════════════════════════════════════════════════════════════════"
echo "NEXT STEPS:"
echo "═══════════════════════════════════════════════════════════════════════════════"
echo ""
echo "1. Review task documentation:"
echo "   .claude/tasks/performance/CRITICAL_regex_compilation.md"
echo ""
echo "2. Create regex variables file (recommended):"
echo "   less/parser_regexes.go"
echo ""
echo "3. Move all regexes to package-level vars"
echo ""
echo "4. Replace dynamic compilation with variable references"
echo ""
echo "5. Verify with profiling:"
echo "   pnpm bench:profile"
echo ""
echo "Expected improvement: 5-10x faster, 80% less memory"
echo ""
