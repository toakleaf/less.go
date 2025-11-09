# Agent Prompts for Parallel Work
**Generated**: 2025-11-09
**Session**: claude/assess-less-go-port-progress-011CUxTjGxoKqLyy1sTeRmAi

## ⚠️ CRITICAL: Start with Prompt #1

**There is a critical regression that must be fixed FIRST before any other work.**
See `.claude/CRITICAL_REGRESSION_REPORT.md` for full details.

---

## Prompt #1: 🚨 FIX CRITICAL COMBINATOR REGRESSION (URGENT)

**Priority**: CRITICAL - Must be done FIRST
**Time**: 30min - 1 hour
**Impact**: Restores 42 regressed tests
**Complexity**: Low

### Task

A recent change to `combinator.go` broke space combinator output, causing 42 tests to regress from passing to failing. The `GenCSS()` function incorrectly returns early for space combinators, preventing them from outputting a space character.

### Current Behavior
```css
/* Input: */
.parent .child { color: blue; }

/* Expected Output: */
.parent .child { color: blue; }

/* Actual Output: */
.parent.child { color: blue; }  /* Missing space! */
```

### Files to Modify
- `packages/less/src/less/less_go/combinator.go` (lines 111-133)

### Root Cause
Lines 128-130 in `combinator.go`:
```go
if c.Value == "" || (c.Value == " " && spaceOrEmpty == "") {
    return
}
```

This returns early for space combinators even when NOT in compress mode.

### Fix Strategy
1. Read `.claude/CRITICAL_REGRESSION_REPORT.md` for full analysis
2. Modify `GenCSS()` function to:
   - Empty combinator ("") → output nothing
   - Space combinator (" ") → output space (unless compress mode)
   - Other combinators (>, +, ~) → output with surrounding spaces
3. Run unit tests: `go test -v -run TestCombinator`
4. Run full unit test suite: `pnpm -w test:go:unit`
5. Run integration tests: `pnpm -w test:go`
6. Verify ~57 perfect matches restored

### Success Criteria
- ✅ All combinator unit tests pass
- ✅ All selector unit tests pass
- ✅ ~2290+ unit tests pass (99.9%+)
- ✅ ~57 integration tests show perfect matches
- ✅ Zero regressions from current state

### Branch Name
`claude/fix-combinator-regression-urgent-{session-id}`

---

## Prompt #2: Fix extend-chaining (Complete Extend Category)

**Priority**: HIGH (after regression fix)
**Time**: 2-3 hours
**Impact**: Completes extend category (7/7 tests)
**Complexity**: Medium

### Task

Implement multi-level extend chains. When A extends B, and B extends C, then A should also extend C (transitive extends).

### Current Status
- 6/7 extend tests passing ✅
- Only `extend-chaining` remains

### Files to Modify
- `packages/less/src/less/less_go/extend_visitor.go`

### Context
See `.claude/tasks/output-differences/extend-functionality.md` for full task specification.

### Test Command
```bash
pnpm -w test:go:filter -- "extend-chaining"
```

### Success Criteria
- ✅ extend-chaining test shows perfect match
- ✅ All 7 extend tests passing (100% category completion)
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-extend-chaining-{session-id}`

---

## Prompt #3: Fix Math Operations (6 tests)

**Priority**: HIGH
**Time**: 4-6 hours
**Impact**: 6 tests across math suites
**Complexity**: Medium-High

### Task

Fix math mode handling (strict, parens, parens-division, always). Tests compile but produce incorrect CSS output. The math operations don't respect the configured math mode properly.

### Current Status
- 6 tests compile but have output differences:
  - css (math-parens)
  - mixins-args (math-parens)
  - parens (math-parens)
  - mixins-args (math-parens-division)
  - parens (math-parens-division)
  - no-strict (units-no-strict)

### Files to Modify
- `packages/less/src/less/less_go/operation.go`
- `packages/less/src/less/less_go/contexts.go`
- `packages/less/src/less/less_go/dimension.go`

### Context
See `.claude/tasks/output-differences/math-operations.md` for full task specification.

### Test Commands
```bash
pnpm -w test:go:filter -- "math-parens"
pnpm -w test:go:filter -- "math-parens-division"
pnpm -w test:go:filter -- "units-no-strict"
```

### Success Criteria
- ✅ All 6 math tests show perfect matches
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-math-operations-{session-id}`

---

## Prompt #4: Fix URL Rewriting (7 tests)

**Priority**: HIGH
**Time**: 3-4 hours
**Impact**: 7 tests across URL suites
**Complexity**: Medium

### Task

Fix URL rewriting logic for different modes (rewrite-urls-all, rewrite-urls-local, rootpath variants). All URL tests compile but produce incorrect CSS output.

### Current Status
- 7 tests with output differences:
  - urls (main suite)
  - urls (static-urls suite)
  - urls (url-args suite)
  - rewrite-urls-all
  - rewrite-urls-local
  - rootpath-rewrite-urls-all
  - rootpath-rewrite-urls-local

