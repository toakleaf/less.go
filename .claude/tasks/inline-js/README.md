# Inline JavaScript Implementation

## Overview

This task implements inline JavaScript expression evaluation in less.go, enabling backtick expressions like `` `1 + 1` `` to be executed via the existing Node.js runtime infrastructure.

## Background

Less.js supports inline JavaScript expressions using backtick syntax:

```less
.example {
    @foo: 42;
    width: `1 + 1`;                    // Simple expression: 2
    height: `parseInt(this.foo.toJS())`; // Variable access: 42
    content: ~`"hello" + " world"`;    // Escaped: hello world
}
```

This feature is distinct from JavaScript plugins (which we've already implemented). It executes arbitrary JavaScript expressions inline during LESS compilation.

## Current State

### Already Implemented
- **Node.js Runtime**: `NodeJSRuntime` manages a Node.js subprocess with JSON IPC
- **Plugin Host**: `plugin-host.js` handles plugin loading, function calls, visitors
- **JavaScript AST Node**: `javascript.go` defines the `JavaScript` struct with `Eval()` method
- **JsEvalNode Base**: `js_eval_node.go` has `EvaluateJavaScript()` that handles variable interpolation

### Current Behavior
- `EvaluateJavaScript()` returns error: "JavaScript evaluation is not supported in the Go port"
- Variable interpolation (`@{varName}`) is partially implemented
- Error handling for `javascriptEnabled: false` is already working

### Quarantined Tests (4 test suites)
1. **`javascript`** - Main inline JS test (basic expressions, variable access, escaping)
2. **`js-type-errors/*`** - JavaScript runtime error tests
3. **`no-js-errors/*`** - Tests when `javascriptEnabled: false` (should already work!)

## Architecture

We leverage the existing Node.js runtime infrastructure:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Go (less.go)                             │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │  JsEvalNode.EvaluateJavaScript(expression, context)     │  │
│   │    1. Check javascriptEnabled                            │  │
│   │    2. Interpolate @{variables}                           │  │
│   │    3. Build variable context for this.foo.toJS()         │  │
│   │    4. Call NodeJSRuntime.SendCommand("evalJS", ...)      │  │
│   │    5. Convert result to Go type                          │  │
│   └──────────────────────────────────────┬──────────────────┘  │
│                                          │                      │
│                      JSON over stdin/stdout                     │
│                                          │                      │
├──────────────────────────────────────────┼──────────────────────┤
│                                          ▼                      │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │              Node.js (plugin-host.js)                    │  │
│   │                                                          │  │
│   │   handleEvalJS(id, data):                                │  │
│   │     1. Build evaluation context with variables           │  │
│   │     2. Create function: new Function("return (" + expr)  │  │
│   │     3. Execute with variable context as 'this'           │  │
│   │     4. Return result to Go                               │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│                        Node.js Process                          │
└─────────────────────────────────────────────────────────────────┘
```

## Key Features to Implement

### 1. Basic Expression Evaluation
```less
js: `42`;           // → 42 (number → Dimension)
js: `1 + 1`;        // → 2
js: `"hello"`;      // → "hello" (string → Quoted)
js: `[1, 2, 3]`;    // → 1, 2, 3 (array → Anonymous)
```

### 2. Variable Access via `this`
```less
@foo: 42;
var: `parseInt(this.foo.toJS())`; // → 42
```

Variables are exposed on `this` with a `.toJS()` method that returns the CSS string value.

### 3. Variable Interpolation
```less
@world: "world";
width: ~`"hello" + " " + @{world}`; // → hello world
```

`@{varName}` is replaced with the variable's CSS value before execution.

### 4. Escaped Expressions
```less
escaped: ~`2 + 5 + 'px'`; // → 7px (no quotes)
```

The `~` prefix marks the result as "escaped" (not quoted).

### 5. Error Handling
```less
// When javascriptEnabled: false
a: `1 + 1`;  // Error: Inline JavaScript is not enabled

// When JavaScript throws
var: `this.undefined.property`; // Error: TypeError: Cannot read property...
```

## Agent Assignments

### Agent 1: JavaScript Side (plugin-host.js)
- Add `evalJS` command handler
- Build variable context with `toJS()` methods
- Execute expressions using `new Function()`
- Handle errors with proper formatting

### Agent 2: Go Side (js_eval_node.go)
- Modify `EvaluateJavaScript()` to call Node.js runtime
- Serialize variable context
- Handle response and convert to Go types
- Access runtime via `Eval.PluginBridge` or `Eval.LazyPluginBridge`

### Agent 3: Integration & Testing
- Un-quarantine `javascript` test suite
- Un-quarantine `js-type-errors` test suite
- Verify `no-js-errors` works (should already work)
- Fix any remaining issues

## Success Criteria

- [ ] `javascript` test suite passes (perfect CSS match)
- [ ] `js-type-errors` tests correctly detect and report JavaScript errors
- [ ] `no-js-errors` tests correctly error when `javascriptEnabled: false`
- [ ] No regressions in existing tests (`pnpm -w test:go:unit` 100%)
- [ ] No regressions in integration tests (`pnpm -w test:go` 183/183)

## Files to Modify

### JavaScript Side
- `packages/less/src/less/less_go/runtime/plugin-host.js` - Add `evalJS` handler

### Go Side
- `packages/less/src/less/less_go/js_eval_node.go` - Modify `EvaluateJavaScript()`

### Test Configuration
- `packages/less/src/less/less_go/integration_suite_test.go` - Remove quarantine

## Timeline

This is a focused feature with minimal scope. Estimated effort:
- Agent 1 (JS): ~2-3 hours
- Agent 2 (Go): ~2-3 hours
- Agent 3 (Testing): ~2-3 hours

Agents 1 and 2 can work in parallel. Agent 3 runs sequentially after 1 and 2 complete.

## Related Documentation

- `.claude/tasks/js-plugins/` - Plugin implementation (completed)
- `packages/less/src/less/tree/javascript.js` - JavaScript reference implementation
- `packages/less/src/less/tree/js-eval-node.js` - JS evaluation reference

## Quick Reference

### Test Commands
```bash
# Run specific test
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go

# Run unit tests (no regressions)
pnpm -w test:go:unit

# Run integration tests
pnpm -w test:go
```

### Key Files
| File | Purpose |
|------|---------|
| `runtime/plugin-host.js` | Node.js side - add evalJS command |
| `js_eval_node.go` | Go side - call Node.js for evaluation |
| `javascript.go` | JavaScript AST node definition |
| `contexts.go` | Eval context with PluginBridge |
