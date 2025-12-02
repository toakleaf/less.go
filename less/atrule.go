package less_go

import (
	"fmt"
	"os"
	"strings"
)

type AtRule struct {
	*Node
	Name        string
	Value       any
	Rules       []any
	IsRooted    bool
	AllowRoot   bool
	DebugInfo   any
	AllExtends  []*Extend // For storing extends found by ExtendFinderVisitor
}

func NewAtRule(name string, value any, rules any, index int, currentFileInfo map[string]any, debugInfo any, isRooted bool, visibilityInfo map[string]any) *AtRule {
	node := NewNode()
	node.TypeIndex = GetTypeIndexForNodeType("AtRule")

	atRule := &AtRule{
		Node:      node,
		Name:      name,
		IsRooted:  isRooted,
		AllowRoot: true,
		DebugInfo: debugInfo,
	}

	// Handle value - convert to Anonymous if string/non-Node
	if value != nil {
		// Check if value is already a Node-based type
		// This matches JavaScript: (value instanceof Node) ? value : (value ? new Anonymous(value) : value)
		// Preserve Anonymous nodes and any node that has a GetType() method (Quoted, Variable, Expression, etc.)
		if _, ok := value.(*Anonymous); ok {
			atRule.Value = value
		} else if _, ok := value.(interface{ GetType() string }); ok {
			atRule.Value = value
		} else {
			atRule.Value = NewAnonymous(value, index, currentFileInfo, false, false, nil)
		}
	} else {
		atRule.Value = value
	}

	// Handle rules
	if rules != nil {
		if rulesSlice, ok := rules.([]any); ok {
			atRule.Rules = rulesSlice
		} else {
			// Single rule - convert to array
			atRule.Rules = []any{rules}
		}

		// Check if this is a bubblable at-rule (@supports or @document)
		// Only these at-rules need empty selectors for selector joining
		nonVendorName := stripVendorPrefix(name)
		isBubblable := nonVendorName == "@supports" || nonVendorName == "@document"

		// Set allowImports and Root for ALL at-rules
		// Only create empty selectors for bubblable at-rules (@supports/@document)
		for _, rule := range atRule.Rules {
			if ruleset, ok := rule.(*Ruleset); ok {
				// These settings are needed for all at-rules
				ruleset.AllowImports = true
				ruleset.Root = true

				// Only add empty selectors for bubblable at-rules
				// This is needed for JoinSelectorVisitor to properly join selectors
				if isBubblable && len(ruleset.Selectors) == 0 {
					// Create empty selectors directly (same as Selector.CreateEmptySelectors)
					// Create an Element with & as the value
					el := NewElement("", "&", false, index, currentFileInfo, nil)
					// Create Selector with the element
					sel, err := NewSelector(el, nil, nil, index, currentFileInfo, nil)
					if err == nil {
						sel.MediaEmpty = true
						ruleset.Selectors = []any{sel}
					}
				}
			}
		}
		atRule.SetParent(atRule.Rules, atRule.Node)
	}

	// Set node properties
	atRule.Index = index
	if currentFileInfo != nil {
		atRule.SetFileInfo(currentFileInfo)
	}
	atRule.CopyVisibilityInfo(visibilityInfo)

	return atRule
}

func (a *AtRule) Type() string {
	return "AtRule"
}

func (a *AtRule) GetType() string {
	return "AtRule"
}

func (a *AtRule) GetName() string {
	return a.Name
}

func (a *AtRule) GetDebugInfo() any {
	return a.DebugInfo
}

func (a *AtRule) GetIsRooted() bool {
	return a.IsRooted
}

func (a *AtRule) GetRules() []any {
	return a.Rules
}

func (a *AtRule) SetRules(rules []any) {
	a.Rules = rules
}

func (a *AtRule) ToCSS(context any) string {
	var strs []string
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			strs = append(strs, fmt.Sprintf("%v", chunk))
		},
		IsEmpty: func() bool {
			return len(strs) == 0
		},
	}
	a.GenCSS(context, output)
	return strings.Join(strs, "")
}

