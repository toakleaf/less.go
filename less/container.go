package less_go

import (
	"fmt"
	"os"
)

type Container struct {
	*AtRule
	Features *Value
	Rules    []any
}

func NewContainer(value any, features any, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) (*Container, error) {
	selector, err := NewSelector([]any{}, nil, nil, index, currentFileInfo, nil)
	if err != nil {
		return nil, err
	}
	
	emptySelectors, err := selector.CreateEmptySelectors()
	if err != nil {
		return nil, err
	}

	selectors := make([]any, len(emptySelectors))
	for i, sel := range emptySelectors {
		selectors[i] = sel
	}

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

	atRule := NewAtRule("@container", nil, nil, index, currentFileInfo, nil, false, visibilityInfo)
	atRule.AllowRoot = true

	container := &Container{
		AtRule:   atRule,
		Features: containerFeatures,
		Rules:    []any{ruleset},
	}

	container.SetParent(selectors, container.Node)
	container.SetParent(containerFeatures.Node, container.Node)
	container.SetParent(container.Rules, container.Node)

	return container, nil
}

func (c *Container) Type() string {
	return "Container"
}

func (c *Container) GetType() string {
	return "Container"
}

func (c *Container) GetTypeIndex() int {
	return GetTypeIndexForNodeType("Container")
}

func (c *Container) GetRules() []any {
	return c.Rules
}

// Accept must be overridden because Container.Rules shadows AtRule.Rules
func (c *Container) Accept(visitor any) {
	if c.Features != nil {
		if v, ok := visitor.(interface{ Visit(any) any }); ok {
			if result := v.Visit(c.Features); result != nil {
				if features, ok := result.(*Value); ok {
					c.Features = features
				}
			}
		}
	}
	if c.Rules != nil {
		if v, ok := visitor.(interface{ VisitArray([]any, ...bool) []any }); ok {
			c.Rules = v.VisitArray(c.Rules)
		} else if v, ok := visitor.(interface{ VisitArray([]any, bool) []any }); ok {
			c.Rules = v.VisitArray(c.Rules, false)
		} else if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
			c.Rules = v.VisitArray(c.Rules)
		}
	}
}

func (c *Container) GenCSS(context any, output *CSSOutput) {
	if len(c.Rules) == 0 {
		return
	}

	if ruleset, ok := c.Rules[0].(*Ruleset); ok {
		if hasOnlyEmptyContent(ruleset.Rules) {
			return
		}
	}

	output.Add("@container ", c.FileInfo(), c.GetIndex())
	c.Features.GenCSS(context, output)
	c.OutputRuleset(context, output, c.Rules)
}

