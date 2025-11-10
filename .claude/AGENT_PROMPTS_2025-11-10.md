# Agent Prompts for less.go Port - November 10, 2025

**Current Status**: 79/185 perfect matches (42.7%), 76.2% overall success rate
**Branch Pattern**: `claude/fix-{task-name}-{session-id}`

## ⚠️ CRITICAL: Validation Requirements

Before creating ANY pull request, you MUST:
1. ✅ Run ALL unit tests: `pnpm -w test:go:unit` (must pass 100%)
2. ✅ Run ALL integration tests: `pnpm -w test:go`
3. ✅ Check for regressions: Baseline is **79 perfect matches**, **ZERO tolerance for regressions**
4. ✅ Verify your specific test(s) now show "Perfect match!"

---

## Prompt 1: Fix Import-Reference Functionality (HIGH PRIORITY)

**Task**: Fix the `@import (reference)` functionality so referenced imports don't output CSS but their selectors/mixins remain available.

**Tests to Fix** (2 tests):
- `main/import-reference`
- `main/import-reference-issues`

**Expected Impact**: +2 perfect matches → 81/185 (43.8%)

**Investigation**:
1. Read `.claude/tasks/runtime-failures/import-reference.md` for detailed context
2. Compare Go vs JS implementations:
   - `packages/less/src/less/import-visitor.js` (JS)
   - `packages/less/src/less/less_go/import_visitor.go` (Go)
3. Key question: Is the `reference` flag being preserved and checked during CSS generation?

**Files to Modify**:
- `import.go` - Import node definition
- `import_visitor.go` - Import processing
- `ruleset.go` - CSS generation (check reference flag)

**Debug Commands**:
```bash
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
LESS_GO_TRACE=1 pnpm -w test:go:filter -- "import-reference"
```

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "import-reference"  # Both tests should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 81
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 2: Fix Functions Test Output Issues

**Task**: Fix CSS output differences in the `functions` test. The test compiles but produces incorrect CSS output.

