package less_go

import (
	"fmt"
	"os"
)

// NOTE: Slice/map pooling was considered but is not used because:
// - arguments slice is retained by NewExpression
// - mixinFrames slice is stored in evalContext["frames"]
// - evalFrames slice is stored in evalContext["frames"]
// Instead, we optimize via pre-allocation with exact capacity.

// toCSS converts a value to its CSS string representation
func toCSS(val any, context any) string {
	if cssGenerator, ok := val.(interface{ ToCSS(any) string }); ok {
		return cssGenerator.ToCSS(context)
	}
	return fmt.Sprintf("%v", val)
}

// continueEvaluatingVariables continues evaluating if the result is a Variable
// This handles cases where a Variable's value is another Variable (e.g., nested mixin parameters)
func continueEvaluatingVariables(val any, context any) any {
	seen := make(map[*Variable]bool)
	for {
		if varResult, ok := val.(*Variable); ok {
			// Avoid infinite loops - if we've seen this variable before, stop
			if seen[varResult] {
				break
			}
			seen[varResult] = true
			if evalVar, ok := val.(interface{ Eval(any) any }); ok {
				val = evalVar.Eval(context)
			} else {
				break
			}
		} else {
			break
		}
	}
	return val
}

// MixinDefinition represents a mixin definition node in the Less AST
type MixinDefinition struct {
	*Ruleset
	Name               string
	Params             []any
	Condition          any
	Variadic           bool
	Arity              int
	Lookups            map[string][]any
	Required           int
	OptionalParameters []string
	Frames             []any
}

// NewMixinDefinition creates a new MixinDefinition instance
func NewMixinDefinition(name string, params []any, rules []any, condition any, variadic bool, frames []any, visibilityInfo map[string]any) (*MixinDefinition, error) {
	// Handle default name
	if name == "" {
		name = "anonymous mixin"
	}
	name = Intern(name)

	// Create selectors array with single selector containing element
	element := NewElement(nil, name, false, 0, make(map[string]any), nil)
	selector, err := NewSelector([]*Element{element}, nil, nil, 0, make(map[string]any), visibilityInfo)
	if err != nil {
		return nil, err
	}

	// Create base ruleset
	ruleset := NewRuleset([]any{selector}, rules, false, visibilityInfo)
	ruleset.AllowRoot = true

	// NOTE: We do NOT mark nested rulesets as InsideMixinDefinition here.
	// This is because NewMixinDefinition is used for two cases:
	// 1. Creating real mixin definitions from source (parser.go) - needs marking
	// 2. Wrapping existing rulesets to call them as mixins (mixin_call.go) - should NOT be marked
	//
	// The marking is done in the parser where the mixin definition is created from source.

	// Calculate arity
	arity := 0
	if params != nil {
		arity = len(params)
	}

	// Calculate required parameters and collect optional ones
	required := 0
	optionalParameters := []string{}
	
	for _, p := range params {
		if param, ok := p.(map[string]any); ok {
			paramName := param["name"]
			paramValue := param["value"]
			
			// Extract name string from either string or Variable
			var nameStr string
			if s, ok := paramName.(string); ok {
				nameStr = s
			} else if v, ok := paramName.(*Variable); ok {
				nameStr = v.GetName()
			}
			
			// JavaScript logic: !p.name || (p.name && !p.value)
			// Required if: no name OR empty string name OR (has name but no value)
			if paramName == nil || nameStr == "" || (paramName != nil && paramValue == nil) {
				required++
			} else {
				// Has both name and value - it's optional
				if nameStr != "" {
					optionalParameters = append(optionalParameters, nameStr)
				}
			}
		}
	}

	md := &MixinDefinition{
		Ruleset:            ruleset,
		Name:               name,
		Params:             params,
		Condition:          condition,
		Variadic:           variadic,
		Arity:              arity,
		Lookups:            make(map[string][]any),
		Required:           required,
		OptionalParameters: optionalParameters,
		Frames:             frames,
	}

	return md, nil
}

// GetType returns the type of the node
func (md *MixinDefinition) GetType() string {
	return "MixinDefinition"
}

// Type returns the type of the node (for compatibility)
func (md *MixinDefinition) Type() string {
	return "MixinDefinition"
}

// EvalFirst indicates this node should be evaluated first
func (md *MixinDefinition) EvalFirst() bool {
	return true
}

// IsRuleset returns true (MixinDefinition inherits from Ruleset in JavaScript)
func (md *MixinDefinition) IsRuleset() bool {
	return true
}

// Accept visits the mixin definition with a visitor
func (md *MixinDefinition) Accept(visitor any) {
	// Try the variadic signature first (matches Visitor.VisitArray signature)
	if v, ok := visitor.(interface{ VisitArray([]any, ...bool) []any }); ok {
		if len(md.Params) > 0 {
			md.Params = v.VisitArray(md.Params)
		}
		if md.Rules != nil {
			md.Rules = v.VisitArray(md.Rules)
		}
	} else if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
		// Fallback to non-variadic signature for compatibility
		if len(md.Params) > 0 {
			md.Params = v.VisitArray(md.Params)
		}
		if md.Rules != nil {
			md.Rules = v.VisitArray(md.Rules)
		}
	}

	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		if md.Condition != nil {
			md.Condition = v.Visit(md.Condition)
		}
	}
}

