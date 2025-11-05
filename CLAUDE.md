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

4. **Current Integration Test Status** (as of 2025-11-05, latest audit):
   - **14 perfect CSS matches** - SEE REGRESSION WARNING BELOW ⚠️
   - **11 compilation failures** - REGRESSIONS DETECTED ⚠️
   - **~100+ tests with output differences** - compiles but CSS doesn't match
   - **~35 correct error handling** - tests that should fail, do fail correctly
   - **5 tests quarantined** (plugin system & JavaScript execution - punted for later)

   **⚠️ CRITICAL: REGRESSIONS DETECTED**
   Recent work introduced regressions that must be fixed before new feature work:
   - ❌ namespacing-6: Perfect match → Compilation failed (REGRESSION)
   - ❌ extend-clearfix: Perfect match → Output differs (REGRESSION)
   - ❌ namespacing-functions: Worse (now fails to compile)
   - ❌ import-reference: Worse (now fails to compile)
   - ❌ import-reference-issues: Worse (now fails to compile)
   - ✅ charsets: Improved to perfect match (+1)
   - **Net result: REGRESSION** (see `.claude/tracking/TEST_AUDIT_2025-11-05.md`)

   **🎉 Parser Status: ALL BUGS FIXED!**
   - Parser correctly handles full LESS syntax
   - Remaining work is in runtime evaluation and functional implementation

   **Recent Progress** (Runtime Fixes):
   - ✅ Issue #1: `if()` function context passing - FIXED
   - ✅ Issue #1b: Type function wrapping (unit, iscolor, etc.) - FIXED
   - ✅ Issue #2: Detached ruleset variable calls and frame scoping - FIXED
   - ✅ Issue #2b: `functions-each` context propagation and variable scope - FIXED
   - ✅ Issue #4: Parenthesized expression evaluation in function arguments - FIXED
   - ✅ Issue #5: `mixins-named-args` @arguments population for named arguments - FIXED
   - ✅ Issue #6: `mixins-closure`, `mixins-interpolated` - Mixin closure frame capture - FIXED
   - ✅ Issue #7: `mixins` - Mixin recursion detection for wrapped rulesets - FIXED
   - ⚠️ Issue #8: `namespacing-6` - Was fixed but now REGRESSED (needs re-fix)

5. **Organized Task System**:
   All project coordination and task management is now organized in the `.claude/` directory:

   - **`.claude/PROMPT_FIX_REGRESSIONS.md`** - URGENT: Fix critical regressions first
   - **`.claude/PROMPT_CREATE_AGENT_TASKS.md`** - Create task specs for agent distribution
   - **`.claude/README_AGENT_PROMPTS.md`** - Guide to using agent prompts
   - **`.claude/AGENT_WORK_QUEUE.md`** - Ready-to-assign work for parallel agents
   - **`.claude/VALIDATION_REQUIREMENTS.md`** - Test requirements for all PRs
   - **`.claude/strategy/agent-workflow.md`** - Step-by-step workflow for working on tasks
   - **`.claude/strategy/MASTER_PLAN.md`** - Overall strategy and current status
   - **`.claude/tasks/runtime-failures/`** - High-priority failing tests
   - **`.claude/tasks/output-differences/`** - Tests that compile but produce wrong CSS
   - **`.claude/tracking/assignments.json`** - Track which tasks are available/in-progress/completed
   - **`.claude/tracking/TEST_AUDIT_2025-11-05.md`** - Latest test status audit

   **If you're fixing regressions**: Start with `.claude/PROMPT_FIX_REGRESSIONS.md`

   **If you're creating new tasks**: Start with `.claude/PROMPT_CREATE_AGENT_TASKS.md`

   **If you're a new agent**: Check `.claude/AGENT_WORK_QUEUE.md` for ready-to-assign tasks

6. **Current Focus: CRITICAL REGRESSIONS MUST BE FIXED FIRST**:
   - **Runtime tracing available**: Use `LESS_GO_TRACE=1` to debug evaluation flow
   - Compare with JavaScript implementation when fixing issues
   - See `.claude/tracking/TEST_AUDIT_2025-11-05.md` for full regression analysis

   **Priority Order** (High to Low):
   1. **CRITICAL**: Fix namespacing-6 regression - `.claude/PROMPT_FIX_REGRESSIONS.md`
   2. **CRITICAL**: Fix extend-clearfix regression
   3. **CRITICAL**: Fix import-reference compilation failures (2 tests)
   4. **CRITICAL**: Fix namespacing-functions compilation failure
   5. **HIGH**: Mixin args with division (3 tests) - See existing tasks
   6. **HIGH**: Namespace value lookups (10 tests) - See existing tasks
   7. **HIGH**: Include path option (1 test) - See existing tasks
   8. Other runtime & output issues - See `.claude/tracking/assignments.json`

   ⚠️ **Before starting ANY new tasks**: Check `.claude/tracking/TEST_AUDIT_2025-11-05.md` for current status

7. **Quarantined Features** (for future implementation):
   - Plugin system tests (`plugin`, `plugin-module`, `plugin-preeval`)
   - JavaScript execution tests (`javascript`, `js-type-errors/*`, `no-js-errors/*`)
   - Import test that depends on plugins (`import`)
   - These are marked in `integration_suite_test.go` and excluded from test counts

Please review the imported rules above for detailed guidelines specific to the task at hand.