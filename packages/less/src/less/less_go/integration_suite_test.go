package less_go

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Global results collector for integration suite summary with mutex protection
var (
	integrationResults []TestResult
	resultsMutex       sync.Mutex
)

// Debug configuration from environment
var (
	debugMode   = os.Getenv("LESS_GO_DEBUG") == "1"
	showAST     = os.Getenv("LESS_GO_AST") == "1"
	showTrace   = os.Getenv("LESS_GO_TRACE") == "1"
	showDiff    = os.Getenv("LESS_GO_DIFF") == "1"
	strictMode  = os.Getenv("LESS_GO_STRICT") == "1"  // Fail tests on output differences
	quietMode   = os.Getenv("LESS_GO_QUIET") == "1"   // Suppress individual test output
	jsonOutput  = os.Getenv("LESS_GO_JSON") == "1"    // Output results as JSON
)

// addTestResult safely adds a test result to the global results slice
func addTestResult(result TestResult) {
	resultsMutex.Lock()
	defer resultsMutex.Unlock()
	integrationResults = append(integrationResults, result)
}

// TestIntegrationSuite runs the comprehensive test suite that matches JavaScript test/index.js
func TestIntegrationSuite(t *testing.T) {
	// Reset results collector safely
	resultsMutex.Lock()
	integrationResults = []TestResult{}
	resultsMutex.Unlock()
	
	// Track overall progress
	startTime := time.Now()
	if debugMode {
		fmt.Println("\n🚀 Starting Integration Test Suite with Debug Mode")
		fmt.Printf("   Debug Options: AST=%v, Trace=%v, Diff=%v\n", showAST, showTrace, showDiff)
	}
	
	// Base paths for test data - from packages/less/src/less/less_go to packages/test-data
	testDataRoot := "../../../../test-data"
	lessRoot := filepath.Join(testDataRoot, "less")
	cssRoot := filepath.Join(testDataRoot, "css")

	// Define test map that mirrors JavaScript test/index.js
	testMap := []TestSuite{
		{
			Name: "main",
			Options: map[string]any{
				"relativeUrls":      true,
				"silent":           true,
				"javascriptEnabled": true,
			},
			Folder: "_main/",
		},
		{
			Name:   "namespacing",
			Options: map[string]any{},
			Folder: "namespacing/",
		},
		{
			Name: "math-parens",
			Options: map[string]any{
				"math": "parens",
			},
			Folder: "math/strict/",
		},
		{
			Name: "math-parens-division",
			Options: map[string]any{
				"math": "parens-division",
			},
			Folder: "math/parens-division/",
		},
		{
			Name: "math-always",
			Options: map[string]any{
				"math": "always",
			},
			Folder: "math/always/",
		},
		{
			Name: "compression",
			Options: map[string]any{
				"math":     "strict",
				"compress": true,
			},
			Folder: "compression/",
		},
		{
			Name: "static-urls",
			Options: map[string]any{
				"math":         "strict",
				"relativeUrls": false,
				"rootpath":     "folder (1)/",
			},
			Folder: "static-urls/",
		},
		{
			Name: "units-strict",
			Options: map[string]any{
				"math":        0,
				"strictUnits": true,
			},
			Folder: "units/strict/",
		},
		{
			Name: "units-no-strict",
			Options: map[string]any{
				"math":        0,
				"strictUnits": false,
			},
			Folder: "units/no-strict/",
		},
		{
			Name: "url-args",
			Options: map[string]any{
				"urlArgs": "424242",
			},
			Folder: "url-args/",
		},
		{
			Name: "rewrite-urls-all",
			Options: map[string]any{
				"rewriteUrls": "all",
			},
			Folder: "rewrite-urls-all/",
		},
		{
			Name: "rewrite-urls-local",
			Options: map[string]any{
				"rewriteUrls": "local",
			},
			Folder: "rewrite-urls-local/",
		},
		{
			Name: "rootpath-rewrite-urls-all",
			Options: map[string]any{
				"rootpath":    "http://example.com/assets/css/",
				"rewriteUrls": "all",
			},
			Folder: "rootpath-rewrite-urls-all/",
		},
		{
			Name: "rootpath-rewrite-urls-local",
			Options: map[string]any{
				"rootpath":    "http://example.com/assets/css/",
				"rewriteUrls": "local",
			},
			Folder: "rootpath-rewrite-urls-local/",
		},
		{
			Name: "include-path",
			Options: map[string]any{
				"paths": []string{"data/", "_main/import/"},
			},
			Folder: "include-path/",
		},
		{
			Name: "include-path-string",
			Options: map[string]any{
				"paths": "data/",
			},
			Folder: "include-path-string/",
		},
		{
			Name: "third-party",
			Options: map[string]any{
				"math": 0,
			},
			Folder: "3rd-party/",
		},
		{
			Name: "process-imports",
			Options: map[string]any{
				"processImports": false,
			},
			Folder: "process-imports/",
		},
	}

	// Error test suites (these should fail compilation)
	errorTestMap := []TestSuite{
		{
			Name: "eval-errors",
			Options: map[string]any{
				"strictMath":        true,
				"strictUnits":       true,
				"javascriptEnabled": true,
			},
			Folder:      "../errors/eval/",
			ExpectError: true,
		},
		{
			Name: "parse-errors",
			Options: map[string]any{
				"strictMath":        true,
				"strictUnits":       true,
				"javascriptEnabled": true,
			},
			Folder:      "../errors/parse/",
			ExpectError: true,
		},
		{
			Name: "js-type-errors",
			Options: map[string]any{
				"math":              "strict",
				"strictUnits":       true,
				"javascriptEnabled": true,
			},
			Folder:      "js-type-errors/",
			ExpectError: true,
		},
		{
			Name: "no-js-errors",
			Options: map[string]any{
				"math":              "strict",
				"strictUnits":       true,
				"javascriptEnabled": false,
			},
			Folder:      "no-js-errors/",
			ExpectError: true,
		},
	}

	// Run success test suites
	for _, suite := range testMap {
		t.Run(suite.Name, func(t *testing.T) {
			runTestSuite(t, suite, lessRoot, cssRoot)
		})
	}

	// Run error test suites
	for _, suite := range errorTestMap {
		t.Run(suite.Name, func(t *testing.T) {
			runErrorTestSuite(t, suite, lessRoot)
		})
	}

	// Run summary as a final subtest to ensure it appears at the end
	t.Run("zzz_summary", func(t *testing.T) {
		// Create a copy of results under lock
		resultsMutex.Lock()
		resultsCopy := make([]TestResult, len(integrationResults))
		copy(resultsCopy, integrationResults)
		resultsMutex.Unlock()
		
		printTestSummary(t, resultsCopy)
		
		// Show timing information
		duration := time.Since(startTime)
		t.Logf("\n⏱️  Total test duration: %v", duration)
		
		// Memory usage in debug mode
		if debugMode {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			t.Logf("💾 Memory used: %.2f MB", float64(m.Alloc)/1024/1024)
			t.Logf("🔄 GC runs: %d", m.NumGC)
		}
	})
}

