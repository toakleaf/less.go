# Ready-to-Use Agent Tasks

This document contains copy-paste task descriptions for spawning independent agents.

---

## Wave 1: Independent Runtime Fixes (4 agents in parallel)

### Agent 1: Import Reference Issues
**Branch**: `claude/fix-imports-<session-id>`
**See**: `issues/ISSUE_IMPORTS.md`
**Tests**: 2-3 of 5 (defer import-interpolation and google)

```
Fix import reference functionality for less.go

CONTEXT: You're working on a Go port of less.js. The parser works (92.4% compilation rate), but import reference handling has bugs.

YOUR MISSION: Fix 2-3 of 5 import tests (defer import-interpolation as noted)

FAILING TESTS:
- import-reference: CSS imports not handled correctly
- import-reference-issues: Referenced mixins not accessible
- import-module: Module path resolution
- (DEFER) import-interpolation: Architectural issue documented
- (LOW PRIORITY) google: Remote import network issue

CONSTRAINTS:
- NEVER modify .js files
- All changes in packages/less/src/less/less_go/*.go
- Must pass: pnpm -w test:go:unit
- Must pass target tests: pnpm -w test:go:filter -- "import-reference"

INVESTIGATION STARTING POINTS:
- import.go: Reference option handling
- import_visitor.go: Import processing and visibility
- import_manager.go: File resolution and CSS import detection
- JS reference: packages/less/src/less/import-manager.js

KEY ISSUE: (reference) imports should make mixins accessible without outputting CSS

TEST COMMAND:
go test -run "TestIntegrationSuite/main/import-reference" -v
LESS_GO_DEBUG=1 go test -run "TestIntegrationSuite/main/import-reference" -v

SUCCESS CRITERIA: 2+ tests passing, no regressions

Read issues/ISSUE_IMPORTS.md for full details.
```

---

### Agent 2: Namespace Resolution
**Branch**: `claude/fix-namespacing-<session-id>`
**See**: `issues/ISSUE_NAMESPACING.md`
**Tests**: 2 of 2

```
Fix namespace variable call evaluation for less.go

CONTEXT: You're working on a Go port of less.js. Mixin calls assigned to variables and then called fail with "Could not evaluate variable call".

YOUR MISSION: Fix 2 namespace resolution tests

FAILING TESTS:
- namespacing-6: @alias: .something(foo); @alias(); fails
- namespacing-functions: Similar issue with function calls

ERROR: "Could not evaluate variable call @alias"

CONSTRAINTS:
- NEVER modify .js files
- All changes in packages/less/src/less/less_go/*.go
- Must pass: pnpm -w test:go:unit
- Must pass target tests: pnpm -w test:go:filter -- "namespacing-6"

INVESTIGATION STARTING POINTS:
- variable_call.go: Variable call evaluation (look for error message)
- variable.go: Variable evaluation and storage
- mixin_call.go: Mixin call return values
- detached_ruleset.go: DetachedRuleset callable interface
- JS reference: packages/less/src/less/tree/variable-call.js

KEY ISSUE: Similar to Issue #2 (fixed) - mixin call results not properly stored/called

DEBUG HINT: Use LESS_GO_TRACE=1 to see evaluation flow

TEST COMMAND:
LESS_GO_TRACE=1 go test -run "TestIntegrationSuite/namespacing/namespacing-6" -v

SUCCESS CRITERIA: 2/2 tests passing, no regressions

Read issues/ISSUE_NAMESPACING.md for full details.
```

---

### Agent 3: URL Parsing with Escapes
**Branch**: `claude/fix-urls-<session-id>`
**See**: `issues/ISSUE_URLS.md`
**Tests**: 2 of 2 (same test, 2 suites)

```
Fix URL parsing with escaped characters for less.go

CONTEXT: You're working on a Go port of less.js. URLs containing escaped parentheses cause parser errors.

YOUR MISSION: Fix URL parsing for 2 tests

FAILING TESTS:
- urls (main suite): Parser error on url(http://...family=\"Font\":\(400\),700)
- urls (compression suite): Same test

ERROR: "expected ')' got '('" - parser sees escaped \( as syntax

CONSTRAINTS:
- NEVER modify .js files
- All changes in packages/less/src/less/less_go/*.go
- Must pass: pnpm -w test:go:unit
- Must pass target tests: pnpm -w test:go:filter -- "urls"

INVESTIGATION STARTING POINTS:
- url.go: URL parsing logic
- parser.go: url() parsing, escape handling
- parser_input.go: Character escaping and tokenization
- JS reference: packages/less/src/less/parser/parser.js

KEY ISSUE: Backslash-escaped characters in URLs not handled correctly

MINIMAL TEST CASE:
cat > /tmp/test.less << 'EOF'
.test { background: url(http://example.com/css?family=\"Font\":\(400\),700); }
EOF
go run cmd/lessc/lessc.go /tmp/test.less

TEST COMMAND:
go test -run "TestIntegrationSuite/main/urls" -v

SUCCESS CRITERIA: 2/2 tests passing, no regressions

Read issues/ISSUE_URLS.md for full details.
```