func (a *AtRule) Accept(visitor any) {
	if a.Rules != nil {
		// Try the variadic signature first (matches Visitor.VisitArray)
		if v, ok := visitor.(interface{ VisitArray([]any, ...bool) []any }); ok {
			a.Rules = v.VisitArray(a.Rules)
		} else if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
			// Fallback to non-variadic signature for compatibility
			a.Rules = v.VisitArray(a.Rules)
		}
	}

	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		if a.Value != nil {
			a.Value = v.Visit(a.Value)
		}
	}
}

func (a *AtRule) IsRulesetLike() any {
	if a.Rules != nil {
		return a.Rules
	}
	return !a.IsCharset()
}

func (a *AtRule) IsCharset() bool {
	return a.Name == "@charset"
}

// stripVendorPrefix removes vendor prefix from at-rule names
// e.g., "@-x-document" -> "@document", "@-webkit-keyframes" -> "@keyframes"
func stripVendorPrefix(name string) string {
	if len(name) > 1 && name[1] == '-' {
		// Find the second dash (after vendor prefix)
		for i := 2; i < len(name); i++ {
			if name[i] == '-' {
				return "@" + name[i+1:]
			}
		}
	}
	return name
}

func (a *AtRule) GenCSS(context any, output *CSSOutput) {
	// Check visibility - skip if node blocks visibility and is not explicitly visible
	// This implements the reference import functionality where nodes from referenced
	// imports are hidden unless explicitly used (via extend or mixin call)
	if a.Node != nil && a.Node.BlocksVisibility() {
		nodeVisible := a.Node.IsVisible()
		if nodeVisible == nil || !*nodeVisible {
			// Node blocks visibility and is not explicitly visible, skip output
			return
		}
	}

	// Check if this directive has rules but they ONLY contain silent content (line comments)
	// In that case, skip output entirely since line comments are stripped from CSS output
	// Exception: @keyframes (and vendor-prefixed variants) should always be output,
	// even if empty, because they define animation names
	// Note: CSS block comments (/* ... */) ARE output, so we don't skip those
	isKeyframes := strings.Contains(a.Name, "keyframes")
	if a.Rules != nil && !isKeyframes {
		hasOnlySilentContent := false
		for _, rule := range a.Rules {
			if ruleset, ok := rule.(*Ruleset); ok {
				// Only check if ruleset has rules - empty rulesets should still be output
				if len(ruleset.Rules) > 0 {
					hasVisibleContent := false
					// Check if ruleset has any visible (non-silent) content
					for _, r := range ruleset.Rules {
						if comment, isComment := r.(*Comment); isComment {
							// Only line comments are silent (stripped from output)
							// Block comments (/* ... */) are visible content
							if !comment.IsLineComment {
								hasVisibleContent = true
								break
							}
						} else {
							// Any non-comment rule is visible content
							hasVisibleContent = true
							break
						}
					}
					if !hasVisibleContent {
						// This ruleset has rules, but they're all silent (line comments)
						hasOnlySilentContent = true
					} else {
						// This ruleset has visible content
						hasOnlySilentContent = false
						break
					}
				}
			}
		}
		if hasOnlySilentContent {
			// All rulesets contain only silent content, skip output
			return
		}
	}

	output.Add(a.Name, a.FileInfo(), a.GetIndex())

	if a.Value != nil {
		output.Add(" ", nil, nil)
		if gen, ok := a.Value.(interface{ GenCSS(any, *CSSOutput) }); ok {
			gen.GenCSS(context, output)
		}
	}

	if a.Rules != nil {
		a.OutputRuleset(context, output, a.Rules)
	} else {
		// Add semicolon only (parent will add newline between rules)
		output.Add(";", nil, nil)
	}
}