type TestSuite struct {
	Name        string
	Options     map[string]any
	Folder      string
	ExpectError bool
}

type TestResult struct {
	Suite        string
	TestName     string
	Status       string // "pass", "fail", "skip"
	Category     string // "perfect_match", "compilation_failed", "output_differs", "correctly_failed", "expected_error", "quarantined"
	ExpectError  bool   // Whether this test should fail
	Error        string
	ExpectedCSS  string
	ActualCSS    string
}

// Quarantined tests - features not yet implemented that we're punting on for now
var quarantinedTests = map[string][]string{
	"main": {
		// JavaScript execution - to be implemented later
		"javascript",
		// Import test that depends on plugins - TESTING IF THIS NOW WORKS
		// "import",
	},
	"third-party": {
		// Bootstrap 4 requires context-aware plugins that can look up variables via `this.context.frames`.
		// The bootstrap-less-port plugins (breakpoints, theme-color-level, etc.) need access to
		// the Less.js evaluation context for variable lookups. This requires serializing the Go
		// frames/variables to JavaScript, which is a significant feature beyond basic plugin support.
		// The percentage() function bug has been fixed, but context-aware plugins are not yet supported.
		"bootstrap4",
	},
	"js-type-errors": {
		// JavaScript error tests - skip entire suite
		"*",
	},
	"no-js-errors": {
		// JavaScript error tests - skip entire suite
		"*",
	},
}

// isQuarantined checks if a test should be skipped
func isQuarantined(suiteName, testName string) bool {
	if tests, exists := quarantinedTests[suiteName]; exists {
		// Check for wildcard (skip entire suite)
		for _, pattern := range tests {
			if pattern == "*" {
				return true
			}
			if pattern == testName {
				return true
			}
		}
	}
	return false
}

