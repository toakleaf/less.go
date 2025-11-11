// Example demonstrating parallel compilation of multiple LESS files
// This is not a test file, just an example for documentation purposes

package examples

import (
	"fmt"
	"runtime"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go"
)

// ExampleBasicParallelCompilation demonstrates the simplest use case
func ExampleBasicParallelCompilation() {
	// Prepare compilation jobs
	jobs := []less_go.CompileJob{
		{
			Input:   `.button { color: blue; font-size: 14px; }`,
			Options: map[string]any{"filename": "button.less"},
			ID:      "button.less",
		},
		{
			Input:   `.header { color: red; font-size: 18px; }`,
			Options: map[string]any{"filename": "header.less"},
			ID:      "header.less",
		},
		{
			Input:   `.footer { color: gray; font-size: 12px; }`,
			Options: map[string]any{"filename": "footer.less"},
			ID:      "footer.less",
		},
	}

	// Create factory
	factory := less_go.Factory(nil, nil)

	// Compile in parallel
	opts := &less_go.ParallelCompileOptions{
		Enable:      true,
		MaxWorkers:  0, // Use all CPUs
		StopOnError: false,
	}

	results := less_go.BatchCompile(factory, jobs, opts)

	// Process results
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("❌ Error compiling %s: %v\n", result.ID, result.Error)
		} else {
			fmt.Printf("✅ Compiled %s: %d bytes\n", result.ID, len(result.CSS))
		}
	}
}

// ExampleBuildToolIntegration shows how a build tool might use parallel compilation
func ExampleBuildToolIntegration() {
	// Simulate discovering LESS files in a project
	lessFiles := map[string]string{
		"styles/main.less":   `@import "variables"; .main { color: @primary; }`,
		"styles/theme.less":  `@theme: dark; .theme { background: @theme; }`,
		"styles/layout.less": `.container { width: 1200px; margin: 0 auto; }`,
	}

	// Convert to jobs
	jobs := make([]less_go.CompileJob, 0, len(lessFiles))
	for filename, content := range lessFiles {
		jobs = append(jobs, less_go.CompileJob{
			Input: content,
			Options: map[string]any{
				"filename":      filename,
				"compress":      true,  // Minify output
				"sourceMap":     false, // Disable source maps for production
				"relativeUrls":  true,
				"strictUnits":   false,
				"strictMath":    false,
			},
			ID: filename,
		})
	}

	factory := less_go.Factory(nil, nil)

	// Use parallel compilation for faster builds
	results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
		Enable:      true,
		MaxWorkers:  runtime.NumCPU(),
		StopOnError: true, // Fail fast in CI/CD
	})

	// Collect results
	compiled := make(map[string]string)
	var errors []error

	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, result.Error)
		} else {
			compiled[result.ID] = result.CSS
		}
	}

	if len(errors) > 0 {
		fmt.Printf("Compilation failed with %d errors\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
	} else {
		fmt.Printf("Successfully compiled %d files\n", len(compiled))
	}
}

// ExampleConvenienceFunction demonstrates the simplified API
func ExampleConvenienceFunction() {
	inputs := []struct {
		Content  string
		Options  map[string]any
		Filename string
	}{
		{
			Content:  `.test1 { color: red; }`,
			Options:  map[string]any{"compress": false},
			Filename: "test1.less",
		},
		{
			Content:  `.test2 { color: blue; }`,
			Options:  map[string]any{"compress": false},
			Filename: "test2.less",
		},
		{
			Content:  `.test3 { color: green; }`,
			Options:  map[string]any{"compress": false},
			Filename: "test3.less",
		},
	}

	// Simple one-liner for parallel compilation
	results := less_go.ParallelCompileMultipleFiles(inputs, true)

	for _, result := range results {
		if result.Error == nil {
			fmt.Printf("%s: %d bytes\n", result.ID, len(result.CSS))
		}
	}
}

// ExampleSequentialMode shows how to explicitly use sequential mode
func ExampleSequentialMode() {
	jobs := []less_go.CompileJob{
		{Input: `.a { color: red; }`, ID: "a.less"},
		{Input: `.b { color: blue; }`, ID: "b.less"},
	}

	factory := less_go.Factory(nil, nil)

	// Use sequential mode (original behavior)
	results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
		Enable: false, // Explicitly disable parallelization
	})

	fmt.Printf("Compiled %d files sequentially\n", len(results))
}

