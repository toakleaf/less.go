package less_go

import (
	"fmt"
	"os"
)

// Media represents a media query node in the Less AST
type Media struct {
	*AtRule
	Features any
	Rules    []any
	DebugInfo any
}

// NewMedia creates a new Media instance
func NewMedia(value any, features any, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) *Media {
	// Match JavaScript: (new Selector([], null, null, this._index, this._fileInfo)).createEmptySelectors()
	selector, _ := NewSelector([]any{}, nil, nil, index, currentFileInfo, nil)
	emptySelectors, _ := selector.CreateEmptySelectors()
	
	// Convert selectors to []any for Ruleset
	selectors := make([]any, len(emptySelectors))
	for i, sel := range emptySelectors {
		selectors[i] = sel
	}

	// Match JavaScript: this.features = new Value(features)
	featuresValue, _ := NewValue(features)

	// Match JavaScript: this.rules = [new Ruleset(selectors, value)]
	// Convert value to []any for Ruleset
	var rules []any
	if value != nil {
		if valueSlice, ok := value.([]any); ok {
			rules = valueSlice
		} else {
			rules = []any{value}
		}
	}
	ruleset := NewRuleset(selectors, rules, false, visibilityInfo)
	ruleset.AllowImports = true

	// Create Media instance
	media := &Media{
		AtRule:   NewAtRule("@media", nil, nil, index, currentFileInfo, nil, false, visibilityInfo),
		Features: featuresValue,
		Rules:    []any{ruleset},
	}

	// Match JavaScript: this.allowRoot = true
	media.AllowRoot = true
	media.CopyVisibilityInfo(visibilityInfo)

	// Match JavaScript: this.setParent calls
	media.SetParent(selectors, media.AtRule.Node)
	media.SetParent(media.Features, media.AtRule.Node)
	media.SetParent(media.Rules, media.AtRule.Node)

	return media
}

// GetType returns the type of the node
func (m *Media) GetType() string {
	return "Media"
}

// Type returns the type of the node (for compatibility)
func (m *Media) Type() string {
	return "Media"
}

// GetTypeIndex returns the type index for visitor pattern
func (m *Media) GetTypeIndex() int {
	return GetTypeIndexForNodeType("Media")
}

// GetRules returns the rules for this media query
func (m *Media) GetRules() []any {
	return m.Rules
}

// IsRulesetLike returns true (implementing NestableAtRulePrototype)
func (m *Media) IsRulesetLike() bool {
	return true
}

// Accept visits the node with a visitor (implementing NestableAtRulePrototype)
func (m *Media) Accept(visitor any) {
	if m.Features != nil {
		if v, ok := visitor.(interface{ Visit(any) any }); ok {
			m.Features = v.Visit(m.Features)
		}
	}
	if m.Rules != nil {
		// Try variadic bool version first (like Ruleset.Accept)
		if v, ok := visitor.(interface{ VisitArray([]any, ...bool) []any }); ok {
			m.Rules = v.VisitArray(m.Rules)
		} else if v, ok := visitor.(interface{ VisitArray([]any, bool) []any }); ok {
			m.Rules = v.VisitArray(m.Rules, false)
		} else if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
			m.Rules = v.VisitArray(m.Rules)
		}
	}
}

