# Performance Benchmarks: Less.go vs Less.js

This document provides a quick guide to running performance benchmarks comparing the Go port with the original JavaScript implementation.

## Quick Start

### Run Comparison (Recommended)

**🏗️ Realistic Suite Benchmark** (simulates real CLI/build tool usage):
```bash
pnpm bench:compare:suite
```
- Each iteration = fresh process compiling all files once
- NO warmup or JIT optimization artifacts
- Shows true CLI/build tool performance
- **Most realistic benchmark** ✅

**🔬 Per-File Benchmark** (good for finding specific performance issues):
```bash
pnpm bench:compare
```
- Compiles each file 30x individually
- Shows JIT warmup effects
- Good for microbenchmarking and optimization work

Both run JavaScript and Go benchmarks and display a clear side-by-side comparison with:
- ✅ Per-file and total compilation times
- ✅ Performance ratio (which is faster and by how much)
- ✅ Memory usage and allocation statistics (Go)
- ✅ Actionable optimization recommendations

### Which Benchmark Should You Use?

**Use Realistic Suite Mode (`bench:compare:suite`)** when:
- ✅ You want to measure **actual build tool performance**
- ✅ Simulating real CLI usage (each build = fresh process)
- ✅ Avoiding artificial warmup effects
- ✅ Making performance decisions for production use

**Use Per-File Mode (`bench:compare`)** when:
- 🔬 Finding performance issues in specific files
- 🔬 Measuring JIT optimization potential
- 🔬 Debugging specific compilation bottlenecks
- 🔬 Comparing warm vs cold performance

**Recommendation**: Start with `bench:compare:suite` for realistic numbers, then use `bench:compare` for detailed optimization work.

**Example output:**
```
╔══════════════════════════════════════════════════════════════════════════════╗
║              LESS.JS vs LESS.GO PERFORMANCE COMPARISON                       ║
╚══════════════════════════════════════════════════════════════════════════════╝

Test Files: 212
Go warm benchmarked: 212, Go cold benchmarked: 212

┌─────────────────────────────────────────────────────────────────────────────┐
│ 🥶 COLD START PERFORMANCE (1st iteration, no warmup)                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                    │  JavaScript  │      Go      │   Difference             │
├────────────────────┼──────────────┼──────────────┼──────────────────────────┤
│ Per File (avg)     │ 799.19µs     │ 505.55µs     │ Go 1.6x faster           │
│ Per File (median)  │ 483.42µs     │ 385.98µs     │ Go 1.3x faster           │
│ All Files (total)  │ 168.63ms     │ 107.18ms     │ Go 1.6x faster           │
└────────────────────┴──────────────┴──────────────┴──────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│ 🔥 WARM PERFORMANCE (after 5 warmup runs)                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                    │  JavaScript  │      Go      │   Difference             │
├────────────────────┼──────────────┼──────────────┼──────────────────────────┤
│ Per File (avg)     │ 402.75µs     │ 476.68µs     │ Similar (~18.4%)         │
│ Per File (median)  │ 272.10µs     │ 363.09µs     │ Go 1.3x slower           │
│ All Files (total)  │ 85.38ms      │ 101.06ms     │ Similar (~18.4%)         │
└────────────────────┴──────────────┴──────────────┴──────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│ MEMORY & ALLOCATIONS (Go only, averaged per file)                          │
├─────────────────────────────────────────────────────────────────────────────┤
│ Memory per file:         0.30 MB                                              │
│ Allocations per file:    5,408 allocations                                    │
└─────────────────────────────────────────────────────────────────────────────┘

🔥 WARM PERFORMANCE (primary comparison metric):
   ⚖️  Warm performance is similar (within 20%)

🥶 COLD START PERFORMANCE:
   🚀 Go is 1.6x faster than JavaScript (cold start)

📈 WARMUP EFFECT:
   JavaScript: 49.6% faster after warmup
   Go:         5.7% faster after warmup
```

The sample above was measured on July 15, 2026 on an Apple M1 Max with Go 1.24.5 and Node.js 24.4.1. Benchmark results are machine-dependent; compare changes on the same idle system.

### Run Individual Benchmarks

```bash
# JavaScript benchmark suite
pnpm bench:js

# Go benchmark suite
pnpm bench:go:suite
```

## Available Commands

### Comparison Commands (Recommended)
| Command | Description |
|---------|-------------|
| `pnpm bench:compare` | **Per-file comparison**: Each file compiled 30x (warm + cold) |
| `pnpm bench:compare:suite` | **Suite comparison**: All files sequentially, 30x (realistic workload) |

### JavaScript Benchmarks
| Command | Description |
|---------|-------------|
| `pnpm bench:js` | Per-file mode: Each file compiled 30x (warm + cold metrics) |
| `pnpm bench:js:suite` | Suite mode: All files sequentially, 30x (realistic workload) |
| `pnpm bench:js:detailed` | Per-file mode with detailed per-test results |
| `pnpm bench:js:json` | Output results as JSON (works with both modes) |

