# Agent Work Queue - Ready for Assignment

**Updated**: 2025-11-27
**Status**: Only 2 output differences remaining!

## Summary

**EXCELLENT PROGRESS!** The project has reached **97.3% overall success rate** with **90 perfect CSS matches**.

## Current Test Status

- **Perfect CSS Matches**: 90 tests (48.9%)
- **Output Differences**: 2 tests (1.1%)
- **Compilation Failures**: 3 tests (all expected - network/external dependencies)
- **Correct Error Handling**: 89 tests (48.4%)
- **Overall Success Rate**: 97.3% (179/184 tests)
- **Compilation Rate**: 98.4% (181/184 tests)
- **Unit Tests**: 3,012 tests passing (100%)

## Remaining Work - Only 2 Output Differences!

### HIGH PRIORITY: Import Reference (2 tests)

1. **import-reference**
   - Reference imports outputting CSS when they shouldn't
   - See `.claude/tasks/runtime-failures/import-reference.md`

2. **import-reference-issues**
   - Import reference with extends/mixins not working correctly
   - Related to import-reference fix

---

## Categories 100% Complete

1. **Namespacing** - 11/11 tests
2. **Guards & Conditionals** - 3/3 tests
3. **Extend** - 7/7 tests
4. **Colors** - 2/2 tests
5. **Compression** - 1/1 test
6. **Math Operations** - 12/12 tests (all variants)
7. **Units** - 2/2 tests (strict and non-strict)
8. **URL Rewriting** - 4/4 tests (rewrite-urls, url-args)
9. **Include Path** - 2/2 tests
10. **Detached Rulesets** - 1/1 test
11. **Media Queries** - 1/1 test
12. **Container Queries** - 1/1 test
13. **Directives Bubbling** - 1/1 test
14. **URLs** - 3/3 tests (JUST FIXED!)

---

## Task Details

### Task 1: Fix Import Reference (HIGH PRIORITY)
**Impact**: +2 tests (import-reference, import-reference-issues)
**Time**: 2-3 hours
**Difficulty**: Medium

Files imported with `(reference)` option should not output CSS but selectors/mixins should be available for extends/mixin calls.

**Key Files**:
- `import.go`, `import_visitor.go`, `ruleset.go`
- See `.claude/tasks/runtime-failures/import-reference.md` for full details

**Debugging**:
```bash
LESS_GO_DIFF=1 pnpm -w test:go 2>&1 | grep -A 20 "import-reference"
```

---

## External Dependencies (Expected Failures)

These 3 tests fail due to infrastructure, not bugs:

1. **bootstrap4** - External bootstrap package not available
2. **google** - Network access to Google Fonts required
3. **import-module** - Node modules resolution not implemented

---

## How to Claim a Task

1. Create branch: `claude/fix-{task-name}-{session-id}`
2. Read task file in `.claude/tasks/runtime-failures/`
3. Make changes, test incrementally
4. **CRITICAL**: Run ALL tests before PR:
   ```bash
   pnpm -w test:go:unit    # Must pass 100% (3,012 tests)
   pnpm -w test:go         # Must show >= 89 perfect matches
   ```
5. Commit and push
6. Create PR with clear description

---

## Validation Checklist

### Before Starting
```bash
pnpm -w test:go:unit          # Baseline: 3,012 tests passing
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # Baseline: 89 perfect
```

### After Fixing
```bash
pnpm -w test:go:unit          # MUST: Still 3,012 passing (NO REGRESSIONS)
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # MUST: >= 89 perfect
```

---

## Path to 100% Success

**Current**: 97.3% (179/184 tests perfect or correctly erroring)
**Target**: Fix remaining 2 output differences → 98.4% (181/184 tests)

Only the 3 external dependency tests would remain as expected failures.

---

## Recent Accomplishments

### 2025-11-27 (Current)
- 90 perfect CSS matches (up from 84!)
- Only 2 output differences remaining (down from 8!)
- urls (main suite) - JUST FIXED!
- detached-rulesets, media, container, directives-bubbling - ALL FIXED!
- static-urls, url-args - FIXED!

### Previous Progress
- 84 perfect matches (2025-11-26)
- 83 perfect matches (2025-11-13)
- 79 perfect matches (2025-11-10)
- 69 perfect matches (2025-11-09)

---

**The project is in OUTSTANDING shape! 97.3% success rate with only 2 import-reference tests remaining!**
