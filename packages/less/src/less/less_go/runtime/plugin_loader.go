package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Plugin represents a loaded JavaScript plugin with its registered components.
type Plugin struct {
	Path           string    // Resolved path to the plugin
	Filename       string    // Original filename/identifier
	Functions      []string  // Names of registered functions
	Visitors       int       // Number of registered visitors
	PreProcessors  int       // Number of registered pre-processors
	PostProcessors int       // Number of registered post-processors
	FileManagers   int       // Number of registered file managers
	Cached         bool      // Whether this was loaded from cache
	IPCMode        JSIPCMode // Preferred IPC mode for this plugin's functions (json or shared-memory)
}

// PluginLoadResult contains the result of loading a plugin via Node.js.
type PluginLoadResult struct {
	Success        bool     `json:"success"`
	Error          string   `json:"error,omitempty"`
	Cached         bool     `json:"cached"`
	Functions      []string `json:"functions,omitempty"`
	Visitors       int      `json:"visitors,omitempty"`
	PreProcessors  int      `json:"preProcessors,omitempty"`
	PostProcessors int      `json:"postProcessors,omitempty"`
	FileManagers   int      `json:"fileManagers,omitempty"`
	IPCMode        string   `json:"ipcMode,omitempty"` // "json" or "shm" - preferred IPC mode
}

// JSPluginLoader loads JavaScript plugins via the Node.js runtime.
// It implements the PluginLoader interface from the main less_go package.
type JSPluginLoader struct {
	runtime       *NodeJSRuntime
	loadedPlugins map[string]*Plugin
	mu            sync.RWMutex
}

// NewJSPluginLoader creates a new JavaScript plugin loader.
func NewJSPluginLoader(runtime *NodeJSRuntime) *JSPluginLoader {
	return &JSPluginLoader{
		runtime:       runtime,
		loadedPlugins: make(map[string]*Plugin),
	}
}

// LoadPlugin loads a plugin from the specified path.
// It sends a command to the Node.js runtime to load the plugin using require().
func (pl *JSPluginLoader) LoadPlugin(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	return pl.LoadPluginSync(path, currentDirectory, context, environment, fileManager)
}

// LoadPluginSync synchronously loads a plugin from the specified path.
// IMPORTANT: We always call Node.js even for cached plugins, because plugin
// functions need to be registered in the CURRENT scope. The scope changes
// during evaluation (e.g., when entering mixins), so each @plugin directive
// needs to register its functions at the current scope depth.
func (pl *JSPluginLoader) LoadPluginSync(path, currentDirectory string, context map[string]any, environment any, fileManager any) any {
	if pl.runtime == nil {
		return fmt.Errorf("Node.js runtime not initialized")
	}

	// Note: We intentionally do NOT return early for cached plugins here.
	// We always need to call Node.js to register the plugin's functions
	// in the current scope. The Node.js side handles caching and will
	// re-register the cached functions in the current scope.
	cacheKey := pl.getCacheKey(path, currentDirectory)

	// Extract options from context if present
	// Plugin options can be either:
	// - A map with "_args" key containing the raw plugin args string
	// - A full options map
	var options any
	if context != nil {
		if opts, ok := context["options"].(map[string]any); ok {
			// Check if this is the _args wrapper format
			if argsStr, ok := opts["_args"].(string); ok {
				options = argsStr // Pass the raw string to setOptions
			} else {
				options = opts // Pass the full map
			}
		} else if opts, ok := context["options"].(string); ok {
			options = opts // Direct string value
		}
	}

	// Determine if we need to try prefixes for npm modules
	// If the path doesn't start with . / or end with .js, it might be an npm module
	prefixes := []string{""}
	if !strings.HasPrefix(path, ".") && !strings.HasPrefix(path, "/") && !strings.HasSuffix(strings.ToLower(path), ".js") {
		prefixes = []string{"less-plugin-", ""}
	}

	var lastErr error
	for _, prefix := range prefixes {
		pluginPath := prefix + path

		// Send load command to Node.js
		data := map[string]any{
			"path":    pluginPath,
			"baseDir": currentDirectory,
		}
		if options != nil {
			data["options"] = options
		}

		resp, err := pl.runtime.SendCommand(Command{
			Cmd:  "loadPlugin",
			Data: data,
		})

		if err != nil {
			lastErr = err
			continue
		}

		if !resp.Success {
			lastErr = fmt.Errorf("plugin load failed: %s", resp.Error)
			continue
		}

		// Parse the result
		result, ok := resp.Result.(map[string]any)
		if !ok {
			lastErr = fmt.Errorf("unexpected response type: %T", resp.Result)
			continue
		}

		// Create plugin object from result
		plugin := &Plugin{
			Path:     pluginPath,
			Filename: path,
		}

		if cached, ok := result["cached"].(bool); ok {
			plugin.Cached = cached
		}

		if functions, ok := result["functions"].([]any); ok {
			for _, f := range functions {
				if name, ok := f.(string); ok {
					plugin.Functions = append(plugin.Functions, name)
				}
			}
		}

		if visitors, ok := result["visitors"].(float64); ok {
			plugin.Visitors = int(visitors)
		}

		if preProcessors, ok := result["preProcessors"].(float64); ok {
			plugin.PreProcessors = int(preProcessors)
		}

		if postProcessors, ok := result["postProcessors"].(float64); ok {
			plugin.PostProcessors = int(postProcessors)
		}

		if fileManagers, ok := result["fileManagers"].(float64); ok {
			plugin.FileManagers = int(fileManagers)
		}

		// Parse IPC mode preference from plugin
		// Plugins can specify their preferred IPC mode for optimal performance
		// Default is JSON mode (faster for most plugin use cases)
		plugin.IPCMode = JSIPCModeJSON // Default
		if ipcMode, ok := result["ipcMode"].(string); ok {
			plugin.IPCMode = ParseIPCMode(ipcMode)
		}

		// Cache the plugin
		pl.mu.Lock()
		pl.loadedPlugins[cacheKey] = plugin
		pl.mu.Unlock()

		return plugin
	}

	return lastErr
}

