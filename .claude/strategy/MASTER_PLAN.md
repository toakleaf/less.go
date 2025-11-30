# Master Strategy: Parallelized Test Fixing for less.go

## Current Status (Updated: 2025-11-30 - PROJECT COMPLETE! 🎉)

### Test Results Summary (Verified)
- **Total Active Tests**: 191
- **Perfect CSS Matches**: 100 tests (52.4%)
- **Correct Error Handling**: 91 tests (47.6%)
- **Output Differs**: 0 tests
- **Compilation Failures**: 0 tests
- **Overall Success Rate**: 100% (191/191) 🎉
- **Compilation Rate**: 100%
- **Quarantined Tests**: 0 (ALL FEATURES IMPLEMENTED!)
- **Unit Tests**: 3,012 tests passing (100%)
- **Performance**: Optimized with sync.Pool and regex caching

**Note**: Run `pnpm install` before tests to ensure npm dependencies are linked.

### Parser Status
**ALL PARSER BUGS FIXED!** The parser correctly handles full LESS syntax. Fixed `chunkInput` default to match JavaScript behavior (was causing parse errors with comments inside parentheses).

## Strategy Overview

This document outlines a strategy for **parallelizing the work** of fixing remaining test failures by enabling multiple independent AI agents to work on different issues simultaneously.

### Core Principles

1. **Independent Work Units**: Each task is self-contained with clear success criteria
2. **Minimal Human Intervention**: Agents pull repo, fix issues, test, and create PRs autonomously
3. **No Conflicts**: Tasks are designed to minimize merge conflicts
4. **Incremental Progress**: Small, focused changes that can be validated independently
5. **Clear Documentation**: All context needed for each task is provided

## Work Breakdown Structure

### Phase 1: Compilation Failures - COMPLETE! ✅
**Status**: ALL compilation failures fixed!

**✅ All Features Implemented** (2025-11-30):
- `javascript` - Inline JavaScript evaluation working!
- `js-type-errors/*` - JavaScript error handling tests PASSING!
- `no-js-errors/*` - Tests for `javascriptEnabled: false` PASSING!
- `plugin`, `plugin-module`, `plugin-preeval` - Full plugin system working!
- `bootstrap4` - Third-party test with JS plugins working!
- `import` - All import functionality working!

### Phase 2: Output Differences - COMPLETE! ✅
**Status**: ALL output differences fixed!

**Completed Categories** (ALL at 100%):
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
14. **URLs (main)** - 1/1 test
15. **Import Reference** - 2/2 tests (FIXED!)

### Phase 3: Polish & Edge Cases - COMPLETE! ✅
All active tests are now passing!

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
- [x] Fix import-reference tests (2/2)
- [x] **Reach 100% success rate (183/183 tests)** 🎉
- [x] **Implement JavaScript evaluation** 🎉 (inline `\`...\`` expressions)

**Stretch Goals** (ALL COMPLETE!):
- [x] Implement plugin system (bootstrap4 now working!)
- [x] Performance optimization (sync.Pool, regex caching)
- [x] Fix remaining edge cases (@arguments in complex mixins)

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
| 2025-11-27 | 90 | 97.3% | urls fixed |
| 2025-11-28 | 94 | 100.0% | ALL TESTS PASSING! (183 tests) |
| **2025-11-30** | **100** | **100%** | **🎉 Port Complete! JavaScript enabled (191 tests)** |

### Major Milestones

- **Week 1-2**: Fixed core evaluation issues (`if()`, type functions, detached rulesets)
- **Week 3**: MASSIVE BREAKTHROUGH - 69 perfect matches, all major categories fixed
- **Week 4**: Continued progress - 79 perfect matches
- **Week 5-6**: Polish and edge cases - 90 perfect matches
- **2025-11-28**: 🎉 **100% SUCCESS RATE ACHIEVED!** All 183 active tests passing!
- **2025-11-30**: 🎉 **PORT COMPLETE!** Core features fully implemented:
  - Inline JavaScript evaluation via Node.js runtime
  - Performance optimization (sync.Pool, regex caching)
  - 191 tests passing at 100% (100 perfect matches + 91 correct errors)

## Next Steps

### Completed Features
- ✅ Core LESS compilation (100% tests passing)
- ✅ JavaScript evaluation (inline `\`...\`` expressions working)
- ✅ Plugin system (custom functions, visitors, processors, file managers)
- ✅ All error handling validation
- ✅ All import functionality (including reference imports)
- ✅ Performance optimization (sync.Pool, regex caching)

### Project Status: COMPLETE 🎉
All features have been implemented. The less.go port is feature-complete with 1:1 parity to less.js!

---

**Remember**: The goal is a faithful 1:1 port of less.js to Go. When in doubt, compare with the JavaScript implementation and match its behavior exactly.
