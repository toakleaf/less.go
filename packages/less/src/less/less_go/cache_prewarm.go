package less_go

import (
	"fmt"
	"os"
	"reflect"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// PluginCallInfo represents information about a plugin function call
// collected during AST traversal for cache pre-warming.
type PluginCallInfo struct {
	// FunctionName is the name of the plugin function being called
	FunctionName string

	// Args are the arguments to the function (not yet evaluated)
	Args []any

	// CacheKey is the pre-computed cache key for this call
	CacheKey string
}

// PluginCallCollector walks the AST and collects all Call nodes
// that reference JavaScript plugin functions.
type PluginCallCollector struct {
	// pluginFunctions is the set of plugin function names
	pluginFunctions map[string]bool

	// collected stores unique calls by cache key
	collected map[string]*PluginCallInfo

	// debug enables verbose logging
	debug bool
}

// NewPluginCallCollector creates a new collector with the given plugin function names.
func NewPluginCallCollector(pluginFunctionNames []string) *PluginCallCollector {
	funcs := make(map[string]bool, len(pluginFunctionNames))
	for _, name := range pluginFunctionNames {
		funcs[name] = true
	}
	return &PluginCallCollector{
		pluginFunctions: funcs,
		collected:       make(map[string]*PluginCallInfo),
		debug:           os.Getenv("LESS_GO_DEBUG") == "1",
	}
}

// Collect walks the AST starting from root and collects all plugin function calls.
func (c *PluginCallCollector) Collect(root any) []*PluginCallInfo {
	c.visit(root)

	// Convert map to slice
	result := make([]*PluginCallInfo, 0, len(c.collected))
	for _, info := range c.collected {
		result = append(result, info)
	}

	if c.debug {
		fmt.Printf("[PluginCallCollector] Collected %d unique plugin calls\n", len(result))
	}

	return result
}

// visit recursively traverses an AST node and its children.
func (c *PluginCallCollector) visit(node any) {
	if node == nil {
		return
	}

	// Check if this is a Call node
	if call, ok := node.(*Call); ok {
		c.visitCall(call)
	}

	// Traverse children using reflection
	c.visitChildren(node)
}

// visitCall checks if a Call node references a plugin function and collects it.
func (c *PluginCallCollector) visitCall(call *Call) {
	if call == nil || call.Name == "" {
		return
	}

	// Check if this is a plugin function
	if !c.pluginFunctions[call.Name] {
		return
	}

	// Try to create a cache key from the arguments
	// Note: Some args may not be evaluatable yet (contain variables, etc.)
	// We'll skip calls that have complex args that we can't serialize
	cacheKey, ok := c.makeCacheKeyFromUnevaluatedArgs(call.Name, call.Args)
	if !ok {
		if c.debug {
			fmt.Printf("[PluginCallCollector] Skipping %s - args not serializable\n", call.Name)
		}
		return
	}

	// Check for duplicate
	if _, exists := c.collected[cacheKey]; exists {
		return
	}

	// Store the call info
	c.collected[cacheKey] = &PluginCallInfo{
		FunctionName: call.Name,
		Args:         call.Args,
		CacheKey:     cacheKey,
	}

	if c.debug {
		fmt.Printf("[PluginCallCollector] Collected call: %s with key: %s\n", call.Name, cacheKey[:min(60, len(cacheKey))])
	}
}

// makeCacheKeyFromUnevaluatedArgs attempts to create a cache key from unevaluated args.
// Returns (key, true) if successful, or ("", false) if the args can't be serialized.
func (c *PluginCallCollector) makeCacheKeyFromUnevaluatedArgs(funcName string, args []any) (string, bool) {
	key := funcName + ":"

	for i, arg := range args {
		if i > 0 {
			key += "|"
		}

		argStr, ok := c.argToString(arg)
		if !ok {
			return "", false
		}
		key += argStr
	}

	return key, true
}

// argToString converts an argument to a string for cache key purposes.
// Returns ("", false) if the argument can't be serialized (e.g., contains variables).
func (c *PluginCallCollector) argToString(arg any) (string, bool) {
	if arg == nil {
		return "nil", true
	}

	// Handle common node types that are safe to serialize
	switch v := arg.(type) {
	case string:
		return v, true

	case float64:
		return fmt.Sprintf("%g", v), true

	case int:
		return fmt.Sprintf("%d", v), true

	case bool:
		return fmt.Sprintf("%t", v), true

	case *Dimension:
		if v == nil {
			return "nil", true
		}
		unit := ""
		if v.Unit != nil {
			unit = v.Unit.ToString()
		}
		return fmt.Sprintf("%g%s", v.Value, unit), true

	case *Color:
		if v == nil {
			return "nil", true
		}
		if len(v.RGB) >= 3 {
			return fmt.Sprintf("rgba(%g,%g,%g,%g)", v.RGB[0], v.RGB[1], v.RGB[2], v.Alpha), true
		}
		return "", false

	case *Quoted:
		if v == nil {
			return "nil", true
		}
		return fmt.Sprintf("%s%s%s", v.GetQuote(), v.GetValue(), v.GetQuote()), true

	case *Keyword:
		if v == nil {
			return "nil", true
		}
		return v.GetValue(), true

	case *Anonymous:
		if v == nil {
			return "nil", true
		}
		return fmt.Sprintf("%v", v.GetValue()), true

	case *Variable:
		// Variables need to be evaluated first - can't pre-warm these
		return "", false

	case *Operation:
		// Operations need to be evaluated first
		return "", false

	case *Call:
		// Nested calls need to be evaluated first
		return "", false

	case *Expression:
		// Try to serialize expression if it's simple (single value)
		if v == nil || len(v.Value) == 0 {
			return "nil", true
		}
		if len(v.Value) == 1 {
			return c.argToString(v.Value[0])
		}
		// Multi-value expressions are complex
		return "", false
	}

	// For other types, try using reflection to get a stable representation
	rv := reflect.ValueOf(arg)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return "nil", true
		}
		// Try to call ToCSS method if available
		if m := rv.MethodByName("ToCSS"); m.IsValid() {
			results := m.Call([]reflect.Value{reflect.ValueOf(make(map[string]any))})
			if len(results) > 0 {
				return fmt.Sprintf("%v", results[0].Interface()), true
			}
		}
		// Try String method
		if m := rv.MethodByName("String"); m.IsValid() {
			results := m.Call(nil)
			if len(results) > 0 {
				return fmt.Sprintf("%v", results[0].Interface()), true
			}
		}
	}

	// Unknown type - can't serialize
	return "", false
}