func (a *AtRule) Eval(context any) (any, error) {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG AtRule.Eval] name=%q, hasRules=%v, isRooted=%v\n", a.Name, len(a.Rules) > 0, a.IsRooted)
	}

	// Standard directives use regular evaluation
	// Note: @supports/@document stay in place during eval; their selectors are joined
	// by JoinSelectorVisitor later (NOT via mediaBlocks bubbling like @media)
	var mediaPathBackup, mediaBlocksBackup any
	var value any = a.Value
	var rules []any = a.Rules

	// Media stored inside other atrule should not bubble over it
	// Backup media bubbling information
	if evalCtx, ok := context.(*Eval); ok {
		mediaPathBackup = evalCtx.MediaPath
		mediaBlocksBackup = evalCtx.MediaBlocks
		// Delete media bubbling information
		evalCtx.MediaPath = []any{}
		evalCtx.MediaBlocks = []any{}
	} else if ctx, ok := context.(map[string]any); ok {
		mediaPathBackup = ctx["mediaPath"]
		mediaBlocksBackup = ctx["mediaBlocks"]
		// Delete media bubbling information
		ctx["mediaPath"] = []any{}
		ctx["mediaBlocks"] = []any{}
	}

	if value != nil {
		if eval, ok := value.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(context)
			if err != nil {
				return nil, err
			}
			value = evaluated
		}
	}

	if len(rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if eval, ok := rules[0].(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(context)
			if err != nil {
				return nil, err
			}
			// Convert back to Ruleset if possible
			if rs, ok := evaluated.(*Ruleset); ok {
				rules = []any{rs}
				// IMPORTANT: Set Root=true for rooted directives (@font-face, @keyframes)
				// Also set Root=true for vendor-prefixed @keyframes (@-webkit-keyframes, etc.)
				// For non-rooted directives (@supports, @document), leave Root unset
				// so JoinSelectorVisitor can properly handle selector joining
				isKeyframes := strings.Contains(a.Name, "keyframes")
				if a.IsRooted || isKeyframes {
					rs.Root = true
				}
			} else {
				rules = []any{evaluated}
			}
		}
	}

	// Restore media bubbling information
	if evalCtx, ok := context.(*Eval); ok {
		if mb, ok := mediaBlocksBackup.([]any); ok {
			evalCtx.MediaBlocks = mb
		}
		if mp, ok := mediaPathBackup.([]any); ok {
			evalCtx.MediaPath = mp
		}
	} else if ctx, ok := context.(map[string]any); ok {
		ctx["mediaPath"] = mediaPathBackup
		ctx["mediaBlocks"] = mediaBlocksBackup
	}

	return NewAtRule(a.Name, value, rules, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo()), nil
}

func (a *AtRule) EvalTop(context any) any {
	// For AtRules, we DON'T clear mediaBlocks like Media does
	// Instead, we return an empty ruleset as a placeholder
	// The directive stays in mediaBlocks and will be collected by the root ruleset
	// This is different from Media because AtRules can be nested in regular rulesets,
	// whereas Media queries handle their own nesting with permutations
	return NewRuleset([]any{}, []any{}, false, nil)
}

