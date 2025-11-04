# Namespace Resolution Issues - Agent Task

## 🎯 Mission
Fix 2 failing tests related to namespace resolution and variable calls

## 📊 Status
- **Tests Failing**: 2
- **Priority**: High (Wave 1 - Independent)
- **Complexity**: Medium
- **Independence**: HIGH - Can be fixed in parallel with other issues

## ❌ Failing Tests

### 1. namespacing-6
**File**: `packages/test-data/less/namespacing/namespacing-6.less`
**Error**: `Could not evaluate variable call @alias`

**Test Content**:
```less
.wrapper(@another-mixin) {
  @another-mixin();  // Call mixin passed as parameter
}

.something(foo) {
  width: 10px;
}

.output-height() {
  height: 10px;
}

.rule-1 {
  @alias: .something(foo);  // Assign mixin call to variable
  @alias();  // Call the variable (detached ruleset call)
}

.rule-2 {
  @alias: .something(foo);
  .wrapper(@alias);  // Pass variable to another mixin
}

.rule-3 {
  .wrapper(.something(foo));  // Pass mixin call directly
  .wrapper(.output-height());
}
```

**Problem**: Variable `@alias` is assigned a mixin call `.something(foo)`, then called as `@alias()`. The error "Could not evaluate variable call @alias" suggests:
1. Mixin calls assigned to variables are not being stored correctly
2. OR variable calls are not being evaluated properly
3. OR the evaluation is returning an unevaluated node that can't be called

**This is similar to Issue #2 (detached-rulesets) which was FIXED** - but this test still fails.

---

### 2. namespacing-functions
**File**: `packages/test-data/less/namespacing/namespacing-functions.less`
**Error**: `Could not evaluate variable call @dr`

**Problem**: Similar to namespacing-6, but involves function calls assigned to variables.

---

## 🔍 Root Cause Analysis

### Theory 1: Mixin Call Assignment Not Evaluated
When a mixin call like `.something(foo)` is assigned to a variable `@alias`, it might be:
- Stored as an unevaluated Call node
- Not converted to a DetachedRuleset
- Missing the proper evaluation context

### Theory 2: Variable Call Evaluation Issue
When `@alias()` is called:
- The variable lookup succeeds
- But the returned value can't be called
- Might be a type issue (not recognizing it as callable)

### Theory 3: Frame Scoping
Similar to Issue #2b (functions-each), the variable might not be accessible in the calling context.

---

## 🔍 Go Files to Investigate

### Primary Files
1. **`variable_call.go`** - Variable call implementation
   - Check how variable calls are evaluated
   - Verify what types can be called
   - Look for error "Could not evaluate variable call"

2. **`variable.go`** - Variable evaluation
   - Check what happens when a mixin call is assigned
   - Verify mixin calls are being evaluated before storage
   - Check if result is callable

3. **`mixin_call.go`** - Mixin call evaluation
   - Check if mixin calls can be stored in variables
   - Verify they return DetachedRulesets when needed

4. **`declaration.go`** - Variable declaration handling
   - Check how values are stored in declarations
   - Verify evaluation happens at declaration time

### Secondary Files
5. **`detached_ruleset.go`** - DetachedRuleset functionality
   - Check if mixin calls return DetachedRulesets
   - Verify CallEval works correctly

6. **`ruleset.go`** - Ruleset evaluation and variable lookup
   - Check variable resolution in frames
   - Verify callable detection

---

## 📚 JavaScript Reference Files

Study these to understand correct behavior:

1. **`packages/less/src/less/tree/variable-call.js`**
   - How variable calls work
   - What can be called
   - Error handling

2. **`packages/less/src/less/tree/variable.js`**
   - Variable evaluation
   - How mixin calls are stored

3. **`packages/less/src/less/tree/mixin-call.js`**
   - Mixin call evaluation
   - How results are returned

4. **`packages/less/src/less/tree/detached-ruleset.js`**
   - DetachedRuleset structure
   - Calling mechanism

---

## ✅ Success Criteria

### Minimum Success (1/2 tests)
- `namespacing-6` - Basic variable call working

### Target Success (2/2 tests)
- `namespacing-6` - Variable calls to mixin results
- `namespacing-functions` - Variable calls to function results

---

## 🚫 Constraints

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target tests**: `pnpm -w test:go:filter -- "namespacing-6"`
4. **No regressions**: All currently passing tests must still pass

---

## 🧪 Testing Strategy

### Run Specific Tests
```bash
# Test individual cases
go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v
go test -run "TestIntegrationSuite/namespacing/namespacing-functions" -v

# With debug output
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v

# With trace (very useful for this issue!)
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v
```

