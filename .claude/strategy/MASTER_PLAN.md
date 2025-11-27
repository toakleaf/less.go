# Master Strategy: Parallelized Test Fixing for less.go

## Current Status (Updated: 2025-11-27 - Verified Run)

### Test Results Summary (Verified)
- **Total Active Tests**: 184
- **Perfect CSS Matches**: 90 tests (48.9%)
- **Correct Error Handling**: 89 tests (48.4%)
- **Output Differs (but compiles)**: 2 tests (1.1%) - import-reference & import-reference-issues
- **Compilation Failures**: 3 tests (1.6%) - All external (network/packages)
- **Tests Passing or Correctly Erroring**: 179 tests (97.3%)
- **Overall Success Rate**: 97.3% (179/184)
- **Compilation Rate**: 98.4% (181/184)
- **Unit Tests**: 3,012 tests passing (100%)
- **Benchmarks**: ~111ms/op, ~38MB/op, ~600k allocs/op

### Parser Status
**ALL PARSER BUGS FIXED!** The parser correctly handles full LESS syntax. All remaining work is in **CSS output edge cases** for 2 import-reference tests.

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
- `import-module` - requires node_modules resolution (low priority)

### Phase 2: Output Differences - NEARLY COMPLETE! (Only 2 remaining)
**Impact**: Features work but produce incorrect output
**Location**: `.claude/tasks/runtime-failures/`

**Completed Categories** (14 categories at 100%):
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
14. **URLs (main)** - 1/1 test (JUST FIXED!)

**Remaining (2 tests)**:
1. **Import Reference** - 2 tests (`import-reference`, `import-reference-issues`)

### Phase 3: Polish & Edge Cases - NEARLY COMPLETE!
The remaining 2 tests are edge cases in import reference handling.

## Task Assignment System

### How to Claim a Task

1. Check `.claude/AGENT_WORK_QUEUE.md` for available tasks
2. Create a feature branch: `claude/fix-{task-name}-{session-id}`
3. Work on the task independently
4. Run tests to validate fix
5. Commit, push, and create PR

### Task States

- `available`: No one working on this task yet
- `in-progress`: Agent actively working
- `completed`: PR created and merged
- `blocked`: Depends on another task or has technical blockers

## Success Criteria

### For Individual Tasks

Each task must:
- Fix the specific test(s) identified in the task
- Pass all existing unit tests (no regressions)
- Not break any currently passing integration tests
- Include clear commit message explaining the fix
- Follow the porting process (never modify original JS code)

### Goals Progress

**Completed Goals**:
- [x] Reduce compilation failures from 5 to 0 (real bugs)
- [x] Increase success rate to 42%
- [x] Fix all guards and conditionals issues
- [x] Complete all namespacing fixes (11/11 tests)
- [x] Complete extend functionality fixes (7/7 tests)
- [x] Reach 50% success rate
- [x] Fix all math operations issues (12/12 tests)
- [x] Fix all URL rewriting issues (4/4 tests)
- [x] Reach 80% success rate
- [x] Reach 90% success rate
- [x] Reach 96% success rate

**Current Goal**:
- [ ] Fix remaining 2 output differences (import-reference, import-reference-issues)
- [ ] Reach 98.4% success rate (181/184 tests)

**Stretch Goals**:
- [ ] Implement quarantined features (plugins, JS execution)
- [ ] All 190+ tests passing

## Testing & Validation

### Required Test Commands

Before creating PR, agents must run:

```bash
# 1. All unit tests (must pass - no regressions allowed)
pnpm -w test:go:unit

# 2. Full integration suite summary (check overall impact)
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30

# 3. Specific test being fixed
LESS_GO_DIFF=1 pnpm -w test:go 2>&1 | grep -A 20 "test-name"
```

### Debug Tools Available

```bash
LESS_GO_TRACE=1   # Enhanced execution tracing with call stacks
LESS_GO_DEBUG=1   # Enhanced error reporting
LESS_GO_DIFF=1    # Visual CSS diffs
```

## Merge Conflict Prevention

### Strategies

1. **File-level isolation**: Each task focuses on specific Go files
2. **Test-level isolation**: Different tests use different code paths
3. **Category-based grouping**: Related fixes grouped to share context
4. **Clear ownership**: One agent per task at a time
5. **Frequent syncs**: Agents pull latest changes before starting
6. **Small PRs**: Fast review and merge cycle

### High-Risk Files (coordinate carefully)

These files are touched by many fixes:
- `ruleset.go` - Core ruleset evaluation
- `mixin_call.go` - Mixin resolution and calling
- `import.go` / `import_visitor.go` - Import handling
- `call.go` - Function calls

## Agent Onboarding

See `.claude/templates/AGENT_PROMPT.md` for the standard prompt to use when spinning up new agents.

## Project Structure Reference

```
less.go/
├── .claude/                    # Project coordination
│   ├── strategy/              # High-level strategy docs
│   ├── tasks/                 # Individual task specifications
│   ├── templates/             # Agent prompts and templates
│   └── tracking/              # Assignment tracking
├── packages/less/src/less/less_go/  # Go implementation (EDIT THESE)
├── packages/test-data/        # Test input/output (DON'T EDIT)
├── packages/less/src/less/    # Original JS (NEVER EDIT)
└── CLAUDE.md                  # Project overview for Claude
```

## Historical Context

### Progress Timeline

| Date | Perfect Matches | Success Rate | Notes |
|------|-----------------|--------------|-------|
| 2025-10-23 | 8 | 38.4% | Initial assessment |
| 2025-10-30 | 14 | 42.2% | Week 1 fixes |
| 2025-11-06 | 20 | 42.2% | Week 2 fixes |
| 2025-11-08 | 69 | 75.0% | Major breakthrough |
| 2025-11-09 | 69 | 75.0% | Stabilization |
| 2025-11-10 | 79 | 75.7% | Week 4 wins |
| 2025-11-13 | 83 | 93.0% | Continued progress |
| 2025-11-26 | 84 | 93.5% | Minor fix |
| **2025-11-27** | **90** | **97.3%** | **Current - urls fixed!** |

### Major Milestones

- **Week 1-2**: Fixed core evaluation issues (`if()`, type functions, detached rulesets)
- **Week 3**: MASSIVE BREAKTHROUGH - 69 perfect matches, all major categories fixed
- **Week 4**: Continued progress - 79 perfect matches
- **Week 5-6**: Polish and edge cases - 90 perfect matches
- **Current**: Only 2 output differences remaining (both import-reference)!

## Next Steps

1. **Fix import-reference** (2 tests) - See `.claude/tasks/runtime-failures/import-reference.md`
2. **Document completion** - Update all tracking when done

---

**Remember**: The goal is a faithful 1:1 port of less.js to Go. When in doubt, compare with the JavaScript implementation and match its behavior exactly.
