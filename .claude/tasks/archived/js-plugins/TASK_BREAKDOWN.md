# JavaScript Plugin Implementation - Task Breakdown

## Overview

This document breaks down the JavaScript plugin implementation into discrete, parallelizable tasks that can be assigned to multiple agents.

## Task Status Legend

- 🔴 Not Started
- 🟡 In Progress
- 🟢 Complete
- ⏸️ Blocked

---

## Phase 1: JavaScript Runtime Integration 🔴

**Assigned to**: Available
**Dependencies**: None
**Estimated effort**: 2-3 days

### Tasks

#### 1.1: Add goja dependency 🔴
- Add `github.com/dop251/goja` to go.mod
- Verify build works
- Check for any compatibility issues

**Acceptance Criteria**:
- `go mod tidy` succeeds
- `pnpm -w test:go:unit` passes (no regressions)

#### 1.2: Create runtime package structure 🔴
- Create `packages/less/src/less/less_go/runtime/` directory
- Create `runtime.go` with basic JSRuntime struct
- Create `runtime_test.go` for unit tests

**Files to create**:
- `runtime/runtime.go`
- `runtime/runtime_test.go`
- `runtime/doc.go`

**Acceptance Criteria**:
- Package compiles
- Basic struct defined

#### 1.3: Implement basic JavaScript execution 🔴
- Implement `NewRuntime()` constructor
- Implement `Execute(script string) (interface{}, error)`
- Implement `Call(funcName string, args ...interface{}) (interface{}, error)`
- Add error handling and stack trace extraction

**Acceptance Criteria**:
- Can execute simple JS: `1 + 1`
- Can call JS functions
- Errors include JS stack traces
- Unit tests cover success and error cases

#### 1.4: Implement context injection 🔴
- Implement `Set(name string, value interface{}) error`
- Implement `Get(name string) (interface{}, error)`
- Support Go → JS type conversion (string, int, bool, map, slice)

**Acceptance Criteria**:
- Can inject Go values into JS context
- Can read JS values from Go
- Unit tests for each type conversion

---

## Phase 2: AST Serialization 🔴

**Assigned to**: Available
**Dependencies**: Phase 1
**Estimated effort**: 4-5 days

### Tasks

#### 2.1: Design flat buffer format 🔴
- Define `FlatNode` struct
- Define `FlatAST` container struct
- Document buffer layout specification

**Files to create**:
- `runtime/serialization.go`
- `runtime/serialization_test.go`
- `runtime/BUFFER_FORMAT.md`

**Acceptance Criteria**:
- Data structures compile
- Format documentation complete
- Design reviewed for efficiency

#### 2.2: Implement AST flattening 🔴
- Implement `FlattenAST(root Node) (*FlatAST, error)`
- Walk AST depth-first
- Assign unique indices to each node
- Build type and string tables
- Handle all node types from `tree/` package

**Acceptance Criteria**:
- Can flatten simple AST (single ruleset)
- Can flatten complex AST (imports, mixins, variables)
- No duplicate strings in string table
- Unit tests for each node type

#### 2.3: Implement AST unflattening 🔴
- Implement `UnflattenAST(flat *FlatAST) (Node, error)`
- Reconstruct tree from flat buffer
- Restore parent/child relationships
- Handle circular references (if any)

**Acceptance Criteria**:
- Can unflatten simple AST
- Can unflatten complex AST
- Roundtrip preserves tree structure
- Unit tests verify correctness

#### 2.4: Add buffer serialization 🔴
- Implement `SerializeToBytes(*FlatAST) ([]byte, error)`
- Implement `DeserializeFromBytes([]byte) (*FlatAST, error)`
- Add magic number and version header
- Implement checksum validation

**Acceptance Criteria**:
- Binary format is stable
- Can serialize/deserialize to bytes
- Checksum detects corruption
- Unit tests verify roundtrip

#### 2.5: Performance optimization 🔴
- Add buffer pooling
- Optimize string deduplication
- Benchmark serialization performance
- Compare against JSON baseline

**Acceptance Criteria**:
- < 10ms to flatten typical AST (1000 nodes)
- < 5ms to unflatten typical AST
- < 50% size of equivalent JSON
- Benchmark results documented

---

## Phase 3: JavaScript Bindings 🔴

**Assigned to**: Available
**Dependencies**: Phase 2
**Estimated effort**: 3-4 days

### Tasks

#### 3.1: Generate node facade code 🔴
- Create JavaScript facade generator
- Generate `NodeFacade` class for each node type
- Implement lazy property access
- Handle child/sibling navigation

