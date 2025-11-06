# Prompts for Fixing mixins-nested Issue

## Background
The `mixins-nested` test fails because the `.mix` MixinDefinition is being converted to a bare Ruleset between `Ruleset.Eval()` and `Ruleset.GenCSS()`, causing an extra empty ruleset `{ height: * 10; }` to appear in the output.

See `mixins-nested-issue.md` in this directory for full investigation details.

---

## Prompt 1: Find Where r.Rules Is Modified (Recommended First Step)

```
Fix the mixins-nested test by finding where Ruleset.Rules is being modified between Eval() and GenCSS().

PROBLEM:
The test file is at: packages/test-data/less/_main/mixins-nested.less
Run with: cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/mixins-nested$"

The output has an extra empty ruleset at the beginning:
{
  height:  * 10;
}
.class .inner {
  height: 300;
}
...

ROOT CAUSE IDENTIFIED:
After Ruleset.Eval() returns, the root ruleset has 4 rules:
  [0] MixinDefinition (.mix-inner)
  [1] MixinDefinition (.mix)
  [2] Ruleset (.class)
  [3] Ruleset (.class2)

But when Ruleset.GenCSS() is called, r.Rules has changed to 6 rules:
  [0] MixinDefinition (.mix-inner) ✓
  [1] Ruleset (was .mix MixinDefinition) ✗
  [2-5] Additional Rulesets

TASK:
1. Add logging to track when r.Rules is accessed/modified between these points
2. Check all code that might modify Ruleset.Rules after evaluation
3. Look in visitor implementations, tree transformations, or post-processing steps
4. Find where the MixinDefinition at index [1] is being replaced with a Ruleset
5. Prevent this modification or ensure MixinDefinitions are preserved

FILES TO CHECK:
- packages/less/src/less/less_go/ruleset.go (Eval and GenCSS methods)
- packages/less/src/less/less_go/mixin_definition.go
- Any visitor pattern implementations
- Any code that calls Ruleset.Accept() or processes rules

VERIFICATION:
The test should pass with perfect match - no empty ruleset should appear in output.
```

---

## Prompt 2: Fix the Embedded Ruleset Issue (Structural Fix)

```
Fix the mixins-nested test by changing how MixinDefinition embeds Ruleset.

PROBLEM:
Test: cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/mixins-nested$"
The test outputs an extra empty ruleset before the correct CSS.

ROOT CAUSE:
MixinDefinition embeds *Ruleset as an anonymous field:
```go
type MixinDefinition struct {
    *Ruleset  // This allows MixinDefinition to be treated as Ruleset!
    Name string
    Params []any
    ...
}
```

This embedding allows Go's type system to treat MixinDefinition as a Ruleset in type assertions, causing the MixinDefinition to be replaced with its embedded Ruleset somewhere in the pipeline.

TASK:
1. Change MixinDefinition to use a named field instead of embedding:
   ```go
   type MixinDefinition struct {
       Ruleset *Ruleset  // Named field instead of embedding
       Name string
       ...
   }
   ```

2. Update all code that accesses MixinDefinition fields to use md.Ruleset.FieldName instead of md.FieldName

3. Search for all places MixinDefinition accesses Ruleset fields and update them

4. Verify that MixinDefinitions can no longer be type-asserted to *Ruleset

FILES TO MODIFY:
- packages/less/src/less/less_go/mixin_definition.go (struct definition and all methods)
- Anywhere that accesses MixinDefinition fields that come from Ruleset

VERIFICATION:
Run: go test -v -run "TestIntegrationSuite/main/mixins-nested$"
Should pass with perfect match. Also verify mixins-important still passes.
```

---

## Prompt 3: Add Type Marker to Prevent Conversion (Quick Fix)

```
Fix the mixins-nested test by adding a marker to distinguish MixinDefinition Rulesets from regular Rulesets.

PROBLEM:
Test: cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/mixins-nested$"
Produces an extra empty ruleset in output.

ROOT CAUSE:
MixinDefinition embeds *Ruleset, and somewhere between Ruleset.Eval() and Ruleset.GenCSS(), the MixinDefinition is being accessed as a Ruleset and stored back into r.Rules, losing the MixinDefinition type.

QUICK FIX APPROACH:
1. Add a boolean flag to Ruleset to mark it as belonging to a MixinDefinition:
   ```go
   type Ruleset struct {
       ...
       IsMixinDefinitionRuleset bool  // Add this field
   }
   ```

2. In MixinDefinition.NewMixinDefinition(), set this flag:
   ```go
   ruleset := NewRuleset([]any{selector}, rules, false, visibilityInfo)
   ruleset.IsMixinDefinitionRuleset = true  // Mark it
   ```

3. In Ruleset.GenCSS(), skip rulesets marked as mixin definitions:
   ```go
   if r.Rules != nil {
       for i, rule := range r.Rules {
           if rs, ok := rule.(*Ruleset); ok && rs.IsMixinDefinitionRuleset {
               continue  // Skip mixin definition rulesets
           }
           ...
       }
   }
   ```

FILES TO MODIFY:
- packages/less/src/less/less_go/ruleset.go (add field, check in GenCSS)
- packages/less/src/less/less_go/mixin_definition.go (set flag in constructor)

VERIFICATION:
Test should pass. This is a quick fix - the root issue still exists but output will be correct.
```