// EvalPlugin evaluates plugin contents directly (for inline plugin code).
func (pl *JSPluginLoader) EvalPlugin(contents string, newEnv any, importManager any, pluginArgs map[string]any, newFileInfo any) any {
	// For now, we don't support inline plugin evaluation
	// This would require sending the code to Node.js for evaluation
	return fmt.Errorf("inline plugin evaluation not yet supported")
}

// getCacheKey creates a cache key for a plugin path.
func (pl *JSPluginLoader) getCacheKey(path, currentDirectory string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(currentDirectory, path)
}

// GetLoadedPlugins returns a copy of all loaded plugins.
func (pl *JSPluginLoader) GetLoadedPlugins() map[string]*Plugin {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	result := make(map[string]*Plugin)
	for k, v := range pl.loadedPlugins {
		result[k] = v
	}
	return result
}

// GetPlugin retrieves a loaded plugin by its cache key.
func (pl *JSPluginLoader) GetPlugin(cacheKey string) (*Plugin, bool) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	plugin, ok := pl.loadedPlugins[cacheKey]
	return plugin, ok
}

// ClearCache clears the plugin cache.
func (pl *JSPluginLoader) ClearCache() {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.loadedPlugins = make(map[string]*Plugin)
}

// CallFunction calls a JavaScript function registered by a plugin.
func (pl *JSPluginLoader) CallFunction(name string, args []any) (any, error) {
	if pl.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := pl.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name": name,
			"args": args,
		},
	})

	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("function call failed: %s", resp.Error)
	}

	return resp.Result, nil
}

// GetRegisteredFunctions returns the names of all functions registered by plugins.
func (pl *JSPluginLoader) GetRegisteredFunctions() ([]string, error) {
	if pl.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := pl.runtime.SendCommand(Command{
		Cmd: "getRegisteredFunctions",
	})

	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get functions: %s", resp.Error)
	}

	var functions []string
	if funcs, ok := resp.Result.([]any); ok {
		for _, f := range funcs {
			if name, ok := f.(string); ok {
				functions = append(functions, name)
			}
		}
	}

	return functions, nil
}

// GetVisitors returns information about registered visitors.
func (pl *JSPluginLoader) GetVisitors() ([]map[string]any, error) {
	if pl.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := pl.runtime.SendCommand(Command{
		Cmd: "getVisitors",
	})

	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get visitors: %s", resp.Error)
	}

	var visitors []map[string]any
	if v, ok := resp.Result.([]any); ok {
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				visitors = append(visitors, m)
			}
		}
	}

	return visitors, nil
}
