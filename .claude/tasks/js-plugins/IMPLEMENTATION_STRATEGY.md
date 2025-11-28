# JavaScript Plugin Implementation Strategy for less.go

## Executive Summary

This document outlines the strategy for implementing JavaScript plugin support in the less.go project, inspired by the OXC project's approach to running JavaScript linting rules from Rust. Our goal is to enable the Go LESS compiler to execute JavaScript plugins with minimal performance overhead while maintaining 100% API compatibility with less.js.

## Background

### What are LESS Plugins?

LESS plugins are JavaScript modules that extend LESS functionality through:
1. **Custom Functions**: Add new functions callable from LESS (e.g., `pi()`, `custom-color()`)
2. **AST Visitors**: Transform the parse tree before/after evaluation (pre-eval, post-eval)
3. **Pre/Post Processors**: Transform raw text before parsing or after compilation
4. **File Managers**: Custom import resolution logic

### Current Status

- ✅ **183/183 tests passing** (100% success rate for non-plugin tests)
- ⏸️ **8 quarantined tests** requiring JavaScript plugins:
  - `plugin` - Basic plugin functionality
  - `plugin-module` - NPM module plugins
  - `plugin-preeval` - Pre-evaluation visitors
  - `bootstrap4` - Real-world usage (uses map-get, breakpoint-next, etc.)

## The OXC Approach: Key Learnings

### 1. Raw Transfer (Zero-Copy Serialization)

**OXC's Innovation**: Instead of serializing ASTs to JSON, they pass Rust's native memory layout directly to JavaScript as Node.js Buffer objects.

**Performance Impact**: Eliminates 80%+ of the overhead in cross-language communication.

**Our Application**:
- Go AST → Flatten to buffer → Pass to JavaScript
- JavaScript modifications → Update buffer → Go reads changes

### 2. Lazy Deserialization

**OXC's Innovation**: Create JavaScript proxy objects that read from buffers on-demand, only materializing objects when accessed.

**Performance Impact**: Reduces garbage collector pressure and CPU time by 4-5x.

**Our Application**:
- Visitor pattern only materializes nodes that plugins actually visit
- Most nodes stay in buffer format, never converted to JS objects

### 3. Flattened AST with Index References

**OXC's Innovation**: Convert tree pointers to array indices, enabling linear buffer storage.

**Structure**:
```
Each node: [Type, Child Index, Next Sibling Index, Parent Index]
```

**Our Application**:
- Similar structure but adapted for LESS nodes
- String values stored in separate string table to minimize boundary crossings

## Why Node.js + Shared Memory? (Hybrid Approach)

After analyzing the options, we're choosing **Node.js with shared memory** instead of embedded runtimes (goja/v8go):

### The Decision

| Approach | Performance | Compatibility | Complexity | Verdict |
|----------|------------|---------------|------------|---------|
| **JSON over stdio** | ❌ 10-20x overhead | ✅ Perfect | ✅ Simple | Too slow |
| **Embedded goja** | ⚠️ 2-3x overhead | ⚠️ Good but quirks | ⚠️ Medium | Acceptable |
| **Embedded v8go** | ✅ 1.5-2x overhead | ✅ Perfect | ❌ CGO, complex | Good but hard |
| **Node.js + Shared Mem** | ✅ 1.5-2x overhead | ✅ Perfect | ✅ Medium | **Best!** ✅ |

### Advantages of Node.js + Shared Memory

1. **Perfect Compatibility**: Real Node.js means perfect plugin compatibility (no quirks!)
2. **Fast V8 Engine**: 10x faster JavaScript execution than goja
3. **OXC Performance**: Still gets zero-copy buffer transfer (no JSON serialization!)
4. **Simple Plugin Loading**: Native `require()` - no need to reimplement npm resolution
5. **Already Required**: Node.js is already needed for the pnpm workspace
6. **Better Debugging**: Standard Node.js debugging tools work
7. **No CGO**: Unlike v8go, doesn't require C dependencies

### How It Combines Both Approaches

```
┌─────────────────────────────────────────────────────┐
│ Traditional Node.js Approach (esbuild style)        │
│ - JSON over stdio                                   │
│ - 10-20x overhead                                   │
│ ❌ Too slow                                         │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ OXC Approach (embedded runtime with buffers)        │
│ - Buffer-based transfer                             │
│ - 2-5x overhead                                     │
│ ✅ Good, but still overhead from slow JS runtime   │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│ Our Hybrid: Node.js + Shared Memory                │
│ - OXC's buffer approach ✅                          │
│ - Fast V8 execution ✅                              │
│ - Perfect compatibility ✅                          │
│ - 1.5-2x overhead ✅                                │
│ 🎯 BEST OF BOTH WORLDS                             │
└─────────────────────────────────────────────────────┘
```

