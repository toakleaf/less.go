package less_go

import (
	"fmt"
	"os"
)

// CSSVisitorUtils provides utility functions for CSS visitor
type CSSVisitorUtils struct {
	visitor *Visitor
	context any
}

// NewCSSVisitorUtils creates a new CSSVisitorUtils instance
func NewCSSVisitorUtils(context any) *CSSVisitorUtils {
	utils := &CSSVisitorUtils{
		context: context,
	}
	utils.visitor = NewVisitor(utils)
	return utils
}

// Reset resets the CSSVisitorUtils for reuse from the pool.
// The visitor's methodLookup map is preserved (it's expensive to rebuild).
func (u *CSSVisitorUtils) Reset(context any) {
	u.context = context
	// Note: u.visitor is preserved - its methodLookup is reused
}

// ContainsSilentNonBlockedChild checks if body rules contain silent non-blocked children
func (u *CSSVisitorUtils) ContainsSilentNonBlockedChild(bodyRules []any) bool {
	if bodyRules == nil {
		return false
	}
	
	for _, rule := range bodyRules {
		if silentRule, hasSilent := rule.(interface{ IsSilent(any) bool }); hasSilent {
			if blockedRule, hasBlocked := rule.(interface{ BlocksVisibility() bool }); hasBlocked {
				if silentRule.IsSilent(u.context) && !blockedRule.BlocksVisibility() {
					// the atrule contains something that was referenced (likely by extend)
					// therefore it needs to be shown in output too
					return true
				}
			}
		}
	}
	return false
}

// KeepOnlyVisibleChilds filters out invisible children from owner
func (u *CSSVisitorUtils) KeepOnlyVisibleChilds(owner any) {
	if owner == nil {
		return
	}

	// Try to access rules field
	if ownerWithRules, ok := owner.(interface{ GetRules() []any; SetRules([]any) }); ok {
		rules := ownerWithRules.GetRules()
		if rules != nil {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[KeepOnlyVisibleChilds] Processing %d rules\n", len(rules))
			}
			var visibleRules []any
			for _, rule := range rules {
				// Try to get IsVisible() value - handle different ways a node might have this
				var vis *bool
				var hasVisibility bool

				// Try direct method call
				if visibleRule, ok := rule.(interface{ IsVisible() *bool }); ok {
					vis = visibleRule.IsVisible()
					hasVisibility = true
				} else if decl, ok := rule.(*Declaration); ok && decl.Node != nil {
					// Declaration embeds *Node, but type assertion might not work, so access directly
					vis = decl.Node.IsVisible()
					hasVisibility = true
				}

				if hasVisibility {
					// Match JavaScript: isVisible() returns nodeVisible, which when undefined is falsy
					// In JavaScript, the filter is: rules.filter(thing => thing.isVisible())
					// This means: keep only if isVisible() is truthy (non-nil and true)
					// undefined (nil) or false = filter out
					if vis != nil && *vis {
						visibleRules = append(visibleRules, rule)
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							ruleType := fmt.Sprintf("%T", rule)
							ruleSel := "?"
							if rs, ok := rule.(*Ruleset); ok && len(rs.Paths) > 0 && len(rs.Paths[0]) > 0 {
								if sel, ok := rs.Paths[0][0].(*Selector); ok {
									ruleSel = sel.ToCSS(nil)
								}
							}
							fmt.Fprintf(os.Stderr, "[KeepOnlyVisibleChilds]   KEPT: type=%s selector=%s visibility=true\n", ruleType, ruleSel)
						}
					} else {
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							ruleType := fmt.Sprintf("%T", rule)
							ruleSel := "?"
							if rs, ok := rule.(*Ruleset); ok && len(rs.Paths) > 0 && len(rs.Paths[0]) > 0 {
								if sel, ok := rs.Paths[0][0].(*Selector); ok {
									ruleSel = sel.ToCSS(nil)
								}
							}
							visStr := "nil"
							if vis != nil {
								if *vis {
									visStr = "true"
								} else {
									visStr = "false"
								}
							}
							fmt.Fprintf(os.Stderr, "[KeepOnlyVisibleChilds]   FILTERED: type=%s selector=%s visibility=%s\n", ruleType, ruleSel, visStr)
						}
					}
				} else {
					// If rule doesn't have IsVisible method, keep it (for primitives, etc.)
					visibleRules = append(visibleRules, rule)
				}
			}
			ownerWithRules.SetRules(visibleRules)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[KeepOnlyVisibleChilds] After filtering: %d rules remain\n", len(visibleRules))
			}
		}
	}
}

// IsEmpty checks if owner is empty
func (u *CSSVisitorUtils) IsEmpty(owner any) bool {
	if owner == nil {
		return true
	}

	if ownerWithRules, ok := owner.(interface{ GetRules() []any }); ok {
		rules := ownerWithRules.GetRules()
		isEmpty := rules == nil || len(rules) == 0
		if !isEmpty {
			// Check if all rules are nil, block visibility (and not explicitly visible), or are variable declarations
			// A ruleset is considered empty if it contains only invisible content
			allInsignificant := true
			for _, r := range rules {
				if r == nil {
					continue
				}
				// Check if rule blocks visibility (reference imports)
				// BUT if the rule has been explicitly made visible (IsVisible() == true),
				// it's significant even if it blocks visibility
				if blocksNode, ok := r.(interface{ BlocksVisibility() bool; IsVisible() *bool }); ok {
					if blocksNode.BlocksVisibility() {
						// Check if explicitly visible (e.g., via extend)
						vis := blocksNode.IsVisible()
						if vis == nil || !*vis {
							// Not explicitly visible - skip
							continue
						}
						// Explicitly visible - this is significant
						allInsignificant = false
						break
					}
				} else if blocksNode, ok := r.(interface{ BlocksVisibility() bool }); ok {
					if blocksNode.BlocksVisibility() {
						continue
					}
				}
				// Check if rule is a variable declaration
				if declNode, ok := r.(interface{ GetVariable() bool }); ok {
					if declNode.GetVariable() {
						continue
					}
				}
				// Found a significant rule (non-nil, visible, non-variable)
				allInsignificant = false
				break
			}
			if allInsignificant {
				isEmpty = true
			}
		}
		return isEmpty
	}

	return true
}

