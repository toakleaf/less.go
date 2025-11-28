package less_go

import (
	"testing"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// TestJSVisitorAdapter_IsPreEvalVisitor tests the IsPreEvalVisitor method.
func TestJSVisitorAdapter_IsPreEvalVisitor(t *testing.T) {
	tests := []struct {
		name     string
		isPreEval bool
		expected bool
	}{
		{"pre-eval visitor", true, true},
		{"post-eval visitor", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visitor := &runtime.JSVisitor{
				Index:            0,
				IsPreEvalVisitor: tt.isPreEval,
				IsReplacing:      false,
			}
			adapter := NewJSVisitorAdapter(visitor, nil)

			if got := adapter.IsPreEvalVisitor(); got != tt.expected {
				t.Errorf("IsPreEvalVisitor() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestJSVisitorAdapter_IsReplacing tests the IsReplacing method.
func TestJSVisitorAdapter_IsReplacing(t *testing.T) {
	tests := []struct {
		name       string
		isReplacing bool
		expected   bool
	}{
		{"replacing visitor", true, true},
		{"non-replacing visitor", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visitor := &runtime.JSVisitor{
				Index:            0,
				IsPreEvalVisitor: false,
				IsReplacing:      tt.isReplacing,
			}
			adapter := NewJSVisitorAdapter(visitor, nil)

			if got := adapter.IsReplacing(); got != tt.expected {
				t.Errorf("IsReplacing() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestJSVisitorAdapter_IsPreVisitor tests that JS visitors are not "pre" visitors.
func TestJSVisitorAdapter_IsPreVisitor(t *testing.T) {
	visitor := &runtime.JSVisitor{
		Index:            0,
		IsPreEvalVisitor: true,
		IsReplacing:      false,
	}
	adapter := NewJSVisitorAdapter(visitor, nil)

	if got := adapter.IsPreVisitor(); got != false {
		t.Errorf("IsPreVisitor() = %v, want false", got)
	}
}

// TestJSVisitorAdapter_RunWithNilVisitor tests that Run handles nil visitor gracefully.
func TestJSVisitorAdapter_RunWithNilVisitor(t *testing.T) {
	adapter := &JSVisitorAdapter{
		visitor: nil,
		runtime: nil,
	}

	root := &Ruleset{}
	result := adapter.Run(root)

	// Should return root unchanged when visitor is nil
	if result != root {
		t.Errorf("Run() with nil visitor should return root unchanged")
	}
}

// TestJSVisitorAdapter_RunWithNilRuntime tests that Run handles nil runtime gracefully.
func TestJSVisitorAdapter_RunWithNilRuntime(t *testing.T) {
	visitor := &runtime.JSVisitor{
		Index:            0,
		IsPreEvalVisitor: false,
		IsReplacing:      false,
	}
	adapter := &JSVisitorAdapter{
		visitor: visitor,
		runtime: nil,
	}

	root := &Ruleset{}
	result := adapter.Run(root)

	// Should return root unchanged when runtime is nil
	if result != root {
		t.Errorf("Run() with nil runtime should return root unchanged")
	}
}

// TestNewJSVisitorAdapter tests the constructor.
func TestNewJSVisitorAdapter(t *testing.T) {
	visitor := &runtime.JSVisitor{
		Index:            1,
		IsPreEvalVisitor: true,
		IsReplacing:      true,
	}

	adapter := NewJSVisitorAdapter(visitor, nil)

	if adapter == nil {
		t.Fatal("NewJSVisitorAdapter returned nil")
	}
	if adapter.visitor != visitor {
		t.Error("adapter.visitor not set correctly")
	}
	if !adapter.IsPreEvalVisitor() {
		t.Error("IsPreEvalVisitor should be true")
	}
	if !adapter.IsReplacing() {
		t.Error("IsReplacing should be true")
	}
}

// TestJSVisitorRegistry tests the registry functionality.
func TestJSVisitorRegistry(t *testing.T) {
	registry := NewJSVisitorRegistry(nil)

	if registry == nil {
		t.Fatal("NewJSVisitorRegistry returned nil")
	}
	if registry.adapters == nil {
		t.Error("adapters slice should be initialized")
	}
	if len(registry.adapters) != 0 {
		t.Errorf("initial adapters length = %d, want 0", len(registry.adapters))
	}
}

// TestJSVisitorRegistry_GetAdapters tests getting adapters.
func TestJSVisitorRegistry_GetAdapters(t *testing.T) {
	registry := NewJSVisitorRegistry(nil)

	adapters := registry.GetAdapters()
	if adapters == nil {
		t.Error("GetAdapters should not return nil")
	}
	if len(adapters) != 0 {
		t.Errorf("initial adapters length = %d, want 0", len(adapters))
	}
}

// TestJSVisitorRegistry_GetPreEvalAdapters tests filtering pre-eval adapters.
func TestJSVisitorRegistry_GetPreEvalAdapters(t *testing.T) {
	registry := NewJSVisitorRegistry(nil)

	// Manually add some adapters
	preEval := &JSVisitorAdapter{
		visitor: &runtime.JSVisitor{IsPreEvalVisitor: true},
	}
	postEval := &JSVisitorAdapter{
		visitor: &runtime.JSVisitor{IsPreEvalVisitor: false},
	}
	registry.adapters = []*JSVisitorAdapter{preEval, postEval}

	preAdapters := registry.GetPreEvalAdapters()
	if len(preAdapters) != 1 {
		t.Errorf("GetPreEvalAdapters() length = %d, want 1", len(preAdapters))
	}
	if preAdapters[0] != preEval {
		t.Error("GetPreEvalAdapters() returned wrong adapter")
	}
}

// TestJSVisitorRegistry_GetPostEvalAdapters tests filtering post-eval adapters.
func TestJSVisitorRegistry_GetPostEvalAdapters(t *testing.T) {
	registry := NewJSVisitorRegistry(nil)

	// Manually add some adapters
	preEval := &JSVisitorAdapter{
		visitor: &runtime.JSVisitor{IsPreEvalVisitor: true},
	}
	postEval := &JSVisitorAdapter{
		visitor: &runtime.JSVisitor{IsPreEvalVisitor: false},
	}
	registry.adapters = []*JSVisitorAdapter{preEval, postEval}

	postAdapters := registry.GetPostEvalAdapters()
	if len(postAdapters) != 1 {
		t.Errorf("GetPostEvalAdapters() length = %d, want 1", len(postAdapters))
	}
	if postAdapters[0] != postEval {
		t.Error("GetPostEvalAdapters() returned wrong adapter")
	}
}

// TestJSVisitorRegistry_RegisterWithPluginManager tests registration with PluginManager.
func TestJSVisitorRegistry_RegisterWithPluginManager(t *testing.T) {
	registry := NewJSVisitorRegistry(nil)
	pm := NewPluginManager(nil)

	// Add an adapter
	adapter := &JSVisitorAdapter{
		visitor: &runtime.JSVisitor{IsPreEvalVisitor: true},
	}
	registry.adapters = []*JSVisitorAdapter{adapter}

	// Register with plugin manager
	registry.RegisterWithPluginManager(pm)

	// Check that the adapter was added
	visitors := pm.GetVisitors()
	if len(visitors) != 1 {
		t.Errorf("PluginManager visitors length = %d, want 1", len(visitors))
	}
}
