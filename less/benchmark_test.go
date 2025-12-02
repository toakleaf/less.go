package less_go

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

// BenchmarkSuites defines test suites with their options and files to benchmark
// These are selected from our passing integration tests for fair comparison
//
// PURE GO TESTS: These tests do NOT require Node.js/plugin support and can be
// benchmarked in pure Go mode for accurate Go vs JavaScript comparison.
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
			"relativeUrls":       true,
			"silent":             true,
			"javascriptEnabled":  true,
		},
		// All passing _main tests EXCEPT those requiring plugins/Node.js or network
		// Excluded: import, import-module, javascript, plugin, plugin-module, plugin-preeval
		// Also excluded: import-remote (makes network requests to cdn.jsdelivr.net)
		files: []string{
			"calc",
			"charsets",
			"colors",
			"colors2",
			"comments",
			"comments2",
			"container",
			"css-3",
			"css-escapes",
			"css-grid",
			"css-guards",
			"detached-rulesets",
			"directives-bubling",
			"empty",
			"extend",
			"extend-chaining",
			"extend-clearfix",
			"extend-exact",
			"extend-media",
			"extend-nest",
			"extend-selector",
			"extract-and-length",
			"functions",
			"functions-each",
			"ie-filters",
			"impor",
			"import-inline",
			"import-interpolation",
			"import-once",
			"import-reference",
			"import-reference-issues",
			"lazy-eval",
			"media",
			"merge",
			"mixin-noparens",
			"mixins",
			"mixins-closure",
			"mixins-guards",
			"mixins-guards-default-func",
			"mixins-important",
			"mixins-interpolated",
			"mixins-named-args",
			"mixins-nested",
			"mixins-pattern",
			"no-output",
			"operations",
			"parse-interpolation",
			"permissive-parse",
			"plugi",
			"property-accessors",
			"property-name-interp",
			"rulesets",
			"scope",
			"selectors",
			"strings",
			"urls",
			"variables",
			"variables-in-at-rules",
			"whitespace",
		},
	},
	{
		suite:   "namespacing",
		folder:  "namespacing/",
		options: map[string]any{},
		// Excluded: namespacing-3 (context bug), namespacing-media (undefined namespace)
		files: []string{
			"namespacing-1",
			"namespacing-2",
			"namespacing-4",
			"namespacing-5",
			"namespacing-6",
			"namespacing-7",
			"namespacing-8",
			"namespacing-functions",
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
		suite:  "static-urls",
		folder: "static-urls/",
		options: map[string]any{
			"math":         "strict",
			"relativeUrls": false,
			"rootpath":     "folder (1)/",
		},
		files: []string{
			"urls",
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
		suite:  "url-args",
		folder: "url-args/",
		options: map[string]any{
			"urlArgs": "424242",
		},
		files: []string{
			"urls",
		},
	},
	{
		suite:  "rewrite-urls-all",
		folder: "rewrite-urls-all/",
		options: map[string]any{
			"rewriteUrls": "all",
		},
		files: []string{
			"rewrite-urls-all",
		},
	},
	{
		suite:  "rewrite-urls-local",
		folder: "rewrite-urls-local/",
		options: map[string]any{
			"rewriteUrls": "local",
		},
		files: []string{
			"rewrite-urls-local",
		},
	},
	{
		suite:  "rootpath-rewrite-urls-all",
		folder: "rootpath-rewrite-urls-all/",
		options: map[string]any{
			"rootpath":    "http://example.com/assets/css/",
			"rewriteUrls": "all",
		},
		files: []string{
			"rootpath-rewrite-urls-all",
		},
	},
	{
		suite:  "rootpath-rewrite-urls-local",
		folder: "rootpath-rewrite-urls-local/",
		options: map[string]any{
			"rootpath":    "http://example.com/assets/css/",
			"rewriteUrls": "local",
		},
		files: []string{
			"rootpath-rewrite-urls-local",
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
	{
		suite:  "include-path-string",
		folder: "include-path-string/",
		options: map[string]any{
			"paths": "data/",
		},
		files: []string{
			"include-path-string",
		},
	},
	{
		suite:  "process-imports",
		folder: "process-imports/",
		options: map[string]any{
			"processImports": false,
		},
		files: []string{
			"google",
		},
	},
}

// PLUGIN TESTS: These tests require Node.js/plugin support and should be
// benchmarked separately as they involve IPC overhead.
var benchmarkPluginTestFiles = []struct {
	suite   string
	options map[string]any
	folder  string
	files   []string
}{
	{
		suite:  "main-plugins",
		folder: "_main/",
		options: map[string]any{
			"relativeUrls":       true,
			"silent":             true,
			"javascriptEnabled":  true,
		},
		// Tests that require plugin system / Node.js IPC
		files: []string{
			"import",
			"javascript",
			"plugin",
			"plugin-preeval",
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
// This simulates a realistic build workload where all files are compiled sequentially
// in a single session. Each benchmark iteration represents one complete build.
// NO warmup runs - each iteration is independent, simulating a fresh build process.
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create factory fresh for each iteration (simulating a fresh build)
		factory := Factory(nil, nil)

		for _, test := range tests {
			// Compile the LESS file
			_, compileErr := compileLessForTest(factory, test.content, test.options)
			// Ignore errors in batch benchmark to keep running, but we could track them
			_ = compileErr
		}
	}
}

// BenchmarkPluginTests benchmarks tests that require Node.js/plugin support
// These are benchmarked separately because they involve IPC overhead to Node.js
// Use this benchmark to measure plugin system performance specifically
func BenchmarkPluginTests(b *testing.B) {
	testDataRoot := "../../../../test-data"
	lessRoot := filepath.Join(testDataRoot, "less")

	for _, suite := range benchmarkPluginTestFiles {
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

				// Warmup runs
				const warmupRuns = 5
				for i := 0; i < warmupRuns; i++ {
					compileOpts := &CompileOptions{
						EnableJavaScriptPlugins: true,
						Filename:                lessFile,
					}
					_, _ = Compile(string(lessData), compileOpts)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					compileOpts := &CompileOptions{
						EnableJavaScriptPlugins: true,
						Filename:                lessFile,
					}
					_, compileErr := Compile(string(lessData), compileOpts)
					if compileErr != nil {
						b.Fatalf("Compile error: %v", compileErr)
					}
				}
			})
		}
	}
}
