package less_go

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// JsEvalNode represents a JavaScript evaluation node in the Less AST
type JsEvalNode struct {
	*Node
}

// NewJsEvalNode creates a new JsEvalNode instance
func NewJsEvalNode() *JsEvalNode {
	return &JsEvalNode{
		Node: NewNode(),
	}
}

// Type returns the node type
func (j *JsEvalNode) Type() string {
	return "JsEvalNode"
}

// GetType returns the node type
func (j *JsEvalNode) GetType() string {
	return "JsEvalNode"
}

// contextWrapper wraps an arbitrary context to implement the EvalContext interface
type contextWrapper struct {
	ctx any
}

func (w *contextWrapper) IsMathOn() bool {
	if mathCtx, ok := w.ctx.(interface{ IsMathOn() bool }); ok {
		return mathCtx.IsMathOn()
	}
	return true // Default to true
}

func (w *contextWrapper) SetMathOn(on bool) {
	if mathCtx, ok := w.ctx.(interface{ SetMathOn(bool) }); ok {
		mathCtx.SetMathOn(on)
	}
}

func (w *contextWrapper) IsInCalc() bool {
	if calcCtx, ok := w.ctx.(interface{ IsInCalc() bool }); ok {
		return calcCtx.IsInCalc()
	}
	return false // Default to false
}

func (w *contextWrapper) EnterCalc() {
	if calcCtx, ok := w.ctx.(interface{ EnterCalc() }); ok {
		calcCtx.EnterCalc()
	}
}

func (w *contextWrapper) ExitCalc() {
	if calcCtx, ok := w.ctx.(interface{ ExitCalc() }); ok {
		calcCtx.ExitCalc()
	}
}

func (w *contextWrapper) GetFrames() []ParserFrame {
	if framesCtx, ok := w.ctx.(interface{ GetFrames() []ParserFrame }); ok {
		return framesCtx.GetFrames()
	}
	
	// Try to get frames from map context
	if mapCtx, ok := w.ctx.(map[string]any); ok {
		if frames, ok := mapCtx["frames"].([]ParserFrame); ok {
			return frames
		}
	}
	
	return nil // Return nil if no frames are found
}

func (w *contextWrapper) GetImportantScope() []map[string]bool {
	if scopeCtx, ok := w.ctx.(interface{ GetImportantScope() []map[string]bool }); ok {
		return scopeCtx.GetImportantScope()
	}
	return nil
}

func (w *contextWrapper) GetDefaultFunc() *DefaultFunc {
	if defaultCtx, ok := w.ctx.(interface{ GetDefaultFunc() *DefaultFunc }); ok {
		return defaultCtx.GetDefaultFunc()
	}
	return nil
}

// EvaluateJavaScript evaluates JavaScript expressions.
// Since JavaScript evaluation is not supported in the Go port,
// this will always return an error if JavaScript is enabled,
// or a "not enabled" error if JavaScript is disabled.
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
		// Return a JavaScript-type error so it propagates through SafeEval
		return nil, &LessError{
			Type:     "JavaScript",
			Message:  "inline JavaScript is not enabled. Is it set in your options?",
			Filename: getFilename(),
			Index:    j.GetIndex(),
		}
	}

	// Replace Less variables with their values for better error messages
	// Track the first error from variable evaluation
	var varEvalError error
	expressionForError := reVariableAtBrace.ReplaceAllStringFunc(expression, func(match string) string {
		// If we already have an error, just return the match
		if varEvalError != nil {
			return match
		}
		// Extract variable name without @ and {}
		varName := match[2 : len(match)-1]
		// Create a Variable node
		variable := NewVariable("@"+varName, j.GetIndex(), j.FileInfo())
		// Evaluate variable
		result, err := variable.Eval(wrappedContext)
		if err != nil {
			// Capture the error - this is likely an undefined variable
			varEvalError = err
			return match // Keep original on error
		}
		return j.jsify(result)
	})

	// If there was a variable evaluation error (e.g., undefined variable), wrap it
	// as a JavaScript error so it propagates through SafeEval
	// This matches JavaScript behavior where undefined variables in JS expressions
	// throw a NameError
	if varEvalError != nil {
		// Wrap the error as a JavaScript-type error so SafeEval propagates it
		if lessErr, ok := varEvalError.(*LessError); ok {
			// Convert to JavaScript type while preserving the original message and info
			return nil, &LessError{
				Type:     "JavaScript",
				Message:  lessErr.Message,
				Filename: lessErr.Filename,
				Index:    lessErr.Index,
				Line:     lessErr.Line,
				Column:   lessErr.Column,
			}
		}
		// For non-LessError, wrap it as a JavaScript error
		return nil, &LessError{
			Type:     "JavaScript",
			Message:  varEvalError.Error(),
			Filename: getFilename(),
			Index:    j.GetIndex(),
		}
	}

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
}

