# Task: Fix Media Query Output Formatting

**Status**: Available
**Priority**: Medium
**Tests**: 1 (media)
**Estimated Time**: 1-2 hours

## Problem

Media query CSS output has formatting differences compared to less.js.

## Test Commands

```bash
# Run specific test
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "media"

# Debug mode
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "media"
```

## Key Files

**Go (to fix)**:
- `packages/less/src/less/less_go/media.go`
- `packages/less/src/less/less_go/at_rule.go`

**JavaScript (reference)**:
- `packages/less/src/less/tree/media.js`

## Test Data

```
Input:  packages/test-data/less/_main/media.less
Output: packages/test-data/css/_main/media.css
```

## Likely Issues

1. Media query condition formatting
2. Nested media query merging order
3. Whitespace in media conditions
4. Feature value formatting

## Success Criteria

- `media` test shows "Perfect match!"
- All unit tests pass (`pnpm -w test:go:unit`)
- No regressions in other media-related tests
