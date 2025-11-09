package less_go

import (
	"fmt"
	"strings"
)

// AtRule represents an at-rule node in the Less AST
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

// NewAtRule creates a new AtRule instance
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
			// Single rule - convert to array and add empty selectors
			atRule.Rules = []any{rules}
			if ruleset, ok := rules.(*Ruleset); ok {
				// Create empty selectors like JavaScript version - skip if errors occur
				emptySelector, err := NewSelector("", nil, nil, index, currentFileInfo, nil)
				if err == nil {
					emptySelectors, err := emptySelector.CreateEmptySelectors()
					if err == nil {
						ruleset.Selectors = make([]any, len(emptySelectors))
						for i, sel := range emptySelectors {
							ruleset.Selectors[i] = sel
						}
					}
				}
			}
		}

		// Set allowImports to true for all rules and parent relationships
		// Also set Root = true for the ruleset (matches JavaScript atrule.js:91)
		for _, rule := range atRule.Rules {
			if ruleset, ok := rule.(*Ruleset); ok {
				ruleset.AllowImports = true
				ruleset.Root = true
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

// Type returns the node type
func (a *AtRule) Type() string {
	return "AtRule"
}

// GetType returns the node type
func (a *AtRule) GetType() string {
	return "AtRule"
}

// GetName returns the at-rule name
func (a *AtRule) GetName() string {
	return a.Name
}

// GetDebugInfo returns debug info for the at-rule
func (a *AtRule) GetDebugInfo() any {
	return a.DebugInfo
}

// ToCSS converts the at-rule to CSS string
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

// Accept visits the node with a visitor
func (a *AtRule) Accept(visitor any) {
	if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
		if a.Rules != nil {
			a.Rules = v.VisitArray(a.Rules)
		}
	}

	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		if a.Value != nil {
			a.Value = v.Visit(a.Value)
		}
	}
}

// IsRulesetLike checks if the at-rule is ruleset-like
func (a *AtRule) IsRulesetLike() any {
	if a.Rules != nil {
		return a.Rules
	}
	return !a.IsCharset()
}

// IsCharset checks if this is a @charset rule
func (a *AtRule) IsCharset() bool {
	return a.Name == "@charset"
}

// GenCSS generates CSS representation
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
		// Check if compress mode is enabled
		compress := false
		if ctx, ok := context.(map[string]any); ok {
			if c, ok := ctx["compress"].(bool); ok {
				compress = c
			}
		}

		// Add semicolon and newline (except in compress mode)
		if compress {
			output.Add(";", nil, nil)
		} else {
			output.Add(";\n", nil, nil)
		}
	}
}

// Eval evaluates the at-rule
func (a *AtRule) Eval(context any) (any, error) {
	// If this is a non-rooted at-rule (like @document, @supports),
	// it should bubble like media queries
	if !a.IsRooted {
		return a.evalBubbling(context)
	}

	var mediaPathBackup, mediaBlocksBackup any
	var value any = a.Value
	var rules []any = a.Rules

	// Media stored inside other atrule should not bubble over it
	// Backup media bubbling information
	if ctx, ok := context.(map[string]any); ok {
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
		if eval, ok := rules[0].(interface{ Eval(any) (*Ruleset, error) }); ok {
			evaluated, err := eval.Eval(context)
			if err != nil {
				return nil, err
			}
			rules = []any{evaluated}
			evaluated.Root = true
		}
	}

	// Restore media bubbling information
	if ctx, ok := context.(map[string]any); ok {
		ctx["mediaPath"] = mediaPathBackup
		ctx["mediaBlocks"] = mediaBlocksBackup
	}

	return NewAtRule(a.Name, value, rules, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo()), nil
}

// evalBubbling evaluates non-rooted at-rules with bubbling behavior (like @document, @supports)
func (a *AtRule) evalBubbling(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for bubbling at-rules")
	}

	// Convert to *Eval context if needed
	var evalCtx *Eval
	if ec, ok := context.(*Eval); ok {
		evalCtx = ec
		return a.evalBubblingWithEvalContext(evalCtx)
	} else if mapCtx, ok := context.(map[string]any); ok {
		// For backward compatibility with map-based contexts
		return a.evalBubblingWithMapContext(mapCtx)
	} else {
		return nil, fmt.Errorf("context must be *Eval or map[string]any, got %T", context)
	}
}

