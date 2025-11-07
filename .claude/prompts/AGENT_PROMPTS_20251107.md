# 10 Independent Agent Prompts - 2025-11-07

**Base Instructions**: Each agent should run the full test suite before submitting a PR:
- `pnpm -w test:go:unit` - ALL unit tests must pass
- `pnpm -w test:go` - ALL integration tests must pass
- Compare results to `/home/user/less.go/.claude/tracking/TEST_STATUS_REPORT_20251107.md`
- **ZERO regressions tolerance** - No test should go from passing to failing

---

## Agent Prompt 1: Fix Math-Parens Suite Output Differences

**Priority**: HIGH
**Impact**: 2 tests (css, mixins-args)
**Time Estimate**: 2-3 hours
**Difficulty**: Medium

### Task
The `math-parens` test suite has 4 tests, 2 are now perfect, but 2 have output differences:
- `math-parens/css` - Output differs
- `math-parens/mixins-args` - Output differs
- `math-parens/parens` - Output differs

### What to do
1. Read test files in `test-data/less/math-parens/` to understand what they test
2. Run `pnpm -w test:go` and look at the actual vs expected output for each test
3. Identify why the output differs from the JavaScript implementation
4. Fix the runtime evaluation or CSS output generation in the less-go code
5. **REGRESSION CHECK**: Before committing, run `pnpm -w test:go:unit` and `pnpm -w test:go`
6. Compare to baseline in `TEST_STATUS_REPORT_20251107.md` - should have 0 regressions

### Files to investigate
- Test files: `packages/less/test-data/less/math-parens/`
- Runtime: `packages/less/src/less/less_go/evaluator.go`
- Math operations: Look for math mode handling
- Expected: Both tests should show "✅ Perfect match!"

### Success Criteria
- Both failing math-parens tests become perfect matches
- No regressions in other tests (80 perfect matches still)

---

## Agent Prompt 2: Fix URL Processing Output Differences

**Priority**: HIGH
**Impact**: 6+ tests (urls suite, rewrite-urls, rootpath-rewrite-urls)
**Time Estimate**: 3-4 hours
**Difficulty**: Medium

### Task
Multiple URL-related tests have output differences. These are mostly working but the CSS output doesn't match:
- `static-urls/urls` - Output differs
- `url-args/urls` - Output differs
- `rewrite-urls-all/rewrite-urls-all` - Output differs
- `rewrite-urls-local/rewrite-urls-local` - Output differs
- `rootpath-rewrite-urls-all/rootpath-rewrite-urls-all` - Output differs
- `rootpath-rewrite-urls-local/rootpath-rewrite-urls-local` - Output differs

### What to do
1. Examine the test files in `test-data/less/` for url-related tests
2. Run `pnpm -w test:go` and compare expected vs actual CSS output
3. Look for patterns in what's different (URL rewriting, path handling, etc.)
4. Fix URL processing in the evaluator or CSS generation
5. **REGRESSION CHECK**: Run full test suite before committing
6. Verify: 6 more tests should become "✅ Perfect match!"

### Files to investigate
- Test files: `packages/less/test-data/less/` (static-urls, url-args, rewrite-urls*)
- URL handling: `packages/less/src/less/less_go/evaluator.go` (URL processing)
- CSS output: Check if URLs are being rewritten correctly

### Success Criteria
- All 6 URL-related tests become perfect matches
- No regressions (still have 80 perfect baseline)

---

## Agent Prompt 3: Fix Import Output Formatting Issues

**Priority**: HIGH
**Impact**: 5+ tests (import-inline, import-reference*, import-remote)
**Time Estimate**: 3-4 hours
**Difficulty**: Medium

### Task
Several import-related tests compile successfully but the CSS output format differs:
- `import-inline` - Output differs
- `import-reference` - Output differs
- `import-reference-issues` - Output differs
- `import-remote` - Output differs
- `import-interpolation` - Output differs

### What to do
1. Look at the test files in `test-data/less/` for import tests
2. Run `pnpm -w test:go` and examine the differences
3. Identify if it's:
   - Missing CSS content from imports
   - Wrong CSS formatting
   - Variable interpolation issues
   - File reference issues
4. Fix the import handling or output formatting
5. **REGRESSION CHECK**: Full test suite before PR
6. Goal: 5 more perfect matches

