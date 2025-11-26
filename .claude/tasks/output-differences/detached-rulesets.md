# Task: Fix Detached Rulesets Media Query Output

**Status**: Available
**Priority**: High
**Tests**: 1 (detached-rulesets)
**Estimated Time**: 2-3 hours

## Problem

When detached rulesets containing `@media` queries are called from within a parent `@media` block, the nested media queries should merge with the parent context. Currently, merged media queries are missing from output.

## Example

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

**Actual output:** The merged media queries are missing entirely.

## Root Cause

The issue is in media path management during detached ruleset evaluation:
1. Media blocks ARE being created during detached ruleset evaluation
2. Media blocks ARE being propagated back to parent context
3. BUT parent `@media` nodes are calling `evalNested()` instead of `evalTop()`
4. This causes them to return empty placeholder Rulesets

The problem is in `media.go` around line 584 - when `mediaPath` is checked, it's not empty when it should be.

## Test Commands

```bash
# Run specific test
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"

# Debug mode
LESS_GO_DEBUG=1 pnpm -w test:go:filter -- "detached-rulesets"
```

## Key Files

**Go (to fix)**:
- `packages/less/src/less/less_go/media.go` - Media.Eval(), evalTop(), evalNested()
- `packages/less/src/less/less_go/detached_ruleset.go` - CallEval()
- `packages/less/src/less/less_go/ruleset.go` - VariableCall handling

**JavaScript (reference)**:
- `packages/less/src/less/tree/media.js`

## Key JavaScript Behavior

```javascript
context.mediaPath.push(media);
context.mediaBlocks.push(media);
// ... evaluate rules ...
context.mediaPath.pop();
return context.mediaPath.length === 0 ? media.evalTop(context) : media.evalNested(context);
```

The Go implementation should match this exactly.

## Success Criteria

- `detached-rulesets` test shows "Perfect match!"
- All unit tests pass (`pnpm -w test:go:unit`)
- No regressions in other media or detached ruleset tests
