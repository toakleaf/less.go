package less_go

import (
	"fmt"
	"os"
)

func flattenPath(path []any) []any {
	result := make([]any, 0, len(path))
	for _, item := range path {
		if arr, ok := item.([]any); ok {
			result = append(result, arr...)
		} else {
			result = append(result, item)
		}
	}
	return result
}

type contextInfo struct {
	paths      []any
	multiMedia bool
}

type JoinSelectorVisitor struct {
	contexts []*contextInfo
	visitor  *Visitor
}

func NewJoinSelectorVisitor() *JoinSelectorVisitor {
	jsv := &JoinSelectorVisitor{
		contexts: []*contextInfo{{paths: []any{}, multiMedia: false}},
	}
	jsv.visitor = NewVisitor(jsv)
	return jsv
}

// Reset resets the JoinSelectorVisitor for reuse from the pool.
// The visitor's methodLookup map is preserved (it's expensive to rebuild).
func (jsv *JoinSelectorVisitor) Reset() {
	// Clear contexts slice but keep capacity
	for i := range jsv.contexts {
		jsv.contexts[i] = nil
	}
	jsv.contexts = jsv.contexts[:0]
	// Initialize with the default context
	jsv.contexts = append(jsv.contexts, &contextInfo{paths: []any{}, multiMedia: false})
	// Note: jsv.visitor is preserved - its methodLookup is reused
}

func (jsv *JoinSelectorVisitor) Run(root any) any {
	return jsv.visitor.Visit(root)
}

func (jsv *JoinSelectorVisitor) IsReplacing() bool {
	return false
}

func (jsv *JoinSelectorVisitor) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {
	switch n := node.(type) {
	case *Declaration:
		return jsv.VisitDeclaration(n, visitArgs), true
	case *MixinDefinition:
		return jsv.VisitMixinDefinition(n, visitArgs), true
	case *Ruleset:
		return jsv.VisitRuleset(n, visitArgs), true
	case *Media:
		return jsv.VisitMedia(n, visitArgs), true
	case *Container:
		return jsv.VisitContainer(n, visitArgs), true
	case *AtRule:
		return jsv.VisitAtRule(n, visitArgs), true
	default:
		return node, false
	}
}

func (jsv *JoinSelectorVisitor) VisitNodeOut(node any) bool {
	switch n := node.(type) {
	case *Ruleset:
		jsv.VisitRulesetOut(n)
		return true
	default:
		return false
	}
}

func (jsv *JoinSelectorVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) any {
	visitArgs.VisitDeeper = false
	return declNode
}

func (jsv *JoinSelectorVisitor) VisitMixinDefinition(mixinDefinitionNode any, visitArgs *VisitArgs) any {
	visitArgs.VisitDeeper = false
	return mixinDefinitionNode
}