// EvalTop evaluates the media rule at the top level (implementing NestableAtRulePrototype)
func (m *Media) EvalTop(context any) any {
	if os.Getenv("LESS_GO_TRACE") != "" || os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[MEDIA.EvalTop] Starting\n")
	}

	var result any = m

	// Handle both *Eval and map[string]any contexts
	var mediaBlocks []any

	if evalCtx, ok := context.(*Eval); ok {
		mediaBlocks = evalCtx.MediaBlocks

		if os.Getenv("LESS_GO_TRACE") != "" || os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MEDIA.EvalTop] mediaBlocks count: %d\n", len(mediaBlocks))
			for i, mb := range mediaBlocks {
				fmt.Fprintf(os.Stderr, "[MEDIA.EvalTop]   mediaBlock[%d]: type=%T\n", i, mb)
			}
		}

		// Render all dependent Media blocks
		// Match JavaScript: use ALL mediaBlocks (don't filter out current media)
		// The parent media will have empty Rules (because nested evalNested() returned empty Ruleset)
		// so it won't output, but the nested media (with merged features) will output
		if len(mediaBlocks) > 1 {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[MEDIA.EvalTop] Creating MultiMedia Ruleset with %d media blocks\n", len(mediaBlocks))
				for i, mb := range mediaBlocks {
					if media, ok := mb.(*Media); ok {
						fmt.Fprintf(os.Stderr, "[MEDIA.EvalTop]   Media[%d]: Rules count=%d\n", i, len(media.Rules))
					}
				}
			}

			// Create empty selectors
			selector, err := NewSelector(nil, nil, nil, m.GetIndex(), m.FileInfo(), nil)
			if err != nil {
				return result
			}
			emptySelectors, err := selector.CreateEmptySelectors()
			if err != nil {
				return result
			}

			// Create new Ruleset - convert selectors to []any
			selectors := make([]any, len(emptySelectors))
			for i, sel := range emptySelectors {
				selectors[i] = sel
			}
			// Create MultiMedia Ruleset with root=true so inner Media nodes are not extracted by ToCSSVisitor
			// The MultiMedia Ruleset is a container for merged media queries
			ruleset := NewRuleset(selectors, mediaBlocks, false, m.VisibilityInfo())
			ruleset.MultiMedia = true // Set MultiMedia to true for multiple media blocks
			ruleset.Root = true       // Set Root to true so Media nodes stay as Rules (not extracted as separate rulesets)
			ruleset.AllowImports = true // Keep this ruleset visible even without selectors
			ruleset.CopyVisibilityInfo(m.VisibilityInfo())
			m.SetParent(ruleset.Node, m.Node)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[MEDIA.EvalTop] MultiMedia Ruleset created, returning it\n")
			}
			result = ruleset
		}

		// Clear mediaBlocks and mediaPath from context
		evalCtx.MediaBlocks = nil
		evalCtx.MediaPath = nil

		return result
	}

	// Handle map[string]any context (for backward compatibility)
	ctx, ok := context.(map[string]any)
	if !ok {
		return result
	}

	mediaBlocks, _ = ctx["mediaBlocks"].([]any)
	// Match JavaScript: use ALL mediaBlocks without filtering
	if len(mediaBlocks) > 1 {
		// Create empty selectors
		selector, err := NewSelector(nil, nil, nil, m.GetIndex(), m.FileInfo(), nil)
		if err != nil {
			return result
		}
		emptySelectors, err := selector.CreateEmptySelectors()
		if err != nil {
			return result
		}

		// Create new Ruleset - convert selectors to []any
		selectors := make([]any, len(emptySelectors))
		for i, sel := range emptySelectors {
			selectors[i] = sel
		}
		// Create MultiMedia Ruleset with root=true so inner Media nodes are not extracted by ToCSSVisitor
		// The MultiMedia Ruleset is a container for merged media queries
		ruleset := NewRuleset(selectors, mediaBlocks, false, m.VisibilityInfo())
		ruleset.MultiMedia = true // Set MultiMedia to true for multiple media blocks
		ruleset.Root = true       // Set Root to true so Media nodes stay as Rules (not extracted as separate rulesets)
		ruleset.AllowImports = true // Keep this ruleset visible even without selectors
		ruleset.CopyVisibilityInfo(m.VisibilityInfo())
		m.SetParent(ruleset.Node, m.Node)
		result = ruleset
	}

	// Delete mediaBlocks and mediaPath from context
	delete(ctx, "mediaBlocks")
	delete(ctx, "mediaPath")

	return result
}

