# Low-Hanging Fruit Performance Optimizations

**For:** Fresh LLM sessions to tackle independently
**Goal:** Achieve ~10% speedup with minimal risk
**Prerequisite:** Read `PERFORMANCE_BOTTLENECKS.md` for context

---

## Task 1: Pre-allocate Slices in findMatch

**Impact:** 5-7% speedup
**Risk:** Low
**Time:** 30 minutes
**File:** `packages/less/src/less/less_go/extend_visitor.go`

### The Problem

Lines 698-700 create slices without capacity hints:
```go
potentialMatches := make([]any, 0)
matches := make([]any, 0)
```

Every `append()` may trigger reallocation and copying. This function is called frequently during extend processing.

### The Fix

Add reasonable capacity hints based on typical usage:

```go
potentialMatches := make([]any, 0, 10)  // Most extend matches are small
matches := make([]any, 0, 5)            // Typical matches per extend
```

### How to Determine Capacity

1. Run integration tests with instrumentation:
```go
// Temporary debugging code
defer func() {
    fmt.Printf("potentialMatches cap: %d, len: %d\n", cap(potentialMatches), len(potentialMatches))
    fmt.Printf("matches cap: %d, len: %d\n", cap(matches), len(matches))
}()
```

2. Look at the output to see typical sizes
3. Choose a capacity that covers 80% of cases without being wasteful

### Testing

```bash
# Must pass all tests
LESS_GO_QUIET=1 pnpm -w test:go

# Benchmark before and after
pnpm -w bench:go:suite

# Look for reduced allocations
# Before: ~815k allocs/op
# After: Should be ~750-800k allocs/op
```

### Expected Results

- **Allocations:** Reduced by ~5-10k per op
- **Memory:** Reduced by ~5-10 MB per op
- **Speed:** ~5-7% faster

---

## Task 2: Replace Hot-Path fmt.Sprintf with strconv

**Impact:** 2-3% speedup
**Risk:** Very Low
**Time:** 1-2 hours
**Files:** Multiple

### The Problem

`fmt.Sprintf` uses reflection and is slow for simple conversions. It's called 1.6M times and allocates 29MB.

### The Targets

Profile shows these patterns are common:

```go
// Converting integers
fmt.Sprintf("%d", someInt)         // BAD
strconv.Itoa(someInt)              // GOOD

// Converting floats
fmt.Sprintf("%f", someFloat)       // BAD
strconv.FormatFloat(someFloat, 'f', -1, 64)  // GOOD

// Converting to string when already a string
fmt.Sprintf("%s", someString)      // BAD
someString                         // GOOD (if it's already a string)

// Simple value to string
fmt.Sprintf("%v", value)           // BAD
// Use type assertion first, then appropriate converter
```

### Find Hot Paths First

```bash
# Find Sprintf calls in non-test files
grep -r "fmt\.Sprintf" packages/less/src/less/less_go/*.go | grep -v _test.go

# Focus on these files (from profiling):
# - parser.go (38 uses)
# - declaration.go (7 uses)
# - dimension.go (7 uses)
# - quoted.go (8 uses)
```

### Example Refactoring

**Before:**
```go
// In parser.go:624
valueStr := fmt.Sprintf("%v", value)
```

**After:**
```go
var valueStr string
switch v := value.(type) {
case string:
    valueStr = v
case int:
    valueStr = strconv.Itoa(v)
case float64:
    valueStr = strconv.FormatFloat(v, 'f', -1, 64)
default:
    valueStr = fmt.Sprintf("%v", v)  // Fallback for complex types
}
```

### Testing Strategy

1. **Replace one file at a time**
2. **Run tests after each file:**
```bash
LESS_GO_QUIET=1 pnpm -w test:go
```
3. **Commit after each successful file:**
```bash
git add <file>
git commit -m "Optimize: Replace fmt.Sprintf in <file>"
```

### Expected Results

- **Allocations:** Reduced by ~100k per op
- **Memory:** Reduced by ~10-15 MB per op
- **Speed:** ~2-3% faster

---

## Task 3: Pre-allocate Slices in Hot Loops

**Impact:** 1-2% speedup
**Risk:** Low
**Time:** 2-3 hours
**Files:** Multiple

### The Problem

Many hot-path functions create slices in loops without capacity hints:

```go
results := []Node{}
for _, item := range items {
    results = append(results, processItem(item))
}
```

Each append may trigger reallocation.

### How to Find These

```bash
# Search for slice creation patterns
grep -n "make(\[\]" packages/less/src/less/less_go/*.go | grep -v _test.go

# Look for append in loops
grep -B3 -A3 "append(" packages/less/src/less/less_go/*.go | grep -v _test.go
```

