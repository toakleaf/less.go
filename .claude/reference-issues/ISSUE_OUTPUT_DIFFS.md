# CSS Output Differences - Phase 5 Agent Tasks

## 🎯 Mission
Fix 102 tests that compile successfully but produce incorrect CSS output

## 📊 Status
- **Tests with Output Diffs**: 102
- **Priority**: Phase 5 (After runtime failures are fixed)
- **Complexity**: VARIES by category
- **Independence**: VARIES - Some categories independent, some related

## 📋 Overview

These tests **compile and evaluate without errors**, but the generated CSS doesn't match the expected output. This means:
- ✅ Parser is working
- ✅ AST is created correctly
- ✅ Evaluation completes
- ❌ CSS generation has bugs

The issues are in:
- Selector generation
- Property value output
- Media query handling
- Extend functionality
- Guard evaluation
- Import content insertion
- Compression/minification
- And other CSS generation logic

---

## 📊 Categorization of 102 Tests

### Category 1: Core Features (26 tests)
Main LESS features that are fundamental:

**Selectors & Mixins** (8 tests):
- `selectors` - Selector generation
- `mixins-nested` - Nested mixin output
- `mixins-important` - !important flag handling
- `mixins-named-args` - Named argument output
- `mixins-guards` - Guard evaluation (main)
- `mixins-guards-default-func` - Guards with default()
- `scope` - Variable scoping output
- `property-accessors` - Property accessor syntax

**Operations & Math** (5 tests):
- `operations` - Arithmetic operations
- `calc` - calc() function preservation
- `merge` - Property merging
- `strings` - String operations
- `variables` - Variable substitution

**Directives & At-Rules** (5 tests):
- `directives-bubling` - Media query bubbling
- `media` - Media query generation
- `variables-in-at-rules` - Variables in @rules
- `container` - Container queries
- `css-guards` - Guards on selectors

**Functions** (4 tests):
- `functions` - Function output
- `functions-each` - each() function
- `extract-and-length` - extract()/length()
- `colors` - Color functions

**Comments & Whitespace** (4 tests):
- `comments` - Comment preservation
- `comments2` - Complex comments
- `whitespace` - Whitespace handling
- `permissive-parse` - Permissive mode

---

### Category 2: Extend Functionality (6 tests)
Tests for the `extend` feature:

- `extend` - Basic extend
- `extend-chaining` - Extended extends
- `extend-clearfix` - Extend with pseudo-classes
- `extend-exact` - Exact matching
- `extend-media` - Extend in media queries
- `extend-nest` - Nested extend
- `extend-selector` - Complex selectors

**Common Issues**:
- Extend not applying
- Wrong selector generation
- Media query interaction
- Pseudo-class handling

---

### Category 3: CSS Standards & Escapes (6 tests)
Modern CSS features and escaping:

- `css-3` - CSS3 features
- `css-escapes` - Escape sequences
- `colors2` - Advanced color handling
- `parse-interpolation` - Interpolation in selectors/properties
- `property-name-interp` - Property name interpolation
- `detached-rulesets` - Detached ruleset output

**Common Issues**:
- Escape sequences incorrect
- Interpolation not working
- Modern CSS syntax

---

### Category 4: Import Handling (3 tests)
Tests that compile but imports aren't processed correctly:

- `import-inline` - Inline import content
- `import-once` - Import once behavior
- `import-remote` - Remote imports

**Common Issues**:
- Import content not inserted
- Multiple imports handling
- URL vs file imports

---

### Category 5: Namespacing (8 tests)
Namespace resolution tests (some pass, some differ):

- `namespacing-1` - Basic namespacing
- `namespacing-2` - Nested namespaces (ERROR)
- `namespacing-3` - Namespace operations (ERROR)
- `namespacing-4` - Complex namespaces (ERROR)
- `namespacing-5` - Namespace mixins
- `namespacing-7` - Advanced patterns
- `namespacing-8` - Edge cases
- `namespacing-media` - Namespaces in media
- `namespacing-operations` - Namespaces in operations

**Note**: Some of these might be related to the **ISSUE_NAMESPACING.md** runtime failures. Fix those first.

---

### Category 6: Math Modes (9 tests)
Math mode specific tests:

**Parens Mode** (3 tests):
- `css` - Math with parens mode
- `media-math` - Math in media queries
- `parens` - Parenthesized expressions

**Parens-Division Mode** (3 tests):
- `media-math` - Math in media (division mode)
- `new-division` - New division behavior
- `parens` - Parenthesized expressions (division)

**Always Mode** (1 test):
- `mixins-guards` - Guards with always math

**Compression** (2 tests):
- `compression` - Minification
- `urls` - URL handling in compression

**Note**: Once **ISSUE_URLS.md** is fixed, compression/urls should pass.

---

### Category 7: URL Rewriting (6 tests)
URL rewriting and path handling:

