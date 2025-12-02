package less_go

import (
	"fmt"
	"os"
	"sync"

	"github.com/toakleaf/less.go/less/runtime"
)

// NodeJSPluginBridge bridges the runtime.JSPluginLoader with the less_go.PluginLoader interface.
// It enables the parsing and evaluation pipeline to use JavaScript plugins loaded via Node.js.
type NodeJSPluginBridge struct {
	runtime          *runtime.NodeJSRuntime
	loader           *runtime.JSPluginLoader
	scope            *runtime.PluginScope
	funcRegistry     *runtime.PluginFunctionRegistry
	visitorManager   *runtime.VisitorManager
	processorManager *runtime.ProcessorManager
	mu               sync.RWMutex
	scopeDepth       int  // Track current scope depth for local plugin detection
	baseScopeDepth   int  // Scope depth when first plugin was loaded (treated as "root")
	hasPlugins       bool // True once first plugin is loaded
	needsScopeSync   bool // True if any plugin was loaded deeper than baseScopeDepth
}

// NewNodeJSPluginBridge creates a new bridge with a fresh Node.js runtime.
// This should be called once per compilation to spawn a new Node.js process.
func NewNodeJSPluginBridge() (*NodeJSPluginBridge, error) {
	rt, err := runtime.NewNodeJSRuntime()
	if err != nil {
		return nil, fmt.Errorf("failed to create Node.js runtime: %w", err)
	}

	// Start the runtime (spawns the Node.js process)
	if err := rt.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Node.js runtime: %w", err)
	}

	return &NodeJSPluginBridge{
		runtime:          rt,
		loader:           runtime.NewJSPluginLoader(rt),
		scope:            runtime.NewRootPluginScope(),
		funcRegistry:     runtime.NewPluginFunctionRegistry(nil, rt),
		visitorManager:   runtime.NewVisitorManager(rt),
		processorManager: runtime.NewProcessorManager(rt),
	}, nil
}

