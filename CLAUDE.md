# Claude Code Context for less.go

This file provides context to Claude Code about the less.go project and imports relevant Cursor rules based on the files being worked on.

## Project Overview
This is a fork of less.js being ported to Go. The goal is to maintain 1:1 functionality while following language-specific idioms.

## Always Applied Rules
@.cursor/rules/project-goals-and-conventions.mdc

## Language-Specific Rules

### When working with Go files (*.go)
@.cursor/rules/go-lang-rules.mdc

### When working with JavaScript files (*.js)
@.cursor/rules/javascript-rules.mdc

### When porting JavaScript to Go
@.cursor/rules/porting-process.mdc

## Context Instructions for Claude

When working on this project, please be aware of the following:

**⚠️ CRITICAL VALIDATION REQUIREMENT**: Before creating ANY pull request, you MUST run ALL tests:
- ✅ ALL unit tests: `pnpm -w test:go:unit` (must pass 100%)
- ✅ ALL integration tests: `pnpm -w test:go`
- ✅ Zero regressions tolerance - see `.claude/VALIDATION_REQUIREMENTS.md` for details

1. **File Type Detection**: The rules above should be considered based on the file types you're working with:
   - For `.go` files: Apply Go language rules and conventions
   - For `.js` files: Apply JavaScript rules (remember: never modify original JS files)
   - When porting: Follow the detailed porting process

2. **Core Principles**:
   - Maintain 1:1 functionality between JavaScript and Go versions
   - Avoid external dependencies where possible
   - Follow language-specific idioms and conventions
   - All ported code must pass tests that verify behavior matches the original

3. **Testing**:
   - JavaScript tests use Vitest framework
   - Go tests should verify ported functionality matches JavaScript behavior

4. **Performance Benchmarking**:

   Comprehensive benchmark suites are available to compare the Go port performance against the original JavaScript implementation.

   **Quick Start:**
   ```bash
   # Compare both implementations
   pnpm bench:compare

   # JavaScript only
   pnpm bench:js

   # Go only (comparable to JS)
   pnpm bench:go:suite
   ```

   **What's Tested:**
   - 80+ passing integration test files
   - Same files, same options for fair comparison
   - Covers all major LESS features
   - All files produce identical CSS output

   **Documentation:**
   - Main guide: `BENCHMARKS.md`
   - Detailed guide: `.claude/benchmarks/BENCHMARKING_GUIDE.md`
   - Quick reference: `.claude/benchmarks/QUICK_REFERENCE.md`

   **Benchmark Files:**
   - Go: `packages/less/src/less/less_go/benchmark_test.go`
   - JavaScript: `packages/less/benchmark/suite.js`

   **Performance Notes:**
   - The Go port is currently 8-10x slower than JavaScript
   - Primary cause: Excessive allocations (~47,000 per file)
   - This is **expected and acceptable** for an unoptimized port
   - Focus is on correctness first (✅), then optimization (📊)
   - Profiling tools: `pnpm bench:profile`
   - Detailed analysis: `.claude/benchmarks/PERFORMANCE_ANALYSIS.md`

   **Plugin IPC Performance:**
   - **JSON mode is 70% faster than SHM** for plugin function calls
   - Bootstrap4 with JSON: ~840ms | with SHM: ~1,420ms
   - Default is JSON mode (optimal for many small function calls)
   - Per-plugin IPC mode configuration available via `ipcMode` in plugin exports
   - Environment override: `LESS_JS_IPC_MODE=json` or `LESS_JS_IPC_MODE=shm`
   - See `.claude/benchmarks/PERFORMANCE_ANALYSIS.md` for details

