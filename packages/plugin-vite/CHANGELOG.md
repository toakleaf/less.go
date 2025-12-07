# @lessgo/plugin-vite

## 0.2.5

### Patch Changes

-   Updated dependencies [[`f2c69ea`](https://github.com/toakleaf/less.go/commit/f2c69eab73f2da776255a6489bb59e6ccfceb14f)]:
    -   lessgo@0.4.0

## 0.2.4

### Patch Changes

-   [#470](https://github.com/toakleaf/less.go/pull/470) [`002bbdc`](https://github.com/toakleaf/less.go/commit/002bbdce4d420d60f05ea0382431e494104a4c48) Thanks [@toakleaf](https://github.com/toakleaf)! - Fix CSS relative color syntax and @layer parent selector issues

    -   Add ColorOperand parser for CSS relative color syntax (oklch, hsl, rgb with calc expressions using channel identifiers l, c, h, r, g, b, s)
    -   Fix @layer parent selector (&) resolution to properly join with parent selectors

-   Updated dependencies [[`002bbdc`](https://github.com/toakleaf/less.go/commit/002bbdce4d420d60f05ea0382431e494104a4c48)]:
    -   lessgo@0.3.1

## 0.2.3

### Patch Changes

-   Updated dependencies [[`38c34be`](https://github.com/toakleaf/less.go/commit/38c34be9d5488468a742720a52d735195777bc52)]:
    -   lessgo@0.3.0

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

-   Updated dependencies [[`73595ec`](https://github.com/toakleaf/less.go/commit/73595ec62030a5be0b94c99b6f64410b78bbe0e4)]:
    -   lessgo@0.2.2

## 0.2.1

### Patch Changes

-   [#459](https://github.com/toakleaf/less.go/pull/459) [`cb8bd0e`](https://github.com/toakleaf/less.go/commit/cb8bd0e72b346b943070eb2a2c5a2c2f1e577e4c) Thanks [@toakleaf](https://github.com/toakleaf)! - Fix release script to resolve workspace:\* protocol for @lessgo/plugin-vite

    The npm publish command doesn't understand pnpm's workspace:_ protocol, which caused
    @lessgo/plugin-vite to be published with "lessgo": "workspace:_" as a dependency.
    Now using pnpm pack (which resolves workspace:\* to actual versions) before publishing.

-   Updated dependencies [[`cb8bd0e`](https://github.com/toakleaf/less.go/commit/cb8bd0e72b346b943070eb2a2c5a2c2f1e577e4c)]:
    -   lessgo@0.2.1

## 0.2.0

### Minor Changes

-   Version alignment with lessgo 0.2.0

## 0.1.7

### Minor Changes

-   Initial release of @lessgo/plugin-vite
-   Vite plugin for using less.go as the LESS preprocessor
-   Full TypeScript support with type definitions
-   Source map support
-   HMR support for development
-   Configurable include/exclude patterns
-   Support for LESS plugins, global variables, and modify variables
