# Selector Interpolation Bug - Complete Investigation

## Problem Statement

The `parse-interpolation` integration test fails because selector interpolation doesn't correctly expand multiple comma-separated selectors when applying pseudo-classes or combinators to an interpolated variable.

### Example
```less
@inputs: input[type=text], input[type=email], input[type=password], textarea;

@{inputs} {
  &:focus {
    foo: bar;
  }
}
```

**Expected Output:**
```css
input[type=text]:focus,
input[type=email]:focus,
input[type=password]:focus,
textarea:focus {
  foo: bar;
}
```

**Actual Output:**
```css
input[type=text], input[type=email], input[type=password], textarea:focus {
  foo: bar;
}
```

Notice how `:focus` is only applied to the **last** selector instead of all four.

## Root Cause Analysis

### The 4-Step Problem Chain

#### Step 1: Parser Creates String Element (Not Variable Node)
**File:** `packages/less/src/less/less_go/parser.go` (Line 3043)

The parser regex pattern includes `@\{[\w-]+\}` as a valid selector character:
```regex
^(?:[.#]?|:*)(?:[\w-]|@\{[\w-]+\}|[^\x00-\x9f]|\\(?:[A-Fa-f0-9]{1,6} ?|[^A-Fa-f0-9]))+
```

This causes `@{inputs}` to be matched as a **STRING** rather than calling `VariableCurly()` to create a Variable node.

**Result:**
```go
Element{
  Combinator: nil,
  Value: "@{inputs}",  // String, not Variable node!
  IsVariable: false    // Because value is not a Variable node
}
```

#### Step 2: Element Evaluation Interpolates But Loses Variable Flag
**File:** `packages/less/src/less/less_go/element.go` (Lines 151-166, 207)

When `Element.Eval()` processes the string element:
```go
} else if strValue, ok := e.Value.(string); ok {
    // Check if string contains variable interpolation
    if strings.Contains(strValue, "@{") {
        quoted := NewQuoted("", strValue, true, e.GetIndex(), e.FileInfo())
        evaluated, err := quoted.Eval(context)
        // ... extracts string "input[type=text], input[type=email]"
    }
}
```

The Quoted node interpolates the variable correctly, but:

**Result:**
```go
Element{
  Value: "input[type=text], input[type=email]",  // Comma-separated string
  IsVariable: false  // Still false! Preserved from original
}
```

**Critical Issue:** The element contains multiple selectors (comma-separated) but `IsVariable=false` means the rest of the code doesn't know this needs special handling.

#### Step 3: Re-parsing Never Triggered (hasVariable Detection Fails)
**File:** `packages/less/src/less/less_go/ruleset.go` (Lines 360-378)

In `Ruleset.Eval()`, after evaluating selectors, the code checks whether selectors contain variables that need re-parsing:

```go
// Check for variables in elements
if selector, ok := evaluated.(*Selector); ok {
    for _, elem := range selector.Elements {
        if elem.IsVariable {  // <-- ONLY checks IsVariable flag
            hasVariable = true
            break
        }
    }
}
```

