// Package main demonstrates JavaScript plugin support in less.go.
//
// This example shows how the @plugin directive works and the plugin
// infrastructure. Plugins are loaded via a Node.js runtime.
//
// Prerequisites:
//   - Node.js must be installed and available in PATH
//
// IMPORTANT: The plugin system in less.go is still under development.
// While @plugin directives are parsed and plugins are loaded, custom
// plugin functions are not yet being executed. This example demonstrates
// the infrastructure that's in place and the intended usage patterns.
//
// Usage:
//
//	cd examples/plugin
//	go run main.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	less "github.com/toakleaf/less.go/packages/less/src/less/less_go"
)

func main() {
	// Get the directory where this example is located
	exampleDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	pluginPath := filepath.Join(exampleDir, "sample-plugin.js")

	// Check if the plugin file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		log.Fatalf("Plugin file not found: %s\nMake sure to run this example from the examples/plugin directory", pluginPath)
	}

	// Demonstrate that @plugin directive is parsed and plugins load
	fmt.Println("=== Plugin Loading Demo ===")
	fmt.Println("The @plugin directive is parsed and plugins are loaded.")
	fmt.Println("Custom function execution is still under development.")
	fmt.Println()

	lessSource := fmt.Sprintf(`
// The @plugin directive loads a JavaScript plugin
@plugin "%s";

// Built-in LESS functions work normally
.builtin-functions {
    // pi() is a built-in LESS function
    pi-value: pi();

    // Built-in color functions
    color: lighten(#4a90d9, 20%%);

    // Built-in math
    calc: (100px / 2);
}

// Plugin functions are parsed but not yet executed
// Once the plugin system is complete, these will work:
.plugin-functions {
    // Will output the function call as-is for now
    custom: my-plugin-function();
}
`, pluginPath)

	result, err := less.Compile(lessSource, &less.CompileOptions{
		Filename:                "styles.less",
		EnableJavaScriptPlugins: true,
		Paths:                   []string{exampleDir},
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println("Compiled CSS:")
	fmt.Println(result.CSS)

	// Show built-in functions that work
	fmt.Println("=== Built-in Functions Demo ===")
	fmt.Println("These LESS built-in functions work without plugins:")
	fmt.Println()

	builtinSource := `
// Math functions
.math {
    pi: pi();
    percentage: percentage(0.5);
    round: round(1.67);
    ceil: ceil(2.4);
    floor: floor(2.6);
    sqrt: sqrt(25);
    abs: abs(-5);
    min: min(3, 5, 1);
    max: max(3, 5, 1);
    mod: mod(5, 2);
}

// Color functions
.colors {
    lighten: lighten(#4a90d9, 20%);
    darken: darken(#4a90d9, 20%);
    saturate: saturate(#4a90d9, 20%);
    desaturate: desaturate(#4a90d9, 20%);
    fade: fade(#4a90d9, 50%);
    mix: mix(#ff0000, #0000ff, 50%);
    spin: spin(#4a90d9, 30);
}

// String functions
.strings {
    escape: escape("hello world");
    replace: replace("hello", "l", "L");
}

// Type functions
.types {
    is-color: iscolor(#fff);
    is-number: isnumber(42);
    is-string: isstring("hello");
    is-pixel: ispixel(100px);
    is-percentage: ispercentage(50%);
}
`

	result2, err := less.Compile(builtinSource, nil)
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println(result2.CSS)

	// Print plugin status
	fmt.Println("=== Plugin System Status ===")
	fmt.Println(`
Current Implementation:
  [x] @plugin directive parsing
  [x] Plugin file loading via Node.js
  [x] Plugin function registration
  [ ] Custom function execution (in progress)
  [ ] Functions with arguments
  [ ] Color return values
  [ ] Visitor plugins

For now, use the many built-in LESS functions which cover most needs:
  - Math: pi(), round(), ceil(), floor(), sqrt(), abs(), min(), max()
  - Colors: lighten(), darken(), saturate(), fade(), mix(), spin()
  - Strings: escape(), replace(), e()
  - Type checks: iscolor(), isnumber(), isstring(), etc.

Full plugin support is being actively developed. Check the less.go
repository for updates.
`)
}
