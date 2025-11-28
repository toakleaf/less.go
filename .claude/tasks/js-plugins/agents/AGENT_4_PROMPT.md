# AGENT 4: Function Registry

**Status**: ⏸️ Blocked - Wait for Agent 3 (Bindings)
**Dependencies**: Agent 3 (must have NodeFacade and constructors)
**Estimated Time**: 3-4 days

---

You are implementing bidirectional function calls between Go and JavaScript for the less.go plugin system.

## Your Mission

Implement Phase 5 (Function Registry Integration) from the strategy document.

## Prerequisites

✅ Verify Agent 3 has completed:
- NodeFacade works (can read nodes from buffer)
- Node constructors work (can create nodes)
- Shared memory buffer read/write works

Check: `go test ./runtime -run TestBindings`

## Required Reading

BEFORE starting, read:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 5
2. `packages/less/src/less/less_go/functions/` - Existing function registry
3. Plugin examples: `packages/test-data/plugin/plugin-simple.js`

## Your Tasks

### 1. Extend Function Registry (Go Side)

Modify `packages/less/src/less/less_go/functions/function_registry.go`:

```go
type FunctionRegistry struct {
    builtins  map[string]Function
    jsPlugins map[string]*JSFunction
}

type JSFunction struct {
    name       string
    functionID string
    runtime    *runtime.NodeJSRuntime
}

func (jf *JSFunction) Call(args []Node, ctx *EvalContext) (Node, error) {
    // 1. Flatten args to shared memory
    argsFlat, err := runtime.FlattenNodes(args)
    if err != nil {
        return nil, err
    }

    argsOffset, err := jf.runtime.WriteArgsBuffer(argsFlat)
    if err != nil {
        return nil, err
    }

    // 2. Send command to Node.js
    resp, err := jf.runtime.SendCommand(runtime.Command{
        Cmd:        "callFunction",
        FunctionID: jf.functionID,
        ArgsOffset: argsOffset,
    })

    if err != nil {
        return nil, err
    }

    // 3. Read result from shared memory
    resultFlat, err := jf.runtime.ReadResultBuffer(resp.ResultOffset)
    if err != nil {
        return nil, err
    }

    // 4. Unflatten to Go node
    result, err := runtime.UnflattenNode(resultFlat)
    if err != nil {
        return nil, err
    }

    return result, nil
}

func (fr *FunctionRegistry) AddJSFunction(name string, functionID string, runtime *runtime.NodeJSRuntime) {
    fr.jsPlugins[name] = &JSFunction{
        name:       name,
        functionID: functionID,
        runtime:    runtime,
    }
}

func (fr *FunctionRegistry) Call(name string, args []Node, ctx *EvalContext) (Node, error) {
    // Check JS plugins first (allow shadowing)
    if jsFn, ok := fr.jsPlugins[name]; ok {
        return jsFn.Call(args, ctx)
    }

    // Fall back to built-in functions
    if fn, ok := fr.builtins[name]; ok {
        return fn.Call(args, ctx)
    }

    return nil, fmt.Errorf("function not found: %s", name)
}
```

### 2. Implement Function Calls (Node.js Side)

Update `plugin-host.js`:

