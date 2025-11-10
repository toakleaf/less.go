# Handoff: Container Query Bubbling Fix

## Current Status: 95% Complete - One Key Issue Remains

Branch: `claude/fix-container-query-bubbling-011CUyGgoTEyCskHL73d9XHHY`

### The Problem
Container queries don't bubble out of parent rulesets. They render inline instead of at the root level like Media queries.

**Current (Wrong) Output:**
```css
.widget {
  container-type: inline-size;
  @container (max-width: 350px) {    /* ❌ Should bubble out */
    .cite .wdr-authors {
      display: none;
    }
  }
}
```

**Expected Output:**
```css
.widget {
  container-type: inline-size;
}
@container (max-width: 350px) {      /* ✅ Bubbled to root */
  .widget .cite .wdr-authors {       /* ✅ Selectors combined */
    display: none;
  }
}
```

## Root Cause Identified

**The bug is a single boolean flag!**

When Container's inner ruleset is created during evaluation, its `Root` flag is being set to `true` when it should be `false`.

**Evidence:**
```
[Media.GenCSS] Rules[0] is Ruleset, Root=false    ✅ Correct
[Container.GenCSS] Rules[0] is Ruleset, Root=true ❌ BUG!
```

This causes Container to output inline instead of bubbling to the root level.

## What's Been Done

✅ **Added empty content check to Container.GenCSS** (lines 91-105 in `container.go`)
   - Matches Media.GenCSS pattern
   - Prevents output of empty nested container queries

✅ **Fixed Container unit test** (`container_test.go` line 195)
   - Test now uses non-empty container content

✅ **Verified no regressions**
   - ✅ All 2,290+ unit tests pass
   - ✅ 76 perfect CSS matches in integration tests (no decrease)
   - ✅ Container still in "output differs" category (expected until fixed)

## The Fix Needed (Should Take <30 Minutes)

### Step 1: Find Where Root Flag Gets Set
All `NewRuleset` calls in `container.go` correctly use `Root=false`:
- Line 57: `NewRuleset(selectors, rulesetRules, false, nil)`
- Line 362: `NewRuleset(selectors, mediaBlocks, false, ...)`
- Line 399: `NewRuleset(selectors, mediaBlocks, false, ...)`
- Line 540: `NewRuleset([]any{}, []any{}, false, nil)`
- Line 608: `NewRuleset(anySelectors, []any{c.Rules[0]}, false, nil)`

**Therefore:** The `Root` flag must be getting changed AFTER creation, somewhere during evaluation.

### Step 2: Debug to Find the Culprit
Add temporary debug logging to track when Root changes:

```go
// In container.go, after line 176 (after inner ruleset evaluation):
if ruleset, ok := media.Rules[0].(*Ruleset); ok {
    fmt.Fprintf(os.Stderr, "[Container.Eval] After eval: inner ruleset Root=%v\n", ruleset.Root)
}
```

Compare with Media.Eval at the same point.

### Step 3: Fix the Root Flag
Likely culprits to investigate:
1. **Ruleset.Eval** - Check if it sets Root=true during evaluation
2. **EvalTop/EvalNested** - Check if these methods modify the Root flag
3. **Some visitor or transformer** - Check if any code paths modify Root

### Step 4: Test the Fix
```bash
# Run container test
pnpm -w test:go:filter -- "container"

# Should show:
# ✅ container: Perfect match!

# Verify no regressions
pnpm -w test:go:unit          # Must pass 100%
pnpm -w test:go:summary       # Must show 77+ perfect matches
```

## Files to Review

**Primary:**
- `packages/less/src/less/less_go/container.go` - Container implementation
- `packages/less/src/less/less_go/media.go` - Working Media implementation (for comparison)
- `packages/less/src/less/less_go/ruleset.go` - Ruleset.Eval (likely culprit)

**Test Files:**
- `packages/test-data/less/_main/container.less` - Input
- `packages/test-data/css/_main/container.css` - Expected output

## Quick Start Commands

```bash
# Checkout the branch
git checkout claude/fix-container-query-bubbling-011CUyGgoTEyCskHL73d9XHY

# Test current status (will fail)
pnpm -w test:go:filter -- "container"

# Run unit tests (should pass)
pnpm -w test:go:unit

# After fixing, verify
pnpm -w test:go:summary
```

## Additional Context

### How Media Bubbling Works (Correctly)
1. Media.Eval creates inner ruleset with `Root=false`
2. Media.EvalTop returns the Media node and clears `mediaBlocks`
3. Root ruleset appends Media to Rules (line 775 in ruleset.go)
4. During GenCSS, Media outputs at root level with combined selectors

### Why Container Fails
Container follows the same pattern, but somehow the inner ruleset has `Root=true`, which causes different rendering behavior.

### Investigation Notes
The following were explored and ruled out:
- ❌ MediaBlocks clearing - Both Media and Container clear correctly
- ❌ BubbleSelectors not being called - Not the issue
- ❌ Empty content check - Added, but not the main issue
- ✅ Root flag discrepancy - **This is the bug!**

## Success Criteria

When fixed, running `pnpm -w test:go:filter -- "container"` should show:
```
✅ container: Perfect match!
```

Integration test summary should show:
```
✅ Perfect CSS Matches: 77 (or higher)
```

## Questions?

The investigation is thoroughly documented. The fix should be straightforward once you find where the Root flag is being changed. Look for code that modifies ruleset.Root after creation, especially in the evaluation path.

Good luck! You're 95% there. 🎯
