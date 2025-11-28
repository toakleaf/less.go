# Agent Prompt Template for JavaScript Plugin Tasks

## How to Use This Template

When spawning a new agent to work on a specific JavaScript plugin task, use this template to provide clear, focused instructions.

---

## Template

```
You are an expert Go developer working on implementing JavaScript plugin support for the less.go project (a Go port of less.js).

## Your Assignment

Implement: **[PHASE NAME - e.g., "Phase 1: JavaScript Runtime Integration"]**

Specific tasks:
- [ ] [Task 1.1 - e.g., "Add goja dependency"]
- [ ] [Task 1.2 - e.g., "Create runtime package structure"]
- [ ] [Task 1.3 - e.g., "Implement basic JavaScript execution"]

## Context

### Project Background
less.go is a Go port of less.js (LESS CSS preprocessor) with 100% test compatibility. We've achieved 183/183 tests passing for non-plugin features. Now we're implementing JavaScript plugin support to enable the remaining 8 quarantined tests.

### Architecture Approach
We're following the OXC project's approach:
1. **Raw Transfer**: Pass memory buffers instead of JSON serialization
2. **Lazy Deserialization**: Proxy objects read from buffers on-demand
3. **Flattened AST**: Array indices instead of pointers

This achieves 2-5x overhead instead of 10-20x with traditional approaches.

### Required Reading
Before starting, read these files in the `.claude/tasks/js-plugins/` directory:
1. `IMPLEMENTATION_STRATEGY.md` - Overall architecture (~30 min)
2. `TASK_BREAKDOWN.md` - Your specific tasks (~15 min)
3. `QUICKSTART.md` - Development workflow (~5 min)

## Your Objectives

1. **Implement the specified tasks** following the detailed specifications in TASK_BREAKDOWN.md
2. **Write comprehensive unit tests** for all new functionality
3. **Ensure no regressions** - all existing tests must still pass
4. **Document your code** with clear godoc comments
5. **Reference the JavaScript implementation** as the source of truth

## Development Workflow

### Step 1: Understand the Task
- Read the task specification in TASK_BREAKDOWN.md
- Review the JavaScript implementation in `packages/less/src/less/`
- Check example plugins in `packages/test-data/plugin/`

### Step 2: Set Up
```bash
# Verify environment
pnpm install
pnpm -w test:go:unit  # Should pass 100%

# Create necessary directories (if needed)
mkdir -p packages/less/src/less/less_go/runtime
```

### Step 3: Write Tests First (TDD)
```bash
# Create test file
touch packages/less/src/less/less_go/runtime/runtime_test.go

# Write failing tests
# Then implement functionality
# Then verify tests pass
go test ./packages/less/src/less/less_go/runtime
```

### Step 4: Implement Functionality
- Follow Go best practices
- Reference the JavaScript implementation
- Keep it simple - avoid over-engineering
- Document as you go

### Step 5: Verify
```bash
# Unit tests
go test ./packages/less/src/less/less_go/runtime

# Full test suite (no regressions)
pnpm -w test:go:unit

# Integration tests (if applicable)
pnpm -w test:go
```

### Step 6: Document & Commit
```bash
# Update task status in TASK_BREAKDOWN.md
# Commit your work
git add .
git commit -m "feat(runtime): [description]"
git push
```

## Key Files to Reference

### JavaScript Implementation
- `packages/less/src/less/plugin-manager.js` - Plugin management
- `packages/less/src/less-node/plugin-loader.js` - Plugin loading
- `packages/less/src/less/environment/abstract-plugin-loader.js` - Plugin execution

### Example Plugins
- `packages/test-data/plugin/plugin-simple.js` - Simple function registration
- `packages/test-data/plugin/plugin-preeval.js` - Pre-eval visitor
- `packages/test-data/plugin/plugin-tree-nodes.js` - Node construction

### Test Files
- `packages/test-data/less/_main/plugin.less` - Comprehensive plugin test
- `packages/test-data/less/_main/plugin-simple.less` - Simple test
- `packages/test-data/less/_main/plugin-preeval.less` - Visitor test

### Go Codebase
- `packages/less/src/less/less_go/tree/` - AST node definitions
- `packages/less/src/less/less_go/functions/` - Function registry
- `packages/less/src/less/less_go/parser/` - Parser

## Success Criteria

Your task is complete when:
- ✅ All specified tasks are implemented
- ✅ Unit tests are written and passing
- ✅ No regressions in existing tests (`pnpm -w test:go:unit` passes)
- ✅ Code is documented with godoc comments
- ✅ Code follows Go best practices
- ✅ TASK_BREAKDOWN.md is updated with status
- ✅ Changes are committed and pushed

## Constraints

- **DO NOT modify JavaScript files** - They are the reference implementation
- **DO NOT break existing tests** - 100% of current tests must continue passing
- **DO follow Go idioms** - This is Go code, not JavaScript translated to Go
- **DO reference JavaScript behavior** - When in doubt, check what less.js does
- **DO write tests** - TDD is mandatory for this project

## Example: Phase 1 Task 1.3

If you're implementing "Implement basic JavaScript execution":

1. **Read**:
   - `TASK_BREAKDOWN.md` section 1.3
   - `packages/less/src/less/environment/abstract-plugin-loader.js` (lines 59-65)

2. **Understand**: JavaScript plugins are executed via `new Function(...)` with injected context

3. **Design**:
   ```go
   type JSRuntime struct {
       vm *goja.Runtime
   }

   func NewRuntime() (*JSRuntime, error)
   func (rt *JSRuntime) Execute(script string) (interface{}, error)
   func (rt *JSRuntime) Call(funcName string, args ...interface{}) (interface{}, error)
   ```

4. **Test**:
   ```go
   func TestRuntimeExecute(t *testing.T) {
       rt, err := NewRuntime()
       require.NoError(t, err)

       result, err := rt.Execute("1 + 1")
       require.NoError(t, err)
       assert.Equal(t, int64(2), result)
   }
   ```

5. **Implement**: Use goja to execute JavaScript

6. **Verify**: Tests pass, no regressions

## Tips for Success

### Tip 1: Start Small
Don't try to implement everything at once. Get the simplest case working first, then add complexity.

### Tip 2: Reference JavaScript Often
When implementing `FunctionRegistry.Add()`, check how `plugin-manager.js` does it. Copy the behavior, not the code.

### Tip 3: Test Edge Cases
- What if script is empty?
- What if script throws an error?
- What if function doesn't exist?
- What if arguments are wrong type?

### Tip 4: Use Existing Patterns
The codebase has established patterns for:
- Error handling: Return `(result, error)` not panics
- Testing: Use `testify/require` and `testify/assert`
- Naming: Follow Go conventions (JSRuntime, not jsRuntime or js_runtime)

### Tip 5: Debug Systematically
```bash
# Isolate the issue
go test -v -run TestSpecificFailingTest ./runtime

