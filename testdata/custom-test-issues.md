# Custom Integration Test Issues

This document describes issues found during testing of custom integration tests that were created as variations of the original Less.js tests (tests 81-90 alphabetically: plugin-module, plugin-preeval, postProcessor, preProcessor, property-accessors, property-name-interp, rewrite-urls-all, rewrite-urls-local, root, rootpath-rewrite-urls-all).

## Summary

- **10 new custom tests created**
- **5 tests passed** (perfect CSS match with Less.js)
- **5 tests failed** (CSS output differs from Less.js)

## Issues Found

### Issue 1: URL Variable Interpolation Not Working

**Affected Tests:** `url-handling-variation`, `rewrite-paths-variation`, `rootpath-url-variation`

**Description:** Variable interpolation using `@{variable}` syntax inside `url()` strings is not being evaluated. The Go port outputs the literal interpolation syntax instead of the resolved value.

**Example:**
```less
.url-variable {
  @img-path: "assets/images";
  background: url("@{img-path}/logo.png");
}
```

**Expected CSS (Less.js):**
```css
.url-variable {
  background: url("assets/images/logo.png");
}
```

**Actual CSS (less.go):**
```css
.url-variable {
  background: url("@{img-path}/logo.png");
}
```

---

### Issue 2: Namespace Property Access with Math Operations

**Affected Tests:** `module-pattern-variation`

**Description:** When accessing a namespace property via `#namespace[@variable]` and using it in a math operation, the math is not evaluated.

**Example:**
```less
#typography {
  @base-size: 16px;
}

.card-title {
  font-size: #typography[@base-size] * 1.25;
}
```

**Expected CSS (Less.js):**
```css
.card-title {
  font-size: 20px;
}
```

**Actual CSS (less.go):**
```css
.card-title {
  font-size: 16px*1.25;
}
```

---

### Issue 3: Property Merge (+) Duplicating Values

**Affected Tests:** `property-accessors-variation`

**Description:** When using property merge syntax (`property+:`) and then accessing the merged value with `$property`, the merged value contains duplicated entries.

**Example:**
```less
.accessor-merged {
  transform+: rotate(45deg);
  transform+: scale(1.5);
  .result {
    animation-transform: $transform;
  }
}
```

**Expected CSS (Less.js):**
```css
.accessor-merged {
  transform: rotate(45deg), scale(1.5);
}
.accessor-merged .result {
  animation-transform: rotate(45deg), scale(1.5);
}
```

**Actual CSS (less.go):**
```css
.accessor-merged {
  transform: rotate(45deg), scale(1.5), scale(1.5);
}
.accessor-merged .result {
  animation-transform: rotate(45deg), scale(1.5);
}
```

Note: The parent selector has `scale(1.5)` duplicated.

---

## LLM Prompts for Fixing Issues

Below are detailed prompts you can copy and paste to another LLM session to investigate and fix each issue.

---

### Prompt for Issue 1: URL Variable Interpolation

```
I'm working on less.go, a Go port of Less.js. I've found an issue where variable interpolation inside url() strings is not working correctly.

## The Problem

When a variable is used inside a url() string like this:

```less
.example {
  @img-path: "assets/images";
  background: url("@{img-path}/logo.png");
}
```

Less.js correctly outputs:
```css
.example {
  background: url("assets/images/logo.png");
}
```

But less.go outputs the literal interpolation without resolving the variable:
```css
.example {
  background: url("@{img-path}/logo.png");
}
```

## Task

Please investigate and fix the URL variable interpolation issue. The fix should:

1. Find where url() values are processed during evaluation
2. Ensure variable interpolation (the @{variable} syntax) is applied to quoted strings inside url()
3. Handle multiple interpolations in a single url string
4. Handle interpolation at any position in the url string

## Key Files to Investigate

Look at:
- How quoted strings are evaluated and where interpolation is applied
- The url() function handling in the evaluator
- Compare with how interpolation works in other string contexts (it works in regular property values)

## Test Cases

Run the custom integration tests:
```bash
LESS_GO_CUSTOM_ONLY=1 LESS_GO_DIFF=1 pnpm test:go
```

Specifically check: url-handling-variation, rewrite-paths-variation, rootpath-url-variation
```

