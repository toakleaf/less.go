# Fix Issue: !important Placement in Property Merge

## Problem Description

When merging properties using `+_` (space merge) where one of the values has `!important`, the Go port places `!important` at the end of the entire merged value, while Less.js preserves the position after the original value.

## Reproduction

**LESS input:**
```less
.important-merge {
  font-family+_: Georgia !important;
  font-family+_: serif;
}
```

**Less.js output:**
```css
.important-merge {
  font-family: Georgia !important serif;
}
```

**Go port output:**
```css
.important-merge {
  font-family: Georgia serif !important;
}
```

## Test File

- `testdata/less/custom/var-merge.less`

## Analysis

The issue is in how `!important` is handled during property merging:

1. Less.js keeps `!important` attached to the value it was originally associated with
2. The Go port moves `!important` to the end of the entire merged value

This is a semantic difference - while both result in the property being important, the exact output differs from Less.js.

## Investigation Areas

1. **Property merge logic:**
   - Find where `+` and `+_` merge operations are handled
   - Look for how `!important` flags are tracked and output

2. **Declaration/Value handling:**
   - Check how values store their `!important` status
   - See if `!important` is being detached from the value during merge

3. **Merge stringification:**
   - Look at how merged values are converted to CSS strings
   - Find where `!important` is appended to the output

## Files to Examine

- `less/tree_declaration.go` - Declaration handling
- `less/tree_value.go` - Value handling and merging
- `less/visitors.go` or `less/visitor_*.go` - Merge processing visitors
- Look for any merge-related functions that handle `+` and `+_` property syntax

## Expected Behavior

When merging properties where the first value has `!important`:
```less
property+_: value1 !important;
property+_: value2;
```

Should output:
```css
property: value1 !important value2;
```

Not:
```css
property: value1 value2 !important;
```

## Verification

After fixing, run:
```bash
LESS_GO_DIFF=1 LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run TestIntegrationSuite/custom/var-merge
```

The output should exactly match:
```css
.important-merge {
  font-family: Georgia !important serif;
}
```