func (c *Container) Eval(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for Container.Eval")
	}

	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[CONTAINER.Eval] Starting eval\n")
	}

	var evalCtx *Eval
	if ec, ok := context.(*Eval); ok {
		evalCtx = ec
	} else if mapCtx, ok := context.(map[string]any); ok {
		return c.evalWithMapContext(mapCtx)
	} else {
		return nil, fmt.Errorf("context must be *Eval or map[string]any, got %T", context)
	}

	if evalCtx.MediaBlocks == nil {
		evalCtx.MediaBlocks = []any{}
		evalCtx.MediaPath = []any{}
	}

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

	evalCtx.MediaPath = append(evalCtx.MediaPath, media)
	evalCtx.MediaBlocks = append(evalCtx.MediaBlocks, media)

	if len(c.Rules) > 0 {
		if ruleset, ok := c.Rules[0].(*Ruleset); ok {
			if len(evalCtx.Frames) > 0 {
				if frameRuleset, ok := evalCtx.Frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			newFrames := make([]any, len(evalCtx.Frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], evalCtx.Frames)
			evalCtx.Frames = newFrames

			evaluated, err := ruleset.Eval(context)
			if err != nil {
				return nil, err
			}
			media.Rules = []any{evaluated}

			if len(evalCtx.Frames) > 0 {
				evalCtx.Frames = evalCtx.Frames[1:]
			}
		}
	}

	if len(evalCtx.MediaPath) > 0 {
		evalCtx.MediaPath = evalCtx.MediaPath[:len(evalCtx.MediaPath)-1]
	}

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

func (c *Container) evalWithMapContext(ctx map[string]any) (any, error) {
	if ctx["mediaBlocks"] == nil {
		ctx["mediaBlocks"] = []any{}
		ctx["mediaPath"] = []any{}
	}

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

	if mediaPath, ok := ctx["mediaPath"].([]any); ok {
		ctx["mediaPath"] = append(mediaPath, media)
	}
	if mediaBlocks, ok := ctx["mediaBlocks"].([]any); ok {
		ctx["mediaBlocks"] = append(mediaBlocks, media)
	}

	if len(c.Rules) > 0 {
		if ruleset, ok := c.Rules[0].(*Ruleset); ok {
			var frames []any
			if f, ok := ctx["frames"].([]any); ok {
				frames = f
			} else {
				return nil, fmt.Errorf("frames is required for container evaluation")
			}

			if len(frames) > 0 {
				if frameRuleset, ok := frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					ruleset.FunctionRegistry = frameRuleset.FunctionRegistry
				}
			}

			newFrames := make([]any, len(frames)+1)
			newFrames[0] = ruleset
			copy(newFrames[1:], frames)
			ctx["frames"] = newFrames

			evaluated, err := ruleset.Eval(ctx)
			if err != nil {
				return nil, err
			}
			media.Rules = []any{evaluated}

			if currentFrames, ok := ctx["frames"].([]any); ok && len(currentFrames) > 0 {
				ctx["frames"] = currentFrames[1:]
			}
		}
	}

	if mediaPath, ok := ctx["mediaPath"].([]any); ok && len(mediaPath) > 0 {
		ctx["mediaPath"] = mediaPath[:len(mediaPath)-1]
	}

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

func (c *Container) EvalTop(context any) any {
	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[CONTAINER.EvalTop] Starting\n")
	}

	var result any = c

	var mediaBlocks []any
	var hasMediaBlocks bool

	if evalCtx, ok := context.(*Eval); ok {
		mediaBlocks = evalCtx.MediaBlocks
		hasMediaBlocks = len(mediaBlocks) > 0

		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[CONTAINER.EvalTop] mediaBlocks count: %d\n", len(mediaBlocks))
		}

		if hasMediaBlocks && len(mediaBlocks) > 1 {
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

		if hasMediaBlocks && len(mediaBlocks) > 1 {
			selector, err := NewSelector(nil, nil, nil, c.GetIndex(), c.FileInfo(), nil)
			if err != nil {
				return result
			}
			emptySelectors, err := selector.CreateEmptySelectors()
			if err != nil {
				return result
			}

			selectors := make([]any, len(emptySelectors))
			for i, sel := range emptySelectors {
				selectors[i] = sel
			}
			ruleset := NewRuleset(selectors, mediaBlocks, false, c.VisibilityInfo())
			ruleset.MultiMedia = true
			c.SetParent(ruleset.Node, c.Node)
			result = ruleset
		}

		delete(ctx, "mediaBlocks")
		delete(ctx, "mediaPath")
	}

	return result
}

func (c *Container) EvalNested(context any) any {
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

	path := append(mediaPath, c)

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

		if container, ok := path[i].(*Container); ok {
			features = container.Features
		}
		if valueNode, ok := features.(*Value); ok {
			value = valueNode.Value
		} else {
			value = features
		}

		if arr, ok := value.([]any); ok {
			path[i] = arr
		} else {
			path[i] = []any{value}
		}
	}

	permuteResult := c.Permute(path)
	if permuteResult == nil {
		return c
	}

	permuteArray, ok := permuteResult.([]any)
	if !ok {
		return c
	}

	for _, p := range permuteArray {
		if _, ok := p.([]any); !ok {
			return c
		}
	}

	expressions := make([]any, len(permuteArray))
	for idx, pathItem := range permuteArray {
		pathArray, ok := pathItem.([]any)
		if !ok {
			continue
		}

		mappedPath := make([]any, len(pathArray))
		for i, fragment := range pathArray {
			if _, ok := fragment.(interface{ ToCSS(any) string }); ok {
				mappedPath[i] = fragment
			} else {
				mappedPath[i] = NewAnonymous(fragment, 0, nil, false, false, nil)
			}
		}

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

	newValue, err := NewValue(expressions)
	if err == nil {
		c.Features = newValue
		c.SetParent(c.Features, c.Node)
	}

	return NewRuleset([]any{}, []any{}, false, nil)
}

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