---

### Agent 4: Include Path Resolution
**Branch**: `claude/fix-paths-<session-id>`
**See**: `issues/ISSUE_PATHS.md`
**Tests**: 1 of 1

```
Fix include path resolution for imports in less.go

CONTEXT: You're working on a Go port of less.js. Imports are not being resolved through configured include paths.

YOUR MISSION: Fix 1 include path test

FAILING TEST:
- include-path: @import "import-test-e"; not found

ERROR: "open import-test-e: no such file or directory"

CONSTRAINTS:
- NEVER modify .js files
- All changes in packages/less/src/less/less_go/*.go
- Must pass: pnpm -w test:go:unit
- Must pass target test: pnpm -w test:go:filter -- "include-path"

INVESTIGATION STARTING POINTS:
- integration_suite_test.go: Check if include path is configured for test
- import_manager.go: Import path resolution logic
- file_manager.go: File finding with include paths
- JS reference: packages/less/src/less/import-manager.js

KEY ISSUE: Either test config missing include path OR import manager not searching include paths

DEBUG STEPS:
1. Find where import-test-e.less actually is: find packages/test-data -name "*import-test-e*"
2. Check test config: grep -A5 "include-path" packages/less/src/less/less_go/integration_suite_test.go
3. Fix test config or import manager accordingly

TEST COMMAND:
go test -run "TestIntegrationSuite.*include-path" -v

SUCCESS CRITERIA: 1/1 test passing, no regressions

Read issues/ISSUE_PATHS.md for full details.
```

---

## Wave 2: Complex Fixes (2 agents, after Wave 1)

### Agent 5: Mixin Argument Expansion
**Branch**: `claude/fix-mixins-args-<session-id>`
**See**: `issues/ISSUE_MIXINS_ARGS.md`
**Tests**: 2 of 2

```
Fix mixin argument expansion with ... operator for less.go

CONTEXT: You're working on a Go port of less.js. The ... expansion operator for mixin arguments is not working.

YOUR MISSION: Fix 2 mixin argument expansion tests

FAILING TESTS:
- mixins-args (math-parens): .m3(@x...) where @x: 1, 2, 3; should expand to 3 arguments
- mixins-args (math-parens-division): Same test

ERROR: "No matching definition was found for '.m3()'" - expansion not happening

CONSTRAINTS:
- NEVER modify .js files
- All changes in packages/less/src/less/less_go/*.go
- Must pass: pnpm -w test:go:unit
- Test carefully - mixins are central feature

INVESTIGATION STARTING POINTS:
- mixin_call.go: Argument expansion logic
- mixin_definition.go: Parameter matching with expanded args
- expression.go or value.go: List expansion
- JS reference: packages/less/src/less/tree/mixin-call.js

KEY ISSUE: @x... should expand list (1, 2, 3) into three separate arguments for mixin matching

MINIMAL TEST CASE:
cat > /tmp/test.less << 'EOF'
.m3(@a, @b, @c) { result: @a, @b, @c; }
.test { @x: 1, 2, 3; .m3(@x...); }
EOF
go run cmd/lessc/lessc.go /tmp/test.less

TEST COMMAND:
go test -run "TestIntegrationSuite/math-parens/mixins-args" -v

SUCCESS CRITERIA: 2/2 tests passing, no regressions in other mixin tests

Read issues/ISSUE_MIXINS_ARGS.md for full details.
```

---

### Agent 6: Bootstrap 4 Investigation
**Branch**: `claude/investigate-bootstrap4-<session-id>`
**See**: `issues/ISSUE_BOOTSTRAP4.md`
**Tests**: 0-1 of 1 (investigation + possible fix)

