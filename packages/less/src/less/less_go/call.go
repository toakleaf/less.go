package less_go

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

// mathEvalContextPool pools *Eval contexts for createMathEnabledContext.
// These are short-lived contexts used during function argument evaluation.
var mathEvalContextPool = sync.Pool{
	New: func() any {
		return &Eval{}
	},
}

// getMathEnabledEvalContext gets a pooled Eval context and copies fields from source.
// The returned context has MathOn set to true.
func getMathEnabledEvalContext(source *Eval) *Eval {
	ctx := mathEvalContextPool.Get().(*Eval)
	// Copy all fields from source (shallow copy)
	*ctx = *source
	// Enable math for function argument evaluation
	ctx.MathOn = true
	return ctx
}

// putMathEnabledEvalContext returns a context to the pool after resetting.
// This must be called when done using a context from getMathEnabledEvalContext.
func putMathEnabledEvalContext(ctx *Eval) {
	if ctx == nil {
		return
	}
	// Reset pointer fields to avoid retaining references
	ctx.Paths = nil
	ctx.ImportantScope = nil
	ctx.Frames = nil
	ctx.CalcStack = nil
	ctx.ParensStack = nil
	ctx.DefaultFunc = nil
	ctx.FunctionRegistry = nil
	ctx.MediaBlocks = nil
	ctx.MediaPath = nil
	ctx.PluginManager = nil
	ctx.PluginBridge = nil
	ctx.LazyPluginBridge = nil
	// Reset scalar fields to zero values
	ctx.Compress = false
	ctx.Math = 0
	ctx.StrictUnits = false
	ctx.SourceMap = false
	ctx.ImportMultiple = false
	ctx.UrlArgs = ""
	ctx.JavascriptEnabled = false
	ctx.RewriteUrls = 0
	ctx.NumPrecision = 0
	ctx.InCalc = false
	ctx.MathOn = false
	mathEvalContextPool.Put(ctx)
}

// EvalContext represents the interface needed for evaluation context
type EvalContext interface {
	IsMathOn() bool
	SetMathOn(bool)
	IsInCalc() bool
	EnterCalc()
	ExitCalc()
	GetFrames() []ParserFrame
	GetImportantScope() []map[string]bool
	GetDefaultFunc() *DefaultFunc
}

// evalContextAdapter adapts EvalContext to runtime.EvalContextProvider
type evalContextAdapter struct {
	evalContext EvalContext
}

// GetFramesAny returns frames as []any for JavaScript serialization
func (a *evalContextAdapter) GetFramesAny() []any {
	frames := a.evalContext.GetFrames()
	result := make([]any, len(frames))
	for i, frame := range frames {
		result[i] = frame
	}
	return result
}

// GetImportantScopeAny returns important scope as []map[string]any for JavaScript serialization
func (a *evalContextAdapter) GetImportantScopeAny() []map[string]any {
	scope := a.evalContext.GetImportantScope()
	result := make([]map[string]any, len(scope))
	for i, m := range scope {
		result[i] = make(map[string]any)
		for k, v := range m {
			result[i][k] = v
		}
	}
	return result
}

// PluginFunctionProvider is an interface for looking up plugin functions.
// This allows the function caller to access JavaScript plugin functions.
type PluginFunctionProvider interface {
	LookupPluginFunction(name string) (any, bool)
	HasPluginFunction(name string) bool
	CallPluginFunction(name string, args ...any) (any, error)
}

// ParserFunctionCaller represents the interface needed to call functions
type ParserFunctionCaller interface {
	IsValid() bool
	Call(args []any) (any, error)
}

// FunctionCallerFactory creates function callers
type FunctionCallerFactory interface {
	NewFunctionCaller(name string, context EvalContext, index int, fileInfo map[string]any) (ParserFunctionCaller, error)
}

// DefaultFunctionCallerFactory implements FunctionCallerFactory using a registry
type DefaultFunctionCallerFactory struct {
	adapter *RegistryFunctionAdapter
}

// NewDefaultFunctionCallerFactory creates a new DefaultFunctionCallerFactory
func NewDefaultFunctionCallerFactory(registry *Registry) *DefaultFunctionCallerFactory {
	return &DefaultFunctionCallerFactory{
		adapter: NewRegistryFunctionAdapter(registry),
	}
}

// NewFunctionCaller creates a ParserFunctionCaller
func (f *DefaultFunctionCallerFactory) NewFunctionCaller(name string, context EvalContext, index int, fileInfo map[string]any) (ParserFunctionCaller, error) {
	// Get function definition from registry via adapter
	lowerName := strings.ToLower(name)
	funcDef := f.adapter.Get(lowerName)


	if funcDef == nil {
		// Check if this might be a plugin function
		if pluginProvider, ok := context.(PluginFunctionProvider); ok {
			if pluginProvider.HasPluginFunction(lowerName) {
				// Return a plugin function caller
				return &PluginFunctionCaller{
					name:     lowerName,
					provider: pluginProvider,
					context:  context,
					index:    index,
					fileInfo: fileInfo,
				}, nil
			}
		}

		// Return an invalid caller - this matches JavaScript behavior where unknown functions are not called
		return &DefaultParserFunctionCaller{
			name:     lowerName,
			valid:    false,
			funcDef:  nil,
			context:  context,
			index:    index,
			fileInfo: fileInfo,
		}, nil
	}

	// funcDef is already a FunctionDefinition from the adapter
	definition := funcDef

	return &DefaultParserFunctionCaller{
		name:     lowerName,
		valid:    true,
		funcDef:  definition,
		context:  context,
		index:    index,
		fileInfo: fileInfo,
	}, nil
}

// PluginFunctionCaller calls JavaScript plugin functions
type PluginFunctionCaller struct {
	name     string
	provider PluginFunctionProvider
	context  EvalContext
	index    int
	fileInfo map[string]any
}

