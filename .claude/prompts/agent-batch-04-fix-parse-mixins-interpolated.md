# Agent Prompt 4: Fix Parser Failures in mixins-interpolated Test

**Priority**: HIGH
**Impact**: 1 test (mixins are critical feature)
**Time Estimate**: 2-3 hours
**Difficulty**: Medium-High

## Task

Fix the parse failure in `mixins-interpolated` test which currently fails with:
```
Parse: Unrecognised input in ../../../../test-data/less/_main/mixins-interpolated.less
```

Mixin interpolation is an advanced feature. The parser needs to handle interpolated mixin names.

## What You Need to Do

1. **Prepare**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-parse-mixins-interpolated-SESSION_ID
   pnpm -w test:go:filter -- "mixins-interpolated"
   ```

2. **Investigate the syntax**:
   - Read: `packages/test-data/less/_main/mixins-interpolated.less`
   - Compare with: `packages/less/src/less/less/mixins-interpolated.less`
   - What interpolated mixin syntax is not being parsed?
   - Examples: `.@{name}()`, `@{mixin-call}`, etc.

3. **Debug**:
   - Use traces: `LESS_GO_TRACE=1 pnpm -w test:go:filter -- "mixins-interpolated"`
   - Find exact location where parser fails
   - Check which mixin pattern causes the issue

4. **Fix Parser**:
   - Edit `packages/less/src/less/less_go/parser.go` or mixin-related parsers
   - Add support for interpolated mixin syntax
   - Test incrementally

5. **Test for Regressions**:
   - Unit tests: `pnpm -w test:go:unit`
   - Integration: `pnpm -w test:go`
   - Mixins are used everywhere - be very careful!

6. **Commit**:
   ```bash
   git add -A
   git commit -m "Fix parser: Add support for interpolated mixin syntax"
   git push -u origin claude/fix-parse-mixins-interpolated-SESSION_ID
   ```

## Success Criteria

- ✅ `mixins-interpolated` compiles
- ✅ No regressions in other tests
- ✅ All unit tests pass
- ✅ Clear commit message

## Files

- Test: `packages/test-data/less/_main/mixins-interpolated.less`
- Reference: `packages/less/src/less/less/mixins-interpolated.less`
- Expected: `packages/test-data/less/_main/mixins-interpolated.css`
- Parser: `packages/less/src/less/less_go/parser.go`

## Key Points

- Mixin interpolation allows dynamic mixin names
- This is a parser feature, not runtime
- Mixins are used by many tests - test thoroughly
- Look for patterns like: `.@{var}()`, `@mixin-@{suffix}()`