// HasVisibleSelector checks if ruleset node has visible selectors
func (u *CSSVisitorUtils) HasVisibleSelector(rulesetNode any) bool {
	if rulesetNode == nil {
		return false
	}
	
	if nodeWithPaths, ok := rulesetNode.(interface{ GetPaths() []any }); ok {
		paths := nodeWithPaths.GetPaths()
		return paths != nil && len(paths) > 0
	}
	
	return false
}

// ResolveVisibilityMedia resolves visibility for Media nodes.
// Media nodes are special because VisitRuleset extracts nested rulesets
// and places them as direct children of Media (m.Rules[1...]), not as
// grandchildren through the wrapper ruleset (m.Rules[0]).
// This function filters all direct children of Media, not just m.Rules[0].Rules.
func (u *CSSVisitorUtils) ResolveVisibilityMedia(node any) any {
	if node == nil {
		return nil
	}

	// Check if node blocks visibility (reference import)
	blocksVis := false
	if blockedNode, hasBlocked := node.(interface{ BlocksVisibility() bool }); hasBlocked {
		blocksVis = blockedNode.BlocksVisibility()
	}

	if !blocksVis {
		// Non-reference import: just check if empty
		if u.IsEmpty(node) {
			return nil
		}
		return node
	}

	// Reference import: filter ALL direct children of Media to keep only visible ones
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		nodeType := fmt.Sprintf("%T", node)
		features := ""
		if m, ok := node.(*Media); ok && m.Features != nil {
			if feat, ok := m.Features.(interface{ ToCSS(any) string }); ok {
				features = feat.ToCSS(nil)
			}
		}
		fmt.Fprintf(os.Stderr, "[ResolveVisibilityMedia] Processing blocked Media type=%s features=%s\n", nodeType, features)
	}

	// Apply KeepOnlyVisibleChilds directly to the Media node's rules
	// This filters m.Rules to keep only visible rulesets (including extracted ones)
	u.KeepOnlyVisibleChilds(node)

	// Check if Media is empty after filtering
	if u.IsEmpty(node) {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ResolveVisibilityMedia] Media is empty after filtering, returning nil\n")
		}
		return nil
	}

	// Media has visible content - make it visible and remove visibility block
	if ensureVisNode, hasEnsure := node.(interface{ EnsureVisibility() }); hasEnsure {
		ensureVisNode.EnsureVisibility()
	}
	if removeVisNode, hasRemove := node.(interface{ RemoveVisibilityBlock() }); hasRemove {
		removeVisNode.RemoveVisibilityBlock()
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[ResolveVisibilityMedia] Returning Media (not nil)\n")
	}
	return node
}

// ResolveVisibility resolves visibility for a node
func (u *CSSVisitorUtils) ResolveVisibility(node any) any {
	if node == nil {
		return nil
	}

	if blockedNode, hasBlocked := node.(interface{ BlocksVisibility() bool }); hasBlocked {
		if !blockedNode.BlocksVisibility() {
			isEmpty := u.IsEmpty(node)
			if isEmpty {
				return nil
			}
			return node
		}
	} else {
		// If node doesn't have BlocksVisibility method, treat as not blocked
		isEmpty := u.IsEmpty(node)
		if isEmpty {
			return nil
		}
		return node
	}

	// Node blocks visibility, process it
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		nodeType := fmt.Sprintf("%T", node)
		features := ""
		if m, ok := node.(*Media); ok && m.Features != nil {
			if feat, ok := m.Features.(interface{ ToCSS(any) string }); ok {
				features = feat.ToCSS(nil)
			}
		}
		if ar, ok := node.(*AtRule); ok {
			features = ar.Name
		}
		fmt.Fprintf(os.Stderr, "[ResolveVisibility] Processing blocked node type=%s features=%s\n", nodeType, features)
	}

	if nodeWithRules, ok := node.(interface{ GetRules() []any }); ok {
		rules := nodeWithRules.GetRules()
		if len(rules) > 0 {
			compiledRulesBody := rules[0]
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[ResolveVisibility] Calling KeepOnlyVisibleChilds on compiledRulesBody type=%T\n", compiledRulesBody)
			}
			u.KeepOnlyVisibleChilds(compiledRulesBody)

			if u.IsEmpty(compiledRulesBody) {
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[ResolveVisibility] compiledRulesBody is empty after filtering, returning nil\n")
				}
				return nil
			}

			// Match JavaScript: node.ensureVisibility(); node.removeVisibilityBlock();
			// When a node from a reference import has visible children (e.g., from mixin calls/extends),
			// we make the node visible and remove the visibility block so it outputs
			if ensureVisNode, hasEnsure := node.(interface{ EnsureVisibility() }); hasEnsure {
				ensureVisNode.EnsureVisibility()
			}
			if removeVisNode, hasRemove := node.(interface{ RemoveVisibilityBlock() }); hasRemove {
				removeVisNode.RemoveVisibilityBlock()
			}

			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[ResolveVisibility] Returning node (not nil)\n")
			}
			return node
		}
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[ResolveVisibility] Node has no rules or empty rules, returning nil\n")
	}
	return nil
}

