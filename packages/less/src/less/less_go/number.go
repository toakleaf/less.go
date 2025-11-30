package less_go

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// NumberFunctions provides all the number-related functions
var NumberFunctions = map[string]interface{}{
	"min":        Min,
	"max":        Max,
	"convert":    Convert,
	"pi":         Pi,
	"mod":        Mod,
	"pow":        Pow,
	"percentage": Percentage,
}

// NumberFunctionWrapper wraps number functions to implement FunctionDefinition interface
type NumberFunctionWrapper struct {
	name string
	fn   func(args ...interface{}) (interface{}, error)
}

func (w *NumberFunctionWrapper) Call(args ...any) (any, error) {
	return w.fn(args...)
}

func (w *NumberFunctionWrapper) CallCtx(ctx *Context, args ...any) (any, error) {
	return w.Call(args...)
}

func (w *NumberFunctionWrapper) NeedsEvalArgs() bool {
	return true
}

func wrapMinMax(fn func(args ...interface{}) (interface{}, error)) func(args ...interface{}) (interface{}, error) {
	return fn
}

func wrapConvert(fn func(val *Dimension, unit *Dimension) (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("convert expects 2 arguments, got %d", len(args))
		}
		val, ok1 := args[0].(*Dimension)
		if !ok1 {
			return nil, fmt.Errorf("convert expects first argument to be a dimension")
		}

		var unitArg *Dimension
		if dim, ok := args[1].(*Dimension); ok {
			unitArg = dim
		} else if kw, ok := args[1].(*Keyword); ok {
			unitArg = &Dimension{Value: 1.0, Unit: NewUnit([]string{kw.value}, nil, kw.value)}
		} else if quoted, ok := args[1].(*Quoted); ok {
			unitArg = &Dimension{Value: 1.0, Unit: NewUnit([]string{quoted.value}, nil, quoted.value)}
		} else {
			return nil, fmt.Errorf("convert expects second argument to be a dimension, keyword, or string")
		}

		return fn(val, unitArg)
	}
}

func wrapPi(fn func() (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		return fn()
	}
}

func wrapMod(fn func(a *Dimension, b *Dimension) (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("mod expects 2 arguments, got %d", len(args))
		}
		a, ok1 := args[0].(*Dimension)
		b, ok2 := args[1].(*Dimension)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("mod expects dimension arguments")
		}
		return fn(a, b)
	}
}

func wrapPow(fn func(x interface{}, y interface{}) (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("pow expects 2 arguments, got %d", len(args))
		}
		return fn(args[0], args[1])
	}
}

func wrapPercentage(fn func(n *Dimension) (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("percentage expects 1 argument, got %d", len(args))
		}

		dim, ok := args[0].(*Dimension)
		if !ok {
			return nil, &LessError{
				Type:    "Argument",
				Message: "argument must be a number",
			}
		}
		return fn(dim)
	}
}

func GetWrappedNumberFunctions() map[string]interface{} {
	wrappedFunctions := make(map[string]interface{})
	wrappedFunctions["min"] = &NumberFunctionWrapper{name: "min", fn: wrapMinMax(Min)}
	wrappedFunctions["max"] = &NumberFunctionWrapper{name: "max", fn: wrapMinMax(Max)}
	wrappedFunctions["convert"] = &NumberFunctionWrapper{name: "convert", fn: wrapConvert(Convert)}
	wrappedFunctions["pi"] = &NumberFunctionWrapper{name: "pi", fn: wrapPi(Pi)}
	wrappedFunctions["mod"] = &NumberFunctionWrapper{name: "mod", fn: wrapMod(Mod)}
	wrappedFunctions["pow"] = &NumberFunctionWrapper{name: "pow", fn: wrapPow(Pow)}
	wrappedFunctions["percentage"] = &NumberFunctionWrapper{name: "percentage", fn: wrapPercentage(Percentage)}
	
	return wrappedFunctions
}

