# Performance Benchmarks: Less.go vs Less.js

This document provides a quick guide to running performance benchmarks comparing the Go port with the original JavaScript implementation.

## Quick Start

### Run Comparison (Recommended)

```bash
pnpm bench:compare
```

This runs both JavaScript and Go benchmarks and displays a clear side-by-side comparison with:
- ✅ Per-file and total compilation times
- ✅ Performance ratio (which is faster and by how much)
- ✅ Memory usage and allocation statistics (Go)
- ✅ Actionable optimization recommendations

**Example output:**
```
╔══════════════════════════════════════════════════════════════════════════════╗
║              LESS.JS vs LESS.GO PERFORMANCE COMPARISON                       ║
╚══════════════════════════════════════════════════════════════════════════════╝

Test Files: 73
Go warm benchmarked: 72, Go cold benchmarked: 72

┌─────────────────────────────────────────────────────────────────────────────┐
│ 🥶 COLD START PERFORMANCE (1st iteration, no warmup)                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                    │  JavaScript  │      Go      │   Difference             │
├────────────────────┼──────────────┼──────────────┼──────────────────────────┤
│ Per File (avg)     │ 993.37µs     │ 930.77µs     │ Similar (~6.3%)          │
│ Per File (median)  │ 546.48µs     │ 484.29µs     │ Similar (~11.4%)         │
│ All Files (total)  │ 71.52ms      │ 67.02ms      │ Similar (~6.3%)          │
└────────────────────┴──────────────┴──────────────┴──────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│ 🔥 WARM PERFORMANCE (after 5 warmup runs)                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                    │  JavaScript  │      Go      │   Difference             │
├────────────────────┼──────────────┼──────────────┼──────────────────────────┤
│ Per File (avg)     │ 428.03µs     │ 882.52µs     │ Go 2.1x slower           │
│ Per File (median)  │ 232.29µs     │ 441.15µs     │ Go 1.9x slower           │
│ All Files (total)  │ 31.25ms      │ 63.54ms      │ Go 2.0x slower           │
└────────────────────┴──────────────┴──────────────┴──────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│ MEMORY & ALLOCATIONS (Go only, averaged per file)                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ Memory per file:         0.56 MB                                              │
│ Allocations per file:    10,296 allocations                                   │
└─────────────────────────────────────────────────────────────────────────────┘

🔥 WARM PERFORMANCE (primary comparison metric):
   🐌 Go is 2.1x SLOWER than JavaScript (warm)

🥶 COLD START PERFORMANCE:
   ⚖️  Cold-start performance is SIMILAR (within 20%)

📈 WARMUP EFFECT:
   JavaScript: 56.9% faster after warmup
   Go:         5.2% faster after warmup
```

### Run Individual Benchmarks

```bash
# JavaScript benchmark suite
pnpm bench:js

# Go benchmark suite
pnpm bench:go:suite
```

## Available Commands

| Command | Description |
|---------|-------------|
| `pnpm bench:compare` | Run both JS and Go benchmarks for comparison (warm + cold) |
| `pnpm bench:js` | Run JavaScript benchmark suite (warm + cold metrics) |
| `pnpm bench:js:detailed` | Run JavaScript benchmarks with per-test results |
| `pnpm bench:js:json` | Output JavaScript results as JSON |
| `pnpm bench:go` | Run Go warm benchmarks (with 5 warmup runs) |
| `pnpm bench:go:cold` | Run Go cold-start benchmarks (no warmup) |
| `pnpm bench:go:suite` | Run Go benchmarks as a suite (warm, faster execution) |

## What's Being Tested?

The benchmarks test **80+ LESS files** from our integration test suite, including:

- ✅ Core LESS features (variables, mixins, operations)
- ✅ Extend functionality
- ✅ Namespacing
- ✅ Math operations (different modes)
- ✅ URL rewriting
- ✅ Import functionality
- ✅ Guards and conditionals
- ✅ Functions
- ✅ Compression

All test files produce **identical CSS output** in both implementations, ensuring we're comparing equivalent functionality.

## Understanding Results

### JavaScript Output Example

```
📊 OVERALL STATISTICS (all tests combined)

🔄 Total Time (Parse + Eval):
   Average: 2.45ms ± 15.3%
   Median:  2.31ms
   Min:     0.52ms
   Max:     12.34ms
```

- **Average:** Mean compilation time per test
- **Median:** Middle value (less affected by outliers)
- **±X%:** Variation (lower is more consistent)

### Go Output Example

```
BenchmarkLargeSuite-8    	     100	  12345678 ns/op	  234567 B/op	    5678 allocs/op
```

- **12345678 ns/op:** Nanoseconds per operation (÷1,000,000 for milliseconds)
  - Example: 12,345,678 ns = 12.3 ms
- **234567 B/op:** Bytes allocated per operation
- **5678 allocs/op:** Number of allocations per operation

## Fair Comparison Tips

✅ **DO:**
- Run benchmarks multiple times and average results
- Close other applications to minimize CPU interference
- Use `pnpm bench:compare` for direct comparison
- Compare median times (more stable than averages)

❌ **DON'T:**
- Compare different benchmark types (e.g., `bench:go` vs `bench:js`)
- Trust single runs - variance is normal
- Run during high system load

