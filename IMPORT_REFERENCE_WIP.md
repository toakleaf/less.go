# Import Reference Investigation - Work in Progress

## Problem Statement
The `import-reference` and `import-reference-issues` tests are failing because files imported with `@import (reference)` are outputting their CSS content when they shouldn't. Only explicitly used parts (via mixin calls or extends) should appear in the output.

## Current Status
**Tests Status**: Still failing
- `import-reference`: Outputs all content from referenced imports instead of just used parts
- `import-reference-issues`: Same issue
- `import-once`: Also affected

## Changes Made

### 1. Fixed Path Filtering Visibility Logic (`to_css_visitor.go:720-748`)
**Problem**: When filtering selector paths, `nil` visibility was treated as `false` (hidden), but it should mean "inherit from parent".

**Fix**: Changed logic to:
- `nil` visibility → Check `BlocksVisibility()` to determine if hidden
- `true` → Explicitly visible
- `false` → Explicitly hidden

**Code**:
```go
isVisible := true // default to visible if not explicitly set
if vis := sel.IsVisible(); vis != nil {
    isVisible = *vis
} else {
    // nil visibility means inherit - check if blocked
    if blocker, ok := selector.(interface{ BlocksVisibility() bool }); ok {
        if blocker.BlocksVisibility() {
            isVisible = false
        }
    }
}
```

### 2. Maintained BlocksVisibility Check in IsVisibleRuleset (`to_css_visitor.go:176-186`)
Rulesets from reference imports that haven't been explicitly made visible should be filtered out.

## Root Cause Analysis

### How Reference Imports Should Work

1. **Import Evaluation** (`import.go:331-343`):
   - When `@import (reference)` is evaluated, `AddVisibilityBlock()` is called on all imported nodes
   - These nodes are inserted into the parent ruleset

2. **Mixin Calls** (`mixin_definition.go:539-647`):
   - When a mixin from a reference import is called, it evaluates and returns a ruleset
   - The returned ruleset's RULES inherit visibility blocks from the mixin definition's rules
   - But the wrapper ruleset itself has no visibility blocks

3. **CSS Generation** (`to_css_visitor.go`):
   - ToCSSVisitor filters nodes based on visibility
   - Rulesets with `BlocksVisibility() == true` and `nodeVisible != true` are filtered out
   - Individual rules (declarations, etc.) with `BlocksVisibility() == true` are also filtered out

### The Missing Piece

**JavaScript Behavior** (`to-css-visitor.js:271-274`):
```javascript
if (this.utils.isVisibleRuleset(rulesetNode)) {
    rulesetNode.ensureVisibility();
    rulesets.splice(0, 0, rulesetNode);
}
```

When a ruleset IS visible (has content and selectors), JavaScript calls `ensureVisibility()` on it, which sets `nodeVisible = true`. This overrides visibility blocks.

**Current Issue**:
- Rulesets from reference imports have `blocksVisibility() = true` and `nodeVisible = nil`
- They're correctly filtered out by `IsVisibleRuleset()`
- BUT this also filters out rulesets returned from mixin calls that SHOULD be visible

## Theories on What's Missing

### Theory 1: Mixin Call Should Remove Visibility Blocks
When a mixin is called, perhaps the returned rules should NOT have visibility blocks, OR should have `nodeVisible = true` set.

**Evidence Against**: JavaScript doesn't appear to do this either.

### Theory 2: JavaScript `isVisible()` Behaves Differently with undefined
Maybe JavaScript's truthiness evaluation of `undefined` in boolean contexts is different from what we expect.

**Need to verify**: How does `if (selector.isVisible() && selector.getIsOutput())` evaluate when `isVisible()` returns `undefined`?

### Theory 3: Selector/Path Creation Timing
Perhaps in JavaScript, selectors get their visibility set at a different point in the eval/visit cycle.

## Next Steps

1. **Debug JavaScript Execution**:
   - Run import-reference test in JavaScript with added logging
   - Check when/where `nodeVisible` gets set on selectors and rulesets
   - Verify how `undefined` is handled in path filtering

2. **Compare Node States**:
   - Add debug logging to Go implementation to print:
     - `visibilityBlocks` value for rulesets from reference imports
     - `nodeVisible` value before/after mixin calls
     - Path filtering decisions

3. **Potential Solutions**:
   - Option A: When mixin returns rules, call `EnsureVisibility()` on the wrapper ruleset
   - Option B: Don't add visibility blocks to mixin definition rules when they're copied
   - Option C: Change how `isVisibleRuleset()` checks visibility for mixin-returned rulesets

## Files Modified
- `packages/less/src/less/less_go/to_css_visitor.go`
  - Lines 720-748: Path filtering visibility logic
  - Lines 176-186: IsVisibleRuleset visibility check

## Related Code Locations

### Go Implementation
- `import.go:331-343` - Adds visibility blocks to reference imports
- `mixin_definition.go:539-647` - Mixin evaluation and rule copying
- `to_css_visitor.go:160-195` - Ruleset visibility checking
- `to_css_visitor.go:705-770` - Path filtering

### JavaScript Implementation
- `tree/import.js:116-121` - Reference import visibility blocks
- `tree/mixin-definition.js:157-175` - Mixin eval call
- `tree/mixin-call.js:161,185-193` - Mixin call and visibility propagation
- `visitors/to-css-visitor.js:224-298` - Ruleset visiting and path compilation

## Test Cases
- `packages/test-data/less/_main/import-reference.less` - Main test file
- `packages/test-data/less/_main/import/import-reference.less` - Referenced file with mixins
- Expected output: `packages/test-data/css/_main/import-reference.css`

Run with:
```bash
pnpm -w test:go:filter -- "import-reference"
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/import-reference$" ./packages/less/src/less/less_go
```