// EvalNested evaluates the media rule in a nested context (implementing NestableAtRulePrototype)
func (m *Media) EvalNested(context any) any {
	// Handle both *Eval and map[string]any contexts
	var mediaPath []any
	var hasMediaPath bool

	if evalCtx, ok := context.(*Eval); ok {
		mediaPath = evalCtx.MediaPath
		hasMediaPath = len(mediaPath) > 0
	} else if ctx, ok := context.(map[string]any); ok {
		mediaPath, hasMediaPath = ctx["mediaPath"].([]any)
	} else {
		return m
	}

	if !hasMediaPath {
		mediaPath = []any{}
	}

	// Create path with current node
	path := append(mediaPath, m)

	// Extract the media-query conditions separated with `,` (OR)
	for i := 0; i < len(path); i++ {
		var pathType string
		switch p := path[i].(type) {
		case *Media:
			pathType = p.GetType()
		case interface{ GetType() string }:
			pathType = p.GetType()
		default:
			continue
		}

		if pathType != m.GetType() {
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
			return m
		}

		var value any
		var features any

		// Get features from the path item
		if media, ok := path[i].(*Media); ok {
			features = media.Features
		}

		if valueNode, ok := features.(*Value); ok {
			value = valueNode.Value
		} else {
			value = features
		}

		// Convert to array if needed
		if arr, ok := value.([]any); ok {
			path[i] = arr
		} else {
			path[i] = []any{value}
		}
	}

	// Trace all permutations to generate the resulting media-query
	permuteResult := m.Permute(path)
	if permuteResult == nil {
		return m
	}

	permuteArray, ok := permuteResult.([]any)
	if !ok {
		return m
	}

	// Ensure every path is an array before mapping
	for _, p := range permuteArray {
		if _, ok := p.([]any); !ok {
			return m
		}
	}

	// Map paths to expressions
	expressions := make([]any, len(permuteArray))
	for idx, pathItem := range permuteArray {
		pathArray, ok := pathItem.([]any)
		if !ok {
			continue
		}

		// Convert fragments
		mappedPath := make([]any, len(pathArray))
		for i, fragment := range pathArray {
			if _, ok := fragment.(interface{ ToCSS(any) string }); ok {
				mappedPath[i] = fragment
			} else {
				mappedPath[i] = NewAnonymous(fragment, 0, nil, false, false, nil)
			}
		}

		// Insert 'and' between fragments
		for i := len(mappedPath) - 1; i > 0; i-- {
			andAnon := NewAnonymous("and", 0, nil, false, false, nil)
			mappedPath = append(mappedPath[:i], append([]any{andAnon}, mappedPath[i:]...)...)
		}

		expr, err := NewExpression(mappedPath, false)
		if err != nil {
			continue
		}
		expressions[idx] = expr
	}

	// Create new Value with expressions
	newValue, err := NewValue(expressions)
	if err == nil {
		m.Features = newValue
		m.SetParent(m.Features, m.Node)
	}

	// Return fake tree-node that doesn't output anything
	return NewRuleset([]any{}, []any{}, false, nil)
}