### Files to Modify
- `packages/less/src/less/less_go/url.go`
- `packages/less/src/less/less_go/ruleset.go`

### Context
See `.claude/tasks/runtime-failures/url-processing.md` and `.claude/tasks/archived/url-processing-progress.md` for background.

### Test Commands
```bash
pnpm -w test:go:filter -- "urls"
pnpm -w test:go:filter -- "rewrite-urls"
pnpm -w test:go:filter -- "rootpath"
```

### Success Criteria
- ✅ All 7 URL tests show perfect matches
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-url-rewriting-{session-id}`

---

## Prompt #5: Fix Formatting and Comments (6 tests)

**Priority**: MEDIUM
**Time**: 3-4 hours
**Impact**: 6 tests
**Complexity**: Low-Medium

### Task

Fix CSS output formatting issues. Logic is correct but whitespace, line breaks, and comment placement differs from expected output.

### Current Status
- 6 tests with formatting differences:
  - comments
  - comments2
  - parse-interpolation
  - whitespace
  - container
  - directives-bubling

### Files to Modify
- `packages/less/src/less/less_go/comment.go`
- `packages/less/src/less/less_go/ruleset.go`
- `packages/less/src/less/less_go/atrule.go`
- `packages/less/src/less/less_go/to_css_visitor.go`

### Strategy
1. Compare actual vs expected output carefully
2. Identify patterns: extra blank lines? missing newlines? comment placement?
3. Fix CSS output generation to match less.js exactly
4. May be able to batch fix multiple tests with same root cause

### Test Commands
```bash
pnpm -w test:go:filter -- "comments"
pnpm -w test:go:filter -- "whitespace"
pnpm -w test:go:filter -- "parse-interpolation"
```

### Success Criteria
- ✅ All 6 formatting tests show perfect matches
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-formatting-output-{session-id}`

---

## Prompt #6: Fix Import Reference (2 tests)

**Priority**: MEDIUM
**Time**: 2-3 hours
**Impact**: 2 tests
**Complexity**: Medium

### Task

Fix mixin availability from referenced imports. When importing with `@import (reference)`, mixins should be available for use but CSS rules should not be included in output.

### Current Status
- 2 tests with output differences:
  - import-reference
  - import-reference-issues

### Files to Modify
- `packages/less/src/less/less_go/import.go`
- `packages/less/src/less/less_go/import_visitor.go`
- `packages/less/src/less/less_go/ruleset.go`

### Context
See `.claude/tasks/runtime-failures/import-reference.md` for full task specification.

### Test Commands
```bash
pnpm -w test:go:filter -- "import-reference"
```

### Success Criteria
- ✅ Both import-reference tests show perfect matches
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-import-reference-{session-id}`

---

## Prompt #7: Fix Detached Rulesets (1 test)

**Priority**: MEDIUM
**Time**: 2-3 hours
**Impact**: 1 test
**Complexity**: Medium

### Task

Fix detached ruleset output differences. Detached rulesets should capture their definition context and apply correctly when called.

### Current Status
- 1 test with output differences:
  - detached-rulesets

### Files to Modify
- `packages/less/src/less/less_go/detached_ruleset.go`
- `packages/less/src/less/less_go/variable_call.go`

### Test Commands
```bash
pnpm -w test:go:filter -- "detached-rulesets"
```

### Success Criteria
- ✅ detached-rulesets test shows perfect match
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-detached-rulesets-{session-id}`

---

## Prompt #8: Fix Function Gaps (2 tests)

**Priority**: MEDIUM
**Time**: 3-4 hours
**Impact**: 2 tests
**Complexity**: Low-Medium per function

### Task

Fix remaining function implementation gaps. Various built-in functions have bugs or missing features.

### Current Status
- 2 tests with output differences:
  - functions
  - functions-each

### Files to Modify
- Various function implementation files in `packages/less/src/less/less_go/`
- May need to add missing functions or fix existing ones

### Strategy
1. Run failing tests to see which specific functions fail
2. Compare with JavaScript implementation
3. Fix or implement missing functionality
4. Test thoroughly

### Test Commands
```bash
pnpm -w test:go:filter -- "functions"
pnpm -w test:go:filter -- "functions-each"
```

### Success Criteria
- ✅ Both function tests show perfect matches
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-function-gaps-{session-id}`

---

## Prompt #9: Fix Import Inline (1 test)

**Priority**: MEDIUM
**Time**: 1-2 hours
**Impact**: 1 test
**Complexity**: Low-Medium

### Task

Fix media query wrapper for inline CSS imports. When importing CSS files with `@import (inline)`, the content should respect media query wrappers.

### Current Status
- 1 test with output differences:
  - import-inline

### Files to Modify
- `packages/less/src/less/less_go/import.go`
- `packages/less/src/less/less_go/import_visitor.go`
- `packages/less/src/less/less_go/media.go`

### Context
See `.claude/tasks/IMPORT_INLINE_INVESTIGATION.md` for investigation notes.

### Test Commands
```bash
pnpm -w test:go:filter -- "import-inline"
```

### Success Criteria
- ✅ import-inline test shows perfect match
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-import-inline-{session-id}`