// evalBubblingWithEvalContext handles bubbling evaluation with *Eval context
func (a *AtRule) evalBubblingWithEvalContext(evalCtx *Eval) (any, error) {
	// Match Media.Eval: initialize mediaBlocks and mediaPath if needed
	if evalCtx.MediaBlocks == nil {
		evalCtx.MediaBlocks = []any{}
		evalCtx.MediaPath = []any{}
	}

	// Create new at-rule instance
	atRule := NewAtRule(a.Name, nil, nil, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo())

	// Evaluate value
	if a.Value != nil {
		if eval, ok := a.Value.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(evalCtx)
			if err != nil {
				return nil, err
			}
			atRule.Value = evaluated
		} else {
			atRule.Value = a.Value
		}
	}

	// Push to mediaPath and mediaBlocks (like Media.Eval)
	evalCtx.MediaPath = append(evalCtx.MediaPath, atRule)
	evalCtx.MediaBlocks = append(evalCtx.MediaBlocks, atRule)

	// Evaluate rules
	if len(a.Rules) > 0 {
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			// Handle function registry inheritance if frames exist
			if len(evalCtx.Frames) > 0 {
				if frameRuleset, ok := evalCtx.Frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			// Match Media.Eval: context.frames.unshift(this.rules[0]);
			newFrames := make([]any, len(evalCtx.Frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], evalCtx.Frames)
			evalCtx.Frames = newFrames

			// Evaluate the ruleset
			evaluated, err := ruleset.Eval(evalCtx)
			if err != nil {
				return nil, err
			}
			atRule.Rules = []any{evaluated}

			// Match Media.Eval: context.frames.shift();
			if len(evalCtx.Frames) > 0 {
				evalCtx.Frames = evalCtx.Frames[1:]
			}
		}
	}

	// Pop from mediaPath (like Media.Eval)
	if len(evalCtx.MediaPath) > 0 {
		evalCtx.MediaPath = evalCtx.MediaPath[:len(evalCtx.MediaPath)-1]
	}

	// Return evalTop or evalNested based on mediaPath length
	if len(evalCtx.MediaPath) == 0 {
		return atRule.EvalTop(evalCtx), nil
	} else {
		return atRule.EvalNested(evalCtx), nil
	}
}

// evalBubblingWithMapContext handles bubbling evaluation with map-based context (for backward compatibility)
func (a *AtRule) evalBubblingWithMapContext(ctx map[string]any) (any, error) {
	// Match Media.Eval: initialize mediaBlocks and mediaPath if needed
	if ctx["mediaBlocks"] == nil {
		ctx["mediaBlocks"] = []any{}
		ctx["mediaPath"] = []any{}
	}

	// Create new at-rule instance
	atRule := NewAtRule(a.Name, nil, nil, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo())

	// Evaluate value
	if a.Value != nil {
		if eval, ok := a.Value.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(ctx)
			if err != nil {
				return nil, err
			}
			atRule.Value = evaluated
		} else {
			atRule.Value = a.Value
		}
	}

	// Push to mediaPath and mediaBlocks (like Media.Eval)
	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		ctx["mediaPath"] = append(mediaPath, atRule)
	}
	if mediaBlocks, ok := ctx["mediaBlocks"].([]any); ok {
		ctx["mediaBlocks"] = append(mediaBlocks, atRule)
	}

	// Evaluate rules
	if len(a.Rules) > 0 {
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			var frames []any
			if f, ok := ctx["frames"].([]any); ok {
				frames = f
			}

			// Match Media.Eval: handle function registry inheritance
			if len(frames) > 0 {
				if frameRuleset, ok := frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			// Push ruleset to frames
			newFrames := make([]any, len(frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], frames)
			ctx["frames"] = newFrames

			// Evaluate the ruleset
			evaluated, err := ruleset.Eval(ctx)
			if err != nil {
				return nil, err
			}
			atRule.Rules = []any{evaluated}

			// Pop from frames
			if currentFrames, ok := ctx["frames"].([]any); ok && len(currentFrames) > 0 {
				ctx["frames"] = currentFrames[1:]
			}
		}
	}

	// Pop from mediaPath (like Media.Eval)
	if mediaPath, ok := ctx["mediaPath"].([]any); ok && len(mediaPath) > 0 {
		ctx["mediaPath"] = mediaPath[:len(mediaPath)-1]
	}

	// Return evalTop or evalNested based on mediaPath length
	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		if len(mediaPath) == 0 {
			return atRule.EvalTop(ctx), nil
		} else {
			return atRule.EvalNested(ctx), nil
		}
	}

	return atRule.EvalTop(ctx), nil
}

// Variable returns a variable from the first rule (if rules exist)
func (a *AtRule) Variable(name string) any {
	if len(a.Rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			return ruleset.Variable(name)
		}
	}
	return nil
}

// Find finds rules matching a selector (delegates to first rule if exists)
func (a *AtRule) Find(selector any, self any, filter func(any) bool) []any {
	if len(a.Rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			return ruleset.Find(selector, self, filter)
		}
	}
	return nil
}

// Rulesets returns rulesets from the first rule (if rules exist)
func (a *AtRule) Rulesets() []any {
	if len(a.Rules) > 0 {
		// Assuming that there is only one rule at this point - that is how parser constructs the rule
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			return ruleset.Rulesets()
		}
	}
	return nil
}

