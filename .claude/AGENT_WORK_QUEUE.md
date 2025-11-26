# Agent Work Queue - Ready for Assignment

**Updated**: 2025-11-26

## Summary

**Project is at 93.5% overall success rate** with **84 perfect CSS matches** and **zero regressions**.

## Current Test Status

- **Perfect Matches**: 84 tests (45.7%)
- **Correct Error Handling**: 88 tests (47.8%)
- **Output Differences**: 8 tests (4.3%)
- **Compilation Failures**: 3 tests (all expected - network/external)
- **Expected Error**: 1 test (javascript-undefined-var)
- **Overall Success Rate**: 93.5% (172/184 tests)
- **Compilation Rate**: 98.4% (181/184 tests)

## Categories Completed (100% Passing)

1. **Namespacing** - 11/11 tests
2. **Guards & Conditionals** - 3/3 tests
3. **Extend** - 7/7 tests
4. **Colors** - 2/2 tests
5. **Compression** - 1/1 test
6. **Math Operations** - 12/12 tests
7. **Units** - 2/2 tests
8. **URL Rewriting** - 4/4 tests
9. **Include Path** - 2/2 tests

---

## Available Tasks (8 Output Differences)

### 1. Import Reference - HIGH PRIORITY
**Tests**: 2 (import-reference, import-reference-issues)
**Time**: 2-3 hours
**Difficulty**: Medium
**Task File**: `.claude/tasks/output-differences/import-reference.md`

Files imported with `(reference)` option should not output CSS. Only explicitly used selectors/mixins should appear.

---

### 2. Detached Rulesets - HIGH PRIORITY
**Tests**: 1 (detached-rulesets)
**Time**: 2-3 hours
**Difficulty**: Medium-High
**Task File**: `.claude/tasks/output-differences/detached-rulesets.md`

Media queries nested in detached rulesets should merge with parent media context.

---

### 3. URLs - HIGH PRIORITY
**Tests**: 3 (urls in main/static-urls/url-args)
**Time**: 2-3 hours
**Difficulty**: Medium
**Task File**: `.claude/tasks/output-differences/urls.md`

URL handling edge cases in different contexts.

---

### 4. Media - MEDIUM PRIORITY
**Tests**: 1 (media)
**Time**: 1-2 hours
**Difficulty**: Medium
**Task File**: `.claude/tasks/output-differences/media.md`

Media query output formatting differences.

---

### 5. Container - MEDIUM PRIORITY
**Tests**: 1 (container)
**Time**: 2-3 hours
**Difficulty**: Medium
**Task File**: `.claude/tasks/output-differences/container.md`

@container query handling.

---

### 6. Directives Bubbling - MEDIUM PRIORITY
**Tests**: 1 (directives-bubling)
**Time**: 2-3 hours
**Difficulty**: Medium-High
**Task File**: `.claude/tasks/output-differences/directives-bubling.md`

Directive bubble order and selector grouping.

---

## Path to 50% Perfect Matches

**Current**: 84/184 (45.7%)
**Target**: 92/184 (50%)
**Needed**: +8 tests

If all 8 output differences are fixed:
- **Perfect Matches**: 84 → 92 tests (50%)
- **Output Differences**: 8 → 0 tests
- **Overall Success**: 93.5% → 97%+

---

## How to Claim a Task

1. Pick a task from above
2. Read the task file in `.claude/tasks/output-differences/`
3. Create branch: `claude/fix-{task-name}-{session-id}`
4. Make changes and test
5. **Run validation**:
   ```bash
   pnpm -w test:go:unit                    # Must pass 100%
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # Check no regressions
   ```
6. Commit and push
7. Create PR

---

## Debug Commands

```bash
# Test specific feature
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"

# Full debug
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"
```

---

## Baseline Metrics (2025-11-26)

**DO NOT REGRESS FROM THESE:**
- Unit tests: 2,304 passing
- Perfect matches: 84
- Error tests: 88/89 correct
- Compilation rate: 98.4%

---

**The project is in excellent shape! 8 tasks remaining to reach 50% perfect matches.**
