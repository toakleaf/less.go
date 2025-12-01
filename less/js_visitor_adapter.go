package less_go

import (
	"fmt"

	"github.com/toakleaf/less.go/less/runtime"
)

// JSVisitorAdapter adapts a runtime.JSVisitor to work with the PluginManager's
// visitor system. It implements the interfaces expected by transform_tree.go.
//
// NOTE: This is part of an EXPERIMENTAL binary buffer visitor approach that is
// NOT currently used in production. The actual visitor execution uses the JSON
// pathway via NodeJSPluginBridge.RunPreEvalVisitorsJSON(). See the comment on
// applyReplacements() for more details.
type JSVisitorAdapter struct {
	visitor *runtime.JSVisitor
	runtime *runtime.NodeJSRuntime
}

// NewJSVisitorAdapter creates a new adapter for a JavaScript visitor.
func NewJSVisitorAdapter(visitor *runtime.JSVisitor, rt *runtime.NodeJSRuntime) *JSVisitorAdapter {
	return &JSVisitorAdapter{
		visitor: visitor,
		runtime: rt,
	}
}

// IsPreEvalVisitor returns true if this is a pre-evaluation visitor.
// This matches the interface expected by transform_tree.go.
func (a *JSVisitorAdapter) IsPreEvalVisitor() bool {
	return a.visitor.IsPreEvalVisitor
}

// IsPreVisitor returns false - JS visitors are not "pre" visitors in the
// sense of going before built-in visitors. They follow the pre-eval/post-eval pattern.
func (a *JSVisitorAdapter) IsPreVisitor() bool {
	return false
}

// IsReplacing returns true if this visitor can replace nodes.
func (a *JSVisitorAdapter) IsReplacing() bool {
	return a.visitor.IsReplacing
}

// Run executes the JavaScript visitor on the AST root.
// It serializes the AST, sends it to Node.js, runs the visitor, and applies
// any replacements to the Go AST.
func (a *JSVisitorAdapter) Run(root any) any {
	if a.visitor == nil || a.runtime == nil {
		return root
	}

	// Run the visitor through the runtime
	result, err := a.visitor.Visit(root)
	if err != nil {
		// Log error but don't panic - visitors failing shouldn't break compilation
		fmt.Printf("[JSVisitorAdapter] Warning: visitor error: %v\n", err)
		return root
	}

	// Apply replacements if any were returned
	if result != nil && len(result.Replacements) > 0 {
		return a.applyReplacements(root, result.Replacements)
	}

	return root
}

// applyReplacements applies the node replacements from the visitor result
// to the Go AST. This modifies the tree based on what the JS visitor changed.
//
// NOTE: This is an EXPERIMENTAL binary buffer approach that is NOT currently used.
// The production visitor system uses NodeJSPluginBridge.RunPreEvalVisitorsJSON()
// which handles visitor execution and AST modification entirely in Node.js,
// returning the complete modified AST as JSON. That approach works correctly
// and all plugin tests pass.
//
// This binary buffer approach was designed as a potential optimization to avoid
// full AST serialization, but was never completed. The JSVisitorAdapter and
// JSVisitorRegistry are only used in test files, not in production code.
//
// If performance optimization via binary buffer visitors is needed in the future,
// implement this method by:
// 1. Building a map of node indices to actual Go AST nodes during serialization
// 2. For each replacement, finding the parent node and replacing the child at the given index
// 3. Deserializing replacement data back into Go AST node types
func (a *JSVisitorAdapter) applyReplacements(root any, replacements []runtime.VisitorReplacementSet) any {
	for _, set := range replacements {
		for _, repl := range set.Replacements {
			// Log for debugging - in production, replacements would be applied here
			fmt.Printf("[JSVisitorAdapter] Replacement requested: parent=%d, child=%d, type=%T\n",
				repl.ParentIndex, repl.ChildIndex, repl.Replacement)
		}
	}

	// Binary buffer replacement logic not implemented - see NOTE above.
	// Production code uses the JSON pathway instead.
	return root
}

// JSVisitorRegistry manages JavaScript visitors registered by plugins.
// It integrates with the PluginManager to provide visitors to transform_tree.
//
// NOTE: This is part of an EXPERIMENTAL binary buffer visitor approach that is
// NOT currently used in production. See JSVisitorAdapter for details.
type JSVisitorRegistry struct {
	runtime  *runtime.NodeJSRuntime
	manager  *runtime.VisitorManager
	adapters []*JSVisitorAdapter
}

// NewJSVisitorRegistry creates a new registry for JavaScript visitors.
func NewJSVisitorRegistry(rt *runtime.NodeJSRuntime) *JSVisitorRegistry {
	return &JSVisitorRegistry{
		runtime:  rt,
		manager:  runtime.NewVisitorManager(rt),
		adapters: make([]*JSVisitorAdapter, 0),
	}
}

// RefreshFromNodeJS fetches the current list of registered visitors from Node.js
// and creates adapters for them.
func (r *JSVisitorRegistry) RefreshFromNodeJS() error {
	if err := r.manager.RefreshVisitors(); err != nil {
		return err
	}

	// Clear existing adapters
	r.adapters = make([]*JSVisitorAdapter, 0)

	// Create adapters for pre-eval visitors
	for _, v := range r.manager.GetPreEvalVisitors() {
		r.adapters = append(r.adapters, NewJSVisitorAdapter(v, r.runtime))
	}

	// Create adapters for post-eval visitors
	for _, v := range r.manager.GetPostEvalVisitors() {
		r.adapters = append(r.adapters, NewJSVisitorAdapter(v, r.runtime))
	}

	return nil
}

// GetAdapters returns all visitor adapters.
func (r *JSVisitorRegistry) GetAdapters() []*JSVisitorAdapter {
	return r.adapters
}

// GetPreEvalAdapters returns only pre-evaluation visitor adapters.
func (r *JSVisitorRegistry) GetPreEvalAdapters() []*JSVisitorAdapter {
	result := make([]*JSVisitorAdapter, 0)
	for _, a := range r.adapters {
		if a.IsPreEvalVisitor() {
			result = append(result, a)
		}
	}
	return result
}

// GetPostEvalAdapters returns only post-evaluation visitor adapters.
func (r *JSVisitorRegistry) GetPostEvalAdapters() []*JSVisitorAdapter {
	result := make([]*JSVisitorAdapter, 0)
	for _, a := range r.adapters {
		if !a.IsPreEvalVisitor() {
			result = append(result, a)
		}
	}
	return result
}

// RegisterWithPluginManager adds all visitor adapters to the PluginManager.
// This allows them to be picked up by transform_tree.go's visitor loop.
func (r *JSVisitorRegistry) RegisterWithPluginManager(pm *PluginManager) {
	for _, adapter := range r.adapters {
		pm.AddVisitor(adapter)
	}
}
