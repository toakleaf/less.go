# Error Validation Tasks - 27 Tests That Should Fail But Don't

**Status**: 27 tests are currently succeeding when they should be throwing errors
**Priority**: LOW (documented in CLAUDE.md section 8)
**Last Updated**: 2025-11-11

## Overview

These tests validate that the Go port correctly handles invalid LESS code by throwing appropriate errors. Currently, these tests compile and produce output when they should fail with specific error messages.

## Test Categories & Individual Tasks

### Category 1: Unit Validation Errors (4 tests)

**Tests**: `add-mixed-units`, `add-mixed-units2`, `divide-mixed-units`, `multiply-mixed-units`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw `SyntaxError` for incompatible unit operations

**Example**:
- Input: `.a { error: (1px + 3em); }`
- Expected Error: `SyntaxError: Incompatible units. Change the units or use the unit function. Bad units: 'px' and 'em'.`

**Task Prompt for LLM**:
```
TASK: Fix Unit Validation Error Handling in LESS Go Port

BACKGROUND:
The Go port of less.js is currently allowing mathematical operations between incompatible units (e.g., px + em) when it should throw a SyntaxError.

TESTS TO FIX (4 tests):
- eval-errors/add-mixed-units
- eval-errors/add-mixed-units2
- eval-errors/divide-mixed-units
- eval-errors/multiply-mixed-units

CURRENT BEHAVIOR:
These tests are compiling and producing CSS output when they should fail.

EXPECTED BEHAVIOR:
Should throw: "SyntaxError: Incompatible units. Change the units or use the unit function. Bad units: 'X' and 'Y'."

STEPS:
1. Run the tests to see current behavior:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/(add-mixed-units|divide-mixed-units|multiply-mixed-units)"

2. Find where unit operations are evaluated in the Go codebase
   - Look for arithmetic operations on Dimension types
   - Search for "operate" or "doOperation" methods

3. Compare with JavaScript implementation:
   - Check packages/less/src/less-node/dimension.js or similar
   - Look for unit compatibility validation

4. Add unit validation that throws errors for incompatible units

5. Verify all 4 tests now fail with the correct error message

6. Run full test suite to ensure no regressions:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- All 4 unit validation tests should now appear in "✅ CORRECTLY FAILED" category
- No decrease in "Perfect CSS Matches" count
- All unit tests still passing
```

---

### Category 2: SVG Gradient Function Validation (6 tests)

**Tests**: `svg-gradient1`, `svg-gradient2`, `svg-gradient3`, `svg-gradient4`, `svg-gradient5`, `svg-gradient6`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw `ArgumentError` for invalid svg-gradient parameters

**Example**:
- Input: `.a { a: svg-gradient(horizontal, black, white); }`
- Expected Error: `ArgumentError: Error evaluating function 'svg-gradient': svg-gradient direction must be 'to bottom', 'to right', 'to bottom right', 'to top right' or 'ellipse at center'`

**Task Prompt for LLM**:
```
TASK: Fix SVG Gradient Function Validation in LESS Go Port

BACKGROUND:
The svg-gradient() function should validate its direction parameter and throw an ArgumentError for invalid values.

TESTS TO FIX (6 tests):
- eval-errors/svg-gradient1
- eval-errors/svg-gradient2
- eval-errors/svg-gradient3
- eval-errors/svg-gradient4
- eval-errors/svg-gradient5
- eval-errors/svg-gradient6

CURRENT BEHAVIOR:
These tests compile successfully when they should fail with argument validation errors.

EXPECTED BEHAVIOR:
Should throw: "ArgumentError: Error evaluating function `svg-gradient`: svg-gradient direction must be 'to bottom', 'to right', 'to bottom right', 'to top right' or 'ellipse at center'"

STEPS:
1. Review the test files to understand each error case:
   ls packages/test-data/errors/eval/svg-gradient*.{less,txt}

2. Run tests to see current behavior:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/svg-gradient"

3. Find the svg-gradient function implementation:
   - Search for "svg-gradient" or "svgGradient" in the Go codebase
   - Look in function registry or built-in functions

4. Compare with JavaScript implementation:
   - Check packages/less/src/less/functions/svg.js or similar
   - Identify validation logic

5. Add parameter validation that throws ArgumentError for invalid directions

6. Verify all 6 tests now fail with correct error messages

7. Run full test suite:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- All 6 svg-gradient tests should appear in "✅ CORRECTLY FAILED" category
- No regressions in passing tests
```

