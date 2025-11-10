package less_go

import (
	"fmt"
	"os"
)

// Container represents a CSS container at-rule node
type Container struct {
	*AtRule
	Features *Value
	Rules    []any
}

// NewContainer creates a new Container instance
func NewContainer(value any, features any, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) (*Container, error) {
	// Create empty selectors via Selector
	selector, err := NewSelector([]any{}, nil, nil, index, currentFileInfo, nil)
	if err != nil {
		return nil, err
	}
	
	emptySelectors, err := selector.CreateEmptySelectors()
	if err != nil {
		return nil, err
	}

	// Convert emptySelectors to []any for Ruleset
	selectors := make([]any, len(emptySelectors))
	for i, sel := range emptySelectors {
		selectors[i] = sel
	}

	// Create features as Value instance - handle nil case
	var containerFeatures *Value
	if features == nil {
		containerFeatures, err = NewValue([]any{})
		if err != nil {
			return nil, err
		}
	} else {
		containerFeatures, err = NewValue(features)
		if err != nil {
			return nil, err
		}
	}

	// Create rules with Ruleset - convert value to slice if needed
	var rulesetRules []any
	if value != nil {
		if valueSlice, ok := value.([]any); ok {
			rulesetRules = valueSlice
		} else {
			rulesetRules = []any{value}
		}
	}
	ruleset := NewRuleset(selectors, rulesetRules, false, visibilityInfo)
	ruleset.AllowImports = true

	// Create the base AtRule - pass nil for rules like Media does
	// This prevents NewAtRule from setting Root=true on the inner ruleset
	atRule := NewAtRule("@container", nil, nil, index, currentFileInfo, nil, false, visibilityInfo)
	atRule.AllowRoot = true

	// Create Container instance
	container := &Container{
		AtRule:   atRule,
		Features: containerFeatures,
		Rules:    []any{ruleset},  // Set rules directly, not through NewAtRule
	}

	// Set parent relationships
	container.SetParent(selectors, container.Node)
	container.SetParent(containerFeatures.Node, container.Node)
	container.SetParent(container.Rules, container.Node)

	return container, nil
}

// Type returns the node type
func (c *Container) Type() string {
	return "Container"
}

// GetType returns the node type
func (c *Container) GetType() string {
	return "Container"
}

// GetTypeIndex returns the type index for visitor pattern
func (c *Container) GetTypeIndex() int {
	return GetTypeIndexForNodeType("Container")
}

// GetRules returns the rules for this container query
func (c *Container) GetRules() []any {
	return c.Rules
}

// GenCSS generates CSS representation
func (c *Container) GenCSS(context any, output *CSSOutput) {
	// Skip container queries with empty rulesets (happens when nested container queries are merged)
	// When evalNested merges nested container queries, it returns an empty Ruleset as a placeholder
	// but the Container node itself should not be output if it has no content
	if len(c.Rules) == 0 {
		// No rules at all, skip
		return
	}

	if ruleset, ok := c.Rules[0].(*Ruleset); ok {
		// Check if the ruleset has only empty content (regardless of selectors)
		// A ruleset with selectors but no actual declarations/rules should not be output
		if hasOnlyEmptyContent(ruleset.Rules) {
			return // Skip empty container blocks
		}
	}

	output.Add("@container ", c.FileInfo(), c.GetIndex())
	c.Features.GenCSS(context, output)
	c.OutputRuleset(context, output, c.Rules)
}