// IsValid returns true because plugin functions are always valid if we get here
func (c *PluginFunctionCaller) IsValid() bool {
	return true
}

// Call calls the JavaScript plugin function with the given arguments
func (c *PluginFunctionCaller) Call(args []any) (any, error) {
	// Evaluate arguments first (match JavaScript behavior)
	evaluatedArgs := make([]any, 0, len(args))
	for _, arg := range args {
		if evaluable, ok := arg.(interface{ Eval(any) any }); ok {
			evaluatedArgs = append(evaluatedArgs, evaluable.Eval(c.context))
		} else if evaluableWithErr, ok := arg.(interface{ Eval(any) (any, error) }); ok {
			result, err := evaluableWithErr.Eval(c.context)
			if err != nil {
				return nil, err
			}
			evaluatedArgs = append(evaluatedArgs, result)
		} else {
			evaluatedArgs = append(evaluatedArgs, arg)
		}
	}

	// Call the plugin function
	result, err := c.provider.CallPluginFunction(c.name, evaluatedArgs...)
	if err != nil {
		return nil, fmt.Errorf("error calling plugin function '%s': %w", c.name, err)
	}

	// Convert JSResultNode to proper Go AST nodes
	converted := convertJSResultToAST(result, c.context)
	return converted, nil
}

// DefaultParserFunctionCaller implements ParserFunctionCaller
type DefaultParserFunctionCaller struct {
	name     string
	valid    bool
	funcDef  FunctionDefinition
	context  EvalContext
	index    int
	fileInfo map[string]any
}

// IsValid returns whether this caller has a valid function
func (c *DefaultParserFunctionCaller) IsValid() bool {
	return c.valid
}

// createMathEnabledContext creates a context with math enabled for function arguments.
// This matches JavaScript behavior where mathOn is set to true (for non-calc functions)
// but the math mode and parensStack logic is still respected.
// Returns (context, pooled) where pooled is true if the context came from a pool
// and must be returned via putMathEnabledEvalContext when done.
func (c *DefaultParserFunctionCaller) createMathEnabledContext() (any, bool) {
	// Handle *Eval context (the most common case)
	if evalCtx, ok := c.context.(*Eval); ok {
		// Get pooled context and copy fields from source
		// DON'T override Math mode - keep the original (strict, parens-division, etc.)
		// This ensures that in strict math mode, operations like 16/17 are not evaluated
		// unless they're in parentheses like (16/17)
		return getMathEnabledEvalContext(evalCtx), true
	}

	// Try to create a map context with math enabled
	if mapCtx, ok := c.context.(*MapEvalContext); ok {
		// Clone the context and enable math
		newCtx := make(map[string]any, len(mapCtx.ctx))
		for k, v := range mapCtx.ctx {
			newCtx[k] = v
		}
		// Set mathOn to true (matching JavaScript: context.mathOn = !this.calc)
		newCtx["mathOn"] = true

		// Initialize parensStack if not present (needed for division math check)
		if _, exists := newCtx["parensStack"]; !exists {
			newCtx["parensStack"] = []bool{}
		}

		// Create proper inParenthesis function that updates newCtx's parensStack
		newCtx["inParenthesis"] = func() {
			stack, _ := newCtx["parensStack"].([]bool)
			newCtx["parensStack"] = append(stack, true)
		}

		// Create proper outOfParenthesis function
		newCtx["outOfParenthesis"] = func() {
			if stack, ok := newCtx["parensStack"].([]bool); ok && len(stack) > 0 {
				newCtx["parensStack"] = stack[:len(stack)-1]
			}
		}

		// Create a new isMathOn function that references newCtx instead of the original
		// This preserves the math mode logic while using the updated mathOn value
		newCtx["isMathOn"] = func(op string) bool {
			debugTrace := os.Getenv("LESS_GO_TRACE") == "1"

			mathOn, exists := newCtx["mathOn"]
			if !exists || !mathOn.(bool) {
				if debugTrace {
					fmt.Printf("[MATH-DEBUG] isMathOn(%s): mathOn not set or false\n", op)
				}
				return false
			}

			// Check for division operator with math mode restrictions
			if op == "/" {
				math, mathExists := newCtx["math"]
				if mathExists && math != Math.Always {
					// Check if we're in parentheses
					parensStack, parensExists := newCtx["parensStack"]
					if !parensExists {
						if debugTrace {
							fmt.Printf("[MATH-DEBUG] isMathOn(/): no parensStack, returning false\n")
						}
						return false
					}
					if stack, ok := parensStack.([]bool); ok && len(stack) == 0 {
						if debugTrace {
							fmt.Printf("[MATH-DEBUG] isMathOn(/): empty parensStack, math=%v, returning false\n", math)
						}
						return false
					}
				}
			}

			// Check if math is disabled for everything except in parentheses
			if math, mathExists := newCtx["math"]; mathExists {
				if mathType, ok := math.(MathType); ok && mathType > Math.ParensDivision {
					parensStack, parensExists := newCtx["parensStack"]
					if !parensExists {
						if debugTrace {
							fmt.Printf("[MATH-DEBUG] isMathOn(%s): PARENS mode, no parensStack, returning false\n", op)
						}
						return false
					}
					if stack, ok := parensStack.([]bool); ok {
						result := len(stack) > 0
						if debugTrace {
							fmt.Printf("[MATH-DEBUG] isMathOn(%s): PARENS mode, parensStack len=%d, returning %v\n", op, len(stack), result)
						}
						return result
					}
					return false
				}
			}

			if debugTrace {
				fmt.Printf("[MATH-DEBUG] isMathOn(%s): returning true (default)\n", op)
			}
			return true
		}

		return newCtx, false
	}

	// For other EvalContext implementations, try to enable math if possible
	if evalCtx, ok := c.context.(EvalContext); ok && evalCtx != nil {
		// Can't clone easily, but we can try to modify if it's a mutable reference
		// For now, just return the original - this is a fallback case
		return evalCtx, false
	}

	// Fallback: return the original context
	return c.context, false
}

