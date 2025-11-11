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

This runs both JavaScript and Go benchmarks in sequence, making it easy to compare results.

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
# Full compilation (parse + eval) - recommended for comparison
pnpm bench:go:suite

# Individual test benchmarks (more detailed, slower)
pnpm bench:go

# Parse-only benchmarks
pnpm bench:go:parse

# Eval-only benchmarks
pnpm bench:go:eval
```

## Benchmark Types

### 1. Full Suite Benchmark (Recommended for Overall Comparison)

**JavaScript:** `pnpm bench:js`
**Go:** `pnpm bench:go:suite`

Runs all test files in a batch, providing overall statistics:
- Average, median, min, max times
- Variation and standard deviation
- Separate timing for parse and eval phases

**Best for:** Getting an overall sense of performance differences.

### 2. Individual Test Benchmarks (Detailed Analysis)

**Go only:** `pnpm bench:go`

Runs benchmarks on each test file individually using Go's built-in benchmark framework.

**Best for:**
- Identifying specific files that are slower/faster
- Memory profiling with `-benchmem`
- CPU profiling with `-cpuprofile`

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

### JavaScript Output

```
📊 OVERALL STATISTICS (all tests combined)

🔄 Total Time (Parse + Eval):
   Average: 2.45ms ± 15.3%
   Median:  2.31ms
   Min:     0.52ms
   Max:     12.34ms

📝 Parse Time:
   Average: 1.20ms ± 18.2%

⚡ Eval Time:
   Average: 1.25ms ± 14.1%
```

- **Average:** Mean time across all runs (excluding warmup)
- **Median:** Middle value, less affected by outliers
- **±X%:** Variation percentage (lower is more consistent)
- **Min/Max:** Range of times observed

### Go Output

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

## Fair Comparison Guidelines

### ✅ DO:

1. **Run multiple times** - Performance can vary between runs
   ```bash
   pnpm bench:compare  # Run several times and average the results
   ```

2. **Close other applications** - Minimize background CPU usage

3. **Use the suite benchmarks** - They test the same files with the same options
   ```bash
   pnpm bench:js         # JavaScript suite
   pnpm bench:go:suite   # Go suite (comparable to JS)
   ```

4. **Compare median times** - Less affected by outliers than averages

5. **Check consistency** - Lower variation percentages mean more reliable results

### ❌ DON'T:

1. **Don't compare individual Go benchmarks to JS suite** - They use different methodologies
   - `pnpm bench:go` runs each file separately (more detailed, higher overhead)
   - `pnpm bench:go:suite` runs all files together (comparable to JS)

2. **Don't compare single runs** - Always run multiple times

3. **Don't compare during high system load** - Close browsers, IDEs, etc.

4. **Don't mix debug/release builds** - Always use optimized builds

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