// IsVisibleRuleset checks if a ruleset is visible
func (u *CSSVisitorUtils) IsVisibleRuleset(rulesetNode any) bool {
	if rulesetNode == nil {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[IsVisibleRuleset] Returning false - rulesetNode is nil\n")
		}
		return false
	}

	// Debug: trace div rulesets
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" {
		if rs, ok := rulesetNode.(*Ruleset); ok && len(rs.Selectors) > 0 {
			if sel, ok := rs.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
				// Check if first element value contains "div" in any form
				elemVal := sel.Elements[0].Value
				var valStr string
				if str, ok := elemVal.(string); ok {
					valStr = str
				} else {
					valStr = fmt.Sprintf("%v", elemVal)
				}
				if valStr == "div" || fmt.Sprintf("%T", elemVal) == "*less_go.Keyword" {
					hasVisibleSel := u.HasVisibleSelector(rulesetNode)
					fmt.Fprintf(os.Stderr, "[IsVisibleRuleset ENTRY] ruleset=%p, elemVal=%v (%T), BlocksVis=%v, HasVisibleSelector=%v, Paths=%d\n",
						rs, elemVal, elemVal, rs.Node.BlocksVisibility(), hasVisibleSel, len(rs.Paths))
				}
			}
		}
	}

	if firstRootNode, ok := rulesetNode.(interface{ GetFirstRoot() bool }); ok {
		if firstRootNode.GetFirstRoot() {
			return true
		}
	}

	if u.IsEmpty(rulesetNode) {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			if rs, ok := rulesetNode.(*Ruleset); ok && rs.MultiMedia {
				fmt.Fprintf(os.Stderr, "[IsVisibleRuleset] MultiMedia Ruleset is empty (Rules=%d), returning false\n", len(rs.Rules))
			}
		}
		return false
	}

	// Check if the ruleset blocks visibility and has undefined nodeVisible
	// This indicates it's from a referenced import and hasn't been explicitly used
	if blockedNode, ok := rulesetNode.(interface{ BlocksVisibility() bool; IsVisible() *bool }); ok {
		if blockedNode.BlocksVisibility() {
			vis := blockedNode.IsVisible()
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				var sel string = "?"
				if rs, ok := rulesetNode.(*Ruleset); ok && len(rs.Paths) > 0 && len(rs.Paths[0]) > 0 {
					if s, ok := rs.Paths[0][0].(*Selector); ok {
						sel = s.ToCSS(nil)
					}
				}
				visVal := "nil"
				if vis != nil {
					if *vis {
						visVal = "true"
					} else {
						visVal = "false"
					}
				}
				fmt.Fprintf(os.Stderr, "[IsVisibleRuleset] selector=%s blocksVisibility=true visibility=%s\n", sel, visVal)
			}
			// If visibility is undefined (nil) or explicitly false, check for visible paths
			// before hiding the ruleset. Extends can add visible selectors to paths, and
			// compileRulesetPaths has already filtered to only visible paths at this point.
			if vis == nil || !*vis {
				// CRITICAL FIX: Check if the ruleset has visible paths from extends
				// compileRulesetPaths has already run and filtered paths to only those
				// with visible selectors. If there are paths remaining, the ruleset
				// should be visible because those paths came from extend matching.
				if u.HasVisibleSelector(rulesetNode) {
					// Ruleset has visible paths from extends - make it visible
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Fprintf(os.Stderr, "[IsVisibleRuleset] Keeping ruleset with visible paths from extends\n")
					}
					// Continue to other checks instead of returning false
				} else {
					return false
				}
			}
		}
	}

	// Special case: rulesets that are direct children of Media/AtRule nodes
	// These rulesets have AllowImports == true and don't need visible selectors
	// because their parent node provides the selector context
	if allowImportsNode, ok := rulesetNode.(interface{ GetAllowImports() bool }); ok {
		if allowImportsNode.GetAllowImports() {
			// This is a wrapper ruleset inside a Media/AtRule node
			// Keep it even if it has no visible selectors
			return true
		}
	}

	// Special case: MultiMedia rulesets should always be visible
	// They contain merged media queries and don't need selectors
	if multiMediaNode, ok := rulesetNode.(*Ruleset); ok && multiMediaNode.MultiMedia {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[IsVisibleRuleset] MultiMedia Ruleset - returning true\n")
		}
		return true
	}

	if rootNode, ok := rulesetNode.(interface{ GetRoot() bool }); ok {
		if !rootNode.GetRoot() && !u.HasVisibleSelector(rulesetNode) {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if rs, ok := rulesetNode.(*Ruleset); ok && rs.MultiMedia {
					fmt.Fprintf(os.Stderr, "[IsVisibleRuleset] MultiMedia Ruleset - has root=%v, hasVisibleSelector=%v\n",
						rootNode.GetRoot(), u.HasVisibleSelector(rulesetNode))
				}
			}
			return false
		}
	}

	return true
}

// ToCSSVisitor implements CSS output visitor
type ToCSSVisitor struct {
	visitor    *Visitor
	context    any
	utils      *CSSVisitorUtils
	charset    bool
	isReplacing bool
}

// NewToCSSVisitor creates a new ToCSSVisitor
func NewToCSSVisitor(context any) *ToCSSVisitor {
	v := &ToCSSVisitor{
		context:     context,
		utils:       NewCSSVisitorUtils(context),
		charset:     false,
		isReplacing: true,
	}
	v.visitor = NewVisitor(v)
	return v
}

// Reset resets the ToCSSVisitor for reuse from the pool.
// The visitor's methodLookup map is preserved (it's expensive to rebuild).
func (v *ToCSSVisitor) Reset(context any) {
	v.context = context
	v.charset = false
	v.isReplacing = true
	// Reset the utils with the new context
	if v.utils != nil {
		v.utils.Reset(context)
	}
	// Note: v.visitor is preserved - its methodLookup is reused
}

// Run runs the visitor on the root node
func (v *ToCSSVisitor) Run(root any) any {
	return v.visitor.Visit(root)
}

// IsReplacing returns true as ToCSSVisitor is a replacing visitor
func (v *ToCSSVisitor) IsReplacing() bool {
	return true
}

// VisitNode implements direct dispatch without reflection for better performance
func (v *ToCSSVisitor) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {
	switch n := node.(type) {
	case *Anonymous:
		return v.VisitAnonymous(n, visitArgs), true
	case *AtRule:
		return v.VisitAtRule(n, visitArgs), true
	case *Comment:
		return v.VisitComment(n, visitArgs), true
	case *Container:
		return v.VisitContainer(n, visitArgs), true
	case *Declaration:
		return v.VisitDeclaration(n, visitArgs), true
	case *Extend:
		return v.VisitExtend(n, visitArgs), true
	case *Import:
		return v.VisitImport(n, visitArgs), true
	case *Media:
		return v.VisitMedia(n, visitArgs), true
	case *MixinDefinition:
		return v.VisitMixinDefinition(n, visitArgs), true
	case *Ruleset:
		return v.VisitRuleset(n, visitArgs), true
	default:
		_ = n
		return node, true // Node type handled (no-op, avoids reflection)
	}
}

