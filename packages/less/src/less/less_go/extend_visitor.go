package less_go

import (
	"fmt"
	"os"
	"strings"
)

// PotentialMatch represents a potential match during extend processing.
// This struct replaces map[string]any to reduce allocations.
type PotentialMatch struct {
	pathIndex            int
	index                int
	matched              int
	initialCombinator    string
	finished             bool
	length               int
	endPathIndex         int
	endPathElementIndex  int
}

type ExtendFinderVisitor struct {
	visitor          *Visitor
	contexts         []any
	allExtendsStack  [][]any
	foundExtends     bool
}

func NewExtendFinderVisitor() *ExtendFinderVisitor {
	efv := &ExtendFinderVisitor{
		contexts:        make([]any, 0),
		allExtendsStack: make([][]any, 1),
	}
	efv.allExtendsStack[0] = make([]any, 0)
	efv.visitor = NewVisitor(efv)
	return efv
}

func (efv *ExtendFinderVisitor) Run(root any) any {
	root = efv.visitor.Visit(root)
	// Convert []any to []*Extend for type consistency
	// Safety check: ensure stack is not empty (visitor Out methods might have popped too many times)
	if len(efv.allExtendsStack) == 0 {
		return root
	}
	extends := make([]*Extend, len(efv.allExtendsStack[0]))
	for i, ext := range efv.allExtendsStack[0] {
		if extend, ok := ext.(*Extend); ok {
			extends[i] = extend
		}
	}
	if rootWithExtends, ok := root.(interface{ SetAllExtends([]*Extend) }); ok {
		rootWithExtends.SetAllExtends(extends)
	}
	return root
}

// IsReplacing returns false as ExtendFinderVisitor is not a replacing visitor
func (efv *ExtendFinderVisitor) IsReplacing() bool {
	return false
}

// VisitNode implements direct dispatch without reflection for better performance
func (efv *ExtendFinderVisitor) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {
	switch n := node.(type) {
	case *Declaration:
		efv.VisitDeclaration(n, visitArgs)
		return n, true
	case *MixinDefinition:
		efv.VisitMixinDefinition(n, visitArgs)
		return n, true
	case *Ruleset:
		efv.VisitRuleset(n, visitArgs)
		return n, true
	case *Media:
		efv.VisitMedia(n, visitArgs)
		return n, true
	case *AtRule:
		efv.VisitAtRule(n, visitArgs)
		return n, true
	default:
		return node, false
	}
}

// VisitNodeOut implements direct dispatch for visitOut methods
func (efv *ExtendFinderVisitor) VisitNodeOut(node any) bool {
	switch n := node.(type) {
	case *Ruleset:
		efv.VisitRulesetOut(n)
		return true
	case *Media:
		efv.VisitMediaOut(n)
		return true
	case *AtRule:
		efv.VisitAtRuleOut(n)
		return true
	default:
		return false
	}
}

func (efv *ExtendFinderVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) {
	visitArgs.VisitDeeper = false
}

func (efv *ExtendFinderVisitor) VisitMixinDefinition(mixinDefinitionNode any, visitArgs *VisitArgs) {
	visitArgs.VisitDeeper = false
}

func (efv *ExtendFinderVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) {
	ruleset, ok := rulesetNode.(*Ruleset)
	if !ok {
		return
	}

	if ruleset.Root {
		return
	}

	var i, j int
	var extend *Extend
	allSelectorsExtendList := make([]*Extend, 0)
	var extendList []*Extend

	// get &:extend(.a); rules which apply to all selectors in this ruleset
	rules := ruleset.Rules
	ruleCnt := 0
	if rules != nil {
		ruleCnt = len(rules)
	}

	for i = 0; i < ruleCnt; i++ {
		if extendRule, ok := rules[i].(*Extend); ok {
			allSelectorsExtendList = append(allSelectorsExtendList, extendRule)
			ruleset.ExtendOnEveryPath = true
		}
	}

	// now find every selector and apply the extends that apply to all extends
	// and the ones which apply to an individual extend
	paths := ruleset.Paths
	for i = 0; i < len(paths); i++ {
		selectorPath := paths[i]
		if len(selectorPath) == 0 {
			continue // Skip empty selector paths
		}
		selector := selectorPath[len(selectorPath)-1]
		var selExtendList []*Extend
		
		if selectorWithExtends, ok := selector.(interface{ GetExtendList() []*Extend }); ok {
			selExtendList = selectorWithExtends.GetExtendList()
		}

		if selExtendList != nil {
			extendList = make([]*Extend, len(selExtendList))
			copy(extendList, selExtendList)
			extendList = append(extendList, allSelectorsExtendList...)
		} else {
			extendList = allSelectorsExtendList
		}

		if extendList != nil {
			clonedExtendList := make([]*Extend, len(extendList))
			for idx, ext := range extendList {
				clonedExtendList[idx] = ext.Clone(nil)
			}
			extendList = clonedExtendList
		}

		for j = 0; j < len(extendList); j++ {
			efv.foundExtends = true
			extend = extendList[j]
			extend.FindSelfSelectors(selectorPath)
			extend.Ruleset = ruleset
			if j == 0 {
				extend.FirstExtendOnThisSelectorPath = true
			}
			// Defensive bounds check - ensure stack is not empty
			if len(efv.allExtendsStack) == 0 {
				// Initialize stack with empty slice if needed
				efv.allExtendsStack = append(efv.allExtendsStack, make([]any, 0))
			}
			efv.allExtendsStack[len(efv.allExtendsStack)-1] = append(efv.allExtendsStack[len(efv.allExtendsStack)-1], extend)
		}
	}

	efv.contexts = append(efv.contexts, ruleset.Selectors)
}

func (efv *ExtendFinderVisitor) VisitRulesetOut(rulesetNode any) {
	ruleset, ok := rulesetNode.(*Ruleset)
	if !ok {
		return
	}

	if !ruleset.Root && len(efv.contexts) > 0 {
		efv.contexts = efv.contexts[:len(efv.contexts)-1]
	}
}

func (efv *ExtendFinderVisitor) VisitMedia(mediaNode any, visitArgs *VisitArgs) {
	if media, ok := mediaNode.(interface{ SetAllExtends([]*Extend) }); ok {
		media.SetAllExtends(make([]*Extend, 0))
		efv.allExtendsStack = append(efv.allExtendsStack, make([]any, 0))
	}
}

func (efv *ExtendFinderVisitor) VisitMediaOut(mediaNode any) {
	// Before popping the stack, set the collected extends back onto the media node
	if len(efv.allExtendsStack) > 1 {
		// Get the extends collected for this media context
		mediaExtends := efv.allExtendsStack[len(efv.allExtendsStack)-1]

		// Convert []any to []*Extend
		extends := make([]*Extend, 0, len(mediaExtends))
		for _, ext := range mediaExtends {
			if extend, ok := ext.(*Extend); ok {
				extends = append(extends, extend)
			}
		}

		// Set extends back onto the media node
		if media, ok := mediaNode.(interface{ SetAllExtends([]*Extend) }); ok {
			media.SetAllExtends(extends)
		}

		// Pop the stack
		efv.allExtendsStack = efv.allExtendsStack[:len(efv.allExtendsStack)-1]
	}
}

