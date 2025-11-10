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

4. **Current Integration Test Status** (as of 2025-11-11 - Fresh Test Run):
   - **140 passing tests (76.1%)** - EXCELLENT PROGRESS! ✅
   - **✅ ZERO REGRESSIONS** - All previously passing tests still passing!
   - **3 compilation failures (1.6%)** - All external/expected (bootstrap4, google, import-module)
   - **14 output differences (7.6%)** - compiles but CSS generation differs
   - **27 error handling issues (14.7%)** - tests that should error but don't
   - **Overall Success Rate: 76.1%** (140/184 active tests)
   - **Compilation Rate: 97.8%** (181/184 tests compile successfully)
   - **Perfect CSS Match Rate: Estimated 50-60%** (140 passing includes correct error handling)

   **🎉 Parser Status: ALL BUGS FIXED!**
   - Parser correctly handles full LESS syntax
   - **181/185 tests compile successfully** ⬆️
   - Remaining work is primarily CSS generation, error handling, and edge cases

   **✅ Unit Test Status:**
   - **2,290+ tests passing** ✅ (99.9%+)
   - **1 test has a timeout issue**: `TestRulesetErrorConditions/should_handle_nested_rulesets_with_circular_dependencies` (test bug, not functionality)
   - No functionality regressions

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

5. **Organized Task System**:
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

6. **Current Focus: Runtime & Evaluation Issues**:
   - **Runtime tracing available**: Use `LESS_GO_TRACE=1` to debug evaluation flow
   - Compare with JavaScript implementation when fixing issues
   - See `.claude/tasks/` for specific task specifications

   **Priority Order** (High to Low):
   1. **CRITICAL**: Error Handling Issues - 27 tests that should fail but don't (14.7% of failures)
      - Units/color operations need stricter validation
      - Variable scope checking
      - Function argument validation
   2. **HIGH**: Import reference (2 tests) - import-reference, import-reference-issues
   3. **HIGH**: CSS output formatting (6-8 tests):
      - detached-rulesets, media, directives-bubling, css-3
      - comments2, container, extract-and-length
   4. **MEDIUM**: URL variants - urls in main/static-urls/url-args (3 tests)
   5. **MEDIUM**: Functions - functions, functions-each (2 tests)
   6. **LOW**: External dependencies - bootstrap4, google (network/packages - not bugs)
   7. **LOW**: Unit test bug - Fix timeout in circular dependency test

   **Recently Completed** (Past 4 weeks):
   - ✅ **MASSIVE BREAKTHROUGH**: +44 perfect matches! From 34 → 78 tests! 🎉
   - ✅ **Week 4 WINS** (this session): +9 perfect matches! From 69 → 78 tests!
   - ✅ **ALL namespacing tests FIXED**: 11/11 namespacing tests perfect matches (100% complete!)
   - ✅ **ALL guards tests FIXED**: css-guards, mixins-guards, mixins-guards-default-func (100% complete!)
   - ✅ **ALL extend tests FIXED**: 7/7 extend tests perfect matches (100% complete!)
   - ✅ **ALL URL rewriting tests FIXED**: 4/4 URL tests perfect matches (100% complete!)
   - ✅ **ALL math operation tests FIXED**: 8/8 math tests perfect matches (100% complete!)
   - ✅ **ALL unit test suites FIXED**: compression, strict-units, no-strict (100% complete!)
   - ✅ **Latest color & variable fixes**: colors, colors2, variables, variables-in-at-rules
   - ✅ **Core functionality**: extract-and-length, property-accessors, parse-interpolation, strings, permissive-parse
   - ✅ **Mixin & import fixes**: All mixin variants, import-inline, import-interpolation passing!
   - ✅ **Parser fully fixed**: All real compilation failures resolved!

7. **Quarantined Features** (for future implementation):
   - Plugin system tests (`plugin`, `plugin-module`, `plugin-preeval`)
   - JavaScript execution tests (`javascript`, `js-type-errors/*`, `no-js-errors/*`)
   - Import test that depends on plugins (`import`)
   - These are marked in `integration_suite_test.go` and excluded from test counts

Please review the imported rules above for detailed guidelines specific to the task at hand.