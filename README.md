# less.go

A complete Go port of [Less.js](https://github.com/less/less.js) - the popular CSS preprocessor. This implementation maintains 100% feature parity with Less.js v4.2.2 while providing the performance benefits of a native Go binary.

## Status

**Production Ready** (v1.0.0)

- 191/191 integration tests passing (100%)
- 100 perfect CSS matches with Less.js output
- 91 error handling tests correctly failing as expected
- 3,012 unit tests passing

## Installation

### Via npm (Recommended)

Install the pre-built binary for your platform:

```bash
npm install @toakleaf/less.go
```

This automatically installs the correct binary for your operating system and architecture.

### Via Go

```bash
go install github.com/toakleaf/less.go/cmd/lessc-go@latest
```

Or add the library to your Go project:

```bash
go get github.com/toakleaf/less.go/less
```

## CLI Usage

```bash
# Basic compilation
npx lessc-go input.less output.css

# With compression
npx lessc-go --compress input.less output.css

# Read from stdin, write to stdout
cat input.less | npx lessc-go -

# With source map
npx lessc-go --source-map input.less output.css

# Include paths for @import resolution
npx lessc-go --include-path=./mixins:./node_modules input.less output.css
```

### CLI Options

| Option | Description |
|--------|-------------|
| `--compress` | Minify output CSS |
| `--source-map` | Generate source map |
| `--include-path=PATHS` | Colon-separated paths for `@import` resolution |
| `--global-var='VAR=VALUE'` | Define global variables |
| `--modify-var='VAR=VALUE'` | Override variables |
| `--strict-units` | Enable strict unit checking |
| `--math=MODE` | Math mode: `always`, `parens`, `parens-division` |
| `--rootpath=PATH` | Base path for URL rewriting |
| `--rewrite-urls=MODE` | URL rewriting: `off`, `local`, `all` |
| `--js` | Enable inline JavaScript evaluation |
| `--plugin` | Enable JavaScript plugin support |

## Library Usage (Go)

```go
package main

import (
    "fmt"
    "log"

    less "github.com/toakleaf/less.go/less"
)

func main() {
    source := `
        @primary: #4a90d9;

        .button {
            background: @primary;
            color: white;
            &:hover {
                background: darken(@primary, 10%);
            }
        }
    `

    result, err := less.Compile(source, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.CSS)
}
```

### With Options

```go
result, err := less.Compile(source, &less.CompileOptions{
    Filename:    "styles.less",
    Compress:    true,
    StrictUnits: true,
    Math:        less.Math.ParensDivision,
    Paths:       []string{"./imports", "./node_modules"},
    GlobalVars: map[string]any{
        "theme-color": "#ff6600",
    },
})
```

## Performance

less.go provides native binary performance without requiring a JavaScript runtime:

| Metric | Less.js | less.go | Notes |
|--------|---------|---------|-------|
| Cold start | ~993µs/file | ~931µs/file | Go ~6% faster |
| No JIT warmup required | - | - | Consistent performance from first run |
| Memory efficiency | - | 0.56 MB/file | Efficient memory usage |

Bootstrap 4's full LESS source compiles in approximately **1.2 seconds**.

## Features

less.go implements **100% feature parity** with Less.js v4.2.2:

- **Variables** - `@primary: #333;`
- **Nesting** - Nested rules and selectors
- **Mixins** - Parametric, guards, closures, recursion
- **Extend** - `&:extend(.class)`
- **Import** - Including npm module resolution
- **Functions** - All 60+ built-in functions
- **Detached Rulesets** - Reusable rule blocks
- **CSS Guards** - Conditional CSS
- **Media Query Bubbling** - Automatic media query handling
- **Property Merge** - `+` and `+_` operators
- **Compression** - CSS minification
- **Source Maps** - Full source map support
- **JavaScript Plugins** - Custom functions via Node.js bridge

## Project Structure

```
less.go/
├── less/              # Go implementation (core library)
├── cmd/lessc-go/      # CLI tool
├── testdata/          # Test fixtures
├── test/js/           # JavaScript unit tests
├── npm/               # NPM package templates
├── reference/less.js/ # Original Less.js (git submodule, reference only)
├── examples/          # Usage examples
└── scripts/           # Build and test scripts
```

## Development

### Prerequisites

- Go 1.21+
- Node.js 18+ (for JavaScript plugin support and tests)
- pnpm

### Setup

```bash
# Clone with submodules
git clone --recurse-submodules https://github.com/toakleaf/less.go.git
cd less.go

# Install dependencies
pnpm install
```

### Running Tests

```bash
# Run all integration tests
pnpm test:go

# Run Go unit tests
pnpm test:go:unit

# Run JavaScript unit tests
pnpm test:js-unit

# Quick summary (recommended)
LESS_GO_QUIET=1 pnpm test:go 2>&1 | tail -100
```

### Benchmarking

```bash
# Compare Go vs JavaScript performance
pnpm bench:compare

# Go benchmarks
pnpm bench:go:suite
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:

- Setting up the development environment
- Running tests
- Submitting pull requests

## Related Projects

- [Less.js](https://github.com/less/less.js) - Original JavaScript implementation
- [lesscss.org](http://lesscss.org) - LESS language documentation

## License

Apache License 2.0 - See [LICENSE](LICENSE)

---

**less.go** is a complete Go port, not a fork. It shares no code with Less.js but maintains 100% compatibility through comprehensive testing against the original implementation.
