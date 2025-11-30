# Agent 2: Go Side - js_eval_node.go

**Status**: 🔴 Can start immediately
**Can run in parallel with**: Agent 1
**Estimated Time**: 2-3 hours
**Blocks**: Agent 3

---

## Your Mission

Modify `js_eval_node.go` to call the Node.js runtime instead of returning an error. This enables Less.js backtick expressions to be executed via the existing plugin infrastructure.

## Required Reading

Before starting, read these files:
1. `.claude/tasks/inline-js/README.md` - Overview (~5 min)
2. `.claude/tasks/inline-js/TASK_BREAKDOWN.md` - Your tasks are 2.1-2.5 (~10 min)
3. `packages/less/src/less/less_go/js_eval_node.go` - Current implementation (~15 min)
4. `packages/less/src/less/less_go/nodejs_plugin_bridge.go` - How to access runtime (~10 min)

## Background Context

### Current State

The `EvaluateJavaScript()` method in `js_eval_node.go`:
1. Checks if `javascriptEnabled` is true ✅
2. Handles variable interpolation `@{varName}` ✅
3. **Returns error**: "JavaScript evaluation is not supported in the Go port" ❌

### What You're Changing

Replace step 3 with:
1. Get Node.js runtime from evaluation context
2. Build variable context for `this.foo.toJS()` access
3. Send `evalJS` command to Node.js
4. Process the response and return appropriate Go type

## Your Tasks

### Task 2.1: Access Node.js Runtime

**File**: `packages/less/src/less/less_go/js_eval_node.go`

Add import for runtime package (if not already present):
```go
import (
    // ... existing imports ...
    "github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)
```

**Modify `EvaluateJavaScript` function** (starting around line 102).

After the `javascriptEnabled` check and variable interpolation, replace the error return with runtime access:

```go
// Get Node.js runtime from context
var rt *runtime.NodeJSRuntime

// Try *Eval context first (most common)
if evalCtx, ok := context.(*Eval); ok {
    if evalCtx.PluginBridge != nil {
        rt = evalCtx.PluginBridge.GetRuntime()
    } else if evalCtx.LazyPluginBridge != nil {
        rt = evalCtx.LazyPluginBridge.GetRuntime()
    }
}

// Try map context (used in some evaluation paths)
if rt == nil {
    if mapCtx, ok := context.(map[string]any); ok {
        if bridge, ok := mapCtx["pluginBridge"].(*NodeJSPluginBridge); ok {
            rt = bridge.GetRuntime()
        } else if lazyBridge, ok := mapCtx["pluginBridge"].(*LazyNodeJSPluginBridge); ok {
            rt = lazyBridge.GetRuntime()
        }
    }
}

// Check wrapped context
if rt == nil {
    if evalCtx, ok := wrappedContext.ctx.(*Eval); ok {
        if evalCtx.PluginBridge != nil {
            rt = evalCtx.PluginBridge.GetRuntime()
        } else if evalCtx.LazyPluginBridge != nil {
            rt = evalCtx.LazyPluginBridge.GetRuntime()
        }
    }
}

if rt == nil {
    return nil, &LessError{
        Type:     "JavaScript",
        Message:  "JavaScript runtime not available. Ensure plugins are enabled.",
        Filename: getFilename(),
        Index:    j.GetIndex(),
    }
}
```

---

### Task 2.2: Build Variable Context

Add a helper method to build the variable context for Node.js.

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
    } else if mapCtx, ok := context.(map[string]any); ok {
        if f, ok := mapCtx["frames"].([]ParserFrame); ok {
            frames = f
        }
    }

    if len(frames) == 0 {
        return variables
    }

    // Get variables from the first frame (current scope)
    frame := frames[0]
    varsMap := frame.Variables()

    if varsMap == nil {
        return variables
    }

    for name, decl := range varsMap {
        // Try to get the CSS value of the variable
        var cssValue string

        // Check if it's a declaration with a value
        if declNode, ok := decl.(interface{ Value() any }); ok {
            value := declNode.Value()
            cssValue = j.jsify(value)
        } else if evalable, ok := decl.(interface{ Eval(any) (any, error) }); ok {
            // Try to evaluate it
            result, err := evalable.Eval(context)
            if err == nil {
                cssValue = j.jsify(result)
            }
        } else {
            // Last resort: use jsify directly
            cssValue = j.jsify(decl)
        }

        // Store with @ prefix removed (Go stores with @, JS expects without)
        cleanName := name
        if strings.HasPrefix(name, "@") {
            cleanName = name[1:]
        }

        variables[cleanName] = map[string]any{
            "value": cssValue,
        }
    }

    return variables
}
```

---

### Task 2.3: Send Command to Node.js

Replace the error return with the command:

```go
// Build variable context for this.varName.toJS() access
varContext := j.buildVariableContext(context)

