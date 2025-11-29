package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"unsafe"
)

// ====================================================================================
// JS Function IPC Mode Configuration
// ====================================================================================
//
// JavaScript plugin functions can communicate with Go using two different IPC modes:
//
// 1. SHARED MEMORY MODE (default):
//    - Arguments and results are serialized to FlatAST binary format
//    - Data is written to a memory-mapped file shared between Go and Node.js
//    - Node.js reads arguments directly from the buffer (zero-copy on read)
//    - Results are written back to the same buffer
//    - Best for: Complex AST trees, large data structures, high-frequency calls
//
// 2. JSON MODE:
//    - Arguments and results are serialized to JSON
//    - Data is passed through stdio pipes between Go and Node.js
//    - Simpler implementation, easier to debug
//    - Best for: Simple function calls, debugging, environments without shared memory
//
// The mode can be controlled in three ways (in order of precedence):
//
// 1. Per-function option: NewJSFunctionDefinition("fn", rt, WithJSONMode())
// 2. Environment variable: LESS_JS_IPC_MODE=json or LESS_JS_IPC_MODE=sharedmem
// 3. Default: Shared memory mode
//
// Environment variable values:
//   - "json" or "JSON": Use JSON mode for all functions
//   - "sharedmem", "shm", or "shared": Use shared memory mode (default)
//
// ====================================================================================

// JSIPCMode represents the IPC mode for JS function calls.
type JSIPCMode int

const (
	// JSIPCModeSharedMemory uses shared memory for zero-copy data transfer.
	// Arguments are serialized to FlatAST format and written to a memory-mapped file.
	// This is the default mode.
	JSIPCModeSharedMemory JSIPCMode = iota

	// JSIPCModeJSON uses JSON serialization over stdio pipes.
	// Simpler but involves serialization/deserialization overhead.
	JSIPCModeJSON
)

// String returns a human-readable name for the IPC mode.
func (m JSIPCMode) String() string {
	switch m {
	case JSIPCModeSharedMemory:
		return "shared-memory"
	case JSIPCModeJSON:
		return "json"
	default:
		return "unknown"
	}
}

// getDefaultIPCMode returns the default IPC mode based on the LESS_JS_IPC_MODE
// environment variable. If not set or unrecognized, defaults to JSON mode.
// Note: Shared memory mode currently has a bug where complex nested objects
// (like DetachedRuleset with rules) are not fully transferred. JSON mode
// works reliably for all node types.
func getDefaultIPCMode() JSIPCMode {
	mode := os.Getenv("LESS_JS_IPC_MODE")
	switch mode {
	case "json", "JSON":
		return JSIPCModeJSON
	case "sharedmem", "shm", "shared", "SHM", "SHARED":
		return JSIPCModeSharedMemory
	default:
		// Default to JSON mode (shared memory has bugs with complex objects)
		return JSIPCModeJSON
	}
}

// JSFunctionDefinition implements the FunctionDefinition interface for JavaScript functions.
// It calls JavaScript functions registered by plugins via the Node.js runtime.
//
// The function supports two IPC modes for communicating with Node.js:
//   - Shared Memory: Zero-copy transfer using memory-mapped files (default)
//   - JSON: Traditional JSON serialization over stdio
//
// See the package-level documentation for details on configuring the IPC mode.
type JSFunctionDefinition struct {
	name    string
	runtime *NodeJSRuntime
	ipcMode JSIPCMode
}

// JSFunctionOption configures a JSFunctionDefinition.
type JSFunctionOption func(*JSFunctionDefinition)

// WithJSONMode configures the function to use JSON serialization for IPC.
// This mode serializes arguments and results as JSON, which is simpler
// but has serialization overhead compared to shared memory mode.
//
// Use this when:
//   - Debugging IPC issues (JSON is easier to inspect)
//   - Running in environments without shared memory support
//   - Working with simple function calls where overhead doesn't matter
func WithJSONMode() JSFunctionOption {
	return func(jf *JSFunctionDefinition) {
		jf.ipcMode = JSIPCModeJSON
	}
}

// WithSharedMemoryMode configures the function to use shared memory for IPC.
// This is the default mode but can be explicitly set to override environment
// variable configuration.
//
// This mode serializes arguments to FlatAST binary format and writes them
// to a memory-mapped file that Node.js can read directly.
//
// Use this when:
//   - Working with complex AST trees
//   - Performance is critical
//   - Making many function calls
func WithSharedMemoryMode() JSFunctionOption {
	return func(jf *JSFunctionDefinition) {
		jf.ipcMode = JSIPCModeSharedMemory
	}
}

// WithIPCMode configures the function to use the specified IPC mode.
// This allows programmatic control over the IPC mode.
func WithIPCMode(mode JSIPCMode) JSFunctionOption {
	return func(jf *JSFunctionDefinition) {
		jf.ipcMode = mode
	}
}

