# Agent Prompts for Next Priorities - November 9, 2025

**Generated**: 2025-11-09
**Current Status**: 69/184 perfect matches (37.5%), 75% overall success
**Target**: 80% success rate

These are 10 short, focused prompts to kick off independent agents for the highest-priority fixes. Each prompt is designed to be self-contained and includes a reminder to check for regressions.

---

## Prompt 1: Fix import-reference Tests

```
Fix the import-reference functionality in less.go. The tests `import-reference` and `import-reference-issues` compile but produce incorrect CSS output.

The issue: Files imported with `@import (reference) "file.less";` should not output their CSS by default, but their selectors/mixins should be available when explicitly referenced.

Key files to investigate:
- packages/less/src/less/less_go/import.go
- packages/less/src/less/less_go/import_visitor.go
- packages/less/src/less/less_go/ruleset.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/import.js
- packages/less/src/less/import-visitor.js

Test commands:
- Run specific tests: `pnpm -w test:go:filter -- "import-reference"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: Both import-reference tests show "Perfect match!"
```

---

## Prompt 2: Fix math-parens Suite Tests

```
Fix math operations in the math-parens suite in less.go. Three tests compile but produce incorrect CSS: `css`, `mixins-args`, and `parens`.

The issue: Math mode handling in parenthesized expressions isn't matching less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/operation.go
- packages/less/src/less/less_go/contexts.go
- packages/less/src/less/less_go/dimension.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/operation.js
- packages/less/src/less/contexts.js

Test commands:
- Run specific tests: `pnpm -w test:go:filter -- "math-parens"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "parens"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: css, mixins-args, and parens in math-parens suite show "Perfect match!"
```

---

## Prompt 3: Fix units-no-strict Test

```
Fix unit handling in non-strict mode in less.go. The test `no-strict` in the units-no-strict suite compiles but produces incorrect CSS output.

The issue: Unit handling in non-strict mode doesn't match less.js behavior, particularly with division operations.

Key files to investigate:
- packages/less/src/less/less_go/dimension.go
- packages/less/src/less/less_go/operation.go
- packages/less/src/less/less_go/contexts.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/dimension.js
- packages/less/src/less/tree/operation.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "no-strict"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "no-strict"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: no-strict test shows "Perfect match!"
```

---

## Prompt 4: Fix URL Handling Tests

```
Fix URL handling edge cases in less.go. Three tests compile but produce incorrect CSS: `urls` in main suite, static-urls suite, and url-args suite.

The issue: URL processing has edge cases that don't match less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/url.go
- packages/less/src/less/less_go/ruleset.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/url.js

Test commands:
- Run specific tests: `pnpm -w test:go:filter -- "urls"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "urls"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: All three urls tests show "Perfect match!"
```

---

## Prompt 5: Fix detached-rulesets Test

```
Fix detached ruleset output in less.go. The test `detached-rulesets` compiles but produces incorrect CSS output.

The issue: Detached rulesets output formatting doesn't match less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/detached_ruleset.go
- packages/less/src/less/less_go/ruleset.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/detached-ruleset.js
- packages/less/src/less/tree/ruleset.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "detached-rulesets"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: detached-rulesets test shows "Perfect match!"
```

---

## Prompt 6: Fix functions Test

```
Fix various function edge cases in less.go. The test `functions` in the main suite compiles but produces incorrect CSS output.

The issue: Various built-in functions have edge cases that don't match less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/functions/*.go (multiple function files)
- Run the test first to see which specific functions are failing

Compare with JavaScript implementation in:
- packages/less/src/less/functions/*.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "^functions$"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^functions$"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: functions test shows "Perfect match!"
```

---

## Prompt 7: Fix functions-each Test

```
Fix the each() function in less.go. The test `functions-each` compiles but produces incorrect CSS output.

The issue: The each() function iteration behavior doesn't match less.js.

Key files to investigate:
- packages/less/src/less/less_go/functions/list.go (or wherever each() is implemented)
- packages/less/src/less/less_go/call.go

Compare with JavaScript implementation in:
- packages/less/src/less/functions/list.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "functions-each"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "functions-each"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: functions-each test shows "Perfect match!"
```

---

## Prompt 8: Fix extract-and-length Test

```
Fix the extract() and length() list functions in less.go. The test `extract-and-length` compiles but produces incorrect CSS output.

The issue: List function behavior doesn't match less.js.

Key files to investigate:
- packages/less/src/less/less_go/functions/list.go
- packages/less/src/less/less_go/expression.go

Compare with JavaScript implementation in:
- packages/less/src/less/functions/list.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "extract-and-length"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "extract-and-length"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: extract-and-length test shows "Perfect match!"
```

---

## Prompt 9: Fix directives-bubling Test

```
Fix directive bubbling behavior in less.go. The test `directives-bubling` compiles but produces incorrect CSS output.

The issue: At-rule directive bubbling (moving nested directives to parent level) doesn't match less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/at_rule.go
- packages/less/src/less/less_go/ruleset.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/atrule.js
- packages/less/src/less/tree/ruleset.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "directives-bubling"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: directives-bubling test shows "Perfect match!"
```

---

## Prompt 10: Fix container Test

```
Fix container query handling in less.go. The test `container` compiles but produces incorrect CSS output.

The issue: CSS container query syntax handling doesn't match less.js behavior.

Key files to investigate:
- packages/less/src/less/less_go/at_rule.go
- packages/less/src/less/less_go/media.go

Compare with JavaScript implementation in:
- packages/less/src/less/tree/atrule.js
- packages/less/src/less/tree/media.js

Test commands:
- Run specific test: `pnpm -w test:go:filter -- "container"`
- See diff: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "container"`

**CRITICAL**: Before creating PR, run ALL tests to check for regressions:
- Unit tests: `pnpm -w test:go:unit` (must pass 100%)
- Integration tests: `pnpm -w test:go` (compare against baseline: 69 perfect matches, no regressions)

Expected outcome: container test shows "Perfect match!"
```

---

## Usage Instructions

1. **Pick a prompt** from above (preferably in order 1-10)
2. **Copy the prompt** verbatim into a new Claude session
3. **Set the branch**: Each agent should work on `claude/fix-{task-name}-{session-id}`
4. **Let the agent work**: The prompt contains all context needed
5. **Review the PR**: Check that tests pass and no regressions occurred

## Success Metrics

If all 10 prompts are completed successfully:
- **Perfect matches**: 69 → 79+ (42.9%+)
- **Overall success**: 75% → 82%+
- **Output differs**: 23 → 13 (7.1%)

This would be a **huge achievement** and put us well past 80% success rate!

---

**Generated**: 2025-11-09 by Claude
**Current baseline**: 69 perfect matches, 138/184 overall success (75%)
