# Master Strategy: Parallelized Test Fixing for less.go

## Current Status (Updated: 2025-11-26)

### Test Results Summary
- **Total Active Tests**: 184
- **Perfect CSS Matches**: 84 tests (45.7%)
- **Correct Error Handling**: 88 tests (47.8%)
- **Output Differences**: 8 tests (4.3%)
- **Compilation Failures**: 3 tests (1.6%) - All external (network/packages)
- **Expected Error Tests**: 1 test (javascript-undefined-var - JS execution quarantined)
- **Tests Passing or Correctly Erroring**: 172 tests (93.5%)
- **Overall Success Rate**: 93.5%
- **Compilation Rate**: 98.4% (181/184 tests)

### Parser Status
**ALL PARSER BUGS FIXED!** The parser correctly handles full LESS syntax. All remaining work is in **CSS output generation** and **feature edge cases**.

## Strategy Overview

This document outlines a strategy for **parallelizing the work** of fixing remaining test failures by enabling multiple independent AI agents to work on different issues simultaneously.

### Core Principles

1. **Independent Work Units**: Each task is self-contained with clear success criteria
2. **Minimal Human Intervention**: Agents pull repo, fix issues, test, and create PRs autonomously
3. **No Conflicts**: Tasks are designed to minimize merge conflicts
4. **Incremental Progress**: Small, focused changes that can be validated independently
5. **Clear Documentation**: All context needed for each task is provided

## Work Breakdown Structure

### Phase 1: Compilation Failures - COMPLETE!
**Status**: ALL real compilation failures fixed!

**Remaining expected failures** (infrastructure/external, not bugs):
- `bootstrap4` - requires external bootstrap dependency
- `google` - requires network access to Google Fonts
- `import-module` - requires node_modules resolution

### Phase 2: Output Differences - 8 TESTS REMAINING
**Impact**: Features work but produce incorrect output
**Location**: `.claude/tasks/output-differences/`

**Remaining Tasks** (8 tests):
1. **Import Reference** (2 tests) - `import-reference`, `import-reference-issues`
2. **Detached Rulesets** (1 test) - media query merging in detached rulesets
3. **URLs** (3 tests) - urls in main/static-urls/url-args suites
4. **Media** (1 test) - media query formatting
5. **Container** (1 test) - @container query handling
6. **Directives Bubbling** (1 test) - directive bubble order

**Completed Categories** (100% Passing):
1. **Namespacing** - 11/11 tests
2. **Guards & Conditionals** - 3/3 tests
3. **Extend** - 7/7 tests
4. **Colors** - 2/2 tests
5. **Compression** - 1/1 test
6. **Math Operations** - 12/12 tests
7. **Units** - 2/2 tests
8. **URL Rewriting** - 4/4 tests
9. **Include Path** - 2/2 tests

### Phase 3: Error Handling - NEARLY COMPLETE
**Status**: 88/89 error tests correct (98.9%)
**Remaining**: `javascript-undefined-var` (JS execution is quarantined)

## Task Assignment System

### How to Claim a Task

1. Check `.claude/tasks/output-differences/` for available tasks
2. Create a feature branch: `claude/fix-{task-name}-{session-id}`
3. Work on the task independently
4. Run tests to validate fix
5. Commit, push, and create PR

## Success Criteria

### For Individual Tasks

Each task must:
- Fix the specific test(s) identified in the task
- Pass all existing unit tests (no regressions)
- Not break any currently passing integration tests
- Include clear commit message explaining the fix
- Follow the porting process (never modify original JS code)

### Goals

**Immediate** (8 remaining output differences):
- [ ] Fix import-reference (2 tests)
- [ ] Fix detached-rulesets (1 test)
- [ ] Fix urls (3 tests)
- [ ] Fix media (1 test)
- [ ] Fix container (1 test)
- [ ] Fix directives-bubling (1 test)

**Stretch** (reach 50% perfect matches):
- Need +8 tests (84 → 92 perfect matches = 50%)

## Testing & Validation

### Required Test Commands

Before creating PR, agents must run:

```bash
# 1. All unit tests (must pass - no regressions allowed)
pnpm -w test:go:unit

# 2. Specific test being fixed (must show improvement)
pnpm -w test:go:filter -- "test-name"

# 3. Full integration suite summary (check overall impact)
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30
```

### Debug Tools Available

```bash
LESS_GO_TRACE=1   # Enhanced execution tracing with call stacks
LESS_GO_DEBUG=1   # Enhanced error reporting
LESS_GO_DIFF=1    # Visual CSS diffs
```

## Project Structure Reference

```
less.go/
├── .claude/                    # Project coordination
│   ├── strategy/              # High-level strategy docs
│   ├── tasks/                 # Individual task specifications
│   │   └── output-differences/  # 8 remaining output diff tasks
│   ├── templates/             # Agent prompts and templates
│   └── archived/              # Completed/outdated task files
├── packages/less/src/less/less_go/  # Go implementation (EDIT THESE)
├── packages/test-data/        # Test input/output (DON'T EDIT)
└── packages/less/src/less/    # Original JS (NEVER EDIT)
```

## Historical Progress Summary

- **Weeks 1-4**: Parser fixes, runtime evaluation, mixin handling
- **Week 5**: All namespacing, guards, extend, colors complete
- **Week 6**: Math operations, URL rewriting, units complete
- **Week 7+**: Error handling improvements, reached 84 perfect matches

**ZERO REGRESSIONS** maintained throughout all progress.

---

**Remember**: The goal is a faithful 1:1 port of less.js to Go. When in doubt, compare with the JavaScript implementation and match its behavior exactly.