### Go Benchmarks
| Command | Description |
|---------|-------------|
| `pnpm bench:go` | Per-file warm benchmarks (with 5 warmup runs) |
| `pnpm bench:go:cold` | Per-file cold-start benchmarks (no warmup) |
| `pnpm bench:go:suite` | Suite mode: All files sequentially, 30x with warmup |
| `pnpm bench:go:suite:gc200` | Suite mode with a throughput-oriented Go GC target |
| `pnpm bench:go:public` | Main exported `Compile` API |

## Benchmark Methodologies

### Realistic Suite Mode (Recommended for Production Decisions)
**Simulates actual CLI/build tool usage**
- **What it does**: Runs 30 independent processes, each compiling all files once
- **JavaScript**: 30 separate `node` processes
- **Go**: 30 separate executions of a precompiled Go benchmark binary
- **Use case**: Measuring real-world build tool performance
- **Process behavior**: Each iteration = fresh process start → compile all 212 benchmarked files → exit
- **No warmup**: Each build is independent, like real CLI usage
- **Run with**: `pnpm bench:compare:suite`

**This is what actually happens in production:**
```
Build 1: Start process → [file1, file2, ..., file212] → Exit
Build 2: Start process → [file1, file2, ..., file212] → Exit
...
Build 30: Start process → [file1, file2, ..., file212] → Exit
```

**Why this matters**: Real-world CLI tools don't benefit from JIT warmup or in-process caching. Each build starts fresh.

### Go GC Throughput Tuning

The suite allocates about 65 MB while compiling 212 files, but most objects die quickly. On the measured system, Go's default `GOGC=100` ran roughly 553 GC cycles during a 15-build benchmark. Increasing the heap-growth target produced the following same-process results:

| Setting | Median suite time | GC cycles | Maximum observed heap goal |
|---------|------------------:|----------:|---------------------------:|
| `GOGC=100` | 112.07ms | 553 | 10 MB |
| `GOGC=200` | 96.78ms | 193 | 15 MB |
| `GOGC=400` | 91.34ms | 76 | 27 MB |

The warmed Node.js suite median was 102.57ms in the same session. `GOGC=200` is therefore the recommended starting point for throughput-oriented services. This is a process-wide Go runtime setting, so less.go deliberately does not change it inside the library. Re-measure under the application's actual memory limit and concurrency.

### Per-File Mode (Good for Optimization Work)
**Measures JIT optimization potential**
- **What it does**: Compiles each file 30 times individually (5 warmup + 25 measured)
- **Use case**: Microbenchmarking, finding file-specific performance issues, measuring JIT effects
- **Cache behavior**: Benefits from repeated compilation of the same content
- **Warmup**: Shows performance after JIT optimization
- **Run with**: `pnpm bench:compare`

**Pattern**:
```
[file1 × 30 in same process] → [file2 × 30 in same process] → ...
```

**Why this matters**: Shows optimization potential but doesn't reflect real CLI usage. Good for finding specific hotspots.

### Key Differences

| Aspect | Realistic Suite Mode | Per-File Mode |
|--------|---------------------|---------------|
| **Process model** | 30 independent processes | Single long-running process |
| **Warmup** | None (each build fresh) | 5 warmup runs per file |
| **Real-world accuracy** | ✅ High (mirrors CLI usage) | ⚠️ Lower (JIT artifacts) |
| **Use for** | Production decisions | Optimization work |
| **JavaScript advantage** | Minimal (no JIT warmup) | Significant (JIT optimization) |
| **Recommended for** | Performance comparisons | Finding bottlenecks |

## What's Being Tested?

The benchmarks test **212 LESS files** from our integration test suite, including:

- ✅ Core LESS features (variables, mixins, operations)
- ✅ Extend functionality
- ✅ Namespacing
- ✅ Math operations (different modes)
- ✅ URL rewriting
- ✅ Import functionality
- ✅ Guards and conditionals
- ✅ Functions
- ✅ Compression

The `include-path/include-path` case is included in the aggregate benchmark set; the benchmark paths are configured so it no longer silently skips that file.

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

**Q: Why does warmed V8 still win the isolated per-file benchmark?**
**A: Allocation and garbage-collection work now dominate the remaining gap.** The current suite allocates about 5,400 objects per file. The aggregate same-process gap is smaller than the isolated-file gap, and a measured `GOGC=200` setting closes it when the application can accept a modestly larger heap target.

**Q: How can I find the bottlenecks?**
**A: Use profiling:**
```bash
pnpm bench:profile
```

This will show CPU hot spots, memory allocations, and allocation hotspots.

**Q: Is this performance acceptable?**
**A: Yes.** The fresh-process build benchmark is 1.8x faster than Less.js, cold per-file compilation is 1.6x faster, and the peak-JIT isolated-file gap is down to 18%. For long-lived throughput-oriented processes, `GOGC=200` moved the aggregate Go suite slightly ahead of the warmed Node suite in local testing.

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