---

### Prompt for Issue 2: Namespace Property Math Operations

```
I'm working on less.go, a Go port of Less.js. I've found an issue where math operations with namespace property access don't evaluate correctly.

## The Problem

When accessing a variable from a namespace and using it in a math operation:

```less
#typography {
  @base-size: 16px;
}

.card-title {
  font-size: #typography[@base-size] * 1.25;
}
```

Less.js correctly outputs:
```css
.card-title {
  font-size: 20px;
}
```

But less.go outputs the math expression unevaluated:
```css
.card-title {
  font-size: 16px*1.25;
}
```

## Task

Please investigate and fix the namespace property access with math operations. The fix should:

1. Find where namespace property access (#namespace[@variable]) is evaluated
2. Ensure the resulting value is properly typed as a Dimension (number with unit)
3. Ensure the math operation (* 1.25) is then evaluated with this Dimension

## Key Files to Investigate

Look at:
- Namespace/ID selector property access (the [@variable] syntax)
- How the Operation node handles multiplication
- Check if there's an issue with the order of evaluation or type detection

## Test Cases

Run the custom integration tests:
```bash
LESS_GO_CUSTOM_ONLY=1 LESS_GO_DIFF=1 pnpm test:go
```

Specifically check: module-pattern-variation
```

---

### Prompt for Issue 3: Property Merge Duplication

```
I'm working on less.go, a Go port of Less.js. I've found an issue where property merge (+) is duplicating values in certain cases.

## The Problem

When using property merge syntax and the value appears in nested contexts:

```less
.accessor-merged {
  transform+: rotate(45deg);
  transform+: scale(1.5);
  .result {
    animation-transform: $transform;
  }
}
```

Less.js correctly outputs:
```css
.accessor-merged {
  transform: rotate(45deg), scale(1.5);
}
.accessor-merged .result {
  animation-transform: rotate(45deg), scale(1.5);
}
```

But less.go outputs with a duplicated scale(1.5):
```css
.accessor-merged {
  transform: rotate(45deg), scale(1.5), scale(1.5);
}
.accessor-merged .result {
  animation-transform: rotate(45deg), scale(1.5);
}
```

Note that the parent selector's transform has scale(1.5) twice, but the child's $transform accessor is correct.

## Task

Please investigate and fix the property merge duplication. The fix should:

1. Find where property merging (the + suffix) is processed
2. Check why values are being duplicated in the parent selector
3. Ensure the merge only happens once per property value

## Key Files to Investigate

Look at:
- Property merging logic (properties with + suffix)
- The evaluation pass that combines merged properties
- Check if there's a double-evaluation or copy issue

## Test Cases

Run the custom integration tests:
```bash
LESS_GO_CUSTOM_ONLY=1 LESS_GO_DIFF=1 pnpm test:go
```

Specifically check: property-accessors-variation
```

---

## Passing Tests

The following custom tests pass with perfect CSS matches:

1. `detached-ruleset-preeval-variation` - Detached ruleset patterns
2. `plugin-functions-simulation` - Built-in function variations
3. `pre-post-process-simulation` - Variable injection patterns
4. `property-name-interp-variation` - Property name interpolation
5. `root-variable-scope-variation` - Root scope and variable hoisting

## Test Files Location

- LESS inputs: `testdata/less/custom/*.less`
- Expected CSS: `testdata/css/custom/*.css`

## Running Tests

```bash
# Run only custom tests
LESS_GO_CUSTOM_ONLY=1 pnpm test:go

# Run with diff output
LESS_GO_CUSTOM_ONLY=1 LESS_GO_DIFF=1 pnpm test:go

# Run full test suite (includes custom)
pnpm test:go
```
