package runtime

import (
	"encoding/json"
	"fmt"
	"sync"
)

// JSFunctionDefinition implements the FunctionDefinition interface for JavaScript functions.
// It calls JavaScript functions registered by plugins via the Node.js runtime.
type JSFunctionDefinition struct {
	name            string
	runtime         *NodeJSRuntime
	useSharedMemory bool // If true (default), use shared memory for zero-copy transfer
}

// JSFunctionOption configures a JSFunctionDefinition.
type JSFunctionOption func(*JSFunctionDefinition)

// WithJSONFallback disables shared memory and uses JSON IPC instead.
// This can be useful for debugging or when shared memory is unavailable.
func WithJSONFallback() JSFunctionOption {
	return func(jf *JSFunctionDefinition) {
		jf.useSharedMemory = false
	}
}

// NewJSFunctionDefinition creates a new JSFunctionDefinition.
// By default, shared memory is used for zero-copy argument and result transfer.
// Use WithJSONFallback() to disable shared memory and use JSON IPC instead.
func NewJSFunctionDefinition(name string, runtime *NodeJSRuntime, opts ...JSFunctionOption) *JSFunctionDefinition {
	jf := &JSFunctionDefinition{
		name:            name,
		runtime:         runtime,
		useSharedMemory: true, // Default to shared memory
	}
	for _, opt := range opts {
		opt(jf)
	}
	return jf
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
// By default, uses shared memory for zero-copy transfer.
// Falls back to JSON IPC when shared memory is disabled via WithJSONFallback().
// Returns the result node or error.
func (jf *JSFunctionDefinition) Call(args ...any) (any, error) {
	if jf.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	if jf.useSharedMemory {
		return jf.callViaSharedMemory(args...)
	}
	return jf.callViaJSON(args...)
}

// callViaJSON calls the JavaScript function using JSON serialization for IPC.
// This is the fallback when shared memory is disabled.
func (jf *JSFunctionDefinition) callViaJSON(args ...any) (any, error) {
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

// callViaSharedMemory calls the JavaScript function using shared memory for zero-copy transfer.
// Arguments are flattened to FlatAST format and written to shared memory.
// Node.js reads arguments directly from the buffer and writes results back.
func (jf *JSFunctionDefinition) callViaSharedMemory(args ...any) (any, error) {
	// If no arguments, use a simplified path
	if len(args) == 0 {
		return jf.callViaSharedMemoryNoArgs()
	}

	// 1. Flatten args to FlatAST format
	argsFlat := NewFlatAST()
	argIndices := make([]uint32, len(args))

	for i, arg := range args {
		if arg == nil {
			argIndices[i] = 0
			continue
		}

		flattener := NewASTFlattener()
		flattener.flat = argsFlat // Use the same flat AST for all args

		idx, err := flattener.FlattenNode(arg, 0)
		if err != nil {
			// Fall back to JSON for non-node arguments
			return jf.callViaJSON(args...)
		}
		argIndices[i] = idx
	}

	// Set the root index to 0 since we have multiple roots (one per arg)
	argsFlat.RootIndex = 0

	// 2. Write to shared memory buffer
	argsBytes, err := argsFlat.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize args: %w", err)
	}

	// Create shared memory segment with extra space for result
	// We allocate 2x the args size to leave room for the result
	bufferSize := len(argsBytes) * 2
	if bufferSize < 4096 {
		bufferSize = 4096 // Minimum buffer size
	}

	shm, err := jf.runtime.CreateSharedMemory(bufferSize)
	if err != nil {
		// Fall back to JSON if shared memory fails
		return jf.callViaJSON(args...)
	}
	defer jf.runtime.DestroySharedMemory(shm)

	if err := shm.WriteAll(argsBytes); err != nil {
		return nil, fmt.Errorf("failed to write to shared memory: %w", err)
	}

	// 3. Attach buffer to Node.js
	if err := jf.runtime.AttachBuffer(shm); err != nil {
		return nil, fmt.Errorf("failed to attach buffer: %w", err)
	}
	defer jf.runtime.DetachBuffer(shm.Key())

	// 4. Send command with buffer reference (not data)
	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunctionSharedMem",
		Data: map[string]any{
			"name":       jf.name,
			"bufferKey":  shm.Key(),
			"argIndices": argIndices,
			"argsSize":   len(argsBytes),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("JavaScript function error: %s", resp.Error)
	}

	// 5. Read result from response
	// The result can come back in two ways:
	// - As a JSON result (for simple types or when buffer writing fails)
	// - As a buffer offset (for complex nodes written to shared memory)
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		// Simple result, use JSON deserialization
		return jf.deserializeResult(resp.Result)
	}

	// Check if result is in shared memory
	if resultOffset, ok := resultMap["resultOffset"]; ok {
		offset := int(resultOffset.(float64))
		resultSize := int(resultMap["resultSize"].(float64))

		// Read the result from shared memory
		resultData, err := shm.Read(offset, resultSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read result from shared memory: %w", err)
		}

		// Parse the result FlatAST
		resultFlat, err := FromBytes(resultData)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize result: %w", err)
		}

		// Unflatten result to GenericNode
		result, err := UnflattenAST(resultFlat)
		if err != nil {
			return nil, fmt.Errorf("failed to unflatten result: %w", err)
		}

		return convertGenericNodeToResult(result, resultFlat), nil
	}

	// Result came back as JSON
	if jsonResult, ok := resultMap["jsonResult"]; ok {
		return jf.deserializeResult(jsonResult)
	}

	return jf.deserializeResult(resp.Result)
}

