# @lessgo/plugin-vite

Vite plugin for using [less.go](https://github.com/toakleaf/less.go) (lessc-go) as the LESS preprocessor.

This plugin intercepts `.less` imports and compiles them using the lessgo Node.js API, providing fast LESS compilation with the Go-based compiler.

## Installation

```bash
npm install @lessgo/plugin-vite lessgo
# or
pnpm add @lessgo/plugin-vite lessgo
# or
yarn add @lessgo/plugin-vite lessgo
```

## Usage

```ts
// vite.config.ts
import { defineConfig } from 'vite';
import lessgo from '@lessgo/plugin-vite';

export default defineConfig({
  plugins: [
    lessgo(),
  ],
});
```

## Options

### `compress`

Type: `boolean`
Default: `false`

Minify CSS output.

```ts
lessgo({
  compress: true,
})
```

### `paths`

Type: `string[]`

Additional include paths for `@import` resolution. These paths will be searched when resolving `@import` statements.

```ts
lessgo({
  paths: ['./src/styles', './node_modules'],
})
```

### `globalVars`

Type: `Record<string, string>`

Global variables to inject into LESS compilation. Variables are available in all compiled LESS files.

```ts
lessgo({
  globalVars: {
    primaryColor: '#007bff',
    theme: '"dark"', // String values need quotes
  },
})
```

### `modifyVars`

Type: `Record<string, string>`

Variables to modify (override) in LESS compilation. These take precedence over variables defined in LESS files.

```ts
lessgo({
  modifyVars: {
    primaryColor: '#ff0000',
  },
})
```

### `plugins`

Type: `(string | { name: string; options?: string })[]`

LESS plugins to load. Can be plugin names (with or without `less-plugin-` prefix), paths to plugin files, or plugin specification objects.

```ts
lessgo({
  plugins: [
    'clean-css',
    { name: 'autoprefix', options: 'last 2 versions' },
  ],
})
```

### `sourceMap`

Type: `boolean`
Default: `true` in development, `false` in production

Generate source maps.

```ts
lessgo({
  sourceMap: true,
})
```

### `include`

Type: `RegExp | string | string[]`
Default: `/\.less$/`

File patterns to include.

```ts
lessgo({
  include: /src\/.*\.less$/,
})
```

### `exclude`

Type: `RegExp | string | string[]`

File patterns to exclude.

```ts
lessgo({
  exclude: /node_modules/,
})
```

## TypeScript

This package includes TypeScript type definitions. Types are exported for all options:

```ts
import lessgo, { type LessgoPluginOptions } from '@lessgo/plugin-vite';

const options: LessgoPluginOptions = {
  compress: true,
  globalVars: {
    theme: '"dark"',
  },
};

export default defineConfig({
  plugins: [lessgo(options)],
});
```

## Why use this plugin?

- **Fast**: Powered by the Go-based lessgo compiler for high performance
- **Drop-in replacement**: Works with existing LESS files without modification
- **Full LESS support**: All LESS features including mixins, functions, and plugins
- **HMR support**: Automatic hot module replacement in development
- **Source maps**: Full source map support for debugging

## License

Apache-2.0
