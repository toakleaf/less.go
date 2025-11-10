# Import Reference Bug - Root Cause Analysis

**Date**: 2025-11-10
**Status**: Root cause identified, structural fix needed

## Problem Summary

The `import-reference` and `import-reference-issues` tests are failing because:

1. Rulesets containing ONLY reference imports are being output as empty rulesets (e.g., `#do-not-show-import { }`)
2. Reference import content is appearing when it shouldn't
3. Whitespace/formatting issues in the output

## Expected Behavior

When a file is imported with `@import (reference)`:
- The imported content should NOT appear in the CSS output by default
- Only explicitly used content (via extends or mixin calls) should appear
- Parent rulesets that contain ONLY reference imports should also not appear

Example:
```less
#do-not-show-import {
  @import (reference) "file.less";
}
```
**Expected output**: Nothing (the ruleset should not appear at all)
**Actual output**: `#do-not-show-import { }` (empty ruleset)

## How JavaScript Handles It

### 1. Import Evaluation (`packages/less/src/less/tree/import.js:114-127`)

```javascript
eval(context) {
    const result = this.doEval(context);
    if (this.options.reference || this.blocksVisibility()) {
        // Add visibility blocks to all imported content
        result.forEach(node => node.addVisibilityBlock());
    }
    return result;
}
```

**Key insight**: Import.eval returns the IMPORTED CONTENT (not the Import node). For reference imports, all imported nodes get visibility blocks added.

### 2. Visibility Mechanism (`packages/less/src/less/tree/node.js:128-169`)

- `visibilityBlocks`: Counter that tracks how many visibility blocks a node has
- `blocksVisibility()`: Returns true if `visibilityBlocks > 0`
- `nodeVisible`: Can be `true` (must show), `false` (must hide), or `undefined` (inherit)
- `ensureVisibility()`: Sets `nodeVisible = true` to make content visible (used by extends/mixin calls)

### 3. ToCSSVisitor Filtering (`packages/less/src/less/visitors/to-css-visitor.js`)

The visitor has methods that filter out invisible content:

```javascript
visitDeclaration(declNode, visitArgs) {
    if (declNode.blocksVisibility() || declNode.variable) {
        return;  // Return undefined = filter out
    }
    return declNode;
}

visitRuleset(rulesetNode, visitArgs) {
    // ... process rules ...

    // Decide whether to keep the ruleset
    if (this.utils.isVisibleRuleset(rulesetNode)) {
        rulesetNode.ensureVisibility();
        rulesets.splice(0, 0, rulesetNode);
    }

    return rulesets;
}
```

**Key insight**: When a visitor method returns `undefined/null`, the Visitor.visitArray method filters it out of the array.

### 4. Ruleset Visibility Check (`packages/less/src/less/visitors/to-css-visitor.js:68-82`)

```javascript
isVisibleRuleset(rulesetNode) {
    if (rulesetNode.firstRoot) {
        return true;
    }

    if (this.isEmpty(rulesetNode)) {  // <--- CRITICAL CHECK
        return false;
    }

    if (!rulesetNode.root && !this.hasVisibleSelector(rulesetNode)) {
        return false;
    }

    return true;
}
```

**Key insight**: Empty rulesets return `false` and are not output.

## How Go Implements It

The Go implementation (`packages/less/src/less/less_go/`) follows the same architecture:

### 1. Import Evaluation ✅ CORRECT

`import.go:396-413` correctly adds visibility blocks to imported content:
```go
func (i *Import) Eval(context any) (any, error) {
    result, err := i.DoEval(context)
    if err != nil {
        return nil, err
    }

    if i.getBoolOption("reference") || i.BlocksVisibility() {
        if resultSlice, ok := result.([]any); ok {
            for _, node := range resultSlice {
                addVisibilityBlockRecursive(node)
            }
        } else {
            addVisibilityBlockRecursive(result)
        }
    }

    return result, nil
}
```

### 2. Visibility Mechanism ✅ CORRECT

`node.go` implements the same visibility methods correctly.

### 3. ToCSSVisitor Filtering ✅ CORRECT

`to_css_visitor.go` correctly filters out invisible content:
```go
func (v *ToCSSVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) any {
    if blockedNode, hasBlocked := declNode.(interface{ BlocksVisibility() bool }); hasBlocked {
        if blockedNode.BlocksVisibility() {
            return nil  // Filtered out
        }
    }
    // ...
}
```

### 4. Ruleset Visibility Check ⚠️ **ISSUE HERE**

`to_css_visitor.go:181-249` has a logic flow problem:

```go
func (u *CSSVisitorUtils) IsVisibleRuleset(rulesetNode any) bool {
    if rulesetNode == nil {
        return false
    }

    // Check 1: First root nodes are always visible
    if firstRootNode, ok := rulesetNode.(interface{ GetFirstRoot() bool }); ok {
        if firstRootNode.GetFirstRoot() {
            return true  // <--- Returns WITHOUT checking isEmpty!
        }
    }

    // Check 2: Empty rulesets are not visible
    if u.IsEmpty(rulesetNode) {
        return false
    }

    // Check 3: Reference imports that haven't been used
    if blockedNode, ok := rulesetNode.(interface{ BlocksVisibility() bool; IsVisible() *bool }); ok {
        if blockedNode.BlocksVisibility() {
            vis := blockedNode.IsVisible()
            if vis == nil || !*vis {
                return false
            }
        }
    }

    // Check 4-6: Special cases for MultiMedia, AllowImports, etc.
    // ...

    // Check 7: No selector
    if rootNode, ok := rulesetNode.(interface{ GetRoot() bool }); ok {
        if !rootNode.GetRoot() && !u.HasVisibleSelector(rulesetNode) {
            return false
        }
    }

    return true  // <--- Default: keep the ruleset
}
```