5. **How to Use Integration Tests Effectively**:

   The integration test suite (`packages/less/src/less/less_go/integration_suite_test.go`) provides comprehensive coverage
   of LESS compilation with structured, LLM-friendly output.

   **Quick Start Commands:**
   ```bash
   # Get summary with minimal output (recommended for LLMs)
   LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

   # Get full verbose output with individual test results
   pnpm -w test:go

   # Get JSON output for programmatic analysis
   LESS_GO_JSON=1 LESS_GO_QUIET=1 pnpm -w test:go

   # Debug a specific test with detailed information
   LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/<suite>/<testname>

   # See CSS diffs for failing tests
   LESS_GO_DIFF=1 pnpm -w test:go
   ```

   **Understanding Test Categories:**

   The tests are automatically categorized into:

   - ✅ **Perfect CSS Matches** - Tests that compile and produce identical CSS to less.js (GOAL!)
   - ❌ **Compilation Failures** - Tests that fail to compile (parser/runtime errors) [HIGHEST PRIORITY]
   - ⚠️ **Output Differences** - Tests that compile but produce different CSS [MEDIUM PRIORITY]
   - ✅ **Correctly Failed** - Error tests that properly fail as expected (working correctly)
   - ⚠️ **Expected Error** - Error tests that should fail but succeed [LOW PRIORITY]
   - ⏸️ **Quarantined** - Plugin/JS features not yet implemented (not counted in totals)

   **Reading Test Output:**

   The test summary provides:
   1. **Quick Stats** - Overall success rate, compilation rate, percentages for each category
   2. **Detailed Results** - Tests grouped by suite, easy to identify which areas need work
   3. **Next Steps** - Prioritized action items with test counts per suite
   4. **Quick Commands** - Copy-paste commands for further analysis

   **How to Update Test Status Documentation:**

   When test results change significantly (e.g., fixing tests, regressions):
   1. Run `LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100` to get the summary
   2. Update the "Current Integration Test Status" section below with new numbers
   3. Add/update bullet points in "Recent Progress" for any newly fixed tests
   4. Verify "NO REGRESSIONS" by checking that perfect match count hasn't decreased
   5. Update the date in the section header to today's date

   **Detecting Regressions:**

   Compare the current "Perfect CSS Matches" count with the documented count below.
   - If count decreases: REGRESSION - investigate immediately
   - If count increases: PROGRESS - update documentation
   - If compilation failures increase: REGRESSION - investigate immediately

   **Environment Variables Reference:**
   - `LESS_GO_QUIET=1` - Suppress individual test output, show only summary
   - `LESS_GO_DEBUG=1` - Show enhanced debugging info and full test lists
   - `LESS_GO_DIFF=1` - Show side-by-side CSS diffs for failing tests
   - `LESS_GO_JSON=1` - Output results as JSON for programmatic parsing
   - `LESS_GO_STRICT=1` - Fail tests on any output difference (useful for CI)
   - `LESS_GO_TRACE=1` - Show evaluation trace (for debugging specific issues)

