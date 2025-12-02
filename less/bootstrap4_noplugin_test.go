package less_go

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"
)

// TestBootstrap4NoPluginPerf tests bootstrap4 without plugin support
// This isolates Go parser/eval performance from plugin IPC overhead
func TestBootstrap4NoPluginPerf(t *testing.T) {
	testDataRoot := "../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")

	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		t.Fatalf("Cannot read bootstrap4.less: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0,
		"paths":    []string{filepath.Dir(lessFile)},
	}

	factory := Factory(nil, nil)

	start := time.Now()
	_, err = compileLessForTest(factory, string(lessData), options)
	elapsed := time.Since(start)

	if err != nil {
		t.Logf("Compilation failed (expected - no plugins): %v", err)
	}

	fmt.Printf("Bootstrap4 (no plugins) single run: %v\n", elapsed)
	t.Logf("Bootstrap4 (no plugins) single run: %v", elapsed)
}

// TestBootstrap4NoPluginMultiRun runs multiple times without plugins
func TestBootstrap4NoPluginMultiRun(t *testing.T) {
	testDataRoot := "../testdata"
	lessFile := filepath.Join(testDataRoot, "less/3rd-party/bootstrap4.less")

	lessData, err := ioutil.ReadFile(lessFile)
	if err != nil {
		t.Fatalf("Cannot read bootstrap4.less: %v", err)
	}

	options := map[string]any{
		"filename": lessFile,
		"math":     0,
		"paths":    []string{filepath.Dir(lessFile)},
	}

	factory := Factory(nil, nil)

	const warmupRuns = 3
	const testRuns = 5

	t.Logf("Bootstrap4 Go (NO PLUGINS) Performance Test")
	t.Logf("==============================================")

	var times []time.Duration
	var coldStartTime time.Duration

	for i := 0; i < warmupRuns+testRuns; i++ {
		start := time.Now()
		_, err := compileLessForTest(factory, string(lessData), options)
		elapsed := time.Since(start)

		if i == 0 {
			coldStartTime = elapsed
		}

		if i >= warmupRuns {
			times = append(times, elapsed)
		}

		status := ""
		if i < warmupRuns {
			status = "(warmup)"
		}
		if err != nil {
			t.Logf("Run %d: %v %s - ERROR: %v", i+1, elapsed, status, err)
		} else {
			t.Logf("Run %d: %v %s", i+1, elapsed, status)
		}
	}

	if len(times) > 0 {
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
		t.Logf("Results:")
		t.Logf("  Cold start:  %v", coldStartTime)
		t.Logf("  Average:     %v", avg)
		t.Logf("  Min:         %v", min)
		t.Logf("  Max:         %v", max)
	}
}
