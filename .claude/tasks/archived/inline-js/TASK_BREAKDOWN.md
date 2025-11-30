# Inline JavaScript Implementation - Task Breakdown

## Overview

This document breaks down the inline JavaScript implementation into discrete tasks. Due to the tight coupling between JavaScript and Go sides, we use 3 sequential agents with Agents 1 and 2 working in parallel.

## Task Status Legend

- 🔴 Not Started
- 🟡 In Progress
- 🟢 Complete
- ⏸️ Blocked

---

## Agent 1: JavaScript Side - plugin-host.js 🔴

**Can run in parallel with Agent 2**
**Estimated time**: 2-3 hours

### Task 1.1: Add evalJS Command Handler 🔴

Add a new command handler in `plugin-host.js` for evaluating inline JavaScript expressions.

**Location**: `packages/less/src/less/less_go/runtime/plugin-host.js`

**Add to handleCommand switch statement**:
```javascript
case 'evalJS':
  handleEvalJS(id, data);
  break;
```

**Implement handleEvalJS function**:
```javascript
/**
 * Evaluate an inline JavaScript expression
 * @param {number} id - Command ID
 * @param {Object} data - { expression: string, variables: { name: { value: string } } }
 */
function handleEvalJS(id, data) {
  const { expression, variables } = data || {};

  if (expression === undefined) {
    sendResponse(id, false, null, 'Expression is required');
    return;
  }

  try {
    // Build evaluation context with variables
    const evalContext = buildEvalContext(variables);

    // Create and execute function
    const fn = new Function(`return (${expression})`);
    const result = fn.call(evalContext);

    // Convert result to serializable format
    const converted = convertJSResult(result);

    sendResponse(id, true, converted);
  } catch (err) {
    // Format error for Less error reporting
    sendResponse(id, false, null, formatJSError(err));
  }
}
```

**Acceptance Criteria**:
- Command registered in handleCommand switch
- Function exists and compiles
- Unit test: Can evaluate simple expression `1 + 1`

---

### Task 1.2: Build Evaluation Context 🔴

Implement the variable context building that exposes Less variables as `this.varName` with a `toJS()` method.

**Add function**:
```javascript
/**
 * Build evaluation context from Less variables
 * Variables are exposed as this.varName with a toJS() method
 * @param {Object} variables - { name: { value: string, type: string } }
 * @returns {Object} Context object for function.call()
 */
function buildEvalContext(variables) {
  const context = {};

  if (!variables) {
    return context;
  }

  for (const [name, info] of Object.entries(variables)) {
    // Remove @ prefix if present
    const cleanName = name.startsWith('@') ? name.slice(1) : name;

    context[cleanName] = {
      value: info.value,
      toJS: function() {
        return this.value;
      }
    };
  }

  return context;
}
```

**Acceptance Criteria**:
- Variables accessible as `this.foo`
- `this.foo.toJS()` returns the CSS string value
- Handles empty/null variables

---

### Task 1.3: Convert JavaScript Results 🔴

Implement result conversion that matches less.js behavior for different JavaScript types.

**Add function**:
```javascript
/**
 * Convert JavaScript result to serializable format
 * Matches less.js behavior in tree/javascript.js eval()
 * @param {*} result - The JavaScript result
 * @returns {Object} Serializable result with type info
 */
function convertJSResult(result) {
  const type = typeof result;

  if (type === 'number' && !isNaN(result)) {
    return { type: 'number', value: result };
  } else if (type === 'string') {
    return { type: 'string', value: result };
  } else if (Array.isArray(result)) {
    return { type: 'array', value: result.join(', ') };
  } else if (type === 'boolean') {
    return { type: 'boolean', value: result };
  } else if (result === null || result === undefined) {
    return { type: 'empty', value: '' };
  } else {
    // Object or other - convert to string
    return { type: 'other', value: String(result) };
  }
}
```

**Acceptance Criteria**:
- Numbers return `{ type: 'number', value: 42 }`
- Strings return `{ type: 'string', value: 'hello' }`
- Arrays return `{ type: 'array', value: '1, 2, 3' }`
- Booleans return `{ type: 'boolean', value: true }`

---

### Task 1.4: Format JavaScript Errors 🔴

Implement error formatting that matches less.js error messages.