## Detailed Documentation

For comprehensive benchmarking information, profiling guides, and advanced usage, see:

📖 [`.claude/benchmarks/BENCHMARKING_GUIDE.md`](./.claude/benchmarks/BENCHMARKING_GUIDE.md)

## Example Workflow

```bash
# 1. Quick comparison
pnpm bench:compare

# 2. Detailed JavaScript results
pnpm bench:js:detailed

# 3. Detailed Go results (individual tests)
pnpm bench:go

# 4. Focus on specific tests
go test -bench=BenchmarkLessCompilation/main/colors ./packages/less/src/less/less_go
```

## System Requirements

- **Node.js:** v14 or higher
- **Go:** 1.18 or higher
- **pnpm:** Latest version

## Benchmark Implementation

### JavaScript
- **Location:** `packages/less/benchmark/suite.js`
- **Method:** Runs 30 iterations (5 warmup) per test file
- **Timing:** High-resolution `process.hrtime()`
- **Output:** Detailed statistics with parse/eval breakdown

### Go
- **Location:** `packages/less/src/less/less_go/benchmark_test.go`
- **Method:** Go's built-in `testing.B` framework (individual file benchmarks)
- **Timing:** Automatic iteration adjustment for accuracy
- **Output:** ns/op, memory allocations, and alloc count per file

### Fair Comparison Methodology

Both benchmarks use **identical methodology** with proper warmup for fair JIT vs AOT comparison:

**Warmup & Measurement:**
- ✅ JavaScript: 5 warmup runs + 25 measured runs per file
- ✅ Go: 5 warmup runs + 25 measured runs per file
- ✅ Identical methodology ensures fair comparison between JIT-compiled (JS) and AOT-compiled (Go) code

**Test Coverage:**
- ✅ Each of the 73 test files is benchmarked individually
- ✅ Same files, same options, same measurement granularity
- ✅ Statistics calculated from individual file results
- ✅ Both produce identical CSS output

**Metrics Reported:**
- 🥶 **Cold-start performance**: First iteration (no warmup) - important for CLI usage
- 🔥 **Warm performance**: After warmup - PRIMARY metric for fair comparison
- 📈 **Warmup effect**: Shows performance improvement from cold to warm

The comparison script (`pnpm bench:compare`) runs both warm and cold-start benchmarks, ensuring you get the complete performance picture for different use cases (~2-3 minutes total).

## Performance Analysis

**Q: Is Go compilation time included in the benchmark?**
**A: No.** The Go benchmark uses `b.ResetTimer()` which excludes all compilation and setup time.

**Q: Are the benchmarks fair now?**
**A: Yes!** Both JavaScript and Go now use identical warmup methodology:
- JavaScript gets 5 warmup runs to allow V8 JIT optimization
- Go gets 5 warmup runs to warm up caches and stabilize performance
- The PRIMARY comparison metric is warm performance (after warmup)
- Cold-start metrics are also reported for real-world CLI usage scenarios

**Q: Why compare warm performance vs cold-start?**
**A: Different use cases:**
- **Warm performance**: Fair comparison of optimized JIT (JS) vs AOT (Go) code
- **Cold-start**: Real-world CLI usage where process starts fresh each time
- JavaScript's JIT needs warmup to reach peak performance, Go doesn't
- Both metrics matter: warm for long-running processes, cold for CLI tools

**Q: Why is Go currently 2.1x slower (warm)?**
**A: Primarily allocations (~10,300 per file) and reflection usage.** The port has been significantly optimized but still has room for improvement. Recent optimizations have reduced allocations by ~78% and improved speed by ~4x. See detailed analysis:
- 📄 [`.claude/benchmarks/PERFORMANCE_ANALYSIS.md`](./.claude/benchmarks/PERFORMANCE_ANALYSIS.md)

**Q: How can I find the bottlenecks?**
**A: Use profiling:**
```bash
pnpm bench:profile
```

This will show CPU hot spots, memory allocations, and allocation hotspots.

**Q: Is this performance acceptable?**
**A: Yes, and improving rapidly!** Recent optimizations (#229-#233) have achieved:
- ✅ **78% reduction** in memory allocations (47k → 10.3k per file)
- ✅ **4x performance improvement** (8.1x slower → 2.1x slower warm)
- ✅ **Cold-start parity** with JavaScript (actually slightly faster!)
- ✅ **80+ tests passing** with identical CSS output

With continued targeted optimization, Go can match or exceed JavaScript warm performance while maintaining its cold-start advantage.

## Contributing

When sharing benchmark results, please include:

1. System info (OS, CPU, RAM)
2. Node.js and Go versions
3. Full output from `pnpm bench:compare`
4. Multiple runs (at least 3)

Example:
```
System: macOS 14.0, M2 Pro (10 cores), 16GB RAM
Node: v20.5.0
Go: 1.22.0

Run 1: JS 2.45ms avg, Go 1.89ms avg
Run 2: JS 2.51ms avg, Go 1.92ms avg
Run 3: JS 2.48ms avg, Go 1.87ms avg

Average: Go ~24% faster
```

---

**Need help?** See the [detailed benchmarking guide](./.claude/benchmarks/BENCHMARKING_GUIDE.md) or open an issue.