func (efv *ExtendFinderVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) {
	if atRule, ok := atRuleNode.(interface{ SetAllExtends([]*Extend) }); ok {
		atRule.SetAllExtends(make([]*Extend, 0))
		efv.allExtendsStack = append(efv.allExtendsStack, make([]any, 0))
	}
}

func (efv *ExtendFinderVisitor) VisitAtRuleOut(atRuleNode any) {
	// Before popping the stack, set the collected extends back onto the atrule node
	if len(efv.allExtendsStack) > 1 {
		// Get the extends collected for this atrule context
		atRuleExtends := efv.allExtendsStack[len(efv.allExtendsStack)-1]

		// Convert []any to []*Extend
		extends := make([]*Extend, 0, len(atRuleExtends))
		for _, ext := range atRuleExtends {
			if extend, ok := ext.(*Extend); ok {
				extends = append(extends, extend)
			}
		}

		// Set extends back onto the atrule node
		if atRule, ok := atRuleNode.(interface{ SetAllExtends([]*Extend) }); ok {
			atRule.SetAllExtends(extends)
		}

		// Pop the stack
		efv.allExtendsStack = efv.allExtendsStack[:len(efv.allExtendsStack)-1]
	}
}

type ProcessExtendsVisitor struct {
	visitor           *Visitor
	extendIndices     map[string]bool
	allExtendsStack   [][]*Extend
	extendChainCount  int
	// Track Media/AtRule containers we're currently inside for visibility propagation
	mediaAtRuleStack []any
}

func NewProcessExtendsVisitor() *ProcessExtendsVisitor {
	pev := &ProcessExtendsVisitor{
		extendIndices: make(map[string]bool),
	}
	pev.visitor = NewVisitor(pev)
	return pev
}

func (pev *ProcessExtendsVisitor) Run(root any) any {
	extendFinder := NewExtendFinderVisitor()
	pev.extendIndices = make(map[string]bool)
	root = extendFinder.Run(root)

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[ProcessExtendsVisitor.Run] foundExtends=%v, root type=%T\n", extendFinder.foundExtends, root)
		if rootRS, ok := root.(*Ruleset); ok {
			fmt.Fprintf(os.Stderr, "[ProcessExtendsVisitor.Run] root has %d rules\n", len(rootRS.Rules))
		}
	}

	if !extendFinder.foundExtends {
		return root
	}

	// Get allExtends from root - this should now be populated by ExtendFinderVisitor
	var rootAllExtends []*Extend
	if rootWithExtends, ok := root.(interface{ GetAllExtends() []*Extend }); ok {
		rootAllExtends = rootWithExtends.GetAllExtends()
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ProcessExtendsVisitor.Run] Found %d extends in root\n", len(rootAllExtends))
			for i, ext := range rootAllExtends {
				if ext.Selector != nil {
					if sel, ok := ext.Selector.(*Selector); ok {
						fmt.Fprintf(os.Stderr, "[ProcessExtendsVisitor.Run] extend[%d]: selector=%s\n", i, sel.ToCSS(nil))
					}
				}
			}
		}
	}

	// Chain extends and concatenate with original extends
	chained := pev.doExtendChaining(rootAllExtends, rootAllExtends, 0)
	newAllExtends := append(rootAllExtends, chained...)

	// Set the new extends back on root
	if rootWithExtends, ok := root.(interface{ SetAllExtends([]*Extend) }); ok {
		rootWithExtends.SetAllExtends(newAllExtends)
	}

	pev.allExtendsStack = [][]*Extend{newAllExtends}
	newRoot := pev.visitor.Visit(root)
	pev.checkExtendsForNonMatched(newAllExtends)
	return newRoot
}

// IsReplacing returns true as ProcessExtendsVisitor is a replacing visitor
func (pev *ProcessExtendsVisitor) IsReplacing() bool {
	return true
}

// VisitNode implements direct dispatch without reflection for better performance
func (pev *ProcessExtendsVisitor) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {
	switch n := node.(type) {
	case *Declaration:
		pev.VisitDeclaration(n, visitArgs)
		return n, true
	case *MixinDefinition:
		pev.VisitMixinDefinition(n, visitArgs)
		return n, true
	case *Selector:
		pev.VisitSelector(n, visitArgs)
		return n, true
	case *Ruleset:
		pev.VisitRuleset(n, visitArgs)
		return n, true
	case *Media:
		pev.VisitMedia(n, visitArgs)
		return n, true
	case *AtRule:
		pev.VisitAtRule(n, visitArgs)
		return n, true
	default:
		return node, false
	}
}

// VisitNodeOut implements direct dispatch for visitOut methods
func (pev *ProcessExtendsVisitor) VisitNodeOut(node any) bool {
	switch n := node.(type) {
	case *Media:
		pev.VisitMediaOut(n)
		return true
	case *AtRule:
		pev.VisitAtRuleOut(n)
		return true
	default:
		return false
	}
}

func (pev *ProcessExtendsVisitor) checkExtendsForNonMatched(extendList []*Extend) {
	indices := pev.extendIndices
	
	// Filter extends that haven't found matches and have exactly one parent_id
	for _, extend := range extendList {
		if !extend.HasFoundMatches && len(extend.ParentIds) == 1 {
			selector := "_unknown_"
			if extend.Selector != nil {
				if selectorWithCSS, ok := extend.Selector.(interface{ ToCSS(map[string]any) string }); ok {
					// Try to generate CSS, but catch any errors (equivalent to JS try/catch)
					func() {
						defer func() {
							if recover() != nil {
								// CSS generation failed, keep default "_unknown_"
							}
						}()
						selector = selectorWithCSS.ToCSS(make(map[string]any))
					}()
				}
			}

			key := fmt.Sprintf("%d %s", extend.Index, selector)
			if !indices[key] {
				indices[key] = true
				Warn(fmt.Sprintf("WARNING: extend '%s' has no matches", selector))
			}
		}
	}
}

