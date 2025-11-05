# Prompt: Fix Critical Test Regressions

## Context

Recent work on the less.go project has introduced **major regressions**. A test audit on 2025-11-05 revealed that while some improvements were made, multiple previously working tests are now broken.

**Current Status**:
- Perfect Matches: 14 (DOWN from 15)
- Compilation Failures: 11 (UP from 6)
- Net Result: MAJOR REGRESSION

## Your Task

**Fix the critical regressions and restore the codebase to a stable state before any new feature work proceeds.**

## Critical Regressions to Fix

### Priority 1: namespacing-6 (CRITICAL)
- **Previous Status**: ✅ Perfect match (recently fixed!)
- **Current Status**: ❌ Compilation failed
- **Error**: `Syntax: Could not evaluate variable call @alias`
- **Location**: `../../../../test-data/less/namespacing/namespacing-6.less`
- **Impact**: HIGH - This undid a recent fix
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "namespacing/namespacing-6"
  ```

### Priority 2: extend-clearfix
- **Previous Status**: ✅ Perfect match
- **Current Status**: ⚠️ Output differs
- **Impact**: MEDIUM - Lost a perfect match
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/extend-clearfix"
  ```

### Priority 3: namespacing-functions
- **Previous Status**: ⚠️ Output differs (at least compiled)
- **Current Status**: ❌ Compilation failed
- **Error**: `Syntax: Could not evaluate variable call @dr`
- **Impact**: MEDIUM - Made worse
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "namespacing/namespacing-functions"
  ```

### Priority 4: import-reference
- **Previous Status**: ⚠️ Output differs (was compiling)
- **Current Status**: ❌ Compilation failed
- **Error**: `open test.css: no such file or directory`
- **Impact**: MEDIUM - Made worse
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/import-reference"
  ```

### Priority 5: import-reference-issues
- **Previous Status**: ⚠️ Output differs (was compiling)
- **Current Status**: ❌ Compilation failed
- **Error**: `#Namespace > .mixin is undefined`
- **Impact**: MEDIUM - Made worse
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "main/import-reference-issues"
  ```

## Root Cause Analysis

The audit suggests recent changes to namespace/variable evaluation logic broke these tests. Likely culprits:

**Files to investigate**:
- `packages/less/src/less/less_go/variable_call.go`
- `packages/less/src/less/less_go/namespace_value.go`
- `packages/less/src/less/less_go/variable.go`
- `packages/less/src/less/less_go/import.go` (for import-reference issues)
- `packages/less/src/less/less_go/extend_visitor.go` (for extend-clearfix)

**Hypothesis**: Recent commits attempted to fix namespace resolution but:
1. Fixed one case but broke others
2. Changed variable evaluation in a way that breaks certain patterns
3. May have introduced issues with variable calls in specific contexts

## Investigation Strategy

### Step 1: Review Recent Changes
```bash
# Check recent commits
git log --oneline --graph -20

# Find commits that modified namespace/variable code
git log --oneline -- packages/less/src/less/less_go/variable_call.go
git log --oneline -- packages/less/src/less/less_go/namespace_value.go
git log --oneline -- packages/less/src/less/less_go/import.go

# Review the diffs
git diff HEAD~5 HEAD -- packages/less/src/less/less_go/variable_call.go
```

### Step 2: Compare with JavaScript
```bash
# Look at the JavaScript implementation for these features
cat packages/less/src/less/tree/variable.js
cat packages/less/src/less/tree/namespace-value.js
```

### Step 3: Examine Test Files
```bash
# Understand what the tests are trying to do
cat packages/test-data/less/namespacing/namespacing-6.less
cat packages/test-data/css/namespacing/namespacing-6.css
```

### Step 4: Debug with Tracing
```bash
# Use trace mode to see execution flow
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "namespacing/namespacing-6"
```

## Fix Approach

### Option 1: Targeted Fix
If you can identify the exact change that broke things:
1. Understand what the change was trying to fix
2. Find a way to fix that case WITHOUT breaking these cases
3. Ensure all tests pass

### Option 2: Partial Revert
If the breaking change is too invasive:
1. Consider reverting the problematic commits
2. Document what was being attempted
3. Create a new task to implement it correctly with tests

### Option 3: Comprehensive Refactor
If the logic is fundamentally flawed:
1. Study the JavaScript implementation carefully
2. Reimplement the logic to match JS exactly
3. Add unit tests for edge cases
4. Ensure all integration tests pass

## Validation Requirements

**CRITICAL**: Before creating a PR, you MUST verify:

```bash
# 1. All 5 regression tests now pass or at least compile
pnpm -w test:go:filter -- "namespacing/namespacing-6"  # Must pass
pnpm -w test:go:filter -- "main/extend-clearfix"       # Must pass
pnpm -w test:go:filter -- "namespacing/namespacing-functions"  # Should compile
pnpm -w test:go:filter -- "main/import-reference"      # Should compile
pnpm -w test:go:filter -- "main/import-reference-issues"  # Should compile

# 2. ALL unit tests must pass
pnpm -w test:go:unit
# Expected: ✅ All pass (zero failures)

# 3. Run FULL integration test suite
pnpm -w test:go
# Expected: No NEW failures, ideally improvements

# 4. Verify baseline
# Perfect matches should be >= 14
# Compilation failures should be <= 11
```

## Success Criteria

Your PR is successful when:

- ✅ namespacing-6 shows "Perfect match!" again
- ✅ extend-clearfix shows "Perfect match!" again
- ✅ namespacing-functions, import-reference, and import-reference-issues at least compile (may have output diffs)
- ✅ All unit tests pass
- ✅ No new test failures introduced
- ✅ Perfect match count is >= 15
- ✅ Compilation failure count is <= 6

## Branch Setup

```bash
# Work on the base branch
git checkout claude/port-lessjs-golang-011CUoW25gKApW3ByzxRpUVk
git pull origin claude/port-lessjs-golang-011CUoW25gKApW3ByzxRpUVk

# Create your fix branch
git checkout -b claude/fix-regressions-YOUR_SESSION_ID
```

## Reference Documents

- Full audit report: `.claude/tracking/TEST_AUDIT_2025-11-05.md`
- Validation requirements: `.claude/VALIDATION_REQUIREMENTS.md`
- Agent workflow: `.claude/strategy/agent-workflow.md`

## Communication

When you create your PR:
1. Title: "Fix critical test regressions in namespace/variable evaluation"
2. List which regressions you fixed
3. Explain the root cause you found
4. Show before/after test results
5. Confirm zero new failures

## Priority

**This task is CRITICAL and blocks all other work.**

Do not proceed with any new features until these regressions are fixed. The codebase must be stable before adding new functionality.
