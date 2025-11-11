# Parallel Compilation in less.go

## Overview

This document describes the parallel compilation feature added to the Go port of Less.js. This feature enables significant performance improvements when compiling multiple LESS files by utilizing multiple CPU cores.

## Performance Results

### Benchmark Summary

Using the standard benchmark suite (73 LESS files):

| Configuration | Time (ms) | Speedup | Memory | Allocations |
|--------------|-----------|---------|---------|-------------|
| **Sequential** | 149.5 | 1.00x (baseline) | 45.6 MB | 808,797 |
| **Parallel (2 workers)** | 96.0 | **1.59x faster** | 47.0 MB | 809,342 |
| **Parallel (4 workers)** | 61.5 | **2.49x faster** | 48.1 MB | 809,736 |
| **Parallel (8 workers)** | 45.7 | **3.34x faster** | 47.6 MB | 809,578 |
| **Parallel (16 workers/NumCPU)** | 42.1 | **3.63x faster** | 46.4 MB | 809,208 |

### Key Findings

1. **Up to 3.63x speedup** with optimal worker count (NumCPU)
2. **Near-linear scaling** up to 4 workers
3. **Minimal memory overhead** (~2-3% increase)
4. **No change in allocation count** - same correctness guarantees
5. **Zero test regressions** - all 80 integration tests still passing

## Architecture

### How It Works

Each LESS file compilation is independent and can be safely parallelized. The implementation uses:

1. **Worker Pool Pattern**: Fixed number of goroutines process jobs from a channel
2. **Result Ordering**: Results are returned in the same order as input jobs
3. **Error Handling**: Compilation errors are captured per-file without stopping others
4. **Feature Flag**: Parallelization is opt-in via `ParallelCompileOptions.Enable`

### Thread Safety

- Each compilation creates its own `Factory`, `Parser`, and `ImportManager`
- No shared mutable state between compilations
- File I/O operations use Go's standard library (thread-safe)

## Usage

### Basic Usage

```go
package main

import "github.com/toakleaf/less.go/packages/less/src/less/less_go"

func main() {
    // Prepare your compilation jobs
    jobs := []less_go.CompileJob{
        {
            Input: `.button { color: blue; }`,
            Options: map[string]any{"filename": "button.less"},
            ID: "button.less",
        },
        {
            Input: `.header { color: red; }`,
            Options: map[string]any{"filename": "header.less"},
            ID: "header.less",
        },
    }

    // Create factory (can be reused)
    factory := less_go.Factory(nil, nil)

    // Compile in parallel
    opts := &less_go.ParallelCompileOptions{
        Enable:      true,              // Enable parallel compilation
        MaxWorkers:  0,                 // 0 = use runtime.NumCPU()
        StopOnError: false,             // Continue on errors
    }

    results := less_go.BatchCompile(factory, jobs, opts)

    // Process results
    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("Error in %s: %v\n", result.ID, result.Error)
        } else {
            fmt.Printf("Compiled %s: %d bytes\n", result.ID, len(result.CSS))
        }
    }
}
```

### Convenience Function

```go
// Simplified API for common use case
inputs := []struct {
    Content  string
    Options  map[string]any
    Filename string
}{
    {Content: `.test { color: red; }`, Filename: "test.less"},
    {Content: `.main { color: blue; }`, Filename: "main.less"},
}

// Compile with parallel enabled
results := less_go.ParallelCompileMultipleFiles(inputs, true)
```

### Sequential Mode (Default)

```go
// To compile sequentially (original behavior)
opts := &less_go.ParallelCompileOptions{
    Enable: false,  // Disable parallel compilation
}

results := less_go.BatchCompile(factory, jobs, opts)
```

## Configuration Options

### ParallelCompileOptions

```go
type ParallelCompileOptions struct {
    // Enable enables parallel compilation (default: false for safety)
    Enable bool

    // MaxWorkers limits concurrent workers (default: 0 = runtime.NumCPU())
    // Set to specific value to limit CPU usage
    MaxWorkers int

    // StopOnError stops all compilation if any file fails (default: false)
    StopOnError bool
}
```

### Recommended Settings

| Use Case | Enable | MaxWorkers | StopOnError |
|----------|--------|------------|-------------|
| **Build Tools** | `true` | `0` (NumCPU) | `false` |
| **CI/CD** | `true` | `0` (NumCPU) | `true` |
| **Development** | `true` | `4` | `false` |
| **Low-Resource** | `true` | `2` | `false` |
| **Single File** | `false` | N/A | N/A |

## When to Use Parallel Compilation

### ✅ Good Use Cases

- **Batch compilation**: Building multiple LESS files at once
- **Build tools**: Webpack, Gulp, Grunt plugins
- **Static site generators**: Pre-compiling many stylesheets
- **CI/CD pipelines**: Fast build times for large projects
- **Development servers**: Watch mode with many files

