package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go"
)

func main() {
	// Compile import-reference.less
	lessFile := "packages/test-data/less/_main/import-reference.less"

	data, err := os.ReadFile(lessFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	options := less_go.Options{
		Paths: []string{
			filepath.Dir(lessFile),
		},
		Filename: lessFile,
	}

	result, err := less_go.Render(string(data), options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error compiling: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== ACTUAL OUTPUT ===")
	fmt.Println(result.CSS)
	fmt.Println("\n=== EXPECTED OUTPUT ===")
	expectedFile := "packages/test-data/css/_main/import-reference.css"
	expected, err := os.ReadFile(expectedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading expected: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(expected))
}
