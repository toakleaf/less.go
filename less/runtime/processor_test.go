//go:build !windows

package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

// findPluginHostJS finds the plugin-host.js file for testing.
func findPluginHostJS() string {
	// Look in common locations
	candidates := []string{
		"plugin-host.js",
		filepath.Join("runtime", "plugin-host.js"),
	}

	// Check relative to current directory
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}

	// Check relative to test directory
	cwd, _ := os.Getwd()
	testPath := filepath.Join(cwd, "plugin-host.js")
	if _, err := os.Stat(testPath); err == nil {
		return testPath
	}

	return ""
}

func TestProcessorManager_RefreshProcessors(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	pm := NewProcessorManager(rt)

	// Initially should have no processors
	if err := pm.RefreshProcessors(); err != nil {
		t.Fatalf("RefreshProcessors failed: %v", err)
	}

	if pm.PreProcessorCount() != 0 {
		t.Errorf("Expected 0 pre-processors, got %d", pm.PreProcessorCount())
	}

	if pm.PostProcessorCount() != 0 {
		t.Errorf("Expected 0 post-processors, got %d", pm.PostProcessorCount())
	}
}

func TestProcessorManager_GetProcessors(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	pm := NewProcessorManager(rt)

	// GetPreProcessors should work even without refresh
	procs := pm.GetPreProcessors()
	if procs == nil {
		t.Error("GetPreProcessors returned nil")
	}

	postProcs := pm.GetPostProcessors()
	if postProcs == nil {
		t.Error("GetPostProcessors returned nil")
	}
}

func TestJSPreProcessor_NilRuntime(t *testing.T) {
	proc := NewJSPreProcessor(nil, 0, 1000)

	_, err := proc.Process("test", nil)
	if err == nil {
		t.Error("Expected error for nil runtime")
	}
}

func TestJSPostProcessor_NilRuntime(t *testing.T) {
	proc := NewJSPostProcessor(nil, 0, 1000)

	_, err := proc.Process("test", nil)
	if err == nil {
		t.Error("Expected error for nil runtime")
	}
}

func TestProcessorManager_RunPreProcessors_Empty(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	pm := NewProcessorManager(rt)
	if err := pm.RefreshProcessors(); err != nil {
		t.Fatalf("RefreshProcessors failed: %v", err)
	}

	// With no processors, input should pass through unchanged
	input := ".test { color: red; }"
	output, err := pm.RunPreProcessors(input, nil)
	if err != nil {
		t.Fatalf("RunPreProcessors failed: %v", err)
	}

	if output != input {
		t.Errorf("Expected input to pass through unchanged, got %q", output)
	}
}

func TestProcessorManager_RunPostProcessors_Empty(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	pm := NewProcessorManager(rt)
	if err := pm.RefreshProcessors(); err != nil {
		t.Fatalf("RefreshProcessors failed: %v", err)
	}

	// With no processors, CSS should pass through unchanged
	css := ".test { color: red; }"
	output, err := pm.RunPostProcessors(css, nil)
	if err != nil {
		t.Fatalf("RunPostProcessors failed: %v", err)
	}

	if output != css {
		t.Errorf("Expected CSS to pass through unchanged, got %q", output)
	}
}

func TestProcessorManager_WithPlugin(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Create a temporary test plugin that adds a pre-processor
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-processor-plugin.js")
	pluginCode := `
module.exports = {
    install(less, pluginManager, functions) {
        // Add a pre-processor that adds a comment
        pluginManager.addPreProcessor({
            process(src, extra) {
                return '/* pre-processed */\n' + src;
            }
        }, 500);

        // Add a post-processor that adds a comment at the end
        pluginManager.addPostProcessor({
            process(css, extra) {
                return css + '\n/* post-processed */';
            }
        }, 500);
    }
};
`
	if err := os.WriteFile(pluginPath, []byte(pluginCode), 0644); err != nil {
		t.Fatalf("Failed to write test plugin: %v", err)
	}

	// Load the plugin
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPluginSync(pluginPath, tmpDir, nil, nil, nil)
	if err, ok := result.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Create processor manager and refresh
	pm := NewProcessorManager(rt)
	if err := pm.RefreshProcessors(); err != nil {
		t.Fatalf("RefreshProcessors failed: %v", err)
	}

	// Verify processors were registered
	if pm.PreProcessorCount() != 1 {
		t.Errorf("Expected 1 pre-processor, got %d", pm.PreProcessorCount())
	}

	if pm.PostProcessorCount() != 1 {
		t.Errorf("Expected 1 post-processor, got %d", pm.PostProcessorCount())
	}

	// Test pre-processor
	input := ".test { color: red; }"
	preOutput, err := pm.RunPreProcessors(input, nil)
	if err != nil {
		t.Fatalf("RunPreProcessors failed: %v", err)
	}

	expectedPre := "/* pre-processed */\n" + input
	if preOutput != expectedPre {
		t.Errorf("Pre-processor output mismatch:\nExpected: %q\nGot: %q", expectedPre, preOutput)
	}

	// Test post-processor
	css := ".test { color: red; }"
	postOutput, err := pm.RunPostProcessors(css, nil)
	if err != nil {
		t.Fatalf("RunPostProcessors failed: %v", err)
	}

	expectedPost := css + "\n/* post-processed */"
	if postOutput != expectedPost {
		t.Errorf("Post-processor output mismatch:\nExpected: %q\nGot: %q", expectedPost, postOutput)
	}
}

func TestProcessorManager_Priority(t *testing.T) {
	pluginHostPath := findPluginHostJS()
	if pluginHostPath == "" {
		t.Skip("plugin-host.js not found")
	}

	rt, err := NewNodeJSRuntime(WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer rt.Stop()

	// Create a temporary test plugin with multiple processors at different priorities
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, "test-priority-plugin.js")
	pluginCode := `
module.exports = {
    install(less, pluginManager, functions) {
        // Add pre-processors with different priorities
        // Lower priority runs first
        pluginManager.addPreProcessor({
            process(src, extra) {
                return src + ' [p2000]';
            }
        }, 2000);

        pluginManager.addPreProcessor({
            process(src, extra) {
                return src + ' [p1000]';
            }
        }, 1000);

        pluginManager.addPreProcessor({
            process(src, extra) {
                return src + ' [p500]';
            }
        }, 500);
    }
};
`
	if err := os.WriteFile(pluginPath, []byte(pluginCode), 0644); err != nil {
		t.Fatalf("Failed to write test plugin: %v", err)
	}

	// Load the plugin
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPluginSync(pluginPath, tmpDir, nil, nil, nil)
	if err, ok := result.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Create processor manager and refresh
	pm := NewProcessorManager(rt)
	if err := pm.RefreshProcessors(); err != nil {
		t.Fatalf("RefreshProcessors failed: %v", err)
	}

	// Verify all processors were registered
	if pm.PreProcessorCount() != 3 {
		t.Errorf("Expected 3 pre-processors, got %d", pm.PreProcessorCount())
	}

	// Test that processors run in priority order
	input := "start"
	output, err := pm.RunPreProcessors(input, nil)
	if err != nil {
		t.Fatalf("RunPreProcessors failed: %v", err)
	}

	// Priority 500 runs first, then 1000, then 2000
	expected := "start [p500] [p1000] [p2000]"
	if output != expected {
		t.Errorf("Priority order mismatch:\nExpected: %q\nGot: %q", expected, output)
	}
}