```javascript
const { NodeFacade } = require('./bindings/node-facade');
const constructors = require('./bindings/constructors');

// Function registry for plugins
const registeredFunctions = new Map();

const functionRegistry = {
    add(name, func) {
        const id = `func_${name}_${Date.now()}`;
        registeredFunctions.set(id, { name, func });
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

// Command: callFunction
function handleCallFunction(cmd) {
    const { functionID, argsOffset } = cmd;

    try {
        const funcInfo = registeredFunctions.get(functionID);
        if (!funcInfo) {
            throw new Error(`Function not found: ${functionID}`);
        }

        // Read args from shared memory buffer
        const args = readArgsFromBuffer(sharedBuffer, argsOffset);

        // Wrap args in NodeFacade
        const facadeArgs = args.map((arg, idx) =>
            new NodeFacade(sharedBuffer, arg.index)
        );

        // Call the JavaScript function
        const result = funcInfo.func(...facadeArgs);

        // Result might be a NodeFacade or a primitive
        // Write result to shared memory
        const resultOffset = writeResultToBuffer(sharedBuffer, result);

        return {
            success: true,
            resultOffset
        };

    } catch (error) {
        return {
            success: false,
            error: error.message,
            stack: error.stack
        };
    }
}

// Helper: Read args from shared memory
function readArgsFromBuffer(buffer, offset) {
    // Read arg count
    const argCount = buffer.readUInt32(offset);
    const args = [];

    let currentOffset = offset + 4;
    for (let i = 0; i < argCount; i++) {
        const nodeIndex = buffer.readUInt32(currentOffset);
        args.push({ index: nodeIndex });
        currentOffset += 4;
    }

    return args;
}

// Helper: Write result to shared memory
function writeResultToBuffer(buffer, result) {
    // If result is a NodeFacade, it's already in the buffer
    if (result instanceof NodeFacade) {
        return result._index;
    }

    // If result is a constructor result, return its index
    if (result && result._index !== undefined) {
        return result._index;
    }

    // Handle primitives (shouldn't happen, but just in case)
    throw new Error('Function must return a node, got: ' + typeof result);
}

// Add to command handler
process.stdin.on('data', (data) => {
    // ... existing code ...

    switch (cmd.cmd) {
        case 'callFunction':
            response = handleCallFunction(cmd);
            break;
        // ... other cases
    }
});

// Expose to plugins
const less = {
    dimension: constructors.dimension,
    color: constructors.color,
    // ... all constructors from Agent 3
};
```

### 3. Test with Real Plugin Functions

Test with `plugin-simple.js`:

```go
func TestJSFunction_PluginSimple(t *testing.T) {
    runtime := setupRuntime(t)
    defer runtime.Stop()

    // Load plugin-simple.js
    loader := NewPluginLoader(runtime)
    plugin, err := loader.LoadPlugin(
        "../../test-data/plugin/plugin-simple.js",
        nil,
        ".",
    )
    require.NoError(t, err)

    // Get function registry
    registry := functions.NewRegistry()

    // Register JS function
    for _, funcID := range plugin.Functions {
        registry.AddJSFunction("pi", funcID, runtime)
    }

    // Call function from Go
    ctx := &EvalContext{}
    result, err := registry.Call("pi", []Node{}, ctx)
    require.NoError(t, err)

    // Verify result
    dim, ok := result.(*Dimension)
    require.True(t, ok)
    assert.InDelta(t, math.Pi, dim.Value, 0.0001)
}
```

## Success Criteria

✅ **Complete When**:
- Go can call JavaScript functions via registry
- Arguments pass through shared memory (no serialization overhead)
- Results return through shared memory
- All node types can be passed as arguments
- All node types can be returned as results
- Error handling works (JS errors propagate to Go)
- Stack traces are preserved
- Unit tests pass for various function signatures

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- NO integration test regressions: `pnpm -w test:go` (183/183)

## Test Requirements

```go
func TestJSFunction_Simple(t *testing.T)
func TestJSFunction_WithArgs(t *testing.T)
func TestJSFunction_AllNodeTypes(t *testing.T)
func TestJSFunction_ErrorHandling(t *testing.T)
func TestJSFunction_Shadowing(t *testing.T)
```

Test with real plugins:
```bash
# Test plugin-simple
go test -v ./runtime -run TestPluginSimple

# Verify no regressions
pnpm -w test:go:unit
```

## Deliverables

1. Extended FunctionRegistry supporting JS functions
2. Bidirectional function calls via shared memory
3. Error handling with stack traces
4. All unit tests passing
5. No regressions
6. Brief summary

You're making plugins callable! 📞