### Files to investigate
- Test files: `test-data/less/_main/import-*.less`
- Import handling: `packages/less/src/less/less_go/evaluator.go`
- File loading: Look for import statement processing

### Success Criteria
- At least 3-4 of the import tests become perfect matches
- No regressions from the 80-test baseline

---

## Agent Prompt 4: Fix Math-Parens-Division Suite Remaining Tests

**Priority**: MEDIUM
**Impact**: 2 tests (parens, mixins-args)
**Time Estimate**: 2-3 hours
**Difficulty**: Medium

### Task
Similar to math-parens but for the parens-division mode:
- `math-parens-division/parens` - Output differs
- `math-parens-division/mixins-args` - Output differs

Note: `media-math` and `new-division` already perfect!

### What to do
1. Compare the test files in `test-data/less/math-parens-division/` with the working ones
2. Understand the difference between parens and parens-division modes
3. Run tests and examine output differences
4. Fix math mode handling for division
5. **REGRESSION CHECK**: Full test suite
6. Goal: 2 more perfect matches

### Files to investigate
- Test files: `test-data/less/math-parens-division/`
- Math modes: Look for division vs math handling
- Runtime: `evaluator.go` - math mode context

### Success Criteria
- Both tests become perfect matches
- No regressions (keep 80 as baseline)

---

## Agent Prompt 5: Fix Selector and Interpolation Issues

**Priority**: MEDIUM
**Impact**: 8+ tests
**Time Estimate**: 2-3 hours
**Difficulty**: Medium-High

### Task
Several tests fail due to selector and property name interpolation issues:
- `selectors` - Output differs
- `property-accessors` - Output differs
- `property-name-interp` - Output differs
- `extend-exact` - Output differs (extra empty rulesets)
- `css-3` - Output differs
- `css-escapes` - Output differs

### What to do
1. Review the test files in `test-data/less/` for selector and property tests
2. Look at what the tests expect vs what we output
3. Identify if the issue is:
   - Selector interpolation not working
   - Property name interpolation failing
   - CSS escaping issues
   - Extra/missing output
4. Fix the interpolation or output generation
5. **REGRESSION CHECK**: Full suite before PR
6. Goal: 5+ more perfect matches

### Files to investigate
- Test files: `test-data/less/selectors.less`, `property-*.less`, etc.
- Selector handling: `packages/less/src/less/less_go/evaluator.go`
- Interpolation: Look for selector/property interpolation code

### Success Criteria
- 5+ tests become perfect matches
- No regressions

---

## Agent Prompt 6: Fix Formatting and Whitespace Issues

**Priority**: MEDIUM
**Impact**: 5+ tests
**Time Estimate**: 2-3 hours
**Difficulty**: Low-Medium

### Task
Several tests have correct logic but wrong formatting in the CSS output:
- `comments` - Formatting differs
- `comments2` - Comment placement differs
- `whitespace` - Whitespace differs
- `parse-interpolation` - Formatting differs
- `variables-in-at-rules` - Formatting differs

### What to do
1. Review test files to understand the CSS being generated
2. Run tests and compare expected vs actual output
3. Identify formatting differences:
   - Wrong newlines
   - Missing/extra spaces
   - Comment placement
4. Fix CSS output generation or formatting code
5. **REGRESSION CHECK**: Full suite
6. Goal: 5 more perfect matches

### Files to investigate
- Test files: `test-data/less/comments.less`, etc.
- Output generation: Look for CSS generation code
- Formatting: Check indentation and whitespace handling

### Success Criteria
- All 5 tests become perfect matches
- No regressions

---

## Agent Prompt 7: Fix Mixin-Important and Nested Issues

**Priority**: MEDIUM
**Impact**: 2+ tests
**Time Estimate**: 1-2 hours
**Difficulty**: Low-Medium

### Task
Two mixin-related tests have output differences:
- `mixins-important` - Output differs
- `mixins-nested` - Has extra empty ruleset in output

### What to do
1. Look at the test files in `test-data/less/`
2. Understand what `!important` should do
3. Look at what extra output is being generated in mixins-nested
4. Fix the mixin evaluation or output
5. **REGRESSION CHECK**: Full test suite
6. Goal: 2 perfect matches

### Files to investigate
- Test files: `test-data/less/mixins-*.less`
- Mixin evaluation: `evaluator.go`
- Output: Check for empty rulesets being generated

### Success Criteria
- Both tests become perfect matches
- No regressions

