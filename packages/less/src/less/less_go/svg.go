package less_go

import (
	"fmt"
	"net/url"
	"strings"
)

// SvgFunctions provides all the svg-related functions
var SvgFunctions = map[string]interface{}{
	"svg-gradient": SvgGradient,
}

// RegisterSvgFunctions registers svg functions with the given registry
func RegisterSvgFunctions(registry *Registry) {
	for name, fn := range GetWrappedSvgFunctions() {
		if regFn, ok := fn.(FunctionDefinition); ok {
			registry.Add(name, regFn)
		}
	}
}

// init registers svg functions with the default registry
func init() {
	RegisterSvgFunctions(DefaultRegistry)
}

// SvgFunctionWrapper wraps svg functions to implement FunctionDefinition interface
type SvgFunctionWrapper struct {
	name string
}

func (w *SvgFunctionWrapper) Call(args ...any) (any, error) {
	return nil, fmt.Errorf("svg-gradient function requires context - use CallCtx instead")
}

func (w *SvgFunctionWrapper) CallCtx(ctx *Context, args ...any) (any, error) {
	switch w.name {
	case "svg-gradient":
		// Extract the proper evaluation context from ctx
		var evalCtx any = ctx
		if ctx != nil && len(ctx.Frames) > 0 && ctx.Frames[0] != nil {
			if ec, ok := ctx.Frames[0].EvalContext.(EvalContext); ok {
				evalCtx = ec
			}
		}

		// Evaluate arguments first - deeply evaluate Expression contents
		evaluatedArgs := make([]any, len(args))
		for i, arg := range args {
			evaluated := arg

			// First, evaluate the argument itself if it has an Eval method
			// Try the new (any, error) signature first
			if evaluator, ok := arg.(interface{ Eval(any) (any, error) }); ok {
				evalResult, err := evaluator.Eval(evalCtx)
				if err != nil {
					// Panic with LessError to fail compilation
					filename := ""
					if ctx != nil && len(ctx.Frames) > 0 && ctx.Frames[0] != nil {
						if ctx.Frames[0].CurrentFileInfo != nil {
							if fn, ok := ctx.Frames[0].CurrentFileInfo["filename"].(string); ok {
								filename = fn
							}
						}
					}
					panic(NewLessError(ErrorDetails{
						Type:    "ArgumentError",
						Message: fmt.Sprintf("error evaluating argument %d: %v", i, err),
					}, nil, filename))
				}
				evaluated = evalResult
			} else if evaluator, ok := arg.(interface{ Eval(any) any }); ok {
				// Try the old (any) signature for backward compatibility
				evaluated = evaluator.Eval(evalCtx)
			}

			// If result is an Expression, evaluate its contents deeply
			if expr, ok := evaluated.(*Expression); ok {
				evaluatedValues := make([]any, len(expr.Value))
				for j, val := range expr.Value {
					// Try the new (any, error) signature first
					if evaluator, ok := val.(interface{ Eval(any) (any, error) }); ok {
						evalResult, err := evaluator.Eval(evalCtx)
						if err != nil {
							filename := ""
							if ctx != nil && len(ctx.Frames) > 0 && ctx.Frames[0] != nil {
								if ctx.Frames[0].CurrentFileInfo != nil {
									if fn, ok := ctx.Frames[0].CurrentFileInfo["filename"].(string); ok {
										filename = fn
									}
								}
							}
							panic(NewLessError(ErrorDetails{
								Type:    "ArgumentError",
								Message: fmt.Sprintf("error evaluating expression value %d in argument %d: %v", j, i, err),
							}, nil, filename))
						}
						evaluatedValues[j] = evalResult
					} else if evaluator, ok := val.(interface{ Eval(any) any }); ok {
						// Try the old (any) signature for backward compatibility
						evalResult := evaluator.Eval(evalCtx)
						evaluatedValues[j] = evalResult
					} else {
						evaluatedValues[j] = val
					}
				}
				// Create new Expression with evaluated values
				newExpr, _ := NewExpression(evaluatedValues, expr.NoSpacing)
				evaluated = newExpr
			}

			evaluatedArgs[i] = evaluated
		}

		svgCtx := buildSvgContext(ctx)
		result, err := SvgGradient(svgCtx, evaluatedArgs...)
		if err != nil {
			// Panic with LessError to fail compilation
			filename := ""
			if ctx != nil && len(ctx.Frames) > 0 && ctx.Frames[0] != nil {
				if ctx.Frames[0].CurrentFileInfo != nil {
					if fn, ok := ctx.Frames[0].CurrentFileInfo["filename"].(string); ok {
						filename = fn
					}
				}
			}
			panic(NewLessError(ErrorDetails{
				Type:    "ArgumentError",
				Message: err.Error(),
			}, nil, filename))
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown svg function: %s", w.name)
	}
}

func (w *SvgFunctionWrapper) NeedsEvalArgs() bool {
	// SVG functions need context for error handling
	return false
}

// buildSvgContext builds an SvgContext from the Context
func buildSvgContext(ctx *Context) SvgContext {
	svgCtx := SvgContext{
		Index:           0,
		CurrentFileInfo: make(map[string]any),
	}

	if ctx != nil && len(ctx.Frames) > 0 && ctx.Frames[0] != nil {
		frame := ctx.Frames[0]
		if frame.CurrentFileInfo != nil {
			svgCtx.CurrentFileInfo = frame.CurrentFileInfo
		}
		if evalCtx, ok := frame.EvalContext.(EvalContext); ok {
			svgCtx.Context = evalCtx
		}
	}

	return svgCtx
}

// GetWrappedSvgFunctions returns svg functions wrapped with FunctionDefinition interface
func GetWrappedSvgFunctions() map[string]interface{} {
	wrappedFunctions := make(map[string]interface{})
	for name := range SvgFunctions {
		wrappedFunctions[name] = &SvgFunctionWrapper{name: name}
	}
	return wrappedFunctions
}

// SvgContext represents the context needed for svg function execution
type SvgContext struct {
	Index           int
	CurrentFileInfo map[string]any
	Context         EvalContext
}

// SvgGradient implements the svg-gradient() function which creates SVG gradient data URIs
func SvgGradient(ctx SvgContext, args ...interface{}) (*URL, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
	}

	// Get direction value - can be any node type that can be converted to string
	var directionValue string

	// Try different node types
	switch dir := args[0].(type) {
	case *Quoted:
		directionValue = dir.value
	case *Keyword:
		directionValue = dir.value
	case *Anonymous:
		if str, ok := dir.Value.(string); ok {
			directionValue = str
		} else {
			directionValue = fmt.Sprintf("%v", dir.Value)
		}
	case *Expression:
		// Handle Expression (e.g., "to bottom" as two keywords in an expression)
		directionValue = dir.ToCSS(ctx.Context)
	default:
		// Try ToCSS as last resort
		if node, ok := args[0].(interface{ ToCSS(bool, *Eval) (string, error) }); ok {
			css, err := node.ToCSS(false, nil)
			if err == nil {
				directionValue = css
			} else {
				return nil, fmt.Errorf("svg-gradient first argument must be a direction (got %T)", args[0])
			}
		} else {
			return nil, fmt.Errorf("svg-gradient first argument must be a direction (got %T)", args[0])
		}
	}

	var stops []interface{}

	if len(args) == 2 {
		// Check if second argument is an Expression (color list)
		if expr, ok := args[1].(*Expression); ok {
			if len(expr.Value) < 2 {
				return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
			}
			stops = expr.Value
		} else {
			return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
		}
	} else if len(args) < 3 {
		return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
	} else {
		stops = args[1:]
	}

	var gradientDirectionSvg string
	var gradientType string = "linear"
	var rectangleDimension string = `x="0" y="0" width="1" height="1"`

	switch directionValue {
	case "to bottom":
		gradientDirectionSvg = `x1="0%" y1="0%" x2="0%" y2="100%"`
	case "to right":
		gradientDirectionSvg = `x1="0%" y1="0%" x2="100%" y2="0%"`
	case "to bottom right":
		gradientDirectionSvg = `x1="0%" y1="0%" x2="100%" y2="100%"`
	case "to top right":
		gradientDirectionSvg = `x1="0%" y1="100%" x2="100%" y2="0%"`
	case "ellipse", "ellipse at center":
		gradientType = "radial"
		gradientDirectionSvg = `cx="50%" cy="50%" r="75%"`
		rectangleDimension = `x="-50" y="-50" width="101" height="101"`
	default:
		return nil, fmt.Errorf("svg-gradient direction must be 'to bottom', 'to right', 'to bottom right', 'to top right' or 'ellipse at center'")
	}

	returner := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><%sGradient id="g" %s>`, gradientType, gradientDirectionSvg)

	for i, stop := range stops {
		var color interface{}
		var position interface{}

		if expr, ok := stop.(*Expression); ok {
			if len(expr.Value) > 0 {
				color = expr.Value[0]
			}
			if len(expr.Value) > 1 {
				position = expr.Value[1]
			}
		} else {
			color = stop
			position = nil
		}

		// Validate color
		colorNode, isColor := color.(*Color)
		if !isColor {
			return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
		}

		// Validate position requirements
		isFirst := i == 0
		isLast := i+1 == len(stops)
		isMiddle := !isFirst && !isLast

		if isFirst || isLast {
			// First and last colors can have position undefined
			if position != nil {
				if _, isDimension := position.(*Dimension); !isDimension {
					return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
				}
			}
		} else if isMiddle {
			// Middle colors must have position
			if position == nil {
				return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
			}
			if _, isDimension := position.(*Dimension); !isDimension {
				return nil, fmt.Errorf("svg-gradient expects direction, start_color [start_position], [color position,]..., end_color [end_position] or direction, color list")
			}
		}

		var positionValue string
		if position != nil {
			if dim, ok := position.(*Dimension); ok {
				// Access fields directly since ToCSS doesn't work properly
				unit := ""
				if dim.Unit != nil && len(dim.Unit.Numerator) > 0 {
					unit = dim.Unit.Numerator[0]
				} else if dim.Unit != nil && dim.Unit.BackupUnit != "" {
					unit = dim.Unit.BackupUnit
				}
				positionValue = fmt.Sprintf("%.0f%s", dim.Value, unit)
			}
		} else {
			if isFirst {
				positionValue = "0%"
			} else {
				positionValue = "100%"
			}
		}

		alpha := colorNode.Alpha
		toRGB := colorNode.ToRGB()

		stopOpacity := ""
		if alpha < 1 {
			stopOpacity = fmt.Sprintf(` stop-opacity="%g"`, alpha)
		}

		returner += fmt.Sprintf(`<stop offset="%s" stop-color="%s"%s/>`, positionValue, toRGB, stopOpacity)
	}

	returner += fmt.Sprintf(`</%sGradient><rect %s fill="url(#g)" /></svg>`, gradientType, rectangleDimension)

	// URL encode the SVG using JavaScript encodeURIComponent behavior
	encodedSVG := url.QueryEscape(returner)
	// Convert to match JavaScript encodeURIComponent: spaces should be %20 not +
	encodedSVG = strings.ReplaceAll(encodedSVG, "+", "%20")

	// Create data URI
	dataURI := fmt.Sprintf("data:image/svg+xml,%s", encodedSVG)

	// Return as URL node
	quotedURI := NewQuoted("'", dataURI, false, ctx.Index, ctx.CurrentFileInfo)
	return NewURL(quotedURI, ctx.Index, ctx.CurrentFileInfo, false), nil
}

// SvgGradientWithCatch implements svg-gradient with error handling like the JavaScript version
func SvgGradientWithCatch(ctx SvgContext, args ...interface{}) *URL {
	result, err := SvgGradient(ctx, args...)
	if err != nil {
		return nil // Return nil on error, like JavaScript version returns undefined
	}
	return result
}