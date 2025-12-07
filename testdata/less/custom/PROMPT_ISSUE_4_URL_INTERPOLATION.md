# Fix Issue: Variable Interpolation in URL Strings

## Problem Description

Variable interpolation using `@{variable}` syntax inside `url()` strings is not being processed. The variable placeholder is left in the output instead of being replaced with the variable's value.

## Reproduction

**LESS input:**
```less
@prefix: "prefix";

.string-ops {
  background: url("path/@{prefix}.png");
}
```

**Less.js output:**
```css
.string-ops {
  background: url("path/prefix.png");
}
```

**Go port output:**
```css
.string-ops {
  background: url("path/@{prefix}.png");
}
```

## Test File

- `testdata/less/custom/var-operations.less`

## Analysis

The issue is that variable interpolation (`@{variable}`) is not being applied to strings inside `url()` function calls. This is a significant bug because URL interpolation is commonly used for:

1. Building dynamic paths based on variables
2. Adding version hashes to cache-bust URLs
3. Creating themed asset paths

## Investigation Areas

1. **URL node evaluation:**
   - Find the URL node type (`tree_url.go` or similar)
   - Check if it processes variable interpolation in its `Eval()` method

2. **String/Quoted value interpolation:**
   - Look at how `@{variable}` syntax is parsed and evaluated
   - Check if there's a special case missing for URL values

3. **Interpolation visitor:**
   - There may be a visitor that processes variable interpolation
   - Verify it visits URL node children correctly

4. **Compare with Less.js:**
   - Look at how Less.js handles URL value interpolation
   - Check `lib/less/tree/url.js` in Less.js source

## Files to Examine

- `less/tree_url.go` - URL node definition and evaluation
- `less/tree_quoted.go` - Quoted string handling
- `less/parser.go` - How URL values are parsed
- Look for interpolation-related code that might skip URL values

## Expected Behavior

Any quoted string containing `@{variable}` syntax should have the variable replaced with its value during evaluation, including strings inside `url()`:

```less
@var: "value";
background: url("path/@{var}.png");
```

Should output:
```css
background: url("path/value.png");
```

## Verification

After fixing, run:
```bash
LESS_GO_DIFF=1 LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run TestIntegrationSuite/custom/var-operations
```

The output should show:
```css
.string-ops {
  background: url("path/prefix.png");
}
```