// VisitNodeOut implements direct dispatch for visitOut methods
func (v *ToCSSVisitor) VisitNodeOut(node any) bool {
	return true // No VisitOut methods, handled as no-op (avoids reflection)
}

// containsOnlyProperties checks if rules contain only properties (no nested rulesets)
func (v *ToCSSVisitor) containsOnlyProperties(rules []any) bool {
	if len(rules) == 0 {
		return false
	}
	
	for _, rule := range rules {
		if ruleNode, ok := rule.(interface{ GetType() string }); ok {
			nodeType := ruleNode.GetType()
			// If we find anything that's not a Declaration, it's not "only properties"
			if nodeType != "Declaration" && nodeType != "Comment" {
				return false
			}
		}
	}
	
	// Check that we have at least one non-variable declaration
	for _, rule := range rules {
		if declNode, ok := rule.(interface{ GetType() string; GetVariable() bool }); ok {
			if declNode.GetType() == "Declaration" && !declNode.GetVariable() {
				return true
			}
		}
	}
	
	return false
}

// VisitDeclaration visits a declaration node
func (v *ToCSSVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) any {
	if declNode == nil {
		return nil
	}
	
	if blockedNode, hasBlocked := declNode.(interface{ BlocksVisibility() bool }); hasBlocked {
		if blockedNode.BlocksVisibility() {
			return nil
		}
	}
	
	if varNode, hasVar := declNode.(interface{ GetVariable() bool }); hasVar {
		if varNode.GetVariable() {
			return nil
		}
	}
	
	return declNode
}

// VisitMixinDefinition visits a mixin definition node
func (v *ToCSSVisitor) VisitMixinDefinition(mixinNode any, visitArgs *VisitArgs) any {
	// mixin definitions do not get eval'd - this means they keep state
	// so we have to clear that state here so it isn't used if toCSS is called twice
	if framesNode, hasFrames := mixinNode.(interface{ SetFrames([]any) }); hasFrames {
		framesNode.SetFrames([]any{})
	}
	// Don't visit nested rules inside mixin definitions - they should only be output when the mixin is called
	// Match JoinSelectorVisitor behavior
	visitArgs.VisitDeeper = false
	return nil
}

// VisitExtend visits an extend node
func (v *ToCSSVisitor) VisitExtend(extendNode any, visitArgs *VisitArgs) any {
	return nil
}

// VisitComment visits a comment node
func (v *ToCSSVisitor) VisitComment(commentNode any, visitArgs *VisitArgs) any {
	if commentNode == nil {
		return nil
	}

	if blockedNode, hasBlocked := commentNode.(interface{ BlocksVisibility() bool }); hasBlocked {
		if blockedNode.BlocksVisibility() {
			return nil
		}
	}

	if silentNode, hasSilent := commentNode.(interface{ IsSilent(any) bool }); hasSilent {
		if silentNode.IsSilent(v.context) {
			return nil
		}
	}

	return commentNode
}

// VisitMedia visits a media node
func (v *ToCSSVisitor) VisitMedia(mediaNode any, visitArgs *VisitArgs) any {
	if mediaNode == nil {
		return nil
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		if m, ok := mediaNode.(*Media); ok {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitMedia] Media=%p, Before Accept, Rules count=%d\n", m, len(m.Rules))
			if len(m.Rules) > 0 {
				if innerRs, ok := m.Rules[0].(*Ruleset); ok {
					fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitMedia]   inner ruleset=%p, Rules count=%d\n", innerRs, len(innerRs.Rules))
				}
			}
		}
	}

	if acceptor, ok := mediaNode.(interface{ Accept(any) }); ok {
		acceptor.Accept(v.visitor)
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		if m, ok := mediaNode.(*Media); ok {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitMedia] After Accept, Rules count=%d\n", len(m.Rules))
			for i, rule := range m.Rules {
				ruleType := fmt.Sprintf("%T", rule)
				ruleSel := "?"
				if rs, ok := rule.(*Ruleset); ok && len(rs.Paths) > 0 && len(rs.Paths[0]) > 0 {
					if sel, ok := rs.Paths[0][0].(*Selector); ok {
						ruleSel = sel.ToCSS(nil)
					}
				}
				visStr := "?"
				if visNode, ok := rule.(interface{ IsVisible() *bool }); ok {
					if vis := visNode.IsVisible(); vis != nil {
						if *vis {
							visStr = "true"
						} else {
							visStr = "false"
						}
					} else {
						visStr = "nil"
					}
				}
				fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitMedia]   rule[%d]: type=%s selector=%s visibility=%s\n", i, ruleType, ruleSel, visStr)
			}
		}
	}
	visitArgs.VisitDeeper = false

	return v.utils.ResolveVisibilityMedia(mediaNode)
}

// VisitContainer visits a container node (same logic as media)
func (v *ToCSSVisitor) VisitContainer(containerNode any, visitArgs *VisitArgs) any {
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitContainer] Called\n")
	}

	if containerNode == nil {
		return nil
	}

	if acceptor, ok := containerNode.(interface{ Accept(any) }); ok {
		acceptor.Accept(v.visitor)
	}
	visitArgs.VisitDeeper = false

	result := v.utils.ResolveVisibility(containerNode)
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitContainer] Result: %v (type=%T)\n", result != nil, result)
	}
	return result
}

// VisitImport visits an import node
func (v *ToCSSVisitor) VisitImport(importNode any, visitArgs *VisitArgs) any {
	if importNode == nil {
		return nil
	}
	
	if blockedNode, hasBlocked := importNode.(interface{ BlocksVisibility() bool }); hasBlocked {
		if blockedNode.BlocksVisibility() {
			return nil
		}
	}
	
	return importNode
}

