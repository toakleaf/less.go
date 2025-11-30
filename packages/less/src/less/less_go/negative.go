package less_go

// Negative represents a negative node in the Less AST
type Negative struct {
	*Node
	Value any
}

// NewNegative creates a new Negative instance
func NewNegative(node any) *Negative {
	return &Negative{
		Node:  NewNode(),
		Value: node,
	}
}

// Type returns the type of the node
func (n *Negative) Type() string {
	return "Negative"
}

// GetType returns the type of the node for visitor pattern consistency
func (n *Negative) GetType() string {
	return "Negative"
}

// GenCSS generates CSS representation
func (n *Negative) GenCSS(context any, output *CSSOutput) {
	output.Add("-", nil, nil)
	if valueWithGenCSS, ok := n.Value.(interface{ GenCSS(any, *CSSOutput) }); ok {
		valueWithGenCSS.GenCSS(context, output)
	}
}

func (n *Negative) Eval(context any) any {
	mathOn := false
	if evalCtx, ok := context.(*Eval); ok {
		mathOn = evalCtx.IsMathOn()
	} else if ctx, ok := context.(map[string]any); ok {
		if mathOnFunc, ok := ctx["isMathOn"].(func() bool); ok {
			mathOn = mathOnFunc()
		} else if mathOnFunc, ok := ctx["isMathOn"].(func(string) bool); ok {
			mathOn = mathOnFunc("*")
		}
	}

	if mathOn {
		dim, _ := NewDimension(-1, nil)
		op := NewOperation("*", []any{dim, n.Value}, false)
		result, err := op.Eval(context)
		if err != nil {
			return n.evalValue(context)
		}
		return result
	}
	return n.evalValue(context)
}

func (n *Negative) evalValue(context any) any {
	if n.Value == nil {
		if zeroDim, err := NewDimension(0, nil); err == nil {
			return NewNegative(zeroDim)
		}
		return NewNegative(nil)
	}
	if eval, ok := n.Value.(interface{ Eval(any) any }); ok {
		evaluated := eval.Eval(context)
		return NewNegative(evaluated)
	} else if eval, ok := n.Value.(interface{ Eval(any) (any, error) }); ok {
		evaluated, err := eval.Eval(context)
		if err != nil {
			return NewNegative(n.Value)
		}
		return NewNegative(evaluated)
	}
	return NewNegative(n.Value)
} 