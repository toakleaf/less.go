# Performance Analysis: less.go Port

> **Last Updated**: 2025-11-27
> **Status**: Current profiling data with actionable optimization opportunities

## TL;DR - Current State

The Go port is currently **8-10x slower** than JavaScript. This is expected for a correctness-first port without optimization.

**Key Metrics (as of 2025-11-27):**
- **~862,000 allocations** per benchmark run (73 files)
- **~11,800 allocations per file** (down from 47k reported previously)
- **44 MB total** / **604 KB per file**
- **~2.15ms per file** compilation time

**Previous regex compilation issue: FIXED!** Regexes are now pre-compiled in `parser_regexes.go`.

---

## Top 5 Optimization Opportunities (Ranked by Impact)

### 1. Reflection in Visitor Pattern (~22% of allocations)

**Evidence from profiling:**
```
reflect.Value.call:              61.50MB cumulative (10.47%)
reflect.(*rtype).Method:         56.00MB cumulative (9.54%)
reflect.New:                     39.50MB flat (6.73%)
reflect.(*structType).FieldByNameFunc: 28.50MB flat (4.85%)
reflect.methodReceiver:          590k allocations (5.68%)
```

**Root cause:** The visitor pattern uses `reflect.Value.MethodByName` and `reflect.Value.Call` to dispatch visit methods dynamically.

**Location:** `visitor.go:118-225`

**Current code (simplified):**
```go
// Slow path: Use reflection-based dispatch
visitMethod, visitMethodExists := v.methodLookup[fnName]
if visitMethodExists && visitMethod.IsValid() {
    results := visitMethod.Call([]reflect.Value{
        reflect.ValueOf(n),
        reflect.ValueOf(args),
    })
}
```

**Fix options:**
1. **Implement DirectDispatchVisitor interface** for all visitors (already defined in code)
2. **Use type switches** instead of reflection for node dispatch
3. **Code generation** to create typed dispatch functions

**Expected improvement:** 15-25% reduction in allocations

---

### 2. Node Pool Inefficiency (~12% of memory)

**Evidence from profiling:**
```
init.func1 (NewNode pool): 70.51MB flat (12.01%)
NewNode (via sync.Pool):   71.01MB cumulative (12.09%)
```

**Root cause:** The `sync.Pool` for `Node` objects is allocating new nodes rather than effectively reusing pooled ones.

**Location:** `node.go:31-52`

**Current code:**
```go
var nodePool = sync.Pool{
    New: func() interface{} {
        return &Node{}
    },
}

func NewNode() *Node {
    n := nodePool.Get().(*Node)
    // Reset all fields...
    return n
}
```

**Issues:**
1. Pool objects may not be returned properly
2. Resetting fields on every Get() creates overhead
3. Other node types (Ruleset, Selector, etc.) don't use pooling

**Fix options:**
1. Audit all `NewNode()` usages to ensure `ReleaseNode()` is called
2. Implement struct pooling for frequently allocated types (Ruleset, Selector, Element)
3. Consider arena-based allocation for parse trees

**Expected improvement:** 10-15% reduction in memory

---

### 3. fmt.Sprintf Overhead (~6% of allocations)

**Evidence from profiling:**
```
fmt.Sprintf: 595,286 allocations (5.73%)
```

**Root cause:** 274 usages of `fmt.Sprintf` across 66 files, many in hot paths.

**Location:** Distributed across codebase. High-frequency files:
- `parser.go`: 38 usages
- `ruleset.go`: 13 usages
- `namespace_value.go`: 11 usages
- `less_error.go`: 10 usages

**Fix options:**
1. Replace with string concatenation (`"prefix" + value + "suffix"`)
2. Use `strconv.Itoa/FormatFloat` for number formatting
3. Use `strings.Builder` for complex string building
4. Pre-allocate error message templates

**Example optimization:**
```go
// Before: allocates
return fmt.Sprintf("VariableCall %s (rules count: %d)", vc.Variable, len(vc.Rules))

// After: no allocation for simple cases
return "VariableCall " + vc.Variable + " (rules count: " + strconv.Itoa(len(vc.Rules)) + ")"
```

**Expected improvement:** 3-5% reduction in allocations

---

### 4. Mixin Evaluation Overhead (~28% cumulative)

**Evidence from profiling:**
```
(*MixinCall).Eval:         163.60MB cumulative (27.86%)
(*MixinDefinition).EvalCall: 99.05MB cumulative (16.87%)
(*MixinDefinition).EvalParams: 35.03MB cumulative (5.97%)
```

**Root cause:** Mixin evaluation involves:
1. Frame copying on each call
2. Parameter evaluation and argument matching
3. Context cloning

**Location:** `mixin_call.go`, `mixin_definition.go`

**Fix options:**
1. **Cache mixin lookups** - Don't re-search for mixins on repeated calls
2. **Reduce frame copying** - Use copy-on-write or structural sharing
3. **Pre-allocate argument arrays** with known capacity
4. **Inline simple mixins** at parse time

**Expected improvement:** 10-15% reduction in mixin-heavy files

---

### 5. Ruleset Evaluation (~39% cumulative)

**Evidence from profiling:**
```
(*Ruleset).Eval:    231.15MB cumulative (39.36%)
(*Ruleset).Accept:  126.51MB cumulative (21.54%)
NewRuleset:         31.51MB flat (5.37%)
```

**Root cause:** Ruleset is the most frequently allocated and evaluated node type.

**Location:** `ruleset.go`

**Fix options:**
1. **Pre-allocate slices** with proper capacity hints:
   ```go
   // Before
   var results []any
   for _, rule := range rs.Rules {
       results = append(results, rule.Eval())
   }

   // After
   results := make([]any, 0, len(rs.Rules))
   ```

