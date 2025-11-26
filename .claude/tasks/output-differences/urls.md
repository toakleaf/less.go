# Task: Fix URL Handling Edge Cases

**Status**: Available
**Priority**: High
**Tests**: 3 (urls in main/static-urls/url-args suites)
**Estimated Time**: 2-3 hours

## Problem

URL handling has edge cases that produce different output than less.js. This affects URL encoding, path handling, and URL rewriting in various contexts.

## Affected Tests

1. `urls` (main suite)
2. `urls` (static-urls suite)
3. `urls` (url-args suite)

## Test Commands

```bash
# Run specific tests
LESS_GO_DIFF=1 pnpm -w test:go:filter -- "urls"

# Debug mode
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go:filter -- "urls"
```

## Key Files

**Go (to fix)**:
- `packages/less/src/less/less_go/url.go`
- `packages/less/src/less/less_go/ruleset.go`

**JavaScript (reference)**:
- `packages/less/src/less/tree/url.js`

## Test Data

```
Input:  packages/test-data/less/_main/urls.less
Output: packages/test-data/css/_main/urls.css

Input:  packages/test-data/less/static-urls/urls.less
Output: packages/test-data/css/static-urls/urls.css

Input:  packages/test-data/less/url-args/urls.less
Output: packages/test-data/css/url-args/urls.css
```

## Likely Issues

1. URL encoding differences (special characters)
2. Path resolution in different contexts
3. URL rewriting with `rewrite-urls` option
4. Data URLs handling
5. Quote handling in URLs

## Success Criteria

- All 3 `urls` tests show "Perfect match!"
- All unit tests pass (`pnpm -w test:go:unit`)
- No regressions in other URL-related tests
