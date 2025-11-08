# Agent Prompt 7: Complete Mixin Issues (2 Remaining Tests)

**Priority**: MEDIUM
**Impact**: 2 tests
**Time Estimate**: 1-2 hours
**Difficulty**: MEDIUM

## Task

Complete the partially-done mixin issues. Two tests remain with output differences:
- `mixins-nested` - Nested mixin calls not working correctly
- `mixins-important` - !important flag not propagating through mixins

Note: `mixins-named-args` is already fixed ✅

## What You Need to Do

1. **Setup**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/complete-mixin-issues-SESSION_ID
   ```

2. **Debug mixins-nested**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "mixins-nested"
   ```
   - What's missing from output?
   - Are nested mixin calls being made?
   - Are the correct properties appearing?

   Test file: `packages/test-data/less/_main/mixins-nested.less`
   Reference: `packages/less/src/less/less/mixins-nested.less`

3. **Debug mixins-important**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "mixins-important"
   ```
   - Is `!important` being preserved through mixin calls?
   - Are properties marked as important?
   - Check declaration evaluation

   Test file: `packages/test-data/less/_main/mixins-important.less`
   Reference: `packages/less/src/less/less/mixins-important.less`

4. **Identify Root Causes**:

   **For mixins-nested**:
   - Check: How are nested mixins handled?
   - Check: `mixin_call.go` - does it support nested calls?
   - Check: Frame scoping - are frames being maintained?
   - Look for recursive mixin handling

   **For mixins-important**:
   - Check: How is the `!important` flag stored?
   - Check: `declaration.go` - does it preserve importance?
   - Check: `mixin_call.go` - when calling mixins, is importance passed?
   - Look for: `Important` field in Declaration

5. **Fix the Issues**:
   - For nested: Ensure mixin calls within mixins are evaluated correctly
   - For important: Ensure `!important` flag is propagated through mixin evaluation

6. **Test Each Fix**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "mixins-nested"
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "mixins-important"
   ```

7. **Validate**:
   - Unit tests: `pnpm -w test:go:unit`
   - Full suite: `pnpm -w test:go` (check for regressions)

8. **Commit**:
   ```bash
   git add -A
   git commit -m "Fix mixin issues: Support nested mixins and !important propagation

   - Fixed nested mixin evaluation in mixins-nested test
   - Fixed !important flag preservation in mixins-important test"
   git push -u origin claude/complete-mixin-issues-SESSION_ID
   ```

## Success Criteria

- ✅ Both `mixins-nested` and `mixins-important` produce correct output
- ✅ No regressions in other mixin tests
- ✅ All unit tests pass
- ✅ Clear commit message

## Key Files

- Tests:
  - `packages/test-data/less/_main/mixins-nested.less` / `.css`
  - `packages/test-data/less/_main/mixins-important.less` / `.css`

- Code:
  - `packages/less/src/less/less_go/mixin_call.go`
  - `packages/less/src/less/less_go/mixin_definition.go`
  - `packages/less/src/less/less_go/declaration.go`
  - `packages/less/src/less/less_go/ruleset.go`

## Context

- `mixins-named-args` was already fixed in previous work
- These are the last 2 mixin output difference tests
- After these, mixin functionality will be 99% complete

## Important Notes

- Mixin tests can be complex due to scoping/framing
- Check the task file: `.claude/tasks/output-differences/mixin-issues.md` (if it exists)
- Use debug output to trace mixin evaluation
- Look for similar patterns in working mixin tests (mixins, mixins-closure, etc.)