// VisitAtRule visits an at-rule node
func (v *ToCSSVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) any {
	if atRuleNode == nil {
		return nil
	}

	// Check for SimpleBlock AtRules (like @starting-style with only declarations)
	// These use Declarations instead of Rules, so we need to handle them specially
	if at, ok := atRuleNode.(*AtRule); ok && at.SimpleBlock && len(at.Declarations) > 0 {
		// Process children (this will visit the declarations)
		if acceptor, ok := atRuleNode.(interface{ Accept(any) }); ok {
			acceptor.Accept(v.visitor)
		}
		visitArgs.VisitDeeper = false

		// Run merge processing on declarations
		at.Declarations = v.mergeRules(at.Declarations)

		// Check visibility - but don't use ResolveVisibility since it checks GetRules()
		// which returns nil for SimpleBlock AtRules (they use Declarations instead)
		if at.Node != nil && at.Node.BlocksVisibility() {
			vis := at.Node.IsVisible()
			if vis == nil || !*vis {
				return nil
			}
		}

		// Check if all declarations were filtered out
		if len(at.Declarations) == 0 {
			return nil
		}

		return atRuleNode
	}

	if nodeWithRules, ok := atRuleNode.(interface{ GetRules() []any }); ok {
		rules := nodeWithRules.GetRules()
		if rules != nil && len(rules) > 0 {
			return v.VisitAtRuleWithBody(atRuleNode, visitArgs)
		}
	}
	return v.VisitAtRuleWithoutBody(atRuleNode, visitArgs)
}

// VisitAnonymous visits an anonymous node
func (v *ToCSSVisitor) VisitAnonymous(anonymousNode any, visitArgs *VisitArgs) any {
	if anonymousNode == nil {
		return nil
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		anonVal := "?"
		if anon, ok := anonymousNode.(*Anonymous); ok {
			anonVal = fmt.Sprintf("[type=%T]", anon.Value)
			if str, ok := anon.Value.(string); ok {
				if len(str) > 30 {
					anonVal = str[:30] + "..."
				} else {
					anonVal = str
				}
			}
		}
		fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitAnonymous] Processing Anonymous value=%q\n", anonVal)
	}

	if blockedNode, hasBlocked := anonymousNode.(interface{ BlocksVisibility() bool }); hasBlocked {
		blocksVis := blockedNode.BlocksVisibility()
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitAnonymous] hasBlocked=true, blocksVisibility=%v\n", blocksVis)
		}
		if !blocksVis {
			if acceptor, ok := anonymousNode.(interface{ Accept(any) }); ok {
				acceptor.Accept(v.visitor)
			}
			return anonymousNode
		}
		// Blocked - return nil to filter out
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitAnonymous] FILTERED - returning nil\n")
		}
	} else {
		// If node doesn't have BlocksVisibility method, treat as not blocked
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitAnonymous] hasBlocked=false, returning node\n")
		}
		if acceptor, ok := anonymousNode.(interface{ Accept(any) }); ok {
			acceptor.Accept(v.visitor)
		}
		return anonymousNode
	}

	return nil
}

// VisitAtRuleWithBody visits an at-rule with body
func (v *ToCSSVisitor) VisitAtRuleWithBody(atRuleNode any, visitArgs *VisitArgs) any {
	if atRuleNode == nil {
		return nil
	}

	// Process children
	if acceptor, ok := atRuleNode.(interface{ Accept(any) }); ok {
		acceptor.Accept(v.visitor)
	}
	visitArgs.VisitDeeper = false

	if !v.utils.IsEmpty(atRuleNode) {
		if nodeWithRules, ok := atRuleNode.(interface{ GetRules() []any }); ok {
			rules := nodeWithRules.GetRules()
			if len(rules) > 0 {
				if firstRule, ok := rules[0].(interface{ GetRules() []any }); ok {
					v.mergeRules(firstRule.GetRules())
				}
			}
		}
	}

	// Use the same approach as Media for AtRule nodes with bodies
	// After VisitRuleset extracts nested rulesets, they're direct children of AtRule
	return v.utils.ResolveVisibilityMedia(atRuleNode)
}

// VisitAtRuleWithoutBody visits an at-rule without body
func (v *ToCSSVisitor) VisitAtRuleWithoutBody(atRuleNode any, visitArgs *VisitArgs) any {
	if atRuleNode == nil {
		return nil
	}

	// Check if the at-rule blocks visibility and has undefined/false nodeVisible
	// This filters out at-rules from referenced imports that haven't been explicitly used
	if blockedNode, ok := atRuleNode.(interface{ BlocksVisibility() bool; IsVisible() *bool }); ok {
		if blockedNode.BlocksVisibility() {
			vis := blockedNode.IsVisible()
			// If visibility is undefined (nil) or explicitly false, hide the at-rule
			if vis == nil || !*vis {
				return nil
			}
		}
	}
	
	if nameNode, hasName := atRuleNode.(interface{ GetName() string }); hasName {
		if nameNode.GetName() == "@charset" {
			// CSS spec: only the first @charset declaration should be output
			// Any subsequent @charset declarations should be completely removed (not even as comments)
			if v.charset {
				// Already seen a @charset, skip this one entirely
				return nil
			}
			v.charset = true
		}
	}
	
	return atRuleNode
}