func (pev *ProcessExtendsVisitor) doExtendChaining(extendsList []*Extend, extendsListTarget []*Extend, iterationCount int) []*Extend {
	var extendIndex, targetExtendIndex int
	var matches []*PotentialMatch
	extendsToAdd := make([]*Extend, 0)
	var newSelector []any
	var selectorPath []any
	var extend, targetExtend, newExtend *Extend

	// loop through comparing every extend with every target extend.
	for extendIndex = 0; extendIndex < len(extendsList); extendIndex++ {
		for targetExtendIndex = 0; targetExtendIndex < len(extendsListTarget); targetExtendIndex++ {
			extend = extendsList[extendIndex]
			targetExtend = extendsListTarget[targetExtendIndex]

			// look for circular references
			found := false
			for _, parentId := range extend.ParentIds {
				if parentId == targetExtend.ObjectId {
					found = true
					break
				}
			}
			if found {
				continue
			}

			// find a match in the target extends self selector (the bit before :extend)
			if len(targetExtend.SelfSelectors) > 0 {
				selectorPath = []any{targetExtend.SelfSelectors[0]}
				matches = pev.findMatch(extend, selectorPath)

				if len(matches) > 0 {
					extend.HasFoundMatches = true

					// we found a match, so for each self selector..
					for _, selfSelector := range extend.SelfSelectors {
						// Use the TARGET extend's visibility info for the chained extend.
						// This matches JavaScript extend-visitor.js line 194:
						//   const info = targetExtend.visibilityInfo();
						//
						// When .c extends .b and .b extends .a (reference import):
						// - extend = .c extends .b (visible, from main file)
						// - targetExtend = .b extends .a (invisible, from reference import)
						// The chained extend .c extends .a inherits visibility from targetExtend
						// so it has visibilityBlocks > 0. When it later matches .a in visitRuleset,
						// its isVisible() returns nil (falsy), marking new selectors invisible.
						var info any
						info = targetExtend.VisibilityInfo()

						// process the extend as usual
						// Match JavaScript: use extend.isVisible() to preserve visibility in chaining
						newSelector = pev.extendSelector(matches, selectorPath, selfSelector, extend.IsVisible())

						// but now we create a new extend from it
						var infoMap map[string]any
						if infoMapTyped, ok := info.(map[string]any); ok {
							infoMap = infoMapTyped
						}
						newExtend = NewExtend(targetExtend.Selector, targetExtend.Option, 0, targetExtend.FileInfo(), infoMap)
						newExtend.SelfSelectors = newSelector

						// add the extend onto the list of extends for that selector
						if len(newSelector) > 0 {
							if selectorWithExtends, ok := newSelector[len(newSelector)-1].(interface{ SetExtendList([]*Extend) }); ok {
								selectorWithExtends.SetExtendList([]*Extend{newExtend})
							}
						}

						// record that we need to add it.
						extendsToAdd = append(extendsToAdd, newExtend)
						newExtend.Ruleset = targetExtend.Ruleset

						// Detailed logging for investigation
						if os.Getenv("LESS_GO_TRACE") == "1" {
							extendStr := "?"
							if len(extend.SelfSelectors) > 0 {
								if sel, ok := extend.SelfSelectors[0].(*Selector); ok {
									extendStr = sel.ToCSS(nil)
								}
							}
							targetStr := "?"
							if targetExtend.Selector != nil {
								if sel, ok := targetExtend.Selector.(*Selector); ok {
									targetStr = sel.ToCSS(nil)
								}
							}
							rulesetFirstSel := "?"
							rulesetPtr := "nil"
							rulesetHasVisibility := false
							if targetExtend.Ruleset != nil {
								rulesetPtr = fmt.Sprintf("%p", targetExtend.Ruleset)
								if targetExtend.Ruleset.Node != nil {
									rulesetHasVisibility = targetExtend.Ruleset.Node.BlocksVisibility()
								}
								if len(targetExtend.Ruleset.Paths) > 0 {
									if len(targetExtend.Ruleset.Paths[0]) > 0 {
										if sel, ok := targetExtend.Ruleset.Paths[0][0].(*Selector); ok {
											rulesetFirstSel = sel.ToCSS(nil)
										}
									}
								}
							}
							fmt.Printf("[CHAIN] %s extends %s → new chained extend points to ruleset %s (selector=%s, hasVisibilityBlock=%v)\n",
								extendStr, targetStr, rulesetPtr, rulesetFirstSel, rulesetHasVisibility)
						}

						// remember its parents for circular references
						newExtend.ParentIds = append(newExtend.ParentIds, targetExtend.ParentIds...)
						newExtend.ParentIds = append(newExtend.ParentIds, extend.ParentIds...)

						// only process the selector once.. if we have :extend(.a,.b) then multiple
						// extends will look at the same selector path, so when extending
						// we know that any others will be duplicates in terms of what is added to the css
						if targetExtend.FirstExtendOnThisSelectorPath {
							newExtend.FirstExtendOnThisSelectorPath = true
							// Check if this selector already exists before adding it
							if !pev.selectorExists(targetExtend.Ruleset.Paths, newSelector) {
								targetExtend.Ruleset.Paths = append(targetExtend.Ruleset.Paths, newSelector)
							}
						}
					}
				}
			}
		}
	}

	if len(extendsToAdd) > 0 {
		// try to detect circular references to stop a stack overflow.
		pev.extendChainCount++
		if iterationCount > 100 {
			selectorOne := "{unable to calculate}"
			selectorTwo := "{unable to calculate}"
			
			// Try to get selector CSS for error message (equivalent to JS try/catch)
			if len(extendsToAdd) > 0 && len(extendsToAdd[0].SelfSelectors) > 0 {
				if selectorWithCSS, ok := extendsToAdd[0].SelfSelectors[0].(interface{ ToCSS() string }); ok {
					func() {
						defer func() {
							if recover() != nil {
								// CSS generation failed, keep default
							}
						}()
						selectorOne = selectorWithCSS.ToCSS()
					}()
				}
			}
			
			if len(extendsToAdd) > 0 && extendsToAdd[0].Selector != nil {
				if selectorWithCSS, ok := extendsToAdd[0].Selector.(interface{ ToCSS() string }); ok {
					func() {
						defer func() {
							if recover() != nil {
								// CSS generation failed, keep default
							}
						}()
						selectorTwo = selectorWithCSS.ToCSS()
					}()
				}
			}
			
			panic(fmt.Sprintf("extend circular reference detected. One of the circular extends is currently:%s:extend(%s)", selectorOne, selectorTwo))
		}

		// now process the new extends on the existing rules so that we can handle a extending b extending c extending d extending e...
		recursive := pev.doExtendChaining(extendsToAdd, extendsListTarget, iterationCount+1)
		return append(extendsToAdd, recursive...)
	} else {
		return extendsToAdd
	}
}

func (pev *ProcessExtendsVisitor) VisitDeclaration(ruleNode any, visitArgs *VisitArgs) {
	visitArgs.VisitDeeper = false
}

func (pev *ProcessExtendsVisitor) VisitMixinDefinition(mixinDefinitionNode any, visitArgs *VisitArgs) {
	visitArgs.VisitDeeper = false
}

func (pev *ProcessExtendsVisitor) VisitSelector(selectorNode any, visitArgs *VisitArgs) {
	visitArgs.VisitDeeper = false
}

