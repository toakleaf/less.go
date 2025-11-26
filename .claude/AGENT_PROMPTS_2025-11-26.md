# 10 Ready-to-Use Agent Prompts - Session 2025-11-26

**Status**: All 8 output differences are ripe for parallel fixing
**Success Rate Target**: Get from 45.7% → 55%+ perfect matches
**Baseline to maintain**: 84 perfect matches, 88 error tests passing, NO regressions

---

## Prompt 1: Fix import-reference CSS Output Suppression

**Time**: 2-3 hours | **Impact**: +2 perfect matches | **Difficulty**: Medium

```
You are working on the less.go port of less.js. This task focuses on fixing how
import references work. Currently 2 tests fail: import-reference and
import-reference-issues.

The issue: Files imported with @import (reference) "file.less" should NOT output
their CSS by default. Only explicitly referenced parts (via :extend() or as mixins)
should appear in the output.

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (all unit tests must pass - baseline: 2,304 tests)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect matches)
- Check there are NO regressions in these baseline numbers before you start

Steps:
1. Look at test data:
   - Input: packages/test-data/less/_main/import-reference.less
   - Expected: packages/test-data/css/_main/import-reference.css
   - Run: LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"

2. Compare with JavaScript implementation:
   - packages/less/src/less/import-visitor.js (how reference option works)
   - packages/less/src/less/tree/ruleset.js (check reference flag before output)

3. Fix in Go (likely files):
   - packages/less/src/less/less_go/import.go (Reference field)
   - packages/less/src/less/less_go/ruleset.go (check reference flag in GenCSS)
   - packages/less/src/less/less_go/import_visitor.go (propagate reference)

4. Test incrementally:
   - pnpm -w test:go:filter -- "import-reference"
   - pnpm -w test:go:unit (verify NO unit test regressions)
   - pnpm -w test:go (check overall counts stay >= baseline)

5. Commit with clear message about preserving and checking reference flag

REGRESSION CHECK before committing:
- Unit tests: pnpm -w test:go:unit (must still be 2,304 passing)
- Perfect matches: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS"
  (must stay >= 84, ideally 86 after this fix)
- If ANY regression detected, stop and investigate before committing
```

---

## Prompt 2: Fix Detached Ruleset Media Query Handling

**Time**: 2-3 hours | **Impact**: +1 perfect match | **Difficulty**: Medium-High

```
You are working on the less.go port of less.js. This task focuses on fixing how
detached rulesets handle media queries. The test detached-rulesets currently fails
because media queries aren't being merged correctly when detached rulesets are used.

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (baseline: 2,304 tests)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect)
- Ensure NO regressions in these numbers before starting

Steps:
1. Examine the failing test:
   - Input: packages/test-data/less/_main/detached-rulesets.less
   - Expected: packages/test-data/css/_main/detached-rulesets.css
   - Run: LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"
   - Understand what media query merging should happen

2. Compare with JavaScript:
   - packages/less/src/less/tree/detached-ruleset.js
   - packages/less/src/less/tree/media.js
   - Look for how media queries wrap detached ruleset outputs

3. Check Go implementation:
   - packages/less/src/less/less_go/detached_ruleset.go
   - packages/less/src/less/less_go/media.go
   - Look for missing media wrapping logic

4. Fix the media query merging logic

5. Test incrementally:
   - pnpm -w test:go:filter -- "detached-rulesets"
   - pnpm -w test:go:unit (check for unit regressions)
   - pnpm -w test:go (verify overall score)

6. Commit with clear message about media query wrapping

REGRESSION CHECK before committing:
- pnpm -w test:go:unit (must still be 2,304)
- LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS" (must be >= 85)
```

---

## Prompt 3: Fix URL Handling in Static and Dynamic Contexts

**Time**: 2-3 hours | **Impact**: +2 perfect matches | **Difficulty**: Medium

