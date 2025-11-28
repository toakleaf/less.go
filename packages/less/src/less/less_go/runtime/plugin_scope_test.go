package runtime

import (
	"testing"
)

func TestPluginScope_Hierarchy(t *testing.T) {
	t.Run("should create root scope with no parent", func(t *testing.T) {
		root := NewRootPluginScope()
		if root.Parent() != nil {
			t.Error("root scope should have no parent")
		}
		if !root.IsRoot() {
			t.Error("root scope should return true for IsRoot()")
		}
	})

	t.Run("should create child scope with parent", func(t *testing.T) {
		root := NewRootPluginScope()
		child := root.CreateChild()

		if child.Parent() != root {
			t.Error("child scope should have root as parent")
		}
		if child.IsRoot() {
			t.Error("child scope should not be root")
		}
	})

	t.Run("should create multi-level hierarchy", func(t *testing.T) {
		root := NewRootPluginScope()
		child1 := root.CreateChild()
		child2 := child1.CreateChild()

		if child2.Parent() != child1 {
			t.Error("child2 should have child1 as parent")
		}
		if child2.Parent().Parent() != root {
			t.Error("child2's grandparent should be root")
		}
	})
}

func TestPluginScope_FunctionLookup(t *testing.T) {
	t.Run("should find function in current scope", func(t *testing.T) {
		scope := NewRootPluginScope()
		fn := &JSFunctionDefinition{name: "test-func"}
		scope.AddFunction("test-func", fn)

		found, ok := scope.LookupFunction("test-func")
		if !ok {
			t.Error("function should be found")
		}
		if found != fn {
			t.Error("should return the correct function")
		}
	})

	t.Run("should find function in parent scope", func(t *testing.T) {
		root := NewRootPluginScope()
		fn := &JSFunctionDefinition{name: "parent-func"}
		root.AddFunction("parent-func", fn)

		child := root.CreateChild()

		found, ok := child.LookupFunction("parent-func")
		if !ok {
			t.Error("function from parent should be found in child")
		}
		if found != fn {
			t.Error("should return the correct function from parent")
		}
	})

	t.Run("should return false for non-existent function", func(t *testing.T) {
		scope := NewRootPluginScope()

		_, ok := scope.LookupFunction("non-existent")
		if ok {
			t.Error("should not find non-existent function")
		}
	})
}

func TestPluginScope_Shadowing(t *testing.T) {
	t.Run("should shadow parent function with local function", func(t *testing.T) {
		root := NewRootPluginScope()
		globalFn := &JSFunctionDefinition{name: "shadow-test"}
		root.AddFunction("shadow-test", globalFn)

		child := root.CreateChild()
		localFn := &JSFunctionDefinition{name: "shadow-test-local"}
		child.AddFunction("shadow-test", localFn)

		// Child should find local version
		found, ok := child.LookupFunction("shadow-test")
		if !ok {
			t.Error("function should be found in child")
		}
		if found != localFn {
			t.Error("child should return local function, not parent's")
		}

		// Parent should still have its own version
		foundInParent, ok := root.LookupFunction("shadow-test")
		if !ok {
			t.Error("function should be found in parent")
		}
		if foundInParent != globalFn {
			t.Error("parent should return its own function")
		}
	})

	t.Run("should shadow at multiple levels", func(t *testing.T) {
		root := NewRootPluginScope()
		rootFn := &JSFunctionDefinition{name: "root"}
		root.AddFunction("foo", rootFn)

		child1 := root.CreateChild()
		child1Fn := &JSFunctionDefinition{name: "child1"}
		child1.AddFunction("foo", child1Fn)

		child2 := child1.CreateChild()
		child2Fn := &JSFunctionDefinition{name: "child2"}
		child2.AddFunction("foo", child2Fn)

		// Each level should find its own version
		found, _ := child2.LookupFunction("foo")
		if found != child2Fn {
			t.Error("child2 should find its own version")
		}

		found, _ = child1.LookupFunction("foo")
		if found != child1Fn {
			t.Error("child1 should find its own version")
		}

		found, _ = root.LookupFunction("foo")
		if found != rootFn {
			t.Error("root should find its own version")
		}
	})
}

func TestPluginScope_LocalVsGlobal(t *testing.T) {
	t.Run("should distinguish local vs inherited functions", func(t *testing.T) {
		root := NewRootPluginScope()
		globalFn := &JSFunctionDefinition{name: "global"}
		root.AddFunction("global-func", globalFn)

		child := root.CreateChild()
		localFn := &JSFunctionDefinition{name: "local"}
		child.AddFunction("local-func", localFn)

		// Local lookup should only find local
		found, ok := child.GetLocalFunction("local-func")
		if !ok || found != localFn {
			t.Error("should find local function with GetLocalFunction")
		}

		_, ok = child.GetLocalFunction("global-func")
		if ok {
			t.Error("should NOT find global function with GetLocalFunction")
		}

		// Regular lookup should find both
		_, ok = child.LookupFunction("local-func")
		if !ok {
			t.Error("should find local function with LookupFunction")
		}

		_, ok = child.LookupFunction("global-func")
		if !ok {
			t.Error("should find global function with LookupFunction")
		}
	})

	t.Run("should get all functions including inherited", func(t *testing.T) {
		root := NewRootPluginScope()
		root.AddFunction("global1", &JSFunctionDefinition{name: "global1"})
		root.AddFunction("global2", &JSFunctionDefinition{name: "global2"})

		child := root.CreateChild()
		child.AddFunction("local1", &JSFunctionDefinition{name: "local1"})
		child.AddFunction("global1", &JSFunctionDefinition{name: "shadowed"}) // Shadow global1

		allFuncs := child.GetAllFunctions()

		if len(allFuncs) != 3 {
			t.Errorf("should have 3 functions (2 global, 1 local, 1 shadowed), got %d", len(allFuncs))
		}

		if allFuncs["local1"].name != "local1" {
			t.Error("should have local1")
		}
		if allFuncs["global2"].name != "global2" {
			t.Error("should have global2 from parent")
		}
		if allFuncs["global1"].name != "shadowed" {
			t.Error("global1 should be shadowed by local version")
		}
	})
}