// EvalParams evaluates mixin parameters and creates parameter frame
func (md *MixinDefinition) EvalParams(context any, mixinEnv any, args []any, evaldArguments []any) (*Ruleset, error) {
	frame := NewRuleset(nil, nil, false, nil)
	
	
	var varargs []any
	var arg any
	params := md.Params // No need to copy - params are only read, not modified
	var val any
	var name string
	var isNamedFound bool
	argIndex := 0
	argsLength := 0
	
	// Set up function registry inheritance
	if env, ok := mixinEnv.(map[string]any); ok {
		if frames, ok := env["frames"].([]any); ok && len(frames) > 0 {
			if frameMap, ok := frames[0].(map[string]any); ok {
				if funcRegistry, ok := frameMap["functionRegistry"]; ok {
					frame.FunctionRegistry = funcRegistry
				}
			}
		}
	}

	// Create new evaluation context
	// Pre-calculate frames capacity and use copy instead of append
	var newFrames []any
	if env, ok := mixinEnv.(map[string]any); ok {
		if frames, ok := env["frames"].([]any); ok {
			newFrames = make([]any, 1+len(frames))
			newFrames[0] = frame
			copy(newFrames[1:], frames)
		} else {
			newFrames = []any{frame}
		}
	} else {
		newFrames = []any{frame}
	}

	// Pre-allocate evalContext with capacity based on mixinEnv size
	var evalContext map[string]any
	if env, ok := mixinEnv.(map[string]any); ok {
		evalContext = make(map[string]any, len(env))
		evalContext["frames"] = newFrames
		for k, v := range env {
			if k != "frames" {
				evalContext[k] = v
			}
		}
	} else {
		evalContext = map[string]any{
			"frames": newFrames,
		}
	}

	if args != nil {
		args = CopyArray(args)
		argsLength = len(args)

		// Process named arguments first
		for i := 0; i < argsLength; i++ {
			if argMap, ok := args[i].(map[string]any); ok {
				if argName, ok := argMap["name"].(string); ok && argName != "" {
					name = argName
					isNamedFound = false
					
					// Ensure evaldArguments has enough slots
					for len(evaldArguments) < len(params) {
						evaldArguments = append(evaldArguments, nil)
					}
					
					for j := 0; j < len(params); j++ {
						if evaldArguments[j] != nil {
							continue
						}
						
						if paramMap, ok := params[j].(map[string]any); ok {
							// Handle both string names and Variable objects
							var paramNameStr string
							if paramName, ok := paramMap["name"].(string); ok {
								paramNameStr = paramName
							} else if paramVar, ok := paramMap["name"].(*Variable); ok {
								paramNameStr = paramVar.GetName()
							}
							
							
							if paramNameStr != "" && name == paramNameStr {
								// Handle values that implement Eval with error return
								if argValue, ok := argMap["value"].(interface{ Eval(any) (any, error) }); ok {
									evalResult, err := argValue.Eval(context)
									if err != nil {
										return nil, err
									}
									// Continue evaluating if result is still a Variable
									evalResult = continueEvaluatingVariables(evalResult, context)
									evaldArguments[j] = evalResult
									// Create declaration and prepend to frame
									decl, err := NewDeclaration(name, evaldArguments[j], nil, false, 0, nil, false, true)
									if err != nil {
										return nil, err
									}
									frame.PrependRule(decl)
									isNamedFound = true
									break
								} else if argValue, ok := argMap["value"].(interface{ Eval(any) any }); ok {
									// Handle values that implement Eval without error return
									evalResult := argValue.Eval(context)
									// Continue evaluating if result is still a Variable
									evalResult = continueEvaluatingVariables(evalResult, context)
									evaldArguments[j] = evalResult
									// Create declaration and prepend to frame
									decl, err := NewDeclaration(name, evaldArguments[j], nil, false, 0, nil, false, true)
									if err != nil {
										return nil, err
									}
									frame.PrependRule(decl)
									isNamedFound = true
									break
								} else {
									// If value doesn't implement Eval, use it directly
									evaldArguments[j] = argMap["value"]
									// Create declaration and prepend to frame
									decl, err := NewDeclaration(name, evaldArguments[j], nil, false, 0, nil, false, true)
									if err != nil {
										return nil, err
									}
									frame.PrependRule(decl)
									isNamedFound = true
									break
								}
							}
						}
					}
					
					if isNamedFound {
						// Remove processed argument
						newArgs := make([]any, len(args)-1)
						copy(newArgs, args[:i])
						copy(newArgs[i:], args[i+1:])
						args = newArgs
						i--
						argsLength--
						continue
					} else {
						return nil, fmt.Errorf("named argument for %s %s not found", md.Name, name)
					}
				}
			}
		}
	}

	// Process positional arguments  
	argIndex = 0
	for i := 0; i < len(params); i++ {
		if i < len(evaldArguments) && evaldArguments[i] != nil {
			continue
		}

		arg = nil
		if args != nil && argIndex < len(args) {
			arg = args[argIndex]
		}

		if paramMap, ok := params[i].(map[string]any); ok {
			// Extract name string from either string or Variable
			var paramNameStr string
			if s, ok := paramMap["name"].(string); ok {
				paramNameStr = s
			} else if v, ok := paramMap["name"].(*Variable); ok {
				paramNameStr = v.GetName()
			}
			
			if paramNameStr != "" {
				name = paramNameStr
				// Debug: check if name has @ prefix
				// fmt.Printf("DEBUG: Parameter name: %s\n", name)
				
				if variadic, ok := paramMap["variadic"].(bool); ok && variadic {
					varargs = make([]any, 0, argsLength-argIndex)
					for j := argIndex; j < argsLength; j++ {
						if argMap, ok := args[j].(map[string]any); ok {
							argValue := argMap["value"]
							// Try to evaluate the value - handle both Eval signatures
							// Match JavaScript: args[j].value.eval(context)
							if evalMethod, ok := argValue.(interface{ Eval(any) (any, error) }); ok {
								evaluated, err := evalMethod.Eval(context)
								if err == nil {
									argValue = evaluated
								}
								// If error, use unevaluated value
							} else if evalMethod, ok := argValue.(interface{ Eval(any) any }); ok {
								argValue = evalMethod.Eval(context)
							}
							// If no Eval method, use the value as-is (e.g., Keywords)

							// IMPORTANT: Flatten Expression/Value types with multiple items
							// This handles cases like .mixin(a b c d) where space-separated
							// values should become individual arguments for ...
							// BUT: Only flatten if there's a single argument total
							if argsLength == 1 {
								if exprValue, ok := argValue.(*Expression); ok && len(exprValue.Value) > 0 {
									varargs = append(varargs, exprValue.Value...)
								} else if valueValue, ok := argValue.(*Value); ok && len(valueValue.Value) > 0 {
									varargs = append(varargs, valueValue.Value...)
								} else {
									varargs = append(varargs, argValue)
								}
							} else {
								// Multiple arguments - don't flatten
								varargs = append(varargs, argValue)
							}
						}
					}
					expr, err := NewExpression(varargs, false)
					if err != nil {
						return nil, err
					}
					// Match JavaScript: new Expression(varargs).eval(context)
					// Use context, not evalContext, to preserve math mode and other settings
					evalExpr, err := expr.Eval(context)
					if err != nil {
						return nil, err
					}
					decl, err := NewDeclaration(name, evalExpr, nil, false, 0, nil, false, true)
					if err != nil {
						return nil, err
					}
					frame.PrependRule(decl)
				} else {
					// Non-variadic parameter
					if arg != nil {
						if argMap, ok := arg.(map[string]any); ok {
							if argValue, ok := argMap["value"]; ok {
								// Handle detached ruleset
								if argSlice, ok := argValue.([]any); ok {
									ruleset := NewRuleset(nil, argSlice, false, nil)
									val = NewDetachedRuleset(ruleset, nil)
								} else if evalValue, ok := argValue.(interface{ Eval(any) (any, error) }); ok {
									// Check for Eval with error return first (more common in Go)
									// This handles Expression, Variable, and other nodes that return (any, error)
									evalResult, err := evalValue.Eval(context)
									if err != nil {
										return nil, err
									}
									val = evalResult
									// Continue evaluating if result is still a Variable (handles nested variable references)
									val = continueEvaluatingVariables(val, context)
								} else if evalValue, ok := argValue.(interface{ Eval(any) any }); ok {
									// Check for Eval without error return (less common)
									// This handles DetachedRuleset and other nodes that return any
									val = evalValue.Eval(context)
									// Continue evaluating if result is still a Variable (handles nested variable references)
									val = continueEvaluatingVariables(val, context)
								} else {
									val = argValue
								}
							} else {
								// No 'value' key in argMap
							}
						} else {
							// Try to handle direct values
							if evalValue, ok := arg.(interface{ Eval(any) any }); ok {
								val = evalValue.Eval(context)
							} else {
								val = arg
							}
						}
					} else if paramValue, ok := paramMap["value"]; ok && paramValue != nil {
						// Try both Eval signatures
						if evalValue, ok := paramValue.(interface{ Eval(any) (any, error) }); ok {
							// Default parameters should use an environment that includes the frame being built
							// This allows them to reference earlier parameters
							defaultParamEnv := mixinEnv
							if mixinEnvMap, ok := mixinEnv.(map[string]any); ok {
								// Create new environment with frame prepended to frames
								defaultParamEnv = make(map[string]any)
								for k, v := range mixinEnvMap {
									defaultParamEnv.(map[string]any)[k] = v
								}
								if frames, ok := mixinEnvMap["frames"].([]any); ok {
									updatedFrames := []any{frame}
									updatedFrames = append(updatedFrames, frames...)
									defaultParamEnv.(map[string]any)["frames"] = updatedFrames
								}
							}
							evalResult, err := evalValue.Eval(defaultParamEnv)
							if err != nil {
								return nil, err
							}
							// Continue evaluating if result is still a Variable
							evalResult = continueEvaluatingVariables(evalResult, defaultParamEnv)
							val = evalResult
							frame.ResetCache()
						} else if evalValue, ok := paramValue.(interface{ Eval(any) any }); ok {
							// Create updated environment for default parameter evaluation
							defaultParamEnv := mixinEnv
							if mixinEnvMap, ok := mixinEnv.(map[string]any); ok {
								// Create new environment with frame prepended to frames
								defaultParamEnv = make(map[string]any)
								for k, v := range mixinEnvMap {
									defaultParamEnv.(map[string]any)[k] = v
								}
								if frames, ok := mixinEnvMap["frames"].([]any); ok {
									updatedFrames := []any{frame}
									updatedFrames = append(updatedFrames, frames...)
									defaultParamEnv.(map[string]any)["frames"] = updatedFrames
								}
							}
							val = evalValue.Eval(defaultParamEnv)
							// Continue evaluating if result is still a Variable
							val = continueEvaluatingVariables(val, defaultParamEnv)
							frame.ResetCache()
						} else {
						}
					} else {
						return nil, fmt.Errorf("wrong number of arguments for %s (%d for %d)", md.Name, argsLength, md.Arity)
					}

					decl, err := NewDeclaration(name, val, nil, false, 0, nil, false, true)
					if err != nil {
						return nil, err
					}
					frame.PrependRule(decl)
					// Populate evaldArguments if within bounds
					if i < len(evaldArguments) {
						evaldArguments[i] = val
					}
				}
			} else {
				// Handle ... syntax without a parameter name
				if variadic, ok := paramMap["variadic"].(bool); ok && variadic {
					// For anonymous variadic parameters, do NOT populate evaldArguments[i]
					// The individual arguments will be populated later in the variadic handling block (lines 537-561)
					// This matches JavaScript behavior where anonymous variadic params don't set evaldArguments[i]
					// Note: Don't create a parameter declaration since there's no name
					// The @arguments variable will be created later in EvalCall
				}
			}

			// Handle variadic parameter arguments BEFORE incrementing argIndex
			// This matches JavaScript behavior where the variadic handling (lines 130-134)
			// happens before argIndex++ (line 135)
			if variadic, ok := paramMap["variadic"].(bool); ok && variadic && args != nil {
				// For variadic parameters, populate evaldArguments with evaluated args
				// This matches JavaScript lines 130-134
				for j := argIndex; j < argsLength; j++ {
					if j < len(evaldArguments) {
						if argMap, ok := args[j].(map[string]any); ok {
							if argValue, ok := argMap["value"]; ok {
								if evalValue, ok := argValue.(interface{ Eval(any) any }); ok {
									argValue = evalValue.Eval(context)
								}
								// Apply flattening here too
								if exprValue, ok := argValue.(*Expression); ok && len(exprValue.Value) > 0 {
									// Can't easily expand here, just store the Expression
									// The primary handling above should have already flattened
									evaldArguments[j] = argValue
								} else if valueValue, ok := argValue.(*Value); ok && len(valueValue.Value) > 0 {
									evaldArguments[j] = argValue
								} else {
									evaldArguments[j] = argValue
								}
							}
						}
					}
				}
				argIndex = argsLength // Set to end since variadic consumes all remaining
			}

			// Always increment argIndex to match JavaScript behavior (line 135)
			// This happens AFTER variadic handling
			argIndex++
		}
	}

	return frame, nil
}

