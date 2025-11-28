package less_go

import (
	"testing"
)

func TestLazyNodeJSPluginBridge_NotInitializedByDefault(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	// Should not be initialized immediately
	if bridge.IsInitialized() {
		t.Error("Bridge should not be initialized by default")
	}

	// WasUsed should return false
	if bridge.WasUsed() {
		t.Error("WasUsed should return false when not initialized")
	}
}

func TestLazyNodeJSPluginBridge_FunctionLookupWithoutInit(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	// Looking up functions should not initialize the bridge
	fn, found := bridge.LookupFunction("test")
	if found {
		t.Error("Should not find function when bridge is not initialized")
	}
	if fn != nil {
		t.Error("Function should be nil when not found")
	}

	// Bridge should still not be initialized
	if bridge.IsInitialized() {
		t.Error("Function lookup should not initialize bridge")
	}
}

func TestLazyNodeJSPluginBridge_HasFunctionWithoutInit(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	// HasFunction should return false when not initialized
	if bridge.HasFunction("test") {
		t.Error("HasFunction should return false when bridge is not initialized")
	}

	// Bridge should still not be initialized
	if bridge.IsInitialized() {
		t.Error("HasFunction should not initialize bridge")
	}
}

func TestLazyNodeJSPluginBridge_CallFunctionWithoutInit(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	// CallFunction should return error when not initialized
	_, err := bridge.CallFunction("test")
	if err == nil {
		t.Error("CallFunction should return error when bridge is not initialized")
	}

	// Bridge should still not be initialized
	if bridge.IsInitialized() {
		t.Error("CallFunction should not initialize bridge")
	}
}

func TestLazyNodeJSPluginBridge_ScopeMethodsWithoutInit(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	// Scope methods should return nil when not initialized
	if bridge.EnterScope() != nil {
		t.Error("EnterScope should return nil when bridge is not initialized")
	}
	if bridge.ExitScope() != nil {
		t.Error("ExitScope should return nil when bridge is not initialized")
	}
	if bridge.GetScope() != nil {
		t.Error("GetScope should return nil when bridge is not initialized")
	}
	if bridge.GetRuntime() != nil {
		t.Error("GetRuntime should return nil when bridge is not initialized")
	}

	// Bridge should still not be initialized
	if bridge.IsInitialized() {
		t.Error("Scope methods should not initialize bridge")
	}
}

func TestLazyNodeJSPluginBridge_VisitorMethodsWithoutInit(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	// Visitor methods should return nil when not initialized
	if bridge.GetVisitors() != nil {
		t.Error("GetVisitors should return nil when bridge is not initialized")
	}
	if bridge.GetPreEvalVisitors() != nil {
		t.Error("GetPreEvalVisitors should return nil when bridge is not initialized")
	}
	if bridge.GetPostEvalVisitors() != nil {
		t.Error("GetPostEvalVisitors should return nil when bridge is not initialized")
	}

	// Bridge should still not be initialized
	if bridge.IsInitialized() {
		t.Error("Visitor methods should not initialize bridge")
	}
}

func TestLazyNodeJSPluginBridge_CloseWithoutInit(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()

	// Closing without initialization should not error
	err := bridge.Close()
	if err != nil {
		t.Errorf("Close should not error when bridge was not initialized: %v", err)
	}
}

func TestLazyNodeJSPluginBridge_DoubleClose(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()

	// First close
	err := bridge.Close()
	if err != nil {
		t.Errorf("First close should not error: %v", err)
	}

	// Second close should also not error
	err = bridge.Close()
	if err != nil {
		t.Errorf("Second close should not error: %v", err)
	}
}

func TestLazyPluginLoaderFactory(t *testing.T) {
	bridge := NewLazyNodeJSPluginBridge()
	defer bridge.Close()

	factory := LazyPluginLoaderFactory(bridge)
	if factory == nil {
		t.Fatal("Factory should not be nil")
	}

	// Create a loader from the factory
	loader := factory(nil)
	if loader == nil {
		t.Fatal("Loader should not be nil")
	}

	// The loader should be the bridge itself
	if _, ok := loader.(*LazyNodeJSPluginBridge); !ok {
		t.Error("Loader should be a LazyNodeJSPluginBridge")
	}
}

func TestNewLessContextWithPlugins(t *testing.T) {
	ctx, cleanup := NewLessContextWithPlugins(map[string]any{
		"filename": "test.less",
	})

	if ctx == nil {
		t.Fatal("Context should not be nil")
	}

	if cleanup == nil {
		t.Fatal("Cleanup function should not be nil")
	}

	// Context should have plugin bridge
	if ctx.PluginBridge == nil {
		t.Error("Context should have a plugin bridge")
	}

	// Plugin bridge should not be initialized yet
	if ctx.PluginBridge.IsInitialized() {
		t.Error("Plugin bridge should not be initialized immediately")
	}

	// Cleanup should work
	err := cleanup()
	if err != nil {
		t.Errorf("Cleanup should not error: %v", err)
	}
}

func TestLessContext_GetPluginLoader(t *testing.T) {
	ctx, cleanup := NewLessContextWithPlugins(nil)
	defer cleanup()

	factory := ctx.GetPluginLoader()
	if factory == nil {
		t.Fatal("GetPluginLoader should return a factory")
	}

	loader := factory(ctx)
	if loader == nil {
		t.Fatal("Factory should return a loader")
	}
}
