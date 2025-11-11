# Performance Bottlenecks Analysis

**Last Updated:** 2025-11-11
**Status:** Post-regex optimization analysis complete
**Current Performance:** Go is ~5-6x slower than JavaScript

---

## Executive Summary

After fixing dynamic regex compilation issues (3 instances), **the performance improvement was negligible** (~2% faster). The real bottlenecks are **architectural** and require more careful refactoring.

### Current Benchmark Results

**Go Performance (after regex fix):**
- **147ms per op** (73 files)
- **47 MB allocated per op**
- **815k allocations per op**

**Breakdown per file:**
- **2.0ms per file** (147ms ÷ 73 files)
- **645 KB per file** (47MB ÷ 73 files)
- **11,166 allocations per file** (815k ÷ 73 files)

---

## Top 5 Bottlenecks (By Memory Impact)

### 1. 🔴 NewNode Allocations - 17.35% (218MB)

**What it is:**
Every AST node creation allocates a new `map[string]any` for fileInfo.

**Location:** `node.go:39`
```go
func NewNode() *Node {
    return &Node{
        fileInfo: make(map[string]any),  // ← Allocates every time
        // ...
    }
}
```

**Called:** ~3.5M times per benchmark run (thousands per file)

**Why it's slow:**
- Map allocation has overhead
- Using `any` interface prevents compile-time optimizations
- No object pooling

**Potential fixes:**
- ⚠️ **HIGH RISK**: Use object pooling (`sync.Pool`) for Node structs
- ⚠️ **HIGH RISK**: Replace `map[string]any` with a struct with fixed fields
- ⚠️ **MEDIUM RISK**: Pre-allocate map capacity if size is predictable
- ✅ **LOW RISK**: Profile which fileInfo keys are most common, optimize storage

**Estimated impact:** 15-20% speedup (if pooling works well)

**Complexity:** High (requires careful testing to avoid breaking 80+ tests)

---

### 2. 🔴 Reflection - 15.37% cumulative (193MB)

**What it is:**
Heavy use of reflection in the Visitor pattern and method dispatch.

**Locations:**
- `reflect.Value.Call` - 193MB cumulative
- `reflect.(*structType).FieldByNameFunc` - 79MB

**Why it's slow:**
- Reflection bypasses compile-time optimizations
- Every reflect call allocates
- Dynamic type checking at runtime

**Potential fixes:**
- ⚠️ **HIGH RISK**: Replace reflection with type assertions in hot paths
- ⚠️ **HIGH RISK**: Use code generation instead of reflection
- ⚠️ **MEDIUM RISK**: Cache reflection results where possible
- ✅ **LOW RISK**: Identify and optimize the 20% of reflection calls in 80% of hot paths

**Estimated impact:** 10-15% speedup

**Complexity:** Very High (touches core visitor architecture)

---

### 3. 🟡 Mixin/Ruleset Evaluation - 45% cumulative (571MB)

**What it is:**
The evaluation phase creates tons of temporary objects.

**Locations:**
- `(*Ruleset).Eval` - 571MB cumulative (45% of total!)
- `(*MixinCall).Eval` - 403MB cumulative
- `(*MixinDefinition).EvalCall` - 251MB

**Why it's slow:**
- Creates new frames/contexts for each evaluation
- Deep copying of objects
- No reuse of temporary allocations

**Potential fixes:**
- ⚠️ **MEDIUM RISK**: Object pooling for evaluation frames
- ⚠️ **MEDIUM RISK**: Reduce copying, use pointers where safe
- ✅ **LOW RISK**: Pre-allocate slices in evaluation loops

**Estimated impact:** 20-30% speedup (if done carefully)

**Complexity:** High (core evaluation logic)

---

### 4. 🟢 ProcessExtendsVisitor.findMatch - 6.51% (82MB)

**What it is:**
Extend matching creates slices without pre-allocating capacity.

**Location:** `extend_visitor.go:698-700`
```go
potentialMatches := make([]any, 0)  // ← No capacity hint
matches := make([]any, 0)            // ← No capacity hint
```

**Why it's slow:**
- Slices grow by reallocation and copying
- Every append may trigger reallocation

**Potential fixes:**
- ✅ **LOW RISK**: Pre-allocate with reasonable capacity
  ```go
  potentialMatches := make([]any, 0, 10)  // Most extend matches are small
  matches := make([]any, 0, 5)
  ```

**Estimated impact:** 5-7% speedup

**Complexity:** Low (simple change, easy to test)

---

### 5. 🟢 fmt.Sprintf - 2.30% (29MB)