func (pev *ProcessExtendsVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) {
	ruleset, ok := rulesetNode.(*Ruleset)
	if !ok {
		return
	}

	if ruleset.Root {
		return
	}

	var matches []*PotentialMatch
	var pathIndex, extendIndex int
	allExtends := pev.allExtendsStack[len(pev.allExtendsStack)-1]
	var selectorPath []any

	// Cache the original paths length to avoid processing paths added during extend chaining
	// Paths added during chaining have extends and should not be extended again
	originalPathsLength := len(ruleset.Paths)

	// DEBUG: Trace rulesets being visited and their paths count
	if os.Getenv("LESS_GO_TRACE_EXTEND") == "1" {
		var sel string = "?"
		if len(ruleset.Selectors) > 0 {
			if s, ok := ruleset.Selectors[0].(*Selector); ok {
				sel = s.ToCSS(nil)
			}
		}
		fmt.Fprintf(os.Stderr, "[EXTEND VisitRuleset] selector=%s, paths=%d, blocksVisibility=%v\n",
			sel, originalPathsLength, ruleset.Node != nil && ruleset.Node.BlocksVisibility())
	}

	// Track which rulesets were modified so we can deduplicate their paths
	modifiedRulesets := make(map[*Ruleset]bool)

	// look at each selector path in the ruleset, find any extend matches and then copy, find and replace
	for extendIndex = 0; extendIndex < len(allExtends); extendIndex++ {
		for pathIndex = 0; pathIndex < originalPathsLength; pathIndex++ {
			selectorPath = ruleset.Paths[pathIndex]

			// extending extends happens initially, before the main pass
			// Match JavaScript: unconditionally skip rulesets with extendOnEveryPath (line 279)
			if ruleset.ExtendOnEveryPath {
				continue
			}

			// Match JavaScript: unconditionally skip selectors with extends (line 280-281)
			// The JavaScript code skips selectors with extends because doExtendChaining
			// already processed them and added new selector paths (without extends) to the rulesets.
			// Only these NEW paths should be processed in visitRuleset.
			if len(selectorPath) > 0 {
				lastElement := selectorPath[len(selectorPath)-1]
				if selectorWithExtends, ok := lastElement.(interface{ GetExtendList() []*Extend }); ok {
					if extendList := selectorWithExtends.GetExtendList(); extendList != nil && len(extendList) > 0 {
						continue
					}
				}
			}

			matches = pev.findMatch(allExtends[extendIndex], selectorPath)

			if os.Getenv("LESS_GO_DEBUG") == "1" && len(matches) > 0 {
				extendSel := "?"
				if allExtends[extendIndex].Selector != nil {
					if sel, ok := allExtends[extendIndex].Selector.(*Selector); ok {
						extendSel = sel.ToCSS(nil)
					}
				}
				pathSel := "?"
				if len(selectorPath) > 0 {
					if sel, ok := selectorPath[0].(*Selector); ok {
						pathSel = sel.ToCSS(nil)
					}
				}
				fmt.Fprintf(os.Stderr, "[EXTEND MATCH] extend=%s matched selector=%s (matches=%d)\n",
					extendSel, pathSel, len(matches))
			}

			if len(matches) > 0 {
				allExtends[extendIndex].HasFoundMatches = true

				// Match JavaScript: use the extend's visibility to determine if created selectors should be visible
				// This ensures that extends from reference imports don't create visible selectors
				// unless they've been explicitly made visible by being used/extended from outside the reference
				// Note: Extend.IsVisible() returns bool (not *bool), taking visibility blocks into account
				isVisible := allExtends[extendIndex].IsVisible()

				// Note: We intentionally do NOT check extendRuleset.Node.BlocksVisibility() here
				// For chained extends, the Ruleset may be from a reference import even if the original
				// extend is from the main file. We should respect the extend's own visibility.

				if os.Getenv("LESS_GO_TRACE") == "1" {
					var extendSel string = "?"
					if len(allExtends[extendIndex].SelfSelectors) > 0 {
						if sel, ok := allExtends[extendIndex].SelfSelectors[0].(*Selector); ok {
							extendSel = sel.ToCSS(nil)
						}
					}
					var targetSel string = "?"
					if allExtends[extendIndex].Selector != nil {
						if sel, ok := allExtends[extendIndex].Selector.(*Selector); ok {
							targetSel = sel.ToCSS(nil)
						}
					}
					var pathSel string = "?"
					if len(selectorPath) > 0 {
						if sel, ok := selectorPath[0].(*Selector); ok {
							pathSel = sel.ToCSS(nil)
						}
					}
					var extendRulesetPtr string = "nil"
					if allExtends[extendIndex].Ruleset != nil {
						extendRulesetPtr = fmt.Sprintf("%p", allExtends[extendIndex].Ruleset)
					}
					fmt.Printf("[MATCH] Extend %s (%s) matched path %s, isVisible=%v, extendRuleset=%s, numParents=%d\n",
						extendSel, targetSel, pathSel, isVisible, extendRulesetPtr, len(allExtends[extendIndex].ParentIds))
				}

				// Check if the matched selector path has visibility blocks (is from a reference import)
				// Also check if the ruleset itself has visibility blocks
				// This determines if an invisible extend should match this selector/ruleset
				selectorHasVisibilityBlocks := false
				for _, pathSelector := range selectorPath {
					if sel, ok := pathSelector.(*Selector); ok {
						if sel.Node != nil && sel.Node.BlocksVisibility() {
							selectorHasVisibilityBlocks = true
							break
						}
					}
				}
				rulesetHasVisibilityBlocks := ruleset.Node != nil && ruleset.Node.BlocksVisibility()

				// Only process the match if:
				// 1. The extend is visible (from a non-reference import), OR
				// 2. The extend, selector, and ruleset are all from reference imports (all have visibility blocks)
				// NOTE: With the architectural fix of adding paths to extend's ruleset (not matched ruleset),
				// we don't need the complex visibility checks from master's workaround approach.
				if isVisible || (selectorHasVisibilityBlocks && rulesetHasVisibilityBlocks) {
					// CRITICAL FIX: When a visible extend (from outside a reference import) matches
					// selectors from a reference import, mark the EXTEND'S RULESET as visible (not the matched ruleset).
					// For chained extends, the extend's Ruleset field points to the intermediate ruleset,
					// ensuring properties come from the correct ruleset in the chain.
					//
					// Example: .c extends .b, .b extends .a (reference import)
					// - Direct extend: .c extends .b → extend.Ruleset points to .b's ruleset
					// - Chained extend: .c extends .a → extend.Ruleset points to .b's ruleset (NOT .a's ruleset)
					// - When chained extend matches .a, we mark .b's ruleset visible, not .a's ruleset
					targetRuleset := allExtends[extendIndex].Ruleset
					if targetRuleset == nil {
						targetRuleset = ruleset // Fallback to matched ruleset if extend has no ruleset
					}

					if isVisible && (selectorHasVisibilityBlocks || rulesetHasVisibilityBlocks) {
						// Mark the EXTEND'S ruleset as visible so its content is output with the extended selector
						if targetRuleset.Node != nil {
							if os.Getenv("LESS_GO_DEBUG") == "1" {
								var targetSel string = "?"
								if len(targetRuleset.Paths) > 0 && len(targetRuleset.Paths[0]) > 0 {
									if sel, ok := targetRuleset.Paths[0][0].(*Selector); ok {
										targetSel = sel.ToCSS(nil)
									}
								}
								fmt.Fprintf(os.Stderr, "[VISIBILITY] Marking extend's ruleset %p (selector=%s) as visible due to extend match\n",
									targetRuleset, targetSel)
							}
							targetRuleset.Node.EnsureVisibility()

							// Walk up the parent chain and make all parent Media/AtRule nodes visible
							// This ensures that @media and @supports blocks containing extended selectors are output
							pev.makeParentNodesVisible(targetRuleset.Node)
						}

						// CRITICAL FIX: Also make the MATCHED ruleset's parent Media/AtRule nodes visible
						// When extending into a @media block from a reference import, we need to make that
						// @media block visible so the extended selectors inside it are output.
						// Example: .class:extend(.class all) should make @media print { .class {...} } visible
						if ruleset.Node != nil && rulesetHasVisibilityBlocks {
							if os.Getenv("LESS_GO_DEBUG") == "1" {
								var matchedSel string = "?"
								if len(ruleset.Paths) > 0 && len(ruleset.Paths[0]) > 0 {
									if sel, ok := ruleset.Paths[0][0].(*Selector); ok {
										matchedSel = sel.ToCSS(nil)
									}
								}
								fmt.Fprintf(os.Stderr, "[VISIBILITY] Making matched ruleset %p (selector=%s) parents visible due to extend\n",
									ruleset, matchedSel)
							}
							// Make parent Media/AtRule nodes visible so they output the extended selectors
							pev.makeParentNodesVisible(ruleset.Node)
						}
					}

					for _, selfSelector := range allExtends[extendIndex].SelfSelectors {
						extendedSelectors := pev.extendSelector(matches, selectorPath, selfSelector, isVisible)
						// CRITICAL: Add paths to the MATCHED ruleset (like JavaScript extend-visitor.js line 296)
						// This ensures extend chaining works correctly (.c:extend(.b) + .b:extend(.a) = .a,.b,.c)
						// The targetRuleset is only used for visibility management (reference imports)
						ruleset.Paths = append(ruleset.Paths, extendedSelectors)
						// Mark this ruleset as modified so we can deduplicate it later
						modifiedRulesets[ruleset] = true
					}

					// CRITICAL FIX: When extending into a ruleset from a reference import,
					// we must ensure the ruleset is visible so it passes the KeepOnlyVisibleChilds filter
					// in ToCSSVisitor.ResolveVisibility. Without this, the ruleset would be filtered out
					// even though it has visible paths from the extend.
					if isVisible && rulesetHasVisibilityBlocks && ruleset.Node != nil {
						ruleset.Node.EnsureVisibility()
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							var matchedSel string = "?"
							if len(ruleset.Paths) > 0 && len(ruleset.Paths[0]) > 0 {
								if sel, ok := ruleset.Paths[0][0].(*Selector); ok {
									matchedSel = sel.ToCSS(nil)
								}
							}
							fmt.Fprintf(os.Stderr, "[VISIBILITY] Made matched ruleset %p (selector=%s) visible for KeepOnlyVisibleChilds\n",
								ruleset, matchedSel)
						}
					}
				}
			}
		}
	}

	// Deduplicate paths in all modified rulesets
	for modifiedRuleset := range modifiedRulesets {
		modifiedRuleset.Paths = pev.deduplicatePaths(modifiedRuleset.Paths)
	}
}

