package less_go

import (
	"testing"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// TestPluginScopeIntegration tests the integration between PluginScope and the evaluation context.
// These tests verify that plugin functions are correctly scoped according to LESS semantics:
// - Global plugins affect the entire file
// - Local plugins only affect their scope and children
// - Child scopes can shadow parent functions
// - Visitors from parent scopes are inherited
func TestPluginScopeIntegration(t *testing.T) {
	t.Run("should create root scope in evaluation context", func(t *testing.T) {
		// Create a bridge without actually starting Node.js
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		// Create evaluation context with the bridge
		evalCtx := &Eval{
			PluginBridge: bridge,
		}

		// Verify the bridge is accessible
		if evalCtx.PluginBridge == nil {
			t.Error("PluginBridge should be set in evaluation context")
		}

		// Verify the scope is root
		if !evalCtx.PluginBridge.GetScope().IsRoot() {
			t.Error("Initial scope should be root")
		}
	})

	t.Run("should enter and exit plugin scopes during evaluation", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		evalCtx := &Eval{
			PluginBridge: bridge,
		}

		rootScope := evalCtx.PluginBridge.GetScope()

		// Enter a child scope (simulating entering a ruleset with @plugin)
		childScope := evalCtx.EnterPluginScope()
		if childScope == nil {
			t.Fatal("EnterPluginScope should return a scope")
		}

		// Verify we're now in child scope
		currentScope := evalCtx.PluginBridge.GetScope()
		if currentScope == rootScope {
			t.Error("Should be in child scope, not root")
		}

		// Exit the scope
		evalCtx.ExitPluginScope()

		// Verify we're back in root scope
		currentScope = evalCtx.PluginBridge.GetScope()
		if currentScope != rootScope {
			t.Error("Should be back in root scope")
		}
	})

	t.Run("should lookup plugin functions with proper scoping", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		evalCtx := &Eval{
			PluginBridge: bridge,
		}

		// Add a global function to root scope
		globalFn := &runtime.JSFunctionDefinition{}
		bridge.GetScope().AddFunction("test-global", globalFn)

		// Function should be found
		found, ok := evalCtx.LookupPluginFunction("test-global")
		if !ok {
			t.Error("Should find global function")
		}
		if found != globalFn {
			t.Error("Should return correct function")
		}

		// Enter child scope
		evalCtx.EnterPluginScope()

		// Global function should still be accessible from child
		found, ok = evalCtx.LookupPluginFunction("test-global")
		if !ok {
			t.Error("Should find global function from child scope")
		}

		// Add local function that shadows global
		localFn := &runtime.JSFunctionDefinition{}
		bridge.GetScope().AddFunction("test-global", localFn)

		// Now should find local version
		found, ok = evalCtx.LookupPluginFunction("test-global")
		if !ok {
			t.Error("Should find local function")
		}
		if found != localFn {
			t.Error("Should return local (shadowed) function")
		}

		// Exit back to root
		evalCtx.ExitPluginScope()

		// Should find global version again
		found, ok = evalCtx.LookupPluginFunction("test-global")
		if !ok {
			t.Error("Should find global function after exit")
		}
		if found != globalFn {
			t.Error("Should return global function after exit")
		}
	})

	t.Run("should correctly report HasPluginFunction", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		evalCtx := &Eval{
			PluginBridge: bridge,
		}

		// Initially no functions
		if evalCtx.HasPluginFunction("test-func") {
			t.Error("Should not find non-existent function")
		}

		// Add a function
		bridge.GetScope().AddFunction("test-func", &runtime.JSFunctionDefinition{})

		// Now should find it
		if !evalCtx.HasPluginFunction("test-func") {
			t.Error("Should find existing function")
		}
	})

	t.Run("should preserve scope across NewEvalFromEval", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		parentCtx := &Eval{
			PluginBridge: bridge,
		}

		// Create child context using NewEvalFromEval
		childCtx := NewEvalFromEval(parentCtx, []any{})

		// Child should have same bridge
		if childCtx.PluginBridge != parentCtx.PluginBridge {
			t.Error("Child context should share same PluginBridge")
		}

		// Changes to bridge should be visible from both
		bridge.GetScope().AddFunction("shared-func", &runtime.JSFunctionDefinition{})

		if !parentCtx.HasPluginFunction("shared-func") {
			t.Error("Parent should see shared function")
		}
		if !childCtx.HasPluginFunction("shared-func") {
			t.Error("Child should see shared function")
		}
	})
}

