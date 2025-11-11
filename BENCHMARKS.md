# Performance Benchmarks: Less.go vs Less.js

This document provides a quick guide to running performance benchmarks comparing the Go port with the original JavaScript implementation.

## Quick Start

### Run Both Benchmarks (Recommended)

```bash
pnpm bench:compare
```

This runs both JavaScript and Go benchmarks on the same test files, making it easy to compare performance.

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
| `pnpm bench:compare` | Run both JS and Go benchmarks for comparison |
| `pnpm bench:js` | Run JavaScript benchmark suite |
| `pnpm bench:js:detailed` | Run JavaScript benchmarks with per-test results |
| `pnpm bench:js:json` | Output JavaScript results as JSON |
| `pnpm bench:go` | Run Go benchmarks on individual files |
| `pnpm bench:go:parse` | Benchmark only the parsing phase (Go) |
| `pnpm bench:go:eval` | Benchmark only the evaluation phase (Go) |
| `pnpm bench:go:suite` | Run Go benchmarks as a suite (comparable to JS) |

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
- **Method:** Runs 30 iterations (5 warmup) per test
- **Timing:** High-resolution `process.hrtime()`
- **Output:** Detailed statistics with parse/eval breakdown

### Go
- **Location:** `packages/less/src/less/less_go/benchmark_test.go`
- **Method:** Go's built-in `testing.B` framework
- **Timing:** Automatic iteration adjustment for accuracy
- **Output:** ns/op, memory allocations, and alloc count

Both benchmarks test the **exact same files** with the **exact same options** for fair comparison.

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