## Proposed Architecture for less.go

### Phase 1: Node.js Process Integration

**Goal**: Integrate with Node.js for JavaScript plugin execution.

**Architecture Decision**: **Node.js + Shared Memory**

**Why Node.js?**
- ✅ Already required for the pnpm workspace
- ✅ Perfect plugin compatibility (real `require()`, full npm ecosystem)
- ✅ Production-ready V8 engine (faster than embedded runtimes)
- ✅ Easier debugging (standard Node.js tools)
- ✅ Proven approach (similar to esbuild, swc)

**Why Shared Memory?**
- ✅ Combines OXC's buffer approach with Node.js
- ✅ No JSON serialization overhead
- ✅ Zero-copy data transfer
- ✅ Best of both worlds: compatibility + performance

**Design**:

```go
type NodeJSRuntime struct {
    process     *exec.Cmd
    stdin       io.Writer
    stdout      io.Reader
    shmSegment  *shm.Segment  // Shared memory for buffers
    alive       bool
}

func NewNodeJSRuntime() (*NodeJSRuntime, error) {
    // 1. Spawn Node.js process running plugin-host.js
    // 2. Establish shared memory segment
    // 3. Keep process alive for duration of compilation
}
```

**Node.js Side** (`plugin-host.js`):
```javascript
const shm = require('shm-typed-array');

// Connect to shared memory from Go
const buffer = shm.get(process.env.LESS_SHM_KEY);

// Lazy facade pattern (same as OXC approach!)
class NodeFacade {
    constructor(buffer, index) {
        this._buffer = buffer;
        this._index = index;
    }

    get type() {
        const typeID = this._buffer.nodes[this._index].typeID;
        return this._buffer.typeTable[typeID];
    }
    // ... rest of facade
}

// Standard Node.js require() - no reimplementation needed!
function loadPlugin(pluginPath) {
    return require(pluginPath);
}
```

**Tasks**:
1. Implement Node.js process spawning and lifecycle management
2. Create shared memory segment for buffer transfer
3. Implement basic IPC protocol (commands over stdin/stdout)
4. Create `plugin-host.js` Node.js script
5. Unit tests for process management and communication

### Phase 2: AST Serialization (Raw Transfer)

**Goal**: Implement zero-copy AST transfer using buffer-based serialization.

**Design**:

```go
// Flattened AST buffer structure
type FlatNode struct {
    TypeID        uint16  // Index into type string table
    ChildIndex    uint32  // Index of first child, 0 if none
    NextIndex     uint32  // Index of next sibling, 0 if none
    ParentIndex   uint32  // Index of parent, 0 if root
    ValueIndex    uint32  // Index into value buffer (for literals)
    PropsOffset   uint32  // Offset into properties buffer
}

type FlatAST struct {
    Nodes       []FlatNode
    TypeTable   []string    // Node type names
    StringTable []string    // String values
    PropBuffer  []byte      // Node-specific properties (JSON)
}
```

**Serialization Strategy**:
1. **Tree Flattening**: Walk AST depth-first, assign indices
2. **Type Deduplication**: Map node types to string table indices
3. **Value Storage**: Separate buffer for string/number literals
4. **Properties**: JSON-encode node-specific data (optional properties)

**Buffer Layout**:
```
+------------------+
| Header (32 bytes)|  Magic, version, node count, offsets
+------------------+
| Node Array       |  Fixed-size FlatNode structs
+------------------+
| Type Table       |  Length-prefixed strings
+------------------+
| String Table     |  Length-prefixed strings
+------------------+
| Prop Buffer      |  Variable-length property data
+------------------+
```

**Tasks**:
1. Implement `FlattenAST(root Node) *FlatAST`
2. Implement `UnflattenAST(flat *FlatAST) Node`
3. Write comprehensive roundtrip tests
4. Benchmark serialization performance

### Phase 3: JavaScript Bindings (Lazy Deserialization)

**Goal**: Create JavaScript facade objects that read from shared memory buffers on-demand.

**Design**:

