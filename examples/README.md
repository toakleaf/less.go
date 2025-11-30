# less.go Examples

This directory contains complete, runnable examples demonstrating how to use the less.go library.

## Examples

### 1. Basic Compilation (`basic/`)

Demonstrates the simplest way to compile LESS to CSS.

```bash
cd basic
go run main.go
```

**Features shown:**
- Simple variable usage
- Mixins with parameters
- Nested selectors
- Color functions (`darken()`, `lighten()`)
- Parent selector (`&`)

### 2. Options Usage (`options/`)

Shows how to configure the LESS compiler with various options.

```bash
cd options
go run main.go
```

**Features shown:**
- CSS compression/minification
- Math operation modes (`Always`, `Parens`, `ParensDivision`)
- Global variables injection
- Modify variables (override existing)
- URL arguments for cache busting
- URL rewriting with rootpath
- Strict units mode
- Include paths for `@import`
- Combining multiple options

### 3. File Watcher (`watcher/`)

Automatically recompiles LESS files when they change.

```bash
cd watcher
go run main.go [watch-dir] [output-dir]

# Examples:
go run main.go                      # Watch current dir, output to ./css/
go run main.go ./styles ./dist      # Watch ./styles, output to ./dist
go run main.go -compress ./src      # Watch ./src with compression
```

**Features shown:**
- Directory watching with polling
- Automatic recompilation on file changes
- Partial file detection (files starting with `_`)
- Compilation timing
- Configurable watch interval

### 4. HTTP Server (`server/`)

Development server that compiles LESS to CSS on-the-fly.

```bash
cd server
go run main.go

# Then visit http://localhost:8080/styles.less
```

**Options:**
```bash
go run main.go -port 3000           # Use port 3000
go run main.go -root ./assets       # Serve from ./assets directory
go run main.go -compress            # Enable CSS minification
go run main.go -cache=false         # Disable caching
```

**Features shown:**
- On-demand LESS compilation
- ETag caching with 304 responses
- File modification tracking
- Compilation error handling (shown as CSS comments)
- Static file serving
- Compile time headers

### 5. JavaScript Plugins (`plugin/`)

Demonstrates how to extend LESS with custom JavaScript functions via plugins.

**Prerequisites:** Node.js must be installed and available in PATH.

```bash
# Run from the project root for best results
go run examples/plugin/main.go
```

**Features shown:**
- Loading plugins via `@plugin` directive
- Custom math functions (`double()`, `add()`, `sqrt-val()`)
- Custom color functions (`brand-color()`, `make-rgb()`)
- Custom string functions (`greet()`, `prefix()`)
- Plugin scoping (global vs local visibility)
- Combining plugins with variables and mixins

**Creating your own plugin:**

```javascript
// my-plugin.js
functions.add('triple', function(n) {
    return less.dimension(n.value * 3, n.unit);
});

functions.add('my-color', function() {
    return less.color([255, 100, 50]);
});
```

```less
// styles.less
@plugin "my-plugin.js";

.example {
    width: triple(10px);    // 30px
    color: my-color();      // #ff6432
}
```

## Quick Start

1. Navigate to the examples directory:
   ```bash
   cd examples
   ```

2. Create a test LESS file:
   ```bash
   cat > test.less << 'EOF'
   @primary: #007bff;

   .button {
       color: @primary;
       padding: 10px 20px;
       border-radius: 4px;

       &:hover {
           background: lighten(@primary, 40%);
       }
   }
   EOF
   ```

3. Run the basic example to see compilation:
   ```bash
   cd basic && go run main.go
   ```

4. Or start the development server:
   ```bash
   cd server && go run main.go
   # Visit http://localhost:8080/test.less
   ```

## API Reference

### Compile Function

```go
import less "github.com/toakleaf/less.go/packages/less/src/less/less_go"

result, err := less.Compile(lessSource, &less.CompileOptions{
    Filename:    "styles.less",
    Compress:    true,
    Math:        less.Math.Parens,
    StrictUnits: true,
    Paths:       []string{"/path/to/mixins"},
    GlobalVars: map[string]any{
        "primary-color": "#333",
    },
})

if err != nil {
    log.Fatal(err)
}

fmt.Println(result.CSS)
```

### CompileFile Function

```go
result, err := less.CompileFile("styles.less", &less.CompileOptions{
    Compress: true,
})
```

### CompileOptions

| Option | Type | Description |
|--------|------|-------------|
| `Filename` | `string` | Source filename for errors/source maps |
| `Compress` | `bool` | Enable CSS minification |
| `Math` | `MathType` | Math evaluation mode |
| `StrictUnits` | `bool` | Strict unit checking |
| `Paths` | `[]string` | Additional `@import` search paths |
| `GlobalVars` | `map[string]any` | Variables injected before parsing |
| `ModifyVars` | `map[string]any` | Variables overriding after parsing |
| `UrlArgs` | `string` | Query string for URLs |
| `Rootpath` | `string` | Base path for URL rewriting |
| `RewriteUrls` | `RewriteUrlsType` | URL rewriting mode |
| `EnableJavaScriptPlugins` | `bool` | Enable `@plugin` directive support (requires Node.js) |
| `JavascriptEnabled` | `bool` | Enable inline JavaScript evaluation (deprecated) |

### Math Modes

- `less.Math.Always` - Evaluate all math operations
- `less.Math.Parens` - Only evaluate math in parentheses
- `less.Math.ParensDivision` - Always evaluate, but division needs parens

### RewriteUrls Modes

- `less.RewriteUrls.Off` - Don't rewrite URLs
- `less.RewriteUrls.Local` - Rewrite local URLs only
- `less.RewriteUrls.All` - Rewrite all URLs