**Files to create**:
- `runtime/codegen/facade_generator.go`
- `runtime/bindings/node_facade.js` (generated)

**Acceptance Criteria**:
- Generated JS code is valid
- Can access node properties
- Can navigate tree structure
- Unit tests (JS side)

#### 3.2: Implement visitor pattern support 🔴
- Generate `Visitor` base class
- Generate `visit*` method stubs for all node types
- Implement tree traversal
- Support node replacement

**Files to create**:
- `runtime/bindings/visitor.js` (generated)

**Acceptance Criteria**:
- Can create custom visitor
- Can visit all node types
- Can replace nodes during traversal
- Unit tests for visitor pattern

#### 3.3: Implement buffer transfer to JS 🔴
- Transfer `FlatAST` buffer to goja runtime
- Expose buffer as JS array/view
- Implement efficient access patterns

**Acceptance Criteria**:
- JS can read buffer data
- No unnecessary copies
- Performance benchmarks

---

## Phase 4: Plugin Loader 🔴

**Assigned to**: Available
**Dependencies**: Phase 1
**Estimated effort**: 4-5 days

### Tasks

#### 4.1: Parse @plugin directive 🔴
- Check if `@plugin` is already parsed (likely is)
- If not, add to parser
- Extract plugin path and options
- Create `PluginDirective` node type (if needed)

**Files to check/modify**:
- `packages/less/src/less/less_go/parser/` (if needed)
- `packages/less/src/less/less_go/tree/plugin.go` (create if needed)

**Acceptance Criteria**:
- Can parse `@plugin "path"`
- Can parse `@plugin (option) "path"`
- AST node created correctly
- Unit tests for parsing

#### 4.2: Implement plugin file resolution 🔴
- Create `PluginLoader` struct
- Implement local file resolution (./path)
- Implement relative file resolution (../../path)
- Implement absolute file resolution (/path)

**Files to create**:
- `runtime/plugin_loader.go`
- `runtime/plugin_loader_test.go`

**Acceptance Criteria**:
- Can resolve local plugins
- Can resolve relative plugins
- Path traversal is safe
- Unit tests for each case

#### 4.3: Implement NPM module resolution 🔴
- Implement node_modules traversal
- Support scoped packages (@org/package)
- Support package.json main field
- Handle less-plugin-* prefix convention

**Acceptance Criteria**:
- Can load `@plugin "clean-css"`
- Can load `@plugin "less-plugin-foo"`
- Works with workspace packages
- Unit tests with mock node_modules

#### 4.4: Implement plugin execution context 🔴
- Create sandboxed execution environment
- Inject `module`, `exports`, `require`
- Inject `functions`, `tree`, `less`, `fileInfo`
- Execute plugin code

**Acceptance Criteria**:
- Plugin code executes in isolation
- Can use module.exports
- Can use require()
- Errors are captured properly

#### 4.5: Implement plugin caching 🔴
- Cache loaded plugins by filename
- Implement `install()` once, `use()` multiple times
- Handle plugin options via `setOptions()`

**Acceptance Criteria**:
- Plugin loaded only once per file
- `install()` called once
- `use()` called per @plugin directive
- Unit tests verify caching

---

## Phase 5: Function Registry Integration 🔴

**Assigned to**: Available
**Dependencies**: Phase 3, Phase 4
**Estimated effort**: 2-3 days

### Tasks

#### 5.1: Create JSFunction wrapper 🔴
- Implement `JSFunction` struct
- Wrap goja.Callable
- Implement `Call(args []Node, ctx *EvalContext) (Node, error)`

**Files to create**:
- `runtime/js_function.go`
- `runtime/js_function_test.go`

**Acceptance Criteria**:
- Can wrap JS function
- Can call from Go
- Errors propagate correctly

#### 5.2: Implement argument serialization 🔴
- Convert Go `Node` arguments to flat buffer
- Pass buffer to JavaScript
- Convert JS arguments to `NodeFacade` objects

**Acceptance Criteria**:
- All node types can be passed as args
- Minimal serialization overhead
- Unit tests for each node type

#### 5.3: Implement return value deserialization 🔴
- Convert JS return value to flat buffer
- Convert buffer back to Go `Node`
- Handle primitive returns (number, string, bool)

**Acceptance Criteria**:
- All node types can be returned
- Primitives handled correctly
- Errors for invalid returns
- Unit tests for each type

#### 5.4: Extend FunctionRegistry 🔴
- Modify `FunctionRegistry` to support `JSFunction`
- Add `AddJSFunction(name string, fn *JSFunction)`
- Modify function lookup to check JS functions
- Maintain priority (builtins vs. plugins)

**Files to modify**:
- `functions/function_registry.go`