// CheckValidNodes checks if nodes are valid for their context
func (v *ToCSSVisitor) CheckValidNodes(rules []any, isRoot bool) error {
	if rules == nil {
		return nil
	}
	
	for _, ruleNode := range rules {
		if isRoot {
			// Check for direct declarations
			if declNode, ok := ruleNode.(interface{ GetType() string; GetVariable() bool; GetIndex() int; FileInfo() map[string]any }); ok {
				if declNode.GetType() == "Declaration" && !declNode.GetVariable() {
					var filename string
					if fileInfo := declNode.FileInfo(); fileInfo != nil {
						if fileNameValue, ok := fileInfo["filename"]; ok {
							if fileNameStr, ok := fileNameValue.(string); ok {
								filename = fileNameStr
							}
						}
					}
					return &LessError{
						Message:  "Properties must be inside selector blocks. They cannot be in the root",
						Index:    declNode.GetIndex(),
						Filename: filename,
					}
				}
			}
			
			// Check for rulesets with no selectors at root level - their contents should also be checked
			if rulesetNode, ok := ruleNode.(interface{ GetType() string; GetSelectors() []any; GetRules() []any }); ok {
				if rulesetNode.GetType() == "Ruleset" && (rulesetNode.GetSelectors() == nil || len(rulesetNode.GetSelectors()) == 0) {
					// A ruleset with no selectors at root level - check its rules as if they were at root
					if err := v.CheckValidNodes(rulesetNode.GetRules(), true); err != nil {
						return err
					}
				}
			}
		}
		
		if callNode, ok := ruleNode.(interface{ GetType() string; GetName() string; GetIndex() int; FileInfo() map[string]any }); ok {
			if callNode.GetType() == "Call" {
				var filename string
				if fileInfo := callNode.FileInfo(); fileInfo != nil {
					if fileNameValue, ok := fileInfo["filename"]; ok {
						if fileNameStr, ok := fileNameValue.(string); ok {
							filename = fileNameStr
						}
					}
				}
				return &LessError{
					Message:  "Function '" + callNode.GetName() + "' did not return a root node",
					Index:    callNode.GetIndex(),
					Filename: filename,
				}
			}
		}
		
		if typeNode, ok := ruleNode.(interface{ GetType() string; GetAllowRoot() bool; GetIndex() int; FileInfo() map[string]any }); ok {
			if typeNode.GetType() != "" && !typeNode.GetAllowRoot() {
				var filename string
				if fileInfo := typeNode.FileInfo(); fileInfo != nil {
					if fileNameValue, ok := fileInfo["filename"]; ok {
						if fileNameStr, ok := fileNameValue.(string); ok {
							filename = fileNameStr
						}
					}
				}
				return &LessError{
					Message:  typeNode.GetType() + " node returned by a function is not valid here",
					Index:    typeNode.GetIndex(),
					Filename: filename,
				}
			}
		}
	}
	
	return nil
}