// MakeImportant creates a new MixinDefinition with important rules
func (md *MixinDefinition) MakeImportant() any {
	var rules []any
	if md.Rules != nil {
		rules = make([]any, len(md.Rules))
		for i, rule := range md.Rules {
			if imp, ok := rule.(interface{ MakeImportant() any }); ok {
				rules[i] = imp.MakeImportant()
			} else {
				rules[i] = rule
			}
		}
	}

	result, _ := NewMixinDefinition(md.Name, md.Params, rules, md.Condition, md.Variadic, md.Frames, md.VisibilityInfo())
	return result
}

// Eval evaluates the mixin definition
func (md *MixinDefinition) Eval(context any) (*MixinDefinition, error) {
	frames := md.Frames
	if frames == nil {
		// Try *Eval context first (most common during evaluation)
		if evalCtx, ok := context.(*Eval); ok {
			if evalCtx.Frames != nil {
				frames = CopyArray(evalCtx.Frames)
			}
		} else if ctx, ok := context.(map[string]any); ok {
			if ctxFrames, ok := ctx["frames"].([]any); ok {
				frames = CopyArray(ctxFrames)
			}
		}
	}

	// Important: Do NOT evaluate the rules here. The rules should only be evaluated
	// when the mixin is called via EvalCall
	// NewMixinDefinition will copy the rules internally via NewRuleset, so no need for CopyArray
	result, err := NewMixinDefinition(md.Name, md.Params, md.Rules, md.Condition, md.Variadic, frames, md.VisibilityInfo())
	if err != nil {
		return nil, err
	}
	return result, nil
}

