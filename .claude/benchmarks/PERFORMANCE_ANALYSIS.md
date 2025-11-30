# Performance Analysis: REGEX OPTIMIZATION COMPLETE ✅

## TL;DR - STATUS

**REGEX COMPILATION ISSUE: FIXED** (as of 2025-11-30)

All regex compilations have been moved to package-level variables and are now compiled once at startup.

### What Was Done

All 99 dynamic regex compilations have been moved to package-level `var` blocks:

| File | Pre-compiled Regexes |
|------|---------------------|
| `parser_regexes.go` | 50+ patterns (parser hot paths) |
| `quoted.go` | 3 patterns (variable/property interpolation) |
| `selector.go` | 1 pattern (mixin elements) |
| `import.go` | 1 pattern (CSS detection) |

### Only Exception

`string.go:222` - Dynamic regex in `Replace()` function. **Cannot be pre-compiled** because the pattern is user-provided via LESS `replace()` function arguments.

### Verification

To verify no dynamic compilations remain in hot paths:
```bash
grep -n "regexp\.\(Must\)\?Compile" packages/less/src/less/less_go/*.go | grep -v "_test.go" | grep -v "trace_examples"
```

Should only show:
- Package-level `var` declarations (parser_regexes.go, quoted.go, selector.go, import.go)
- `string.go:222` (user-provided pattern, cannot pre-compile)

## Historical Profiling Evidence

The following profiling data motivated the regex optimization:

```
TOP MEMORY ALLOCATIONS (BEFORE FIX):
  1.40GB (30.80%) - regexp/syntax.(*compiler).inst
  1.10GB (24.21%) - regexp/syntax.(*parser).newRegexp
  3.71GB (81.73%) - regexp.compile (CUMULATIVE)

ALLOCATION COUNTS (BEFORE FIX):
  10,542,240 allocs (19.69%) - regexp/syntax.(*parser).newRegexp
   5,315,351 allocs (9.93%) - regexp/syntax.(*compiler).inst
  40,664,606 allocs (75.96%) - regexp.compile (CUMULATIVE)
```

**Original Problem:**
- **99 total occurrences** of dynamic regex compilation
- **60 in parser.go alone**
- Every regex compiled from scratch on every parse call

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

### Phase 1: Profile and Identify (COMPLETE ✅)
- [x] Set up benchmarks
- [x] Add profiling tools
- [x] Run profiling on representative workload
- [x] Identify top 10 hot spots
- [x] **Regex optimization complete** - All 99 regexes pre-compiled at package level

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

---

## Plugin System IPC Performance (2025-11-29)

### TL;DR

**JSON IPC is 70% faster than Shared Memory (SHM) for plugin function calls.**

| Mode | Bootstrap4 Time | Allocations |
|------|-----------------|-------------|
| **JSON** | ~840ms | ~2.93M |
| **SHM** | ~1,420ms | ~2.92M |

**Default is JSON mode** - this is the optimal choice for all current use cases.

### Why SHM is Slower

The SHM protocol was designed for zero-copy data transfer, which theoretically should be faster. However, for Bootstrap4-style compilations:

1. **High per-call overhead**: SHM requires creating a memory-mapped buffer, serializing to binary FlatAST format, syncing memory, etc.
2. **Many small calls**: Bootstrap4 makes thousands of small function calls (map-get, color-yiq, breakpoint-next, etc.)
3. **Serialization not the bottleneck**: For small payloads, JSON serialization is fast enough that simpler pipe-based IPC wins

### When SHM Would Help (Theoretical)

SHM could outperform JSON for:
- Plugins with large data payloads (e.g., processing entire AST trees)
- Fewer, larger function calls
- Currently **no real-world plugins fit this profile**

### Per-Plugin IPC Configuration

Plugins can now specify their preferred IPC mode:

```javascript
// In plugin JavaScript
module.exports = {
  install: function(less, pluginManager) {
    pluginManager.addFunctions({ ... });
  },
  ipcMode: "json"  // or "shm" for shared memory
};
```

The Go side reads this and uses it when creating function definitions.

**Configuration Priority:**
1. Per-plugin config (from JS plugin response) - highest priority
2. Environment variable (`LESS_JS_IPC_MODE=json` or `LESS_JS_IPC_MODE=shm`)
3. Default: JSON

### Environment Variables

| Variable | Values | Description |
|----------|--------|-------------|
| `LESS_JS_IPC_MODE` | `json`, `shm` | Override IPC mode for all plugins |
| `LESS_SHM_PROTOCOL` | `1` | Enable SHM protocol initialization |

### Key Files

- `runtime/js_function.go` - IPC mode configuration, `ParseIPCMode()`, `GetOrCreateJSFunctionDefinition()`
- `runtime/plugin_loader.go` - `Plugin.IPCMode` field, reads `ipcMode` from plugin response
- `runtime/plugin_scope.go` - Propagates IPC mode to function definitions
- `lazy_plugin_bridge.go` - SHM protocol initialization (disabled by default)

### Benchmark Commands

```bash
# JSON mode (default)
go test -bench="BenchmarkBootstrap4$" -benchmem -count=3 ./packages/less/src/less/less_go/...

# SHM mode
LESS_JS_IPC_MODE=shm LESS_SHM_PROTOCOL=1 go test -bench="BenchmarkBootstrap4$" -benchmem -count=3 ./packages/less/src/less/less_go/...
```

### SHM Protocol Status

The SHM binary protocol is **not fully implemented**:
- Go side can write binary FlatAST format to shared memory
- Node.js side currently returns JSON instead of binary format
- One test fails: `TestJSFunctionDefinition_SharedMemoryCall`
- This is a known limitation, not a priority to fix since JSON is faster anyway
