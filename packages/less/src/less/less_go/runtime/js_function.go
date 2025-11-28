package runtime

import (
	"encoding/json"
	"fmt"
	"sync"
)

// JSFunctionDefinition implements the FunctionDefinition interface for JavaScript functions.
// It calls JavaScript functions registered by plugins via the Node.js runtime.
type JSFunctionDefinition struct {
	name    string
	runtime *NodeJSRuntime
}

// NewJSFunctionDefinition creates a new JSFunctionDefinition.
func NewJSFunctionDefinition(name string, runtime *NodeJSRuntime) *JSFunctionDefinition {
	return &JSFunctionDefinition{
		name:    name,
		runtime: runtime,
	}
}

// Name returns the function name.
func (jf *JSFunctionDefinition) Name() string {
	return jf.name
}

// NeedsEvalArgs returns true - JS functions always expect evaluated arguments.
func (jf *JSFunctionDefinition) NeedsEvalArgs() bool {
	return true
}

// Call calls the JavaScript function with the given arguments.
// Arguments are serialized to JSON for IPC transfer.
// Returns the result node or error.
func (jf *JSFunctionDefinition) Call(args ...any) (any, error) {
	if jf.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	// Serialize arguments for transfer
	serializedArgs, err := jf.serializeArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize arguments: %w", err)
	}

	// Call the function via Node.js runtime
	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name": jf.name,
			"args": serializedArgs,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("JavaScript function error: %s", resp.Error)
	}

	// Deserialize the result
	result, err := jf.deserializeResult(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize result: %w", err)
	}

	return result, nil
}

// CallCtx calls the JavaScript function with context.
// For JS functions, we ignore the context and just call Call.
func (jf *JSFunctionDefinition) CallCtx(ctx any, args ...any) (any, error) {
	// JS functions don't use Go context, so we just call Call
	return jf.Call(args...)
}

// serializeArgs serializes Go AST nodes to a format suitable for JSON transfer.
func (jf *JSFunctionDefinition) serializeArgs(args []any) ([]any, error) {
	serialized := make([]any, len(args))
	for i, arg := range args {
		serialized[i] = serializeNode(arg)
	}
	return serialized, nil
}

// deserializeResult deserializes a JavaScript node result back to a Go node.
func (jf *JSFunctionDefinition) deserializeResult(result any) (any, error) {
	if result == nil {
		return nil, nil
	}

	// Handle primitive results (numbers, strings, booleans)
	switch v := result.(type) {
	case float64, int, int64:
		// Return as-is for now (will be wrapped in Anonymous by caller)
		return v, nil
	case string:
		return v, nil
	case bool:
		return v, nil
	}

	// Handle node objects (maps from JSON)
	if nodeMap, ok := result.(map[string]any); ok {
		return deserializeNodeMap(nodeMap)
	}

	// Return as-is for other types
	return result, nil
}

