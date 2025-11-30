// Package main demonstrates JavaScript plugin support in less.go.
//
// This example shows how to use JavaScript plugins that extend LESS with
// custom functions. Plugins are executed via a Node.js runtime.
//
// Prerequisites:
//   - Node.js must be installed and available in PATH
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
	// Get the directory where this Go file is located
	// We need absolute paths for reliable plugin loading
	exampleDir, err := filepath.Abs(".")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Use absolute path for the plugin to ensure reliable loading
	pluginPath, err := filepath.Abs(filepath.Join(exampleDir, "sample-plugin.js"))
	if err != nil {
		log.Fatalf("Failed to get absolute plugin path: %v", err)
	}

	// Check if the plugin file exists
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		log.Fatalf("Plugin file not found: %s\nMake sure to run this example from the examples/plugin directory", pluginPath)
	}

	// Example 1: Using custom plugin functions
	fmt.Println("=== Example 1: Custom Plugin Functions ===")

	lessSource := fmt.Sprintf(`
// Load the custom plugin (using absolute path for reliability)
@plugin "%s";

// Use custom functions from the plugin
.math-demo {
    // double(n) multiplies a value by 2
    doubled: double(21px);

    // add(a, b) adds two numbers
    sum: add(10, 5);

    // sqrt-val(n) returns square root
    root: sqrt-val(16);
}

.color-demo {
    // brand-color() returns a predefined brand color
    background: brand-color();

    // make-rgb(r, g, b) creates a color
    custom: make-rgb(255, 128, 0);
}

.string-demo {
    // greet(name) returns a greeting string
    content: greet("World");

    // prefix(str) adds a vendor prefix
    custom-prop: prefix("transform");
}
`, pluginPath)

	// Note: Filename must be set to a path in the same directory as the plugin
	// for relative plugin paths to resolve correctly
	result, err := less.Compile(lessSource, &less.CompileOptions{
		Filename:                filepath.Join(exampleDir, "styles.less"),
		EnableJavaScriptPlugins: true,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println(result.CSS)

	// Example 2: Plugin with scoped visibility
	fmt.Println("=== Example 2: Plugin Scoping ===")

	scopedSource := fmt.Sprintf(`
// Plugin loaded at root level - available everywhere
@plugin "%s";

.global-scope {
    value: double(5px);
}

.namespace {
    // Plugins loaded inside a block are scoped to that block
    .nested {
        // Inherits parent scope
        value: add(3, 7);
    }
}
`, pluginPath)

	result2, err := less.Compile(scopedSource, &less.CompileOptions{
		Filename:                filepath.Join(exampleDir, "scoped.less"),
		EnableJavaScriptPlugins: true,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println(result2.CSS)

	// Example 3: Combining plugins with LESS features
	fmt.Println("=== Example 3: Plugins with Variables and Mixins ===")

	combinedSource := fmt.Sprintf(`
@plugin "%s";

// Use plugin function results in variables
@brand: brand-color();
@spacing: double(8px);

// Mixin using plugin functions
.button-style(@size) {
    padding: double(@size);
    background: @brand;
    border-radius: add(2, 2) * 1px;
}

.button {
    .button-style(5px);
}

.container {
    padding: @spacing;
    margin: double(@spacing);
}
`, pluginPath)

	result3, err := less.Compile(combinedSource, &less.CompileOptions{
		Filename:                filepath.Join(exampleDir, "combined.less"),
		EnableJavaScriptPlugins: true,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println(result3.CSS)

	// Print summary
	fmt.Println("=== Plugin System Summary ===")
	fmt.Println(`
Plugin support is fully functional! You can:
  - Load plugins via @plugin directive
  - Call custom functions with arguments
  - Return dimensions, colors, strings, and keywords
  - Maintain state across function calls
  - Use plugin results in variables and expressions

Creating your own plugin:

  // my-plugin.js
  functions.add('triple', function(n) {
      return less.dimension(n.value * 3, n.unit);
  });

  functions.add('my-color', function() {
      return less.color([255, 100, 50]);
  });

Then use it:

  @plugin "my-plugin.js";
  .example {
      width: triple(10px);    // 30px
      color: my-color();      // #ff6432
  }
`)
}
