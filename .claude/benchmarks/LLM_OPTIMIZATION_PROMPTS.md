# LLM Session Prompts for Performance Optimization

Copy and paste these prompts into fresh LLM sessions to tackle specific optimization tasks.

---

## Prompt 1: Pre-allocate Slices in findMatch (Easiest, 5-7% Impact)

```
I'm working on optimizing a Go port of Less.js. Profiling shows that the `findMatch` function in `extend_visitor.go` is allocating 82MB (6.51% of total) because it creates slices without capacity hints.

Your task:
1. Read `packages/less/src/less/less_go/extend_visitor.go` lines 690-750
2. Find the slice creations at lines 698 and 700:
   ```go
   potentialMatches := make([]any, 0)
   matches := make([]any, 0)
   ```
3. Add capacity hints based on typical usage (start with 10 and 5)
4. Run tests: `LESS_GO_QUIET=1 pnpm -w test:go`
5. Benchmark before and after: `pnpm -w bench:go:suite`
6. Measure the improvement in allocations

Context:
- We need all 80 integration tests to still pass
- The current benchmark shows 815k allocs/op, 47MB/op
- This optimization should reduce allocations by ~5-10k per op

References:
- `.claude/benchmarks/LOW_HANGING_FRUIT_TASKS.md` - Task 1
- `.claude/benchmarks/PERFORMANCE_BOTTLENECKS.md` - For background

Expected outcome:
- Tests pass ✅
- Allocations reduced by 5-10k
- 5-7% faster execution
```

---

## Prompt 2: Replace fmt.Sprintf with strconv (Medium Difficulty, 2-3% Impact)

```
I'm optimizing a Go port of Less.js. Profiling shows `fmt.Sprintf` is called 1.6M times and allocates 29MB (2.3% of total). Most of these calls are simple type conversions that should use `strconv` instead.

Your task:
1. Focus on these files (highest impact first):
   - `packages/less/src/less/less_go/parser.go` (38 Sprintf calls)
   - `packages/less/src/less/less_go/quoted.go` (8 calls)
   - `packages/less/src/less/less_go/declaration.go` (7 calls)

2. Replace simple conversions:
   - `fmt.Sprintf("%d", n)` → `strconv.Itoa(n)`
   - `fmt.Sprintf("%v", value)` → type assertion + appropriate converter
   - `fmt.Sprintf("%s", s)` → direct use if already a string

3. Keep fmt.Sprintf for complex formatting (multiple values, padding, etc.)

4. Test after EACH file:
   ```bash
   LESS_GO_QUIET=1 pnpm -w test:go
   ```

5. Benchmark to confirm improvement:
   ```bash
   pnpm -w bench:go:suite
   ```

Important constraints:
- Don't break the 80 passing integration tests
- One file at a time, commit after each
- Profile-first: focus on hot paths, not every Sprintf

References:
- `.claude/benchmarks/LOW_HANGING_FRUIT_TASKS.md` - Task 2
- Go strconv docs: https://pkg.go.dev/strconv

Expected outcome:
- All tests pass ✅
- Allocations reduced by ~100k per op
- 2-3% faster execution
```

---

## Prompt 3: Pre-allocate Slices in Hot Loops (Medium, 1-2% Impact)

```
I'm optimizing a Go port of Less.js. Many hot-path functions create slices in loops without pre-allocating capacity, causing repeated reallocation and copying.

Your task:
1. Profile to identify hot-path loops:
   ```bash
   pnpm bench:profile
   go tool pprof -top profiles/mem.prof | head -30
   ```

2. Focus on these functions (from prior profiling):
   - `(*Ruleset).Eval` in `ruleset.go`
   - `(*MixinDefinition).EvalParams` in `mixin_definition.go`
   - `(*Selector).MixinElements` in `selector.go`
   - `(*JoinSelectorVisitor).VisitRuleset` in `join_selector_visitor.go`

3. Find patterns like:
   ```go
   results := []Type{}
   for _, item := range items {
       results = append(results, process(item))
   }
   ```

4. Replace with:
   ```go
   results := make([]Type, 0, len(items))  // Pre-allocate
   for _, item := range items {
       results = append(results, process(item))  // No reallocation
   }
   ```

5. When output size equals input size, use even better pattern:
   ```go
   results := make([]Type, len(items))  // Exact size
   for i, item := range items {
       results[i] = process(item)  // Direct assignment
   }
   ```

Testing approach:
- One function at a time
- Test after each: `LESS_GO_QUIET=1 pnpm -w test:go`
- Commit after each successful change
- Benchmark: `pnpm -w bench:go:suite`

References:
- `.claude/benchmarks/LOW_HANGING_FRUIT_TASKS.md` - Task 3
- `.claude/benchmarks/PERFORMANCE_BOTTLENECKS.md` - Section 3

Expected outcome:
- All tests pass ✅
- Allocations reduced by ~50k per op
- 1-2% faster execution
```