// selectorExists checks if a selector path already exists in the paths list
func (pev *ProcessExtendsVisitor) selectorExists(paths [][]any, newPath []any) bool {
	newCSS := pev.pathToCSS(newPath)
	for _, path := range paths {
		if pev.pathToCSS(path) == newCSS {
			return true
		}
	}
	return false
}

// deduplicatePaths removes duplicate selector paths based on their CSS output
// When duplicates are found, it prefers the path with visible selectors
func (pev *ProcessExtendsVisitor) deduplicatePaths(paths [][]any) [][]any {
	if len(paths) == 0 {
		return paths
	}

	// Use a map to track CSS strings we've seen and their index in unique
	seenIndex := make(map[string]int)
	unique := make([][]any, 0, len(paths))

	for _, path := range paths {
		// Generate CSS for this path
		css := pev.pathToCSS(path)

		// Check if this path has any visible selectors
		pathHasVisible := false
		for _, selector := range path {
			if sel, ok := selector.(*Selector); ok {
				if vis := sel.IsVisible(); vis != nil && *vis {
					pathHasVisible = true
					break
				}
			}
		}

		if idx, exists := seenIndex[css]; exists {
			// We've seen this CSS before - check if we should replace it
			// Prefer paths with visible selectors
			if pathHasVisible {
				existingPath := unique[idx]
				existingHasVisible := false
				for _, selector := range existingPath {
					if sel, ok := selector.(*Selector); ok {
						if vis := sel.IsVisible(); vis != nil && *vis {
							existingHasVisible = true
							break
						}
					}
				}
				// Replace if existing doesn't have visible but new one does
				if !existingHasVisible {
					unique[idx] = path
				}
			}
		} else {
			// First time seeing this CSS
			seenIndex[css] = len(unique)
			unique = append(unique, path)
		}
	}

	return unique
}

// pathToCSS converts a selector path to CSS string for comparison
func (pev *ProcessExtendsVisitor) pathToCSS(path []any) string {
	if len(path) == 0 {
		return ""
	}

	var chunks []string
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			if chunk != nil {
				chunks = append(chunks, fmt.Sprintf("%v", chunk))
			}
		},
		IsEmpty: func() bool {
			return len(chunks) == 0
		},
	}

	ctx := make(map[string]any)
	ctx["compress"] = true // Use compressed output for comparison
	ctx["firstSelector"] = true

	for i, pathElement := range path {
		if i > 0 {
			ctx["firstSelector"] = false
		}

		if gen, ok := pathElement.(interface{ GenCSS(any, *CSSOutput) }); ok {
			gen.GenCSS(ctx, output)
		}
	}

	return strings.Join(chunks, "")
}