**Add function**:
```javascript
/**
 * Format JavaScript error for Less error reporting
 * Matches less.js format: "JavaScript evaluation error: 'TypeError: ...'"
 * @param {Error} err - The JavaScript error
 * @returns {string} Formatted error message
 */
function formatJSError(err) {
  // Match less.js format from tree/js-eval-node.js
  const errName = err.name || 'Error';
  const errMsg = err.message || String(err);
  return `JavaScript evaluation error: '${errName}: ${errMsg.replace(/["]/g, "'")}'`;
}
```

**Acceptance Criteria**:
- TypeError formatted as `JavaScript evaluation error: 'TypeError: ...'`
- Quotes in error message replaced with single quotes
- Unknown errors still produce valid format

---

### Task 1.5: Handle Edge Cases 🔴

Handle special cases that appear in the test files.

**Edge cases to handle**:

1. **Empty function result**: `` `+function(){}` `` should return empty string
2. **Multiline expressions**: Expressions can span multiple lines
3. **Complex expressions**: IIFEs, ternary operators, etc.

**Update handleEvalJS**:
```javascript
// Handle special case: empty/undefined results
if (result === undefined || result === null) {
  sendResponse(id, true, { type: 'empty', value: '' });
  return;
}

// Handle NaN (should be treated as empty)
if (typeof result === 'number' && isNaN(result)) {
  sendResponse(id, true, { type: 'empty', value: '' });
  return;
}
```

**Acceptance Criteria**:
- `` `+function(){}` `` returns empty
- Multiline expressions execute correctly
- IIFEs work: `` `(function(){ return 1 })()` ``

---

## Agent 2: Go Side - js_eval_node.go 🔴

**Can run in parallel with Agent 1**
**Estimated time**: 2-3 hours

### Task 2.1: Access Node.js Runtime 🔴

Modify `EvaluateJavaScript()` to access the Node.js runtime from the evaluation context.

**Location**: `packages/less/src/less/less_go/js_eval_node.go`

**Current code** (line 102-188):
```go
func (j *JsEvalNode) EvaluateJavaScript(expression string, context any) (any, error) {
    // ... existing code checks javascriptEnabled ...
    // ... currently returns error "JavaScript evaluation is not supported" ...
}
```

**Modify to get runtime**:
```go
// After javascriptEnabled check, get the runtime
var rt *runtime.NodeJSRuntime

// Try to get runtime from Eval context
if evalCtx, ok := context.(*Eval); ok {
    if evalCtx.PluginBridge != nil {
        rt = evalCtx.PluginBridge.GetRuntime()
    } else if evalCtx.LazyPluginBridge != nil {
        rt = evalCtx.LazyPluginBridge.GetRuntime()
    }
}

// Also check map context for pluginBridge
if rt == nil {
    if mapCtx, ok := context.(map[string]any); ok {
        if bridge, ok := mapCtx["pluginBridge"].(*NodeJSPluginBridge); ok {
            rt = bridge.GetRuntime()
        }
        if bridge, ok := mapCtx["pluginBridge"].(*LazyNodeJSPluginBridge); ok {
            rt = bridge.GetRuntime()
        }
    }
}

if rt == nil {
    return nil, &LessError{
        Type:     "JavaScript",
        Message:  "JavaScript runtime not available",
        Filename: getFilename(),
        Index:    j.GetIndex(),
    }
}
```

**Acceptance Criteria**:
- Can access runtime from `*Eval` context
- Can access runtime from map context
- Returns proper error if runtime not available

---

### Task 2.2: Build Variable Context 🔴

Build the variable context to send to Node.js for `this.foo.toJS()` access.

**Add helper function**:
```go
// buildVariableContext extracts variables from the evaluation context
// for access via this.varName.toJS() in JavaScript
func (j *JsEvalNode) buildVariableContext(context any) map[string]map[string]any {
    variables := make(map[string]map[string]any)

    // Get frames from context
    var frames []ParserFrame
    if evalCtx, ok := context.(*Eval); ok {
        frames = evalCtx.GetFrames()
    } else if wrapper, ok := context.(*contextWrapper); ok {
        frames = wrapper.GetFrames()
    }

    if len(frames) == 0 {
        return variables
    }

    // Get variables from the first frame (current scope)
    frame := frames[0]
    frameVars := frame.Variables()

    for name, decl := range frameVars {
        // Evaluate the variable to get its CSS value
        if declNode, ok := decl.(interface{ Eval(any) (any, error) }); ok {
            result, err := declNode.Eval(context)
            if err == nil {
                // Get the CSS representation
                cssValue := j.jsify(result)
                variables[name] = map[string]any{
                    "value": cssValue,
                }
            }
        }
    }

    return variables
}
```

**Acceptance Criteria**:
- Extracts variables from current scope
- Evaluates variables to get CSS values
- Returns map suitable for JSON serialization

---

### Task 2.3: Send Command to Node.js 🔴

Replace the error return with a call to the Node.js runtime.

**Modify EvaluateJavaScript**:
```go
// Send evalJS command to Node.js
resp, err := rt.SendCommand(runtime.Command{
    Cmd: "evalJS",
    Data: map[string]any{
        "expression": expressionForError, // After variable interpolation
        "variables":  j.buildVariableContext(context),
    },
})

