# Task: Fix Mixin Args - Forward References in Rulesets

**Status**: Available
**Priority**: HIGH
**Estimated Time**: 3-4 hours
**Complexity**: Medium-High
**Tests Affected**: 2 instances (both math/strict and math/parens-division suites use same file)

## Overview

The `mixins-args` test fails with "No matching definition was found for `.m3()`". The issue is that mixins defined later in the same ruleset are not accessible to code earlier in that ruleset - a **forward reference** or **scoping** problem.

This is NOT a regression - this test has been failing. It involves complex mixin argument handling, variadic arguments (`@args...`), and mixin scoping within rulesets.

## Failing Tests

### 1. mixins-args (math/strict)
- **Status**: ❌ Compilation failed
- **Error**: `Syntax: No matching definition was found for .m3()`
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "math-strict/mixins-args"
  ```

### 2. mixins-args (math/parens-division)
- **Status**: ❌ Compilation failed
- **Error**: `Syntax: No matching definition was found for .m3()`
- **Test Command**:
  ```bash
  pnpm -w test:go:filter -- "math-parens-division/mixins-args"
  ```

**Note**: Both tests use the exact same LESS file: `packages/test-data/less/math/strict/mixins-args.less`

## Current Behavior

```less
// From lines 214-230 of mixins-args.less
mixins-args-expand-op- {
  @x: 1, 2, 3;
  @y: 4  5  6;

  &1 {.m3(@x...)}  // ❌ Line 218: ERROR - .m3 is undefined
  &2 {.m3(@y...)}  // ❌ Error
  &3 {.wr(a, b, c)}  // This might work if .wr is found

  // ...more uses of .m3

  .m3(@a, @b, @c) {  // Line 228: DEFINITION is here (10 lines later)
    m3: @a, @b, @c;
  }

  .m4(@a, @b, @c, @d) {  // Line 232
    m4: @a, @b, @c, @d;
  }

  .wr(@a...) {  // Line 236
    &a {.m3(@a...)}
    &b {.m4(0, @a...)}
    &c {.m4(@a..., 4)}
  }
}
```

**Problem**: On line 218, `&1 {.m3(@x...)}` tries to call `.m3()`, but `.m3()` isn't defined until line 228 (inside the same parent ruleset). The mixin lookup fails because the mixin hasn't been processed yet.

## Expected Behavior

In LESS, mixins should be available throughout their containing ruleset, regardless of definition order. This is a **two-pass** or **hoisting** scenario:

**Pass 1**: Collect all mixin definitions in the ruleset
**Pass 2**: Evaluate mixin calls, now that all definitions are known

**Expected Output**: The test should compile and output should show:
```css
mixins-args-expand-op-1 {
  m3: 1, 2, 3;
}
mixins-args-expand-op-2 {
  m3: 4, 5, 6;
}
/* ...etc */
```

## Investigation Starting Points

### Files to Examine

1. **`ruleset.go`** (PRIMARY) - Lines 200-500
   - Ruleset evaluation order
   - How rules within a ruleset are processed
   - Check if there's a "collect mixins first" pass

2. **`mixin_definition.go`** - Lines 100-250
   - How mixin definitions are registered
   - When they become available in frames

3. **`mixin_call.go`** - Lines 50-200
   - Mixin lookup logic
   - Where it searches for mixin definitions
   - Error handling when mixin not found

4. **`contexts.go`** - Lines 100-200
   - Frame management
   - How frames are pushed/popped during evaluation

### Debug Commands

```bash
# Run with trace to see evaluation order
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "math-strict/mixins-args" 2>&1 | grep -E "(\.m3|\.m4|\.wr)" | head -50

# Just the error
pnpm -w test:go:filter -- "math-strict/mixins-args" 2>&1 | grep -A5 "No matching definition"

