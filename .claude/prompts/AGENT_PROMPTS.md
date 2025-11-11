# Agent Prompts for Parallel Work

## Instructions for Each Agent

**CRITICAL**: Before starting, each agent MUST:
1. Run `pnpm -w test:go:unit` to verify all unit tests pass
2. Run `pnpm -w test:go` to get current integration test baseline
3. After making changes, run BOTH test suites again to check for regressions
4. Compare results against `.claude/status/TEST_RESULTS_2025-11-10.md` baseline
5. Report any regressions immediately

## Prompt 1: Fix import-reference-issues Test

**Task**: Fix the `import-reference-issues` integration test which has extra blank lines and whitespace in output.

**Context**: Test compiles successfully but has minor whitespace differences in CSS output. The expected output should not have extra blank lines between selectors. See `test-data/less/_main/import-reference-issues.less` for input.

**Success Criteria**:
- Test shows "Perfect match!" instead of "Output differs"
- Run full test suite: zero regressions in unit tests (2,290+ passing) and integration tests (80 perfect matches maintained)
- Verify extend-chaining test still passes (known regression risk)

**Files to Investigate**:
- CSS generation code in tree/import.go
- Reference import handling logic

---

## Prompt 2: Fix import-reference Test

**Task**: Fix the `import-reference` integration test which has import reference filtering issues.

**Context**: This test validates that `@import (reference)` correctly filters out unused CSS. Currently producing incorrect output. Related to import-reference-issues but may have different root cause.

**Success Criteria**:
- Test shows "Perfect match!"
- Zero regressions: verify all 80 currently passing tests still pass
- All 11 namespacing tests still perfect matches

**Files to Investigate**:
- tree/import.go reference handling
- Ruleset filtering logic during CSS generation

---

## Prompt 3: Fix detached-rulesets Media Query Merging

**Task**: Fix the `detached-rulesets` test which has media query merging issues.

**Context**: Detached rulesets should properly merge media queries when evaluated. Current behavior may be duplicating or incorrectly nesting media queries.

**Success Criteria**:
- Test shows "Perfect match!"
- All detached-ruleset functionality tests pass
- Zero regressions in 80 passing tests

**Files to Investigate**:
- tree/detached_ruleset.go
- Media query merging logic in contexts/eval.go
- tree/media.go

---

## Prompt 4: Fix urls Test (Main Suite)

**Task**: Fix the `urls` test in the main suite which has URL handling issues.

**Context**: URL processing/rewriting has issues in the main test. Note: URL rewriting tests (rewrite-urls-all, rewrite-urls-local) pass perfectly, so the issue is likely with base URL handling or specific URL formats.

**Success Criteria**:
- Main suite urls test shows "Perfect match!"
- All 4 URL rewriting tests still pass
- Zero regressions

**Files to Investigate**:
- tree/url.go
- URL processing in contexts/eval.go
- Compare with passing URL rewriting test cases

---

## Prompt 5: Fix urls Test (static-urls Suite)

**Task**: Fix the `urls` test in the static-urls suite.

**Context**: Similar to main urls test but with static URL configuration. May have different root cause related to static URL handling.

**Success Criteria**:
- static-urls suite test shows "Perfect match!"
- All other URL tests remain passing
- Zero regressions

**Files to Investigate**:
- tree/url.go static URL handling
- URL option processing

---

## Prompt 6: Fix urls Test (url-args Suite)

**Task**: Fix the `urls` test in the url-args suite which handles URL function arguments.

**Context**: URL function with arguments may not be processing correctly. Test validates URL function calls with various argument types.

**Success Criteria**:
- url-args suite test shows "Perfect match!"
- All other URL tests remain passing
- Zero regressions

**Files to Investigate**:
- functions/url.go
- URL argument parsing and evaluation

---

## Prompt 7: Fix functions Test Output Issue

**Task**: Fix the `functions` test which has function output issues.

**Context**: One or more built-in functions producing incorrect output. The functions-each test passes perfectly, so issue is likely with a specific function not covered by that test.

**Success Criteria**:
- functions test shows "Perfect match!"
- functions-each test still passes
- All 80 passing tests maintained

**Files to Investigate**:
- Compare functions.less test with functions-each.less to identify difference
- Specific function implementations in functions/ directory
- Function call evaluation in tree/call.go

---

## Prompt 8: Fix comments2 Keyframes Issue

**Task**: Fix the `comments2` test which is missing webkit keyframes rule in output.

**Context**: Test expects `@-webkit-keyframes hover` to be in output but it's missing. Other comments are handled correctly. Likely an issue with directive comment handling or keyframes-specific logic.

**Success Criteria**:
- comments2 test shows "Perfect match!"
- comments test still passes
- Zero regressions

**Files to Investigate**:
- tree/comment.go
- tree/directive.go (keyframes handling)
- At-rule processing with comments

---

## Prompt 9: Fix Media Query Formatting Tests

**Task**: Fix the `media`, `directives-bubling`, and `container` tests which have CSS output formatting issues.

**Context**: These tests compile correctly but have minor formatting differences in media queries, directive bubbling, or container queries. Likely whitespace, indentation, or line break issues.

**Success Criteria**:
- All 3 tests show "Perfect match!"
- All other media/directive tests still pass
- Zero regressions

**Files to Investigate**:
- tree/media.go genCSS method
- tree/directive.go and tree/at_rule.go
- CSS output formatting logic in output/output.go

---

## Prompt 10: Fix css-3 CSS3 Features Formatting

**Task**: Fix the `css-3` test which has CSS3 feature formatting issues.

**Context**: CSS3 specific features (likely calc, grid, variables, etc.) have minor output formatting differences. May be related to CSS3-specific property handling or vendor prefixes.

**Success Criteria**:
- css-3 test shows "Perfect match!"
- css-grid test still passes
- All other CSS tests maintained

**Files to Investigate**:
- CSS3-specific feature handling
- tree/at_rule.go for @supports, @layer, etc.
- Vendor prefix logic if any

---

## Bonus Prompt 11: Investigate Error Handling Tests

**Task**: Analyze why 27 error tests pass when they should fail and create a plan to fix them.

**Context**: Tests in eval-errors and parse-errors suites should fail but currently succeed. Need proper error detection and reporting. This is likely a systematic issue rather than individual test problems.

**Success Criteria**:
- Document root cause analysis
- Create fix plan for error detection
- No need to fix all 27 - just identify the pattern

**Files to Investigate**:
- Error handling in parser and evaluator
- Error propagation from tree nodes
- Compare with JavaScript less.js error handling

---

## Notes for All Agents

- **Regression Prevention**: The extend-chaining test has been a regression risk in the past. Always verify it still passes.
- **Test Baseline**: Current status is 80 perfect matches out of 90 usable tests (88.9% success rate)
- **Unit Tests**: Must maintain 2,290+ passing unit tests with zero regressions
- **Documentation**: Current results documented in `.claude/status/TEST_RESULTS_2025-11-10.md`
- **Parallel Work**: These prompts are designed to be independent - agents can work simultaneously
- **Priority Order**: Prompts 1-6 are HIGH priority (CSS correctness), 7-10 are MEDIUM priority (formatting), 11 is analysis only

## Command Quick Reference

```bash
# Run specific integration test
go test -v ./packages/less/src/less/less_go -run TestIntegrationSuite/main/[test-name]

# Run all unit tests
pnpm -w test:go:unit

# Run all integration tests
pnpm -w test:go

# Check current test counts
pnpm -w test:go 2>&1 | grep "Perfect match\|Output differs\|Compilation failed"
```
