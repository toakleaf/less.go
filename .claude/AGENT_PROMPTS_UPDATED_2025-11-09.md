# Agent Prompts for Next Priorities - November 9, 2025 (UPDATED)

**Generated**: 2025-11-09 (Updated after latest test run)
**Current Status**: 74/177 perfect matches (40.2%), 76.8% overall success
**Target**: 80% success rate (need +6 tests)

These are 10 short, focused prompts to kick off independent agents for the highest-priority fixes. Each prompt includes validation requirements to prevent regressions.

---

## Prompt 1: Fix import-reference Tests ⚡ HIGH PRIORITY

```
Fix the import-reference functionality in less.go. The tests `import-reference` and `import-reference-issues` compile but produce incorrect CSS output.

**The issue**: Files imported with `@import (reference) "file.less";` should not output their CSS by default, but their selectors/mixins should be available when explicitly referenced.

**Key files to investigate**:
- packages/less/src/less/less_go/import.go
- packages/less/src/less/less_go/import_visitor.go
- packages/less/src/less/less_go/ruleset.go

**Compare with JavaScript**:
- packages/less/src/less/tree/import.js
- packages/less/src/less/import-visitor.js

**Test commands**:
```bash
# Run specific tests
pnpm -w test:go:filter -- "import-reference"

# See detailed diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
# 1. Unit tests must pass 100%
pnpm -w test:go:unit

# 2. Integration tests - verify no regressions
pnpm -w test:go

# 3. Check baseline: Must have 76 perfect matches (74 current + 2 new)
#    Look for: "Perfect CSS Matches: 76" or similar output
```

**Expected outcome**:
- import-reference shows "✅ Perfect match!"
- import-reference-issues shows "✅ Perfect match!"
- No regressions in other tests
```

---

## Prompt 2: Fix math-parens Suite ⚡ HIGH PRIORITY

```
Fix math operations in the math-parens suite. The tests `css` and `mixins-args` compile but produce incorrect CSS output.

**The issue**: Math mode handling in parenthesized expressions isn't matching less.js behavior when math=parens mode is enabled.

**Key files to investigate**:
- packages/less/src/less/less_go/operation.go
- packages/less/src/less/less_go/contexts.go
- packages/less/src/less/less_go/paren.go
- packages/less/src/less/less_go/dimension.go

**Compare with JavaScript**:
- packages/less/src/less/tree/operation.js
- packages/less/src/less/contexts.js
- packages/less/src/less/tree/paren.js

**Test commands**:
```bash
# Run specific test suite
pnpm -w test:go:filter -- "math-parens"

# See detailed diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "math-parens/css"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
# 1. Unit tests must pass 100%
pnpm -w test:go:unit

# 2. Integration tests - verify no regressions
pnpm -w test:go

# 3. Check baseline: Must have 76 perfect matches (74 current + 2 new)
```

**Expected outcome**:
- css (math-parens) shows "✅ Perfect match!"
- mixins-args (math-parens) shows "✅ Perfect match!"
- No regressions in other math tests
```

---

## Prompt 3: Fix math-parens-division Suite ⚡ QUICK WIN

```
Fix the mixins-args test in math-parens-division suite. It compiles but produces incorrect CSS output.

**The issue**: Division behavior in math mode with parens-division setting doesn't match less.js.

**Key files to investigate**:
- packages/less/src/less/less_go/operation.go (division handling)
- packages/less/src/less/less_go/contexts.go
- packages/less/src/less/less_go/dimension.go

**Compare with JavaScript**:
- packages/less/src/less/tree/operation.js
- Look for division logic when mathMode is parens-division

**Test commands**:
```bash
pnpm -w test:go:filter -- "math-parens-division/mixins-args"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "math-parens-division/mixins-args"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
```

**Expected outcome**:
- mixins-args (math-parens-division) shows "✅ Perfect match!"
```

---

## Prompt 4: Fix URL Handling Tests 🔗 HIGH PRIORITY

