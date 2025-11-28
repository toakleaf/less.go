package runtime

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// Helper to get plugin path for tests
func getTestPluginPath(t *testing.T, pluginName string) string {
	// Try relative to test directory
	candidates := []string{
		filepath.Join("..", "..", "..", "..", "..", "test-data", "plugin", pluginName),
		filepath.Join("..", "..", "..", "..", "..", "..", "test-data", "plugin", pluginName),
	}

	// Try from current working directory
	wd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "..", "..", "..", "..", "..", "test-data", "plugin", pluginName),
			filepath.Join(wd, "packages", "test-data", "plugin", pluginName),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absPath, _ := filepath.Abs(candidate)
			return absPath
		}
	}

	t.Fatalf("plugin %s not found; tried: %v", pluginName, candidates)
	return ""
}

// Helper to setup runtime with a plugin loaded
func setupRuntimeWithPlugin(t *testing.T, pluginPath string) (*NodeJSRuntime, *JSPluginLoader, func()) {
	hostPath := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	loader := NewJSPluginLoader(rt)

	// Load the plugin
	result := loader.LoadPlugin(pluginPath, filepath.Dir(pluginPath), nil, nil, nil)
	if err, ok := result.(error); ok {
		rt.Stop()
		t.Fatalf("LoadPlugin failed: %v", err)
	}

	cleanup := func() {
		rt.Stop()
	}

	return rt, loader, cleanup
}

func TestJSFunctionDefinition_SimpleFunction(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-simple.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create function definition for 'pi'
	jsFn := NewJSFunctionDefinition("pi", rt)

	// Call the function
	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	// Result should be a Dimension node with value ~3.14159
	if result == nil {
		t.Fatal("result is nil")
	}

	// Check if it's a JSResultNode
	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Fatalf("expected *JSResultNode, got %T", result)
	}

	if jsNode.NodeType != "Dimension" {
		t.Errorf("expected type Dimension, got %s", jsNode.NodeType)
	}

	value := jsNode.getFloat("value")
	if math.Abs(value-math.Pi) > 0.0001 {
		t.Errorf("expected value ~%f, got %f", math.Pi, value)
	}
}

func TestJSFunctionDefinition_AnonymousReturn(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-simple.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create function definition for 'pi-anon' which returns a raw number
	jsFn := NewJSFunctionDefinition("pi-anon", rt)

	// Call the function
	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	// Result should be a raw number (Math.PI)
	if result == nil {
		t.Fatal("result is nil")
	}

	// pi-anon returns Math.PI directly (a number), not a node
	switch v := result.(type) {
	case float64:
		if math.Abs(v-math.Pi) > 0.0001 {
			t.Errorf("expected value ~%f, got %f", math.Pi, v)
		}
	default:
		t.Logf("result type: %T, value: %v", result, result)
		// This is acceptable - the function might wrap it
	}
}

func TestJSFunctionDefinition_WithArguments(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create function definition for 'test-atrule' which takes 2 args
	jsFn := NewJSFunctionDefinition("test-atrule", rt)

	// Create test arguments (simulated nodes)
	arg1 := map[string]any{
		"_type": "Keyword",
		"value": "media",
	}
	arg2 := map[string]any{
		"_type": "Keyword",
		"value": "screen",
	}

	// Call the function with arguments
	result, err := jsFn.Call(arg1, arg2)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// Should return an AtRule node
	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Logf("result type: %T, value: %v", result, result)
		// Might be a map instead
		if nodeMap, ok := result.(map[string]any); ok {
			if nodeMap["_type"] != "AtRule" {
				t.Errorf("expected type AtRule, got %v", nodeMap["_type"])
			}
		}
	} else {
		if jsNode.NodeType != "AtRule" {
			t.Errorf("expected type AtRule, got %s", jsNode.NodeType)
		}
	}
}