// visitChildren traverses all child nodes of a given node.
func (c *PluginCallCollector) visitChildren(node any) {
	if node == nil {
		return
	}

	// Use reflection to find children
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	// Iterate over all fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}

		c.visitField(field)
	}
}

// visitField handles traversing a single field value.
func (c *PluginCallCollector) visitField(field reflect.Value) {
	switch field.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !field.IsNil() {
			c.visit(field.Interface())
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < field.Len(); i++ {
			elem := field.Index(i)
			if elem.CanInterface() {
				c.visit(elem.Interface())
			}
		}

	case reflect.Map:
		iter := field.MapRange()
		for iter.Next() {
			val := iter.Value()
			if val.CanInterface() {
				c.visit(val.Interface())
			}
		}
	}
}

// WarmPluginCache collects plugin function calls from the AST and pre-warms
// the result cache with a batch IPC call.
//
// Parameters:
//   - root: The root AST node (typically a *Ruleset)
//   - bridge: The NodeJSPluginBridge with the runtime and registered functions
//   - evalContext: The evaluation context (required for variable lookup)
//
// Returns the number of cache entries warmed and any error.
func WarmPluginCache(root any, bridge *NodeJSPluginBridge, evalContext any) (int, error) {
	if bridge == nil {
		return 0, nil
	}

	rt := bridge.GetRuntime()
	if rt == nil {
		return 0, nil
	}

	// Get the list of registered plugin functions
	funcRegistry := bridge.GetFunctionRegistry()
	if funcRegistry == nil {
		return 0, nil
	}

	pluginFuncNames := funcRegistry.GetJSFunctionNames()
	if len(pluginFuncNames) == 0 {
		return 0, nil
	}

	debug := os.Getenv("LESS_GO_DEBUG") == "1"
	if debug {
		fmt.Printf("[WarmPluginCache] Found %d plugin functions: %v\n", len(pluginFuncNames), pluginFuncNames)
	}

	// Collect all plugin function calls from the AST
	collector := NewPluginCallCollector(pluginFuncNames)
	calls := collector.Collect(root)

	if len(calls) == 0 {
		if debug {
			fmt.Printf("[WarmPluginCache] No cacheable plugin calls found in AST\n")
		}
		return 0, nil
	}

	// Build batch calls
	// NOTE: Most plugin functions require evaluation context for variable lookup.
	// Currently we skip context-dependent warming because serializing the full
	// evaluation context is complex. The existing per-call cache achieves ~95%
	// hit rate after warming, so this pre-warming is limited to pure functions
	// that don't need context.
	batchCalls := make([]runtime.BatchCall, 0, len(calls))

	for _, call := range calls {
		// Serialize arguments
		serializedArgs, err := serializeArgsForBatch(call.Args, evalContext)
		if err != nil {
			if debug {
				fmt.Printf("[WarmPluginCache] Skipping %s - serialization error: %v\n", call.FunctionName, err)
			}
			continue
		}

		// For now, skip functions that are known to require context (variable lookup)
		// These include most Bootstrap4 functions like theme-color, color-yiq, etc.
		// TODO: Add a registry of context-free functions that can be safely pre-warmed
		if isContextDependentFunction(call.FunctionName) {
			if debug {
				fmt.Printf("[WarmPluginCache] Skipping %s - requires context for variable lookup\n", call.FunctionName)
			}
			continue
		}

		batchCalls = append(batchCalls, runtime.BatchCall{
			Key:  call.CacheKey,
			Name: call.FunctionName,
			Args: serializedArgs,
		})
	}

	if len(batchCalls) == 0 {
		if debug {
			fmt.Printf("[WarmPluginCache] No batch calls to make after serialization\n")
		}
		return 0, nil
	}

	if debug {
		fmt.Printf("[WarmPluginCache] Warming cache with %d batch calls\n", len(batchCalls))
		for i, call := range batchCalls {
			fmt.Printf("[WarmPluginCache]   Call %d: %s, args=%v\n", i, call.Name, call.Args)
		}
	}

	// Execute batch call and cache results
	cached, err := rt.BatchCallFunctionsAndCache(batchCalls)
	if err != nil {
		return 0, fmt.Errorf("batch cache warming failed: %w", err)
	}

	if debug {
		fmt.Printf("[WarmPluginCache] Successfully cached %d results\n", cached)
	}

	return cached, nil
}

