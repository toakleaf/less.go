# Agent Prompts - Priority Tasks (2025-11-07)

## Current Status Summary

**Test Results:**
- ✅ **72 perfect matches** (up from 30! +140% improvement!)
- ⚠️ **80 output differences** (down from ~106)
- ❌ **4 compilation failures** (network/path issues - expected)
- ⏸️ **5 quarantined** (plugin system - deferred)
- 🎯 **Overall: ~53% success rate** (97/184)

**Unit Tests:**
- ✅ All passing except TestMergeRulesTruthiness (3/6 sub-tests)

**Major Wins Since Last Update:**
- All 10 namespacing tests now passing! 🎉
- CSS guards and mixin guards working! 🎉
- All function tests passing! 🎉
- Mixins-named-args fixed! 🎉
- No regressions detected ✅

---

## 10 Ready-to-Assign Agent Prompts

### Prompt 1: Fix Math Operations in Parens Mode
**Priority:** HIGH | **Time:** 2-3 hours | **Impact:** 4+ tests

Fix math operations to respect parenthesis mode - operations should only evaluate inside parentheses. Currently `16 + 1` is evaluating to `17` in media queries when it should stay as `16 + 1`.

**Tests:** `math-parens/media-math`, `math-parens/parens`, `math-parens/css`, `math-parens/mixins-args`

**Starting point:** Check `operation.go` Eval() method - it needs to check `context.Math` mode and `context.InParens` flag before evaluating.

**Task file:** `.claude/tasks/output-differences/math-operations.md`

---

### Prompt 2: Fix Division in Parens-Division Mode
**Priority:** HIGH | **Time:** 1-2 hours | **Impact:** 3+ tests

Fix division behavior in parens-division mode - division should stay as `16 / 9` unless inside parentheses, when it should evaluate to `1.777...`.

**Tests:** `math-parens-division/media-math`, `math-parens-division/parens`, `math-parens-division/mixins-args`

**Starting point:** In `operation.go`, add special handling for `/` operator based on math mode.

**Task file:** `.claude/tasks/output-differences/math-operations.md`

---

### Prompt 3: Fix Extend-Chaining Edge Cases
**Priority:** MEDIUM | **Time:** 2-3 hours | **Impact:** 1 test

Fix extend chaining where one extend references another extend. The test compiles but CSS output is missing some extended selectors.

**Tests:** `extend-chaining`

**Starting point:** Check `extend_visitor.go` - ensure extended selectors that themselves have extends are processed recursively.

**Task file:** `.claude/tasks/output-differences/extend-functionality.md`

---

### Prompt 4: Fix Extend in Media Queries
**Priority:** MEDIUM | **Time:** 2-3 hours | **Impact:** 1 test

Fix extend functionality inside media queries - extra blank lines appearing and formatting incorrect.

**Tests:** `extend-media`

**Starting point:** Check how `extend_visitor.go` handles extends when inside `@media` blocks.

**Task file:** `.claude/tasks/output-differences/extend-functionality.md`

---

### Prompt 5: Fix Import Reference Functionality
**Priority:** HIGH | **Time:** 2-3 hours | **Impact:** 2 tests

Fix import reference so files imported with `@import (reference)` don't output CSS but their mixins/selectors are available when explicitly referenced.

**Tests:** `import-reference`, `import-reference-issues`

**Starting point:** Check `import.go` and `ruleset.go` - ensure reference flag is preserved and checked during GenCSS().

**Task file:** `.claude/tasks/runtime-failures/import-reference.md`

---

### Prompt 6: Fix Mixins-Guards Complex Cases
**Priority:** MEDIUM | **Time:** 1-2 hours | **Impact:** 1 test

Fix complex mixin guard cases in the main `mixins-guards` test. The simple guards work (`mixins-guards-default-func` passes), but the complex test has output differences.

**Tests:** `mixins-guards` (main suite)

**Starting point:** Compare expected vs actual output to identify which specific guard patterns are failing.

**Task file:** `.claude/tasks/output-differences/guards-conditionals.md`

---

### Prompt 7: Fix Comment Handling in Keyframes
**Priority:** MEDIUM | **Time:** 1-2 hours | **Impact:** 1 test

Fix comment placement in `@keyframes` - comments before keyframes are appearing after the closing brace instead of inside. Also affects empty rulesets with comments.

**Tests:** `comments2`

**Starting point:** Check `atrule.go` GenCSS() - comments in keyframes need special handling.

**Diff:** Expected comments inside keyframes, getting them after or outside.

---

### Prompt 8: Fix Variables in At-Rules
**Priority:** MEDIUM | **Time:** 2-3 hours | **Impact:** 1 test

Fix variable interpolation in at-rules like `@charset`, `@namespace`, `@keyframes`. Currently outputting raw interpolation syntax instead of evaluated values.

**Tests:** `variables-in-at-rules`

**Starting point:** Check `atrule.go` Eval() - ensure variables in name/value are evaluated before CSS generation.

**Example:** `@charset "UTF-@{Eight}"` should become `@charset "UTF-8"`

---

### Prompt 9: Fix Parse Interpolation Selector Issues
**Priority:** MEDIUM | **Time:** 2-3 hours | **Impact:** 1 test

Fix selector interpolation - selectors with interpolated parts are coming out wrong (missing pieces, wrong structure).

**Tests:** `parse-interpolation`

**Starting point:** Check `selector.go` and how interpolated expressions in selectors are evaluated and assembled.

**Example:** Multiple selectors being collapsed incorrectly.

---

### Prompt 10: Fix Nested Mixin Extra Output
**Priority:** MEDIUM | **Time:** 1-2 hours | **Impact:** 1 test

Fix mixins-nested test where an extra empty ruleset with wrong math appears in output. Should only output the properly nested and evaluated mixins.

**Tests:** `mixins-nested`

**Starting point:** Check `mixin_definition.go` - ensure nested mixins don't output intermediate evaluation steps.

**Task file:** `.claude/tasks/output-differences/mixin-issues.md`

---

## Notes for All Agents

### Before Starting:
1. Run `pnpm -w test:go:unit` to ensure unit tests pass
2. Run your specific test to see current output
3. Read the task file if referenced
4. Study the JavaScript implementation in `packages/less/src/less/tree/`

### Testing:
```bash
# See differences
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"

# Trace execution
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "test-name"
```

### Before Submitting:
1. ✅ Your specific test passes
2. ✅ Unit tests still pass: `pnpm -w test:go:unit`
3. ✅ No regressions: Check that other tests still pass
4. ✅ Commit with clear message explaining the fix

### Success Metrics:
- Each fix should move at least 1 test to "Perfect match!"
- No regressions allowed
- Unit tests must remain passing

---

## Already Completed (Do NOT work on these)

- ✅ All namespacing tests (1-8, functions, operations)
- ✅ CSS guards and mixin-guards-default-func
- ✅ Mixins-named-args
- ✅ All function tests
- ✅ Extend-selector