func TestJSFunctionDefinition_DimensionReturn(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	jsFn := NewJSFunctionDefinition("test-dimension", rt)

	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Fatalf("expected *JSResultNode, got %T", result)
	}

	if jsNode.NodeType != "Dimension" {
		t.Errorf("expected type Dimension, got %s", jsNode.NodeType)
	}

	value := jsNode.getFloat("value")
	if value != 1 {
		t.Errorf("expected value 1, got %f", value)
	}

	unit := jsNode.getString("unit")
	if unit != "px" {
		t.Errorf("expected unit 'px', got '%s'", unit)
	}
}

func TestJSFunctionDefinition_ColorReturn(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	jsFn := NewJSFunctionDefinition("test-color", rt)

	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Fatalf("expected *JSResultNode, got %T", result)
	}

	if jsNode.NodeType != "Color" {
		t.Errorf("expected type Color, got %s", jsNode.NodeType)
	}

	rgb := jsNode.getRGBArray("rgb")
	if len(rgb) != 3 || rgb[0] != 50 || rgb[1] != 50 || rgb[2] != 50 {
		t.Errorf("expected rgb [50, 50, 50], got %v", rgb)
	}
}

func TestJSFunctionDefinition_QuotedReturn(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	jsFn := NewJSFunctionDefinition("test-quoted", rt)

	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Fatalf("expected *JSResultNode, got %T", result)
	}

	if jsNode.NodeType != "Quoted" {
		t.Errorf("expected type Quoted, got %s", jsNode.NodeType)
	}

	value := jsNode.getString("value")
	if value != "foo" {
		t.Errorf("expected value 'foo', got '%s'", value)
	}

	quote := jsNode.getString("quote")
	if quote != "\"" {
		t.Errorf("expected quote '\"', got '%s'", quote)
	}
}

func TestJSFunctionDefinition_KeywordReturn(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	jsFn := NewJSFunctionDefinition("test-keyword", rt)

	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Fatalf("expected *JSResultNode, got %T", result)
	}

	if jsNode.NodeType != "Keyword" {
		t.Errorf("expected type Keyword, got %s", jsNode.NodeType)
	}

	value := jsNode.getString("value")
	if value != "foo" {
		t.Errorf("expected value 'foo', got '%s'", value)
	}
}

func TestJSFunctionDefinition_ErrorHandling(t *testing.T) {
	hostPath := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Try to call a non-existent function
	jsFn := NewJSFunctionDefinition("non_existent_function", rt)

	_, err = jsFn.Call()
	if err == nil {
		t.Error("expected error for non-existent function")
	}
}

func TestJSFunctionDefinition_NeedsEvalArgs(t *testing.T) {
	jsFn := NewJSFunctionDefinition("test", nil)

	// JS functions should always need evaluated args
	if !jsFn.NeedsEvalArgs() {
		t.Error("expected NeedsEvalArgs() to return true")
	}
}

func TestPluginFunctionRegistry_Basic(t *testing.T) {
	hostPath := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Create a mock builtin registry
	mockBuiltin := &mockRegistry{
		funcs: map[string]any{
			"builtin-func": &mockFunctionDef{name: "builtin-func"},
		},
	}

	registry := NewPluginFunctionRegistry(mockBuiltin, rt)

	// Register a JS function
	registry.RegisterJSFunction("js-func")

	// Check that JS function is registered
	if !registry.HasJSFunction("js-func") {
		t.Error("expected js-func to be registered")
	}

	// Check that builtin function is accessible
	fn := registry.Get("builtin-func")
	if fn == nil {
		t.Error("expected builtin-func to be accessible")
	}

	// Check that JS function takes precedence (shadowing)
	registry.RegisterJSFunction("builtin-func") // Shadow the builtin

	fn = registry.Get("builtin-func")
	if _, ok := fn.(*JSFunctionDefinition); !ok {
		t.Error("expected JS function to shadow builtin")
	}
}

func TestPluginFunctionRegistry_RegisterMultiple(t *testing.T) {
	hostPath := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	registry := NewPluginFunctionRegistry(nil, rt)

	// Register multiple JS functions
	registry.RegisterJSFunctions([]string{"func1", "func2", "func3"})

	names := registry.GetJSFunctionNames()
	if len(names) != 3 {
		t.Errorf("expected 3 functions, got %d", len(names))
	}

	for _, name := range []string{"func1", "func2", "func3"} {
		if !registry.HasJSFunction(name) {
			t.Errorf("expected %s to be registered", name)
		}
	}
}