func minMax(isMin bool, args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, &LessError{Type: "Argument", Message: "one or more arguments required"}
	}

	var order []interface{}
	values := make(map[string]int)
	var unitStatic string
	var unitClone string

	for i := 0; i < len(args); i++ {
		current := args[i]
		dim, ok := current.(*Dimension)
		if !ok {
			if valuer, ok := current.(interface{ GetValue() interface{} }); ok {
				if arr, ok := valuer.GetValue().([]interface{}); ok {
					args = append(args, arr...)
					continue
				}
			}
			currentVal := reflect.ValueOf(current)
			if currentVal.Kind() == reflect.Ptr {
				currentVal = currentVal.Elem()
			}
			if currentVal.Kind() == reflect.Struct {
				valueField := currentVal.FieldByName("Value")
				if valueField.IsValid() && valueField.CanInterface() {
					if arr, ok := valueField.Interface().([]interface{}); ok {
						args = append(args, arr...)
						continue
					}
				}
			}
			return nil, &LessError{Type: "Argument", Message: "incompatible types"}
		}

		var currentUnified *Dimension
		if dim.Unit.ToString() == "" && unitClone != "" {
			// Create new dimension with unitClone
			clonedUnit := &Unit{Numerator: []string{unitClone}, BackupUnit: unitClone}
			tempDim, _ := NewDimension(dim.Value, clonedUnit)
			currentUnified = tempDim.Unify()
		} else {
			currentUnified = dim.Unify()
		}
		unit := currentUnified.Unit.ToString()
		if unit == "" && unitStatic != "" {
			unit = unitStatic
		}
		if unit != "" && unitStatic == "" {
			unitStatic = unit
		} else if unit != "" && len(order) > 0 {
			firstUnified := order[0].(*Dimension).Unify()
			if firstUnified.Unit.ToString() == "" {
				unitStatic = unit
			}
		}
		if unit != "" && unitClone == "" {
			unitClone = dim.Unit.ToString()
		}
		var j int
		var found bool
		if unitVal, exists := values[""]; exists && unit != "" && unit == unitStatic {
			j = unitVal
			found = true
		} else if unitVal, exists := values[unit]; exists {
			j = unitVal
			found = true
		}

		if !found {
			if unitStatic != "" && unit != unitStatic {
				return nil, &LessError{Type: "Argument", Message: "incompatible types"}
			}
			values[unit] = len(order)
			order = append(order, dim)
			continue
		}
		var referenceUnified *Dimension
		refDim := order[j].(*Dimension)
		if refDim.Unit.ToString() == "" && unitClone != "" {
			clonedUnit := &Unit{Numerator: []string{unitClone}, BackupUnit: unitClone}
			tempDim, _ := NewDimension(refDim.Value, clonedUnit)
			referenceUnified = tempDim.Unify()
		} else {
			referenceUnified = refDim.Unify()
		}
		if (isMin && currentUnified.Value < referenceUnified.Value) ||
			(!isMin && currentUnified.Value > referenceUnified.Value) {
			order[j] = dim
		}
	}

	if len(order) == 1 {
		return order[0], nil
	}
	var cssArgs []string
	for _, arg := range order {
		if dim, ok := arg.(*Dimension); ok {
			cssArgs = append(cssArgs, dim.ToCSS(nil))
		} else if node, ok := arg.(interface{ ToCSS(any) string }); ok {
			cssArgs = append(cssArgs, node.ToCSS(nil))
		} else {
			cssArgs = append(cssArgs, fmt.Sprintf("%v", arg))
		}
	}
	
	separator := ", "
	if ctx, ok := args[0].(interface{ GetContext() interface{} }); ok {
		if context, ok := ctx.GetContext().(map[string]bool); ok && context["compress"] {
			separator = ","
		}
	}
	
	var fnName string
	if isMin {
		fnName = "min"
	} else {
		fnName = "max"
	}
	
	result := fmt.Sprintf("%s(%s)", fnName, joinStrings(cssArgs, separator))
	return NewAnonymous(result, 0, nil, false, false, nil), nil
}

func Min(args ...interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	
	result, err := minMax(true, args)
	if err != nil {
		return nil, nil
	}
	return result, nil
}

func Max(args ...interface{}) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	
	result, err := minMax(false, args)
	if err != nil {
		return nil, nil
	}
	return result, nil
}

func Convert(val *Dimension, unit *Dimension) (*Dimension, error) {
	unitStr := unit.Unit.ToString()
	result := val.ConvertTo(unitStr)
	return result, nil
}

func Pi() (*Dimension, error) {
	return NewDimension(math.Pi, nil)
}

func Mod(a *Dimension, b *Dimension) (*Dimension, error) {
	if b.Value == 0 {
		return nil, &LessError{Type: "Argument", Message: "cannot divide by zero"}
	}
	result := math.Mod(a.Value, b.Value)
	return NewDimension(result, a.Unit)
}

func Pow(x interface{}, y interface{}) (*Dimension, error) {
	var xDim, yDim *Dimension
	var err error
	xIsNum := false
	yIsNum := false
	
	if _, ok := x.(float64); ok {
		xIsNum = true
	} else if _, ok := x.(int); ok {
		xIsNum = true
	}
	if _, ok := y.(float64); ok {
		yIsNum = true
	} else if _, ok := y.(int); ok {
		yIsNum = true
	}
	if xIsNum && yIsNum {
		if xNum, ok := x.(float64); ok {
			xDim, err = NewDimension(xNum, nil)
			if err != nil {
				return nil, err
			}
		} else if xNum, ok := x.(int); ok {
			xDim, err = NewDimension(float64(xNum), nil)
			if err != nil {
				return nil, err
			}
		}
		
		if yNum, ok := y.(float64); ok {
			yDim, err = NewDimension(yNum, nil)
			if err != nil {
				return nil, err
			}
		} else if yNum, ok := y.(int); ok {
			yDim, err = NewDimension(float64(yNum), nil)
			if err != nil {
				return nil, err
			}
		}
	} else if !xIsNum && !yIsNum {
		// Both must be dimensions
		var ok bool
		xDim, ok = x.(*Dimension)
		if !ok {
			return nil, &LessError{Type: "Argument", Message: "arguments must be numbers"}
		}
		yDim, ok = y.(*Dimension)
		if !ok {
			return nil, &LessError{Type: "Argument", Message: "arguments must be numbers"}
		}
	} else {
		return nil, &LessError{Type: "Argument", Message: "arguments must be numbers"}
	}

	result := math.Pow(xDim.Value, yDim.Value)
	return NewDimension(result, xDim.Unit)
}

func Percentage(n *Dimension) (*Dimension, error) {
	percentUnit := &Unit{
		Numerator:  []string{"%"},
		BackupUnit: "%",
	}
	
	return MathHelper(func(num float64) float64 {
		return num * 100
	}, percentUnit, n)
}

func joinStrings(strs []string, separator string) string {
	var builder strings.Builder
	for i, str := range strs {
		if i > 0 {
			builder.WriteString(separator)
		}
		builder.WriteString(str)
	}
	return builder.String()
}

func init() {
	for name, fn := range GetWrappedNumberFunctions() {
		DefaultRegistry.Add(name, fn)
	}
}