// EvalCall evaluates a mixin call
func (md *MixinDefinition) EvalCall(context any, args []any, important bool) (*Ruleset, error) {
	// Debug: trace incoming context
	if os.Getenv("LESS_GO_TRACE") != "" {
		switch ctx := context.(type) {
		case *Eval:
			fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] %s: *Eval context, mediaPath len=%d\n", md.Name, len(ctx.MediaPath))
		case map[string]any:
			if mp, ok := ctx["mediaPath"].([]any); ok {
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] %s: map context, mediaPath len=%d\n", md.Name, len(mp))
			} else {
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] %s: map context, mediaPath is nil\n", md.Name)
			}
		}
	}

	// Arguments array will be populated by EvalParams
	// For variadic mixins, we need to allocate space for all arguments
	// For non-variadic, allocate space for parameters only
	argumentsSize := len(md.Params)
	if md.Variadic && args != nil && len(args) > len(md.Params) {
		argumentsSize = len(args)
	}
	arguments := make([]any, argumentsSize)

	// Determine mixin frames - pre-allocate with exact capacity to avoid reallocation
	var mixinFrames []any
	var ctxFrames []any
	if ctx, ok := context.(map[string]any); ok {
		ctxFrames, _ = ctx["frames"].([]any)
	} else if evalCtx, ok := context.(*Eval); ok {
		ctxFrames = evalCtx.Frames
	}

	if md.Frames != nil {
		if ctxFrames != nil {
			// Pre-allocate with exact capacity
			mixinFrames = make([]any, len(md.Frames)+len(ctxFrames))
			copy(mixinFrames, md.Frames)
			copy(mixinFrames[len(md.Frames):], ctxFrames)
		} else {
			mixinFrames = md.Frames
		}
	} else {
		mixinFrames = ctxFrames
	}

	// Create mixin environment
	// NOTE: Do NOT copy mediaBlocks/mediaPath from parent context, matching JavaScript's
	// contexts.Eval which doesn't include them in evalCopyProperties
	// Pre-allocate mixinEnv with capacity based on context size
	var mixinEnv map[string]any
	if ctx, ok := context.(map[string]any); ok {
		mixinEnv = make(map[string]any, len(ctx))
		mixinEnv["frames"] = mixinFrames
		for k, v := range ctx {
			if k != "frames" && k != "mediaBlocks" && k != "mediaPath" {
				mixinEnv[k] = v
			}
		}
	} else if evalCtx, ok := context.(*Eval); ok {
		// Efficiently copy Eval fields to map, but NOT mediaBlocks/mediaPath
		// (matching JavaScript's contexts.Eval which doesn't include them in evalCopyProperties)
		mixinEnv = make(map[string]any, 8) // Typical Eval context has ~8 fields
		mixinEnv["frames"] = mixinFrames
		evalCtx.CopyEvalToMap(mixinEnv, false)
	} else {
		mixinEnv = map[string]any{
			"frames": mixinFrames,
		}
	}

	// Evaluate parameters
	frame, err := md.EvalParams(context, mixinEnv, args, arguments)
	if err != nil {
		return nil, err
	}

	// Add @arguments declaration
	expr, err := NewExpression(arguments, false)
	if err != nil {
		return nil, err
	}
	evalExpr, err := expr.Eval(context)
	if err != nil {
		return nil, err
	}
	argDecl, err := NewDeclaration("@arguments", evalExpr, nil, false, 0, nil, false, true)
	if err != nil {
		return nil, err
	}
	frame.PrependRule(argDecl)

	// Copy rules
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] %s: md.Rules has %d rules, md=%p\n", md.Name, len(md.Rules), md)
		for i, rule := range md.Rules {
			if imp, ok := rule.(*Import); ok {
				rootType := fmt.Sprintf("%T", imp.root)
				hasInline := imp.getBoolOption("inline")
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall]   rule[%d] Import=%p: root type=%s, inline=%v, importedFilename=%s\n", i, imp, rootType, hasInline, imp.importedFilename)
			} else if rs, ok := rule.(*Ruleset); ok {
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall]   rule[%d] Ruleset=%p: %d rules, %d selectors\n", i, rs, len(rs.Rules), len(rs.Selectors))
				// Check for imports inside the ruleset
				for j, innerRule := range rs.Rules {
					if innerImp, ok := innerRule.(*Import); ok {
						innerRootType := fmt.Sprintf("%T", innerImp.root)
						innerHasInline := innerImp.getBoolOption("inline")
						fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall]     inner rule[%d] Import=%p: root type=%s, inline=%v\n", j, innerImp, innerRootType, innerHasInline)
					}
				}
			} else {
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall]   rule[%d] type=%T\n", i, rule)
			}
		}
	}
	// Create result ruleset - NewRuleset will copy the rules internally, so no need for CopyArray
	ruleset := NewRuleset(nil, md.Rules, false, nil)
	// Match JavaScript: ruleset.originalRuleset = this
	// If this MixinDefinition was created as a wrapper for a Ruleset,
	// use the wrapped Ruleset as the originalRuleset for recursion detection.
	// Otherwise, use the MixinDefinition's own Ruleset field.
	if md.OriginalRuleset != nil {
		ruleset.OriginalRuleset = md.OriginalRuleset
	} else {
		ruleset.OriginalRuleset = md.Ruleset
	}

	// Evaluate ruleset with proper context
	// OPTIMIZATION: When context is *Eval, we pass *Eval directly to Ruleset.Eval
	// instead of converting to map[string]any. This saves ~27MB of allocations per profile.
	// Pre-allocate evalFrames with exact size to avoid reallocation
	evalFramesLen := 2 + len(mixinFrames)
	evalFrames := make([]any, evalFramesLen)
	evalFrames[0] = md
	evalFrames[1] = frame
	if len(mixinFrames) > 0 {
		copy(evalFrames[2:], mixinFrames)
	}

	// Re-load plugins from ancestor frames so that mixin body can access plugin functions
	// This is needed because when #ns { @plugin "..."; .mixin() { ... } } is called,
	// the mixin inherits functions from the parent namespace's @plugin
	//
	// We check for Rulesets that have LoadedPluginFunctions stored on them (set during
	// Ruleset.Eval when plugins are loaded).
	//
	// IMPORTANT: We enter a new plugin scope BEFORE re-registering inherited functions.
	// This ensures that inherited functions are scoped to the mixin call and don't leak
	// to the caller's scope when the mixin completes. See Issue: plugin function scoping.
	debug := os.Getenv("LESS_GO_DEBUG") == "1"
	if debug {
		fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Checking %d mixinFrames for plugins\n", len(mixinFrames))
	}

	// Get plugin bridge from context (handles both *Eval and map contexts)
	var pluginBridge any
	if evalCtx, ok := context.(*Eval); ok {
		if evalCtx.LazyPluginBridge != nil {
			pluginBridge = evalCtx.LazyPluginBridge
		} else if evalCtx.PluginBridge != nil {
			pluginBridge = evalCtx.PluginBridge
		}
	} else if ctx, ok := context.(map[string]any); ok {
		pluginBridge = ctx["pluginBridge"]
	}

	// Enter a new plugin scope for this mixin call to prevent function leakage
	var mixinScopeExitFunc func()
	if pluginBridge != nil {
		if lazyBridge, ok := pluginBridge.(*LazyNodeJSPluginBridge); ok {
			if bridge, err := lazyBridge.GetBridge(); err == nil && bridge != nil {
				bridge.EnterScope()
				mixinScopeExitFunc = func() {
					bridge.ExitScope()
					if debug {
						fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Exiting mixin plugin scope\n")
					}
				}
				if debug {
					fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Entered mixin plugin scope\n")
				}
			}
		} else if bridge, ok := pluginBridge.(*NodeJSPluginBridge); ok {
			bridge.EnterScope()
			mixinScopeExitFunc = func() {
				bridge.ExitScope()
				if debug {
					fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Exiting mixin plugin scope\n")
				}
			}
			if debug {
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Entered mixin plugin scope\n")
			}
		}
	}
	// Ensure we exit the scope when this function returns
	if mixinScopeExitFunc != nil {
		defer mixinScopeExitFunc()
	}

	// Now re-register inherited plugin functions within the mixin's scope
	if pluginBridge != nil {
		for frameIdx, frameAny := range mixinFrames {
			if rs, ok := frameAny.(*Ruleset); ok {
				if debug {
					fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall]   Frame %d is *Ruleset, LoadedPluginFunctions=%v\n", frameIdx, rs.LoadedPluginFunctions)
				}
				// If this ruleset loaded plugins, re-load them at the current scope depth
				if rs.LoadedPluginFunctions != nil && len(rs.LoadedPluginFunctions) > 0 {
					if debug {
						fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Found %d loaded plugin functions in frame %d\n", len(rs.LoadedPluginFunctions), frameIdx)
					}
					// The plugin functions are already registered in the Node.js runtime,
					// we just need to re-add them at the current scope depth (now in mixin's scope)
					if lazyBridge, ok := pluginBridge.(*LazyNodeJSPluginBridge); ok {
						if bridge, err := lazyBridge.GetBridge(); err == nil && bridge != nil {
							for funcName := range rs.LoadedPluginFunctions {
								if debug {
									fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Re-registering function '%s' at mixin scope\n", funcName)
								}
								// Register the function at the mixin's scope depth
								bridge.AddFunctionToCurrentScope(funcName)
							}
						}
					} else if bridge, ok := pluginBridge.(*NodeJSPluginBridge); ok {
						for funcName := range rs.LoadedPluginFunctions {
							if debug {
								fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] Re-registering function '%s' at mixin scope\n", funcName)
							}
							bridge.AddFunctionToCurrentScope(funcName)
						}
					}
				}
			} else if debug {
				fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall]   Frame %d is type %T (not *Ruleset)\n", frameIdx, frameAny)
			}
		}
	} else if debug {
		fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] No pluginBridge in context\n")
	}

	// Pre-allocate evalContext with capacity based on context size
	var evalContext map[string]any
	if ctx, ok := context.(map[string]any); ok {
		evalContext = make(map[string]any, len(ctx))
		evalContext["frames"] = evalFrames
		for k, v := range ctx {
			// Skip "frames" (already set), and "mediaBlocks"/"mediaPath" which should NOT be
			// inherited. In JavaScript's contexts.Eval constructor, mediaBlocks and mediaPath
			// are NOT in evalCopyProperties, meaning mixin body evaluation gets a fresh media
			// context. This is critical for correct media query merging when a mixin contains
			// nested @media rules with detached ruleset calls.
			if k != "frames" && k != "mediaBlocks" && k != "mediaPath" {
				evalContext[k] = v
			}
		}
	} else if evalCtx, ok := context.(*Eval); ok {
		// Efficiently copy Eval fields to map (with closures for compatibility)
		// Pass false for includeMediaContext to NOT copy mediaBlocks/mediaPath
		evalContext = make(map[string]any, 8) // Typical Eval context has ~8 fields
		evalContext["frames"] = evalFrames
		evalCtx.CopyEvalToMap(evalContext, false)
	} else {
		evalContext = map[string]any{
			"frames": evalFrames,
		}
	}

	// Debug: trace evalContext mediaPath after copy
	if os.Getenv("LESS_GO_TRACE") != "" {
		if mp, ok := evalContext["mediaPath"].([]any); ok {
			fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] %s: evalContext after copy, mediaPath len=%d\n", md.Name, len(mp))
		} else {
			fmt.Fprintf(os.Stderr, "[MixinDefinition.EvalCall] %s: evalContext after copy, mediaPath is nil/not []any\n", md.Name)
		}
	}

	evaluated, err := ruleset.Eval(evalContext)
	if err != nil {
		return nil, err
	}

	// NOTE: Do NOT propagate mediaBlocks/mediaPath back to parent context.
	// Since the mixin body evaluation now has a fresh media context (matching JavaScript
	// where mediaBlocks/mediaPath are NOT in evalCopyProperties), the media blocks
	// collected in the mixin body are returned as part of the evaluated rules via
	// Media.evalTop's return value (either the single media node or a MultiMedia ruleset).
	// The parent will receive these media nodes as part of the mixin's evaluated rules.

	evaluatedRuleset, ok := evaluated.(*Ruleset)
	if !ok {
		return nil, fmt.Errorf("expected *Ruleset from Eval, got %T", evaluated)
	}

	// NOTE: Visibility handling for mixin content is done in MixinCall.Eval(), not here.
	// This is because we need to know whether the mixin CALL blocks visibility
	// (i.e., whether the call originates from a reference import).
	//
	// MixinCall.Eval() will:
	// 1. Call ensureMixinContentVisibility() if the call does NOT block visibility
	// 2. Call setVisibilityToReplacement() which adds blocks if the call DOES block visibility

	if important {
		importantResult := evaluatedRuleset.MakeImportant()
		if importantRuleset, ok := importantResult.(*Ruleset); ok {
			evaluatedRuleset = importantRuleset
		} else {
			return nil, fmt.Errorf("expected *Ruleset from MakeImportant, got %T", importantResult)
		}
	}

	return evaluatedRuleset, nil
}