// Send evalJS command to Node.js
resp, err := rt.SendCommand(runtime.Command{
    Cmd: "evalJS",
    Data: map[string]any{
        "expression": expressionForError, // The interpolated expression
        "variables":  varContext,
    },
})

if err != nil {
    return nil, &LessError{
        Type:     "JavaScript",
        Message:  fmt.Sprintf("JavaScript evaluation failed: %v", err),
        Filename: getFilename(),
        Index:    j.GetIndex(),
    }
}

if !resp.Success {
    // Error from Node.js (syntax error, runtime error, etc.)
    return nil, &LessError{
        Type:     "JavaScript",
        Message:  resp.Error,
        Filename: getFilename(),
        Index:    j.GetIndex(),
    }
}

// Process the successful result
return j.processJSResult(resp.Result)
```

---

### Task 2.4: Process JavaScript Result

Add a helper method to convert the Node.js response to Go types:

```go
// processJSResult converts the JavaScript result to the appropriate Go type
// The JavaScript side sends: { type: 'number'|'string'|'array'|'boolean'|'empty', value: ... }
func (j *JsEvalNode) processJSResult(result any) (any, error) {
    // Handle nil result
    if result == nil {
        return nil, nil
    }

    // Parse the result map
    resultMap, ok := result.(map[string]any)
    if !ok {
        // If it's not a map, return as-is (shouldn't happen with proper JS side)
        return result, nil
    }

    resultType, _ := resultMap["type"].(string)
    value := resultMap["value"]

    switch resultType {
    case "number":
        // JavaScript numbers come as float64
        if numVal, ok := value.(float64); ok {
            return numVal, nil
        }
        // Handle int (shouldn't happen but just in case)
        if intVal, ok := value.(int); ok {
            return float64(intVal), nil
        }
        return value, nil

    case "string":
        if strVal, ok := value.(string); ok {
            return strVal, nil
        }
        return fmt.Sprintf("%v", value), nil

    case "array":
        // Arrays are pre-joined by JavaScript side
        if strVal, ok := value.(string); ok {
            return strVal, nil
        }
        return fmt.Sprintf("%v", value), nil

    case "boolean":
        if boolVal, ok := value.(bool); ok {
            return boolVal, nil
        }
        return value, nil

    case "empty":
        return "", nil

    case "other":
        if strVal, ok := value.(string); ok {
            return strVal, nil
        }
        return fmt.Sprintf("%v", value), nil

    default:
        // Unknown type, return as-is
        return value, nil
    }
}
```

---

### Task 2.5: Verify Escaped Expressions

The `JavaScript` struct in `javascript.go` handles the `escaped` flag. Verify this integration works:

**Check `javascript.go` line 66-67**:
```go
case string:
    return NewQuoted(`"`+v+`"`, v, j.escaped, j.GetIndex(), j.FileInfo()), nil
