package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// getPluginHostPath returns the path to plugin-host.js for tests.
func getPluginHostPath(t *testing.T) string {
	// Try relative to test file first
	candidates := []string{
		"plugin-host.js",
		filepath.Join("runtime", "plugin-host.js"),
	}

	// Get the directory of the test file
	wd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "plugin-host.js"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	t.Fatalf("plugin-host.js not found; tried: %v", candidates)
	return ""
}

func TestNodeJSRuntime_NewRuntime(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if rt == nil {
		t.Fatal("runtime is nil")
	}

	if rt.pluginHostPath != path {
		t.Errorf("pluginHostPath = %q, want %q", rt.pluginHostPath, path)
	}
}

func TestNodeJSRuntime_StartStop(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	// Start the runtime
	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify it's alive
	if !rt.IsAlive() {
		t.Error("runtime should be alive after Start")
	}

	// Stop the runtime
	if err := rt.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// Give it a moment to shut down
	time.Sleep(100 * time.Millisecond)

	// Verify it's no longer alive
	if rt.IsAlive() {
		t.Error("runtime should not be alive after Stop")
	}
}

func TestNodeJSRuntime_Ping(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Send ping
	if err := rt.Ping(); err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestNodeJSRuntime_Echo(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	tests := []struct {
		name  string
		value any
	}{
		{"string", "hello world"},
		{"number", float64(42)},
		{"boolean true", true},
		{"boolean false", false},
		{"null", nil},
		{"array", []any{"a", "b", "c"}},
		{"object", map[string]any{"key": "value", "num": float64(123)}},
		{"nested", map[string]any{
			"array": []any{1.0, 2.0, 3.0},
			"obj":   map[string]any{"nested": true},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rt.Echo(tt.value)
			if err != nil {
				t.Errorf("Echo failed: %v", err)
				return
			}

			// For nil input, result should also be nil
			if tt.value == nil {
				if result != nil {
					t.Errorf("Echo() = %v, want nil", result)
				}
				return
			}

			// For other types, do basic validation
			// (JSON round-trip may change types slightly)
			if result == nil {
				t.Errorf("Echo() = nil, want %v", tt.value)
			}
		})
	}
}

func TestNodeJSRuntime_CommandResponse(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Test unknown command
	resp, err := rt.SendCommand(Command{Cmd: "unknown_command"})
	if err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}

	if resp.Success {
		t.Error("unknown command should fail")
	}

	if resp.Error == "" {
		t.Error("unknown command should return error message")
	}
}

func TestNodeJSRuntime_MultipleCommands(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Send multiple commands in sequence
	for i := 0; i < 10; i++ {
		resp, err := rt.SendCommand(Command{
			Cmd:  "echo",
			Data: i,
		})
		if err != nil {
			t.Errorf("Command %d failed: %v", i, err)
			continue
		}
		if !resp.Success {
			t.Errorf("Command %d was not successful: %s", i, resp.Error)
		}
	}
}

func TestNodeJSRuntime_ConcurrentCommands(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Send concurrent commands
	const numCommands = 20
	results := make(chan error, numCommands)

	for i := 0; i < numCommands; i++ {
		go func(idx int) {
			resp, err := rt.SendCommand(Command{
				Cmd:  "echo",
				Data: idx,
			})
			if err != nil {
				results <- err
				return
			}
			if !resp.Success {
				results <- err
				return
			}
			results <- nil
		}(i)
	}

	// Collect results
	var errors []error
	for i := 0; i < numCommands; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent commands had %d errors: %v", len(errors), errors)
	}
}

func TestNodeJSRuntime_DoubleStart(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Try to start again
	err = rt.Start()
	if err == nil {
		t.Error("second Start should return error")
	}
}

func TestNodeJSRuntime_CommandAfterStop(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop the runtime
	if err := rt.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Try to send a command
	_, err = rt.SendCommand(Command{Cmd: "ping"})
	if err == nil {
		t.Error("SendCommand should fail after Stop")
	}
}