// serializeNode serializes a Go AST node to a map for JSON transfer.
func serializeNode(node any) any {
	if node == nil {
		return nil
	}

	// Handle primitive types
	switch v := node.(type) {
	case string, float64, int, int64, bool:
		return v
	}

	// Try to serialize as a node with GetType method
	if typer, ok := node.(interface{ GetType() string }); ok {
		nodeType := typer.GetType()
		nodeMap := map[string]any{
			"_type": nodeType,
		}

		// Extract common node properties based on type
		switch nodeType {
		case "Dimension":
			if getter, ok := node.(interface{ GetValue() float64 }); ok {
				nodeMap["value"] = getter.GetValue()
			}
			if getter, ok := node.(interface{ GetUnit() any }); ok {
				nodeMap["unit"] = serializeUnit(getter.GetUnit())
			}
		case "Color":
			if getter, ok := node.(interface{ GetRGB() []float64 }); ok {
				nodeMap["rgb"] = getter.GetRGB()
			}
			if getter, ok := node.(interface{ GetAlpha() float64 }); ok {
				nodeMap["alpha"] = getter.GetAlpha()
			}
		case "Quoted":
			if getter, ok := node.(interface{ GetValue() string }); ok {
				nodeMap["value"] = getter.GetValue()
			}
			if getter, ok := node.(interface{ GetQuote() string }); ok {
				nodeMap["quote"] = getter.GetQuote()
			}
			if getter, ok := node.(interface{ GetEscaped() bool }); ok {
				nodeMap["escaped"] = getter.GetEscaped()
			}
		case "Keyword":
			if getter, ok := node.(interface{ GetValue() string }); ok {
				nodeMap["value"] = getter.GetValue()
			}
		case "Anonymous":
			if getter, ok := node.(interface{ GetValue() any }); ok {
				nodeMap["value"] = getter.GetValue()
			}
		case "Expression", "Value":
			if getter, ok := node.(interface{ GetValue() []any }); ok {
				vals := getter.GetValue()
				serialized := make([]any, len(vals))
				for i, v := range vals {
					serialized[i] = serializeNode(v)
				}
				nodeMap["value"] = serialized
			}
		}

		return nodeMap
	}

	// Fallback: try JSON marshaling
	data, err := json.Marshal(node)
	if err != nil {
		return fmt.Sprintf("%v", node)
	}
	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Sprintf("%v", node)
	}
	return result
}

// serializeUnit serializes a Unit to a string or map.
func serializeUnit(unit any) any {
	if unit == nil {
		return ""
	}
	if s, ok := unit.(string); ok {
		return s
	}
	if stringer, ok := unit.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%v", unit)
}

// deserializeNodeMap deserializes a JavaScript node map to a JSResultNode.
func deserializeNodeMap(nodeMap map[string]any) (any, error) {
	nodeType, ok := nodeMap["_type"].(string)
	if !ok {
		// Not a typed node, return as-is
		return nodeMap, nil
	}

	// Create a JSResultNode that can be used by the Go evaluator
	return &JSResultNode{
		NodeType:   nodeType,
		Properties: nodeMap,
	}, nil
}

// JSResultNode represents a result from a JavaScript function.
// It implements common node interfaces so it can be used by the Go evaluator.
type JSResultNode struct {
	NodeType   string
	Properties map[string]any
}

// GetType returns the node type.
func (n *JSResultNode) GetType() string {
	return n.NodeType
}

// GetValue returns the node's value property.
func (n *JSResultNode) GetValue() any {
	if v, ok := n.Properties["value"]; ok {
		return v
	}
	return nil
}

