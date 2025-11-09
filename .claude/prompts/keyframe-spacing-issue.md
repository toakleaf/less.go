# Fix Keyframe Spacing Issue in @keyframes At-Rules

## Problem Statement

The Go port of less.js is producing incorrect spacing for keyframe selectors inside `@keyframes` at-rules. The output has 3 spaces of indentation before keyframe selectors (like `from` and `to`) instead of 2, and is missing newlines between keyframe rules.

## Current vs Expected Output

**Current (WRONG):**
```css
@keyframes enlarger {
   from {           /* ← 3 spaces instead of 2 */
    font-size: 12px;
  } to {            /* ← Missing newline before 'to' */
    font-size: 15px;
  }
}
```

**Expected (CORRECT):**
```css
@keyframes enlarger {
  from {            /* ← Exactly 2 spaces */
    font-size: 12px;
  }
  to {              /* ← Newline before 'to' */
    font-size: 15px;
  }
}
```

## Test Case

- **Input file**: `packages/test-data/less/_main/variables-in-at-rules.less`
- **Expected output**: `packages/test-data/css/_main/variables-in-at-rules.css`
- **Test command**: `go test -v -run "TestIntegrationSuite/main/variables-in-at-rules" ./packages/less/src/less/less_go/...`

## Root Cause Analysis

After investigation, the issue appears to be:

1. **Space Combinator Problem**: When the parser encounters keyframe selectors like `from` and `to`, it creates `Element` nodes with space combinators (`" "`) because there's whitespace before them in the input
2. **Double Spacing**:
   - `AtRule.OutputRuleset()` outputs `\n  ` (newline + 2 spaces) for indentation at line 318
   - When the child Ruleset's selector is output, its first Element has a space combinator that adds another space
   - Result: 3 spaces total instead of 2
3. **Missing Newlines**: Child rulesets aren't properly separated by newlines

## Key Files

### Go Implementation:
- **`packages/less/src/less/less_go/atrule.go`** - `OutputRuleset()` method (lines 277-347)
- **`packages/less/src/less/less_go/ruleset.go`** - `GenCSS()` method (lines 1367-1640)
- **`packages/less/src/less/less_go/combinator.go`** - `GenCSS()` method (lines 110-125)
- **`packages/less/src/less/less_go/element.go`** - `ToCSS()` method (lines 236-299)
- **`packages/less/src/less/less_go/parser.go`** - `Combinator()` parser (lines 3499-3538)

### JavaScript Reference:
- **`packages/less/src/less/tree/atrule.js`** - `outputRuleset()` method (lines 122-153)
- **`packages/less/src/less/tree/ruleset.js`** - `genCSS()` method (lines 454-565)
- **`packages/less/src/less/tree/combinator.js`** - `genCSS()` method (lines 21-24)

## JavaScript Reference Implementation

The JavaScript `atrule.js` `outputRuleset()` method (lines 138-150):
```javascript
// Non-compressed
const tabSetStr = `\n${Array(context.tabLevel).join('  ')}`, tabRuleStr = `${tabSetStr}  `;
if (!ruleCnt) {
    output.add(` {${tabSetStr}}`);
} else {
    output.add(` {${tabRuleStr}`);
    rules[0].genCSS(context, output);
    for (i = 1; i < ruleCnt; i++) {
        output.add(tabRuleStr);
        rules[i].genCSS(context, output);
    }
    output.add(`${tabSetStr}}`);
}
```

Key observation: JavaScript outputs `tabRuleStr` BEFORE the opening brace for the first rule, then for subsequent rules.

## Current Go Implementation

```go
// atrule.go lines 310-344
tabSetStr := "\n" + strings.Repeat("  ", tabLevel-1)
tabRuleStr := tabSetStr + "  "

if ruleCnt == 0 {
    output.Add(" {"+tabSetStr+"}", nil, nil)
} else {
    output.Add(" {"+tabRuleStr, nil, nil)  // Adds indentation with opening brace

    for i := 0; i < ruleCnt; i++ {
        if i > 0 {
            output.Add(tabRuleStr, nil, nil)
        }
        // ... genCSS call ...
    }
}
```

## What Was Tried (Did NOT Work)

1. **Attempt**: Temporarily replace space combinators with empty combinators before calling `GenCSS`
   - Modified both `Paths` and `Selectors` properties
   - Saved/restored combinators around the `GenCSS` call
   - **Result**: No change in output - the modification didn't take effect

2. **Why it failed**: The issue is likely more fundamental in how the output is structured, not just the combinator values

## Important Constraints

⚠️ **DO NOT** modify:
- `combinator.go` - Already matches JavaScript exactly
- `element.go` - Already matches JavaScript exactly
- Parser combinator logic - Already matches JavaScript exactly

The fix must be in the **output generation phase**, not the parsing or combinator logic.

## Recommended Investigation Steps

1. **Compare Exact JavaScript Output Flow**:
   - The JavaScript and Go combinator logic are identical
   - The difference must be in HOW the output is assembled
   - Focus on the sequence of `output.Add()` calls

2. **Check Ruleset.GenCSS Behavior Inside At-Rules**:
   - When a Ruleset is inside an at-rule, does it behave differently?
   - Check if `firstSelector` context flag affects combinator output
   - Verify if Ruleset uses `Paths` or `Selectors` (likely `Paths` after JoinSelectorVisitor)

3. **Debug Output Sequence**:
   - Add debug logging to see exact sequence of output chunks
   - Compare with what JavaScript would output
   - Look for where the extra space is being added

4. **Investigate Context Flags**:
   - Check if there's a context flag that should suppress leading combinators
   - JavaScript Element.toCSS uses `context.firstSelector` - verify this is working correctly
   - See `element.go` lines 246-249 and 269-270

## Verification Commands

```bash
# Run the specific test
go test -v -run "TestIntegrationSuite/main/variables-in-at-rules" ./packages/less/src/less/less_go/...

# Run unit tests (must pass 100%)
pnpm -w test:go:unit

# Run full integration suite
pnpm -w test:go

# Check current perfect match count (should be >= 63)
pnpm -w test:go 2>&1 | grep -c "Perfect match"
```

## Success Criteria

✅ `variables-in-at-rules` test achieves perfect match
✅ ALL previously passing tests remain passing (no regressions)
✅ Perfect matches count >= 63
✅ Unit tests pass 100%

## Additional Context

- The codebase is a Go port of less.js maintaining 1:1 functionality
- Current status: 63/185 tests perfect matches (34.1%)
- Integration test suite: `packages/less/src/less/less_go/integration_suite_test.go`
- The issue is purely CSS formatting - compilation succeeds

## Hint for Solution

The key insight: JavaScript's `outputRuleset` adds `tabRuleStr` (which includes the newline) RIGHT AFTER the opening brace for the first element, then BEFORE each subsequent element. This means each child ruleset receives the indentation BEFORE it starts outputting its selector.

When the Ruleset then outputs its selector with a space combinator, the combinator's spaces should not interfere because the indentation has already positioned the cursor correctly.

The Go implementation may need to ensure that:
1. Indentation is added at the right time
2. The `firstSelector` flag is set correctly to suppress leading combinator spaces
3. Newlines are added between child rulesets