```

The `j.escaped` flag should be passed correctly. This should work without modification.

**Test case**:
- `` `"hello"` `` → `"hello"` (quoted)
- `` ~`"hello"` `` → `hello` (unquoted/escaped)

---

## Complete Modified Function

Here's a template for the complete modified `EvaluateJavaScript` function:

```go
func (j *JsEvalNode) EvaluateJavaScript(expression string, context any) (any, error) {
    // Wrap the context to implement EvalContext
    wrappedContext := &contextWrapper{ctx: context}

    // Check if JavaScript is enabled
    javascriptEnabled := false
    if evalCtx, ok := context.(map[string]any); ok {
        if jsEnabled, ok := evalCtx["javascriptEnabled"].(bool); ok {
            javascriptEnabled = jsEnabled
        }
    } else if jsCtx, ok := context.(interface{ IsJavaScriptEnabled() bool }); ok {
        javascriptEnabled = jsCtx.IsJavaScriptEnabled()
    } else if evalCtx, ok := context.(*Eval); ok {
        javascriptEnabled = evalCtx.JavascriptEnabled
    }

    // Helper function to get filename safely
    getFilename := func() string {
        info := j.FileInfo()
        if info != nil {
            if filename, ok := info["filename"].(string); ok {
                return filename
            }
        }
        return "<unknown>"
    }

    if !javascriptEnabled {
        return nil, &LessError{
            Type:     "JavaScript",
            Message:  "inline JavaScript is not enabled. Is it set in your options?",
            Filename: getFilename(),
            Index:    j.GetIndex(),
        }
    }

    // Replace Less variables with their values
    var varEvalError error
    expressionForError := reVariableAtBrace.ReplaceAllStringFunc(expression, func(match string) string {
        if varEvalError != nil {
            return match
        }
        varName := match[2 : len(match)-1]
        variable := NewVariable("@"+varName, j.GetIndex(), j.FileInfo())
        result, err := variable.Eval(wrappedContext)
        if err != nil {
            varEvalError = err
            return match
        }
        return j.jsify(result)
    })

    if varEvalError != nil {
        if lessErr, ok := varEvalError.(*LessError); ok {
            return nil, &LessError{
                Type:     "JavaScript",
                Message:  lessErr.Message,
                Filename: lessErr.Filename,
                Index:    lessErr.Index,
                Line:     lessErr.Line,
                Column:   lessErr.Column,
            }
        }
        return nil, &LessError{
            Type:     "JavaScript",
            Message:  varEvalError.Error(),
            Filename: getFilename(),
            Index:    j.GetIndex(),
        }
    }

    // === NEW CODE: Get Node.js runtime ===
    var rt *runtime.NodeJSRuntime

    if evalCtx, ok := context.(*Eval); ok {
        if evalCtx.PluginBridge != nil {
            rt = evalCtx.PluginBridge.GetRuntime()
        } else if evalCtx.LazyPluginBridge != nil {
            rt = evalCtx.LazyPluginBridge.GetRuntime()
        }
    }

    if rt == nil {
        if mapCtx, ok := context.(map[string]any); ok {
            if bridge, ok := mapCtx["pluginBridge"].(*NodeJSPluginBridge); ok {
                rt = bridge.GetRuntime()
            } else if lazyBridge, ok := mapCtx["pluginBridge"].(*LazyNodeJSPluginBridge); ok {
                rt = lazyBridge.GetRuntime()
            }
        }
    }

    if rt == nil {
        if evalCtx, ok := wrappedContext.ctx.(*Eval); ok {
            if evalCtx.PluginBridge != nil {
                rt = evalCtx.PluginBridge.GetRuntime()
            } else if evalCtx.LazyPluginBridge != nil {
                rt = evalCtx.LazyPluginBridge.GetRuntime()
            }
        }
    }

    if rt == nil {
        return nil, &LessError{
            Type:     "JavaScript",
            Message:  "JavaScript runtime not available. Ensure plugins are enabled.",
            Filename: getFilename(),
            Index:    j.GetIndex(),
        }
    }

    // Build variable context for this.varName.toJS() access
    varContext := j.buildVariableContext(context)

    // Send evalJS command to Node.js
    resp, err := rt.SendCommand(runtime.Command{
        Cmd: "evalJS",
        Data: map[string]any{
            "expression": expressionForError,
            "variables":  varContext,
        },
    })

    if err != nil {
        return nil, &LessError{
            Type:     "JavaScript",
            Message:  fmt.Sprintf("JavaScript evaluation failed: %v", err),
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

    return j.processJSResult(resp.Result)
}
```

---

## Verification Checklist

Before marking complete:

- [ ] Added import for `runtime` package (if needed)
- [ ] Can access runtime from `*Eval` context
- [ ] Can access runtime from `map[string]any` context
- [ ] Implemented `buildVariableContext` method
- [ ] Implemented `processJSResult` method
- [ ] Error handling preserves file/line info
- [ ] Unit tests still pass: `pnpm -w test:go:unit`
- [ ] No regressions in integration tests

---

## Test Commands

```bash
# Run unit tests (should still pass)
pnpm -w test:go:unit

# Check for compile errors
go build ./packages/less/src/less/less_go/...

# Test specific file compiles
go build ./packages/less/src/less/less_go/js_eval_node.go
```

---

## Important Notes

1. **Import Path**: The runtime package is at `github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime`

2. **LazyNodeJSPluginBridge**: This is a deferred initialization wrapper. Check if it has a `GetRuntime()` method. If not, you may need to add one or access the inner bridge.

3. **Variable Evaluation**: The `buildVariableContext` tries multiple approaches because variables can be stored in different formats depending on the evaluation context.

4. **Error Type**: Always use `Type: "JavaScript"` for errors so they propagate correctly through `SafeEval`.

---

## Deliverables

When complete, provide:

1. **Summary**: What you implemented (2-3 sentences)
2. **Files modified**: List with brief description
3. **Test results**: Output of `pnpm -w test:go:unit` (should pass)
4. **Issues encountered**: Any blockers or surprises

Good luck! You're connecting the Go evaluation pipeline to Node.js. 🚀
