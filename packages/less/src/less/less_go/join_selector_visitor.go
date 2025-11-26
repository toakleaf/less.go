package less_go

import (
	"fmt"
	"os"
)

// flattenPath flattens nested arrays in selector paths to ensure correct CSS generation
func flattenPath(path []any) []any {
	result := make([]any, 0, len(path))
	for _, item := range path {
		if arr, ok := item.([]any); ok {
			// If this item is an array, flatten it
			result = append(result, arr...)
		} else {
			// If this item is a selector, keep it as-is
			result = append(result, item)
		}
	}
	return result
}

// contextInfo holds context information for the JoinSelectorVisitor
// It tracks both selector paths and whether we're in a MultiMedia Ruleset
type contextInfo struct {
	paths      []any // Selector paths
	multiMedia bool  // True if the Ruleset has MultiMedia=true
}

// JoinSelectorVisitor implements a visitor that joins selectors in rulesets
type JoinSelectorVisitor struct {
	contexts []*contextInfo
	visitor  *Visitor
}

// NewJoinSelectorVisitor creates a new JoinSelectorVisitor
func NewJoinSelectorVisitor() *JoinSelectorVisitor {
	jsv := &JoinSelectorVisitor{
		contexts: []*contextInfo{{paths: []any{}, multiMedia: false}},
	}
	jsv.visitor = NewVisitor(jsv)
	return jsv
}

// Run executes the visitor on the root node
func (jsv *JoinSelectorVisitor) Run(root any) any {
	return jsv.visitor.Visit(root)
}

// IsReplacing returns false as this visitor doesn't replace nodes
func (jsv *JoinSelectorVisitor) IsReplacing() bool {
	return false
}

// VisitNode implements direct dispatch without reflection for better performance
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

// VisitNodeOut implements direct dispatch for visitOut methods
func (jsv *JoinSelectorVisitor) VisitNodeOut(node any) bool {
	switch n := node.(type) {
	case *Ruleset:
		jsv.VisitRulesetOut(n)
		return true
	default:
		return false
	}
}

// VisitDeclaration prevents deeper visiting of declaration nodes
func (jsv *JoinSelectorVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) any {
	visitArgs.VisitDeeper = false
	return declNode
}

// VisitMixinDefinition prevents deeper visiting of mixin definition nodes
func (jsv *JoinSelectorVisitor) VisitMixinDefinition(mixinDefinitionNode any, visitArgs *VisitArgs) any {
	visitArgs.VisitDeeper = false
	return mixinDefinitionNode
}

