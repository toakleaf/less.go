package less_go

import (
	"fmt"
	"os"
)

// DetachedRuleset represents a detached ruleset in the Less AST
type DetachedRuleset struct {
	*Node
	ruleset any // Can be *Node or *Ruleset
	frames  []any
}

// NewDetachedRuleset creates a new DetachedRuleset instance
func NewDetachedRuleset(ruleset any, frames []any) *DetachedRuleset {
	dr := &DetachedRuleset{
		Node:    NewNode(),
		ruleset: ruleset,
		frames:  frames,
	}
	if node, ok := ruleset.(*Node); ok {
		dr.Node.SetParent(node, dr.Node)
	}
	return dr
}

// Accept implements the visitor pattern
func (dr *DetachedRuleset) Accept(visitor any) {
	// Match JavaScript: this.ruleset = visitor.visit(this.ruleset);
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		if result := v.Visit(dr.ruleset); result != nil {
			dr.ruleset = result
		}
	}
}

// Eval evaluates the detached ruleset
func (dr *DetachedRuleset) Eval(context any) any {
	// Match JavaScript: const frames = this.frames || utils.copyArray(context.frames);
	frames := dr.frames
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DetachedRuleset.Eval] dr.frames=%v (nil=%v)\n", dr.frames != nil, dr.frames == nil)
	}
	if frames == nil {
		// Copy frames from context
		switch ctx := context.(type) {
		case *Eval:
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[DetachedRuleset.Eval] *Eval context, Frames len=%d (nil=%v)\n", len(ctx.Frames), ctx.Frames == nil)
			}
			if ctx.Frames != nil {
				frames = CopyArray(ctx.Frames)
			}
		case map[string]any:
			if contextFrames, ok := ctx["frames"].([]any); ok {
				frames = CopyArray(contextFrames)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[DetachedRuleset.Eval] map context, frames len=%d\n", len(contextFrames))
				}
			}
		}
	}
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DetachedRuleset.Eval] Returning new DetachedRuleset with frames=%v (nil=%v)\n", frames != nil, frames == nil)
	}
	// Match JavaScript: return new DetachedRuleset(this.ruleset, frames);
	return NewDetachedRuleset(dr.ruleset, frames)
}