// MatchCondition checks if the mixin condition matches
func (md *MixinDefinition) MatchCondition(args []any, context any) bool {
	if md.Condition == nil {
		return true
	}

	// Create evaluation context similar to JavaScript version
	// Match JavaScript: new contexts.Eval(context, this.frames ? this.frames.concat(context.frames) : context.frames)
	// This preserves all context properties (including defaultFunc) while updating frames
	var mixinFrames []any
	var ctxFrames []any
	if ctx, ok := context.(map[string]any); ok {
		ctxFrames, _ = ctx["frames"].([]any)
	} else if evalCtx, ok := context.(*Eval); ok {
		ctxFrames = evalCtx.Frames
	}

	if md.Frames != nil {
		if ctxFrames != nil {
			// Pre-allocate with exact capacity
			mixinFrames = make([]any, len(md.Frames)+len(ctxFrames))
			copy(mixinFrames, md.Frames)
			copy(mixinFrames[len(md.Frames):], ctxFrames)
		} else {
			mixinFrames = md.Frames
		}
	} else {
		mixinFrames = ctxFrames
	}

	// Create new context preserving all properties from original context
	var mixinEnv any
	if evalCtx, ok := context.(*Eval); ok {
		// Preserve *Eval context with DefaultFunc
		newEvalCtx := &Eval{}
		*newEvalCtx = *evalCtx
		newEvalCtx.Frames = mixinFrames
		mixinEnv = newEvalCtx
		debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
		if debugTrace {
			fmt.Printf("[TRACE] MixinDefinition.Eval: cloned Eval context, MathOn=%v\n", newEvalCtx.MathOn)
		}
	} else {
		// Map context - copy properties with pre-allocated capacity
		var mixinEnvMap map[string]any
		if ctx, ok := context.(map[string]any); ok {
			mixinEnvMap = make(map[string]any, len(ctx)+1)
			for k, v := range ctx {
				mixinEnvMap[k] = v
			}
		} else {
			mixinEnvMap = make(map[string]any, 1)
		}
		mixinEnvMap["frames"] = mixinFrames
		mixinEnv = mixinEnvMap
	}

	paramFrame, err := md.EvalParams(context, mixinEnv, args, []any{})
	if err != nil {
		return false
	}

	// Pre-calculate total frames length for single allocation
	evalFramesLen := 1 // paramFrame
	if md.Frames != nil {
		evalFramesLen += len(md.Frames)
	}
	// ctxFrames already extracted above
	if ctxFrames != nil {
		evalFramesLen += len(ctxFrames)
	}

	evalFrames := make([]any, evalFramesLen)
	evalFrames[0] = paramFrame
	idx := 1
	if md.Frames != nil {
		copy(evalFrames[idx:], md.Frames)
		idx += len(md.Frames)
	}
	if ctxFrames != nil {
		copy(evalFrames[idx:], ctxFrames)
	}

	// Create evaluation context preserving type
	var evalContext any
	if evalCtx, ok := context.(*Eval); ok {
		// Preserve *Eval context with DefaultFunc
		newEvalCtx := &Eval{}
		*newEvalCtx = *evalCtx
		newEvalCtx.Frames = evalFrames
		evalContext = newEvalCtx
	} else {
		// Map context - pre-allocate with capacity based on context size
		var evalContextMap map[string]any
		if ctx, ok := context.(map[string]any); ok {
			evalContextMap = make(map[string]any, len(ctx))
			evalContextMap["frames"] = evalFrames
			for k, v := range ctx {
				if k != "frames" {
					evalContextMap[k] = v
				}
			}
		} else {
			evalContextMap = map[string]any{
				"frames": evalFrames,
			}
		}
		evalContext = evalContextMap
	}

	if condEval, ok := md.Condition.(interface{ Eval(any) any }); ok {
		result := condEval.Eval(evalContext)

		debug := os.Getenv("LESS_DEBUG_GUARDS") == "1"
		if debug {
			condType := fmt.Sprintf("%T", md.Condition)
			if cond, ok := md.Condition.(*Condition); ok {
				condType = fmt.Sprintf("*Condition(op=%s)", cond.Op)
			}
			fmt.Printf("DEBUG:  MatchCondition for '%s': condition type=%s, result=%v (%T)\n", md.Name, condType, result, result)
		}

		// Check if result is falsy
		if result == nil {
			return false
		}
		if b, ok := result.(bool); ok {
			return b
		}
		// Non-nil, non-boolean values are truthy
		return true
	}

	return false
}


