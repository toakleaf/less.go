# 🚨 CRITICAL REGRESSION DETECTED

**Date**: 2025-11-09
**Severity**: CRITICAL - Blocks all further work
**Impact**: ~42 tests regressed from passing to failing (57 → 15 perfect matches)

## Executive Summary

A recent change to `combinator.go` has introduced a critical regression that breaks selector spacing across the entire codebase. This affects approximately **42 previously passing tests** and must be fixed **immediately** before any other work can proceed.

## Current Test Status

### Unit Tests: ❌ FAILING
- **Status**: 5 tests failing in combinator/selector logic
- **Failures**:
  - `TestCombinator/genCSS/should_generate_CSS_with_no_spaces_for_empty_combinator`
  - `TestCombinator/genCSS/should_generate_CSS_with_no_spaces_for_space_combinator`
  - `TestCombinator/genCSS/should_handle_all_special_no-space_combinators`
  - `TestSelector/genCSS/should_handle_special_characters_in_selectors`
  - `TestParentSelectorExpansion/parent_selector_with_space`
  - `TestParentSelectorExpansion/nested_parent_selectors`

### Integration Tests: ❌ MASSIVE REGRESSION
- **Previous State** (documented): 57 perfect matches (31.0%)
- **Current State**: 15 perfect matches (8.2%)
- **Regression**: -42 tests (-73.7% decrease)

### Tests Lost (Sample)
All tests that rely on proper spacing in selectors have regressed:
- ❌ colors (was ✅)
- ❌ comments2 (was ✅)
- ❌ css-guards (was ✅)
- ❌ extend-clearfix (was ✅)
- ❌ extend-exact (was ✅)
- ❌ extend-media (was ✅)
- ❌ extend-nest (was ✅)
- ❌ extend-selector (was ✅)
- ❌ mixins-closure (was ✅)
- ❌ mixins-interpolated (was ✅)
- ❌ mixins-named-args (was ✅)
- ❌ mixins-nested (was ✅)
- ❌ namespacing-1 through namespacing-8 (all were ✅)
- And ~20 more...

## Root Cause Analysis

### File: `combinator.go`
### Function: `GenCSS()` (lines 111-133)

**The Bug** (lines 128-130):
```go
if c.Value == "" || (c.Value == " " && spaceOrEmpty == "") {
    return
}
```

This logic is **INCORRECT**. It returns early for space combinators (" ") even when NOT in compress mode, preventing them from outputting anything.

**Example of Broken Output**:
```less
// Input:
.parent .child { color: blue; }

// Expected Output:
.parent .child { color: blue; }

// Actual Output:
.parent.child { color: blue; }  // Missing space!
```

### Impact Chain
1. Space combinators don't output → selectors concatenate without spaces
2. `.parent .child` becomes `.parent.child` (descendant selector becomes class)
3. CSS semantics completely change
4. All tests with descendant selectors fail

## Correct Logic

The `GenCSS()` function should:
1. **Empty combinator** (`c.Value == ""`): Output nothing ✅
2. **Space combinator** (`c.Value == " "`):
   - In **compress mode**: Output nothing (spaces removed)
   - In **normal mode**: Output a space ❌ **THIS IS BROKEN**
3. **Other combinators** (`>`, `+`, `~`, etc.): Output with surrounding spaces

## Fix Required

### Location
`packages/less/src/less/less_go/combinator.go:111-133`

### Proposed Fix
```go
func (c *Combinator) GenCSS(context any, output *CSSOutput) {
	// Determine if we're in compress mode
	compress := false
	if ctx, ok := context.(map[string]any); ok {
		if c, ok := ctx["compress"].(bool); ok {
			compress = c
		}
	}

	// Handle empty combinator
	if c.Value == "" {
		return
	}

	// Handle space combinator
	if c.Value == " " {
		if !compress {
			output.Add(" ", nil, nil)
		}
		return
	}

	// Handle other combinators (>, +, ~, etc.)
	var spaceOrEmpty string
	if compress || NoSpaceCombinators[c.Value] {
		spaceOrEmpty = ""
	} else {
		spaceOrEmpty = " "
	}
	output.Add(spaceOrEmpty+c.Value+spaceOrEmpty, nil, nil)
}
```

## Verification Steps

### 1. Fix the Code
Apply the fix above to `combinator.go`

### 2. Run Unit Tests
```bash
go test -v -run TestCombinator
go test -v -run TestSelector
go test -v -run TestParentSelectorExpansion
```
**Expected**: All tests pass ✅

### 3. Run Full Unit Test Suite
```bash
pnpm -w test:go:unit
```
**Expected**: 2290+ tests pass (99.9%+) ✅

### 4. Run Integration Tests
```bash
pnpm -w test:go
```
**Expected**: ~57 perfect matches restored ✅

### 5. Verify No New Regressions
Compare integration test results before and after:
- Perfect matches should return to ~57 tests
- No currently passing tests should break

## Priority: IMMEDIATE

**This regression must be fixed before any other work proceeds.**

All other planned work (math operations, URL rewriting, etc.) depends on having a stable baseline. With 42 tests regressed, we cannot accurately assess progress on other issues.

## Next Steps After Fix

Once this regression is resolved:
1. ✅ Verify 57 perfect matches restored
2. ✅ Run full test suite to confirm stability
3. ✅ Update `.claude/AGENT_WORK_QUEUE.md` with current status
4. ✅ Resume work on high-priority tasks:
   - extend-chaining (complete 7/7 extend tests)
   - Math operations (6 tests)
   - URL rewriting (7 tests)
   - Formatting/comments (6 tests)

## Responsible Agent

This task should be assigned to the **highest priority agent** or handled immediately by the maintainer.

**Estimated Time to Fix**: 30 minutes - 1 hour
**Complexity**: Low (logic bug, straightforward fix)
**Impact**: CRITICAL (blocks all other work)

---

**Remember**: Always run BOTH unit and integration tests before committing changes. This regression could have been caught with proper testing before the commit.
