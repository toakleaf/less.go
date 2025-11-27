# Independent Agent Prompts - 2025-11-27

These 4 prompts are ready for independent agents to work on in parallel. Each agent should:
1. Use the full prompt below (copy the entire section)
2. Create a new agent with that prompt
3. Work independently to fix the issue
4. Run all tests to ensure no regressions before committing
5. Push to their assigned branch

**Baseline for regression detection:**
- Unit tests: 2,304 passing (100%)
- Integration tests: 89 perfect matches (48.4%)
- Output differences: 3 tests remaining
- Overall success: 96.7%

If any of these numbers decrease, STOP and investigate before committing.

---

## Prompt 1: Fix Import Reference Functionality (2 Tests)

```
You are fixing 2 failing tests in the less.go port: `import-reference` and `import-reference-issues`.

**Current Status:**
- These tests compile but CSS output doesn't match less.js
- Files imported with `@import (reference)` should not output CSS
- Their selectors/mixins should still be available for extends and mixin calls

**What to do:**
1. Run baseline tests:
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: 89 perfect matches

2. Look at the failing tests:
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
   # This shows what the test expects vs actual output

3. Check the test input file:
   cat packages/test-data/less/_main/import-reference.less
   cat packages/test-data/css/_main/import-reference.css

4. Look for where the `reference` flag is being handled:
   - packages/less/src/less/less_go/import.go - Import node
   - packages/less/src/less/less_go/import_visitor.go - Import processing
   - packages/less/src/less/less_go/ruleset.go - CSS generation (check if it respects reference flag)

5. Compare with JavaScript implementation:
   cat packages/less/src/less/tree/import.js | grep -A 10 "reference"
   cat packages/less/src/less/tree/ruleset.js | grep -A 10 "isReferenced\|reference"

6. The likely issue: When a ruleset is imported with (reference), it should:
   - Mark the ruleset as being from a referenced import
   - Skip CSS output when generating
   - But still allow the selectors to be referenced/extended

7. Fix the Go code to handle this correctly

8. Run tests after fix:
   pnpm -w test:go:unit  # Must pass 100%
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: >= 89 perfect matches (ideally 91!)

9. If tests pass:
   git add -A
   git commit -m "Fix import reference functionality - reference imports no longer output CSS"
   git push -u origin claude/fix-import-reference-SESSION_ID

**Validation:**
✅ Both import-reference tests show "Perfect match!"
✅ All 2,304 unit tests still pass
✅ Perfect match count >= 89 (no regressions)
✅ No new failures in other import tests
```

---

## Prompt 2: Fix URL Handling Output (1 Test)

```
You are fixing 1 failing test in the less.go port: `urls` (in the main suite).

**Current Status:**
- Test compiles but CSS output doesn't match less.js exactly
- Likely issue: URL encoding, path handling, or formatting differences

**What to do:**
1. Run baseline tests:
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: 89 perfect matches

2. Look at the failing test in detail:
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/urls"
   # This shows side-by-side diff of what's expected vs actual

3. Check the test input file:
   cat packages/test-data/less/_main/urls.less
   cat packages/test-data/css/_main/urls.css

4. Look for URL handling code:
   - packages/less/src/less/less_go/url.go - URL processing
   - packages/less/src/less/less_go/ruleset.go - CSS generation that uses URLs
   - packages/less/src/less/less_go/declaration.go - Declaration processing

5. Common issues to check:
   - URL encoding/escaping differences
   - Path normalization differences
   - Quote handling around URLs
   - Relative vs absolute path handling

6. Compare with JavaScript:
   cat packages/less/src/less/tree/url.js | head -50
   Look at how JavaScript handles URL generation

7. Fix the Go code to match JavaScript behavior exactly

8. Run tests after fix:
   pnpm -w test:go:unit  # Must pass 100%
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: >= 89 perfect matches (ideally 90!)

9. If tests pass:
   git add -A
   git commit -m "Fix URL handling output formatting"
   git push -u origin claude/fix-urls-output-SESSION_ID

**Validation:**
✅ urls test shows "Perfect match!"
✅ All 2,304 unit tests still pass
✅ Perfect match count >= 89 (no regressions)
✅ No new failures in other URL/path tests
```

---

## Prompt 3: Fix Color Function Error Validation (2 Tests)

