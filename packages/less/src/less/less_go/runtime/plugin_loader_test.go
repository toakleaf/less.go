package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findTestDataPluginDir finds the plugin test data directory
func findTestDataPluginDir() string {
	// Start from current working directory and go up
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Check different possible relative paths
	candidates := []string{
		filepath.Join(cwd, "..", "..", "..", "..", "testdata", "plugin"),
		filepath.Join(cwd, "packages", "testdata", "plugin"),
		"/home/user/less.go/packages/testdata/plugin",
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}

func TestJSPluginLoader_LoadSimplePlugin(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	// Create and start runtime
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create plugin loader
	loader := NewJSPluginLoader(rt)
	if loader == nil {
		t.Fatal("NewJSPluginLoader returned nil")
	}

	// Load plugin-simple.js
	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	// Should return a Plugin, not an error
	plugin, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T: %v", result, result)
	}
	if plugin == nil {
		t.Fatal("Plugin is nil")
	}

	// Check that functions were registered
	hasPi := false
	hasPiAnon := false
	for _, fn := range plugin.Functions {
		if fn == "pi" {
			hasPi = true
		}
		if fn == "pi-anon" {
			hasPiAnon = true
		}
	}
	if !hasPi {
		t.Error("Expected 'pi' function to be registered")
	}
	if !hasPiAnon {
		t.Error("Expected 'pi-anon' function to be registered")
	}

	if plugin.Path != pluginPath {
		t.Errorf("Plugin.Path = %q, want %q", plugin.Path, pluginPath)
	}
	if plugin.Cached {
		t.Error("First load should not be cached")
	}
}

func TestJSPluginLoader_LoadTreeNodesPlugin(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load plugin-tree-nodes.js
	pluginPath := filepath.Join(pluginDir, "plugin-tree-nodes.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	plugin, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T: %v", result, result)
	}
	if plugin == nil {
		t.Fatal("Plugin is nil")
	}

	// Check that multiple functions were registered
	if len(plugin.Functions) < 5 {
		t.Errorf("Expected multiple functions from tree-nodes plugin, got %d", len(plugin.Functions))
	}

	funcNames := strings.Join(plugin.Functions, ", ")
	t.Logf("Registered functions: %s", funcNames)
}

func TestJSPluginLoader_PluginCaching(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load the same plugin twice
	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")

	result1 := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)
	plugin1, ok := result1.(*Plugin)
	if !ok {
		t.Fatalf("First load: Expected *Plugin, got %T", result1)
	}
	if plugin1.Cached {
		t.Error("First load should not be cached")
	}

	result2 := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)
	plugin2, ok := result2.(*Plugin)
	if !ok {
		t.Fatalf("Second load: Expected *Plugin, got %T", result2)
	}
	if !plugin2.Cached {
		t.Error("Second load should be cached")
	}
}

func TestJSPluginLoader_ErrorHandling(t *testing.T) {
	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Try to load non-existent plugin
	result := loader.LoadPlugin("/nonexistent/plugin.js", "/tmp", nil, nil, nil)

	// Should return an error
	loadErr, ok := result.(error)
	if !ok {
		t.Fatalf("Expected error, got %T: %v", result, result)
	}
	if loadErr == nil {
		t.Fatal("Error is nil")
	}
	if !strings.Contains(loadErr.Error(), "Failed to load plugin") {
		t.Errorf("Error message should contain 'Failed to load plugin': %v", loadErr)
	}
}

func TestJSPluginLoader_RelativePath(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load using relative path
	result := loader.LoadPlugin("./plugin-simple.js", pluginDir, nil, nil, nil)

	plugin, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T: %v", result, result)
	}
	if plugin == nil {
		t.Fatal("Plugin is nil")
	}

	hasPi := false
	for _, fn := range plugin.Functions {
		if fn == "pi" {
			hasPi = true
			break
		}
	}
	if !hasPi {
		t.Error("Expected 'pi' function to be registered")
	}
}