func TestPluginScope_Visitors(t *testing.T) {
	t.Run("should inherit visitors from parent", func(t *testing.T) {
		root := NewRootPluginScope()
		rootVisitor := &JSVisitor{Index: 1, IsPreEvalVisitor: true}
		root.AddVisitor(rootVisitor)

		child := root.CreateChild()
		childVisitor := &JSVisitor{Index: 2, IsPreEvalVisitor: false}
		child.AddVisitor(childVisitor)

		// Child should have both visitors
		visitors := child.GetVisitors()
		if len(visitors) != 2 {
			t.Errorf("child should have 2 visitors, got %d", len(visitors))
		}

		// Local should only have child's visitor
		localVisitors := child.GetLocalVisitors()
		if len(localVisitors) != 1 {
			t.Errorf("local should have 1 visitor, got %d", len(localVisitors))
		}
	})

	t.Run("should filter pre-eval and post-eval visitors", func(t *testing.T) {
		scope := NewRootPluginScope()
		preEval := &JSVisitor{Index: 1, IsPreEvalVisitor: true}
		postEval := &JSVisitor{Index: 2, IsPreEvalVisitor: false}
		scope.AddVisitor(preEval)
		scope.AddVisitor(postEval)

		preEvalVisitors := scope.GetPreEvalVisitors()
		if len(preEvalVisitors) != 1 {
			t.Errorf("should have 1 pre-eval visitor, got %d", len(preEvalVisitors))
		}

		postEvalVisitors := scope.GetPostEvalVisitors()
		if len(postEvalVisitors) != 1 {
			t.Errorf("should have 1 post-eval visitor, got %d", len(postEvalVisitors))
		}
	})
}

func TestPluginScope_Processors(t *testing.T) {
	t.Run("should sort processors by priority", func(t *testing.T) {
		scope := NewRootPluginScope()
		scope.AddPreProcessor("high", 1000)
		scope.AddPreProcessor("low", 100)
		scope.AddPreProcessor("medium", 500)

		processors := scope.GetPreProcessors()
		if len(processors) != 3 {
			t.Errorf("should have 3 processors, got %d", len(processors))
		}

		// Should be sorted by priority (low to high)
		if processors[0] != "low" {
			t.Errorf("first should be 'low', got %v", processors[0])
		}
		if processors[1] != "medium" {
			t.Errorf("second should be 'medium', got %v", processors[1])
		}
		if processors[2] != "high" {
			t.Errorf("third should be 'high', got %v", processors[2])
		}
	})

	t.Run("should inherit processors from parent", func(t *testing.T) {
		root := NewRootPluginScope()
		root.AddPreProcessor("parent", 500)

		child := root.CreateChild()
		child.AddPreProcessor("child", 500)

		processors := child.GetPreProcessors()
		if len(processors) != 2 {
			t.Errorf("should have 2 processors, got %d", len(processors))
		}
	})
}

func TestPluginScope_FileManagers(t *testing.T) {
	t.Run("should inherit file managers from parent", func(t *testing.T) {
		root := NewRootPluginScope()
		root.AddFileManager("parent-fm")

		child := root.CreateChild()
		child.AddFileManager("child-fm")

		managers := child.GetFileManagers()
		if len(managers) != 2 {
			t.Errorf("should have 2 file managers, got %d", len(managers))
		}
	})
}

func TestPluginScope_AddPlugin(t *testing.T) {
	t.Run("should register plugin functions", func(t *testing.T) {
		scope := NewRootPluginScope()
		plugin := &Plugin{
			Functions: []string{"test-global", "test-local"},
		}

		// Create a mock runtime (nil is ok for this test since we're not calling functions)
		scope.AddPlugin(plugin, nil)

		plugins := scope.GetPlugins()
		if len(plugins) != 1 {
			t.Errorf("should have 1 plugin, got %d", len(plugins))
		}

		// Functions should be registered (but with nil runtime in this case)
		_, ok := scope.LookupFunction("test-global")
		if !ok {
			t.Error("test-global function should be registered")
		}

		_, ok = scope.LookupFunction("test-local")
		if !ok {
			t.Error("test-local function should be registered")
		}
	})
}

func TestScopedPluginManager(t *testing.T) {
	t.Run("should provide visitor iterator", func(t *testing.T) {
		scope := NewRootPluginScope()
		scope.AddVisitor(&JSVisitor{Index: 1})
		scope.AddVisitor(&JSVisitor{Index: 2})

		spm := NewScopedPluginManager(scope, nil)
		iter := spm.Visitor()

		iter.First()
		v1 := iter.Get()
		if v1 == nil {
			t.Error("first visitor should not be nil")
		}

		v2 := iter.Get()
		if v2 == nil {
			t.Error("second visitor should not be nil")
		}

		v3 := iter.Get()
		if v3 != nil {
			t.Error("third call should return nil")
		}
	})

	t.Run("should return visitors as any slice", func(t *testing.T) {
		scope := NewRootPluginScope()
		scope.AddVisitor(&JSVisitor{Index: 1})

		spm := NewScopedPluginManager(scope, nil)
		visitors := spm.GetVisitors()

		if len(visitors) != 1 {
			t.Errorf("should have 1 visitor, got %d", len(visitors))
		}
	})
}