---

## Prompt 4: Use strings.Builder for Concatenation (Easy, 1-2% Impact)

```
I'm optimizing a Go port of Less.js. String concatenation with the `+` operator creates a new string on each operation, causing excessive allocations in loops.

Your task:
1. Find string concatenation in loops:
   ```bash
   grep -B5 '+ "' packages/less/src/less/less_go/*.go | grep -v _test.go
   grep -B5 '+= "' packages/less/src/less/less_go/*.go | grep -v _test.go
   ```

2. Target these files (known to have issues):
   - `to_css_visitor.go` - CSS output generation
   - `selector.go` - Selector string building
   - `element.go` - Element string building

3. Replace concatenation loops:

   **Before:**
   ```go
   result := ""
   for _, item := range items {
       result += item + ","
   }
   ```

   **After:**
   ```go
   var builder strings.Builder
   builder.Grow(len(items) * 10)  // Rough estimate
   for _, item := range items {
       builder.WriteString(item)
       builder.WriteString(",")
   }
   result := builder.String()
   ```

4. For single concatenations (not in loops), keep using `+` - it's fine

Testing:
- Test after each file: `LESS_GO_QUIET=1 pnpm -w test:go`
- Benchmark: `pnpm -w bench:go:suite`
- Look for allocation reduction

References:
- `.claude/benchmarks/LOW_HANGING_FRUIT_TASKS.md` - Task 4
- Go strings.Builder docs: https://pkg.go.dev/strings#Builder

Expected outcome:
- All tests pass ✅
- Allocations reduced by ~30k per op
- 1-2% faster execution
```

---

## Prompt 5: Combined Optimization Sprint (For Experienced Developers)

```
I'm optimizing the performance of a Go port of Less.js. The current benchmark shows:
- 147ms per op (73 LESS files)
- 47 MB allocated per op
- 815k allocations per op

I need to tackle all low-hanging fruit optimizations to achieve ~10-15% speedup while maintaining 100% test pass rate (80 perfect CSS matches).

Your task - complete all 4 optimizations:

1. **Pre-allocate slices in findMatch** (5-7% impact)
   - File: `extend_visitor.go` lines 698-700
   - Add capacity hints to slice creation

2. **Replace fmt.Sprintf with strconv** (2-3% impact)
   - Files: `parser.go`, `quoted.go`, `declaration.go`
   - Replace simple type conversions

3. **Pre-allocate slices in hot loops** (1-2% impact)
   - Files: `ruleset.go`, `mixin_definition.go`, `selector.go`
   - Add capacity to slice creation in loops

4. **Use strings.Builder for concatenation** (1-2% impact)
   - Files: `to_css_visitor.go`, `selector.go`, `element.go`
   - Replace `+` in loops with strings.Builder

Workflow:
1. Start with Task 1 (easiest)
2. Run baseline benchmark: `pnpm -w bench:go:suite > baseline.txt`
3. Complete each task one at a time
4. Test after EACH change: `LESS_GO_QUIET=1 pnpm -w test:go`
5. Benchmark after each task
6. Commit after each successful task
7. Document actual improvements

Success criteria:
- All 80 integration tests pass ✅
- Total speedup: 10-15%
- Allocation reduction: 185-240k per op
- Memory reduction: 20-33 MB per op

Read these files first:
- `.claude/benchmarks/PERFORMANCE_BOTTLENECKS.md` - Understanding the issues
- `.claude/benchmarks/LOW_HANGING_FRUIT_TASKS.md` - Detailed task specs
- `.claude/VALIDATION_REQUIREMENTS.md` - Testing requirements

Expected final results:
- Before: 147ms, 47MB, 815k allocs
- After: 125-135ms, 27-35MB, 575-630k allocs
```

---

## Prompt 6: Profile-Guided Optimization (Advanced)

```
I'm optimizing a Go port of Less.js using profile-guided optimization. I need you to identify and fix the actual bottlenecks based on profiling data, not assumptions.

Your approach:
1. Run profiling:
   ```bash
   pnpm bench:profile
   ```

2. Analyze memory hotspots:
   ```bash
   go tool pprof -top profiles/mem.prof | head -30
   go tool pprof -top -cum profiles/mem.prof | head -30
   ```

3. Analyze CPU hotspots:
   ```bash
   go tool pprof -top profiles/cpu.prof | head -30
   ```

4. For each function consuming >3% of resources:
   - Read the source code
   - Identify allocation patterns
   - Propose low-risk optimization
   - Estimate impact

5. Prioritize by impact/risk ratio:
   - High impact + Low risk = Do first
   - High impact + High risk = Document for later
   - Low impact + Any risk = Skip

6. Implement top 3-5 optimizations

7. Re-profile to confirm improvement:
   ```bash
   pnpm bench:profile
   # Compare before/after in profiles/
   ```

Constraints:
- Don't touch architectural code (visitor pattern, Node allocation)
- Focus on algorithmic improvements and allocation reduction
- Must maintain all 80 passing tests
- Must show measurable improvement (>2%)

References:
- `.claude/benchmarks/PERFORMANCE_BOTTLENECKS.md` - Known bottlenecks
- `.claude/benchmarks/BENCHMARKING_GUIDE.md` - Profiling tools
- Current profiling data: `profiles/cpu.prof` and `profiles/mem.prof`

Expected outcome:
- Clear identification of top 5 hotspots
- Low-risk fixes for top 3
- 5-10% total speedup
- Detailed before/after profiling comparison
```

---

## General Tips for All Sessions

### Before You Start
1. Pull latest: `git fetch origin master && git merge origin/master`
2. Verify tests pass: `LESS_GO_QUIET=1 pnpm -w test:go`
3. Run baseline benchmark: `pnpm -w bench:go:suite > baseline.txt`

### While Working
1. One change at a time
2. Test frequently
3. Commit small, focused changes
4. Measure everything

### If Tests Fail
1. Check the specific test output
2. Use `LESS_GO_DEBUG=1` for more details
3. Compare CSS with `LESS_GO_DIFF=1`
4. Revert if stuck: `git checkout <file>`

### Success Metrics
- ✅ All 80 integration tests pass
- ✅ Benchmark shows improvement
- ✅ Code remains readable
- ✅ No regressions in any category

---

## Questions to Ask

If you're an LLM working on these tasks, ask the user:

1. **"Should I start with the easiest task (Task 1) or tackle all of them?"**
2. **"Would you like me to profile first to confirm these are still the hotspots?"**
3. **"Should I commit after each file or after each complete task?"**
4. **"Do you want detailed explanations or just the code changes?"**
5. **"If tests fail, should I try to fix or revert and document?"**

---

## Expected Timeline

- **Task 1 (Pre-allocate findMatch):** 30 minutes
- **Task 2 (Replace fmt.Sprintf):** 1-2 hours
- **Task 3 (Pre-allocate loops):** 2-3 hours
- **Task 4 (strings.Builder):** 1-2 hours
- **All tasks combined:** 5-8 hours for a fresh LLM session

---

## Final Checklist

After completing optimizations:

- [ ] All tests pass: `LESS_GO_QUIET=1 pnpm -w test:go`
- [ ] 80 perfect CSS matches confirmed
- [ ] Benchmark shows improvement: `pnpm -w bench:go:suite`
- [ ] Changes committed with clear messages
- [ ] Before/after metrics documented
- [ ] Code reviewed for readability
- [ ] No new compiler warnings
- [ ] Updated this file with actual results
