# extend-chaining Investigation Results

**Date**: 2025-11-10
**Session**: claude/fix-extend-chaining-011CUzmqFzXPQYDzJptCPsV1

## Was it a regression?

**YES** - This was a true regression introduced on November 10, 2025.

### Timeline:
- **Nov 9, 2025**: extend-chaining was a **perfect match** (documented in assessment report)
- **Nov 10, 2025**: extend-chaining began **failing with output differences**

## Root cause:

**Commit 587acf4** ("Fix import-reference extend chain architecture #210") introduced an architectural change that broke regular extend chaining.

### The Bug:

The commit changed line 604 in `extend_visitor.go` from adding selector paths to the **matched ruleset** to adding them to the **extend's ruleset**:

```go
// WRONG (commit 587acf4):
targetRuleset.Paths = append(targetRuleset.Paths, extendedSelectors)

// CORRECT (JavaScript behavior):
ruleset.Paths = append(ruleset.Paths, extendedSelectors)
```

### Why This Broke extend-chaining:

In the test case:
```less
.a { color: black; }
.b:extend(.a) {}
.c:extend(.b) {}
```

Expected output: `.a, .b, .c { color: black; }` (all selectors on same ruleset)

With the bug:
- When `.b:extend(.a)` matched `.a`, the path `.b` was added to `.b`'s ruleset (not `.a`'s)
- When `.c:extend(.b)` matched `.b`, the path `.c` was added to `.c`'s ruleset (not `.b`'s)
- Result: Each selector outputted separately instead of grouped together

### JavaScript Comparison:

The JavaScript `extend-visitor.js` line 296 clearly shows:
```javascript
rulesetNode.paths = rulesetNode.paths.concat(selectorsToAdd);
```

It adds to `rulesetNode.paths` (the **matched ruleset**), NOT to the extend's ruleset.

## Fix applied:

Changed `extend_visitor.go` line 604-608 to add paths to the matched ruleset:

```go
// CRITICAL: Add paths to the MATCHED ruleset (like JavaScript extend-visitor.js line 296)
// This ensures extend chaining works correctly (.c:extend(.b) + .b:extend(.a) = .a,.b,.c)
// The targetRuleset is only used for visibility management (reference imports)
ruleset.Paths = append(ruleset.Paths, extendedSelectors)
// Mark this ruleset as modified so we can deduplicate it later
modifiedRulesets[ruleset] = true
```

### Important Note:

The `targetRuleset` variable (which points to the extend's ruleset) is still used correctly for **visibility management** on lines 570-599. This is needed for reference import handling. The fix only changes WHERE selector paths are added, not which rulesets are marked visible.

## Test results:

### extend-chaining: ✅ PASS (Perfect match!)

All 7 extend tests now pass:
1. ✅ extend-chaining (FIXED!)
2. ✅ extend-clearfix
3. ✅ extend-exact
4. ✅ extend-media
5. ✅ extend-nest
6. ✅ extend-selector
7. ✅ extend

### Perfect matches: 79 (up from 78) ⬆️

**Zero regressions** - all other tests remain in the same state

### import-reference tests:

The import-reference and import-reference-issues tests still have output differences (as they did before). The fix did not make them worse. These tests need separate investigation and fixes.

## Conclusion:

This was a **regression** caused by an architectural misunderstanding in commit 587acf4. The commit was trying to fix import-reference tests by changing where paths are added, but this broke normal extend chaining.

The correct approach (now implemented) is:
- **Add paths to the matched ruleset** (like JavaScript) for proper extend chaining
- **Mark the extend's ruleset as visible** (for reference import handling)

These two concepts are separate and should not be conflated.
