# Performance Optimization Opportunities Analysis
**Date**: 2025-11-26
**Analyst**: Claude Code Performance Profiling

## Executive Summary

The less.go port shows significant optimization opportunities. Current benchmarks indicate the Go implementation processes test files at approximately **130ms** per suite iteration with **681,488 allocations** and **39.6MB** memory per run.

**Key Finding**: The regex compilation bottleneck identified in earlier analysis has been partially addressed (68 regexes are pre-compiled in `parser_regexes.go`). However, profiling reveals new bottlenecks that account for the current performance gap.

## Current Benchmark Results

```
BenchmarkLargeSuite-16    30    129,982,240 ns/op    39,673,060 B/op    681,488 allocs/op
```

**Per-file breakdown** (assuming ~80 test files):
- ~1.6ms per file average
- ~495KB memory per file
- ~8,500 allocations per file

## Top 5 Performance Bottlenecks

### 1. **fmt.Sprintf Abuse** (HIGH PRIORITY - Quick Win)

**Impact**: 1.9M allocations (9.36% of total allocation count)

**Problem**: 263 occurrences of `fmt.Sprintf` across 65 files, including in hot paths:
- parser.go: 38 calls (many in tracing/debugging)
- declaration.go: 7 calls
- ruleset.go: 11 calls

**Evidence from profiling**:
```
1,933,319 allocs (9.36%) - fmt.Sprintf
```

**Fix**: Replace with `strings.Builder`, `strconv.Itoa`, or string concatenation
- Tracing calls can be gated behind a compile-time flag
- Simple formatting can use `strconv` package

**Expected improvement**: 5-10% allocation reduction, measurable CPU reduction

---

### 2. **Reflection Overhead** (HIGH PRIORITY - Significant Effort)

**Impact**: 2M+ allocations (10%+ of total), 77MB memory

**Problem**: Visitor pattern relies heavily on reflection for method dispatch:
- `reflect.FieldByNameFunc`: 1.1M allocations, 77MB
- `reflect.(*rtype).Method`: 841K allocations
- `reflect.New`: 267K allocations, 21MB

**Location**: `visitor.go` lines 100-250 (slow path with reflection dispatch)

**Evidence**:
```
77.01MB (6.25%) - reflect.(*structType).FieldByNameFunc
1,166,956 allocs (5.65%) - reflect.FieldByNameFunc
841,052 allocs (4.07%) - reflect.(*rtype).Method
```

**Fix**:
1. Expand `DirectDispatchVisitor` interface usage (fast path exists but underutilized)
2. Generate type-specific visitor methods at compile time
3. Replace `FieldByNameFunc` with direct field access via type switches

**Expected improvement**: 15-25% overall speedup

---

### 3. **Ruleset.Eval Cumulative Allocations** (HIGH PRIORITY - Core Path)

**Impact**: 40.77% cumulative allocation count, 586MB cumulative memory

**Problem**: The `Ruleset.Eval` function is the main evaluation entry point and accumulates allocations from:
- `NewRuleset`: 589K allocs, 95MB
- `NewSelector`: 298K allocs, 38MB
- `NewDeclaration`: 432K allocs, 20MB
- Frame copying and variable lookups

**Evidence**:
```
594,729 allocs (2.88% flat, 40.77% cumulative) - Ruleset.Eval
95.02MB (7.71%) - NewRuleset
```

**Fix**:
1. Object pooling for `Ruleset`, `Selector`, `Declaration` nodes
2. Reduce frame copying in `CopyEvalToMap` (49MB, 502K allocs)
3. Lazy initialization of lookup maps in `Ruleset`
4. Pre-allocate slices with capacity hints

**Expected improvement**: 20-30% memory reduction, 15-20% speedup

---

### 4. **Context Copying (CopyEvalToMap)** (MEDIUM PRIORITY - Structural)

**Impact**: 49MB, 502K allocations

