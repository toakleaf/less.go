# Integration Test Guide

This guide explains how to use the improved integration test output system for less.go.

## Quick Reference

### Most Common Commands

```bash
# Get summary only (recommended for LLMs and humans)
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

# Run all tests with verbose output
pnpm -w test:go

# Debug a specific failing test
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/<suite>/<test>

# Get JSON output for programmatic analysis
LESS_GO_JSON=1 LESS_GO_QUIET=1 pnpm -w test:go
```

## Environment Variables

| Variable | Purpose | When to Use |
|----------|---------|-------------|
| `LESS_GO_QUIET=1` | Suppress individual test logs | When you only want the summary |
| `LESS_GO_DEBUG=1` | Show enhanced debug info | When investigating test failures |
| `LESS_GO_DIFF=1` | Show side-by-side CSS diffs | When comparing expected vs actual output |
| `LESS_GO_TRACE=1` | Show evaluation trace | When debugging runtime issues |
| `LESS_GO_JSON=1` | Output as JSON | For programmatic parsing |
| `LESS_GO_STRICT=1` | Fail on any difference | For CI/CD pipelines |

## Understanding Test Categories

Tests are automatically categorized based on their behavior:

### ✅ Perfect CSS Matches
Tests that compile successfully and produce CSS identical to less.js.
- **Goal**: Maximize this number
- **Priority**: Celebrate and maintain these!

### ❌ Compilation Failures
Tests that fail to compile due to parser or runtime errors.
- **Priority**: HIGHEST - fix these first
- **Impact**: Blocks functionality

### ⚠️ Output Differences
Tests that compile successfully but produce different CSS than less.js.
- **Priority**: MEDIUM - fix after compilation failures
- **Impact**: Wrong CSS output

### ✅ Correctly Failed
Error tests that properly fail as expected.
- **Priority**: Maintain - these are working correctly
- **Impact**: Error handling is working

### ⚠️ Expected Error
Error tests that should fail but succeed.
- **Priority**: LOW - fix after output differences
- **Impact**: Missing error handling

### ⏸️ Quarantined
Tests for features not yet implemented (plugins, JavaScript execution).
- **Priority**: Future work
- **Impact**: Not counted in metrics

## Reading the Summary

The test summary provides structured output in three sections:

### 1. Quick Stats
```
OVERALL SUCCESS: 142/184 tests (77.2%)

✅ Perfect CSS Matches:       80  (43.5% of active tests)
❌ Compilation Failures:       3  (1.6% of active tests)
⚠️  Output Differences:        12  (6.5% of active tests)
✅ Correctly Failed (Error):  62  (33.7% of active tests)
⚠️  Expected Error:            27  (14.7% of active tests)
```

**Key Metrics:**
- **OVERALL SUCCESS**: Perfect matches + correctly failed error tests
- **Perfect CSS Matches**: Tests producing correct output
- **Compilation Rate**: Percentage of tests that compile (currently 98.4%)

### 2. Detailed Results by Category

Tests are grouped by suite for easy navigation:
```
✅ PERFECT CSS MATCHES (80 tests) - Fully working!
   main: (50 tests)
     - calc
     - charsets
     - colors
     ...
   namespacing: (11 tests)
     - namespacing-1
     - namespacing-2
     ...
```

### 3. Next Steps

Prioritized action items:
```
💡 NEXT STEPS

PRIORITY: Fix 3 compilation failures first
  • main: 1 tests
  • third-party: 1 tests
  • process-imports: 1 tests

MEDIUM: Fix 12 output differences
  • main: 10 tests
  • static-urls: 1 tests
  ...
```

## Detecting Regressions

### Before Creating a PR

Always check these metrics:
1. **Perfect CSS Matches**: Should increase or stay same
2. **Compilation Failures**: Should decrease or stay same
3. **Overall Success Rate**: Should increase or stay same

### Regression Detected

If any of these conditions occur:
- Perfect CSS Matches decreased → REGRESSION
- Compilation Failures increased → REGRESSION
- Overall Success Rate decreased → REGRESSION

**Action**: DO NOT create PR. Fix the regression first.

## Updating Documentation

After fixing tests, update `CLAUDE.md`:

1. Run the summary:
   ```bash
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
   ```

2. Update the "Current Integration Test Status" section with:
   - New perfect match count and percentage
   - New compilation failure count
   - New overall success rate
   - Update the date

3. Add newly fixed tests to "Recent Progress"

## Example Workflow

### Debugging a Failing Test

```bash
# 1. See which category the test is in
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

# 2. Run the specific test with diffs
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/main/import-reference

# 3. Trace execution to understand the issue
LESS_GO_TRACE=1 go test -v -run TestIntegrationSuite/main/import-reference

# 4. Fix the code

# 5. Verify the fix
go test -v -run TestIntegrationSuite/main/import-reference

# 6. Check for regressions
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

### Checking Overall Progress

```bash
# Quick check
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "OVERALL SUCCESS"

# Full summary
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

# JSON for analysis
LESS_GO_JSON=1 LESS_GO_QUIET=1 pnpm -w test:go
```

## Tips for LLMs

When analyzing test results:

1. **Always use QUIET mode** to reduce noise:
   ```bash
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
   ```

2. **Focus on categories** rather than individual test output

3. **Look for patterns** in grouped test failures (e.g., all namespacing tests)

4. **Check the "NEXT STEPS"** section for prioritized work

5. **Use JSON mode** for programmatic analysis:
   ```bash
   LESS_GO_JSON=1 LESS_GO_QUIET=1 pnpm -w test:go
   ```

6. **Compare metrics** before and after changes to detect regressions

## Common Issues

### "Too much output"
**Solution**: Use `LESS_GO_QUIET=1` to suppress individual test logs

### "Can't find failing test"
**Solution**: Check the "Detailed Results" section in the summary - tests are grouped by category and suite

### "Want to see CSS diff"
**Solution**: Use `LESS_GO_DIFF=1` when running individual tests

### "Need to trace execution"
**Solution**: Use `LESS_GO_TRACE=1` for runtime tracing

### "Binary file matches" error with grep
**Solution**: This is normal - the test output contains emoji characters. Use `tail` instead or the built-in summary.

## Changes in This Update

### Improvements Made

1. **Structured Summary**: Clear categorization of all test results
2. **Environment Variables**: Fine-grained control over output
3. **LLM-Friendly**: Easy to parse output with percentages and counts
4. **Regression Detection**: Clear metrics to track progress
5. **Quick Commands**: Copy-paste commands for common tasks
6. **Grouped Results**: Tests grouped by suite and category

### Migration Notes

Old commands still work, but new commands are recommended:

| Old | New (Recommended) |
|-----|-------------------|
| `pnpm -w test:go \| grep summary` | `LESS_GO_QUIET=1 pnpm -w test:go 2>&1 \| tail -100` |
| Multiple grep commands | Single summary command |
| Manual counting | Automatic categorization |

## Future Enhancements

Planned improvements:
- [ ] HTML report generation
- [ ] Historical trend tracking
- [ ] Automatic regression detection in CI
- [ ] Test timing analytics
- [ ] Suite-specific filtering

---

For more details, see:
- `CLAUDE.md` - Section 4: How to Use Integration Tests Effectively
- `.claude/strategy/agent-workflow.md` - Step 8: Validate Fix
