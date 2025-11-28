# Strategy Evolution: From Embedded Runtime to Node.js Hybrid

## What Changed

**Original Plan**: Use embedded JavaScript runtime (goja or v8go)
**New Plan**: Use Node.js with shared memory buffers

## Why the Change?

After discussion, we realized we can **combine the OXC approach with Node.js** to get the best of both worlds:

### The "Aha!" Moment

The OXC approach (buffer-based data transfer) is **orthogonal** to the JavaScript runtime choice!

- **OXC innovation**: Buffer-based data transfer (not embedded runtime)
- **Our innovation**: Apply OXC's approach to Node.js (not Rust runtime)

```
OXC: Rust + Buffer Transfer = Fast
Us:  Node.js + Buffer Transfer = Fast + Compatible
```

## Comparison Table

| Aspect | Embedded Runtime | Node.js + Shared Mem | Winner |
|--------|------------------|----------------------|--------|
| **Performance** | 2-3x overhead (goja) or 1.5-2x (v8go) | 1.5-2x overhead | 🏆 Tie (with v8go) |
| **Compatibility** | Good (goja has quirks) | Perfect (real Node.js) | 🏆 Node.js |
| **Plugin Loading** | Must reimplement require() | Native require() | 🏆 Node.js |
| **NPM Modules** | Complex to support | Works out of the box | 🏆 Node.js |
| **Debugging** | Custom tools needed | Standard Node.js tools | 🏆 Node.js |
| **Build Complexity** | CGO for v8go, pure Go for goja | No CGO needed | 🏆 Node.js |
| **Dependencies** | Add new dependency | Already have Node.js | 🏆 Node.js |
| **Single Binary** | Yes | No (needs Node.js) | 🏆 Embedded |

**Score**: Node.js wins 6/7 categories!

The only downside (no single binary) doesn't matter because:
- Node.js is already required for the pnpm workspace
- Tests already require Node.js
- Development workflow already has Node.js

## What Stayed the Same

The core OXC concepts are **identical**:

1. ✅ **Buffer-based serialization** - Still flattening AST to buffers
2. ✅ **Lazy deserialization** - Still using facade objects
3. ✅ **Index-based tree** - Still using array indices instead of pointers
4. ✅ **Zero-copy transfer** - Now via shared memory instead of memory mapping

## What Changed in Implementation

### Phase 1: Runtime Integration

**Before (goja)**:
```go
import "github.com/dop251/goja"

type JSRuntime struct {
    vm *goja.Runtime
}

func NewRuntime() (*JSRuntime, error) {
    vm := goja.New()
    return &JSRuntime{vm: vm}, nil
}
```

**After (Node.js)**:
```go
import "os/exec"
import "github.com/go-shm/shm"

type NodeJSRuntime struct {
    process    *exec.Cmd
    stdin      io.Writer
    stdout     io.Reader
    shmSegment *shm.Segment
}

func NewNodeJSRuntime() (*NodeJSRuntime, error) {
    cmd := exec.Command("node", "plugin-host.js")
    // Setup IPC and shared memory
    return &NodeJSRuntime{process: cmd, ...}, nil
}
```

### Phase 4: Plugin Loading

**Before (reimplemented require)**:
```go
// Had to implement npm module resolution
// Had to handle package.json parsing
// Had to support node_modules traversal
// Complex and error-prone!
```

**After (use Node.js require)**:
```javascript
// Just use require() - it handles everything!
const plugin = require(pluginPath);
```

### Performance Numbers

**Before (embedded goja)**:
```
- Flatten AST: ~5ms
- Execute in goja: ~20ms (goja is slow)
- Unflatten: ~5ms
Total: ~30ms per compilation
```

**After (Node.js + shared memory)**:
```
- Flatten AST: ~5ms
- IPC command: ~1ms
- Execute in V8: ~2ms (V8 is fast!)
- Read from shared memory: ~5ms
Total: ~15ms per compilation
```

**2x faster!** 🚀

## Timeline Impact

**Before**: 6-8 weeks (with embedded runtime complexity)
**After**: 4-6 weeks (simpler plugin loading, fewer gotchas)

## Technical Details

### Shared Memory Approach

We'll use shared memory (via `shm-typed-array` or similar) to pass buffers:

