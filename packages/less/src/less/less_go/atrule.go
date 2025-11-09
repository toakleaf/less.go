package less_go

import (
	"fmt"
	"os"
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
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG AtRule.Eval] name=%q, hasRules=%v\n", a.Name, len(a.Rules) > 0)
	}

	// Check if this is a bubbling directive (@supports, @document)
	// These specific directives bubble to the root level like Media nodes
	// We check by name rather than just isRooted to be more explicit
	isBubblingDirective := !a.IsRooted && (a.Name == "@supports" || a.Name == "@document")

	if isBubblingDirective {
		return a.evalBubbling(context)
	}

	// Standard directives use the regular evaluation
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
		if eval, ok := rules[0].(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(context)
			if err != nil {
				return nil, err
			}
			// Convert back to Ruleset if possible
			if rs, ok := evaluated.(*Ruleset); ok {
				rules = []any{rs}
				rs.Root = true
			} else {
				rules = []any{evaluated}
			}
		}
	}

	// Restore media bubbling information
	if ctx, ok := context.(map[string]any); ok {
		ctx["mediaPath"] = mediaPathBackup
		ctx["mediaBlocks"] = mediaBlocksBackup
	}

	return NewAtRule(a.Name, value, rules, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo()), nil
}

// evalBubbling handles evaluation for non-rooted directives (like @supports, @document)
// These directives bubble up to the root level, following the same pattern as Media nodes
func (a *AtRule) evalBubbling(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for AtRule.evalBubbling")
	}

	// Route to appropriate context handler
	if evalCtx, ok := context.(*Eval); ok {
		return a.evalBubblingWithEvalContext(evalCtx)
	} else if mapCtx, ok := context.(map[string]any); ok {
		return a.evalBubblingWithMapContext(mapCtx)
	}

	return nil, fmt.Errorf("context must be *Eval or map[string]any, got %T", context)
}

// evalBubblingWithEvalContext handles bubbling evaluation with *Eval context
func (a *AtRule) evalBubblingWithEvalContext(evalCtx *Eval) (any, error) {
	// Match JavaScript/Media pattern: initialize mediaBlocks and mediaPath if needed
	if evalCtx.MediaBlocks == nil {
		evalCtx.MediaBlocks = []any{}
		evalCtx.MediaPath = []any{}
	}

	// Create new AtRule instance (like Media creates new Media)
	atRule := NewAtRule(a.Name, nil, nil, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo())

	// Set debug info
	if a.DebugInfo != nil {
		if len(a.Rules) > 0 {
			if ruleset, ok := a.Rules[0].(*Ruleset); ok {
				ruleset.DebugInfo = a.DebugInfo
			}
		}
		atRule.DebugInfo = a.DebugInfo
	}

	// Evaluate value
	if a.Value != nil {
		if eval, ok := a.Value.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(evalCtx)
			if err != nil {
				return nil, err
			}
			atRule.Value = evaluated
		} else if eval, ok := a.Value.(interface{ Eval(any) any }); ok {
			atRule.Value = eval.Eval(evalCtx)
		}
	}

	// Add to mediaPath and mediaBlocks (like Media.Eval)
	evalCtx.MediaPath = append(evalCtx.MediaPath, atRule)
	evalCtx.MediaBlocks = append(evalCtx.MediaBlocks, atRule)

	// Evaluate rules
	if len(a.Rules) > 0 {
		if ruleset, ok := a.Rules[0].(*Ruleset); ok {
			// Handle function registry inheritance
			if len(evalCtx.Frames) > 0 {
				if frameRuleset, ok := evalCtx.Frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			// Push ruleset to frames
			newFrames := make([]any, len(evalCtx.Frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], evalCtx.Frames)
			evalCtx.Frames = newFrames

			// Evaluate ruleset
			evaluated, err := ruleset.Eval(evalCtx)
			if err != nil {
				return nil, err
			}
			atRule.Rules = []any{evaluated}

			// Pop frames
			if len(evalCtx.Frames) > 0 {
				evalCtx.Frames = evalCtx.Frames[1:]
			}
		}
	}

	// Pop mediaPath
	if len(evalCtx.MediaPath) > 0 {
		evalCtx.MediaPath = evalCtx.MediaPath[:len(evalCtx.MediaPath)-1]
	}

	// Return result based on mediaPath length
	if len(evalCtx.MediaPath) == 0 {
		return atRule.EvalTop(evalCtx), nil
	} else {
		return atRule.EvalNested(evalCtx), nil
	}
}

// evalBubblingWithMapContext handles bubbling evaluation with map context
func (a *AtRule) evalBubblingWithMapContext(ctx map[string]any) (any, error) {
	// Match JavaScript/Media pattern: initialize mediaBlocks and mediaPath if needed
	if ctx["mediaBlocks"] == nil {
		ctx["mediaBlocks"] = []any{}
		ctx["mediaPath"] = []any{}
	}

	// Create new AtRule instance
	atRule := NewAtRule(a.Name, nil, nil, a.GetIndex(), a.FileInfo(), a.DebugInfo, a.IsRooted, a.VisibilityInfo())

	// Set debug info
	if a.DebugInfo != nil {
		if len(a.Rules) > 0 {
			if ruleset, ok := a.Rules[0].(*Ruleset); ok {
				ruleset.DebugInfo = a.DebugInfo
			}
		}
		atRule.DebugInfo = a.DebugInfo
	}

	// Evaluate value
	if a.Value != nil {
		if eval, ok := a.Value.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := eval.Eval(ctx)
			if err != nil {
				return nil, err
			}
			atRule.Value = evaluated
		} else if eval, ok := a.Value.(interface{ Eval(any) any }); ok {
			atRule.Value = eval.Eval(ctx)
		}
	}

	// Add to mediaPath and mediaBlocks
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
			} else {
				return nil, fmt.Errorf("frames is required for atRule evaluation")
			}

			// Handle function registry inheritance
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

			// Evaluate ruleset
			evaluated, err := ruleset.Eval(ctx)
			if err != nil {
				return nil, err
			}
			atRule.Rules = []any{evaluated}

			// Pop frames
			if currentFrames, ok := ctx["frames"].([]any); ok && len(currentFrames) > 0 {
				ctx["frames"] = currentFrames[1:]
			}
		}
	}

	// Pop mediaPath
	if mediaPath, ok := ctx["mediaPath"].([]any); ok && len(mediaPath) > 0 {
		ctx["mediaPath"] = mediaPath[:len(mediaPath)-1]
	}

	// Return result based on mediaPath length
	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		if len(mediaPath) == 0 {
			result := atRule.EvalTop(ctx)
			return result, nil
		} else {
			return atRule.EvalNested(ctx), nil
		}
	}

	return atRule.EvalTop(ctx), nil
}

// EvalTop evaluates the at-rule at the top level (implementing NestableAtRulePrototype pattern)
func (a *AtRule) EvalTop(context any) any {
	// For AtRules, we DON'T clear mediaBlocks like Media does
	// Instead, we return an empty ruleset as a placeholder
	// The directive stays in mediaBlocks and will be collected by the root ruleset
	// This is different from Media because AtRules can be nested in regular rulesets,
	// whereas Media queries handle their own nesting with permutations
	return NewRuleset([]any{}, []any{}, false, nil)
}

// EvalNested evaluates the at-rule in a nested context (implementing NestableAtRulePrototype pattern)
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

// BubbleSelectors bubbles selectors up the tree (implementing NestableAtRulePrototype pattern)
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

// Permute creates permutations of the given array (implementing NestableAtRulePrototype pattern)
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