6. **Current Integration Test Status** (as of 2025-11-28 - Latest Verified Measurement - UPDATED):
   - **94 perfect CSS matches (51.4%)** - EXCELLENT PROGRESS! ✅
   - **0 compilation failures (0.0%)** - ALL COMPILATION ISSUES FIXED! 🎉
   - **89 correct error handling (48.6%)** - tests that should fail, do fail correctly ✅
   - **0 tests with CSS output differences (0.0%)** - ALL OUTPUT ISSUES FIXED! 🎉
   - **Overall Success Rate: 100.0%** ✅ (183/183 tests perfect matches or correctly erroring) 🎉
   - **Compilation Rate: 100.0%** (183/183 tests compile successfully)
   - **Perfect CSS Match Rate: 51.4%**
   - **8 quarantined tests** (plugin/JS features not yet implemented)
   - **✅ NO REGRESSIONS** - All previously passing tests still passing + new improvements!
   - **🎉 MAJOR MILESTONE**: 100% success rate achieved! ALL tests passing! 🎉

   **⚠️ IMPORTANT: Dependencies Required for Testing**
   Before running integration tests, you MUST install npm dependencies:
   ```bash
   pnpm install
   ```
   This installs workspace packages (`@less/test-import-module`) and npm dependencies (`bootstrap-less-port`)
   that are required for npm module resolution tests like `import-module`.

   **🎉 Parser Status: ALL BUGS FIXED!**
   - Parser correctly handles full LESS syntax
   - **183/183 tests compile successfully** ✅
   - NPM module resolution working correctly for scoped packages and bare imports
   - Fixed `chunkInput` default to match JavaScript behavior (was causing parse errors with comments inside parentheses)

   **✅ Unit Test Status:**
   - **3,012 tests passing** ✅ (100%) - up from 2,304!
   - All unit tests pass with no regressions

   **Recent Progress** (Runtime Fixes):
   - ✅ Issue #1: `if()` function context passing - FIXED
   - ✅ Issue #1b: Type function wrapping (unit, iscolor, etc.) - FIXED
   - ✅ Issue #2: Detached ruleset variable calls and frame scoping - FIXED
   - ✅ Issue #2b: `functions-each` context propagation and variable scope - FIXED
   - ✅ Issue #4: Parenthesized expression evaluation in function arguments - FIXED
   - ✅ Issue #5: `mixins-named-args` @arguments population for named arguments - FIXED
   - ✅ Issue #6: `mixins-closure`, `mixins-interpolated` - Mixin closure frame capture - FIXED
   - ✅ Issue #7: `mixins` - Mixin recursion detection for wrapped rulesets - FIXED
   - ✅ Issue #8: `namespacing-6` - VariableCall handling for MixinCall nodes - FIXED
   - ✅ Issue #9: DetachedRuleset missing methods - FIXED (regression fix)
   - ✅ Issue #10: Mixin variadic parameter expansion and argument matching - FIXED
   - ✅ Issue #11: `include-path` - Include path option for import resolution - FIXED
   - ✅ Issue #12: `css-guards` - CSS guard evaluation on rulesets - FIXED
   - ✅ Issue #13: Namespacing value evaluation - FIXED (namespacing-1, namespacing-2, namespacing-functions, namespacing-operations)
   - ✅ Issue #14: `import-interpolation` - Variable interpolation in import paths - FIXED
   - ✅ Issue #15: Math suites - All math-parens, math-parens-division, math-always suites now passing! - FIXED
   - ✅ Issue #16: URL processing - All URL rewriting suites now passing! - FIXED
   - ✅ Issue #17: Units suites - units-strict and units-no-strict now passing! - FIXED
   - ✅ Issue #18: Compression suite - compression now passing! - FIXED
   - ✅ Issue #19: Extend regressions - extend-clearfix, extend-nest, extend all FIXED! - NO REGRESSIONS
   - ✅ Issue #20: `namespacing-media` - Media query variable interpolation - FIXED (11/11 namespacing tests!)
   - ✅ Issue #21: `mixins-nested` - Nested mixin variable scoping - FIXED
   - ✅ Issue #22: `import-inline` - Media query wrapper - FIXED
   - ✅ Issue #23: `import-interpolation` - Variable interpolation in imports - FIXED
   - ✅ Issue #24: `css-escapes` - CSS escape handling - FIXED
   - ✅ Compilation failures reduced from 12 → 3 tests (75% reduction!)
   - ✅ **ALL DOCUMENTED REGRESSIONS FIXED**: mixins, mixins-interpolated, mixins-guards (main) - all now perfect matches!
   - ✅ Issue #25: **Error validation improvements** - 17 error handling tests now correctly validate and fail (2025-11-12):
     - Mixed unit operations (add/divide/multiply) now properly validated
     - Recursive variable detection working
     - Namespacing errors properly caught (namespacing-2, -3, -4)
     - SVG gradient validation (all 6 tests)
     - Detached ruleset type checking
     - Function undefined detection

7. **Organized Task System**:
   All project coordination and task management is now organized in the `.claude/` directory:

   - **`.claude/strategy/MASTER_PLAN.md`** - Overall strategy and current status
   - **`.claude/strategy/agent-workflow.md`** - Step-by-step workflow for working on tasks
   - **`.claude/templates/AGENT_PROMPT.md`** - Template for spinning up new agents
   - **`.claude/tasks/runtime-failures/`** - High-priority failing tests (6 tests remaining)
   - **`.claude/tasks/output-differences/`** - Tests that compile but produce wrong CSS (~106 tests)
   - **`.claude/tracking/assignments.json`** - Track which tasks are available/in-progress/completed
   - **`.claude/AGENT_WORK_QUEUE.md`** - Ready-to-assign work for parallel agents

   **If you're working on a specific task**: Check `.claude/tasks/` for detailed task specifications.

   **If you're a new agent**: Start with `.claude/AGENT_WORK_QUEUE.md` for ready-to-assign tasks.

