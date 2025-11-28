# JavaScript Plugin Implementation Strategy for less.go

## Executive Summary

This document outlines a strategy for implementing JavaScript plugin support in less.go, inspired by the OXC project's approach to running JavaScript plugins from native code. The goal is to enable the quarantined tests (plugin, plugin-module, plugin-preeval, javascript, bootstrap4) to pass.

## Background: The OXC Approach

The OXC project (Rust-based linter) faced a similar challenge: running ESLint-compatible JavaScript plugins from a native codebase. Their key innovations:

1. **Raw Transfer**: Instead of JSON serialization, they pass memory blocks directly between Rust and JavaScript
2. **Lazy Deserialization**: Only deserialize AST nodes that plugins actually access
3. **Dual API**: ESLint-compatible API for migration + optimized API for performance

**Key difference for less.go**: OXC only needs to *read* the AST for linting. Less plugins need to *create and modify* tree nodes. This requires bidirectional data flow.

## Requirements Analysis

### Plugin Capabilities Required

Based on the test fixtures and bootstrap4 requirements:

| Capability | Tests Using It | Complexity |
|-----------|---------------|------------|
| Custom functions | plugin, plugin-simple, bootstrap4 | Medium |
| Pre-eval visitors | plugin-preeval | High |
| Post-eval visitors | plugin | High |
| Tree node creation | plugin-tree-nodes | Medium |
| Pre-processors | plugin | Low |
| Post-processors | plugin | Low |
| File managers | - | Low |
| Inline JS (`backticks`) | javascript, no-js-errors | Medium |

### Quarantined Tests to Enable

1. **`plugin`** - Core plugin functionality (function registry, visitors, tree nodes)
2. **`plugin-module`** - Loading plugins from npm packages
3. **`plugin-preeval`** - Pre-evaluation visitor that transforms AST before eval
4. **`javascript`** - Inline JavaScript evaluation with backticks
5. **`js-type-errors/*`** - Error handling for JS evaluation
6. **`no-js-errors/*`** - Valid JS expression evaluation
7. **`bootstrap4`** - Real-world test requiring 11 plugin functions

## Proposed Architecture

### Technology Choice: Goja (Pure Go JavaScript Runtime)