// NewJSFunctionDefinition creates a new JSFunctionDefinition for calling
// JavaScript functions registered by plugins.
//
// The default IPC mode is determined by:
//  1. Any options passed (WithJSONMode, WithSharedMemoryMode, WithIPCMode)
//  2. The LESS_JS_IPC_MODE environment variable
//  3. Shared memory mode (if nothing else is specified)
//
// Example usage:
//
//	// Use default mode (shared memory, or env var override)
//	fn := NewJSFunctionDefinition("myFunc", runtime)
//
//	// Explicitly use JSON mode
//	fn := NewJSFunctionDefinition("myFunc", runtime, WithJSONMode())
//
//	// Explicitly use shared memory mode
//	fn := NewJSFunctionDefinition("myFunc", runtime, WithSharedMemoryMode())
func NewJSFunctionDefinition(name string, runtime *NodeJSRuntime, opts ...JSFunctionOption) *JSFunctionDefinition {
	jf := &JSFunctionDefinition{
		name:    name,
		runtime: runtime,
		ipcMode: getDefaultIPCMode(), // Respects LESS_JS_IPC_MODE env var
	}
	// Options override the default/env var setting
	for _, opt := range opts {
		opt(jf)
	}
	return jf
}

// IPCMode returns the current IPC mode for this function.
func (jf *JSFunctionDefinition) IPCMode() JSIPCMode {
	return jf.ipcMode
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
//
// The IPC mode (shared memory or JSON) is determined by the function's
// configuration. See NewJSFunctionDefinition for details on mode selection.
//
// Returns the result node or error.
func (jf *JSFunctionDefinition) Call(args ...any) (any, error) {
	if jf.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	switch jf.ipcMode {
	case JSIPCModeSharedMemory:
		return jf.callViaSharedMemory(args...)
	case JSIPCModeJSON:
		return jf.callViaJSON(args...)
	default:
		// Default to shared memory if somehow an invalid mode is set
		return jf.callViaSharedMemory(args...)
	}
}

// callViaJSON calls the JavaScript function using JSON serialization for IPC.
//
// This mode:
//   - Serializes arguments to JSON format
//   - Sends data through the stdio pipe to Node.js
//   - Receives JSON response back through stdio
//   - Deserializes the response to Go types
//
// Pros: Simple, easy to debug, no shared memory setup
// Cons: Serialization overhead for large/complex data
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

// callViaSharedMemory calls the JavaScript function using shared memory for IPC.
//
// This mode:
//  1. Flattens arguments to FlatAST binary format
//  2. Creates a memory-mapped file (shared memory segment)
//  3. Writes the FlatAST data to the shared memory
//  4. Attaches the buffer to Node.js (sends only the file path, not the data)
//  5. Node.js reads arguments directly from the memory-mapped file
//  6. Node.js writes results back to the same buffer
//  7. Go reads the results from shared memory
//
// Pros: Zero-copy on read, efficient for complex AST trees
// Cons: Setup overhead, more complex implementation
//
// Note: If shared memory operations fail, this method automatically falls back
// to JSON mode to ensure the function call succeeds.
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

// EvalContextProvider is an interface for objects that can provide evaluation context
// for JavaScript plugin functions that need to access variables.
type EvalContextProvider interface {
	GetFramesAny() []any
	GetImportantScopeAny() []map[string]any
}

// CallWithContext calls the JavaScript function with evaluation context.
// This is used by plugin functions that need to access Less variables.
//
// OPTIMIZATION: Uses pre-fetch + on-demand lookup for optimal performance:
// 1. Pre-fetch commonly needed variables (avoiding IPC for them)
// 2. For any other variables, use on-demand callback lookup
func (jf *JSFunctionDefinition) CallWithContext(evalContext EvalContextProvider, args ...any) (any, error) {
	if jf.runtime == nil {
		return nil, fmt.Errorf("Node.js runtime not initialized")
	}

	// Use pre-fetch mode for known plugin functions to avoid IPC overhead
	return jf.callWithPrefetchContext(evalContext, args...)
}

// knownVariables contains variables that are commonly accessed by plugin functions.
// We pre-serialize these to avoid IPC round-trips for each lookup.
var knownVariables = []string{
	// Bootstrap theme colors and utilities
	"@theme-colors",
	"@theme-color-interval",
	"@black",
	"@white",
	"@gray-100",
	"@gray-200",
	"@gray-600",
	"@gray-800",
	"@gray-900",
	"@yiq-contrasted-threshold",
	"@yiq-text-dark",
	"@yiq-text-light",
	// Bootstrap grid breakpoints
	"@grid-breakpoints",
	// Color variants
	"@primary",
	"@secondary",
	"@success",
	"@info",
	"@warning",
	"@danger",
	"@light",
	"@dark",
}

// callWithPrefetchContext pre-fetches commonly needed variables and sends them
// with the function call using SHARED MEMORY in BINARY format.
// This avoids JSON serialization overhead which is critical for performance.
//
// The binary format is defined in binary_variables.go:
// - Header: magic + version + variable count
// - Each variable: name + important flag + type + binary-encoded value
//
// JavaScript reads directly from the memory-mapped file using DataView.
func (jf *JSFunctionDefinition) callWithPrefetchContext(evalContext EvalContextProvider, args ...any) (any, error) {
	// Serialize arguments for transfer (these are typically small)
	serializedArgs, err := jf.serializeArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize arguments: %w", err)
	}

	// Get frames for variable lookup
	frames := evalContext.GetFramesAny()
	importantScope := evalContext.GetImportantScopeAny()

	// Look up all known variables and collect their declarations
	varDecls := jf.collectPrefetchVariables(frames)

	// Write variables to binary format
	binaryData := WritePrefetchedVariables(varDecls)

	// Create shared memory buffer for the prefetched variables
	shmMgr := jf.runtime.SharedMemoryManager()
	if shmMgr == nil {
		// Fall back to JSON if shared memory not available
		return jf.callWithPrefetchContextJSON(evalContext, args...)
	}

	// Allocate buffer with some extra space
	bufferSize := len(binaryData) + 1024
	if bufferSize < 4096 {
		bufferSize = 4096
	}

	shm, err := shmMgr.Create(bufferSize)
	if err != nil {
		// Fall back to JSON if shared memory creation fails
		return jf.callWithPrefetchContextJSON(evalContext, args...)
	}
	defer shmMgr.Destroy(shm.Key())

	// Write binary data to shared memory
	if err := shm.WriteAll(binaryData); err != nil {
		return nil, fmt.Errorf("failed to write to shared memory: %w", err)
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[callWithPrefetchContext] Function %s: %d frames, %d prefetched vars, %d bytes binary\n",
			jf.name, len(frames), len(varDecls), len(binaryData))
	}

	// Register callback for any variables not in prefetch list
	// Uses binary format for the response as well
	jf.runtime.RegisterCallback("lookupVariable", func(data any) (any, error) {
		reqData, ok := data.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid variable lookup request")
		}

		varName, _ := reqData["name"].(string)
		frameIdx := 0
		if idx, ok := reqData["frameIndex"].(float64); ok {
			frameIdx = int(idx)
		}

		for i := frameIdx; i < len(frames); i++ {
			frame := frames[i]
			if frame == nil {
				continue
			}

			variablesProvider, ok := frame.(interface{ Variables() map[string]any })
			if !ok {
				continue
			}

			variables := variablesProvider.Variables()
			if variables == nil {
				continue
			}

			if decl, exists := variables[varName]; exists && decl != nil {
				// Write the variable to shared memory for binary transfer
				offset, written, err := writeVariableToSharedMemory(shm, decl)
				if err == nil {
					return map[string]any{
						"found":      true,
						"frameIndex": i,
						"shmOffset":  offset,
						"shmLength":  written,
					}, nil
				}
				// Fall back to JSON for complex types
				serializedDecl := jf.serializeVariableDeclaration(decl)
				if serializedDecl != nil {
					return map[string]any{
						"found":      true,
						"frameIndex": i,
						"value":      serializedDecl,
						"useJSON":    true,
					}, nil
				}
			}
		}

		return map[string]any{"found": false}, nil
	})
	defer jf.runtime.UnregisterCallback("lookupVariable")

	// Send context with shared memory reference (NOT JSON data)
	minimalContext := map[string]any{
		"frameCount":         len(frames),
		"importantScope":     serializeImportantScope(importantScope),
		"usePrefetch":        true,
		"useSharedMemory":    true,
		"prefetchBufferKey":  shm.Key(),
		"prefetchBufferPath": shm.Path(),
		"prefetchBufferSize": len(binaryData),
	}

	// Call the function via Node.js runtime
	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name":    jf.name,
			"args":    serializedArgs,
			"context": minimalContext,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("JavaScript function error: %s", resp.Error)
	}

	result, err := jf.deserializeResult(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize result: %w", err)
	}

	return result, nil
}