---

### Category 3: Color Function Validation (2 tests)

**Tests**: `color-func-invalid-color`, `color-func-invalid-color-2`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw `ArgumentError` for invalid color arguments

**Example**:
- Input: `.test-rule { color: color("NOT A COLOR"); }`
- Expected Error: `ArgumentError: Error evaluating function 'color': argument must be a color keyword or 3|4|6|8 digit hex e.g. #FFF`

**Task Prompt for LLM**:
```
TASK: Fix Color Function Validation in LESS Go Port

BACKGROUND:
The color() function should validate its argument is a valid color and throw an ArgumentError for invalid values.

TESTS TO FIX (2 tests):
- eval-errors/color-func-invalid-color
- eval-errors/color-func-invalid-color-2

CURRENT BEHAVIOR:
These tests compile when they should fail with color validation errors.

EXPECTED BEHAVIOR:
Should throw: "ArgumentError: Error evaluating function `color`: argument must be a color keyword or 3|4|6|8 digit hex e.g. #FFF"

STEPS:
1. Examine test files:
   cat packages/test-data/errors/eval/color-func-invalid-color.less
   cat packages/test-data/errors/eval/color-func-invalid-color.txt
   cat packages/test-data/errors/eval/color-func-invalid-color-2.less
   cat packages/test-data/errors/eval/color-func-invalid-color-2.txt

2. Run tests:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/color-func-invalid-color"

3. Find color() function implementation in Go codebase

4. Compare with JavaScript implementation to see validation logic

5. Add validation that throws ArgumentError for invalid color strings

6. Test and validate:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- Both color validation tests in "✅ CORRECTLY FAILED" category
- No regressions
```

---

### Category 4: Percentage Function Validation (1 test)

**Tests**: `percentage-non-number-argument`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw `ArgumentError` when percentage() receives a non-number

**Example**:
- Input: `div { percentage: percentage(16/17); }`
- Expected Error: `ArgumentError: Error evaluating function 'percentage': argument must be a number`

**Task Prompt for LLM**:
```
TASK: Fix Percentage Function Validation in LESS Go Port

BACKGROUND:
The percentage() function should validate its argument is a number (not a division operation) and throw an ArgumentError otherwise.

TESTS TO FIX (1 test):
- eval-errors/percentage-non-number-argument

CURRENT BEHAVIOR:
Test compiles when it should fail (16/17 is treated as an operation, not a number).

EXPECTED BEHAVIOR:
Should throw: "ArgumentError: Error evaluating function `percentage`: argument must be a number"

STEPS:
1. Review test:
   cat packages/test-data/errors/eval/percentage-non-number-argument.less
   cat packages/test-data/errors/eval/percentage-non-number-argument.txt

2. Run test:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/percentage-non-number-argument"

3. Find percentage() function implementation

4. Add type validation before processing the argument

5. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- Test appears in "✅ CORRECTLY FAILED" category
```

---

### Category 5: Unit Function Validation (1 test)

**Tests**: `unit-function`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw an error for invalid unit() function usage

**Task Prompt for LLM**:
```
TASK: Fix Unit Function Validation in LESS Go Port

BACKGROUND:
The unit() function has specific parameter requirements that should be validated.

TESTS TO FIX (1 test):
- eval-errors/unit-function

STEPS:
1. Review test files to understand expected error:
   cat packages/test-data/errors/eval/unit-function.less
   cat packages/test-data/errors/eval/unit-function.txt

2. Run test:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/unit-function"

3. Find unit() function implementation in Go

4. Compare with JavaScript to identify validation requirements

5. Add appropriate validation and error throwing

6. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go
```

---

### Category 6: Undefined Variable Errors (2 tests)

**Tests**: `property-interp-not-defined`, `root-func-undefined-1`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw `NameError` for undefined variables

**Example**:
- Input: `a {outline-@{color}: green}` (where @color is undefined)
- Expected Error: `NameError: variable @color is undefined`