// VisitRuleset visits a ruleset node
func (v *ToCSSVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) any {
	if rulesetNode == nil {
		return nil
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		if rs, ok := rulesetNode.(*Ruleset); ok {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset ENTRY] ruleset=%p, Rules count=%d, Root=%v\n", rs, len(rs.Rules), rs.Root)
			// Only log first 10 rules to avoid flooding
			maxLog := 10
			if len(rs.Rules) < maxLog {
				maxLog = len(rs.Rules)
			}
			for i := 0; i < maxLog; i++ {
				rule := rs.Rules[i]
				if anon, ok := rule.(*Anonymous); ok {
					anonVal := fmt.Sprintf("[type=%T]", anon.Value)
					if str, ok := anon.Value.(string); ok {
						if len(str) > 30 {
							anonVal = str[:30] + "..."
						} else {
							anonVal = str
						}
					}
					fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset]   rule[%d] Anonymous value=%q blocksVis=%v\n", i, anonVal, anon.BlocksVisibility())
				} else {
					fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset]   rule[%d] type=%T\n", i, rule)
				}
			}
			if len(rs.Rules) > maxLog {
				fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset]   ... and %d more rules\n", len(rs.Rules)-maxLog)
			}
		}
	}

	var rulesets []any
	
	
	// Check valid nodes
	if nodeWithRules, ok := rulesetNode.(interface{ GetRules() []any }); ok {
		rules := nodeWithRules.GetRules()
		var isFirstRoot bool
		if firstRootNode, ok := rulesetNode.(interface{ GetFirstRoot() bool }); ok {
			isFirstRoot = firstRootNode.GetFirstRoot()
		}

		// Special check: if this is the file root and contains selector-less rulesets
		// with only properties, that would be invalid CSS
		// Note: We only check isFirstRoot here, matching JavaScript behavior exactly.
		// Do NOT use isRulesetAtRoot() which incorrectly returns true for @font-face/@keyframes inner rulesets
		if isFirstRoot {
			if selNode, ok := rulesetNode.(interface{ GetSelectors() []any }); ok {
				selectors := selNode.GetSelectors()
				if len(selectors) == 0 && v.containsOnlyProperties(rules) {
					// This is a selector-less ruleset at root with only properties
					// Find the first property to report its location
					for _, rule := range rules {
						if declNode, ok := rule.(interface{ GetType() string; GetVariable() bool; GetIndex() int; FileInfo() map[string]any }); ok {
							if declNode.GetType() == "Declaration" && !declNode.GetVariable() {
								var filename string
								if fileInfo := declNode.FileInfo(); fileInfo != nil {
									if fileNameValue, ok := fileInfo["filename"]; ok {
										if fileNameStr, ok := fileNameValue.(string); ok {
											filename = fileNameStr
										}
									}
								}
								panic(&LessError{
									Message:  "Properties must be inside selector blocks. They cannot be in the root",
									Index:    declNode.GetIndex(),
									Filename: filename,
								})
							}
						}
					}
				}
			}
		}

		if err := v.CheckValidNodes(rules, isFirstRoot); err != nil {
			panic(err) // Matches JS behavior of throwing errors
		}
	}
	
	if rootNode, ok := rulesetNode.(interface{ GetRoot() bool }); ok {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			if rs, ok := rulesetNode.(*Ruleset); ok {
				fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset] Ruleset root=%v, MultiMedia=%v, Rules=%d\n",
					rootNode.GetRoot(), rs.MultiMedia, len(rs.Rules))
			}
		}
		if !rootNode.GetRoot() {
			// remove invisible paths and clean up combinators
			v.compileRulesetPaths(rulesetNode)
			
			// remove rulesets from this ruleset body and compile them separately
			if nodeWithRules, ok := rulesetNode.(interface{ GetRules() []any; SetRules([]any) }); ok {
				nodeRules := nodeWithRules.GetRules()
				
				if nodeRules != nil {
					nodeRuleCnt := len(nodeRules)
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Fprintf(os.Stderr, "[VisitRuleset] Processing %d rules\n", nodeRuleCnt)
					}
					for i := 0; i < nodeRuleCnt; {
						rule := nodeRules[i]
						if ruleWithRules, ok := rule.(interface{ GetRules() []any }); ok {
							if ruleWithRules.GetRules() != nil {
								// Check if this is a @starting-style at-rule
								// @starting-style should stay nested inside its parent ruleset
								// (like CSS native nesting) and NOT bubble out like @media does
								isStartingStyle := false
								if atRule, ok := rule.(*AtRule); ok && atRule.Name == "@starting-style" {
									isStartingStyle = true
								}

								if !isStartingStyle {
									if os.Getenv("LESS_GO_DEBUG") == "1" {
										fmt.Fprintf(os.Stderr, "[VisitRuleset] Extracting child ruleset at index %d\n", i)
									}
									// visit because we are moving them out from being a child
									rulesets = append(rulesets, v.visitor.Visit(rule))
									// Remove from nodeRules
									nodeRules = append(nodeRules[:i], nodeRules[i+1:]...)
									nodeRuleCnt--
									continue
								}
							}
						}
						i++
					}
					
					// accept the visitor to remove rules and refactor itself
					// then we can decide now whether we want it or not
					// compile body
					if nodeRuleCnt > 0 {
						nodeWithRules.SetRules(nodeRules)
						if acceptor, ok := rulesetNode.(interface{ Accept(any) }); ok {
							acceptor.Accept(v.visitor)
						}
					} else {
						nodeWithRules.SetRules(nil)
					}
				}
			}
			visitArgs.VisitDeeper = false
		} else {
			// if (! rulesetNode.root) {
			// For the root ruleset, we need to clean up paths of its direct children
			// This ensures top-level rulesets don't have extra space combinators
			if acceptor, ok := rulesetNode.(interface{ Accept(any) }); ok {
				acceptor.Accept(v.visitor)
			}
			visitArgs.VisitDeeper = false
		}
	}
	
	if nodeWithRules, ok := rulesetNode.(interface{ GetRules() []any; SetRules([]any) }); ok {
		rules := nodeWithRules.GetRules()
		if rules != nil {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if rs, ok := rulesetNode.(*Ruleset); ok && rs.MultiMedia {
					fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset] MultiMedia before merge/dedup: Rules=%d\n", len(rules))
				}
			}
			rules = v.mergeRules(rules)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if rs, ok := rulesetNode.(*Ruleset); ok && rs.MultiMedia {
					fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset] MultiMedia after mergeRules: Rules=%d\n", len(rules))
				}
			}
			rules = v.removeDuplicateRules(rules)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if rs, ok := rulesetNode.(*Ruleset); ok && rs.MultiMedia {
					fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset] MultiMedia after removeDuplicateRules: Rules=%d\n", len(rules))
				}
			}
			nodeWithRules.SetRules(rules)
		}
	}
	
	// now decide whether we keep the ruleset
	keepRuleset := v.utils.IsVisibleRuleset(rulesetNode)

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		if rs, ok := rulesetNode.(*Ruleset); ok {
			fmt.Fprintf(os.Stderr, "[ToCSSVisitor.VisitRuleset] ruleset=%p keepRuleset=%v MultiMedia=%v AllowImports=%v Rules=%d Paths=%d Root=%v\n",
				rs, keepRuleset, rs.MultiMedia, rs.AllowImports, len(rs.Rules), len(rs.Paths), rs.Root)
		}
	}
	
	
	
	// Special case: if we extracted nested rulesets and the parent has non-variable declarations,
	// we should keep it even if paths were filtered
	if !keepRuleset && len(rulesets) > 0 {
		if nodeWithRules, ok := rulesetNode.(interface{ GetRules() []any }); ok {
			rules := nodeWithRules.GetRules()
			if rules != nil {
				for _, rule := range rules {
					// Check if it's a non-variable declaration
					if decl, ok := rule.(interface{ GetVariable() bool }); ok {
						if !decl.GetVariable() {
							// Has at least one non-variable declaration
							keepRuleset = true
							break
						}
					}
				}
			}
		}
	}
	
	if keepRuleset {
		// Only mark as visible if it doesn't block visibility (reference imports)
		blocksVis := false
		if blocksNode, hasBlocks := rulesetNode.(interface{ BlocksVisibility() bool }); hasBlocks {
			blocksVis = blocksNode.BlocksVisibility()
		}

		if !blocksVis {
			if ensureVisNode, hasEnsure := rulesetNode.(interface{ EnsureVisibility() }); hasEnsure {
				ensureVisNode.EnsureVisibility()
			}
		}
		// Insert at beginning
		rulesets = append([]any{rulesetNode}, rulesets...)
	}
	

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[VisitRuleset] Returning %d rulesets\n", len(rulesets))
	}
	if len(rulesets) == 1 {
		return rulesets[0]
	}
	return rulesets
}

