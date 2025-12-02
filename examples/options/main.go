// Package main demonstrates various compilation options available in less.go.
//
// This example shows how to configure the LESS compiler with different options
// including compression, math modes, URL rewriting, and variable injection.
// Run with: go run main.go
package main

import (
	"fmt"
	"log"

	less "github.com/toakleaf/less.go/less"
)

func main() {
	lessSource := `
@base-color: #333;
@spacing: 8px;

.container {
    color: @base-color;
    padding: @spacing;
    margin: @spacing * 2;
    width: 100px / 2;
    background: url("images/bg.png");
}
`

	// Example 1: Compression (minified output)
	fmt.Println("=== Example 1: Compressed Output ===")
	result, err := less.Compile(lessSource, &less.CompileOptions{
		Compress: true,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)
	fmt.Println()

	// Example 2: Math modes
	mathSource := `
.math-demo {
    // Math.Always: all math is evaluated
    width: 100px + 50px;
    height: 200px / 2;
    margin: 10px * 2;

    // Parentheses force evaluation in all modes
    padding: (20px / 4);
}
`
	fmt.Println("=== Example 2a: Math.Always (default) ===")
	result, err = less.Compile(mathSource, &less.CompileOptions{
		Math: less.Math.Always,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	fmt.Println("=== Example 2b: Math.Parens (only evaluate in parentheses) ===")
	result, err = less.Compile(mathSource, &less.CompileOptions{
		Math: less.Math.Parens,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	// Example 3: Global variables (injected before parsing)
	fmt.Println("=== Example 3: Global Variables ===")
	varSource := `
.themed {
    color: @theme-color;
    font-size: @base-size;
}
`
	result, err = less.Compile(varSource, &less.CompileOptions{
		GlobalVars: map[string]any{
			"theme-color": "#ff6600",
			"base-size":   "16px",
		},
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	// Example 4: Modify variables (override after parsing)
	fmt.Println("=== Example 4: Modify Variables (override) ===")
	modifySource := `
@brand-color: blue;  // This will be overridden

.brand {
    color: @brand-color;
}
`
	result, err = less.Compile(modifySource, &less.CompileOptions{
		ModifyVars: map[string]any{
			"brand-color": "green",
		},
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	// Example 5: URL arguments (cache busting)
	fmt.Println("=== Example 5: URL Arguments ===")
	urlSource := `
.background {
    background-image: url("sprite.png");
}
`
	result, err = less.Compile(urlSource, &less.CompileOptions{
		UrlArgs: "v=1.2.3",
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	// Example 6: URL rewriting with rootpath
	fmt.Println("=== Example 6: URL Rewriting ===")
	rewriteSource := `
.icon {
    background: url("icons/star.png");
}
`
	result, err = less.Compile(rewriteSource, &less.CompileOptions{
		Rootpath:    "/assets/",
		RewriteUrls: less.RewriteUrls.All,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	// Example 7: Strict units
	fmt.Println("=== Example 7: Strict Units Mode ===")
	unitsSource := `
.strict-demo {
    // With strict units, mixing incompatible units is an error
    // This example uses compatible units
    width: 100px + 50px;
    margin: 2em * 2;
}
`
	result, err = less.Compile(unitsSource, &less.CompileOptions{
		StrictUnits: true,
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)

	// Example 8: Include paths for @import resolution
	fmt.Println("=== Example 8: Include Paths ===")
	fmt.Println("Include paths allow @import to find files in additional directories:")
	fmt.Print(`
    result, err := less.Compile(source, &less.CompileOptions{
        Paths: []string{
            "/path/to/vendor/less",
            "/path/to/project/mixins",
        },
    })
`)

	// Example 9: Combining options
	fmt.Println("=== Example 9: Combined Options ===")
	combinedSource := `
.production-ready {
    color: @primary;
    padding: (10px * 2);
    background: url("bg.png");
}
`
	result, err = less.Compile(combinedSource, &less.CompileOptions{
		Compress: true,
		Math:     less.Math.Parens,
		UrlArgs:  "v=2.0.0",
		GlobalVars: map[string]any{
			"primary": "#007bff",
		},
	})
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}
	fmt.Println(result.CSS)
}