// Eval evaluates the container at-rule
func (c *Container) Eval(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for Container.Eval")
	}

	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[CONTAINER.Eval] Starting eval\n")
	}

	// Convert to *Eval context if needed
	var evalCtx *Eval
	if ec, ok := context.(*Eval); ok {
		evalCtx = ec
	} else if mapCtx, ok := context.(map[string]any); ok {
		// For backward compatibility with map-based contexts
		// This path is used by EvalTop and EvalNested
		return c.evalWithMapContext(mapCtx)
	} else {
		return nil, fmt.Errorf("context must be *Eval or map[string]any, got %T", context)
	}

	// Match JavaScript: if (!context.mediaBlocks) { context.mediaBlocks = []; context.mediaPath = []; }
	if evalCtx.MediaBlocks == nil {
		evalCtx.MediaBlocks = []any{}
		evalCtx.MediaPath = []any{}
	}

	// Match JavaScript: const media = new Container(null, [], this._index, this._fileInfo, this.visibilityInfo())
	media, err := NewContainer(nil, []any{}, c.GetIndex(), c.FileInfo(), c.VisibilityInfo())
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}

	// Match JavaScript: if (this.debugInfo) { this.rules[0].debugInfo = this.debugInfo; media.debugInfo = this.debugInfo; }
	if c.DebugInfo != nil {
		if len(c.Rules) > 0 {
			if ruleset, ok := c.Rules[0].(*Ruleset); ok {
				ruleset.DebugInfo = c.DebugInfo
			}
		}
		media.DebugInfo = c.DebugInfo
	}

	// Match JavaScript: media.features = this.features.eval(context)
	if c.Features != nil {
		evaluated, err := c.Features.Eval(context)
		if err != nil {
			return nil, err
		}

		if featuresValue, ok := evaluated.(*Value); ok {
			media.Features = featuresValue
		} else {
			// If eval doesn't return a Value, wrap it
			media.Features, err = NewValue(evaluated)
			if err != nil {
				return nil, fmt.Errorf("error wrapping features in Value: %w", err)
			}
		}
	}

	// Match JavaScript: context.mediaPath.push(media); context.mediaBlocks.push(media);
	evalCtx.MediaPath = append(evalCtx.MediaPath, media)
	evalCtx.MediaBlocks = append(evalCtx.MediaBlocks, media)

	// Match JavaScript: this.rules[0].functionRegistry = context.frames[0].functionRegistry.inherit();
	if len(c.Rules) > 0 {
		if ruleset, ok := c.Rules[0].(*Ruleset); ok {
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
		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[CONTAINER.Eval] Calling evalTop, mediaBlocks count: %d\n", len(evalCtx.MediaBlocks))
		}
		return media.EvalTop(evalCtx), nil
	} else {
		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[CONTAINER.Eval] Calling evalNested, mediaPath length: %d\n", len(evalCtx.MediaPath))
		}
		return media.EvalNested(evalCtx), nil
	}
}

