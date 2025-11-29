package less_go

import (
	"fmt"
	"os"
	"sync"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// LazyNodeJSPluginBridge provides lazy initialization of the NodeJS plugin bridge.
// It only starts the Node.js runtime when a plugin is actually loaded.
// This avoids the overhead of spawning a Node.js process for compilations
// that don't use plugins.
type LazyNodeJSPluginBridge struct {
	bridge     *NodeJSPluginBridge
	mu         sync.RWMutex
	initOnce   sync.Once
	initErr    error
	closed     bool
}

// NewLazyNodeJSPluginBridge creates a new lazy bridge that will initialize
// the Node.js runtime on first use.
func NewLazyNodeJSPluginBridge() *LazyNodeJSPluginBridge {
	return &LazyNodeJSPluginBridge{}
}

// ensureInitialized lazily initializes the Node.js runtime.
// This is thread-safe and will only initialize once.
//
// If the LESS_SHM_PROTOCOL=1 environment variable is set, the high-performance
// shared memory protocol will be enabled for faster plugin function calls.
func (lb *LazyNodeJSPluginBridge) ensureInitialized() error {
	lb.initOnce.Do(func() {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[LazyNodeJSPluginBridge] ensureInitialized called, initializing bridge...\n")
		}
		lb.mu.Lock()
		defer lb.mu.Unlock()

		if lb.closed {
			lb.initErr = fmt.Errorf("plugin bridge has been closed")
			return
		}

		bridge, err := NewNodeJSPluginBridge()
		if err != nil {
			lb.initErr = err
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[LazyNodeJSPluginBridge] Failed to initialize: %v\n", err)
			}
			return
		}
		lb.bridge = bridge

		// Optionally enable the high-performance SHM protocol
		if os.Getenv("LESS_SHM_PROTOCOL") == "1" {
			if err := bridge.InitSHMProtocol(); err != nil {
				// Log but don't fail - fallback to JSON protocol
				fmt.Fprintf(os.Stderr, "[LazyNodeJSPluginBridge] Warning: failed to init SHM protocol: %v\n", err)
			} else if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[LazyNodeJSPluginBridge] SHM protocol enabled\n")
			}
		}

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[LazyNodeJSPluginBridge] Successfully initialized bridge=%p\n", bridge)
		}
	})

	return lb.initErr
}

// IsInitialized returns true if the bridge has been initialized.
func (lb *LazyNodeJSPluginBridge) IsInitialized() bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.bridge != nil
}

// GetBridge returns the underlying NodeJSPluginBridge, initializing it if necessary.
// Returns an error if initialization fails.
func (lb *LazyNodeJSPluginBridge) GetBridge() (*NodeJSPluginBridge, error) {
	if err := lb.ensureInitialized(); err != nil {
		return nil, err
	}
	return lb.bridge, nil
}

// EvalPlugin evaluates plugin contents directly.
// Lazily initializes the Node.js runtime if needed.
func (lb *LazyNodeJSPluginBridge) EvalPlugin(contents string, newEnv *Parse, importManager any, pluginArgs map[string]any, newFileInfo any) any {
	if err := lb.ensureInitialized(); err != nil {
		return err
	}
	return lb.bridge.EvalPlugin(contents, newEnv, importManager, pluginArgs, newFileInfo)
}

// LoadPluginSync synchronously loads a plugin.
// Lazily initializes the Node.js runtime if needed.
func (lb *LazyNodeJSPluginBridge) LoadPluginSync(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	if err := lb.ensureInitialized(); err != nil {
		return err
	}
	return lb.bridge.LoadPluginSync(path, currentDirectory, context, environment, fileManager)
}

// LoadPlugin loads a plugin.
// Lazily initializes the Node.js runtime if needed.
func (lb *LazyNodeJSPluginBridge) LoadPlugin(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	if err := lb.ensureInitialized(); err != nil {
		return err
	}
	return lb.bridge.LoadPlugin(path, currentDirectory, context, environment, fileManager)
}