func (a *AtRule) EvalNested(context any) any {
	// Handle both *Eval and map[string]any contexts
	var mediaPath []any
	var hasMediaPath bool

	if evalCtx, ok := context.(*Eval); ok {
		mediaPath = evalCtx.MediaPath
		hasMediaPath = len(mediaPath) > 0
	} else if ctx, ok := context.(map[string]any); ok {
		mediaPath, hasMediaPath = ctx["mediaPath"].([]any)
	} else {
		return a
	}

	if !hasMediaPath {
		mediaPath = []any{}
	}

	// Create path with current node
	path := append(mediaPath, a)

	// Extract the conditions separated with `,` (OR)
	for i := 0; i < len(path); i++ {
		var pathType string
		switch p := path[i].(type) {
		case *AtRule:
			pathType = p.GetType()
		case interface{ GetType() string }:
			pathType = p.GetType()
		default:
			continue
		}

		if pathType != a.GetType() {
			// Remove from mediaBlocks if types don't match
			if evalCtx, ok := context.(*Eval); ok {
				if i < len(evalCtx.MediaBlocks) {
					evalCtx.MediaBlocks = append(evalCtx.MediaBlocks[:i], evalCtx.MediaBlocks[i+1:]...)
				}
			} else if ctx, ok := context.(map[string]any); ok {
				if mediaBlocks, hasMediaBlocks := ctx["mediaBlocks"].([]any); hasMediaBlocks && i < len(mediaBlocks) {
					ctx["mediaBlocks"] = append(mediaBlocks[:i], mediaBlocks[i+1:]...)
				}
			}
			return a
		}
	}

	// For AtRules (unlike Media), we don't need to permute features
	// Just return a fake tree-node that doesn't output anything
	return NewRuleset([]any{}, []any{}, false, nil)
}

func (a *AtRule) BubbleSelectors(selectors any) {
	if selectors == nil {
		return
	}
	if len(a.Rules) == 0 {
		return
	}

	// Handle both []*Selector and []any types
	var anySelectors []any

	switch s := selectors.(type) {
	case []*Selector:
		copiedSelectors := make([]*Selector, len(s))
		copy(copiedSelectors, s)

		// Convert selectors to []any
		anySelectors = make([]any, len(copiedSelectors))
		for i, sel := range copiedSelectors {
			anySelectors[i] = sel
		}
	case []any:
		// Copy the slice
		anySelectors = make([]any, len(s))
		copy(anySelectors, s)
	default:
		return
	}

	// Create a new Ruleset wrapper with the selectors containing the original rules
	// This matches JavaScript: this.rules = [new Ruleset(utils.copyArray(selectors), [this.rules[0]])];
	newRuleset := NewRuleset(anySelectors, []any{a.Rules[0]}, false, nil)
	a.Rules = []any{newRuleset}
	a.SetParent(a.Rules, a.Node)
}

func (a *AtRule) Permute(arr []any) any {
	if len(arr) == 0 {
		return []any{}
	} else if len(arr) == 1 {
		return arr[0]
	} else {
		result := []any{}
		rest := a.Permute(arr[1:])

		restArray, ok := rest.([]any)
		if !ok {
			return nil
		}

		firstArray, ok := arr[0].([]any)
		if !ok {
			return nil
		}

		for i := 0; i < len(restArray); i++ {
			restItem, ok := restArray[i].([]any)
			if !ok {
				restItem = []any{restArray[i]}
			}

			for j := 0; j < len(firstArray); j++ {
				combined := append([]any{firstArray[j]}, restItem...)
				result = append(result, combined)
			}
		}
		return result
	}
}

func (a *AtRule) Variable(name string) any {
	if len(a.Rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			return ruleset.Variable(name)
		}
	}
	return nil
}

func (a *AtRule) Find(selector any, self any, filter func(any) bool) []any {
	if len(a.Rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			return ruleset.Find(selector, self, filter)
		}
	}
	return nil
}

func (a *AtRule) Rulesets() []any {
	if len(a.Rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			return ruleset.Rulesets()
		}
	}
	return nil
}

