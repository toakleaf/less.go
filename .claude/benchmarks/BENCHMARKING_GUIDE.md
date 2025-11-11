# Less.go Performance Benchmarking Guide

This guide explains how to run performance benchmarks comparing the Go port of Less.js with the original JavaScript implementation.

## Overview

We've created comprehensive benchmark suites that test the same LESS files with the same options in both implementations. This ensures fair, apples-to-apples comparisons.

**Test Coverage:**
- 80+ passing integration test files
- Multiple test suites with different options (math modes, compression, URL rewriting, etc.)
- Separate benchmarks for parsing, evaluation, and full compilation

## Quick Start

### Compare Both Implementations (Recommended)

```bash
pnpm bench:compare
```

This runs both JavaScript and Go benchmarks and displays a comprehensive comparison table showing:
- **Performance metrics**: Per-file and total compilation times for both implementations
- **Direct comparison**: Clear indication of which is faster and by what ratio
- **Memory statistics**: Go's memory usage and allocation counts
- **Recommendations**: Actionable optimization suggestions if performance gaps are significant

**Fair Comparison Methodology**: Both implementations benchmark each test file individually using identical methodology, ensuring true apples-to-apples comparison.

### Run Individual Benchmarks

**JavaScript benchmarks:**
```bash
# Basic summary
pnpm bench:js

# Detailed per-test results
pnpm bench:js:detailed

# JSON output for programmatic analysis
pnpm bench:js:json
```

**Go benchmarks:**
```bash
# Individual file benchmarks - used by pnpm bench:compare for fair comparison
pnpm bench:go

# Batch suite benchmark (all files in one run, faster but less granular)
pnpm bench:go:suite

# Parse-only benchmarks
pnpm bench:go:parse

# Eval-only benchmarks
pnpm bench:go:eval
```

## Benchmark Types

### 1. Individual File Benchmarks (Recommended for Fair Comparison)

**JavaScript:** `pnpm bench:js`
**Go:** `pnpm bench:go`
**Comparison:** `pnpm bench:compare` (uses both of the above)

Runs benchmarks on each test file individually:
- JavaScript: 30 iterations per file (5 warmup + 25 measured)
- Go: Auto-determined iterations per file via `testing.B` framework
- Provides per-file statistics and overall averages

**Best for:**
- Fair apples-to-apples comparison between JS and Go
- Identifying specific files that are slower/faster
- Memory profiling with `-benchmem`
- CPU profiling with `-cpuprofile`

### 2. Batch Suite Benchmark (Faster, Less Granular)

**Go only:** `pnpm bench:go:suite`

Runs all test files as one large batch, providing overall statistics:
- Lower benchmark framework overhead
- Faster execution
- Good for quick performance checks

**Best for:** Quick overall performance snapshots (not used for JS/Go comparison)

### 3. Phase-Specific Benchmarks

**Parse only:**
```bash
pnpm bench:go:parse
```

**Eval only:**
```bash
pnpm bench:go:eval
```

**Best for:** Understanding which phase (parsing vs evaluation) accounts for performance differences.

## Understanding the Results

### Comparison Tool Output (`pnpm bench:compare`)

The comparison tool produces a formatted table that makes it easy to see relative performance:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ COMPILATION TIME                                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                    │  JavaScript  │      Go      │   Difference             │
├────────────────────┼──────────────┼──────────────┼──────────────────────────┤
│ Per File (avg)     │ 446.32µs     │ 3.62ms       │ Go 8.1x slower           │
│ Per File (median)  │ 277.04µs     │ N/A          │                          │
│ All Files (total)  │ 32.58ms      │ 264.06ms     │ Go 8.1x slower           │
└────────────────────┴──────────────┴──────────────┴──────────────────────────┘

🐌 Go is 8.1x SLOWER than JavaScript

💡 Optimization Opportunities:
  • Profile with: go test -bench=BenchmarkLargeSuite -cpuprofile=cpu.prof
  • Analyze with: go tool pprof cpu.prof
  • Check for: excessive allocations, string operations, reflection
```

**What to look for:**
- **Per File (avg)**: Average compilation time per test file - most important metric
- **Performance verdict**: Clear emoji indicator (🚀 faster, ⚖️ similar, 🐌 slower)
- **Optimization recommendations**: Shown automatically if there's a significant performance gap

### Individual Benchmark Output

**JavaScript Output** (`pnpm bench:js`):

```
📊 OVERALL STATISTICS (all tests combined)

🔄 Total Time (Parse + Eval):
   Average: 2.45ms ± 15.3%
   Median:  2.31ms
   Min:     0.52ms
   Max:     12.34ms
