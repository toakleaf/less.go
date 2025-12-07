# Issues Found in Custom Variation Tests

This document summarizes the issues discovered by running variation tests against the 4th set of 10 integration tests from Less.js (tests 31-40 alphabetically: import-reference-issues, import-reference, import-remote, import, javascript, layer, lazy-eval, media, merge, mixin-noparens).

## Summary

- **12 tests passed** (perfectly matching Less.js output)
- **1 compilation failure** (var-javascript)
- **4 output differences** (var-import, var-import-reference, var-merge, var-operations)

---

## Issue 1: JavaScript Evaluation Not Enabled for Custom Tests

**Test file:** `var-javascript.less`

**Error:** `Syntax: JavaScript: inline JavaScript is not enabled. Is it set in your options?`

**Description:**
The `javascriptEnabled: true` option is being set in the test options, but the JavaScript evaluation context shows `JavascriptEnabled=false`. The option is not being propagated correctly to the evaluation context.

**Debug output shows:**
```
[DEBUG] Starting compilation with options: map[... javascriptEnabled:true ...]
[DEBUG JsEvalNode] *Eval context, JavascriptEnabled=false, PluginBridge=<nil>
```

---

## Issue 2: Extend Property Formatting Difference

**Test files:** `var-import.less`, `var-import-reference.less`

**Description:**
When using `:extend()`, properties from the extended selector are formatted differently between Less.js and the Go port.

**Less.js output:**
```css
.extended-card {
  display: block;position: relative;
}
```

**Go port output:**
```css
.extended-card {
  display: block;
  position: relative;
}
```

**Impact:** This is primarily a whitespace/formatting difference, but it causes test comparison failures.

---

## Issue 3: !important Placement in Property Merge

**Test file:** `var-merge.less`

**Description:**
When merging properties with `+_` (space merge) where one property has `!important`, the placement of `!important` differs.

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

**Impact:** The Go port moves `!important` to the end of the entire merged value, while Less.js preserves its position after the original value.

---

## Issue 4: Variable Interpolation in URL Strings

**Test file:** `var-operations.less`

**Description:**
Variable interpolation inside `url()` strings is not being processed.

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

**Impact:** Variables are not interpolated inside URL strings, leaving the `@{variable}` syntax in the output.

---

## Prompts for Fixing Issues

See the individual prompt files below for detailed instructions to give to an LLM to fix each issue.