// Call executes the function with the given arguments
func (c *DefaultParserFunctionCaller) Call(args []any) (any, error) {
	if !c.valid || c.funcDef == nil {
		return nil, fmt.Errorf("function %s is not valid", c.name)
	}

	// Determine if we need to evaluate arguments
	needsEval := c.funcDef.NeedsEvalArgs()

	// Create a simplified context for function calling if needed
	if !needsEval {
		// For functions that don't need evaluated args, create a context with proper EvalContext
		// We need a registry that contains this function
		tempRegistry := NewRegistryFunctionAdapter(DefaultRegistry.Inherit())
		tempRegistry.registry.Add(c.name, c.funcDef)

		funcContext := &Context{
			Frames: []*Frame{
				{
					FunctionRegistry: tempRegistry,
					EvalContext:      c.context,   // Pass the evaluation context for variable resolution
					CurrentFileInfo:  c.fileInfo,  // Pass the current file information
				},
			},
		}

		return c.funcDef.CallCtx(funcContext, args...)
	}

	// For functions that need evaluated args, evaluate them first
	// Function arguments should always be evaluated with math enabled
	// because functions operate on computed values
	mathCtx, pooled := c.createMathEnabledContext()
	if pooled {
		defer putMathEnabledEvalContext(mathCtx.(*Eval))
	}

	evaluatedArgs := make([]any, len(args))
	for i, arg := range args {
		var evalResult any

		// Try all Eval signatures, always using mathCtx for consistent math evaluation
		// This ensures that function arguments are evaluated with math enabled
		if evalable, ok := arg.(interface {
			Eval(any) (any, error)
		}); ok {
			// Use the math-enabled context for all evaluations
			var err error
			evalResult, err = evalable.Eval(mathCtx)
			if err != nil {
				return nil, fmt.Errorf("error evaluating argument %d: %w", i, err)
			}
		} else if evalable, ok := arg.(interface {
			Eval(any) any
		}); ok {
			// Handle nodes with single-return Eval (like Paren, DetachedRuleset, etc.)
			// Use the math-enabled context
			evalResult = evalable.Eval(mathCtx)
		} else {
			evalResult = arg
		}

		// Unwrap Paren nodes to get the inner value
		// This matches JavaScript behavior where parens are transparent to functions
		if paren, ok := evalResult.(*Paren); ok {
			evalResult = paren.Value
		}

		evaluatedArgs[i] = evalResult
	}

	// Filter comments from evaluated arguments (matches JavaScript behavior)
	// This handles cases like @color2: #FFF/* comment2 */;
	filteredArgs := c.filterCommentsFromArgs(evaluatedArgs)

	return c.funcDef.Call(filteredArgs...)
}

// filterCommentsFromArgs removes comments from evaluated arguments
// This is specifically for handling inline comments in variable values
func (c *DefaultParserFunctionCaller) filterCommentsFromArgs(args []any) []any {
	isComment := func(node any) bool {
		if comment, ok := node.(*Comment); ok {
			return comment != nil
		}
		if hasType, ok := node.(interface{ GetType() string }); ok {
			return hasType.GetType() == "Comment"
		}
		return false
	}

	filtered := make([]any, 0, len(args))
	for _, arg := range args {
		// Skip top-level comments
		if isComment(arg) {
			continue
		}

		// Handle Expression nodes that might contain comments
		if expr, ok := arg.(*Expression); ok {
			// Filter comments from expression value
			hasComments := false
			for _, val := range expr.Value {
				if isComment(val) {
					hasComments = true
					break
				}
			}

			if hasComments {
				// Create new expression without comments
				nonCommentVals := make([]any, 0, len(expr.Value))
				for _, val := range expr.Value {
					if !isComment(val) {
						nonCommentVals = append(nonCommentVals, val)
					}
				}

				if len(nonCommentVals) == 1 {
					// Single value, return it directly
					filtered = append(filtered, nonCommentVals[0])
				} else if len(nonCommentVals) > 0 {
					// Multiple values, create new expression
					newExpr := &Expression{
						Node:       NewNode(),
						Value:      nonCommentVals,
						Parens:     expr.Parens,
						ParensInOp: expr.ParensInOp,
					}
					newExpr.Node.Index = expr.GetIndex()
					newExpr.Node.SetFileInfo(expr.FileInfo())
					filtered = append(filtered, newExpr)
				}
				// Skip if all values were comments
			} else {
				// No comments, keep as-is
				filtered = append(filtered, arg)
			}
		} else {
			// Not a comment or expression, keep as-is
			filtered = append(filtered, arg)
		}
	}

	return filtered
}

// MapEvalContext implements EvalContext for map[string]any contexts
type MapEvalContext struct {
	ctx map[string]any
}

// IsMathOn returns whether math operations are enabled
func (m *MapEvalContext) IsMathOn() bool {
	if mathOn, exists := m.ctx["mathOn"]; exists {
		if enabled, ok := mathOn.(bool); ok {
			return enabled
		}
	}
	return false
}

// SetMathOn sets whether math operations are enabled
func (m *MapEvalContext) SetMathOn(enabled bool) {
	m.ctx["mathOn"] = enabled
}

// IsInCalc returns whether we're inside a calc() function
func (m *MapEvalContext) IsInCalc() bool {
	if inCalc, exists := m.ctx["inCalc"]; exists {
		if enabled, ok := inCalc.(bool); ok {
			return enabled
		}
	}
	return false
}

// EnterCalc marks that we're entering a calc() function
func (m *MapEvalContext) EnterCalc() {
	m.ctx["inCalc"] = true
}

// ExitCalc marks that we're exiting a calc() function
func (m *MapEvalContext) ExitCalc() {
	m.ctx["inCalc"] = false
}

