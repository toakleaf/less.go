# AGENT 2: Plugin Loader

**Status**: ⏸️ Blocked - Wait for Agent 1 to have working IPC (Phase 1 Tasks 1-3)
**Dependencies**: Agent 1 Phase 1 (Node.js process + IPC)
**Estimated Time**: 3-4 days
**Can work in parallel with**: Agent 1 Phase 2 (serialization)

---

You are implementing plugin loading for JavaScript plugins in less.go using Node.js require().

## Your Mission

Implement Phase 4 (Plugin Loader) from the strategy document.

## Prerequisites

✅ Verify Agent 1 has completed:
- Node.js process spawning works
- Basic IPC (stdin/stdout) works
- Can send/receive commands between Go and Node.js

Check: `go test ./runtime -run TestNodeJSRuntime_CommandResponse`

## Required Reading

BEFORE starting, read these files in `.claude/tasks/js-plugins/`:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 4
2. JavaScript plugin examples in `packages/test-data/plugin/*.js`
3. JavaScript plugin loader: `packages/less/src/less-node/plugin-loader.js`

## Your Tasks

### 1. Parse @plugin Directive (Go Side)

Check if already parsed:
```bash
grep -r "plugin" packages/less/src/less/less_go/parser/
```

If not parsed, add to parser. If already parsed, just use existing node.

### 2. Create Plugin Loader (Go Side)

```go
// In packages/less/src/less/less_go/runtime/plugin_loader.go

type PluginLoader struct {
    runtime       *NodeJSRuntime
    loadedPlugins map[string]*Plugin
}

type Plugin struct {
    filename   string
    id         string
    functions  []string  // Function IDs registered by this plugin
    visitors   []string  // Visitor IDs registered by this plugin
}

func NewPluginLoader(runtime *NodeJSRuntime) *PluginLoader

func (pl *PluginLoader) LoadPlugin(path string, options map[string]interface{}, baseDir string) (*Plugin, error) {
    // 1. Send loadPlugin command to Node.js
    // 2. Receive response with registered functions/visitors
    // 3. Cache plugin
    // 4. Return Plugin object
}
```

### 3. Implement Plugin Host (Node.js Side)

Extend `plugin-host.js`:

```javascript
const path = require('path');
const Module = require('module');

// Plugin registry
const loadedPlugins = new Map();
const registeredFunctions = new Map();
const registeredVisitors = new Map();

// Function registry API for plugins
const functionRegistry = {
    add(name, func) {
        const id = `func_${name}_${Date.now()}`;
        registeredFunctions.set(id, func);
        return id;
    },

    addMultiple(obj) {
        const ids = {};
        for (const [name, func] of Object.entries(obj)) {
            ids[name] = this.add(name, func);
        }
        return ids;
    }
};

// Plugin manager API for plugins
const pluginManager = {
    addVisitor(visitor) {
        const id = `visitor_${Date.now()}`;
        registeredVisitors.set(id, visitor);
        return id;
    }
};

// Less API stub (will be completed by other agents)
const less = {
    // Tree constructors will be added by Agent 3
    // For now, just stub
};

// Command: loadPlugin
function handleLoadPlugin(cmd) {
    const { path: pluginPath, options, baseDir } = cmd;

    try {
        // Resolve path using Node.js resolution
        const resolvedPath = require.resolve(pluginPath, {
            paths: [baseDir, process.cwd()]
        });

        // Check cache
        if (loadedPlugins.has(resolvedPath)) {
            const cached = loadedPlugins.get(resolvedPath);
            return { success: true, plugin: cached };
        }

        // Load plugin using require()
        const plugin = require(resolvedPath);

        // Validate plugin structure
        if (!plugin || typeof plugin !== 'object') {
            throw new Error('Plugin must export an object');
        }

        // Check minVersion if specified
        if (plugin.minVersion) {
            // Less version check (stub for now)
        }

        // Call install() if it exists
        const registeredItems = {
            functions: [],
            visitors: []
        };

        if (plugin.install) {
            // Reset registries to track what this plugin adds
            const beforeFuncs = registeredFunctions.size;
            const beforeVisitors = registeredVisitors.size;

            plugin.install(less, pluginManager, functionRegistry);

            // Collect what was registered
            // (simplified - real version would track IDs)
            registeredItems.functions = Array.from(registeredFunctions.keys()).slice(beforeFuncs);
            registeredItems.visitors = Array.from(registeredVisitors.keys()).slice(beforeVisitors);
        }

        // Call setOptions() if provided
        if (options && plugin.setOptions) {
            plugin.setOptions(options);
        }

        // Cache plugin
        loadedPlugins.set(resolvedPath, {
            path: resolvedPath,
            plugin,
            ...registeredItems
        });

        return {
            success: true,
            plugin: {
                path: resolvedPath,
                ...registeredItems
            }
        };

    } catch (error) {
        return {
            success: false,
            error: error.message,
            stack: error.stack
        };
    }
}

// Add to command handler
process.stdin.on('data', (data) => {
    const lines = data.toString().split('\n').filter(Boolean);

    for (const line of lines) {
        try {
            const cmd = JSON.parse(line);
            let response;

            switch (cmd.cmd) {
                case 'loadPlugin':
                    response = handleLoadPlugin(cmd);
                    break;

                // Other commands added by Agent 1
                default:
                    response = { success: false, error: 'Unknown command' };
            }

            process.stdout.write(JSON.stringify(response) + '\n');

        } catch (error) {
            process.stdout.write(JSON.stringify({
                success: false,
                error: error.message
            }) + '\n');
        }
    }
});
```

