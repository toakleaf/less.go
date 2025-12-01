package less_go

import (
	"fmt"
	"os"
)

// ParserFrame represents a scope frame that can look up variables and properties
type ParserFrame interface {
	Variable(name string) map[string]any
	Property(name string) []any
}

// Variable represents a variable node in the Less AST
type Variable struct {
	*Node
	name      string
	_index    int
	_fileInfo map[string]any
	evaluating bool
}

func NewVariable(name string, index int, currentFileInfo map[string]any) *Variable {
	return &Variable{
		Node:      NewNode(),
		name:      Intern(name),
		_index:    index,
		_fileInfo: currentFileInfo,
	}
}

func (v *Variable) Type() string {
	return "Variable"
}

func (v *Variable) GetType() string {
	return "Variable"
}

func (v *Variable) GetIndex() int {
	return v._index
}

func (v *Variable) FileInfo() map[string]any {
	return v._fileInfo
}

func (v *Variable) GetName() string {
	return v.name
}

func (v *Variable) Eval(context any) (any, error) {
	name := v.name


	if len(name) >= 2 && name[:2] == "@@" {
		innerVar := NewVariable(name[1:], v.GetIndex(), v.FileInfo())
		innerResult, err := innerVar.Eval(context)
		if err != nil {
			return nil, err
		}

		var valueStr string
		switch res := innerResult.(type) {
		case map[string]any:
			if value, exists := res["value"]; exists {
				switch v := value.(type) {
				case *Quoted:
					valueStr = v.value
				case *Anonymous:
					if str, ok := v.Value.(string); ok {
						valueStr = str
					} else {
						valueStr = fmt.Sprintf("%v", v.Value)
					}
				case string:
					valueStr = v
				default:
					valueStr = fmt.Sprintf("%v", v)
				}
			}
		case *Quoted:
			valueStr = res.value
		case *Keyword:
			valueStr = res.value
		case *Color:
			// For colors, use the original form if it's a keyword like "red"
			if res.Value != "" {
				valueStr = res.Value
			} else {
				valueStr = res.ToCSS(context)
			}
		case *Anonymous:
			if quoted, ok := res.Value.(*Quoted); ok {
				valueStr = quoted.value
			} else if str, ok := res.Value.(string); ok {
				valueStr = str
			} else {
				valueStr = fmt.Sprintf("%v", res.Value)
			}
		case interface{ GetValue() any }:
			if val := res.GetValue(); val != nil {
				if str, ok := val.(string); ok {
					valueStr = str
				} else {
					valueStr = fmt.Sprintf("%v", val)
				}
			}
		default:
			if valuer, ok := innerResult.(interface{ GetValue() string }); ok {
				valueStr = valuer.GetValue()
			} else {
				valueStr = fmt.Sprintf("%v", innerResult)
			}
		}
		
		if valueStr != "" {
			name = "@" + valueStr
		}
	}

	if v.evaluating {
		filename := ""
		if v._fileInfo != nil {
			if fn, ok := v._fileInfo["filename"].(string); ok {
				filename = fn
			}
		}
		// Match JavaScript behavior: throw error for circular dependency
		// Use panic to ensure error propagates and isn't silently caught
		panic(&LessError{
			Type:     "Name",
			Message:  fmt.Sprintf("Recursive variable definition for %s", name),
			Filename: filename,
			Index:    v.GetIndex(),
		})
	}

	v.evaluating = true
	defer func() { v.evaluating = false }()

	// Handle different context types - interface-based (tests) or map-based (runtime)
	if interfaceContext, ok := context.(interface{ GetFrames() []ParserFrame }); ok {
		// Interface-based context (for tests)
		frames := interfaceContext.GetFrames()

		for _, frame := range frames {
			if varResult := frame.Variable(name); varResult != nil {
				// Extract values from the pooled map before returning it
				importantVal, hasImportant := varResult["important"]
				val, hasValue := varResult["value"]
				// Return the map to the pool immediately after extracting values
				PutVariableResultMap(varResult)

				if hasImportant && importantVal != nil {
					// For interface context (*Eval), use typed SetImportantInCurrentScope
					if evalCtx, ok := context.(*Eval); ok {
						if boolVal, ok := importantVal.(bool); ok && boolVal {
							evalCtx.SetImportantInCurrentScope("!important")
						} else if strVal, ok := importantVal.(string); ok {
							evalCtx.SetImportantInCurrentScope(strVal)
						}
					} else if ctx, ok := context.(map[string]any); ok {
						// For map context, access importantScope from map
						if importantScopeAny, exists := ctx["importantScope"]; exists {
							if importantScope, ok := importantScopeAny.([]any); ok && len(importantScope) > 0 {
								lastScope := importantScope[len(importantScope)-1]
								if scope, ok := lastScope.(map[string]any); ok {
									if boolVal, ok := importantVal.(bool); ok && boolVal {
										scope["important"] = "!important"
									} else if strVal, ok := importantVal.(string); ok {
										scope["important"] = strVal
									}
								}
							}
						}
					}
				}

				if !hasValue {
					continue
				}

				// If in calc context, wrap vars in a function call to cascade evaluate args first
				if isInCalc, ok := context.(interface{ IsInCalc() bool }); ok && isInCalc.IsInCalc() {
					selfCall := NewCall("_SELF", []any{val}, v.GetIndex(), v.FileInfo())
					// Set the CallerFactory so _SELF can be resolved later
					selfCall.CallerFactory = NewDefaultFunctionCallerFactory(DefaultRegistry)
					// Evaluate the Call object immediately (matches JavaScript: return (new Call('_SELF', [v.value])).eval(context))
					return selfCall.Eval(context)
				}

				if evalable, ok := val.(interface{ Eval(any) (any, error) }); ok {
					result, err := evalable.Eval(context)
					return result, err
				} else if evalCtx, ok := val.(interface{ Eval(EvalContext) (any, error) }); ok {
					if ctx, ok := context.(EvalContext); ok {
						result, err := evalCtx.Eval(ctx)
						return result, err
					} else {
						return val, nil
					}
				} else if evalSingle, ok := val.(interface{ Eval(any) any }); ok {
					// Handle single-return Eval (e.g., DetachedRuleset.Eval)
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Fprintf(os.Stderr, "[DEBUG Variable.Eval] Calling single-return Eval on %T\n", val)
					}
					result := evalSingle.Eval(context)
					return result, nil
				} else {
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Fprintf(os.Stderr, "[DEBUG Variable.Eval] No Eval method found, returning val as-is: %T\n", val)
					}
					return val, nil
				}
			}
		}
	} else if ctx, ok := context.(map[string]any); ok {
		// Map-based context (for runtime)
		framesAny, exists := ctx["frames"]
		if !exists {
			return nil, fmt.Errorf("no frames in evaluation context")
		}

		// Handle frames as []any or []ParserFrame
		var framesList []any
		if frames, ok := framesAny.([]any); ok {
			framesList = frames
		} else if frames, ok := framesAny.([]ParserFrame); ok {
			// Convert []ParserFrame to []any for uniform handling
			framesList = make([]any, len(frames))
			for i, f := range frames {
				framesList[i] = f
			}
		} else {
			return nil, fmt.Errorf("frames is not []any or []ParserFrame")
		}

		for _, frameAny := range framesList {
			// Frames can be Rulesets that have Variable lookup methods
			if frame, ok := frameAny.(interface{ Variable(string) map[string]any }); ok {
				if varResult := frame.Variable(name); varResult != nil {
					// Extract values from the pooled map before returning it
					importantVal, hasImportant := varResult["important"]
					val, hasValue := varResult["value"]
					// Return the map to the pool immediately after extracting values
					PutVariableResultMap(varResult)

					if hasImportant && importantVal != nil {
						if importantScopeAny, exists := ctx["importantScope"]; exists {
							if importantScope, ok := importantScopeAny.([]any); ok && len(importantScope) > 0 {
								lastScope := importantScope[len(importantScope)-1]
								if scope, ok := lastScope.(map[string]any); ok {
									if boolVal, ok := importantVal.(bool); ok && boolVal {
										scope["important"] = "!important"
									} else if strVal, ok := importantVal.(string); ok {
										scope["important"] = strVal
									}
								}
							}
						}
					}

					if !hasValue {
						continue
					}

					// If in calc context, wrap vars in a function call to cascade evaluate args first
					if isInCalc, exists := ctx["inCalc"]; exists && isInCalc == true {
						selfCall := NewCall("_SELF", []any{val}, v.GetIndex(), v.FileInfo())
						// Set the CallerFactory so _SELF can be resolved later
						selfCall.CallerFactory = NewDefaultFunctionCallerFactory(DefaultRegistry)
						// Evaluate the Call object immediately (matches JavaScript: return (new Call('_SELF', [v.value])).eval(context))
						return selfCall.Eval(context)
					}

					if evalable, ok := val.(interface{ Eval(any) (any, error) }); ok {
						result, err := evalable.Eval(context)

						// Continue evaluating if result is still a Variable (handles nested variable references)
						// Use a set to detect circular references
						seen := make(map[*Variable]bool)
						seen[v] = true
						for {
							if resultVar, ok := result.(*Variable); ok && err == nil {
								// Avoid infinite loops - if we've seen this variable before, stop
								if seen[resultVar] {
									// Circular reference detected - stop evaluation
									// This should not happen with the fix to evaluate Expression arguments
									break
								}
								seen[resultVar] = true
								result, err = resultVar.Eval(context)
							} else {
								break
							}
						}

						return result, err
					} else if _, ok := val.(interface{ Eval(EvalContext) (any, error) }); ok {
						// For map context, we can't pass EvalContext, so return value as-is
						return val, nil
					} else if evalSingle, ok := val.(interface{ Eval(any) any }); ok {
						// Handle single-return Eval (e.g., DetachedRuleset.Eval)
						result := evalSingle.Eval(context)
						return result, nil
					} else {
						return val, nil
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("context is neither map[string]any nor interface context")
	}

	filename := ""
	if v._fileInfo != nil {
		if fn, ok := v._fileInfo["filename"].(string); ok {
			filename = fn
		}
	}
	return nil, &LessError{
		Type:     "Name",
		Message:  fmt.Sprintf("variable %s is undefined", name),
		Filename: filename,
		Index:    v.GetIndex(),
	}
}

func (v *Variable) ToCSS(context any) string {
	result, err := v.Eval(context)
	if err != nil {
		// Return the variable name as fallback, similar to how failed evaluation might be handled
		return v.name
	}

	if cssObj, ok := result.(interface{ ToCSS(any) string }); ok {
		return cssObj.ToCSS(context)
	} else if str, ok := result.(string); ok {
		return str
	} else {
		return fmt.Sprintf("%v", result)
	}
}

func (v *Variable) GenCSS(context any, output *CSSOutput) {
	result, err := v.Eval(context)
	if err != nil {
		// On error, output the variable name as fallback
		output.Add(v.name, v.FileInfo(), v.GetIndex())
		return
	}

	if gen, ok := result.(interface{ GenCSS(any, *CSSOutput) }); ok {
		gen.GenCSS(context, output)
	} else if str, ok := result.(string); ok {
		output.Add(str, v.FileInfo(), v.GetIndex())
	} else if cssObj, ok := result.(interface{ ToCSS(any) string }); ok {
		output.Add(cssObj.ToCSS(context), v.FileInfo(), v.GetIndex())
	} else {
		output.Add(fmt.Sprintf("%v", result), v.FileInfo(), v.GetIndex())
	}
}