// evalWithMapContext handles evaluation with map-based context (for backward compatibility)
func (c *Container) evalWithMapContext(ctx map[string]any) (any, error) {
	// Match JavaScript: if (!context.mediaBlocks) { context.mediaBlocks = []; context.mediaPath = []; }
	if ctx["mediaBlocks"] == nil {
		ctx["mediaBlocks"] = []any{}
		ctx["mediaPath"] = []any{}
	}

	// Match JavaScript: const media = new Container(null, [], this._index, this._fileInfo, this.visibilityInfo())
	media, err := NewContainer(nil, []any{}, c.GetIndex(), c.FileInfo(), c.VisibilityInfo())
	if err != nil {
		return nil, fmt.Errorf("error creating container: %w", err)
	}

	// Match JavaScript: if (this.debugInfo) { this.rules[0].debugInfo = this.debugInfo; media.debugInfo = this.debugInfo; }
	if c.DebugInfo != nil {
		if len(c.Rules) > 0 {
			if ruleset, ok := c.Rules[0].(*Ruleset); ok {
				ruleset.DebugInfo = c.DebugInfo
			}
		}
		media.DebugInfo = c.DebugInfo
	}

	// Match JavaScript: media.features = this.features.eval(context)
	if c.Features != nil {
		evaluated, err := c.Features.Eval(ctx)
		if err != nil {
			return nil, err
		}

		if featuresValue, ok := evaluated.(*Value); ok {
			media.Features = featuresValue
		} else {
			// If eval doesn't return a Value, wrap it
			media.Features, err = NewValue(evaluated)
			if err != nil {
				return nil, fmt.Errorf("error wrapping features in Value: %w", err)
			}
		}
	}

	// Match JavaScript: context.mediaPath.push(media); context.mediaBlocks.push(media);
	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		ctx["mediaPath"] = append(mediaPath, media)
	}
	if mediaBlocks, ok := ctx["mediaBlocks"].([]any); ok {
		ctx["mediaBlocks"] = append(mediaBlocks, media)
	}

	// Match JavaScript: this.rules[0].functionRegistry = context.frames[0].functionRegistry.inherit();
	if len(c.Rules) > 0 {
		if ruleset, ok := c.Rules[0].(*Ruleset); ok {
			var frames []any
			if f, ok := ctx["frames"].([]any); ok {
				frames = f
			} else {
				return nil, fmt.Errorf("frames is required for container evaluation")
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

// EvalTop evaluates the container at the top level (implementing NestableAtRulePrototype)
func (c *Container) EvalTop(context any) any {
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[CONTAINER.EvalTop] Starting\n")
	}

	var result any = c

	// Handle both *Eval and map[string]any contexts
	var mediaBlocks []any
	var hasMediaBlocks bool

	if evalCtx, ok := context.(*Eval); ok {
		mediaBlocks = evalCtx.MediaBlocks
		hasMediaBlocks = len(mediaBlocks) > 0

		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[CONTAINER.EvalTop] mediaBlocks count: %d\n", len(mediaBlocks))
		}

		// Render all dependent Container blocks
		if hasMediaBlocks && len(mediaBlocks) > 1 {
			// Create empty selectors
			selector, err := NewSelector(nil, nil, nil, c.GetIndex(), c.FileInfo(), nil)
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
			ruleset := NewRuleset(selectors, mediaBlocks, false, c.VisibilityInfo())
			ruleset.MultiMedia = true // Set MultiMedia to true for multiple container blocks
			ruleset.CopyVisibilityInfo(c.VisibilityInfo())
			c.SetParent(ruleset.Node, c.Node)
			result = ruleset
		}

		// Delete mediaBlocks and mediaPath from context
		evalCtx.MediaBlocks = nil
		evalCtx.MediaPath = nil

	} else if ctx, ok := context.(map[string]any); ok {
		mediaBlocksAny, hasMediaBlocks := ctx["mediaBlocks"]
		if hasMediaBlocks {
			if blocks, ok := mediaBlocksAny.([]any); ok {
				mediaBlocks = blocks
				hasMediaBlocks = len(mediaBlocks) > 0
			}
		}

		// Render all dependent Container blocks
		if hasMediaBlocks && len(mediaBlocks) > 1 {
			// Create empty selectors
			selector, err := NewSelector(nil, nil, nil, c.GetIndex(), c.FileInfo(), nil)
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
			ruleset := NewRuleset(selectors, mediaBlocks, false, c.VisibilityInfo())
			ruleset.MultiMedia = true // Set MultiMedia to true for multiple container blocks
			c.SetParent(ruleset.Node, c.Node)
			result = ruleset
		}

		// Delete mediaBlocks and mediaPath from context
		delete(ctx, "mediaBlocks")
		delete(ctx, "mediaPath")
	}

	return result
}

// EvalNested evaluates the container in a nested context (implementing NestableAtRulePrototype)
func (c *Container) EvalNested(context any) any {
	// Handle both *Eval and map[string]any contexts
	var mediaPath []any
	var hasMediaPath bool

	if evalCtx, ok := context.(*Eval); ok {
		mediaPath = evalCtx.MediaPath
		hasMediaPath = len(mediaPath) > 0
	} else if ctx, ok := context.(map[string]any); ok {
		mediaPath, hasMediaPath = ctx["mediaPath"].([]any)
	} else {
		return c
	}

	if !hasMediaPath {
		mediaPath = []any{}
	}

	// Create path with current node
	path := append(mediaPath, c)

	// Extract the container-query conditions separated with `,` (OR)
	for i := 0; i < len(path); i++ {
		var pathType string
		switch p := path[i].(type) {
		case *Container:
			pathType = p.GetType()
		case interface{ GetType() string }:
			pathType = p.GetType()
		default:
			continue
		}

		if pathType != c.GetType() {
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
			return c
		}

		var value any
		var features any

		// Get features from the path item
		if container, ok := path[i].(*Container); ok {
			features = container.Features
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

	// Trace all permutations to generate the resulting container-query
	permuteResult := c.Permute(path)
	if permuteResult == nil {
		return c
	}

	permuteArray, ok := permuteResult.([]any)
	if !ok {
		return c
	}

	// Ensure every path is an array before mapping
	for _, p := range permuteArray {
		if _, ok := p.([]any); !ok {
			return c
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
		c.Features = newValue
		c.SetParent(c.Features, c.Node)
	}

	// Return fake tree-node that doesn't output anything
	return NewRuleset([]any{}, []any{}, false, nil)
}

// Permute creates permutations of the given array (implementing NestableAtRulePrototype)
func (c *Container) Permute(arr []any) any {
	if len(arr) == 0 {
		return []any{}
	} else if len(arr) == 1 {
		return arr[0]
	} else {
		result := []any{}
		rest := c.Permute(arr[1:])

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

// BubbleSelectors bubbles selectors up the tree (implementing NestableAtRulePrototype)
func (c *Container) BubbleSelectors(selectors any) {
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[CONTAINER.BubbleSelectors] Called with selectors: %v\n", selectors)
	}

	if selectors == nil {
		return
	}
	if len(c.Rules) == 0 {
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

	newRuleset := NewRuleset(anySelectors, []any{c.Rules[0]}, false, nil)
	c.Rules = []any{newRuleset}
	c.SetParent(c.Rules, c.Node)
} 