**Acceptance Criteria**:
- Can register JS functions
- Can call JS functions from LESS
- Plugin functions can shadow builtins (if scoped)
- Unit tests verify integration

#### 5.5: Implement function injection for plugins 🔴
- Create `functions` object for plugins
- Implement `functions.add(name, func)`
- Implement `functions.addMultiple(obj)`

**Acceptance Criteria**:
- Plugins can register functions
- Functions are added to registry
- Unit tests with sample plugin

---

## Phase 6: Visitor Integration 🔴

**Assigned to**: Available
**Dependencies**: Phase 3, Phase 4
**Estimated effort**: 3-4 days

### Tasks

#### 6.1: Create JSVisitor wrapper 🔴
- Implement `JSVisitor` struct
- Wrap JS visitor object
- Implement `Visit(node Node) (Node, error)`
- Support `isPreEvalVisitor` and `isReplacing` flags

**Files to create**:
- `runtime/js_visitor.go`
- `runtime/js_visitor_test.go`

**Acceptance Criteria**:
- Can wrap JS visitor
- Can call `run(root)` method
- Flags are respected

#### 6.2: Implement visitor tree traversal 🔴
- Flatten AST subtree starting from visited node
- Pass buffer to JS visitor
- Call `visitor.run(root)`
- Unflatten result

**Acceptance Criteria**:
- Entire subtree is accessible to visitor
- Visitor can navigate tree
- Performance is acceptable
- Unit tests verify correctness

#### 6.3: Add visitor registration to PluginManager 🔴
- Create/modify `PluginManager` struct
- Add `AddVisitor(visitor *JSVisitor)`
- Separate pre-eval and post-eval visitors
- Implement visitor iteration

**Files to create/modify**:
- `runtime/plugin_manager.go`

**Acceptance Criteria**:
- Can register visitors
- Can retrieve pre-eval visitors
- Can retrieve post-eval visitors
- Unit tests verify registration

#### 6.4: Integrate pre-eval visitors 🔴
- Find evaluation entry point
- Add pre-eval visitor loop
- Call each visitor in order
- Handle node replacement

**Files to modify**:
- `tree/*.go` (likely `ruleset.go` or main eval)

**Acceptance Criteria**:
- Pre-eval visitors run before evaluation
- Node replacements work
- Errors are propagated
- Integration test passes

#### 6.5: Integrate post-eval visitors 🔴
- Add post-eval visitor loop after evaluation
- Call each visitor in order
- Handle node replacement

**Acceptance Criteria**:
- Post-eval visitors run after evaluation
- Node replacements work
- Errors are propagated
- Integration test passes

---

## Phase 7: Tree Node Constructors 🔴

**Assigned to**: Available
**Dependencies**: Phase 3
**Estimated effort**: 2-3 days

### Tasks

#### 7.1: Generate node constructor functions 🔴
- For each node type in `tree/`, generate JS constructor
- Example: `less.dimension(value, unit)`
- Return `NodeFacade` wrapping new node

**Files to create/modify**:
- `runtime/codegen/constructor_generator.go`
- `runtime/bindings/constructors.js` (generated)

**Acceptance Criteria**:
- All node types have constructors
- Constructors return valid facades
- Unit tests (JS side)

#### 7.2: Implement node allocation in buffer 🔴
- Allocate new node in `FlatAST` buffer
- Assign unique index
- Set node properties
- Update parent/child relationships if needed

**Acceptance Criteria**:
- Can create nodes from JS
- Nodes are properly indexed
- Buffer remains valid
- Unit tests verify allocation

#### 7.3: Handle complex node properties 🔴
- Support array properties (e.g., `children`)
- Support nested node properties (e.g., `value` containing another node)
- Validate required vs. optional properties

**Acceptance Criteria**:
- Can create nodes with array properties
- Can create nodes with nested nodes
- Validation catches missing required props
- Unit tests for complex cases

#### 7.4: Inject constructors into plugin context 🔴
- Create `tree` object with all constructors
- Create `less` object with constructors and utilities
- Inject into plugin execution context

**Acceptance Criteria**:
- Plugins can access `less.dimension()`, etc.
- All constructors work
- Integration test with sample plugin

---

## Phase 8: Pre/Post Processors 🔴

**Assigned to**: Available
**Dependencies**: Phase 4
**Estimated effort**: 2-3 days

### Tasks

#### 8.1: Create JSPreProcessor wrapper 🔴
- Implement `JSPreProcessor` struct
- Wrap JS function
- Implement `Process(input string, options) (string, error)`

**Files to create**:
- `runtime/js_preprocessor.go`
- `runtime/js_preprocessor_test.go`

