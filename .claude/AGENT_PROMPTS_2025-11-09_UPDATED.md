# Agent Prompts for Next Priorities - November 9, 2025 (UPDATED)

**Generated**: 2025-11-09 (Final Update)
**Current Status**: 69/184 perfect matches (37.5%), 75% overall success, ZERO regressions
**Target**: 80% success rate (need +9 perfect matches)

These are 10 short, focused prompts to kick off independent agents for the highest-priority fixes. Each prompt is designed to be self-contained and includes a reminder to check for regressions against the baseline of **69 perfect matches**.

---

## Prompt 1: Fix import-reference Tests ⚡ HIGH IMPACT

```
Fix the import-reference functionality in less.go. The tests `import-reference` and `import-reference-issues` compile but produce incorrect CSS output.

The issue: Files imported with `@import (reference) "file.less";` should not output their CSS by default. Selectors and mixins from referenced imports should only appear in the output when explicitly used/extended.

Current behavior: Referenced imports are outputting CSS they shouldn't, and selectors that should be filtered are appearing in the output.

Key files to investigate:
- packages/less/src/less/less_go/import.go
- packages/less/src/less/less_go/import_visitor.go
- packages/less/src/less/less_go/ruleset.go
- packages/less/src/less/less_go/extend_visitor.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/import.js
- packages/less/src/less/import-visitor.js

Test commands:
- Run specific tests: `pnpm -w test:go:filter -- "import-reference"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: Both import-reference and import-reference-issues tests show "Perfect match!"
Impact: +2 tests → 71/184 perfect matches (38.6%), 76.1% overall success
```

---

## Prompt 2: Fix math-parens Suite Tests ⚡ HIGH IMPACT

```
Fix math operations in the math-parens suite in less.go. Three tests compile but produce incorrect CSS: `css`, `mixins-args`, and `parens`.

The issue: When math mode is "parens-only", operations should only evaluate inside parentheses. Currently the behavior doesn't match less.js - operations are evaluating when they shouldn't, or not evaluating when they should.

Key files to investigate:
- packages/less/src/less/less_go/operation.go - Operation evaluation logic
- packages/less/src/less/less_go/paren.go - Parenthesis handling
- packages/less/src/less/less_go/contexts.go - Math mode context tracking
- packages/less/src/less/less_go/dimension.go - Dimension operations

Compare with JavaScript implementation in:
- packages/less/src/less/tree/operation.js
- packages/less/src/less/tree/paren.js
- packages/less/src/less/contexts.js

Test commands:
- Run specific tests: `pnpm -w test:go:filter -- "math-parens"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "parens"`
- Compare suites: `pnpm -w test:go:filter -- "math-"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: css, mixins-args, and parens in math-parens suite show "Perfect match!"
Impact: +3 tests → 72/184 perfect matches (39.1%), 77.7% overall success
```

---

## Prompt 3: Fix units-no-strict Test ⚡ QUICK WIN

```
Fix unit handling in non-strict mode in less.go. The test `no-strict` in the units-no-strict suite compiles but produces incorrect CSS output.

The issue: In non-strict mode, division operations with mixed units should produce different output than in strict mode. The test shows operations like "4 / 2 + 5em" being output as-is instead of being evaluated.

Key files to investigate:
- packages/less/src/less/less_go/dimension.go - Unit handling and operations
- packages/less/src/less/less_go/operation.go - Math operation evaluation
- packages/less/src/less/less_go/contexts.go - strictUnits flag handling

Compare with JavaScript implementation in:
- packages/less/src/less/tree/dimension.js
- packages/less/src/less/tree/operation.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "no-strict"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "no-strict"`
- Compare with strict: `pnpm -w test:go:filter -- "units-"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: no-strict test shows "Perfect match!"
Impact: +1 test → 70/184 perfect matches (38%), completes units category 2/2
```

---

## Prompt 4: Fix URL Handling Tests ⚡ HIGH IMPACT

```
Fix URL handling edge cases in less.go. Three tests compile but produce incorrect CSS: `urls` in main, static-urls, and url-args suites.

The issue: URL processing has edge cases around quoting, escaping, and path handling that don't match less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/url.go - URL node and processing
- packages/less/src/less/less_go/quoted.go - Quote handling
- packages/less/src/less/less_go/functions/string.go - URL-related functions

Compare with JavaScript implementation in:
- packages/less/src/less/tree/url.js
- packages/less/src/less/functions/string.js

Test commands:
- Run all URL tests: `pnpm -w test:go:filter -- "urls"`
- See main diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/urls"`
- See static diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "static-urls"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: All three urls tests show "Perfect match!"
Impact: +3 tests → 72/184 perfect matches (39.1%), 79.3% overall success
```

---

## Prompt 5: Fix detached-rulesets Test

```
Fix detached ruleset output formatting in less.go. The test `detached-rulesets` compiles but produces incorrect CSS output.

The issue: Detached rulesets (variables that hold rulesets) are being output with incorrect formatting or structure.

Key files to investigate:
- packages/less/src/less/less_go/detached_ruleset.go - DetachedRuleset node
- packages/less/src/less/less_go/ruleset.go - Ruleset output generation
- packages/less/src/less/less_go/mixin_call.go - Detached ruleset calls

Compare with JavaScript implementation in:
- packages/less/src/less/tree/detached-ruleset.js
- packages/less/src/less/tree/ruleset.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "detached-rulesets"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: detached-rulesets test shows "Perfect match!"
Impact: +1 test → 70/184 perfect matches (38%)
```

---

## Prompt 6: Fix functions Test

```
Fix various function edge cases in less.go. The test `functions` in the main suite compiles but produces incorrect CSS output.