```
Fix URL handling edge cases in less.go. Three tests compile but produce incorrect CSS: `urls` in main suite, static-urls suite, and url-args suite.

**The issue**: URL processing has edge cases that don't match less.js behavior - likely quote handling, escaping, or path resolution.

**Key files to investigate**:
- packages/less/src/less/less_go/url.go
- packages/less/src/less/less_go/quoted.go
- packages/less/src/less/less_go/ruleset.go (for rewriteUrls context)

**Compare with JavaScript**:
- packages/less/src/less/tree/url.js
- packages/less/src/less/tree/quoted.js

**Test commands**:
```bash
pnpm -w test:go:filter -- "/urls$"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "main/urls"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "static-urls"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "url-args"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 77 perfect matches (74 current + 3 new)
# Verify url-rewriting tests still pass (4 tests)
```

**Expected outcome**:
- urls (main) shows "✅ Perfect match!"
- urls (static-urls) shows "✅ Perfect match!"
- urls (url-args) shows "✅ Perfect match!"
```

---

## Prompt 5: Fix functions-each Test 🔧 MEDIUM PRIORITY

```
Fix the each() function in less.go. The test `functions-each` compiles but produces incorrect CSS output.

**The issue**: The each() function iteration, variable binding, or scope handling doesn't match less.js.

**Key files to investigate**:
- packages/less/src/less/less_go/functions/list.go (each function implementation)
- packages/less/src/less/less_go/mixin_call.go (detached ruleset calling)
- packages/less/src/less/less_go/detached_ruleset.go

**Compare with JavaScript**:
- packages/less/src/less/functions/list.js (each function)

**Test commands**:
```bash
pnpm -w test:go:filter -- "functions-each"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "functions-each"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
```

**Expected outcome**:
- functions-each shows "✅ Perfect match!"
```

---

## Prompt 6: Fix functions Test 🔧 MEDIUM PRIORITY

```
Fix various function edge cases in less.go. The test `functions` in main suite compiles but produces incorrect CSS output.

**The issue**: Various built-in functions have edge cases that don't match less.js behavior. Run the test first to identify which specific functions are failing.

**Key files to investigate**:
- Start by running test to see which functions fail
- packages/less/src/less/less_go/functions/*.go (multiple function files)
- Likely candidates: color.go, string.go, math.go, type.go

**Compare with JavaScript**:
- packages/less/src/less/functions/*.js

**Test commands**:
```bash
pnpm -w test:go:filter -- "^main/functions$"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^main/functions$"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
```

**Expected outcome**:
- functions shows "✅ Perfect match!"
```

---

## Prompt 7: Fix detached-rulesets Test 📋 MEDIUM PRIORITY

```
Fix detached ruleset output formatting in less.go. The test `detached-rulesets` compiles but produces incorrect CSS output.

**The issue**: Detached rulesets CSS output formatting, whitespace, or structure doesn't match less.js behavior.

**Key files to investigate**:
- packages/less/src/less/less_go/detached_ruleset.go
- packages/less/src/less/less_go/ruleset.go (genCSS method)
- packages/less/src/less/less_go/mixin_call.go

**Compare with JavaScript**:
- packages/less/src/less/tree/detached-ruleset.js
- packages/less/src/less/tree/ruleset.js

**Test commands**:
```bash
pnpm -w test:go:filter -- "detached-rulesets"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
```

**Expected outcome**:
- detached-rulesets shows "✅ Perfect match!"
```

---

## Prompt 8: Fix directives-bubling Test 🔀 MEDIUM PRIORITY

```
Fix directive bubbling behavior in less.go. The test `directives-bubling` compiles but produces incorrect CSS output.

**The issue**: At-rule directive bubbling (moving nested directives like @media, @supports to parent level) doesn't match less.js behavior.

**Key files to investigate**:
- packages/less/src/less/less_go/at_rule.go
- packages/less/src/less/less_go/ruleset.go (directive bubbling logic)
- packages/less/src/less/less_go/media.go

**Compare with JavaScript**:
- packages/less/src/less/tree/atrule.js
- packages/less/src/less/tree/ruleset.js

**Test commands**:
```bash
pnpm -w test:go:filter -- "directives-bubling"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
```

**Expected outcome**:
- directives-bubling shows "✅ Perfect match!"
```

