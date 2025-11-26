# Agent Prompts for less.go Port - 2025-11-26

These are ready-to-use prompts for independent agents to fix remaining issues in the less.go port.

**Current Status:**
- 83 perfect CSS matches (45.1%)
- 9 output differences remaining
- 2 error handling tests remaining
- All 2,304 unit tests passing

**CRITICAL**: Before creating any PR, agents MUST:
1. Run `pnpm -w test:go:unit` - must pass 100%
2. Run `pnpm -w test:go` - baseline is 83 perfect matches, NO regressions allowed
3. Verify the specific test(s) being fixed now show "Perfect match!"

---

## Prompt 1: Fix Import Reference (HIGH PRIORITY - 2 tests)

```
You are fixing the import-reference functionality in the less.go port. This affects 2 tests: import-reference and import-reference-issues.

**Problem**: Files imported with `(reference)` option should NOT output CSS, but their selectors/mixins should be available for extends and mixin calls.

**Task file**: Read `.claude/tasks/runtime-failures/import-reference.md` for detailed investigation guidance.

**Files to focus on**:
- `packages/less/src/less/less_go/import.go`
- `packages/less/src/less/less_go/import_visitor.go`
- `packages/less/src/less/less_go/ruleset.go`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/import-reference"
```

**Success criteria**:
- Both import-reference tests show "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 85+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 2: Fix Detached Rulesets Media Query Merging (HIGH PRIORITY - 1 test)

```
You are fixing the detached-rulesets test in less.go. The test compiles but media queries from detached rulesets called within parent @media blocks are not being merged correctly.

**Problem**: When a detached ruleset containing @media is called from within another @media block, the queries should merge. Currently the merged queries don't appear in output.

**Task file**: Read `.claude/tasks/runtime-failures/detached-rulesets-continuation.md` for root cause analysis.

**Root cause identified**: Parent @media nodes are calling `evalNested()` instead of `evalTop()` due to mediaPath not being empty when expected.

**Files to focus on**:
- `packages/less/src/less/less_go/media.go` - Media.Eval(), evalTop(), evalNested()
- `packages/less/src/less/less_go/detached_ruleset.go` - CallEval()
- Compare with `packages/less/src/less/tree/media.js`

**Debug command**:
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/main/detached-rulesets"
```

**Success criteria**:
- detached-rulesets test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 3: Fix URLs Main Suite (HIGH PRIORITY - 1 test)

```
You are fixing the `urls` test in the main suite of less.go. The test compiles but produces different CSS output for URL handling.

**Problem**: URL processing has edge cases not matching less.js behavior.

**Files to focus on**:
- `packages/less/src/less/less_go/url.go` - URL node evaluation
- `packages/less/src/less/less_go/ruleset.go` - URL handling in rules
- Compare with `packages/less/src/less/tree/url.js`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/urls"
```

**Test data**:
- Input: `packages/test-data/less/_main/urls.less`
- Expected: `packages/test-data/css/_main/urls.css`

**Success criteria**:
- urls (main) test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 4: Fix URLs Static-URLs Suite (HIGH PRIORITY - 1 test)

```
You are fixing the `urls` test in the static-urls suite of less.go. This tests URL handling with the staticUrls option.

**Problem**: URL processing with staticUrls option not matching less.js behavior.

**Files to focus on**:
- `packages/less/src/less/less_go/url.go` - URL node evaluation
- `packages/less/src/less/less_go/contexts.go` - staticUrls option
- Compare with `packages/less/src/less/tree/url.js`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/static-urls/urls"
```

**Success criteria**:
- urls (static-urls) test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 5: Fix URLs URL-Args Suite (HIGH PRIORITY - 1 test)

```
You are fixing the `urls` test in the url-args suite of less.go. This tests URL handling with the urlArgs option.

**Problem**: URL processing with urlArgs option not matching less.js behavior.

**Files to focus on**:
- `packages/less/src/less/less_go/url.go` - URL node evaluation
- `packages/less/src/less/less_go/contexts.go` - urlArgs option
- Compare with `packages/less/src/less/tree/url.js`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/url-args/urls"
```

**Success criteria**:
- urls (url-args) test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 6: Fix Container CSS Output (MEDIUM PRIORITY - 1 test)

```
You are fixing the `container` test in less.go. The test compiles but produces different CSS output for @container rules.

**Problem**: CSS @container at-rule output formatting doesn't match less.js.

**Files to focus on**:
- `packages/less/src/less/less_go/at_rule.go` - AtRule processing
- `packages/less/src/less/less_go/ruleset.go` - Selector handling for container
- Compare with `packages/less/src/less/tree/atrule.js`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/container"
```

**Test data**:
- Input: `packages/test-data/less/_main/container.less`
- Expected: `packages/test-data/css/_main/container.css`

**Success criteria**:
- container test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 7: Fix Directives Bubbling (MEDIUM PRIORITY - 1 test)

