# Prompt for Next Session: Fix Import-Reference Bug

Copy and paste this into a fresh LLM session to continue fixing the import-reference bug:

---

I'm working on fixing the import-reference bug in the less.go port. A previous investigation identified the root cause but didn't implement the fix yet. Please help me fix it.

## Background

This is a Go port of less.js. The `import-reference` and `import-reference-issues` integration tests are failing because rulesets that contain ONLY reference imports are being output as empty rulesets when they shouldn't appear at all.

**Example:**
```less
#do-not-show-import {
  @import (reference) "file.less";
}
```

**Expected output:** Nothing (the ruleset should not appear)
**Actual output:** `#do-not-show-import { }` (empty ruleset appears)

## Root Cause Identified

Read the complete analysis in `/home/user/less.go/IMPORT_REFERENCE_ROOT_CAUSE_ANALYSIS.md`

**Summary:** The issue is in `packages/less/src/less/less_go/to_css_visitor.go`. The problem areas are:

1. **`IsVisibleRuleset` method (lines 181-249)** - Logic flow may not correctly handle empty rulesets
2. **`VisitRuleset` special case (lines 753-769)** - May incorrectly force `keepRuleset=true` for empty rulesets

The JavaScript implementation (`packages/less/src/less/visitors/to-css-visitor.js`) handles this correctly - empty rulesets are filtered out.

## Your Task

1. **Add debug output** to `to_css_visitor.go` in the `VisitRuleset` method to trace:
   - What rules exist before and after filtering
   - What `IsEmpty()` returns
   - What `IsVisibleRuleset()` returns
   - Whether the special case at line 753 is triggered
   - The final `keepRuleset` value

2. **Run the failing test with debug**:
   ```bash
   env LESS_GO_DEBUG=1 pnpm -w test:go -- -run "import-reference-issues"
   ```

3. **Analyze the debug output** to determine:
   - Why empty rulesets are being kept
   - Which check in `IsVisibleRuleset` is returning true when it should return false
   - OR if the special case logic is overriding the correct decision

4. **Implement the fix** - likely one of:
   - Reorder checks in `IsVisibleRuleset` so empty check happens at the right time
   - Fix the special case at lines 753-769 to not apply to truly empty rulesets
   - Add an additional check before the final `return true` in `IsVisibleRuleset`

5. **Test the fix**:
   ```bash
   # Run the specific failing tests
   pnpm -w test:go -- -run "import-reference"

   # Should see:
   # ✅ import-reference: Perfect match!
   # ✅ import-reference-issues: Perfect match!
   ```

6. **Check for regressions**:
   ```bash
   # Run ALL integration tests
   pnpm -w test:go

   # Run ALL unit tests
   pnpm -w test:go:unit

   # Verify no previously passing tests now fail
   ```

7. **Commit and push** when both tests pass with no regressions:
   ```bash
   git add -A
   git commit -m "Fix import-reference bug: correctly filter empty rulesets from reference imports"
   git push -u origin claude/fix-integration-test-bug-011CUzvKjGaAM67Vw3akG2ys
   ```

## Key Files

- `packages/less/src/less/less_go/to_css_visitor.go` - The file to fix (lines 181-249 and 753-769)
- `packages/less/src/less/visitors/to-css-visitor.js` - JavaScript reference implementation
- `IMPORT_REFERENCE_ROOT_CAUSE_ANALYSIS.md` - Complete analysis document
- `packages/test-data/less/_main/import-reference-issues.less` - Test input
- `packages/test-data/css/_main/import-reference-issues.css` - Expected output

## Test Files to Examine

```bash
# Input file
cat packages/test-data/less/_main/import-reference-issues.less

# Expected output
cat packages/test-data/css/_main/import-reference-issues.css

# Compare with actual output by running test
pnpm -w test:go -- -run "import-reference-issues"
```

## Success Criteria

- ✅ `import-reference` test shows "Perfect match!"
- ✅ `import-reference-issues` test shows "Perfect match!"
- ✅ All unit tests pass: `pnpm -w test:go:unit`
- ✅ No regressions in integration tests: `pnpm -w test:go`
- ✅ The fix is minimal and surgical (likely < 20 lines changed)

## Important Notes

- Previous attempts fixed CSS detection and import loading - those work correctly now
- The issue is specifically in the visitor pattern logic for determining ruleset visibility
- Compare carefully with JavaScript - the logic should match exactly
- This is a subtle bug in the interaction between multiple checks, not a missing feature
- Make sure to run full test suite after changes to catch any regressions

Please start by reading the root cause analysis document, then add the debug output and run the test to confirm the hypothesis before implementing the fix.