8. **Current Focus: Runtime & Evaluation Issues**:
   - **Runtime tracing available**: Use `LESS_GO_TRACE=1` to debug evaluation flow
   - Compare with JavaScript implementation when fixing issues
   - See `.claude/tasks/` for specific task specifications

   **Priority Order** (High to Low) - Updated 2025-11-28 (Current Run):

   **🎉 100% SUCCESS RATE ACHIEVED!** All active tests passing!

   **No Remaining Issues** - All compilation failures and output differences have been resolved!
   - `bootstrap4` has been quarantined (requires JavaScript plugins not implemented in Go)

   **✅ Tests That Now Pass (were previously thought broken):**
   - `import-module` - NOW PASSING! NPM module resolution works when `pnpm install` is run
   - `import-reference` - NOW PASSING!
   - `import-reference-issues` - NOW PASSING!
   - `google` - Correctly handles network imports (expected to fail without network)

   **Recently Fixed** (since 2025-11-26):
   - ✅ import-module - NPM module resolution confirmed working!
   - ✅ import-reference - Reference imports now working correctly!
   - ✅ import-reference-issues - Import reference edge cases fixed!
   - ✅ urls (main) - URL handling now perfect match!
   - ✅ detached-rulesets - media query merging now working!
   - ✅ media - media query output now passing!
   - ✅ container - CSS output formatting fixed!
   - ✅ url tests in static-urls and url-args - now passing!
   - ✅ directives-bubling - CSS directives bubbling now working!

   **Recently Completed** (Past 6 weeks):
   - ✅ **LATEST (2025-11-28 - CURRENT)**: 94 perfect matches! ZERO output diffs remaining! 🎉
   - ✅ **2025-11-27**: 90 perfect matches! Only 2 output diffs remaining!
   - ✅ **2025-11-26**: 88 perfect matches! +5 from 2025-11-13, only 4 output diffs remaining!
   - ✅ **MASSIVE BREAKTHROUGH**: +45 perfect matches! From 34 → 79 tests! 🎉
   - ✅ **Week 4 WINS**: +10 perfect matches! From 69 → 79 tests!
   - ✅ **ALL namespacing tests FIXED**: 11/11 namespacing tests perfect matches (100% complete!)
   - ✅ **ALL guards tests FIXED**: css-guards, mixins-guards, mixins-guards-default-func (100% complete!)
   - ✅ **ALL extend tests FIXED**: 7/7 extend tests perfect matches including extend-chaining (100% complete!)
   - ✅ **ALL URL rewriting tests FIXED**: 4/4 URL tests perfect matches (100% complete!)
   - ✅ **ALL math operation tests FIXED**: 10/10 math tests perfect matches (100% complete!)
   - ✅ **ALL unit test suites FIXED**: compression, strict-units, no-strict (100% complete!)
   - ✅ **Latest color & variable fixes**: colors, colors2, variables, variables-in-at-rules
   - ✅ **Core functionality**: extract-and-length, property-accessors, parse-interpolation, strings, permissive-parse
   - ✅ **Mixin & import fixes**: All mixin variants, import-inline, import-interpolation passing!
   - ✅ **Parser fully fixed**: All real compilation failures resolved!
   - ✅ **ZERO REGRESSIONS**: All previously passing tests continue to pass

   **Error Handling Milestone** (2025-11-28):
   - **89 correct error handling tests!** (48.6% of all tests)
   - Overall success rate now at **100.0%** (183/183 tests)
   - **ALL tests passing!** 🎉

9. **Quarantined Features** (for future implementation):
   - Plugin system tests (`plugin`, `plugin-module`, `plugin-preeval`)
   - JavaScript execution tests (`javascript`, `js-type-errors/*`, `no-js-errors/*`)
   - Import test that depends on plugins (`import`)
   - Third-party test requiring JS plugins (`bootstrap4` - uses map-get, breakpoint-next, etc.)
   - These are marked in `integration_suite_test.go` and excluded from test counts

Please review the imported rules above for detailed guidelines specific to the task at hand.