// jsify converts Less values to a simple string representation suitable for error messages.
func (j *JsEvalNode) jsify(obj any) string {
	if obj == nil {
		return "null"
	}

	// Check for Node types and get their value
	if node, ok := obj.(*Node); ok {
		obj = node.Value // Use the actual value within the node
	}

	// If obj is a map containing a "value" key, extract it
	if mapVal, ok := obj.(map[string]any); ok {
		if val, exists := mapVal["value"]; exists {
			obj = val
		}
	}

	// Handle specific types
	switch v := obj.(type) {
	case string:
		return v // Return string directly
	case float64:
		if math.IsNaN(v) {
			return "NaN"
		}
		// Format float without unnecessary trailing zeros
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "null"
	case *Quoted:
		return v.value // Return the raw string content
	case *Dimension:
		return v.ToCSS(nil) // Use the CSS representation
	case *Color:
		return v.ToCSS(nil) // Use the CSS representation
	case *Anonymous:
		// If Anonymous contains a simple type, stringify that
		switch anonVal := v.Value.(type) {
		case string:
			return anonVal
		case float64:
			return strconv.FormatFloat(anonVal, 'f', -1, 64)
		case int:
			return strconv.Itoa(anonVal)
		case bool:
			return fmt.Sprintf("%t", anonVal)
		case nil:
			return "null"
		default:
			// Fallback for complex Anonymous values: use ToCSS
			return v.ToCSS(nil)
		}
	case []any:
		// Handle arrays recursively
		var parts []string
		for _, item := range v {
			parts = append(parts, j.jsify(item)) // Recursively call jsify
		}
		// Return comma-separated string wrapped in square brackets
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		// Fallback: Try ToCSS(any) first
		if cssableAny, ok := obj.(interface{ ToCSS(any) string }); ok {
			return cssableAny.ToCSS(nil)
		}
		// Then try ToCSS() for simpler mocks/types
		if cssableSimple, ok := obj.(interface{ ToCSS() string }); ok {
			return cssableSimple.ToCSS()
		}
		// Last resort: Use default Go formatting
		return fmt.Sprintf("%v", obj)
	}
}

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

	// Wrap context for variable evaluation
	wrappedContext := &contextWrapper{ctx: context}

	// Collect variables from all frames (inner scopes first)
	// We only keep the first definition of each variable (closest scope wins)
	for _, frame := range frames {
		// ParserFrame doesn't have Variables() method, but Ruleset and other frames do
		variablesProvider, ok := frame.(interface{ Variables() map[string]any })
		if !ok {
			continue
		}
		varsMap := variablesProvider.Variables()

		if varsMap == nil {
			continue
		}

		for name, decl := range varsMap {
			// Skip if we already have this variable from an inner scope
			cleanName := name
			if strings.HasPrefix(name, "@") {
				cleanName = name[1:]
			}

			if _, exists := variables[cleanName]; exists {
				continue
			}

			// Try to get the CSS value of the variable
			var cssValue string

			// Check if it's a declaration with a value
			if declNode, ok := decl.(interface{ Value() any }); ok {
				value := declNode.Value()
				cssValue = j.jsify(value)
			} else if evalable, ok := decl.(interface{ Eval(any) (any, error) }); ok {
				// Try to evaluate it
				result, err := evalable.Eval(wrappedContext)
				if err == nil {
					cssValue = j.jsify(result)
				}
			} else {
				// Last resort: use jsify directly
				cssValue = j.jsify(decl)
			}

			variables[cleanName] = map[string]any{
				"value": cssValue,
			}
		}
	}

	return variables
}

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