func TestJSPluginLoader_GetRegisteredFunctions(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load a plugin first
	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)
	_, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T", result)
	}

	// Get all registered functions
	functions, err := loader.GetRegisteredFunctions()
	if err != nil {
		t.Fatalf("GetRegisteredFunctions failed: %v", err)
	}

	hasPi := false
	hasPiAnon := false
	for _, fn := range functions {
		if fn == "pi" {
			hasPi = true
		}
		if fn == "pi-anon" {
			hasPiAnon = true
		}
	}
	if !hasPi {
		t.Error("Expected 'pi' function in registry")
	}
	if !hasPiAnon {
		t.Error("Expected 'pi-anon' function in registry")
	}
}

func TestJSPluginLoader_CallFunction(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load plugin
	pluginPath := filepath.Join(pluginDir, "plugin-simple.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)
	_, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T", result)
	}

	// Call the pi function
	piResult, err := loader.CallFunction("pi", nil)
	if err != nil {
		t.Fatalf("CallFunction failed: %v", err)
	}
	if piResult == nil {
		t.Fatal("pi function returned nil")
	}

	// The result should be a Dimension node with value ~3.14
	resultMap, ok := piResult.(map[string]any)
	if !ok {
		t.Fatalf("Expected map, got %T: %v", piResult, piResult)
	}

	// Check it's a Dimension type
	if resultMap["_type"] != "Dimension" {
		t.Errorf("Expected _type = Dimension, got %v", resultMap["_type"])
	}

	// Check value is approximately pi
	if value, ok := resultMap["value"].(float64); ok {
		if value < 3.14 || value > 3.15 {
			t.Errorf("Expected value ~3.14, got %v", value)
		}
	} else {
		t.Errorf("Expected value to be float64, got %T", resultMap["value"])
	}
}

func TestJSPluginLoader_LoadPreEvalPlugin(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load plugin-preeval.js (uses module.exports with install method)
	pluginPath := filepath.Join(pluginDir, "plugin-preeval.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	plugin, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T: %v", result, result)
	}
	if plugin == nil {
		t.Fatal("Plugin is nil")
	}

	// This plugin registers a visitor
	if plugin.Visitors != 1 {
		t.Errorf("plugin-preeval should register 1 visitor, got %d", plugin.Visitors)
	}
}

func TestJSPluginLoader_LoadGlobalPlugin(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load plugin-global.js (uses global functions.addMultiple and tree.Anonymous)
	pluginPath := filepath.Join(pluginDir, "plugin-global.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)

	plugin, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T: %v", result, result)
	}
	if plugin == nil {
		t.Fatal("Plugin is nil")
	}

	// Should have registered functions
	hasTestShadow := false
	hasTestGlobal := false
	for _, fn := range plugin.Functions {
		if fn == "test-shadow" {
			hasTestShadow = true
		}
		if fn == "test-global" {
			hasTestGlobal = true
		}
	}
	if !hasTestShadow {
		t.Error("Expected 'test-shadow' function to be registered")
	}
	if !hasTestGlobal {
		t.Error("Expected 'test-global' function to be registered")
	}
}

func TestJSPluginLoader_GetVisitors(t *testing.T) {
	pluginDir := findTestDataPluginDir()
	if pluginDir == "" {
		t.Skip("Test data plugin directory not found")
	}

	rt, err := NewNodeJSRuntime()
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	err = rt.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	loader := NewJSPluginLoader(rt)

	// Load plugin-preeval.js which registers a visitor
	pluginPath := filepath.Join(pluginDir, "plugin-preeval.js")
	result := loader.LoadPlugin(pluginPath, pluginDir, nil, nil, nil)
	_, ok := result.(*Plugin)
	if !ok {
		t.Fatalf("Expected *Plugin, got %T", result)
	}

	// Get visitors
	visitors, err := loader.GetVisitors()
	if err != nil {
		t.Fatalf("GetVisitors failed: %v", err)
	}
	if len(visitors) != 1 {
		t.Fatalf("Expected 1 visitor, got %d", len(visitors))
	}

	// Check visitor properties
	v := visitors[0]
	if isPreEval, ok := v["isPreEvalVisitor"].(bool); !ok || !isPreEval {
		t.Error("Should be a pre-eval visitor")
	}
	if isReplacing, ok := v["isReplacing"].(bool); !ok || !isReplacing {
		t.Error("Should be a replacing visitor")
	}
}