2. **Lazy initialization** for lookups map (already partially done)
3. **Reduce copying** of selectors and rules arrays
4. **Pool Ruleset objects** like Node objects

**Expected improvement:** 8-12% reduction in ruleset-heavy files

---

## Secondary Optimization Opportunities

### 6. Regex FindStringSubmatch Allocations
```
regexp.(*Regexp).FindStringSubmatch: 114k allocations (1.42%)
```
Each regex match creates a new slice. Consider using `FindStringSubmatchIndex` and extracting substrings manually.

### 7. Selector and Element Creation
```
NewSelector: 24.50MB cumulative (4.17%)
NewElement: 13MB cumulative (2.21%)
```
High-frequency allocations. Good candidates for object pooling.

### 8. Context/Eval Map Operations
```
(*Eval).CopyEvalToMap: 26.02MB flat (4.43%)
(*Eval).GetFrames:     5.01MB flat (0.85%)
```
Context copying happens frequently. Consider structural sharing or lazy copying.

---

## Profiling Commands Reference

```bash
# Run benchmarks with memory stats
pnpm bench:go:suite

# Generate CPU and memory profiles
cd packages/less/src/less/less_go
go test -bench=BenchmarkLargeSuite -benchtime=10x -cpuprofile=cpu.prof -memprofile=mem.prof

# Analyze profiles
go tool pprof -top cpu.prof                    # Top CPU consumers
go tool pprof -top -alloc_objects mem.prof     # Top allocators by count
go tool pprof -top -alloc_space mem.prof       # Top allocators by memory
go tool pprof -top -cum mem.prof               # Cumulative allocation paths

# Interactive analysis
go tool pprof -http=:8080 cpu.prof             # Web UI with flamegraph
```

---

## Expected Optimization Roadmap

### Phase 1: Low-Hanging Fruit (Est. 2-3x speedup)
1. Replace `fmt.Sprintf` in hot paths
2. Pre-allocate slices with capacity hints
3. Implement DirectDispatchVisitor for main visitors

### Phase 2: Structural Improvements (Est. 1.5-2x additional)
1. Object pooling for Ruleset, Selector, Element
2. Reduce reflection usage in visitor pattern
3. Optimize mixin lookup caching

### Phase 3: Algorithmic (Est. 1.2-1.5x additional)
1. Lazy evaluation where possible
2. Structural sharing for context/frames
3. Parallel processing for independent rulesets

**Realistic Goal:** Match or beat JavaScript performance with focused optimization work.

---

## Appendix: Full Profiling Output

### Memory Allocation by Object Count (Top 20)
```
      flat  flat%   sum%        cum   cum%
    961197  9.25%  9.25%    1470224 14.15%  reflect.(*rtype).Method
    770028  7.41% 16.67%     770028  7.41%  init.func1 (node pool)
    609863  5.87% 22.54%     609863  5.87%  reflect.(*structType).FieldByNameFunc
    595286  5.73% 28.27%     595286  5.73%  fmt.Sprintf
    589824  5.68% 33.95%     589824  5.68%  reflect.methodReceiver
    509027  4.90% 38.85%     509027  4.90%  reflect.New
    240298  2.31% 41.16%     294911  2.84%  NewRuleset
    229376  2.21% 43.37%     229376  2.21%  reflect.Value.assignTo
    212084  2.04% 45.41%     979488  9.43%  (*MixinDefinition).EvalCall
    193340  1.86% 47.27%     193340  1.86%  (*ProcessExtendsVisitor).findMatch
```

### Memory Allocation by Size (Top 20)
```
      flat  flat%   sum%        cum   cum%
   70.51MB 12.01% 12.01%    70.51MB 12.01%  init.func1 (node pool)
   39.50MB  6.73% 18.73%    39.50MB  6.73%  reflect.New
   31.51MB  5.37% 24.10%    36.51MB  6.22%  NewRuleset
   28.50MB  4.85% 28.95%    28.50MB  4.85%  reflect.(*structType).FieldByNameFunc
   26.02MB  4.43% 33.38%    26.02MB  4.43%  (*Eval).CopyEvalToMap
   16.51MB  2.81% 36.20%    35.03MB  5.97%  (*MixinDefinition).EvalParams
   16.50MB  2.81% 39.01%       56MB  9.54%  reflect.(*rtype).Method
   16.01MB  2.73% 41.73%    99.05MB 16.87%  (*MixinDefinition).EvalCall
   13.55MB  2.31% 44.04%    13.55MB  2.31%  (*Registry).Add
   13.50MB  2.30% 46.34%    24.50MB  4.17%  NewSelector
```

### CPU Profile (Top Functions)
```
Parser.Parse:          54% cumulative (parsing phase)
Ruleset.Accept:        37% cumulative (visitor traversal)
Ruleset.Eval:          28% cumulative (evaluation)
MixinCall.Eval:        11% cumulative
reflect.Value.Call:    15% cumulative (reflection overhead)
runtime.mallocgc:      15% cumulative (GC pressure)
runtime.gcDrain:       10% cumulative (GC work)
```

---

## Conclusion

The Go port is slower primarily due to:
1. **Heavy reflection usage** in the visitor pattern (~22% overhead)
2. **Ineffective object pooling** for Node types (~12% overhead)
3. **Excessive string allocations** via fmt.Sprintf (~6% overhead)
4. **Non-optimized mixin evaluation** (28% cumulative)

These are addressable issues with focused optimization work. The foundation is solid (tests pass, correctness achieved), making it safe to optimize based on profiling data.

**Remember:** Premature optimization is the root of all evil. You got the port working first (✅), now you can optimize based on data (📊).