The bindings run in **Node.js** and read from shared memory mapped by Go.

**JavaScript Side** (`plugin-bindings.js` - runs in Node.js):

```javascript
const shm = require('shm-typed-array');

// Access shared memory segment
let sharedBuffer = null;

function attachBuffer(shmKey) {
    sharedBuffer = shm.get(shmKey, 'Uint32Array');
}

// Lazy facade - reads from shared memory on-demand
class NodeFacade {
    constructor(buffer, index) {
        this._buffer = buffer || sharedBuffer;
        this._index = index;
    }

    get type() {
        const typeID = this._buffer.nodes[this._index].typeID;
        return this._buffer.typeTable[typeID];
    }

    get value() {
        const valueIndex = this._buffer.nodes[this._index].valueIndex;
        if (valueIndex === 0) return undefined;
        return this._buffer.stringTable[valueIndex];
    }

    get children() {
        const childIndex = this._buffer.nodes[this._index].childIndex;
        if (childIndex === 0) return [];
        return this._collectSiblings(childIndex);
    }

    _collectSiblings(startIndex) {
        const result = [];
        let index = startIndex;
        while (index !== 0) {
            result.push(new NodeFacade(this._buffer, index));
            index = this._buffer.nodes[index].nextIndex;
        }
        return result;
    }
}

// Visitor pattern support (same as OXC approach)
class VisitorContext {
    constructor(buffer) {
        this._buffer = buffer;
    }

    visit(nodeIndex) {
        const node = new NodeFacade(this._buffer, nodeIndex);
        const method = 'visit' + node.type;
        if (this[method]) {
            return this[method](node);
        }
        return node;
    }
}

module.exports = { NodeFacade, VisitorContext, attachBuffer };
```

**Go Side** (coordinates shared memory):
```go
func (rt *NodeJSRuntime) SendBuffer(flatAST *FlatAST) error {
    // 1. Write flat AST to shared memory
    // 2. Send command to Node.js: {"cmd": "setBuffer", "shmKey": "..."}
    // 3. Node.js maps shared memory and reads buffer
}
```

**Tasks**:
1. Implement shared memory wrapper for Node.js (`shm-typed-array` or similar)
2. Generate JavaScript binding code from Go AST node definitions
3. Implement visitor pattern support
4. Add node construction helpers (for plugin-created nodes)
5. Write JavaScript unit tests for bindings (can run with `node` directly!)

### Phase 4: Plugin Loader

**Goal**: Parse `@plugin` directives and load JavaScript files.

**Design**:

With Node.js, plugin loading is **dramatically simpler** - we use native `require()`!

**Go Side**:
```go
type PluginLoader struct {
    nodeRuntime   *NodeJSRuntime
    loadedPlugins map[string]*Plugin
    fileManager   *FileManager
}

type Plugin struct {
    filename      string
    id            string  // Unique ID for this plugin instance
    functions     map[string]*JSFunction
    visitors      []JSVisitor
    fileManagers  []JSFileManager
}

func (pl *PluginLoader) LoadPlugin(path string, options map[string]interface{}) (*Plugin, error) {
    // 1. Send command to Node.js: {"cmd": "loadPlugin", "path": path, "options": options}
    // 2. Node.js uses require() to load the plugin
    // 3. Node.js calls plugin.install() and returns registered items
    // 4. Go receives list of functions, visitors, etc.
    // 5. Create Plugin object with references
}
```

**Node.js Side** (`plugin-host.js`):
```javascript
const path = require('path');

// Plugin loading - just use require()!
function loadPlugin(pluginPath, options, baseDir) {
    const resolvedPath = path.resolve(baseDir, pluginPath);

    // Native require() handles:
    // - npm modules (node_modules resolution)
    // - Relative paths (./plugin.js)
    // - Absolute paths (/full/path.js)
    // - package.json main field
    // - .js extension inference
    const plugin = require(resolvedPath);

    // Call install if it exists
    if (plugin.install) {
        plugin.install(lessAPI, pluginManager, functionRegistry);
    }

    // Call setOptions if provided
    if (options && plugin.setOptions) {
        plugin.setOptions(options);
    }

    // Return what was registered
    return {
        functions: functionRegistry.getAll(),
        visitors: pluginManager.getVisitors(),
        fileManagers: pluginManager.getFileManagers()
    };
}

// Handle commands from Go
process.stdin.on('data', (data) => {
    const cmd = JSON.parse(data);

    if (cmd.cmd === 'loadPlugin') {
        const result = loadPlugin(cmd.path, cmd.options, cmd.baseDir);
        process.stdout.write(JSON.stringify({success: true, result}));
    }
});
```

