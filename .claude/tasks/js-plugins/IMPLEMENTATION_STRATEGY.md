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

## Proposed Architecture for less.go

### Phase 1: JavaScript Runtime Integration

**Goal**: Choose and integrate a JavaScript runtime for Go.

**Options Analysis**:

| Runtime | Pros | Cons | Recommendation |
|---------|------|------|----------------|
| **goja** | Pure Go, no CGO, easy integration | ~10-20x slower than V8 | ✅ **Start here** (simplicity) |
| **v8go** | Near-native JS performance | Requires CGO, complex build | Phase 2 optimization |
| **quickjs-go** | Smaller footprint than V8 | Still requires CGO | Alternative to v8go |

**Decision**: Start with `goja` for rapid prototyping, measure performance, then optionally migrate to `v8go` if needed.

**Tasks**:
1. Add goja dependency
2. Create `packages/less/src/less/less_go/runtime/` package
3. Implement basic JavaScript execution wrapper
4. Unit tests for runtime initialization and script execution

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

**Goal**: Create JavaScript facade objects that read from buffers on-demand.

**Design**:

```javascript
// JavaScript side - generated/written in Go, executed in runtime
class NodeFacade {
  constructor(buffer, index) {
    this._buffer = buffer;
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

// Visitor pattern support
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
    return node; // Return unchanged
  }
}
```

**Tasks**:
1. Generate JavaScript binding code from Go AST node definitions
2. Implement visitor pattern support
3. Add node construction helpers (for plugin-created nodes)
4. Write JavaScript unit tests for bindings

### Phase 4: Plugin Loader

**Goal**: Parse `@plugin` directives and load JavaScript files.

**Design**:

```go
type PluginLoader struct {
    runtime       *runtime.JSRuntime
    loadedPlugins map[string]*Plugin
    fileManager   *FileManager
}

type Plugin struct {
    filename      string
    source        string
    exports       map[string]interface{}
    functions     map[string]JSFunction
    visitors      []JSVisitor
    fileManagers  []JSFileManager
}

func (pl *PluginLoader) LoadPlugin(path string, options map[string]interface{}) (*Plugin, error) {
    // 1. Resolve path (handle relative, absolute, npm modules)
    // 2. Load JS file content
    // 3. Execute in sandboxed context
    // 4. Extract exports (install, use, setOptions, minVersion)
    // 5. Cache plugin by filename
}
```

**Plugin Execution Context**:
```javascript
// Injected into plugin execution scope
const context = {
  module: { exports: {} },
  require: createRequire(pluginPath),
  registerPlugin: (obj) => { /* capture */ },
  functions: functionRegistry,
  tree: treeConstructors,
  less: lessAPI,
  fileInfo: currentFileInfo
};

// Execute plugin code
(function() {
  // Plugin source code here
}).call(context);
```

**Tasks**:
1. Implement plugin file resolution (local, relative, npm)
2. Create sandboxed execution context
3. Implement `require()` emulation for dependencies
4. Parse `@plugin` directive options
5. Cache loaded plugins

### Phase 5: Function Registry Integration

**Goal**: Enable plugins to register custom LESS functions.

**Design**:

```go
type JSFunction struct {
    name     string
    jsFunc   goja.Callable
    runtime  *runtime.JSRuntime
}

func (jf *JSFunction) Call(args []Node, ctx *EvalContext) (Node, error) {
    // 1. Serialize args to flat buffer
    // 2. Call JavaScript function with buffer
    // 3. Deserialize return value
    // 4. Convert to Go Node
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

**Tasks**:
1. Extend FunctionRegistry to support JS functions
2. Implement bidirectional argument serialization
3. Handle return value conversion
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

**Agent 1: Runtime & Serialization**
- Phase 1: JavaScript runtime integration
- Phase 2: AST serialization
- Deliverable: Working buffer-based AST transfer

**Agent 2: Plugin Loader & Module Resolution**
- Phase 4: Plugin loader
- NPM module resolution
- Deliverable: `@plugin` directive loading

**Agent 3: Bindings & Constructors**
- Phase 3: JavaScript bindings
- Phase 7: Tree node constructors
- Deliverable: Complete node API for plugins

**Agent 4: Functions & Registry**
- Phase 5: Function registry integration
- Deliverable: Plugins can add custom functions

**Agent 5: Visitors & Evaluation**
- Phase 6: Visitor integration
- Deliverable: Pre-eval and post-eval transformations

**Agent 6: Processors & File Managers**
- Phase 8: Pre/post processors
- Phase 9: File manager support
- Deliverable: Text transformation and custom imports

**Agent 7: Scoping & Integration**
- Phase 10: Plugin scope management
- Integration testing
- Deliverable: Complete plugin system with proper scoping

## Performance Considerations

### Expected Overhead

Based on OXC's results:

- **Traditional approach (JSON)**: 10-20x slower than native
- **Raw transfer approach**: 2-5x slower than native
- **With lazy deserialization**: 1.5-3x slower than native

### Optimization Opportunities

1. **Buffer Reuse**: Pool buffers to reduce allocations
2. **Incremental Updates**: Only serialize changed subtrees
3. **V8 Migration**: Switch to v8go for ~10x JavaScript performance boost
4. **JIT Warm-up**: Keep runtime alive between compilations
5. **Partial Visitor**: Only flatten visited node paths

### When to Optimize

1. **First**: Get it working with goja
2. **Second**: Measure real-world usage (bootstrap4 test)
3. **Third**: Optimize hot paths if needed
4. **Last**: Consider v8go migration if 2-3x overhead is unacceptable

## Risks & Mitigations

### Risk 1: Performance Unacceptable

**Mitigation**:
- Benchmark early and often
- Have v8go migration path ready
- Consider native Go reimplementations of popular plugins

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