// Permute creates permutations of the given array (implementing NestableAtRulePrototype)
func (m *Media) Permute(arr []any) any {
	if len(arr) == 0 {
		return []any{}
	} else if len(arr) == 1 {
		return arr[0]
	} else {
		result := []any{}
		rest := m.Permute(arr[1:])
		
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

// hasOnlyEmptyContent recursively checks if rules contain only empty content
func hasOnlyEmptyContent(rules []any) bool {
	if len(rules) == 0 {
		return true
	}

	for _, rule := range rules {
		// Check if it's a Ruleset
		if rs, ok := rule.(*Ruleset); ok {
			// AllowImports rulesets are considered to have content
			// even if they appear empty (they're structural wrappers)
			if rs.AllowImports {
				return false
			}
			// If it has selectors with content, it's not empty
			if len(rs.Selectors) > 0 && len(rs.Rules) > 0 && !hasOnlyEmptyContent(rs.Rules) {
				return false
			}
			// Recursively check nested rulesets
			if !hasOnlyEmptyContent(rs.Rules) {
				return false
			}
		} else if media, ok := rule.(*Media); ok {
			// Check if nested media has content
			if len(media.Rules) > 0 {
				if rs, ok := media.Rules[0].(*Ruleset); ok {
					if !hasOnlyEmptyContent(rs.Rules) {
						return false
					}
				} else {
					// Media has non-Ruleset content
					return false
				}
			}
		} else if _, ok := rule.(*Declaration); ok {
			// Declarations are content
			return false
		} else if _, ok := rule.(*Comment); ok {
			// Comments might be considered content in some contexts
			// For now, we'll consider them as non-content for empty detection
			continue
		} else if _, ok := rule.(*VariableCall); ok {
			// VariableCall nodes don't have GenCSS, so they output nothing
			// They should have been evaluated during Eval, but if they're still here
			// at GenCSS time, treat them as empty
			continue
		} else if _, ok := rule.(*MixinCall); ok {
			// MixinCall nodes don't have GenCSS either
			continue
		} else {
			// Any other rule type is considered content
			return false
		}
	}

	return true
}

// BubbleSelectors bubbles selectors up the tree (implementing NestableAtRulePrototype)
func (m *Media) BubbleSelectors(selectors any) {
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors] Called with selectors: %v\n", selectors)
	}

	// Match JavaScript: if (!selectors) return;
	// This includes nil and empty arrays
	if selectors == nil {
		return
	}

	// Check for empty slices - match JavaScript behavior where empty arrays are falsy
	switch s := selectors.(type) {
	case []*Selector:
		if len(s) == 0 {
			if os.Getenv("LESS_GO_TRACE") != "" {
				fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors] Skipping empty selector array\n")
			}
			return
		}
	case []any:
		if len(s) == 0 {
			if os.Getenv("LESS_GO_TRACE") != "" {
				fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors] Skipping empty selector array\n")
			}
			return
		}
	}

	if len(m.Rules) == 0 {
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

	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors] Wrapping Rules with %d selectors\n", len(anySelectors))
		for i, selAny := range anySelectors {
			if sel, ok := selAny.(*Selector); ok {
				fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors]   Selector %d has %d elements\n", i, len(sel.Elements))
				for j, elem := range sel.Elements {
					fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors]     Element %d: Value=%v\n", j, elem.Value)
				}
			}
		}
	}

	// Filter out parent selector references (&) from selectors
	// When detached rulesets are called in contexts without parent selectors (e.g., directly inside @media),
	// the & elements need to be removed as they have no parent to reference
	cleanedSelectors := make([]any, 0, len(anySelectors))
	for _, selAny := range anySelectors {
		if sel, ok := selAny.(*Selector); ok {
			// Filter out elements that are just "&" (parent selector reference)
			cleanedElements := make([]*Element, 0, len(sel.Elements))
			for _, elem := range sel.Elements {
				// Skip elements that are ONLY a parent selector reference
				if str, ok := elem.Value.(string); ok && str == "&" {
					continue
				}
				cleanedElements = append(cleanedElements, elem)
			}

			// Only include selectors that have elements after filtering
			if len(cleanedElements) > 0 {
				// Create new selector with filtered elements
				cleanedSel := &Selector{
					Elements:       cleanedElements,
					EvaldCondition: sel.EvaldCondition,
					Condition:      sel.Condition,
					ExtendList:     sel.ExtendList,
					MediaEmpty:     sel.MediaEmpty,
					Node:           sel.Node,
				}
				cleanedSelectors = append(cleanedSelectors, cleanedSel)
			}
		} else {
			// If it's not a *Selector, keep it as-is
			cleanedSelectors = append(cleanedSelectors, selAny)
		}
	}

	// If all selectors were filtered out (all were just "&"), don't wrap
	if len(cleanedSelectors) == 0 {
		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors] All selectors were parent references, skipping wrap\n")
		}
		return
	}

	// Check if m.Rules[0] is a Ruleset with matching selectors
	// If so, use its Rules directly instead of double-wrapping
	var innerContent []any
	if innerRuleset, ok := m.Rules[0].(*Ruleset); ok {
		// If the inner ruleset has selectors, use its Rules (unwrap one level)
		// This prevents double-nesting like .my-selector { .my-selector { ... } }
		if len(innerRuleset.Selectors) > 0 || len(innerRuleset.Paths) > 0 {
			innerContent = innerRuleset.Rules
			if os.Getenv("LESS_GO_TRACE") != "" {
				fmt.Fprintf(os.Stderr, "[MEDIA.BubbleSelectors] Inner ruleset has selectors, unwrapping its Rules\n")
			}
		} else {
			// Inner ruleset has no selectors, wrap it as-is
			innerContent = []any{m.Rules[0]}
		}
	} else {
		// Not a Ruleset, wrap as-is
		innerContent = []any{m.Rules[0]}
	}

	newRuleset := NewRuleset(cleanedSelectors, innerContent, false, nil)

	// CRITICAL: Set Paths from selectors for proper CSS generation
	// Ruleset.GenCSS prefers Paths over Selectors. Since these selectors are already
	// evaluated/compiled, we can directly use them as paths (each selector becomes its own path)
	paths := make([][]any, len(cleanedSelectors))
	for i, sel := range cleanedSelectors {
		paths[i] = []any{sel}
	}
	newRuleset.Paths = paths

	m.Rules = []any{newRuleset}
	m.SetParent(m.Rules, m.Node)
}

