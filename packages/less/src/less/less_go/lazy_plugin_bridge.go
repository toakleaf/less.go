package less_go

import (
	"fmt"
	"os"
	"sync"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// LazyNodeJSPluginBridge provides lazy initialization of the NodeJS plugin bridge,
// avoiding the overhead of spawning Node.js for compilations that don't use plugins.
type LazyNodeJSPluginBridge struct {
	bridge        *NodeJSPluginBridge
	mu            sync.RWMutex
	initOnce      sync.Once
	initErr       error
	closed        bool
	pendingScopes int
}

func NewLazyNodeJSPluginBridge() *LazyNodeJSPluginBridge {
	return &LazyNodeJSPluginBridge{}
}

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

		// Apply pending scopes entered before initialization
		if lb.pendingScopes > 0 {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[LazyNodeJSPluginBridge] Applying %d pending scopes\n", lb.pendingScopes)
			}
			for i := 0; i < lb.pendingScopes; i++ {
				bridge.EnterScope()
			}
		}

		// SHM protocol disabled by default (benchmarks show ~75% slower than JSON)
		if os.Getenv("LESS_SHM_PROTOCOL") == "1" {
			if err := bridge.InitSHMProtocol(); err != nil {
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[LazyNodeJSPluginBridge] SHM protocol unavailable, using JSON: %v\n", err)
				}
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

func (lb *LazyNodeJSPluginBridge) IsInitialized() bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.bridge != nil
}

func (lb *LazyNodeJSPluginBridge) GetBridge() (*NodeJSPluginBridge, error) {
	if err := lb.ensureInitialized(); err != nil {
		return nil, err
	}
	return lb.bridge, nil
}

func (lb *LazyNodeJSPluginBridge) EvalPlugin(contents string, newEnv *Parse, importManager any, pluginArgs map[string]any, newFileInfo any) any {
	if err := lb.ensureInitialized(); err != nil {
		return err
	}
	return lb.bridge.EvalPlugin(contents, newEnv, importManager, pluginArgs, newFileInfo)
}

func (lb *LazyNodeJSPluginBridge) LoadPluginSync(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	if err := lb.ensureInitialized(); err != nil {
		return err
	}
	return lb.bridge.LoadPluginSync(path, currentDirectory, context, environment, fileManager)
}

func (lb *LazyNodeJSPluginBridge) LoadPlugin(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	if err := lb.ensureInitialized(); err != nil {
		return err
	}
	return lb.bridge.LoadPlugin(path, currentDirectory, context, environment, fileManager)
}

func (lb *LazyNodeJSPluginBridge) LookupFunction(name string) (*runtime.JSFunctionDefinition, bool) {
	if !lb.IsInitialized() {
		return nil, false
	}
	return lb.bridge.LookupFunction(name)
}

func (lb *LazyNodeJSPluginBridge) HasFunction(name string) bool {
	if !lb.IsInitialized() {
		return false
	}
	return lb.bridge.HasFunction(name)
}

func (lb *LazyNodeJSPluginBridge) CallFunction(name string, args ...any) (any, error) {
	if !lb.IsInitialized() {
		return nil, fmt.Errorf("no plugins have been loaded")
	}
	return lb.bridge.CallFunction(name, args...)
}

func (lb *LazyNodeJSPluginBridge) CallFunctionWithContext(name string, evalContext runtime.EvalContextProvider, args ...any) (any, error) {
	if !lb.IsInitialized() {
		return nil, fmt.Errorf("no plugins have been loaded")
	}
	return lb.bridge.CallFunctionWithContext(name, evalContext, args...)
}

func (lb *LazyNodeJSPluginBridge) EnterScope() *runtime.PluginScope {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.bridge == nil {
		lb.pendingScopes++
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[LazyNodeJSPluginBridge] EnterScope (pending): pendingScopes=%d\n", lb.pendingScopes)
		}
		return nil
	}
	return lb.bridge.EnterScope()
}

func (lb *LazyNodeJSPluginBridge) ExitScope() *runtime.PluginScope {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if lb.bridge == nil {
		if lb.pendingScopes > 0 {
			lb.pendingScopes--
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[LazyNodeJSPluginBridge] ExitScope (pending): pendingScopes=%d\n", lb.pendingScopes)
			}
		}
		return nil
	}
	return lb.bridge.ExitScope()
}

func (lb *LazyNodeJSPluginBridge) GetScope() *runtime.PluginScope {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetScope()
}

func (lb *LazyNodeJSPluginBridge) GetRuntime() *runtime.NodeJSRuntime {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetRuntime()
}

func (lb *LazyNodeJSPluginBridge) GetVisitors() []*runtime.JSVisitor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetVisitors()
}

func (lb *LazyNodeJSPluginBridge) GetPreEvalVisitors() []*runtime.JSVisitor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPreEvalVisitors()
}

func (lb *LazyNodeJSPluginBridge) GetPostEvalVisitors() []*runtime.JSVisitor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPostEvalVisitors()
}

func (lb *LazyNodeJSPluginBridge) GetProcessorManager() *runtime.ProcessorManager {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetProcessorManager()
}

func (lb *LazyNodeJSPluginBridge) GetPreProcessors() []*runtime.JSPreProcessor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPreProcessors()
}

func (lb *LazyNodeJSPluginBridge) GetPostProcessors() []*runtime.JSPostProcessor {
	if !lb.IsInitialized() {
		return nil
	}
	return lb.bridge.GetPostProcessors()
}

func (lb *LazyNodeJSPluginBridge) RunPreProcessors(input string, options map[string]any) (string, error) {
	if !lb.IsInitialized() {
		return input, nil
	}
	return lb.bridge.RunPreProcessors(input, options)
}

func (lb *LazyNodeJSPluginBridge) RunPostProcessors(css string, options map[string]any) (string, error) {
	if !lb.IsInitialized() {
		return css, nil
	}
	return lb.bridge.RunPostProcessors(css, options)
}

func (lb *LazyNodeJSPluginBridge) Close() error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.closed = true

	if lb.bridge != nil {
		return lb.bridge.Close()
	}
	return nil
}

func (lb *LazyNodeJSPluginBridge) WasUsed() bool {
	return lb.IsInitialized()
}

func LazyPluginLoaderFactory(bridge *LazyNodeJSPluginBridge) PluginLoaderFactory {
	return func(less LessInterface) PluginLoader {
		return bridge
	}
}
