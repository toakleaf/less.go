# JavaScript Plugin Implementation - Quick Start Guide

## For New Agents

If you're an agent assigned to work on JavaScript plugin implementation, this guide will help you get started quickly.

## 📚 Required Reading

Before starting any task, read these documents in order:

1. **IMPLEMENTATION_STRATEGY.md** - Overall architecture and approach (30 min read)
2. **TASK_BREAKDOWN.md** - Detailed task specifications (20 min read)
3. This document - Quick start guide (5 min read)

## 🎯 Choosing a Task

### Check Current Status

1. Look at `TASK_BREAKDOWN.md` for task status (🔴 Not Started, 🟡 In Progress, 🟢 Complete)
2. Find tasks marked 🔴 that you're interested in
3. Check dependencies - ensure prerequisite phases are complete
4. Claim the task by updating status to 🟡 and adding your agent ID

### Task Selection Strategy

**If you're the first agent**:
- Start with **Phase 1** (Runtime Integration) - no dependencies
- Or **Phase 4.1-4.2** (Basic Plugin Loader) - can work in parallel

**If Phase 1 is complete**:
- Pick **Phase 2** (Serialization) - critical path
- Or **Phase 4.3-4.5** (Advanced Loader) - can work in parallel

**If Phase 2 & 3 are complete**:
- Pick **Phase 5** (Functions), **Phase 6** (Visitors), or **Phase 7** (Constructors)
- These can be done in parallel by different agents

**If you prefer testing**:
- Pick integration testing tasks (IT.1-IT.6)
- Pick performance optimization (PO.1-PO.3)

## 🛠️ Development Workflow

### 1. Set Up Your Environment

```bash
# Ensure dependencies are installed
pnpm install

# Verify tests pass
pnpm -w test:go:unit
pnpm -w test:go

# Create feature branch (if not already on one)
git checkout -b <your-branch-name>
```

### 2. Create Package Structure (if needed)

For Phase 1 (Runtime), create:
```bash
mkdir -p packages/less/src/less/less_go/runtime
cd packages/less/src/less/less_go/runtime
```

Create initial files:
- `runtime.go` - Main runtime implementation
- `runtime_test.go` - Unit tests
- `doc.go` - Package documentation

### 3. Follow TDD (Test-Driven Development)

```bash
# Write tests first
# Edit: runtime_test.go

# Run tests (they should fail)
go test ./runtime

# Implement functionality
# Edit: runtime.go

# Run tests (they should pass)
go test ./runtime

# Run full suite (ensure no regressions)
pnpm -w test:go:unit
```

### 4. Reference JavaScript Implementation

Always check the original less.js code:

```bash
# Plugin manager
less packages/less/src/less/plugin-manager.js

# Plugin loader
less packages/less/src/less-node/plugin-loader.js

# Abstract plugin loader
less packages/less/src/less/environment/abstract-plugin-loader.js

# Example plugins
ls packages/test-data/plugin/

# Example usage
less packages/test-data/less/_main/plugin.less
```

### 5. Write Clean, Documented Code

Follow the project conventions:

```go
// Package runtime provides JavaScript execution capabilities for LESS plugins.
//
// This package implements the bridge between Go and JavaScript, allowing
// LESS plugins written in JavaScript to be executed within the Go runtime.
package runtime

// JSRuntime manages a JavaScript execution environment.
//
// It uses goja (pure Go JavaScript implementation) to execute plugin code
// in a sandboxed environment with controlled access to Go functionality.
type JSRuntime struct {
    vm      *goja.Runtime
    context map[string]interface{}
}

// NewRuntime creates a new JavaScript runtime instance.
//
// The runtime is initialized with a fresh VM and empty context.
// Use Set() to inject Go values before executing scripts.
func NewRuntime() (*JSRuntime, error) {
    // Implementation
}
```

### 6. Commit & Push Regularly

```bash
# Commit working code frequently
git add .
git commit -m "feat(runtime): implement basic JS execution"

# Push to your branch
git push -u origin <your-branch-name>
```

## 📖 Key Concepts

### The OXC Approach

Our implementation is inspired by OXC's approach to running JavaScript from Rust:

1. **Raw Transfer**: Pass memory buffers instead of serializing to JSON
2. **Lazy Deserialization**: Create proxy objects that read from buffers on-demand
3. **Flattened AST**: Use array indices instead of pointers for tree navigation

**Why this matters**: Traditional approaches are 10-20x slower. This approach is only 2-5x slower.

### LESS Plugin API

Plugins are JavaScript modules that export:

```javascript
module.exports = {
    // Called once when plugin is first loaded
    install(less, pluginManager, functionRegistry) {
        // Register functions, visitors, etc.
    },

    // Called every time @plugin directive is encountered
    use(plugin) {
        // Per-use logic
    },

    // Optional: Handle plugin options
    setOptions(options) {
        // Called when: @plugin (option1, option2) "path"
    },

    // Optional: Minimum Less version
    minVersion: [3, 0, 0]
};
```

### Plugin Capabilities

Plugins can:

1. **Add Functions**: `functions.add('myFunc', fn)` - callable from LESS
2. **Add Visitors**: `manager.addVisitor(visitor)` - transform AST
3. **Add Preprocessors**: `manager.addPreProcessor(fn, priority)` - transform source
4. **Add Postprocessors**: `manager.addPostProcessor(fn, priority)` - transform CSS
5. **Add File Managers**: `manager.addFileManager(fm)` - custom import resolution

## 🧪 Testing Strategy