// ExampleLimitedWorkers shows how to limit CPU usage
func ExampleLimitedWorkers() {
	jobs := make([]less_go.CompileJob, 20)
	for i := 0; i < 20; i++ {
		jobs[i] = less_go.CompileJob{
			Input:   fmt.Sprintf(`.test%d { color: red; }`, i),
			Options: map[string]any{"filename": fmt.Sprintf("test%d.less", i)},
			ID:      fmt.Sprintf("test%d", i),
		}
	}

	factory := less_go.Factory(nil, nil)

	// Use only 4 workers to limit CPU usage
	results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
		Enable:     true,
		MaxWorkers: 4, // Limit to 4 concurrent compilations
	})

	fmt.Printf("Compiled %d files with 4 workers\n", len(results))
}

// ExampleErrorHandling demonstrates error handling strategies
func ExampleErrorHandling() {
	jobs := []less_go.CompileJob{
		{Input: `.valid { color: red; }`, ID: "valid.less"},
		{Input: `.invalid { color: @undefined`, ID: "invalid.less"}, // Syntax error
		{Input: `.another { color: blue; }`, ID: "another.less"},
	}

	factory := less_go.Factory(nil, nil)

	// Continue on error (default)
	results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
		Enable:      true,
		StopOnError: false, // Continue compiling other files
	})

	successCount := 0
	errorCount := 0

	for _, result := range results {
		if result.Error != nil {
			errorCount++
			fmt.Printf("❌ %s: %v\n", result.ID, result.Error)
		} else {
			successCount++
			fmt.Printf("✅ %s\n", result.ID)
		}
	}

	fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, errorCount)
}

// ExampleWatchModeSimulation shows how a file watcher might use this
func ExampleWatchModeSimulation() {
	// Simulate a file watcher detecting changes
	changedFiles := []string{"main.less", "theme.less"}

	// Create jobs for changed files
	jobs := make([]less_go.CompileJob, len(changedFiles))
	for i, filename := range changedFiles {
		// In reality, you'd read from disk
		content := fmt.Sprintf(`.%s { color: blue; }`, filename)

		jobs[i] = less_go.CompileJob{
			Input:   content,
			Options: map[string]any{"filename": filename},
			ID:      filename,
		}
	}

	factory := less_go.Factory(nil, nil)

	// Quick parallel recompilation
	results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
		Enable:     true,
		MaxWorkers: 0, // Use all CPUs for fast rebuild
	})

	// Write results back to disk
	for _, result := range results {
		if result.Error == nil {
			outputFile := result.ID + ".css"
			fmt.Printf("Would write %s (%d bytes)\n", outputFile, len(result.CSS))
			// In reality: ioutil.WriteFile(outputFile, []byte(result.CSS), 0644)
		}
	}
}

// ExamplePerformanceComparison shows the performance difference
func ExamplePerformanceComparison() {
	// Create 50 compilation jobs
	jobs := make([]less_go.CompileJob, 50)
	for i := 0; i < 50; i++ {
		jobs[i] = less_go.CompileJob{
			Input: fmt.Sprintf(`
				@color%d: #%06x;
				.test%d {
					color: @color%d;
					background: darken(@color%d, 10%%);
					border: 1px solid lighten(@color%d, 20%%);
				}
			`, i, i*1000, i, i, i, i),
			Options: map[string]any{"filename": fmt.Sprintf("test%d.less", i)},
			ID:      fmt.Sprintf("test%d", i),
		}
	}

	factory := less_go.Factory(nil, nil)

	// Sequential compilation
	fmt.Println("Running sequential compilation...")
	// results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{Enable: false})

	// Parallel compilation
	fmt.Println("Running parallel compilation...")
	results := less_go.BatchCompile(factory, jobs, &less_go.ParallelCompileOptions{
		Enable:     true,
		MaxWorkers: 0,
	})

	fmt.Printf("Compiled %d files\n", len(results))
	fmt.Println("Try benchmarking to see the speedup!")
}