// CallEval calls eval on the ruleset with the appropriate context
func (dr *DetachedRuleset) CallEval(context any) any {
	// Match JavaScript: return this.ruleset.eval(this.frames ? new contexts.Eval(context, this.frames.concat(context.frames)) : context);

	// Debug: trace incoming mediaPath
	if os.Getenv("LESS_GO_TRACE") != "" {
		switch ctx := context.(type) {
		case *Eval:
			fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] Incoming mediaPath len=%d\n", len(ctx.MediaPath))
			for i, mp := range ctx.MediaPath {
				if m, ok := mp.(*Media); ok && m.Features != nil {
					fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval]   mediaPath[%d] features: %T\n", i, m.Features)
				}
			}
		case map[string]any:
			if mp, ok := ctx["mediaPath"].([]any); ok {
				fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] Incoming mediaPath len=%d\n", len(mp))
				for i, m := range mp {
					if media, ok := m.(*Media); ok && media.Features != nil {
						fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval]   mediaPath[%d] features: %T\n", i, media.Features)
					}
				}
			}
		}
	}

	var evalContext any = context

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] dr.frames=%v (nil=%v)\n", dr.frames != nil, dr.frames == nil)
		if ec, ok := context.(*Eval); ok {
			fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] context is *Eval, MediaBlocks=%d, MediaPath=%d\n", len(ec.MediaBlocks), len(ec.MediaPath))
		} else if mc, ok := context.(map[string]any); ok {
			mb, _ := mc["mediaBlocks"].([]any)
			mp, _ := mc["mediaPath"].([]any)
			fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] context is map, mediaBlocks=%d, mediaPath=%d\n", len(mb), len(mp))
		}
	}

	if dr.frames != nil {
		// Create concatenated frames: this.frames.concat(context.frames)
		var contextFrames []any

		switch ctx := context.(type) {
		case *Eval:
			contextFrames = ctx.Frames
			// Create new Eval with concatenated frames and MOST properties copied
			// IMPORTANT: Do NOT copy MediaBlocks and MediaPath. In JavaScript,
			// new contexts.Eval(context, frames) does NOT include these in evalCopyProperties.
			// This means the child context starts with undefined/nil mediaBlocks,
			// and Media.eval creates fresh arrays when it checks if (!context.mediaBlocks).
			// This isolation is essential for correct ordering: Media nodes inside
			// detached rulesets should NOT add to the parent's mediaBlocks during the
			// first loop of Ruleset.Eval. They should only be added when the spliced
			// result is re-evaluated in the second loop with the actual parent context.
			newEval := &Eval{
				Frames:            append(dr.frames, contextFrames...),
				Compress:          ctx.Compress,
				Math:              ctx.Math,
				StrictUnits:       ctx.StrictUnits,
				Paths:             ctx.Paths,
				SourceMap:         ctx.SourceMap,
				ImportMultiple:    ctx.ImportMultiple,
				UrlArgs:           ctx.UrlArgs,
				JavascriptEnabled: ctx.JavascriptEnabled,
				PluginManager:     ctx.PluginManager,
				ImportantScope:    ctx.ImportantScope,
				RewriteUrls:       ctx.RewriteUrls,
				CalcStack:         ctx.CalcStack,
				ParensStack:       ctx.ParensStack,
				InCalc:            ctx.InCalc,
				MathOn:            ctx.MathOn,
				DefaultFunc:       ctx.DefaultFunc,
				PluginBridge:      ctx.PluginBridge,      // Share plugin bridge for scope management
				LazyPluginBridge:  ctx.LazyPluginBridge,  // Share lazy plugin bridge
				// MediaBlocks: nil - intentionally not copied, see comment above
				// MediaPath: nil - intentionally not copied, see comment above
			}
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] Created isolated *Eval context, MediaBlocks=%v, MediaPath=%v\n", newEval.MediaBlocks, newEval.MediaPath)
			}
			evalContext = newEval
		case map[string]any:
			// Copy context keys EXCEPT mediaBlocks and mediaPath (see comment above)
			contextMap := make(map[string]any, len(ctx))
			for k, v := range ctx {
				// Skip mediaBlocks and mediaPath - child context should start fresh
				if k == "mediaBlocks" || k == "mediaPath" {
					continue
				}
				contextMap[k] = v
			}

			if frames, ok := ctx["frames"].([]any); ok {
				contextFrames = frames
			}
			newFrames := append(dr.frames, contextFrames...)
			contextMap["frames"] = newFrames
			evalContext = contextMap
		default:
			// Fallback for unknown context types
			evalContext = context
		}
	}

	// Call eval on the ruleset
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval] dr.ruleset type=%T, value=%+v\n", dr.ruleset, dr.ruleset)
	}
	if dr.ruleset != nil {
		// Check if ruleset is a Ruleset
		if ruleset, ok := dr.ruleset.(*Ruleset); ok {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval] Evaluating Ruleset with %d selectors, %d rules\n", len(ruleset.Selectors), len(ruleset.Rules))
				for i, r := range ruleset.Rules {
					fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval]   Rule %d: type=%T\n", i, r)
					if gt, ok := r.(interface{ GetType() string }); ok {
						fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval]     GetType=%s\n", gt.GetType())
					}
				}
			}

			// Convert evalContext to map for Ruleset.Eval (it expects map[string]any or *Eval)
			// The child context has its own isolated mediaBlocks/mediaPath, so Media nodes
			// inside will add to the child's arrays. These Media nodes will be returned as
			// part of the result rules and re-evaluated with the parent's context in the
			// second loop of Ruleset.Eval, where they'll properly add to the parent's mediaBlocks.
			mapContext := evalContextToMap(evalContext)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				mb, _ := mapContext["mediaBlocks"].([]any)
				mp, _ := mapContext["mediaPath"].([]any)
				fmt.Fprintf(os.Stderr, "[DetachedRuleset.CallEval] mapContext after evalContextToMap: mediaBlocks=%v (len=%d), mediaPath=%v (len=%d)\n", mapContext["mediaBlocks"], len(mb), mapContext["mediaPath"], len(mp))
			}
			result, err := ruleset.Eval(mapContext)
			if err != nil {
				// Match JavaScript behavior - throw the error
				panic(err)
			}

			// NOTE: We intentionally do NOT copy mediaBlocks/mediaPath back to the parent.
			// The child's isolated context ensures Media nodes are not prematurely added
			// to the parent's mediaBlocks. The returned rules containing Media nodes will
			// be re-evaluated in the parent's Ruleset.Eval second loop with the correct context.

			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if rs, ok := result.(*Ruleset); ok {
					fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval] Result Ruleset has %d selectors, %d rules\n", len(rs.Selectors), len(rs.Rules))
				}
			}
			return result
		}

		// Check if ruleset is a Node with Value
		if node, ok := dr.ruleset.(*Node); ok && node.Value != nil {
			// Check if the value is a Ruleset
			if ruleset, ok := node.Value.(*Ruleset); ok {
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval] Evaluating Node.Value Ruleset with %d selectors, %d rules\n", len(ruleset.Selectors), len(ruleset.Rules))
				}

				// Convert evalContext to map for Ruleset.Eval
				// The child context has its own isolated mediaBlocks/mediaPath (see comment above)
				mapContext := evalContextToMap(evalContext)
				result, err := ruleset.Eval(mapContext)
				if err != nil {
					panic(err)
				}

				// NOTE: We intentionally do NOT copy mediaBlocks/mediaPath back to the parent.

				if os.Getenv("LESS_GO_DEBUG") == "1" {
					if rs, ok := result.(*Ruleset); ok {
						fmt.Fprintf(os.Stderr, "[DEBUG DetachedRuleset.CallEval] Result Ruleset has %d selectors, %d rules\n", len(rs.Selectors), len(rs.Rules))
					}
				}
				return result
			}

			// Try single-return eval first (matches most nodes)
			if evaluator, ok := node.Value.(interface{ Eval(any) any }); ok {
				return evaluator.Eval(evalContext)
			}
			// Try double-return eval
			if evaluator, ok := node.Value.(interface{ Eval(any) (any, error) }); ok {
				result, err := evaluator.Eval(evalContext)
				if err != nil {
					// Match JavaScript behavior - throw the error
					panic(err)
				}
				return result
			}
		}

		// Try to eval the ruleset directly if it has an Eval method
		if evaluator, ok := dr.ruleset.(interface{ Eval(any) (any, error) }); ok {
			result, err := evaluator.Eval(evalContext)
			if err != nil {
				panic(err)
			}
			return result
		}

		// If nothing worked, return the ruleset itself
		return dr.ruleset
	}

	return nil
}

