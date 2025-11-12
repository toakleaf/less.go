# Expected Error Tests - Tests That Should Fail But Don't

**Last Updated:** 2025-11-12
**Total Count:** 16 tests (8.7% of active tests) ⬇️ **-11 from previous count!**
**Priority:** MEDIUM (error handling is important but doesn't affect successful compilations)

## Overview

These are tests in the `eval-errors` and `parse-errors` suites that are expected to fail with an error, but currently compile successfully. This indicates missing validation and error handling in the Go port.

## Current Status

**🎉 MAJOR PROGRESS!** Recent validation improvements (commits 240-246) fixed 11 error handling tests!

From the latest test run (2025-11-12):

- **eval-errors suite:** 12 tests failing to error properly (down from 23!)
- **parse-errors suite:** 4 tests failing to error properly (unchanged)

**Total:** 16 tests that should error but succeed (down from 27!)

## Test Categories

### Category 1: Strict Unit/Math Validation (2 tests remaining, 4 FIXED! ✅)
**Priority:** HIGH - These are fundamental validation errors
**Progress:** 4/6 tests now correctly error!

**✅ FIXED (now correctly error):**
- ~~add-mixed-units~~ - FIXED in commit f8b76cb
- ~~add-mixed-units2~~ - FIXED in commit f8b76cb
- ~~divide-mixed-units~~ - FIXED in commit f8b76cb
- ~~multiply-mixed-units~~ - FIXED in commit f8b76cb

**Still need fixing:**

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

### Category 2: Variable/Reference Errors (0 tests remaining, 5 FIXED! ✅)
**Priority:** HIGH - These are fundamental scoping/reference errors
**Progress:** 5/5 tests now correctly error! **CATEGORY COMPLETE!**

**✅ ALL FIXED (now correctly error):**
- ~~recursive-variable~~ - FIXED in commit e0e371b
- ~~property-interp-not-defined~~ - FIXED in commit e4bb4ce
- ~~namespacing-2~~ - FIXED in commit cb3055b
- ~~namespacing-3~~ - FIXED in commit cb3055b
- ~~namespacing-4~~ - FIXED in commit cb3055b

### Category 3: Invalid Function Calls (8 tests remaining, 5 FIXED! ✅)
**Priority:** MEDIUM - Function validation errors
**Progress:** 5/13 tests now correctly error!

**✅ FIXED (now correctly error):**
- ~~root-func-undefined-1~~ - FIXED in commit 6d9afec
- ~~root-func-undefined-2~~ - FIXED in commit 6d9afec
- ~~at-rules-undefined-var~~ - FIXED in commits 240-246
- ~~css-guard-default-func~~ - FIXED in commits 240-246
- ~~javascript-undefined-var~~ - FIXED in commits 240-246

**Still need fixing:**

1. **color-func-invalid-color** - color() with invalid color string
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
