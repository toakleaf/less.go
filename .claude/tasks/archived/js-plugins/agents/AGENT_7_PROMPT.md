# AGENT 7: Scoping & Integration Testing

**Status**: ⏸️ Blocked - Wait for Agents 2, 4, 5
**Dependencies**: Need plugin loading, functions, and visitors working
**Estimated Time**: 4-5 days
**This is the final phase!**

---

You are implementing plugin scope management and enabling the quarantined integration tests.

## Your Mission

Implement Phase 10 (Plugin Scoping) and enable all quarantined plugin tests.

## Prerequisites

✅ Verify these are complete:
- Agent 2: Plugin loading works
- Agent 4: Functions work
- Agent 5: Visitors work

Check:
```bash
go test ./runtime -run TestPluginLoader
go test ./runtime -run TestJSFunction
go test ./runtime -run TestVisitor
```

## Required Reading

BEFORE starting, read:
1. IMPLEMENTATION_STRATEGY.md - Focus on Phase 10
2. `packages/test-data/less/_main/plugin.less` - Shows scoping behavior
3. `packages/less/src/less/plugin-manager.js` - JavaScript scoping

## Your Tasks

### Phase 10: Plugin Scope Management

#### 1. Implement PluginScope

```go
// runtime/plugin_scope.go

type PluginScope struct {
    parent    *PluginScope
    plugins   []*Plugin
    functions map[string]*JSFunction
    visitors  []*JSVisitor
}

func NewPluginScope(parent *PluginScope) *PluginScope {
    return &PluginScope{
        parent:    parent,
        plugins:   []*Plugin{},
        functions: make(map[string]*JSFunction),
        visitors:  []*JSVisitor{},
    }
}

func (ps *PluginScope) AddPlugin(plugin *Plugin) {
    ps.plugins = append(ps.plugins, plugin)

    // Register plugin's functions in this scope
    for name, fn := range plugin.Functions {
        ps.functions[name] = fn
    }

    // Register plugin's visitors in this scope
    ps.visitors = append(ps.visitors, plugin.Visitors...)
}

func (ps *PluginScope) LookupFunction(name string) (*JSFunction, bool) {
    // Check local scope
    if fn, ok := ps.functions[name]; ok {
        return fn, true
    }

    // Check parent scopes
    if ps.parent != nil {
        return ps.parent.LookupFunction(name)
    }

    return nil, false
}

func (ps *PluginScope) GetVisitors() []*JSVisitor {
    // Collect visitors from this scope and all parent scopes
    visitors := append([]*JSVisitor{}, ps.visitors...)

    if ps.parent != nil {
        visitors = append(visitors, ps.parent.GetVisitors()...)
    }

    return visitors
}
```

#### 2. Add PluginScope to EvalContext

```go
// Modify EvalContext (wherever it's defined)

type EvalContext struct {
    // ... existing fields
    PluginScope    *PluginScope
    PluginManager  *PluginManager
}

func (ctx *EvalContext) Clone() *EvalContext {
    return &EvalContext{
        // ... clone existing fields
        PluginScope:   ctx.PluginScope,  // Keep reference to current scope
        PluginManager: ctx.PluginManager,
    }
}

func (ctx *EvalContext) EnterScope() *EvalContext {
    child := ctx.Clone()
    child.PluginScope = NewPluginScope(ctx.PluginScope)  // Create child scope
    return child
}
```

#### 3. Create Child Scopes in Evaluation

```go
// In ruleset.go (or wherever rulesets are evaluated)

func (r *Ruleset) Eval(ctx *EvalContext) (Node, error) {
    // Create child scope for this ruleset
    childCtx := ctx.EnterScope()

    // Evaluate @plugin directives in this ruleset
    for _, rule := range r.Rules {
        if pluginDir, ok := rule.(*PluginDirective); ok {
            plugin, err := childCtx.PluginManager.LoadPlugin(
                pluginDir.Path,
                pluginDir.Options,
                childCtx.CurrentDirectory,
            )
            if err != nil {
                return nil, err
            }

            // Add plugin to CHILD scope (local scope)
            childCtx.PluginScope.AddPlugin(plugin)
        }
    }

    // Evaluate rules with child scope
    // ...
}

// Similar for mixins, media queries, etc.
```