**Recommendation**: Use [goja](https://github.com/dop251/goja) - a pure Go ECMAScript 5.1+ implementation.

**Why Goja?**
- Pure Go - no CGO dependency, easy cross-compilation
- Battle-tested (used by k6, CockroachDB, Grafana)
- Good ES6+ support (classes, arrow functions, spread operator)
- Reasonable performance for plugin workloads
- Native Go object bridging

**Alternatives Considered:**
- **V8Go**: Faster but requires CGO + V8 library
- **QuickJS-Go**: Faster but requires CGO
- **Wazero (WASM)**: Would require compiling plugins to WASM first

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        less.go Core                              │
├─────────────────────────────────────────────────────────────────┤
│  Parser  │  Evaluator  │  Visitor  │  CSS Output Generator      │
└────┬─────┴──────┬──────┴─────┬─────┴─────────────────────────────┘
     │            │            │
     ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Plugin Bridge Layer                          │
├─────────────────────────────────────────────────────────────────┤
│  GoJSBridge  │  FunctionProxy  │  TreeNodeProxy  │  VisitorProxy │
└──────┬───────┴────────┬────────┴────────┬───────┴───────────────┘
       │                │                 │
       ▼                ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Goja JavaScript Runtime                       │
├─────────────────────────────────────────────────────────────────┤
│  Plugin Execution  │  Function Registry  │  less.* API           │
└─────────────────────────────────────────────────────────────────┘
```

### Component Design

#### 1. JS Runtime Manager (`js_runtime.go`)

```go
// JSRuntime manages the Goja runtime and plugin execution
type JSRuntime struct {
    vm           *goja.Runtime
    less         *LessAPI        // Exposed as 'less' global
    tree         *TreeAPI        // Exposed as 'tree' global
    functions    *FunctionProxy  // Exposed as 'functions' global
    nodeFactory  *NodeFactory    // Creates Go tree nodes from JS
}

func NewJSRuntime() *JSRuntime
func (r *JSRuntime) Execute(code string) (goja.Value, error)
func (r *JSRuntime) RegisterPlugin(plugin *goja.Object) error
func (r *JSRuntime) CallFunction(name string, args ...any) (any, error)
```

#### 2. Tree Node Proxy (`tree_proxy.go`)

The critical component - allows JS to create and manipulate Go tree nodes.

```go
// TreeAPI provides the less.* constructors to JavaScript
type TreeAPI struct {
    runtime *goja.Runtime
}

// Each method returns a wrapped Go tree node
func (t *TreeAPI) Dimension(value, unit any) *goja.Object
func (t *TreeAPI) Color(rgb any) *goja.Object
func (t *TreeAPI) Keyword(value string) *goja.Object
func (t *TreeAPI) Quoted(quote, value string) *goja.Object
// ... all 20+ tree node types
```

**Lazy Binding Strategy** (inspired by OXC):
- Don't convert entire AST to JS objects upfront
- Wrap Go nodes in JS proxy objects
- Only materialize properties when accessed
- Cache materialized properties

#### 3. Function Registry Bridge (`function_registry_bridge.go`)

```go
// FunctionProxy allows plugins to register custom LESS functions
type FunctionProxy struct {
    registry *FunctionRegistry
    runtime  *goja.Runtime
}

func (f *FunctionProxy) Add(name string, fn goja.Callable)
func (f *FunctionProxy) AddMultiple(fns *goja.Object)
```

When a plugin-registered function is called during LESS evaluation:
1. Convert Go arguments to Goja values
2. Call the JavaScript function
3. Convert the return value back to a Go tree node

#### 4. Visitor Bridge (`visitor_bridge.go`)

```go
// JSVisitor wraps a JavaScript visitor for use in Go's visitor system
type JSVisitor struct {
    jsVisitor      goja.Value
    runtime        *goja.Runtime
    isPreEval      bool
    isReplacing    bool
}

func (v *JSVisitor) Visit(node tree.Node) tree.Node
func (v *JSVisitor) Run(root tree.Node) tree.Node
```

#### 5. Plugin Loader (`plugin_loader_js.go`)

```go
// JSPluginLoader loads and executes JavaScript plugins
type JSPluginLoader struct {
    runtime *JSRuntime
    less    *Less
}

func (l *JSPluginLoader) LoadPlugin(path string) (*Plugin, error)
func (l *JSPluginLoader) EvalPlugin(contents string, context *ParseContext) (*Plugin, error)
```

## Implementation Plan

### Phase 1: Core Infrastructure (Foundation)

**Goal**: Get basic JavaScript execution working with function registration.

**Deliverables**:
1. Add goja dependency
2. Implement `JSRuntime` with basic execution
3. Implement `FunctionProxy` for `functions.add()` and `functions.addMultiple()`
4. Wire up plugin-registered functions to LESS evaluator
5. **Test**: `plugin-simple.js` functions work

**Estimated complexity**: Medium

### Phase 2: Tree Node Creation

**Goal**: Enable plugins to create tree nodes via `less.*` constructors.

**Deliverables**:
1. Implement `TreeAPI` with all node constructors
2. Implement `NodeFactory` for JS→Go node conversion
3. Implement lazy property access on node proxies
4. **Test**: `plugin-tree-nodes.js` functions work

**Node types to implement**:
- `less.dimension(value, unit)`
- `less.color(rgb)`
- `less.keyword(value)`
- `less.quoted(quote, value)`
- `less.url(url)`
- `less.value(expressions)`
- `less.expression(values)`
- `less.declaration(property, value)`
- `less.assignment(key, value)`
- `less.call(name)`
- `less.condition(op, left, right)`
- `less.attribute(name, op, value)`
- `less.element(combinator, value)`
- `less.selector(selector)`
- `less.ruleset(selector, rules)`
- `less.detachedruleset(ruleset)`
- `less.atrule(name, value)`
- `less.combinator(value)`
- `less.operation(op, values)`

**Estimated complexity**: High (many node types)

### Phase 3: Visitor System

**Goal**: Enable pre-eval and post-eval visitors from JavaScript.

**Deliverables**:
1. Implement `JSVisitor` wrapper
2. Bridge `visitors.Visitor` class to JavaScript
3. Implement visitor method dispatch (visitDeclaration, visitVariable, etc.)
4. Wire visitors into transform-tree pipeline
5. **Test**: `plugin-preeval.js` works

**Estimated complexity**: High

### Phase 4: Plugin Loading & @plugin Directive

**Goal**: Load plugins from files and npm packages.

**Deliverables**:
1. Implement `JSPluginLoader`
2. Parse `@plugin` directive (already exists in parser)
3. Implement plugin caching
4. Handle `module.exports` and `registerPlugin()` patterns
5. Support plugin options and version checking
6. **Test**: `plugin.less` and `plugin-module.less` work

**Estimated complexity**: Medium

### Phase 5: Inline JavaScript Evaluation

**Goal**: Support backtick JavaScript expressions in LESS.

**Deliverables**:
1. Implement JavaScript tree node evaluation
2. Bind LESS variables to JS context
3. Convert JS results back to LESS values
4. Handle JS errors gracefully
5. **Test**: `javascript` and `no-js-errors` suites pass

**Estimated complexity**: Medium

### Phase 6: Bootstrap4 & Real-World Testing

**Goal**: Get bootstrap4 test passing.

**Deliverables**:
1. Ensure all 11 bootstrap-less-port plugin functions work
2. Test map/ruleset access from plugins
3. Performance optimization if needed
4. **Test**: `bootstrap4.less` compiles correctly

**Estimated complexity**: Medium (mostly integration)

### Phase 7: Pre/Post Processors

**Goal**: Complete plugin API coverage.

**Deliverables**:
1. Implement pre-processor bridge
2. Implement post-processor bridge
3. Wire into parse pipeline
4. **Test**: Full plugin test suite passes

**Estimated complexity**: Low

## Performance Considerations

### Potential Bottlenecks

1. **JS↔Go boundary crossing**: Each function call and property access crosses the boundary
2. **Object creation**: Creating JS wrapper objects for tree nodes
3. **String conversion**: UTF-8 encoding/decoding

### Optimization Strategies (OXC-Inspired)

1. **Lazy Binding**: Don't create JS objects until needed
   ```go
   // Instead of eagerly converting:
   jsNode := convertToJS(goNode)  // ❌ Expensive

   // Use a proxy that defers conversion:
   jsNode := wrapAsProxy(goNode)  // ✅ Cheap
   ```

2. **Property Caching**: Cache converted properties
   ```go
   type NodeProxy struct {
       goNode    tree.Node
       propCache map[string]goja.Value
   }
   ```

3. **Batch String Handling**: Pre-decode all strings once (OXC's approach)

4. **Pool JS Runtimes**: Reuse Goja VMs across compilations

### Expected Performance

Based on OXC's benchmarks:
- With plugins: ~5x slower than without (acceptable)
- Still faster than running full Node.js less.js

## Testing Strategy

### Unit Tests

For each component:
```go
func TestJSRuntime_Execute(t *testing.T)
func TestFunctionProxy_Add(t *testing.T)
func TestTreeAPI_Dimension(t *testing.T)
func TestJSVisitor_Visit(t *testing.T)
```

### Integration Tests

Use existing test fixtures:
- `packages/test-data/plugin/*.js` - Plugin implementations
- `packages/test-data/less/_main/plugin*.less` - LESS files using plugins

### Compatibility Tests

Verify output matches less.js:
```bash
# Compare Go output with JS output for plugin tests
LESS_GO_DIFF=1 go test -run TestIntegrationSuite/main/plugin
```

## File Structure

```
packages/less/src/less/less_go/
├── js/                          # New directory for JS integration
│   ├── runtime.go               # JSRuntime - core Goja wrapper
│   ├── runtime_test.go
│   ├── tree_api.go              # TreeAPI - less.* constructors
│   ├── tree_api_test.go
│   ├── node_proxy.go            # Lazy node wrapper
│   ├── node_proxy_test.go
│   ├── function_proxy.go        # functions.add/addMultiple
│   ├── function_proxy_test.go
│   ├── visitor_bridge.go        # JS visitor wrapper
│   ├── visitor_bridge_test.go
│   └── plugin_loader.go         # Plugin file loading
├── plugin_manager.go            # Existing - needs updates
└── ...
```

## Dependencies

Add to `go.mod`:
```
require github.com/dop251/goja v0.0.0-20240220182346-e401ed450204
```

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Goja doesn't support needed ES6 features | Low | High | Test early with real plugins |
| Performance too slow | Medium | Medium | Implement lazy binding, optimize hot paths |
| Complex visitor patterns fail | Medium | Medium | Test incrementally, match JS behavior exactly |
| Memory leaks from JS↔Go cycles | Medium | Low | Use weak refs, clear caches |

## Success Criteria

1. **All quarantined tests pass**: plugin, plugin-module, plugin-preeval, javascript
2. **Bootstrap4 compiles correctly**: Output matches less.js
3. **No regressions**: Existing 94 perfect matches maintained
4. **Acceptable performance**: <10x slowdown with plugins enabled
5. **Clean API**: Plugins written for less.js work without modification

## Work Distribution for Parallel Agents

The phases are designed for parallel work:

| Agent | Phase | Dependencies |
|-------|-------|--------------|
| Agent 1 | Phase 1: Core Infrastructure | None |
| Agent 2 | Phase 2: Tree Node Creation | Phase 1 |
| Agent 3 | Phase 3: Visitor System | Phase 1, 2 |
| Agent 4 | Phase 4: Plugin Loading | Phase 1 |
| Agent 5 | Phase 5: Inline JavaScript | Phase 1 |
| Agent 6 | Phase 6-7: Integration | All above |

**Suggested parallel groupings**:
- **Wave 1**: Agents 1 (complete Phase 1 first as foundation)
- **Wave 2**: Agents 2, 4, 5 (can work in parallel once Phase 1 done)
- **Wave 3**: Agent 3 (needs Phase 2)
- **Wave 4**: Agent 6 (integration and polish)

## Appendix A: Bootstrap4 Required Functions

| Function | Purpose | Implementation Notes |
|----------|---------|---------------------|
| `breakpoint-next` | Next breakpoint in grid | Map iteration |
| `breakpoint-min` | Min width for breakpoint | Map lookup |
| `breakpoint-max` | Max width for breakpoint | Map lookup + calc |
| `breakpoint-infix` | Responsive class suffix | String manipulation |
| `map-keys` | Extract keys from ruleset | Ruleset introspection |
| `color` | Get color from @colors map | Variable lookup |
| `theme-color` | Get theme color | Variable lookup |
| `theme-color-level` | Color at contrast level | Color math |
| `color-yiq` | Calculate YIQ contrast | Color algorithm |
| `gray` | Get grayscale color | Variable lookup |
| `escape-svg` | Escape SVG for CSS | String escaping |

## Appendix B: Key less.js Files to Reference

- `environment/abstract-plugin-loader.js` - Plugin evaluation logic
- `plugin-manager.js` - Plugin registration and lifecycle
- `tree/javascript.js` - Backtick JS evaluation
- `functions/function-registry.js` - Function registration
- `visitors/visitor.js` - Visitor pattern implementation
