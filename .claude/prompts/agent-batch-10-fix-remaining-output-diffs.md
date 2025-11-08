# Agent Prompt 10: Fix Remaining Output Differences (6+ Tests)

**Priority**: MEDIUM
**Impact**: 6+ tests
**Time Estimate**: 3-4 hours
**Difficulty**: MEDIUM

## Task

Fix output differences in various remaining tests. These are good targets for parallel work since they're independent issues:

Target tests (prioritized):
1. `extend-chaining` - Selector extension chaining
2. `mixins-guards` - Guard evaluation with mixins
3. `import-inline` - Inline import formatting
4. `detached-rulesets` - Detached ruleset output
5. `directives-bubling` - Directive handling
6. `media` - Media query output

Choose 1-2 of these to focus on, or work through all if time permits.

## What You Need to Do

### For Each Test:

1. **Setup**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/fix-output-diffs-SESSION_ID
   ```

2. **Investigate**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "extend-chaining"
   ```
   - See what output differs
   - Read the test file: `packages/test-data/less/_main/extend-chaining.less`
   - Compare with reference: `packages/less/src/less/less/extend-chaining.less`

3. **Understand the Feature**:
   - What LESS feature is being tested?
   - How should it work?
   - What's missing or different?

4. **Find Root Cause**:
   - Each test has a different issue:
     - **extend-chaining**: Selector chaining in extends
     - **mixins-guards**: Guard evaluation correctness
     - **import-inline**: Import formatting
     - **detached-rulesets**: Ruleset object handling
     - **directives-bubling**: Directive propagation
     - **media**: Media query generation

5. **Fix**:
   - Make targeted changes to fix the specific issue
   - Test frequently: `LESS_GO_DIFF=1 pnpm -w test:go:filter -- "test-name"`

6. **Move to Next Test** (if time permits):
   - Each test is independent
   - You can fix 1 or multiple tests
   - Commit all fixes together

7. **Final Validation**:
   - Unit tests: `pnpm -w test:go:unit`
   - Full suite: `pnpm -w test:go` (check for regressions)

8. **Commit All Changes**:
   ```bash
   git add -A
   git commit -m "Fix output differences: Multiple tests with CSS output issues

   - Fixed extend-chaining selector output
   - Fixed mixins-guards evaluation
   - Fixed import-inline formatting
   [etc. for all tests you fixed]"
   git push -u origin claude/fix-output-diffs-SESSION_ID
   ```

## Success Criteria

- ✅ At least 2/6 tests now produce correct output
- ✅ No regressions
- ✅ All unit tests pass
- ✅ Clear commit message listing which tests were fixed

## Test-Specific Notes

### extend-chaining
- File: `packages/test-data/less/_main/extend-chaining.less`
- Issue: Chained selector extension not outputting correctly
- Check: `extend.go`, `extend_visitor.go`

### mixins-guards
- File: `packages/test-data/less/_main/mixins-guards.less`
- Issue: Guard conditions evaluated wrong for mixins
- Check: `mixin_definition.go`, `condition.go`, guard evaluation

### import-inline
- File: `packages/test-data/less/_main/import-inline.less`
- Issue: Inline imports not formatted correctly
- Check: `import.go`, `import_visitor.go`, output formatting

### detached-rulesets
- File: `packages/test-data/less/_main/detached-rulesets.less`
- Issue: Detached ruleset objects not rendering
- Check: `detached_ruleset.go`, ruleset evaluation

### directives-bubling
- File: `packages/test-data/less/_main/directives-bubling.less`
- Issue: Directives not propagating through nesting
- Check: `ruleset.go`, directive handling

### media
- File: `packages/test-data/less/_main/media.less`
- Issue: Media query output format
- Check: `media.go`, media query generation

## Strategy

1. **Quick Wins First**: Tests with simpler fixes
2. **Build Momentum**: Each fixed test gives confidence for next
3. **Test Carefully**: After each fix, check for regressions
4. **Commit Strategically**: Group related fixes, one commit per fix

## Expected Impact

If you fix all 6:
- Current: 42 perfect matches
- New: 48 perfect matches
- Success rate: 67.6% 🎉

If you fix 3:
- Current: 42 perfect matches
- New: 45 perfect matches
- Success rate: 63.4%

Even 1 test fixed is a win!

## Notes

- These tests are all independent
- No single test blocks others
- Work through them in priority order
- Use LESS_GO_DIFF=1 frequently to see progress
- Tests provide good complexity progression

## Resources

- Use previous completed task files as reference patterns
- Check how similar features are implemented in working tests
- Compare with JavaScript implementation when stuck