#### 4. Use Scoped Functions

```go
// In function call evaluation

func (fc *FunctionCall) Eval(ctx *EvalContext) (Node, error) {
    // Try plugin functions from current scope first
    if jsFn, ok := ctx.PluginScope.LookupFunction(fc.Name); ok {
        return jsFn.Call(fc.Args, ctx)
    }

    // Fall back to built-in functions
    return ctx.FunctionRegistry.Call(fc.Name, fc.Args, ctx)
}
```

### Integration Testing

#### 1. Enable plugin-simple Test

Edit `integration_suite_test.go`:

```go
// Find plugin-simple test and remove quarantine
{
    name:        "plugin-simple",
    quarantined: false,  // Changed from true!
},
```

Run test:
```bash
go test -v -run TestIntegrationSuite/plugin-simple
```

If it fails, debug and fix issues. Expected: ✅ Perfect CSS match

#### 2. Enable plugin-tree-nodes Test

```go
{
    name:        "plugin-tree-nodes",
    quarantined: false,
},
```

Test: `go test -v -run TestIntegrationSuite/plugin-tree-nodes`

#### 3. Enable plugin-preeval Test

```go
{
    name:        "plugin-preeval",
    quarantined: false,
},
```

Test: `go test -v -run TestIntegrationSuite/plugin-preeval`

#### 4. Enable Main plugin Test

```go
{
    name:        "plugin",
    quarantined: false,
},
```

Test: `go test -v -run TestIntegrationSuite/plugin`

This is the comprehensive test! It tests:
- Global plugins
- Local plugins
- Plugin scoping
- Function shadowing
- Visitor scope propagation

#### 5. Enable plugin-module Test

```go
{
    name:        "plugin-module",
    quarantined: false,
},
```

Test: `go test -v -run TestIntegrationSuite/plugin-module`

#### 6. Try bootstrap4 (Stretch Goal)

```go
{
    name:        "bootstrap4",
    quarantined: false,  // Only if you're feeling ambitious!
},
```

This is a stretch goal - bootstrap4 has complex plugin requirements.

## Success Criteria

✅ **Phase 10 Complete When**:
- Plugin scope hierarchy works
- Local plugins only affect current scope
- Global plugins affect entire file
- Function shadowing works (local overrides global)
- Visitor scope propagation works
- Unit tests pass for scoping

✅ **Integration Tests Pass**:
- ✅ plugin-simple - Basic function registration
- ✅ plugin-tree-nodes - Node construction
- ✅ plugin-preeval - Pre-eval visitor
- ✅ plugin - Comprehensive scoping test
- ✅ plugin-module - NPM module loading
- 🎯 bootstrap4 - Real-world usage (stretch goal)

✅ **No Regressions**:
- ALL existing tests still pass: `pnpm -w test:go:unit` (100%)
- All non-plugin integration tests: still 183/183

✅ **New Test Count**:
- Should go from 183/183 → 188+/191 (at least 5 plugin tests passing!)

## Test Requirements

```go
func TestPluginScope_Hierarchy(t *testing.T)
func TestPluginScope_Shadowing(t *testing.T)
func TestPluginScope_LocalVsGlobal(t *testing.T)
```

Integration tests:
```bash
# Test each plugin test individually
LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/plugin-simple
LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/plugin-tree-nodes
LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/plugin-preeval
LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/plugin

# Full suite
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

## Deliverables

1. Working plugin scope management
2. All 5+ plugin integration tests passing
3. Scoping tests passing
4. No regressions
5. Updated CLAUDE.md with new test counts
6. Victory celebration 🎉

## Final Check

Before declaring victory:

```bash
# All unit tests
pnpm -w test:go:unit
# Should be: 100% passing

# All integration tests
pnpm -w test:go
# Should show: 188+/191 tests passing (was 183/183)

# Check perfect matches increased
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | grep "Perfect CSS Matches"
# Should show increase from 94 → 99+ tests
```

You're finishing the plugin system! 🎯🎉