// WarmPluginCacheFromLazyBridge is like WarmPluginCache but takes a LazyNodeJSPluginBridge.
// It only warms the cache if the bridge is already initialized (i.e., plugins have been loaded).
func WarmPluginCacheFromLazyBridge(root any, lazyBridge *LazyNodeJSPluginBridge, evalContext any) (int, error) {
	if lazyBridge == nil || !lazyBridge.IsInitialized() {
		return 0, nil
	}

	bridge, err := lazyBridge.GetBridge()
	if err != nil {
		return 0, err
	}

	return WarmPluginCache(root, bridge, evalContext)
}

// contextDependentFunctions is a set of plugin function names known to require
// evaluation context for variable lookup. These can't be safely pre-warmed.
// This list is based on Bootstrap4 plugins that use this.context.frames.
var contextDependentFunctions = map[string]bool{
	// Bootstrap4 plugins that look up variables from context
	"theme-color":       true,
	"theme-color-level": true,
	"color-yiq":         true,
	"color":             true,
	"gray":              true,
	"breakpoint-min":    true,
	"breakpoint-max":    true,
	"breakpoint-infix":  true,
	"breakpoint-next":   true,
	"map-get":           true, // Looks up variable maps
	"index":             true,
}

// isContextDependentFunction returns true if the function requires evaluation
// context for variable lookup and cannot be safely pre-warmed.
func isContextDependentFunction(name string) bool {
	return contextDependentFunctions[name]
}