**Acceptance Criteria**:
- Can wrap JS processor
- Can call from Go
- Strings pass correctly
- Unit tests

#### 8.2: Create JSPostProcessor wrapper 🔴
- Implement `JSPostProcessor` struct
- Same as pre-processor but for CSS output

**Files to create**:
- `runtime/js_postprocessor.go`
- `runtime/js_postprocessor_test.go`

**Acceptance Criteria**:
- Same as 8.1
- Unit tests

#### 8.3: Implement priority-based ordering 🔴
- Store processors with priority values
- Sort by priority on retrieval
- Document priority conventions (1 = before import, etc.)

**Acceptance Criteria**:
- Processors run in priority order
- Unit tests verify ordering

#### 8.4: Integrate pre-processors 🔴
- Find parse entry point
- Add pre-processor loop before parsing
- Transform input string

**Files to modify**:
- `parse.go` or equivalent

**Acceptance Criteria**:
- Pre-processors run before parsing
- String transformations work
- Integration test

#### 8.5: Integrate post-processors 🔴
- Find CSS output generation
- Add post-processor loop after CSS generation
- Transform output string

**Acceptance Criteria**:
- Post-processors run after CSS output
- String transformations work
- Integration test

---

## Phase 9: File Manager Support 🔴

**Assigned to**: Available
**Dependencies**: Phase 4
**Estimated effort**: 2-3 days

### Tasks

#### 9.1: Create JSFileManager wrapper 🔴
- Implement `JSFileManager` struct
- Wrap JS file manager object
- Implement `Supports(filename, currentDir) bool`
- Implement `LoadFile(filename, currentDir) (*LoadedFile, error)`

**Files to create**:
- `runtime/js_filemanager.go`
- `runtime/js_filemanager_test.go`

**Acceptance Criteria**:
- Can wrap JS file manager
- Can check support
- Can load files
- Unit tests

#### 9.2: Add file manager registration 🔴
- Add to `PluginManager`
- Implement `AddFileManager(fm *JSFileManager)`

**Acceptance Criteria**:
- Can register file managers
- Can retrieve file managers
- Unit tests

#### 9.3: Integrate in import resolution 🔴
- Find import resolution code
- Try plugin file managers before default
- Fall back to default if no plugin supports

**Files to modify**:
- File manager or import resolution code

**Acceptance Criteria**:
- Plugin file managers are consulted
- Default file manager is fallback
- Integration test

#### 9.4: Handle async loading (if needed) 🔴
- Determine if async is needed
- If yes, implement Promise/callback handling
- If no, document why not

**Acceptance Criteria**:
- Async handling works or is unnecessary
- Documentation updated

---

## Phase 10: Plugin Scope Management 🔴

**Assigned to**: Available
**Dependencies**: All previous phases
**Estimated effort**: 3-4 days

### Tasks

#### 10.1: Implement PluginScope struct 🔴
- Create `PluginScope` with parent pointer
- Implement `Lookup(name) (*JSFunction, bool)`
- Implement `AddPlugin(plugin *Plugin)`
- Implement scope hierarchy traversal

**Files to create**:
- `runtime/plugin_scope.go`
- `runtime/plugin_scope_test.go`

**Acceptance Criteria**:
- Can create scope hierarchy
- Lookup traverses parents
- Unit tests verify behavior

#### 10.2: Add PluginScope to EvalContext 🔴
- Modify `EvalContext` to include `*PluginScope`
- Initialize with root scope
- Implement context cloning with child scope

**Files to modify**:
- `context.go` or equivalent

**Acceptance Criteria**:
- Context has plugin scope
- Cloning works correctly
- No regressions in existing tests

#### 10.3: Create child scopes in evaluation 🔴
- When entering ruleset, create child scope
- When entering mixin, create child scope
- When entering directive, create child scope

**Acceptance Criteria**:
- Child scopes created correctly
- Scopes are destroyed on exit
- Unit tests

#### 10.4: Handle @plugin directives in scopes 🔴
- Evaluate @plugin directives during eval
- Load plugins into current scope
- Functions added to scope are local

**Acceptance Criteria**:
- Local plugins only affect current scope
- Global plugins affect entire file
- Integration tests verify scoping

#### 10.5: Implement function shadowing 🔴
- Local functions shadow parent functions
- Test with same function name in different scopes

**Acceptance Criteria**:
- Shadowing works correctly
- Lookup returns innermost definition
- Integration test (plugin.less has this)

---

## Integration Testing Tasks 🔴

**Assigned to**: Available
**Dependencies**: All phases
**Estimated effort**: 2-3 days

