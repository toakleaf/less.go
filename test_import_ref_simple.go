package main

import (
	"fmt"
	"os"

	less "github.com/toakleaf/less.go/packages/less/src/less/less_go"
)

func main() {
	// Simple test case
	input := `
@import (reference) "test-ref.less";

.b {
  .z();
}
`

	// Create test-ref.less file
	err := os.WriteFile("test-ref.less", []byte(`.z {
  color: red;
}`), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating test file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove("test-ref.less")

	// Compile
	factory := less.Factory(nil, nil)
	fmt.Println("Factory created")
	if render, ok := factory["render"].(func(string, ...any) any); ok {
		fmt.Println("Found render function")
			options := map[string]any{
				"paths": []string{"."},
			}
			result := render(input, options)
			fmt.Printf("Result type: %T\n", result)
			fmt.Printf("Result: %+v\n", result)

			if resultMap, ok := result.(map[string]any); ok {
				if css, ok := resultMap["css"].(string); ok {
					fmt.Println("=== OUTPUT ===")
					fmt.Println(css)
					fmt.Println("\n=== EXPECTED ===")
					fmt.Println(".b {\n  color: red;\n}")
					fmt.Println("\n=== ANALYSIS ===")
					fmt.Println("Should NOT show .z selector")
					fmt.Println("Should ONLY show .b selector with color from .z() mixin")
			} else if err, ok := resultMap["error"]; ok {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
