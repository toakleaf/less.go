package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	less_go "github.com/toakleaf/less.go/packages/less/src/less/less_go"
)

func main() {
	// Create factory
	factory := less_go.Factory(nil, nil)

	// Read the test file
	lessFile := "/home/user/less.go/packages/test-data/less/rewrite-urls-all/rewrite-urls-all.less"
	lessContent, err := ioutil.ReadFile(lessFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Set up options
	options := map[string]any{
		"filename":    lessFile,
		"rewriteUrls": "all",
		"paths":       []string{filepath.Dir(lessFile)},
	}

	// Compile
	renderFunc := factory["render"].(func(string, ...any) any)
	result := renderFunc(string(lessContent), options)

	// Check result
	if resultMap, ok := result.(map[string]any); ok {
		if err, hasErr := resultMap["error"]; hasErr {
			fmt.Printf("Compilation failed: %v\n", err)
			return
		}
	}

	if resultStr, ok := result.(string); ok {
		fmt.Printf("Result:\n%s\n", resultStr)
	} else {
		fmt.Printf("Result type: %T, value: %+v\n", result, result)
	}

	// Read expected output
	cssFile := "/home/user/less.go/packages/test-data/css/rewrite-urls-all/rewrite-urls-all.css"
	expectedCSS, err := ioutil.ReadFile(cssFile)
	if err != nil {
		fmt.Printf("Error reading expected CSS: %v\n", err)
		return
	}

	fmt.Printf("\n\nExpected:\n%s\n", string(expectedCSS))
}