// GenCSS generates CSS representation
func (m *Media) GenCSS(context any, output *CSSOutput) {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		featDesc := "nil"
		if m.Features != nil {
			if val, ok := m.Features.(*Value); ok && len(val.Value) > 0 {
				featDesc = fmt.Sprintf("Value with %d expressions", len(val.Value))
			} else {
				featDesc = fmt.Sprintf("%T", m.Features)
			}
		}
		fmt.Fprintf(os.Stderr, "[MEDIA.GenCSS] Called, Rules count=%d, Features=%s\n", len(m.Rules), featDesc)
	}
	// Match JavaScript: Media nodes are always output
	// Visibility filtering happens at the rule level inside the media block, not at the media block itself
	// JavaScript media.genCSS() has no visibility check

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[Media.GenCSS] Called, Rules count=%d\n", len(m.Rules))
	}

	// Skip media queries with empty rulesets (happens when nested media queries are merged)
	// When evalNested merges nested media queries, it returns an empty Ruleset as a placeholder
	// but the Media node itself should not be output if it has no content
	if len(m.Rules) == 0 {
		// No rules at all, skip
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[Media.GenCSS] Skipping - no rules\n")
		}
		return
	}

	if ruleset, ok := m.Rules[0].(*Ruleset); ok {
		// Check if the ruleset has only empty content (regardless of selectors)
		// A ruleset with selectors but no actual declarations/rules should not be output
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[Media.GenCSS] Checking ruleset with %d rules, Selectors=%d, Paths=%d\n",
				len(ruleset.Rules), len(ruleset.Selectors), len(ruleset.Paths))
			for i, selAny := range ruleset.Selectors {
				if sel, ok := selAny.(*Selector); ok {
					fmt.Fprintf(os.Stderr, "[Media.GenCSS]   Selector %d: %T, Elements=%d\n", i, sel, len(sel.Elements))
					if len(sel.Elements) > 0 {
						for j, elem := range sel.Elements {
							fmt.Fprintf(os.Stderr, "[Media.GenCSS]     Element %d: %T, Value=%v\n", j, elem, elem.Value)
						}
					}
				} else {
					fmt.Fprintf(os.Stderr, "[Media.GenCSS]   Selector %d: %T\n", i, selAny)
				}
			}
			for i, rule := range ruleset.Rules {
				fmt.Fprintf(os.Stderr, "[Media.GenCSS]   Rule %d: %T\n", i, rule)
			}
		}

		// Skip parent media blocks that haven't been merged
		// Parent media have Features as *Expression (unmerged)
		// Merged media have Features as *Value
		// Standalone media have Features as *Paren or other types (should output)
		isUnmergedParent := false
		if m.Features != nil {
			if _, ok := m.Features.(*Expression); ok {
				// Features are *Expression, this is an unmerged parent media from a MultiMedia context
				isUnmergedParent = true
			}
		}

		// Also skip media with empty content
		isEmpty := hasOnlyEmptyContent(ruleset.Rules)

		if isUnmergedParent || isEmpty {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[Media.GenCSS] Skipping - unmerged parent or empty (isUnmergedParent=%v, isEmpty=%v)\n",
					isUnmergedParent, isEmpty)
			}
			return // Skip unmerged parent media or empty media blocks
		}
	}

	output.Add("@media ", m.FileInfo(), m.GetIndex())

	if m.Features != nil {
		if gen, ok := m.Features.(interface{ GenCSS(any, *CSSOutput) }); ok {
			gen.GenCSS(context, output)
		}
	}

	m.OutputRuleset(context, output, m.Rules)
}