**Task Prompt for LLM**:
```
TASK: Fix Undefined Variable Detection in LESS Go Port

BACKGROUND:
The Go port should detect and report undefined variables in property name interpolation and function contexts.

TESTS TO FIX (2 tests):
- eval-errors/property-interp-not-defined
- eval-errors/root-func-undefined-1

CURRENT BEHAVIOR:
Code compiles with undefined variables when it should throw NameError.

EXPECTED BEHAVIOR:
Should throw: "NameError: variable @X is undefined in {path} on line Y, column Z"

STEPS:
1. Review test cases:
   cat packages/test-data/errors/eval/property-interp-not-defined.{less,txt}
   cat packages/test-data/errors/eval/root-func-undefined-1.{less,txt}

2. Run tests:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/(property-interp-not-defined|root-func-undefined-1)"

3. Find where variable interpolation and function evaluation occur

4. Add undefined variable detection that throws NameError

5. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- Both tests in "✅ CORRECTLY FAILED" category
- No regressions
```

---

### Category 7: Recursive Variable Detection (1 test)

**Tests**: `recursive-variable`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw `NameError` for recursive variable definitions

**Example**:
- Input: `@bodyColor: darken(@bodyColor, 30%);`
- Expected Error: `NameError: Error evaluating function 'darken': Recursive variable definition for @bodyColor`

**Task Prompt for LLM**:
```
TASK: Fix Recursive Variable Detection in LESS Go Port

BACKGROUND:
The Go port should detect when a variable references itself in its own definition.

TESTS TO FIX (1 test):
- eval-errors/recursive-variable

CURRENT BEHAVIOR:
Recursive variable definitions compile when they should throw NameError.

EXPECTED BEHAVIOR:
Should throw: "NameError: Error evaluating function `darken`: Recursive variable definition for @bodyColor"

STEPS:
1. Review test:
   cat packages/test-data/errors/eval/recursive-variable.less
   cat packages/test-data/errors/eval/recursive-variable.txt

2. Run test:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/recursive-variable"

3. Research how JavaScript detects recursive variables:
   - Look for cycle detection in variable evaluation
   - Check how evaluation stack is tracked

4. Implement cycle detection in variable evaluation

5. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go
```

---

### Category 8: JavaScript Variable Errors (1 test)

**Tests**: `javascript-undefined-var`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw error for undefined variables in JavaScript context

**Task Prompt for LLM**:
```
TASK: Fix JavaScript Undefined Variable Detection

BACKGROUND:
When JavaScript evaluation is enabled, undefined variables in JS context should throw errors.

TESTS TO FIX (1 test):
- eval-errors/javascript-undefined-var

NOTE: This may be related to the quarantined JavaScript features. Check if this should be handled differently.

STEPS:
1. Review test:
   cat packages/test-data/errors/eval/javascript-undefined-var.{less,txt}

2. Run test:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/javascript-undefined-var"

3. Determine if this is a JavaScript execution error or LESS variable error

4. Implement appropriate validation

5. Validate:
   pnpm -w test:go
```

---

### Category 9: Namespacing Errors (3 tests)

**Tests**: `namespacing-2`, `namespacing-3`, `namespacing-4`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw errors for invalid namespacing operations

**Task Prompt for LLM**:
```
TASK: Fix Namespacing Error Detection in LESS Go Port

BACKGROUND:
These tests validate error handling in namespacing scenarios (nested mixins/variables).

TESTS TO FIX (3 tests):
- eval-errors/namespacing-2
- eval-errors/namespacing-3
- eval-errors/namespacing-4

NOTE: The regular namespacing tests (namespacing-1 through namespacing-operations) all pass perfectly. These error tests validate that INVALID namespacing operations properly fail.

STEPS:
1. Review what errors each test expects:
   for f in namespacing-{2,3,4}; do
     echo "=== $f ==="
     cat packages/test-data/errors/eval/$f.less
     cat packages/test-data/errors/eval/$f.txt
   done

2. Run tests:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/namespacing-[234]"

3. Compare with passing namespacing tests to understand the difference

4. Implement validation for invalid namespacing cases

5. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- All 3 tests in "✅ CORRECTLY FAILED" category
- All 11 passing namespacing tests still pass (NO REGRESSIONS!)
```

---

### Category 10: Detached Ruleset Errors (2 tests)

**Tests**: `detached-ruleset-1`, `detached-ruleset-2`

**Location**: `/home/user/less.go/packages/test-data/errors/eval/`

**Error Type**: Should throw errors for invalid detached ruleset operations

