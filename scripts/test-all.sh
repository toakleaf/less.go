#!/bin/bash
# Run all Go tests (unit + integration) with concise output on success

# Colors (only if terminal)
if [ -t 1 ]; then
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    NC='\033[0m'
else
    GREEN=''
    RED=''
    NC=''
fi

# Temp files for capturing output
UNIT_OUTPUT=$(mktemp)
INT_OUTPUT=$(mktemp)
trap "rm -f $UNIT_OUTPUT $INT_OUTPUT" EXIT

# Track failures
FAILED=0

# Run unit tests (silently)
if ! go test ./less -run 'Test[^I]' -timeout 2m > "$UNIT_OUTPUT" 2>&1; then
    FAILED=1
fi

# Run integration tests in strict mode (silently)
# LESS_GO_STRICT=1 makes output differences cause actual test failures
# LESS_GO_QUIET=1 suppresses individual test output
if ! LESS_GO_STRICT=1 LESS_GO_QUIET=1 go test ./less -run 'TestIntegrationSuite' -timeout 5m > "$INT_OUTPUT" 2>&1; then
    FAILED=1
fi

# Check results
if [ $FAILED -eq 0 ]; then
    # Extract timing info (handles both "1.234s" and "(cached)")
    UNIT_TIME=$(grep -oE '([0-9.]+s|\(cached\))' "$UNIT_OUTPUT" | tail -1)
    INT_TIME=$(grep -oE '([0-9.]+s|\(cached\))' "$INT_OUTPUT" | tail -1)
    [ -z "$UNIT_TIME" ] && UNIT_TIME="ok"
    [ -z "$INT_TIME" ] && INT_TIME="ok"
    echo -e "${GREEN}✓ All Go tests passed${NC} (unit: ${UNIT_TIME}, integration: ${INT_TIME})"
else
    echo -e "${RED}✗ Tests failed${NC}"
    echo ""
    echo "=== Unit Test Output ==="
    cat "$UNIT_OUTPUT"
    echo ""
    echo "=== Integration Test Output ==="
    cat "$INT_OUTPUT"
    exit 1
fi
