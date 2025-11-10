# Continue: Fix import-reference Tests

## Current Status

Branch: `claude/fix-import-reference-tests-011CUzCPSHawpUCUx5m92Ur9`

**Last Updated:** After merging origin/master (commit 0cab374 - "WIP: Partial fix for import reference flag visibility")

### ✅ What's Working
- **79 perfect matches** (gained 1 from master merge) ⬆️
- All extend tests passing perfectly (7/7)
- All unit tests passing (2,290+ tests)
- Partial fix from master: Some reference import selectors (`.z` from import/import-reference.less) ARE appearing

### ❌ What's Broken
- **import-reference** test (main suite) - Output differs
- **import-reference-issues** test (main suite) - Output differs
- `.test-rule-c` still missing from output (main issue)
- Some selectors from css-3.less and media.less not appearing

## The Problem

`.test-rule-c` is completely missing from the output.

**Expected behavior:**
```less
// main file
.test-rule-c {
  &:extend(.test-rule-b all);
}

// global-scope-import.less (imported with @import (reference))
.test-rule-b {
  background-color: green;
}
```

**Expected CSS output:**
```css
.test-rule-c {
  background-color: green;
}
```

**Actual output:** `.test-rule-c` is completely missing.

## Root Cause Identified

Through debugging, we found:
1. ✅ The extend IS finding `.test-rule-b` during the chaining phase
2. ✅ The selector `.test-rule-c` IS being added to `.test-rule-b`'s ruleset
3. ❌ The created selector has **`visibility=false`** when it should be `visibility=true`
4. ❌ This causes it to be filtered out by `compileRulesetPaths` in `to_css_visitor.go` (lines 776-801)

### Why is visibility=false?

The extend from `.test-rule-c` (which is in the main file, NOT in a reference import) has `visibility=false`. This is wrong - extends in the main file should have `visibility=true` after `SetTreeVisibilityVisitor(true)` runs.

## What Was Already Fixed (via master merge)

### 1. extend_visitor.go - Partial Fix from Master (commit 0cab374)

**Problem:** Selectors from reference imports that are extended by visible selectors weren't appearing in output.

**Fix Applied:** The master branch has a partial fix that:
1. Keeps the conditional check for visibility blocks (different from our initial approach)
2. When a visible extend matches selectors from reference imports, it:
   - Calls `EnsureVisibility()` on matched selectors
   - Sets `EvaldCondition = true` on matched selectors (for isOutput check)
   - Calls `EnsureVisibility()` and `RemoveVisibilityBlock()` on the matched ruleset

**Result:** Works for SOME cases (`.z` selectors appear) but NOT all (`.test-rule-c` still missing).

### 2. Other improvements from master
- Fixed detached-rulesets formatting
- Fixed number formatting (no scientific notation)
- Added task documentation

## What Still Needs Investigation

### Primary Issue: Visibility Not Set Correctly

**Question:** Why does `.test-rule-c`'s extend have `visibility=false`?

**Areas to investigate:**

1. **SetTreeVisibilityVisitor** (`set_tree_visibility_visitor.go`):
   - Lines 56-59: Skips nodes where `BlocksVisibility()` returns true
   - Is `.test-rule-c` or its extend somehow getting visibility blocks?
   - Should extends inside rulesets be visited separately?

2. **Extend.Eval** (`extend.go`):
   - Does evaluation of extends preserve or override visibility?
   - Check if there's any cloning that doesn't preserve visibility

3. **Ruleset parsing/evaluation**:
   - When `.test-rule-c { &:extend(...); }` is parsed, does the ruleset or extend get visibility blocks?
   - Check if there's something marking main file content as having visibility blocks

4. **Visitor pipeline** (`transform_tree.go` lines 173-180):
   ```go
   NewJoinSelectorVisitor(),
   NewSetTreeVisibilityVisitor(true),  // Should set visibility=true
   NewExtendVisitor(),
   NewToCSSVisitor(...),
   ```
   - Is the visitor order correct?
   - Is SetTreeVisibilityVisitor actually being called on extends?

### Secondary Issue: Empty Rulesets

The actual output also shows:
```css
#do-not-show-import {

}
```

This empty ruleset should not appear. This might be a separate issue or related to the visibility filtering.

## Debugging Tools Available

### 1. Enable Debug Output
```bash
LESS_GO_DEBUG_EXTEND=1 go test -v -run "TestIntegrationSuite/main/import-reference-issues"
```

