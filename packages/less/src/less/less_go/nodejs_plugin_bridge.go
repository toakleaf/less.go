package less_go

import (
	"fmt"
	"sync"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// NodeJSPluginBridge bridges the runtime.JSPluginLoader with the less_go.PluginLoader interface.
// It enables the parsing and evaluation pipeline to use JavaScript plugins loaded via Node.js.
type NodeJSPluginBridge struct {
	runtime        *runtime.NodeJSRuntime
	loader         *runtime.JSPluginLoader
	scope          *runtime.PluginScope
	funcRegistry   *runtime.PluginFunctionRegistry
	visitorManager *runtime.VisitorManager
	mu             sync.RWMutex
}

// NewNodeJSPluginBridge creates a new bridge with a fresh Node.js runtime.
// This should be called once per compilation to spawn a new Node.js process.
func NewNodeJSPluginBridge() (*NodeJSPluginBridge, error) {
	rt, err := runtime.NewNodeJSRuntime()
	if err != nil {
		return nil, fmt.Errorf("failed to create Node.js runtime: %w", err)
	}

	return &NodeJSPluginBridge{
		runtime:        rt,
		loader:         runtime.NewJSPluginLoader(rt),
		scope:          runtime.NewRootPluginScope(),
		funcRegistry:   runtime.NewPluginFunctionRegistry(nil, rt),
		visitorManager: runtime.NewVisitorManager(rt),
	}, nil
}

// NewNodeJSPluginBridgeWithRuntime creates a bridge using an existing runtime.
// This allows sharing the Node.js process across multiple compilations.
func NewNodeJSPluginBridgeWithRuntime(rt *runtime.NodeJSRuntime) *NodeJSPluginBridge {
	return &NodeJSPluginBridge{
		runtime:        rt,
		loader:         runtime.NewJSPluginLoader(rt),
		scope:          runtime.NewRootPluginScope(),
		funcRegistry:   runtime.NewPluginFunctionRegistry(nil, rt),
		visitorManager: runtime.NewVisitorManager(rt),
	}
}

// EvalPlugin evaluates plugin contents directly (inline plugin code).
// This is used when a plugin's JavaScript code is provided inline rather than as a file path.
func (b *NodeJSPluginBridge) EvalPlugin(contents string, newEnv *Parse, importManager any, pluginArgs map[string]any, newFileInfo any) any {
	b.mu.Lock()
	defer b.mu.Unlock()

	// For inline plugins, we'd need to send the code to Node.js for evaluation
	// This is less common - most plugins are loaded from files
	// For now, return an error indicating this isn't fully supported yet
	return fmt.Errorf("inline plugin evaluation via Node.js not yet implemented")
}

// LoadPluginSync synchronously loads a plugin from the specified path.
// This wraps the runtime.JSPluginLoader and integrates the results with the scope.
func (b *NodeJSPluginBridge) LoadPluginSync(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.runtime == nil {
		return fmt.Errorf("Node.js runtime not initialized")
	}

	// Call the underlying loader
	result := b.loader.LoadPluginSync(path, currentDirectory, context, environment, fileManager)

	// Handle errors
	if err, ok := result.(error); ok {
		return err
	}

	// Process the loaded plugin
	if plugin, ok := result.(*runtime.Plugin); ok {
		// Register the plugin with the current scope
		b.scope.AddPlugin(plugin, b.runtime)

		// Register functions in the function registry
		b.funcRegistry.RegisterJSFunctions(plugin.Functions)

		// Refresh visitors from Node.js
		if err := b.visitorManager.RefreshVisitors(); err != nil {
			// Log but don't fail - plugin functions are still available
			fmt.Printf("[NodeJSPluginBridge] Warning: failed to refresh visitors: %v\n", err)
		} else {
			// Add new visitors to the scope
			for _, v := range b.visitorManager.GetPreEvalVisitors() {
				b.scope.AddVisitor(v)
			}
			for _, v := range b.visitorManager.GetPostEvalVisitors() {
				b.scope.AddVisitor(v)
			}
		}

		return plugin
	}

	return result
}

// LoadPlugin loads a plugin asynchronously.
// For now, this just calls LoadPluginSync since we're in a synchronous Go context.
func (b *NodeJSPluginBridge) LoadPlugin(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	return b.LoadPluginSync(path, currentDirectory, context, environment, fileManager)
}

// GetRuntime returns the underlying Node.js runtime.
func (b *NodeJSPluginBridge) GetRuntime() *runtime.NodeJSRuntime {
	return b.runtime
}

// GetLoader returns the underlying plugin loader.
func (b *NodeJSPluginBridge) GetLoader() *runtime.JSPluginLoader {
	return b.loader
}

// GetScope returns the current plugin scope.
func (b *NodeJSPluginBridge) GetScope() *runtime.PluginScope {
	return b.scope
}

// GetFunctionRegistry returns the function registry.
func (b *NodeJSPluginBridge) GetFunctionRegistry() *runtime.PluginFunctionRegistry {
	return b.funcRegistry
}

// GetVisitorManager returns the visitor manager.
func (b *NodeJSPluginBridge) GetVisitorManager() *runtime.VisitorManager {
	return b.visitorManager
}

// LookupFunction looks up a function by name in the current scope.
// This is used by the function caller during evaluation.
func (b *NodeJSPluginBridge) LookupFunction(name string) (*runtime.JSFunctionDefinition, bool) {
	return b.scope.LookupFunction(name)
}

// HasFunction checks if a function exists in the current scope or Node.js registry.
func (b *NodeJSPluginBridge) HasFunction(name string) bool {
	// Check scope first
	if _, ok := b.scope.LookupFunction(name); ok {
		return true
	}
	// Check registry
	return b.funcRegistry.HasJSFunction(name)
}

// CallFunction calls a JavaScript function by name.
func (b *NodeJSPluginBridge) CallFunction(name string, args ...any) (any, error) {
	// Look up function in scope
	if fn, ok := b.scope.LookupFunction(name); ok {
		return fn.Call(args...)
	}

	// Fall back to function registry
	if fnDef := b.funcRegistry.Get(name); fnDef != nil {
		if jsFn, ok := fnDef.(*runtime.JSFunctionDefinition); ok {
			return jsFn.Call(args...)
		}
	}

	return nil, fmt.Errorf("function '%s' not found", name)
}

// EnterScope creates and enters a new child scope.
// This is used when entering a ruleset or mixin that might have local plugins.
func (b *NodeJSPluginBridge) EnterScope() *runtime.PluginScope {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scope = b.scope.CreateChild()
	return b.scope
}

// ExitScope exits the current scope and returns to the parent.
// Returns the parent scope, or nil if already at root.
func (b *NodeJSPluginBridge) ExitScope() *runtime.PluginScope {
	b.mu.Lock()
	defer b.mu.Unlock()
	if parent := b.scope.Parent(); parent != nil {
		b.scope = parent
	}
	return b.scope
}

// SetScope sets the current scope directly.
// This allows restoring a previous scope state.
func (b *NodeJSPluginBridge) SetScope(scope *runtime.PluginScope) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scope = scope
}

