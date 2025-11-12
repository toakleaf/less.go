# Expected Error Tests - Tests That Should Fail But Don't

**Last Updated:** 2025-11-11
**Total Count:** 27 tests (14.7% of active tests)
**Priority:** MEDIUM (error handling is important but doesn't affect successful compilations)

## Overview

These are tests in the `eval-errors` and `parse-errors` suites that are expected to fail with an error, but currently compile successfully. This indicates missing validation and error handling in the Go port.

## Current Status

From the latest test run:

- **eval-errors suite:** 23 tests failing to error properly
- **parse-errors suite:** 4 tests failing to error properly

**Total:** 27 tests that should error but succeed

## Test Categories

### Category 1: Strict Unit/Math Validation (6 tests)
**Priority:** HIGH - These are fundamental validation errors

Tests that should fail due to incompatible unit operations in strict mode:

1. **add-mixed-units** - Adding incompatible units (px + em)
   - File: `/home/user/less.go/packages/test-data/errors/eval/add-mixed-units.less`
   - Expected Error: Cannot add units px and em
   - Current Behavior: Compiles successfully

2. **add-mixed-units2** - Another mixed unit addition test
   - File: `/home/user/less.go/packages/test-data/errors/eval/add-mixed-units2.less`
   - Expected Error: Cannot add incompatible units
   - Current Behavior: Compiles successfully

3. **divide-mixed-units** - Dividing incompatible units
   - File: `/home/user/less.go/packages/test-data/errors/eval/divide-mixed-units.less`
   - Expected Error: Cannot divide incompatible units
   - Current Behavior: Compiles successfully

4. **multiply-mixed-units** - Multiplying incompatible units
   - File: `/home/user/less.go/packages/test-data/errors/eval/multiply-mixed-units.less`
   - Expected Error: Cannot multiply incompatible units
   - Current Behavior: Compiles successfully

5. **percentage-non-number-argument** - percentage() function with division in strict math
   - File: `/home/user/less.go/packages/test-data/errors/eval/percentage-non-number-argument.less`
   - Content: `percentage(16/17)` in strict math mode
   - Expected Error: percentage() expects a number, not an operation
   - Current Behavior: Compiles successfully

6. **unit-function** - unit() function with division in strict math
   - File: `/home/user/less.go/packages/test-data/errors/eval/unit-function.less`
   - Content: `unit(80/16, rem)` in strict math mode
   - Expected Error: unit() expects a number, not an operation
   - Current Behavior: Compiles successfully

### Category 2: Variable/Reference Errors (5 tests)
**Priority:** HIGH - These are fundamental scoping/reference errors

1. **recursive-variable** - Self-referencing variable
   - File: `/home/user/less.go/packages/test-data/errors/eval/recursive-variable.less`
   - Content: `@bodyColor: darken(@bodyColor, 30%);`
   - Expected Error: Recursive variable reference detected
   - Current Behavior: Compiles successfully (likely infinite loop or wrong value)

2. **property-interp-not-defined** - Undefined variable in interpolation
   - File: `/home/user/less.go/packages/test-data/errors/eval/property-interp-not-defined.less`
   - Content: `outline-@{color}: green` where @color is undefined
   - Expected Error: Variable @color is undefined
   - Current Behavior: Compiles successfully

3. **namespacing-2** - Accessing non-existent property in detached ruleset
   - File: `/home/user/less.go/packages/test-data/errors/eval/namespacing-2.less`
   - Content: `@dr[not-found]` where property doesn't exist
   - Expected Error: Property 'not-found' not found in detached ruleset
   - Current Behavior: Compiles successfully

4. **namespacing-3** - Another namespacing access error
   - File: `/home/user/less.go/packages/test-data/errors/eval/namespacing-3.less`
   - Expected Error: Invalid namespacing access
   - Current Behavior: Compiles successfully

5. **namespacing-4** - Another namespacing access error
   - File: `/home/user/less.go/packages/test-data/errors/eval/namespacing-4.less`
   - Expected Error: Invalid namespacing access
   - Current Behavior: Compiles successfully

### Category 3: Invalid Function Calls (13 tests)
**Priority:** MEDIUM - Function validation errors

1. **root-func-undefined-1** - Calling undefined function at root level
   - File: `/home/user/less.go/packages/test-data/errors/eval/root-func-undefined-1.less`
   - Content: `func();` where func is undefined
   - Expected Error: Function 'func' is undefined
   - Current Behavior: Compiles successfully

2. **color-func-invalid-color** - color() with invalid color string
   - File: `/home/user/less.go/packages/test-data/errors/eval/color-func-invalid-color.less`
   - Content: `color("NOT A COLOR")`
   - Expected Error: Invalid color string
   - Current Behavior: Compiles successfully

3. **color-func-invalid-color-2** - Another invalid color test
   - File: `/home/user/less.go/packages/test-data/errors/eval/color-func-invalid-color-2.less`
   - Expected Error: Invalid color string
   - Current Behavior: Compiles successfully

4. **svg-gradient1** through **svg-gradient6** - Invalid svg-gradient calls (6 tests)
   - Files: `/home/user/less.go/packages/test-data/errors/eval/svg-gradient[1-6].less`
   - Content: `svg-gradient(horizontal, black, white)` and similar
   - Expected Error: Invalid svg-gradient syntax (requires proper direction format)
   - Current Behavior: Compiles successfully
   - Note: These all have similar issues with svg-gradient validation

### Category 4: Type Errors (2 tests)
**Priority:** MEDIUM - Type checking errors

1. **detached-ruleset-1** - Using detached ruleset as property value
   - File: `/home/user/less.go/packages/test-data/errors/eval/detached-ruleset-1.less`
   - Content: `a: @a;` where @a is a detached ruleset
   - Expected Error: Cannot use detached ruleset as property value
   - Current Behavior: Compiles successfully

2. **detached-ruleset-2** - Another detached ruleset type error
   - File: `/home/user/less.go/packages/test-data/errors/eval/detached-ruleset-2.less`
   - Expected Error: Invalid detached ruleset usage
   - Current Behavior: Compiles successfully

### Category 5: Parse Errors (4 tests)
**Priority:** MEDIUM - Parser validation

1. **invalid-color-with-comment** - Invalid color syntax
   - File: `/home/user/less.go/packages/test-data/errors/parse/invalid-color-with-comment.less`
   - Content: `color: #fffff /* comment */;` (5-digit hex, invalid)
   - Expected Error: Invalid hex color (must be 3, 4, 6, or 8 digits)
   - Current Behavior: Compiles successfully

2. **parens-error-1** - Missing operator between expressions
   - File: `/home/user/less.go/packages/test-data/errors/parse/parens-error-1.less`
   - Content: `(12 (13 + 5 -23) + 5)` - missing operator between 12 and (13...)
   - Expected Error: Expected operator between expressions
   - Current Behavior: Compiles successfully

3. **parens-error-2** - Parse error in parenthesized expression
   - File: `/home/user/less.go/packages/test-data/errors/parse/parens-error-2.less`
   - Content: `(12 * (13 + 5 -23))`
   - Expected Error: Parse error (possibly related to -23 parsing)
   - Current Behavior: Compiles successfully

4. **parens-error-3** - Parse error in parenthesized expression
   - File: `/home/user/less.go/packages/test-data/errors/parse/parens-error-3.less`
   - Content: `(12 + (13 + 10 -23))`
   - Expected Error: Parse error (possibly related to -23 parsing)
   - Current Behavior: Compiles successfully

### Category 6: JavaScript-Related (1 test)
**Priority:** LOW - JavaScript feature (quarantined feature)

1. **javascript-undefined-var** - Undefined variable in JavaScript context
   - File: `/home/user/less.go/packages/test-data/errors/eval/javascript-undefined-var.less`
   - Expected Error: JavaScript variable undefined
   - Current Behavior: Compiles successfully
   - Note: May be quarantined as JavaScript feature

---

## Quick Reference Commands

### Run all error tests
```bash
# Get summary of error test status
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep -A 50 "EXPECTED ERROR BUT SUCCEEDED"

# Run eval-errors suite
go test -v -run TestIntegrationSuite/eval-errors

# Run parse-errors suite
go test -v -run TestIntegrationSuite/parse-errors
```

### Debug individual test
```bash
# See detailed output for a specific test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/eval-errors/add-mixed-units

# See what it's currently outputting (should be error instead)
LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/eval-errors/add-mixed-units
```

### Verify fixes
```bash
# After fixing, verify no regressions
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

# Check that expected error count decreased
# Should show fewer tests in "EXPECTED ERROR BUT SUCCEEDED" section
```

---

## Notes for Implementation

### General Strategy
1. **Find the JavaScript implementation first** - See what error message less.js throws
2. **Locate the Go equivalent** - Find where the same logic exists in less.go
3. **Add validation** - Implement the missing error checks
4. **Test incrementally** - Fix one test at a time, verify no regressions
5. **Use proper error types** - Match the JavaScript error structure

### Error Handling Pattern
In less.js, errors are typically thrown as Error objects with descriptive messages. In less.go, you should:
- Return errors from eval functions
- Include file location info when possible
- Match JavaScript error message format for consistency

### Testing Pattern
```bash
# Before fixing - should show "Expected error but compilation succeeded"
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/eval-errors/test-name

# After fixing - should show "Correctly failed with error: [message]"
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/eval-errors/test-name
```

### Common Pitfalls
- Don't just add errors without checking JavaScript behavior
- Some errors might be in evaluation, not parsing
- Some might need option checking (strictUnits, strictMath)
- Test both the error case AND that valid cases still work

---

## Test Dependencies

Most of these tests are independent and can be fixed in parallel. However, note:
- The parens-error tests (parse-errors) might share common root cause
- The namespacing-* tests might share validation code
- The svg-gradient tests are all similar

Consider grouping related tests when assigning to agents for efficiency.
