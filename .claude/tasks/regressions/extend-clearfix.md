# Task: Fix Extend Clearfix Regression

**Status**: Available
**Priority**: CRITICAL
**Estimated Time**: 2-3 hours
**Complexity**: Medium
**Tests Affected**: 1 (extend-clearfix)

## Overview

The `extend-clearfix` test was previously a **perfect match** but has regressed to producing incorrect CSS output. The `:extend()` pseudo-class with `all` flag is no longer properly extending all selectors (including nested selectors like `:after`).

This is a **REGRESSION** - this test was working perfectly before recent changes.

## Failing Test

### extend-clearfix (CRITICAL REGRESSION)
- **Previous Status**: ✅ Perfect match
- **Current Status**: ⚠️ Output differs
- **Issue**: `:extend(.clearfix all)` not extending nested selectors
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/extend-clearfix"
  ```

## Current Behavior

### Input LESS:
```less
.clearfix {
  *zoom: 1;
  &:after {
    content: '';
    display: block;
    clear: both;
    height: 0;
  }
}

.foo {
  &:extend(.clearfix all);  // Should extend ALL clearfix selectors
  color: red;
}

.bar {
  &:extend(.clearfix all);
  color: blue;
}
```

### Current Output (WRONG ❌):
```css
.clearfix {
  *zoom: 1;
}
.clearfix:after {
  content: '';
  display: block;
  clear: both;
  height: 0;
}
.foo {
  color: red;
}
.bar {
  color: blue;
}
```

**Problem**: `.foo` and `.bar` are NOT getting the clearfix styles. The `&:extend(.clearfix all)` is being ignored or not working properly.

## Expected Behavior

### Expected Output (CORRECT ✅):
```css
.clearfix,
.foo,
.bar {
  *zoom: 1;
}
.clearfix:after,
.foo:after,
.bar:after {
  content: '';
  display: block;
  clear: both;
  height: 0;
}
.foo {
  color: red;
}
.bar {
  color: blue;
}
```

**Expected**: When using `:extend(.clearfix all)`, both `.foo` and `.bar` should:
1. Be added to all `.clearfix` selectors (including `.clearfix` itself)
2. Be added to all nested/derived selectors (like `.clearfix:after`)

The `all` keyword means "extend this selector everywhere it appears, including with pseudo-classes, pseudo-elements, and other modifications".

## Investigation Starting Points

### Files to Examine

1. **`extend_visitor.go`** (PRIMARY) - Lines 150-400
   - The ExtendVisitor processes :extend() pseudo-classes
   - Check how the `all` flag is handled
   - Verify selector matching logic for nested selectors

2. **`extend.go`** - Lines 30-100
   - Extend node representation
   - Check if `AllExtends` flag is being set/read correctly

3. **`selector.go`** - Lines 150-250
   - Selector matching logic
   - How selectors are compared/matched for extending

4. **`ruleset.go`** - Lines 200-350
   - How rulesets handle extend processing
   - Nested selector handling

### Debug Commands

```bash
# See actual vs expected output
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/extend-clearfix"

# Trace extend processing
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "main/extend-clearfix" 2>&1 | grep -i "extend"

# Compare with working extend test
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/extend"
```

### Recent Changes

This was working before. Check what changed:

```bash
# Check recent changes to extend_visitor.go
git log -p --since="2025-11-01" -- packages/less/src/less/less_go/extend_visitor.go | head -300

# Find when this test last passed
git log --all --oneline | head -20
```

## Root Cause Hypothesis

**Most Likely**: Recent changes to extend processing broke the `all` flag functionality. Possibilities:

1. **The `all` flag isn't being respected** - The extend visitor might be ignoring the `all` flag and only extending the base selector

2. **Nested selector matching is broken** - The visitor might not be finding nested selectors like `.clearfix:after` to extend

3. **Selector cloning/modification is broken** - Even if nested selectors are found, the logic to create `.foo:after` from `.clearfix:after` might be failing

## Success Criteria

- [ ] `extend-clearfix` test produces exact expected output
- [ ] `.foo` and `.bar` appear in both `.clearfix` and `.clearfix:after` selector lists
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] FULL integration test suite shows no new regressions: `pnpm -w test:go`
- [ ] Other extend tests still pass (extend, extend-chaining, extend-exact, etc.)

## Validation Checklist

Before creating PR:

```bash
# 1. Verify specific test passes
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/extend-clearfix"
# Expected: ✅ Perfect match (no diff output)

# 2. Verify other extend tests still work
pnpm -w test:go 2>&1 | grep "extend" | grep -E "(✅|❌)"
# Expected: No regressions in other extend tests

# 3. Run ALL unit tests (catch any regressions) - REQUIRED
pnpm -w test:go:unit
# Expected: ✅ All unit tests pass (no failures)

# 4. Run FULL integration test suite - REQUIRED
pnpm -w test:go
# Expected:
# - ✅ 15+ perfect matches (14 currently, +1 for this fix)
# - ✅ No new compilation failures
# - ✅ namespacing-6 and other critical tests not broken
```

**If any test fails that was passing before**: STOP and fix the regression before proceeding.

## Additional Context

### Compare with JavaScript Implementation

The less.js implementation is in:
- `packages/less/src/less-tree/extend.js`
- `packages/less/src/less-tree/extend-visitor.js`

Key JavaScript logic to understand:
```javascript
// In extend-visitor.js
if (extend.option === 'all') {
    // Should match the selector in all contexts
    // Including with pseudo-classes, pseudo-elements, etc.
}
```

### Related Tests

These extend tests are currently passing and should remain passing:
- `extend` - Basic extend functionality
- `extend-chaining` - Extending extended selectors
- `extend-exact` - Exact selector matching
- `extend-media` - Extends within media queries
- `extend-nest` - Nested extend scenarios
- `extend-selector` - Complex selector extending

Make sure your fix doesn't break any of these.

## Notes

This is a regression, so the functionality existed before. The fix might involve:
- Reverting a recent problematic change
- Re-enabling disabled code
- Fixing a conditional that was changed

Priority is to restore the perfect match status quickly while maintaining all other test results.