// NewNodeJSPluginBridgeWithRuntime creates a bridge using an existing runtime.
// This allows sharing the Node.js process across multiple compilations.
func NewNodeJSPluginBridgeWithRuntime(rt *runtime.NodeJSRuntime) *NodeJSPluginBridge {
	return &NodeJSPluginBridge{
		runtime:          rt,
		loader:           runtime.NewJSPluginLoader(rt),
		scope:            runtime.NewRootPluginScope(),
		funcRegistry:     runtime.NewPluginFunctionRegistry(nil, rt),
		visitorManager:   runtime.NewVisitorManager(rt),
		processorManager: runtime.NewProcessorManager(rt),
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
		// Track the "base" scope depth when first plugin is loaded.
		// Plugins loaded at this depth are treated as global plugins.
		// Plugins loaded at depth > baseScopeDepth are local and enable scope syncing.
		if !b.hasPlugins {
			b.baseScopeDepth = b.scopeDepth
			b.hasPlugins = true
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[NodeJSPluginBridge] First plugin loaded at scopeDepth=%d, baseScopeDepth=%d\n", b.scopeDepth, b.baseScopeDepth)
			}
		} else if b.scopeDepth > b.baseScopeDepth && !b.needsScopeSync {
			// Local plugin - enable scope syncing for shadowing to work
			b.needsScopeSync = true
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[NodeJSPluginBridge] Local plugin loaded at scopeDepth=%d (base=%d), enabling scope sync\n", b.scopeDepth, b.baseScopeDepth)
			}
			// CRITICAL: Sync up Node.js to the current scope depth.
			// We missed some scope changes before needsScopeSync was enabled,
			// so we need to catch up Node.js to match our current Go scope depth.
			if b.runtime != nil {
				scopesToSync := b.scopeDepth - b.baseScopeDepth
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[NodeJSPluginBridge] Syncing %d missed scopes to Node.js\n", scopesToSync)
				}
				for i := 0; i < scopesToSync; i++ {
					b.runtime.IncrementScopeDepth()
					b.runtime.IncrementScopeSeq()
					err := b.runtime.SendCommandFireAndForget(runtime.Command{
						Cmd:  "enterScope",
						Data: map[string]any{},
					})
					if err != nil && os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Fprintf(os.Stderr, "[NodeJSPluginBridge] Catch-up enterScope IPC error: %v\n", err)
					}
				}
			}
		}

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

		// Refresh processors from Node.js if the plugin has any
		if plugin.PreProcessors > 0 || plugin.PostProcessors > 0 {
			if err := b.processorManager.RefreshProcessors(); err != nil {
				fmt.Printf("[NodeJSPluginBridge] Warning: failed to refresh processors: %v\n", err)
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

// HasFunction checks if a function exists in the current scope hierarchy.
// This respects plugin scoping - local plugins are only visible in their scope.
func (b *NodeJSPluginBridge) HasFunction(name string) bool {
	// Only check scope hierarchy - this respects local vs global plugin visibility
	// The scope hierarchy is: current scope -> parent scope -> ... -> root scope
	// Functions registered in child scopes are NOT visible from parent scopes
	_, ok := b.scope.LookupFunction(name)
	return ok
}

// CallFunction calls a JavaScript function by name.
// Only functions visible in the current scope hierarchy can be called.
func (b *NodeJSPluginBridge) CallFunction(name string, args ...any) (any, error) {
	// Only look up function in scope hierarchy - respects plugin scoping
	if fn, ok := b.scope.LookupFunction(name); ok {
		return fn.Call(args...)
	}

	return nil, fmt.Errorf("function '%s' not found in current scope", name)
}

// CallFunctionWithContext calls a JavaScript function by name with evaluation context.
// This is used by plugin functions that need to access Less variables.
// The context provides frames and importantScope for variable lookup.
func (b *NodeJSPluginBridge) CallFunctionWithContext(name string, evalContext runtime.EvalContextProvider, args ...any) (any, error) {
	// Only look up function in scope hierarchy - respects plugin scoping
	if fn, ok := b.scope.LookupFunction(name); ok {
		return fn.CallWithContext(evalContext, args...)
	}

	return nil, fmt.Errorf("function '%s' not found in current scope", name)
}

// EnterScope creates and enters a new child scope.
// This is used when entering a ruleset or mixin that might have local plugins.
//
// OPTIMIZATION: Only syncs with Node.js when needsScopeSync is true (i.e., when
// a local plugin was loaded that needs scoping). For global plugins only, we skip
// IPC entirely for massive performance improvement (~8x faster).
func (b *NodeJSPluginBridge) EnterScope() *runtime.PluginScope {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Create child scope in Go
	b.scope = b.scope.CreateChild()
	b.scopeDepth++

	// Only sync with Node.js if needed (local plugins require scoping)
	if b.needsScopeSync && b.runtime != nil {
		// Track scope depth and sequence in runtime for cache key generation
		b.runtime.IncrementScopeDepth()
		b.runtime.IncrementScopeSeq() // Ensures sibling scopes don't share cache

		// Use fire-and-forget for performance - scope sync doesn't need response
		err := b.runtime.SendCommandFireAndForget(runtime.Command{
			Cmd:  "enterScope",
			Data: map[string]any{},
		})
		if err != nil && os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[NodeJSPluginBridge] EnterScope IPC error: %v\n", err)
		}
	}

	return b.scope
}

// ExitScope exits the current scope and returns to the parent.
// Returns the parent scope, or nil if already at root.
//
// OPTIMIZATION: Only syncs with Node.js when needsScopeSync is true.
func (b *NodeJSPluginBridge) ExitScope() *runtime.PluginScope {
	b.mu.Lock()
	defer b.mu.Unlock()

	if parent := b.scope.Parent(); parent != nil {
		oldScope := b.scope
		b.scope = parent
		b.scopeDepth--
		// Release old scope to pool for reuse
		oldScope.Release()

		// Only sync with Node.js if needed (local plugins require scoping)
		if b.needsScopeSync && b.runtime != nil {
			// Track scope depth in runtime for cache key generation
			b.runtime.DecrementScopeDepth()

			// Use fire-and-forget for performance - scope sync doesn't need response
			err := b.runtime.SendCommandFireAndForget(runtime.Command{
				Cmd:  "exitScope",
				Data: map[string]any{},
			})
			if err != nil && os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[NodeJSPluginBridge] ExitScope IPC error: %v\n", err)
			}
		}
	}

	return b.scope
}

// AddFunctionToCurrentScope registers a function name at the current scope depth.
// This is used when re-registering plugin functions inherited from ancestor frames
// (e.g., when a mixin defined inside a namespace with @plugin is called).
// The function must already exist in the Node.js runtime - this just makes it
// visible at the current scope level in BOTH Go and Node.js.
//
// OPTIMIZATION: Only syncs with Node.js when needsScopeSync is true.
func (b *NodeJSPluginBridge) AddFunctionToCurrentScope(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add to Go scope
	fn := runtime.GetOrCreateJSFunctionDefinition(name, b.runtime)
	b.scope.AddFunction(name, fn)

	// Only sync with Node.js if needed (local plugins require scoping)
	if b.needsScopeSync && b.runtime != nil {
		err := b.runtime.SendCommandFireAndForget(runtime.Command{
			Cmd: "addFunctionToScopeNoReply",
			Data: map[string]any{
				"name": name,
			},
		})
		if err != nil && os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[NodeJSPluginBridge] AddFunctionToCurrentScope IPC error: %v\n", err)
		}
	}
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
		// Close SHM protocol first if initialized
		b.runtime.CloseSHMProtocol()
		return b.runtime.Stop()
	}
	return nil
}

