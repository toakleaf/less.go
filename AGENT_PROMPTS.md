# Agent Prompts - Next Priority Issues
**Generated**: 2025-11-08
**Session**: claude/assess-less-go-port-progress-011CUwJ5jjGHkMNZFvH8oNmz

## Instructions
Each prompt below is ready for a new agent to pick up and work on independently. Before starting, agents should:
1. Run `pnpm -w test:go` to establish baseline (currently 57 perfect matches, 0 regressions)
2. Run `pnpm -w test:go:unit` to verify all unit tests pass
3. After completing work, re-run both test suites to verify no regressions
4. Ensure perfect match count increases (or at minimum, no regressions)

---

## Prompt 1: Fix extend-chaining Test (QUICK WIN - Complete Extend Category!)

**Goal**: Fix the `extend-chaining` test to complete the extend category (7/7 tests passing).

**Current status**: The test compiles successfully but CSS output doesn't match. This is the LAST extend test - fixing it completes the entire extend category!

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "extend-chaining"` to see differences
2. Study the test input at `packages/test-data/less/_main/extend-chaining.less`
3. The issue is multi-level extend chains (A extends B, B extends C → A should also extend C)
4. Compare with JavaScript implementation in `packages/less/src/less/visitors/extend-visitor.js`
5. Fix the chain resolution logic in Go's `packages/less/src/less/less_go/extend_visitor.go`
6. Verify: `pnpm -w test:go:filter -- "extend"` shows all 7 extend tests passing
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: extend-chaining shows "Perfect match!", completing extend category to 7/7 (100%)

**See also**: `.claude/tasks/output-differences/extend-functionality.md`

---

## Prompt 2: Fix Math Operations in Parens Mode (6 Tests)

**Goal**: Fix math operation handling in math-parens mode to get 6 tests passing.

**Current status**: Tests compile but output differs. Math operations in parentheses aren't being evaluated correctly in different math modes.

**Affected tests**:
- css (math-parens suite)
- mixins-args (math-parens suite)
- parens (math-parens suite)
- mixins-args (math-parens-division suite)
- parens (math-parens-division suite)
- no-strict (units-no-strict suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "math-parens/css"` to see first failure
2. Study how JavaScript handles math modes in `packages/less/src/less/tree/operation.js`
3. Check Go's math mode handling in `packages/less/src/less/less_go/operation.go` and `contexts.go`
4. The issue is likely: operations in parentheses should evaluate differently based on math mode
5. Fix the evaluation logic to match JavaScript behavior
6. Test all affected suites: `pnpm -w test:go:filter -- "math-parens" && pnpm -w test:go:filter -- "math-parens-division"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +6 perfect matches, bringing total to 63/184 (34.2%)

---

## Prompt 3: Fix URL Output Issues (3 Tests)

**Goal**: Fix URL output formatting in static-urls, url-args, and main/urls tests.

**Current status**: All three URL tests compile but produce incorrect output. Likely URL escaping or path handling issue.

**Affected tests**:
- urls (main suite)
- urls (static-urls suite)
- urls (url-args suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "static-urls"` to see differences
2. Compare with JavaScript URL handling in `packages/less/src/less/tree/url.go`
3. Check Go's URL implementation in `packages/less/src/less/less_go/url.go`
4. The 4 rewrite-urls tests all pass, so core URL rewriting works - this is likely output formatting
5. Fix URL generation in `url.go` GenCSS method
6. Test all URL suites: `pnpm -w test:go:filter -- "urls"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +3 perfect matches, bringing total to 60/184 (32.6%)

---

## Prompt 4: Fix Import Reference Functionality (2 Tests)

**Goal**: Fix import reference handling so referenced imports don't output CSS but remain available for extends/mixins.

**Current status**: Tests compile but output differs. Files imported with `@import (reference)` are either outputting CSS when they shouldn't, or not making selectors available.

**Affected tests**:
- import-reference (main suite)
- import-reference-issues (main suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"` to see differences
2. Study JavaScript implementation in `packages/less/src/less/tree/import.js` and `import-visitor.js`
3. Check Go's import handling in `packages/less/src/less/less_go/import.go` and `import_visitor.go`
4. The issue: reference flag needs to be preserved and checked during CSS generation
5. Add/fix reference flag handling in Import and Ruleset nodes
6. Ensure rulesets from referenced imports don't output CSS unless explicitly extended/called
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +2 perfect matches, bringing total to 59/184 (32.1%)

**See also**: `.claude/tasks/runtime-failures/import-reference.md`

---

## Prompt 5: Fix CSS Output Formatting Issues (4 Tests)

**Goal**: Fix whitespace/formatting differences in CSS output for comments, parse-interpolation, variables-in-at-rules tests.

**Current status**: Tests compile successfully and logic is correct, but formatting differs (extra newlines, missing spaces, etc).