---

## Prompt #10: Fix CSS Output Issues (Multiple tests)

**Priority**: MEDIUM-LOW
**Time**: 4-6 hours
**Impact**: ~10 tests
**Complexity**: Variable

### Task

Fix various CSS output issues across multiple tests. These are tests that compile successfully but produce CSS that differs from expected output in various ways (not just formatting).

### Current Status
Tests with output differences in various categories:
- calc
- colors
- css-3
- css-escapes
- css-grid
- extract-and-length
- ie-filters
- media
- merge
- property-accessors
- property-name-interp
- selectors
- strings
- variables

### Strategy
1. Pick 2-3 related tests to work on together
2. Identify common patterns in output differences
3. Fix root causes systematically
4. Test thoroughly after each fix

### Test Commands
```bash
pnpm -w test:go:filter -- "{test-name}"
```

### Success Criteria
- ✅ At least 3-5 tests fixed to perfect matches
- ✅ No unit test regressions: `pnpm -w test:go:unit` passes
- ✅ No integration test regressions

### Branch Name
`claude/fix-css-output-issues-{session-id}`

---

## General Instructions for All Prompts

### Before Starting
1. ✅ Pull latest changes: `git pull origin main`
2. ✅ Checkout to working directory: `cd packages/less/src/less/less_go`
3. ✅ Read any referenced task files in `.claude/tasks/`
4. ✅ Review `.claude/VALIDATION_REQUIREMENTS.md`

### While Working
1. ⚠️ **CRITICAL**: Run tests frequently during development
2. 🔍 Use debug tools: `LESS_GO_TRACE=1 pnpm -w test:go:filter -- "{test}"`
3. 📝 Compare with JavaScript implementation when uncertain
4. 🎯 Focus on 1:1 functionality match with less.js

### Before Committing
1. ✅ **MANDATORY**: Run ALL unit tests: `pnpm -w test:go:unit` (must pass 100%)
2. ✅ **MANDATORY**: Run ALL integration tests: `pnpm -w test:go`
3. ✅ Verify target test(s) now pass
4. ✅ Verify no regressions in currently passing tests
5. ✅ Review changes for any unintended side effects

### Committing and Pushing
1. 📝 Write clear commit message explaining the fix
2. 🔧 Commit changes: `git add . && git commit -m "Fix: {description}"`
3. 📤 Push to branch: `git push -u origin {branch-name}`

### After Push
1. 🎉 Celebrate your contribution!
2. 📊 Document results (optional but appreciated)
3. 🔄 Move to next task if working on multiple issues

---

## Current Project Status (Before Regression)

### Baseline Metrics (Documented State)
- **Perfect Matches**: 57 tests (31.0%)
- **Compilation Failures**: 3 tests (expected - network/external)
- **Output Differences**: 35 tests (19.0%)
- **Correct Error Handling**: 39 tests (21.2%)
- **Overall Success Rate**: 52.7% (97/184 tests)
- **Compilation Rate**: 98.4% (181/184 tests)

### Current Metrics (With Regression)
- **Perfect Matches**: 15 tests (8.2%) ⚠️ REGRESSION
- **Compilation Failures**: 3 tests (expected)
- **Output Differences**: ~77 tests ⚠️ INCREASED
- **Correct Error Handling**: 39 tests (21.2%)
- **Overall Success Rate**: ~31.0% ⚠️ REGRESSION

### After Fixing Regression
Expected to return to documented baseline of 57 perfect matches.

---

## Priority Order

**Work in this order for maximum impact:**

1. **Prompt #1** - Fix combinator regression (CRITICAL - blocks everything)
2. **Prompt #2** - Fix extend-chaining (completes category)
3. **Prompt #3** - Fix math operations (6 tests)
4. **Prompt #4** - Fix URL rewriting (7 tests)
5. **Prompt #5** - Fix formatting (6 tests)
6. **Prompt #6** - Fix import-reference (2 tests)
7. **Prompt #7** - Fix detached rulesets (1 test)
8. **Prompt #8** - Fix function gaps (2 tests)
9. **Prompt #9** - Fix import-inline (1 test)
10. **Prompt #10** - Fix CSS output issues (multiple tests)

**If running agents in parallel**: Start with #1 FIRST, then parallelize #2-#6 after regression is fixed.

---

## Questions or Issues?

If you encounter any blockers or have questions:
1. Check existing documentation in `.claude/` directory
2. Review similar passing tests for patterns
3. Compare with JavaScript implementation
4. Document blocker clearly for human review if needed

**Remember**: All tests must pass before committing. Zero regression tolerance! 🎯