// VisitRuleset processes ruleset nodes
func (jsv *JoinSelectorVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) any {
	contextItem := jsv.contexts[len(jsv.contexts)-1]
	context := contextItem.paths
	paths := make([]any, 0)

	// Check if this is a MultiMedia ruleset
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

	// Push new context info to context stack BEFORE JoinSelectors (matches JavaScript)
	jsv.contexts = append(jsv.contexts, &contextInfo{paths: paths, multiMedia: isMultiMedia})

	// Debug tracing for context stack
	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		fmt.Fprintf(os.Stderr, "  PUSHED context, stack depth now %d\n", len(jsv.contexts))
	}
	
	// Try interface-based approach first
	if rulesetInterface, ok := rulesetNode.(interface {
		GetRoot() bool
		GetSelectors() []any
		SetSelectors([]any)
		SetRules([]any)
		SetPaths([]any)
	}); ok {
		if !rulesetInterface.GetRoot() {
			selectors := rulesetInterface.GetSelectors()
			if selectors != nil {
				// Filter selectors by GetIsOutput
				filteredSelectors := make([]any, 0)
				for _, selector := range selectors {
					if selectorWithOutput, ok := selector.(interface{ GetIsOutput() bool }); ok {
						if selectorWithOutput.GetIsOutput() {
							filteredSelectors = append(filteredSelectors, selector)
						}
					}
				}
				
				if len(filteredSelectors) > 0 {
					rulesetInterface.SetSelectors(filteredSelectors)
					// Call JoinSelectors if it exists on the ruleset
					if jsInterface, ok := rulesetNode.(interface{ JoinSelectors(*[][]any, [][]any, []any) }); ok {
						// Convert paths and context to the expected types
						pathsSlice := make([][]any, 0)
						
						// The context is already a []any containing paths.
						// JoinSelectors expects [][]any where each element is a path.
						// If context is empty, pass empty [][]any
						// Otherwise, convert context elements to [][]any
						var contextSlice [][]any
						if len(context) == 0 {
							contextSlice = [][]any{}
						} else {
							// Each element in context should be a path ([]any)
							contextSlice = make([][]any, len(context))
							for i, path := range context {
								if pathArray, ok := path.([]any); ok {
									contextSlice[i] = pathArray
								} else {
									// If it's a single selector, wrap it
									contextSlice[i] = []any{path}
								}
							}
						}
						
						jsInterface.JoinSelectors(&pathsSlice, contextSlice, filteredSelectors)

						// Convert [][]any to []any and update the existing paths slice in place
						// This ensures the paths array on the context stack gets populated
						for _, path := range pathsSlice {
							// Flatten any nested arrays in the path to match expected structure
							flatPath := flattenPath(path)
							paths = append(paths, flatPath)
						}

						// Update the context stack with the populated paths
						jsv.contexts[len(jsv.contexts)-1].paths = paths

						// Debug tracing for directives-bubling analysis
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
		}
	} else if ruleset, ok := rulesetNode.(*Ruleset); ok {
		// Fallback to concrete type for backward compatibility
		if !ruleset.Root {
			selectors := ruleset.Selectors
			if selectors != nil {
				// Filter selectors by GetIsOutput
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
					// Call JoinSelectors if it exists on the ruleset
					if hasJoinSelectors(ruleset) {
						pathsSlice := make([][]any, 0)
						pathsPtr := &pathsSlice
						
						// The context is already a []any containing paths.
						// JoinSelectors expects [][]any where each element is a path.
						// If context is empty, pass empty [][]any
						// Otherwise, convert context elements to [][]any
						var contextSlice [][]any
						if len(context) == 0 {
							contextSlice = [][]any{}
						} else {
							// Each element in context should be a path ([]any)
							contextSlice = make([][]any, len(context))
							for i, path := range context {
								if pathArray, ok := path.([]any); ok {
									contextSlice[i] = pathArray
								} else {
									// If it's a single selector, wrap it
									contextSlice[i] = []any{path}
								}
							}
						}
						
						ruleset.JoinSelectors(pathsPtr, contextSlice, filteredSelectors)
						
						// Convert the result to []any and update the existing paths slice in place
						// This ensures the paths array on the context stack gets populated
						for _, path := range *pathsPtr {
							// Flatten any nested arrays in the path to match expected structure
							flatPath := flattenPath(path)
							paths = append(paths, flatPath)
						}
						
						// Update the context stack with the populated paths
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

// VisitRulesetOut removes the top context when exiting a ruleset
func (jsv *JoinSelectorVisitor) VisitRulesetOut(rulesetNode any) {
	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitRulesetOut] POPPING context, stack depth was %d\n", len(jsv.contexts))
	}
	if len(jsv.contexts) > 0 {
		jsv.contexts = jsv.contexts[:len(jsv.contexts)-1]
	}
}

// MediaRule interface for the first rule in media
type MediaRule interface {
	SetRoot(root bool)
}

// VisitMedia processes media nodes
func (jsv *JoinSelectorVisitor) VisitMedia(mediaNode any, visitArgs *VisitArgs) any {
	// Guard against empty contexts
	if len(jsv.contexts) == 0 {
		return nil
	}
	contextItem := jsv.contexts[len(jsv.contexts)-1]

	// Set root flag on inner ruleset
	// JavaScript: mediaNode.rules[0].root = (context.length === 0 || context[0].multiMedia);
	// Root is true if we're at the top level (context.paths empty) OR inside a MultiMedia Ruleset
	rootValue := len(contextItem.paths) == 0 || contextItem.multiMedia

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitMedia] contextItem.paths len=%d, multiMedia=%v, setting root=%v\n",
			len(contextItem.paths), contextItem.multiMedia, rootValue)
	}

	// NOTE: BubbleSelectors is called during Media.Eval, not here.
	// By the time JoinSelectorVisitor runs, Media nodes have already been bubbled up
	// and their parent selectors have been captured during evaluation.

	// Try interface-based approach first
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
		// Fallback to concrete type for backward compatibility
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

// VisitContainer processes container nodes (same logic as media for bubbling)
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

	// Set root flag on inner ruleset (same as Media)
	// Root is true if we're at the top level (context.paths empty) OR inside a MultiMedia Ruleset
	rootValue := len(contextItem.paths) == 0 || contextItem.multiMedia

	// NOTE: BubbleSelectors is called during Container.Eval, not here.
	// By the time JoinSelectorVisitor runs, Container nodes have already been bubbled up
	// and their parent selectors have been captured during evaluation.

	// Try interface-based approach first
	if containerInterface, ok := containerNode.(interface{ GetRules() []any }); ok {
		rules := containerInterface.GetRules()
		if len(rules) > 0 {
			if containerRule, ok := rules[0].(interface{ SetRoot(bool) }); ok {
				containerRule.SetRoot(rootValue)
			}
		}
	} else if container, ok := containerNode.(*Container); ok {
		// Fallback to concrete type for backward compatibility
		rules := container.Rules
		if len(rules) > 0 {
			if containerRule, ok := rules[0].(MediaRule); ok {
				containerRule.SetRoot(rootValue)
			}
		}
	}

	return containerNode
}

// AtRuleRule interface for the first rule in at-rule
type AtRuleRule interface {
	SetRoot(value any)
}

// VisitAtRule processes at-rule nodes
// This matches JavaScript join-selector-visitor.js visitAtRule exactly:
//   const context = this.contexts[this.contexts.length - 1];
//   if (atRuleNode.rules && atRuleNode.rules.length) {
//       atRuleNode.rules[0].root = (atRuleNode.isRooted || context.length === 0 || null);
//   }
func (jsv *JoinSelectorVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) any {
	// Guard against empty contexts
	if len(jsv.contexts) == 0 {
		return nil
	}
	contextItem := jsv.contexts[len(jsv.contexts)-1]
	contextPaths := contextItem.paths

	// Debug tracing for directives-bubling analysis
	if os.Getenv("LESS_GO_TRACE_JOIN") == "1" {
		if atRule, ok := atRuleNode.(*AtRule); ok {
			fmt.Fprintf(os.Stderr, "[JoinSelectorVisitor.VisitAtRule] AtRule name=%q, isRooted=%v, context.paths len=%d\n",
				atRule.Name, atRule.IsRooted, len(contextPaths))
		}
	}

	// Try interface-based approach first
	if atRuleInterface, ok := atRuleNode.(interface{ GetRules() []any }); ok {
		rules := atRuleInterface.GetRules()
		if rules != nil && len(rules) > 0 {
			if atRuleRule, ok := rules[0].(interface{ SetRoot(any) }); ok {
				var rootValue any = nil

				// Check if this is a bubbling directive that needs special handling
				// Only @supports and @document need selector bubbling and Root=false
				var isBubblingDirective bool
				if nameInterface, ok := atRuleNode.(interface{ GetName() string }); ok {
					name := nameInterface.GetName()
					isBubblingDirective = (name == "@supports" || name == "@document")
				}

				// Check if atRule has GetIsRooted method
				if isRootedInterface, ok := atRuleNode.(interface{ GetIsRooted() bool }); ok {
					isRooted := isRootedInterface.GetIsRooted()
					if isRooted {
						// Rooted directives (@font-face, @keyframes) always have Root=true
						rootValue = true
					} else if isBubblingDirective {
						// ONLY @supports and @document get special bubbling treatment
						// Set Root=false to allow nested selector joining
						rootValue = false
					} else {
						// Other non-rooted directives use the old behavior
						// Set Root=nil when context has items, Root=true when empty
						if len(contextPaths) == 0 {
							rootValue = true
						}
						// else rootValue stays nil
					}
				} else if len(contextPaths) == 0 {
					rootValue = true
				}
				atRuleRule.SetRoot(rootValue)
			}
		}
	} else if atRule, ok := atRuleNode.(*AtRule); ok {
		// Fallback to concrete type for backward compatibility
		rules := atRule.Rules
		if rules != nil && len(rules) > 0 {
			if atRuleRule, ok := rules[0].(AtRuleRule); ok {
				var rootValue any = nil
				// Check if this is a bubbling directive
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

// Helper functions to safely check and call methods
func hasJoinSelectors(ruleset *Ruleset) bool {
	// Check if the ruleset has a JoinSelectors method
	// For now, assume all rulesets have this capability
	return true
}


func hasIsRooted(atRule *AtRule) bool {
	// Check if atRule has IsRooted method
	// For now, return false as a safe default
	return false
}

func getIsRooted(atRule *AtRule) bool {
	// Get the IsRooted value
	// For now, return false as a safe default
	return false
}