# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-12-02

Initial release of **lessgo**, a complete Go port of [Less.js](https://github.com/less/less.js) v4.2.2.

### Highlights

- **191/191 integration tests passing** (100% success rate)
- **100 perfect CSS matches** with Less.js output
- **91 error handling tests** correctly failing as expected
- **3,012 unit tests passing**

### Added

#### Core Compiler
- Full LESS syntax parsing and compilation
- All built-in functions (60+)
- Mixins (parametric, guards, closures, recursion)
- Namespacing and scope management
- Extend functionality
- Import system (including npm module resolution)
- Variable interpolation
- Detached rulesets
- CSS guards
- Media query handling and bubbling
- URL rewriting
- Math operations (all modes: `always`, `parens`, `parens-division`)
- Compression output
- Source maps

#### JavaScript Integration
- Inline JavaScript evaluation via Node.js runtime
- Variable interpolation `@{varName}` in JavaScript expressions
- Variable access via `this.varName.toJS()` syntax
- Full JavaScript plugin support
- Plugin scope management
- Function registry integration
- Post-processor and pre-eval visitor support

#### CLI Tool (`lessc-go`)
- Drop-in replacement for `lessc`
- All standard Less.js CLI options
- Plugin loading support

#### npm Packages
- `lessgo` - Main package with automatic platform detection
- `@lessgo/darwin-arm64` - macOS Apple Silicon
- `@lessgo/darwin-x64` - macOS Intel
- `@lessgo/linux-x64` - Linux x64
- `@lessgo/linux-arm64` - Linux ARM64
- `@lessgo/win32-x64` - Windows x64
- `@lessgo/win32-arm64` - Windows ARM64

### Performance

- Native Go binary - no JavaScript runtime needed for core functionality
- ~2.2x faster than Less.js for typical workloads
- Efficient memory usage (~0.56 MB per file)
- No JIT warmup required - consistent performance from first run

---

**lessgo** is a complete Go port, not a fork. It shares no code with Less.js but maintains 100% compatibility through comprehensive testing against the original implementation.

[0.1.0]: https://github.com/toakleaf/less.go/releases/tag/v0.1.0
