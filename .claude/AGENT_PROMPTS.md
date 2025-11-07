# 10 Ready-to-Assign Agent Prompts for less.go

These are independent, parallelizable tasks for fixing the next most important issues in the less.go port. Each prompt is designed to be copy-pasted to start a new agent working on a specific issue.

---

## Prompt 1: Fix Math Operations in Parens Mode

**Priority**: HIGH | **Estimated Time**: 3-4 hours | **Tests Affected**: 4+

Fix math operations to respect the `parens` mode where operations should only evaluate when inside parentheses, except division which always evaluates.

**Failing Tests**:
- `math-parens/css`
- `math-parens/media-math`
- `math-parens/parens`
- `math-parens/mixins-args` (already compiling, needs output fix)

**Current Issue**:
Operations are evaluating when they shouldn't. In `parens` mode, `1 + 1` should output as `1 + 1`, but `(1 + 1)` should evaluate to `2`. Division is special: `1 / 2` should always evaluate to `0.5`.

**Investigation Start**:
```bash
# See the issue
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/math-parens/parens"

# Compare with JavaScript
cd packages/less && npx lessc --math=parens test-data/less/math/parens/parens.less -
```

**Key Files**: `operation.go`, `paren.go`, `contexts.go`

**Success**: All 4 tests show "Perfect match!"

---

## Prompt 2: Fix Math Operations in Parens-Division Mode

**Priority**: HIGH | **Estimated Time**: 2-3 hours | **Tests Affected**: 3

Fix math operations for `parens-division` mode (LESS 4.x default) where ALL operations including division are literal unless in parentheses.

**Failing Tests**:
- `math-parens-division/media-math`
- `math-parens-division/parens`
- `math-parens-division/mixins-args`

**Current Issue**:
Division is evaluating when it shouldn't. In this mode, `1 / 2` should output as `1 / 2` (for CSS font shorthand), and only `(1 / 2)` should evaluate to `0.5`.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/math-parens-division/parens"
```

**Key Files**: `operation.go`, `paren.go`, `contexts.go`

**Success**: All 3 tests show "Perfect match!"

---

## Prompt 3: Fix Extend Chaining

**Priority**: HIGH | **Estimated Time**: 2-3 hours | **Tests Affected**: 1

Fix extend chaining where one extend references another extend (A extends B, B extends C, so A should also extend C).

**Failing Tests**:
- `extend-chaining`

**Current Issue**:
Multi-level extends aren't being resolved. When `.a:extend(.b)` and `.b:extend(.c)`, the final CSS should show `.a, .b, .c` in selectors with class `.c`.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/extend-chaining"
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/main/extend-chaining"
```

**Key Files**: `extend_visitor.go`, `extend.go`, `selector.go`

**Success**: Test shows "Perfect match!"

---

## Prompt 4: Fix Extend in Media Queries

**Priority**: HIGH | **Estimated Time**: 3-4 hours | **Tests Affected**: 1

Fix extend functionality when extends cross media query boundaries or happen inside media queries.

**Failing Tests**:
- `extend-media`

**Current Issue**:
Extends inside media queries produce incorrect nesting. The test shows `@media (tv) { @media (hires) { ... } }` instead of `@media (tv) and (hires) { ... }`.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/extend-media"
```

**Key Files**: `extend_visitor.go`, `media.go`, `selector.go`

**Success**: Test shows "Perfect match!"

---

## Prompt 5: Fix Import Interpolation

**Priority**: HIGH | **Estimated Time**: 2-3 hours | **Tests Affected**: 1

Fix variable interpolation in import paths so `@import "import-@{var}.less"` resolves the variable before importing.

**Failing Tests**:
- `import-interpolation` (compilation failure)

**Current Issue**:
The import path isn't being interpolated, so it tries to open a file literally named `import-@{in}@{terpolation}.less` instead of evaluating the variables first.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/import-interpolation"

# See what file it's trying to open
cat packages/test-data/less/_main/import-interpolation.less
```

**Key Files**: `import.go`, `import_visitor.go`, variable interpolation in file paths

