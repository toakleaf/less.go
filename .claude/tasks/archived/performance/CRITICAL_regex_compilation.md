# CRITICAL: Regex Compilation Performance Bug

## Priority: IMMEDIATE / CRITICAL

## Impact: 10x Performance Hit

## Problem Statement

Profiling reveals that **81% of memory allocations** and **76% of allocation count** comes from `regexp.compile`. This means we are compiling regular expressions repeatedly instead of once.

### Profiling Evidence

```
TOP MEMORY ALLOCATIONS:
  1.40GB (30.80%) - regexp/syntax.(*compiler).inst
  1.10GB (24.21%) - regexp/syntax.(*parser).newRegexp
  3.71GB (81.73%) - regexp.compile (CUMULATIVE)

ALLOCATION COUNTS:
  40,664,606 allocations (76%) from regexp.compile
```

**This is the primary cause of the 8-10x slowdown vs JavaScript.**

## Root Cause

In Go, `regexp.MustCompile()` and `regexp.Compile()` are expensive operations that:
1. Parse the regex pattern
2. Compile to bytecode
3. Optimize the pattern
4. Allocate internal structures

If called repeatedly (e.g., in a loop or on every parse), this causes massive overhead.

## Expected Pattern (CORRECT)

```go
// Package-level pre-compiled regexes (compiled once at init)
var (
    identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*`)
    numberRegex     = regexp.MustCompile(`^-?\d+(\.\d+)?`)
    colorRegex      = regexp.MustCompile(`^#[0-9a-fA-F]{3,8}`)
)

func parseIdentifier(input string) string {
    match := identifierRegex.FindString(input)  // Reuse compiled regex
    return match
}
```

## Anti-Pattern (INCORRECT - What We're Likely Doing)

```go
func parseIdentifier(input string) string {
    // WRONG: Compiling on every call!
    regex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*`)
    match := regex.FindString(input)
    return match
}
```

## Finding the Culprits

### Step 1: Search for Dynamic Regex Compilation

```bash
# Find all regexp.Compile calls
grep -rn "regexp\.Compile\|regexp\.MustCompile" packages/less/src/less/less_go/*.go

# Find string-based regex compilation
grep -rn "regexp\.MustCompile.*fmt\.Sprintf" packages/less/src/less/less_go/*.go
```

### Step 2: Check Parser

The parser is likely the main culprit:

```bash
# Check parser.go for regex usage
grep -n "regexp\|Compile" packages/less/src/less/less_go/parser.go
```

### Step 3: Profile to Confirm

```bash
# Get detailed call stack
go tool pprof -text -nodecount=50 profiles/mem.prof | grep -A5 -B5 "regexp.compile"
```

## Solution

### Phase 1: Move Regexes to Package Level (HIGH PRIORITY)

1. **Identify all regexes** in parser and other files
2. **Declare as package-level vars**:
   ```go
   var (
       identRegex = regexp.MustCompile(`...`)
       numRegex   = regexp.MustCompile(`...`)
       // etc.
   )
   ```
3. **Use the pre-compiled regexes** instead of compiling inline

### Phase 2: Lazy Compilation (if needed)

For regexes that depend on runtime values:

```go
var (
    cachedRegexes = make(map[string]*regexp.Regexp)
    regexMutex    sync.RWMutex
)

func getOrCompileRegex(pattern string) *regexp.Regexp {
    regexMutex.RLock()
    if re, exists := cachedRegexes[pattern]; exists {
        regexMutex.RUnlock()
        return re
    }
    regexMutex.RUnlock()

    regexMutex.Lock()
    defer regexMutex.Unlock()

    // Double-check after acquiring write lock
    if re, exists := cachedRegexes[pattern]; exists {
        return re
    }

    re := regexp.MustCompile(pattern)
    cachedRegexes[pattern] = re
    return re
}
```

## Expected Results

Based on profiling:
- **Current**: 3.71GB memory, 40M+ allocations from regex
- **After fix**: ~0.1GB memory, ~100 allocations from regex
- **Expected speedup**: **5-10x faster** overall
- **Memory reduction**: **80%+ reduction**

This single fix should bring us from **8-10x slower** to **1-2x slower** (or possibly faster than JS).

## Verification

After fixing, run benchmarks again:

```bash
pnpm bench:compare
pnpm bench:profile
```

Look for:
1. `regexp.compile` should be near 0% in profiling
2. Allocations should drop from 3.4M to <500k
3. Performance should be 5-10x faster

## Files to Check

Priority order (most likely culprits):

1. **`parser.go`** - Main parser, likely compiling regexes per parse
2. **`tokenizer.go`** or similar - Token matching
3. **`*.go` files with pattern matching**

### Search Commands

```bash
# Find all regex compilations
cd packages/less/src/less/less_go
grep -l "regexp.Compile\|regexp.MustCompile" *.go

# Count regex compilations per file
for f in *.go; do
    count=$(grep -c "regexp.MustCompile\|regexp.Compile" "$f" 2>/dev/null || echo 0)
    if [ "$count" -gt 0 ]; then
        echo "$f: $count"
    fi
done | sort -t: -k2 -rn
```

## Task for LLM Agent

**Objective**: Fix regex compilation performance bug

**Steps**:
1. Run search commands above to find all regex compilations
2. For each file with regexes:
   - Identify which regexes are static (same pattern every time)
   - Move static regexes to package-level `var` declarations
   - Replace compilation calls with pre-compiled regex usage
3. For dynamic regexes (pattern changes at runtime):
   - Implement caching mechanism
   - Use `sync.Map` or mutex-protected map
4. Run benchmarks to verify improvement
5. Run profiler to confirm `regexp.compile` is no longer dominant

**Success Criteria**:
- `regexp.compile` < 5% of total allocations in profiler
- Benchmark shows 5-10x speedup
- All tests still pass: `pnpm test:go`

## Example Fix

**Before**:
```go
func (p *Parser) parseIdentifier() string {
    regex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*`)  // COMPILED EVERY CALL!
    match := regex.FindString(p.input)
    return match
}
```

**After**:
```go
// At package level
var identifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*`)

func (p *Parser) parseIdentifier() string {
    match := identifierRegex.FindString(p.input)  // Reuse pre-compiled
    return match
}
```

## References

- Go regexp package: https://pkg.go.dev/regexp
- Performance best practices: https://go.dev/doc/effective_go#regexp
- Profiling guide: https://go.dev/blog/pprof

---

**This is NOT an acceptable performance issue** - this is a critical bug that should be fixed immediately. The fix is straightforward and will likely solve most of the performance gap.