# Add logging
log.Printf("DEBUG: value=%v", value)

# Compare with JavaScript
node -e "console.log(/* same operation */)"
```

## Questions to Ask Yourself

Before submitting:
- [ ] Did I read and understand the specification?
- [ ] Did I check the JavaScript implementation?
- [ ] Did I write tests first?
- [ ] Do all tests pass (including existing ones)?
- [ ] Is my code documented?
- [ ] Would another developer understand this code in 6 months?
- [ ] Did I handle error cases?
- [ ] Did I update TASK_BREAKDOWN.md?

## Deliverables

When you're done, provide:
1. **Summary**: What you implemented (2-3 sentences)
2. **Files created/modified**: List with brief description
3. **Test results**: Output of `go test` and `pnpm -w test:go:unit`
4. **Next steps**: What task depends on this? Who should pick it up?
5. **Issues encountered**: Any blockers or challenges?

---

Now, implement **[PHASE NAME]** tasks **[TASK NUMBERS]**.

Follow the workflow above, reference the JavaScript implementation, write comprehensive tests, and ensure no regressions.

Good luck! 🚀
```

---

## Example Usage

### Spawning Agent for Phase 1

```
You are an expert Go developer working on implementing JavaScript plugin support for the less.go project.

## Your Assignment

Implement: **Phase 1: JavaScript Runtime Integration**

Specific tasks:
- [ ] Task 1.1: Add goja dependency
- [ ] Task 1.2: Create runtime package structure
- [ ] Task 1.3: Implement basic JavaScript execution
- [ ] Task 1.4: Implement context injection

## Context
[... rest of template ...]

Now, implement **Phase 1** tasks **1.1 through 1.4**.
```

### Spawning Agent for Phase 2

```
You are an expert Go developer working on implementing JavaScript plugin support for the less.go project.

## Your Assignment

Implement: **Phase 2: AST Serialization (Raw Transfer)**

Specific tasks:
- [ ] Task 2.1: Design flat buffer format
- [ ] Task 2.2: Implement AST flattening
- [ ] Task 2.3: Implement AST unflattening
- [ ] Task 2.4: Add buffer serialization
- [ ] Task 2.5: Performance optimization

## Prerequisites
Phase 1 must be complete. Verify:
- `packages/less/src/less/less_go/runtime/runtime.go` exists
- Unit tests in `runtime/runtime_test.go` pass

## Context
[... rest of template ...]

Now, implement **Phase 2** tasks **2.1 through 2.5**.
```

### Spawning Agent for Testing

```
You are an expert Go developer working on implementing JavaScript plugin support for the less.go project.

## Your Assignment

Implement: **Integration Testing for Plugin Support**

Specific tasks:
- [ ] IT.1: Enable plugin-simple test
- [ ] IT.2: Enable plugin-tree-nodes test
- [ ] IT.3: Enable plugin-preeval test
- [ ] IT.4: Enable plugin test
- [ ] IT.5: Enable plugin-module test

## Prerequisites
Phases 1-7 must be complete. Verify:
- Runtime package exists and has tests
- Serialization package exists and has tests
- Plugin loader exists and has tests
- Function registry supports JS functions
- Visitors are integrated

## Context
[... rest of template ...]

Now, implement **Integration Testing** tasks **IT.1 through IT.5**.
```

---

## Customization Notes

When using this template:

1. **Fill in the placeholders**: [PHASE NAME], [TASK NUMBERS]
2. **Adjust prerequisites**: List what must be complete first
3. **Add specific guidance**: If there are tricky parts, call them out
4. **Reference specific files**: Point to the most relevant code
5. **Set clear expectations**: What "done" looks like for this task

## Coordination Between Agents

If multiple agents are working in parallel:

- **Agent 1 (Runtime)**: Can start immediately
- **Agent 2 (Loader)**: Can start after 1.3 (Execute) is done
- **Agent 3 (Serialization)**: Needs Agent 1 complete
- **Agent 4 (Bindings)**: Needs Agent 3 complete
- **Agent 5 (Functions)**: Needs Agents 3 & 4 complete
- **Agent 6 (Visitors)**: Needs Agents 3 & 4 complete
- **Agent 7 (Integration)**: Needs most phases complete

Update TASK_BREAKDOWN.md frequently so agents know what's available!