// GetFrames returns the current frames stack
func (m *MapEvalContext) GetFrames() []ParserFrame {
	if framesAny, exists := m.ctx["frames"]; exists {
		if frameSlice, ok := framesAny.([]any); ok {
			frames := make([]ParserFrame, 0, len(frameSlice))
			for _, f := range frameSlice {
				if frame, ok := f.(ParserFrame); ok {
					frames = append(frames, frame)
				}
			}
			return frames
		}
	}
	return []ParserFrame{}
}

// GetImportantScope returns the current important scope stack
func (m *MapEvalContext) GetImportantScope() []map[string]bool {
	if importantScope, exists := m.ctx["importantScope"]; exists {
		if scope, ok := importantScope.([]map[string]bool); ok {
			return scope
		}
	}
	return []map[string]bool{}
}

// GetDefaultFunc returns the default function instance
func (m *MapEvalContext) GetDefaultFunc() *DefaultFunc {
	if defaultFunc, exists := m.ctx["defaultFunc"]; exists {
		if df, ok := defaultFunc.(*DefaultFunc); ok {
			return df
		}
	}
	return nil
}

// RegistryAdapter adapts a single FunctionDefinition to the FunctionRegistry interface
type RegistryAdapter struct {
	registry FunctionDefinition
	name     string
}

// Get implements FunctionRegistry interface
func (r *RegistryAdapter) Get(name string) FunctionDefinition {
	if strings.EqualFold(name, r.name) {
		return r.registry
	}
	return nil
}

// Call represents a function call node in the Less AST.
type Call struct {
	*Node
	Name          string
	Args          []any
	Calc          bool
	_index        int
	_fileInfo     map[string]any
	CallerFactory FunctionCallerFactory // Factory for creating FunctionCaller instances
}

// NewCall creates a new Call instance.
func NewCall(name string, args []any, index int, currentFileInfo map[string]any) *Call {
	return &Call{
		Node:      NewNode(),
		Name:      name,
		Args:      args,
		Calc:      name == "calc",
		_index:    index,
		_fileInfo: currentFileInfo,
	}
}

// GetType returns the node type.
func (c *Call) GetType() string {
	return "Call"
}

// GetName returns the function name.
func (c *Call) GetName() string {
	return c.Name
}

// Accept processes the node's children with a visitor.
func (c *Call) Accept(visitor any) {
	if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok && c.Args != nil {
		c.Args = v.VisitArray(c.Args)
	}
}

// GetIndex returns the node's index.
func (c *Call) GetIndex() int {
	return c._index
}

// FileInfo returns the node's file information.
func (c *Call) FileInfo() map[string]any {
	return c._fileInfo
}