func (pev *ProcessExtendsVisitor) findMatch(extend *Extend, haystackSelectorPath []any) []*PotentialMatch {
	// This matches the JavaScript findMatch method exactly
	var haystackSelectorIndex, hackstackElementIndex int
	var hackstackSelector, haystackElement any
	var targetCombinator string
	var i int
	needleElements := extend.Selector.(*Selector).Elements
	// Pre-allocate with reasonable capacity based on typical usage patterns
	potentialMatches := make([]*PotentialMatch, 0, 8)
	var potentialMatch *PotentialMatch
	matches := make([]*PotentialMatch, 0, 4)

	// loop through the haystack elements
	for haystackSelectorIndex = 0; haystackSelectorIndex < len(haystackSelectorPath); haystackSelectorIndex++ {
		hackstackSelector = haystackSelectorPath[haystackSelectorIndex]
		
		var hackstackElements []*Element
		if selector, ok := hackstackSelector.(*Selector); ok {
			hackstackElements = selector.Elements
		} else {
			// Skip if not a proper selector
			continue
		}

		for hackstackElementIndex = 0; hackstackElementIndex < len(hackstackElements); hackstackElementIndex++ {
			haystackElement = hackstackElements[hackstackElementIndex]

			// if we allow elements before our match we can add a potential match every time. otherwise only at the first element.
			if extend.AllowBefore || (haystackSelectorIndex == 0 && hackstackElementIndex == 0) {
				var initialCombinator string
				if element, ok := haystackElement.(*Element); ok {
					if element.Combinator != nil {
						initialCombinator = element.Combinator.Value
					}
				}

				potentialMatches = append(potentialMatches, &PotentialMatch{
					pathIndex:         haystackSelectorIndex,
					index:            hackstackElementIndex,
					matched:          0,
					initialCombinator: initialCombinator,
				})
			}

			for i = 0; i < len(potentialMatches); i++ {
				potentialMatch = potentialMatches[i]

				// selectors add " " onto the first element. When we use & it joins the selectors together, but if we don't
				// then each selector in haystackSelectorPath has a space before it added in the toCSS phase. so we need to
				// work out what the resulting combinator will be
				targetCombinator = ""
				if element, ok := haystackElement.(*Element); ok && element.Combinator != nil {
					targetCombinator = element.Combinator.Value
				}
				if targetCombinator == "" && hackstackElementIndex == 0 {
					targetCombinator = " "
				}

				matched := potentialMatch.matched
				
				// if we don't match, null our match to indicate failure
				if !pev.isElementValuesEqual(needleElements[matched].Value, haystackElement.(*Element).Value) {
					potentialMatch = nil
				} else if matched > 0 {
					var needleCombinator string
					if needleElements[matched].Combinator != nil {
						needleCombinator = needleElements[matched].Combinator.Value
					}
					if needleCombinator != targetCombinator {
						potentialMatch = nil
					}
				}
				
				if potentialMatch != nil {
					potentialMatch.matched = matched + 1
				}

				// if we are still valid and have finished, test whether we have elements after and whether these are allowed
				if potentialMatch != nil {
					finished := potentialMatch.matched == len(needleElements)
					potentialMatch.finished = finished
					
					if finished && !extend.AllowAfter {
						if hackstackElementIndex+1 < len(hackstackElements) || haystackSelectorIndex+1 < len(haystackSelectorPath) {
							potentialMatch = nil
						}
					}
				}
				
				// if null we remove, if not, we are still valid, so either push as a valid match or continue
				if potentialMatch != nil {
					if potentialMatch.finished {
						potentialMatch.length = len(needleElements)
						potentialMatch.endPathIndex = haystackSelectorIndex
						potentialMatch.endPathElementIndex = hackstackElementIndex + 1 // index after end of match
						potentialMatches = potentialMatches[:0] // we don't allow matches to overlap, so start matching again
						matches = append(matches, potentialMatch)
					} else {
						potentialMatches[i] = potentialMatch
					}
				} else {
					// Remove null match - splice operation equivalent to JS
					potentialMatches = append(potentialMatches[:i], potentialMatches[i+1:]...)
					i--
				}
			}
		}
	}
	return matches
}

func (pev *ProcessExtendsVisitor) isElementValuesEqual(elementValue1, elementValue2 any) bool {
	// Handle string comparison
	if str1, ok1 := elementValue1.(string); ok1 {
		if str2, ok2 := elementValue2.(string); ok2 {
			return str1 == str2
		}
		return false
	}
	if _, ok2 := elementValue2.(string); ok2 {
		return false
	}

	// Handle Attribute comparison
	if attr1, ok1 := elementValue1.(*Attribute); ok1 {
		if attr2, ok2 := elementValue2.(*Attribute); ok2 {
			// Compare operators
			if attr1.Op != attr2.Op {
				return false
			}

			// Compare keys - need to handle both string and node types
			key1, key2 := attr1.Key, attr2.Key
			if k1, ok := key1.(interface{ ToCSS(any) string }); ok {
				key1 = k1.ToCSS(nil)
			}
			if k2, ok := key2.(interface{ ToCSS(any) string }); ok {
				key2 = k2.ToCSS(nil)
			}
			if key1 != key2 {
				return false
			}

			// Compare values
			if attr1.Value == nil || attr2.Value == nil {
				return attr1.Value == attr2.Value
			}

			// Get the actual values (matching JavaScript: elementValue1.value.value || elementValue1.value)
			// JavaScript uses: elementValue1 = elementValue1.value.value || elementValue1.value
			// In Go, we need to check for both GetValue() string and GetValue() any
			var val1, val2 any = attr1.Value, attr2.Value

			// Try to extract value from Quoted or similar types
			// Check for GetValue() string first (like Quoted)
			if valueProvider1, ok := attr1.Value.(interface{ GetValue() string }); ok {
				val1 = valueProvider1.GetValue()
			} else if valueProvider1, ok := attr1.Value.(interface{ GetValue() any }); ok {
				val1 = valueProvider1.GetValue()
			}

			if valueProvider2, ok := attr2.Value.(interface{ GetValue() string }); ok {
				val2 = valueProvider2.GetValue()
			} else if valueProvider2, ok := attr2.Value.(interface{ GetValue() any }); ok {
				val2 = valueProvider2.GetValue()
			}

			// Direct comparison
			return val1 == val2
		}
		return false
	}

	// Get values for comparison
	var val1, val2 any
	if valueProvider1, ok := elementValue1.(interface{ GetValue() any }); ok {
		val1 = valueProvider1.GetValue()
	} else {
		val1 = elementValue1
	}
	if valueProvider2, ok := elementValue2.(interface{ GetValue() any }); ok {
		val2 = valueProvider2.GetValue()
	} else {
		val2 = elementValue2
	}

	// Handle Selector comparison
	if selector1, ok1 := val1.(*Selector); ok1 {
		if selector2, ok2 := val2.(*Selector); ok2 {
			if len(selector1.Elements) != len(selector2.Elements) {
				return false
			}
			for i := 0; i < len(selector1.Elements); i++ {
				var comb1, comb2 string
				if selector1.Elements[i].Combinator != nil {
					comb1 = selector1.Elements[i].Combinator.Value
				}
				if selector2.Elements[i].Combinator != nil {
					comb2 = selector2.Elements[i].Combinator.Value
				}
				
				if comb1 != comb2 {
					if i != 0 {
						return false
					}
					// Handle first element special case
					defaultComb1 := comb1
					if defaultComb1 == "" {
						defaultComb1 = " "
					}
					defaultComb2 := comb2
					if defaultComb2 == "" {
						defaultComb2 = " "
					}
					if defaultComb1 != defaultComb2 {
						return false
					}
				}
				if !pev.isElementValuesEqual(selector1.Elements[i].Value, selector2.Elements[i].Value) {
					return false
				}
			}
			return true
		}
		return false
	}
	
	return false
}