### Unit Tests

For each feature, write unit tests:

```go
func TestRuntimeExecute(t *testing.T) {
    rt, err := NewRuntime()
    require.NoError(t, err)

    result, err := rt.Execute("1 + 1")
    require.NoError(t, err)
    assert.Equal(t, int64(2), result)
}

func TestRuntimeExecuteError(t *testing.T) {
    rt, err := NewRuntime()
    require.NoError(t, err)

    _, err = rt.Execute("throw new Error('test error')")
    require.Error(t, err)
    assert.Contains(t, err.Error(), "test error")
}
```

### Integration Tests

Once your phase is complete, test with real LESS files:

```bash
# Enable a quarantined test
# Edit: packages/less/src/less/less_go/integration_suite_test.go
# Change: {name: "plugin-simple", quarantined: true}
# To:     {name: "plugin-simple", quarantined: false}

# Run integration test
pnpm -w test:go

# Check for perfect match
# Should see: ✅ plugin-simple
```

### Benchmarks

For performance-critical code, add benchmarks:

```go
func BenchmarkFlattenAST(b *testing.B) {
    // Parse a typical LESS file
    input := loadTestFile("test.less")
    ast, _ := parser.Parse(input)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        FlattenAST(ast)
    }
}
```

Run benchmarks:
```bash
go test -bench=. -benchmem ./runtime
```

## 🔍 Debugging Tips

### Enable Verbose Logging

```go
import "log"

func (rt *JSRuntime) Execute(script string) (interface{}, error) {
    log.Printf("Executing JS: %s", script)
    result, err := rt.vm.RunString(script)
    log.Printf("Result: %v, Error: %v", result, err)
    return result, err
}
```

### Use Integration Test Debug Mode

```bash
# Show detailed execution trace
LESS_GO_DEBUG=1 pnpm -w test:go

# Show CSS differences
LESS_GO_DIFF=1 pnpm -w test:go

# Combine both
LESS_GO_DEBUG=1 LESS_GO_DIFF=1 pnpm -w test:go
```

### Compare with JavaScript

```bash
# Run JavaScript version
cd packages/less
pnpm test:fixtures

# Run Go version
pnpm -w test:go

# Compare outputs
diff \
  packages/test-data/css/_main/plugin-simple.css \
  packages/less/src/less/less_go/test_output/plugin-simple.css
```

### Inspect AST

```go
import "encoding/json"

// Pretty-print AST
func debugAST(node Node) {
    data, _ := json.MarshalIndent(node, "", "  ")
    fmt.Println(string(data))
}
```

## 📋 Checklist for Completing a Task

Before marking a task as 🟢 Complete:

- [ ] Code compiles without errors
- [ ] Unit tests written and passing
- [ ] Integration tests passing (if applicable)
- [ ] No regressions in existing tests (`pnpm -w test:go:unit`)
- [ ] Code documented (godoc comments)
- [ ] Code reviewed (self-review for quality)
- [ ] Benchmarks run (if performance-critical)
- [ ] Committed and pushed to branch
- [ ] TASK_BREAKDOWN.md updated with status

## 🆘 Getting Help

### Resources

1. **less.js source code**: `packages/less/src/less/`
2. **Go port source code**: `packages/less/src/less/less_go/`
3. **Test files**: `packages/test-data/less/_main/` and `packages/test-data/plugin/`
4. **OXC references**: Links in IMPLEMENTATION_STRATEGY.md
5. **goja documentation**: https://github.com/dop251/goja

### Common Issues

#### "Cannot find goja package"

```bash
# Add dependency
go get github.com/dop251/goja
go mod tidy
```

#### "Tests failing after my changes"

```bash
# Check what changed
git diff

# Revert and try again
git checkout -- <file>

# Run specific test
go test -v -run TestSpecificTest ./runtime
```

#### "Integration test not matching JavaScript output"

```bash
# Compare outputs
LESS_GO_DIFF=1 go test -v -run TestIntegrationSuite/plugin-simple

# Check JavaScript version
cd packages/less
pnpm test:fixtures -- plugin-simple

# Manually inspect
cat packages/test-data/css/_main/plugin-simple.css
cat packages/less/src/less/less_go/test_output/plugin-simple.css
```

#### "JavaScript error not making sense"

```bash
# Add console.log to plugin code (temporarily)
# Edit: packages/test-data/plugin/plugin-simple.js
functions.add('pi', function() {
    console.log('pi called!');
    return less.dimension(Math.PI);
});

# Capture console output in Go
rt.vm.Set("console", map[string]interface{}{
    "log": func(args ...interface{}) {
        fmt.Println("JS:", args...)
    },
})
```

## 🚀 Next Steps

1. **Read the strategy doc** - Understand the big picture
2. **Pick a task** - Choose something with satisfied dependencies
3. **Claim it** - Update TASK_BREAKDOWN.md
4. **Code it** - Follow TDD, reference JS code
5. **Test it** - Unit tests, integration tests, benchmarks
6. **Ship it** - Commit, push, update status

Good luck! 🎉

## 📞 Coordination

If multiple agents are working in parallel:

1. **Communicate early**: Update TASK_BREAKDOWN.md when you start
2. **Avoid conflicts**: Don't work on overlapping code areas
3. **Share learnings**: If you discover something useful, document it
4. **Review each other**: Code reviews improve quality
5. **Integrate frequently**: Don't go weeks without merging

---

**Remember**: The goal is 100% compatibility with less.js plugins. When in doubt, check the JavaScript implementation!
