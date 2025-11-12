# Expected Error Tests - Tests That Should Fail But Don't

**Last Updated:** 2025-11-12
**Total Count:** 10 tests (5.4% of active tests) - DOWN FROM 27! 🎉
**Priority:** LOW (error handling is important but most critical validations are now working)

## Overview

These are tests in the `eval-errors` and `parse-errors` suites that are expected to fail with an error, but currently compile successfully. This indicates missing validation and error handling in the Go port.

## 🎉 MAJOR PROGRESS: 17 Tests Fixed!

Since the last update, **17 error validation tests** have been fixed and now correctly fail with appropriate errors. This is a **63% reduction** in missing error handling!

## Current Status

From the latest test run (2025-11-12):

- **eval-errors suite:** 6 tests failing to error properly (down from 23!)
- **parse-errors suite:** 4 tests failing to error properly (same as before)

**Total:** 10 tests that should error but succeed (down from 27!)

## Tests Still Needing Fixes

### Category 1: Strict Unit/Math Validation (2 tests remaining)
**Priority:** MEDIUM - Basic validation

**FIXED (4 tests):** ✅
- add-mixed-units - NOW CORRECTLY VALIDATES
- add-mixed-units2 - NOW CORRECTLY VALIDATES
- divide-mixed-units - NOW CORRECTLY VALIDATES
- multiply-mixed-units - NOW CORRECTLY VALIDATES

**Still Need Fixing (2 tests):**

1. **percentage-non-number-argument** - percentage() function with division in strict math
   - File: `/home/user/less.go/packages/test-data/errors/eval/percentage-non-number-argument.less`
   - Content: `percentage(16/17)` in strict math mode
   - Expected Error: percentage() expects a number, not an operation
   - Current Behavior: Compiles successfully

2. **unit-function** - unit() function with division in strict math
   - File: `/home/user/less.go/packages/test-data/errors/eval/unit-function.less`
   - Content: `unit(80/16, rem)` in strict math mode
   - Expected Error: unit() expects a number, not an operation
   - Current Behavior: Compiles successfully

### Category 2: Variable/Reference Errors (1 test remaining)
**Priority:** MEDIUM - Scoping/reference validation

**FIXED (4 tests):** ✅
- recursive-variable - NOW CORRECTLY DETECTS RECURSION
- namespacing-2 - NOW CORRECTLY VALIDATES
- namespacing-3 - NOW CORRECTLY VALIDATES
- namespacing-4 - NOW CORRECTLY VALIDATES

**Still Need Fixing (1 test):**

1. **property-interp-not-defined** - Undefined variable in interpolation
   - File: `/home/user/less.go/packages/test-data/errors/eval/property-interp-not-defined.less`
   - Content: `outline-@{color}: green` where @color is undefined
   - Expected Error: Variable @color is undefined
   - Current Behavior: Compiles successfully

### Category 3: Invalid Function Calls (2 tests remaining)
**Priority:** MEDIUM - Function validation errors

**FIXED (7 tests):** ✅
- root-func-undefined-1 - NOW CORRECTLY VALIDATES
- svg-gradient1 - NOW CORRECTLY VALIDATES
- svg-gradient2 - NOW CORRECTLY VALIDATES
- svg-gradient3 - NOW CORRECTLY VALIDATES
- svg-gradient4 - NOW CORRECTLY VALIDATES
- svg-gradient5 - NOW CORRECTLY VALIDATES
- svg-gradient6 - NOW CORRECTLY VALIDATES

**Still Need Fixing (2 tests):**

1. **color-func-invalid-color** - color() with invalid color string
   - File: `/home/user/less.go/packages/test-data/errors/eval/color-func-invalid-color.less`
   - Content: `color("NOT A COLOR")`
   - Expected Error: Invalid color string
   - Current Behavior: Compiles successfully

2. **color-func-invalid-color-2** - Another invalid color test
   - File: `/home/user/less.go/packages/test-data/errors/eval/color-func-invalid-color-2.less`
   - Expected Error: Invalid color string
   - Current Behavior: Compiles successfully

### Category 4: Type Errors (0 tests remaining)
**Priority:** N/A - ALL FIXED! ✅

**FIXED (2 tests):** ✅
- detached-ruleset-1 - NOW CORRECTLY VALIDATES TYPE
- detached-ruleset-2 - NOW CORRECTLY VALIDATES TYPE

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
**Priority:** VERY LOW - JavaScript feature (should be quarantined)

1. **javascript-undefined-var** - Undefined variable in JavaScript context
   - File: `/home/user/less.go/packages/test-data/errors/eval/javascript-undefined-var.less`
   - Expected Error: JavaScript variable undefined
   - Current Behavior: Compiles successfully
   - **Note:** This test is related to JavaScript execution which is a quarantined feature. It should probably be moved to the quarantined tests list rather than treated as a bug.

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
