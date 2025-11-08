# Agent Prompt 2: Fix Parser Failures in selectors Test

**Priority**: CRITICAL
**Impact**: 1 test
**Time Estimate**: 2-3 hours
**Difficulty**: Medium

## Task

Fix the parse failure in `selectors` test which currently fails with:
```
Parse: Unrecognised input in ../../../../test-data/less/_main/selectors.less
```

## What You Need to Do

1. **Investigate the Parse Error**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-parse-selectors-SESSION_ID
   pnpm -w test:go:filter -- "selectors"
   ```

2. **Locate the problematic syntax**:
   - Read: `packages/test-data/less/_main/selectors.less`
   - Compare with: `packages/less/src/less/less/selectors.less` (JS reference)
   - Find where parser stops (use LESS_GO_TRACE=1 if needed)

3. **Determine the missing parser feature**:
   - What CSS/LESS selector syntax is not being parsed?
   - Does less.js support it? Check the JS parser
   - Is it a valid LESS feature?

4. **Fix the Parser**:
   - Edit `packages/less/src/less/less_go/parser.go` or selector-related files
   - Add missing parser rules for the syntax
   - Incrementally test: `pnpm -w test:go:filter -- "selectors"`

5. **Test for Regressions**:
   - Run all unit tests: `pnpm -w test:go:unit`
   - Run full integration suite: `pnpm -w test:go`
   - Verify no tests broke that were passing before

6. **Commit**:
   ```bash
   git add -A
   git commit -m "Fix parser: Add support for [selector feature] in selectors test"
   git push -u origin claude/fix-parse-selectors-SESSION_ID
   ```

## Success Criteria

- ✅ `selectors` test compiles
- ✅ No regressions
- ✅ All unit tests pass
- ✅ Clear commit message

## Files

- Input: `packages/test-data/less/_main/selectors.less`
- Reference: `packages/less/src/less/less/selectors.less`
- Expected: `packages/test-data/less/_main/selectors.css`
- Code: `packages/less/src/less/less_go/parser.go`

## Notes

- This is about parsing CSS/LESS selectors
- The issue is in parser, not evaluation
- Be thorough with regression testing - parser changes affect everything