### The Fix Pattern

**Before:**
```go
results := []Node{}
for _, item := range items {
    results = append(results, processItem(item))
}
```

**After:**
```go
results := make([]Node, 0, len(items))  // Pre-allocate for exact size
for _, item := range items {
    results = append(results, processItem(item))  // No reallocation
}
```

### Target Functions (From Profiling)

Focus on these high-frequency functions:
1. `(*Ruleset).Eval` - Creates slices for rules
2. `(*MixinDefinition).EvalParams` - Creates slices for parameters
3. `(*Selector).MixinElements` - Creates slices for elements
4. `(*JoinSelectorVisitor).VisitRuleset` - Creates slices for selectors

### Testing

```bash
# After each function optimization
LESS_GO_QUIET=1 pnpm -w test:go

# Profile to confirm improvement
pnpm bench:profile
go tool pprof -top profiles/mem.prof | head -30
```

### Expected Results

- **Allocations:** Reduced by ~50k per op
- **Memory:** Reduced by ~3-5 MB per op
- **Speed:** ~1-2% faster

---

## Task 4: Use strings.Builder for Concatenation

**Impact:** 1-2% speedup
**Risk:** Very Low
**Time:** 1-2 hours
**Files:** Multiple

### The Problem

String concatenation with `+` allocates a new string each time:

```go
result := ""
for _, item := range items {
    result += item  // Allocates every iteration!
}
```

### Find These Patterns

```bash
# Find string concatenation in loops
grep -B5 "+ \"" packages/less/src/less/less_go/*.go | grep -v _test.go
grep -B5 "\+= \"" packages/less/src/less/less_go/*.go | grep -v _test.go
```

### The Fix

**Before:**
```go
result := ""
for _, item := range items {
    result += item + ","
}
```

**After:**
```go
var builder strings.Builder
builder.Grow(len(items) * 10)  // Pre-allocate rough estimate
for _, item := range items {
    builder.WriteString(item)
    builder.WriteString(",")
}
result := builder.String()
```

### Target Files (Known Patterns)

- `to_css_visitor.go` - CSS output generation
- `selector.go` - Selector string building
- `element.go` - Element string building

### Testing

```bash
LESS_GO_QUIET=1 pnpm -w test:go
pnpm -w bench:go:suite
```

### Expected Results

- **Allocations:** Reduced by ~30k per op
- **Memory:** Reduced by ~2-3 MB per op
- **Speed:** ~1-2% faster

---

## Combined Impact

If all 4 tasks are completed successfully:

- **Total speedup:** ~10-15%
- **Allocation reduction:** ~185-240k per op (23-29%)
- **Memory reduction:** ~20-33 MB per op (42-70%)

**Before all optimizations:**
- 147ms per op, 47MB, 815k allocs

**After all optimizations (estimated):**
- 125-135ms per op, 27-35MB, 575-630k allocs

---

## General Guidelines

### Before Starting

1. **Pull latest master:** `git fetch origin master && git merge origin/master`
2. **Run baseline benchmark:** `pnpm -w bench:go:suite > baseline.txt`
3. **Verify tests pass:** `LESS_GO_QUIET=1 pnpm -w test:go`

### While Working

1. **One optimization at a time** - Don't mix multiple changes
2. **Test after each change** - Catch regressions early
3. **Commit frequently** - Small, focused commits
4. **Measure everything** - Run benchmarks to confirm improvement

### After Completing

1. **Run full test suite:** `pnpm -w test:go`
2. **Check for regressions:** Verify 80 perfect CSS matches remain
3. **Benchmark final result:** `pnpm -w bench:go:suite`
4. **Document changes:** Update this file with actual results

---

## What NOT To Do

❌ **Don't optimize without profiling first**
❌ **Don't change code you don't understand**
❌ **Don't skip testing**
❌ **Don't sacrifice correctness for speed**
❌ **Don't make large architectural changes** (that's Phase 3)

---

## If You Get Stuck

1. **Check profiling data:** `pnpm bench:profile`
2. **Read test output:** Tests often show what broke
3. **Use git diff:** See exactly what changed
4. **Revert and retry:** `git checkout <file>` to start over
5. **Ask for help:** Document what you tried and what failed

---

## Success Criteria

An optimization is successful if:

1. ✅ All integration tests still pass (80 perfect matches)
2. ✅ Benchmark shows measurable improvement (>1% faster or <5% allocations)
3. ✅ Code remains readable and maintainable
4. ✅ No regressions in any test category

If any criterion fails, revert and try a different approach.