func runTestSuite(t *testing.T, suite TestSuite, lessRoot, cssRoot string) {
	lessDir := filepath.Join(lessRoot, suite.Folder)
	cssDir := filepath.Join(cssRoot, suite.Folder)

	// Find all .less files in the directory
	lessFiles, err := filepath.Glob(filepath.Join(lessDir, "*.less"))
	if err != nil {
		t.Fatalf("Failed to find .less files in %s: %v", lessDir, err)
	}

	if len(lessFiles) == 0 {
		t.Skipf("No .less files found in %s", lessDir)
		return
	}
	successCount := 0
	totalCount := len(lessFiles)

	for _, lessFile := range lessFiles {
		fileName := filepath.Base(lessFile)
		testName := strings.TrimSuffix(fileName, ".less")

		t.Run(testName, func(t *testing.T) {
			result := TestResult{
				Suite:       suite.Name,
				TestName:    testName,
				ExpectError: false,
			}

			// Check if test is quarantined
			if isQuarantined(suite.Name, testName) {
				result.Status = "skip"
				result.Category = "quarantined"
				result.Error = "Quarantined: Feature not yet implemented (plugin system or JavaScript execution)"
				addTestResult(result)
				if !quietMode {
					t.Skipf("⏸️  Quarantined: %s (feature not yet implemented)", testName)
				}
				return
			}

			// Read the .less file
			lessContent, err := ioutil.ReadFile(lessFile)
			if err != nil {
				result.Status = "skip"
				result.Category = "skip"
				result.Error = "Failed to read .less file: " + err.Error()
				addTestResult(result)
				if !quietMode {
					t.Skipf("Failed to read %s: %v", lessFile, err)
				}
				return
			}

			// Expected CSS file
			cssFile := filepath.Join(cssDir, testName+".css")
			expectedCSS, err := ioutil.ReadFile(cssFile)
			if err != nil {
				result.Status = "skip"
				result.Category = "skip"
				result.Error = "Expected CSS file not found: " + err.Error()
				addTestResult(result)
				if !quietMode {
					t.Skipf("Expected CSS file %s not found: %v", cssFile, err)
				}
				return
			}

			result.ExpectedCSS = strings.TrimSpace(string(expectedCSS))

			// Set up options
			options := make(map[string]any)
			for k, v := range suite.Options {
				options[k] = v
			}
			options["filename"] = lessFile

			// Handle include paths - merge suite paths with file directory
			// Start with the file's directory as the primary search path
			paths := []string{filepath.Dir(lessFile)}

			// Add suite paths if specified (resolve relative to lessRoot)
			if suitePaths, ok := suite.Options["paths"]; ok {
				switch v := suitePaths.(type) {
				case []string:
					for _, p := range v {
						// Resolve relative paths relative to lessRoot
						if !filepath.IsAbs(p) {
							p = filepath.Join(lessRoot, p)
						}
						paths = append(paths, p)
					}
				case string:
					// Single string path
					p := v
					if !filepath.IsAbs(p) {
						p = filepath.Join(lessRoot, p)
					}
					paths = append(paths, p)
				case []any:
					// Handle []any type (from JSON unmarshaling)
					for _, pAny := range v {
						if pStr, ok := pAny.(string); ok {
							if !filepath.IsAbs(pStr) {
								pStr = filepath.Join(lessRoot, pStr)
							}
							paths = append(paths, pStr)
						}
					}
				}
			}

			options["paths"] = paths

			// Compile with appropriate method based on test type
			var actualResult string

			if isPluginTest(testName) {
				// Use plugin-enabled compilation for plugin tests
				actualResult, err = compileLessWithPlugins(string(lessContent), options)
			} else {
				// Use standard compilation
				factory := Factory(nil, nil)
				actualResult, err = compileLessWithDebug(factory, string(lessContent), options)
			}
			if err != nil {
				result.Status = "fail"
				result.Category = "compilation_failed"
				result.Error = err.Error()
				result.ActualCSS = ""
				addTestResult(result)

				if strictMode {
					// In strict mode, fail the test on compilation errors
					t.Errorf("❌ %s: Compilation failed: %v", testName, err)
				} else if !quietMode {
					// In normal mode, just log the error (expected during development)
					t.Logf("❌ %s: Compilation failed: %v", testName, err)
				}
				if debugMode {
					enhancedErrorReport(t, err, lessFile, string(lessContent))
				}
				return
			}

			// Compare results
			result.ActualCSS = strings.TrimSpace(actualResult)

			if result.ActualCSS == result.ExpectedCSS {
				result.Status = "pass"
				result.Category = "perfect_match"
				addTestResult(result)
				if !quietMode {
					t.Logf("✅ %s: Perfect match!", testName)
				}
				successCount++
			} else {
				result.Status = "fail"
				result.Category = "output_differs"
				result.Error = "Output differs from expected"
				addTestResult(result)

				if strictMode {
					// In strict mode, fail the test immediately
					t.Errorf("❌ %s: Output differs from expected", testName)
					if showDiff {
						t.Errorf("%s", formatDiff(result.ExpectedCSS, result.ActualCSS))
					} else {
						t.Errorf("   Expected: %s", result.ExpectedCSS)
						t.Errorf("   Actual:   %s", result.ActualCSS)
					}
				} else if !quietMode {
					// During the Go port development, many tests will have output differences
					// as features are still being implemented. These are marked as failures
					// but noted as "expected during development" to distinguish them from
					// compilation errors or crashes.
					t.Logf("⚠️  %s: Output differs (expected during development)", testName)
					if showDiff || (len(result.ActualCSS) < 500 && len(result.ExpectedCSS) < 500) {
						if showDiff {
							t.Logf("%s", formatDiff(result.ExpectedCSS, result.ActualCSS))
						} else {
							t.Logf("   Expected: %s", result.ExpectedCSS)
							t.Logf("   Actual:   %s", result.ActualCSS)
						}
					}
				}
			}
		})
	}

	t.Logf("Suite %s: %d/%d tests compiled successfully", suite.Name, successCount, totalCount)
}