```
Investigate and optionally fix bootstrap4 test for less.go

CONTEXT: You're working on a Go port of less.js. Bootstrap 4 is a large real-world LESS codebase that tests many features.

YOUR MISSION: Investigate why bootstrap4 test fails, fix if simple (max 1 hour)

FAILING TEST:
- bootstrap4: Large integration test

ERROR: "open bootstrap-less-port/less/bootstrap: no such file or directory"

TIME LIMIT: 1 hour maximum

CONSTRAINTS:
- NEVER modify .js files
- If it's a simple setup issue: fix it
- If it's complex bugs: document and defer

INVESTIGATION STEPS:
1. Find the test file: cat packages/test-data/less/3rd-party/bootstrap4.less
2. Check if bootstrap files exist: find packages/test-data -name "*bootstrap*"
3. Check git submodules: git submodule status
4. Check JavaScript test setup: grep -r "bootstrap4" packages/less/test/

POSSIBLE OUTCOMES:
A) Setup issue: Bootstrap files missing → Get them, configure paths
B) Single bug: Maps to existing issue → Note in that issue, defer
C) Multiple bugs: Document in ISSUE_BOOTSTRAP4_BUG.md, defer
D) Works after other fixes: Report success

TEST COMMAND:
go test -run "TestIntegrationSuite/3rd-party/bootstrap4" -v

SUCCESS CRITERIA:
- Either: Test passing
- Or: Clear documentation of what's needed
- Don't spend >1 hour on this

Read issues/ISSUE_BOOTSTRAP4.md for full details and decision tree.
```

---

## Phase 5: CSS Output Differences (Batched Agents)

**See**: `issues/ISSUE_OUTPUT_DIFFS.md` for full categorization and batching strategy.

Run after Wave 1 & 2 are complete and integrated.

### Batch 1: Core Features (26 tests)
**Branch**: `claude/fix-output-core-<session-id>`

```
Fix CSS output differences for core LESS features

MISSION: Fix 18+ of 26 core feature tests that compile but have wrong CSS output

TESTS: selectors, mixins-nested, mixins-important, operations, calc, merge,
       strings, variables, directives-bubling, media, functions, etc.

STRATEGY:
1. Use LESS_GO_DIFF=1 to see differences
2. Group tests by similar issues
3. Fix GenCSS methods in node files
4. Aim for 70%+ success rate

DEBUG COMMAND:
LESS_GO_DIFF=1 go test -run "TestIntegrationSuite/main/selectors" -v

Read issues/ISSUE_OUTPUT_DIFFS.md for full details.
```

### Subsequent Batches
See `ISSUE_OUTPUT_DIFFS.md` for:
- Batch 2: Extend (6 tests)
- Batch 3: CSS Standards (6 tests)
- Batch 4: Imports (3 tests)
- Batch 5: Namespacing (8 tests)
- Batch 6: Math Modes (9 tests)
- Batch 7: URL Rewriting (6 tests)
- Batch 8: Error Messages (20 tests)
- Batch 9: Misc (13 tests)

---

## Usage Instructions

### For Human Orchestrator

1. **Start Wave 1** (4 agents in parallel):
   ```
   - Copy Agent 1-4 task descriptions
   - Spawn 4 Claude Code agents simultaneously
   - Each gets its own branch and task
   ```

2. **Monitor Progress**:
   - Each agent commits to its branch
   - Check: Did tests pass? Any regressions?
   - Cost tracking: Wave 1 should be ~$100-150

3. **Integration**:
   - Cherry-pick successful fixes
   - Merge to development branch
   - Run full test suite

4. **Start Wave 2** (2 agents):
   - After Wave 1 integrated
   - Similar process

5. **Phase 5** (batched):
   - After all runtime issues fixed
   - Run batch agents one at a time or in small groups
   - Check progress after each batch

### For Agents

Each task description above is self-contained:
- ✅ Clear mission
- ✅ Specific tests
- ✅ Error messages
- ✅ Investigation starting points
- ✅ Test commands
- ✅ Success criteria
- ✅ Link to full issue doc

Just copy the task and go!

---

## Quick Reference

| Agent | Tests | Priority | Complexity | Independence |
|-------|-------|----------|------------|--------------|
| 1. Imports | 2-3/5 | High | High | HIGH |
| 2. Namespacing | 2/2 | High | Medium | HIGH |
| 3. URLs | 2/2 | Medium | Low-Med | HIGH |
| 4. Paths | 1/1 | Low | Low | HIGH |
| 5. Mixin Args | 2/2 | Medium | Medium | MEDIUM |
| 6. Bootstrap4 | 0-1/1 | Medium | High | MEDIUM |

**Wave 1 can run fully in parallel (Agents 1-4)**
**Wave 2 should wait for Wave 1 integration (Agents 5-6)**
**Phase 5 should wait for all runtime fixes (Batches 1-9)**