// callWithPrefetchContextJSON is the fallback when shared memory is not available.
// Uses JSON serialization for prefetched variables.
func (jf *JSFunctionDefinition) callWithPrefetchContextJSON(evalContext EvalContextProvider, args ...any) (any, error) {
	serializedArgs, err := jf.serializeArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize arguments: %w", err)
	}

	frames := evalContext.GetFramesAny()
	importantScope := evalContext.GetImportantScopeAny()

	// Pre-fetch known variables using JSON serialization
	prefetchedVars := jf.prefetchVariables(frames)

	minimalContext := map[string]any{
		"frameCount":     len(frames),
		"importantScope": serializeImportantScope(importantScope),
		"prefetchedVars": prefetchedVars,
		"usePrefetch":    true,
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[callWithPrefetchContextJSON] Function %s: %d frames, %d prefetched vars (JSON fallback)\n",
			jf.name, len(frames), len(prefetchedVars))
	}

	jf.runtime.RegisterCallback("lookupVariable", func(data any) (any, error) {
		reqData, ok := data.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid variable lookup request")
		}

		varName, _ := reqData["name"].(string)
		frameIdx := 0
		if idx, ok := reqData["frameIndex"].(float64); ok {
			frameIdx = int(idx)
		}

		for i := frameIdx; i < len(frames); i++ {
			frame := frames[i]
			if frame == nil {
				continue
			}

			variablesProvider, ok := frame.(interface{ Variables() map[string]any })
			if !ok {
				continue
			}

			variables := variablesProvider.Variables()
			if variables == nil {
				continue
			}

			if decl, exists := variables[varName]; exists && decl != nil {
				serializedDecl := jf.serializeVariableDeclaration(decl)
				if serializedDecl != nil {
					return map[string]any{
						"found":      true,
						"frameIndex": i,
						"value":      serializedDecl,
					}, nil
				}
			}
		}

		return map[string]any{"found": false}, nil
	})
	defer jf.runtime.UnregisterCallback("lookupVariable")

	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name":    jf.name,
			"args":    serializedArgs,
			"context": minimalContext,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("JavaScript function error: %s", resp.Error)
	}

	result, err := jf.deserializeResult(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize result: %w", err)
	}

	return result, nil
}