// serializeEvalContextForBatch converts an Eval context to a format suitable for batch IPC.
// This is needed for plugin functions that need to look up variables from frames.
func serializeEvalContextForBatch(evalCtx *Eval) map[string]any {
	if evalCtx == nil {
		return nil
	}

	result := make(map[string]any)

	// Serialize frames - this is critical for variable lookup
	if len(evalCtx.Frames) > 0 {
		frames := make([]any, len(evalCtx.Frames))
		for i, frame := range evalCtx.Frames {
			frames[i] = serializeFrameForBatch(frame)
		}
		result["frames"] = frames
	}

	return result
}

// serializeFrameForBatch converts a frame to a serializable format.
func serializeFrameForBatch(frame any) map[string]any {
	if frame == nil {
		return nil
	}

	result := map[string]any{}

	// Handle Ruleset frames
	if ruleset, ok := frame.(*Ruleset); ok {
		result["_type"] = "Ruleset"

		// Serialize variable declarations from the ruleset
		vars := make(map[string]any)
		for _, rule := range ruleset.Rules {
			if decl, ok := rule.(*Declaration); ok {
				nameStr := decl.GetName()
				// Only include variable declarations (starting with @)
				if len(nameStr) > 0 && nameStr[0] == '@' {
					vars[nameStr] = serializeNodeForBatch(decl.Value, nil)
				}
			}
		}
		if len(vars) > 0 {
			result["variables"] = vars
		}

		return result
	}

	// For other frame types, just mark the type
	if typer, ok := frame.(interface{ GetType() string }); ok {
		result["_type"] = typer.GetType()
	}

	return result
}

// serializeArgsForBatch converts arguments to a format suitable for batch IPC.
func serializeArgsForBatch(args []any, evalContext any) ([]any, error) {
	result := make([]any, len(args))

	for i, arg := range args {
		serialized := serializeNodeForBatch(arg, evalContext)
		result[i] = serialized
	}

	return result, nil
}

// serializeNodeForBatch converts a single node to a serializable format.
func serializeNodeForBatch(node any, evalContext any) any {
	if node == nil {
		return nil
	}

	// Handle common node types
	switch v := node.(type) {
	case string, float64, int, int64, bool:
		return v

	case *Dimension:
		if v == nil {
			return nil
		}
		unit := ""
		if v.Unit != nil {
			unit = v.Unit.ToString()
		}
		return map[string]any{
			"_type": "Dimension",
			"value": v.Value,
			"unit":  unit,
		}

	case *Color:
		if v == nil {
			return nil
		}
		return map[string]any{
			"_type": "Color",
			"rgb":   v.RGB,
			"alpha": v.Alpha,
		}

	case *Quoted:
		if v == nil {
			return nil
		}
		return map[string]any{
			"_type":   "Quoted",
			"value":   v.GetValue(),
			"quote":   v.GetQuote(),
			"escaped": v.GetEscaped(),
		}

	case *Keyword:
		if v == nil {
			return nil
		}
		return map[string]any{
			"_type": "Keyword",
			"value": v.GetValue(),
		}

	case *Anonymous:
		if v == nil {
			return nil
		}
		return map[string]any{
			"_type": "Anonymous",
			"value": v.GetValue(),
		}

	case *Expression:
		if v == nil || len(v.Value) == 0 {
			return nil
		}
		serialized := make([]any, len(v.Value))
		for i, val := range v.Value {
			serialized[i] = serializeNodeForBatch(val, evalContext)
		}
		return map[string]any{
			"_type": "Expression",
			"value": serialized,
		}

	case *Value:
		if v == nil || len(v.Value) == 0 {
			return nil
		}
		serialized := make([]any, len(v.Value))
		for i, val := range v.Value {
			serialized[i] = serializeNodeForBatch(val, evalContext)
		}
		return map[string]any{
			"_type": "Value",
			"value": serialized,
		}
	}

	// Try to get type info for other nodes
	if typer, ok := node.(interface{ GetType() string }); ok {
		nodeType := typer.GetType()
		result := map[string]any{
			"_type": nodeType,
		}

		// Try to extract common properties
		if getter, ok := node.(interface{ GetValue() any }); ok {
			result["value"] = getter.GetValue()
		}

		return result
	}

	// Fallback - return as-is
	return node
}

