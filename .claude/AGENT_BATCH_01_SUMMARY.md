# Agent Batch 01 - Ready for Assignment

**Date**: 2025-11-08
**Current Status**: 42+ perfect matches (59.2% success rate)
**Session ID**: Assessment by human maintainer

## Summary

The less.go port is in **excellent shape** with 59.2% of tests passing perfectly! Recent work has fixed:
- ✅ All namespace variable issues (10 tests)
- ✅ All guard/condition evaluation (3 tests)
- ✅ Mixin argument matching (enabling math suites)
- ✅ Include path resolution
- ✅ Plus many other improvements

## 10 Independent Tasks Ready for Parallel Work

All 10 tasks are completely independent and can be worked in parallel by different agents.

### Group A: Parser Fixes (CRITICAL) - 5 Tasks

These are high-impact parser issues preventing tests from even compiling. Fixing these unblocks many other tests.

1. **`fix-parse-functions-each`** - `/agent-batch-01-fix-parse-failures.md`
   - Impact: 1 test
   - Difficulty: Medium
   - Time: 2-3 hours
   - Current: Parse error - Unrecognised input

2. **`fix-parse-selectors`** - `/agent-batch-02-fix-parse-selectors.md`
   - Impact: 1 test (CSS selector syntax)
   - Difficulty: Medium
   - Time: 2-3 hours
   - Current: Parse error - Unrecognised input

3. **`fix-parse-variables`** - `/agent-batch-03-fix-parse-variables.md`
   - Impact: 1 test (core feature!)
   - Difficulty: Medium
   - Time: 2-3 hours
   - Current: Parse error - Unrecognised input
   - ⚠️ HIGH RISK: Variables are everywhere - test thoroughly!

4. **`fix-parse-mixins-interpolated`** - `/agent-batch-04-fix-parse-mixins-interpolated.md`
   - Impact: 1 test (advanced mixin feature)
   - Difficulty: Medium-High
   - Time: 2-3 hours
   - Current: Parse error - Unrecognised input

5. **`fix-parse-import-remote`** - (Not documented, but mentioned)
   - Remote imports requiring network (lower priority)

### Group B: Output Differences (HIGH IMPACT) - 5 Tasks

These tests compile fine but CSS output differs. Each is independent and can yield quick wins.

6. **`fix-math-operations`** - `/agent-batch-05-fix-math-operations.md`
   - Impact: 10+ tests across multiple suites
   - Difficulty: Medium-High
   - Time: 3-4 hours
   - Current: Compiles, output differs
   - Status: UNBLOCKED (mixin fixes cleared path)

7. **`fix-formatting-output`** - `/agent-batch-06-fix-formatting-output.md`
   - Impact: 6+ tests
   - Difficulty: LOW ⭐ QUICK WINS!
   - Time: 2-3 hours
   - Current: Whitespace/formatting issues only
   - Tests: comments, whitespace, variables-in-at-rules, charsets, parse-interpolation
   - Note: comments2 already works - use as reference!

8. **`complete-mixin-issues`** - `/agent-batch-07-complete-mixin-issues.md`
   - Impact: 2 tests
   - Difficulty: Medium
   - Time: 1-2 hours
   - Current: Output differs
   - Tests: mixins-nested (nested calls), mixins-important (!important propagation)

9. **`fix-color-functions`** - `/agent-batch-08-fix-color-functions.md`
   - Impact: 1 test
   - Difficulty: Medium
   - Time: 1-2 hours
   - Current: Output differs
   - Note: colors2 already fixed - use as reference!

10. **`complete-import-reference`** - `/agent-batch-09-complete-import-reference.md`
    - Impact: 2 tests
    - Difficulty: Medium
    - Time: 1-2 hours
    - Current: Compiles but output differs
    - Status: 80% done, needs final fix
    - Issue: Mixins from referenced imports not available

## Bonus Task

**`fix-remaining-output-diffs`** - `/agent-batch-10-fix-remaining-output-diffs.md`
- If agents finish early, batch fix 6 more tests
- Tests: extend-chaining, mixins-guards, import-inline, detached-rulesets, directives-bubling, media
- Impact: 6+ more tests if completed

## Assignment Recommendations

### Conservative Approach (1-2 Agents)
```
Agent 1: fix-parse-functions-each → fix-parse-selectors
Agent 2: fix-math-operations
```

### Aggressive Approach (10 Agents in Parallel)
```
Assign all 10 tasks in parallel:
- 5 agents on parser fixes
- 5 agents on output differences
- Expected impact: 15-20 new perfect matches
- Target: 57-62 perfect matches total (80%+ success rate!)
```

### Balanced Approach (4-6 Agents)
```
Agent 1: fix-formatting-output (quick wins, 6 tests)
Agent 2: fix-parse-functions-each
Agent 3: fix-parse-variables
Agent 4: fix-math-operations
Agent 5: complete-import-reference (80% done)
Agent 6: complete-mixin-issues + fix-color-functions
```