# Simpler test case to debug
cat > /tmp/test-forward-ref.less << 'EOF'
.parent {
  &-child { .mixin(); }
  .mixin() { color: red; }
}
EOF
cd packages/less/src/less/less_go && go run cmd/lessc/lessc.go /tmp/test-forward-ref.less
```

### Compare with JavaScript

Check how less.js handles this:
```bash
# Run with JavaScript less
cd packages/less
npx lessc test-data/less/math/strict/mixins-args.less /tmp/js-output.css
cat /tmp/js-output.css | head -50
```

## Root Cause Hypothesis

**Most Likely**: The Go implementation evaluates ruleset contents in a single pass, sequentially. When it encounters `.m3(@x...)` on line 218, it looks for `.m3` in the current frames, doesn't find it (because it hasn't processed line 228 yet), and throws an error.

**Needed Fix**: Implement two-pass evaluation for rulesets:

### Approach 1: Two-Pass Evaluation (RECOMMENDED)

```go
// In ruleset.go Eval() method

// Pass 1: Collect all mixin definitions
for _, rule := range r.Rules {
    if mixinDef, ok := rule.(*MixinDefinition); ok {
        // Register mixin in frame BEFORE evaluating anything
        context.Frames.AddMixin(mixinDef)
    }
}

// Pass 2: Evaluate all rules
for _, rule := range r.Rules {
    evaluated, err := rule.Eval(context)
    // ... rest of evaluation
}
```

### Approach 2: Lazy Mixin Resolution

Defer mixin lookup until after all rules in the current ruleset are known.

### Approach 3: Pre-scan Rulesets

Before evaluating any ruleset, scan it for mixin definitions and register them.

## Success Criteria

- [ ] `mixins-args` test (strict mode) compiles successfully
- [ ] `mixins-args` test (parens-division mode) compiles successfully
- [ ] Mixins can be called before they're defined (within same ruleset)
- [ ] Other mixin tests still pass (mixins, mixins-closure, mixins-pattern, etc.)
- [ ] All unit tests pass: `pnpm -w test:go:unit`
- [ ] FULL integration test suite shows no regressions: `pnpm -w test:go`

## Validation Checklist

Before creating PR:

```bash
# 1. Verify both specific tests compile and pass
pnpm -w test:go:filter -- "math-strict/mixins-args"
# Expected: ✅ Compilation succeeds, output matches expected

pnpm -w test:go:filter -- "math-parens-division/mixins-args"
# Expected: ✅ Compilation succeeds, output matches expected

# 2. Verify other mixin tests still work (NO REGRESSIONS)
pnpm -w test:go 2>&1 | grep "mixin" | grep -E "(✅|❌)"
# Expected: All currently passing mixin tests still pass

# 3. Run ALL unit tests (catch any regressions) - REQUIRED
pnpm -w test:go:unit
# Expected: ✅ All unit tests pass (no failures)

# 4. Run FULL integration test suite - REQUIRED
pnpm -w test:go
# Expected:
# - ✅ Compilation failure count drops by 2 (or 1 if they're counted as one test)
# - ✅ No new compilation failures
# - ✅ Perfect matches remain at 14+ (hopefully 15+ if regressions are fixed)
```

**If any test fails that was passing before**: STOP and fix the regression before proceeding.

## Additional Context

### Currently Passing Mixin Tests (Don't Break These!)

- ✅ `mixin-noparens` - Perfect match
- ✅ `mixins` - Perfect match
- ✅ `mixins-closure` - Perfect match
- ✅ `mixins-interpolated` - Perfect match
- ✅ `mixins-pattern` - Perfect match

### Related Failed Tests (Different Issues)

- ⚠️ `mixins-guards` - Output differs
- ⚠️ `mixins-guards-default-func` - Output differs
- ⚠️ `mixins-important` - Output differs
- ⚠️ `mixins-named-args` - Output differs
- ⚠️ `mixins-nested` - Output differs

### Test File Details

- **Location**: `packages/test-data/less/math/strict/mixins-args.less`
- **Size**: 260 lines
- **Tests Many Features**:
  - Mixin arguments with defaults
  - Variadic arguments (`@args...`)
  - Argument spreading (`.m3(@x...)`)
  - Named arguments
  - Guards and pattern matching
  - Division vs. literal slash in arguments
  - Forward references (the failing part)

The forward reference issue is in the last section (lines 214-260).

## Notes

This is a significant scoping issue that affects mixin availability. The fix needs to:
1. Allow forward references within the same ruleset
2. Not break existing mixin behavior
3. Handle nested rulesets correctly (mixins inside mixins)

Consider looking at how variables are handled - they might have similar forward-reference handling that could be used as a pattern.
