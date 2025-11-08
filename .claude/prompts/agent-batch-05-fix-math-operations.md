# Agent Prompt 5: Fix Math Operations Output Differences

**Priority**: HIGH
**Impact**: 10+ tests across multiple suites
**Time Estimate**: 3-4 hours
**Difficulty**: Medium-High

## Task

Fix output differences in math operation tests. These tests now compile successfully (thanks to previous fixes!) but produce incorrect CSS output.

Affected tests:
- `math-parens`: media-math ✅, parens ⚠️, css ⚠️, mixins-args ⚠️
- `math-parens-division`: media-math ✅, new-division ✅, parens ⚠️, mixins-args ⚠️
- `math-always`: mixins-guards ✅, no-sm-operations ✅

## What You Need to Do

1. **Setup**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-math-operations-SESSION_ID

   # Test current state
   pnpm -w test:go:filter -- "math-parens"
   pnpm -w test:go:filter -- "math-parens-division"
   ```

2. **Understand Math Modes**:
   - Less.js has 4 math modes: `strict`, `parens`, `parens-division`, `always`
   - Each has different rules for when operations are performed
   - Read `.claude/tasks/archived/math-operations.md` for details

3. **Debug One Suite at a Time**:
   ```bash
   # Start with math-parens suite (simplest)
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "parens"
   ```

   This shows expected vs actual CSS diffs.

4. **Compare with JavaScript**:
   - For each failing test, run it through less.js
   - Understand what the correct output should be
   - Check how math modes affect operation evaluation

5. **Find the Root Cause**:
   - Is the issue in `operation.go`?
   - Is it in how math mode is set?
   - Is it in `paren.go` or expression evaluation?
   - Check: `packages/less/src/less/less_go/operation.go`
   - Check: `packages/less/src/less/less_go/contexts.go` (mathOn setting)

6. **Fix the Code**:
   - Identify which math mode is being applied incorrectly
   - Fix the logic in operation evaluation
   - Test each suite: `pnpm -w test:go:filter -- "parens"`, etc.

7. **Validate**:
   - Unit tests: `pnpm -w test:go:unit`
   - Full suite: `pnpm -w test:go`
   - Check for regressions

8. **Commit**:
   ```bash
   git add -A
   git commit -m "Fix math operations: Correct math mode handling for parens/division/always modes"
   git push -u origin claude/fix-math-operations-SESSION_ID
   ```

## Success Criteria

- ✅ All math-parens tests compile AND produce correct output
- ✅ All math-parens-division tests produce correct output
- ✅ All math-always tests still work
- ✅ No regressions (full test suite passes)
- ✅ Commit explains which math mode(s) were fixed

## Key Files

- Main test suites:
  - `packages/test-data/less/math-parens/`
  - `packages/test-data/less/math-parens-division/`
  - `packages/test-data/less/math-always/`

- Code to fix:
  - `packages/less/src/less/less_go/operation.go` (main logic)
  - `packages/less/src/less/less_go/paren.go` (parens handling)
  - `packages/less/src/less/less_go/contexts.go` (math mode config)
  - `packages/less/src/less/less_go/dimension.go` (if number ops broken)

## Important Notes

- Math operations are everywhere - test thoroughly!
- Some tests are already passing (media-math, new-division) - DON'T break them
- The failing tests show CSS diffs - use LESS_GO_DIFF=1 to see them
- Some operations should NOT be performed (only in certain modes)
- Division (/) handling is different in different modes

## Reference Task File

See `.claude/tasks/archived/math-operations.md` for detailed analysis of:
- Each math mode and its rules
- Which operations should/shouldn't be performed
- Specific test failures and causes