func (a *AtRule) OutputRuleset(context any, output *CSSOutput, rules []any) {
	ruleCnt := len(rules)

	ctx, ok := context.(map[string]any)
	if !ok {
		ctx = make(map[string]any)
	}

	tabLevel := 0
	if tl, ok := ctx["tabLevel"].(int); ok {
		tabLevel = tl
	}
	tabLevel = tabLevel + 1
	ctx["tabLevel"] = tabLevel

	compress := false
	if c, ok := ctx["compress"].(bool); ok {
		compress = c
	}

	if compress {
		output.Add("{", nil, nil)
		for i := 0; i < ruleCnt; i++ {
			if gen, ok := rules[i].(interface{ GenCSS(any, *CSSOutput) }); ok {
				gen.GenCSS(ctx, output)
			}
		}
		output.Add("}", nil, nil)
		ctx["tabLevel"] = tabLevel - 1
		return
	}

	// Non-compressed
	// JavaScript: Array(context.tabLevel).join('  ') creates (tabLevel-1) pairs of spaces
	tabSetStr := "\n" + strings.Repeat("  ", tabLevel-1)
	tabRuleStr := tabSetStr + "  "

	if ruleCnt == 0 {
		output.Add(" {"+tabSetStr+"}", nil, nil)
	} else {
		// Output opening brace without initial indent - we'll add indent before first rule that outputs
		output.Add(" {", nil, nil)

		// Track whether we've output any rule content yet
		hasOutputContent := false

		// Process all rules uniformly
		for i := 0; i < ruleCnt; i++ {
			rule := rules[i]

			// Check if this rule will output anything (for visibility filtering)
			// Skip adding newline/indent for rules that will be filtered out
			willOutput := true
			if rs, ok := rule.(*Ruleset); ok {
				// Check if ruleset is a media-empty ruleset with only silent content
				// These rulesets have MediaEmpty selectors but no visible rules to output
				if len(rs.Selectors) == 1 {
					if sel, ok := rs.Selectors[0].(*Selector); ok && sel.MediaEmpty {
						// Check if ruleset will produce any output by looking at its rules
						hasNonSilentContent := false
						for _, ruleContent := range rs.Rules {
							// Check if this is a comment and whether it's silent
							if comment, isComment := ruleContent.(*Comment); isComment {
								// Non-silent comments (block comments) ARE content
								if !comment.IsSilent(ctx) {
									hasNonSilentContent = true
									break
								}
							} else {
								// Any other rule type is non-silent content
								hasNonSilentContent = true
								break
							}
						}
						if !hasNonSilentContent {
							willOutput = false
						}
					}
				}
				// Check if ruleset blocks visibility and has no visible paths
				if willOutput && rs.Node != nil && rs.Node.BlocksVisibility() {
					nodeVisible := rs.Node.IsVisible()
					if nodeVisible == nil || !*nodeVisible {
						// Check if any path has visible selectors
						hasVisiblePath := false
						if rs.Paths != nil && len(rs.Paths) > 0 {
							for _, path := range rs.Paths {
								for _, pathElem := range path {
									if sel, ok := pathElem.(*Selector); ok {
										if sel.Node != nil {
											selVis := sel.Node.IsVisible()
											if selVis != nil && *selVis {
												hasVisiblePath = true
												break
											}
										}
									}
								}
								if hasVisiblePath {
									break
								}
							}
						}
						if !hasVisiblePath {
							willOutput = false
						}
					}
				}
			}

			// Add newline/indent before rules that will output
			if willOutput {
				output.Add(tabRuleStr, nil, nil)
				hasOutputContent = true
			}

			// Set lastRule flag for the last rule
			if i+1 == ruleCnt {
				ctx["lastRule"] = true
			}

			if gen, ok := rule.(interface{ GenCSS(any, *CSSOutput) }); ok {
				gen.GenCSS(ctx, output)
			}

			// Clear lastRule after processing
			if i+1 == ruleCnt {
				ctx["lastRule"] = false
			}
		}

		// If no rules output anything, still need to add a newline before closing brace
		if !hasOutputContent {
			output.Add(tabSetStr, nil, nil)
		}

		output.Add(tabSetStr+"}", nil, nil)
		// Note: Don't add newline after closing brace here.
		// The parent ruleset's GenCSS will add the appropriate newline
		// between top-level rules.
	}

	ctx["tabLevel"] = tabLevel - 1
}

func (a *AtRule) SetAllExtends(extends []*Extend) {
	a.AllExtends = extends
}

func (a *AtRule) GetAllExtends() []*Extend {
	return a.AllExtends
} 