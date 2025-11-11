# Performance Analysis: Why is Go 8-10x Slower?

## TL;DR

**The Go benchmark is measuring correctly** - compilation time is NOT included. The Go port is genuinely slower right now, primarily due to:

1. **3.4 MILLION allocations** per benchmark run (~47k per file)
2. Heavy use of reflection (462 occurrences across codebase)
3. String operations and conversions
4. Port is still under active development, not optimized yet

## Is Compilation Time Included?

**No.** Here's why:

```go
// From benchmark_test.go
b.ResetTimer()  // <-- This line resets the timer AFTER all setup
for i := 0; i < b.N; i++ {
    _, compileErr := compileLessForTest(factory, string(lessData), options)
    // ...
}
```

The `b.ResetTimer()` call happens AFTER:
- The Go test binary is compiled
- Test files are read from disk
- Options are configured
- Factory is created

The timer only measures the actual LESS compilation loop.

## The Real Culprit: Allocations

From your benchmark output:
```
BenchmarkLargeSuite-10    45    264059064 ns/op    299574207 B/op    3437404 allocs/op
```

Breaking this down:
- **3,437,404 allocations** for 73 files = **47,088 allocations per file**
- **299 MB** allocated = **4.1 MB per file**

Each allocation means:
- Heap allocation overhead (~16-24 bytes minimum)
- Garbage collection pressure
- Pointer dereferencing
- CPU cache misses
- Memory fragmentation

### Comparison

**JavaScript (V8)**:
- Highly optimized JIT compiler
- Inline caching
- Hidden classes for fast property access
- Escape analysis (stack allocates where possible)
- Years of optimization work

**Go (current port)**:
- Direct port without optimization
- Heavy reflection usage (462 instances)
- Lots of string operations
- Interface conversions
- Not profiled or optimized yet

## Finding the Hot Spots

Run profiling to see exactly where time is spent:

```bash
pnpm bench:profile
```

This will show:
1. Top 20 functions by CPU time
2. Top 20 functions by memory allocations
3. Allocation hotspots

Example output might show:
```
      flat  flat%   sum%        cum   cum%
     500ms 25.00% 25.00%     1500ms 75.00%  github.com/toakleaf/less.go/.../parser.Parse
     300ms 15.00% 40.00%      800ms 40.00%  github.com/toakleaf/less.go/.../visitor.Visit
     200ms 10.00% 50.00%      500ms 25.00%  reflect.ValueOf
     150ms  7.50% 57.50%      400ms 20.00%  fmt.Sprintf
```

## Common Performance Issues in Go

### 1. Reflection
**Cost**: 10-100x slower than direct access

Found 462 uses in codebase:
```bash
$ grep -r "reflect\." packages/less/src/less/less_go/*.go | wc -l
```

**Fix**: Replace reflection with type assertions or code generation

### 2. String Operations
**Cost**: Strings are immutable, so string concatenation allocates

**Bad**:
```go
result := ""
for i := 0; i < 1000; i++ {
    result += "text"  // 1000 allocations!
}
```

**Good**:
```go
var builder strings.Builder
for i := 0; i < 1000; i++ {
    builder.WriteString("text")  // 1 allocation
}
result := builder.String()
```

### 3. Interface Conversions
**Cost**: Heap allocation when concrete type escapes

**Bad**:
```go
func process(node any) {  // Forces heap allocation
    // ...
}
```

**Good**:
```go
func process(node *ConcreteType) {  // Can stay on stack
    // ...
}
```

### 4. Slice Growth
**Cost**: Reallocation and copying when capacity exceeded

**Bad**:
```go
var results []Node
for _, item := range items {
    results = append(results, item)  // May reallocate many times
}
```

**Good**:
```go
results := make([]Node, 0, len(items))  // Pre-allocate
for _, item := range items {
    results = append(results, item)  // No reallocation
}
```

## Optimization Roadmap

### Phase 1: Profile and Identify (Current)
- [x] Set up benchmarks
- [x] Add profiling tools
- [ ] Run profiling on representative workload
- [ ] Identify top 10 hot spots

### Phase 2: Low-Hanging Fruit
Focus on changes that give big wins with low risk:

1. **String Builder**: Replace string concatenation
2. **Pre-allocation**: Add capacity hints to slices/maps
3. **Reduce Reflection**: Replace with type switches where safe
4. **Avoid Sprintf**: Use strconv or direct string ops

### Phase 3: Structural Improvements
Bigger changes requiring more testing:

1. **Object Pooling**: Reuse frequently allocated objects
2. **Intern Strings**: Reuse common strings (colors, units, etc.)
3. **Optimize Visitors**: Reduce interface overhead
4. **Better Data Structures**: Arrays instead of slices where size is known

### Phase 4: Algorithmic
Re-think algorithms if needed:

1. **Lazy Evaluation**: Defer work until needed
2. **Memoization**: Cache expensive computations
3. **Parallel Processing**: Use goroutines for independent work

## Expected Improvement Potential

Based on typical Go optimization experiences:

- **Phase 2**: 2-3x speedup (8x → 3-4x slower)
- **Phase 3**: 1.5-2x additional (3-4x → 2x slower)
- **Phase 4**: 1.2-1.5x additional (2x → 1.5x slower)

**Realistic goal**: Match or beat JavaScript performance
- Go has advantages: compiled, no JIT warmup, predictable performance
- But requires optimization work

## Current Status: Expected and Acceptable

The 8-10x slower result is **expected** for an unoptimized port:

✅ **Good news**:
- Tests pass (80+ perfect CSS matches)
- Functionality is correct
- Foundation is solid

⚠️ **To improve**:
- Reduce allocations (47k → target <5k per file)
- Profile and optimize hot paths
- Replace reflection where possible
- Use proper string builders

## Next Steps

1. **Run profiling**:
   ```bash
   pnpm bench:profile
   ```

2. **Examine top CPU consumers**:
   ```bash
   go tool pprof -top profiles/cpu.prof
   ```

3. **Find allocation hot spots**:
   ```bash
   go tool pprof -alloc_objects profiles/mem.prof
   ```

4. **Visualize with flamegraph**:
   ```bash
   go tool pprof -http=:8080 profiles/cpu.prof
   ```

5. **Fix top 5 issues** and re-benchmark

## Conclusion

The benchmark is accurate - the Go port is slower due to:
1. Not being optimized yet (expected for a port)
2. Excessive allocations (47k per file is the main issue)
3. Heavy reflection usage
4. String operation overhead

This is **normal** for a first-pass port and there's **huge optimization potential** available. The profiling tools will show exactly where to focus optimization efforts for maximum impact.

---

**Remember**: Premature optimization is the root of all evil. You got the port working first (✅), now you can optimize based on data (📊).
