# Task: Fix Directive Bubbling Output

**Status**: Available
**Priority**: Medium
**Tests**: 1 (directives-bubling)
**Estimated Time**: 2-3 hours

## Problem

Directive bubbling (how `@supports`, `@document`, and similar at-rules bubble up through selectors) produces different output order or grouping than less.js.

## Test Commands

```bash
# Run specific test
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"

# Debug mode
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "directives-bubling"
```

## Key Files

**Go (to fix)**:
- `packages/less/src/less/less_go/at_rule.go`
- `packages/less/src/less/less_go/ruleset.go`

**JavaScript (reference)**:
- `packages/less/src/less/tree/atrule.js`
- `packages/less/src/less/tree/ruleset.js`

## Test Data

```
Input:  packages/test-data/less/_main/directives-bubling.less
Output: packages/test-data/css/_main/directives-bubling.css
```

## Background

When you nest a directive like `@supports` inside a selector:

```less
.foo {
  @supports (display: grid) {
    display: grid;
  }
}
```

It should bubble up to become:

```css
@supports (display: grid) {
  .foo {
    display: grid;
  }
}
```

## Likely Issues

1. Directive bubble order (which comes first when multiple)
2. Selector joining when bubbling up
3. Rule grouping within bubbled directives
4. Output ordering differences

## Success Criteria

- `directives-bubling` test shows "Perfect match!"
- All unit tests pass (`pnpm -w test:go:unit`)
- No regressions in other at-rule or directive tests
