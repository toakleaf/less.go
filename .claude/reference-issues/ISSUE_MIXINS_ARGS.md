# Mixin Argument Expansion Issues - Agent Task

## 🎯 Mission
Fix 2 failing tests related to mixin argument expansion with the `...` operator

## 📊 Status
- **Tests Failing**: 2 (same test file in different math mode suites)
- **Priority**: Medium (Wave 2 - May touch shared files)
- **Complexity**: Medium
- **Independence**: MEDIUM - Touches mixin code that other agents might work on

## ❌ Failing Tests

### 1. mixins-args (math-parens suite)
**File**: `packages/test-data/less/math/strict/mixins-args.less`
**Error**: `No matching definition was found for '.m3()'`

### 2. mixins-args (math-parens-division suite)
**File**: `packages/test-data/less/math/parens-division/mixins-args.less`
**Error**: Same as above (same file content)

**Failing Section** (lines 214-247):
```less
mixins-args-expand-op- {
  @x: 1, 2, 3;
  @y: 4  5  6;

  &1 {.m3(@x...)}        // Expand @x into 3 arguments
  &2 {.m3(@y...)}        // Expand @y into 3 arguments
  &3 {.wr(a, b, c)}
  &4 {.wr(a; b; c, d)}
  &5 {.wr(@x...)}        // Expand in variadic mixin
  &6 {.m4(0; @x...)}     // Expand after positional arg
  &7 {.m4(@x..., @a: 0)} // Expand before named arg
  &8 {.m4(@b: 1.5; @x...)} // Expand after named arg
  &9 {.aa(@y, @x..., and again, @y...)} // Multiple expansions

  .m3(@a, @b, @c) {
    m3: @a, @b, @c;
  }

  .m4(@a, @b, @c, @d) {
    m4: @a, @b, @c, @d;
  }

  .wr(@a...) {
    &a {.m3(@a...)}      // Re-expand collected args
    &b {.m4(0, @a...)}
    &c {.m4(@a..., 4)}
  }

  .aa(@a...) {
    aa: @a;
    a4: extract(@a, 5);
    a8: extract(@a, 8);
  }
}
```

**Problem**: The error "No matching definition was found for `.m3()`" suggests:
1. The `...` expansion operator is not working
2. `@x...` is not being expanded into `1, 2, 3`
3. The mixin is being called with no arguments instead of 3 arguments
4. OR the mixin matching logic doesn't see the expanded arguments

---

## 🔍 Root Cause Analysis

### Theory 1: Expansion Not Happening
When `.m3(@x...)` is called:
- Parser might create a Call with `@x...` as a single argument
- The `...` operator is not being recognized
- The variable is not being expanded into multiple arguments
- Result: `.m3()` called with 1 arg (the variable) instead of 3 args (1, 2, 3)

### Theory 2: Expansion Timing Issue
The expansion might need to happen at:
- Parse time (unlikely)
- Call time (likely)
- Evaluation time (very likely)

If expansion isn't happening at the right time, the mixin matcher sees 1 argument when it should see 3.

### Theory 3: Rest Parameter Collection Issue
Lines like `.wr(@a...)` collect variadic arguments. Then `.m3(@a...)` should re-expand them. This suggests:
- Rest parameter collection works (since `.wr` is defined)
- But re-expansion might not work

---

## 🔍 Go Files to Investigate

### Primary Files
1. **`mixin_call.go`** - Mixin call evaluation and argument processing
   - Look for `...` operator handling
   - Check argument expansion logic
   - Find mixin matching code

2. **`mixin_definition.go`** - Mixin definition and parameter matching
   - Check how arguments are matched to parameters
   - Look for variadic parameter handling (`@rest...`)
   - Check EvalParams method

3. **`expression.go`** - Expression evaluation
   - `@x...` might be parsed as an Expression with a flag
   - Check if there's a "expand" or "spread" flag
   - Look for list expansion logic

4. **`value.go`** - Value lists
   - Lists like `1, 2, 3` are Value nodes
   - Check how they're expanded
   - Look for spread operator handling