---

## Prompt 4: Trace the Visitor Pattern (Deep Investigation)

```
Investigate the visitor pattern to find where MixinDefinitions are being converted to Rulesets.

CONTEXT:
The mixins-nested test fails because a MixinDefinition is replaced with a Ruleset.
Investigation shows r.Rules changes from 4 to 6 rules between Eval() and GenCSS().
See: .claude/investigation/mixins-nested-issue.md

TASK:
1. Find all visitor implementations in the codebase:
   ```bash
   cd packages/less/src/less/less_go
   grep -r "Accept\|Visit" *.go | grep -E "func.*Visit|\.Accept\("
   ```

2. Check if Ruleset.Accept() is called between Eval() and GenCSS()

3. Look for any visitor that might:
   - Access the Rules array
   - Replace MixinDefinitions with Rulesets
   - Flatten or transform the rule tree

4. Add debug logging to track:
   - When Accept() is called
   - What visitors are applied
   - When r.Rules is modified

5. Find the specific visitor/transformation that's causing the issue

6. Fix it by either:
   - Skipping MixinDefinitions in that visitor
   - Preserving the MixinDefinition type
   - Preventing the problematic transformation

FILES TO INVESTIGATE:
- All files matching: packages/less/src/less/less_go/*visitor*.go
- packages/less/src/less/less_go/ruleset.go (Accept method)
- Any code calling ruleset.Accept()

DEBUGGING APPROACH:
Add logging like this in Ruleset.Accept():
```go
func (r *Ruleset) Accept(visitor any) {
    if os.Getenv("LESS_GO_DEBUG") == "1" {
        fmt.Fprintf(os.Stderr, "Accept called on ruleset with %d rules\n", len(r.Rules))
    }
    // ... existing code
}
```

VERIFICATION:
Find the exact line of code causing the conversion, then fix it so the test passes.
```

---

## Prompt 5: Compare with JavaScript Implementation

```
Fix mixins-nested by comparing with the JavaScript Less.js implementation.

PROBLEM:
Test: cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/mixins-nested$"
Extra empty ruleset appears in output.

TASK:
1. Find the JavaScript implementation of Ruleset evaluation and CSS generation:
   - File: packages/less/src/less/tree/ruleset.js
   - Look at the `eval()` and `genCSS()` methods

2. Find the JavaScript MixinDefinition implementation:
   - File: packages/less/src/less/tree/mixin-definition.js

3. Compare how JavaScript handles:
   - Storing mixin definitions in the rules array
   - Evaluating rulesets containing mixin definitions
   - Filtering out mixin definitions during CSS generation
   - Preventing mixin definitions from appearing in output

4. Identify what the Go implementation is doing differently

5. Port the correct JavaScript logic to Go

SPECIFIC AREAS TO CHECK:
- How does JavaScript distinguish MixinDefinitions from Rulesets in the rules array?
- Does JavaScript have any special handling in ruleset.eval() for mixin definitions?
- How does JavaScript ensure mixin definitions don't generate CSS?
- Are there any type checks or filters in the JavaScript genCSS() method?

VERIFICATION:
The Go implementation should match JavaScript behavior - mixin definitions should be evaluated but not output.
```

---

## How to Use These Prompts

1. **Start with Prompt 1**: This is the investigative approach to find exactly where the bug occurs
2. **Try Prompt 3**: If you need a quick fix while investigating the deeper issue
3. **Use Prompt 2**: For a structural fix if the embedded Ruleset is confirmed as the core problem
4. **Apply Prompt 4**: For deep tracing when the issue is in the visitor pattern
5. **Reference Prompt 5**: To understand the intended behavior from the JavaScript implementation

Each prompt is self-contained and includes:
- Clear problem description
- Root cause identified from investigation
- Specific tasks to complete
- Files to check/modify
- How to verify the fix

All prompts can be given to a fresh LLM session with no additional context needed.
