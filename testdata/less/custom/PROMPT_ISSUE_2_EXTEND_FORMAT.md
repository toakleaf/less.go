# Fix Issue: Extend Property Formatting Difference

## Problem Description

When using `:extend()` to inherit properties from another selector, the Go port outputs each property on a separate line, while Less.js outputs them on the same line.

## Reproduction

**LESS input:**
```less
.ref-extend-base {
  display: block;
  position: relative;
}

.my-card:extend(.ref-extend-base) {
  background: #f5f5f5;
}
```

**Less.js output:**
```css
.ref-extend-base,
.my-card {
  display: block;position: relative;
}
.my-card {
  background: #f5f5f5;
}
```

**Go port output:**
```css
.ref-extend-base,
.my-card {
  display: block;
  position: relative;
}
.my-card {
  background: #f5f5f5;
}
```

## Test Files

- `testdata/less/custom/var-import.less`
- `testdata/less/custom/var-import-reference.less`

## Analysis

This appears to be a CSS generation/formatting issue in how properties are output when they come from an extended selector. The difference is in the newline handling.

Looking at the diff:
```
Expected: "  display: block;position: relative;"
Actual:   "  display: block;"
         (next line) "  position: relative;"
```

## Investigation Areas

1. **CSS output generation:**
   - Find where CSS declarations are stringified
   - Look for formatting options that control newlines between properties
   - Check if there's special handling for extended/merged selectors

2. **Extend processing:**
   - Look at how extend merges properties from source selector
   - Check if the merged properties maintain their original formatting

3. **Compare with Less.js:**
   - In Less.js source, see how they format properties in extended selectors
   - Determine if this is intentional behavior or a quirk of their implementation

## Files to Examine

- `less/tree_ruleset.go` - Ruleset generation
- `less/tree_declaration.go` - Declaration/property output
- `less/extend.go` or similar - Extend processing
- `less/visitor_extend.go` - Extend visitor

## Consideration

This is a whitespace-only difference that doesn't affect the semantic meaning of the CSS. You may choose to:

1. **Match Less.js exactly** - Modify output to put extended properties on the same line
2. **Accept the difference** - Document it as an intentional formatting difference and update the expected CSS files
3. **Add a formatting option** - Allow users to choose their preferred formatting

## Verification

After fixing (or updating expected files), run:
```bash
LESS_GO_DIFF=1 LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run "TestIntegrationSuite/custom/var-import"
```