The issue: Various built-in functions have edge cases or missing functionality that don't match less.js behavior.

Approach:
1. Run the test first to see specific function failures in the diff
2. Focus on the failing functions one at a time
3. Compare each with the JavaScript implementation

Key files to investigate:
- packages/less/src/less/less_go/functions/*.go (multiple function files)
- Run test first to identify which specific functions are failing

Compare with JavaScript implementation in:
- packages/less/src/less/functions/*.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "^functions$"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^functions$"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: functions test shows "Perfect match!"
Impact: +1 test → 70/184 perfect matches (38%)
```

---

## Prompt 7: Fix functions-each Test

```
Fix the each() function in less.go. The test `functions-each` compiles but produces incorrect CSS output.

The issue: The each() function doesn't properly iterate over lists/values or doesn't create the right scope/variables during iteration.

Key files to investigate:
- packages/less/src/less/less_go/functions/list.go - each() implementation
- packages/less/src/less/less_go/call.go - Function call evaluation
- packages/less/src/less/less_go/contexts.go - Function evaluation context

Compare with JavaScript implementation in:
- packages/less/src/less/functions/list.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "functions-each"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "functions-each"`
- Debug: `LESS_GO_TRACE=1 pnpm -w test:go:filter -- "functions-each"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: functions-each test shows "Perfect match!"
Impact: +1 test → 70/184 perfect matches (38%)
```

---

## Prompt 8: Fix extract-and-length Test

```
Fix the extract() and length() list functions in less.go. The test `extract-and-length` compiles but produces incorrect CSS output.

The issue: List extraction and length calculation don't match less.js behavior. These functions work with comma/space-separated lists.

Key files to investigate:
- packages/less/src/less/less_go/functions/list.go - extract() and length() implementations
- packages/less/src/less/less_go/expression.go - Expression/list handling
- packages/less/src/less/less_go/value.go - Value list handling

Compare with JavaScript implementation in:
- packages/less/src/less/functions/list.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "extract-and-length"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "extract-and-length"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: extract-and-length test shows "Perfect match!"
Impact: +1 test → 70/184 perfect matches (38%)
```

---

## Prompt 9: Fix directives-bubling Test

```
Fix directive bubbling behavior in less.go. The test `directives-bubling` compiles but produces incorrect CSS output.

The issue: At-rule directives (like @media, @supports) should "bubble up" from nested contexts to parent levels. The bubbling logic doesn't match less.js.

Key files to investigate:
- packages/less/src/less/less_go/at_rule.go - AtRule node and bubbling logic
- packages/less/src/less/less_go/ruleset.go - Ruleset directive handling
- packages/less/src/less/less_go/media.go - Media query handling

Compare with JavaScript implementation in:
- packages/less/src/less/tree/atrule.js
- packages/less/src/less/tree/ruleset.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "directives-bubling"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: directives-bubling test shows "Perfect match!"
Impact: +1 test → 70/184 perfect matches (38%)
```

---

## Prompt 10: Fix colors Precision Tests

```
Fix color precision handling in less.go. The tests `colors` and `colors2` compile but produce CSS with incorrect floating-point precision in HSL values.

The issue: HSL color output shows values like "198.00000000000006" instead of "198". Color calculations accumulate floating-point errors that need to be rounded.

Key files to investigate:
- packages/less/src/less/less_go/color.go - Color node and operations
- packages/less/src/less/less_go/functions/color.go - Color functions
- Focus on HSL output formatting (toHSL method)

Compare with JavaScript implementation in:
- packages/less/src/less/tree/color.js
- packages/less/src/less/functions/color.js

Test commands:
- Run specific tests: `pnpm -w test:go:filter -- "colors"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "colors2"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (baseline: 69 perfect matches, no regressions)

Expected outcome: Both colors and colors2 tests show "Perfect match!"
Impact: +2 tests → 71/184 perfect matches (38.6%)
```

---

## Usage Instructions

1. **Pick a prompt** from above (preferably in priority order 1-4 for fastest path to 80%)
2. **Copy the prompt** verbatim into a new Claude session
3. **Set the branch**: Each agent should work on `claude/fix-{task-name}-{session-id}`
4. **Let the agent work**: The prompt contains all context needed
5. **Review the PR**: Check that tests pass and no regressions occurred (baseline: 69 perfect matches)

## Success Metrics

**Path to 80% (Recommended):**
- Prompts 1, 2, 3, 4 = +9 tests = 78/184 perfect matches = **80%+ overall success** ✨

**Stretch to 85%:**
- Add Prompts 5, 7, 8 = +3 more tests = 81/184 perfect matches = **81.5% overall success**

**Complete all 10:**
- All prompts = +16 tests = 85/184 perfect matches (46.2%) = **85%+ overall success** 🎉

## Priority Tiers

**TIER 1 - Path to 80%** (8-12 hours total):
1. Prompt 1: import-reference (+2)
2. Prompt 2: math-parens (+3)
3. Prompt 3: units-no-strict (+1)
4. Prompt 4: urls (+3)

**TIER 2 - Quick Wins** (6-9 hours total):
5. Prompt 5: detached-rulesets (+1)
6. Prompt 7: functions-each (+1)
7. Prompt 8: extract-and-length (+1)

**TIER 3 - Polish** (8-12 hours total):
8. Prompt 6: functions (+1)
9. Prompt 9: directives-bubling (+1)
10. Prompt 10: colors precision (+2)

---

**Generated**: 2025-11-09 by Claude
**Current baseline**: 69 perfect matches, 138/184 overall success (75%), ZERO regressions
**Next milestone**: 80% overall success (147/184 tests)