// Eval evaluates the media rule - matching JavaScript implementation closely
func (m *Media) Eval(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for Media.Eval")
	}

	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[MEDIA.Eval] Starting eval\n")
	}

	// Convert to *Eval context if needed
	var evalCtx *Eval
	if ec, ok := context.(*Eval); ok {
		evalCtx = ec
	} else if mapCtx, ok := context.(map[string]any); ok {
		// For backward compatibility with map-based contexts
		// This path is used by EvalTop and EvalNested
		return m.evalWithMapContext(mapCtx)
	} else {
		return nil, fmt.Errorf("context must be *Eval or map[string]any, got %T", context)
	}

	// Match JavaScript: if (!context.mediaBlocks) { context.mediaBlocks = []; context.mediaPath = []; }
	if evalCtx.MediaBlocks == nil {
		evalCtx.MediaBlocks = []any{}
		evalCtx.MediaPath = []any{}
	}

	// Match JavaScript: const media = new Media(null, [], this._index, this._fileInfo, this.visibilityInfo())
	media := NewMedia(nil, []any{}, m.GetIndex(), m.FileInfo(), m.VisibilityInfo())

	// Match JavaScript: if (this.debugInfo) { this.rules[0].debugInfo = this.debugInfo; media.debugInfo = this.debugInfo; }
	if m.DebugInfo != nil {
		if len(m.Rules) > 0 {
			if ruleset, ok := m.Rules[0].(*Ruleset); ok {
				ruleset.DebugInfo = m.DebugInfo
			}
		}
		media.DebugInfo = m.DebugInfo
	}

	// Match JavaScript: media.features = this.features.eval(context)
	if m.Features != nil {
		if eval, ok := m.Features.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(context)
			if err != nil {
				return nil, err
			}
			media.Features = evaluated
		} else if eval, ok := m.Features.(interface{ Eval(any) any }); ok {
			media.Features = eval.Eval(context)
		}

		// CRITICAL: After evaluating features, ensure we deeply evaluate all nested values
		// This is necessary for media queries with namespace calls that return expressions
		// containing variables from the mixin scope
		media.Features = m.deeplyEvaluateFeatures(media.Features, context)
	}

	// Match JavaScript: context.mediaPath.push(media); context.mediaBlocks.push(media);
	evalCtx.MediaPath = append(evalCtx.MediaPath, media)
	evalCtx.MediaBlocks = append(evalCtx.MediaBlocks, media)
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		featStr := "nil"
		if media.Features != nil {
			featStr = fmt.Sprintf("%T", media.Features)
		}
		fmt.Fprintf(os.Stderr, "[MEDIA.Eval] Added to mediaBlocks, features=%s, mediaPath len=%d\n", featStr, len(evalCtx.MediaPath))
	}

	// Match JavaScript: this.rules[0].functionRegistry = context.frames[0].functionRegistry.inherit();
	if len(m.Rules) > 0 {
		if ruleset, ok := m.Rules[0].(*Ruleset); ok {
			// Set AllowImports=true so the inner ruleset is considered visible even without selectors
			// This matches JavaScript behavior where Media inner rulesets don't need visible selectors
			ruleset.AllowImports = true

			// Handle function registry inheritance if frames exist
			if len(evalCtx.Frames) > 0 {
				if frameRuleset, ok := evalCtx.Frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					// Stub: ruleset.FunctionRegistry = frameRuleset.FunctionRegistry.Inherit()
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			// Match JavaScript: context.frames.unshift(this.rules[0]);
			newFrames := make([]any, len(evalCtx.Frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], evalCtx.Frames)
			evalCtx.Frames = newFrames

			// Match JavaScript: media.rules = [this.rules[0].eval(context)];
			evaluated, err := ruleset.Eval(context)
			if err != nil {
				return nil, err
			}
			media.Rules = []any{evaluated}

			// Propagate AllowImports to direct children to preserve them during ToCSSVisitor
			// This ensures that rulesets like .my-selector inside detached rulesets are kept
			if evaluatedRuleset, ok := evaluated.(*Ruleset); ok {
				for _, child := range evaluatedRuleset.Rules {
					if childRuleset, ok := child.(*Ruleset); ok {
						childRuleset.AllowImports = true
					}
				}
			}

			// Match JavaScript: context.frames.shift();
			if len(evalCtx.Frames) > 0 {
				evalCtx.Frames = evalCtx.Frames[1:]
			}
		}
	}

	// Match JavaScript: context.mediaPath.pop();
	if len(evalCtx.MediaPath) > 0 {
		evalCtx.MediaPath = evalCtx.MediaPath[:len(evalCtx.MediaPath)-1]
	}

	// Match JavaScript: return context.mediaPath.length === 0 ? media.evalTop(context) : media.evalNested(context);
	if len(evalCtx.MediaPath) == 0 {
		if os.Getenv("LESS_GO_TRACE") != "" || os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MEDIA.Eval] Calling evalTop, mediaPath len=%d, mediaBlocks count: %d\n", len(evalCtx.MediaPath), len(evalCtx.MediaBlocks))
		}
		result := media.EvalTop(evalCtx)
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MEDIA.Eval] evalTop returned type=%T\n", result)
			if rs, ok := result.(*Ruleset); ok {
				fmt.Fprintf(os.Stderr, "[MEDIA.Eval]   Ruleset: MultiMedia=%v, Root=%v, Rules count=%d\n", rs.MultiMedia, rs.Root, len(rs.Rules))
			}
		}
		return result, nil
	} else {
		if os.Getenv("LESS_GO_TRACE") != "" || os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MEDIA.Eval] Calling evalNested, mediaPath length: %d\n", len(evalCtx.MediaPath))
		}
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[MEDIA.Eval] Calling evalNested\n")
		}
		result := media.EvalNested(evalCtx)
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			featStr := "nil"
			if media.Features != nil {
				if val, ok := media.Features.(*Value); ok && len(val.Value) > 0 {
					featStr = fmt.Sprintf("Value with %d items", len(val.Value))
				} else {
					featStr = fmt.Sprintf("%T", media.Features)
				}
			}
			fmt.Fprintf(os.Stderr, "[MEDIA.Eval] After evalNested, media.features=%s, result type=%T\n", featStr, result)
		}
		return result, nil
	}
}

