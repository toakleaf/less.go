# Directive Bubbling Fix Status

## Issue
The test `directives-bubling` compiles but produces incorrect CSS output. @supports and @document directives should bubble to the top level with merged selectors, but currently stay nested.

## Root Cause Analysis

### Mechanism in JavaScript
@supports/@document use a **different bubbling mechanism** than @media:

1. **Parser**: Sets `isRooted=false` for @supports/@document
2. **JoinSelectorVisitor**: Recalculates `root=false` for nested directives → triggers selector joining
3. **Ruleset.joinSelectors()**: Merges parent selectors (e.g., `.top` + `.inside &` → `.inside .top`)
4. **ToCSSVisitor**: Extracts AtRules from parent rulesets → moves them to sibling level

This is NOT the mediaBlocks bubbling mechanism used by @media.

## Progress Made

### ✅ Completed
1. Added `GetIsRooted()` method to AtRule (atrule.go:113-116)
   - Allows JoinSelectorVisitor to detect non-rooted directives
   - Parser already sets IsRooted=false for @supports/@document (parser.go:1830-1832)

2. ToCSSVisitor extraction logic already exists (to_css_visitor.go:622-629)
   - Extracts rules with `GetRules()` to sibling level
   - Working correctly - directives DO move to top level

### ❌ Remaining Issues

**Selector joining not working correctly**

Current output:
```css
@document url-prefix() {
  .parent {
      .child {
        color: red;
      }
    }
  }
}
```

Expected output:
```css
@document url-prefix() {
  .parent .child {
    color: red;
  }
}
```

**Problem**: Selectors are not being flattened/merged inside the bubbled directive.

## What Needs to Be Fixed

### Investigation Needed
1. **JoinSelectorVisitor context propagation**
   - When visiting @document (nested in `.parent`), does context contain `.parent`?
   - Is `joinSelectors()` being called on the @document's internal ruleset?
   - Are the paths being calculated correctly?

2. **Selector flattening logic**
   - `Ruleset.JoinSelectors()` should merge parent context with child selectors
   - Need to verify it's working for rulesets inside AtRules
   - May need special handling for AtRule contexts

### Likely Fix Location
**packages/less/src/less/less_go/join_selector_visitor.go**

Possibilities:
- VisitAtRule may need to manage context stack (push/pop)
- VisitRuleset may need special case for rulesets inside non-rooted AtRules
- JoinSelectors call may need different parameters for AtRule contexts

### Testing Approach
1. Add debug logging to JoinSelectorVisitor to trace:
   - Context stack at each visit
   - Whether joinSelectors is called for @document's ruleset
   - What paths are generated

2. Compare with working @media bubbling to understand differences

## Files Involved
- ✅ `packages/less/src/less/less_go/atrule.go` - Added GetIsRooted()
- ❌ `packages/less/src/less/less_go/join_selector_visitor.go` - Needs investigation
- ✅ `packages/less/src/less/less_go/to_css_visitor.go` - Extraction works
- ✅ `packages/less/src/less/less_go/parser.go` - IsRooted setting works

## Test Command
```bash
pnpm -w test:go:filter -- "directives-bubling"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"
```

## References
- JavaScript: `packages/less/src/less/tree/nested-at-rule.js`
- JavaScript: `packages/less/src/less/visitors/join-selector-visitor.js`
- JavaScript: `packages/less/src/less/visitors/to-css-visitor.js`
