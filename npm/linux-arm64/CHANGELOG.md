# @lessgo/linux-arm64

## 0.4.0

### Minor Changes

-   [#511](https://github.com/toakleaf/less.go/pull/511) [`f2c69ea`](https://github.com/toakleaf/less.go/commit/f2c69eab73f2da776255a6489bb59e6ccfceb14f) Thanks [@toakleaf](https://github.com/toakleaf)! - ## Bug Fixes

    -   Fix `!important` handling in variable declarations, merge declarations, and anonymous values
    -   Fix Anonymous color values in `contrast()` and other color functions
    -   Preserve `calc()` expressions in mixin default values
    -   Fix `@page` pseudo-selectors and variable hoisting in imports
    -   Fix parser handling of unmatched closing braces in `ParseUntil` to prevent over-consumption
    -   Fix `Sub()` parser to match Less.js behavior for parens-error tests
    -   Preserve parentheses in parenthesized list values with comma separation
    -   Fix extend property formatting to match Less.js behavior
    -   Fix namespace property access evaluation to Dimension for math operations
    -   Prevent property merge duplication when using property accessors
    -   Fix arrow function syntax handling with parameters in Less variables
    -   Preserve frames for variable imports during path interpolation
    -   Fix Quoted node evaluation inside URL for variable interpolation
    -   Prevent incorrect deduplication of multi-target extends
    -   Remove extra whitespace when extend-only rulesets are inside media queries
    -   Stabilize import-remote test by pinning `@less/test-data@4.4.2`

    ## Performance Improvements

    -   Add `ReleaseTree` function for improved object pool utilization
    -   Convert CSS output methods from slice concatenation to `strings.Builder` for reduced allocations
    -   Add context map pool for evaluation phase allocations
    -   Optimize `Selector.CreateDerived` to reduce map allocations
    -   Optimize `EvalCall` memory allocations in `MixinDefinition`
    -   Reduce unnecessary map allocations in CSS generation phase

    ## Testing

    -   Expand benchmark test suite to include custom integration tests
    -   Add extensive test variations for overfitting detection across math modes, mixins, and other features

## 0.3.1

## 0.3.0

### Minor Changes

-   [#469](https://github.com/toakleaf/less.go/pull/469) [`38c34be`](https://github.com/toakleaf/less.go/commit/38c34be9d5488468a742720a52d735195777bc52) Thanks [@toakleaf](https://github.com/toakleaf)! - Full compatibility with Less.js v4.4.2

    -   All 195 integration tests now passing (100%)
    -   Fixed CSS layer syntax and nested @layer handling
    -   Fixed @starting-style at-rule nesting behavior
    -   Fixed container queries with scroll-state() syntax
    -   Fixed color channel identifiers (l,c,h,r,g,b,s) as operands
    -   Fixed fade function alpha value handling

## 0.2.2

### Patch Changes

-   [#463](https://github.com/toakleaf/less.go/pull/463) [`73595ec`](https://github.com/toakleaf/less.go/commit/73595ec62030a5be0b94c99b6f64410b78bbe0e4) Thanks [@toakleaf](https://github.com/toakleaf)! - Fix rgba() ignoring variable alpha values

    The rgba() function (and rgb(), hsl(), hsla()) was ignoring alpha values when passed as variables. For example:

    ```less
    @alpha: 0.5;
    color: rgba(
        255,
        0,
        0,
        @alpha
    ); // was producing #ff0000 instead of rgba(255, 0, 0, 0.5)
    ```

    This fix ensures that variable arguments are properly evaluated before being passed to color functions.

## 0.2.1

### Patch Changes

-   [#459](https://github.com/toakleaf/less.go/pull/459) [`cb8bd0e`](https://github.com/toakleaf/less.go/commit/cb8bd0e72b346b943070eb2a2c5a2c2f1e577e4c) Thanks [@toakleaf](https://github.com/toakleaf)! - Fix release script to resolve workspace:\* protocol for @lessgo/plugin-vite

    The npm publish command doesn't understand pnpm's workspace:_ protocol, which caused
    @lessgo/plugin-vite to be published with "lessgo": "workspace:_" as a dependency.
    Now using pnpm pack (which resolves workspace:\* to actual versions) before publishing.

## 0.2.0

### Minor Changes

-   Added @lessgo/plugin-vite - Vite plugin for using less.go as the LESS preprocessor

## 0.1.7

### Patch Changes

-   [#456](https://github.com/toakleaf/less.go/pull/456) [`0356932`](https://github.com/toakleaf/less.go/commit/0356932525100e19147d4c6ad9056228f244b4e0) Thanks [@toakleaf](https://github.com/toakleaf)! - Add plugins option to Node.js compile() API

## 0.1.6

### Patch Changes

-   [`3acfcd0`](https://github.com/toakleaf/less.go/commit/3acfcd0ae8af315f10c247e956b6c93f9399ad48) Thanks [@toakleaf](https://github.com/toakleaf)! - Include plugin-host.js in built package

## 0.1.5

### Patch Changes

-   [`e1cdd2e`](https://github.com/toakleaf/less.go/commit/e1cdd2eccf4614fc66f2865b7e5215960efdd218) Thanks [@toakleaf](https://github.com/toakleaf)! - Added --plugin flag and fixed tests

## 0.1.4

### Patch Changes

-   [`7d4593c`](https://github.com/toakleaf/less.go/commit/7d4593c4453db0da50ab3f2c198af246427bab5a) Thanks [@toakleaf](https://github.com/toakleaf)! - Fixed stdin issue and source maps

## 0.1.3

### Patch Changes

-   [`8d61e93`](https://github.com/toakleaf/less.go/commit/8d61e939f533d16abfe241f2f2aefbf65875f40b) Thanks [@toakleaf](https://github.com/toakleaf)! - Fixing release process

## 0.1.2

### Patch Changes

-   [`a91922d`](https://github.com/toakleaf/less.go/commit/a91922d104766d00ab9faeb44f03858bdffce87c) Thanks [@toakleaf](https://github.com/toakleaf)! - Fixed release process - hopefully

## 0.1.1

### Patch Changes

-   [`09cd983`](https://github.com/toakleaf/less.go/commit/09cd983d1e460c72ddbac8f315a1e9b7b2eaea8e) Thanks [@toakleaf](https://github.com/toakleaf)! - Release setup with binary publishing
