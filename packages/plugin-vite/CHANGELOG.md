# @lessgo/plugin-vite

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