```

- **Average:** Mean time across all runs (excluding warmup)
- **Median:** Middle value, less affected by outliers
- **±X%:** Variation percentage (lower is more consistent)
- **Min/Max:** Range of times observed

**Go Output** (`pnpm bench:go:suite`):

```
BenchmarkLargeSuite-8    	     100	  12345678 ns/op	  234567 B/op	    5678 allocs/op
```

- **BenchmarkLargeSuite-8:** Test name with number of CPU cores used
- **100:** Number of iterations run
- **12345678 ns/op:** Nanoseconds per operation (12.3ms in this example)
- **234567 B/op:** Bytes allocated per operation
- **5678 allocs/op:** Number of allocations per operation

**To convert ns/op to ms:** Divide by 1,000,000
- 1,000,000 ns/op = 1.0 ms
- 10,000,000 ns/op = 10.0 ms

**Note**: Don't compare these individual outputs directly - use `pnpm bench:compare` for accurate comparison!

## Fair Comparison Guidelines

### ✅ DO:

1. **Use the comparison tool** - It handles metric normalization automatically
   ```bash
   pnpm bench:compare
   ```

2. **Run multiple times** - Performance can vary between runs
   ```bash
   # Run 3-5 times and look for consistent results
   pnpm bench:compare
   pnpm bench:compare
   pnpm bench:compare
   ```

3. **Close other applications** - Minimize background CPU usage
   - Close browsers, IDEs, Slack, Docker containers, etc.
   - Check Activity Monitor/Task Manager before running

4. **Look at multiple metrics**
   - Per-file average (most important)
   - Total time (for batch processing)
   - Variation/consistency
   - Memory usage (for large-scale deployments)

5. **Check consistency** - Lower variation percentages mean more reliable results

### ❌ DON'T:

1. **Don't manually compare raw outputs** - Use `pnpm bench:compare` instead
   - The comparison tool normalizes metrics automatically
   - Raw Go/JS outputs use different measurement approaches

2. **Don't trust single runs** - Variance is normal, run multiple times

3. **Don't compare during high system load**
   - Close browsers, IDEs, etc.
   - Disable automatic backups/updates
   - Check `top`/`htop` to verify low CPU usage

4. **Don't compare different test sets**
   - Both benchmarks must test the same files
   - Verify test counts match in output

## Advanced Usage

### Custom Number of Runs (JavaScript)

```bash
node packages/less/benchmark/suite.js --runs=50
```

### Custom Benchmark Time (Go)

```bash
go test -bench=BenchmarkLargeSuite -benchtime=30s ./packages/less/src/less/less_go
```

### Memory Profiling (Go)

```bash
go test -bench=BenchmarkLargeSuite -benchmem -memprofile=mem.prof ./packages/less/src/less/less_go
go tool pprof mem.prof
```

### CPU Profiling (Go)

```bash
go test -bench=BenchmarkLargeSuite -cpuprofile=cpu.prof ./packages/less/src/less/less_go
go tool pprof cpu.prof
```

### Filter Specific Tests (Go)

```bash
# Only benchmark main suite tests
go test -bench=BenchmarkLessCompilation/main -benchtime=10s ./packages/less/src/less/less_go

# Only benchmark namespacing tests
go test -bench=BenchmarkLessCompilation/namespacing -benchtime=10s ./packages/less/src/less/less_go
```

## Interpreting Performance Differences

### Expected Patterns

1. **Parsing:** Go is typically faster at raw parsing (compiled vs interpreted)
2. **Evaluation:** Performance depends on implementation complexity
3. **Memory:** Go may use more memory but with better predictability
4. **Consistency:** Go often shows lower variation (more predictable)

### What to Look For

- **Overall compilation time:** The most important metric for end users
- **Consistency:** Lower variation means more predictable build times
- **Memory usage:** Important for large projects or CI environments
- **Phase breakdown:** Identifies optimization opportunities

## Test File Coverage

The benchmarks test 80+ files across these suites:

- **main** (47 files): Core LESS features
- **namespacing** (11 files): Namespace operations
- **math-parens** (4 files): Math with parens
- **math-parens-division** (4 files): Division operations
- **math-always** (2 files): Always-on math
- **compression** (1 file): Compressed output
- **units-strict** (1 file): Strict unit checking
- **units-no-strict** (1 file): Lenient units
- **rewrite-urls** (1 file): URL rewriting
- **include-path** (1 file): Import paths

All test files are verified to produce identical CSS output in both implementations.

## Troubleshooting

### JavaScript benchmark fails

**Check Node.js version:**
```bash
node --version  # Should be 14+
```

**Install dependencies:**
```bash
pnpm install
```

### Go benchmark fails

**Verify Go is installed:**
```bash
go version  # Should be 1.18+
```

**Check test files exist:**
```bash
ls packages/test-data/less/_main/
```

### Results seem inconsistent

1. Close other applications
2. Run multiple times and average
3. Increase benchmark time: `--runs=50` (JS) or `-benchtime=30s` (Go)
4. Check system load: `top` or Activity Monitor

## Example Workflow

```bash
# 1. Run quick comparison
pnpm bench:compare

# 2. If you want more detail, run with verbose output
pnpm bench:js:detailed

# 3. For specific analysis, use Go's detailed benchmarks
pnpm bench:go

# 4. If investigating a specific area, filter tests
go test -bench=BenchmarkLessCompilation/main/colors -benchtime=20s ./packages/less/src/less/less_go

# 5. Profile memory if needed
go test -bench=BenchmarkLargeSuite -benchmem -memprofile=mem.prof ./packages/less/src/less/less_go
```

## Contributing Benchmark Results

When sharing benchmark results:

1. Include system information:
   - OS and version
   - CPU model and cores
   - RAM amount
   - Node.js version
   - Go version

2. Share full output from `pnpm bench:compare`

3. Run at least 3 times and include all results

4. Note any unusual system conditions (background processes, thermal throttling, etc.)

Example:
```
System: macOS 13.0, M1 Pro (8 cores), 16GB RAM
Node: v18.12.0
Go: 1.21.0

Run 1: JS avg 2.45ms, Go avg 1.89ms
Run 2: JS avg 2.51ms, Go avg 1.92ms
Run 3: JS avg 2.48ms, Go avg 1.87ms

Result: Go ~24% faster on average
```

## Summary

For most use cases, run:
```bash
pnpm bench:compare
```

This gives you a fair, easy-to-understand comparison of the two implementations.