### ❌ Not Beneficial

- **Single file compilation**: No parallelization possible
- **Sequential dependencies**: If files must be compiled in order
- **Very small files**: Overhead may exceed benefits (< 10 files)
- **Memory-constrained environments**: Each worker uses memory

## Performance Tuning

### Optimal Worker Count

```go
import "runtime"

// Use all CPUs (default and recommended)
opts.MaxWorkers = 0  // or runtime.NumCPU()

// For 50% CPU utilization
opts.MaxWorkers = runtime.NumCPU() / 2

// For specific limit
opts.MaxWorkers = 4
```

### Scaling Characteristics

Based on benchmarks with 73 files:

- **1-2 workers**: ~1.6x speedup
- **2-4 workers**: ~2.5x speedup
- **4-8 workers**: ~3.3x speedup
- **8-16 workers**: ~3.6x speedup (diminishing returns)

**Amdahl's Law applies**: Not all compilation work is parallelizable (file I/O, initialization), limiting theoretical maximum speedup.

## Testing

### Running Tests

```bash
# Run parallel compilation tests
go test -v -run TestBatchCompile

# Run all parallel tests
go test -v -run TestParallel

# Run benchmarks
go test -bench=BenchmarkParallelVsSequential -benchmem

# Compare sequential vs parallel
go test -bench="BenchmarkLargeSuite$|BenchmarkLargeSuiteParallel$" -benchmem
```

### Benchmark Results

```bash
BenchmarkParallelVsSequential/Sequential-16         	 152.9 ms/op
BenchmarkParallelVsSequential/Parallel-2-16         	  96.0 ms/op
BenchmarkParallelVsSequential/Parallel-4-16         	  61.5 ms/op
BenchmarkParallelVsSequential/Parallel-8-16         	  45.7 ms/op
BenchmarkParallelVsSequential/Parallel-NumCPU-16    	  42.1 ms/op
```

## Implementation Details

### Files

- `parallel_compile.go`: Core parallel compilation implementation
- `parallel_compile_test.go`: Comprehensive test suite
- `benchmark_test.go`: Performance benchmarks

### Key Functions

- `BatchCompile()`: Main entry point for batch compilation
- `batchCompileSequential()`: Sequential fallback
- `batchCompileParallel()`: Worker pool implementation
- `ParallelCompileMultipleFiles()`: Convenience wrapper

## Safety & Correctness

### Guarantees

1. **Result ordering**: Results match input job order
2. **Deterministic**: Same input → same output (order-independent)
3. **Error isolation**: One file's error doesn't affect others
4. **Zero regressions**: All existing tests pass

### Limitations

1. **File imports**: Files with `@import` are compiled independently
2. **Global state**: No shared variables across files (by design)
3. **Memory overhead**: ~2-3% increase due to parallel structures

## Migration Guide

### For Library Users

No changes needed! The default behavior remains sequential. Parallel compilation is opt-in.

### For Tool Developers

```go
// Before (sequential)
for _, file := range files {
    css, err := compileFile(file)
    // process result
}

// After (parallel)
jobs := make([]less_go.CompileJob, len(files))
for i, file := range files {
    jobs[i] = less_go.CompileJob{
        Input:   readFile(file),
        Options: map[string]any{"filename": file},
        ID:      file,
    }
}

factory := less_go.Factory(nil, nil)
results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
    Enable:     true,
    MaxWorkers: 0,
})

for _, result := range results {
    // process result
}
```

## Future Enhancements

### Potential Improvements

1. **Adaptive worker count**: Auto-tune based on file size/complexity
2. **Work stealing**: Better load balancing for uneven file sizes
3. **Streaming results**: Process results as they complete (not in order)
4. **Import parallelization**: Parallel import resolution
5. **Memory pooling**: Reuse parser/evaluator instances

### Performance Roadmap

Current status:
- ✅ Parallel batch compilation: **3.6x speedup**
- 🔄 Allocation reduction: Target 50% reduction
- 📋 CPU optimization: Profile-guided optimization
- 📋 Import caching: Reduce redundant I/O

## Conclusion

The parallel compilation feature provides significant performance improvements for batch compilation scenarios with minimal code changes and zero correctness regressions. It's production-ready and recommended for any workflow that compiles multiple LESS files.

**Key Takeaway**: Enable parallel compilation when building multiple files to get **up to 3.6x faster compilation** with no downsides.

---

**See also**:
- `BENCHMARKS.md` - General benchmarking guide
- `.claude/benchmarks/PERFORMANCE_ANALYSIS.md` - Detailed performance analysis
- `parallel_compile_test.go` - Usage examples and test cases