func TestNodeJSRuntime_GetRegisteredFunctions(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Get registered functions (should be empty initially)
	resp, err := rt.SendCommand(Command{Cmd: "getRegisteredFunctions"})
	if err != nil {
		t.Fatalf("SendCommand failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("getRegisteredFunctions failed: %s", resp.Error)
	}

	// Result should be an empty array
	functions, ok := resp.Result.([]any)
	if !ok {
		t.Errorf("result is not an array: %T", resp.Result)
		return
	}

	if len(functions) != 0 {
		t.Errorf("expected 0 functions, got %d", len(functions))
	}
}

// Benchmark tests

func BenchmarkNodeJSRuntime_Ping(b *testing.B) {
	path := "plugin-host.js"
	if _, err := os.Stat(path); err != nil {
		b.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rt.Ping(); err != nil {
			b.Fatalf("Ping failed: %v", err)
		}
	}
}

func BenchmarkNodeJSRuntime_Echo(b *testing.B) {
	path := "plugin-host.js"
	if _, err := os.Stat(path); err != nil {
		b.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	data := map[string]any{
		"string": "hello world",
		"number": 42,
		"array":  []any{1, 2, 3, 4, 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := rt.Echo(data); err != nil {
			b.Fatalf("Echo failed: %v", err)
		}
	}
}

// TestNodeJSRuntime_BatchCallFunctions tests the batch IPC functionality
func TestNodeJSRuntime_BatchCallFunctions(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Load a simple plugin that has some functions
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	// Load plugin-simple.js which defines 'pi' and 'pi-anon' functions
	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	_, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Failed to load plugin: %v", result)
	}

	// Test empty batch
	results, err := rt.BatchCallFunctions([]BatchCall{})
	if err != nil {
		t.Fatalf("Empty batch call failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected empty results, got %d", len(results))
	}

	// Test batch with valid function calls
	calls := []BatchCall{
		{Key: "pi:1", Name: "pi", Args: []any{}},
		{Key: "pi:2", Name: "pi", Args: []any{}},
	}

	results, err = rt.BatchCallFunctions(calls)
	if err != nil {
		t.Fatalf("BatchCallFunctions failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check that both calls succeeded
	for _, key := range []string{"pi:1", "pi:2"} {
		result, ok := results[key]
		if !ok {
			t.Errorf("Missing result for key %q", key)
			continue
		}
		if !result.Success {
			t.Errorf("Call %q failed: %s", key, result.Error)
		}
	}

	// Test batch with unknown function (should fail gracefully)
	mixedCalls := []BatchCall{
		{Key: "pi:3", Name: "pi", Args: []any{}},
		{Key: "unknown:1", Name: "unknown_function", Args: []any{}},
	}

	results, err = rt.BatchCallFunctions(mixedCalls)
	if err != nil {
		t.Fatalf("BatchCallFunctions with unknown function failed: %v", err)
	}

	// pi should succeed
	if !results["pi:3"].Success {
		t.Errorf("pi:3 should have succeeded: %s", results["pi:3"].Error)
	}

	// unknown should fail but not crash
	if results["unknown:1"].Success {
		t.Error("unknown:1 should have failed")
	}
	if results["unknown:1"].Error == "" {
		t.Error("unknown:1 should have error message")
	}
}

// TestNodeJSRuntime_BatchCallFunctionsAndCache tests the batch call and cache functionality
func TestNodeJSRuntime_BatchCallFunctionsAndCache(t *testing.T) {
	path := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Load a simple plugin
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	_, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Failed to load plugin: %v", result)
	}

	// Clear any existing cache
	rt.ClearFunctionCache()

	// Test BatchCallFunctionsAndCache
	calls := []BatchCall{
		{Key: "pi:cache1", Name: "pi", Args: []any{}},
		{Key: "pi:cache2", Name: "pi", Args: []any{}},
		{Key: "unknown:cache1", Name: "unknown", Args: []any{}}, // This will fail
	}

	cached, err := rt.BatchCallFunctionsAndCache(calls)
	if err != nil {
		t.Fatalf("BatchCallFunctionsAndCache failed: %v", err)
	}

	// Should cache 2 results (the two successful pi calls)
	if cached != 2 {
		t.Errorf("Expected 2 cached results, got %d", cached)
	}

	// Verify cache was populated
	cacheSize := rt.FunctionCacheSize()
	if cacheSize != 2 {
		t.Errorf("Expected cache size 2, got %d", cacheSize)
	}

	// Verify we can get cached results
	if _, ok := rt.GetCachedResult("pi:cache1"); !ok {
		t.Error("Expected pi:cache1 to be in cache")
	}
	if _, ok := rt.GetCachedResult("pi:cache2"); !ok {
		t.Error("Expected pi:cache2 to be in cache")
	}
	if _, ok := rt.GetCachedResult("unknown:cache1"); ok {
		t.Error("Expected unknown:cache1 NOT to be in cache")
	}
}

// BenchmarkNodeJSRuntime_BatchCallFunctions benchmarks batch IPC vs individual calls
func BenchmarkNodeJSRuntime_BatchCallFunctions(b *testing.B) {
	path := "plugin-host.js"
	if _, err := os.Stat(path); err != nil {
		b.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(path))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Load a simple plugin
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		b.Skip("Test data plugin directory not found")
	}

	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	_, ok := result.(*Plugin)
	if !ok {
		b.Fatalf("Failed to load plugin: %v", result)
	}

	// Create batch of 10 calls
	calls := make([]BatchCall, 10)
	for i := 0; i < 10; i++ {
		calls[i] = BatchCall{
			Key:  "pi:" + string(rune('0'+i)),
			Name: "pi",
			Args: []any{},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := rt.BatchCallFunctions(calls); err != nil {
			b.Fatalf("BatchCallFunctions failed: %v", err)
		}
	}
}

