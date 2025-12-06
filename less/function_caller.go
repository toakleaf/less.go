package less_go

// Context represents the evaluation context.
// This is a simplified representation; more fields might be needed.
type Context struct {
	Frames []*Frame
	// Add other context fields as needed (e.g., isMathOn, inParenthesis)
}

// GetFrames implements the interface expected by Variable.Eval
// It delegates to the underlying EvalContext if available
func (c *Context) GetFrames() []ParserFrame {
	if c == nil || len(c.Frames) == 0 {
		return []ParserFrame{}
	}
	// Get frames from the first Frame's EvalContext
	if c.Frames[0].EvalContext != nil {
		if ec, ok := c.Frames[0].EvalContext.(interface{ GetFrames() []ParserFrame }); ok {
			return ec.GetFrames()
		}
	}
	return []ParserFrame{}
}

// Frame represents a scope frame.
type Frame struct {
	FunctionRegistry FunctionRegistry
	variables        map[string]any // Add variable storage
	EvalContext      EvalContext    // Reference to the evaluation context
	CurrentFileInfo  map[string]any // Current file information for this frame
}

// Variable gets a variable from the frame
func (f *Frame) Variable(name string) any {
	if f.variables == nil {
		return nil
	}
	return f.variables[name]
}

// SetVariable sets a variable in the frame
func (f *Frame) SetVariable(name string, value any) {
	if f.variables == nil {
		f.variables = make(map[string]any)
	}
	f.variables[name] = value
}

// FunctionRegistry provides access to registered functions.
type FunctionRegistry interface {
	Get(name string) FunctionDefinition
}

// FunctionDefinition defines a Less function.
type FunctionDefinition interface {
	// Call handles functions where args are evaluated (evalArgs=true, default)
	Call(args ...any) (any, error)
	// CallCtx handles functions where args are not evaluated (evalArgs=false)
	CallCtx(ctx *Context, args ...any) (any, error)
	// NeedsEvalArgs returns true if arguments should be evaluated before calling.
	NeedsEvalArgs() bool
}

// NodeWithType defines types that have a GetType method.
type NodeWithType interface {
	GetType() string
}

// NodeWithOp defines types that have an GetOp method.
type NodeWithOp interface {
	GetOp() string
}

// NodeWithParens defines types that have a GetParens method.
type NodeWithParens interface {
	GetParens() bool
}
