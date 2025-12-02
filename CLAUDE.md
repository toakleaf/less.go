# Claude Code Context for less.go

This file provides context to Claude Code about the less.go project.

## Project Overview

A complete Go port of less.js - the popular CSS preprocessor. The port maintains 1:1 functionality with the original JavaScript implementation while following Go idioms.

**Status: Port Complete** (2025-11-30)
- 191/191 integration tests passing (100%)
- 100 perfect CSS matches with less.js output
- 91 error handling tests correctly failing as expected
- 3,012 unit tests passing

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
- Never modify original less.js source files

## Testing

**Prerequisites:**
```bash
pnpm install  # Required for npm module resolution tests
git submodule update --init  # Required for JS unit tests
```

**Commands:**
```bash
# Run all integration tests
pnpm -w test:go

# Run Go unit tests
pnpm -w test:go:unit

# Run JavaScript unit tests
pnpm test:js-unit

# Quick summary (recommended)
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100

# Debug a specific test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/<suite>/<testname> ./less

# See CSS diffs
LESS_GO_DIFF=1 pnpm -w test:go
```

**Environment Variables:**
- `LESS_GO_QUIET=1` - Show only summary
- `LESS_GO_DEBUG=1` - Enhanced debugging info
- `LESS_GO_DIFF=1` - Show CSS diffs for failing tests
- `LESS_GO_JSON=1` - Output results as JSON
- `LESS_GO_TRACE=1` - Show evaluation trace

**Test Categories:**
- **Perfect CSS Matches** - Tests producing identical CSS to less.js (100 tests)
- **Correctly Failed** - Error tests that properly fail as expected (91 tests)

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
- URL rewriting
- Math operations (all modes)
- Inline JavaScript evaluation (via Node.js IPC)
- Plugin system (custom functions, visitors, processors, file managers)
- Compression output
- Source maps

All features are fully implemented - no quarantined tests!

## Project Structure

- `less/` - Go implementation
- `reference/less.js/` - Original less.js source (git submodule, reference only)
- `test/js/` - Custom JavaScript unit tests for less.js
- `packages/test-data/` - Test fixtures shared by both implementations
- `cmd/lessc-go/` - CLI tool
- `examples/` - Example usage
- `.claude/` - Claude Code configuration and documentation