```
You are working on the less.go port of less.js. This task fixes URL handling in
two contexts: standard URL rewriting (urls in main) and static URL processing
(urls in static-urls).

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (baseline: 2,304 tests)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect)
- NO regressions allowed - verify before starting

Steps:
1. Understand the failing tests:
   - main/urls: packages/test-data/less/_main/urls.less
   - static-urls/urls: packages/test-data/less/static-urls/urls.less
   - Run both with diffs: LESS_GO_DIFF=1 pnpm -w test:go:filter -- "urls"
   - Compare actual vs expected CSS

2. Identify URL handling differences:
   - Data URLs vs file URLs vs relative URLs
   - Encoding/escaping differences
   - Path resolution context

3. Compare with JavaScript:
   - packages/less/src/less/tree/url.go
   - packages/less/src/less/less_go/url.go (Go implementation)
   - Check quote handling, escaping, path processing

4. Fix URL processing logic:
   - May need to adjust quote handling
   - May need to fix data URL handling
   - May need to adjust relative path handling

5. Test incrementally:
   - pnpm -w test:go:filter -- "urls"
   - pnpm -w test:go:unit (check regressions)
   - pnpm -w test:go (verify overall impact)

6. Commit explaining URL handling improvements

REGRESSION CHECK before committing:
- pnpm -w test:go:unit (2,304 expected)
- LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS" (should be >= 86)
```

---

## Prompt 4: Fix Media Query Output Formatting

**Time**: 2 hours | **Impact**: +1 perfect match | **Difficulty**: Medium

```
You are working on the less.go port of less.js. The media test fails because
media query CSS output formatting doesn't match less.js exactly. This is likely
about selector grouping, whitespace, or rule ordering.

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (baseline: 2,304)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect)
- Verify NO regressions before starting work

Steps:
1. Examine the failing test:
   - Input: packages/test-data/less/_main/media.less
   - Expected: packages/test-data/css/_main/media.css
   - Run: LESS_GO_DIFF=1 pnpm -w test:go:filter -- "TestIntegrationSuite/main/media"
   - Look at the CSS differences carefully

2. Identify what's different:
   - Selector formatting?
   - Whitespace/indentation?
   - Rule ordering?
   - Media query nesting structure?

3. Compare with JavaScript implementation:
   - packages/less/src/less/tree/media.js (JavaScript Media class)
   - packages/less/src/less/less_go/media.go (Go implementation)
   - Look at the GenCSS() method implementations

4. Fix the formatting logic in media.go

5. Test:
   - pnpm -w test:go:filter -- "media"
   - pnpm -w test:go:unit (check for regressions)
   - pnpm -w test:go (verify overall)

6. Commit with clear formatting fix message

REGRESSION CHECK before committing:
- pnpm -w test:go:unit (must be 2,304)
- LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS" (must be >= 85)
```

---

## Prompt 5: Fix Container Query Handling

**Time**: 2-3 hours | **Impact**: +1 perfect match | **Difficulty**: Medium

```
You are working on the less.go port of less.js. The container test fails because
@container queries (CSS container queries) aren't being processed correctly.

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (baseline: 2,304)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect)
- NO regressions allowed

Steps:
1. Look at the failing test:
   - Input: packages/test-data/less/_main/container.less
   - Expected: packages/test-data/css/_main/container.css
   - Run: LESS_GO_DIFF=1 pnpm -w test:go:filter -- "container"
   - Understand what's being tested

2. Check if @container is being parsed and handled:
   - packages/less/src/less/less_go/parser.go (should parse @container)
   - packages/less/src/less/less_go/at_rule.go (check AtRule handling)
   - Look for container query support

3. Compare with JavaScript:
   - packages/less/src/less/tree/at-rule.js
   - packages/less/src/less/parser.js (search for container)

4. Fix container query support (likely needs):
   - Proper parsing of @container queries
   - Correct CSS generation
   - Proper nesting of selectors within containers

5. Test:
   - pnpm -w test:go:filter -- "container"
   - pnpm -w test:go:unit (check for regressions)
   - pnpm -w test:go (verify overall)

6. Commit explaining container query support

REGRESSION CHECK before committing:
- pnpm -w test:go:unit (2,304 expected)
- LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS" (must be >= 85)
```

---

## Prompt 6: Fix Directive Bubbling and Selector Order

**Time**: 2-3 hours | **Impact**: +1 perfect match | **Difficulty**: Medium-High

