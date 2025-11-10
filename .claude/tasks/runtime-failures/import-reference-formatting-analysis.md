# Import-Reference Formatting Issue - Analysis

## Status: NEEDS FURTHER INVESTIGATION

## Problem Summary
The `import-reference` and `import-reference-issues` integration tests produce functionally correct CSS but with formatting differences:

### Issue 1: Missing Newlines Between Rulesets
When nested rulesets are extracted from their parent and become siblings, they're missing newlines before them.

**Expected:**
```css
show-all-content {
  /* comment */
}
show-all-content .fix {
  fix: fix;
}
```

**Actual:**
```css
show-all-content {
  /* comment */
}show-all-content .fix {
  fix: fix;
}
```

### Issue 2: Extra Indentation
Rulesets and their contents have extra indentation (4 spaces instead of 2).

### Issue 3: Extra Blank Lines
Some rulesets have extra blank lines inside them.

## Root Cause Analysis

### How Nested Rulesets Are Extracted
1. When a file is imported inside a ruleset (e.g., `show-all-content { @import "file.less"; }`), the imported content becomes children of that ruleset
2. During ToCSSVisitor, nested rulesets are extracted at `to_css_visitor.go:691-715`
3. They're visited and added to a `rulesets` slice
4. They become siblings of the parent ruleset in the output

### Newline Logic in Ruleset.GenCSS
In `ruleset.go:1777-1802`, newlines are added after rules based on:
1. If `rule.IsVisible()` returns true (for declarations)
2. If the rule is a ruleset inside a container AND `tabLevel > 0`

**Problem:** Extracted rulesets become children of the root (tabLevel == 0), so they don't get newlines from condition #2.

### JavaScript Behavior
In JavaScript (`ruleset.js:550-554`), newlines are added if `rule.isVisible()` returns true.
- JavaScript: `Node.isVisible()` returns `this.nodeVisible`
- Go: `Node.IsVisible()` returns `*bool`

The type assertion in Go checks for `IsVisible() bool` but Node returns `*bool`, causing the check to fail for rulesets.

## Attempted Fixes and Results

### Attempt 1: Change Type Assertion
Changed `ruleset.go` to check for `IsVisible() *bool` instead of `IsVisible() bool`.

**Result:** FAILED - Caused major regressions (79 → 17 passing tests)
**Why:** Declarations have `IsVisible() bool`, not `*bool`. The change broke declarations.

### Attempt 2: Mark Extracted Rulesets as Visible
Added explicit `EnsureVisibility()` call for extracted rulesets in `to_css_visitor.go:700-705`.

**Result:** NO IMPROVEMENT - import-reference tests still failed, no regressions
**Why:** The visibility was likely already being set elsewhere, or the newline logic doesn't depend on it.

### Attempt 3: Both Changes Together
Combined both changes.

**Result:** Partial success but too many regressions
- Newlines were added between extracted rulesets ✅
- But extra newlines appeared everywhere ❌
- Indentation was still wrong ❌

## Key Insights

1. **Type Mismatch:** The Go codebase has inconsistent return types for `IsVisible()`:
   - `Node.IsVisible()` returns `*bool`
   - Declarations seem to have `IsVisible() bool`

   The current code only checks for `bool`, missing rulesets.

2. **Context Inheritance:** When extracted rulesets are visited, they inherit the parent's context including `tabLevel`, which may cause indentation issues.

3. **Root Level Special Case:** The root ruleset has `tabLevel == 0`, so the container check `if isContainer && tabLevel > 0` fails for top-level rulesets.

## Recommended Approach for Future Fix

### Option 1: Fix Type Assertions (Safest)
Update `ruleset.go:1782` to check for BOTH `*bool` and `bool`:
```go
// Check for *bool (Node.IsVisible)
if vis, ok := rule.(interface{ IsVisible() *bool }); ok {
    if visPtr := vis.IsVisible(); visPtr != nil && *visPtr {
        shouldAddNewline = true
    }
}
// Also check for bool (other nodes)
if !shouldAddNewline {
    if vis, ok := rule.(interface{ IsVisible() bool }); ok && vis.IsVisible() {
        shouldAddNewline = true
    }
}
```

**Risk:** Need to ensure this doesn't add extra newlines for declarations

### Option 2: Special Case for Root Level
Modify the container check to also apply when `tabLevel == 0` but only for specific cases:
```go
if isContainer && (tabLevel > 0 || r.Root) {
    shouldAddNewline = true
}
```

**Risk:** May add extra newlines in unintended places

### Option 3: Investigate JavaScript More Deeply
The JavaScript code might have additional logic we're missing. Need to:
- Trace through how JavaScript handles extracted rulesets
- Check if there's special handling for reference imports
- Verify how `nodeVisible` is set for different node types

## Test Files
- Input: `packages/test-data/less/_main/import-reference-issues.less`
- Expected: `packages/test-data/css/_main/import-reference-issues.css`
- Test: Run with `pnpm -w test:go -- -run "import-reference-issues$"`

## Related Files
- `packages/less/src/less/less_go/ruleset.go` - Ruleset.GenCSS (line 1509+)
- `packages/less/src/less/less_go/to_css_visitor.go` - VisitRuleset (line 622+)
- `packages/less/src/less/tree/ruleset.js` - JavaScript reference implementation
- `packages/less/src/less/visitors/to-css-visitor.js` - JavaScript ToCSSVisitor

## Priority
**MEDIUM** - The functionality is correct (empty rulesets are filtered), this is purely a cosmetic formatting issue affecting 2 tests.
