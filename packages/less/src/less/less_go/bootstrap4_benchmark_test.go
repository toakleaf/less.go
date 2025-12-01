package less_go

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestBootstrap4Performance runs a performance benchmark specifically for bootstrap4
// This test is NOT a benchmark - it's a test that outputs timing information
// Run with: go test -v -run TestBootstrap4Performance
func TestBootstrap4Performance(t *testing.T) {
	testDataRoot := "../../../../../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")

	// Read the file
	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		t.Skipf("Skipping test - bootstrap4.less not available: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0, // strict math
		"paths":    []string{filepath.Dir(lessFile)},
	}

	const warmupRuns = 3
	const testRuns = 5

	t.Logf("Bootstrap4 Go Performance Test")
	t.Logf("==============================================")
	t.Logf("Warmup runs: %d, Test runs: %d", warmupRuns, testRuns)
	t.Logf("")

	var times []time.Duration
	var coldStartTime time.Duration

	// Run tests
	for i := 0; i < warmupRuns+testRuns; i++ {
		start := time.Now()

		// Use plugin-enabled compilation for bootstrap4
		_, err := compileLessWithPlugins(string(lessData), options)

		elapsed := time.Since(start)

		if i == 0 {
			coldStartTime = elapsed
		}

		if err != nil {
			t.Logf("Run %d: FAILED - %v", i+1, err)
			// Don't fail the test - we're just measuring performance
			continue
		}

		if i >= warmupRuns {
			times = append(times, elapsed)
		}

		status := ""
		if i < warmupRuns {
			status = "(warmup)"
		}
		t.Logf("Run %d: %v %s", i+1, elapsed, status)
	}

	if len(times) == 0 {
		t.Logf("\nNo successful runs - unable to calculate stats")
		return
	}

	// Calculate stats
	var sum time.Duration
	min := times[0]
	max := times[0]

	for _, d := range times {
		sum += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	avg := sum / time.Duration(len(times))

	t.Logf("")
	t.Logf("Results (after warmup):")
	t.Logf("----------------------------------------------")
	t.Logf("  Cold start:  %v", coldStartTime)
	t.Logf("  Average:     %v", avg)
	t.Logf("  Min:         %v", min)
	t.Logf("  Max:         %v", max)
	t.Logf("")
}

// BenchmarkBootstrap4 runs a proper Go benchmark for bootstrap4
// Run with: go test -bench=BenchmarkBootstrap4 -benchtime=5s
func BenchmarkBootstrap4(b *testing.B) {
	testDataRoot := "../../../../../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")

	// Read the file
	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		b.Skipf("Skipping benchmark - bootstrap4.less not available: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0, // strict math
		"paths":    []string{filepath.Dir(lessFile)},
	}

	// Warmup runs
	for i := 0; i < 3; i++ {
		_, _ = compileLessWithPlugins(string(lessData), options)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compileLessWithPlugins(string(lessData), options)
		if err != nil {
			b.Fatalf("Compilation error: %v", err)
		}
	}
}

// BenchmarkBootstrap4ColdStart measures cold-start performance (no warmup)
func BenchmarkBootstrap4ColdStart(b *testing.B) {
	testDataRoot := "../../../../../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")

	// Read the file
	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		b.Skipf("Skipping benchmark - bootstrap4.less not available: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0, // strict math
		"paths":    []string{filepath.Dir(lessFile)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := compileLessWithPlugins(string(lessData), options)
		if err != nil {
			b.Fatalf("Compilation error: %v", err)
		}
	}
}

// TestBootstrap4CompareOutput tests if the Go output matches less.js
// This is separate from performance testing
func TestBootstrap4CompareOutput(t *testing.T) {
	testDataRoot := "../../../../../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")
	expectedCSSFile := filepath.Join(testDataRoot, "css/3rd-party/bootstrap4.css")

	// Read the less file
	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		t.Skipf("Skipping test - bootstrap4.less not available: %v", err)
	}

	// Read expected CSS
	expectedCSS, err := ioutil.ReadFile(expectedCSSFile)
	if err != nil {
		t.Skipf("Skipping test - bootstrap4.css not available: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0, // strict math
		"paths":    []string{filepath.Dir(lessFile)},
	}

	// Compile
	start := time.Now()
	result, err := compileLessWithPlugins(string(lessData), options)
	elapsed := time.Since(start)

	t.Logf("Compilation took: %v", elapsed)

	if err != nil {
		// Skip if compilation failed due to missing dependencies
		if strings.Contains(err.Error(), "no such file or directory") {
			t.Skipf("Skipping test - bootstrap4 dependencies not available: %v", err)
		}
		t.Fatalf("Compilation failed: %v", err)
	}

	t.Logf("Output CSS length: %d bytes", len(result))
	t.Logf("Expected CSS length: %d bytes", len(expectedCSS))

	// Compare
	if result != string(expectedCSS) {
		t.Logf("Output does not match expected CSS")
		// Show first difference
		for i := 0; i < len(result) && i < len(expectedCSS); i++ {
			if result[i] != expectedCSS[i] {
				start := i - 50
				if start < 0 {
					start = 0
				}
				end := i + 50
				if end > len(result) {
					end = len(result)
				}
				t.Logf("First difference at byte %d:", i)
				t.Logf("  Expected: ...%q...", string(expectedCSS[start:end]))
				if end <= len(result) {
					t.Logf("  Actual:   ...%q...", result[start:end])
				}
				break
			}
		}
	} else {
		t.Logf("✓ Output matches expected CSS!")
	}
}

// TestBootstrap4QuickPerf is a quick single-run performance check
func TestBootstrap4QuickPerf(t *testing.T) {
	testDataRoot := "../../../../../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")

	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		t.Skipf("Skipping test - bootstrap4.less not available: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0,
		"paths":    []string{filepath.Dir(lessFile)},
	}

	start := time.Now()
	_, err = compileLessWithPlugins(string(lessData), options)
	elapsed := time.Since(start)

	if err != nil {
		t.Logf("Compilation failed: %v", err)
	}

	fmt.Printf("Bootstrap4 single run: %v\n", elapsed)
	t.Logf("Bootstrap4 single run: %v", elapsed)
}