```
You are working on the less.go port of less.js. The directives-bubling test fails
because directives like @media and @supports aren't being bubbled up correctly in
the CSS output, or selectors are in the wrong order.

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (baseline: 2,304)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect)
- Verify NO regressions before starting

Steps:
1. Examine the test:
   - Input: packages/test-data/less/_main/directives-bubling.less
   - Expected: packages/test-data/css/_main/directives-bubling.css
   - Run: LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"
   - Understand what the directive bubbling should do

2. Understand directive bubbling:
   - Directives like @media, @supports should bubble up to wrap output
   - Selectors should be grouped correctly by their directives
   - Order matters: outer directives wrap inner content

3. Compare with JavaScript:
   - packages/less/src/less/tree/ruleset.js (look for bubbling logic)
   - packages/less/src/less/tree/directive.js (if exists)
   - Check how directives are collected and output

4. Check Go implementation:
   - packages/less/src/less/less_go/ruleset.go (GenCSS method)
   - packages/less/src/less/less_go/at_rule.go
   - Look for missing or incorrect bubbling logic

5. Fix the directive bubbling and selector ordering

6. Test:
   - pnpm -w test:go:filter -- "directives-bubling"
   - pnpm -w test:go:unit (check for regressions)
   - pnpm -w test:go (verify overall)

7. Commit explaining directive bubbling fix

REGRESSION CHECK before committing:
- pnpm -w test:go:unit (2,304 expected)
- LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS" (must be >= 85)
```

---

## Prompt 7: Investigate Output Difference Root Causes (Analysis Only)

**Time**: 1-2 hours | **Impact**: Knowledge/planning | **Difficulty**: Low

```
You are working on the less.go port of less.js. This is an analysis and
investigation task (not a fix task). Your goal is to understand the root causes
of the 8 remaining output differences and create a prioritized plan for fixing them.

CRITICAL VALIDATION: Just running tests, no changes made
- Run: pnpm -w test:go:unit (baseline: 2,304)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 84 perfect)

Steps:
1. Run all 8 failing tests with diffs:
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference|detached-rulesets|urls|media|container|directives-bubling"

2. Categorize the differences:
   - Are they formatting differences? (whitespace, line breaks)
   - Are they structural differences? (missing selectors, wrong nesting)
   - Are they logical differences? (wrong evaluation, missing features)
   - Are they flag/option differences? (reference not checked, options not used)

3. For EACH test, create a summary documenting:
   - What's different between expected and actual
   - Pattern of the difference
   - Likely root cause area (import handling, media, url, selector, etc)
   - Estimated fix complexity (simple flag, medium logic, complex restructuring)

4. Create a file: .claude/tasks/output-differences/ANALYSIS_2025-11-26.md
   with findings from all 8 tests

5. Suggest priority order for fixes based on:
   - Quick wins (simple fixes) vs complex ones
   - Related tests that could be fixed together
   - Dependencies between fixes

This analysis will help the team prioritize which agents should work on which tests.

NO REGRESSION CHECK needed (analysis only, no code changes):
- No code modifications
- Just document findings
```

---

## Prompt 8: Fix Error Handling - javascript-undefined-var

**Time**: 1-2 hours | **Impact**: +1 error test fix | **Difficulty**: Low

```
You are working on the less.go port of less.js. The javascript-undefined-var test
should throw an error (because JavaScript execution is not supported), but currently
compiles successfully.

CRITICAL VALIDATION REQUIREMENT before starting work:
- Run: pnpm -w test:go:unit (baseline: 2,304)
- Run: LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30 (baseline: 88 error tests)
- Verify NO regressions before starting

Steps:
1. Look at the test:
   - File: packages/test-data/errors/eval/javascript-undefined-var.less
   - This test tries to use JavaScript execution
   - It should fail because JavaScript is a quarantined feature

2. Understand the test:
   - What JavaScript is being used?
   - Why should it fail in Go?

3. Compare with JavaScript version:
   - How does less.js handle this?
   - Should it throw an error?

4. Fix in Go:
   - packages/less/src/less/less_go/call.go (function calls)
   - packages/less/src/less/less_go/runtime.go (if exists)
   - Add validation to detect JavaScript execution and throw appropriate error

5. Test:
   - pnpm -w test:go:filter -- "javascript-undefined-var"
   - Should now show "Correctly failed" instead of succeeding
   - pnpm -w test:go:unit (check for regressions)

6. Commit with message about adding JavaScript feature detection

REGRESSION CHECK before committing:
- pnpm -w test:go:unit (2,304 expected)
- LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep -A 5 "EXPECTED ERROR"
  (should show this test in correctly failed list)
```

---

## Prompt 9: Performance Analysis and Optimization Opportunities

**Time**: 1-2 hours | **Impact**: Knowledge/planning | **Difficulty**: Low

