# Task: Fix Namespacing Variable Calls Regressions

**Status**: Available
**Priority**: CRITICAL
**Estimated Time**: 3-4 hours
**Complexity**: High
**Tests Affected**: 2 (namespacing-6, namespacing-functions)

## Overview

Two tests that were previously working (or better) have regressed with identical symptoms: **variable calls to mixin results are failing**. When a mixin call is assigned to a variable and then that variable is called, evaluation fails with "Could not evaluate variable call @<varname>".

This is a **REGRESSION** - at least namespacing-6 was working as a perfect match before recent namespace/variable work.

## Failing Tests

### 1. namespacing-6 (CRITICAL REGRESSION)
- **Previous Status**: ✅ Perfect match
- **Current Status**: ❌ Compilation failed
- **Error**: `Syntax: Could not evaluate variable call @alias`
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "namespacing/namespacing-6"
  ```

### 2. namespacing-functions (REGRESSION - worse)
- **Previous Status**: ⚠️ Output differs
- **Current Status**: ❌ Compilation failed
- **Error**: `Syntax: Could not evaluate variable call @dr`
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "namespacing/namespacing-functions"
  ```

## Current Behavior

```less
// Example from namespacing-6
.something(foo) {
  width: 10px;
}

.rule-1 {
  @alias: .something(foo);  // Assign mixin call to variable
  @alias();                 // Call the variable - FAILS ❌
}
```

**Error**: `Could not evaluate variable call @alias`

The mixin call result is being assigned to a variable, but when that variable is invoked as a call, the evaluator can't handle it.

## Expected Behavior

The variable should hold a callable reference (DetachedRuleset) that can be invoked. This is similar to how detached rulesets work:

```less
@detached: {
  color: red;
};
.rule {
  @detached();  // This works ✅
}
```

Mixin results should be storable and callable the same way.

## Investigation Starting Points

### Files to Examine

1. **`variable_call.go`** - Where the error occurs (lines ~35-55)
   - The error message is thrown from here
   - Check how it evaluates the variable value
   - Check type assertions and what types it accepts

2. **`variable.go`** - Variable storage/retrieval (lines ~80-120)
   - How values are stored when variables are assigned
   - Whether mixin call results are properly captured

3. **`mixin_call.go`** - Mixin call evaluation (lines ~150-250)
   - What type of value is returned from a mixin call
   - Whether it returns a DetachedRuleset or something else

4. **`detached_ruleset.go`** - For comparison
   - How detached rulesets work (they DO support calling)
   - What interfaces they implement

### Debug Commands

```bash
# Run with trace to see evaluation flow
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "namespacing/namespacing-6"

# Look specifically for variable assignment and calling
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "namespacing/namespacing-6" 2>&1 | grep -E "(@alias|Variable|MixinCall)"
```

### Key Code Location

From `variable_call.go` (~line 35-55):
```go
func (vc *VariableCall) Eval(context any) (any, error) {
    value := context.(*EvalContext).Frames.Variable(vc.Variable)
    if value == nil {
        return nil, &LessError{
            Type:    "Syntax",
            Message: fmt.Sprintf("Could not evaluate variable call @%s", vc.Variable),
        }
    }

    // Problem likely here - what types does it check for?
    // Does it handle mixin call results?
}
```

## Root Cause Hypothesis

**Most Likely**: Recent namespace/variable work changed how mixin call results are stored or what type they return. The value stored in the variable is no longer compatible with what `VariableCall.Eval()` expects.

**Possible Issues**:
1. Mixin calls might be returning the wrong type (not DetachedRuleset)
2. Variable storage might be unwrapping/transforming mixin results incorrectly
3. VariableCall evaluation might have regressed to not check for the right callable types

**Similar Fixed Issue**: This is similar to **Issue #2** (detached-rulesets) which was FIXED. That fix involved checking `Eval(any) (any, error)` signature BEFORE `Eval(any) any`. The same pattern might be needed here.

## Success Criteria

- [ ] `namespacing-6` test passes (restore perfect match)
- [ ] `namespacing-functions` test passes (compilation succeeds)
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] FULL integration test suite shows no new regressions: `pnpm -w test:go`
- [ ] Perfect match count back to 15 or better

## Validation Checklist

Before creating PR:

```bash
# 1. Verify both specific tests pass
pnpm -w test:go:filter -- "namespacing/namespacing-6"
# Expected: ✅ Perfect match

pnpm -w test:go:filter -- "namespacing/namespacing-functions"
# Expected: ✅ Pass (compilation succeeds)

# 2. Run ALL unit tests (catch any regressions) - REQUIRED
pnpm -w test:go:unit
# Expected: ✅ All unit tests pass (no failures)

# 3. Run FULL integration test suite - REQUIRED
pnpm -w test:go
# Expected:
# - ✅ 15+ perfect matches (was 14 before fix)
# - ✅ No new compilation failures
# - ✅ extend-clearfix still passing
```

**If any test fails that was passing before**: STOP and fix the regression before proceeding.

## Additional Context

- See `.claude/agents/agent-namespacing-6/TASK.md` for previous attempt at this task
- See `.claude/reference-issues/ISSUE_NAMESPACING.md` for deep dive analysis
- Compare with successful detached ruleset calling in `detached_ruleset.go`
- This is blocking progress - high priority to restore previously working functionality

## Notes

These are regressions, meaning they worked before. Check git history to see:
```bash
# Find when namespacing-6 last passed
git log --all --grep="namespacing-6" --oneline

# Check recent changes to variable_call.go
git log -p --follow -- packages/less/src/less/less_go/variable_call.go | head -200
```

The fix might be as simple as reverting a recent change that broke this functionality.