// deeplyEvaluateFeatures recursively evaluates all nodes in features to ensure
// variables are fully resolved. This is critical for media queries created from
// namespace calls where variables from mixin scope need to be resolved.
func (m *Media) deeplyEvaluateFeatures(features any, context any) any {
	if features == nil {
		return features
	}

	// For Value nodes, evaluate each value in the array
	if valueNode, ok := features.(*Value); ok {
		if len(valueNode.Value) == 0 {
			return features
		}
		evaluatedValues := make([]any, len(valueNode.Value))
		for i, val := range valueNode.Value {
			evaluatedValues[i] = m.deeplyEvaluateFeatures(val, context)
		}
		newValue, _ := NewValue(evaluatedValues)
		return newValue
	}

	// For Expression nodes, evaluate each value
	if exprNode, ok := features.(*Expression); ok {
		evaluatedValues := make([]any, len(exprNode.Value))
		for i, val := range exprNode.Value {
			evaluatedValues[i] = m.deeplyEvaluateFeatures(val, context)
		}
		newExpr, _ := NewExpression(evaluatedValues, exprNode.NoSpacing)
		if newExpr != nil {
			newExpr.Parens = exprNode.Parens
			newExpr.ParensInOp = exprNode.ParensInOp
		}
		return newExpr
	}

	// For Paren nodes, evaluate the inner value
	if parenNode, ok := features.(*Paren); ok {
		evaluatedValue := m.deeplyEvaluateFeatures(parenNode.Value, context)
		return NewParen(evaluatedValue)
	}

	// For Anonymous nodes, evaluate the inner value if it's evaluable
	if anonNode, ok := features.(*Anonymous); ok {
		if anonNode.Value != nil {
			evaluatedValue := m.deeplyEvaluateFeatures(anonNode.Value, context)
			// If the evaluation changed the value, create a new Anonymous with the evaluated value
			if evaluatedValue != anonNode.Value {
				return NewAnonymous(evaluatedValue, anonNode.Index, anonNode.FileInfo, anonNode.MapLines, anonNode.RulesetLike, anonNode.VisibilityInfo())
			}
		}
		return features
	}

	// For Variable nodes, try to evaluate them
	if _, ok := features.(*Variable); ok {
		if evalNode, ok := features.(interface{ Eval(any) (any, error) }); ok {
			if result, err := evalNode.Eval(context); err == nil && result != nil {
				// Recursively evaluate the result in case it contains more variables
				return m.deeplyEvaluateFeatures(result, context)
			}
		}
		return features
	}

	// For any other Evaluable node, try to evaluate
	if evalNode, ok := features.(interface{ Eval(any) (any, error) }); ok {
		if result, err := evalNode.Eval(context); err == nil && result != nil {
			return result
		}
	} else if evalNode, ok := features.(interface{ Eval(any) any }); ok {
		if result := evalNode.Eval(context); result != nil {
			return result
		}
	}

	return features
}