// collectPrefetchVariables looks up known commonly-used variables and returns their declarations.
// Unlike prefetchVariables, this returns raw declarations instead of serialized JSON.
func (jf *JSFunctionDefinition) collectPrefetchVariables(frames []any) map[string]any {
	collected := make(map[string]any)

	for _, varName := range knownVariables {
		for _, frame := range frames {
			if frame == nil {
				continue
			}

			variablesProvider, ok := frame.(interface{ Variables() map[string]any })
			if !ok {
				continue
			}

			variables := variablesProvider.Variables()
			if variables == nil {
				continue
			}

			if decl, exists := variables[varName]; exists && decl != nil {
				collected[varName] = decl
				break // Found it, don't look in more frames
			}
		}
	}

	return collected
}

// prefetchVariables looks up and serializes known commonly-used variables.
func (jf *JSFunctionDefinition) prefetchVariables(frames []any) map[string]any {
	prefetched := make(map[string]any)

	for _, varName := range knownVariables {
		// Look up the variable in frames
		for _, frame := range frames {
			if frame == nil {
				continue
			}

			variablesProvider, ok := frame.(interface{ Variables() map[string]any })
			if !ok {
				continue
			}

			variables := variablesProvider.Variables()
			if variables == nil {
				continue
			}

			if decl, exists := variables[varName]; exists && decl != nil {
				serializedDecl := jf.serializeVariableDeclaration(decl)
				if serializedDecl != nil {
					prefetched[varName] = serializedDecl
					break // Found it, don't look in more frames
				}
			}
		}
	}

	return prefetched
}

