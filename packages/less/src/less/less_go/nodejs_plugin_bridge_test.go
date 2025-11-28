package less_go

import (
	"testing"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

func TestNodeJSPluginBridge_ScopeManagement(t *testing.T) {
	t.Run("should create root scope on initialization", func(t *testing.T) {
		// Create a bridge without actually starting Node.js
		bridge := &NodeJSPluginBridge{
			scope: runtime.NewRootPluginScope(),
		}

		if bridge.scope == nil {
			t.Error("scope should be initialized")
		}
		if !bridge.scope.IsRoot() {
			t.Error("initial scope should be root")
		}
	})

	t.Run("should enter and exit scopes", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope: runtime.NewRootPluginScope(),
		}

		rootScope := bridge.scope

		// Enter a new scope
		childScope := bridge.EnterScope()
		if childScope == rootScope {
			t.Error("entering scope should create a new scope")
		}
		if !rootScope.IsRoot() {
			t.Error("root should still be root")
		}
		if childScope.IsRoot() {
			t.Error("child should not be root")
		}

		// Exit scope
		returnedScope := bridge.ExitScope()
		if returnedScope != rootScope {
			t.Error("exiting should return to parent scope")
		}
	})

	t.Run("should handle nested scopes", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope: runtime.NewRootPluginScope(),
		}

		level1 := bridge.EnterScope()
		level2 := bridge.EnterScope()
		level3 := bridge.EnterScope()

		// Exit back through the levels
		bridge.ExitScope()
		if bridge.scope != level2 {
			t.Error("should be at level 2")
		}
		bridge.ExitScope()
		if bridge.scope != level1 {
			t.Error("should be at level 1")
		}

		_ = level3 // Mark as used
	})
}

func TestNodeJSPluginBridge_FunctionLookup(t *testing.T) {
	t.Run("should look up functions in scope", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		// Add a function to scope
		fn := &runtime.JSFunctionDefinition{}
		bridge.scope.AddFunction("test-func", fn)

		// Should find it
		found, ok := bridge.LookupFunction("test-func")
		if !ok {
			t.Error("should find function in scope")
		}
		if found != fn {
			t.Error("should return correct function")
		}
	})

	t.Run("should not find non-existent functions", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		_, ok := bridge.LookupFunction("non-existent")
		if ok {
			t.Error("should not find non-existent function")
		}
	})

	t.Run("should check HasFunction correctly", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		bridge.scope.AddFunction("exists", &runtime.JSFunctionDefinition{})

		if !bridge.HasFunction("exists") {
			t.Error("should find existing function")
		}
		if bridge.HasFunction("does-not-exist") {
			t.Error("should not find non-existent function")
		}
	})
}

func TestNodeJSPluginBridge_ScopeInheritance(t *testing.T) {
	t.Run("should inherit functions from parent scope", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		// Add function to root scope
		rootFn := &runtime.JSFunctionDefinition{}
		bridge.scope.AddFunction("root-func", rootFn)

		// Enter child scope
		bridge.EnterScope()

		// Should still find root function
		found, ok := bridge.LookupFunction("root-func")
		if !ok {
			t.Error("should find parent function from child scope")
		}
		if found != rootFn {
			t.Error("should return correct function")
		}
	})

	t.Run("should shadow parent functions with local definitions", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		// Add function to root
		rootFn := &runtime.JSFunctionDefinition{}
		bridge.scope.AddFunction("shadow-test", rootFn)

		// Enter child and add local version
		bridge.EnterScope()
		localFn := &runtime.JSFunctionDefinition{}
		bridge.scope.AddFunction("shadow-test", localFn)

		// Should find local version
		found, _ := bridge.LookupFunction("shadow-test")
		if found != localFn {
			t.Error("child scope should shadow parent function")
		}

		// Exit scope - should find root version again
		bridge.ExitScope()
		found, _ = bridge.LookupFunction("shadow-test")
		if found != rootFn {
			t.Error("after exit, should find parent function")
		}
	})
}

func TestNodeJSPluginBridge_SetScope(t *testing.T) {
	t.Run("should allow setting scope directly", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		// Save original scope
		originalScope := bridge.scope

		// Enter some scopes
		bridge.EnterScope()
		bridge.EnterScope()
		deepScope := bridge.scope

		// Restore to original
		bridge.SetScope(originalScope)
		if bridge.scope != originalScope {
			t.Error("should be able to set scope directly")
		}

		// Set back to deep scope
		bridge.SetScope(deepScope)
		if bridge.scope != deepScope {
			t.Error("should be able to jump to any scope")
		}
	})
}

func TestNodeJSPluginBridge_Visitors(t *testing.T) {
	t.Run("should get visitors from scope", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		visitor1 := &runtime.JSVisitor{IsPreEvalVisitor: true}
		visitor2 := &runtime.JSVisitor{IsPreEvalVisitor: false}
		bridge.scope.AddVisitor(visitor1)
		bridge.scope.AddVisitor(visitor2)

		visitors := bridge.GetVisitors()
		if len(visitors) != 2 {
			t.Errorf("should have 2 visitors, got %d", len(visitors))
		}

		preEval := bridge.GetPreEvalVisitors()
		if len(preEval) != 1 {
			t.Errorf("should have 1 pre-eval visitor, got %d", len(preEval))
		}

		postEval := bridge.GetPostEvalVisitors()
		if len(postEval) != 1 {
			t.Errorf("should have 1 post-eval visitor, got %d", len(postEval))
		}
	})
}

func TestNodeJSPluginBridge_CreateScopedPluginManager(t *testing.T) {
	t.Run("should create ScopedPluginManager", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		bridge.scope.AddVisitor(&runtime.JSVisitor{Index: 1})

		spm := bridge.CreateScopedPluginManager()
		if spm == nil {
			t.Error("should create ScopedPluginManager")
		}

		visitors := spm.GetVisitors()
		if len(visitors) != 1 {
			t.Error("ScopedPluginManager should have access to scope visitors")
		}
	})
}
