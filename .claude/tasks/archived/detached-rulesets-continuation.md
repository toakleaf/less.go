# Continue: Fix Detached Rulesets Media Query Output

## Context
You're continuing work on fixing the `detached-rulesets` integration test in the less.go port. The test compiles successfully but CSS output differs because media queries nested within detached rulesets are not appearing in the final output.

## Current Branch
`claude/fix-detached-rulesets-formatting-011CUzCXfrsrkhzFVV2uNgs7`

## The Problem

When a detached ruleset containing `@media` queries is called from within a parent `@media` block, the nested media queries should merge with the parent context and output combined queries:

```less
@my-ruleset: {
  .my-selector {
    @media (tv) {
      background-color: black;
    }
  }
};

@media (orientation:portrait) {
  @my-ruleset();
}
```

**Expected output:**
```css
@media (orientation: portrait) and (tv) {
  .my-selector {
    background-color: black;
  }
}
```

**Actual output:** The merged media queries are missing entirely (lines 49-68 in expected output).

## Root Cause Identified

Through extensive debugging (see commit `bca5ea4`), the issue is clearly identified:

1. ✅ Media blocks ARE being created during detached ruleset evaluation (confirmed: 3 blocks)
2. ✅ The blocks ARE being propagated back to parent context via `DetachedRuleset.CallEval()`
3. ✅ The blocks ARE being included in rules during VariableCall processing (lines 684-709 in ruleset.go)
4. ✅ `Media.EvalTop()` IS creating a MultiMedia Ruleset with the blocks
5. ❌ **BUT** all parent `@media` nodes are calling `evalNested()` instead of `evalTop()`
6. ❌ This causes them to return empty placeholder Rulesets instead of the actual merged media queries

**Debug Evidence:**
```
[MEDIA.EvalTop] Creating MultiMedia Ruleset with 3 media blocks
[MEDIA.EvalTop] MultiMedia Ruleset created, returning it
[RULESET.Eval] Media node evaluated to type=*less_go.Ruleset
[RULESET.Eval]   Ruleset MultiMedia=false, Rules=0  ← Problem here!
```

The MultiMedia Ruleset with `Rules=3` is created but gets replaced by empty Rulesets.

## What's Been Done

**Files modified:**

1. **`ruleset.go` (lines 684-709)**: Added logic to include mediaBlocks created during VariableCall evaluation
2. **`media.go` & `ruleset.go`**: Added comprehensive debug logging with `LESS_GO_DEBUG=1`

**How to test:**
```bash
cd packages/less/src/less/less_go
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets" 2>&1 | grep -E "MEDIA|VariableCall"
```

## Your Task

**Fix the issue where parent @media nodes incorrectly call evalNested() instead of evalTop().**

The problem is in `media.go` around line 584:
```go
if len(evalCtx.MediaPath) == 0 {
    return media.EvalTop(evalCtx), nil
} else {
    return media.EvalNested(evalCtx), nil
}
```

When a detached ruleset containing media queries is evaluated, something is causing `mediaPath` to not be empty when the parent @media finishes evaluation, forcing it to call `evalNested()` which returns an empty placeholder Ruleset.

**Investigation steps:**

1. **Trace mediaPath manipulation**: Add debug logging to see when items are added/removed from `mediaPath` during detached ruleset evaluation
2. **Check Media.Eval() lines 544-582**: The push/pop logic for `mediaPath` may not be handling detached ruleset evaluation correctly
3. **Compare with JavaScript**: Look at `packages/less/src/less/tree/media.js` to see how JavaScript handles this
4. **Key insight**: When `DetachedRuleset.CallEval()` evaluates its inner ruleset, any nested Media nodes will push themselves onto `mediaPath`. They need to pop themselves off before the parent Media checks the path length.

## Success Criteria

```bash
# Test passes with perfect match
pnpm -w test:go:filter -- "detached-rulesets"
# Expected: ✅ detached-rulesets: Perfect match!

# No regressions (should still be 78 perfect matches)
pnpm -w test:go 2>&1 | grep "✅.*Perfect match" | wc -l
# Expected: 79 (one more than current 78)

# All unit tests pass
pnpm -w test:go:unit
```

## Files to Focus On

- `packages/less/src/less/less_go/media.go` - Media.Eval(), evalTop(), evalNested()
- `packages/less/src/less/less_go/detached_ruleset.go` - CallEval()
- `packages/less/src/less/less_go/ruleset.go` - VariableCall handling (lines 657-710)
- `packages/less/src/less/tree/media.js` - JavaScript reference implementation

## Debug Commands

```bash
# See what's happening with media paths
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets" 2>&1 | grep -E "mediaPath|mediaBlock"

# See the actual diff
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets" 2>&1 | less

# Minimal reproduction
cat > /tmp/test.less << 'TESTEOF'
@my-ruleset: {
  .my-selector {
    @media (tv) {
      background-color: black;
    }
  }
};
@media (orientation:portrait) {
  @my-ruleset();
}
TESTEOF

# Then test manually
cd packages/less/src/less/less_go
go test -v ./... -run "TestIntegrationSuite"
```

## Key JavaScript Behavior to Match

In JavaScript (`tree/media.js`):
```javascript
context.mediaPath.push(media);
context.mediaBlocks.push(media);
// ... evaluate rules ...
context.mediaPath.pop();
return context.mediaPath.length === 0 ? media.evalTop(context) : media.evalNested(context);
```

The Go implementation should match this exactly. The issue is likely that when detached rulesets containing media queries are evaluated, the mediaPath push/pop is getting out of sync.

## Commit & Push When Done

```bash
git add -A
git commit -m "Fix detached-rulesets media query output

When detached rulesets containing @media queries are called from
within parent @media blocks, the nested queries should merge and
output combined media queries.

The issue was [describe your fix here].

Fixes media query bubbling in detached ruleset contexts.
Test detached-rulesets now passes with perfect match."

git push -u origin claude/fix-detached-rulesets-formatting-011CUzCXfrsrkhzFVV2uNgs7
```

## Current Status
- **Perfect matches:** 78/184 tests (42.4%)
- **This fix will bring it to:** 79/184 tests (42.9%)
- **Regressions:** 0 (verified)

Good luck! The root cause is clearly identified - you just need to fix the mediaPath management so parent @media nodes correctly call evalTop().
