# Error Handling Tasks - Quick Start

## Summary

**Total Tests**: 27 tests that should fail but currently succeed
**Priority**: LOW (per CLAUDE.md)
**Impact**: Improves error handling and validation completeness

## Files in This Directory

1. **ERROR_VALIDATION_TASKS.md** - Detailed analysis of all 27 tests, organized by error type with complete context
2. **PARALLEL_TASK_ASSIGNMENTS.md** - Ready-to-copy prompts for assigning to independent LLM sessions
3. **README.md** (this file) - Quick start guide

## Quick Start

### Option 1: Assign All Tasks in Parallel (Fastest)

Open **PARALLEL_TASK_ASSIGNMENTS.md** and:
1. Copy Task 1 → Paste into LLM Session 1
2. Copy Task 2 → Paste into LLM Session 2
3. Copy Task 3 → Paste into LLM Session 3
... etc.

Total: 11 independent tasks

### Option 2: Prioritize by Impact

**High Impact** (10 tests):
- Task 1: Unit validation (4 tests)
- Task 2: SVG gradient validation (6 tests)

**Medium Impact** (11 tests):
- Task 3-8: Function and variable validation

**Low Impact** (6 tests):
- Task 9-11: Specific edge cases

### Option 3: Work Sequentially

Start with Task 1 and work through them in order.

## Current Status (2025-11-11)

```
⚠️  EXPECTED ERROR BUT SUCCEEDED (27 tests)

eval-errors (23 tests):
  - add-mixed-units
  - add-mixed-units2
  - color-func-invalid-color-2
  - color-func-invalid-color
  - detached-ruleset-1
  - detached-ruleset-2
  - divide-mixed-units
  - javascript-undefined-var
  - multiply-mixed-units
  - namespacing-2
  - namespacing-3
  - namespacing-4
  - percentage-non-number-argument
  - property-interp-not-defined
  - recursive-variable
  - root-func-undefined-1
  - svg-gradient1
  - svg-gradient2
  - svg-gradient3
  - svg-gradient4
  - svg-gradient5
  - svg-gradient6
  - unit-function

parse-errors (4 tests):
  - invalid-color-with-comment
  - parens-error-1
  - parens-error-2
  - parens-error-3
```

## Testing Commands

**Check current status:**
```bash
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

**Run all error tests:**
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/(eval-errors|parse-errors)"
```

**Run specific test:**
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/eval-errors/add-mixed-units"
```

## Success Criteria

- ⚠️  "EXPECTED ERROR BUT SUCCEEDED" count: 27 → 0
- ✅ "CORRECTLY FAILED" count: 62 → 89
- ✅ "Perfect CSS Matches": Stay at 80 (no regressions)
- ✅ All unit tests passing

## Notes

- These tests validate that invalid LESS code properly throws errors
- Each test has a .less file (invalid input) and .txt file (expected error)
- Test data location: `/home/user/less.go/packages/test-data/errors/`
- All work should be on branch: `claude/incomplete-description-011CV2rK88Fuho4UdMRdRz9X`