// InitSHMProtocol initializes the high-performance shared memory protocol.
// This should be called once at the start of compilation for best performance.
// After initialization, plugin function calls will use binary IPC instead of JSON.
func (b *NodeJSPluginBridge) InitSHMProtocol() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.runtime == nil {
		return fmt.Errorf("runtime not initialized")
	}
	return b.runtime.InitSHMProtocol()
}

// UseSHMProtocol returns whether the binary SHM protocol is enabled.
func (b *NodeJSPluginBridge) UseSHMProtocol() bool {
	if b.runtime == nil {
		return false
	}
	return b.runtime.UseSHMProtocol()
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

// HasPreEvalVisitors returns true if there are any pre-eval visitors registered.
func (b *NodeJSPluginBridge) HasPreEvalVisitors() bool {
	visitors := b.scope.GetPreEvalVisitors()
	return len(visitors) > 0
}

// VariableInfo represents a variable to check for replacement.
type VariableInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// VariableReplacement represents a replacement for a variable.
type VariableReplacement struct {
	Type  string         `json:"_type"`
	Value string         `json:"value,omitempty"`
	Quote string         `json:"quote,omitempty"`
	RGB   []float64      `json:"rgb,omitempty"`
	Alpha float64        `json:"alpha,omitempty"`
	Unit  string         `json:"unit,omitempty"`
	Props map[string]any `json:"-"`
}

// CheckVariableReplacements checks which variables should be replaced by pre-eval visitors.
// Returns a map of variable ID to replacement info.
func (b *NodeJSPluginBridge) CheckVariableReplacements(variables []VariableInfo) (map[string]map[string]any, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	// Convert to interface slice for JSON
	varsData := make([]any, len(variables))
	for i, v := range variables {
		varsData[i] = map[string]any{
			"id":   v.ID,
			"name": v.Name,
		}
	}

	resp, err := b.runtime.SendCommand(runtime.Command{
		Cmd: "checkVariableReplacements",
		Data: map[string]any{
			"variables": varsData,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check variable replacements: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("check variable replacements error: %s", resp.Error)
	}

	// Parse the result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return nil, nil
	}

	replacements, ok := resultMap["replacements"].(map[string]any)
	if !ok {
		return nil, nil
	}

	// Convert to the expected format
	result := make(map[string]map[string]any)
	for id, repl := range replacements {
		if replMap, ok := repl.(map[string]any); ok {
			result[id] = replMap
		}
	}

	return result, nil
}

// RunPreEvalVisitorsJSON runs pre-eval visitors on a JSON-serialized AST.
// Returns the modified AST as a map.
func (b *NodeJSPluginBridge) RunPreEvalVisitorsJSON(ast map[string]any) (map[string]any, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.runtime == nil {
		return ast, false, fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := b.runtime.SendCommand(runtime.Command{
		Cmd: "runPreEvalVisitorsJSON",
		Data: map[string]any{
			"ast": ast,
		},
	})
	if err != nil {
		return ast, false, fmt.Errorf("failed to run pre-eval visitors: %w", err)
	}

	if !resp.Success {
		return ast, false, fmt.Errorf("pre-eval visitors error: %s", resp.Error)
	}

	// Parse the result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return ast, false, nil
	}

	modified, _ := resultMap["modified"].(bool)
	modifiedAst, ok := resultMap["modifiedAst"].(map[string]any)
	if !ok {
		return ast, false, nil
	}

	return modifiedAst, modified, nil
}

// GetProcessorManager returns the processor manager for pre/post processing.
func (b *NodeJSPluginBridge) GetProcessorManager() *runtime.ProcessorManager {
	return b.processorManager
}

// GetPreProcessors returns all registered pre-processors.
func (b *NodeJSPluginBridge) GetPreProcessors() []*runtime.JSPreProcessor {
	return b.processorManager.GetPreProcessors()
}

// GetPostProcessors returns all registered post-processors.
func (b *NodeJSPluginBridge) GetPostProcessors() []*runtime.JSPostProcessor {
	return b.processorManager.GetPostProcessors()
}

// RunPreProcessors runs all pre-processors on the input source.
func (b *NodeJSPluginBridge) RunPreProcessors(input string, options map[string]any) (string, error) {
	return b.processorManager.RunPreProcessors(input, options)
}

// RunPostProcessors runs all post-processors on the CSS output.
func (b *NodeJSPluginBridge) RunPostProcessors(css string, options map[string]any) (string, error) {
	return b.processorManager.RunPostProcessors(css, options)
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