**Success**: Test compiles and shows "Perfect match!"

---

## Prompt 6: Fix Import Module Resolution

**Priority**: HIGH | **Estimated Time**: 3-4 hours | **Tests Affected**: 1

Fix import resolution to support node_modules-style package imports like `@import "@less/test-import-module/one/1.less"`.

**Failing Tests**:
- `import-module` (compilation failure)

**Current Issue**:
Imports starting with `@` package names aren't being resolved from node_modules or configured module paths.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/import-module"

# Check if test setup includes module directory
ls packages/test-data/less/
```

**Key Files**: `import.go`, `file_manager.go`, module resolution logic

**Success**: Test compiles and shows "Perfect match!"

---

## Prompt 7: Fix Namespacing-3 (Media Query Variables)

**Priority**: MEDIUM | **Estimated Time**: 1-2 hours | **Tests Affected**: 1

Fix variable interpolation in media queries when accessed via namespaces, and ensure media query formatting is correct.

**Failing Tests**:
- `namespacing-3`

**Current Issue**:
Output shows `@media (min-width: 320px) { ... }.cell { ... }` instead of proper newline between media query and next rule.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/namespacing/namespacing-3"
```

**Key Files**: `media.go`, `namespace_value.go`, CSS generation

**Success**: Test shows "Perfect match!"

---

## Prompt 8: Fix Namespacing-8 (CSS Variable Interpolation)

**Priority**: MEDIUM | **Estimated Time**: 2-3 hours | **Tests Affected**: 1

Fix CSS custom property (variable) name interpolation when using namespace lookups.

**Failing Tests**:
- `namespacing-8`

**Current Issue**:
Expected `--color: #fff` but actual shows `--color: contrast(, #000, #fff)`. The `contrast()` function isn't being evaluated in the namespace lookup context.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/namespacing/namespacing-8"
```

**Key Files**: `namespace_value.go`, `functions.go` (contrast function), variable evaluation

**Success**: Test shows "Perfect match!"

---

## Prompt 9: Fix Mixins Nested Output

**Priority**: MEDIUM | **Estimated Time**: 2-3 hours | **Tests Affected**: 1

Fix nested mixin calls so they don't create extra empty rulesets or unevaluated expressions in output.

**Failing Tests**:
- `mixins-nested`

**Current Issue**:
Output includes `.inner { height:  * 10; }` instead of proper evaluated value, and empty rulesets appear.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/mixins-nested"
```

**Key Files**: `mixin_definition.go`, `mixin_call.go`, selector nesting

**Success**: Test shows "Perfect match!"

---

## Prompt 10: Fix Mixins Important Flag

**Priority**: MEDIUM | **Estimated Time**: 1-2 hours | **Tests Affected**: 1

Fix the `!important` flag on mixin calls so it propagates to all properties in the mixin output.

**Failing Tests**:
- `mixins-important`

**Current Issue**:
When calling `.mixin() !important;`, the `!important` flag isn't being added to the output properties.

**Investigation Start**:
```bash
cd packages/less/src/less/less_go && go test -v -run "TestIntegrationSuite/main/mixins-important"
```

**Key Files**: `mixin_call.go`, `declaration.go`, important flag handling

**Success**: Test shows "Perfect match!"

---

## How to Use These Prompts

1. Pick a prompt (ideally start with HIGH priority ones)
2. Create a new branch: `git checkout -b claude/fix-{task-name}-{session-id}`
3. Work on the task following the investigation steps
4. Run tests: `pnpm -w test:go:unit && pnpm -w test:go:filter -- "{test-name}"`
5. Commit with clear message explaining the fix
6. Push: `git push -u origin claude/fix-{task-name}-{session-id}`
7. Verify overall impact: `pnpm -w test:go` to check no regressions

## Validation Requirements

Before creating any PR:
- ✅ ALL unit tests pass: `pnpm -w test:go:unit`
- ✅ Target test(s) show improvement or "Perfect match!"
- ✅ No regressions in other passing tests
- ✅ Code follows Go idioms and matches JavaScript behavior