// =============================================================================
// Context-Aware Batch Pre-Warming
// =============================================================================
// These functions enable pre-warming the plugin function cache by evaluating
// arguments with context and batching multiple calls into a single IPC request.
// This is critical for Bootstrap4-style compilations where thousands of plugin
// function calls are made (map-get, color-yiq, breakpoint-next, etc.).
// =============================================================================

// EvaluatedPluginCall represents a plugin function call with fully evaluated arguments.
type EvaluatedPluginCall struct {
	FunctionName string
	CacheKey     string
	Args         []any // Evaluated and serialized arguments
	Context      map[string]any
}

// PluginCallEvaluator walks the AST and collects plugin function calls,
// evaluating their arguments using the provided context.
type PluginCallEvaluator struct {
	pluginFunctions map[string]bool
	evalContext     *Eval
	collected       map[string]*EvaluatedPluginCall
	debug           bool
}

// NewPluginCallEvaluator creates a new evaluator with the given plugin function names.
func NewPluginCallEvaluator(pluginFunctionNames []string, evalContext *Eval) *PluginCallEvaluator {
	funcs := make(map[string]bool, len(pluginFunctionNames))
	for _, name := range pluginFunctionNames {
		funcs[name] = true
	}
	return &PluginCallEvaluator{
		pluginFunctions: funcs,
		evalContext:     evalContext,
		collected:       make(map[string]*EvaluatedPluginCall),
		debug:           os.Getenv("LESS_GO_DEBUG") == "1",
	}
}

// CollectAndEvaluate walks the AST and collects all plugin function calls,
// evaluating their arguments using the provided context.
func (e *PluginCallEvaluator) CollectAndEvaluate(root any) []*EvaluatedPluginCall {
	e.visit(root)

	// Convert map to slice
	result := make([]*EvaluatedPluginCall, 0, len(e.collected))
	for _, call := range e.collected {
		result = append(result, call)
	}

	if e.debug {
		fmt.Printf("[PluginCallEvaluator] Collected and evaluated %d unique plugin calls\n", len(result))
	}

	return result
}

// visit recursively traverses an AST node.
func (e *PluginCallEvaluator) visit(node any) {
	if node == nil {
		return
	}

	// Check if this is a Call node
	if call, ok := node.(*Call); ok {
		e.visitCall(call)
	}

	// Traverse children using reflection
	e.visitChildren(node)
}

// visitCall evaluates a plugin function call and collects it.
func (e *PluginCallEvaluator) visitCall(call *Call) {
	if call == nil || call.Name == "" {
		return
	}

	// Check if this is a plugin function
	if !e.pluginFunctions[call.Name] {
		return
	}

	// Try to evaluate all arguments
	evaluatedArgs := make([]any, 0, len(call.Args))
	for _, arg := range call.Args {
		evaluated, ok := e.tryEvaluateArg(arg)
		if !ok {
			// Can't evaluate this argument - skip this call
			if e.debug {
				fmt.Printf("[PluginCallEvaluator] Skipping %s - couldn't evaluate arg of type %T\n", call.Name, arg)
			}
			return
		}
		evaluatedArgs = append(evaluatedArgs, evaluated)
	}

	// Create cache key from evaluated arguments
	cacheKey := e.makeCacheKey(call.Name, evaluatedArgs)

	// Check for duplicate
	if _, exists := e.collected[cacheKey]; exists {
		return
	}

	// Serialize arguments for batch IPC
	serializedArgs := make([]any, len(evaluatedArgs))
	for i, arg := range evaluatedArgs {
		serializedArgs[i] = serializeNodeForBatch(arg, nil)
	}

	// Serialize context for variable lookup
	var serializedContext map[string]any
	if e.evalContext != nil && isContextDependentFunction(call.Name) {
		serializedContext = serializeEvalContextForBatch(e.evalContext)
	}

	e.collected[cacheKey] = &EvaluatedPluginCall{
		FunctionName: call.Name,
		CacheKey:     cacheKey,
		Args:         serializedArgs,
		Context:      serializedContext,
	}

	if e.debug {
		fmt.Printf("[PluginCallEvaluator] Collected: %s with key: %s\n", call.Name, truncateString(cacheKey, 60))
	}
}