**Task Prompt for LLM**:
```
TASK: Fix Detached Ruleset Error Detection in LESS Go Port

BACKGROUND:
Detached rulesets have specific usage constraints that should be validated.

TESTS TO FIX (2 tests):
- eval-errors/detached-ruleset-1
- eval-errors/detached-ruleset-2

NOTE: There are also eval-errors/detached-ruleset-3 and detached-ruleset-5 that correctly fail, and main/detached-rulesets that has output differences. Don't confuse them.

STEPS:
1. Review test expectations:
   cat packages/test-data/errors/eval/detached-ruleset-1.{less,txt}
   cat packages/test-data/errors/eval/detached-ruleset-2.{less,txt}

2. Run tests:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/detached-ruleset-[12]"

3. Find detached ruleset implementation

4. Compare with JavaScript to identify validation rules

5. Add appropriate error handling

6. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- Both tests in "✅ CORRECTLY FAILED" category
- detached-ruleset-3 and -5 still correctly fail
- No regressions
```

---

### Category 11: Parse Errors (4 tests)

**Tests**: `invalid-color-with-comment`, `parens-error-1`, `parens-error-2`, `parens-error-3`

**Location**: `/home/user/less.go/packages/test-data/errors/parse/`

**Error Type**: Should throw `ParseError` for invalid syntax

**Example**:
- Input: `.a { something: (12 (13 + 5 -23) + 5); }`
- Expected Error: `ParseError: Expected ')' in {path}parens-error-1.less on line 2, column 18`

**Task Prompt for LLM**:
```
TASK: Fix Parse Error Detection in LESS Go Port

BACKGROUND:
The parser should reject invalid syntax and throw ParseError.

TESTS TO FIX (4 tests):
- parse-errors/invalid-color-with-comment
- parse-errors/parens-error-1
- parse-errors/parens-error-2
- parse-errors/parens-error-3

CURRENT BEHAVIOR:
Invalid syntax is being parsed successfully when it should fail.

EXPECTED BEHAVIOR:
Should throw: "ParseError: Expected '...' in {path} on line X, column Y"

STEPS:
1. Review each test case:
   for f in invalid-color-with-comment parens-error-{1,2,3}; do
     echo "=== $f ==="
     cat packages/test-data/errors/parse/$f.less
     cat packages/test-data/errors/parse/$f.txt
   done

2. Run tests:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/parse-errors/(invalid-color|parens-error)"

3. Identify where parser is too permissive:
   - For parens-error: Check parenthesis parsing/validation
   - For invalid-color-with-comment: Check color parsing with comments

4. Add stricter validation to throw ParseError

5. Validate:
   pnpm -w test:go:unit
   pnpm -w test:go

VALIDATION:
- All 4 tests in "✅ CORRECTLY FAILED" category
- All 23 correctly failing parse-errors still fail
- Parser still correctly handles all valid syntax (NO REGRESSIONS!)
```

---

## Summary of Tasks by Priority

### High-Impact (Most Common Error Types)
1. **Unit Validation** (4 tests) - Core arithmetic operations
2. **SVG Gradients** (6 tests) - Function validation pattern

### Medium-Impact (Function Validation)
3. **Color Function** (2 tests)
4. **Percentage Function** (1 test)
5. **Unit Function** (1 test)

### Medium-Impact (Variable Handling)
6. **Undefined Variables** (2 tests)
7. **Recursive Variables** (1 test)
8. **Namespacing Errors** (3 tests)

### Lower-Impact (Specific Cases)
9. **Detached Rulesets** (2 tests)
10. **JavaScript Variables** (1 test)
11. **Parse Errors** (4 tests)

## Testing Commands

**Run all error tests:**
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/(eval-errors|parse-errors)"
```

**Run specific category:**
```bash
# Unit validation
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/(add-mixed|divide-mixed|multiply-mixed)"

# SVG gradients
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/svg-gradient"

# Namespacing
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/namespacing-[234]"
```

**Check overall progress:**
```bash
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

## Success Criteria

For each task:
1. ✅ Test moves from "⚠️ EXPECTED ERROR BUT SUCCEEDED" to "✅ CORRECTLY FAILED"
2. ✅ Error message matches expected format in .txt file
3. ✅ No regressions in "Perfect CSS Matches" count
4. ✅ All unit tests still pass

Overall success:
- All 27 tests in "✅ CORRECTLY FAILED" category
- "⚠️ EXPECTED ERROR BUT SUCCEEDED" count = 0
- No decrease in perfect match count (currently 80)
