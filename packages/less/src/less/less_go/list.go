package less_go

import (
	"fmt"
	"strconv"
)

func GetItemsFromNode(node any) []any {
	if node == nil {
		return []any{node}
	}

	if nodeMap, ok := node.(map[string]any); ok {
		if value, exists := nodeMap["value"]; exists {
			if valueSlice, ok := value.([]any); ok {
				return valueSlice
			}
		}
	}

	if valueNode, ok := node.(*Value); ok && valueNode != nil {
		return valueNode.Value
	}

	if exprNode, ok := node.(*Expression); ok && exprNode != nil {
		return exprNode.Value
	}

	if valueSlice, ok := node.([]any); ok {
		return valueSlice
	}

	return []any{node}
}

func Self(n any) any {
	return n
}

func SpaceSeparatedValues(expr ...any) any {
	if len(expr) == 1 {
		return expr[0]
	}
	value, _ := NewValue(expr)
	return value
}

func Extract(values any, indexNode any) any {
	items := GetItemsFromNode(values)

	var index int
	var indexFloat float64
	if dimNode, ok := indexNode.(*Dimension); ok {
		indexFloat = dimNode.Value - 1
		index = int(indexFloat)
		if indexFloat != float64(index) {
			return nil
		}
	} else if indexMap, ok := indexNode.(map[string]any); ok {
		if val, exists := indexMap["value"]; exists {
			if floatVal, ok := val.(float64); ok {
				index = int(floatVal) - 1
			} else if intVal, ok := val.(int); ok {
				index = intVal - 1
			} else if strVal, ok := val.(string); ok {
				if parsedVal, err := strconv.ParseFloat(strVal, 64); err == nil {
					index = int(parsedVal) - 1
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			return nil
		}
	} else {
		return nil
	}

	if index < 0 || index >= len(items) {
		return nil
	}

	return items[index]
}

func Length(values any) *Dimension {
	items := GetItemsFromNode(values)
	dim, _ := NewDimension(float64(len(items)), nil)
	return dim
}

func Range(start, end, step any) *Expression {
	var from float64 = 1
	var to float64
	var stepValue float64 = 1
	var unit *Unit

	if end != nil {
		if startDim, ok := start.(*Dimension); ok {
			from = startDim.Value
		}
		if endDim, ok := end.(*Dimension); ok {
			to = endDim.Value
			unit = endDim.Unit
		}
		if step != nil {
			if stepDim, ok := step.(*Dimension); ok {
				stepValue = stepDim.Value
			}
		}
	} else {
		if startDim, ok := start.(*Dimension); ok {
			to = startDim.Value
			unit = startDim.Unit
		}
	}

	var list []any
	for i := from; i <= to; i += stepValue {
		dim, _ := NewDimension(i, unit)
		list = append(list, dim)
	}

	expr, err := NewExpression(list, false)
	if err != nil {
		expr, _ = NewExpression([]any{}, false)
	}
	return expr
}

type EachFunctionDef struct {
	name string
}

func (e *EachFunctionDef) Call(args ...any) (any, error) {
	return nil, fmt.Errorf("each() function requires context")
}

func (e *EachFunctionDef) CallCtx(ctx *Context, args ...any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("each() requires exactly 2 arguments")
	}
	return EachWithContext(args[0], args[1], ctx), nil
}

func (e *EachFunctionDef) NeedsEvalArgs() bool {
	return false
}

func Each(list any, rs any) any {
	return EachWithContext(list, rs, nil)
}

func EachWithContext(list any, rs any, ctx *Context) any {
	rules := make([]any, 0)
	var iterator []any
	var sourceRuleset *Ruleset

	var evalContext any
	if ctx != nil {
		if len(ctx.Frames) > 0 {
			rawContext := ctx.Frames[0].EvalContext
			if mapCtx, ok := rawContext.(*MapEvalContext); ok {
				evalContext = mapCtx.ctx
			} else {
				evalContext = rawContext
			}
		} else {
			evalContext = nil
		}
	} else {
		evalContext = nil
	}

	tryEval := func(val any) any {
		if mixinCall, ok := val.(*MixinCall); ok {
			if evalContext != nil {
				rules, err := mixinCall.Eval(evalContext)
				if err == nil && rules != nil {
					ampersandElement := NewElement(nil, "&", false, 0, nil, nil)
					ampersandSelector, sErr := NewSelector([]*Element{ampersandElement}, nil, nil, 0, nil, nil)
					if sErr == nil {
						return NewRuleset([]any{ampersandSelector}, rules, false, nil)
					}
				}
			}
			return mixinCall
		}

		if evalable, ok := val.(interface{ Eval(any) (any, error) }); ok {
			if evalContext != nil {
				result, err := evalable.Eval(evalContext)
				if err == nil {
					return result
				}
			}
			return evalable
		}
		return val
	}

	list = tryEval(list)

	if valueNode, ok := list.(*Value); ok && valueNode.Value != nil {
		if _, isQuoted := list.(*Quoted); !isQuoted {
			iterator = make([]any, len(valueNode.Value))
			for i, item := range valueNode.Value {
				iterator[i] = tryEval(item)
			}
		}
	} else if exprNode, ok := list.(*Expression); ok && exprNode.Value != nil {
		iterator = make([]any, len(exprNode.Value))
		for i, item := range exprNode.Value {
			iterator[i] = tryEval(item)
		}
	} else if detachedRuleset, ok := list.(*DetachedRuleset); ok && detachedRuleset.ruleset != nil {
		if rulesetNode, ok := detachedRuleset.ruleset.(*Ruleset); ok {
			sourceRuleset = rulesetNode
			iterator = rulesetNode.Rules
		} else if node, ok := detachedRuleset.ruleset.(*Node); ok && node.Value != nil {
			if rulesetNode, ok := node.Value.(*Ruleset); ok {
				sourceRuleset = rulesetNode
				iterator = rulesetNode.Rules
			}
		}
	} else if rulesetNode, ok := list.(*Ruleset); ok {
		iterator = make([]any, len(rulesetNode.Rules))
		for i, rule := range rulesetNode.Rules {
			iterator[i] = tryEval(rule)
		}
	} else if listSlice, ok := list.([]any); ok {
		iterator = make([]any, len(listSlice))
		for i, item := range listSlice {
			iterator[i] = tryEval(item)
		}
	} else {
		iterator = []any{tryEval(list)}
	}

	valueName := "@value"
	keyName := "@key"
	indexName := "@index"

	if mixinDef, ok := rs.(*MixinDefinition); ok {
		if len(mixinDef.Params) > 0 {
			if param0, ok := mixinDef.Params[0].(map[string]any); ok {
				if name, ok := param0["name"].(string); ok {
					valueName = name
				}
			}
		}
		if len(mixinDef.Params) > 1 {
			if param1, ok := mixinDef.Params[1].(map[string]any); ok {
				if name, ok := param1["name"].(string); ok {
					keyName = name
				}
			}
		}
		if len(mixinDef.Params) > 2 {
			if param2, ok := mixinDef.Params[2].(map[string]any); ok {
				if name, ok := param2["name"].(string); ok {
					indexName = name
				}
			}
		}
	}

	var targetRuleset *Ruleset
	if mixinDef, ok := rs.(*MixinDefinition); ok {
		targetRuleset = mixinDef.Ruleset
	} else if detachedRuleset, ok := rs.(*DetachedRuleset); ok && detachedRuleset.ruleset != nil {
		if rulesetNode, ok := detachedRuleset.ruleset.(*Ruleset); ok {
			targetRuleset = rulesetNode
		} else if node, ok := detachedRuleset.ruleset.(*Node); ok && node.Value != nil {
			if rulesetNode, ok := node.Value.(*Ruleset); ok {
				targetRuleset = rulesetNode
			}
		}
	} else if rulesetNode, ok := rs.(*Ruleset); ok {
		targetRuleset = rulesetNode
	} else if rsMap, ok := rs.(map[string]any); ok {
		if rulesAny, exists := rsMap["rules"]; exists {
			if rulesSlice, ok := rulesAny.([]any); ok {
					ampersandElement := NewElement(nil, "&", false, 0, nil, nil)
				ampersandSelector, err := NewSelector([]*Element{ampersandElement}, nil, nil, 0, nil, nil)
				if err == nil {
					targetRuleset = NewRuleset([]any{ampersandSelector}, rulesSlice, false, nil)
				}
			}
		}
	}

	if targetRuleset == nil {
		return createEmptyRuleset()
	}

	for i, item := range iterator {
		if _, isComment := item.(*Comment); isComment {
			continue
		}

		var key any
		var value any

		if decl, ok := item.(*Declaration); ok {
			if decl.name != nil {
				if nameStr, ok := decl.name.(string); ok {
					key = nameStr
				} else if nameSlice, ok := decl.name.([]any); ok && len(nameSlice) > 0 {
					first := nameSlice[0]
					if keyword, ok := first.(*Keyword); ok {
						key = keyword.value
					} else if nodeMap, ok := first.(map[string]any); ok {
						if val, exists := nodeMap["value"]; exists {
							key = val
						} else {
							key = first
						}
					} else {
						key = first
					}
				}
			}

			value = decl.Value
		} else {
			keyDim, _ := NewDimension(float64(i+1), nil)
			key = keyDim
			value = item
		}

		newRules := make([]any, len(targetRuleset.Rules), len(targetRuleset.Rules)+3)
		copy(newRules, targetRuleset.Rules)

		if valueName != "" && value != nil {
			valueDecl, err := NewDeclaration(valueName, value, false, false, 0, nil, false, nil)
			if err == nil {
				newRules = append(newRules, valueDecl)
			}
		}

		if indexName != "" {
			indexDim, err := NewDimension(float64(i+1), nil)
			if err == nil {
				indexDecl, err := NewDeclaration(indexName, indexDim, false, false, 0, nil, false, nil)
				if err == nil {
					newRules = append(newRules, indexDecl)
				}
			}
		}

		if keyName != "" && key != nil {
			keyDecl, err := NewDeclaration(keyName, key, false, false, 0, nil, false, nil)
			if err == nil {
				newRules = append(newRules, keyDecl)
			}
		}

		ampersandElement := NewElement(nil, "&", false, 0, nil, nil)
		ampersandSelector, err := NewSelector([]*Element{ampersandElement}, nil, nil, 0, nil, nil)
		if err == nil {
			newRuleset := NewRuleset([]any{ampersandSelector}, newRules, targetRuleset.StrictImports, targetRuleset.VisibilityInfo())
			rules = append(rules, newRuleset)
		}
	}

	ampersandElement := NewElement(nil, "&", false, 0, nil, nil)
	ampersandSelector, err := NewSelector([]*Element{ampersandElement}, nil, nil, 0, nil, nil)
	if err != nil {
		return createEmptyRuleset()
	}

	finalRuleset := NewRuleset([]any{ampersandSelector}, rules, targetRuleset.StrictImports, targetRuleset.VisibilityInfo())

	if evalContext != nil {
		finalEvalContext := evalContext
		if sourceRuleset != nil {
			// Add the source ruleset to the frames for property resolution
			if contextMap, ok := evalContext.(map[string]any); ok {
				newContextMap := make(map[string]any, len(contextMap))
				for k, v := range contextMap {
					newContextMap[k] = v
				}

				// Get existing frames or create empty slice
				var existingFrames []any
				if frames, ok := contextMap["frames"].([]any); ok {
					existingFrames = frames
				} else {
					existingFrames = []any{}
				}

				newFrames := append([]any{sourceRuleset}, existingFrames...)
				newContextMap["frames"] = newFrames
				finalEvalContext = newContextMap
			} else if evalCtx, ok := evalContext.(*Eval); ok {
				newEvalCtx := &Eval{
					Frames:            append([]any{sourceRuleset}, evalCtx.Frames...),
					Compress:          evalCtx.Compress,
					Math:              evalCtx.Math,
					StrictUnits:       evalCtx.StrictUnits,
					Paths:             evalCtx.Paths,
					SourceMap:         evalCtx.SourceMap,
					ImportMultiple:    evalCtx.ImportMultiple,
					UrlArgs:           evalCtx.UrlArgs,
					JavascriptEnabled: evalCtx.JavascriptEnabled,
					PluginManager:     evalCtx.PluginManager,
					ImportantScope:    evalCtx.ImportantScope,
					RewriteUrls:       evalCtx.RewriteUrls,
					CalcStack:         evalCtx.CalcStack,
					ParensStack:       evalCtx.ParensStack,
					InCalc:            evalCtx.InCalc,
					MathOn:            evalCtx.MathOn,
					DefaultFunc:       evalCtx.DefaultFunc,
					PluginBridge:      evalCtx.PluginBridge,
					LazyPluginBridge:  evalCtx.LazyPluginBridge,
				}
				finalEvalContext = newEvalCtx
			}
		}

		result, err := finalRuleset.Eval(finalEvalContext)
		if err != nil {
			return finalRuleset
		}
		return result
	}
	return finalRuleset
}

func createEmptyRuleset() *Ruleset {
	ampersandElement := NewElement(nil, "&", false, 0, nil, nil)
	ampersandSelector, _ := NewSelector([]*Element{ampersandElement}, nil, nil, 0, nil, nil)
	return NewRuleset([]any{ampersandSelector}, []any{}, false, nil)
}

func GetListFunctions() map[string]any {
	return map[string]any{
		"_SELF":   Self,
		"~":       SpaceSeparatedValues, 
		"extract": Extract,
		"length":  Length,
		"range":   Range,
		"each":    Each,
	}
}

func GetWrappedListFunctions() map[string]interface{} {
	return GetListFunctions()
}

func init() {
	listFunctions := GetListFunctions()
	for name, fn := range listFunctions {
		switch name {
		case "_SELF":
			if selfFn, ok := fn.(func(any) any); ok {
				DefaultRegistry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   1,
					maxArgs:   1,
					variadic:  false,
					fn:        selfFn,
					needsEval: true,
				})
			}
		case "~":
			if spaceFn, ok := fn.(func(...any) any); ok {
				DefaultRegistry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   0,
					maxArgs:   -1,
					variadic:  true,
					fn:        spaceFn,
					needsEval: true,
				})
			}
		case "range":
			if rangeFn, ok := fn.(func(any, any, any) any); ok {
				DefaultRegistry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   1,
					maxArgs:   3,
					variadic:  false,
					fn:        rangeFn,
					needsEval: true,
				})
			}
		case "each":
			DefaultRegistry.Add(name, &EachFunctionDef{
				name: name,
			})
		case "length":
			if lengthFn, ok := fn.(func(any) *Dimension); ok {
				DefaultRegistry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   1,
					maxArgs:   1,
					variadic:  false,
					fn:        func(args ...any) any {
						return lengthFn(args[0])
					},
					needsEval: true,
				})
			}
		case "extract":
			if extractFn, ok := fn.(func(any, any) any); ok {
				DefaultRegistry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   2,
					maxArgs:   2,
					variadic:  false,
					fn:        func(args ...any) any {
						return extractFn(args[0], args[1])
					},
					needsEval: true,
				})
			}
		default:
			if functionImpl, ok := fn.(func(any, any) any); ok {
				DefaultRegistry.Add(name, &SimpleFunctionDef{
					name: name,
					fn:   functionImpl,
				})
			}
		}
	}
}