// Eval evaluates the function call.
func (c *Call) Eval(context any) (any, error) {
	// Convert context to EvalContext if it's a map
	var evalContext EvalContext
	if ctx, ok := context.(EvalContext); ok {
		evalContext = ctx
	} else if ctxMap, ok := context.(map[string]any); ok {
		// Create a bridge EvalContext from the map
		evalContext = &MapEvalContext{ctx: ctxMap}
	} else {
		return nil, fmt.Errorf("invalid context type: %T", context)
	}

	// Set up CallerFactory from context if not already set
	if c.CallerFactory == nil {
		if evalCtx, ok := context.(*Eval); ok {
			// Try to get function registry from *Eval context
			if evalCtx.FunctionRegistry != nil {
				c.CallerFactory = NewDefaultFunctionCallerFactory(evalCtx.FunctionRegistry)
			} else if len(evalCtx.Frames) > 0 {
				// Try to get from frames
				for _, frame := range evalCtx.Frames {
					if ruleset, ok := frame.(*Ruleset); ok && ruleset.FunctionRegistry != nil {
						if registry, ok := ruleset.FunctionRegistry.(*Registry); ok {
							c.CallerFactory = NewDefaultFunctionCallerFactory(registry)
							break
						}
					}
				}
			}
		} else if ctxMap, ok := context.(map[string]any); ok {
			// Try to get function registry from context
			if funcRegistry, exists := ctxMap["functionRegistry"]; exists {
				if registry, ok := funcRegistry.(*Registry); ok {
					c.CallerFactory = NewDefaultFunctionCallerFactory(registry)
				}
			} else {
				// Try to get from frames
				if frames, exists := ctxMap["frames"]; exists {
					if frameList, ok := frames.([]any); ok {
						for _, frame := range frameList {
							if ruleset, ok := frame.(*Ruleset); ok && ruleset.FunctionRegistry != nil {
								if registry, ok := ruleset.FunctionRegistry.(*Registry); ok {
									c.CallerFactory = NewDefaultFunctionCallerFactory(registry)
									break
								}
							}
						}
					}
				}
			}
		}
		// If still nil, use default registry
		if c.CallerFactory == nil {
			c.CallerFactory = NewDefaultFunctionCallerFactory(DefaultRegistry)
		}
	}
	// Turn off math for calc(), and switch back on for evaluating nested functions
	// Match JavaScript: save the mathOn FIELD value, not the computed result from isMathOn()
	var currentMathContext bool
	if evalCtx, ok := evalContext.(*Eval); ok {
		currentMathContext = evalCtx.MathOn
	} else if mapCtx, ok := evalContext.(*MapEvalContext); ok {
		// For map context, get the raw mathOn value from the map
		if mathOn, exists := mapCtx.ctx["mathOn"]; exists {
			if enabled, ok := mathOn.(bool); ok {
				currentMathContext = enabled
			}
		}
	} else {
		// Fallback for other context types
		currentMathContext = evalContext.IsMathOn()
	}
	evalContext.SetMathOn(!c.Calc)

	if c.Calc || evalContext.IsInCalc() {
		evalContext.EnterCalc()
	}

	exitCalc := func() {
		if c.Calc || evalContext.IsInCalc() {
			evalContext.ExitCalc()
		}
		evalContext.SetMathOn(currentMathContext)
	}

	var result any
	var err error

	// Check if we have a function caller factory
	if c.CallerFactory == nil {
		// No Go function caller - try JS plugin functions first
		result, err = c.tryJSPluginFunction(context, evalContext)
		if err != nil {
			exitCalc()
			return nil, err
		}
		if result != nil {
			// JS plugin function returned a result
			exitCalc()
			return result, nil
		}

		// No JS plugin function either - evaluate args and return as CSS function call
		exitCalc()
		evaledArgs := make([]any, len(c.Args))
		for i, arg := range c.Args {
			if evalable, ok := arg.(interface{ Eval(any) (any, error) }); ok {
				evaledVal, err := evalable.Eval(context)
				if err != nil {
					return nil, err
				}
				evaledArgs[i] = evaledVal
			} else {
				evaledArgs[i] = arg
			}
		}
		return NewCall(c.Name, evaledArgs, c.GetIndex(), c.FileInfo()), nil
	}

	funcCaller, err := c.CallerFactory.NewFunctionCaller(c.Name, evalContext, c.GetIndex(), c.FileInfo())
	if err != nil {
		exitCalc()
		return nil, err
	}

	if funcCaller.IsValid() {
		// Preprocess arguments to match JavaScript behavior
		processedArgs := c.preprocessArgs(c.Args)
		result, err = funcCaller.Call(processedArgs)
	} else {
		// Function not found in Go registry - try JS plugin functions
		result, err = c.tryJSPluginFunction(context, evalContext)
		if err != nil {
			exitCalc()
			return nil, err
		}
	}

	if result != nil || funcCaller.IsValid() {

		if err != nil {
			exitCalc()
			// Check if error has line and column properties
			if e, ok := err.(interface{ HasLineColumn() bool }); ok && e.HasLineColumn() {
				return nil, err
			}

			var lineNumber, columnNumber int
			e, ok := err.(interface {
				LineNumber() int
				ColumnNumber() int
			})
			if ok {
				lineNumber = e.LineNumber()
				columnNumber = e.ColumnNumber()
			}

			errorType := "Runtime"
			// Check for GetErrorType() method first (used by LessError)
			if typedErr, ok := err.(interface{ GetErrorType() string }); ok {
				errorType = typedErr.GetErrorType()
			} else if typedErr, ok := err.(interface{ Type() string }); ok {
				// Fallback to Type() for other error types
				errorType = typedErr.Type()
			}

			errorMsg := fmt.Sprintf("Error evaluating function `%s`", c.Name)
			if err.Error() != "" {
				errorMsg += fmt.Sprintf(": %s", err.Error())
			}

			// Return a *LessError to preserve type for proper error propagation
			lessErr := &LessError{
				Type:     errorType,
				Message:  errorMsg,
				Index:    c.GetIndex(),
				Filename: c.FileInfo()["filename"].(string),
				Column:   columnNumber,
			}
			if lineNumber > 0 {
				lessErr.Line = &lineNumber
			}
			return nil, lessErr
		}
		exitCalc()
		
		// Check if result is a LessError and return it as an error
		if result != nil {
			if lessErr, ok := result.(*LessError); ok {
				return nil, lessErr
			}
		}

		if result != nil {
			// Results that are not nodes are cast as Anonymous nodes
			// Falsy values or booleans are returned as empty nodes
			// Check if result implements common node interfaces or is a known node type
			isNodeType := false
			switch result.(type) {
			case *Node, *Color, *Dimension, *Quoted, *Anonymous, *Keyword, *Value, *Expression, *Call, *Ruleset, *Declaration:
				isNodeType = true
			case interface{ GetType() string }:
				// Has GetType method, likely a node
				isNodeType = true
			}
			
			if !isNodeType {
				// Check for falsy values or true - these should return empty Anonymous nodes
				// JavaScript behavior: if (!result || result === true)
				if result == nil || result == false || result == true || result == "" {
					result = NewAnonymous(nil, 0, nil, false, false, nil)
				} else {
					// Non-falsy values are converted to string
					result = NewAnonymous(fmt.Sprintf("%v", result), 0, nil, false, false, nil)
				}
			}

			// Set index and file info
			resultNode, ok := result.(interface {
				SetIndex(int)
				SetFileInfo(map[string]any)
			})
			if ok {
				resultNode.SetIndex(c._index)
				resultNode.SetFileInfo(c._fileInfo)
			}
			return result, nil
		}
	}

	// If function not found or no result, evaluate args and return new Call
	// This matches JavaScript behavior (line 93 in call.js)
	evaledArgs := make([]any, len(c.Args))
	for i, arg := range c.Args {
		if evalable, ok := arg.(interface{ Eval(any) (any, error) }); ok {
			evaledVal, err := evalable.Eval(context)
			if err != nil {
				exitCalc()
				return nil, err
			}
			evaledArgs[i] = evaledVal
		} else {
			evaledArgs[i] = arg
		}
	}
	
	// Important: exit calc AFTER evaluating arguments
	exitCalc()

	return NewCall(c.Name, evaledArgs, c.GetIndex(), c.FileInfo()), nil
}

