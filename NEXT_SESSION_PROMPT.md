# Prompt for Next LLM Session: Complete Directive Bubbling Fix

## Task Overview
Fix the remaining selector joining issue in the `directives-bubling` test. The directive extraction is working, but selectors inside bubbled `@supports`/`@document` directives are not being flattened/merged correctly.

## What's Already Done ✅

1. **Root cause identified**: @supports/@document use a visitor-based bubbling mechanism (NOT mediaBlocks like @media)
   - Parser sets `isRooted=false` ✅ (parser.go:1830-1832)
   - `GetIsRooted()` method added to AtRule ✅ (atrule.go:113-116)
   - ToCSSVisitor extraction works ✅ (to_css_visitor.go:622-629) - directives DO move to top level

2. **Comprehensive documentation**: See `DIRECTIVE_BUBBLING_STATUS.md` for full analysis

3. **No regressions**: All tests passing
   - Unit tests: ✅ 2,290+ passing
   - Integration: ✅ 74 perfect matches (baseline maintained)

## The Remaining Problem ❌

Selectors inside bubbled directives are NOT being flattened.

**Current output:**
```css
@document url-prefix() {
  .parent {
      .child {
        color: red;
      }
    }
  }
}
```

**Expected output:**
```css
@document url-prefix() {
  .parent .child {
    color: red;
  }
}
```

## Where to Focus

**File**: `packages/less/src/less/less_go/join_selector_visitor.go`

### Investigation Steps

1. **Add debug logging to JoinSelectorVisitor**:
   ```go
   // In VisitRuleset, before joinSelectors call:
   fmt.Printf("[DEBUG] Context stack depth: %d, Context: %v\n", len(jsv.contexts), jsv.contexts)
   fmt.Printf("[DEBUG] Ruleset root=%v, selectors=%v\n", rulesetInterface.GetRoot(), selectors)
   ```

2. **Test with simple case**:
   ```less
   .parent {
     @document url-prefix() {
       .child {
         color: red;
       }
     }
   }
   ```

3. **Key questions to answer**:
   - When visiting the `.child` ruleset (inside @document), what's in the context stack?
   - Is the context stack empty? (It should contain `.parent`)
   - Is `root` being set to `false` on `.child`'s ruleset?
   - Is `JoinSelectors()` being called?

### Likely Root Cause

The JoinSelectorVisitor may not be properly managing the context stack when visiting AtRules. Compare these two scenarios:

**Working case (nested rulesets)**:
```
.parent { .child {} }
→ Context has [.parent] when visiting .child
→ root=false on .child
→ joinSelectors merges: .parent + .child → .parent .child ✅
```

**Broken case (AtRule with nested ruleset)**:
```
.parent { @document { .child {} } }
→ Context may be empty when visiting .child? 🤔
→ root may be true on .child?
→ joinSelectors not called or not merging correctly ❌
```

### Possible Fixes

**Option 1**: Modify `VisitAtRule` to push/pop context
```go
func (jsv *JoinSelectorVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) any {
    // Current code sets root...

    // NEW: For non-rooted AtRules, we may need to maintain context?
    // But AtRules don't have selectors themselves...
    // The CHILD ruleset needs the PARENT context
}
```

**Option 2**: Ensure AtRule.Accept properly visits rules with context
- AtRule.Accept already calls `v.VisitArray(a.Rules)` ✅
- But does VisitArray properly descend and maintain context?

**Option 3**: The ruleset inside AtRule has `root=true` when it should be `false`
- Check in `VisitAtRule`: After setting root, is it actually applied?
- Debug: `fmt.Printf("After SetRoot: %v\n", rules[0].GetRoot())`

### JavaScript Reference Code

**File**: `packages/less/src/less/visitors/join-selector-visitor.js:53-58`
```javascript
visitAtRule(atRuleNode, visitArgs) {
    const context = this.contexts[this.contexts.length - 1];
    if (atRuleNode.rules && atRuleNode.rules.length) {
        atRuleNode.rules[0].root = (atRuleNode.isRooted || context.length === 0 || null);
        // If isRooted=false AND context.length > 0, then root=null (falsy)
        // This triggers selector joining!
    }
}
```

**Key insight**: In JavaScript, when `root=null` (falsy), the ruleset's selectors get joined with parent context.

Check if our Go code is doing the same:
```go
// Our current code (join_selector_visitor.go:288-290):
if isRootedInterface.GetIsRooted() || len(context) == 0 {
    rootValue = true
}
// else rootValue stays nil

// This looks correct! But is SetRoot actually being called?
// And is the ruleset's VisitRuleset checking root correctly?
```

## Test Commands

```bash
# Run the specific test
pnpm -w test:go:filter -- "directives-bubling"

# Run with Go debug output
cd packages/less
go test -v -run "TestIntegrationSuite/main/directives-bubling" ./src/less/less_go

# Test simple case
cat > /tmp/test.less << 'EOF'
.parent {
  @document url-prefix() {
    .child {
      color: red;
    }
  }
}
EOF

go run << 'GOEOF'
package main
import (
    "fmt"
    less "github.com/toakleaf/less.go/packages/less/src/less/less_go"
    "os"
)
func main() {
    input, _ := os.ReadFile("/tmp/test.less")
    factory := less.Factory(nil, nil)
    renderFunc := factory["render"].(func(string, ...any) any)
    result := renderFunc(string(input), map[string]any{})
    fmt.Print(result)
}
GOEOF
```

## Success Criteria

✅ Test `directives-bubling` shows "Perfect match!"
✅ All 74 existing perfect matches maintained (no regressions)
✅ All unit tests pass (2,290+)

## Files to Modify

Primary:
- `packages/less/src/less/less_go/join_selector_visitor.go` (likely lines 273-317)

Possibly:
- `packages/less/src/less/less_go/ruleset.go` (JoinSelectors method if needed)

Do NOT modify:
- `atrule.go` (GetIsRooted already added)
- `parser.go` (isRooted setting already correct)
- `to_css_visitor.go` (extraction already working)

## Context Files to Review

- `DIRECTIVE_BUBBLING_STATUS.md` - Full analysis of the issue
- `packages/less/src/less/visitors/join-selector-visitor.js` - JavaScript reference
- `packages/less/src/less/tree/ruleset.js:567-658` - JavaScript joinSelectors method

## Git Branch

Branch: `claude/fix-directive-bubbling-011CUyDgJT8iEgWahUYV6rpr`

Latest commit: "WIP: Partial fix for directive bubbling (@supports/@document)"

## Recommended Approach

1. Start by adding detailed debug logging to `VisitAtRule` and `VisitRuleset`
2. Run the simple test case and analyze the output
3. Compare with what happens for working nested rulesets (without @document)
4. Identify where the context is being lost or root is being set incorrectly
5. Implement the fix
6. Verify with full test suite
7. Commit and push

The fix is likely small (5-15 lines) once you identify where the issue is. The challenge is understanding the visitor pattern flow.

Good luck! 🚀
