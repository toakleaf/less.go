package less_go

import (
	"fmt"
	"os"
)

func init() {
	DefaultRegistry.Add("boolean", &FlexibleFunctionDef{
		name:      "boolean",
		minArgs:   1,
		maxArgs:   1,
		variadic:  false,
		fn:        func(args ...any) any { return Boolean(args[0]) },
		needsEval: true,
	})
	DefaultRegistry.Add("if", &IfFunctionDef{})
	DefaultRegistry.Add("isdefined", &IsDefinedFunctionDef{})
}

func Boolean(condition any) *Keyword {
	if isTruthy(condition) {
		return KeywordTrue
	}
	return KeywordFalse
}

// If takes unevaluated nodes for lazy evaluation
func If(context *Context, condition any, trueValue any, falseValue any) any {
	var evalContext any = context
	if context != nil && len(context.Frames) > 0 && context.Frames[0].EvalContext != nil {
		evalContext = context.Frames[0].EvalContext
	}

	var conditionResult any
	if evaluable, ok := condition.(interface{ Eval(any) (any, error) }); ok {
		result, _ := evaluable.Eval(evalContext)
		conditionResult = result
	} else if evaluable, ok := condition.(interface{ Eval(any) any }); ok {
		conditionResult = evaluable.Eval(evalContext)
	} else {
		conditionResult = condition
	}

	debug := os.Getenv("LESS_DEBUG_IF") == "1"
	if debug {
		fmt.Printf("[If] condition=%v (type: %T)\n", condition, condition)
		fmt.Printf("[If] conditionResult=%v (type: %T)\n", conditionResult, conditionResult)
		fmt.Printf("[If] isTruthy(conditionResult)=%v\n", isTruthy(conditionResult))
	}

	if isTruthy(conditionResult) {
		if evaluable, ok := trueValue.(interface{ Eval(any) (any, error) }); ok {
			result, _ := evaluable.Eval(evalContext)
			return result
		} else if evaluable, ok := trueValue.(interface{ Eval(any) any }); ok {
			return evaluable.Eval(evalContext)
		}
		return trueValue
	}

	if falseValue != nil {
		if evaluable, ok := falseValue.(interface{ Eval(any) (any, error) }); ok {
			result, _ := evaluable.Eval(evalContext)
			return result
		} else if evaluable, ok := falseValue.(interface{ Eval(any) any }); ok {
			result := evaluable.Eval(evalContext)
			return result
		}
		return falseValue
	}

	return NewAnonymous("", 0, nil, false, false, nil)
}

func IsDefined(context *Context, variable any) *Keyword {
	defer func() {
		recover()
	}()

	var evalContext any = context
	if context != nil && len(context.Frames) > 0 && context.Frames[0].EvalContext != nil {
		evalContext = context.Frames[0].EvalContext
	}

	if evaluable, ok := variable.(interface{ Eval(any) (any, error) }); ok {
		_, err := evaluable.Eval(evalContext)
		if err != nil {
			return KeywordFalse
		}
		return KeywordTrue
	}

	return KeywordTrue
}

func GetBooleanFunctions() map[string]any {
	return map[string]any{
		"boolean":   Boolean,
		"if":        If,
		"isdefined": IsDefined,
	}
}

func GetWrappedBooleanFunctions() map[string]interface{} {
	return map[string]interface{}{
		"boolean": &FlexibleFunctionDef{
			name:      "boolean",
			minArgs:   1,
			maxArgs:   1,
			variadic:  false,
			fn:        func(args ...any) any { return Boolean(args[0]) },
			needsEval: true,
		},
		"if":        &IfFunctionDef{},
		"isdefined": &IsDefinedFunctionDef{},
	}
}

type IfFunctionDef struct{}

func (f *IfFunctionDef) Call(args ...any) (any, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("if function expects 2-3 arguments, got %d", len(args))
	}

	ctx := &Context{}

	condition := args[0]
	trueValue := args[1]
	var falseValue any
	if len(args) > 2 {
		falseValue = args[2]
	}

	return If(ctx, condition, trueValue, falseValue), nil
}

func (f *IfFunctionDef) CallCtx(ctx *Context, args ...any) (any, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("if function expects 2-3 arguments, got %d", len(args))
	}

	condition := args[0]
	trueValue := args[1]
	var falseValue any
	if len(args) > 2 {
		falseValue = args[2]
	}

	return If(ctx, condition, trueValue, falseValue), nil
}

func (f *IfFunctionDef) NeedsEvalArgs() bool { return false }

type IsDefinedFunctionDef struct{}

func (f *IsDefinedFunctionDef) Call(args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isdefined function expects 1 argument, got %d", len(args))
	}

	return IsDefined(&Context{}, args[0]), nil
}

func (f *IsDefinedFunctionDef) CallCtx(ctx *Context, args ...any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isdefined function expects 1 argument, got %d", len(args))
	}

	return IsDefined(ctx, args[0]), nil
}

func (f *IsDefinedFunctionDef) NeedsEvalArgs() bool { return false }