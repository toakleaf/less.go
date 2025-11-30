package less_go

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Operation represents an operation node in the Less AST
type Operation struct {
	*Node
	Op       string
	Operands []any
	IsSpaced bool
}

// NewOperation creates a new Operation instance
func NewOperation(op string, operands []any, isSpaced bool) *Operation {
	operation := &Operation{
		Node:     NewNode(),
		Op:       strings.TrimSpace(op),
		Operands: operands,
		IsSpaced: isSpaced,
	}
	
	return operation
}

// Type returns the type of the node
func (o *Operation) Type() string {
	return "Operation"
}

// GetType returns the type of the node for visitor pattern consistency
func (o *Operation) GetType() string {
	return "Operation"
}

func (o *Operation) Accept(visitor any) {
	if visitorWithArray, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
		o.Operands = visitorWithArray.VisitArray(o.Operands)
		return
	}
	visitorType := reflect.ValueOf(visitor)
	if visitorType.Kind() == reflect.Struct {
		visitArrayField := visitorType.FieldByName("VisitArray")
		if visitArrayField.IsValid() && visitArrayField.CanInterface() {
			if visitArrayFunc, ok := visitArrayField.Interface().(func([]any) []any); ok {
				o.Operands = visitArrayFunc(o.Operands)
				return
			}
		}
	}
}

func (o *Operation) Eval(context any) (any, error) {
	var a, b any
	if len(o.Operands) < 2 || o.Operands[0] == nil || o.Operands[1] == nil {
		return nil, &LessError{
			Type:    "Operation",
			Message: "Operation requires two operands",
		}
	}
	if aNode, ok := o.Operands[0].(interface{ Eval(any) (any, error) }); ok {
		var err error
		a, err = aNode.Eval(context)
		if err != nil {
			return nil, err
		}
	} else if aNode, ok := o.Operands[0].(interface{ Eval(any) any }); ok {
		a = aNode.Eval(context)
	} else {
		a = o.Operands[0]
	}
	if bNode, ok := o.Operands[1].(interface{ Eval(any) (any, error) }); ok {
		var err error
		b, err = bNode.Eval(context)
		if err != nil {
			return nil, err
		}
	} else if bNode, ok := o.Operands[1].(interface{ Eval(any) any }); ok {
		b = bNode.Eval(context)
	} else {
		b = o.Operands[1]
	}
	mathOn := false
	debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
	if evalCtx, ok := context.(*Eval); ok {
		mathOn = evalCtx.IsMathOnWithOp(o.Op)
		if debugTrace {
			fmt.Printf("[TRACE] Operation.Eval: op=%s, mathOn=%v, ParensStack len=%d\n", o.Op, mathOn, len(evalCtx.ParensStack))
		}
	} else if ctx, ok := context.(map[string]any); ok {
		if mathOnFunc, ok := ctx["isMathOn"].(func(string) bool); ok {
			mathOn = mathOnFunc(o.Op)
		}
	}

	if mathOn {
		op := o.Op
		if op == "./" {
			op = "/"
		}
		if aDim, aOk := a.(*Dimension); aOk {
			if _, bOk := b.(*Color); bOk {
				a = aDim.ToColor()
			}
		}
		if bDim, bOk := b.(*Dimension); bOk {
			if _, aOk := a.(*Color); aOk {
				b = bDim.ToColor()
			}
		}
		aHasOperate := false
		bHasOperate := false
		
		if _, ok := a.(*Dimension); ok {
			aHasOperate = true
		} else if _, ok := a.(*Color); ok {
			aHasOperate = true
		} else if _, ok := a.(*Anonymous); ok {
			aHasOperate = true
		}
		if _, ok := b.(*Dimension); ok {
			bHasOperate = true
		} else if _, ok := b.(*Color); ok {
			bHasOperate = true
		} else if _, ok := b.(*Anonymous); ok {
			bHasOperate = true
		}
		
		if !aHasOperate || !bHasOperate {
			aOp, aIsOp := a.(*Operation)
			_, bIsOp := b.(*Operation)
			
			if (aIsOp || bIsOp) && aIsOp && aOp.Op == "/" && IsMathParensDivision(context) {
				return NewOperation(o.Op, []any{a, b}, o.IsSpaced), nil
			}
			
			return nil, &LessError{
				Type:    "Operation",
				Message: "Operation on an invalid type",
			}
		}
		if aDim, aOk := a.(*Dimension); aOk {
			if bDim, bOk := b.(*Dimension); bOk {
				result := aDim.Operate(context, op, bDim)
				if result == nil {
					return nil, &LessError{
						Type:    "Operation",
						Message: "Operation produced invalid result (NaN)",
					}
				}
				return result, nil
			}
		}
		if aColor, aOk := a.(*Color); aOk {
			if bColor, bOk := b.(*Color); bOk {
				result := aColor.OperateColor(context, op, bColor)
				return result, nil
			}
		}
		if aOperable, aOk := a.(interface{ Operate(any, string, any) any }); aOk {
			return aOperable.Operate(context, op, b), nil
		}
		return nil, &LessError{
			Type:    "Operation",
			Message: "Operation on an invalid type",
		}
	}
	return NewOperation(o.Op, []any{a, b}, o.IsSpaced), nil
}

func (o *Operation) GenCSS(context any, output *CSSOutput) {
	if operand0, ok := o.Operands[0].(interface{ GenCSS(any, *CSSOutput) }); ok {
		operand0.GenCSS(context, output)
	}

	if o.IsSpaced {
		output.Add(" ", nil, nil)
	}
	output.Add(o.Op, nil, nil)
	if o.IsSpaced {
		output.Add(" ", nil, nil)
	}

	if operand1, ok := o.Operands[1].(interface{ GenCSS(any, *CSSOutput) }); ok {
		operand1.GenCSS(context, output)
	}
}

func IsMathParensDivision(context any) bool {
	if evalCtx, ok := context.(*Eval); ok {
		return evalCtx.Math == MathParensDivision
	}
	if ctx, ok := context.(map[string]any); ok {
		if mathType, ok := ctx["math"].(MathType); ok {
			return mathType == Math.ParensDivision
		}
	}
	return false
} 