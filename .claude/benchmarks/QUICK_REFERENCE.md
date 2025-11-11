# Benchmark Quick Reference

## TL;DR - Run This

```bash
pnpm bench:compare
```

This runs both JavaScript and Go benchmarks on the same 73+ test files for direct comparison.

## All Commands

```bash
# Direct comparison (recommended)
pnpm bench:compare

# JavaScript only
pnpm bench:js                # Summary stats
pnpm bench:js:detailed       # Per-test breakdown
pnpm bench:js:json           # JSON output

# Go only
pnpm bench:go                # Individual file benchmarks (detailed)
pnpm bench:go:suite          # Suite benchmark (comparable to JS)
pnpm bench:go:parse          # Skipped - can't separate phases
pnpm bench:go:eval           # Skipped - can't separate phases
```

## Understanding Output

### JavaScript (ms/µs)
- Average: Mean time per test
- Median: Middle value (more stable)
- Min/Max: Range observed
- ±X%: Variation (lower = more consistent)

### Go (ns/op)
- Divide by 1,000,000 to convert to ms
- Example: 10,000,000 ns/op = 10 ms
- Also shows memory (B/op) and allocations (allocs/op)

## Quick Comparison

1. Run: `pnpm bench:compare`
2. Note the JavaScript average time
3. Note the Go ns/op time (convert to ms)
4. Compare!

Example:
- JS: 1.34ms average
- Go: 549,660,616 ns/op = ~550ms for all 73 files = ~7.5ms per file
- Result: JS faster for small files, Go might be faster for large batches

## Files

- **Go benchmark**: `packages/less/src/less/less_go/benchmark_test.go`
- **JS benchmark**: `packages/less/benchmark/suite.js`
- **Documentation**: `BENCHMARKS.md` and `.claude/benchmarks/BENCHMARKING_GUIDE.md`

## What's Tested

80+ LESS files from passing integration tests:
- Core features (variables, mixins, operations)
- Extend functionality (7 tests)
- Namespacing (11 tests)
- Math operations (10 tests)
- URL rewriting (4 tests)
- Import functionality
- Guards and conditionals
- Compression
- And more!

All files produce **identical CSS** in both implementations.