### Verify No Regressions
```bash
# Run all unit tests
pnpm -w test:go:unit

# Run full integration suite
pnpm -w test:go:summary
```

---

## 📝 Expected Changes

### Likely Changes Needed

1. **mixin_call.go** - Ensure mixin calls return DetachedRulesets when assigned to variables
   - Check EvalCall method
   - Verify proper DetachedRuleset wrapping

2. **variable.go** - Ensure mixin call results are evaluated
   - Similar to Issue #2 fix (check Eval signature order)
   - Make sure Expression nodes with mixin calls are evaluated

3. **variable_call.go** - Ensure variable calls can handle DetachedRulesets
   - Check type assertions
   - Verify CallEval is being called

4. **declaration.go** - Ensure variable declarations evaluate their values
   - Check if mixin calls are evaluated at declaration time
   - Verify proper storage of evaluated results

### Testing Pattern

For each fix:
1. Add trace output to understand flow
2. Make minimal change
3. Run specific test with LESS_GO_TRACE=1
4. If passing, run unit tests
5. If unit tests pass, run full integration suite
6. Commit with clear message

---

## 🎯 Debugging Hints

### Add Trace Output
```go
// In variable_call.go
fmt.Printf("[VARCALL-DEBUG] Variable call: @%s, value type: %T\n",
    vc.Variable, value)

// In mixin_call.go
fmt.Printf("[MIXIN-DEBUG] Mixin call result type: %T\n", result)

// In variable.go
fmt.Printf("[VAR-DEBUG] Variable %s assigned value type: %T\n",
    v.Name, evaluatedValue)
```

### Use LESS_GO_TRACE
```bash
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v 2>&1 | grep -A5 -B5 "@alias"
```

This will show you:
- When @alias is declared
- What value is stored
- When @alias() is called
- What type is being called

---

## 📊 Comparison with Fixed Issues

### Issue #2 (detached-rulesets) - FIXED
Similar problem: "Could not evaluate variable call @ruleset"

**Fix Applied**:
- Changed Eval signature checking order in `mixin_definition.go`
- Check `Eval(any) (any, error)` BEFORE `Eval(any) any`
- This ensured Expression nodes were evaluated

**Possible Connection**: The same fix might be needed elsewhere, or a similar pattern applies.

---

## 📊 Estimated Impact

- **Tests Fixed**: 2
- **Other Tests Potentially Improved**: 3-5 tests using variable calls or namespacing
- **Risk Level**: Low-Medium - Variable calls are used but this is a specific pattern

---

## 🔄 Iteration Strategy

### Round 1: Understand the Flow
1. Run with LESS_GO_TRACE=1
2. Identify where evaluation fails
3. Compare with JavaScript behavior

### Round 2: Implement Fix
1. Apply fix based on understanding
2. Test with namespacing-6
3. If passing, test namespacing-functions

### Round 3: Verify and Commit
1. Run unit tests
2. Run full integration suite
3. Commit and push

---

## 📋 Commit Message Template

```
Fix namespace variable call evaluation

When mixin calls like `.something(foo)` are assigned to variables and then
called as `@alias()`, the evaluation was failing with "Could not evaluate
variable call @alias".

Root cause: [Describe root cause found]

Fix: [Describe fix applied]

Tests fixed:
- namespacing-6: ✅
- namespacing-functions: ✅

Related to Issue #2 (detached-rulesets) which had similar symptoms.
```

---

## 🚀 When Done

1. **Commit** to branch: `claude/fix-namespacing-<your-session-id>`
2. **Push** to remote: `git push -u origin claude/fix-namespacing-<your-session-id>`
3. **Report**: "Fixed 2/2 namespace tests: namespacing-6, namespacing-functions"

---

## 💡 Key Insights

1. **Parser is correct** - 92.4% compilation rate
2. **Similar to Issue #2** - Variable call evaluation
3. **Focus on evaluation chain** - Declaration → Storage → Variable Lookup → Call
4. **Use LESS_GO_TRACE** - Essential for debugging this issue
5. **Likely a type or evaluation order issue** - Not missing functionality

---

## 🔗 Related Issues

- **Issue #2** (detached-rulesets) - FIXED - Similar variable call issue
- **Issue #2b** (functions-each) - FIXED - Variable scope issue
- Both might provide patterns applicable here

---

## ⚠️ Special Notes

1. These tests are DIFFERENT from detached-rulesets even though the error is similar
2. The issue is specific to mixin calls assigned to variables
3. DetachedRuleset functionality is working (tested in other tests)
4. Focus on how mixin call results are stored in variables
5. The evaluation context and frames are likely correct (fixed in Issue #2)