**Problem:** The element has `IsVariable = false` (it's a string that was interpolated), so the condition is never met.

**Result:**
- `hasVariable` remains `false`
- The re-parsing code (lines 384-431) is never executed
- The selector is NOT re-parsed to split "a, b" into multiple selectors

#### Step 4: Selector Joining Treats Comma-String as Single Element
**File:** `packages/less/src/less/less_go/ruleset.go` (Lines 1665+, in `replaceParentSelector`)

When `JoinSelector` processes the selector elements:
```go
for _, el := range inSelector.Elements {
    // el.Value = "input[type=text], input[type=email], ..."
    if valueStr, ok := el.Value.(string); !ok || valueStr != "&" {
        // This element is not "&", so it's added as-is
        currentElements = append(currentElements, el)
    }
}
```

The Element with comma-separated string value is treated as a **single selector element**. When the `&:focus` combinator is processed:
- The `&` triggers parent selector multiplication
- The `:focus` is appended to the entire combined path
- Due to how string concatenation works in CSS output, `:focus` only appears on the last part

**Result:** `"input[type=text], input[type=email]:focus"`

## Data Flow Visualization

```
Parsing:
  @{inputs} ──[Regex 3043]──> Element{Value: "@{inputs}", IsVariable: false}

Evaluation:
  Element{IsVariable: false} ──[Element.Eval()]──> Element{
    Value: "input[type=text], input[type=email]",
    IsVariable: false  (preserved but still false!)
  }

Re-parsing Check (Fails!):
  elem.IsVariable = false  ──[Line 366 check]──> hasVariable never set to true
                                                ──> Re-parsing code never executes
                                                ──> Selectors not split

Selector Joining:
  Element{Value: "input[type=text], input[type=email]"} ──[Treated as single element]──>
  CSS output: "input[type=text], input[type=email]:focus"  (WRONG!)
```

## Why hasVariable Detection Fails

The detection logic assumes that interpolated variables would create Elements with `IsVariable=true`, but:

1. When `@{inputs}` is parsed, it becomes a STRING, not a Variable node
2. IsVariable is only true for Variable node types
3. When the string is interpolated, the result is still a string with IsVariable=false
4. The hasVariable check only looks at the IsVariable flag, not at the string content

## The Solution

There are two robust approaches to fix this:

### Approach A: Mark Elements During Interpolation (Recommended)
**File:** `packages/less/src/less/less_go/element.go`

Modify Element.Eval() to set a flag when interpolation occurs:

```go
func (e *Element) Eval(context any) (any, error) {
    var evaluatedValue any = e.Value
    interpolated := false  // NEW: Track if interpolation occurred
    
    if e.Value != nil {
        // ... existing code ...
        } else if strValue, ok := e.Value.(string); ok {
            // Check if string contains variable interpolation
            if strings.Contains(strValue, "@{") {
                interpolated = true  // NEW: Mark as interpolated
                quoted := NewQuoted("", strValue, true, e.GetIndex(), e.FileInfo())
                evaluated, err := quoted.Eval(context)
                // ... existing code ...
            }
        }
    }
    
    newElement := NewElement(
        e.Combinator,
        evaluatedValue,
        e.IsVariable || interpolated,  // CHANGED: Include interpolation flag
        index,
        fileInfo,
        visibilityInfo,
    )
    
    return newElement, nil
}
```

This way, when a string element undergoes interpolation, the resulting Element will have `IsVariable=true`, triggering the existing re-parsing logic in Ruleset.Eval().

### Approach B: Enhance Re-parsing Detection
**File:** `packages/less/src/less/less_go/ruleset.go`

Modify the hasVariable detection to also check for comma-containing strings:

```go
// Check for variables in elements
if selector, ok := evaluated.(*Selector); ok {
    for _, elem := range selector.Elements {
        if elem.IsVariable {
            hasVariable = true
            break
        }
        // NEW: Also check if this is a string with commas (from interpolation)
        if strVal, ok := elem.Value.(string); ok && strings.Contains(strVal, ",") {
            hasVariable = true
            break
        }
    }
}
```

## Why Approach A is Better

1. **Cleaner:** Single point of change in element evaluation
2. **Complete:** Handles all interpolated selectors, not just those with commas
3. **Maintainable:** Preserves the existing hasVariable detection logic
4. **Robust:** Works with nested interpolations and edge cases

## Affected Test Cases

All test cases in `packages/test-data/less/_main/parse-interpolation.less`:

1. **Lines 1-7:** `@{inputs} { &:focus }` - Pseudo-class expansion broken
2. **Lines 9-15:** `@{classes} { + .z }` - Combinator expansion broken
3. **Lines 17-21:** `.bar { .d@{classes}&:hover }` - Parent + interpolation
4. **Lines 23-31:** Nested interpolations with combinators
5. **Lines 33-39:** `input { @{textClasses} }` - Parent nesting
6. **Lines 41-46:** `.master-page-1 { @{my-selector} }` - Parent nesting
7. **Lines 48-53:** `@{list} { .fruit-& }` - Parent reference with interpolation

## Impact Assessment

**Severity:** High
- Breaks all comma-separated selector interpolation cases
- Affects pseudo-classes, combinators, and parent selectors
- Cascades to ~25-30 other tests that depend on this functionality

**Scope:** 
- ~8 test cases in parse-interpolation alone
- Potentially affects color functions, mixin expansion, and other tests using selector interpolation

## Testing Checklist

After implementing the fix:

- [ ] Run `pnpm -w test:go:unit` - ensure no unit test regressions
- [ ] Run `pnpm -w test:go` - verify parse-interpolation passes all 8 cases
- [ ] Check selector-related tests pass (selector_test.go)
- [ ] Verify JoinSelectorVisitor tests still pass
- [ ] Test edge cases:
  - [ ] Nested interpolations: `@{a}:@{b}`
  - [ ] Complex selectors: `.bar > @{classes} + &:hover`
  - [ ] Escaped interpolations: `\@{var}` (should not interpolate)
- [ ] Ensure no regressions in related tests (colors, mixins, etc.)

## Related Files Reference

- **Parser:** `/home/user/less.go/packages/less/src/less/less_go/parser.go:3043`
- **Element:** `/home/user/less.go/packages/less/src/less/less_go/element.go:151-207`
- **Ruleset:** `/home/user/less.go/packages/less/src/less/less_go/ruleset.go:360-431`
- **Selector Joining:** `/home/user/less.go/packages/less/src/less/less_go/ruleset.go:1657-1944`
- **Test Data:** `/home/user/less.go/packages/test-data/less/_main/parse-interpolation.less`
- **Expected Output:** `/home/user/less.go/packages/test-data/css/_main/parse-interpolation.css`
