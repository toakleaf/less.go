# Selector Interpolation Bug Investigation - Document Index

## Summary
Complete investigation into why selector interpolation fails in the parse-interpolation test. The bug causes only the last selector in an interpolated comma-separated list to receive pseudo-classes or combinators.

## Investigation Documents

### 1. Main Investigation Report
**File:** `SELECTOR_INTERPOLATION_BUG_SUMMARY.md`
- Complete root cause analysis
- 4-step problem chain explanation
- Data flow visualization
- Two fix approaches with code examples
- Testing checklist
- Impact assessment

**Key Finding:** The `IsVariable` flag is not set when a string element undergoes interpolation, causing the re-parsing logic to never execute.

### 2. Detailed Analysis
**File:** `selector-interpolation-root-cause.md`
- In-depth explanation of each affected system
- Comparison with JavaScript behavior
- Why this matters (all 8 test cases affected)
- Files that need changes
- Testing approach

### 3. Quick Reference Summary
**File:** `SELECTOR_INTERPOLATION_INVESTIGATION.txt`
- Quick overview of problem and solution
- 4-step problem chain summary
- Critical code locations
- Two fix approaches
- Affected test cases list
- Next steps

## The Bug in One Diagram

```
Input:  @inputs: input[type=text], input[type=email];
        @{inputs} { &:focus { ... } }

Stage 1: PARSER (Line 3043 in parser.go)
         @{inputs} ──[Regex]──> Element{Value: "@{inputs}", IsVariable: false}

Stage 2: ELEMENT EVAL (Lines 151-207 in element.go)
         Element{IsVariable: false} ──[Interpolation]──>
         Element{Value: "input[type=text], input[type=email]", IsVariable: false}
                    ↑ PROBLEM: IsVariable should be true now! ↑

Stage 3: RE-PARSE DETECTION (Lines 360-378 in ruleset.go)
         if elem.IsVariable {  // <-- FALSE! Detection fails
             hasVariable = true
         }
         Result: Re-parsing code never executes

Stage 4: SELECTOR JOINING (Lines 1665+ in ruleset.go)
         Element with comma-separated string treated as single element
         Result: ":focus" only on last selector

Output: input[type=text], input[type=email]:focus { ... }  ❌ WRONG!
        Should be: input[type=text]:focus, input[type=email]:focus
```

## Recommended Fix

**Approach A (Recommended):** In `element.go`, mark elements when interpolation occurs

Change in Element.Eval() to set a flag when `@{` interpolation happens:
```go
newElement := NewElement(
    e.Combinator,
    evaluatedValue,
    e.IsVariable || interpolated,  // Include interpolation flag
    index,
    fileInfo,
    visibilityInfo,
)
```

This lets the existing hasVariable detection in ruleset.go (line 366) catch the interpolated element and trigger re-parsing.

## Files Most Relevant

1. **element.go** - Lines 151-207 (where interpolation happens)
2. **ruleset.go** - Lines 360-378 (where detection fails)  
3. **ruleset.go** - Lines 384-431 (re-parsing code that should run but doesn't)
4. **parser.go** - Line 3043 (where @{var} becomes string, not Variable node)

## Test Case
- **Input:** `/home/user/less.go/packages/test-data/less/_main/parse-interpolation.less`
- **Expected:** `/home/user/less.go/packages/test-data/css/_main/parse-interpolation.css`
- **Run test:** `pnpm -w test:go` (look for parse-interpolation test)

## Impact

- 8 failing test cases in parse-interpolation
- Affects 25-30+ other tests indirectly
- Breaks any selector with: `@{var}` + pseudo-class/combinator
- Critical for full CSS output correctness

---

**Status:** Investigation complete, ready for fix implementation

**Last Updated:** 2025-11-09

**Key Insight:** The problem is that `IsVariable` flag doesn't get set when a string element undergoes interpolation. This breaks the detection that triggers necessary re-parsing. The fix is simple: set the flag during interpolation in Element.Eval().
