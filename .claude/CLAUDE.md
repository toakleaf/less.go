# Claude Code Context for less.go

This file provides context to Claude Code about the less.go project.

## Project Overview

A complete Go port of Less.js - the popular CSS preprocessor. The port maintains 1:1 functionality with the original JavaScript implementation while following Go idioms.

**Status: Port Complete** (2025-11-30)
- Tracking Less.js v4.4.2 (latest release, October 2025)
- 191/195 integration tests passing (98%)
- 99 perfect CSS matches with Less.js output
- 91 error handling tests correctly failing as expected
- 3,012 unit tests passing

**Pending v4.4.2 Compatibility Fixes** (4 tests):
- `layer.less` - Extra space in `layer()` syntax, parent selector in nested @layer
- `starting-style.less` - Nested @starting-style incorrectly bubbling to root
- `container.less` - Extra space in `scroll-state()` syntax
- `colors.less` - Color channel identifiers (l,c,h,r,g,b,s) not parsed as operands

## Rules References

### Always Applied
@.cursor/rules/project-goals-and-conventions.mdc

### Language-Specific
- Go files (*.go): @.cursor/rules/go-lang-rules.mdc
- JavaScript files (*.js): @.cursor/rules/javascript-rules.mdc

## Core Principles

- Maintain 1:1 functionality between JavaScript and Go versions
- Avoid external dependencies where possible
- Follow language-specific idioms and conventions
- Never modify original Less.js source files in reference/

## Project Structure

```
less.go/
├── less/               # Go implementation (core library)
├── cmd/lessc-go/       # CLI tool
├── testdata/           # Test fixtures (LESS files, expected CSS)
├── test/js/            # JavaScript unit tests
├── npm/                # NPM package templates (platform-specific)
├── reference/less.js/  # Original Less.js (git submodule, reference only)
├── examples/           # Usage examples
├── scripts/            # Build and test scripts
├── packages/           # Monorepo packages
└── .claude/            # Claude Code configuration and documentation
```

## Testing

**Prerequisites:**
```bash
pnpm install  # Required for npm module resolution tests
git submodule update --init  # Required for JS unit tests
```

**Commands:**
```bash
# Run all integration tests
pnpm test:go

# Run Go unit tests
pnpm test:go:unit

# Run JavaScript unit tests
pnpm test:js-unit

# Quick summary (recommended)
LESS_GO_QUIET=1 pnpm test:go 2>&1 | tail -100

# Debug a specific test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/<suite>/<testname> ./less

# See CSS diffs
LESS_GO_DIFF=1 pnpm test:go
```

**Environment Variables:**
- `LESS_GO_QUIET=1` - Show only summary
- `LESS_GO_DEBUG=1` - Enhanced debugging info
- `LESS_GO_DIFF=1` - Show CSS diffs for failing tests
- `LESS_GO_JSON=1` - Output results as JSON
- `LESS_GO_TRACE=1` - Show evaluation trace
- `LESS_GO_SKIP_CUSTOM=1` - Skip custom integration tests
- `LESS_GO_CUSTOM_ONLY=1` - Run only custom integration tests

**Test Categories:**
- **Perfect CSS Matches** - Tests producing identical CSS to Less.js
- **Correctly Failed** - Error tests that properly fail as expected
- **Custom Tests** - User-defined tests separate from Less.js originals

### Custom Integration Tests

Add your own integration tests by dropping `.less` and `.css` file pairs into:
- `testdata/less/custom/` - Input LESS files
- `testdata/css/custom/` - Expected CSS output files

Files must have matching names (e.g., `my-test.less` and `my-test.css`).

```bash
# Run only custom tests
LESS_GO_CUSTOM_ONLY=1 go test ./less -v -run TestIntegrationSuite -timeout 5m

# Run all tests except custom
LESS_GO_SKIP_CUSTOM=1 pnpm test:go:all

# Debug a specific custom test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/custom/<testname> ./less
```

Custom tests run with default options: `relativeUrls: true`, `silent: true`, `javascriptEnabled: true`.

## Benchmarking

```bash
# Compare Go vs JavaScript performance
pnpm bench:compare

# JavaScript only
pnpm bench:js

# Go only
pnpm bench:go:suite

# Profile Go implementation
pnpm bench:profile
```

**Performance Notes:**
- Bootstrap4 compiles in ~1.2s with the Go port
- JSON IPC mode is 70% faster than SHM for plugin function calls
- Default is JSON mode (optimal for many small function calls)
- Per-plugin IPC mode configuration available via `ipcMode` in plugin exports
- Environment override: `LESS_JS_IPC_MODE=json` or `LESS_JS_IPC_MODE=shm`

**Documentation:**
- `BENCHMARKS.md` - Main benchmarking guide
- `.claude/benchmarks/BENCHMARKING_GUIDE.md` - Detailed guide
- `.claude/benchmarks/PERFORMANCE_ANALYSIS.md` - Performance analysis

## Key Features Implemented

- Full LESS syntax parsing and compilation
- All built-in functions
- Mixins (parametric, guards, closures, recursion)
- Namespacing
- Extend functionality
- Import system (including npm module resolution)
- Variable interpolation
- Detached rulesets
- CSS guards
- Media query handling
- Container queries (@container with size/style queries)
- CSS layers (@layer at-rule)
- @starting-style at-rule
- URL rewriting
- Math operations (all modes)
- Inline JavaScript evaluation (via Node.js IPC)
- Plugin system (custom functions, visitors, processors, file managers)
- Compression output
- Source maps

All core features are implemented. See "Pending v4.4.2 Compatibility Fixes" above for minor issues being addressed.