// Close shuts down the Node.js runtime.
// This should be called when the compilation is complete.
func (b *NodeJSPluginBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.runtime != nil {
		return b.runtime.Stop()
	}
	return nil
}

// GetVisitors returns all visitors from the current scope.
// This is used by transform_tree.go to get plugin visitors.
func (b *NodeJSPluginBridge) GetVisitors() []*runtime.JSVisitor {
	return b.scope.GetVisitors()
}

// GetPreEvalVisitors returns pre-evaluation visitors from the current scope.
func (b *NodeJSPluginBridge) GetPreEvalVisitors() []*runtime.JSVisitor {
	return b.scope.GetPreEvalVisitors()
}

// GetPostEvalVisitors returns post-evaluation visitors from the current scope.
func (b *NodeJSPluginBridge) GetPostEvalVisitors() []*runtime.JSVisitor {
	return b.scope.GetPostEvalVisitors()
}

// CreateScopedPluginManager creates a ScopedPluginManager for the current scope.
// This provides compatibility with the existing PluginManager interface.
func (b *NodeJSPluginBridge) CreateScopedPluginManager() *runtime.ScopedPluginManager {
	return runtime.NewScopedPluginManager(b.scope, b.runtime)
}

// NodeJSPluginLoaderFactory creates a PluginLoaderFactory that returns NodeJSPluginBridge.
// This can be passed to LessContext to enable JavaScript plugin support.
func NodeJSPluginLoaderFactory(runtime *runtime.NodeJSRuntime) PluginLoaderFactory {
	return func(less LessInterface) PluginLoader {
		return NewNodeJSPluginBridgeWithRuntime(runtime)
	}
}

// CreateLessContextWithPlugins creates a LessContext configured to use Node.js plugins.
// This is a convenience function for enabling plugin support in the parsing pipeline.
func CreateLessContextWithPlugins(options map[string]any) (*LessContext, *NodeJSPluginBridge, error) {
	bridge, err := NewNodeJSPluginBridge()
	if err != nil {
		return nil, nil, err
	}

	ctx := &LessContext{
		Options: options,
		PluginLoader: func(less LessInterface) PluginLoader {
			return bridge
		},
		Functions: &DefaultFunctions{},
	}

	return ctx, bridge, nil
}
