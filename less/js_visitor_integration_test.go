package less_go

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toakleaf/less.go/less/runtime"
)

// TestJSVisitorIntegration_WithNodeJS tests the full visitor flow with the Node.js runtime.
// This is an integration test that requires Node.js to be available.
func TestJSVisitorIntegration_WithNodeJS(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("LESS_GO_INTEGRATION") != "1" {
		t.Skip("Skipping integration test (set LESS_GO_INTEGRATION=1 to run)")
	}

	// Get the path to plugin-host.js
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	pluginHostPath := filepath.Join(cwd, "runtime", "plugin-host.js")
	if _, err := os.Stat(pluginHostPath); os.IsNotExist(err) {
		t.Fatalf("plugin-host.js not found at %s", pluginHostPath)
	}

	// Create and start the Node.js runtime
	rt, err := runtime.NewNodeJSRuntime(runtime.WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create Node.js runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start Node.js runtime: %v", err)
	}
	defer rt.Stop()

	// Create the plugin loader
	loader := runtime.NewJSPluginLoader(rt)

	// Load the plugin-preeval.js plugin
	pluginPath := filepath.Join(cwd, "..", "testdata", "plugin", "plugin-preeval.js")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skipf("Plugin not found at %s (run from project root)", pluginPath)
	}

	result := loader.LoadPluginSync(pluginPath, "", nil, nil, nil)
	if err, ok := result.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	plugin, ok := result.(*runtime.Plugin)
	if !ok {
		t.Fatalf("Unexpected result type: %T", result)
	}

	t.Logf("Loaded plugin: %s (visitors: %d)", plugin.Path, plugin.Visitors)

	if plugin.Visitors == 0 {
		t.Error("Expected at least 1 visitor to be registered")
	}

	// Create the visitor registry
	registry := NewJSVisitorRegistry(rt)

	// Refresh visitors from Node.js
	if err := registry.RefreshFromNodeJS(); err != nil {
		t.Fatalf("Failed to refresh visitors: %v", err)
	}

	adapters := registry.GetAdapters()
	t.Logf("Got %d visitor adapters", len(adapters))

	if len(adapters) == 0 {
		t.Error("Expected at least 1 visitor adapter")
	}

	// Check that we have a pre-eval visitor
	preEvalAdapters := registry.GetPreEvalAdapters()
	if len(preEvalAdapters) == 0 {
		t.Error("Expected at least 1 pre-eval visitor adapter")
	}

	// Test registering with plugin manager
	pm := NewPluginManager(nil)
	registry.RegisterWithPluginManager(pm)

	visitors := pm.GetVisitors()
	if len(visitors) != len(adapters) {
		t.Errorf("PluginManager has %d visitors, expected %d", len(visitors), len(adapters))
	}
}

// TestJSVisitorIntegration_VisitorExecution tests running a visitor on an AST.
func TestJSVisitorIntegration_VisitorExecution(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("LESS_GO_INTEGRATION") != "1" {
		t.Skip("Skipping integration test (set LESS_GO_INTEGRATION=1 to run)")
	}

	// Get the path to plugin-host.js
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	pluginHostPath := filepath.Join(cwd, "runtime", "plugin-host.js")
	if _, err := os.Stat(pluginHostPath); os.IsNotExist(err) {
		t.Fatalf("plugin-host.js not found at %s", pluginHostPath)
	}

	// Create and start the Node.js runtime
	rt, err := runtime.NewNodeJSRuntime(runtime.WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create Node.js runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start Node.js runtime: %v", err)
	}
	defer rt.Stop()

	// Create a simple test AST - a ruleset with a variable
	variable := NewVariable("@replace", 0, nil)
	expr, _ := NewExpression([]any{variable}, false)
	decl, _ := NewDeclaration("color", expr, false, false, 0, nil, false, nil)
	ruleset := NewRuleset(nil, []any{decl}, false, nil)

	// Create visitor manager
	vm := runtime.NewVisitorManager(rt)

	// Run pre-eval visitors (none should be registered yet)
	result, err := vm.RunPreEvalVisitors(ruleset)
	if err != nil {
		// This is expected since no plugin is loaded
		t.Logf("RunPreEvalVisitors returned error (expected): %v", err)
	} else {
		t.Logf("RunPreEvalVisitors result: visitorCount=%d", result.VisitorCount)
	}

	// Load the plugin
	loader := runtime.NewJSPluginLoader(rt)
	pluginPath := filepath.Join(cwd, "..", "testdata", "plugin", "plugin-preeval.js")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skipf("Plugin not found at %s", pluginPath)
	}

	pluginResult := loader.LoadPluginSync(pluginPath, "", nil, nil, nil)
	if err, ok := pluginResult.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Now run pre-eval visitors again
	result, err = vm.RunPreEvalVisitors(ruleset)
	if err != nil {
		t.Fatalf("RunPreEvalVisitors failed after loading plugin: %v", err)
	}

	t.Logf("RunPreEvalVisitors result after loading plugin: visitorCount=%d, replacements=%d",
		result.VisitorCount, len(result.Replacements))

	if result.VisitorCount == 0 {
		t.Error("Expected at least 1 visitor to run")
	}
}

// TestJSVisitorAdapter_RunWithNodeJS tests the adapter's Run method with real Node.js.
func TestJSVisitorAdapter_RunWithNodeJS(t *testing.T) {
	// Skip if not in integration test mode
	if os.Getenv("LESS_GO_INTEGRATION") != "1" {
		t.Skip("Skipping integration test (set LESS_GO_INTEGRATION=1 to run)")
	}

	// Get the path to plugin-host.js
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	pluginHostPath := filepath.Join(cwd, "runtime", "plugin-host.js")
	if _, err := os.Stat(pluginHostPath); os.IsNotExist(err) {
		t.Fatalf("plugin-host.js not found at %s", pluginHostPath)
	}

	// Create and start the Node.js runtime
	rt, err := runtime.NewNodeJSRuntime(runtime.WithPluginHostPath(pluginHostPath))
	if err != nil {
		t.Fatalf("Failed to create Node.js runtime: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Failed to start Node.js runtime: %v", err)
	}
	defer rt.Stop()

	// Load the plugin
	loader := runtime.NewJSPluginLoader(rt)
	pluginPath := filepath.Join(cwd, "..", "testdata", "plugin", "plugin-preeval.js")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Skipf("Plugin not found at %s", pluginPath)
	}

	pluginResult := loader.LoadPluginSync(pluginPath, "", nil, nil, nil)
	if err, ok := pluginResult.(error); ok {
		t.Fatalf("Failed to load plugin: %v", err)
	}

	// Create registry and get adapters
	registry := NewJSVisitorRegistry(rt)
	if err := registry.RefreshFromNodeJS(); err != nil {
		t.Fatalf("Failed to refresh visitors: %v", err)
	}

	adapters := registry.GetPreEvalAdapters()
	if len(adapters) == 0 {
		t.Fatal("Expected at least 1 pre-eval visitor adapter")
	}

	// Create a simple test AST
	variable := NewVariable("@replace", 0, nil)
	expr, _ := NewExpression([]any{variable}, false)
	decl, _ := NewDeclaration("color", expr, false, false, 0, nil, false, nil)
	ruleset := NewRuleset(nil, []any{decl}, false, nil)

	// Run the adapter
	adapter := adapters[0]
	result := adapter.Run(ruleset)

	if result == nil {
		t.Error("Run() returned nil")
	}

	// The result should be the same root (replacements are logged but not yet applied)
	t.Logf("Run() completed, result type: %T", result)
}
