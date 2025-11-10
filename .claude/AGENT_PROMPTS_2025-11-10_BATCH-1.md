# Independent Agent Prompts - Batch 1
**Generated**: 2025-11-10
**Session**: claude/assess-less-go-port-progress-011CUz7KxhVPs73Xpoz34U2b
**Baseline**: 78 perfect matches, 14 output differences, 140/184 tests passing (76.1%)

---

## ⚠️ CRITICAL: Regression Testing Required

Before creating ANY pull request, you MUST:

```bash
# 1. Run ALL unit tests (must pass 100%)
pnpm -w test:go:unit

# 2. Run ALL integration tests and verify counts
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match"
# Expected: 78+ (must not decrease!)

# 3. Run specific test you're fixing
pnpm -w test:go:filter -- "your-test-name"
# Expected: ✅ Perfect match!
```

**Zero tolerance for regressions** - If any previously passing test breaks, fix it before creating PR.

---

## Prompt 1: Fix import-reference Tests (HIGH PRIORITY)

**Task**: Fix the import-reference functionality so that files imported with `(reference)` option are handled correctly.

**Context**: Files imported with `@import (reference) "file.less";` should NOT output CSS, but their selectors/mixins should be available for use. Currently these tests compile but produce incorrect CSS output.

**Failing Tests**:
- `import-reference` (main suite)
- `import-reference-issues` (main suite)

**Expected Impact**: +2 perfect matches → 80/184 tests (43.5%)

**Files to Investigate**:
- `packages/less/src/less/less_go/import.go`
- `packages/less/src/less/less_go/import_visitor.go`
- `packages/less/src/less/less_go/ruleset.go`

**Debugging Commands**:
```bash
# See actual vs expected output
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"

# Trace execution
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "import-reference"
```

**Task File**: `.claude/tasks/runtime-failures/import-reference.md`

**Estimated Time**: 2-3 hours
**Difficulty**: Medium
**Priority**: HIGH - Core functionality

**Success Criteria**:
- ✅ import-reference: Perfect match!
- ✅ import-reference-issues: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "import-reference" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 80+
```

---

## Prompt 2: Fix extract() and length() Functions

**Task**: Fix the `extract()` and `length()` built-in functions to properly handle list operations.

**Context**: The extract-and-length test compiles but produces incorrect CSS output. These are core list manipulation functions used frequently in LESS codebases.

**Failing Tests**:
- `extract-and-length` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/functions/list.go`
- `packages/less/src/less/less_go/call.go`
- Compare with `packages/less/src/less/functions/list.js`

**Debugging Commands**:
```bash
# See actual vs expected output
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "extract-and-length"

# Check test input/output files
cat packages/test-data/less/_main/extract-and-length.less
cat packages/test-data/css/_main/extract-and-length.css
```

**Estimated Time**: 2-3 hours
**Difficulty**: Medium
**Priority**: HIGH - Core functionality

**Success Criteria**:
- ✅ extract-and-length: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "extract-and-length" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 3: Fix functions Test (Function Edge Cases)

**Task**: Fix various function implementation edge cases that are causing the `functions` test to produce incorrect CSS output.

**Context**: The functions test exercises many built-in functions with edge cases. Most functions work, but some edge cases produce incorrect output.