```
You are working on the less.go port of less.js. This is a profiling and analysis
task to identify performance optimization opportunities.

CRITICAL NOTE: This is analysis only - NO code changes. Just profiling.

Steps:
1. Run the benchmark suite:
   pnpm bench:go:suite

   This will compare Go vs JavaScript performance on the test suite.

2. Check current performance stats:
   - How much slower is Go than JavaScript?
   - Where are the allocations happening?
   - Which test files are slowest?

3. Look at profiling results in:
   - .claude/benchmarks/PERFORMANCE_ANALYSIS.md (existing analysis)
   - Run: pnpm bench:profile (to get fresh profiles)

4. Document:
   - Top 5 performance bottlenecks
   - Potential quick wins for optimization
   - Long-term optimization strategy

5. Create file: .claude/tasks/performance/OPTIMIZATION_OPPORTUNITIES_2025-11-26.md
   with findings and priority recommendations

6. Note: Performance optimization is lower priority than correctness, but this
   analysis helps prioritize future work

This helps the team understand where optimization effort would have highest impact.

NO REGRESSION CHECK needed (analysis only):
- No code changes
- Just documentation of findings
```

---

## Prompt 10: Clean Up and Archive Completed Task Documentation

**Time**: 1 hour | **Impact**: Project organization | **Difficulty**: Low

```
You are working on the less.go port of less.js. This is a documentation cleanup
task to remove completed and outdated task files.

Steps:
1. Review what's in .claude/tasks/archived/ and .claude/tasks/

2. Files that should be ARCHIVED (move to archived/):
   - Tasks for issues we've already fixed
   - Outdated investigation files
   - Prompts from past sessions that are no longer relevant

3. Files that should be UPDATED:
   - .claude/CLAUDE.md (update test status to 84 perfect matches, 8 output diffs)
   - .claude/strategy/MASTER_PLAN.md (update current status section)
   - .claude/AGENT_WORK_QUEUE.md (update with current priorities)

4. Create new task files for the 8 remaining output differences:
   Move them from implicit list to explicit task files in .claude/tasks/
   - .claude/tasks/output-differences/import-reference.md (if not exists)
   - .claude/tasks/output-differences/detached-rulesets.md (if not exists)
   - .claude/tasks/output-differences/urls.md (if not exists)
   - .claude/tasks/output-differences/media.md (if not exists)
   - .claude/tasks/output-differences/container.md (if not exists)
   - .claude/tasks/output-differences/directives-bubling.md (if not exists)

5. Remove or archive:
   - Any status reports older than 1 week
   - Any outdated assessment reports
   - Prompts files from completed issues

6. Create NEW: .claude/STATUS_REPORT_2025-11-26.md
   with current status: 84/184 perfect, 8 output diffs, 88 error tests passing

7. Commit with message about documentation cleanup and reorganization

NO REGRESSION CHECK needed (doc-only changes):
- No code changes
- Just organizing task files for clarity
```

---

## How to Use These Prompts

### For Users
1. Pick any prompt above (1-6 for feature fixes, 7-10 for analysis/cleanup)
2. Copy the task prompt verbatim
3. Launch a new agent with the prompt
4. Agent will work independently and submit PR when done

### For Agents
1. **Start Fresh**: Clone repo, fetch latest from origin
2. **Create Branch**: `claude/fix-{task-name}-{session-id}`
3. **Run Baseline Tests FIRST**:
   ```bash
   pnpm -w test:go:unit          # Baseline: 2,304 tests
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # Baseline: 84 perfect
   ```
4. **Make changes** to fix the specific issue
5. **Test incrementally** as you work
6. **Run final validation**:
   ```bash
   pnpm -w test:go:unit          # Must still be 2,304
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -30  # Must be >= baseline
   ```
7. **Commit and push** with clear message
8. **Create PR** with summary of changes

### Success Criteria
- ✅ Specific test(s) fixed (tests show "Perfect match!")
- ✅ Unit tests still pass (2,304/2,304)
- ✅ No regressions (perfect match count >= 84)
- ✅ Clear commit message
- ✅ PR submitted

---

## Estimated Impact
If all 10 tasks completed:
- **8 fixes** (Prompts 1-6): 84 → 86 perfect matches (46.7%)
- **1 analysis** (Prompt 7): Knowledge for future work
- **1 error fix** (Prompt 8): 88 → 89 error tests
- **1 optimization analysis** (Prompt 9): Knowledge for future optimization
- **1 cleanup** (Prompt 10): Better organized project

**New Success Rate**: 95/184 tests (51.6% perfect matches, ~95% overall)

---

Generated: 2025-11-26
Baseline: 84 perfect matches, 88 error tests, 2,304 unit tests
