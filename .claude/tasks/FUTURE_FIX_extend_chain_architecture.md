# Deep Fix for Import-Reference Extend Chain Architecture

## Context

The current implementation has a **pragmatic workaround** for import-reference extend chains that works but doesn't match JavaScript's architecture. This prompt describes how to implement a proper fix.

## Current Workaround (Commit c7b2ac7)

Two special cases were added to `extend_visitor.go`:

### Fix #1: Allow reference import rulesets with ExtendOnEveryPath to be matched
```go
// Line ~454
rulesetHasVisibilityBlocks := ruleset.Node != nil && ruleset.Node.BlocksVisibility()
if ruleset.ExtendOnEveryPath && !rulesetHasVisibilityBlocks {
    continue  // Only skip non-reference rulesets
}
```

### Fix #2: Skip deeply chained extends to reference rulesets
```go
// Line ~516
numParents := len(allExtends[extendIndex].ParentIds)
isDeeplyChainedExtend := numParents > 1
shouldSkipDeeplyChainedRefExtend := isDeeplyChainedExtend && rulesetHasVisibilityBlocks && isVisible
```

**Why this is a workaround:** JavaScript unconditionally skips ALL `extendOnEveryPath` rulesets and ALL selectors with extends (lines 279-281 in extend-visitor.js). It doesn't need special cases for reference imports.

## The Problem Being Solved

Given this structure:
```less
// global-scope-nested.less (reference import - doubly nested)
.test-rule-a { color: red; }

// global-scope-import.less (reference import)
@import (reference) "global-scope-nested.less";
.test-rule-b {
  background-color: green;
  &:extend(.test-rule-a all);
}

// import-reference-issues.less (main file)
@import (reference) "global-scope-import.less";
.test-rule-c { &:extend(.test-rule-b all); }
```

**Expected:** `.test-rule-c { background-color: green; }` (from .test-rule-b)
**Without fixes:** `.test-rule-c { color: red; }` (from .test-rule-a - wrong!)

## Root Cause Hypothesis

The issue is that `doExtendChaining` creates chained extends correctly, but **the visibility propagation through extend chains is not working properly for reference imports**.

Evidence:
1. Commit #204 comment: *"The extended rulesets being marked visible are not the same instances being checked by KeepOnlyVisibleChilds"*
2. The chained extend `.test-rule-c extends .test-rule-a` is created (good!)
3. But it's matching against `.test-rule-a`'s ruleset and creating output (bad!)
4. The intermediate `.test-rule-b` ruleset should be the one that becomes visible

## How JavaScript Handles This

JavaScript's architecture:
1. **ExtendFinderVisitor** finds all extends
2. **doExtendChaining** creates chained extends (e.g., `.c extends .b`, `.b extends .a` → creates `.c extends .a`)
3. **visitRuleset** skips ALL selectors with extends (line 281: `if (extendList && extendList.length) { continue; }`)
4. **ToCSSVisitor** filters output using `isVisible() && getIsOutput()` (to-css-visitor.js line 291)

The key: JavaScript's `doExtendChaining` attaches the NEW chained extend to the intermediate ruleset's paths, and visibility propagates through the ruleset reference.

## Investigation Plan

### Step 1: Understand doExtendChaining's Ruleset Attachment

**File:** `packages/less/src/less/less_go/extend_visitor.go` line 343

```go
newExtend.Ruleset = targetExtend.Ruleset
```

**Question:** When `.test-rule-c extends .test-rule-b` creates a chained extend `.test-rule-c extends .test-rule-a`, which ruleset should `newExtend.Ruleset` point to?

**JavaScript (extend-visitor.js line 208):**
```javascript
newExtend.ruleset = targetExtend.ruleset;
```

This means: The chained extend `.test-rule-c → .test-rule-a` should point to `.test-rule-b`'s ruleset, NOT `.test-rule-a`'s ruleset!

**To investigate:**
1. Add logging in `doExtendChaining` to see which rulesets are being assigned
2. Check if the ruleset pointer is correct for reference imports
3. Verify that when `.test-rule-c` matches, it's using the correct ruleset's properties

### Step 2: Understand Visibility Propagation

**The visibility flow should be:**
1. `.test-rule-c` (visible, from main file) extends `.test-rule-b`
2. This should mark `.test-rule-b`'s **ruleset** as visible (not the selector)
3. The chained extend `.test-rule-c extends .test-rule-a` should NOT create output because `.test-rule-a`'s ruleset stays invisible

**Files to check:**
- `extend_visitor.go` lines 514-524 (makeParentNodesVisible)
- `extend.go` lines 164-176 (IsVisible with SelfSelectors check)
- `to_css_visitor.go` lines 765-799 (path filtering)

**Key question:** When a visible extend matches a reference import ruleset, which Node is being marked visible?

### Step 3: Test with Detailed Logging

Add this logging to understand the flow:

```go
// In doExtendChaining, line ~343
if showTrace {
    extendStr := "?"
    if len(extend.SelfSelectors) > 0 {
        if sel, ok := extend.SelfSelectors[0].(*Selector); ok {
            extendStr = sel.ToCSS(nil)
        }
    }
    targetStr := "?"
    if targetExtend.Selector != nil {
        if sel, ok := targetExtend.Selector.(*Selector); ok {
            targetStr = sel.ToCSS(nil)
        }
    }
    rulesetFirstSel := "?"
    if targetExtend.Ruleset != nil && len(targetExtend.Ruleset.Paths) > 0 {
        if len(targetExtend.Ruleset.Paths[0]) > 0 {
            if sel, ok := targetExtend.Ruleset.Paths[0][0].(*Selector); ok {
                rulesetFirstSel = sel.ToCSS(nil)
            }
        }
    }
    fmt.Printf("[CHAIN] %s extends %s → new chained extend will point to ruleset with selector %s\n",
        extendStr, targetStr, rulesetFirstSel)
}
```

### Step 4: Compare Instances

The comment from #204 says: *"Extended rulesets may be nested deeper (inside wrapper rulesets with AllowImports=true)"*

**Hypothesis:** The `.test-rule-b` ruleset that `newExtend.Ruleset` points to might be a DIFFERENT instance than the one being checked by `KeepOnlyVisibleChilds` in ToCSSVisitor.

**To verify:**
1. Add instance ID logging (using `%p` format for pointer addresses)
2. Track which ruleset instances are marked visible
3. Track which ruleset instances are checked in ToCSSVisitor
4. If they're different, find where the duplication happens

### Step 5: Proposed Proper Fix

Once you understand the root cause, the fix should:

1. **Remove the special cases** from my workaround:
   - Remove the `!rulesetHasVisibilityBlocks` exception in ExtendOnEveryPath check
   - Remove the `shouldSkipDeeplyChainedRefExtend` check

2. **Fix doExtendChaining** to ensure:
   - Chained extends point to the correct ruleset instance
   - Visibility propagates correctly through reference import chains
   - The intermediate ruleset (`.test-rule-b`) becomes visible, not the final target (`.test-rule-a`)

3. **Restore unconditional skipping** like JavaScript:
   ```go
   if ruleset.ExtendOnEveryPath {
       continue  // Always skip, no exceptions
   }
   ```

4. **Verify ToCSSVisitor filtering** works correctly:
   - Paths from `.test-rule-b` are visible and output
   - Paths from `.test-rule-a` remain invisible
   - Only properties from the intermediate ruleset appear in output

## Test Cases

### Primary Test Case
**File:** `packages/test-data/less/_main/import-reference-issues.less`

**Expected output:**
```css
.test-rule-c {
  background-color: green;  /* from .test-rule-b */
}
```

**Wrong output (before any fix):**
```css
.test-rule-c {
  color: red;  /* from .test-rule-a - incorrect! */
}
```

### Test Command
```bash
pnpm -w test:go:filter -- "import-reference"
```

### Regression Test
Make sure `extend-chaining` test doesn't regress further:
```bash
LESS_GO_DIFF=1 go test -v ./packages/less/src/less/less_go -run "TestIntegrationSuite/main/extend-chaining$"
```

## Success Criteria

1. ✅ `import-reference` and `import-reference-issues` tests produce correct output
2. ✅ No special cases needed for reference imports in ExtendOnEveryPath check
3. ✅ No new regressions in other extend tests
4. ✅ Architecture matches JavaScript's approach
5. ✅ Code is cleaner and more maintainable

## Additional Context

### Key Files
- `packages/less/src/less/less_go/extend_visitor.go` - Main extend logic
- `packages/less/src/less/less_go/extend.go` - Extend struct and IsVisible
- `packages/less/src/less/less_go/to_css_visitor.go` - Output filtering
- `packages/less/src/less/less_go/node.go` - Visibility tracking
- `packages/less/src/less/visitors/extend-visitor.js` - JavaScript reference

### Related Commits
- `c7b2ac7` - Current workaround (this fix)
- `a4ed54a` - Critical visibility management fixes (#204)
- `c7ec8e2` - Partial fix for import-reference extends (#205)

### Debug Environment Variables
```bash
LESS_GO_TRACE=1    # Enable trace logging
LESS_GO_DIFF=1     # Show diff output in tests
```

## Questions to Answer

1. **Ruleset instances:** Are chained extends pointing to the correct ruleset instance?
2. **Visibility propagation:** When a visible extend matches a reference ruleset, which Node gets marked visible?
3. **Path filtering:** Why isn't ToCSSVisitor filtering working for deeply chained extends?
4. **doExtendChaining:** Does it properly preserve visibility info through reference import chains?

## Final Note

The current workaround is functional and safe, but doesn't match JavaScript's cleaner architecture. The proper fix requires understanding why the normal extend chaining + visibility filtering flow doesn't work for reference imports, then fixing that root cause rather than adding special cases.

Good luck! 🚀