**Problem**: Every evaluation context copy creates a new map and copies 16+ fields:
```go
func (e *Eval) CopyEvalToMap(target map[string]any, withClosures bool) {
    target["paths"] = e.Paths
    target["compress"] = e.Compress
    // ... 14 more fields
}
```

**Evidence**:
```
49.04MB (3.98%) - (*Eval).CopyEvalToMap
502,024 allocs (2.43%) - (*Eval).CopyEvalToMap
```

**Fix**:
1. Pass `*Eval` struct pointers instead of copying to maps
2. Use struct embedding instead of map-based context
3. Implement copy-on-write for context modifications

**Expected improvement**: 10-15% allocation reduction

---

### 5. **Function Registry Initialization** (LOW PRIORITY - One-time Cost)

**Impact**: 167MB, 1.8M allocations in init.func1

**Problem**: Multiple `init()` functions register wrapped functions:
- Color functions: ~50 wrappers
- Math functions: ~15 wrappers
- List functions: ~20 wrappers
- Each wrapper allocation persists in registry

**Evidence**:
```
167.02MB (13.55%) - init.func1
1,824,075 allocs (8.83%) - init.func1
```

**Note**: This is a one-time initialization cost, not a per-file cost. However, it affects benchmark warmup and total memory footprint.

**Fix**:
1. Lazy function registration (register on first use)
2. Reduce wrapper allocations by using direct function references
3. Pre-allocate registry with known capacity

**Expected improvement**: Reduced memory footprint, faster cold start

---

## Quick Wins (Implement First)

| Priority | Bottleneck | Effort | Expected Gain | Risk |
|----------|------------|--------|---------------|------|
| 1 | Replace fmt.Sprintf with strconv/Builder | Low | 5-10% allocs | Low |
| 2 | Gate tracing calls with build tag | Low | 5% CPU in parser | None |
| 3 | Pre-allocate slices in hot paths | Low | 5-10% allocs | Low |
| 4 | Object pool for Selector/Element | Medium | 10-15% allocs | Low |

## Long-term Optimization Strategy

### Phase 1: Low-Hanging Fruit (Est. 20-30% improvement)
1. Replace `fmt.Sprintf` in hot paths
2. Add capacity hints to slice allocations
3. Gate debug/trace logging

### Phase 2: Structural Improvements (Est. 30-40% improvement)
1. Implement object pooling for AST nodes
2. Expand DirectDispatchVisitor usage in visitor.go
3. Reduce reflection in visitor pattern
4. Optimize context passing (avoid map copies)

### Phase 3: Algorithmic (Est. 10-20% improvement)
1. String interning for common values (colors, units)
2. Memoization for variable lookups
3. Lazy evaluation where possible

## Profiling Commands

Run fresh profiling:
```bash
pnpm bench:profile
```

Analyze CPU profile:
```bash
go tool pprof -top profiles/cpu.prof
go tool pprof -list=Ruleset.Eval profiles/cpu.prof
```

Analyze memory profile:
```bash
go tool pprof -alloc_objects profiles/mem.prof
go tool pprof -alloc_space profiles/mem.prof
```

## Success Metrics

Track these metrics after each optimization:

1. **Allocations per file**: Target <2,000 (currently ~8,500)
2. **Memory per file**: Target <100KB (currently ~495KB)
3. **Time per file**: Target <1ms (currently ~1.6ms)
4. **fmt.Sprintf calls in profiler**: Target <1% (currently 9.36%)
5. **reflect.* calls in profiler**: Target <5% (currently ~11%)

## Conclusion

The less.go port has achieved **correctness** (83+ perfect CSS matches, 98.4% compilation rate). Performance optimization is the next frontier. The identified bottlenecks are well-understood Go performance anti-patterns with known solutions.

**Priority order for optimization work:**
1. `fmt.Sprintf` replacement (quick win)
2. Reflection reduction in visitor
3. Object pooling for AST nodes
4. Context passing optimization

**Estimated total improvement potential**: 3-5x speedup with focused optimization work.

---

*This analysis is based on profiling run on 2025-11-26. Re-run benchmarks after any code changes to verify improvements.*