// tryJSPluginFunction attempts to call a JavaScript plugin function.
// Returns (nil, nil) if no plugin function was found.
func (c *Call) tryJSPluginFunction(context any, evalContext EvalContext) (any, error) {
	// Try to get pluginBridge from context
	var pluginBridge *NodeJSPluginBridge
	var lazyBridge *LazyNodeJSPluginBridge

	debug := os.Getenv("LESS_GO_DEBUG") == "1"
	if debug {
		fmt.Printf("[tryJSPluginFunction] Checking function '%s', context type: %T\n", c.Name, context)
	}

	// Check Eval context - check both PluginBridge and LazyPluginBridge
	if evalCtx, ok := context.(*Eval); ok {
		if evalCtx.PluginBridge != nil {
			pluginBridge = evalCtx.PluginBridge
			if debug {
				fmt.Printf("[tryJSPluginFunction] Found PluginBridge in Eval context\n")
			}
		} else if evalCtx.LazyPluginBridge != nil {
			lazyBridge = evalCtx.LazyPluginBridge
			if debug {
				fmt.Printf("[tryJSPluginFunction] Found LazyPluginBridge in Eval context (initialized=%v)\n", lazyBridge.IsInitialized())
			}
		} else if debug {
			fmt.Printf("[tryJSPluginFunction] Eval context has nil PluginBridge and nil LazyPluginBridge\n")
		}
	}

	// Check map context for various bridge types
	if pluginBridge == nil && lazyBridge == nil {
		if ctxMap, ok := context.(map[string]any); ok {
			if pb, exists := ctxMap["pluginBridge"]; exists {
				if lazy, ok := pb.(*LazyNodeJSPluginBridge); ok {
					lazyBridge = lazy
					if debug {
						fmt.Printf("[tryJSPluginFunction] Found LazyPluginBridge in map context\n")
					}
				} else if bridge, ok := pb.(*NodeJSPluginBridge); ok {
					pluginBridge = bridge
					if debug {
						fmt.Printf("[tryJSPluginFunction] Found PluginBridge in map context\n")
					}
				}
			}
		}
	}

	// No plugin bridge available
	if pluginBridge == nil && lazyBridge == nil {
		if debug {
			fmt.Printf("[tryJSPluginFunction] No plugin bridge available for '%s'\n", c.Name)
		}
		return nil, nil
	}

	// Use the lazy bridge if available (it handles initialization internally)
	if lazyBridge != nil {
		// Check if this function exists - this will return false if bridge isn't initialized yet
		if !lazyBridge.HasFunction(c.Name) {
			if debug {
				fmt.Printf("[tryJSPluginFunction] Function '%s' not found in lazy plugin bridge\n", c.Name)
			}
			return nil, nil
		}

		if debug {
			fmt.Printf("[tryJSPluginFunction] Calling JS function '%s' via LazyBridge\n", c.Name)
		}

		// Evaluate arguments
		evaledArgs := make([]any, 0, len(c.Args))
		for i, arg := range c.Args {
			// Filter out comments
			if c.isComment(arg) {
				continue
			}

			if evalable, ok := arg.(interface{ Eval(any) (any, error) }); ok {
				evaledVal, err := evalable.Eval(context)
				if err != nil {
					if debug {
						fmt.Printf("[tryJSPluginFunction] Error evaluating arg %d for '%s': %v\n", i, c.Name, err)
					}
					return nil, err
				}
				if debug {
					fmt.Printf("[tryJSPluginFunction] Evaluated arg %d for '%s': %T = %+v\n", i, c.Name, evaledVal, evaledVal)
				}
				evaledArgs = append(evaledArgs, evaledVal)
			} else {
				evaledArgs = append(evaledArgs, arg)
			}
		}

		// Call the JS function via the lazy bridge with context for variable access
		var result any
		var err error
		// Try to use context-aware call for plugin functions that need variable lookup
		if evalCtx, ok := context.(*Eval); ok {
			// Use CallFunctionWithContext to pass evaluation context for variable lookup
			result, err = lazyBridge.CallFunctionWithContext(c.Name, evalCtx, evaledArgs...)
		} else if evalContext != nil {
			// Create adapter for EvalContext interface
			adapter := &evalContextAdapter{evalContext: evalContext}
			result, err = lazyBridge.CallFunctionWithContext(c.Name, adapter, evaledArgs...)
		} else {
			// Fallback to regular call without context
			result, err = lazyBridge.CallFunction(c.Name, evaledArgs...)
		}
		if err != nil {
			if debug {
				fmt.Printf("[tryJSPluginFunction] JS function '%s' returned error: %v\n", c.Name, err)
			}
			return nil, err
		}
		if debug {
			fmt.Printf("[tryJSPluginFunction] JS function '%s' returned: %T = %+v\n", c.Name, result, result)
		}
		// Convert JSResultNode to proper Go AST nodes
		converted := convertJSResultToAST(result, context)
		if debug && converted != result {
			fmt.Printf("[tryJSPluginFunction] Converted result to: %T\n", converted)
		}
		return converted, nil
	}

	// Check if this function exists in the plugin bridge
	if !pluginBridge.HasFunction(c.Name) {
		if debug {
			fmt.Printf("[tryJSPluginFunction] Function '%s' not found in plugin bridge\n", c.Name)
		}
		return nil, nil
	}

	if debug {
		fmt.Printf("[tryJSPluginFunction] Calling JS function '%s'\n", c.Name)
	}

	// Evaluate arguments
	evaledArgs := make([]any, 0, len(c.Args))
	for _, arg := range c.Args {
		// Filter out comments
		if c.isComment(arg) {
			continue
		}

		if evalable, ok := arg.(interface{ Eval(any) (any, error) }); ok {
			evaledVal, err := evalable.Eval(context)
			if err != nil {
				return nil, err
			}
			evaledArgs = append(evaledArgs, evaledVal)
		} else {
			evaledArgs = append(evaledArgs, arg)
		}
	}

	// Call the JS function with context for variable access
	var result any
	var err error
	// Try to use context-aware call for plugin functions that need variable lookup
	if evalCtx, ok := context.(*Eval); ok {
		// Use CallFunctionWithContext to pass evaluation context for variable lookup
		result, err = pluginBridge.CallFunctionWithContext(c.Name, evalCtx, evaledArgs...)
	} else if evalContext != nil {
		// Create adapter for EvalContext interface
		adapter := &evalContextAdapter{evalContext: evalContext}
		result, err = pluginBridge.CallFunctionWithContext(c.Name, adapter, evaledArgs...)
	} else {
		// Fallback to regular call without context
		result, err = pluginBridge.CallFunction(c.Name, evaledArgs...)
	}
	if err != nil {
		return nil, err
	}

	// Convert JSResultNode to proper Go AST nodes
	converted := convertJSResultToAST(result, context)
	if debug && converted != result {
		fmt.Printf("[tryJSPluginFunction] Converted result to: %T\n", converted)
	}
	return converted, nil
}