// MatchArgs checks if the mixin arguments match
func (md *MixinDefinition) MatchArgs(args []any, context any) bool {
	debug := os.Getenv("LESS_DEBUG_GUARDS") == "1"

	allArgsCnt := 0
	if args != nil {
		allArgsCnt = len(args)
	}

	if debug {
		fmt.Printf("DEBUG: MatchArgs for mixin '%s' with %d args (arity=%d, required=%d)\n", md.Name, allArgsCnt, md.Arity, md.Required)
	}

	// Count required arguments that are provided (match JavaScript logic)
	// JavaScript: args.reduce(function (count, p) {
	//   if (optionalParameters.indexOf(p.name) < 0) {
	//     return count + 1;
	//   } else {
	//     return count;
	//   }
	// }, 0);
	requiredArgsCnt := 0
	for _, arg := range args {
		if argMap, ok := arg.(map[string]any); ok {
			if argName, ok := argMap["name"].(string); ok && argName != "" {
				// Check if this named argument is in optional parameters list
				isOptional := false
				for _, optParam := range md.OptionalParameters {
					if optParam == argName {
						isOptional = true
						break
					}
				}
				// Only count if NOT in optional parameters (like JavaScript)
				if !isOptional {
					requiredArgsCnt++
				}
				// Note: JavaScript only counts named args, so we don't increment for positional args here
			} else {
				// Positional argument (no name) - always counts as required
				requiredArgsCnt++
			}
		} else {
			// Non-map argument - always counts as required
			requiredArgsCnt++
		}
	}

	if !md.Variadic {
		if requiredArgsCnt < md.Required {
			return false
		}
		if allArgsCnt > md.Arity {
			return false
		}
	} else {
		if requiredArgsCnt < (md.Required - 1) {
			return false
		}
	}

	// Check patterns - match JavaScript implementation
	lenCheck := requiredArgsCnt
	if lenCheck > md.Arity {
		lenCheck = md.Arity
	}

	for i := 0; i < lenCheck; i++ {
		if i >= len(md.Params) {
			continue
		}

		param, ok := md.Params[i].(map[string]any)
		if !ok {
			continue
		}

		// Check if this parameter has a name or is variadic
		paramName, hasName := param["name"].(string)
		paramVariadic, isVariadic := param["variadic"].(bool)

		if debug {
			paramValue := param["value"]
			fmt.Printf("DEBUG:   Param[%d]: hasName=%v, name='%s', variadic=%v, value=%T\n", i, hasName, paramName, paramVariadic, paramValue)
		}

		// If the parameter has an empty name and is not variadic, it's a pattern that needs matching
		// Note: hasName can be true even when paramName is empty string
		if paramName == "" && (!isVariadic || !paramVariadic) {
			// Get the parameter value
			paramValue := param["value"]
			if paramValue == nil {
				if debug {
					fmt.Printf("DEBUG:   Pattern param has nil value, skipping\n")
				}
				continue
			}

			// Get the argument at this position
			if i >= len(args) {
				if debug {
					fmt.Printf("DEBUG:   Pattern matching failed: not enough args\n")
				}
				return false
			}

			argValue := args[i]
			if argMap, ok := argValue.(map[string]any); ok {
				if val, hasVal := argMap["value"]; hasVal {
					argValue = val
				}
			}

			// Evaluate both values and compare their CSS output
			// Match JavaScript: args[i].value.eval(context).toCSS() != this.params[i].value.eval(context).toCSS()
			var paramCSS string
			if evaluator, ok := paramValue.(interface{ Eval(any) (any, error) }); ok {
				evalResult, err := evaluator.Eval(context)
				if err != nil {
					if debug {
						fmt.Printf("DEBUG:   Pattern matching failed: param eval error: %v\n", err)
					}
					return false
				}
				paramCSS = toCSS(evalResult, context)
			} else if cssGenerator, ok := paramValue.(interface{ ToCSS(any) string }); ok {
				paramCSS = cssGenerator.ToCSS(context)
			} else {
				paramCSS = fmt.Sprintf("%v", paramValue)
			}

			// Then evaluate the argument value
			var argCSS string
			if evaluator, ok := argValue.(interface{ Eval(any) (any, error) }); ok {
				evalResult, err := evaluator.Eval(context)
				if err != nil {
					if debug {
						fmt.Printf("DEBUG:   Pattern matching failed: arg eval error: %v\n", err)
					}
					return false
				}
				argCSS = toCSS(evalResult, context)
			} else if cssGenerator, ok := argValue.(interface{ ToCSS(any) string }); ok {
				argCSS = cssGenerator.ToCSS(context)
			} else {
				argCSS = fmt.Sprintf("%v", argValue)
			}

			// Compare the CSS output
			if debug {
				fmt.Printf("DEBUG:   Pattern match: param='%s' vs arg='%s'\n", paramCSS, argCSS)
			}
			if paramCSS != argCSS {
				if debug {
					fmt.Printf("DEBUG:   Pattern matching failed: CSS mismatch\n")
				}
				return false
			}
			if debug {
				fmt.Printf("DEBUG:   Pattern matched!\n")
			}
		}
	}

	if debug {
		fmt.Printf("DEBUG:   MatchArgs returning true\n")
	}
	return true
}