**Affected tests**:
- comments (main suite)
- parse-interpolation (main suite)
- variables-in-at-rules (main suite)
- container (main suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "comments"` to see first formatting issue
2. Compare with JavaScript output formatting
3. Check Go's GenCSS methods in affected node types (Comment, AtRule, etc)
4. Fix newline/spacing issues to match JavaScript exactly
5. Focus on: newlines before/after blocks, spacing in at-rules, comment formatting
6. Test all affected tests individually
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +4 perfect matches, bringing total to 61/184 (33.2%)

---

## Prompt 6: Fix Mixins-Guards Output in Main Suite (1 Test)

**Goal**: Fix the mixins-guards test in the main suite (the math-always suite version already passes).

**Current status**: Test compiles but output differs. Interesting: the SAME test passes in math-always suite but fails in main suite, suggesting a math mode context issue.

**Affected test**:
- mixins-guards (main suite only - math-always version passes!)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/mixins-guards"` to see differences
2. Run `pnpm -w test:go:filter -- "math-always/mixins-guards"` to confirm this one passes
3. Compare the test configurations - what's different between main and math-always suites?
4. Check how math mode affects guard evaluation in `packages/less/src/less/less_go/mixin_definition.go`
5. Fix guard evaluation to respect math mode context
6. Verify both versions pass: `pnpm -w test:go:filter -- "mixins-guards"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +1 perfect match, bringing total to 58/184 (31.5%)

---

## Prompt 7: Fix Function Output Issues (3 Tests)

**Goal**: Fix output differences in functions, functions-each, and extract-and-length tests.

**Current status**: Tests compile but function results don't match expected output.

**Affected tests**:
- functions (main suite)
- functions-each (main suite)
- extract-and-length (main suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "functions"` to see differences
2. Identify which specific functions are producing wrong output
3. Compare with JavaScript function implementations in `packages/less/src/less/functions/`
4. Fix the corresponding Go functions in `packages/less/src/less/less_go/functions/`
5. Focus on: list functions (extract, length), iteration (each), type conversions
6. Test each individually: `pnpm -w test:go:filter -- "functions" && pnpm -w test:go:filter -- "functions-each"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +3 perfect matches, bringing total to 60/184 (32.6%)

---

## Prompt 8: Fix Detached Ruleset Output (1 Test)

**Goal**: Fix the detached-rulesets test output differences.

**Current status**: Test compiles but detached ruleset evaluation produces incorrect output.

**Affected test**:
- detached-rulesets (main suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"` to see differences
2. Study JavaScript detached ruleset handling in `packages/less/src/less/tree/detached-ruleset.js`
3. Check Go implementation in `packages/less/src/less/less_go/detached_ruleset.go`
4. Previous fixes (Issue #2, #9) fixed major detached ruleset bugs - this is likely edge cases
5. Focus on: variable scoping, frame capture, evaluation context
6. Verify fix: `pnpm -w test:go:filter -- "detached-rulesets"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +1 perfect match, bringing total to 58/184 (31.5%)

---

## Prompt 9: Fix Data-URI Encoding in Include-Path Tests (2 Tests)

**Goal**: Fix data-uri() function encoding differences (spaces vs + signs).

**Current status**: Tests compile but data-uri() output differs - using `+` instead of `%20` for spaces.

**Affected tests**:
- include-path (include-path suite)
- include-path-string (include-path-string suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "include-path"` to see encoding difference
2. Expected shows `%20` for spaces, actual shows `+`
3. Check JavaScript data-uri implementation in `packages/less/src/less/functions/data-uri.js`
4. Fix Go's data-uri in `packages/less/src/less/less_go/functions/data_uri.go`
5. Use proper URL encoding (space → %20, not +)
6. Test both: `pnpm -w test:go:filter -- "include-path"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +2 perfect matches, bringing total to 59/184 (32.1%)

---

## Prompt 10: Fix Selectors and String Output (2 Tests)

**Goal**: Fix selector interpolation and string escaping issues.

**Current status**: Tests compile but selector/string output doesn't match expected.

**Affected tests**:
- selectors (main suite)
- strings (main suite)

**Task**:
1. Run `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "selectors"` to see differences
2. Then check: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "strings"`
3. Study JavaScript implementations in `selector.js` and `quoted.js`
4. Fix Go implementations in `selector.go` and `quoted.go`
5. Focus on: interpolation in selectors, string escape sequences, quote handling
6. Test both: `pnpm -w test:go:filter -- "selectors" && pnpm -w test:go:filter -- "strings"`
7. Check for regressions: `pnpm -w test:go:unit && pnpm -w test:go`

**Expected result**: +2 perfect matches, bringing total to 59/184 (32.1%)

---

## Summary

These 10 prompts cover:
- **Quick wins**: extend-chaining (1 test, completes category)
- **High impact**: Math operations (6 tests), URL issues (3 tests), Formatting (4 tests)
- **Medium impact**: Import reference (2 tests), Functions (3 tests), Data-URI (2 tests)
- **Individual fixes**: Mixins-guards (1 test), Detached rulesets (1 test), Selectors/strings (2 tests)

**Total potential**: +25 perfect matches (57 → 82, reaching 44.6% perfect match rate)

**Priority order** (for maximum impact):
1. Prompt 1 (extend-chaining) - Quick win, completes category
2. Prompt 2 (Math ops) - Highest test count (+6)
3. Prompt 3 (URLs) - High impact (+3)
4. Prompt 5 (Formatting) - High impact (+4)
5. Remaining prompts as needed

**Current baseline** (verify before starting):
- Unit tests: 2,290+ passing (99.9%+)
- Integration: 57 perfect matches, 0 regressions
- Compilation rate: 98.4% (181/184)

**After each agent completes work**: Re-run full test suite and update counts in CLAUDE.md and AGENT_WORK_QUEUE.md
