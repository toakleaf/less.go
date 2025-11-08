# Agent Prompt 1: Fix Parser Failures in functions-each Test

**Priority**: CRITICAL
**Impact**: 1 test + potentially others with same parse pattern
**Time Estimate**: 2-3 hours
**Difficulty**: Medium

## Task

Fix the parse failure in `functions-each` test which currently fails with:
```
Parse: Unrecognised input in ../../../../test-data/less/_main/functions-each.less
```

## What You Need to Do

1. **Investigate the Parse Error**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-parse-functions-each-SESSION_ID
   pnpm -w test:go:filter -- "functions-each"
   ```

2. **Find the problematic LESS syntax**:
   - Read the test file: `packages/test-data/less/_main/functions-each.less`
   - Identify what syntax the parser doesn't recognize
   - Compare with: `packages/less/src/less/less/functions-each.less` (JS reference)

3. **Debug the Parser**:
   - Use `LESS_GO_TRACE=1 pnpm -w test:go:filter -- "functions-each"` to trace parsing
   - Find where the parser stops recognizing input
   - Compare parser logic with less.js parser

4. **Fix the Parser**:
   - Modify `packages/less/src/less/less_go/parser.go` or related files
   - Add support for the missing syntax pattern
   - Test incrementally: `pnpm -w test:go:filter -- "functions-each"`

5. **Validate**:
   - Ensure test compiles: `pnpm -w test:go:filter -- "functions-each"`
   - Run all unit tests: `pnpm -w test:go:unit`
   - Run all integration tests: `pnpm -w test:go` - CHECK FOR REGRESSIONS
   - Document what syntax was missing in commit message

6. **Commit & Push**:
   ```bash
   git add -A
   git commit -m "Fix parser: Add support for [syntax feature] in functions-each test"
   git push -u origin claude/fix-parse-functions-each-SESSION_ID
   ```

## Success Criteria

- ✅ Test `functions-each` compiles successfully
- ✅ No new regressions in other tests
- ✅ No unit test failures
- ✅ Commit message explains what parser pattern was fixed

## Context Files

- Test input: `packages/test-data/less/_main/functions-each.less`
- JS reference: `packages/less/src/less/less/functions-each.less`
- Expected output: `packages/test-data/less/_main/functions-each.css`
- Parser: `packages/less/src/less/less_go/parser.go`

## Important Notes

- Do NOT modify test data files or JS reference files
- Focus only on parser fixes in Go code
- This is a parser issue, not runtime - the fix is in parsing, not evaluation
- Check git log for recent parser changes that might have broken this