// evalWithMapContext handles evaluation with map-based context (for backward compatibility)
func (m *Media) evalWithMapContext(ctx map[string]any) (any, error) {
	// Match JavaScript: if (!context.mediaBlocks) { context.mediaBlocks = []; context.mediaPath = []; }
	if ctx["mediaBlocks"] == nil {
		ctx["mediaBlocks"] = []any{}
		ctx["mediaPath"] = []any{}
	}

	// Match JavaScript: const media = new Media(null, [], this._index, this._fileInfo, this.visibilityInfo())
	media := NewMedia(nil, []any{}, m.GetIndex(), m.FileInfo(), m.VisibilityInfo())

	// Match JavaScript: if (this.debugInfo) { this.rules[0].debugInfo = this.debugInfo; media.debugInfo = this.debugInfo; }
	if m.DebugInfo != nil {
		if len(m.Rules) > 0 {
			if ruleset, ok := m.Rules[0].(*Ruleset); ok {
				ruleset.DebugInfo = m.DebugInfo
			}
		}
		media.DebugInfo = m.DebugInfo
	}

	// Match JavaScript: media.features = this.features.eval(context)
	if m.Features != nil {
		if eval, ok := m.Features.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(ctx)
			if err != nil {
				return nil, err
			}
			media.Features = evaluated
		} else if eval, ok := m.Features.(interface{ Eval(any) any }); ok {
			media.Features = eval.Eval(ctx)
		}

		// CRITICAL: After evaluating features, ensure we deeply evaluate all nested values
		// This is necessary for media queries with namespace calls that return expressions
		// containing variables from the mixin scope
		media.Features = m.deeplyEvaluateFeatures(media.Features, ctx)
	}

	// Match JavaScript: context.mediaPath.push(media); context.mediaBlocks.push(media);
	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		ctx["mediaPath"] = append(mediaPath, media)
	}
	if mediaBlocks, ok := ctx["mediaBlocks"].([]any); ok {
		ctx["mediaBlocks"] = append(mediaBlocks, media)
	}

	// Match JavaScript: this.rules[0].functionRegistry = context.frames[0].functionRegistry.inherit();
	if len(m.Rules) > 0 {
		if ruleset, ok := m.Rules[0].(*Ruleset); ok {
			var frames []any
			if f, ok := ctx["frames"].([]any); ok {
				frames = f
			} else {
				return nil, fmt.Errorf("frames is required for media evaluation")
			}

			// Handle function registry inheritance if frames exist
			if len(frames) > 0 {
				if frameRuleset, ok := frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			// Match JavaScript: context.frames.unshift(this.rules[0]);
			newFrames := make([]any, len(frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], frames)
			ctx["frames"] = newFrames

			// Match JavaScript: media.rules = [this.rules[0].eval(context)];
			evaluated, err := ruleset.Eval(ctx)
			if err != nil {
				return nil, err
			}
			media.Rules = []any{evaluated}

			// Match JavaScript: context.frames.shift();
			if currentFrames, ok := ctx["frames"].([]any); ok && len(currentFrames) > 0 {
				ctx["frames"] = currentFrames[1:]
			}
		}
	}

	// Match JavaScript: context.mediaPath.pop();
	if mediaPath, ok := ctx["mediaPath"].([]any); ok && len(mediaPath) > 0 {
		ctx["mediaPath"] = mediaPath[:len(mediaPath)-1]
	}

	// Match JavaScript: return context.mediaPath.length === 0 ? media.evalTop(context) : media.evalNested(context);
	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		if len(mediaPath) == 0 {
			result := media.EvalTop(ctx)
			return result, nil
		} else {
			return media.EvalNested(ctx), nil
		}
	}

	return media.EvalTop(ctx), nil
} 