// callWithOnDemandContext calls a function with on-demand variable lookup via shared memory.
// Instead of serializing any context to JSON, this uses shared memory for variable data.
// The lookup process:
// 1. Go creates a shared memory buffer for variable data
// 2. When JavaScript needs a variable, it sends ONLY the variable name (tiny JSON)
// 3. Go looks up the variable and writes it to shared memory in binary format
// 4. JavaScript reads the value directly from shared memory (no JSON parsing)
func (jf *JSFunctionDefinition) callWithOnDemandContext(evalContext EvalContextProvider, args ...any) (any, error) {
	// Serialize arguments for transfer (these are usually small)
	serializedArgs, err := jf.serializeArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize arguments: %w", err)
	}

	// Get frames for variable lookup
	frames := evalContext.GetFramesAny()
	importantScope := evalContext.GetImportantScopeAny()

	// Create a shared memory buffer for variable data (1MB should be plenty)
	const varBufferSize = 1024 * 1024
	shmMgr := jf.runtime.SharedMemoryManager()
	if shmMgr == nil {
		// Fall back to JSON-based lookup if shared memory not available
		return jf.callWithOnDemandContextJSON(evalContext, args...)
	}

	varBuffer, err := shmMgr.Create(varBufferSize)
	if err != nil {
		// Fall back to JSON-based lookup
		return jf.callWithOnDemandContextJSON(evalContext, args...)
	}
	defer shmMgr.Destroy(varBuffer.Key())

	// Attach the buffer to JavaScript
	_, err = jf.runtime.SendCommand(Command{
		Cmd: "attachVarBuffer",
		Data: map[string]any{
			"key":  varBuffer.Key(),
			"path": varBuffer.Path(),
			"size": varBuffer.Size(),
		},
	})
	if err != nil {
		return jf.callWithOnDemandContextJSON(evalContext, args...)
	}

	// Register a callback handler for variable lookup
	// The callback only receives the variable name - value is written to shared memory
	jf.runtime.RegisterCallback("lookupVariable", func(data any) (any, error) {
		reqData, ok := data.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid variable lookup request")
		}

		varName, _ := reqData["name"].(string)
		frameIdx := 0
		if idx, ok := reqData["frameIndex"].(float64); ok {
			frameIdx = int(idx)
		}

		// Search for the variable starting from the specified frame index
		for i := frameIdx; i < len(frames); i++ {
			frame := frames[i]
			if frame == nil {
				continue
			}

			// Check if frame has Variables() method (like Ruleset)
			variablesProvider, ok := frame.(interface {
				Variables() map[string]any
			})
			if !ok {
				continue
			}

			variables := variablesProvider.Variables()
			if variables == nil {
				continue
			}

			if decl, exists := variables[varName]; exists && decl != nil {
				// Found the variable - write it to shared memory in binary format
				offset, written, err := writeVariableToSharedMemory(varBuffer, decl)
				if err != nil {
					// Fall back to JSON serialization for this variable
					serializedDecl := jf.serializeVariableDeclaration(decl)
					return map[string]any{
						"found":      true,
						"frameIndex": i,
						"value":      serializedDecl,
						"useJSON":    true,
					}, nil
				}
				return map[string]any{
					"found":      true,
					"frameIndex": i,
					"shmOffset":  offset,
					"shmLength":  written,
				}, nil
			}
		}

		// Variable not found
		return map[string]any{"found": false}, nil
	})

	// Unregister callback when done
	defer jf.runtime.UnregisterCallback("lookupVariable")
	defer jf.runtime.SendCommand(Command{Cmd: "detachVarBuffer"})

	// Send minimal context info
	minimalContext := map[string]any{
		"frameCount":        len(frames),
		"importantScope":    serializeImportantScope(importantScope),
		"useOnDemandLookup": true,
		"useSharedMemory":   true,
	}

	// Call the function via Node.js runtime
	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name":    jf.name,
			"args":    serializedArgs,
			"context": minimalContext,
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

// callWithOnDemandContextJSON is the fallback when shared memory is not available.
func (jf *JSFunctionDefinition) callWithOnDemandContextJSON(evalContext EvalContextProvider, args ...any) (any, error) {
	serializedArgs, err := jf.serializeArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize arguments: %w", err)
	}

	frames := evalContext.GetFramesAny()
	importantScope := evalContext.GetImportantScopeAny()

	jf.runtime.RegisterCallback("lookupVariable", func(data any) (any, error) {
		reqData, ok := data.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid variable lookup request")
		}

		varName, _ := reqData["name"].(string)
		frameIdx := 0
		if idx, ok := reqData["frameIndex"].(float64); ok {
			frameIdx = int(idx)
		}

		for i := frameIdx; i < len(frames); i++ {
			frame := frames[i]
			if frame == nil {
				continue
			}

			variablesProvider, ok := frame.(interface{ Variables() map[string]any })
			if !ok {
				continue
			}

			variables := variablesProvider.Variables()
			if variables == nil {
				continue
			}

			if decl, exists := variables[varName]; exists && decl != nil {
				serializedDecl := jf.serializeVariableDeclaration(decl)
				if serializedDecl != nil {
					return map[string]any{
						"found":      true,
						"frameIndex": i,
						"value":      serializedDecl,
					}, nil
				}
			}
		}

		return map[string]any{"found": false}, nil
	})

	defer jf.runtime.UnregisterCallback("lookupVariable")

	minimalContext := map[string]any{
		"frameCount":        len(frames),
		"importantScope":    serializeImportantScope(importantScope),
		"useOnDemandLookup": true,
	}

	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name":    jf.name,
			"args":    serializedArgs,
			"context": minimalContext,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("function call failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("JavaScript function error: %s", resp.Error)
	}

	result, err := jf.deserializeResult(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize result: %w", err)
	}

	return result, nil
}