// LookupFunction looks up a function by name in the plugin scope.
// Returns nil, false if the bridge hasn't been initialized (no plugins loaded).
func (lb *LazyNodeJSPluginBridge) LookupFunction(name string) (*runtime.JSFunctionDefinition, bool) {
	if !lb.IsInitialized() {
		return nil, false
	}
	return lb.bridge.LookupFunction(name)
}

// HasFunction checks if a function exists in the plugin scope.
// Returns false if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) HasFunction(name string) bool {
	if !lb.IsInitialized() {
		return false
	}
	return lb.bridge.HasFunction(name)
}

// CallFunction calls a JavaScript function by name.
// Returns an error if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) CallFunction(name string, args ...any) (any, error) {
	if !lb.IsInitialized() {
		return nil, fmt.Errorf("no plugins have been loaded")
	}
	return lb.bridge.CallFunction(name, args...)
}

// CallFunctionWithContext calls a JavaScript function by name with evaluation context.
// This is used by plugin functions that need to access Less variables.
// Returns an error if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) CallFunctionWithContext(name string, evalContext runtime.EvalContextProvider, args ...any) (any, error) {
	if !lb.IsInitialized() {
		return nil, fmt.Errorf("no plugins have been loaded")
	}
	return lb.bridge.CallFunctionWithContext(name, evalContext, args...)
}

// EnterScope creates and enters a new child scope.
// Only effective if the bridge is initialized.
func (lb *LazyNodeJSPluginBridge) EnterScope() *runtime.PluginScope {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.EnterScope()
}

// ExitScope exits the current scope and returns to the parent.
// Only effective if the bridge is initialized.
func (lb *LazyNodeJSPluginBridge) ExitScope() *runtime.PluginScope {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.ExitScope()
}

// GetScope returns the current plugin scope.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetScope() *runtime.PluginScope {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetScope()
}

// GetRuntime returns the underlying Node.js runtime.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetRuntime() *runtime.NodeJSRuntime {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetRuntime()
}

// GetVisitors returns all visitors from the current scope.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetVisitors() []*runtime.JSVisitor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetVisitors()
}

// GetPreEvalVisitors returns pre-evaluation visitors.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetPreEvalVisitors() []*runtime.JSVisitor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPreEvalVisitors()
}

// GetPostEvalVisitors returns post-evaluation visitors.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetPostEvalVisitors() []*runtime.JSVisitor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPostEvalVisitors()
}

// GetProcessorManager returns the processor manager.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetProcessorManager() *runtime.ProcessorManager {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetProcessorManager()
}

// GetPreProcessors returns all registered pre-processors.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetPreProcessors() []*runtime.JSPreProcessor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPreProcessors()
}

// GetPostProcessors returns all registered post-processors.
// Returns nil if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) GetPostProcessors() []*runtime.JSPostProcessor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPostProcessors()
}

// RunPreProcessors runs all pre-processors on the input source.
// Returns the input unchanged if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) RunPreProcessors(input string, options map[string]any) (string, error) {
	if !lb.IsInitialized() {
		return input, nil
	}
	return lb.bridge.RunPreProcessors(input, options)
}

// RunPostProcessors runs all post-processors on the CSS output.
// Returns the CSS unchanged if the bridge hasn't been initialized.
func (lb *LazyNodeJSPluginBridge) RunPostProcessors(css string, options map[string]any) (string, error) {
	if !lb.IsInitialized() {
		return css, nil
	}
	return lb.bridge.RunPostProcessors(css, options)
}

// Close shuts down the Node.js runtime if it was initialized.
// This should be called when compilation is complete.
func (lb *LazyNodeJSPluginBridge) Close() error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.closed = true

	if lb.bridge != nil {
		return lb.bridge.Close()
	}
	return nil
}

// WasUsed returns true if the bridge was actually initialized and used.
// This can be used for diagnostics or logging.
func (lb *LazyNodeJSPluginBridge) WasUsed() bool {
	return lb.IsInitialized()
}

// LazyPluginLoaderFactory creates a PluginLoaderFactory that returns a LazyNodeJSPluginBridge.
// The bridge is shared across the compilation and should be closed when done.
func LazyPluginLoaderFactory(bridge *LazyNodeJSPluginBridge) PluginLoaderFactory {
	return func(less LessInterface) PluginLoader {
		return bridge
	}
}
