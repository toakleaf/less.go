# Agent Prompt 9: Complete Import Reference Implementation (2 Tests)

**Priority**: MEDIUM-HIGH
**Impact**: 2 tests
**Time Estimate**: 1-2 hours
**Difficulty**: MEDIUM

## Task

Complete the import reference implementation. These tests were partially fixed previously (they now compile!) but CSS output still differs.

Affected tests:
- `import-reference` - CSS imports with reference mode
- `import-reference-issues` - Edge cases in reference imports

## Current Status

- Tests compile successfully (big progress!)
- CSS output differs (needs final fix)
- Issue: Mixins from referenced imports aren't available for calling
- Partial PR already created: branch `claude/fix-import-reference-SESSION_ID`

## What You Need to Do

1. **Setup**:
   ```bash
   cd /home/user/less.go
   git checkout -b claude/complete-import-reference-SESSION_ID
   ```

2. **See Current State**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference-issues"
   ```
   - What's in expected but missing in actual?
   - Are mixins not being output?
   - Are imported styles not being used?

3. **Understand the Feature**:
   - Import reference allows importing without outputting styles
   - Mixins/variables from referenced imports should be available
   - Read: `packages/test-data/less/_main/import-reference.less`
   - Reference: `packages/less/src/less/less/import-reference.less`

4. **Debug**:
   ```bash
   LESS_GO_TRACE=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
   ```
   - Are referenced imports being loaded?
   - Are mixins from imports being registered?
   - Are they being found when called?

5. **Find the Root Cause**:
   - Check: `import.go` - is `reference: true` being handled?
   - Check: `import_visitor.go` - are referenced imports registered?
   - Check: `mixin_call.go` - can it find mixins from referenced imports?
   - The issue: Mixins from referenced imports aren't available

6. **Fix the Issue**:
   - Likely need to: Register mixins from referenced imports
   - Or: Make referenced import scope visible to mixin calls
   - Check how non-reference imports work, compare to reference imports

7. **Test**:
   ```bash
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
   LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference-issues"
   ```

8. **Validate**:
   - Unit tests: `pnpm -w test:go:unit`
   - Import tests don't break other imports:
     ```bash
     pnpm -w test:go:filter -- "import"
     ```
   - Full suite: `pnpm -w test:go`

9. **Commit**:
   ```bash
   git add -A
   git commit -m "Complete import reference: Make referenced import mixins available

   - Fixed mixin visibility from referenced imports
   - Resolved import-reference and import-reference-issues output"
   git push -u origin claude/complete-import-reference-SESSION_ID
   ```

## Success Criteria

- ✅ Both tests produce correct output
- ✅ No regressions in other import tests
- ✅ All unit tests pass
- ✅ Referenced imports work with mixins/variables

## Key Files

- Tests:
  - `packages/test-data/less/_main/import-reference.less` / `.css`
  - `packages/test-data/less/_main/import-reference-issues.less` / `.css`
  - See what they import: look for `@import (reference) ...`

- Code:
  - `packages/less/src/less/less_go/import.go`
  - `packages/less/src/less/less_go/import_visitor.go`
  - `packages/less/src/less/less_go/mixin_call.go`
  - `packages/less/src/less/less_go/ruleset.go`

## Important Notes

- These tests were 80% complete from previous work
- They compile now, just CSS output differs
- The main issue: Referenced imports' mixins aren't callable
- This is a visibility/scoping issue, not a parsing issue
- Reference imports shouldn't output CSS, just make contents available

## Context

- See task file: `.claude/tasks/runtime-failures/import-reference.md` (if available)
- Previous session notes what's already been tried
- This is the final push to complete import reference support