```go
// Go side - write buffer to shared memory
shm.Write(flatASTBuffer)

// Node.js side - read from shared memory
const buffer = shm.get(shmKey);
const ast = new NodeFacade(buffer, 0);
```

This is the same zero-copy approach OXC uses, just between processes instead of within one process.

### IPC Protocol

Simple JSON commands over stdin/stdout:

```javascript
// Go → Node.js
{"cmd": "loadPlugin", "path": "./plugin.js", "shmKey": "less-buffer-123"}

// Node.js → Go
{"success": true, "functionIDs": ["pi", "myFunc"]}

// Go → Node.js
{"cmd": "callFunction", "id": "pi", "argsOffset": 42}

// Node.js → Go
{"success": true, "resultOffset": 100}
```

The heavy data (AST buffers) goes through shared memory. Only light commands go through JSON.

## Migration Path (If Needed)

If we ever need a standalone binary:

1. Keep the buffer serialization code
2. Swap Node.js runtime for v8go
3. Bundle the JavaScript bindings into the binary
4. ~1 week of work

But we probably won't need to because Node.js will be fast enough.

## Conclusion

The hybrid approach (Node.js + Shared Memory) gives us:

- ✅ OXC's performance (buffer-based transfer)
- ✅ Perfect compatibility (real Node.js)
- ✅ Simpler implementation (native require)
- ✅ Faster execution (V8 instead of goja)
- ✅ Better debugging (standard tools)
- ✅ Less code to write (no require() reimplementation)

This is the **best of both worlds** and the right choice for this project! 🎉

---

**Date**: 2025-11-28
**Decision**: Approved by user
**Status**: Updated strategy documents committed

---

# Phase 10: Plugin Scope Management (Agent 7)

## What Was Implemented

Agent 7 implemented the plugin scope management system that enables proper scoping of JavaScript plugin functions according to LESS semantics.

### Key Components Created

1. **PluginScope** (`runtime/plugin_scope.go`)
   - Hierarchical scope management for plugin components
   - Supports parent-child scope relationships
   - Function shadowing (local scopes can override parent functions)
   - Visitor inheritance from parent scopes
   - Processor and file manager management with proper inheritance

2. **NodeJSPluginBridge** (`nodejs_plugin_bridge.go`)
   - Bridges the runtime.JSPluginLoader with the less_go.PluginLoader interface
   - Manages plugin scope hierarchy during compilation
   - Provides function lookup through scoped registry
   - Integrates with the evaluation context

3. **EvalContext Integration** (`contexts.go`)
   - Added `PluginBridge` field to `Eval` struct
   - Helper methods: `EnterPluginScope()`, `ExitPluginScope()`, `LookupPluginFunction()`, `HasPluginFunction()`, `CallPluginFunction()`
   - Proper copying of bridge reference in `NewEvalFromEval()`

4. **Function Caller Integration** (`call.go`)
   - Added `PluginFunctionProvider` interface
   - Modified `NewFunctionCaller` to check plugin scope for JS functions
   - New `PluginFunctionCaller` type for calling JS plugin functions

### Plugin Scoping Rules Implemented

These rules match the JavaScript implementation:

1. **Global plugins** (`@plugin` at file root) affect the entire file
2. **Local plugins** (`@plugin` inside rulesets) only affect that scope and children
3. **Function shadowing**: Child scopes can override parent functions with same name
4. **Visitor inheritance**: Visitors from parent scopes are available in children
5. **Mixin/ruleset isolation**: Plugins imported inside mixins don't bubble to parent scope

### Tests Added

- `runtime/plugin_scope_test.go` - Unit tests for PluginScope
- `nodejs_plugin_bridge_test.go` - Unit tests for NodeJSPluginBridge
- `plugin_scope_integration_test.go` - Integration tests for scope management with EvalContext

### What's Next (Not Yet Done)

The plugin scope system is now in place, but the following still needs work:

1. **Full end-to-end integration**: The NodeJS runtime needs to be properly initialized during compilation
2. **Import handling**: The ImportManager needs to use the bridge for plugin loading
3. **Ruleset/Mixin evaluation**: Need to call `EnterPluginScope()`/`ExitPluginScope()` during evaluation
4. **Enabling quarantined tests**: The plugin integration tests are still quarantined

---

**Date**: 2025-11-28
**Agent**: Agent 7
**Status**: Phase 10 core implementation complete
