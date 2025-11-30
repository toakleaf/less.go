package less_go

import (
	"fmt"
	"math"
	"os"
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
	// Check if JavaScript is enabled
	javascriptEnabled := false
	debugMode := os.Getenv("LESS_GO_DEBUG") == "1"

	// Unwrap contextWrapper recursively if needed (can be nested)
	actualContext := context
	for {
		if wrapper, ok := actualContext.(*contextWrapper); ok {
			actualContext = wrapper.ctx
		} else {
			break
		}
	}

	if evalCtx, ok := actualContext.(map[string]any); ok {
		if jsEnabled, ok := evalCtx["javascriptEnabled"].(bool); ok {
			javascriptEnabled = jsEnabled
		}
		if debugMode && !javascriptEnabled {
			fmt.Printf("[DEBUG JsEvalNode] map[string]any context, javascriptEnabled not found or false\n")
		}
	} else if jsCtx, ok := actualContext.(interface{ IsJavaScriptEnabled() bool }); ok {
		javascriptEnabled = jsCtx.IsJavaScriptEnabled()
		if debugMode && !javascriptEnabled {
			fmt.Printf("[DEBUG JsEvalNode] interface context, IsJavaScriptEnabled()=false\n")
		}
	} else if evalCtx, ok := actualContext.(*Eval); ok {
		javascriptEnabled = evalCtx.JavascriptEnabled
		if debugMode && !javascriptEnabled {
			fmt.Printf("[DEBUG JsEvalNode] *Eval context, JavascriptEnabled=%v, PluginBridge=%v, LazyPluginBridge=%v\n", evalCtx.JavascriptEnabled, evalCtx.PluginBridge, evalCtx.LazyPluginBridge)
		}
	} else if debugMode {
		fmt.Printf("[DEBUG JsEvalNode] Unknown context type: %T\n", actualContext)
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
		// Evaluate variable using the unwrapped actual context
		result, err := variable.Eval(actualContext)
		if err != nil {
			// Capture the error - this is likely an undefined variable
			if debugMode {
				fmt.Printf("[DEBUG JsEvalNode] Variable @%s evaluation failed: %v\n", varName, err)
			}
			varEvalError = err
			return match // Keep original on error
		}
		jsResult := j.jsify(result)
		if debugMode {
			fmt.Printf("[DEBUG JsEvalNode] Variable @%s evaluated to: %v, jsified to: %s\n", varName, result, jsResult)
		}
		return jsResult
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

	// Get Node.js runtime from context (use actualContext which is unwrapped)
	var rt *runtime.NodeJSRuntime

	// Helper function to get runtime from a lazy bridge (initializes if needed)
	getRuntimeFromLazyBridge := func(lazyBridge *LazyNodeJSPluginBridge) *runtime.NodeJSRuntime {
		// GetBridge() initializes the bridge if not already done
		bridge, err := lazyBridge.GetBridge()
		if err != nil {
			return nil
		}
		return bridge.GetRuntime()
	}

	// Try *Eval context first (most common) - use actualContext which is unwrapped
	if evalCtx, ok := actualContext.(*Eval); ok {
		if evalCtx.PluginBridge != nil {
			rt = evalCtx.PluginBridge.GetRuntime()
		} else if evalCtx.LazyPluginBridge != nil {
			rt = getRuntimeFromLazyBridge(evalCtx.LazyPluginBridge)
		}
	}

	// Try map context (used in some evaluation paths)
	if rt == nil {
		if mapCtx, ok := actualContext.(map[string]any); ok {
			if bridge, ok := mapCtx["pluginBridge"].(*NodeJSPluginBridge); ok {
				rt = bridge.GetRuntime()
			} else if lazyBridge, ok := mapCtx["pluginBridge"].(*LazyNodeJSPluginBridge); ok {
				rt = getRuntimeFromLazyBridge(lazyBridge)
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

	if debugMode {
		fmt.Printf("[DEBUG JsEvalNode] Sending expression to Node.js: %s\n", expressionForError)
	}

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

// jsify converts Less values to JavaScript-compatible string representation.
// This matches the less.js jsify function behavior:
//   - If obj.value is an array with length > 1: wrap in [] and call toCSS() on each element
//   - Otherwise: call toCSS() on the whole object
func (j *JsEvalNode) jsify(obj any) string {
	if obj == nil {
		return "null"
	}

	// Helper function to call toCSS on an item
	toCSS := func(item any) string {
		if cssableAny, ok := item.(interface{ ToCSS(any) string }); ok {
			return cssableAny.ToCSS(nil)
		}
		if cssableSimple, ok := item.(interface{ ToCSS() string }); ok {
			return cssableSimple.ToCSS()
		}
		// Fallback for primitive types
		switch v := item.(type) {
		case string:
			return v
		case float64:
			if math.IsNaN(v) {
				return "NaN"
			}
			return strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			return strconv.Itoa(v)
		case bool:
			return fmt.Sprintf("%t", v)
		case nil:
			return "null"
		default:
			return fmt.Sprintf("%v", v)
		}
	}

	// If obj is a map containing a "value" key, extract it (for compatibility)
	if mapVal, ok := obj.(map[string]any); ok {
		if val, exists := mapVal["value"]; exists {
			obj = val
		}
	}

	// Match less.js: check if obj has a value array with length > 1
	// For Value and Expression types, wrap in [] and call toCSS on each element
	switch v := obj.(type) {
	case *Value:
		if len(v.Value) > 1 {
			var parts []string
			for _, item := range v.Value {
				parts = append(parts, toCSS(item))
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return toCSS(obj)
	case *Expression:
		if len(v.Value) > 1 {
			var parts []string
			for _, item := range v.Value {
				parts = append(parts, toCSS(item))
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return toCSS(obj)
	case []any:
		if len(v) > 1 {
			var parts []string
			for _, item := range v {
				parts = append(parts, toCSS(item))
			}
			return "[" + strings.Join(parts, ", ") + "]"
		} else if len(v) == 1 {
			return toCSS(v[0])
		}
		return ""
	default:
		return toCSS(obj)
	}
}

// buildVariableContext extracts variables from the evaluation context
// for access via this.varName.toJS() in JavaScript
func (j *JsEvalNode) buildVariableContext(context any) map[string]map[string]any {
	variables := make(map[string]map[string]any)
	debugMode := os.Getenv("LESS_GO_DEBUG") == "1"

	// Get frames from context - less.js uses context.frames[0].variables()
	// frames[0] should be the current/innermost ruleset
	var frames []ParserFrame

	if evalCtx, ok := context.(*Eval); ok {
		frames = evalCtx.GetFrames()
	} else if wrapper, ok := context.(*contextWrapper); ok {
		frames = wrapper.GetFrames()
	} else if mapCtx, ok := context.(map[string]any); ok {
		// Try to get frames as []ParserFrame
		if f, ok := mapCtx["frames"].([]ParserFrame); ok {
			frames = f
		} else if f, ok := mapCtx["frames"].([]any); ok {
			// Convert []any to []ParserFrame
			for _, frame := range f {
				if pf, ok := frame.(ParserFrame); ok {
					frames = append(frames, pf)
				}
			}
		}
	}

	if len(frames) == 0 {
		return variables
	}

	// Collect variables from all frames (inner scopes first)
	// We keep the first definition of each variable (closest scope wins)
	for _, frame := range frames {
		// First try the Variables() method (for cached variables)
		variablesProvider, ok := frame.(interface{ Variables() map[string]any })
		if ok {
			varsMap := variablesProvider.Variables()
			for name, decl := range varsMap {
				cleanName := name
				if strings.HasPrefix(name, "@") {
					cleanName = name[1:]
				}
				if _, exists := variables[cleanName]; exists {
					continue
				}
				cssValue := j.extractVariableValue(decl, debugMode, cleanName)
				if cssValue != "" {
					variables[cleanName] = map[string]any{"value": cssValue}
				}
			}
		}

		// Also directly scan Rules for variable declarations (in case Variables() cache isn't populated)
		if ruleset, ok := frame.(*Ruleset); ok {
			for _, rule := range ruleset.Rules {
				if decl, ok := rule.(*Declaration); ok && decl.variable {
					name, ok := decl.name.(string)
					if !ok {
						continue
					}
					cleanName := name
					if strings.HasPrefix(name, "@") {
						cleanName = name[1:]
					}
					if _, exists := variables[cleanName]; exists {
						continue
					}
					cssValue := j.extractVariableValue(decl, debugMode, cleanName)
					if cssValue != "" {
						variables[cleanName] = map[string]any{"value": cssValue}
					}
				}
			}
		}
	}

	return variables
}

// extractVariableValue extracts the CSS value from a declaration, avoiding JavaScript values
func (j *JsEvalNode) extractVariableValue(decl any, debugMode bool, cleanName string) string {
	var cssValue string

	// Check if it's a Declaration with a Value field
	if declStruct, ok := decl.(*Declaration); ok {
		if declStruct.Value != nil {
			// Check for JavaScript values in the Declaration's Value
			hasJS := false
			for _, valItem := range declStruct.Value.Value {
				if _, isJS := valItem.(*JavaScript); isJS {
					hasJS = true
					break
				}
			}
			if hasJS {
				return "" // Skip JavaScript values to prevent infinite recursion
			}
			cssValue = j.jsify(declStruct.Value)
		} else {
			return ""
		}
	} else if declNode, ok := decl.(interface{ Value() any }); ok {
		// Fallback: Check if it's an interface with Value() method
		value := declNode.Value()
		// Skip JavaScript values to prevent infinite recursion
		if _, isJS := value.(*JavaScript); isJS {
			return ""
		}
		cssValue = j.jsify(value)
	} else {
		// Don't try to evaluate - just use jsify directly to avoid recursion
		cssValue = j.jsify(decl)
	}

	if debugMode && cssValue != "" {
		fmt.Printf("[DEBUG buildVariableContext] Found variable %s = %s\n", cleanName, cssValue)
	}
	return cssValue
}

// JSArrayResult wraps a pre-joined array result from JavaScript
// This allows JavaScript.Eval to distinguish between string and array results
type JSArrayResult struct {
	Value string
}

// JSEmptyResult represents an empty JavaScript result (e.g., undefined, NaN)
type JSEmptyResult struct{}

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
		// Return JSArrayResult so JavaScript.Eval can create Anonymous instead of Quoted
		if strVal, ok := value.(string); ok {
			return &JSArrayResult{Value: strVal}, nil
		}
		return &JSArrayResult{Value: fmt.Sprintf("%v", value)}, nil

	case "boolean":
		if boolVal, ok := value.(bool); ok {
			return boolVal, nil
		}
		return value, nil

	case "empty":
		// Return JSEmptyResult so JavaScript.Eval can create an empty Anonymous
		return &JSEmptyResult{}, nil

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