**Tests to Fix** (1 test):
- `main/functions`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff to see what's wrong:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "functions"
   ```
2. Compare expected vs actual CSS output
3. Check which functions are producing incorrect output
4. Compare JS vs Go implementations in `packages/less/src/less/less_go/functions/`

**Files to Check**:
- `functions/*.go` - Individual function implementations
- `call.go` - Function call evaluation
- Test data: `packages/test-data/less/_main/functions.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "functions"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 3: Fix URL Handling Edge Cases (3 tests)

**Task**: Fix CSS output differences in URL handling tests across three test suites.

**Tests to Fix** (3 tests):
- `main/urls`
- `static-urls/urls`
- `url-args/urls`

**Expected Impact**: +3 perfect matches → 82/185 (44.3%)

**Investigation**:
1. Run each test with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "urls"
   ```
2. Compare the three test outputs to identify common patterns
3. Check URL processing logic in Go vs JS

**Files to Check**:
- `url.go` - URL node and processing
- `ruleset.go` - URL output generation
- Compare with `packages/less/src/less/tree/url.js`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "urls"  # All 3 should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 82
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 4: Fix Detached Rulesets Output

**Task**: Fix CSS output formatting in the `detached-rulesets` test.

**Tests to Fix** (1 test):
- `main/detached-rulesets`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "detached-rulesets"
   ```
2. Examine output differences - likely spacing, formatting, or structure
3. Compare with JS: `packages/less/src/less/tree/detached-ruleset.js`

**Files to Check**:
- `detached_ruleset.go` - DetachedRuleset implementation
- `ruleset.go` - CSS generation
- Test data: `packages/test-data/less/_main/detached-rulesets.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "detached-rulesets"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 5: Fix Media Query Output

**Task**: Fix CSS output issues in the `media` test related to media query handling.

**Tests to Fix** (1 test):
- `main/media`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "media"
   ```
2. Check media query output formatting
3. Compare with JS: `packages/less/src/less/tree/media.js`

**Files to Check**:
- `media.go` - Media node implementation
- `at_rule.go` - At-rule handling
- Test data: `packages/test-data/less/_main/media.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "media"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 6: Fix Directives Bubbling Output

**Task**: Fix CSS output in the `directives-bubling` test (note: spelling is intentional).

**Tests to Fix** (1 test):
- `main/directives-bubling`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"
   ```
2. Examine how directives (@media, @supports, @container) bubble up through nested rulesets
3. Compare with recent fixes to similar issues

**Files to Check**:
- `ruleset.go` - Directive bubbling logic
- `at_rule.go` - At-rule handling
- `media.go` - Media query specific handling
- Test data: `packages/test-data/less/_main/directives-bubling.less`

**Context**: This may be related to recent container and media fixes.

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "directives-bubling"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 7: Fix Container Query Output

**Task**: Fix CSS output in the `container` test related to CSS container queries.

**Tests to Fix** (1 test):
- `main/container`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "container"
   ```
2. Check container query output and bubbling behavior
3. May be related to directives-bubbling and media tests

**Files to Check**:
- `at_rule.go` - Container at-rule handling
- `ruleset.go` - Bubbling logic
- Test data: `packages/test-data/less/_main/container.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "container"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 8: Fix CSS-3 Test Output

**Task**: Fix CSS output issues in the `css-3` test covering CSS3 features.

**Tests to Fix** (1 test):
- `main/css-3`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "css-3"
   ```
2. Examine what CSS3 features are being tested
3. Check output differences

**Files to Check**:
- Various tree nodes depending on test content
- Test data: `packages/test-data/less/_main/css-3.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "css-3"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 9: Fix Property Name Interpolation

**Task**: Fix CSS output in the `property-name-interp` test for property name interpolation.

**Tests to Fix** (1 test):
- `main/property-name-interp`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "property-name-interp"
   ```
2. Check how property names with `@{variable}` interpolation are handled
3. Compare with JS implementation

**Files to Check**:
- `declaration.go` - Declaration/property handling
- `ruleset.go` - Property output
- Test data: `packages/test-data/less/_main/property-name-interp.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "property-name-interp"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## Prompt 10: Fix Selectors Test Output

**Task**: Fix CSS output in the `selectors` test covering various selector edge cases.

**Tests to Fix** (1 test):
- `main/selectors`

**Expected Impact**: +1 perfect match → 80/185 (43.2%)

**Investigation**:
1. Run with diff:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "selectors"
   ```
2. Examine which selector patterns are producing incorrect output
3. Compare with JS selector handling

**Files to Check**:
- `selector.go` - Selector implementation
- `element.go` - Element/selector parts
- `ruleset.go` - Selector output
- Test data: `packages/test-data/less/_main/selectors.less`

**Validation**:
```bash
pnpm -w test:go:unit  # Must pass 100%
pnpm -w test:go:filter -- "selectors"  # Should show "Perfect match!"
pnpm -w test:go:summary | grep "Perfect CSS Matches"  # Should show 80
```

**Reminder**: Check against baseline - currently 79 perfect matches, 76.2% success rate. NO REGRESSIONS ALLOWED.

---

## 🎯 Recommended Order for Maximum Impact

### Fast Path to 80% Success Rate (Need +9 perfect matches)
1. **Prompt 1**: import-reference (+2) → 81/185
2. **Prompt 2**: functions (+1) → 82/185
3. **Prompt 3**: urls (+3) → 85/185
4. **Prompt 4**: detached-rulesets (+1) → 86/185
5. **Prompt 5**: media (+1) → 87/185
6. **Prompt 6**: directives-bubling (+1) → 88/185

**Result**: 88/185 = **47.6% perfect matches**, **~80-82% overall success rate**

### Alternative: Category Completion
1. **Prompt 1**: import-reference - Complete import category
2. **Prompt 6**: directives-bubling - Complete directive category
3. **Prompt 7**: container - Complete container category
4. **Prompt 5**: media - Improve media handling
5. **Prompt 3**: urls - Handle URL edge cases

---

## 📊 Current Stats & Goals

**Baseline** (2025-11-10):
- Perfect matches: **79/185 (42.7%)**
- Overall success: **76.2%**
- Unit tests: **2,290+ passing (99.9%)**

**Short-term goal** (1-2 weeks):
- Perfect matches: **88/185 (47.6%)**
- Overall success: **80%+**
- Zero regressions

**Medium-term goal** (1 month):
- Perfect matches: **100/185 (54%)**
- Overall success: **85%+**
- Most common features 100% working

---

## 🔧 General Workflow for All Prompts

1. **Setup**:
   ```bash
   git checkout -b claude/fix-{task-name}-{session-id}
   git pull origin main  # Get latest changes
   ```

2. **Investigate**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "{test-name}"
   ```

3. **Compare implementations**: Read both JS and Go files side-by-side

4. **Make changes**: Edit Go files only, never JS

5. **Test continuously**:
   ```bash
   pnpm -w test:go:filter -- "{test-name}"  # Quick check
   pnpm -w test:go:unit  # Regression check
   ```

6. **Final validation**:
   ```bash
   pnpm -w test:go:unit  # ALL must pass
   pnpm -w test:go  # Check no regressions
   pnpm -w test:go:summary  # Verify improvement
   ```

7. **Commit & Push**:
   ```bash
   git add .
   git commit -m "Fix {test-name}: {brief description of what was fixed}"
   git push -u origin claude/fix-{task-name}-{session-id}
   ```

---

## ⚠️ Critical Reminders

- **NEVER modify JavaScript files** - they're the reference implementation
- **ALWAYS run all unit tests** before pushing
- **CHECK for regressions** - baseline is 79 perfect matches
- **COMPARE with JS** - when in doubt, match JavaScript behavior exactly
- **READ test files** - understand what's being tested before fixing
- **USE debug flags** - LESS_GO_DIFF, LESS_GO_TRACE, LESS_GO_DEBUG

---

**Generated**: 2025-11-10
**Valid through**: These prompts remain valid until test counts change significantly
**Next review**: After 5+ prompts completed or 80% success rate reached
