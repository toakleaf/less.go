# Task: Fix Import Reference Output

**Status**: Available
**Priority**: High
**Tests**: 2 (import-reference, import-reference-issues)
**Estimated Time**: 2-3 hours

## Problem

Files imported with `(reference)` option output CSS when they shouldn't. Only explicitly used selectors/mixins should appear in output.

## Current Behavior

Referenced imports are outputting their CSS directly instead of only making selectors/mixins available for extends and mixin calls.

## Expected Behavior

From less.js:
1. `@import (reference) "file.less";` should NOT output any CSS
2. Selectors/mixins from referenced files should be available for:
   - Extending via `:extend()`
   - Calling as mixins
   - Variable lookup
3. Only explicitly used parts should appear in output
4. CSS imports like `@import (reference) "file.css";` should remain as `@import` statements

## Test Commands

```bash
# Run specific tests
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"

# Debug mode
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "import-reference"
```

## Key Files

**Go (to fix)**:
- `packages/less/src/less/less_go/import.go`
- `packages/less/src/less/less_go/import_visitor.go`
- `packages/less/src/less/less_go/ruleset.go`

**JavaScript (reference)**:
- `packages/less/src/less/import-visitor.js`
- `packages/less/src/less/tree/import.js`
- `packages/less/src/less/tree/ruleset.js`

## Test Data

```
Input:  packages/test-data/less/_main/import-reference.less
Output: packages/test-data/css/_main/import-reference.css
```

## Likely Root Cause

The `reference` flag is either:
1. Not preserved during import processing
2. Not propagated to the imported ruleset
3. Not checked during CSS generation in `Ruleset.GenCSS()`

## Success Criteria

- `import-reference` test shows "Perfect match!"
- `import-reference-issues` test shows "Perfect match!"
- All unit tests pass (`pnpm -w test:go:unit`)
- No regressions in other import tests