// GenCSS for MixinDefinition - mixin definitions should not output any CSS
func (md *MixinDefinition) GenCSS(context any, output *CSSOutput) {
	// Mixin definitions do not generate CSS output
}

// ensureMixinContentVisibility recursively ensures visibility for mixin expansion content.
// Unlike SetTreeVisibilityVisitor, this function clears visibility blocks first,
// then ensures visibility. This is necessary because mixin content from reference imports
// has visibility blocks that would cause SetTreeVisibilityVisitor to skip the nodes.
func ensureMixinContentVisibility(node any) {
	if node == nil {
		return
	}

	// Debug: check visibility before changes
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" {
		if rs, ok := node.(*Ruleset); ok && len(rs.Selectors) > 0 {
			if sel, ok := rs.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
				elemVal := sel.Elements[0].Value
				if str, ok := elemVal.(string); ok && str == "div" {
					fmt.Fprintf(os.Stderr, "[ensureMixinContentVisibility] Found div ruleset=%p, Node=%p\n", rs, rs.Node)
					if rs.Node != nil {
						fmt.Fprintf(os.Stderr, "[ensureMixinContentVisibility] BEFORE: BlocksVisibility=%v, VisibilityBlocks=%v\n",
							rs.Node.BlocksVisibility(), rs.Node.VisibilityBlocks)
					}
				}
			}
		}
	}

	// Clear visibility blocks first (so the node no longer blocks visibility)
	if clearNode, ok := node.(interface{ ClearVisibilityBlocks() }); ok {
		clearNode.ClearVisibilityBlocks()
	}

	// Then ensure visibility
	if visNode, ok := node.(interface{ EnsureVisibility() }); ok {
		visNode.EnsureVisibility()
	}

	// Debug: check visibility after changes
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" {
		if rs, ok := node.(*Ruleset); ok && len(rs.Selectors) > 0 {
			if sel, ok := rs.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
				elemVal := sel.Elements[0].Value
				if str, ok := elemVal.(string); ok && str == "div" {
					if rs.Node != nil {
						fmt.Fprintf(os.Stderr, "[ensureMixinContentVisibility] AFTER: BlocksVisibility=%v, VisibilityBlocks=%v\n",
							rs.Node.BlocksVisibility(), rs.Node.VisibilityBlocks)
					}
				}
			}
		}
	}

	// Recursively process children based on node type
	switch n := node.(type) {
	case *Ruleset:
		// Also ensure visibility on selectors - this is important for path filtering in compileRulesetPaths
		for _, sel := range n.Selectors {
			if selector, ok := sel.(*Selector); ok {
				if selector.Node != nil {
					selector.Node.ClearVisibilityBlocks()
					selector.Node.EnsureVisibility()
				}
			}
		}
		for _, rule := range n.Rules {
			ensureMixinContentVisibility(rule)
		}
	case *Media:
		for _, rule := range n.Rules {
			ensureMixinContentVisibility(rule)
		}
	case *AtRule:
		for _, rule := range n.Rules {
			ensureMixinContentVisibility(rule)
		}
	case []any:
		for _, item := range n {
			ensureMixinContentVisibility(item)
		}
	}
} 