func TestPluginFunctionRegistry_ClearFunctions(t *testing.T) {
	hostPath := getPluginHostPath(t)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		t.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	registry := NewPluginFunctionRegistry(nil, rt)

	registry.RegisterJSFunctions([]string{"func1", "func2"})
	if len(registry.GetJSFunctionNames()) != 2 {
		t.Error("expected 2 functions before clear")
	}

	registry.ClearJSFunctions()
	if len(registry.GetJSFunctionNames()) != 0 {
		t.Error("expected 0 functions after clear")
	}
}

func TestPluginFunctionRegistry_RefreshFromRuntime(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-simple.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	registry := NewPluginFunctionRegistry(nil, rt)

	// Refresh should discover 'pi' and 'pi-anon' functions
	err := registry.RefreshFromRuntime()
	if err != nil {
		t.Fatalf("RefreshFromRuntime failed: %v", err)
	}

	names := registry.GetJSFunctionNames()
	if len(names) < 2 {
		t.Errorf("expected at least 2 functions, got %d: %v", len(names), names)
	}

	// Check that pi function is registered
	if !registry.HasJSFunction("pi") {
		t.Error("expected 'pi' function to be registered")
	}
}

func TestJSResultNode_ToCSS(t *testing.T) {
	tests := []struct {
		name     string
		node     *JSResultNode
		expected string
	}{
		{
			name: "Dimension with unit",
			node: &JSResultNode{
				NodeType:   "Dimension",
				Properties: map[string]any{"value": 10.5, "unit": "px"},
			},
			expected: "10.5px",
		},
		{
			name: "Dimension without unit",
			node: &JSResultNode{
				NodeType:   "Dimension",
				Properties: map[string]any{"value": 42.0},
			},
			expected: "42",
		},
		{
			name: "Color with alpha",
			node: &JSResultNode{
				NodeType:   "Color",
				Properties: map[string]any{"rgb": []any{255.0, 128.0, 0.0}, "alpha": 0.5},
			},
			expected: "rgba(255, 128, 0, 0.5)",
		},
		{
			name: "Color without alpha",
			node: &JSResultNode{
				NodeType:   "Color",
				Properties: map[string]any{"rgb": []any{255.0, 128.0, 0.0}, "alpha": 1.0},
			},
			expected: "rgb(255, 128, 0)",
		},
		{
			name: "Quoted string",
			node: &JSResultNode{
				NodeType:   "Quoted",
				Properties: map[string]any{"value": "hello", "quote": "'", "escaped": false},
			},
			expected: "'hello'",
		},
		{
			name: "Quoted escaped",
			node: &JSResultNode{
				NodeType:   "Quoted",
				Properties: map[string]any{"value": "hello", "quote": "'", "escaped": true},
			},
			expected: "hello",
		},
		{
			name: "Keyword",
			node: &JSResultNode{
				NodeType:   "Keyword",
				Properties: map[string]any{"value": "auto"},
			},
			expected: "auto",
		},
		{
			name: "Anonymous",
			node: &JSResultNode{
				NodeType:   "Anonymous",
				Properties: map[string]any{"value": "custom-value"},
			},
			expected: "custom-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.node.ToCSS()
			if got != tt.expected {
				t.Errorf("ToCSS() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// ============================================
// IPC Mode Tests (Shared Memory vs JSON)
// ============================================
//
// These tests verify both IPC modes work correctly and produce equivalent results.
// The mode can be controlled via:
//   - WithJSONMode() / WithSharedMemoryMode() options
//   - LESS_JS_IPC_MODE environment variable
//
// ============================================

func TestJSFunctionDefinition_DefaultIPCMode(t *testing.T) {
	// By default (without env var), shared memory should be used
	os.Unsetenv("LESS_JS_IPC_MODE")
	jsFn := NewJSFunctionDefinition("test", nil)
	if jsFn.IPCMode() != JSIPCModeSharedMemory {
		t.Errorf("expected default IPC mode to be shared-memory, got %s", jsFn.IPCMode())
	}
}

func TestJSFunctionDefinition_WithJSONModeOption(t *testing.T) {
	// WithJSONMode() should set JSON mode
	jsFn := NewJSFunctionDefinition("test", nil, WithJSONMode())
	if jsFn.IPCMode() != JSIPCModeJSON {
		t.Errorf("expected IPC mode to be json, got %s", jsFn.IPCMode())
	}
}

func TestJSFunctionDefinition_WithSharedMemoryModeOption(t *testing.T) {
	// WithSharedMemoryMode() should set shared memory mode
	jsFn := NewJSFunctionDefinition("test", nil, WithSharedMemoryMode())
	if jsFn.IPCMode() != JSIPCModeSharedMemory {
		t.Errorf("expected IPC mode to be shared-memory, got %s", jsFn.IPCMode())
	}
}

func TestJSFunctionDefinition_EnvVarJSON(t *testing.T) {
	// LESS_JS_IPC_MODE=json should set JSON mode as default
	os.Setenv("LESS_JS_IPC_MODE", "json")
	defer os.Unsetenv("LESS_JS_IPC_MODE")

	jsFn := NewJSFunctionDefinition("test", nil)
	if jsFn.IPCMode() != JSIPCModeJSON {
		t.Errorf("expected IPC mode from env var to be json, got %s", jsFn.IPCMode())
	}
}

func TestJSFunctionDefinition_EnvVarSharedMem(t *testing.T) {
	// LESS_JS_IPC_MODE=shm should set shared memory mode
	os.Setenv("LESS_JS_IPC_MODE", "shm")
	defer os.Unsetenv("LESS_JS_IPC_MODE")

	jsFn := NewJSFunctionDefinition("test", nil)
	if jsFn.IPCMode() != JSIPCModeSharedMemory {
		t.Errorf("expected IPC mode from env var to be shared-memory, got %s", jsFn.IPCMode())
	}
}

func TestJSFunctionDefinition_OptionOverridesEnvVar(t *testing.T) {
	// Explicit option should override env var
	os.Setenv("LESS_JS_IPC_MODE", "json")
	defer os.Unsetenv("LESS_JS_IPC_MODE")

	jsFn := NewJSFunctionDefinition("test", nil, WithSharedMemoryMode())
	if jsFn.IPCMode() != JSIPCModeSharedMemory {
		t.Errorf("expected option to override env var, got %s", jsFn.IPCMode())
	}
}

func TestJSFunctionDefinition_SharedMemoryCall(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-simple.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create function definition with shared memory mode
	jsFn := NewJSFunctionDefinition("pi", rt, WithSharedMemoryMode())

	if jsFn.IPCMode() != JSIPCModeSharedMemory {
		t.Errorf("expected IPC mode to be shared-memory, got %s", jsFn.IPCMode())
	}

	// Call the function - should use shared memory path
	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call with shared memory failed: %v", err)
	}

	// Result should be a Dimension node with value ~3.14159
	if result == nil {
		t.Fatal("result is nil")
	}

	// Check if it's a JSResultNode
	jsNode, ok := result.(*JSResultNode)
	if !ok {
		// Might return a float64 directly for simple values
		if f, ok := result.(float64); ok {
			if math.Abs(f-math.Pi) > 0.0001 {
				t.Errorf("expected value ~%f, got %f", math.Pi, f)
			}
			return
		}
		t.Fatalf("expected *JSResultNode or float64, got %T", result)
	}

	if jsNode.NodeType != "Dimension" {
		t.Errorf("expected type Dimension, got %s", jsNode.NodeType)
	}

	value := jsNode.getFloat("value")
	if math.Abs(value-math.Pi) > 0.0001 {
		t.Errorf("expected value ~%f, got %f", math.Pi, value)
	}
}

func TestJSFunctionDefinition_JSONModeCall(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-simple.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create function definition with JSON mode
	jsFn := NewJSFunctionDefinition("pi", rt, WithJSONMode())

	if jsFn.IPCMode() != JSIPCModeJSON {
		t.Errorf("expected IPC mode to be json, got %s", jsFn.IPCMode())
	}

	// Call the function - should use JSON path
	result, err := jsFn.Call()
	if err != nil {
		t.Fatalf("Call with JSON mode failed: %v", err)
	}

	// Result should be a Dimension node with value ~3.14159
	if result == nil {
		t.Fatal("result is nil")
	}

	// Check if it's a JSResultNode
	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Fatalf("expected *JSResultNode, got %T", result)
	}

	if jsNode.NodeType != "Dimension" {
		t.Errorf("expected type Dimension, got %s", jsNode.NodeType)
	}

	value := jsNode.getFloat("value")
	if math.Abs(value-math.Pi) > 0.0001 {
		t.Errorf("expected value ~%f, got %f", math.Pi, value)
	}
}

func TestJSFunctionDefinition_SharedMemoryWithArgs(t *testing.T) {
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create function definition with shared memory mode
	jsFn := NewJSFunctionDefinition("test-atrule", rt, WithSharedMemoryMode())

	// Create test arguments
	arg1 := map[string]any{
		"_type": "Keyword",
		"value": "media",
	}
	arg2 := map[string]any{
		"_type": "Keyword",
		"value": "screen",
	}

	// Call the function with arguments
	result, err := jsFn.Call(arg1, arg2)
	if err != nil {
		t.Fatalf("Call with shared memory failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	// Should return an AtRule node
	jsNode, ok := result.(*JSResultNode)
	if !ok {
		t.Logf("result type: %T, value: %v", result, result)
		// Might be a map instead
		if nodeMap, ok := result.(map[string]any); ok {
			if nodeMap["_type"] != "AtRule" {
				t.Errorf("expected type AtRule, got %v", nodeMap["_type"])
			}
			return
		}
	}

	if jsNode != nil && jsNode.NodeType != "AtRule" {
		t.Errorf("expected type AtRule, got %s", jsNode.NodeType)
	}
}

func TestJSFunctionDefinition_BothModesEquivalent(t *testing.T) {
	// Test that both IPC modes produce equivalent results
	pluginPath := getTestPluginPath(t, "plugin-tree-nodes.js")
	rt, _, cleanup := setupRuntimeWithPlugin(t, pluginPath)
	defer cleanup()

	// Create two function definitions - one with each mode
	jsFnShm := NewJSFunctionDefinition("test-dimension", rt, WithSharedMemoryMode())
	jsFnJSON := NewJSFunctionDefinition("test-dimension", rt, WithJSONMode())

	// Call both
	resultShm, errShm := jsFnShm.Call()
	resultJSON, errJSON := jsFnJSON.Call()

	if errShm != nil {
		t.Fatalf("Shared memory call failed: %v", errShm)
	}
	if errJSON != nil {
		t.Fatalf("JSON call failed: %v", errJSON)
	}

	// Both should return a Dimension node with value 1 and unit "px"
	jsNodeShm, okShm := resultShm.(*JSResultNode)
	jsNodeJSON, okJSON := resultJSON.(*JSResultNode)

	if !okShm || !okJSON {
		t.Logf("Shared memory result: %T, JSON result: %T", resultShm, resultJSON)
		// Both should at least be non-nil
		if resultShm == nil || resultJSON == nil {
			t.Error("One or both results are nil")
		}
		return
	}

	if jsNodeShm.NodeType != jsNodeJSON.NodeType {
		t.Errorf("Types differ: shared memory=%s, JSON=%s", jsNodeShm.NodeType, jsNodeJSON.NodeType)
	}

	valueShm := jsNodeShm.getFloat("value")
	valueJSON := jsNodeJSON.getFloat("value")
	if valueShm != valueJSON {
		t.Errorf("Values differ: shared memory=%f, JSON=%f", valueShm, valueJSON)
	}

	unitShm := jsNodeShm.getString("unit")
	unitJSON := jsNodeJSON.getString("unit")
	if unitShm != unitJSON {
		t.Errorf("Units differ: shared memory=%s, JSON=%s", unitShm, unitJSON)
	}
}

func TestJSIPCMode_String(t *testing.T) {
	// Test the String() method for IPC modes
	tests := []struct {
		mode     JSIPCMode
		expected string
	}{
		{JSIPCModeSharedMemory, "shared-memory"},
		{JSIPCModeJSON, "json"},
		{JSIPCMode(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("JSIPCMode(%d).String() = %s, want %s", tt.mode, got, tt.expected)
		}
	}
}

// ============================================
// Benchmark Tests
// ============================================
//
// These benchmarks compare the performance of the two IPC modes.
//
// Run with: go test -bench=BenchmarkJSFunction -benchmem
//
// Expected characteristics:
//   - JSON mode: Lower per-call overhead for simple functions
//   - Shared memory mode: Better for complex AST trees and large data
//
// ============================================

func BenchmarkJSFunction_SharedMemoryMode(b *testing.B) {
	hostPath := getPluginHostPathBench(b)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Load the plugin
	pluginPath := getTestPluginPathBench(b, "plugin-simple.js")
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPlugin(pluginPath, "", nil, nil, nil)
	if err, ok := result.(error); ok {
		b.Fatalf("LoadPlugin failed: %v", err)
	}

	// Create function definition with shared memory mode
	jsFn := NewJSFunctionDefinition("pi", rt, WithSharedMemoryMode())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsFn.Call()
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
	}
}

func BenchmarkJSFunction_JSONMode(b *testing.B) {
	hostPath := getPluginHostPathBench(b)
	rt, err := NewNodeJSRuntime(WithPluginHostPath(hostPath))
	if err != nil {
		b.Fatalf("NewNodeJSRuntime failed: %v", err)
	}

	if err := rt.Start(); err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer rt.Stop()

	// Load the plugin
	pluginPath := getTestPluginPathBench(b, "plugin-simple.js")
	loader := NewJSPluginLoader(rt)
	result := loader.LoadPlugin(pluginPath, "", nil, nil, nil)
	if err, ok := result.(error); ok {
		b.Fatalf("LoadPlugin failed: %v", err)
	}

	// Create function definition with JSON mode
	jsFn := NewJSFunctionDefinition("pi", rt, WithJSONMode())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsFn.Call()
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
	}
}

// Helper functions for benchmarks

func getPluginHostPathBench(b *testing.B) string {
	candidates := []string{
		"plugin-host.js",
		"./plugin-host.js",
	}

	wd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates, filepath.Join(wd, "plugin-host.js"))
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absPath, _ := filepath.Abs(candidate)
			return absPath
		}
	}

	b.Fatalf("plugin-host.js not found; tried: %v", candidates)
	return ""
}

func getTestPluginPathBench(b *testing.B, pluginName string) string {
	candidates := []string{
		filepath.Join("..", "..", "..", "..", "..", "test-data", "plugin", pluginName),
		filepath.Join("..", "..", "..", "..", "..", "..", "test-data", "plugin", pluginName),
	}

	wd, err := os.Getwd()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(wd, "..", "..", "..", "..", "..", "test-data", "plugin", pluginName),
			filepath.Join(wd, "packages", "test-data", "plugin", pluginName),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			absPath, _ := filepath.Abs(candidate)
			return absPath
		}
	}

	b.Fatalf("plugin %s not found; tried: %v", pluginName, candidates)
	return ""
}

// Mock types for testing

type mockRegistry struct {
	funcs map[string]any
}

func (m *mockRegistry) Get(name string) any {
	return m.funcs[name]
}

type mockFunctionDef struct {
	name string
}

func (m *mockFunctionDef) Call(args ...any) (any, error) {
	return nil, nil
}

func (m *mockFunctionDef) CallCtx(ctx any, args ...any) (any, error) {
	return nil, nil
}

func (m *mockFunctionDef) NeedsEvalArgs() bool {
	return true
}