// GenCSS generates CSS representation of the function call.
func (c *Call) GenCSS(context any, output *CSSOutput) {
	// Special case: _SELF should output its argument directly, not as a function call
	if c.Name == "_SELF" && len(c.Args) > 0 {
		if genCSSable, ok := c.Args[0].(interface{ GenCSS(any, *CSSOutput) }); ok {
			genCSSable.GenCSS(context, output)
			return
		}
	}
	
	// Special case: alpha() function for IE compatibility
	if c.Name == "alpha" && len(c.Args) == 1 {
		if assignment, ok := c.Args[0].(*Assignment); ok && assignment.Key == "opacity" {
			output.Add("alpha(opacity=", c.FileInfo(), c.GetIndex())
			if genCSSable, ok := assignment.Value.(interface{ GenCSS(any, *CSSOutput) }); ok {
				genCSSable.GenCSS(context, output)
			}
			output.Add(")", nil, nil)
			return
		}
	}
	
	output.Add(c.Name+"(", c.FileInfo(), c.GetIndex())

	for i, arg := range c.Args {
		if genCSSable, ok := arg.(interface{ GenCSS(any, *CSSOutput) }); ok {
			genCSSable.GenCSS(context, output)
			if i+1 < len(c.Args) {
				output.Add(", ", nil, nil)
			}
		}
	}

	output.Add(")", nil, nil)
}

// preprocessArgs processes arguments to match JavaScript behavior
// JavaScript does: args.filter(commentFilter).map(item => ...)
func (c *Call) preprocessArgs(args []any) []any {
	if args == nil {
		return []any{}
	}

	processed := make([]any, 0, len(args))

	for _, arg := range args {
		// Filter out comments (JavaScript uses commentFilter)
		if c.isComment(arg) {
			continue
		}
		
		// Process expressions - flatten single-item expressions
		if expr, ok := arg.(*Expression); ok {
			// Filter out comments from expression value
			subNodes := make([]any, 0, len(expr.Value))
			for _, subNode := range expr.Value {
				if !c.isComment(subNode) {
					subNodes = append(subNodes, subNode)
				}
			}

			if len(subNodes) == 1 {
				// Special handling for parens and division (JavaScript logic)
				// JavaScript: if (a.parens && a.value[0].op === '/')
				// See: https://github.com/less/less.js/issues/3616
				if expr.Parens {
					if op, ok := subNodes[0].(*Operation); ok && op.Op == "/" {
						// Keep the expression with parens for division
						processed = append(processed, arg)
						continue
					}
				}
				// Return the single sub-node
				processed = append(processed, subNodes[0])
			} else if len(subNodes) > 0 {
				// Create new expression with filtered nodes
				newExpr := &Expression{
					Node:       NewNode(),
					Value:      subNodes,
					ParensInOp: expr.ParensInOp,
					Parens:     expr.Parens,
				}
				newExpr.Node.Index = expr.GetIndex()
				newExpr.Node.SetFileInfo(expr.FileInfo())
				processed = append(processed, newExpr)
			}
			// Skip empty expressions (all comments filtered out)
		} else {
			// Non-expression argument, add as-is
			processed = append(processed, arg)
		}
	}
	
	return processed
}

// isComment checks if a node is a comment
func (c *Call) isComment(node any) bool {
	if comment, ok := node.(*Comment); ok {
		return comment != nil
	}
	// Check for Node with comment type
	if hasType, ok := node.(interface{ GetType() string }); ok {
		return hasType.GetType() == "Comment"
	}
	return false
}