func runErrorTestSuite(t *testing.T, suite TestSuite, lessRoot string) {
	lessDir := filepath.Join(lessRoot, suite.Folder)

	// Find all .less files in the directory
	lessFiles, err := filepath.Glob(filepath.Join(lessDir, "*.less"))
	if err != nil {
		t.Fatalf("Failed to find .less files in %s: %v", lessDir, err)
	}

	if len(lessFiles) == 0 {
		t.Skipf("No .less files found in %s", lessDir)
		return
	}

	for _, lessFile := range lessFiles {
		fileName := filepath.Base(lessFile)
		testName := strings.TrimSuffix(fileName, ".less")

		t.Run(testName, func(t *testing.T) {
			result := TestResult{
				Suite:       suite.Name,
				TestName:    testName,
				ExpectError: true,
			}

			// Check if test is quarantined
			if isQuarantined(suite.Name, testName) {
				result.Status = "skip"
				result.Category = "quarantined"
				result.Error = "Quarantined: Feature not yet implemented (plugin system or JavaScript execution)"
				addTestResult(result)
				if !quietMode {
					t.Skipf("⏸️  Quarantined: %s (feature not yet implemented)", testName)
				}
				return
			}

			// Read the .less file
			lessContent, err := ioutil.ReadFile(lessFile)
			if err != nil {
				result.Status = "skip"
				result.Category = "skip"
				result.Error = "Failed to read .less file: " + err.Error()
				addTestResult(result)
				if !quietMode {
					t.Skipf("Failed to read %s: %v", lessFile, err)
				}
				return
			}

			// Set up options
			options := make(map[string]any)
			for k, v := range suite.Options {
				options[k] = v
			}
			options["filename"] = lessFile

			// Handle include paths - merge suite paths with file directory
			// Start with the file's directory as the primary search path
			paths := []string{filepath.Dir(lessFile)}

			// Add suite paths if specified (resolve relative to lessRoot)
			if suitePaths, ok := suite.Options["paths"]; ok {
				switch v := suitePaths.(type) {
				case []string:
					for _, p := range v {
						// Resolve relative paths relative to lessRoot
						if !filepath.IsAbs(p) {
							p = filepath.Join(lessRoot, p)
						}
						paths = append(paths, p)
					}
				case string:
					// Single string path
					p := v
					if !filepath.IsAbs(p) {
						p = filepath.Join(lessRoot, p)
					}
					paths = append(paths, p)
				case []any:
					// Handle []any type (from JSON unmarshaling)
					for _, pAny := range v {
						if pStr, ok := pAny.(string); ok {
							if !filepath.IsAbs(pStr) {
								pStr = filepath.Join(lessRoot, pStr)
							}
							paths = append(paths, pStr)
						}
					}
				}
			}

			options["paths"] = paths

			// Create Less factory
			factory := Factory(nil, nil)

			// Compile - this should fail
			actualResult, err := compileLessWithDebug(factory, string(lessContent), options)
			if err != nil {
				result.Status = "pass" // For error tests, failure is success
				result.Category = "correctly_failed"
				result.Error = err.Error()
				addTestResult(result)
				if !quietMode {
					t.Logf("✅ %s: Correctly failed with error: %v", testName, err)
				}
			} else {
				result.Status = "fail" // For error tests, success is failure
				result.Category = "expected_error"
				result.ActualCSS = actualResult
				result.Error = "Expected error but compilation succeeded"
				addTestResult(result)
				if !quietMode {
					t.Logf("⚠️  %s: Expected error but compilation succeeded", testName)
					t.Logf("   Result: %s", actualResult)
				}
			}
		})
	}

}