- `urls` - URL processing (FAILING - see ISSUE_URLS.md)
- `urls` (strict-units suite) - Same test
- `rewrite-urls-all` - Rewrite all URLs
- `rewrite-urls-local` - Rewrite local URLs only
- `rootpath-rewrite-urls-all` - With rootpath
- `rootpath-rewrite-urls-local` - With rootpath local
- `include-path-string` - Include path in URLs

**Note**: Core URL parsing is FAILING (see ISSUE_URLS.md). Fix that first.

---

### Category 8: Source Maps (Tests not listed individually)
Source map generation tests.

**Common Issues**:
- Source map not generated
- Wrong mappings
- File references incorrect

**Note**: Lower priority - functionality over debugging

---

### Category 9: Error Handling Tests (20 tests)
Tests that **expect errors** but might have wrong error messages or formatting:

These all correctly fail, but the error output might differ. Examples:
- `add-mixed-units` - "Cannot add 1px and 1em"
- `at-rules-undefined-var` - Undefined variable errors
- `color-func-invalid-color` - Invalid color errors
- `detached-ruleset-1`, `detached-ruleset-2` - Ruleset errors
- `divide-mixed-units` - Division errors
- `mixins-guards-default-func-*` - Guard errors
- `multiply-mixed-units` - Multiplication errors
- Various other error message tests

**Common Issues**:
- Error message format differs
- Line numbers differ
- Error context differs

**Priority**: LOW - These are "correct" failures, just different messages

---

### Category 10: Misc & Edge Cases (13 tests)
Various other tests:

- `property-in-root2` - Properties in root scope
- `property-interp-not-defined` - Undefined interpolation
- `javascript-undefined-var` - JS evaluation (QUARANTINED feature)
- `import-subfolder1` - Import from subdirectory
- `recursive-variable` - Recursive variable detection
- `root-func-undefined-1` - Undefined function
- `percentage-non-number-argument` - Type errors
- `svg-gradient1` through `svg-gradient6` - SVG gradients
- `unit-function` - unit() function edge cases
- `invalid-color-with-comment` - Error with comment
- `parens-error-2`, `parens-error-3` - Paren errors
- `no-strict` - Non-strict mode
- `strict-units` - Strict units mode

---

## 🎯 Recommended Batching Strategy

### Batch 1: Core Features (High Value, 26 tests)
**Agent Task**: Fix fundamental LESS features
- Selectors, operations, functions, directives
- **Estimated Effort**: 2-3 hours
- **Value**: High - These are core features

### Batch 2: Extend Functionality (High Value, 6 tests)
**Agent Task**: Fix extend feature
- All extend-* tests
- **Estimated Effort**: 1-2 hours
- **Value**: High - Extend is important

### Batch 3: CSS Standards (Medium Value, 6 tests)
**Agent Task**: Fix modern CSS and escaping
- CSS3, escapes, interpolation
- **Estimated Effort**: 1-2 hours
- **Value**: Medium

### Batch 4: Imports (Medium Value, 3 tests)
**Agent Task**: Fix import content insertion
- Inline, once, remote
- **Estimated Effort**: 1 hour
- **Value**: Medium

### Batch 5: Namespacing (Medium Value, 8 tests)
**Agent Task**: Fix namespace output
- **Wait for ISSUE_NAMESPACING.md fixes first**
- Then fix output differences
- **Estimated Effort**: 1-2 hours
- **Value**: Medium

### Batch 6: Math Modes (Low-Medium Value, 9 tests)
**Agent Task**: Fix math mode differences
- Test different math modes
- **Estimated Effort**: 1-2 hours
- **Value**: Medium

### Batch 7: URL Rewriting (Low Value, 6 tests)
**Agent Task**: Fix URL rewriting
- **Wait for ISSUE_URLS.md fix first**
- Then fix rewriting logic
- **Estimated Effort**: 1 hour
- **Value**: Low-Medium

### Batch 8: Error Messages (Low Priority, 20 tests)
**Agent Task**: Fix error message formatting
- These tests correctly fail
- Just need better error messages
- **Estimated Effort**: 2 hours
- **Value**: Low - cosmetic improvements

### Batch 9: Misc (Low Priority, 13 tests)
**Agent Task**: Fix edge cases
- One-off issues
- **Estimated Effort**: 2-3 hours
- **Value**: Low-Medium

---

## 🚫 General Constraints for All Batches

1. **NEVER modify any .js files**
2. **Must pass unit tests**: `pnpm -w test:go:unit`
3. **Must pass target tests**: `pnpm -w test:go:filter -- "test-name"`
4. **No regressions**: All currently passing tests must still pass

---

## 🧪 General Testing Strategy

### For Each Batch

1. **Pick 2-3 tests** from the batch
2. **Run with diff output**:
   ```bash
   LESS_GO_DIFF=1 go test -run "TestIntegrationSuite/.*/test-name" -v
   ```
3. **Identify pattern** - What's wrong with the output?
4. **Find responsible code** - Which GenCSS method or visitor?
5. **Fix and test**
6. **Repeat** for remaining tests in batch

