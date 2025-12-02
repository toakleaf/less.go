// Package main demonstrates basic LESS to CSS compilation using less.go.
//
// This example shows the simplest way to compile LESS source code to CSS.
// Run with: go run main.go
package main

import (
	"fmt"
	"log"

	less "github.com/toakleaf/less.go/less"
)

func main() {
	// Example 1: Simple LESS compilation
	lessSource := `
@primary-color: #4a90d9;
@padding: 10px;

.button {
    color: @primary-color;
    padding: @padding;
    border: 1px solid darken(@primary-color, 10%);

    &:hover {
        background-color: lighten(@primary-color, 40%);
    }
}
`
	result, err := less.Compile(lessSource, nil)
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println("=== Example 1: Basic Compilation ===")
	fmt.Println(result.CSS)

	// Example 2: Using mixins
	lessWithMixins := `
.border-radius(@radius: 5px) {
    border-radius: @radius;
    -webkit-border-radius: @radius;
    -moz-border-radius: @radius;
}

.box {
    .border-radius(10px);
    padding: 20px;
}

.pill {
    .border-radius(999px);
}
`
	result2, err := less.Compile(lessWithMixins, nil)
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println("=== Example 2: Using Mixins ===")
	fmt.Println(result2.CSS)

	// Example 3: Nested selectors
	lessNested := `
nav {
    ul {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    li {
        display: inline-block;

        a {
            color: #333;
            text-decoration: none;

            &:hover {
                color: #007bff;
            }
        }
    }
}
`
	result3, err := less.Compile(lessNested, nil)
	if err != nil {
		log.Fatalf("Compilation error: %v", err)
	}

	fmt.Println("=== Example 3: Nested Selectors ===")
	fmt.Println(result3.CSS)
}
