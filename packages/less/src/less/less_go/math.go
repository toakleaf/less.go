package less_go

import (
	"fmt"
	"math"
)

// MathFunctions provides all the mathematical functions that were in math.js
var MathFunctions = map[string]any{
	"ceil":  Ceil,
	"floor": Floor,
	"sqrt":  Sqrt,
	"abs":   Abs,
	"tan":   Tan,
	"sin":   Sin,
	"cos":   Cos,
	"atan":  Atan,
	"asin":  Asin,
	"acos":  Acos,
	"round": Round,
}

// MathFunctionWrapper wraps math functions to implement FunctionDefinition interface
type MathFunctionWrapper struct {
	name string
	fn   func(args ...interface{}) (interface{}, error)
}

func (w *MathFunctionWrapper) Call(args ...any) (any, error) {
	return w.fn(args...)
}

func (w *MathFunctionWrapper) CallCtx(ctx *Context, args ...any) (any, error) {
	return w.Call(args...)
}

func (w *MathFunctionWrapper) NeedsEvalArgs() bool {
	return true
}

func wrapUnaryMath(fn func(*Dimension) (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("function expects 1 argument, got %d", len(args))
		}
		dim, ok := args[0].(*Dimension)
		if !ok {
			return nil, fmt.Errorf("function expects dimension argument")
		}
		return fn(dim)
	}
}

func wrapRound(fn func(*Dimension, *Dimension) (*Dimension, error)) func(args ...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("round expects 1 or 2 arguments, got %d", len(args))
		}
		n, ok := args[0].(*Dimension)
		if !ok {
			return nil, fmt.Errorf("round expects dimension as first argument")
		}
		var f *Dimension
		if len(args) == 2 {
			f, ok = args[1].(*Dimension)
			if !ok {
				return nil, fmt.Errorf("round expects dimension as second argument")
			}
		}
		return fn(n, f)
	}
}

func GetWrappedMathFunctions() map[string]interface{} {
	wrappedFunctions := make(map[string]interface{})
	wrappedFunctions["ceil"] = &MathFunctionWrapper{name: "ceil", fn: wrapUnaryMath(Ceil)}
	wrappedFunctions["floor"] = &MathFunctionWrapper{name: "floor", fn: wrapUnaryMath(Floor)}
	wrappedFunctions["sqrt"] = &MathFunctionWrapper{name: "sqrt", fn: wrapUnaryMath(Sqrt)}
	wrappedFunctions["abs"] = &MathFunctionWrapper{name: "abs", fn: wrapUnaryMath(Abs)}
	wrappedFunctions["tan"] = &MathFunctionWrapper{name: "tan", fn: wrapUnaryMath(Tan)}
	wrappedFunctions["sin"] = &MathFunctionWrapper{name: "sin", fn: wrapUnaryMath(Sin)}
	wrappedFunctions["cos"] = &MathFunctionWrapper{name: "cos", fn: wrapUnaryMath(Cos)}
	wrappedFunctions["atan"] = &MathFunctionWrapper{name: "atan", fn: wrapUnaryMath(Atan)}
	wrappedFunctions["asin"] = &MathFunctionWrapper{name: "asin", fn: wrapUnaryMath(Asin)}
	wrappedFunctions["acos"] = &MathFunctionWrapper{name: "acos", fn: wrapUnaryMath(Acos)}
	wrappedFunctions["round"] = &MathFunctionWrapper{name: "round", fn: wrapRound(Round)}
	
	return wrappedFunctions
}

func Ceil(n *Dimension) (*Dimension, error) {
	return MathHelper(math.Ceil, nil, n)
}

func Floor(n *Dimension) (*Dimension, error) {
	return MathHelper(math.Floor, nil, n)
}

func Sqrt(n *Dimension) (*Dimension, error) {
	return MathHelper(math.Sqrt, nil, n)
}

func Abs(n *Dimension) (*Dimension, error) {
	return MathHelper(math.Abs, nil, n)
}

func Tan(n *Dimension) (*Dimension, error) {
	emptyUnit := NewUnit(nil, nil, "")
	return MathHelper(math.Tan, emptyUnit, n)
}

func Sin(n *Dimension) (*Dimension, error) {
	emptyUnit := NewUnit(nil, nil, "")
	return MathHelper(math.Sin, emptyUnit, n)
}

func Cos(n *Dimension) (*Dimension, error) {
	emptyUnit := NewUnit(nil, nil, "")
	return MathHelper(math.Cos, emptyUnit, n)
}

func Atan(n *Dimension) (*Dimension, error) {
	radUnit := NewUnit([]string{"rad"}, nil, "rad")
	return MathHelper(math.Atan, radUnit, n)
}

func Asin(n *Dimension) (*Dimension, error) {
	radUnit := NewUnit([]string{"rad"}, nil, "rad")
	return MathHelper(math.Asin, radUnit, n)
}

func Acos(n *Dimension) (*Dimension, error) {
	radUnit := NewUnit([]string{"rad"}, nil, "rad")
	return MathHelper(math.Acos, radUnit, n)
}

func Round(n *Dimension, f *Dimension) (*Dimension, error) {
	var fraction float64
	if f == nil {
		fraction = 0
	} else {
		fraction = f.Value
	}

	roundFunc := func(num float64) float64 {
		// JavaScript's toFixed rounds half values away from zero
		multiplier := math.Pow(10, fraction)
		var rounded float64
		if num >= 0 {
			rounded = math.Floor(num*multiplier + 0.5) / multiplier
		} else {
			rounded = math.Ceil(num*multiplier - 0.5) / multiplier
		}
		
		return rounded
	}

	return MathHelper(roundFunc, nil, n)
}

func init() {
	for name, fn := range GetWrappedMathFunctions() {
		DefaultRegistry.Add(name, fn)
	}
}