---

## Agent Prompt 8: Fix Math-Always Suite Remaining Test

**Priority**: MEDIUM
**Impact**: 1 test
**Time Estimate**: 1-2 hours
**Difficulty**: Low

### Task
The `math-always` suite has 2 perfect tests already:
- `mixins-guards` ✅
- `no-sm-operations` ✅

But the suite status shows only 2/2 perfect, so it's already complete! However, double-check the math-always suite isn't missing anything.

### What to do
1. Verify all math-always tests are perfect
2. If any are still output differences, fix them
3. Check for regressions
4. Consider this potentially DONE

### Files to investigate
- Test suite: `test-data/less/math-always/`

### Success Criteria
- All tests in math-always are perfect matches
- Suite shows 2/2 or better

---

## Agent Prompt 9: Fix Remaining Extend Issues

**Priority**: MEDIUM
**Impact**: 1 test (extend-exact is the main one)
**Time Estimate**: 1-2 hours
**Difficulty**: Low-Medium

### Task
The extend functionality is mostly working:
- extend ✅
- extend-media ✅
- extend-nest ✅
- extend-selector ✅
- extend-clearfix ✅

But one still has issues:
- `extend-exact` - Has extra content in output

### What to do
1. Look at test-data/less/extend-exact.less
2. Examine the expected vs actual output
3. Identify what extra content is being generated
4. Fix extend visitor to not generate extra rulesets
5. **REGRESSION CHECK**: Full suite
6. Goal: 1 more perfect match

### Files to investigate
- Test file: `test-data/less/extend-exact.less`
- Extend handling: Look for extend visitor code
- Output: Where are extra rulesets coming from?

### Success Criteria
- extend-exact becomes perfect match
- No regressions (keep 80+ perfect matches)

---

## Agent Prompt 10: Investigate and Fix Remaining Output Differences

**Priority**: LOWER
**Impact**: 20+ tests (various categories)
**Time Estimate**: Varies (pick subset)
**Difficulty**: Varies

### Task
Many tests have output differences in various categories:
- `calc` - Output differs
- `container` - Output differs
- `detached-rulesets` - Output differs
- `directives-bubling` - Output differs
- `extract-and-length` - Output differs
- `functions` - Output differs
- `functions-each` - Output differs
- `media` - Output differs
- `merge` - Output differs
- `mixins-guards` - Output differs
- `namespacing-3` - Output differs
- `permissive-parse` - Output differs
- `strings` - Output differs
- And more...

### What to do
1. Pick 2-3 tests that seem related (e.g., all function-related or all formatting)
2. Run the tests and examine the differences
3. Identify what feature is broken or incomplete
4. Fix the implementation
5. **REGRESSION CHECK**: Full suite before each commit
6. Goal: Get a few more tests to perfect match

### Note
This is exploratory work to tackle the remaining issues. Pick tests strategically based on patterns you notice.

### Success Criteria
- At least 2-3 additional tests become perfect matches
- No regressions

---

## Cross-Cutting Recommendations

1. **Group by root cause**: Many tests likely fail for the same reason (e.g., math mode, formatting). Fixing one might fix several.

2. **Use LESS_GO_TRACE**: Set `LESS_GO_TRACE=1` before running tests to see evaluation trace:
   ```
   LESS_GO_TRACE=1 pnpm -w test:go
   ```

3. **Compare with JavaScript**: Compare the Go output with what less.js produces to understand what should happen.

4. **Keep test baseline handy**: Reference `/home/user/less.go/.claude/tracking/TEST_STATUS_REPORT_20251107.md` to understand the baseline.

5. **Commit frequently**: Small, focused commits are better than large ones.

6. **Document findings**: If you discover the root cause of a category of failures, add notes to the task files.

---

## Recommended Parallel Execution

These agents can work in parallel (they don't conflict):
- Agent 1: Fix math-parens suite
- Agent 2: Fix URL processing
- Agent 3: Fix import formatting
- Agent 4: Fix selector/interpolation
- Agent 5: Fix formatting issues

Sequential after the above (dependencies):
- Agent 6: Math-parens-division (builds on math fixes)
- Agent 7: Mixin issues
- Agent 8: Remaining extends
- Agent 9: Investigate other differences

**Expected outcome**: 15-20 additional perfect matches (+19-25% improvement), reaching ~95-100 perfect matches (60-62% success rate).
