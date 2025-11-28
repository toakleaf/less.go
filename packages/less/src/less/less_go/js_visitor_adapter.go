package less_go

import (
	"fmt"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// JSVisitorAdapter adapts a runtime.JSVisitor to work with the PluginManager's
// visitor system. It implements the interfaces expected by transform_tree.go.
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
func (a *JSVisitorAdapter) applyReplacements(root any, replacements []runtime.VisitorReplacementSet) any {
	// For now, we'll implement a basic replacement strategy.
	// Full implementation would need to traverse the AST and find nodes by index.
	//
	// The replacement structure is:
	// - VisitorReplacementSet contains a visitorIndex and list of NodeReplacements
	// - NodeReplacement contains parentIndex, childIndex, and replacement data
	//
	// To apply replacements properly, we need to:
	// 1. Build a map of node indices to actual Go nodes
	// 2. For each replacement, find the parent and replace the child at the given index
	//
	// This is a complex operation that depends on the AST structure.
	// For now, we'll just return the root unchanged and log replacements.

	for _, set := range replacements {
		for _, repl := range set.Replacements {
			// Log for debugging
			fmt.Printf("[JSVisitorAdapter] Replacement requested: parent=%d, child=%d, type=%T\n",
				repl.ParentIndex, repl.ChildIndex, repl.Replacement)
		}
	}

	// TODO: Implement actual replacement logic once we have AST traversal support
	return root
}

// JSVisitorRegistry manages JavaScript visitors registered by plugins.
// It integrates with the PluginManager to provide visitors to transform_tree.
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