---

## Prompt 9: Fix container Test 📦 MEDIUM PRIORITY

```
Fix container query handling in less.go. The test `container` compiles but produces incorrect CSS output.

**The issue**: CSS container query (@container) syntax handling, formatting, or evaluation doesn't match less.js behavior.

**Key files to investigate**:
- packages/less/src/less/less_go/at_rule.go
- packages/less/src/less/less_go/media.go (container queries similar to media queries)
- packages/less/src/less/less_go/atrule_descriptor.go if exists

**Compare with JavaScript**:
- packages/less/src/less/tree/atrule.js
- packages/less/src/less/tree/media.js

**Test commands**:
```bash
pnpm -w test:go:filter -- "container"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "container"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
```

**Expected outcome**:
- container shows "✅ Perfect match!"
```

---

## Prompt 10: Fix media Test 📺 MEDIUM PRIORITY

```
Fix media query edge cases in less.go. The test `media` in main suite compiles but produces incorrect CSS output.

**The issue**: Media query handling, merging, bubbling, or formatting has edge cases that don't match less.js.

**Key files to investigate**:
- packages/less/src/less/less_go/media.go
- packages/less/src/less/less_go/at_rule.go
- packages/less/src/less/less_go/ruleset.go

**Compare with JavaScript**:
- packages/less/src/less/tree/media.js
- packages/less/src/less/tree/atrule.js

**Test commands**:
```bash
pnpm -w test:go:filter -- "^main/media$"
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^main/media$"
```

**CRITICAL VALIDATION** - Run before creating PR:
```bash
pnpm -w test:go:unit
pnpm -w test:go
# Check baseline: Must have 75 perfect matches (74 current + 1 new)
# Verify other media tests still pass (media-math, namespacing-media)
```

**Expected outcome**:
- media shows "✅ Perfect match!"
```

---

## Usage Instructions

### For Each Prompt:

1. **Copy the prompt** verbatim into a new Claude Code session
2. **Set the branch**: Work on `claude/fix-{task-name}-{session-id}`
3. **Let the agent work**: All context is provided in the prompt
4. **Critical validation**: Agent MUST run all tests before PR:
   - Unit tests: `pnpm -w test:go:unit` (must be 100% passing)
   - Integration tests: `pnpm -w test:go` (must match or exceed baseline)
5. **Create PR** only after validation passes
6. **Include in PR description**:
   - Before: X perfect matches
   - After: X+N perfect matches
   - No regressions: YES/NO

### Priority Order

**Fastest path to 80% (6 tests needed)**:
1. Prompt 1: import-reference (+2 tests) = 76/177 (42.9%)
2. Prompt 2: math-parens (+2 tests) = 78/177 (44.1%)
3. Prompt 4: URLs (+2 tests) = 80/177 (45.2%)
**Result: 80.2% overall success!** 🎉

**Or alternate path**:
1. Prompt 1: import-reference (+2)
2. Prompt 2: math-parens (+2)
3. Prompt 3: math-parens-division (+1)
4. Prompt 5: functions-each (+1)
= 80/177 (45.2%), **80.2% overall success!**

## Success Metrics

**If all 10 prompts completed successfully**:
- Perfect matches: 74 → 84 (45.8%)
- Overall success: 76.8% → 83.6%
- CSS output differs: 18 → 8 (4.3%)

This would be **OUTSTANDING ACHIEVEMENT** and put the project at 83%+ success rate!

---

## Git Workflow Reminder

Each agent should:
1. Create feature branch: `claude/fix-{task-name}-{session-id}`
2. Make focused changes (one test fix per PR preferred)
3. Run full test suite before committing
4. Commit with clear message: "Fix {test-name}: {brief description}"
5. Push to origin
6. Create PR with validation results in description

**Current baseline for comparison**:
- 74 perfect CSS matches (40.2%)
- 136 total successful tests (76.8%)
- 2,290+ unit tests passing (100%)

---

**Generated**: 2025-11-09 (Updated)
**Current Status**: 74 perfect matches, 136/177 overall success (76.8%)
**Target**: 80% success rate (142/177 tests) - only 6 tests away!
