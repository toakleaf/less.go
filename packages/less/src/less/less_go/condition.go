package less_go

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type Condition struct {
	*Node
	Op     string
	Lvalue any
	Rvalue any
	Index  int
	Negate bool
}

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

func (c *Condition) GetType() string {
	return "Condition"
}

func (c *Condition) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		c.Lvalue = v.Visit(c.Lvalue)
		c.Rvalue = v.Visit(c.Rvalue)
	}
}

func (c *Condition) Eval(context any) any {
	return c.EvalBool(context)
}

func (c *Condition) EvalBool(context any) bool {
	debug := os.Getenv("LESS_DEBUG_GUARDS") == "1"
	if debug {
		fmt.Printf("DEBUG:   Condition.EvalBool: op=%s, lvalue=%T, rvalue=%T\n", c.Op, c.Lvalue, c.Rvalue)
	}

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
		abool := toBool(a)
		bbool := toBool(b)
		result = abool && bbool
		if debug {
			fmt.Printf("DEBUG:   Condition.EvalBool AND: a=%v (%t), b=%v (%t), result=%t\n", a, abool, b, bbool, result)
		}

	case "or":
		a := eval(c.Lvalue)
		b := eval(c.Rvalue)
		abool := toBool(a)
		bbool := toBool(b)
		result = abool || bbool
		if debug {
			fmt.Printf("DEBUG:   Condition.EvalBool OR: a=%v (%t), b=%v (%t), result=%t\n", a, abool, b, bbool, result)
		}

	default:
		a := eval(c.Lvalue)
		b := eval(c.Rvalue)

		var compareResult int

		// Check Quoted and Anonymous first for symmetric comparison results
		if quoted, ok := a.(*Quoted); ok {
			cmpResult := quoted.Compare(b)
			if cmpResult == nil {
				compareResult = 999 // undefined
			} else {
				compareResult = *cmpResult
			}
		} else if quoted, ok := b.(*Quoted); ok {
			cmpResult := quoted.Compare(a)
			if cmpResult == nil {
				compareResult = 999 // undefined
			} else {
				compareResult = -*cmpResult
			}
		} else if anon, ok := a.(*Anonymous); ok {
			cmpResult := anon.Compare(b)
			if cmpResult == nil {
				compareResult = 999 // undefined
			} else if cmpInt, ok := cmpResult.(int); ok {
				compareResult = cmpInt
			} else {
				compareResult = 999
			}
		} else if anon, ok := b.(*Anonymous); ok {
			cmpResult := anon.Compare(a)
			if cmpResult == nil {
				compareResult = 999 // undefined
			} else if cmpInt, ok := cmpResult.(int); ok {
				compareResult = -cmpInt
			} else {
				compareResult = 999
			}
		} else if dim, ok := a.(*Dimension); ok {
			if otherDim, ok := b.(*Dimension); ok {
				if cmpPtr := dim.Compare(otherDim); cmpPtr != nil {
					compareResult = *cmpPtr
				} else {
					compareResult = 999 // undefined
				}
			} else {
				compareResult = 999
			}
		} else if col, ok := a.(*Color); ok {
			if otherCol, ok := b.(*Color); ok {
				compareResult = col.Compare(otherCol)
			} else {
				compareResult = 999
			}
		} else if expr, ok := a.(*Expression); ok {
			if otherExpr, ok := b.(*Expression); ok {
				if len(expr.Value) != len(otherExpr.Value) {
					compareResult = 999
				} else {
					allEqual := true
					for i := range expr.Value {
						aElem := eval(expr.Value[i])
						bElem := eval(otherExpr.Value[i])

						if debug {
							fmt.Printf("DEBUG:   Expression element %d: aElem=%T(%v), bElem=%T(%v)\n", i, aElem, aElem, bElem, bElem)
						}

						elemCond := NewCondition("=", aElem, bElem, 0, false)
						if !elemCond.EvalBool(context) {
							allEqual = false
							if debug {
								fmt.Printf("DEBUG:   Expression element %d comparison failed\n", i)
							}
							break
						}
					}
					if allEqual {
						compareResult = 0
					} else {
						compareResult = 999
					}
				}
			} else {
				compareResult = 999
			}
		} else if val, ok := a.(*Value); ok {
			if otherVal, ok := b.(*Value); ok {
				if len(val.Value) != len(otherVal.Value) {
					compareResult = 999
				} else {
					allEqual := true
					for i := range val.Value {
						aElem := eval(val.Value[i])
						bElem := eval(otherVal.Value[i])

						if debug {
							fmt.Printf("DEBUG:   Value element %d: aElem=%T(%v), bElem=%T(%v)\n", i, aElem, aElem, bElem, bElem)
						}

						elemCond := NewCondition("=", aElem, bElem, 0, false)
						if !elemCond.EvalBool(context) {
							allEqual = false
							if debug {
								fmt.Printf("DEBUG:   Value element %d comparison failed\n", i)
							}
							break
						}
					}
					if allEqual {
						compareResult = 0
					} else {
						compareResult = 999
					}
				}
			} else {
				compareResult = 999
			}
		} else if kw, ok := a.(*Keyword); ok {
			if otherKw, ok := b.(*Keyword); ok {
				if kw.ToCSS(nil) == otherKw.ToCSS(nil) {
					compareResult = 0
				} else {
					compareResult = 999
				}
			} else {
				compareResult = 999
			}
		} else {
			var aNode, bNode *Node

			if nodeProvider, ok := a.(interface{ GetNode() *Node }); ok {
				aNode = nodeProvider.GetNode()
			} else if n, ok := a.(*Node); ok {
				aNode = n
			} else if dim, ok := a.(*Dimension); ok {
				aNode = dim.Node
			} else if col, ok := a.(*Color); ok {
				aNode = col.Node
			} else if kw, ok := a.(*Keyword); ok {
				aNode = kw.Node
			} else {
				aNode = &Node{Value: a}
			}

			if nodeProvider, ok := b.(interface{ GetNode() *Node }); ok {
				bNode = nodeProvider.GetNode()
			} else if n, ok := b.(*Node); ok {
				bNode = n
			} else if dim, ok := b.(*Dimension); ok {
				bNode = dim.Node
			} else if col, ok := b.(*Color); ok {
				bNode = col.Node
			} else if kw, ok := b.(*Keyword); ok {
				bNode = kw.Node
			} else {
				bNode = &Node{Value: b}
			}

			compareResult = Compare(aNode, bNode)
		}

		switch compareResult {
		case -1:
			result = c.Op == "<" || c.Op == "=<" || c.Op == "<="
		case 0:
			result = c.Op == "=" || c.Op == ">=" || c.Op == "=<" || c.Op == "<="
		case 1:
			result = c.Op == ">" || c.Op == ">="
		default:
			result = false
		}
	}

	if c.Negate {
		return !result
	}
	return result
}

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
		if dim, ok := v.(*Dimension); ok {
			return dim.Value != 0.0
		}
		if dim, ok := v.(interface{ GetValue() float64 }); ok {
			return dim.GetValue() != 0.0
		}
		if hasVal, ok := v.(interface{ Value() any }); ok {
			return toBool(hasVal.Value())
		}
		return true
	}
} 