// compileRulesetPaths compiles paths for a ruleset
func (v *ToCSSVisitor) compileRulesetPaths(rulesetNode any) {
	if pathNode, ok := rulesetNode.(interface{ GetPaths() []any; SetPaths([]any) }); ok {
		paths := pathNode.GetPaths()
		if paths != nil {
			var filteredPaths []any

			for _, p := range paths {
				if pathSlice, ok := p.([]any); ok && len(pathSlice) > 0 {
					// Convert space combinator to empty at start of path
					// pathSlice[0] should be a Selector
					if selector, ok := pathSlice[0].(*Selector); ok && len(selector.Elements) > 0 {
						// Check the first element's combinator
						firstElement := selector.Elements[0]
						if firstElement.Combinator != nil && firstElement.Combinator.Value == " " {
							// Set combinator to empty for top-level selectors
							firstElement.Combinator = NewCombinator("")
						}
					}

					// Check if path has any visible and output selectors
					// In JavaScript: p[i].isVisible() && p[i].getIsOutput()
					// where p is a path array and p[i] is a selector
					//
					// After SetTreeVisibilityVisitor(true) runs, all selectors should have visibility set:
					// - Non-reference selectors: nodeVisible = true
					// - Reference selectors: nodeVisible remains nil/undefined (skipped by SetTreeVisibilityVisitor)
					//
					// When extend creates a new selector from a visible extend:
					// - The new selector gets nodeVisible = true (via ensureVisibility())
					//
					// So: isVisible() returning nil means the selector is from a reference import and hasn't been explicitly made visible
					hasVisibleOutput := false
					for _, selector := range pathSlice {
						// Check if it's a selector with the required methods
						if sel, ok := selector.(*Selector); ok {
							// Check visibility - handle that IsVisible returns *bool
							// nil = from reference import and not explicitly made visible
							// true = explicitly visible (either from non-reference or made visible by extend)
							// false = explicitly invisible
							isVisible := false
							if vis := sel.IsVisible(); vis != nil {
								isVisible = *vis
							}

							// Check output status
							isOutput := sel.GetIsOutput()

							if isVisible && isOutput {
								hasVisibleOutput = true
								break
							}
						}
					}

					if hasVisibleOutput {
						filteredPaths = append(filteredPaths, p)
					}
				}
			}
			
			// If no paths passed the filter but the ruleset has non-variable declarations,
			// keep at least one path to ensure the ruleset is output
			if len(filteredPaths) == 0 && len(paths) > 0 {
				// Check if ruleset has non-variable declarations
				hasDeclarations := false
				if nodeWithRules, ok := rulesetNode.(interface{ GetRules() []any }); ok {
					rules := nodeWithRules.GetRules()
					if rules != nil {
						for _, rule := range rules {
							// Check if it's a declaration (not a ruleset)
							if _, isRuleset := rule.(interface{ GetRules() []any }); !isRuleset {
								if decl, ok := rule.(interface{ GetVariable() bool }); ok {
									if !decl.GetVariable() {
										hasDeclarations = true
										break
									}
								}
							}
						}
					}
				}
				// If it has declarations, keep the original paths
				if hasDeclarations {
					filteredPaths = paths
				}
			}
			
			
			pathNode.SetPaths(filteredPaths)
		}
	}
}

// removeDuplicateRules removes duplicate rules
func (v *ToCSSVisitor) removeDuplicateRules(rules []any) []any {
	if rules == nil {
		return rules
	}
	
	// remove duplicates
	ruleCache := make(map[string]any)
	
	for i := len(rules) - 1; i >= 0; i-- {
		rule := rules[i]
		if declNode, ok := rule.(interface{ GetType() string; GetName() string; ToCSS(any) string }); ok {
			if declNode.GetType() == "Declaration" {
				name := declNode.GetName()
				if existing, exists := ruleCache[name]; !exists {
					ruleCache[name] = rule
				} else {
					var ruleList []string
					if existingDecl, ok := existing.(interface{ ToCSS(any) string }); ok {
						existingCSS := existingDecl.ToCSS(v.context)
						ruleList = []string{existingCSS}
					} else if existingList, ok := existing.([]string); ok {
						ruleList = existingList
					}
					
					ruleCSS := declNode.ToCSS(v.context)
					isDuplicate := false
					for _, css := range ruleList {
						if css == ruleCSS {
							isDuplicate = true
							break
						}
					}
					
					if isDuplicate {
						// Remove rule at index i
						rules = append(rules[:i], rules[i+1:]...)
					} else {
						ruleList = append(ruleList, ruleCSS)
						ruleCache[name] = ruleList
					}
				}
			}
		}
	}
	return rules
}

// mergeRules merges rules with merge property
func (v *ToCSSVisitor) mergeRules(rules []any) []any {
	if rules == nil {
		return rules
	}

	groups := make(map[string]*[]any)
	var groupsArr []*[]any

	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if mergeNode, ok := rule.(interface{ GetMerge() any; GetName() string }); ok {
			merge := mergeNode.GetMerge()
			// Check if merge is truthy (not nil, not false, not empty string)
			isTruthy := false
			switch m := merge.(type) {
			case bool:
				isTruthy = m
			case string:
				isTruthy = m != ""
			case nil:
				isTruthy = false
			default:
				isTruthy = true // Other non-nil values are considered truthy
			}

			if isTruthy {
				key := mergeNode.GetName()
				if groupPtr, exists := groups[key]; !exists {
					// First rule with merge for this property name
					// Unlike previous buggy behavior, we DON'T search backwards for non-merge rules
					// because JavaScript only merges rules that have the merge flag set
					newGroup := []any{rule}
					groups[key] = &newGroup
					groupsArr = append(groupsArr, &newGroup)
					// Keep the current rule in place (it will be updated with merged value later)
				} else {
					// Subsequent rule with merge for this property - add to existing group and remove
					*groupPtr = append(*groupPtr, rule)
					// Remove from rules array
					rules = append(rules[:i], rules[i+1:]...)
					i--
				}
			}
		}
	}
	
	for _, groupPtr := range groupsArr {
		group := *groupPtr
		if len(group) > 0 {
			result := group[0]
			space := []any{}
			comma := []any{}

			for _, rule := range group {
				if mergeRule, ok := rule.(interface{ GetMerge() any; GetValue() any; GetImportant() bool }); ok {
					// If merge is "+" and we have content, start a new expression for comma separation
					if mergeValue, ok := mergeRule.GetMerge().(string); ok && mergeValue == "+" {
						if len(space) > 0 {
							// Finalize current space expression and start a new one
							spaceExpr, _ := NewExpression(space, false)
							comma = append(comma, spaceExpr)
							space = []any{}
						}
					}
					// Add the value to the current space
					space = append(space, mergeRule.GetValue())

					// Merge important flags
					if resultSetter, ok := result.(interface{ SetImportant(bool) }); ok {
						if resultGetter, ok := result.(interface{ GetImportant() bool }); ok {
							resultSetter.SetImportant(resultGetter.GetImportant() || mergeRule.GetImportant())
						}
					}
				}
			}

			// Add the final space expression
			if len(space) > 0 {
				spaceExpr, _ := NewExpression(space, false)
				comma = append(comma, spaceExpr)
			}

			// Set the merged value
			if resultSetter, ok := result.(interface{ SetValue(any) }); ok {
				value, _ := NewValue(comma)
				resultSetter.SetValue(value)
			}
		}
	}
	return rules
}