if err != nil {
    return nil, &LessError{
        Type:     "JavaScript",
        Message:  fmt.Sprintf("JavaScript evaluation error: %v", err),
        Filename: getFilename(),
        Index:    j.GetIndex(),
    }
}

if !resp.Success {
    return nil, &LessError{
        Type:     "JavaScript",
        Message:  resp.Error,
        Filename: getFilename(),
        Index:    j.GetIndex(),
    }
}

// Process the result
return j.processJSResult(resp.Result)
```

**Acceptance Criteria**:
- Sends correct command structure
- Handles error responses
- Passes result to processing function

---

### Task 2.4: Process JavaScript Result 🔴

Implement result processing that converts Node.js response to Go types.

**Add helper function**:
```go
// processJSResult converts the JavaScript result to the appropriate Go type
func (j *JsEvalNode) processJSResult(result any) (any, error) {
    resultMap, ok := result.(map[string]any)
    if !ok {
        return NewAnonymous(result, 0, nil, false, false, nil), nil
    }

    resultType, _ := resultMap["type"].(string)
    value := resultMap["value"]

    switch resultType {
    case "number":
        if numVal, ok := value.(float64); ok {
            dim, err := NewDimension(numVal, nil)
            if err != nil {
                return nil, err
            }
            return dim, nil
        }
    case "string":
        if strVal, ok := value.(string); ok {
            return NewQuoted(`"`+strVal+`"`, strVal, false, j.GetIndex(), j.FileInfo()), nil
        }
    case "array":
        if strVal, ok := value.(string); ok {
            return NewAnonymous(strVal, 0, nil, false, false, nil), nil
        }
    case "boolean":
        if boolVal, ok := value.(bool); ok {
            return NewAnonymous(boolVal, 0, nil, false, false, nil), nil
        }
    case "empty":
        return NewAnonymous("", 0, nil, false, false, nil), nil
    }

    // Default: return as Anonymous
    return NewAnonymous(value, 0, nil, false, false, nil), nil
}
```

**Note**: The `JavaScript.Eval()` method in `javascript.go` already handles the final conversion, so `processJSResult` just needs to return the raw value in the correct format.

**Acceptance Criteria**:
- Numbers → appropriate numeric value
- Strings → string value
- Arrays → comma-separated string
- Empty → empty string
- Booleans → boolean value

---

### Task 2.5: Handle Escaped Expressions 🔴

Ensure escaped expressions (with `~` prefix) are handled correctly.

The `JavaScript.Eval()` method in `javascript.go` already handles the `escaped` flag when creating the `Quoted` node. Verify this works correctly.

**In javascript.go, verify line 67**:
```go
return NewQuoted(`"`+v+`"`, v, j.escaped, j.GetIndex(), j.FileInfo()), nil
```

The `j.escaped` flag should propagate correctly.

**Acceptance Criteria**:
- `~`1 + 1`` produces unquoted output
- Regular backticks produce quoted strings

---

## Agent 3: Integration & Testing 🔴

**Runs after Agents 1 and 2 complete**
**Estimated time**: 2-3 hours

### Task 3.1: Un-quarantine JavaScript Test 🔴

Remove the `javascript` test from quarantine and run it.

**Location**: `packages/less/src/less/less_go/integration_suite_test.go`

**Find and modify** (around line 295-317):
```go
// Remove "javascript" from quarantined tests
quarantinedTests := map[string]bool{
    // Keep these quarantined:
    "plugin":        true,
    "plugin-module": true,
    "plugin-preeval": true,
    "bootstrap4":    true,  // Being fixed separately

    // REMOVE these - now supported:
    // "javascript": true,  // <-- Remove this line
}
```

**Run test**:
```bash
LESS_GO_DEBUG=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go
```

**Acceptance Criteria**:
- Test runs without quarantine skip
- Debug any failures

---

### Task 3.2: Fix JavaScript Test Issues 🔴

Debug and fix any issues with the javascript test.

**Expected test file**: `packages/test-data/less/_main/javascript.less`
**Expected output**: `packages/test-data/css/_main/javascript.css`

**Common issues to check**:
1. Variable interpolation `@{varName}` working correctly
2. `this.foo.toJS()` returning correct values
3. Array joining producing correct format
4. Multiline expressions executing correctly
5. Escaped vs non-escaped output

**Debugging commands**:
```bash
# See diff between expected and actual
LESS_GO_DIFF=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go