func (jsv *JoinSelectorVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) any {
	isDivRuleset := false
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" {
		if rs, ok := rulesetNode.(*Ruleset); ok && len(rs.Selectors) > 0 {
			if sel, ok := rs.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
				elemVal := sel.Elements[0].Value
				if str, ok := elemVal.(string); ok && str == "div" {
					isDivRuleset = true
					fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRuleset] div ruleset=%p, Root=%v, Selectors=%d, EvaldCondition=%v\n",
						rs, rs.Root, len(rs.Selectors), sel.EvaldCondition)
				}
			}
		}
	}
	_ = isDivRuleset

	contextItem := jsv.contexts[len(jsv.contexts)-1]
	context := contextItem.paths
	paths := make([]any, 0)

	isMultiMedia := false
	if rs, ok := rulesetNode.(*Ruleset); ok {
		isMultiMedia = rs.MultiMedia
	}

	// Debug tracing for directives-bubling analysis
	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRuleset] START: context.paths len=%d, isMultiMedia=%v\n",
			len(context), isMultiMedia)
		if rs, ok := rulesetNode.(*Ruleset); ok {
			fmt.Fprintf(os.Stderr, "  Ruleset: Root=%v, Selectors len=%d\n", rs.Root, len(rs.Selectors))
			for i, sel := range rs.Selectors {
				if s, ok := sel.(*Selector); ok {
					fmt.Fprintf(os.Stderr, "    Selector[%d]: Elements=%d, EvaldCondition=%v\n", i, len(s.Elements), s.EvaldCondition)
					for j, el := range s.Elements {
						fmt.Fprintf(os.Stderr, "      Element[%d]: Value=%v\n", j, el.Value)
					}
				}
			}
		}
	}

	// Push context BEFORE JoinSelectors (matches JavaScript)
	jsv.contexts = append(jsv.contexts, &contextInfo{paths: paths, multiMedia: isMultiMedia})

	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		fmt.Fprintf(os.Stderr, "  PUSHED context, stack depth now %d\n", len(jsv.contexts))
	}

	if rulesetInterface, ok := rulesetNode.(interface {
		GetRoot() bool
		GetSelectors() []any
		SetSelectors([]any)
		SetRules([]any)
		SetPaths([]any)
	}); ok {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			if rs, ok := rulesetNode.(*Ruleset); ok {
				fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRuleset] ruleset=%p Root=%v\n", rs, rs.Root)
			}
		}
		if !rulesetInterface.GetRoot() {
			selectors := rulesetInterface.GetSelectors()
			if selectors != nil {
				filteredSelectors := make([]any, 0)
				for _, selector := range selectors {
					if selectorWithOutput, ok := selector.(interface{ GetIsOutput() bool }); ok {
						isOutput := selectorWithOutput.GetIsOutput()
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRuleset] Selector %T GetIsOutput=%v\n", selector, isOutput)
						}
						if isOutput {
							filteredSelectors = append(filteredSelectors, selector)
						}
					}
				}
				
				if len(filteredSelectors) > 0 {
					rulesetInterface.SetSelectors(filteredSelectors)
					if jsInterface, ok := rulesetNode.(interface{ JoinSelectors(*[][]any, [][]any, []any) }); ok {
						pathsSlice := make([][]any, 0)

						var contextSlice [][]any
						if len(context) == 0 {
							contextSlice = [][]any{}
						} else {
							contextSlice = make([][]any, len(context))
							for i, path := range context {
								if pathArray, ok := path.([]any); ok {
									contextSlice[i] = pathArray
								} else {
									contextSlice[i] = []any{path}
								}
							}
						}

						jsInterface.JoinSelectors(&pathsSlice, contextSlice, filteredSelectors)

						for _, path := range pathsSlice {
							flatPath := flattenPath(path)
							paths = append(paths, flatPath)
						}

						jsv.contexts[len(jsv.contexts)-1].paths = paths

						if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
							fmt.Fprintf(os.Stderr, "  After JoinSelectors: paths len=%d\n", len(paths))
							for i, p := range paths {
								fmt.Fprintf(os.Stderr, "    Path[%d]: %v\n", i, p)
							}
						}
					}
				} else {
					rulesetInterface.SetSelectors(nil)
				}
			}
			
			if len(rulesetInterface.GetSelectors()) == 0 {
				rulesetInterface.SetRules(nil)
			}
			rulesetInterface.SetPaths(paths)
			if os.Getenv("LESS_GO_DEBUG_VIS") == "1" && isDivRuleset {
				fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRuleset] Setting paths=%d for div ruleset\n", len(paths))
			}
		}
	} else if ruleset, ok := rulesetNode.(*Ruleset); ok {
		if !ruleset.Root {
			selectors := ruleset.Selectors
			if selectors != nil {
				filteredSelectors := make([]any, 0)
				for _, selector := range selectors {
					if selectorWithOutput, ok := selector.(interface{ GetIsOutput() bool }); ok {
						if selectorWithOutput.GetIsOutput() {
							filteredSelectors = append(filteredSelectors, selector)
						}
					}
				}
				
				if len(filteredSelectors) > 0 {
					ruleset.Selectors = filteredSelectors
					if hasJoinSelectors(ruleset) {
						pathsSlice := make([][]any, 0)
						pathsPtr := &pathsSlice

						var contextSlice [][]any
						if len(context) == 0 {
							contextSlice = [][]any{}
						} else {
							contextSlice = make([][]any, len(context))
							for i, path := range context {
								if pathArray, ok := path.([]any); ok {
									contextSlice[i] = pathArray
								} else {
									contextSlice[i] = []any{path}
								}
							}
						}

						ruleset.JoinSelectors(pathsPtr, contextSlice, filteredSelectors)

						for _, path := range *pathsPtr {
							flatPath := flattenPath(path)
							paths = append(paths, flatPath)
						}

						jsv.contexts[len(jsv.contexts)-1].paths = paths
						ruleset.SetPaths(paths)
					}
				} else {
					ruleset.Selectors = nil
				}
			}
			
			if ruleset.Selectors == nil {
				ruleset.Rules = nil
			}
		}
	}
	
	return rulesetNode
}

func (jsv *JoinSelectorVisitor) VisitRulesetOut(rulesetNode any) {
	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRulesetOut] POPPING context, stack depth was %d\n", len(jsv.contexts))
	}
	if len(jsv.contexts) > 0 {
		jsv.contexts = jsv.contexts[:len(jsv.contexts)-1]
	}
}

type MediaRule interface {
	SetRoot(root bool)
}