// Enhanced debugging helpers

// debugLog prints debug messages when debug mode is enabled
func debugLog(format string, args ...interface{}) {
	if debugMode {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// traceLog prints trace messages when trace mode is enabled
func traceLog(format string, args ...interface{}) {
	if showTrace {
		fmt.Printf("[TRACE] "+format+"\n", args...)
	}
}

// formatDiff creates a visual diff between expected and actual CSS
func formatDiff(expected, actual string) string {
	if !showDiff {
		return ""
	}
	
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	
	var diff strings.Builder
	diff.WriteString("\n--- Expected CSS ---\n")
	for i, line := range expectedLines {
		diff.WriteString(fmt.Sprintf("%3d | %s\n", i+1, line))
	}
	diff.WriteString("\n--- Actual CSS ---\n")
	for i, line := range actualLines {
		diff.WriteString(fmt.Sprintf("%3d | %s\n", i+1, line))
	}
	
	// Find first difference
	for i := 0; i < len(expectedLines) && i < len(actualLines); i++ {
		if expectedLines[i] != actualLines[i] {
			diff.WriteString(fmt.Sprintf("\n⚠️  First difference at line %d:\n", i+1))
			diff.WriteString(fmt.Sprintf("   Expected: %q\n", expectedLines[i]))
			diff.WriteString(fmt.Sprintf("   Actual:   %q\n", actualLines[i]))
			break
		}
	}
	
	return diff.String()
}

// enhancedErrorReport provides detailed error context
func enhancedErrorReport(t *testing.T, err error, lessFile string, lessContent string) {
	if !debugMode {
		return
	}
	
	t.Logf("\n🔍 Enhanced Error Report:")
	t.Logf("   File: %s", lessFile)
	
	// Try to extract line/column from error
	errStr := err.Error()
	if strings.Contains(errStr, "line") || strings.Contains(errStr, "Line") {
		t.Logf("   Position: %s", errStr)
	}
	
	// Show file context if possible
	lines := strings.Split(lessContent, "\n")
	if len(lines) <= 20 {
		t.Logf("   Source Content:")
		for i, line := range lines {
			t.Logf("   %3d | %s", i+1, line)
		}
	}
	
	// Provide suggestions based on error type
	if strings.Contains(errStr, "Parse") || strings.Contains(errStr, "parse") {
		t.Logf("\n💡 Parser Error Suggestions:")
		t.Logf("   • Check for missing semicolons or braces")
		t.Logf("   • Verify syntax matches Less.js grammar")
		t.Logf("   • Look for unsupported syntax features")
	} else if strings.Contains(errStr, "undefined") {
		t.Logf("\n💡 Variable Error Suggestions:")
		t.Logf("   • Check variable scope and definition order")
		t.Logf("   • Verify import statements are processed")
		t.Logf("   • Look for typos in variable names")
	} else if strings.Contains(errStr, "import") {
		t.Logf("\n💡 Import Error Suggestions:")
		t.Logf("   • Verify file paths are correct")
		t.Logf("   • Check import resolution logic")
		t.Logf("   • Ensure imported files exist")
	}
	
	// Stack trace if available
	if debugMode {
		t.Logf("\n📚 Stack Trace:")
		for i := 1; i <= 10; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			if strings.Contains(file, "less_go") {
				t.Logf("   at %s:%d", file, line)
			}
		}
	}
}

// compileLessWithDebug wraps the compilation with additional debugging
func compileLessWithDebug(factory map[string]any, content string, options map[string]any) (string, error) {
	startTime := time.Now()

	debugLog("Starting compilation with options: %+v", options)

	// Add debug hooks if enabled
	if showAST {
		// This would require modification to the actual compiler
		// to expose AST - placeholder for now
		debugLog("AST output enabled (requires compiler support)")
	}

	result, err := compileLessForTest(factory, content, options)

	duration := time.Since(startTime)
	debugLog("Compilation completed in %v", duration)

	if err != nil {
		debugLog("Compilation failed: %v", err)
	} else {
		debugLog("Compilation successful, output length: %d", len(result))
	}

	return result, err
}

// compileLessWithPlugins compiles LESS with JavaScript plugin support enabled.
// This uses the Compile API with EnableJavaScriptPlugins=true.
func compileLessWithPlugins(content string, options map[string]any) (string, error) {
	startTime := time.Now()

	debugLog("Starting plugin-enabled compilation with options: %+v", options)

	// Convert options map to CompileOptions struct
	compileOpts := &CompileOptions{
		EnableJavaScriptPlugins: true,
	}

	// Extract relevant options
	if filename, ok := options["filename"].(string); ok {
		compileOpts.Filename = filename
	}
	if paths, ok := options["paths"].([]string); ok {
		compileOpts.Paths = paths
	}
	if compress, ok := options["compress"].(bool); ok {
		compileOpts.Compress = compress
	}
	if strictUnits, ok := options["strictUnits"].(bool); ok {
		compileOpts.StrictUnits = strictUnits
	}
	if math, ok := options["math"].(MathType); ok {
		compileOpts.Math = math
	} else if mathStr, ok := options["math"].(string); ok {
		switch mathStr {
		case "always":
			compileOpts.Math = Math.Always
		case "parens-division":
			compileOpts.Math = Math.ParensDivision
		case "parens":
			compileOpts.Math = Math.Parens
		}
	}
	if rewriteUrls, ok := options["rewriteUrls"].(RewriteUrlsType); ok {
		compileOpts.RewriteUrls = rewriteUrls
	}
	if rootpath, ok := options["rootpath"].(string); ok {
		compileOpts.Rootpath = rootpath
	}
	if urlArgs, ok := options["urlArgs"].(string); ok {
		compileOpts.UrlArgs = urlArgs
	}

	result, err := Compile(content, compileOpts)

	duration := time.Since(startTime)
	debugLog("Plugin-enabled compilation completed in %v", duration)

	if err != nil {
		debugLog("Compilation failed: %v", err)
		return "", err
	}

	debugLog("Compilation successful, output length: %d", len(result.CSS))
	return result.CSS, nil
}

// isPluginTest returns true if the test requires JavaScript plugin support.
func isPluginTest(testName string) bool {
	// Tests that explicitly start with "plugin"
	if strings.HasPrefix(testName, "plugin") {
		return true
	}
	// Other tests that use @plugin directive
	pluginTests := map[string]bool{
		"import":     true, // Uses @plugin "../../plugin/plugin-simple"
		"bootstrap4": true, // Uses @plugin directives for breakpoints, map-get, color-yiq, etc.
	}
	return pluginTests[testName]
}

// printTestSummary prints a comprehensive, LLM-friendly summary of test results
func printTestSummary(t *testing.T, results []TestResult) {
	// Categorize results
	var (
		perfectMatches    []TestResult
		compilationFailed []TestResult
		outputDiffers     []TestResult
		correctlyFailed   []TestResult
		expectedError     []TestResult
		quarantined       []TestResult
		skipped           []TestResult
	)

	for _, result := range results {
		switch result.Category {
		case "perfect_match":
			perfectMatches = append(perfectMatches, result)
		case "compilation_failed":
			compilationFailed = append(compilationFailed, result)
		case "output_differs":
			outputDiffers = append(outputDiffers, result)
		case "correctly_failed":
			correctlyFailed = append(correctlyFailed, result)
		case "expected_error":
			expectedError = append(expectedError, result)
		case "quarantined":
			quarantined = append(quarantined, result)
		case "skip":
			skipped = append(skipped, result)
		}
	}

	total := len(results)
	activeTotal := total - len(quarantined) - len(skipped)

	// If JSON output is requested, print JSON and return
	if jsonOutput {
		printJSONSummary(t, results, perfectMatches, compilationFailed, outputDiffers,
			correctlyFailed, expectedError, quarantined, skipped)
		return
	}

	// Print quick stats header
	divider := strings.Repeat("=", 80)
	t.Logf("\n" + divider)
	t.Logf("📊 INTEGRATION TEST SUMMARY - Quick Stats")
	t.Logf(divider)
	t.Logf("")
	t.Logf("OVERALL SUCCESS: %d/%d tests (%.1f%%)",
		len(perfectMatches)+len(correctlyFailed),
		activeTotal,
		float64(len(perfectMatches)+len(correctlyFailed))/float64(activeTotal)*100)
	t.Logf("")
	t.Logf("✅ Perfect CSS Matches:      %3d  (%.1f%% of active tests)",
		len(perfectMatches),
		float64(len(perfectMatches))/float64(activeTotal)*100)
	t.Logf("❌ Compilation Failures:     %3d  (%.1f%% of active tests)",
		len(compilationFailed),
		float64(len(compilationFailed))/float64(activeTotal)*100)
	t.Logf("⚠️  Output Differences:       %3d  (%.1f%% of active tests)",
		len(outputDiffers),
		float64(len(outputDiffers))/float64(activeTotal)*100)
	t.Logf("✅ Correctly Failed (Error): %3d  (%.1f%% of active tests)",
		len(correctlyFailed),
		float64(len(correctlyFailed))/float64(activeTotal)*100)
	t.Logf("⚠️  Expected Error:           %3d  (%.1f%% of active tests)",
		len(expectedError),
		float64(len(expectedError))/float64(activeTotal)*100)
	t.Logf("⏸️  Quarantined:              %3d  (not counted - plugin/JS features)",
		len(quarantined))
	if len(skipped) > 0 {
		t.Logf("⏭️  Skipped:                  %3d  (not counted - file errors)", len(skipped))
	}
	t.Logf("")
	t.Logf("TOTAL ACTIVE TESTS: %d", activeTotal)
	t.Logf("COMPILATION RATE:   %.1f%% (%d/%d tests compile successfully)",
		float64(activeTotal-len(compilationFailed))/float64(activeTotal)*100,
		activeTotal-len(compilationFailed),
		activeTotal)
	t.Logf("")

	// Detailed breakdown by category
	t.Logf(divider)
	t.Logf("📋 DETAILED RESULTS BY CATEGORY")
	t.Logf(divider)
	t.Logf("")

	// Perfect matches
	if len(perfectMatches) > 0 {
		t.Logf("✅ PERFECT CSS MATCHES (%d tests) - Fully working!", len(perfectMatches))
		printTestList(t, perfectMatches)
		t.Logf("")
	}

	// Compilation failures
	if len(compilationFailed) > 0 {
		t.Logf("❌ COMPILATION FAILURES (%d tests) - Parser/Runtime errors", len(compilationFailed))
		printTestList(t, compilationFailed)
		t.Logf("")
	}

	// Output differences
	if len(outputDiffers) > 0 {
		t.Logf("⚠️  OUTPUT DIFFERENCES (%d tests) - Compiles but CSS doesn't match", len(outputDiffers))
		printTestList(t, outputDiffers)
		t.Logf("")
	}

	// Correctly failed
	if len(correctlyFailed) > 0 {
		t.Logf("✅ CORRECTLY FAILED (%d tests) - Error tests working correctly", len(correctlyFailed))
		if debugMode {
			printTestList(t, correctlyFailed)
		} else {
			t.Logf("   (run with LESS_GO_DEBUG=1 to see full list)")
		}
		t.Logf("")
	}

	// Expected error but succeeded
	if len(expectedError) > 0 {
		t.Logf("⚠️  EXPECTED ERROR BUT SUCCEEDED (%d tests) - Should fail but doesn't", len(expectedError))
		printTestList(t, expectedError)
		t.Logf("")
	}

	// Quarantined
	if len(quarantined) > 0 {
		t.Logf("⏸️  QUARANTINED (%d tests) - Plugin/JS features not implemented", len(quarantined))
		if debugMode {
			printTestList(t, quarantined)
		} else {
			t.Logf("   (run with LESS_GO_DEBUG=1 to see full list)")
		}
		t.Logf("")
	}

	// Quick commands for LLMs
	t.Logf(divider)
	t.Logf("🤖 QUICK COMMANDS FOR ANALYSIS")
	t.Logf(divider)
	t.Logf("")
	t.Logf("# Get just the summary (no verbose output):")
	t.Logf("LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100")
	t.Logf("")
	t.Logf("# Get JSON output for programmatic analysis:")
	t.Logf("LESS_GO_JSON=1 LESS_GO_QUIET=1 pnpm -w test:go")
	t.Logf("")
	t.Logf("# See detailed diffs for failing tests:")
	t.Logf("LESS_GO_DIFF=1 pnpm -w test:go")
	t.Logf("")
	t.Logf("# Debug a specific test:")
	t.Logf("LESS_GO_DEBUG=1 go test -v -run TestIntegrationSuite/<suite>/<testname>")
	t.Logf("")

	// Next steps
	if len(compilationFailed) > 0 || len(outputDiffers) > 0 || len(expectedError) > 0 {
		t.Logf(divider)
		t.Logf("💡 NEXT STEPS")
		t.Logf(divider)
		t.Logf("")

		if len(compilationFailed) > 0 {
			t.Logf("PRIORITY: Fix %d compilation failures first", len(compilationFailed))
			groupedFailures := groupBySuite(compilationFailed)
			for suite, tests := range groupedFailures {
				t.Logf("  • %s: %d tests", suite, len(tests))
			}
			t.Logf("")
		}

		if len(outputDiffers) > 0 {
			t.Logf("MEDIUM: Fix %d output differences", len(outputDiffers))
			groupedDiffs := groupBySuite(outputDiffers)
			for suite, tests := range groupedDiffs {
				t.Logf("  • %s: %d tests", suite, len(tests))
			}
			t.Logf("")
		}

		if len(expectedError) > 0 {
			t.Logf("LOW: Fix %d error handling tests", len(expectedError))
			groupedErrors := groupBySuite(expectedError)
			for suite, tests := range groupedErrors {
				t.Logf("  • %s: %d tests", suite, len(tests))
			}
			t.Logf("")
		}
	}

	t.Logf(divider)
}

// printTestList prints a formatted list of tests
func printTestList(t *testing.T, tests []TestResult) {
	grouped := groupBySuite(tests)
	for suite, suiteTests := range grouped {
		if len(suiteTests) == 1 {
			t.Logf("   %s/%s", suite, suiteTests[0].TestName)
		} else {
			t.Logf("   %s: (%d tests)", suite, len(suiteTests))
			for _, test := range suiteTests {
				t.Logf("     - %s", test.TestName)
			}
		}
	}
}

// groupBySuite groups test results by suite name
func groupBySuite(tests []TestResult) map[string][]TestResult {
	grouped := make(map[string][]TestResult)
	for _, test := range tests {
		grouped[test.Suite] = append(grouped[test.Suite], test)
	}
	return grouped
}

// printJSONSummary prints results as JSON for programmatic parsing
func printJSONSummary(t *testing.T, results []TestResult, perfectMatches, compilationFailed,
	outputDiffers, correctlyFailed, expectedError, quarantined, skipped []TestResult) {

	summary := map[string]interface{}{
		"total_tests": len(results),
		"active_tests": len(results) - len(quarantined) - len(skipped),
		"categories": map[string]interface{}{
			"perfect_match":       len(perfectMatches),
			"compilation_failed":  len(compilationFailed),
			"output_differs":      len(outputDiffers),
			"correctly_failed":    len(correctlyFailed),
			"expected_error":      len(expectedError),
			"quarantined":         len(quarantined),
			"skipped":             len(skipped),
		},
		"success_rate": float64(len(perfectMatches)+len(correctlyFailed)) / float64(len(results)-len(quarantined)-len(skipped)) * 100,
		"compilation_rate": float64(len(results)-len(quarantined)-len(skipped)-len(compilationFailed)) / float64(len(results)-len(quarantined)-len(skipped)) * 100,
		"perfect_match_rate": float64(len(perfectMatches)) / float64(len(results)-len(quarantined)-len(skipped)) * 100,
		"results": results,
	}

	// Format as JSON (simple manual formatting to avoid import)
	t.Logf("{")
	t.Logf(`  "total_tests": %d,`, summary["total_tests"])
	t.Logf(`  "active_tests": %d,`, summary["active_tests"])
	t.Logf(`  "success_rate": %.2f,`, summary["success_rate"])
	t.Logf(`  "compilation_rate": %.2f,`, summary["compilation_rate"])
	t.Logf(`  "perfect_match_rate": %.2f,`, summary["perfect_match_rate"])
	t.Logf(`  "categories": {`)
	t.Logf(`    "perfect_match": %d,`, len(perfectMatches))
	t.Logf(`    "compilation_failed": %d,`, len(compilationFailed))
	t.Logf(`    "output_differs": %d,`, len(outputDiffers))
	t.Logf(`    "correctly_failed": %d,`, len(correctlyFailed))
	t.Logf(`    "expected_error": %d,`, len(expectedError))
	t.Logf(`    "quarantined": %d,`, len(quarantined))
	t.Logf(`    "skipped": %d`, len(skipped))
	t.Logf(`  }`)
	t.Logf("}")
}