### 2. Check Visibility Flow
Add debug output in:
- `set_tree_visibility_visitor.go` line 63 (when calling EnsureVisibility)
- `extend.go` Eval method (check if visibility is preserved)
- `extend_visitor.go` line 472 (check isVisible value)

### 3. Compare with JavaScript
The JavaScript visitor pipeline in `render.js` or equivalent shows the exact order and how visibility is set.

## Files to Focus On

**Primary:**
1. `/home/user/less.go/packages/less/src/less/less_go/set_tree_visibility_visitor.go`
   - Why isn't it setting visibility=true on extends in main file?

2. `/home/user/less.go/packages/less/src/less/less_go/extend.go`
   - Does Eval preserve visibility?

3. `/home/user/less.go/packages/less/src/less/less_go/to_css_visitor.go`
   - Lines 745-824: compileRulesetPaths filtering logic

**Secondary:**
4. `/home/user/less.go/packages/less/src/less/less_go/extend_visitor.go`
   - Already partially fixed, but verify isVisible is correct

5. `/home/user/less.go/packages/less/src/less/less_go/transform_tree.go`
   - Verify visitor pipeline order

## Test Commands

### Run the failing tests
```bash
pnpm -w test:go:filter -- "import-reference"
```

### Run with diff output
```bash
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference-issues"
```

### Check for regressions
```bash
# Unit tests
pnpm -w test:go:unit

# Integration tests - should still have 78 perfect matches
go test -v -run "TestIntegrationSuite" 2>&1 | grep "Perfect match" | wc -l
```

## Success Criteria

1. ✅ `.test-rule-c { background-color: green; }` appears in output
2. ✅ `#do-not-show-import` empty ruleset does NOT appear
3. ✅ All 78+ perfect matches still passing (no regressions)
4. ✅ All unit tests passing

## Key JavaScript References

### Visibility System
- `/home/user/less.go/packages/less/src/less/tree/node.js` lines 162-168
  - `isVisible()` returns `this.nodeVisible`
  - `undefined` = no explicit visibility (inherits from context)
  - `true` = explicitly visible
  - `false` = explicitly invisible

### Import Evaluation
- `/home/user/less.go/packages/less/src/less/tree/import.js` lines 114-127
  - Reference imports call `addVisibilityBlock()` on result nodes

### Extend Processing
- `/home/user/less.go/packages/less/src/less/visitors/extend-visitor.js` lines 465-500
  - No conditional checks - always processes matches

### Path Filtering
- `/home/user/less.go/packages/less/src/less/visitors/to-css-visitor.js` lines 282-296
  - `_compileRulesetPaths` filters paths
  - Keeps paths where ANY selector is visible (not ALL)

## Suggested Approach

1. **Add debug logging** to trace visibility values:
   ```go
   fmt.Printf("[DEBUG] Extend visibility: %v\n", extend.Node.IsVisible())
   ```

2. **Check if BlocksVisibility is incorrectly true** for main file extends:
   ```go
   fmt.Printf("[DEBUG] Extend blocks visibility: %v\n", extend.Node.BlocksVisibility())
   ```

3. **Verify SetTreeVisibilityVisitor runs** on extends:
   - Add logging in `callEnsureVisibility` method
   - Check if extends are being visited at all

4. **Compare with working test**: Look at a similar extend test that works (e.g., `extend-chaining`) and trace how visibility is set differently.

5. **Minimal reproduction**: Create a simple test case:
   ```less
   // test.less
   @import (reference) "imported.less";
   .visible { &:extend(.from-ref all); }

   // imported.less
   .from-ref { color: red; }
   ```
   Expected: `.visible { color: red; }`

## Additional Notes

- The issue is NOT with the extend matching logic (that works)
- The issue is NOT with the path creation (that works)
- The issue IS with visibility propagation causing paths to be filtered out
- This is a "last mile" problem - everything works except the visibility flag

## Related Files in Previous Commit

The commit `b27240f` has detailed investigation notes in the commit message. Review it for additional context.

## Questions to Answer

1. Why does `extend.Node.IsVisible()` return `false` for extends in the main file?
2. Is `SetTreeVisibilityVisitor` actually visiting extend nodes?
3. If it is visiting them, why isn't it setting `nodeVisible = true`?
4. Is there something calling `AddVisibilityBlock()` on main file content that shouldn't be?

Good luck! The fix is close - it's just a matter of finding where visibility is being incorrectly set or not set at all.
