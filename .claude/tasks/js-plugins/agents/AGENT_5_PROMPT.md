# AGENT 5: Visitor Integration

**Status**: ⏸️ Blocked - Wait for Agent 3 (Bindings)
**Dependencies**: Agent 3 (must have Visitor class working)
**Estimated Time**: 3-4 days
**Can work in parallel with**: Agent 4

---

You are implementing AST visitor integration for pre-eval and post-eval transformations in less.go.

## Your Mission

Implement Phase 6 (Visitor Integration) from the strategy document.

## Prerequisites

✅ Verify Agent 3 has completed:
- NodeFacade works
- Visitor class exists in bindings
- Can traverse AST from JavaScript

Check: `node runtime/bindings/test.js` (should have visitor tests)

## Required Reading

BEFORE starting, read:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 6
2. `packages/less/src/less/transform-tree.js` - JavaScript visitor implementation
3. Plugin example: `packages/test-data/plugin/plugin-preeval.js`

## Your Tasks

### 1. Create JSVisitor Wrapper (Go Side)

Create `packages/less/src/less/less_go/runtime/js_visitor.go`:

```go
package runtime

type JSVisitor struct {
    visitorID       string
    runtime         *NodeJSRuntime
    isPreEval       bool
    isReplacing     bool
}

func (jv *JSVisitor) Visit(node Node) (Node, error) {
    // 1. Flatten AST starting from node
    flat, err := FlattenAST(node)
    if err != nil {
        return nil, err
    }

    // 2. Write to shared memory
    shmKey, err := jv.runtime.WriteASTBuffer(flat)
    if err != nil {
        return nil, err
    }

    // 3. Send command to Node.js
    resp, err := jv.runtime.SendCommand(Command{
        Cmd:       "runVisitor",
        VisitorID: jv.visitorID,
        ShmKey:    shmKey,
    })

    if err != nil {
        return nil, err
    }

    // 4. Read modified AST from shared memory
    modifiedFlat, err := jv.runtime.ReadASTBuffer(resp.ShmKey)
    if err != nil {
        return nil, err
    }

    // 5. Unflatten to Go node
    result, err := UnflattenAST(modifiedFlat)
    if err != nil {
        return nil, err
    }

    return result, nil
}
```

### 2. Create PluginManager (Go Side)

Create `packages/less/src/less/less_go/runtime/plugin_manager.go`:

```go
package runtime

type PluginManager struct {
    preEvalVisitors  []*JSVisitor
    postEvalVisitors []*JSVisitor
}

func NewPluginManager() *PluginManager {
    return &PluginManager{
        preEvalVisitors:  []*JSVisitor{},
        postEvalVisitors: []*JSVisitor{},
    }
}

func (pm *PluginManager) AddVisitor(visitor *JSVisitor) {
    if visitor.isPreEval {
        pm.preEvalVisitors = append(pm.preEvalVisitors, visitor)
    } else {
        pm.postEvalVisitors = append(pm.postEvalVisitors, visitor)
    }
}

func (pm *PluginManager) GetPreEvalVisitors() []*JSVisitor {
    return pm.preEvalVisitors
}

func (pm *PluginManager) GetPostEvalVisitors() []*JSVisitor {
    return pm.postEvalVisitors
}
```

### 3. Implement Visitor Command (Node.js Side)

Update `plugin-host.js`:

```javascript
const { Visitor } = require('./bindings/visitor');

// Visitor registry
const registeredVisitors = new Map();

const pluginManager = {
    addVisitor(visitor) {
        // Validate visitor
        if (!(visitor instanceof Visitor)) {
            throw new Error('Visitor must extend Visitor class');
        }

        const id = `visitor_${Date.now()}`;
        registeredVisitors.set(id, visitor);

        return {
            id,
            isPreEval: visitor.isPreEvalVisitor || false,
            isReplacing: visitor.isReplacing || false
        };
    }
};

// Command: runVisitor
function handleRunVisitor(cmd) {
    const { visitorID, shmKey } = cmd;

    try {
        const visitor = registeredVisitors.get(visitorID);
        if (!visitor) {
            throw new Error(`Visitor not found: ${visitorID}`);
        }

        // Attach buffer
        const buffer = shm.get(shmKey);
        visitor._buffer = buffer;

        // Run visitor (starts at root index 0)
        const result = visitor.visit(0);

        // Buffer is modified in place
        // Return same shmKey
        return {
            success: true,
            shmKey  // Same key, modified buffer
        };

    } catch (error) {
        return {
            success: false,
            error: error.message,
            stack: error.stack
        };
    }
}
```

