# Bugs Found Through Integration Test Variations

This document describes bugs discovered by creating variations of the original Less.js integration tests (tests 21-30: extend, extract-and-length, functions-each, functions, ie-filters, impor, import-inline, import-interpolation, import-module, import-once).

---

## Bug 1: Multi-target Extend Does Not Produce Duplicate Selectors

### Summary
When using `:extend()` with multiple comma-separated targets like `.baz:extend(.foo, .bar) {}`, the Go port produces `.baz` only once in the output, but Less.js produces it twice (once for each extended target).

### Reproduction

**Input (extend-variation.less):**
```less
.foo, .bar {
  font-weight: bold;
}
.baz:extend(.foo, .bar) {}
```

**Expected Output (from Less.js):**
```css
.foo,
.bar,
.baz,
.baz {
  font-weight: bold;
}
```

**Actual Output (from Go port):**
```css
.foo,
.bar,
.baz {
  font-weight: bold;
}
```

### Analysis
The Go version appears to be deduplicating the extended selectors when it shouldn't. When you extend multiple targets, the selector should appear in each target's selector list separately (even if that means repeating it in the output). This is the standard Less.js behavior.

### Files Likely Involved
- Extend handling in the evaluator/visitor
- Selector combination logic

### To Debug
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/custom/extend-variation" ./less
```

---

## Bug 2: Variable Interpolation in Import Paths Not Working

### Summary
The Go port does not interpolate variables in `@import` path strings. It tries to open the literal path including the `@{variable}` syntax instead of resolving the variable first.

### Reproduction

**Input (import-interpolation-variation.less):**
```less
@prefix: "shared";
@suffix: "mixins";
@import "import-support/@{prefix}-@{suffix}.less";
```

**Expected Behavior:**
Should resolve to `@import "import-support/shared-mixins.less"` and import that file.

**Actual Behavior:**
```
Compilation error: Syntax: : open import-support/@{prefix}-@{suffix}.less: no such file or directory
```

### Analysis
The variable interpolation in import paths is not being resolved before the file system lookup. This works with a single variable (e.g., `@import "import-support/theme-@{theme}.less"` works fine), but appears to fail when there are multiple variable interpolations in the path.

### Files Likely Involved
- Import path resolution
- Variable interpolation during parsing
- Import manager

### To Debug
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/custom/import-interpolation-variation" ./less
```

---

## Bug 3: Parenthesized List Values Cause Parse Errors

### Summary
The Go parser fails to parse variable assignments containing comma-separated parenthesized groups.

### Reproduction

**Input:**
```less
.test {
  @outer: (a b c), (1 2 3), (x y z);
  length: length(@outer);
}
```

**Expected Behavior (from Less.js):**
Compiles successfully (treats the whole thing as a single value with length 1).

**Actual Behavior:**
```
Parse: Unrecognised input. Possibly missing opening '{'
```

### Workaround
Use variables instead of inline parenthesized lists:
```less
.test {
  @inner1: a b c;
  @inner2: 1 2 3;
  @outer: @inner1, @inner2;
  length: length(@outer);
}
```

### Files Likely Involved
- Parser value handling
- Parenthesis handling in lists

---

## Prompts for LLMs to Investigate and Fix

### Prompt for Bug 1 (Multi-target Extend):

```
I'm working on less.go, a Go port of Less.js. I've discovered a bug with multi-target extends.

When you have `.baz:extend(.foo, .bar) {}` to extend multiple selectors at once, the Go version produces `.baz` only once in the output, but Less.js produces it twice (once for each extended target).

Input:
```less
.foo, .bar {
  font-weight: bold;
}
.baz:extend(.foo, .bar) {}
```

Expected CSS (from Less.js):
```css
.foo,
.bar,
.baz,
.baz {
  font-weight: bold;
}
```

Actual CSS (from Go port):
```css
.foo,
.bar,
.baz {
  font-weight: bold;
}
```

Please investigate the extend handling code in the less.go codebase. Look for where extend selectors are being processed and where deduplication might be happening incorrectly. The selector should be added to each target's selector group separately without deduplication.

Start by searching for extend-related code in the less/ directory and trace how multi-target extends like `.baz:extend(.foo, .bar)` are processed.
```

### Prompt for Bug 2 (Import Path Interpolation):

```
I'm working on less.go, a Go port of Less.js. I've discovered a bug with variable interpolation in @import paths.

When using multiple variable interpolations in an import path, the variables are not resolved:

Input:
```less
@prefix: "shared";
@suffix: "mixins";
@import "import-support/@{prefix}-@{suffix}.less";
```

Error:
```
Compilation error: Syntax: : open import-support/@{prefix}-@{suffix}.less: no such file or directory
```

The file `import-support/shared-mixins.less` exists, so the issue is that `@{prefix}` and `@{suffix}` are not being interpolated into the path string.

Note: Single variable interpolation works fine (e.g., `@import "theme-@{theme}.less"`), but multiple interpolations in the same path fail.

Please investigate the import path resolution code in the less.go codebase. Look for where import paths are processed and where variable interpolation should be happening. The interpolation may be working for the first variable but failing for subsequent ones, or there may be an issue with how the interpolated path is assembled.

Start by searching for import-related code and variable interpolation handlers in the less/ directory.
```

### Prompt for Bug 3 (Parenthesized Lists):

```
I'm working on less.go, a Go port of Less.js. I've discovered a parsing bug with parenthesized list values.

The Go parser fails to parse variable assignments containing comma-separated parenthesized groups:

Input:
```less
.test {
  @outer: (a b c), (1 2 3), (x y z);
}
```

Error:
```
Parse: Unrecognised input. Possibly missing opening '{'
```

Less.js compiles this successfully, treating the entire expression as a single value.

Workaround using variables instead of inline parentheses works:
```less
.test {
  @inner1: a b c;
  @inner2: 1 2 3;
  @outer: @inner1, @inner2;
}
```

Please investigate the parser code in the less.go codebase, specifically how values with parentheses are handled. The parser may be misinterpreting the parenthesized groups as something else (like a selector or mixin call) rather than treating them as values.

Start by searching for value parsing code and parenthesis handling in the less/ directory.
```