// tryEvaluateArg attempts to evaluate an argument using the context.
// Returns (evaluatedArg, true) on success, (nil, false) if evaluation fails.
func (e *PluginCallEvaluator) tryEvaluateArg(arg any) (any, bool) {
	if arg == nil {
		return nil, true
	}

	// Already evaluated types - pass through
	switch v := arg.(type) {
	case string, float64, int, int64, bool:
		return v, true
	case *Dimension, *Color, *Keyword, *Anonymous:
		return v, true
	case *Quoted:
		if v != nil {
			return v, true
		}
		return nil, true
	}

	// Try to evaluate nodes that need evaluation
	if e.evalContext != nil {
		// Try Eval(any) (any, error) signature
		if evaluable, ok := arg.(interface{ Eval(any) (any, error) }); ok {
			result, err := evaluable.Eval(e.evalContext)
			if err == nil {
				return result, true
			}
			// Evaluation failed - can't use this call
			return nil, false
		}

		// Try Eval(any) any signature
		if evaluable, ok := arg.(interface{ Eval(any) any }); ok {
			result := evaluable.Eval(e.evalContext)
			if result != nil {
				return result, true
			}
			return nil, false
		}
	}

	// Handle Expression with single value
	if expr, ok := arg.(*Expression); ok {
		if expr != nil && len(expr.Value) == 1 {
			return e.tryEvaluateArg(expr.Value[0])
		}
	}

	// Can't evaluate this argument
	return nil, false
}

// makeCacheKey creates a cache key from evaluated arguments.
// This matches the format used by JSFunctionDefinition.makeCacheKey.
func (e *PluginCallEvaluator) makeCacheKey(funcName string, args []any) string {
	key := funcName + ":"

	for i, arg := range args {
		if i > 0 {
			key += "|"
		}
		key += e.argToKeyString(arg)
	}

	return key
}

// argToKeyString converts an evaluated argument to a string for cache key.
func (e *PluginCallEvaluator) argToKeyString(arg any) string {
	if arg == nil {
		return "nil"
	}

	switch v := arg.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case *Dimension:
		if v == nil {
			return "nil"
		}
		unit := ""
		if v.Unit != nil {
			unit = v.Unit.ToString()
		}
		return fmt.Sprintf("%g%s", v.Value, unit)
	case *Color:
		if v == nil {
			return "nil"
		}
		if len(v.RGB) >= 3 {
			return fmt.Sprintf("rgba(%g,%g,%g,%g)", v.RGB[0], v.RGB[1], v.RGB[2], v.Alpha)
		}
		return "color"
	case *Quoted:
		if v == nil {
			return "nil"
		}
		return fmt.Sprintf("%s%s%s", v.GetQuote(), v.GetValue(), v.GetQuote())
	case *Keyword:
		if v == nil {
			return "nil"
		}
		return v.GetValue()
	case *Anonymous:
		if v == nil {
			return "nil"
		}
		return fmt.Sprintf("%v", v.GetValue())
	}

	// Fallback - use type and value
	return fmt.Sprintf("%T:%v", arg, arg)
}