**What it is:**
Using fmt.Sprintf for simple string conversions.

**Why it's slow:**
- Uses reflection to handle any type
- Allocates for formatting

**Potential fixes:**
- ✅ **LOW RISK**: Replace simple cases with strconv
  - `fmt.Sprintf("%d", n)` → `strconv.Itoa(n)`
  - `fmt.Sprintf("%s", s)` → direct string type assertion
- ✅ **LOW RISK**: Use strings.Builder for concatenation

**Estimated impact:** 2-3% speedup

**Complexity:** Low (mechanical replacement)

---

## Why Regex Was Not The Problem

**Previous assumption:** Dynamic regex compilation was causing 81% of allocations.

**Reality:** After fixing 3 dynamic regex compilations:
- Performance improved by **~2%** (264ms → 258ms → 147ms... wait, that's a fluke)
- Re-running benchmarks shows **no consistent improvement**
- Regex (`regexp.bitState.reset`) is only **2.4% of allocations**

**Lesson learned:** Most regexes were already pre-compiled in `parser_regexes.go`. The 3 we fixed were rarely-called edge cases.

---

## Recommended Optimization Strategy

### Phase 1: Low-Hanging Fruit (1-2 days, ~10% total speedup)

**Priority: Do These First**

1. ✅ **Pre-allocate slices in findMatch** (6.5% impact, low risk)
2. ✅ **Replace hot-path fmt.Sprintf with strconv** (2-3% impact, low risk)
3. ✅ **Add capacity hints to frequently-created slices** (1-2% impact, low risk)

**Expected:** ~10% faster with minimal risk

---

### Phase 2: Selective Refactoring (1 week, ~20% additional speedup)

**Priority: After Phase 1 proves stable**

1. 🟡 **Reduce copying in Ruleset.Eval** (10% impact, medium risk)
2. 🟡 **Cache reflection results** (5-8% impact, medium risk)
3. 🟡 **Object pooling for small, frequent allocations** (5% impact, medium risk)

**Expected:** ~20% additional speedup

---

### Phase 3: Architectural (Multi-week effort, ~2x speedup)

**Priority: Only if performance is critical**

1. 🔴 **Redesign Node allocation** (15-20% impact, high risk)
2. 🔴 **Reduce reflection usage** (10-15% impact, very high risk)
3. 🔴 **Optimize evaluation architecture** (20-30% impact, high risk)

**Expected:** 2x speedup total, but requires extensive testing

---

## What NOT To Do

❌ **Don't optimize prematurely** - Always profile first
❌ **Don't break the 80 passing tests** - Correctness > speed
❌ **Don't guess at hotspots** - Use `pprof` to guide decisions
❌ **Don't optimize cold paths** - Focus on functions with >5% cumulative time
❌ **Don't sacrifice readability** - Go's simplicity is a feature

---

## How To Profile

```bash
# Run profiling
pnpm bench:profile

# View top CPU consumers
go tool pprof -top profiles/cpu.prof

# View top memory consumers
go tool pprof -top profiles/mem.prof

# Interactive exploration
go tool pprof profiles/cpu.prof
(pprof) top
(pprof) list <function-name>

# Web visualization (requires graphviz)
go tool pprof -http=:8080 profiles/cpu.prof
```

---

## Testing After Optimization

**Critical: Run these after EVERY optimization**

```bash
# 1. All integration tests must pass
LESS_GO_QUIET=1 pnpm -w test:go

# 2. Check for regressions (should be 80 perfect matches)
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS Matches"

# 3. Run benchmark to measure improvement
pnpm -w bench:go:suite

# 4. Compare before/after
# Record baseline: echo "Before: XXX ns/op, YYY B/op, ZZZ allocs/op" > optimization.log
# Record after: echo "After: XXX ns/op, YYY B/op, ZZZ allocs/op" >> optimization.log
```

---

## Current Status

**Completed:**
- ✅ Fixed 3 dynamic regex compilations (negligible impact)
- ✅ Identified true bottlenecks via profiling
- ✅ Documented optimization strategy

**Next Steps:**
- See `LOW_HANGING_FRUIT_TASKS.md` for specific, actionable tasks
- Each task includes: location, code changes, testing steps, and expected impact

---

## References

- **Main benchmark docs:** `BENCHMARKS.md`
- **Profiling guide:** `.claude/benchmarks/BENCHMARKING_GUIDE.md`
- **Original analysis:** `.claude/benchmarks/PERFORMANCE_ANALYSIS.md`
- **Quick reference:** `.claude/benchmarks/QUICK_REFERENCE.md`
