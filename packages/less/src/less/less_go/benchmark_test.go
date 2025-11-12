package less_go

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

// BenchmarkSuites defines test suites with their options and files to benchmark
// These are selected from our passing integration tests for fair comparison
var benchmarkTestFiles = []struct {
	suite   string
	options map[string]any
	folder  string
	files   []string
}{
	{
		suite:  "main",
		folder: "_main/",
		options: map[string]any{
			"relativeUrls":      true,
			"silent":           true,
			"javascriptEnabled": true,
		},
		files: []string{
			"calc",
			"charsets",
			"colors",
			"colors2",
			"comments",
			"css-escapes",
			"css-grid",
			"css-guards",
			"empty",
			"extend-chaining",
			"extend-clearfix",
			"extend-exact",
			"extend-media",
			"extend-nest",
			"extend-selector",
			"extend",
			"extract-and-length",
			"functions-each",
			"ie-filters",
			"import-inline",
			"import-interpolation",
			"import-once",
			"lazy-eval",
			"merge",
			"mixin-noparens",
			"mixins-closure",
			"mixins-guards-default-func",
			"mixins-guards",
			"mixins-important",
			"mixins-interpolated",
			"mixins-named-args",
			"mixins-nested",
			"mixins-pattern",
			"mixins",
			"no-output",
			"operations",
			"parse-interpolation",
			"permissive-parse",
			"property-accessors",
			"property-name-interp",
			"rulesets",
			"scope",
			"selectors",
			"strings",
			"variables-in-at-rules",
			"variables",
			"whitespace",
		},
	},
	{
		suite:  "namespacing",
		folder: "namespacing/",
		options: map[string]any{},
		files: []string{
			"namespacing-1",
			"namespacing-2",
			"namespacing-3",
			"namespacing-4",
			"namespacing-5",
			"namespacing-6",
			"namespacing-7",
			"namespacing-8",
			"namespacing-functions",
			"namespacing-media",
			"namespacing-operations",
		},
	},
	{
		suite:  "math-parens",
		folder: "math/strict/",
		options: map[string]any{
			"math": "parens",
		},
		files: []string{
			"css",
			"media-math",
			"mixins-args",
			"parens",
		},
	},
	{
		suite:  "math-parens-division",
		folder: "math/parens-division/",
		options: map[string]any{
			"math": "parens-division",
		},
		files: []string{
			"media-math",
			"mixins-args",
			"new-division",
			"parens",
		},
	},
	{
		suite:  "math-always",
		folder: "math/always/",
		options: map[string]any{
			"math": "always",
		},
		files: []string{
			"mixins-guards",
			"no-sm-operations",
		},
	},
	{
		suite:  "compression",
		folder: "compression/",
		options: map[string]any{
			"math":     "strict",
			"compress": true,
		},
		files: []string{
			"compression",
		},
	},
	{
		suite:  "units-strict",
		folder: "units/strict/",
		options: map[string]any{
			"math":        0,
			"strictUnits": true,
		},
		files: []string{
			"strict-units",
		},
	},
	{
		suite:  "units-no-strict",
		folder: "units/no-strict/",
		options: map[string]any{
			"math":        0,
			"strictUnits": false,
		},
		files: []string{
			"no-strict",
		},
	},
	{
		suite:  "rewrite-urls",
		folder: "rewrite-urls-all/",
		options: map[string]any{
			"rewriteUrls": "all",
		},
		files: []string{
			"rewrite-urls-all",
		},
	},
	{
		suite:  "include-path",
		folder: "include-path/",
		options: map[string]any{
			"paths": []string{"data/", "_main/import/"},
		},
		files: []string{
			"include-path",
		},
	},
}

// BenchmarkLessCompilation benchmarks the full compilation process (parse + eval)
// This is the primary benchmark for comparing with less.js
// Includes warmup runs for fair comparison with JIT-compiled JavaScript
func BenchmarkLessCompilation(b *testing.B) {
	testDataRoot := "../../../../test-data"
	lessRoot := filepath.Join(testDataRoot, "less")

	for _, suite := range benchmarkTestFiles {
		for _, fileName := range suite.files {
			testName := fmt.Sprintf("%s/%s", suite.suite, fileName)
			lessFile := filepath.Join(lessRoot, suite.folder, fileName+".less")

			b.Run(testName, func(b *testing.B) {
				// Read file once
				lessData, err := ioutil.ReadFile(lessFile)
				if err != nil {
					b.Skipf("Cannot read %s: %v", lessFile, err)
					return
				}

				// Prepare options
				options := make(map[string]any)
				for k, v := range suite.options {
					options[k] = v
				}
				options["filename"] = lessFile

				// Create factory once
				factory := Factory(nil, nil)

				// Warmup runs (matching JavaScript methodology)
				// JavaScript does 5 warmup runs before measuring to allow V8 JIT optimization
				// We do the same for fair comparison
				const warmupRuns = 5
				for i := 0; i < warmupRuns; i++ {
					_, _ = compileLessForTest(factory, string(lessData), options)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// Compile (parse + eval)
					_, compileErr := compileLessForTest(factory, string(lessData), options)
					if compileErr != nil {
						b.Fatalf("Compile error: %v", compileErr)
					}
				}
			})
		}
	}
}

