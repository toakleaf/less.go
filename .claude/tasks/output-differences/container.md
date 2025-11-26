# Task: Fix Container Query Handling

**Status**: Available
**Priority**: Medium
**Tests**: 1 (container)
**Estimated Time**: 2-3 hours

## Problem

Container queries (`@container`) may not be fully implemented or produce different output than less.js.

## Background

Container queries are a CSS feature that allows styling based on container size. Less.js supports passing these through to the output.

## Test Commands

```bash
# Run specific test
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "container"

# Debug mode
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "container"
```

## Key Files

**Go (to fix)**:
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/container.go` (if exists)

**JavaScript (reference)**:
- `packages/less/src/less/tree/container.js` (if exists)
- `packages/less/src/less/tree/atrule.js`

## Test Data

```
Input:  packages/test-data/less/_main/container.less
Output: packages/test-data/css/_main/container.css
```

## Likely Issues

1. Container at-rule not recognized
2. Container conditions not parsed correctly
3. Nested rules not handled properly
4. Output formatting differences

## Success Criteria

- `container` test shows "Perfect match!"
- All unit tests pass (`pnpm -w test:go:unit`)
- No regressions in other at-rule tests