```
You are fixing 2 failing tests in the less.go port: `color-func-invalid-color` and `color-func-invalid-color-2`.

**Current Status:**
- Both tests SHOULD fail with an error about invalid color strings
- Currently they compile successfully when they shouldn't
- These are in the eval-errors test suite

**What to do:**
1. Run baseline tests:
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: 89 correct error handling tests

2. Look at the failing tests:
   cat packages/test-data/errors/eval/color-func-invalid-color.less
   cat packages/test-data/errors/eval/color-func-invalid-color-2.less
   cat packages/test-data/errors/eval/color-func-invalid-color.error

3. Run the test to see what's happening:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/color-func-invalid-color"
   # Should show: "Expected error but compilation succeeded"

4. Find the color function implementation:
   - packages/less/src/less/less_go/functions/ - Look for color function
   - Check how it validates input strings
   - Compare with JavaScript: packages/less/src/less/functions/color.js

5. The color() function should:
   - Accept valid color strings like "#fff", "red", "rgb(255,0,0)"
   - REJECT invalid color strings like "NOT A COLOR"
   - Throw an error with message about invalid color

6. Add validation to reject invalid colors before compiling

7. Test your fix:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/color-func"
   # Should show: "Correctly failed with error: [message]"

8. Run full test suite:
   pnpm -w test:go:unit  # Must pass 100%
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: error count increased by 2 (90 or 91 total)

9. If tests pass:
   git add -A
   git commit -m "Add color function validation for invalid color strings"
   git push -u origin claude/fix-color-validation-SESSION_ID

**Validation:**
✅ Both color-func-invalid-color tests now correctly fail with errors
✅ All 2,304 unit tests still pass
✅ Error handling count increases by 2 (89 → 91)
✅ No new test failures
```

---

## Prompt 4: Fix Parenthesis Expression Parse Errors (3 Tests)

```
You are fixing 3 failing tests in the less.go port: `parens-error-1`, `parens-error-2`, `parens-error-3`.

**Current Status:**
- All 3 tests SHOULD fail with parse errors
- Currently they compile successfully when they shouldn't
- These are in the parse-errors test suite

**What to do:**
1. Run baseline tests:
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: 0 expected errors (all error tests working)

2. Look at the failing tests:
   cat packages/test-data/errors/parse/parens-error-1.less
   cat packages/test-data/errors/parse/parens-error-2.less
   cat packages/test-data/errors/parse/parens-error-3.less
   cat packages/test-data/errors/parse/parens-error-1.error

3. Run the tests to see what's happening:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/parse-errors/parens-error"
   # Should show: "Expected error but compilation succeeded"

4. The issues:
   - parens-error-1: Missing operator between expressions: (12 (13 + 5 -23) + 5)
   - parens-error-2 & 3: Possibly malformed negative number handling in expressions

5. Find the parser code for parenthesized expressions:
   - packages/less/src/less/less_go/parser/ - Expression parsing
   - Look for how it handles parentheses and operators
   - Compare with JavaScript: packages/less/src/less/parser/

6. The parser should:
   - Require explicit operators between adjacent expressions
   - Not allow implicit multiplication like (12 (13...))
   - Properly validate negative numbers in expressions

7. Add validation to catch these parse errors

8. Test your fix:
   LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/parse-errors/parens-error"
   # Should show: "Correctly failed with error: [message]"

9. Run full test suite:
   pnpm -w test:go:unit  # Must pass 100%
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
   # Should show: error count increased by 3

10. If tests pass:
    git add -A
    git commit -m "Add parse error validation for malformed parenthesized expressions"
    git push -u origin claude/fix-parens-parse-errors-SESSION_ID

**Validation:**
✅ All 3 parens-error tests now correctly fail with parse errors
✅ All 2,304 unit tests still pass
✅ No new test failures
✅ Error handling count increases by 3
```

---

## How to Use These Prompts

1. **Pick a prompt** above that appeals to you (Prompt 1 is highest impact)
2. **Copy the entire prompt section** (starting from the triple backticks)
3. **Create a new agent** with that prompt in Claude Code
4. **Agent works independently** to fix the issue
5. **Agent runs tests** to verify no regressions
6. **Agent commits and pushes** to the assigned branch
7. **Report back** with whether it succeeded

All 4 prompts are independent - multiple agents can work on them in parallel!

---

## Test Baseline Reference

Before starting, baseline should be:
```
OVERALL SUCCESS: 178/184 tests (96.7%)
✅ Perfect CSS Matches: 89 (48.4%)
❌ Compilation Failures: 3 (1.6%) - all expected
⚠️ Output Differences: 3 (1.6%)
✅ Correctly Failed: 89 (48.4%)
```

After all 4 prompts complete:
```
Expected: 92-93 perfect matches
Expected: 0-3 output differences
Expected: 92-95 error tests
Expected: >97% overall success
```

---

Generated: 2025-11-27
Ready for parallel agent work