// convertJSResultToAST converts a JSResultNode to proper Go AST nodes.
// This is needed because JS functions return generic JSResultNode objects
// that need to be converted to actual AST node types for proper evaluation.
func convertJSResultToAST(result any, context any) any {
	jsNode, ok := result.(*runtime.JSResultNode)
	if !ok {
		return result
	}

	debug := os.Getenv("LESS_GO_DEBUG") == "1"

	switch jsNode.NodeType {
	case "DetachedRuleset":
		// Convert to Go *DetachedRuleset
		rulesetData, ok := jsNode.Properties["ruleset"].(map[string]any)
		if !ok {
			if debug {
				fmt.Printf("[convertJSResultToAST] DetachedRuleset missing ruleset property\n")
			}
			return result
		}
		ruleset := convertRulesetData(rulesetData, context)
		if ruleset == nil {
			if debug {
				fmt.Printf("[convertJSResultToAST] Failed to convert ruleset data\n")
			}
			return result
		}
		// Get frames from context if available
		var frames []any
		if evalCtx, ok := context.(*Eval); ok {
			frames = evalCtx.Frames
		}
		detached := NewDetachedRuleset(ruleset, frames)
		if debug {
			fmt.Printf("[convertJSResultToAST] Created DetachedRuleset with %d rules\n", len(ruleset.Rules))
		}
		return detached

	case "Anonymous":
		value := jsNode.Properties["value"]
		if value == nil {
			value = ""
		}
		return NewAnonymous(fmt.Sprintf("%v", value), 0, nil, false, false, nil)

	case "Dimension":
		val := 0.0
		if v, ok := jsNode.Properties["value"].(float64); ok {
			val = v
		}
		unit := ""
		if u, ok := jsNode.Properties["unit"].(string); ok {
			unit = u
		}
		// Create proper unit - use empty numerator for no unit
		var dimensionUnit *Unit
		if unit != "" {
			dimensionUnit = &Unit{Numerator: []string{unit}}
		} else {
			dimensionUnit = &Unit{Numerator: []string{}, Denominator: []string{}}
		}
		return NewDimensionFrom(val, dimensionUnit)

	case "Keyword":
		value := ""
		if v, ok := jsNode.Properties["value"].(string); ok {
			value = v
		}
		return NewKeyword(value)

	case "Quoted":
		value := ""
		quote := "\""
		escaped := false
		if v, ok := jsNode.Properties["value"].(string); ok {
			value = v
		}
		if q, ok := jsNode.Properties["quote"].(string); ok {
			quote = q
		}
		if e, ok := jsNode.Properties["escaped"].(bool); ok {
			escaped = e
		}
		return NewQuoted(quote, value, escaped, 0, nil)

	case "Color":
		rgb := []float64{0, 0, 0}
		alpha := 1.0
		value := ""
		if r, ok := jsNode.Properties["rgb"].([]any); ok && len(r) >= 3 {
			for i := 0; i < 3 && i < len(r); i++ {
				if v, ok := r[i].(float64); ok {
					rgb[i] = v
				}
			}
		}
		if a, ok := jsNode.Properties["alpha"].(float64); ok {
			alpha = a
		}
		// Preserve the original color value (e.g., "#fff") for proper CSS output
		if v, ok := jsNode.Properties["value"].(string); ok {
			value = v
		}
		return NewColor(rgb, alpha, value)

	case "AtRule":
		name := ""
		if n, ok := jsNode.Properties["name"].(string); ok {
			name = n
		}
		value := jsNode.Properties["value"]
		// Convert value if needed
		if valueStr, ok := value.(string); ok {
			value = NewAnonymous(valueStr, 0, nil, false, false, nil)
		}
		if debug {
			fmt.Printf("[convertJSResultToAST] Creating AtRule: name=%s, value=%v\n", name, value)
		}
		return NewAtRule(name, value, nil, 0, nil, nil, false, nil)

	case "Combinator":
		value := " "
		if v, ok := jsNode.Properties["value"].(string); ok {
			value = v
		}
		return NewCombinator(value)

	case "Value":
		// Convert Value's children array to Go values
		if valueArr, ok := jsNode.Properties["value"].([]any); ok {
			convertedValues := make([]any, 0, len(valueArr))
			for _, item := range valueArr {
				if itemMap, ok := item.(map[string]any); ok {
					if nodeType, ok := itemMap["_type"].(string); ok {
						// Convert nested nodes
						nestedNode := &runtime.JSResultNode{
							NodeType:   nodeType,
							Properties: itemMap,
						}
						converted := convertJSResultToAST(nestedNode, context)
						convertedValues = append(convertedValues, converted)
					} else {
						convertedValues = append(convertedValues, item)
					}
				} else {
					convertedValues = append(convertedValues, item)
				}
			}
			if len(convertedValues) > 0 {
				val, err := NewValue(convertedValues)
				if err == nil {
					return val
				}
			}
		}
		return result

	default:
		// For unhandled types, return as-is
		if debug {
			fmt.Printf("[convertJSResultToAST] Unhandled node type: %s\n", jsNode.NodeType)
		}
		return result
	}
}

// convertRulesetData converts a ruleset map to a *Ruleset
func convertRulesetData(data map[string]any, context any) *Ruleset {
	var selectors []any
	var rules []any

	// Convert selectors
	if sels, ok := data["selectors"].([]any); ok {
		for _, sel := range sels {
			if selData, ok := sel.(map[string]any); ok {
				selector := convertSelectorData(selData)
				if selector != nil {
					selectors = append(selectors, selector)
				}
			}
		}
	}

	// Convert rules
	if rs, ok := data["rules"].([]any); ok {
		for _, r := range rs {
			if ruleData, ok := r.(map[string]any); ok {
				rule := convertRuleData(ruleData, context)
				if rule != nil {
					rules = append(rules, rule)
				}
			}
		}
	}

	return NewRuleset(selectors, rules, false, nil)
}

// convertSelectorData converts a selector map to a *Selector
func convertSelectorData(data map[string]any) *Selector {
	var elements []*Element

	if elems, ok := data["elements"].([]any); ok {
		for _, e := range elems {
			if elemData, ok := e.(map[string]any); ok {
				elem := convertElementData(elemData)
				if elem != nil {
					elements = append(elements, elem)
				}
			}
		}
	}

	selector, _ := NewSelector(elements, nil, nil, 0, nil, nil)
	return selector
}

// convertElementData converts an element map to an *Element
func convertElementData(data map[string]any) *Element {
	combinator := " "
	value := ""

	if c, ok := data["combinator"].(map[string]any); ok {
		if v, ok := c["value"].(string); ok {
			combinator = v
		}
	} else if c, ok := data["combinator"].(string); ok {
		combinator = c
	}

	if v, ok := data["value"].(string); ok {
		value = v
	}

	return NewElement(combinator, value, false, 0, nil, nil)
}

// convertRuleData converts a rule map to an AST node
func convertRuleData(data map[string]any, context any) any {
	nodeType, ok := data["_type"].(string)
	if !ok {
		return nil
	}

	switch nodeType {
	case "Declaration":
		name := ""
		value := ""
		if n, ok := data["name"].(string); ok {
			name = n
		}
		if v, ok := data["value"].(string); ok {
			value = v
		} else if vMap, ok := data["value"].(map[string]any); ok {
			// Value is a complex node - convert it
			converted := convertJSResultToAST(&runtime.JSResultNode{
				NodeType:   vMap["_type"].(string),
				Properties: vMap,
			}, context)
			if converted != nil {
				// Create declaration with converted value
				decl, _ := NewDeclaration(name, converted, false, false, 0, nil, false, nil)
				return decl
			}
		}
		// Simple string value
		decl, _ := NewDeclaration(name, NewAnonymous(value, 0, nil, false, false, nil), false, false, 0, nil, false, nil)
		return decl

	case "Ruleset":
		return convertRulesetData(data, context)

	default:
		return nil
	}
}