### Tasks

#### IT.1: Enable plugin-simple test 🔴
- Remove quarantine from `plugin-simple`
- Run integration test
- Fix any issues
- Verify perfect CSS match

**Acceptance Criteria**:
- `plugin-simple.less` passes
- Output matches JavaScript version
- No regressions

#### IT.2: Enable plugin-tree-nodes test 🔴
- Remove quarantine from `plugin-tree-nodes`
- Run integration test
- Fix any issues
- Verify perfect CSS match

**Acceptance Criteria**:
- `plugin-tree-nodes.less` passes
- Output matches JavaScript version
- No regressions

#### IT.3: Enable plugin-preeval test 🔴
- Remove quarantine from `plugin-preeval`
- Run integration test
- Fix any issues
- Verify perfect CSS match

**Acceptance Criteria**:
- `plugin-preeval.less` passes
- Output matches JavaScript version
- No regressions

#### IT.4: Enable plugin test 🔴
- Remove quarantine from `plugin` (main comprehensive test)
- Run integration test
- Fix any issues
- Verify perfect CSS match

**Acceptance Criteria**:
- `plugin.less` passes (all scoping tests)
- Output matches JavaScript version
- No regressions

#### IT.5: Enable plugin-module test 🔴
- Remove quarantine from `plugin-module`
- Run integration test
- Fix any issues
- Verify perfect CSS match

**Acceptance Criteria**:
- `plugin-module.less` passes (NPM module loading)
- Output matches JavaScript version
- No regressions

#### IT.6: Enable bootstrap4 test (stretch) 🔴
- Remove quarantine from `bootstrap4`
- Run integration test
- Fix any issues
- Verify perfect CSS match

**Acceptance Criteria**:
- `bootstrap4` passes
- Real-world usage validated
- Performance acceptable

---

## Performance Optimization Tasks 🔴

**Assigned to**: Available
**Dependencies**: Integration testing
**Estimated effort**: 2-3 days

### Tasks

#### PO.1: Benchmark plugin overhead 🔴
- Create benchmark suite
- Measure native Go vs. plugin function call
- Measure visitor overhead
- Measure serialization overhead

**Acceptance Criteria**:
- Benchmarks run reliably
- Results documented
- Overhead quantified

#### PO.2: Optimize hot paths 🔴
- Profile plugin execution
- Identify bottlenecks
- Optimize critical paths
- Re-benchmark

**Acceptance Criteria**:
- > 20% performance improvement
- No correctness regressions

#### PO.3: Evaluate V8 migration (optional) 🔴
- If goja overhead > 5x, evaluate v8go
- Create proof-of-concept
- Benchmark comparison
- Document trade-offs

**Acceptance Criteria**:
- Decision documented
- If migrating, migration plan created
- If not, rationale documented

---

## Documentation Tasks 🔴

**Assigned to**: Available
**Dependencies**: All phases
**Estimated effort**: 1-2 days

### Tasks

#### DOC.1: Write developer guide 🔴
- Document plugin system architecture
- Explain buffer format
- Explain serialization process
- Add diagrams

**Acceptance Criteria**:
- Guide is comprehensive
- Examples included
- Reviewed by team

#### DOC.2: Write plugin author guide 🔴
- Document plugin API
- Provide examples
- Explain scoping rules
- Migration guide from less.js plugins

**Acceptance Criteria**:
- Guide is clear
- Examples work
- Covers all capabilities

#### DOC.3: Add inline code documentation 🔴
- Document all public APIs
- Add package documentation
- Add examples in godoc

**Acceptance Criteria**:
- `go doc` output is helpful
- All exports documented

---

## Summary

**Total Phases**: 10
**Total Tasks**: 60+
**Estimated Duration**: 6-8 weeks (with parallelization)
**Parallelization Potential**: High (see strategy doc)

### Suggested Work Order

1. **Week 1**: Phase 1, Phase 4.1-4.2 (Runtime + Basic Loader)
2. **Week 2**: Phase 2 (Serialization)
3. **Week 3**: Phase 3, Phase 4.3-4.5 (Bindings + Full Loader)
4. **Week 4**: Phase 5, Phase 7 (Functions + Constructors)
5. **Week 5**: Phase 6 (Visitors)
6. **Week 6**: Phase 8, Phase 9 (Processors + File Managers)
7. **Week 7**: Phase 10 (Scoping)
8. **Week 8**: Integration Testing + Performance + Documentation

### Critical Path

```
Phase 1 → Phase 2 → Phase 3 → Phase 6 → Phase 10 → Integration Testing
```

All other phases can be done in parallel with appropriate coordination.
