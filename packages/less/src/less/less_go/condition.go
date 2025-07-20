package less_go

import (
	"math"
	"strings"
)

// Condition represents a condition node in the Less AST
type Condition struct {
	*Node
	Op     string
	Lvalue any
	Rvalue any
	Index  int
	Negate bool
}

// NewCondition creates a new Condition
func NewCondition(op string, l, r any, i int, negate bool) *Condition {
	return &Condition{
		Node:   NewNode(),
		Op:     strings.TrimSpace(op),
		Lvalue: l,
		Rvalue: r,
		Index:  i,
		Negate: negate,
	}
}

// GetType returns the node type
func (c *Condition) GetType() string {
	return "Condition"
}

// Accept implements the Node Accept method
func (c *Condition) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		c.Lvalue = v.Visit(c.Lvalue)
		c.Rvalue = v.Visit(c.Rvalue)
	}
}

// Eval evaluates the condition
func (c *Condition) Eval(context any) bool {
	// JavaScript implementation:
	// Evaluates lvalue and rvalue, then uses Node.compare for comparison operators
	
	// Helper to evaluate a node
	eval := func(node any) any {
		if evaluator, ok := node.(interface{ Eval(any) (any, error) }); ok {
			result, _ := evaluator.Eval(context)
			return result
		} else if evaluator, ok := node.(interface{ Eval(any) any }); ok {
			return evaluator.Eval(context)
		}
		return node
	}

	var result bool

	switch c.Op {
	case "and":
		a := eval(c.Lvalue)
		b := eval(c.Rvalue)
		// Convert to bool - JavaScript truthy/falsy conversion
		abool := toBool(a)
		bbool := toBool(b)
		result = abool && bbool
		
	case "or":
		a := eval(c.Lvalue)
		b := eval(c.Rvalue)
		// Convert to bool - JavaScript truthy/falsy conversion
		abool := toBool(a)
		bbool := toBool(b)
		result = abool || bbool
		
	default:
		// For comparison operators, use Node.compare
		a := eval(c.Lvalue)
		b := eval(c.Rvalue)
		
		// Convert to nodes if necessary for comparison
		var aNode, bNode *Node
		if n, ok := a.(*Node); ok {
			aNode = n
		} else {
			// Wrap non-node values in a Node for comparison
			aNode = &Node{Value: a}
		}
		if n, ok := b.(*Node); ok {
			bNode = n
		} else {
			// Wrap non-node values in a Node for comparison
			bNode = &Node{Value: b}
		}
		
		compareResult := Compare(aNode, bNode)
		
		switch compareResult {
		case -1:
			result = c.Op == "<" || c.Op == "=<" || c.Op == "<="
		case 0:
			result = c.Op == "=" || c.Op == ">=" || c.Op == "=<" || c.Op == "<="
		case 1:
			result = c.Op == ">" || c.Op == ">="
		default:
			// JavaScript returns false for undefined comparison results
			result = false
		}
	}

	if c.Negate {
		return !result
	}
	return result
}

// toBool converts a value to boolean following JavaScript truthy/falsy rules
func toBool(v any) bool {
	if v == nil {
		return false
	}
	
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int8:
		return val != 0
	case int16:
		return val != 0
	case int32:
		return val != 0
	case int64:
		return val != 0
	case uint:
		return val != 0
	case uint8:
		return val != 0
	case uint16:
		return val != 0
	case uint32:
		return val != 0
	case uint64:
		return val != 0
	case float32:
		return val != 0.0 && !math.IsNaN(float64(val))
	case float64:
		return val != 0.0 && !math.IsNaN(val)
	case string:
		return val != ""
	case []any:
		return len(val) > 0
	case map[string]any:
		return len(val) > 0
	default:
		// Check if it's a dimension with zero value
		if dim, ok := v.(*Dimension); ok {
			return dim.Value != 0.0
		}
		// Check if it's another type with GetValue method
		if dim, ok := v.(interface{ GetValue() float64 }); ok {
			return dim.GetValue() != 0.0
		}
		// Check if it's a Node-like object that could be evaluated
		if hasVal, ok := v.(interface{ Value() any }); ok {
			return toBool(hasVal.Value())
		}
		// For other types (like Node objects), they are truthy unless explicitly falsy
		return true
	}
} 