### Secondary Files
5. **`parser.go`** - Parser
   - Check how `...` is parsed
   - Look for "spread" or "rest" parameter syntax
   - Verify it's creating the right AST structure

6. **`call.go`** - General call handling
   - Might have generic argument expansion logic

---

## 📚 JavaScript Reference Files

Study these to understand correct behavior:

1. **`packages/less/src/less/tree/mixin-call.js`**
   - Look for `...` operator handling
   - Check `args` expansion
   - Find argument matching logic

2. **`packages/less/src/less/tree/mixin-definition.js`**
   - Parameter matching
   - Variadic parameters (`...args`)
   - How rest parameters work

3. **`packages/less/src/less/parser/parser.js`**
   - How `...` is parsed
   - AST structure for spread operator

---

## ✅ Success Criteria

### Minimum Success (1/2 tests)
- `mixins-args` (math-parens suite) - Argument expansion works

### Target Success (2/2 tests)
- `mixins-args` (math-parens suite) - All expansion patterns work
- `mixins-args` (math-parens-division suite) - Same fixes apply

---

## 🚫 Constraints

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target tests**: `pnpm -w test:go:filter -- "mixins-args"`
4. **No regressions**: All currently passing tests must still pass
5. **Be careful with mixin code**: Other agents might work on mixins too

---

## 🧪 Testing Strategy

### Create Minimal Test Case
```bash
cat > /tmp/test-expand.less << 'EOF'
.m3(@a, @b, @c) {
  result: @a, @b, @c;
}

.test {
  @x: 1, 2, 3;

  /* This should expand @x into three arguments */
  .m3(@x...);

  /* Expected output: result: 1, 2, 3; */
}
EOF

# Test it
go run cmd/lessc/lessc.go /tmp/test-expand.less
```

### Run Specific Tests
```bash
# Test individual cases
go test -run "TestIntegrationSuite/math-parens/mixins-args" -v
go test -run "TestIntegrationSuite/math-parens-division/mixins-args" -v

# With debug output
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/math-parens/mixins-args" -v

# With trace
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/math-parens/mixins-args" -v
```

### Verify No Regressions
```bash
# Run all unit tests
pnpm -w test:go:unit

# Check other mixin tests still pass
go test -run "TestIntegrationSuite/main/mixins" -v

# Run full integration suite
pnpm -w test:go:summary
```

---

## 📝 Expected Changes

### Likely Changes Needed

1. **mixin_call.go** - Implement argument expansion
   - Detect when an argument has `...` operator
   - If argument is a list/value, expand it into multiple arguments
   - Pass expanded arguments to mixin matcher

2. **mixin_definition.go** - Handle expanded arguments
   - Ensure parameter matching works with expanded args
   - Variadic parameter collection should work
   - Re-expansion of collected args should work

3. **expression.go** or **value.go** - Expansion helper
   - Might need a method to expand lists into individual values
   - Check if there's a flag for "spread" operation

4. **Parser changes** (less likely)
   - Parser probably already handles `...` syntax
   - Might just need to set a flag on the node

### Testing Pattern

For each fix:
1. Test with minimal case first
2. Gradually add complexity (nested expansion, multiple expansions)
3. Run specific test
4. Run unit tests
5. Run related mixin tests
6. Run full integration suite
7. Commit with clear message

---

## 🎯 Debugging Hints

### Understand the AST
```go
// Add debug output in mixin_call.go
fmt.Printf("[MIXIN-CALL-DEBUG] Call to: %v, num args: %d\n",
    mc.Selector, len(mc.Arguments))
for i, arg := range mc.Arguments {
    fmt.Printf("[MIXIN-CALL-DEBUG] Arg %d type: %T, value: %+v\n",
        i, arg, arg)
}
```

### Check if ... is Parsed
```go
// Check if there's a "spread" or "variadic" flag
fmt.Printf("[MIXIN-CALL-DEBUG] Arg has spread: %v\n", arg.IsSpread)
```