// writeVariableToSharedMemory writes a variable declaration to shared memory in binary format.
// Returns the offset and number of bytes written.
//
// Binary format:
// [1 byte: type] [4 bytes: value length] [value data...]
//
// Types:
// 0 = null/undefined
// 1 = Dimension (8 bytes float64 value + 4 bytes unit length + unit string)
// 2 = Color (8 bytes r, g, b, alpha as float64s = 32 bytes)
// 3 = Quoted (4 bytes length + string + 1 byte quote char)
// 4 = Keyword (4 bytes length + string)
// 5 = Expression (serialized as JSON for now - complex)
// 255 = fallback to JSON
func writeVariableToSharedMemory(shm *SharedMemory, decl any) (int, int, error) {
	if decl == nil {
		return 0, 0, fmt.Errorf("nil declaration")
	}

	// Get the value from the declaration
	valueProvider, ok := decl.(interface{ GetValue() any })
	if !ok {
		return 0, 0, fmt.Errorf("declaration has no GetValue")
	}
	value := valueProvider.GetValue()
	if value == nil {
		return 0, 0, fmt.Errorf("nil value")
	}

	// Get important flag
	important := false
	if ip, ok := decl.(interface{ GetImportant() bool }); ok {
		important = ip.GetImportant()
	}

	// Write to shared memory buffer
	// Start at offset 0 for simplicity (could use a write pointer for multiple values)
	offset := 0
	buf := make([]byte, 0, 256)

	// Write important flag
	if important {
		buf = append(buf, 1)
	} else {
		buf = append(buf, 0)
	}

	// Write value based on type
	switch v := value.(type) {
	case interface{ GetType() string }:
		nodeType := v.GetType()
		switch nodeType {
		case "Dimension":
			buf = append(buf, 1) // type = Dimension
			// Get value
			if valGetter, ok := v.(interface{ GetValue() float64 }); ok {
				val := valGetter.GetValue()
				buf = appendFloat64(buf, val)
			} else {
				return 0, 0, fmt.Errorf("Dimension has no GetValue")
			}
			// Get unit
			if unitGetter, ok := v.(interface{ GetUnit() any }); ok {
				unit := unitGetter.GetUnit()
				unitStr := ""
				if unit != nil {
					if s, ok := unit.(fmt.Stringer); ok {
						unitStr = s.String()
					} else if s, ok := unit.(interface{ ToString() string }); ok {
						unitStr = s.ToString()
					}
				}
				buf = appendString(buf, unitStr)
			} else {
				buf = appendString(buf, "")
			}

		case "Color":
			buf = append(buf, 2) // type = Color
			if rgbGetter, ok := v.(interface{ GetRGB() []float64 }); ok {
				rgb := rgbGetter.GetRGB()
				for _, c := range rgb {
					buf = appendFloat64(buf, c)
				}
				// Pad to 3 values if needed
				for i := len(rgb); i < 3; i++ {
					buf = appendFloat64(buf, 0)
				}
			} else {
				buf = appendFloat64(buf, 0)
				buf = appendFloat64(buf, 0)
				buf = appendFloat64(buf, 0)
			}
			if alphaGetter, ok := v.(interface{ GetAlpha() float64 }); ok {
				buf = appendFloat64(buf, alphaGetter.GetAlpha())
			} else {
				buf = appendFloat64(buf, 1)
			}

		case "Quoted":
			buf = append(buf, 3) // type = Quoted
			if valGetter, ok := v.(interface{ GetValue() string }); ok {
				buf = appendString(buf, valGetter.GetValue())
			} else {
				buf = appendString(buf, "")
			}
			if quoteGetter, ok := v.(interface{ GetQuote() string }); ok {
				q := quoteGetter.GetQuote()
				if len(q) > 0 {
					buf = append(buf, q[0])
				} else {
					buf = append(buf, '"')
				}
			} else {
				buf = append(buf, '"')
			}

		case "Keyword":
			buf = append(buf, 4) // type = Keyword
			if valGetter, ok := v.(interface{ GetValue() string }); ok {
				buf = appendString(buf, valGetter.GetValue())
			} else {
				buf = appendString(buf, "")
			}

		default:
			// Unsupported type - signal to use JSON fallback
			return 0, 0, fmt.Errorf("unsupported type: %s", nodeType)
		}

	default:
		return 0, 0, fmt.Errorf("value is not a node")
	}

	// Write to shared memory
	if err := shm.Write(offset, buf); err != nil {
		return 0, 0, err
	}

	return offset, len(buf), nil
}

// appendFloat64 appends a float64 to a byte slice in little-endian format.
func appendFloat64(buf []byte, v float64) []byte {
	bits := *(*uint64)(unsafe.Pointer(&v))
	return append(buf,
		byte(bits),
		byte(bits>>8),
		byte(bits>>16),
		byte(bits>>24),
		byte(bits>>32),
		byte(bits>>40),
		byte(bits>>48),
		byte(bits>>56),
	)
}

// appendString appends a length-prefixed string to a byte slice.
func appendString(buf []byte, s string) []byte {
	length := uint32(len(s))
	buf = append(buf,
		byte(length),
		byte(length>>8),
		byte(length>>16),
		byte(length>>24),
	)
	return append(buf, s...)
}

