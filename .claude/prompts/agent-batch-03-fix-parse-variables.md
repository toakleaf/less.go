# Agent Prompt 3: Fix Parser Failures in variables Test

**Priority**: CRITICAL
**Impact**: 1 test
**Time Estimate**: 2-3 hours
**Difficulty**: Medium

## Task

Fix the parse failure in `variables` test which currently fails with:
```
Parse: Unrecognised input in ../../../../test-data/less/_main/variables.less
```

This test is critical because variables are a core LESS feature used everywhere.

## What You Need to Do

1. **Setup**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-parse-variables-SESSION_ID
   pnpm -w test:go:filter -- "variables"
   ```

2. **Find the problematic syntax**:
   - Read: `packages/test-data/less/_main/variables.less`
   - Compare with: `packages/less/src/less/less/variables.less`
   - What variable syntax isn't being parsed?

3. **Debug Parser**:
   - Run with traces: `LESS_GO_TRACE=1 pnpm -w test:go:filter -- "variables"`
   - Identify exact line/syntax where parsing fails
   - Check if it's variable declaration, usage, or interpolation

4. **Fix the Parser**:
   - Update `packages/less/src/less/less_go/parser.go`
   - Add support for the missing variable syntax pattern
   - Test frequently: `pnpm -w test:go:filter -- "variables"`

5. **Regression Testing** (CRITICAL):
   - Variables are used everywhere - this is high-risk
   - Run: `pnpm -w test:go:unit` (all unit tests)
   - Run: `pnpm -w test:go` (full suite)
   - Look carefully for ANY test that regressed

6. **Commit**:
   ```bash
   git add -A
   git commit -m "Fix parser: Add support for [variable feature] pattern in variables test"
   git push -u origin claude/fix-parse-variables-SESSION_ID
   ```

## Success Criteria

- ✅ `variables` test compiles
- ✅ ZERO regressions (check other tests carefully!)
- ✅ All 184 tests still pass or fail as expected
- ✅ Unit tests: 100% passing

## Important Files

- Test: `packages/test-data/less/_main/variables.less`
- Reference: `packages/less/src/less/less/variables.less`
- Expected CSS: `packages/test-data/less/_main/variables.css`
- Parser code: `packages/less/src/less/less_go/parser.go`

## Critical Notes

⚠️ **HIGH RISK**: Variables are fundamental. ANY regression here breaks many tests.

- Test thoroughly before pushing
- Run full test suite minimum 3 times to be sure
- Variable syntax includes:
  - `@variable: value;`
  - `@name: @value;`
  - Interpolation: `@{...}`
  - Nested variables: `@{"key": value}`

Check the assigned test status before and after to ensure no regressions.