## Expected Outcomes

### Parser Fixes (Group A)
- **If all 4 parse issues fixed**: +4 new perfect matches
- **Potential cascade effect**: May unblock other tests
- **Total impact**: 5-8 tests

### Output Differences (Group B)
- **If all 5 completed**: +17 new perfect matches
  - 6 from formatting
  - 2 from mixin issues
  - 1 from color functions
  - 2 from import reference
  - 10+ from math operations

### Total Possible Improvement
- Current: 42 perfect matches (59.2%)
- All tasks completed: ~65 perfect matches (91.5%)
- Conservative (3 tasks): 50 perfect matches (70.4%)

## Critical Notes for Agents

### Before Starting
1. ✅ Check the specific prompt file for your assigned task
2. ✅ Read `.claude/VALIDATION_REQUIREMENTS.md` for testing requirements
3. ✅ Read `.claude/strategy/agent-workflow.md` for workflow
4. ✅ Check `.claude/tracking/assignments.json` to claim your task

### During Work
1. ✅ Commit frequently with clear messages
2. ✅ Test EVERY change: `pnpm -w test:go:unit`
3. ✅ Run full suite before pushing: `pnpm -w test:go`
4. ✅ Use LESS_GO_DIFF=1 to see output differences
5. ✅ Use LESS_GO_TRACE=1 to debug parser issues

### Before Pushing
1. ✅ Run all unit tests: `pnpm -w test:go:unit` (MUST PASS)
2. ✅ Run full integration suite: `pnpm -w test:go` (CHECK FOR REGRESSIONS)
3. ✅ Verify your target test(s) now pass
4. ✅ Write clear commit message explaining the fix

## Files You'll Need

### Configuration
- `.claude/VALIDATION_REQUIREMENTS.md` - Testing requirements
- `.claude/strategy/agent-workflow.md` - How to work
- `.claude/tracking/assignments.json` - Track your task

### Prompts (assigned to each agent)
- `/agent-batch-0X-*.md` - Specific task details

### Status Documents
- `.claude/tracking/CURRENT_STATUS_2025_11_08.md` - Current detailed status
- `.claude/strategy/MASTER_PLAN.md` - Overall strategy

## Repository Structure

```
less.go/
├── .claude/                          # Project coordination
│   ├── prompts/                      # Agent task prompts (this batch)
│   ├── strategy/                     # Strategic planning
│   ├── tasks/                        # Detailed task specs
│   │   ├── archived/                 # Completed tasks
│   │   ├── output-differences/       # Output diff tasks
│   │   └── runtime-failures/         # Runtime failure tasks
│   └── tracking/                     # Status tracking
├── packages/less/src/less/less_go/   # 👈 Go code to modify
├── packages/test-data/               # Test inputs/outputs
└── packages/less/src/less/           # Original JavaScript (DON'T EDIT)
```

## Key Success Criteria

1. ✅ **No Regressions** - Only fix, never break existing passing tests
2. ✅ **All Unit Tests Pass** - Run `pnpm -w test:go:unit` must 100% pass
3. ✅ **Target Test Passes** - Your specific test must now pass/have better output
4. ✅ **Clear Commits** - Commit messages explain what was fixed and why
5. ✅ **Follow Workflow** - Use the workflow in `.claude/strategy/agent-workflow.md`

## Quick Start Template

```bash
# 1. Setup
cd /home/user/less.go
git fetch origin
git checkout -b claude/your-task-name-SESSION_ID

# 2. Understand the problem
# Read your prompt file carefully
# Read the test file at packages/test-data/
# Run the test to see current failure

# 3. Fix the code
# Edit files in packages/less/src/less/less_go/
# Test frequently with: pnpm -w test:go:filter -- "test-name"

# 4. Validate
pnpm -w test:go:unit          # Unit tests
pnpm -w test:go               # Integration tests

# 5. Commit
git add -A
git commit -m "Fix [issue]: [description]"
git push -u origin claude/your-task-name-SESSION_ID

# 6. Update tracking
# Update .claude/tracking/assignments.json
# Mark your task as complete with PR branch
```

## Questions?

If you get stuck:
1. Check the task-specific prompt file (lots of details there)
2. Check `.claude/strategy/agent-workflow.md` for workflow issues
3. Check `.claude/tasks/archived/` for completed similar tasks
4. Review LESS.js reference at `packages/less/src/less/less/` directory
5. Use debug environment variables: `LESS_GO_TRACE=1`, `LESS_GO_DEBUG=1`

---

**Status**: Ready for agent assignment ✅
**Quality**: All tasks thoroughly documented ✅
**Risk**: Low - tasks are independent, can be parallelized ✅
**Expected Impact**: +15-20 perfect matches possible ✅

Good luck! 🚀