**Failing Tests**:
- `functions` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/functions/*.go` (various function files)
- `packages/less/src/less/less_go/call.go`
- Compare with `packages/less/src/less/functions/*.js`

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^functions$"

# Trace function calls
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "^functions$"
```

**Estimated Time**: 3-4 hours
**Difficulty**: Medium-High
**Priority**: HIGH - Core functionality

**Success Criteria**:
- ✅ functions: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "^functions$" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 4: Fix Detached Rulesets Output

**Task**: Fix the CSS output formatting for detached rulesets to match less.js exactly.

**Context**: Detached rulesets compile and work correctly, but the CSS output formatting differs from less.js. This is likely a whitespace/indentation issue.

**Failing Tests**:
- `detached-rulesets` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/detached_ruleset.go`
- `packages/less/src/less/less_go/ruleset.go` (genCSS methods)
- Compare with `packages/less/src/less/tree/detached-ruleset.js`

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"

# Check the test files
cat packages/test-data/less/_main/detached-rulesets.less
cat packages/test-data/css/_main/detached-rulesets.css
```

**Estimated Time**: 1-2 hours
**Difficulty**: Low-Medium
**Priority**: MEDIUM - Formatting issue

**Success Criteria**:
- ✅ detached-rulesets: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "detached-rulesets" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 5: Fix Directive Bubbling Output

**Task**: Fix the CSS output formatting for directive bubbling (@supports, @document) to match less.js.

**Context**: Directive bubbling logic works but the CSS output has formatting differences. This affects how nested directives are output.

**Failing Tests**:
- `directives-bubling` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/ruleset.go`
- `packages/less/src/less/less_go/directive.go` (if exists)
- Compare with `packages/less/src/less/tree/atrule.js`

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"

# Check the test files
cat packages/test-data/less/_main/directives-bubling.less
cat packages/test-data/css/_main/directives-bubling.css
```

**Estimated Time**: 1-2 hours
**Difficulty**: Low-Medium
**Priority**: MEDIUM - Formatting issue

**Success Criteria**:
- ✅ directives-bubling: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "directives-bubling" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 6: Fix Container Query Output

**Task**: Fix the CSS output formatting for container queries to match less.js exactly.

**Context**: Container query logic works but CSS output has formatting differences. This is a CSS3 feature.

**Failing Tests**:
- `container` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/media.go` (container queries may share logic with media queries)
- Compare with `packages/less/src/less/tree/atrule.js`

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^container$"

# Check the test files
cat packages/test-data/less/_main/container.less
cat packages/test-data/css/_main/container.css
```

**Estimated Time**: 1-2 hours
**Difficulty**: Low-Medium
**Priority**: MEDIUM - CSS3 feature

**Success Criteria**:
- ✅ container: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "^container$" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 7: Fix CSS-3 Feature Output

**Task**: Fix the CSS output formatting for CSS3 features to match less.js exactly.

**Context**: CSS3 features compile correctly but output has formatting differences. May involve multiple CSS3 features.

**Failing Tests**:
- `css-3` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/ruleset.go` (genCSS methods)
- `packages/less/src/less/less_go/declaration.go`
- Various node genCSS methods
- Compare with corresponding JS files

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^css-3$"

# Check the test files
cat packages/test-data/less/_main/css-3.less
cat packages/test-data/css/_main/css-3.css
```

**Estimated Time**: 2-3 hours
**Difficulty**: Medium
**Priority**: MEDIUM - CSS3 features

**Success Criteria**:
- ✅ css-3: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "^css-3$" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 8: Fix Media Query Output Formatting

**Task**: Fix the CSS output formatting for media queries to match less.js exactly.

**Context**: Media query logic works but CSS output has formatting differences, likely related to nested media queries or media query bubbling.

**Failing Tests**:
- `media` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/media.go`
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/ruleset.go`
- Compare with `packages/less/src/less/tree/media.js`

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "^media$"

# Check the test files
cat packages/test-data/less/_main/media.less
cat packages/test-data/css/_main/media.css
```

**Estimated Time**: 1-2 hours
**Difficulty**: Low-Medium
**Priority**: MEDIUM - Formatting issue

**Success Criteria**:
- ✅ media: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "^media$" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Prompt 9: Fix URL Processing Edge Cases (3 tests)

**Task**: Fix URL processing edge cases across three different test suites that all test URL handling variations.

**Context**: Most URL processing works (4/7 URL tests pass), but three tests have edge cases causing output differences.

**Failing Tests**:
- `urls` (main suite)
- `urls` (static-urls suite)
- `urls` (url-args suite)

**Expected Impact**: +3 perfect matches → 81/184 tests (44.0%)

**Files to Investigate**:
- `packages/less/src/less/less_go/url.go`
- `packages/less/src/less/less_go/ruleset.go`
- Compare with `packages/less/src/less/tree/url.js`

**Debugging Commands**:
```bash
# See all URL test differences
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "urls"

# Check the test files
cat packages/test-data/less/_main/urls.less
cat packages/test-data/less/static-urls/urls.less
cat packages/test-data/less/url-args/urls.less
```

**Estimated Time**: 2-3 hours
**Difficulty**: Medium
**Priority**: MEDIUM - Edge cases

**Success Criteria**:
- ✅ urls (main): Perfect match!
- ✅ urls (static-urls): Perfect match!
- ✅ urls (url-args): Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "urls" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 81+
```

---

## Prompt 10: Fix Comments2 Keyframe Output

**Task**: Fix the comment placement in @keyframes rules to match less.js exactly.

**Context**: The comments2 test has an issue with comment placement in @keyframes/@-webkit-keyframes rules. The main comments test passes, but this edge case doesn't.

**Failing Tests**:
- `comments2` (main suite)

**Expected Impact**: +1 perfect match → 79/184 tests (42.9%)

**Files to Investigate**:
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/comment.go` (if exists)
- `packages/less/src/less/less_go/anonymous.go` (comments may be stored as anonymous nodes)
- Compare with `packages/less/src/less/tree/atrule.js` and `packages/less/src/less/tree/comment.js`

**Debugging Commands**:
```bash
# See actual vs expected output with diff
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "comments2"

# Check the test files
cat packages/test-data/less/_main/comments2.less
cat packages/test-data/css/_main/comments2.css
```

**Estimated Time**: 1-2 hours
**Difficulty**: Low
**Priority**: LOW - Edge case

**Success Criteria**:
- ✅ comments2: Perfect match!
- ✅ All unit tests pass
- ✅ No regressions (78+ perfect matches maintained)

**Validation**:
```bash
pnpm -w test:go:unit && \
pnpm -w test:go:filter -- "comments2" && \
pnpm -w test:go 2>&1 | grep -c "✅.*Perfect match" # Should be 79+
```

---

## Summary: Impact of Completing All 10 Prompts

If all 10 prompts are completed successfully:

| Metric | Current | After All | Change |
|--------|---------|-----------|--------|
| Perfect Matches | 78 (42.4%) | 92 (50.0%) | +14 tests 🎉 |
| Output Differences | 14 (7.6%) | 0 (0%) | -14 tests 🎉 |
| Overall Success Rate | 76.1% | 83.7% | +7.6% 🎉 |

**Target Achievement**: All 10 prompts would bring us to **83.7% success rate** and **50% perfect CSS match rate**!

---

## Coordination Notes

- **All prompts are independent** - Can be worked on in parallel
- **No merge conflicts expected** - Each touches different files
- **Update `.claude/tracking/assignments.json`** when claiming a task
- **Create branch**: `claude/fix-{test-name}-{session-id}`
- **Sync frequently**: Pull latest changes before starting
- **Zero tolerance for regressions**: Must maintain 78+ perfect matches

---

**Generated by**: Assessment Agent (Session 011CUz7KxhVPs73Xpoz34U2b)
**Last Updated**: 2025-11-10
**Baseline**: 78 perfect matches, 76.1% success rate