### Example Workflow
```bash
# Pick test
TEST="selectors"

# See the difference
LESS_GO_DIFF=1 go test -run "TestIntegrationSuite/main/$TEST" -v

# This shows:
# Expected: .class { color: red; }
# Got:      .class{color:red}
# Issue: Missing whitespace

# Find code
grep -r "GenCSS" packages/less/src/less/less_go/*.go | grep -i selector

# Fix code
# ... edit the file ...

# Test
go test -run "TestIntegrationSuite/main/$TEST" -v

# Verify no regressions
pnpm -w test:go:unit
pnpm -w test:go:summary
```

---

## 📝 Example Agent Task: Batch 1 (Core Features)

```markdown
# Fix Core Feature CSS Output

## Mission
Fix 26 tests in the core features category that have CSS output differences.

## Tests
- selectors, mixins-nested, mixins-important, mixins-named-args, mixins-guards,
  mixins-guards-default-func, scope, property-accessors, operations, calc,
  merge, strings, variables, directives-bubling, media, variables-in-at-rules,
  container, css-guards, functions, functions-each, extract-and-length, colors,
  comments, comments2, whitespace, permissive-parse

## Strategy
1. Group by similar issues (e.g., all selector tests together)
2. Fix one, test if others also fixed
3. Aim for 70%+ (18/26) - don't get stuck on hard ones

## Debug
LESS_GO_DIFF=1 go test -run "TestIntegrationSuite/main/selectors" -v

## Success
18+ tests passing, no regressions
```

---

## 🎯 Success Metrics for Phase 5

### After All Batches
- **70+ tests fixed** (68% of 102)
- **Perfect CSS match rate**: 8% → 50%+
- **Overall pass rate**: 38% → 85%+

### Stretch Goal
- **85+ tests fixed** (83% of 102)
- **Perfect CSS match rate**: 60%+
- **Overall pass rate**: 90%+

---

## 💡 Key Insights

1. **These all compile** - Not parsing or evaluation issues
2. **Focus on GenCSS methods** - That's where CSS is generated
3. **Compare with JavaScript** - Look at how JS generates CSS
4. **Use LESS_GO_DIFF=1** - Visual diff is essential
5. **Batch by similarity** - Fix patterns, not individual tests
6. **Don't get stuck** - If one test is hard, move to next

---

## 📚 Files to Investigate

Most changes will be in **GenCSS methods** and **visitors**:

### Node GenCSS Methods
- `selector.go` - Selector output
- `declaration.go` - Property output
- `ruleset.go` - Ruleset structure
- `mixin_definition.go` - Mixin output
- `media.go` - Media query output
- `operation.go` - Operation output
- `call.go` - Function call output
- `color.go` - Color output
- `quoted.go` - String output

### Visitors
- `to_css_visitor.go` - Main CSS generation
- `extend_visitor.go` - Extend processing
- `import_visitor.go` - Import content insertion
- `join_selector_visitor.go` - Selector joining

### Output Helpers
- `output.go` - Output buffer and formatting
- `environment.go` - Output context

---

## 🔄 Execution Plan

### Wave 1: High Value (35 tests)
- Batch 1: Core Features (26 tests)
- Batch 2: Extend (6 tests)
- Batch 3: CSS Standards (6 tests)
→ **Expected**: 24+ tests fixed

### Wave 2: Medium Value (20 tests)
- Batch 4: Imports (3 tests)
- Batch 5: Namespacing (8 tests - after runtime fixes)
- Batch 6: Math Modes (9 tests)
→ **Expected**: 12+ tests fixed

### Wave 3: Polish (47 tests)
- Batch 7: URL Rewriting (6 tests - after runtime fixes)
- Batch 8: Error Messages (20 tests)
- Batch 9: Misc (13 tests)
→ **Expected**: 25+ tests fixed

**Total Expected**: 61+ tests fixed (60% of 102)
**Stretch**: 70+ tests fixed (68% of 102)

---

## 📋 Commit Message Template

```
Fix CSS output for [category] tests

Fixed [X] tests in the [category] category that were producing incorrect CSS output.

Root causes:
- [Issue 1: description]
- [Issue 2: description]

Changes:
- [File 1]: [Description]
- [File 2]: [Description]

Tests fixed:
- test1: ✅
- test2: ✅
- test3: ✅
...

Tests remaining: [count] (will address separately)
```

---

## ⚠️ Special Notes

1. **Phase 5 comes LAST** - Fix runtime failures first (Phase 3-4)
2. **Batch wisely** - Similar issues together
3. **Don't aim for 100%** - Some tests might be very hard
4. **Use visual diff** - LESS_GO_DIFF=1 is your friend
5. **Compare with JS** - When stuck, check JavaScript output generation
6. **Iterate** - Fix pattern, test batch, move on
7. **Document hard ones** - If stuck, note why and move on

---

## 🚀 When Ready

Phase 5 begins AFTER:
- ✅ Phase 3 (Parallel fixes) complete
- ✅ Phase 4 (Integration) complete
- ✅ All runtime failures resolved
- ✅ Ready to polish CSS output

Then spawn agents for each batch in waves.
