# Selector Interpolation Bug Investigation
## Parse-Interpolation Test Failure Analysis

### Executive Summary
The parse-interpolation test fails because selector interpolation doesn't properly expand comma-separated selectors when applying pseudo-classes or combinators. When `@{inputs}` contains multiple selectors like `"input[type=text], input[type=email]"`, only the last selector gets the following pseudo-class/combinator.

Example failure:
- Input: `@inputs: input[type=text], input[type=email];` followed by `@{inputs} { &:focus { foo: bar; } }`
- Expected: `input[type=text]:focus, input[type=email]:focus { foo: bar; }`
- Actual: `input[type=text], input[type=email]:focus { foo: bar; }`

### Root Cause

#### 1. Parser Behavior (parser.go:3043)
When parsing a selector containing `@{inputs}`, the parser uses a regex pattern that includes `@\{[\w-]+\}` as a valid element character:

```
Regex: ^(?:[.#]?|:*)(?:[\w-]|@\{[\w-]+\}|[^\x00-\x9f]|\\(?:[A-Fa-f0-9]{1,6} ?|[^A-Fa-f0-9]))+
```

This causes `@{inputs}` to be matched as a **STRING element**, not as a `VariableCurly` node.

Result: 
- Element.Value = "@{inputs}" (string)
- Element.IsVariable = false (line 3112-3114 shows IsVariable is true only for Variable nodes)

#### 2. Element Evaluation (element.go:151-166)
When Element.Eval() processes the string element "@{inputs}":

```go
} else if strValue, ok := e.Value.(string); ok {
    // Check if string contains variable interpolation
    if strings.Contains(strValue, "@{") {
        // Create a Quoted node to handle interpolation
        quoted := NewQuoted("", strValue, true, e.GetIndex(), e.FileInfo())
        evaluated, err := quoted.Eval(context)
        if err != nil {
            return nil, err
        }
        // Extract the string value from the evaluated Quoted node
        if quotedResult, ok := evaluated.(*Quoted); ok {
            evaluatedValue = quotedResult.value
        }
    }
}
```

The Quoted node is evaluated, which interpolates `@{inputs}` → `"input[type=text], input[type=email]"`.

Result:
- evaluatedValue = "input[type=text], input[type=email]" (string)
- Element.IsVariable = false (preserved from line 207, but was originally false)

#### 3. Missing Re-parse Detection (ruleset.go:360-378)
In Ruleset.Eval(), after evaluating selectors, the code checks for elements that need re-parsing:

```go
// Check for variables in elements
if selector, ok := evaluated.(*Selector); ok {
    for _, elem := range selector.Elements {
        if elem.IsVariable {  // <-- CHECKS IsVariable FLAG
            hasVariable = true
            break
        }
    }
}
```

**Problem**: The element has `IsVariable = false` because it's not a Variable node. It's a string element that was interpolated.

Result:
- `hasVariable` is never set to true
- The re-parsing code (lines 384-431) is never executed
- Selector remains: `Element{Value="input[type=text], input[type=email]", IsVariable=false}`
- This single element with comma-separated selectors is not split into multiple selectors

#### 4. Selector Joining (ruleset.go:1665+)
When JoinSelector processes elements during selector joining:

```go
for _, el := range inSelector.Elements {
    // el.Value = "input[type=text], input[type=email]"
    if valueStr, ok := el.Value.(string); !ok || valueStr != "&" {
        // This element is not "&", so it gets added as a single element
        currentElements = append(currentElements, el)
    }
    // ... continues processing
}
```

The element with comma-separated string is treated as a single element, so when `&:focus` is processed:
- `&` causes parent multiplication
- `:focus` is appended to the last element in the join

Result: `:focus` only appears on the last selector in the string.

### Comparison with Expected Behavior

The JavaScript implementation likely:
1. Either parses `@{inputs}` as a Variable node (not as a string), making `IsVariable = true`
2. OR detects comma-separated strings after interpolation and triggers re-parsing
3. OR performs the re-parsing check differently (checking for string content, not just IsVariable flag)

### Why This Matters

This bug affects all cases where:
- A variable containing comma-separated selectors is interpolated into a selector
- Any pseudo-class (`:focus`, `:hover`, etc.) or combinator is applied after
- Parent selectors (`.bar .d@{classes}`) are combined with interpolated selectors
- Nested interpolations with combinators are used

### Affected Test Cases in parse-interpolation.less

1. **Line 3-7**: `@{inputs}:focus` - :focus only on last input
2. **Line 11-15**: `@{classes} + .z` - `+ .z` only on last class  
3. **Line 18-21**: `.bar .d@{classes}&:hover` - &:hover expansion broken
4. **Line 36-39**: `input { @{textClasses} { ... } }` - Parent joining broken
5. **Line 43-45**: `.master-page-1 { @{my-selector} { ... } }` - Parent joining broken
6. **Line 49-53**: `@{list} .fruit-&` - & expansion broken with interpolation

### The Fix Strategy

The most robust fix is to enhance the variable detection in Ruleset.Eval() (around line 366):

**Current Logic** (line 366):
```go
if elem.IsVariable {
    hasVariable = true
    break
}
```

**Improved Logic** should also detect:
1. When an element contains a string that came from interpolating a variable containing commas
2. By checking if the element value is a string AND:
   - The original element had a string value containing "@{"
   - OR the evaluated string value contains a comma (indicating multiple selectors)

Alternatively, set a flag on the Element during Element.Eval() when interpolation occurs, to preserve the knowledge that this element was interpolated.

### Files That Need Changes

1. **element.go** - Element.Eval() method (lines 137-214)
   - Need to mark elements that underwent interpolation
   
2. **ruleset.go** - Ruleset.Eval() method (lines 360-378)
   - Need to enhance hasVariable detection to catch interpolated strings with commas

3. **parser.go** - Optional: Parser.Element() method (line 3043)
   - Consider whether to parse `@{var}` as Variable or String
   - Current approach (String) requires more post-evaluation detection

### Testing Approach

1. Run `pnpm -w test:go` and check parse-interpolation test
2. Verify all 8 test cases in parse-interpolation.less produce correct output
3. Ensure no regressions in other selector tests (selector_test.go, etc.)
4. Test with complex nesting scenarios
