# Agent Prompt 6: Fix Formatting & Whitespace Output Issues

**Priority**: MEDIUM
**Impact**: 6+ tests
**Time Estimate**: 2-3 hours
**Difficulty**: LOW

## Task

Fix formatting and whitespace issues in CSS output. These tests compile correctly but have minor formatting differences (extra/missing newlines, spacing, etc.).

Affected tests:
- `comments` - Comment formatting
- `comments2` - ✅ Already perfect!
- `whitespace` - Whitespace handling
- `variables-in-at-rules` - Newline issues in @rules
- `charsets` - Charset formatting
- `parse-interpolation` - Selector formatting

## What You Need to Do

1. **Setup**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-formatting-output-SESSION_ID
   ```

2. **Examine Each Test**:
   ```bash
   # See differences
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "comments"
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "whitespace"
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "variables-in-at-rules"
   ```

3. **Analyze the Differences**:
   - Are extra newlines being added?
   - Are newlines being removed when they should be there?
   - Is indentation correct?
   - Are comments being rendered?
   - Are selectors being formatted correctly?

4. **Find the Root Causes**:
   - `comments`: Check how comments are being output - `comment.go`
   - `whitespace`: Check CSS generation - `ruleset.go`, output methods
   - `variables-in-at-rules`: Check @rule formatting - `at_rule.go`
   - `charsets`: Check charset rule handling - `at_rule.go`
   - `parse-interpolation`: Check selector interpolation - `selector.go`

5. **Fix the Code**:
   - Common issues:
     - Missing `\n` after at-rules
     - Extra/missing spaces around selectors
     - Comment placement
     - Media query formatting
   - Look for: `output += "\n"`, `fmt.Sprintf`, selector joining

6. **Fix One Test at a Time**:
   ```bash
   # Test individual formatting issues
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "whitespace"
   # Make changes
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "whitespace"
   # Verify
   ```

7. **Validate**:
   - Unit tests: `pnpm -w test:go:unit`
   - Full suite: `pnpm -w test:go`
   - Check for regressions (formatting changes can be risky)

8. **Commit**:
   ```bash
   git add -A
   git commit -m "Fix formatting: Correct whitespace/newlines in CSS output

   - Fixed comment rendering in [test]
   - Fixed newline handling in @rules
   - Fixed selector spacing in interpolation"
   git push -u origin claude/fix-formatting-output-SESSION_ID
   ```

## Success Criteria

- ✅ At least 4/6 formatting tests now perfect matches
- ✅ No regressions
- ✅ All unit tests pass
- ✅ Clear commit message explaining what was fixed

## Files to Check

Test files:
- `packages/test-data/less/_main/comments.less` / `.css`
- `packages/test-data/less/_main/whitespace.less` / `.css`
- `packages/test-data/less/_main/variables-in-at-rules.less` / `.css`
- `packages/test-data/less/_main/charsets.less` / `.css`
- `packages/test-data/less/_main/parse-interpolation.less` / `.css`

Code files (likely):
- `packages/less/src/less/less_go/comment.go`
- `packages/less/src/less/less_go/ruleset.go` (output methods)
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/selector.go`
- `packages/less/src/less/less_go/anonymous.go`

## Important Notes

- These are mostly output formatting issues, not logic
- Easy to fix but need careful testing (whitespace matters!)
- Some are newline issues (extra/missing `\n`)
- Some are space issues (around selectors, properties)
- Comments2 is already working - check how it does it!
- Use LESS_GO_DIFF=1 to see exact differences

## Quick Wins

If you fix these 6 tests, we'll have:
- Current: 42 perfect matches
- New: 48 perfect matches (+6)
- Success rate: 67.6% 🎉

This is high-impact, low-complexity work!
