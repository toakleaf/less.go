# Import-Reference Fix Handoff Prompt

**Copy everything below this line and paste into a fresh Claude session:**

---

# Task: Complete import-reference Fix in less.go

## Critical Information
- **Project**: `/home/user/less.go` - LESS CSS preprocessor ported from JavaScript to Go
- **Branch**: `claude/fix-import-reference-visibility-011CUzUW1GkAgERNMJy2XUnx`
- **Session ID**: `011CUzWJ7hLvtarEucR1NpB9` (use for new branch if needed)
- **Context file**: Read `/home/user/less.go/CLAUDE.md` for full project context
- **Zero regressions tolerance**: Must maintain 79 perfect CSS matches

## What I Need You To Do

Complete the fix for `import-reference` and `import-reference-issues` integration tests. Previous session made partial progress - the selector now appears but gets wrong properties.

## Current State

### What Works ✅
- `.test-rule-c` now appears in output (was completely missing)
- All 79 perfect matches maintained (zero regressions)
- All unit tests passing
- Extend visibility logic fixed in `extend.go`

### What's Broken ❌
- `.test-rule-c` gets properties from `.test-rule-a` (color: red)
- Should get properties from `.test-rule-b` (background-color: green)
- Issue: Rulesets from reference imports aren't outputting their properties even when matched by visible extends

## The Test Case

**File**: `test/less/import-reference-issues.less`
```less
// Main file
@import (reference) "import-reference-issues/global-scope-import.less";
.test-rule-c {
  &:extend(.test-rule-b all);
}
```

**Reference import** (`global-scope-import.less`):
```less
.test-rule-b {
  background-color: green;
  &:extend(.test-rule-a all);
}
.test-rule-a {
  color: red;
}
```

**Expected CSS**:
```css
.test-rule-c {
  background-color: green;
}
```

**Current CSS** (partial fix):
```css
.test-rule-a,
.test-rule-c {
  color: red;
}
```

## Previous Fix (Already Committed)

**Commit**: `9fe2f46` - "WIP: Partial fix for import-reference visibility"
**File**: `packages/less/src/less/less_go/extend.go`
**Change**: Modified `Extend.IsVisible()` to check if any `SelfSelectors` are visible

```go
// Now correctly identifies extends from main file as visible
func (e *Extend) IsVisible() bool {
    // Check if any self selector is visible (not from a visibility block)
    for _, sel := range e.SelfSelectors {
        if sel.VisibilityInfo() != VisibilityInfoVisible {
            return false
        }
    }
    return true
}
```

## The Remaining Problem

**Root cause**: When a visible extend from the main file matches a selector in a reference import's ruleset, that ruleset should output its properties to CSS. Currently:

1. The extend match is found ✅
2. The selector is added to the output ✅
3. BUT the ruleset containing `background-color: green` is NOT being output ❌

**Why**: Rulesets from reference imports are marked invisible, and even though we call `EnsureVisibility()` and `RemoveVisibilityBlock()`, the individual Rules inside the ruleset may still be filtered out during CSS generation.

## Key Files to Investigate

### Primary suspects (Go code):

1. **`packages/less/src/less/less_go/extend_visitor.go`**
   - Lines 498-518: Code that tries to make matched reference import rulesets visible
   - Lines 962-968: Sets EvaldCondition = true for visible extends
   - **Hypothesis**: The visibility changes here may not be propagating to the Rules

2. **`packages/less/src/less/less_go/to_css_visitor.go`**
   - Lines 584-803: Ruleset CSS generation and filtering
   - Lines 776-801: Path filtering logic (compileRulesetPaths)
   - **Hypothesis**: Visibility checks may be filtering out reference import rules

3. **`packages/less/src/less/less_go/ruleset.go`**
   - Check `GenCSS()` method and how it handles visibility
   - Check if Rules need individual visibility flags

### Reference implementation (JavaScript - source of truth):

4. **`packages/less/src/less/visitors/extend-visitor.js`**
   - Lines 190-210: How JS makes reference import rulesets visible

5. **`packages/less/src/less/visitors/to-css-visitor.js`**
   - Lines 282-296: How JS filters rulesets during CSS generation

## Testing Commands

```bash
# Navigate to project
cd /home/user/less.go

# Test specific failing test
pnpm -w test:go:filter -- "import-reference"

# Show detailed diff
cd packages/less/src/less/less_go
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/import-reference-issues" .

# Check for regressions (MUST stay at 79)
cd /home/user/less.go
pnpm -w test:go 2>&1 | grep -a "Perfect match" | wc -l

# Run all unit tests (must pass 100%)
pnpm -w test:go:unit
```

## Debugging Tips

Enable trace logging to see evaluation flow:
```bash
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "import-reference-issues"
```

Add debug prints in key locations:
```go
// In extend_visitor.go around line 509
fmt.Printf("Making ruleset visible: %v, Rules count: %d\n", ruleset, len(ruleset.Rules))

// In to_css_visitor.go in Ruleset case
fmt.Printf("Generating CSS for ruleset, visible: %v, rules: %d\n", ruleset.IsVisible(), len(ruleset.Rules))
```

## Hypothesis

The issue is in how rulesets from reference imports are made visible:

1. We call `ruleset.Node.EnsureVisibility()` and `RemoveVisibilityBlock()` on the ruleset node
2. BUT the individual `Rule` nodes inside `ruleset.Rules` may still have visibility blocks
3. OR the CSS generation visitor is filtering based on parent visibility, not node visibility
4. OR the paths are being added but the ruleset body isn't being output

**Check**: Do we need to call `EnsureVisibility()` on each Rule in the ruleset's Rules array?

**Check**: Is there a visibility check in `Ruleset.GenCSS()` that's filtering out the rules?

## Success Criteria

✅ `.test-rule-c { background-color: green; }` appears in output
✅ Both `import-reference` and `import-reference-issues` tests pass (perfect CSS match)
✅ All 79 perfect matches still passing (zero regressions)
✅ All 2,290+ unit tests passing

## Suggested Workflow

1. **Read context**: Start by reading `/home/user/less.go/CLAUDE.md`
2. **Verify current state**: Run the test to confirm current output
3. **Compare with JS**: Read the JavaScript implementation in `extend-visitor.js` lines 190-210
4. **Add debug logging**: Trace where the rules are being filtered
5. **Identify the gap**: Find where Rules need visibility set
6. **Implement fix**: Make minimal changes to fix the issue
7. **Test thoroughly**:
   - Both import-reference tests must pass
   - Run full suite to verify zero regressions
   - Run all unit tests
8. **Commit and push**: Use clear commit message on the feature branch

## Branch Information

- **Current branch**: `claude/fix-import-reference-visibility-011CUzUW1GkAgERNMJy2XUnx`
- **If you need a new branch**: Use session ID `011CUzWJ7hLvtarEucR1NpB9`
- **Push with**: `git push -u origin <branch-name>`

## Context

This is issue #202 in the project's tracking system. Import-reference is a LESS feature where `@import (reference)` imports styles that are only included in the output if they're explicitly referenced (via extend or mixin usage).

The reference import mechanism uses "visibility blocks" to mark nodes as invisible by default. When an extend from the main file matches a selector in a reference import, that selector's ruleset should become visible.

---

**Please start by reading CLAUDE.md and running the test to confirm current behavior, then proceed with the fix.**