// OutputRuleset outputs CSS for rules with proper formatting
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
		output.Add(" {"+tabRuleStr, nil, nil)

		// Output first rule
		if ruleCnt > 0 {
			// Set lastRule flag for the last rule (similar to JavaScript ruleset.js line 533)
			if ruleCnt == 1 {
				ctx["lastRule"] = true
			}

			if gen, ok := rules[0].(interface{ GenCSS(any, *CSSOutput) }); ok {
				gen.GenCSS(ctx, output)
			}

			if ruleCnt == 1 {
				ctx["lastRule"] = false
			}
		}

		// Output subsequent rules with indentation before each
		for i := 1; i < ruleCnt; i++ {
			output.Add(tabRuleStr, nil, nil)

			// Set lastRule flag for the last rule
			if i+1 == ruleCnt {
				ctx["lastRule"] = true
			}

			if gen, ok := rules[i].(interface{ GenCSS(any, *CSSOutput) }); ok {
				gen.GenCSS(ctx, output)
			}

			// Clear lastRule after processing
			if i+1 == ruleCnt {
				ctx["lastRule"] = false
			}
		}

		output.Add(tabSetStr+"}", nil, nil)
		// Add newline after closing brace to separate from next top-level rule
		output.Add("\n", nil, nil)
	}

	ctx["tabLevel"] = tabLevel - 1
}

// SetAllExtends sets the AllExtends field (used by ExtendFinderVisitor)
func (a *AtRule) SetAllExtends(extends []*Extend) {
	a.AllExtends = extends
}

// GetAllExtends returns the AllExtends field (used by ProcessExtendsVisitor)
func (a *AtRule) GetAllExtends() []*Extend {
	return a.AllExtends
}

// EvalTop evaluates the at-rule at the top level (for bubbling directives)
func (a *AtRule) EvalTop(context any) any {
	var result any = a

	// Handle both *Eval and map[string]any contexts
	var mediaBlocks []any
	var hasMediaBlocks bool

	if evalCtx, ok := context.(*Eval); ok {
		mediaBlocks = evalCtx.MediaBlocks
		hasMediaBlocks = len(mediaBlocks) > 0

		// Render all dependent media/at-rule blocks
		if hasMediaBlocks && len(mediaBlocks) > 1 {
			// Create empty selectors
			selector, err := NewSelector(nil, nil, nil, a.GetIndex(), a.FileInfo(), nil)
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
			ruleset := NewRuleset(selectors, mediaBlocks, false, a.VisibilityInfo())
			ruleset.MultiMedia = true // Set MultiMedia to true for multiple blocks
			ruleset.CopyVisibilityInfo(a.VisibilityInfo())
			a.SetParent(ruleset.Node, a.Node)
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

	mediaBlocks, hasMediaBlocks = ctx["mediaBlocks"].([]any)
	if hasMediaBlocks && len(mediaBlocks) > 1 {
		// Create empty selectors
		selector, err := NewSelector(nil, nil, nil, a.GetIndex(), a.FileInfo(), nil)
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
		ruleset := NewRuleset(selectors, mediaBlocks, false, a.VisibilityInfo())
		ruleset.MultiMedia = true // Set MultiMedia to true for multiple blocks
		ruleset.CopyVisibilityInfo(a.VisibilityInfo())
		a.SetParent(ruleset.Node, a.Node)
		result = ruleset
	}

	// Delete mediaBlocks and mediaPath from context
	delete(ctx, "mediaBlocks")
	delete(ctx, "mediaPath")

	return result
}

// EvalNested evaluates the at-rule in a nested context (for bubbling directives)
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

	// For non-media at-rules (like @document, @supports), we don't merge like media queries
	// We just check if types match. If they don't match, remove from mediaBlocks and return
	for i := 0; i < len(path); i++ {
		var pathType string

		// Get the type of the path item
		if atRule, ok := path[i].(*AtRule); ok {
			pathType = atRule.Name
		} else if typed, ok := path[i].(interface{ GetType() string }); ok {
			pathType = typed.GetType()
		} else {
			continue
		}

		// If types don't match, remove from mediaBlocks and return
		if pathType != a.Name {
			// Remove from mediaBlocks
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

		// For matching types, extract the value for permutation
		var value any
		if atRule, ok := path[i].(*AtRule); ok {
			value = atRule.Value
		}

		// Convert to Value if needed
		if valueNode, ok := value.(*Value); ok {
			value = valueNode.Value
		}

		// Convert to array if needed
		if arr, ok := value.([]any); ok {
			path[i] = arr
		} else {
			path[i] = []any{value}
		}
	}

	// Trace all permutations to generate the resulting at-rule value
	permuteResult := a.Permute(path)
	if permuteResult == nil {
		return a
	}

	permuteArray, ok := permuteResult.([]any)
	if !ok {
		return a
	}

	// Ensure every path is an array before mapping
	for _, p := range permuteArray {
		if _, ok := p.([]any); !ok {
			return a
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
		a.Value = newValue
		a.SetParent(a.Value, a.Node)
	}

	// Return fake tree-node that doesn't output anything
	return NewRuleset([]any{}, []any{}, false, nil)
}

// BubbleSelectors bubbles selectors up the tree (for bubbling directives)
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

	newRuleset := NewRuleset(anySelectors, []any{a.Rules[0]}, false, nil)
	a.Rules = []any{newRuleset}
	a.SetParent(a.Rules, a.Node)
}

// Permute creates permutations of the given array (for bubbling directives)
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
			return []any{} // Return empty array instead of nil
		}

		firstArray, ok := arr[0].([]any)
		if !ok {
			return []any{} // Return empty array instead of nil
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