```
You are fixing the `directives-bubling` test in less.go. The test compiles but CSS output for bubbled at-rules has formatting differences.

**Problem**: When at-rules (like @media, @supports) are nested inside selectors, they should "bubble" up to the root. The output formatting doesn't match less.js.

**Files to focus on**:
- `packages/less/src/less/less_go/ruleset.go` - Ruleset bubbling logic
- `packages/less/src/less/less_go/at_rule.go` - AtRule bubbling
- `packages/less/src/less/less_go/media.go` - Media bubbling
- Compare with `packages/less/src/less/tree/ruleset.js`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/directives-bubling"
```

**Test data**:
- Input: `packages/test-data/less/_main/directives-bubling.less`
- Expected: `packages/test-data/css/_main/directives-bubling.css`

**Success criteria**:
- directives-bubling test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 8: Fix Media Query Output (MEDIUM PRIORITY - 1 test)

```
You are fixing the `media` test in the main suite of less.go. The test compiles but media query CSS output has differences.

**Problem**: Media query output formatting or evaluation doesn't match less.js.

**Files to focus on**:
- `packages/less/src/less/less_go/media.go` - Media node evaluation
- `packages/less/src/less/less_go/ruleset.go` - Media handling in rulesets
- Compare with `packages/less/src/less/tree/media.js`

**Debug command**:
```bash
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/media"
```

**Test data**:
- Input: `packages/test-data/less/_main/media.less`
- Expected: `packages/test-data/css/_main/media.css`

**Success criteria**:
- media test shows "Perfect match!"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 84+ perfect matches (no regressions from 83)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate.
```

---

## Prompt 9: Fix Color Function Error Validation (MEDIUM PRIORITY - 1 test)

```
You are fixing error validation for the `color()` function in less.go. The test `color-func-invalid-color-2` should fail with an error but currently compiles successfully.

**Problem**: The `color("NOT A COLOR")` call should throw an error for invalid color strings.

**Task file**: Read `.claude/tasks/error-handling/EXPECTED_ERROR_TESTS.md` for context.

**Files to focus on**:
- `packages/less/src/less/less_go/functions/color.go` - color() function
- Compare with `packages/less/src/less/functions/color.js`

**Test data**:
- Input: `packages/test-data/errors/eval/color-func-invalid-color-2.less`

**Debug command**:
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/color-func-invalid-color-2"
```

**Success criteria**:
- color-func-invalid-color-2 test shows "Correctly failed with error"
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 83+ perfect matches (no regressions)

**Note**: This is an error test - success means it now correctly FAILS with an error, not that it produces output.

**Regression check**: Document baseline of 83 perfect matches and 87 correct error handling. If counts drop, stop and investigate.
```

---

## Prompt 10: Fix Regex Compilation Performance (MEDIUM PRIORITY - Performance)

```
You are fixing a critical performance issue in less.go: regex patterns are being compiled on every use instead of once at startup.

**Problem**: Profiling shows 81% of memory allocations come from regexp.compile, causing 8-10x slowdown vs JavaScript.

**Task file**: Read `.claude/tasks/performance/CRITICAL_regex_compilation.md` for full analysis.

**Solution pattern**:
```go
// BEFORE (wrong - compiles every call):
func parseIdentifier(input string) string {
    regex := regexp.MustCompile(`^[a-zA-Z_]...`)
    return regex.FindString(input)
}

// AFTER (correct - compile once at package level):
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_]...`)
func parseIdentifier(input string) string {
    return identifierRegex.FindString(input)
}
```

**Search for culprits**:
```bash
grep -rn "regexp.Compile\|regexp.MustCompile" packages/less/src/less/less_go/*.go
```

**Focus files** (most likely culprits):
- `packages/less/src/less/less_go/parser.go`
- Any file with inline regexp compilation

**Success criteria**:
- All regexes moved to package-level vars
- `pnpm bench:profile` shows regexp.compile < 5% of allocations
- `pnpm -w test:go:unit` passes 100%
- `pnpm -w test:go` shows 83+ perfect matches (no regressions)

**Regression check**: Document baseline of 83 perfect matches. If count drops, stop and investigate. This is a refactoring task - no functional changes should occur.
```

---

## Summary of Remaining Work

| Priority | Test/Task | Prompt |
|----------|-----------|--------|
| HIGH | import-reference (2 tests) | Prompt 1 |
| HIGH | detached-rulesets (1 test) | Prompt 2 |
| HIGH | urls - main (1 test) | Prompt 3 |
| HIGH | urls - static-urls (1 test) | Prompt 4 |
| HIGH | urls - url-args (1 test) | Prompt 5 |
| MEDIUM | container (1 test) | Prompt 6 |
| MEDIUM | directives-bubling (1 test) | Prompt 7 |
| MEDIUM | media (1 test) | Prompt 8 |
| MEDIUM | color-func error (1 test) | Prompt 9 |
| MEDIUM | Regex performance | Prompt 10 |

**Total**: Fixing all output differences would bring us to **92 perfect matches (50%)** with 92.4% overall success rate!