// ToCSS returns a CSS string representation.
func (n *JSResultNode) ToCSS() string {
	switch n.NodeType {
	case "Dimension":
		value := n.getFloat("value")
		unit := n.getString("unit")
		if unit == "" {
			return fmt.Sprintf("%g", value)
		}
		return fmt.Sprintf("%g%s", value, unit)
	case "Color":
		rgb := n.getRGBArray("rgb")
		alpha := n.getFloat("alpha")
		if alpha < 1.0 {
			return fmt.Sprintf("rgba(%d, %d, %d, %g)", int(rgb[0]), int(rgb[1]), int(rgb[2]), alpha)
		}
		return fmt.Sprintf("rgb(%d, %d, %d)", int(rgb[0]), int(rgb[1]), int(rgb[2]))
	case "Quoted":
		value := n.getString("value")
		quote := n.getString("quote")
		escaped := n.getBool("escaped")
		if escaped {
			return value
		}
		return quote + value + quote
	case "Keyword":
		return n.getString("value")
	case "Anonymous":
		if v := n.Properties["value"]; v != nil {
			return fmt.Sprintf("%v", v)
		}
		return ""
	default:
		if v := n.Properties["value"]; v != nil {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
}

// GenCSS generates CSS output for the node.
func (n *JSResultNode) GenCSS(context any, output interface {
	Add(string, any, any)
}) {
	output.Add(n.ToCSS(), nil, nil)
}

// Helper methods for property access

func (n *JSResultNode) getFloat(key string) float64 {
	if v, ok := n.Properties[key]; ok {
		switch f := v.(type) {
		case float64:
			return f
		case int:
			return float64(f)
		case int64:
			return float64(f)
		}
	}
	return 0
}

func (n *JSResultNode) getString(key string) string {
	if v, ok := n.Properties[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (n *JSResultNode) getBool(key string) bool {
	if v, ok := n.Properties[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (n *JSResultNode) getRGBArray(key string) []float64 {
	if v, ok := n.Properties[key]; ok {
		switch arr := v.(type) {
		case []any:
			result := make([]float64, len(arr))
			for i, val := range arr {
				switch f := val.(type) {
				case float64:
					result[i] = f
				case int:
					result[i] = float64(f)
				}
			}
			return result
		case []float64:
			return arr
		}
	}
	return []float64{0, 0, 0}
}

// PluginFunctionRegistry provides a unified interface for both built-in Go functions
// and JavaScript plugin functions.
type PluginFunctionRegistry struct {
	builtinRegistry any               // The built-in Go function registry
	jsRuntime       *NodeJSRuntime    // Node.js runtime for JS functions
	jsFunctions     map[string]*JSFunctionDefinition
	mu              sync.RWMutex
}

// NewPluginFunctionRegistry creates a new PluginFunctionRegistry.
func NewPluginFunctionRegistry(builtinRegistry any, runtime *NodeJSRuntime) *PluginFunctionRegistry {
	return &PluginFunctionRegistry{
		builtinRegistry: builtinRegistry,
		jsRuntime:       runtime,
		jsFunctions:     make(map[string]*JSFunctionDefinition),
	}
}

// RegisterJSFunction registers a JavaScript function by name.
func (r *PluginFunctionRegistry) RegisterJSFunction(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jsFunctions[name] = NewJSFunctionDefinition(name, r.jsRuntime)
}

// RegisterJSFunctions registers multiple JavaScript functions by name.
func (r *PluginFunctionRegistry) RegisterJSFunctions(names []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, name := range names {
		r.jsFunctions[name] = NewJSFunctionDefinition(name, r.jsRuntime)
	}
}

// Get retrieves a function definition by name.
// JavaScript functions take precedence over built-in functions (shadowing).
func (r *PluginFunctionRegistry) Get(name string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check JS functions first (allows shadowing built-ins)
	if jsFn, ok := r.jsFunctions[name]; ok {
		return jsFn
	}

	// Fall back to built-in registry
	if r.builtinRegistry != nil {
		if getter, ok := r.builtinRegistry.(interface{ Get(string) any }); ok {
			return getter.Get(name)
		}
	}

	return nil
}

// HasJSFunction checks if a JavaScript function is registered.
func (r *PluginFunctionRegistry) HasJSFunction(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.jsFunctions[name]
	return ok
}

// GetJSFunctionNames returns the names of all registered JavaScript functions.
func (r *PluginFunctionRegistry) GetJSFunctionNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.jsFunctions))
	for name := range r.jsFunctions {
		names = append(names, name)
	}
	return names
}

// GetBuiltinRegistry returns the underlying built-in registry.
func (r *PluginFunctionRegistry) GetBuiltinRegistry() any {
	return r.builtinRegistry
}

// ClearJSFunctions removes all registered JavaScript functions.
func (r *PluginFunctionRegistry) ClearJSFunctions() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jsFunctions = make(map[string]*JSFunctionDefinition)
}

// RefreshFromRuntime queries the Node.js runtime for registered functions
// and updates the registry.
func (r *PluginFunctionRegistry) RefreshFromRuntime() error {
	if r.jsRuntime == nil {
		return fmt.Errorf("Node.js runtime not initialized")
	}

	resp, err := r.jsRuntime.SendCommand(Command{
		Cmd: "getRegisteredFunctions",
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("failed to get registered functions: %s", resp.Error)
	}

	// Parse function names from response
	if funcs, ok := resp.Result.([]any); ok {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, f := range funcs {
			if name, ok := f.(string); ok {
				if _, exists := r.jsFunctions[name]; !exists {
					r.jsFunctions[name] = NewJSFunctionDefinition(name, r.jsRuntime)
				}
			}
		}
	}

	return nil
}