## The Root Cause

The problem is in the **order of checks and the default behavior** in `IsVisibleRuleset`:

1. `FirstRoot` check returns `true` WITHOUT checking if the ruleset is empty
2. `IsEmpty` check happens early but might not catch all cases
3. The function defaults to returning `true` at the end

For a ruleset like `#do-not-show-import` that contains only reference imports:
- It's not a first root ✓
- After visiting, all its rules are filtered out (imported content blocks visibility)
- `IsEmpty` should return `true`... but does it?

### Investigating IsEmpty

`to_css_visitor.go:88-113`:
```go
func (u *CSSVisitorUtils) IsEmpty(owner any) bool {
    if owner == nil {
        return true
    }

    if ownerWithRules, ok := owner.(interface{ GetRules() []any }); ok {
        rules := ownerWithRules.GetRules()
        isEmpty := rules == nil || len(rules) == 0  // <--- Should catch Rules=0
        if !isEmpty {
            // Check if all rules are nil or invisible
            allNil := true
            for _, r := range rules {
                if r != nil {
                    allNil = false
                    break
                }
            }
            if allNil {
                isEmpty = true
            }
        }
        return isEmpty
    }

    return true
}
```

This SHOULD return `true` when `len(rules) == 0`. But debug output shows rulesets with `Rules=0` are still being output!

## The Actual Bug (Hypothesis)

Based on debug output showing `Rules=0` but rulesets still output, I believe the issue is one of:

1. **Timing**: `IsEmpty` is being called BEFORE rules are visited/filtered
2. **State mutation**: The rules array changes AFTER `IsVisibleRuleset` is called
3. **Wrong node**: `IsVisibleRuleset` is checking a different node than the one being output

Looking at `VisitRuleset` (`to_css_visitor.go:607-792`):

```go
func (v *ToCSSVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) any {
    // Line 662: Debug print showing Rules count

    // Lines 666-713: Visit nested rules, extract nested rulesets
    // This is where rules get filtered!

    // Lines 716-738: Merge and deduplicate rules

    // Line 741: Check if we should keep the ruleset
    keepRuleset := v.utils.IsVisibleRuleset(rulesetNode)

    // Lines 753-769: Special case - might force keepRuleset=true

    if keepRuleset {
        // Lines 771-785: Add to output
        rulesets = append([]any{rulesetNode}, rulesets...)
    }

    return rulesets
}
```

**AH! I found it!** Lines 753-769:

```go
// Special case: if we extracted nested rulesets and the parent has non-variable declarations,
// we should keep it even if paths were filtered
if !keepRuleset && len(rulesets) > 0 {
    if nodeWithRules, ok := rulesetNode.(interface{ GetRules() []any }); ok {
        rules := nodeWithRules.GetRules()
        if rules != nil {
            for _, rule := range rules {
                // Check if it's a non-variable declaration
                if decl, ok := rule.(interface{ GetVariable() bool }); ok {
                    if !decl.GetVariable() {
                        // Has at least one non-variable declaration
                        keepRuleset = true  // <--- FORCES keepRuleset=true!
                        break
                    }
                }
            }
        }
    }
}
```

This special case might be incorrectly forcing `keepRuleset=true` even for empty rulesets!

## The Fix

The issue is likely in how the "special case" logic at lines 753-769 interacts with empty rulesets. This logic should NOT apply to rulesets that are completely empty after filtering.

The fix should ensure:
1. `IsVisibleRuleset` correctly identifies empty rulesets
2. The special case at lines 753-769 doesn't override the empty check
3. Empty rulesets are never output

## Alternative Hypothesis

Another possibility: The ruleset contains something OTHER than filtered rules that makes it appear non-empty. For example:
- Comments that block visibility
- Variable declarations
- Other nodes that aren't being filtered

Need to add more debug output to see exactly what Rules contains when `Rules=0` is printed but the ruleset is still kept.

## Next Steps

1. Add detailed debug output to show:
   - What rules are in the ruleset at each stage
   - What `IsEmpty` returns
   - What `IsVisibleRuleset` returns
   - Whether the special case at line 753 is triggered

2. Run the test with this debug output to identify the exact flow

3. Implement the fix based on findings

4. Test with both `import-reference` and `import-reference-issues` tests

## Related Files

- `packages/less/src/less/less_go/to_css_visitor.go` - The main issue
- `packages/less/src/less/less_go/import.go` - Import evaluation (working correctly)
- `packages/less/src/less/less_go/node.go` - Visibility methods (working correctly)
- `packages/less/src/less/tree/import.js` - JavaScript reference
- `packages/less/src/less/visitors/to-css-visitor.js` - JavaScript reference