**Advantages over embedded runtime**:
- ✅ No need to reimplement `require()` - Node.js handles it
- ✅ Perfect npm module resolution (node_modules, package.json, etc.)
- ✅ Plugins can use `require()` for their own dependencies
- ✅ Works with TypeScript plugins (if they're compiled)
- ✅ Standard Node.js debugging tools work

**Tasks**:
1. Implement IPC protocol for plugin loading commands
2. Create `plugin-host.js` with require()-based loader
3. Handle plugin file resolution (Node.js does the heavy lifting!)
4. Parse `@plugin` directive options
5. Cache loaded plugins
6. Handle plugin errors and forward to Go

### Phase 5: Function Registry Integration

**Goal**: Enable plugins to register custom LESS functions.

**Design**:

```go
type JSFunction struct {
    name       string
    functionID string  // ID in Node.js process
    runtime    *NodeJSRuntime
}

func (jf *JSFunction) Call(args []Node, ctx *EvalContext) (Node, error) {
    // 1. Flatten args to shared memory buffer
    jf.runtime.WriteArgsBuffer(args)

    // 2. Send command to Node.js: {"cmd": "callFunction", "id": functionID, "argsOffset": offset}
    response := jf.runtime.SendCommand(CallFunctionCommand{
        FunctionID: jf.functionID,
        ArgsOffset: offset,
    })

    // 3. Read result from shared memory (Node.js wrote it there)
    result := jf.runtime.ReadResultBuffer(response.ResultOffset)

    // 4. Unflatten to Go Node
    return UnflattenNode(result), nil
}

type FunctionRegistry struct {
    builtins  map[string]Function
    jsPlugins map[string]*JSFunction
}

func (fr *FunctionRegistry) Add(name string, fn interface{}) {
    switch f := fn.(type) {
    case Function:
        fr.builtins[name] = f
    case *JSFunction:
        fr.jsPlugins[name] = f
    }
}
```

**Node.js Side**:
```javascript
// When function is registered by plugin
functionRegistry.add('myFunc', function(arg1, arg2) {
    // These args are NodeFacade objects reading from shared memory!
    const value = arg1.value;
    return less.dimension(value * 2, 'px');
});

// When Go calls function
function callFunction(functionID, argsOffset) {
    const func = registeredFunctions.get(functionID);

    // Args are in shared memory - create facades
    const args = readArgsFromBuffer(sharedBuffer, argsOffset);

    // Call function
    const result = func(...args);

    // Write result back to shared memory
    const resultOffset = writeResultToBuffer(sharedBuffer, result);

    return {resultOffset};
}
```

**Performance**: Function calls are fast because:
- ✅ Args passed via shared memory (no serialization)
- ✅ Result returned via shared memory (no serialization)
- ✅ Only small command sent over IPC (function ID + offset)

**Tasks**:
1. Extend FunctionRegistry to support JS functions
2. Implement argument flattening to shared memory
3. Implement result reading from shared memory
4. Add error handling and stack traces
5. Write unit tests for various function signatures

### Phase 6: Visitor Integration (Pre-eval/Post-eval)

**Goal**: Allow plugins to transform AST before and after evaluation.

**Design**:

```go
type JSVisitor struct {
    jsObj         goja.Value
    isPreEval     bool
    isReplacing   bool
    runtime       *runtime.JSRuntime
}

func (jv *JSVisitor) Visit(node Node) (Node, error) {
    // 1. Flatten AST starting from node
    // 2. Create buffer and pass to JS
    // 3. Call visitor.run(root)
    // 4. Unflatten result
    // 5. Return modified node
}

// Integration point in evaluation
func (e *Evaluator) EvaluateTree(root Node, ctx *EvalContext) (Node, error) {
    // Pre-eval visitors
    for _, visitor := range ctx.PluginManager.PreEvalVisitors() {
        root, err = visitor.Visit(root)
        if err != nil {
            return nil, err
        }
    }

    // Normal evaluation
    result, err := root.Eval(ctx)

    // Post-eval visitors
    for _, visitor := range ctx.PluginManager.PostEvalVisitors() {
        result, err = visitor.Visit(result)
        if err != nil {
            return nil, err
        }
    }

    return result, nil
}
```

**Tasks**:
1. Implement JSVisitor wrapper
2. Add visitor registration to PluginManager
3. Integrate pre-eval hooks in evaluation pipeline
4. Integrate post-eval hooks in evaluation pipeline
5. Handle node replacement vs. mutation
6. Write visitor integration tests

### Phase 7: Tree Node Constructors

**Goal**: Allow plugins to create LESS AST nodes from JavaScript.

**Design**:

```javascript
// Exposed to plugins via `less` object
const less = {
  // Node constructors
  dimension: (value, unit) => createNode('Dimension', {value, unit}),
  color: (rgb, alpha) => createNode('Color', {rgb, alpha}),
  quoted: (quote, value, escaped) => createNode('Quoted', {quote, value, escaped}),
  keyword: (value) => createNode('Keyword', {value}),
  url: (value) => createNode('URL', {value}),
  call: (name, args) => createNode('Call', {name, args}),
  // ... all node types

  // Visitor base class
  visitors: {
    Visitor: class Visitor { /* ... */ }
  }
};

function createNode(type, props) {
  // Allocate new node in buffer
  // Return facade object
}
```

**Tasks**:
1. Generate constructor functions for all node types
2. Implement node allocation in flat buffer
3. Handle complex node properties (arrays, nested nodes)
4. Add validation for required properties
5. Write tests for each constructor

### Phase 8: Pre/Post Processors

**Goal**: Allow plugins to transform raw text before/after compilation.

**Design**:

```go
type JSPreProcessor struct {
    jsFunc   goja.Callable
    priority int
    runtime  *runtime.JSRuntime
}

func (jpp *JSPreProcessor) Process(input string, options map[string]interface{}) (string, error) {
    // Call JS function with input string
    // Return modified string
}

// Integration in parser
func Parse(input string, options *Options) (Node, error) {
    // Pre-processors
    for _, proc := range options.PluginManager.PreProcessors() {
        input, err = proc.Process(input, options)
        if err != nil {
            return nil, err
        }
    }

    // Parse
    ast, err := parser.Parse(input)

    // ... evaluation ...

    // Post-processors
    output := ast.ToCSS()
    for _, proc := range options.PluginManager.PostProcessors() {
        output, err = proc.Process(output, options)
        if err != nil {
            return nil, err
        }
    }

    return output, nil
}
```

**Tasks**:
1. Implement JSPreProcessor wrapper
2. Implement JSPostProcessor wrapper
3. Add priority-based ordering
4. Integrate in parse/compile pipeline
5. Write processor tests

### Phase 9: File Manager Support

**Goal**: Allow plugins to implement custom import resolution.

**Design**:

```go
type JSFileManager struct {
    jsObj    goja.Value
    runtime  *runtime.JSRuntime
}

func (jfm *JSFileManager) Supports(filename string, currentDirectory string) bool {
    // Call JS: fileManager.supports(filename, currentDirectory)
}

func (jfm *JSFileManager) LoadFile(filename string, currentDirectory string) (*LoadedFile, error) {
    // Call JS: fileManager.loadFile(filename, currentDirectory)
    // Return {filename, contents}
}

// Integration in import manager
func (im *ImportManager) ResolveImport(path string) (*LoadedFile, error) {
    // Try plugin file managers first
    for _, fm := range im.PluginManager.FileManagers() {
        if fm.Supports(path, im.currentDir) {
            return fm.LoadFile(path, im.currentDir)
        }
    }

    // Fall back to default file manager
    return im.defaultFileManager.LoadFile(path, im.currentDir)
}
```

**Tasks**:
1. Implement JSFileManager wrapper
2. Add file manager registration
3. Integrate in import resolution pipeline
4. Handle async file loading (if needed)
5. Write file manager tests

### Phase 10: Plugin Scope Management

**Goal**: Handle global vs. local plugin scoping.

**Design**:

```go
type PluginScope struct {
    parent    *PluginScope
    plugins   []*Plugin
    functions map[string]*JSFunction
    visitors  []JSVisitor
}

func (ps *PluginScope) Lookup(name string) (*JSFunction, bool) {
    // Check local scope
    if fn, ok := ps.functions[name]; ok {
        return fn, true
    }

    // Check parent scopes
    if ps.parent != nil {
        return ps.parent.Lookup(name)
    }

    return nil, false
}

// In evaluation context
type EvalContext struct {
    // ...
    PluginScope *PluginScope
}

// When entering ruleset/mixin
func (r *Ruleset) Eval(ctx *EvalContext) (Node, error) {
    // Create child scope
    childScope := &PluginScope{parent: ctx.PluginScope}
    childCtx := ctx.Clone()
    childCtx.PluginScope = childScope

    // Evaluate @plugin directives in this ruleset
    for _, stmt := range r.Rules {
        if pluginStmt, ok := stmt.(*PluginDirective); ok {
            plugin, err := LoadPlugin(pluginStmt.Path, childCtx)
            if err != nil {
                return nil, err
            }
            childScope.AddPlugin(plugin)
        }
    }

    // Evaluate other rules with child scope
    // ...
}
```

**Tasks**:
1. Implement PluginScope hierarchy
2. Add scope creation/destruction in evaluation
3. Handle function shadowing
4. Handle visitor scope propagation
5. Write scope isolation tests

## Testing Strategy

### Unit Tests

For each phase, write focused unit tests:

1. **Runtime Tests**: Execute simple JS, call functions, handle errors
2. **Serialization Tests**: Roundtrip all node types, verify data integrity
3. **Binding Tests**: Access node properties, navigate tree, modify nodes
4. **Loader Tests**: Load local files, NPM modules, handle errors
5. **Function Tests**: Call JS functions with various argument types
6. **Visitor Tests**: Transform nodes, replace nodes, handle recursion
7. **Scope Tests**: Local/global plugin isolation, shadowing

### Integration Tests

Use existing quarantined tests:

1. **`plugin-simple.less`**: Basic function registration
2. **`plugin-tree-nodes.less`**: Node construction from JS
3. **`plugin-preeval.less`**: Pre-eval visitor transformation
4. **`plugin.less`**: Comprehensive plugin features and scoping
5. **`plugin-module.less`**: NPM module loading
6. **`bootstrap4`**: Real-world usage (stretch goal)

### Performance Benchmarks

Measure overhead at each layer:

1. **Serialization**: Time to flatten/unflatten AST
2. **Cross-language calls**: Time per function call
3. **Visitor overhead**: Time per node visit
4. **End-to-end**: Compare plugin vs. native implementation

**Target**: Plugin overhead should be < 2x vs. native Go implementation for typical use cases.

## Work Breakdown & Parallelization

### Phase Dependencies

```
Phase 1 (Runtime)
  ↓
Phase 2 (Serialization) ← Can be developed in parallel with Phase 4
  ↓
Phase 3 (Bindings)
  ↓
Phase 4 (Loader) ← Can start after Phase 1
  ↓
Phase 5 (Functions) ←┐
Phase 6 (Visitors)   ├── Can be done in parallel
Phase 7 (Constructors)←┘
  ↓
Phase 8 (Processors) ←┐
Phase 9 (File Managers)├── Can be done in parallel
Phase 10 (Scoping)    ←┘
```

### Suggested Agent Assignment

**Agent 1: Node.js Process & Serialization** (Critical Path)
- Phase 1: Node.js process management, IPC, shared memory
- Phase 2: AST serialization to buffer format
- Deliverable: Can spawn Node.js, send/receive data via shared memory
- **Estimated**: 1 week

**Agent 2: Plugin Loader** (Can Start Early)
- Phase 4: Plugin loader (Go and Node.js sides)
- Create `plugin-host.js` with require()-based loading
- Deliverable: `@plugin` directive loads plugins via Node.js
- **Estimated**: 3-4 days

**Agent 3: Bindings & Constructors** (Needs Phase 2)
- Phase 3: JavaScript bindings (NodeFacade, lazy deserialization)
- Phase 7: Tree node constructors
- Deliverable: Complete node API for plugins
- **Estimated**: 5-7 days

**Agent 4: Functions & Registry** (Needs Phase 3)
- Phase 5: Function registry integration
- Bidirectional function calls via shared memory
- Deliverable: Plugins can add custom functions
- **Estimated**: 3-4 days

**Agent 5: Visitors & Evaluation** (Needs Phase 3)
- Phase 6: Visitor integration
- Deliverable: Pre-eval and post-eval transformations
- **Estimated**: 3-4 days

**Agent 6: Processors & File Managers** (Can Be Parallel)
- Phase 8: Pre/post processors
- Phase 9: File manager support
- Deliverable: Text transformation and custom imports
- **Estimated**: 3-4 days

**Agent 7: Scoping & Integration** (Needs Most Phases)
- Phase 10: Plugin scope management
- Integration testing with quarantined tests
- Deliverable: Complete plugin system with proper scoping
- **Estimated**: 4-5 days

**Total Timeline**: 4-6 weeks with parallel agents

## Performance Considerations

### Expected Overhead

**Node.js + Shared Memory** approach:

- **Node.js startup** (one-time): ~100ms
- **Per-compilation overhead**: ~5-15ms
  - Flatten AST: ~5ms
  - IPC command: ~1ms
  - Execute plugin (V8): ~2ms
  - Read result: ~5ms

**Comparison to alternatives**:

- **Traditional JSON approach**: ~200ms per compilation (10-20x overhead) ❌
- **Embedded goja**: ~30ms per compilation (2-3x overhead) ⚠️
- **Node.js + shared memory**: ~15ms per compilation (1.5-2x overhead) ✅
- **Native Go**: ~10ms baseline ⭐

**Key Insight**: Node.js + shared memory combines:
- V8's fast JavaScript execution (~10x faster than goja)
- OXC's zero-copy data transfer (no JSON serialization)
- Result: **Best of both worlds!**

### Optimization Opportunities

1. **Buffer Reuse**: Pool shared memory segments to reduce allocations
2. **Incremental Updates**: Only serialize changed subtrees for visitors
3. **Keep Node.js Alive**: Reuse the same process across multiple compilations
4. **Batch Commands**: Send multiple function calls in one IPC roundtrip
5. **Partial Visitor**: Only flatten visited node paths (lazy evaluation)

### When to Optimize

1. **First**: Get it working with basic implementation
2. **Second**: Measure real-world usage (bootstrap4 test)
3. **Third**: Optimize hot paths if measurements show bottlenecks
4. **Note**: With Node.js + shared memory, performance should be acceptable from day 1!

## Risks & Mitigations

### Risk 1: Performance Unacceptable

**Mitigation**:
- Node.js + shared memory should give acceptable performance from start
- Benchmark early and often anyway
- If needed, optimize buffer reuse and batch IPC commands
- Worst case: Keep Node.js process warm, pool shared memory segments

### Risk 2: API Incompatibility

**Mitigation**:
- Use less.js tests as specification
- Test with real-world plugins (bootstrap4)
- Document any intentional deviations

### Risk 3: Complexity Explosion

**Mitigation**:
- Keep phases small and focused
- Write tests for each component
- Regular integration testing
- Code reviews between phases

### Risk 4: Debugging Difficulty

**Mitigation**:
- Implement detailed error messages
- Add source map support for JS stack traces
- Create debugging tools (AST inspector, plugin console)

## Success Criteria

### Minimum Viable Product (MVP)

- ✅ All 8 quarantined tests pass
- ✅ Plugin function registration works
- ✅ Plugin visitors work (pre-eval, post-eval)
- ✅ Plugin scoping works (global, local)
- ✅ No regressions in existing 183 tests

### Stretch Goals

- 🎯 bootstrap4 test passes completely
- 🎯 Performance overhead < 3x vs. native
- 🎯 Support for NPM plugin ecosystem
- 🎯 Developer documentation and examples

## Timeline & Milestones

### Week 1-2: Foundation
- Phase 1: Runtime integration (Agent 1)
- Phase 4: Plugin loader skeleton (Agent 2)

### Week 3-4: Core Functionality
- Phase 2: Serialization (Agent 1)
- Phase 3: Bindings (Agent 3)
- Phase 5: Functions (Agent 4)

### Week 5-6: Advanced Features
- Phase 6: Visitors (Agent 5)
- Phase 7: Constructors (Agent 3)
- Phase 8: Processors (Agent 6)

### Week 7-8: Integration & Polish
- Phase 9: File managers (Agent 6)
- Phase 10: Scoping (Agent 7)
- Integration testing all phases
- Performance optimization

## References

1. [OXC Plugin Issue](https://github.com/oxc-project/oxc/issues/2409#issue-2133367176)
2. [Speeding Up JavaScript Ecosystem Part 11](https://marvinh.dev/blog/speeding-up-javascript-ecosystem-part-11/)
3. [Oxlint JS Plugins - Raw Transfer](https://oxc.rs/blog/2025-10-09-oxlint-js-plugins.html#raw-transfer)
4. [less.js Plugin Documentation](https://lesscss.org/features/#plugin-atrules-feature)
5. [goja - Go JavaScript Runtime](https://github.com/dop251/goja)
6. [v8go - V8 Bindings for Go](https://github.com/rogchap/v8go)

## Appendix A: LESS Plugin API Reference

### Plugin Structure

```javascript
module.exports = {
    install(less, pluginManager, functionRegistry) {
        // Called once when plugin is first loaded
    },

    use(plugin) {
        // Called every time @plugin directive is encountered
    },

    setOptions(options) {
        // Called when plugin has options: @plugin (opt1, opt2) "path"
    },

    minVersion: [3, 0, 0]  // Optional minimum Less version
};
```

### Function Registration

```javascript
install(less, pluginManager, functionRegistry) {
    // Single function
    functionRegistry.add('myFunc', function(arg1, arg2) {
        return less.dimension(42, 'px');
    });

    // Multiple functions
    functionRegistry.addMultiple({
        'func1': function() { return less.keyword('value'); },
        'func2': function(arg) { return arg; }
    });
}
```

### Visitor Registration

```javascript
install(less, pluginManager, functionRegistry) {
    class MyVisitor {
        constructor() {
            this.isPreEvalVisitor = true;  // or false for post-eval
            this.isReplacing = true;       // can replace nodes
        }

        run(root) {
            return this.visit(root);
        }

        visitVariable(node) {
            // Transform or replace variable node
            return node;
        }

        visitRuleset(node) {
            // Transform or replace ruleset node
            return node;
        }
    }

    pluginManager.addVisitor(new MyVisitor());
}
```

### Pre/Post Processor Registration

```javascript
install(less, pluginManager, functionRegistry) {
    pluginManager.addPreProcessor({
        process(src, extra) {
            // Transform source before parsing
            return src.replace(/foo/g, 'bar');
        }
    }, priority);  // 1 = before import, 1000 = import, 2000 = after

    pluginManager.addPostProcessor({
        process(css, extra) {
            // Transform CSS after compilation
            return css;
        }
    }, priority);
}
```

### File Manager Registration

```javascript
install(less, pluginManager, functionRegistry) {
    pluginManager.addFileManager({
        supports(filename, currentDirectory, options, environment) {
            // Return true if this manager can handle the file
            return filename.startsWith('http://');
        },

        loadFile(filename, currentDirectory, options, environment) {
            // Return Promise or callback with {contents, filename}
            return fetch(filename).then(r => r.text()).then(contents => ({
                contents,
                filename
            }));
        }
    });
}
```

### Available Node Constructors

```javascript
// In `less` object passed to install()
less.assignment(key, val)
less.attribute(key, op, value)
less.call(name, args)
less.color(rgb, alpha)
less.combinator(value)
less.condition(op, lvalue, rvalue)
less.detachedruleset(ruleset)
less.dimension(value, unit)
less.element(combinator, value)
less.expression(value)
less.keyword(value)
less.operation(op, operands)
less.paren(node)
less.quoted(quote, value, escaped)
less.ruleset(selectors, rules)
less.selector(elements)
less.url(value, paths)
less.value(value)
less.atrule(name, value)
```

## Appendix B: Example Plugins

### Simple Function Plugin

```javascript
// plugin-simple.js
functions.add('pi', function() {
    return less.dimension(Math.PI);
});
```

### Pre-eval Visitor Plugin

```javascript
// plugin-preeval.js
module.exports = {
    install({ tree: { Quoted }, visitors }, manager) {
        class Visitor {
            constructor() {
                this.isPreEvalVisitor = true;
                this.isReplacing = true;
            }

            run(root) {
                return this.visit(root);
            }

            visitVariable(node) {
                if (node.name === '@replace') {
                    return new Quoted("'", 'bar', true);
                }
                return node;
            }
        }

        manager.addVisitor(new Visitor());
    },
    minVersion: [2, 0, 0]
};
```

### Node Constructor Plugin

```javascript
// plugin-tree-nodes.js
functions.addMultiple({
    'test-dimension': function() {
        return less.dimension(1, 'px');
    },
    'test-color': function() {
        return less.color([50, 50, 50]);
    },
    'test-quoted': function() {
        return less.quoted('"', 'foo');
    },
    'test-detached-ruleset': function() {
        var decl = less.declaration('prop', 'value');
        return less.detachedruleset(less.ruleset('', [decl]));
    }
});
```
