# less.go

A complete Go port of [less.js](https://github.com/less/less.js) - the popular CSS preprocessor. This implementation maintains 1:1 functionality with less.js v4.2.2 while following Go idioms and conventions.

## Status

**Production Ready** (v1.0.0 - 2025-11-30)

- 191/191 integration tests passing (100%)
- 100 perfect CSS matches with less.js output
- 91 error handling tests correctly failing as expected
- 3,012 unit tests passing

## Installation

```bash
go get github.com/toakleaf/less.go/less
```

## Quick Start

### Basic Compilation

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

**Output:**
```css
.button {
  background: #4a90d9;
  color: white;
}
.button:hover {
  background: #3275b9;
}
```

### Compile from File

```go
result, err := less.CompileFile("styles.less", &less.CompileOptions{
    Compress: true,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.CSS)
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

## API Reference

### Compile Function

```go
func Compile(input string, options *CompileOptions) (*CompileResult, error)
```

The main entry point for compiling LESS source code to CSS.

### CompileFile Function

```go
func CompileFile(filename string, options *CompileOptions) (*CompileResult, error)
```

Convenience function to read and compile a LESS file.

### CompileResult

```go
type CompileResult struct {
    CSS     string   // Compiled CSS output
    Map     string   // Source map (if enabled)
    Imports []string // List of imported files
}
```

### CompileOptions

| Option | Type | Description |
|--------|------|-------------|
| `Paths` | `[]string` | Additional include paths for `@import` resolution |
| `Filename` | `string` | File name for error messages and source maps |
| `Compress` | `bool` | Enable CSS minification |
| `StrictUnits` | `bool` | Enable strict unit checking for math operations |
| `Math` | `MathType` | Math evaluation mode |
| `RewriteUrls` | `RewriteUrlsType` | URL rewriting behavior |
| `Rootpath` | `string` | Base path for URL rewriting |
| `UrlArgs` | `string` | Query string to append to URLs |
| `GlobalVars` | `map[string]any` | Variables injected before compilation |
| `ModifyVars` | `map[string]any` | Variables injected after (override existing) |
| `EnableJavaScriptPlugins` | `bool` | Enable JavaScript plugin support via Node.js |
| `JavascriptEnabled` | `bool` | Enable inline JavaScript evaluation |

### Math Modes

```go
less.Math.Always         // Always evaluate math expressions
less.Math.ParensDivision // Require parens for division (default)
less.Math.Parens         // Only evaluate math in parentheses
```

### URL Rewriting Modes

```go
less.RewriteUrls.Off   // No URL rewriting
less.RewriteUrls.Local // Rewrite local URLs only
less.RewriteUrls.All   // Rewrite all URLs
```

## Feature Parity with less.js

less.go implements **100% feature parity** with less.js v4.2.2:

### Core Features
- Variables and variable interpolation
- Nested rules and selectors
- Mixins (parametric, guards, closures, recursion)
- Namespacing
- Extend functionality
- Import system (including npm module resolution)
- Detached rulesets
- CSS guards
- Property merge (`+` and `+_`)

### Built-in Functions
All 60+ built-in functions are implemented:

| Category | Functions |
|----------|-----------|
| **Color** | `lighten`, `darken`, `saturate`, `desaturate`, `fade`, `fadein`, `fadeout`, `spin`, `mix`, `tint`, `shade`, `contrast`, `hue`, `saturation`, `lightness`, `alpha`, etc. |
| **Math** | `ceil`, `floor`, `sqrt`, `abs`, `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `pi`, `pow`, `mod`, `min`, `max`, `round`, `percentage` |
| **String** | `e`, `escape`, `replace`, `%`, `upper`, `lower` |
| **Type** | `isnumber`, `isstring`, `iscolor`, `iskeyword`, `isurl`, `ispixel`, `ispercentage`, `isem`, `isunit`, `isruleset` |
| **List** | `length`, `extract`, `range`, `each` |
| **Misc** | `color`, `image-width`, `image-height`, `data-uri`, `svg-gradient`, `get-unit`, `unit`, `convert`, `if`, `boolean` |
| **Blending** | `multiply`, `screen`, `overlay`, `softlight`, `hardlight`, `difference`, `exclusion`, `average`, `negation` |

### At-Rules
- `@media` (with query bubbling and merging)
- `@keyframes` / `@-webkit-keyframes`
- `@supports`
- `@font-face`
- `@container` (container queries)
- `@document`
- `@page`
- `@charset`
- `@namespace`

### Media Queries
- Full media query support
- Query bubbling out of nested rulesets
- Media query merging with detached rulesets
- Nested media query handling

## Plugin System

less.go provides full JavaScript plugin compatibility through a Node.js runtime bridge.

### Enabling Plugins

```go
result, err := less.Compile(source, &less.CompileOptions{
    EnableJavaScriptPlugins: true,
})
```

### Plugin Types

1. **Custom Functions** - Add custom LESS functions
2. **Visitors** - Transform the AST during compilation
3. **Pre-processors** - Transform source before parsing
4. **Post-processors** - Transform CSS after compilation
5. **File Managers** - Custom import resolution

### Using Plugins in LESS

```less
@plugin "my-plugin";

.example {
    color: my-custom-function();
}
```

### Writing JavaScript Plugins

```javascript
module.exports = {
    install: function(less, pluginManager, functions) {
        // Register custom function
        functions.add('pi', function() {
            return less.dimension(Math.PI);
        });

        // Register visitor
        pluginManager.addVisitor(new MyVisitor());

        // Register processors
        pluginManager.addPreProcessor(new MyPreProcessor(), 1000);
        pluginManager.addPostProcessor(new MyPostProcessor(), 1000);
    },

    minVersion: [2, 0, 0]
};
```

### Plugin Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Go Compiler                        │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────┐ │
│  │    Parser    │───>│  Evaluator   │───>│  ToCSS   │ │
│  └──────────────┘    └──────────────┘    └──────────┘ │
│          │                  │                  │        │
│          ▼                  ▼                  ▼        │
│  ┌─────────────────────────────────────────────────┐   │
│  │              Plugin Manager                      │   │
│  │  • Visitors  • Pre/Post Processors  • Functions │   │
│  └─────────────────────────────────────────────────┘   │
│                         │                              │
└─────────────────────────│──────────────────────────────┘
                          │ IPC (JSON/stdin-stdout)
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  Node.js Runtime                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │                  plugin-host.js                   │  │
│  │  • Plugin loading via require()                   │  │
│  │  • Function execution                             │  │
│  │  • Visitor callbacks                              │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

## Go-Specific Features

These features are unique to the Go implementation:

### 1. Type-Safe API

Unlike the JavaScript version's dynamic options, less.go provides strongly-typed configuration:

```go
options := &less.CompileOptions{
    Math:        less.Math.ParensDivision,  // Type-safe enum
    RewriteUrls: less.RewriteUrls.Local,    // Type-safe enum
    Compress:    true,
}
```

### 2. Lazy Plugin Bridge

The Node.js runtime is only started when plugins are actually used:

```go
// Node.js NOT started - no @plugin directives in source
result, err := less.Compile(source, &less.CompileOptions{
    EnableJavaScriptPlugins: true,
})

// Node.js started only when @plugin is encountered during parsing
```

### 3. IPC Mode Configuration

Control how the Go compiler communicates with the Node.js plugin runtime:

| Mode | Description | Best For |
|------|-------------|----------|
| **JSON** (default) | JSON over stdin/stdout | Many small function calls |
| **SHM** | Shared memory with binary protocol | Large AST transfers |

**Environment variable override:**
```bash
LESS_JS_IPC_MODE=json  # Default, 70% faster for typical usage
LESS_JS_IPC_MODE=shm   # Better for large data transfers
```

**Per-plugin configuration:**
```javascript
module.exports = {
    install: function(less, pm, functions) { ... },
    ipcMode: 'json'  // or 'shm'
};
```

### 4. Context-Free Functions

Plugin functions can be marked as context-free for better performance:

```go
// Context-free functions skip scope serialization
// Great for pure functions like math operations
opts = append(opts, WithContextFree())
```

### 5. Structured Errors

```go
type LessError struct {
    Type     string   // "Syntax", "Argument", etc.
    Message  string   // Error description
    Filename string   // File where error occurred
    Line     *int     // Line number (1-based)
    Column   int      // Column number
    Extract  []string // Context lines
}
```

### 6. Memory Optimization

Object pooling via `sync.Pool` for frequently allocated types:
- Rulesets
- Expressions
- Selectors
- Plugin scopes
- Math contexts

## Performance

### Comparison with less.js

| Metric | less.js | less.go | Notes |
|--------|---------|---------|-------|
| Cold start | ~993µs/file | ~931µs/file | Go ~6% faster |
| Warm (JIT) | ~428µs/file | ~883µs/file | JS JIT advantage |
| Memory/file | - | 0.56 MB | With 10k allocations |

### Benchmarking

```bash
# Suite-mode benchmark (realistic workload)
pnpm bench:compare:suite

# Per-file comparison
pnpm bench:compare

# Go-only benchmarks
pnpm bench:go:suite    # Suite mode
pnpm bench:go          # Per-file warm
pnpm bench:go:cold     # Per-file cold

# JavaScript benchmarks
pnpm bench:js
```

### Bootstrap 4 Compilation

Bootstrap 4's full LESS source compiles in approximately **1.2 seconds**.

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LESS_GO_DEBUG=1` | Enhanced debugging output |
| `LESS_GO_QUIET=1` | Suppress output, show summary only |
| `LESS_GO_DIFF=1` | Show CSS diffs for test failures |
| `LESS_GO_TRACE=1` | Show evaluation trace |
| `LESS_GO_JSON=1` | Output results as JSON |
| `LESS_JS_IPC_MODE` | Plugin IPC mode: `json` or `shm` |

## Testing

```bash
# Run all integration tests
pnpm test:go

# Run unit tests
pnpm test:go:unit

# Quick summary
LESS_GO_QUIET=1 pnpm test:go 2>&1 | tail -100

# Debug specific test
LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/<suite>/<testname>
```

## Examples

### Variables and Nesting

```less
@base-color: #4a90d9;
@spacing: 16px;

.card {
    padding: @spacing;
    background: white;

    .header {
        color: @base-color;
        border-bottom: 1px solid lighten(@base-color, 30%);
    }

    .body {
        padding: @spacing / 2;
    }
}
```

### Mixins with Guards

```less
.text-color(@bg) when (lightness(@bg) >= 50%) {
    color: black;
}

.text-color(@bg) when (lightness(@bg) < 50%) {
    color: white;
}

.dark-theme {
    @bg: #333;
    background: @bg;
    .text-color(@bg);
}
```

### Extend

```less
.button {
    display: inline-block;
    padding: 10px 20px;
    border-radius: 4px;
}

.primary-button {
    &:extend(.button);
    background: blue;
    color: white;
}
```

### Loops with `each()`

```less
@colors: red, green, blue;

each(@colors, {
    .color-@{value} {
        color: @value;
    }
});
```

### Detached Rulesets

```less
@mobile: ~"(max-width: 768px)";

@card-styles: {
    padding: 20px;
    background: white;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
};

.card {
    @card-styles();

    @media @mobile {
        padding: 10px;
    }
}
```

### Using Plugins

```less
@plugin "less-plugin-functions";

.example {
    // Use custom function from plugin
    color: my-custom-color(#ff0000, 50%);
}
```

## License

Apache License 2.0 - See [LICENSE](../LICENSE)

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## Related

- [Less.js](https://github.com/less/less.js) - Original JavaScript implementation
- [lesscss.org](http://lesscss.org) - LESS language documentation
- [CHANGELOG.md](../CHANGELOG.md) - Release history
- [BENCHMARKS.md](../BENCHMARKS.md) - Performance comparison details
