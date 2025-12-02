# Contributing to less.go

Thank you for your interest in contributing to less.go! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- **Go 1.21+** - The Go implementation
- **Node.js 18+** - Required for JavaScript plugin support and tests
- **pnpm** - Package manager for the monorepo

### Clone with Submodules

less.go uses a git submodule to reference the original Less.js source for testing:

```bash
# Clone with submodules (recommended)
git clone --recurse-submodules https://github.com/toakleaf/less.go.git
cd less.go

# Or if you already cloned without submodules
git submodule update --init
```

### Install Dependencies

```bash
pnpm install
```

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
└── packages/           # Monorepo packages
```

### Key Directories

- **`less/`** - The core Go library. All compiler logic lives here.
- **`testdata/`** - Test fixtures including LESS source files and expected CSS output.
- **`reference/less.js/`** - Git submodule pointing to the original Less.js. **Never modify files here** - it's reference only for running comparison tests.
- **`npm/`** - Platform-specific NPM package templates. The build process copies the compiled binary into these packages.

## Submodule Relationship

The `reference/less.js/` submodule serves as a reference implementation:

1. **Testing** - We compare less.go output against Less.js output to ensure compatibility
2. **Benchmarking** - Performance comparisons between Go and JavaScript implementations
3. **Reference** - When implementing features, the original source serves as documentation

**Important:** Never commit changes to the submodule. If you need to update the reference version:

```bash
cd reference/less.js
git fetch
git checkout <tag-or-commit>
cd ../..
git add reference/less.js
git commit -m "chore: update less.js reference to <version>"
```

## Running Tests

### Integration Tests

Integration tests compile LESS files and compare output against Less.js:

```bash
# Run all integration tests
pnpm test:go

# Quick summary (recommended for development)
LESS_GO_QUIET=1 pnpm test:go 2>&1 | tail -100

# Debug a specific test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/<suite>/<testname> ./less

# See CSS diffs for failures
LESS_GO_DIFF=1 pnpm test:go
```

### Unit Tests

```bash
# Go unit tests
pnpm test:go:unit

# JavaScript unit tests
pnpm test:js-unit

# All tests
pnpm test
```

### Test Environment Variables

| Variable | Purpose |
|----------|---------|
| `LESS_GO_QUIET=1` | Show only summary |
| `LESS_GO_DEBUG=1` | Enhanced debugging |
| `LESS_GO_DIFF=1` | Show CSS diffs |
| `LESS_GO_TRACE=1` | Show evaluation trace |
| `LESS_GO_JSON=1` | Output as JSON |

## NPM Package Development

less.go is distributed via npm with platform-specific packages:

```
npm/
├── less.go/              # Main package (installs correct platform binary)
├── less.go-darwin-arm64/ # macOS Apple Silicon
├── less.go-darwin-x64/   # macOS Intel
├── less.go-linux-arm64/  # Linux ARM64
├── less.go-linux-x64/    # Linux x64
├── less.go-win32-arm64/  # Windows ARM64
└── less.go-win32-x64/    # Windows x64
```

### Building for NPM

```bash
# Cross-compile for all platforms
make cross-compile

# Build for current platform only
make build
```

### Local Testing

```bash
# Link for local testing
cd npm/less.go
npm link

# Test in another project
npm link @toakleaf/less.go
npx lessc-go --version
```

## Pull Request Guidelines

### Before Submitting

1. **Run tests** - Ensure all tests pass
   ```bash
   pnpm test
   ```

2. **Check for regressions** - Compare test metrics before and after
   ```bash
   LESS_GO_QUIET=1 pnpm test:go 2>&1 | grep "OVERALL SUCCESS"
   ```

3. **Format code** - Go code should be formatted
   ```bash
   go fmt ./...
   ```

### PR Checklist

- [ ] Tests pass locally
- [ ] No regressions in test metrics
- [ ] Code is formatted
- [ ] Commit messages are clear and descriptive
- [ ] Documentation updated if needed

### Commit Message Format

Use clear, descriptive commit messages:

```
<type>: <description>

[optional body]
```

Types:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `test` - Tests
- `refactor` - Code refactoring
- `perf` - Performance improvement
- `chore` - Maintenance

Examples:
```
feat: add support for container queries
fix: correct color function alpha handling
docs: update installation instructions
```

## Reporting Issues

### Before Opening an Issue

1. **Search existing issues** - Check if already reported
2. **Test with latest version** - Ensure the issue still exists
3. **Create minimal reproduction** - Isolate the problem

### Issue Template

```markdown
**Description**
[Clear description of the issue]

**LESS Input**
```less
// Minimal LESS that reproduces the issue
```

**Expected Output**
```css
// What Less.js produces
```

**Actual Output**
```css
// What less.go produces
```

**Version**
- less.go: [version]
- OS: [operating system]
```

## Feature Requests

When requesting new features:

1. Check if the feature exists in Less.js
2. Provide use cases
3. Consider if it could be a plugin instead

## Questions?

- Open a [GitHub Discussion](https://github.com/toakleaf/less.go/discussions)
- Check [Less.js documentation](http://lesscss.org) for language questions

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