// visitChildren traverses all child nodes.
func (e *PluginCallEvaluator) visitChildren(node any) {
	if node == nil {
		return
	}

	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}
		e.visitFieldValue(field)
	}
}

// visitFieldValue handles traversing a single field value.
func (e *PluginCallEvaluator) visitFieldValue(field reflect.Value) {
	switch field.Kind() {
	case reflect.Ptr, reflect.Interface:
		if !field.IsNil() {
			e.visit(field.Interface())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < field.Len(); i++ {
			elem := field.Index(i)
			if elem.CanInterface() {
				e.visit(elem.Interface())
			}
		}
	case reflect.Map:
		iter := field.MapRange()
		for iter.Next() {
			val := iter.Value()
			if val.CanInterface() {
				e.visit(val.Interface())
			}
		}
	}
}

// WarmPluginCacheWithContext collects plugin function calls from the AST,
// evaluates their arguments using the provided context, and pre-warms the cache
// with a single batch IPC call.
//
// This is the recommended function for pre-warming plugin caches before evaluation.
// Unlike WarmPluginCache, this function properly handles context-dependent functions
// like map-get, color-yiq, and breakpoint-next that need variable lookup.
//
// Parameters:
//   - root: The root AST node (typically a *Ruleset)
//   - bridge: The NodeJSPluginBridge with the runtime and registered functions
//   - evalContext: The evaluation context with frames for variable lookup
//
// Returns the number of cache entries warmed and any error.
func WarmPluginCacheWithContext(root any, bridge *NodeJSPluginBridge, evalContext *Eval) (int, error) {
	if bridge == nil || evalContext == nil {
		return 0, nil
	}

	rt := bridge.GetRuntime()
	if rt == nil {
		return 0, nil
	}

	// Get the list of registered plugin functions
	funcRegistry := bridge.GetFunctionRegistry()
	if funcRegistry == nil {
		return 0, nil
	}

	pluginFuncNames := funcRegistry.GetJSFunctionNames()
	if len(pluginFuncNames) == 0 {
		return 0, nil
	}

	debug := os.Getenv("LESS_GO_DEBUG") == "1"
	if debug {
		fmt.Printf("[WarmPluginCacheWithContext] Found %d plugin functions\n", len(pluginFuncNames))
	}

	// Collect and evaluate all plugin function calls from the AST
	evaluator := NewPluginCallEvaluator(pluginFuncNames, evalContext)
	calls := evaluator.CollectAndEvaluate(root)

	if len(calls) == 0 {
		if debug {
			fmt.Printf("[WarmPluginCacheWithContext] No cacheable plugin calls found\n")
		}
		return 0, nil
	}

	// Build batch calls
	batchCalls := make([]runtime.BatchCall, 0, len(calls))
	for _, call := range calls {
		batchCalls = append(batchCalls, runtime.BatchCall{
			Key:     call.CacheKey,
			Name:    call.FunctionName,
			Args:    call.Args,
			Context: call.Context,
		})
	}

	if debug {
		fmt.Printf("[WarmPluginCacheWithContext] Warming cache with %d batch calls\n", len(batchCalls))
	}

	// Execute batch call and cache results
	cached, err := rt.BatchCallFunctionsAndCache(batchCalls)
	if err != nil {
		return 0, fmt.Errorf("batch cache warming failed: %w", err)
	}

	if debug {
		fmt.Printf("[WarmPluginCacheWithContext] Successfully cached %d results\n", cached)
	}

	return cached, nil
}

// WarmPluginCacheWithContextFromLazyBridge is like WarmPluginCacheWithContext
// but takes a LazyNodeJSPluginBridge. It only warms if the bridge is initialized.
func WarmPluginCacheWithContextFromLazyBridge(root any, lazyBridge *LazyNodeJSPluginBridge, evalContext *Eval) (int, error) {
	if lazyBridge == nil || !lazyBridge.IsInitialized() {
		return 0, nil
	}

	bridge, err := lazyBridge.GetBridge()
	if err != nil {
		return 0, err
	}

	return WarmPluginCacheWithContext(root, bridge, evalContext)
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}
