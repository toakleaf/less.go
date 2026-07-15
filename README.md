# less.go

An attempt at a complete Go port of [Less.js](https://github.com/less/less.js) - the popular CSS preprocessor. This implementation aims to maintain 100% feature parity with Less.js v4.4.2 while providing the performance benefits of a native Go binary.

## Status

**Current Release** - Fully compatible with Less.js v4.4.2

- 195/195 integration tests passing (100%)
- 100 perfect CSS matches with Less.js output
- 91 error handling tests correctly failing as expected
- All unit tests passing

## Installation

### Via npm (Recommended)

Install the pre-built binary for your platform:

```bash
npm install lessgo
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

less.go provides native binary performance without requiring a JavaScript runtime. Current benchmarks cover 212 integration files and use Less.js v4.4.2 as the comparison point. These results were measured on an Apple M1 Max with Go 1.24.5 and Node.js 24.4.1:

| Metric | Less.js | less.go | Difference |
|--------|---------|---------|------------|
| Fresh-process full build | 207.08ms | 117.46ms | Go 1.8x faster |
| Cold per-file average | 799.19µs | 505.55µs | Go 1.6x faster |
| Warm per-file average | 402.75µs | 476.68µs | Go 1.18x slower |
| Memory per file | - | 0.30 MB | ~5,408 allocations |

- **CLI/build workloads are faster in Go** - The fresh-process 212-file workload is 1.8x faster than Less.js
- **Peak-JIT microbenchmarks are close** - V8 retains an 18% lead when each file is compiled repeatedly in isolation
- **GC tuning closes the aggregate warm gap** - `GOGC=200` reduced the same-process Go suite from about 112ms to 97ms in local testing, slightly ahead of the 103ms warmed Node suite, at the cost of a larger heap target

`GOGC` is process-wide, so the library does not change it automatically. Throughput-oriented services can start with `GOGC=200`, measure their own memory envelope, and consider `GOGC=400` only when latency matters more than memory.

Run benchmarks yourself:
```bash
pnpm bench:compare        # Warm/cold per-file comparison across 212 files
pnpm bench:compare:suite  # Realistic full-suite comparison
pnpm bench:go:public      # Main exported Compile API
pnpm bench:go:suite:gc200 # Suite with the throughput-oriented GC setting
```

## Features

less.go implements **100% feature parity** with Less.js v4.4.2:

- **Variables** - `@primary: #333;`
- **Nesting** - Nested rules and selectors
- **Mixins** - Parametric, guards, closures, recursion
- **Extend** - `&:extend(.class)`
- **Import** - Including npm module resolution
- **Functions** - All 60+ built-in functions
- **Detached Rulesets** - Reusable rule blocks
- **CSS Guards** - Conditional CSS
- **Media Query Bubbling** - Automatic media query handling
- **Container Queries** - `@container` with size and style queries
- **CSS Layers** - `@layer` at-rule and import with layer()
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

# Or if already cloned, initialize submodules
git submodule update --init --recursive

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