// evalContextToMap converts an evaluation context to map[string]any for Ruleset.Eval
func evalContextToMap(context any) map[string]any {
	switch ctx := context.(type) {
	case map[string]any:
		return ctx
	case *Eval:
		// Convert Eval to map, preserving all necessary properties
		result := map[string]any{
			"frames":            ctx.Frames,
			"compress":          ctx.Compress,
			"math":              ctx.Math,
			"strictUnits":       ctx.StrictUnits,
			"paths":             ctx.Paths,
			"sourceMap":         ctx.SourceMap,
			"importMultiple":    ctx.ImportMultiple,
			"urlArgs":           ctx.UrlArgs,
			"javascriptEnabled": ctx.JavascriptEnabled,
			"pluginManager":     ctx.PluginManager,
			"importantScope":    ctx.ImportantScope,
			"rewriteUrls":       ctx.RewriteUrls,
			"mediaBlocks":       ctx.MediaBlocks,
			"mediaPath":         ctx.MediaPath,
			"_evalContext":      ctx, // Preserve reference to *Eval for plugin scope management
		}
		// Copy plugin bridges so that detached ruleset bodies can access plugin functions
		if ctx.PluginBridge != nil {
			result["pluginBridge"] = ctx.PluginBridge
		} else if ctx.LazyPluginBridge != nil {
			result["pluginBridge"] = ctx.LazyPluginBridge
		}
		return result
	default:
		// Fallback for unknown types
		return map[string]any{
			"frames": []any{},
		}
	}
}

// Type returns the type of the node
func (dr *DetachedRuleset) Type() string {
	return "DetachedRuleset"
}

// GetType returns the type of the node for visitor pattern consistency
func (dr *DetachedRuleset) GetType() string {
	return "DetachedRuleset"
}

// EvalFirst indicates whether this node should be evaluated first
func (dr *DetachedRuleset) EvalFirst() bool {
	return true
}

// HasRuleset indicates whether this detached ruleset has an inner ruleset
// This is used by NamespaceValue to determine if it should unwrap the ruleset
func (dr *DetachedRuleset) HasRuleset() bool {
	return dr.ruleset != nil
}

// GetRuleset returns the inner ruleset for evaluation
// This is used by NamespaceValue to unwrap detached rulesets and access their variables/properties
func (dr *DetachedRuleset) GetRuleset() any {
	return dr.ruleset
} 