### Test JavaScript Behavior
```bash
# Create test file
cat > /tmp/test.less << 'EOF'
.m3(@a, @b, @c) { m3: @a, @b, @c; }
.test { @x: 1, 2, 3; .m3(@x...); }
EOF

# Run with JavaScript less.js
npx lessc /tmp/test.less

# Expected output:
# .test { m3: 1, 2, 3; }
```

---

## 📊 Estimated Impact

- **Tests Fixed**: 2 (same test in 2 suites)
- **Other Tests Potentially Improved**: Unknown - depends on how many tests use `...`
- **Risk Level**: Medium - Mixin code is central, but `...` is a specific feature

---

## 🔄 Iteration Strategy

### Round 1: Understand Current Behavior
1. Run with LESS_GO_TRACE=1
2. See what arguments are passed to .m3
3. Check if `...` is being parsed
4. Compare with JavaScript AST structure

### Round 2: Implement Expansion
1. Find where arguments are processed in mixin_call.go
2. Add logic to detect and expand `...` arguments
3. Test with minimal case
4. Test with full mixins-args.less

### Round 3: Handle Complex Cases
1. Test nested expansion (line 237: `.wr(@a...)` where `.wr` uses `@a...` again)
2. Test multiple expansions (line 226: `@y, @x..., and again, @y...`)
3. Test mixed positional and named args with expansion

### Round 4: Verify and Commit
1. Run all mixin-related tests
2. Run unit tests
3. Run full integration suite
4. Commit and push

---

## 📋 Commit Message Template

```
Fix mixin argument expansion with ... operator

When calling mixins with expanded arguments like `.m3(@x...)` where
@x is a list (1, 2, 3), the expansion wasn't happening. The mixin
matcher received 1 argument (the variable) instead of 3 (1, 2, 3).

Root cause: [Describe where expansion should happen but wasn't]

Fix:
- Detect arguments with ... operator
- Expand list values into multiple arguments
- Handle re-expansion of collected variadic arguments
- Support mixed positional, named, and expanded arguments

Tests fixed:
- mixins-args (math-parens): ✅
- mixins-args (math-parens-division): ✅
```

---

## 🚀 When Done

1. **Commit** to branch: `claude/fix-mixins-args-<your-session-id>`
2. **Push** to remote: `git push -u origin claude/fix-mixins-args-<your-session-id>`
3. **Report**: "Fixed 2/2 mixins-args tests (both math mode suites)"

---

## 💡 Key Insights

1. **This is an evaluation issue** - Parser probably handles `...` syntax
2. **Focus on mixin_call.go** - Argument expansion likely happens during call evaluation
3. **List expansion** - `@x...` should turn list `1, 2, 3` into three separate arguments
4. **Complex test case** - Has nested expansion, re-expansion, multiple expansions
5. **Test incrementally** - Start simple, add complexity gradually

---

## 🔗 Related Code

From RUNTIME_ISSUES.md:
- Issue #5 (mixins-named-args) - FIXED - Related to mixin argument handling
- Issue #6 (mixins-closure) - FIXED - Mixin evaluation
- Issue #7 (mixins) - FIXED - Mixin recursion

These fixes show the mixin system is working, but `...` expansion is a specific missing feature.

---

## ⚠️ Special Notes

1. **Wave 2 issue** - Might conflict with other mixin work, coordinate timing
2. **Comprehensive test** - The test file has many edge cases, aim to fix most
3. **Both suites same test** - Same fix applies to math-parens and math-parens-division
4. **Variadic parameters** - Both collection (`@a...` in definition) and expansion (`@a...` in call)
5. **Don't break existing mixins** - Many mixin tests are passing, don't regress them

---

## 📖 Reference: JavaScript Implementation

In JavaScript less.js, look for:
- `args` array manipulation
- `isVariadic` or `isSpread` flags
- Argument expansion in `mixin-call.js`
- Parameter matching in `mixin-definition.js`

The pattern in JS might look like:
```javascript
if (arg.isVariadic) {
    // Expand arg.value into multiple arguments
    args = args.concat(arg.value.toArray());
}
```

Your Go implementation should follow the same pattern.