// serializeImportantScope serializes the important scope array.
func serializeImportantScope(scope []map[string]any) []map[string]any {
	result := make([]map[string]any, len(scope))
	for i, s := range scope {
		result[i] = make(map[string]any)
		for k, v := range s {
			result[i][k] = v
		}
	}
	return result
}

// callViaJSONWithContext calls the JavaScript function with context using JSON serialization.
func (jf *JSFunctionDefinition) callViaJSONWithContext(evalContext EvalContextProvider, args ...any) (any, error) {
	// Serialize arguments for transfer
	serializedArgs, err := jf.serializeArgs(args)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize arguments: %w", err)
	}

	// Serialize the evaluation context
	serializedContext := jf.serializeEvalContext(evalContext)

	// Debug: log context info
	if os.Getenv("LESS_GO_DEBUG") == "1" && serializedContext != nil {
		if frames, ok := serializedContext["frames"].([]map[string]any); ok {
			fmt.Printf("[callViaJSONWithContext] Function %s: context has %d frames\n", jf.name, len(frames))
		} else {
			fmt.Printf("[callViaJSONWithContext] Function %s: context frames not available (type: %T)\n", jf.name, serializedContext["frames"])
		}
	} else if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[callViaJSONWithContext] Function %s: context is nil\n", jf.name)
	}

	// Call the function via Node.js runtime with context
	resp, err := jf.runtime.SendCommand(Command{
		Cmd: "callFunction",
		Data: map[string]any{
			"name":    jf.name,
			"args":    serializedArgs,
			"context": serializedContext,
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

// contextCache stores the last serialized context to avoid re-serialization
// when the context hasn't changed between consecutive function calls.
type contextCache struct {
	serialized map[string]any
	frameCount int
	version    uint64
}

// Global context cache (thread-safe via sync.Once pattern in usage)
var (
	contextCacheMu    sync.RWMutex
	lastContextCache  *contextCache
	contextVersion    uint64 = 0
)

// IncrementContextVersion should be called when frames are pushed/popped
// to invalidate the context cache.
func IncrementContextVersion() {
	contextCacheMu.Lock()
	contextVersion++
	contextCacheMu.Unlock()
}

// GetContextVersion returns the current context version.
func GetContextVersion() uint64 {
	contextCacheMu.RLock()
	defer contextCacheMu.RUnlock()
	return contextVersion
}

// serializeEvalContext serializes the evaluation context for JavaScript.
// This includes frames with their variables so plugin functions can look up values.
//
// OPTIMIZATION: Uses lazy serialization to avoid serializing all frames upfront:
// 1. Only the first MAX_SERIALIZED_FRAMES frames are fully serialized
// 2. For remaining frames, only variable names are serialized (not values)
// 3. JavaScript can request specific variable values on-demand
func (jf *JSFunctionDefinition) serializeEvalContext(evalContext EvalContextProvider) map[string]any {
	if evalContext == nil {
		return nil
	}

	frames := evalContext.GetFramesAny()
	importantScope := evalContext.GetImportantScopeAny()

	// Check if we can use cached context
	contextCacheMu.RLock()
	currentVersion := contextVersion
	cache := lastContextCache
	contextCacheMu.RUnlock()

	if cache != nil && cache.version == currentVersion && cache.frameCount == len(frames) {
		// Context hasn't changed, reuse cached serialization
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[serializeEvalContext] Using cached context (version=%d, frames=%d)\n", currentVersion, len(frames))
		}
		return cache.serialized
	}

	// Limit on how many frames to fully serialize
	// Most variable lookups find the variable in the first few frames
	const MAX_SERIALIZED_FRAMES = 50

	// Serialize frames with their variables
	serializedFrames := make([]map[string]any, 0, len(frames))
	for i, frame := range frames {
		var serializedFrame map[string]any
		if i < MAX_SERIALIZED_FRAMES {
			// Fully serialize the first N frames
			serializedFrame = jf.serializeFrame(frame)
		} else {
			// For remaining frames, only serialize variable names (not values)
			// This allows JavaScript to know which variables exist but defer value lookup
			serializedFrame = jf.serializeFrameNamesOnly(frame)
		}
		if serializedFrame != nil {
			serializedFrame["_frameIndex"] = i
			serializedFrames = append(serializedFrames, serializedFrame)
		}
	}

	// Serialize important scope
	serializedImportantScope := make([]map[string]any, len(importantScope))
	for i, scope := range importantScope {
		serializedImportantScope[i] = make(map[string]any)
		for k, v := range scope {
			serializedImportantScope[i][k] = v
		}
	}

	result := map[string]any{
		"frames":         serializedFrames,
		"importantScope": serializedImportantScope,
		"totalFrames":    len(frames),
	}

	// Cache the serialized context
	contextCacheMu.Lock()
	lastContextCache = &contextCache{
		serialized: result,
		frameCount: len(frames),
		version:    currentVersion,
	}
	contextCacheMu.Unlock()

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[serializeEvalContext] Serialized %d frames (full: %d, names-only: %d)\n",
			len(serializedFrames), min(len(frames), MAX_SERIALIZED_FRAMES), max(0, len(frames)-MAX_SERIALIZED_FRAMES))
	}

	return result
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// serializeFrameNamesOnly serializes only the variable names from a frame.
// This is used for frames beyond the MAX_SERIALIZED_FRAMES threshold to allow
// JavaScript to know which variables exist without the overhead of serializing values.
func (jf *JSFunctionDefinition) serializeFrameNamesOnly(frame any) map[string]any {
	if frame == nil {
		return nil
	}

	// Check if frame has Variables() method (like Ruleset)
	variablesProvider, ok := frame.(interface {
		Variables() map[string]any
	})
	if !ok {
		return nil
	}

	variables := variablesProvider.Variables()
	if variables == nil {
		return map[string]any{"variables": map[string]any{}, "_lazyLoad": true}
	}

	// Only serialize variable names (set value to nil to indicate lazy loading)
	varNames := make(map[string]any)
	for name := range variables {
		varNames[name] = nil // nil indicates value needs to be fetched
	}

	return map[string]any{
		"variables": varNames,
		"_lazyLoad": true,
	}
}

// serializeFrame serializes a single frame (typically a Ruleset) with its variables.
func (jf *JSFunctionDefinition) serializeFrame(frame any) map[string]any {
	if frame == nil {
		return nil
	}

	// Check if frame has Variables() method (like Ruleset)
	variablesProvider, ok := frame.(interface {
		Variables() map[string]any
	})
	if !ok {
		return nil
	}

	variables := variablesProvider.Variables()
	if variables == nil {
		return map[string]any{"variables": map[string]any{}}
	}

	// Serialize each variable declaration
	serializedVars := make(map[string]any)
	for name, decl := range variables {
		serializedDecl := jf.serializeVariableDeclaration(decl)
		if serializedDecl != nil {
			serializedVars[name] = serializedDecl
		}
	}

	return map[string]any{
		"variables": serializedVars,
	}
}

// serializeVariableDeclaration serializes a variable declaration.
func (jf *JSFunctionDefinition) serializeVariableDeclaration(decl any) map[string]any {
	if decl == nil {
		return nil
	}

	// Try to get value from declaration
	valueProvider, ok := decl.(interface {
		GetValue() any
	})
	if !ok {
		// Try alternate interface
		if declMap, ok := decl.(map[string]any); ok {
			return declMap
		}
		return nil
	}

	value := valueProvider.GetValue()

	// Check for important flag
	important := false
	if importantProvider, ok := decl.(interface{ GetImportant() bool }); ok {
		important = importantProvider.GetImportant()
	} else if importantProvider, ok := decl.(interface{ GetImportant() string }); ok {
		important = importantProvider.GetImportant() != ""
	}

	return map[string]any{
		"value":     serializeNode(value),
		"important": important,
	}
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
			// Try GetRGB method first, then fall back to field access
			if getter, ok := node.(interface{ GetRGB() []float64 }); ok {
				nodeMap["rgb"] = getter.GetRGB()
			} else if colorNode, ok := node.(interface{ GetColorRGB() []float64 }); ok {
				// Alternative method name
				nodeMap["rgb"] = colorNode.GetColorRGB()
			} else {
				// Try to access RGB field via reflection as last resort
				// For less_go.Color, RGB is a public field
				if hasRGB := extractFieldByName(node, "RGB"); hasRGB != nil {
					nodeMap["rgb"] = hasRGB
				}
			}
			if getter, ok := node.(interface{ GetAlpha() float64 }); ok {
				nodeMap["alpha"] = getter.GetAlpha()
			} else {
				// Try to access Alpha field
				if alpha := extractFieldByName(node, "Alpha"); alpha != nil {
					nodeMap["alpha"] = alpha
				}
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
	// Check for ToString method (for Unit type)
	if toStringer, ok := unit.(interface{ ToString() string }); ok {
		return toStringer.ToString()
	}
	// Fallback - but avoid printing Go struct syntax
	return ""
}

// extractFieldByName uses reflection to extract a field value by name from a struct.
// Returns nil if the field doesn't exist or isn't accessible.
func extractFieldByName(node any, fieldName string) any {
	if node == nil {
		return nil
	}
	v := reflect.ValueOf(node)
	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	field := v.FieldByName(fieldName)
	if !field.IsValid() || !field.CanInterface() {
		return nil
	}
	return field.Interface()
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
