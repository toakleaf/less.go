# Parallel Task Assignments - Error Validation (27 Tests)

**Quick Reference for Assigning Tasks to Independent LLM Sessions**

## Ready-to-Copy Task Prompts

Each section below is a complete, standalone prompt you can copy and paste into an independent LLM session.

---

## Task 1: Unit Validation Errors (4 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix Unit Validation Error Handling in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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
- No decrease in "Perfect CSS Matches" count (should stay at 80)
- All unit tests still passing

DELIVERABLE:
- Commit your changes with message: "Fix unit validation error handling (4 tests)"
- Push to the branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 2: SVG Gradient Validation (6 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix SVG Gradient Function Validation in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix svg-gradient function validation (6 tests)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 3: Color Function Validation (2 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix Color Function Validation in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix color function validation (2 tests)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 4: Percentage Function Validation (1 test)

**Copy this entire section to an LLM:**

```
TASK: Fix Percentage Function Validation in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix percentage function validation (1 test)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 5: Unit Function Validation (1 test)

**Copy this entire section to an LLM:**

```
TASK: Fix Unit Function Validation in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix unit function validation (1 test)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 6: Undefined Variable Detection (2 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix Undefined Variable Detection in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix undefined variable detection (2 tests)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 7: Recursive Variable Detection (1 test)

**Copy this entire section to an LLM:**

```
TASK: Fix Recursive Variable Detection in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix recursive variable detection (1 test)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 8: Namespacing Error Detection (3 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix Namespacing Error Detection in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix namespacing error detection (3 tests)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 9: Detached Ruleset Error Detection (2 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix Detached Ruleset Error Detection in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix detached ruleset error detection (2 tests)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 10: JavaScript Variable Error Detection (1 test)

**Copy this entire section to an LLM:**

```
TASK: Fix JavaScript Undefined Variable Detection

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix JavaScript undefined variable detection (1 test)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Task 11: Parse Error Detection (4 tests)

**Copy this entire section to an LLM:**

```
TASK: Fix Parse Error Detection in LESS Go Port

REPOSITORY: https://github.com/toakleaf/less.go
BRANCH: claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X

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

DELIVERABLE:
- Commit: "Fix parse error detection (4 tests)"
- Push to branch
```

**Status**: [ ] Not Started [ ] In Progress [ ] Complete

---

## Coordination Notes

**For the User:**
1. Each task above can be assigned to a separate LLM session
2. Tasks are independent and can be worked on in parallel
3. All work should be committed and pushed to: `claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X`
4. Track progress by checking boxes in the "Status" line under each task

**Recommended Assignment Order** (if limited parallelism):
1. **High Priority**: Tasks 1, 2 (Unit validation, SVG gradients - 10 tests total)
2. **Medium Priority**: Tasks 3, 4, 5, 6, 7, 8 (Function validation, variables - 11 tests)
3. **Lower Priority**: Tasks 9, 10, 11 (Specific cases - 7 tests)

**Testing After Completion:**
```bash
# Check overall status
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

# Verify "EXPECTED ERROR BUT SUCCEEDED" count
# Should go from 27 → 0 as tasks complete

# Verify no regressions
# "Perfect CSS Matches" should stay at 80
```