### 4. Integrate with Evaluation

Find where imports are evaluated and add plugin loading:

```go
// In appropriate evaluation file
func EvaluatePluginDirective(plugin *PluginDirective, ctx *EvalContext) error {
    loader := ctx.PluginLoader

    p, err := loader.LoadPlugin(
        plugin.Path,
        plugin.Options,
        ctx.CurrentDirectory,
    )

    if err != nil {
        return err
    }

    // Register plugin in current scope (Agent 7 will implement scoping)
    // For now, just add to context
    ctx.LoadedPlugins = append(ctx.LoadedPlugins, p)

    return nil
}
```

### 5. Test with Real Plugins

Test with actual plugin files:

```go
func TestPluginLoader_SimplePlugin(t *testing.T) {
    // Use packages/test-data/plugin/plugin-simple.js
    loader := NewPluginLoader(runtime)

    plugin, err := loader.LoadPlugin(
        "../../test-data/plugin/plugin-simple.js",
        nil,
        ".",
    )

    require.NoError(t, err)
    assert.NotNil(t, plugin)
    assert.Contains(t, plugin.Functions, "pi")
}
```

## Success Criteria

✅ **Complete When**:
- Can load plugins using Node.js `require()`
- NPM modules resolve correctly (e.g., `@plugin "clean-css"`)
- Relative paths work (`@plugin "./plugin.js"`)
- Plugin `install()` and `setOptions()` are called
- Plugin caching works (load once, use multiple times)
- Errors are properly captured and returned to Go
- Unit tests pass for all plugin loading scenarios
- Can load all test plugins in `packages/test-data/plugin/`

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- NO integration test regressions: `pnpm -w test:go` (183/183)

## Test Requirements

```go
func TestPluginLoader_LoadSimplePlugin(t *testing.T)
func TestPluginLoader_LoadNPMModule(t *testing.T)
func TestPluginLoader_RelativePath(t *testing.T)
func TestPluginLoader_PluginCaching(t *testing.T)
func TestPluginLoader_ErrorHandling(t *testing.T)
func TestPluginLoader_SetOptions(t *testing.T)
```

Test with real plugins:
```bash
# Test Node.js can load plugin files
cd packages/test-data/plugin
node -e "console.log(require('./plugin-simple.js'))"

# Test Go can load via runtime
go test -v ./runtime -run TestPluginLoader
```

## Deliverables

1. Working plugin loader (Go and Node.js sides)
2. Support for require(), npm modules, relative paths
3. Plugin caching and lifecycle management
4. Error handling with stack traces
5. All unit tests passing
6. No regressions
7. Brief summary of implementation

Good luck! You're making plugins loadable! 🔌