# Enable tracing
LESS_GO_TRACE=1 go test -v -run "TestIntegrationSuite/_main/javascript" ./packages/less/src/less/less_go
```

**Acceptance Criteria**:
- All javascript test cases pass
- CSS output matches expected exactly

---

### Task 3.3: Un-quarantine Error Tests 🔴

Remove `js-type-errors` test suite from quarantine.

**Modify integration_suite_test.go**:
```go
// Remove js-type-errors/* from quarantined patterns
```

**Run test**:
```bash
go test -v -run "TestIntegrationSuite/js-type-errors" ./packages/less/src/less/less_go
```

**Acceptance Criteria**:
- Error tests run
- Correct errors are detected

---

### Task 3.4: Verify Error Messages 🔴

Ensure error messages match expected format (or close enough).

**Expected error format** (from `js-type-error.txt`):
```
SyntaxError: JavaScript evaluation error: 'TypeError: Cannot read property 'toJS' of undefined' in {path}js-type-error.less on line 2, column 8:
```

**Key requirements**:
- Error type detected (TypeError, SyntaxError, etc.)
- Line and column numbers correct
- File path included

**If exact match not possible**, update the `.txt` files to match Go's error format while preserving the essential error detection.

**Acceptance Criteria**:
- Error tests pass
- Essential error information preserved

---

### Task 3.5: Verify no-js-errors Test 🔴

This test should already work since the `javascriptEnabled: false` check is already implemented.

**Run test**:
```bash
go test -v -run "TestIntegrationSuite/no-js-errors" ./packages/less/src/less/less_go
```

**Expected behavior**: Test should error with "Inline JavaScript is not enabled" message.

**If quarantined**, remove from quarantine and verify.

**Acceptance Criteria**:
- Test passes
- Correct error message

---

### Task 3.6: Verify No Regressions 🔴

Run full test suite to ensure no regressions.

**Commands**:
```bash
# Unit tests
pnpm -w test:go:unit

# Integration tests
LESS_GO_QUIET=1 pnpm -w test:go 2>&1 | tail -100
```

**Acceptance Criteria**:
- All unit tests pass (3,012+)
- All integration tests pass (183/183 or more)
- No previously passing tests now fail

---

### Task 3.7: Update Documentation 🔴

Update CLAUDE.md and any other relevant documentation.

**Updates needed**:
1. Remove `javascript`, `js-type-errors`, `no-js-errors` from quarantined list
2. Update test counts
3. Add note about inline JS support

**Location**: `/home/user/less.go/CLAUDE.md`

**Acceptance Criteria**:
- Documentation reflects new capabilities
- Test counts updated

---

## Summary

**Total Tasks**: 17
**Agents**: 3 (2 parallel, 1 sequential)

### Execution Order

```
┌─────────────────┐     ┌─────────────────┐
│    Agent 1      │     │    Agent 2      │
│   (JS Side)     │     │   (Go Side)     │
│                 │     │                 │
│ Tasks 1.1-1.5   │     │ Tasks 2.1-2.5   │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
                     ▼
         ┌─────────────────────┐
         │      Agent 3        │
         │  (Integration &     │
         │     Testing)        │
         │                     │
         │   Tasks 3.1-3.7     │
         └─────────────────────┘
```

### Critical Path

Agent 1 + Agent 2 (parallel) → Agent 3 (sequential)

### Dependencies

- Agent 1 and Agent 2 have no dependencies on each other
- Agent 3 depends on both Agent 1 and Agent 2 completing
- All agents should reference the JavaScript implementation in `packages/less/src/less/tree/javascript.js`