// BenchmarkLessCompilationColdStart benchmarks cold-start performance (first iteration)
// This measures the real-world performance when the process is starting up
// No warmup runs are performed, capturing cache misses and initial allocations
func BenchmarkLessCompilationColdStart(b *testing.B) {
	testDataRoot := "../../../../test-data"
	lessRoot := filepath.Join(testDataRoot, "less")

	for _, suite := range benchmarkTestFiles {
		for _, fileName := range suite.files {
			testName := fmt.Sprintf("%s/%s", suite.suite, fileName)
			lessFile := filepath.Join(lessRoot, suite.folder, fileName+".less")

			b.Run(testName, func(b *testing.B) {
				// Read file once
				lessData, err := ioutil.ReadFile(lessFile)
				if err != nil {
					b.Skipf("Cannot read %s: %v", lessFile, err)
					return
				}

				// Prepare options
				options := make(map[string]any)
				for k, v := range suite.options {
					options[k] = v
				}
				options["filename"] = lessFile

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					// Create factory fresh each time to measure true cold start
					factory := Factory(nil, nil)

					// Compile (parse + eval) - cold start
					_, compileErr := compileLessForTest(factory, string(lessData), options)
					if compileErr != nil {
						b.Fatalf("Compile error: %v", compileErr)
					}

					b.StopTimer()
					// Small delay to ensure clean state between iterations
					b.StartTimer()
				}
			})
		}
	}
}

// BenchmarkLessParsing benchmarks the compilation (currently we don't have an easy way to separate parsing)
// This is kept for compatibility but measures full compilation like BenchmarkLessCompilation
func BenchmarkLessParsing(b *testing.B) {
	b.Skip("Skipping - parsing cannot be easily separated from evaluation in current implementation")
}

// BenchmarkLessEvaluation benchmarks the compilation (currently we don't have an easy way to separate evaluation)
// This is kept for compatibility but measures full compilation like BenchmarkLessCompilation
func BenchmarkLessEvaluation(b *testing.B) {
	b.Skip("Skipping - evaluation cannot be easily separated from parsing in current implementation")
}

// BenchmarkLargeSuite runs a comprehensive benchmark on multiple files at once
// This simulates a realistic workload where multiple different files are compiled
// sequentially, similar to a build process.
// Includes warmup runs for fair comparison with JIT-compiled JavaScript.
func BenchmarkLargeSuite(b *testing.B) {
	testDataRoot := "../../../../test-data"
	lessRoot := filepath.Join(testDataRoot, "less")

	// Collect all test data
	type testData struct {
		content  string
		options  map[string]any
		filename string
	}

	var tests []testData
	for _, suite := range benchmarkTestFiles {
		for _, fileName := range suite.files {
			lessFile := filepath.Join(lessRoot, suite.folder, fileName+".less")
			lessData, err := ioutil.ReadFile(lessFile)
			if err != nil {
				continue // Skip files that can't be read
			}

			// Prepare options
			options := make(map[string]any)
			for k, v := range suite.options {
				options[k] = v
			}
			options["filename"] = lessFile

			tests = append(tests, testData{
				content:  string(lessData),
				options:  options,
				filename: lessFile,
			})
		}
	}

	// Create factory once
	factory := Factory(nil, nil)

	// Warmup runs (matching JavaScript methodology)
	// JavaScript does 5 warmup runs before measuring to allow V8 JIT optimization
	// We do the same for fair comparison
	const warmupRuns = 5
	for i := 0; i < warmupRuns; i++ {
		for _, test := range tests {
			_, _ = compileLessForTest(factory, test.content, test.options)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			// Compile the LESS file
			_, compileErr := compileLessForTest(factory, test.content, test.options)
			// Ignore errors in batch benchmark to keep running, but we could track them
			_ = compileErr
		}
	}
}