// TestPluginScopeWithoutBridge tests that evaluation works correctly when no plugin bridge is set.
// This is the default case for compilations that don't use JavaScript plugins.
func TestPluginScopeWithoutBridge(t *testing.T) {
	t.Run("should handle nil plugin bridge gracefully", func(t *testing.T) {
		evalCtx := &Eval{
			PluginBridge: nil,
		}

		// EnterPluginScope should not panic
		result := evalCtx.EnterPluginScope()
		if result != nil {
			t.Error("EnterPluginScope with nil bridge should return nil")
		}

		// ExitPluginScope should not panic
		evalCtx.ExitPluginScope()

		// LookupPluginFunction should return false
		_, ok := evalCtx.LookupPluginFunction("any-func")
		if ok {
			t.Error("LookupPluginFunction with nil bridge should return false")
		}

		// HasPluginFunction should return false
		if evalCtx.HasPluginFunction("any-func") {
			t.Error("HasPluginFunction with nil bridge should return false")
		}
	})

	t.Run("should not interfere with normal function lookup", func(t *testing.T) {
		// Create a context without plugin bridge
		evalCtx := &Eval{
			PluginBridge:     nil,
			FunctionRegistry: DefaultRegistry.Inherit(),
		}

		// Normal function registry should still work
		if evalCtx.FunctionRegistry == nil {
			t.Error("Function registry should be available")
		}

		// Built-in functions should be accessible
		fn := evalCtx.FunctionRegistry.Get("rgb")
		if fn == nil {
			t.Error("Built-in function rgb should be accessible")
		}
	})
}

// TestPluginScopeNesting tests deep nesting of plugin scopes.
func TestPluginScopeNesting(t *testing.T) {
	t.Run("should handle deeply nested scopes", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		evalCtx := &Eval{
			PluginBridge: bridge,
		}

		// Add function at each level and track them
		scopes := make([]*runtime.PluginScope, 0)
		scopes = append(scopes, bridge.GetScope())
		bridge.GetScope().AddFunction("level-0", &runtime.JSFunctionDefinition{})

		for i := 1; i <= 5; i++ {
			evalCtx.EnterPluginScope()
			scopes = append(scopes, bridge.GetScope())
			bridge.GetScope().AddFunction("level-"+string(rune('0'+i)), &runtime.JSFunctionDefinition{})
		}

		// Should be able to find all functions from deepest level
		for i := 0; i <= 5; i++ {
			funcName := "level-" + string(rune('0'+i))
			if !evalCtx.HasPluginFunction(funcName) {
				t.Errorf("Should find function %s from deepest level", funcName)
			}
		}

		// Exit back through all levels
		for i := 5; i >= 1; i-- {
			evalCtx.ExitPluginScope()
			// Functions from exited scopes should not be directly accessible
			// (they're still in their scopes, but we've exited)
		}

		// At root level, should only find level-0
		if !evalCtx.HasPluginFunction("level-0") {
			t.Error("Should find level-0 at root")
		}
	})
}

// TestPluginScopeVisitorInheritance tests that visitors are inherited correctly.
func TestPluginScopeVisitorInheritance(t *testing.T) {
	t.Run("should inherit visitors from parent scopes", func(t *testing.T) {
		bridge := &NodeJSPluginBridge{
			scope:        runtime.NewRootPluginScope(),
			funcRegistry: runtime.NewPluginFunctionRegistry(nil, nil),
		}

		// Add visitor to root
		rootVisitor := &runtime.JSVisitor{Index: 1, IsPreEvalVisitor: true}
		bridge.GetScope().AddVisitor(rootVisitor)

		// Enter child scope
		bridge.EnterScope()

		// Add visitor to child
		childVisitor := &runtime.JSVisitor{Index: 2, IsPreEvalVisitor: false}
		bridge.GetScope().AddVisitor(childVisitor)

		// Both visitors should be accessible
		visitors := bridge.GetVisitors()
		if len(visitors) != 2 {
			t.Errorf("Should have 2 visitors, got %d", len(visitors))
		}

		// Exit child scope
		bridge.ExitScope()

		// Only root visitor should be accessible
		visitors = bridge.GetVisitors()
		if len(visitors) != 1 {
			t.Errorf("Should have 1 visitor after exit, got %d", len(visitors))
		}
	})
}