### 4. Integrate Pre-Eval Visitors

Find the evaluation entry point and add pre-eval:

```go
// In appropriate evaluation file (likely tree/ruleset.go or similar)

func (r *Ruleset) Eval(ctx *EvalContext) (Node, error) {
    // Run pre-eval visitors BEFORE evaluation
    node := Node(r)
    for _, visitor := range ctx.PluginManager.GetPreEvalVisitors() {
        var err error
        node, err = visitor.Visit(node)
        if err != nil {
            return nil, err
        }
    }

    // Continue with normal evaluation
    evaluated, err := node.Eval(ctx)
    if err != nil {
        return nil, err
    }

    return evaluated, nil
}
```

### 5. Integrate Post-Eval Visitors

Add post-eval visitors after evaluation:

```go
func EvaluateWithPlugins(root Node, ctx *EvalContext) (Node, error) {
    // Normal evaluation
    evaluated, err := root.Eval(ctx)
    if err != nil {
        return nil, err
    }

    // Run post-eval visitors AFTER evaluation
    result := evaluated
    for _, visitor := range ctx.PluginManager.GetPostEvalVisitors() {
        result, err = visitor.Visit(result)
        if err != nil {
            return nil, err
        }
    }

    return result, nil
}
```

### 6. Test with Real Visitor Plugin

Test with `plugin-preeval.js`:

```go
func TestVisitor_PluginPreeval(t *testing.T) {
    runtime := setupRuntime(t)
    defer runtime.Stop()

    // Load plugin-preeval.js
    loader := NewPluginLoader(runtime)
    plugin, err := loader.LoadPlugin(
        "../../test-data/plugin/plugin-preeval.js",
        nil,
        ".",
    )
    require.NoError(t, err)

    // Create plugin manager
    pm := NewPluginManager()

    // Register visitor
    for _, visitorInfo := range plugin.Visitors {
        visitor := &JSVisitor{
            visitorID:   visitorInfo.ID,
            runtime:     runtime,
            isPreEval:   visitorInfo.IsPreEval,
            isReplacing: visitorInfo.IsReplacing,
        }
        pm.AddVisitor(visitor)
    }

    // Create test AST with @replace variable
    variable := &Variable{Name: "@replace"}

    // Run visitor
    result, err := pm.GetPreEvalVisitors()[0].Visit(variable)
    require.NoError(t, err)

    // Verify variable was replaced with quoted string
    quoted, ok := result.(*Quoted)
    require.True(t, ok)
    assert.Equal(t, "bar", quoted.Value)
}
```

## Success Criteria

✅ **Complete When**:
- Pre-eval visitors run before evaluation
- Post-eval visitors run after evaluation
- Visitors can traverse entire AST
- Visitors can replace nodes
- Buffer modifications are reflected in Go
- Error handling works
- Node replacements work correctly
- Unit tests pass for visitor integration

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- NO integration test regressions: `pnpm -w test:go` (183/183)

## Test Requirements

```go
func TestVisitor_PreEval(t *testing.T)
func TestVisitor_PostEval(t *testing.T)
func TestVisitor_NodeReplacement(t *testing.T)
func TestVisitor_Traversal(t *testing.T)
func TestVisitor_PluginPreeval(t *testing.T)
```

## Deliverables

1. Working JSVisitor wrapper
2. PluginManager for visitor registration
3. Pre-eval integration in evaluation pipeline
4. Post-eval integration in evaluation pipeline
5. Node replacement support
6. All unit tests passing
7. No regressions
8. Brief summary

You're making AST transformations possible! 🎭