func (pev *ProcessExtendsVisitor) extendSelector(matches []*PotentialMatch, selectorPath []any, replacementSelector any, isVisible bool) []any {
	// This matches the JavaScript extendSelector method exactly (lines 417-482)
	currentSelectorPathIndex := 0
	currentSelectorPathElementIndex := 0
	path := make([]any, 0)
	var matchIndex int
	var selector *Selector
	var firstElement *Element
	var match *PotentialMatch
	var newElements []*Element

	for matchIndex = 0; matchIndex < len(matches); matchIndex++ {
		match = matches[matchIndex]
		pathIndex := match.pathIndex
		if pathIndex < 0 || pathIndex >= len(selectorPath) {
			panic(fmt.Sprintf("Invalid pathIndex %d for selectorPath length %d", pathIndex, len(selectorPath)))
		}
		selector = selectorPath[pathIndex].(*Selector)
		
		// Get replacement selector elements
		replacementSel := replacementSelector.(*Selector)
		
		firstElement = NewElement(
			match.initialCombinator,
			replacementSel.Elements[0].Value,
			replacementSel.Elements[0].IsVariable,
			replacementSel.Elements[0].GetIndex(),
			replacementSel.Elements[0].FileInfo(),
			replacementSel.Elements[0].VisibilityInfo(),
		)

		if match.pathIndex > currentSelectorPathIndex && currentSelectorPathElementIndex > 0 {
			// Equivalent to JS: path[path.length - 1].elements = path[path.length - 1].elements.concat(...)
			if len(path) > 0 {
				if pathSel, ok := path[len(path)-1].(*Selector); ok {
					currentSelector := selectorPath[currentSelectorPathIndex].(*Selector)
					sliceStart := currentSelectorPathElementIndex
					if sliceStart < len(currentSelector.Elements) {
						pathSel.Elements = append(pathSel.Elements, currentSelector.Elements[sliceStart:]...)
					}
				}
			}
			currentSelectorPathElementIndex = 0
			currentSelectorPathIndex++
		}

		// Build newElements exactly like JavaScript
		newElements = make([]*Element, 0)
		
		// Add elements before the match (equivalent to selector.elements.slice(currentSelectorPathElementIndex, match.index))
		sliceEnd := match.index
		if sliceEnd > currentSelectorPathElementIndex && currentSelectorPathElementIndex < len(selector.Elements) {
			if sliceEnd > len(selector.Elements) {
				sliceEnd = len(selector.Elements)
			}
			newElements = append(newElements, selector.Elements[currentSelectorPathElementIndex:sliceEnd]...)
		}
		
		// Add the first replacement element
		newElements = append(newElements, firstElement)
		
		// Add remaining replacement elements (equivalent to .concat(replacementSelector.elements.slice(1)))
		if len(replacementSel.Elements) > 1 {
			newElements = append(newElements, replacementSel.Elements[1:]...)
		}

		if currentSelectorPathIndex == match.pathIndex && matchIndex > 0 {
			// Equivalent to JS: path[path.length - 1].elements = path[path.length - 1].elements.concat(newElements)
			if len(path) > 0 {
				if pathSel, ok := path[len(path)-1].(*Selector); ok {
					pathSel.Elements = append(pathSel.Elements, newElements...)
				}
			}
		} else {
			// Equivalent to JS: path = path.concat(selectorPath.slice(currentSelectorPathIndex, match.pathIndex))
			if match.pathIndex > currentSelectorPathIndex {
				path = append(path, selectorPath[currentSelectorPathIndex:match.pathIndex]...)
			}

			// Inherit visibility info from the matched selector (selector from selectorPath)
			// This ensures that selectors created during extend processing preserve visibility blocks
			// from reference imports
			var visibilityInfo map[string]any
			if selector.Node != nil {
				visibilityInfo = selector.VisibilityInfo()
			}

			// Equivalent to JS: path.push(new tree.Selector(newElements))
			newSelector, err := NewSelector(newElements, nil, nil, 0, nil, visibilityInfo)
			if err == nil {
				path = append(path, newSelector)
			}
		}
		
		currentSelectorPathIndex = match.endPathIndex
		currentSelectorPathElementIndex = match.endPathElementIndex
		
		// Handle element index overflow (equivalent to JS lines 458-462)
		if currentSelectorPathIndex < len(selectorPath) {
			currentSelector := selectorPath[currentSelectorPathIndex].(*Selector)
			if currentSelectorPathElementIndex >= len(currentSelector.Elements) {
				currentSelectorPathElementIndex = 0
				currentSelectorPathIndex++
			}
		}
	}

	// Handle remaining elements (equivalent to JS lines 464-468)
	if currentSelectorPathIndex < len(selectorPath) && currentSelectorPathElementIndex > 0 {
		if len(path) > 0 {
			if pathSel, ok := path[len(path)-1].(*Selector); ok {
				currentSelector := selectorPath[currentSelectorPathIndex].(*Selector)
				if currentSelectorPathElementIndex < len(currentSelector.Elements) {
					pathSel.Elements = append(pathSel.Elements, currentSelector.Elements[currentSelectorPathElementIndex:]...)
				}
			}
		}
		currentSelectorPathIndex++
	}

	// Equivalent to JS: path = path.concat(selectorPath.slice(currentSelectorPathIndex, selectorPath.length))
	path = append(path, selectorPath[currentSelectorPathIndex:]...)
	
	// Apply visibility (equivalent to JS lines 471-481)
	for i, currentValue := range path {
		if selector, ok := currentValue.(*Selector); ok {
			// Equivalent to JS: currentValue.createDerived(currentValue.elements)
			derived, err := selector.CreateDerived(selector.Elements, nil, nil)
			if err == nil {
				if isVisible {
					derived.EnsureVisibility()
					// CRITICAL: Also set EvaldCondition = true so the selector passes the isOutput check
					// This ensures that newly created selectors from extends are included in CSS output
					derived.EvaldCondition = true
				} else {
					derived.EnsureInvisibility()
				}
				path[i] = derived
			}
		}
	}
	
	return path
}

