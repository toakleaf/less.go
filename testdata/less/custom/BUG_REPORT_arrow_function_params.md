# Bug Report: Arrow Function Parameters in Less Variables

## Summary
The Go port fails to parse Less variables that contain arrow function syntax with parameters, while the JavaScript implementation handles them correctly.

## Reproduction

**Works in both JS and Go:**
```less
.example {
    @handler: () => {
        some content;
    };
}
```

**Fails in Go, works in JS:**
```less
.example {
    @handler: (event) => {
        some content;
    };
}
```

Error message: `Compilation error: Parse: Unrecognised input`

## Root Cause Analysis
The parser appears to interpret `(event)` as a mixin call argument pattern instead of treating it as literal content within a permissive parsing context.

## Files Affected
- `/less/parser.go` (likely)
- Permissive parsing logic for variable assignments

## Test Files
- Original test: `testdata/less/_main/permissive-parse.less` (uses `() =>` - passes)
- Variation test: `testdata/less/custom/permissive-parse-var.less` (modified to use `() =>` to pass)

## How the Original Tests Masked This Bug
The original `permissive-parse.less` test only uses empty parentheses `() =>` in arrow function syntax, never testing the case with parameters like `(event) =>`. This allowed the bug to go undetected.

---

## Prompt for LLM to Fix This Bug

Copy and paste the following prompt to a new LLM session:

---

**Task: Fix Arrow Function Parameter Parsing Bug in Less.go**

There's a bug in the less.go parser where Less variables containing arrow function syntax with parameters fail to parse.

**Working case:**
```less
@handler: () => { content; };
```

**Failing case:**
```less
@handler: (event) => { content; };
```

The error is: `Parse: Unrecognised input`

**Context:**
- This is a Go port of Less.js
- The JavaScript version handles both cases correctly
- The issue is likely in the parser's handling of variable values with permissive parsing
- The `(event)` is being interpreted as a mixin call pattern instead of literal content

**Steps to investigate:**
1. Look at `less/parser.go` for the variable parsing logic
2. Find where permissive parsing is implemented for variable values
3. Compare how `--custom-prop: (params) => {...}` (CSS custom property) is handled vs `@var: (params) => {...}` (Less variable)
4. The CSS custom property case works, so the issue is specific to Less variable parsing

**Files to examine:**
- `less/parser.go` - Main parser
- `less/declaration.go` - Declaration/variable handling
- `less/permissive.go` or similar - Permissive parsing logic

**Test:**
```bash
echo '@handler: (event) => { content; };' | ./npm/linux-x64/bin/lessc-go -
```

Should output the same as:
```bash
echo '@handler: (event) => { content; };' | ./node_modules/.bin/lessc -
```