// callViaSharedMemoryNoArgs is an optimized path for functions with no arguments.
func (jf *JSFunctionDefinition) callViaSharedMemoryNoArgs() (any, error) {
	// Create a minimal shared memory buffer for the result
	bufferSize := 4096 // Should be enough for most results
	shm, err := jf.runtime.CreateSharedMemory(bufferSize)
	if err != nil {
		// Fall back to JSON if shared memory fails
		return jf.callViaJSON()
	}
	defer jf.runtime.DestroySharedMemory(shm)

	// Write an empty FlatAST as a placeholder
	emptyFlat := NewFlatAST()
	emptyBytes, _ := emptyFlat.ToBytes()
	if err := shm.WriteAll(emptyBytes); err != nil {
		return nil, fmt.Errorf("failed to write to shared memory: %w", err)
	}

	// Attach buffer to Node.js
	if err := jf.runtime.AttachBuffer(shm); err != nil {
		return nil, fmt.Errorf("failed to attach buffer: %w", err)
	}
	defer jf.runtime.DetachBuffer(shm.Key())

	// Send command with buffer reference
	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunctionSharedMem",
		Data: map[string]any{
			"name":       jf.name,
			"bufferKey":  shm.Key(),
			"argIndices": []uint32{},
			"argsSize":   len(emptyBytes),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("JavaScript function error: %s", resp.Error)
	}

	// Handle result
	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		return jf.deserializeResult(resp.Result)
	}

	if resultOffset, ok := resultMap["resultOffset"]; ok {
		offset := int(resultOffset.(float64))
		resultSize := int(resultMap["resultSize"].(float64))

		resultData, err := shm.Read(offset, resultSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read result from shared memory: %w", err)
		}

		resultFlat, err := FromBytes(resultData)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize result: %w", err)
		}

		result, err := UnflattenAST(resultFlat)
		if err != nil {
			return nil, fmt.Errorf("failed to unflatten result: %w", err)
		}

		return convertGenericNodeToResult(result, resultFlat), nil
	}

	if jsonResult, ok := resultMap["jsonResult"]; ok {
		return jf.deserializeResult(jsonResult)
	}

	return jf.deserializeResult(resp.Result)
}

// convertGenericNodeToResult converts a GenericNode to a JSResultNode.
// It resolves string indices using the FlatAST's string table.
func convertGenericNodeToResult(node *GenericNode, flat *FlatAST) any {
	if node == nil {
		return nil
	}

	// Convert GenericNode properties to JSResultNode format
	// Resolve string table indices to actual strings
	props := make(map[string]any)
	if node.Properties != nil {
		for k, v := range node.Properties {
			// Check if the value might be a string table index
			if idx, ok := v.(float64); ok {
				// Try to resolve as string index for known string properties
				if isStringProperty(node.Type, k) {
					if resolved := flat.GetString(uint32(idx)); resolved != "" {
						props[k] = resolved
						continue
					}
				}
			}
			props[k] = v
		}
	}

	// Add flags if set
	if node.Parens {
		props["parens"] = true
	}
	if node.ParensInOp {
		props["parensInOp"] = true
	}

	// Convert children recursively if needed
	if len(node.Children) > 0 {
		children := make([]any, len(node.Children))
		for i, child := range node.Children {
			children[i] = convertGenericNodeToResult(child, flat)
		}
		props["children"] = children
	}

	return &JSResultNode{
		NodeType:   node.Type,
		Properties: props,
	}
}

// isStringProperty returns true if the given property of the given node type
// is expected to be a string value (and thus might be stored as a string table index).
func isStringProperty(nodeType, propName string) bool {
	stringProps := map[string]map[string]bool{
		"Dimension": {"Unit": true, "unit": true},
		"Quoted":    {"Value": true, "value": true, "Quote": true, "quote": true},
		"Keyword":   {"Value": true, "value": true},
		"Anonymous": {"Value": true, "value": true},
		"Variable":  {"Name": true, "name": true},
		"URL":       {"Value": true, "value": true},
		"Call":      {"Name": true, "name": true},
		"Combinator": {"Value": true, "value": true},
		"Element":   {"Value": true, "value": true},
		"AtRule":    {"Name": true, "name": true},
		"Comment":   {"Value": true, "value": true},
		"Assignment": {"Key": true, "key": true},
		"Attribute": {"Key": true, "key": true, "Op": true, "op": true},
		"Operation": {"Op": true, "op": true},
		"Condition": {"Op": true, "op": true},
	}

	if props, ok := stringProps[nodeType]; ok {
		return props[propName]
	}
	return false
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