func (pev *ProcessExtendsVisitor) VisitMedia(mediaNode any, visitArgs *VisitArgs) {
	var mediaAllExtends []*Extend
	if media, ok := mediaNode.(interface{ GetAllExtends() []*Extend }); ok {
		mediaAllExtends = media.GetAllExtends()
	}

	// Guard against empty stack - initialize with empty slice
	if len(pev.allExtendsStack) == 0 {
		pev.allExtendsStack = [][]*Extend{make([]*Extend, 0)}
	}

	currentAllExtends := pev.allExtendsStack[len(pev.allExtendsStack)-1]
	newAllExtends := append(mediaAllExtends, currentAllExtends...)
	chained := pev.doExtendChaining(newAllExtends, mediaAllExtends, 0)
	newAllExtends = append(newAllExtends, chained...)
	pev.allExtendsStack = append(pev.allExtendsStack, newAllExtends)

	// Push this Media node onto the stack for visibility propagation
	pev.mediaAtRuleStack = append(pev.mediaAtRuleStack, mediaNode)
}

func (pev *ProcessExtendsVisitor) VisitMediaOut(mediaNode any) {
	// Match JavaScript: this.allExtendsStack.length = lastIndex;
	// But ensure we never go below 1 element (the root level)
	if len(pev.allExtendsStack) > 1 {
		lastIndex := len(pev.allExtendsStack) - 1
		pev.allExtendsStack = pev.allExtendsStack[:lastIndex]
	}

	// Pop the Media node from the stack
	if len(pev.mediaAtRuleStack) > 0 {
		pev.mediaAtRuleStack = pev.mediaAtRuleStack[:len(pev.mediaAtRuleStack)-1]
	}
}

func (pev *ProcessExtendsVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) {
	var atRuleAllExtends []*Extend
	if atRule, ok := atRuleNode.(interface{ GetAllExtends() []*Extend }); ok {
		atRuleAllExtends = atRule.GetAllExtends()
	}

	// Guard against empty stack - initialize with empty slice
	if len(pev.allExtendsStack) == 0 {
		pev.allExtendsStack = [][]*Extend{make([]*Extend, 0)}
	}

	currentAllExtends := pev.allExtendsStack[len(pev.allExtendsStack)-1]
	newAllExtends := append(atRuleAllExtends, currentAllExtends...)
	chained := pev.doExtendChaining(newAllExtends, atRuleAllExtends, 0)
	newAllExtends = append(newAllExtends, chained...)
	pev.allExtendsStack = append(pev.allExtendsStack, newAllExtends)

	// Push this AtRule node onto the stack for visibility propagation
	pev.mediaAtRuleStack = append(pev.mediaAtRuleStack, atRuleNode)
}

func (pev *ProcessExtendsVisitor) VisitAtRuleOut(atRuleNode any) {
	// Match JavaScript: this.allExtendsStack.length = lastIndex;
	// But ensure we never go below 1 element (the root level)
	if len(pev.allExtendsStack) > 1 {
		lastIndex := len(pev.allExtendsStack) - 1
		pev.allExtendsStack = pev.allExtendsStack[:lastIndex]
	}

	// Pop the AtRule node from the stack
	if len(pev.mediaAtRuleStack) > 0 {
		pev.mediaAtRuleStack = pev.mediaAtRuleStack[:len(pev.mediaAtRuleStack)-1]
	}
}

// makeParentNodesVisible makes all Media and AtRule nodes in the current stack visible.
// This is critical for import (reference) to work correctly - when we extend a selector
// from a reference import, we need to make the entire @media or @supports block visible,
// not just the ruleset.
func (pev *ProcessExtendsVisitor) makeParentNodesVisible(node *Node) {
	// Mark all Media/AtRule containers in the stack as visible
	// NOTE: The specific ruleset containing the extended selector has already been marked
	// visible by the caller, so we don't need to mark child rulesets here.
	// We only need to ensure the parent Media/AtRule containers are visible so they output.

	// DEBUG: Trace what's in the mediaAtRuleStack
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[makeParentNodesVisible] Stack size=%d\n", len(pev.mediaAtRuleStack))
		for i, containerNode := range pev.mediaAtRuleStack {
			switch v := containerNode.(type) {
			case *Media:
				features := "?"
				if feat, ok := v.Features.(interface{ ToCSS(any) string }); ok {
					features = feat.ToCSS(nil)
				}
				fmt.Fprintf(os.Stderr, "  [%d] Media features=%s: blocksVisibility=%v\n", i, features, v.Node.BlocksVisibility())
			case *AtRule:
				fmt.Fprintf(os.Stderr, "  [%d] AtRule name=%s: blocksVisibility=%v\n", i, v.Name, v.Node.BlocksVisibility())
			default:
				fmt.Fprintf(os.Stderr, "  [%d] Unknown: %T\n", i, containerNode)
			}
		}
	}

	for _, containerNode := range pev.mediaAtRuleStack {
		switch v := containerNode.(type) {
		case *Media:
			// Make the Media node visible
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				visBefore := "nil"
				if vis := v.Node.IsVisible(); vis != nil {
					if *vis {
						visBefore = "true"
					} else {
						visBefore = "false"
					}
				}
				fmt.Fprintf(os.Stderr, "[makeParentNodesVisible] Making Media visible: before=%s\n", visBefore)
			}
			v.Node.EnsureVisibility()
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				visAfter := "nil"
				if vis := v.Node.IsVisible(); vis != nil {
					if *vis {
						visAfter = "true"
					} else {
						visAfter = "false"
					}
				}
				fmt.Fprintf(os.Stderr, "[makeParentNodesVisible] Media after EnsureVisibility: visibility=%s\n", visAfter)
			}
			// IMPORTANT: DO NOT call RemoveVisibilityBlock() here!
			// The Media node needs to keep BlocksVisibility() == true so that
			// ToCSSVisitor.ResolveVisibility() takes the correct code path that
			// filters children with KeepOnlyVisibleChilds() before checking if empty.

		case *AtRule:
			// Make the AtRule node visible (this covers @supports, @keyframes, etc.)
			v.Node.EnsureVisibility()
			// IMPORTANT: DO NOT call RemoveVisibilityBlock() here!
			// The AtRule node needs to keep BlocksVisibility() == true so that
			// ToCSSVisitor.ResolveVisibility() takes the correct code path.

		default:
		}
	}
}

// NewExtendVisitor creates a new extend visitor (alias for NewProcessExtendsVisitor)
func NewExtendVisitor() *ProcessExtendsVisitor {
	return NewProcessExtendsVisitor()
}

// Default export equivalent
var Default = NewProcessExtendsVisitor