func (jsv *JoinSelectorVisitor) VisitMedia(mediaNode any, visitArgs *VisitArgs) any {
	if len(jsv.contexts) == 0 {
		return nil
	}
	contextItem := jsv.contexts[len(jsv.contexts)-1]

	// Root is true when context paths is empty (MultiMedia case results in empty paths)
	rootValue := len(contextItem.paths) == 0

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitMedia] contextItem.paths len=%d, multiMedia=%v, rootValue=%v\n",
			len(contextItem.paths), contextItem.multiMedia, rootValue)
	}

	if mediaInterface, ok := mediaNode.(interface{ GetRules() []any }); ok {
		rules := mediaInterface.GetRules()
		if len(rules) > 0 {
			if mediaRule, ok := rules[0].(interface{ SetRoot(bool) }); ok {
				mediaRule.SetRoot(rootValue)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitMedia] Set root=%v on inner ruleset (interface)\n", rootValue)
				}
			}
		}
	} else if media, ok := mediaNode.(*Media); ok {
		rules := media.Rules
		if len(rules) > 0 {
			if mediaRule, ok := rules[0].(MediaRule); ok {
				mediaRule.SetRoot(rootValue)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitMedia] Set root=%v on inner ruleset (concrete)\n", rootValue)
				}
			}
		}
	}

	return mediaNode
}

func (jsv *JoinSelectorVisitor) VisitContainer(containerNode any, visitArgs *VisitArgs) any {
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitContainer] Called\n")
	}

	// Guard against empty contexts
	if len(jsv.contexts) == 0 {
		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitContainer] Empty contexts, returning nil\n")
		}
		return nil
	}
	contextItem := jsv.contexts[len(jsv.contexts)-1]

	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitContainer] Context paths length: %d\n", len(contextItem.paths))
	}

	rootValue := len(contextItem.paths) == 0 || contextItem.multiMedia

	if containerInterface, ok := containerNode.(interface{ GetRules() []any }); ok {
		rules := containerInterface.GetRules()
		if len(rules) > 0 {
			if containerRule, ok := rules[0].(interface{ SetRoot(bool) }); ok {
				containerRule.SetRoot(rootValue)
			}
		}
	} else if container, ok := containerNode.(*Container); ok {
		rules := container.Rules
		if len(rules) > 0 {
			if containerRule, ok := rules[0].(MediaRule); ok {
				containerRule.SetRoot(rootValue)
			}
		}
	}

	return containerNode
}

type AtRuleRule interface {
	SetRoot(value any)
}

// Matches JavaScript join-selector-visitor.js visitAtRule:
// atRuleNode.rules[0].root = (atRuleNode.isRooted || context.length === 0 || null);
func (jsv *JoinSelectorVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) any {
	if len(jsv.contexts) == 0 {
		return nil
	}
	contextItem := jsv.contexts[len(jsv.contexts)-1]
	contextPaths := contextItem.paths

	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		if atRule, ok := atRuleNode.(*AtRule); ok {
			fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitAtRule] AtRule name=%q, isRooted=%v, context.paths len=%d\n",
				atRule.Name, atRule.IsRooted, len(contextPaths))
		}
	}

	if atRuleInterface, ok := atRuleNode.(interface{ GetRules() []any }); ok {
		rules := atRuleInterface.GetRules()
		if rules != nil && len(rules) > 0 {
			if atRuleRule, ok := rules[0].(interface{ SetRoot(any) }); ok {
				var rootValue any = nil

				var isBubblingDirective bool
				if nameInterface, ok := atRuleNode.(interface{ GetName() string }); ok {
					name := nameInterface.GetName()
					isBubblingDirective = (name == "@supports" || name == "@document")
				}

				if isRootedInterface, ok := atRuleNode.(interface{ GetIsRooted() bool }); ok {
					isRooted := isRootedInterface.GetIsRooted()
					if isRooted {
						rootValue = true
					} else if isBubblingDirective {
						rootValue = false
					} else {
						if len(contextPaths) == 0 {
							rootValue = true
						}
					}
				} else if len(contextPaths) == 0 {
					rootValue = true
				}
				atRuleRule.SetRoot(rootValue)
			}
		}
	} else if atRule, ok := atRuleNode.(*AtRule); ok {
		rules := atRule.Rules
		if rules != nil && len(rules) > 0 {
			if atRuleRule, ok := rules[0].(AtRuleRule); ok {
				var rootValue any = nil
				isBubblingDirective := (atRule.Name == "@supports" || atRule.Name == "@document")
				if atRule.IsRooted {
					rootValue = true
				} else if isBubblingDirective {
					rootValue = false
				} else if len(contextPaths) == 0 {
					rootValue = true
				}
				atRuleRule.SetRoot(rootValue)
			}
		}
	}

	return atRuleNode
}

func hasJoinSelectors(ruleset *Ruleset) bool {
	return true
}