package less_go

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
)

// Debug helper functions
func elementToString(el *Element) string {
	if el == nil {
		return "nil"
	}
	combStr := ""
	if el.Combinator != nil {
		combStr = fmt.Sprintf("[comb:%q]", el.Combinator.Value)
	}
	return fmt.Sprintf("%s%v", combStr, el.Value)
}

func elementSliceToString(els []*Element) string {
	if len(els) == 0 {
		return "[]"
	}
	parts := make([]string, len(els))
	for i, el := range els {
		parts[i] = elementToString(el)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// SelectorsParseFunc is a function type for parsing selector strings into selectors
type SelectorsParseFunc func(input string, context map[string]any, imports map[string]any, fileInfo map[string]any, index int) ([]any, error)

// ValueParseFunc is a function type for parsing value strings into values
type ValueParseFunc func(input string, context map[string]any, imports map[string]any, fileInfo map[string]any, index int) ([]any, error)

// Ruleset represents a ruleset node in the Less AST
type Ruleset struct {
	*Node
	Selectors     []any
	Rules         []any
	StrictImports bool
	AllowRoot     bool
	// Private fields for caching
	lookups    map[string][]any
	variables  map[string]any
	properties map[string][]any
	// NOTE: rulesets and variableCache were removed - caching these caused stale data
	// issues when Rules are modified during evaluation (mixin expansion, visitors, etc.)
	// Original ruleset reference for eval
	OriginalRuleset *Ruleset
	Root            bool
	// Extend support
	ExtendOnEveryPath bool
	Paths            [][]any
	FirstRoot       bool
	AllowImports    bool
	AllExtends      []*Extend // For storing extends found by ExtendFinderVisitor
	FunctionRegistry any // Changed from *functions.Registry to avoid import cycle
	// Parser functions for handling dynamic content
	SelectorsParseFunc SelectorsParseFunc
	ValueParseFunc     ValueParseFunc
	ParseContext       map[string]any
	ParseImports       map[string]any
	// Parse object matching JavaScript structure
	Parse map[string]any // Contains context and importManager
	// Debug info
	DebugInfo any
	// Multi-media flag for nested media queries
	MultiMedia bool
	// InsideMixinDefinition marks rulesets that are nested inside mixin definitions
	// These should not be output directly, only when the mixin is called
	InsideMixinDefinition bool
	// LoadedPluginFunctions stores function names loaded via @plugin in this ruleset's scope
	// This is used for function lookup when this ruleset is in the frames of a mixin call
	LoadedPluginFunctions map[string]bool
}

// OPTIMIZATION: Uses sync.Pool to reuse Ruleset objects and reduce GC pressure.
// Call Release() when the Ruleset is no longer needed to return it to the pool.
func NewRuleset(selectors []any, rules []any, strictImports bool, visibilityInfo map[string]any, parseFuncs ...any) *Ruleset {
	node := NewNode()
	node.TypeIndex = GetTypeIndexForNodeType("Ruleset")

	r := GetRulesetFromPool()
	r.Node = node
	r.StrictImports = strictImports
	r.AllowRoot = true

	// Handle selectors - keep nil if input is nil (tests expect this)
	if selectors == nil {
		r.Selectors = nil
	} else if len(selectors) == 0 {
		r.Selectors = r.Selectors[:0]
	} else {
		// Copy selectors to pooled slice
		if cap(r.Selectors) < len(selectors) {
			r.Selectors = make([]any, len(selectors))
		} else {
			r.Selectors = r.Selectors[:len(selectors)]
		}
		copy(r.Selectors, selectors)
	}

	// Handle rules - keep nil if input is nil (tests expect this)
	if rules == nil {
		r.Rules = nil
	} else if len(rules) == 0 {
		r.Rules = r.Rules[:0]
	} else {
		// Copy rules to pooled slice
		if cap(r.Rules) < len(rules) {
			r.Rules = make([]any, len(rules))
		} else {
			r.Rules = r.Rules[:len(rules)]
		}
		copy(r.Rules, rules)
	}

	r.CopyVisibilityInfo(visibilityInfo)
	r.SetParent(r.Selectors, r.Node)
	r.SetParent(r.Rules, r.Node)

	// Handle optional parse functions
	if len(parseFuncs) > 0 {
		if selectorsParseFunc, ok := parseFuncs[0].(SelectorsParseFunc); ok {
			r.SelectorsParseFunc = selectorsParseFunc
		}
	}
	if len(parseFuncs) > 1 {
		if valueParseFunc, ok := parseFuncs[1].(ValueParseFunc); ok {
			r.ValueParseFunc = valueParseFunc
		}
	}
	if len(parseFuncs) > 2 {
		if parseContext, ok := parseFuncs[2].(map[string]any); ok {
			r.ParseContext = parseContext
			// Also set in Parse object for JavaScript compatibility
			if r.Parse == nil {
				r.Parse = make(map[string]any)
			}
			r.Parse["context"] = parseContext
		}
	}
	if len(parseFuncs) > 3 {
		if parseImports, ok := parseFuncs[3].(map[string]any); ok {
			r.ParseImports = parseImports
			// Also set in Parse object for JavaScript compatibility
			if r.Parse == nil {
				r.Parse = make(map[string]any)
			}
			r.Parse["importManager"] = parseImports
		}
	}

	return r
}

func (r *Ruleset) Type() string {
	return "Ruleset"
}

func (r *Ruleset) GetType() string {
	return "Ruleset"
}

func (r *Ruleset) GetTypeIndex() int {
	// Return from Node field if set, otherwise get from registry
	if r.Node != nil && r.Node.TypeIndex != 0 {
		return r.Node.TypeIndex
	}
	return GetTypeIndexForNodeType("Ruleset")
}

func (r *Ruleset) IsRuleset() bool {
	return true
}

func (r *Ruleset) IsRulesetLike() bool {
	return true
}

// ToCSS converts the ruleset to CSS output (original signature)
func (r *Ruleset) ToCSS(options map[string]any) (string, error) {
	var output strings.Builder

	// Create context map from options
	contextMap := make(map[string]any)

	// Copy all options to the context map
	if options != nil {
		for k, v := range options {
			contextMap[k] = v
		}
	}

	// Ensure compress has a default value if not set
	if _, hasCompress := contextMap["compress"]; !hasCompress {
		contextMap["compress"] = false
	}
	
	// Create CSS output implementation
	cssOutput := &CSSOutput{
		Add: func(chunk, fileInfo, index any) {
			if chunk == nil {
				return
			}
			// Optimize: use type switch to avoid fmt.Sprintf allocation for common types
			switch v := chunk.(type) {
			case string:
				output.WriteString(v)
			case fmt.Stringer:
				output.WriteString(v.String())
			default:
				output.WriteString(fmt.Sprintf("%v", chunk))
			}
		},
		IsEmpty: func() bool {
			return output.Len() == 0
		},
	}
	
	// Generate CSS using the GenCSS method
	r.GenCSS(contextMap, cssOutput)
	
	// Return the generated CSS output
	result := output.String()

	return result, nil
}

// ToCSSString converts the ruleset to CSS output (Node interface version)
func (r *Ruleset) ToCSSString(context any) string {
	// Convert context to options map if possible
	var options map[string]any
	if ctx, ok := context.(map[string]any); ok {
		options = ctx
	}
	
	result, _ := r.ToCSS(options)
	return result
}

// Interface methods required by JoinSelectorVisitor and ToCSSVisitor

func (r *Ruleset) GetRoot() bool {
	return r.Root
}

func (r *Ruleset) SetRoot(value any) {
	// Handle both bool and any types
	oldRoot := r.Root
	if value == nil {
		r.Root = false
	} else if boolVal, ok := value.(bool); ok {
		r.Root = boolVal
	} else {
		// If value is truthy (not nil), set to true
		r.Root = true
	}
	if os.Getenv("LESS_GO_DEBUG") == "1" && r.MultiMedia {
		fmt.Fprintf(os.Stderr, "[Ruleset.SetRoot] MultiMedia Ruleset: root %v -> %v\n", oldRoot, r.Root)
	}
}

func (r *Ruleset) GetSelectors() []any {
	return r.Selectors
}

// Required by JoinSelectorVisitor
func (r *Ruleset) SetSelectors(selectors []any) {
	r.Selectors = selectors
}

func (r *Ruleset) GetPaths() []any {
	// Convert [][]any to []any for interface compatibility
	if r.Paths == nil {
		return nil
	}
	result := make([]any, len(r.Paths))
	for i, path := range r.Paths {
		result[i] = path
	}
	return result
}

func (r *Ruleset) SetPaths(paths []any) {
	// Debug: trace when paths are set
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" && len(r.Selectors) > 0 {
		if sel, ok := r.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
			if str, ok := sel.Elements[0].Value.(string); ok && str == "div" {
				fmt.Fprintf(os.Stderr, "[Ruleset.SetPaths] div ruleset=%p, setting paths=%d\n", r, len(paths))
				// Print stack trace to find caller when paths=0
				if len(paths) == 0 {
					buf := make([]byte, 2048)
					n := runtime.Stack(buf, false)
					fmt.Fprintf(os.Stderr, "[Ruleset.SetPaths] STACK TRACE:\n%s\n", buf[:n])
				}
			}
		}
	}

	// Convert []any to [][]any
	if paths == nil {
		r.Paths = nil
		return
	}
	r.Paths = make([][]any, len(paths))
	for i, path := range paths {
		if pathSlice, ok := path.([]any); ok {
			r.Paths[i] = pathSlice
		}
	}
}

// Required by ToCSSVisitor
func (r *Ruleset) GetRules() []any {
	return r.Rules
}

// Required by ToCSSVisitor
func (r *Ruleset) SetRules(rules []any) {
	r.Rules = rules
}

func (r *Ruleset) GetFirstRoot() bool {
	return r.FirstRoot
}

func (r *Ruleset) GetAllowImports() bool {
	return r.AllowImports
}

func (r *Ruleset) Accept(visitor any) {
	// Try the variadic bool version first (for the mock visitor)
	if v, ok := visitor.(interface{ VisitArray([]any, ...bool) []any }); ok {
		if r.Paths != nil {
			newPaths := make([][]any, len(r.Paths))
			for i, path := range r.Paths {
				newPaths[i] = v.VisitArray(path, true)
			}
			r.Paths = newPaths
		} else if r.Selectors != nil {
			r.Selectors = v.VisitArray(r.Selectors, false)
		}
		if len(r.Rules) > 0 {
			r.Rules = v.VisitArray(r.Rules)
		}
	} else if v, ok := visitor.(interface{ VisitArray([]any, bool) []any }); ok {
		if r.Paths != nil {
			newPaths := make([][]any, len(r.Paths))
			for i, path := range r.Paths {
				newPaths[i] = v.VisitArray(path, true)
			}
			r.Paths = newPaths
		} else if r.Selectors != nil {
			r.Selectors = v.VisitArray(r.Selectors, false)
		}
		if len(r.Rules) > 0 {
			r.Rules = v.VisitArray(r.Rules, false)
		}
	} else if v, ok := visitor.(interface{ VisitArray([]any) []any }); ok {
		if r.Paths != nil {
			newPaths := make([][]any, len(r.Paths))
			for i, path := range r.Paths {
				newPaths[i] = v.VisitArray(path)
			}
			r.Paths = newPaths
		} else if r.Selectors != nil {
			r.Selectors = v.VisitArray(r.Selectors)
		}
		if len(r.Rules) > 0 {
			r.Rules = v.VisitArray(r.Rules)
		}
	} else if v, ok := visitor.(interface{ VisitArray(any) any }); ok {
		// Handle reflection-based visitors like SetTreeVisibilityVisitor
		if r.Paths != nil {
			for _, path := range r.Paths {
				v.VisitArray(path)
			}
		} else if r.Selectors != nil {
			v.VisitArray(r.Selectors)
		}
		if len(r.Rules) > 0 {
			v.VisitArray(r.Rules)
		}
	}
}

func (r *Ruleset) Eval(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for Ruleset.Eval")
	}

	// Accept both *Eval and map[string]any contexts
	// For circular dependency tracking, we need access to a map
	var ctx map[string]any
	var evalCtx *Eval

	if ec, ok := context.(*Eval); ok {
		evalCtx = ec
		// Create a minimal map for state tracking (selectors, etc.)
		// We'll pass the *Eval context to child evaluations
		// Pre-allocate with capacity 2 since we typically only use selectors and maybe frames
		ctx = make(map[string]any, 2)
	} else if mapCtx, ok := context.(map[string]any); ok {
		ctx = mapCtx
	} else {
		return nil, fmt.Errorf("context must be *Eval or map[string]any, got %T", context)
	}

	// Enter a new plugin scope for this ruleset.
	// This ensures that any @plugin directives inside this ruleset only affect
	// this scope and its children, not parent scopes.
	// We use defer to ensure ExitPluginScope is called even if an error occurs.
	//
	// We check for the plugin bridge in both *Eval contexts and map contexts.
	// Map contexts may have _evalContext which references the parent *Eval.
	var pluginEvalCtx *Eval
	if evalCtx != nil {
		pluginEvalCtx = evalCtx
	} else if parentEval, ok := ctx["_evalContext"].(*Eval); ok {
		pluginEvalCtx = parentEval
	}
	// Check for both direct PluginBridge and LazyPluginBridge (which wraps PluginBridge)
	scopeEntered := false
	if pluginEvalCtx != nil && (pluginEvalCtx.PluginBridge != nil || pluginEvalCtx.LazyPluginBridge != nil) {
		pluginEvalCtx.EnterPluginScope()
		scopeEntered = true
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[Ruleset.Eval] Entered plugin scope, r=%p\n", r)
		}
		defer func() {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[Ruleset.Eval] Exiting plugin scope (defer), r=%p\n", r)
			}
			pluginEvalCtx.ExitPluginScope()
		}()
	}
	_ = scopeEntered // suppress unused warning

	// NOTE: JavaScript does not have circular dependency checking for ruleset evaluation.
	// Recursive mixin calls are valid and should be allowed. The check was removed because
	// it was causing false positives for valid recursive calls where the same mixin definition
	// is called multiple times with different parameters (e.g., .mixin -> .passthrough -> .mixin).

	var selectors []any
	var selCnt int
	var hasVariable bool
	var hasOnePassingSelector bool

	if len(r.Selectors) > 0 {
		selCnt = len(r.Selectors)
		selectors = make([]any, selCnt)

		// Match JavaScript: defaultFunc.error({type: 'Syntax', message: 'it is currently only allowed in parametric mixin guards,'});
		if evalCtx != nil && evalCtx.DefaultFunc != nil {
			evalCtx.DefaultFunc.Error(map[string]any{
				"type":    "Syntax",
				"message": "it is currently only allowed in parametric mixin guards,",
			})
		} else if df, ok := ctx["defaultFunc"].(interface{ Error(map[string]any) }); ok {
			df.Error(map[string]any{
				"type":    "Syntax",
				"message": "it is currently only allowed in parametric mixin guards,",
			})
		}

		// Evaluate selectors - pass the original context (either *Eval or map)
		for i := 0; i < selCnt; i++ {
			if sel, ok := r.Selectors[i].(interface{ Eval(any) (any, error) }); ok {
				evaluated, err := sel.Eval(context)
				if err != nil {
					return nil, err
				}
				selectors[i] = evaluated

				// Check for variables in elements
				if selector, ok := evaluated.(*Selector); ok {
					for _, elem := range selector.Elements {
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							fmt.Printf("[DEBUG Eval] Element IsVariable=%v, Value type=%T, Value=%#v\n",
								elem.IsVariable, elem.Value, elem.Value)
						}
						if elem.IsVariable {
							hasVariable = true
							break
						}
					}

					// Check for passing condition
					if selector.EvaldCondition {
						hasOnePassingSelector = true
					}
				}
			}
		}

		// Handle variables in selectors - parse using SelectorsParseFunc
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG Eval] hasVariable=%v, SelectorsParseFunc=%v\n", hasVariable, r.SelectorsParseFunc != nil)
		}
		if hasVariable && r.SelectorsParseFunc != nil {
			// Convert selectors to CSS strings for parsing (like JavaScript toParseSelectors)
			// Pre-allocate with capacity based on selectors length
			toParseSelectors := make([]string, 0, len(selectors))
			var startingIndex int
			var selectorFileInfo map[string]any
			// Reuse toCSSCtx map across iterations to reduce allocations
			toCSSCtx := map[string]any{"firstSelector": true}

			for i, sel := range selectors {
				if selector, ok := sel.(*Selector); ok {
					// Get CSS representation of selector
					// Pass firstSelector=true to avoid leading spaces
					cssStr := selector.ToCSS(toCSSCtx)
					toParseSelectors = append(toParseSelectors, cssStr)

					if i == 0 {
						startingIndex = selector.GetIndex()
						selectorFileInfo = selector.FileInfo()
					}
				}
			}
			
			if len(toParseSelectors) > 0 {
				// Parse the selectors string (equivalent to JS parseNode call)
				selectorString := strings.Join(toParseSelectors, ",")
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG Eval] Re-parsing selectors: %s\n", selectorString)
				}
				parsedSelectors, err := r.SelectorsParseFunc(selectorString, r.ParseContext, r.ParseImports, selectorFileInfo, startingIndex)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG Eval] ParseSelectors returned: err=%v, len=%d\n", err, len(parsedSelectors))
				}
				if err == nil && len(parsedSelectors) > 0 {
					// Flatten the result (equivalent to utils.flattenArray in JS)
					selectors = flattenArray(parsedSelectors)
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Printf("[DEBUG Eval] Parsed %d selectors after flatten\n", len(selectors))
						for i, sel := range selectors {
							if s, ok := sel.(*Selector); ok {
								fmt.Printf("[DEBUG Eval]   Selector %d: %d elements\n", i, len(s.Elements))
								for j, el := range s.Elements {
									fmt.Printf("[DEBUG Eval]     Element %d: Value=%#v, Type=%T\n", j, el.Value, el.Value)
								}
							}
						}
					}
				}
			}
		}
		
		// Match JavaScript: defaultFunc.reset();
		if evalCtx != nil && evalCtx.DefaultFunc != nil {
			evalCtx.DefaultFunc.Reset()
		} else if df, ok := ctx["defaultFunc"].(interface{ Reset() }); ok {
			df.Reset()
		}
	} else {
		hasOnePassingSelector = true
	}

	// Copy rules
	var rules []any
	if r.Rules != nil {
		rules = CopyArray(r.Rules)
	}

	// Create new ruleset
	ruleset := NewRuleset(selectors, rules, r.StrictImports, r.VisibilityInfo(), r.SelectorsParseFunc, r.ValueParseFunc, r.ParseContext, r.ParseImports)
	ruleset.OriginalRuleset = r

	// Debug: trace creation of div rulesets
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" && len(selectors) > 0 {
		if sel, ok := selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
			elemVal := sel.Elements[0].Value
			if str, ok := elemVal.(string); ok && str == "div" {
				fmt.Fprintf(os.Stderr, "[Ruleset.Eval] Created div ruleset=%p from r=%p\n", ruleset, r)
				fmt.Fprintf(os.Stderr, "[Ruleset.Eval]   r.Node=%p, r.BlocksVisibility=%v\n", r.Node, r.Node.BlocksVisibility())
				fmt.Fprintf(os.Stderr, "[Ruleset.Eval]   ruleset.Node=%p, ruleset.BlocksVisibility=%v\n", ruleset.Node, ruleset.Node.BlocksVisibility())
			}
		}
	}
	ruleset.Root = r.Root
	ruleset.FirstRoot = r.FirstRoot
	ruleset.AllowImports = r.AllowImports
	ruleset.MultiMedia = r.MultiMedia

	if r.DebugInfo != nil {
		ruleset.DebugInfo = r.DebugInfo
	}

	// Match JavaScript: if (!hasOnePassingSelector) { rules.length = 0; }
	if !hasOnePassingSelector {
		// Clear the rules in the newly created ruleset
		ruleset.Rules = []any{}
	}

	// inherit a function registry from the frames stack when possible;
	// otherwise from the global registry
	// Match JavaScript: ruleset.functionRegistry = (function (frames) {...}(context.frames)).inherit();
	var functionRegistry any
	var frames []any

	// Get frames from the appropriate context type
	if evalCtx != nil {
		frames = evalCtx.Frames
	} else if framesVal, ok := ctx["frames"].([]any); ok {
		frames = framesVal
	}

	// Check evalCtx for FunctionRegistry first
	if evalCtx != nil && evalCtx.FunctionRegistry != nil {
		functionRegistry = evalCtx.FunctionRegistry
	} else if len(frames) > 0 {
		// Fall back to checking frames
		for _, frame := range frames {
			if f, ok := frame.(*Ruleset); ok {
				if f.FunctionRegistry != nil {
					functionRegistry = f.FunctionRegistry
					break
				}
			}
		}
	}

	// Apply .inherit() pattern like JavaScript
	if functionRegistry != nil {
		if inheritRegistry, ok := functionRegistry.(interface{ Inherit() any }); ok {
			ruleset.FunctionRegistry = inheritRegistry.Inherit()
		} else {
			// Fallback if Inherit method not available
			ruleset.FunctionRegistry = functionRegistry
		}
	}
	// If no function registry found in frames, leave it nil for now

	// Push current ruleset to frames stack
	newFrames := make([]any, len(frames)+1)
	newFrames[0] = ruleset
	copy(newFrames[1:], frames)

	// Update frames in the appropriate context type
	if evalCtx != nil {
		evalCtx.Frames = newFrames
	} else {
		ctx["frames"] = newFrames
	}

	// Current selectors - store in map for both context types
	if selectors := ctx["selectors"]; selectors == nil {
		ctx["selectors"] = []any{r.Selectors}
	} else if sels, ok := selectors.([]any); ok {
		newSelectors := make([]any, len(sels)+1)
		newSelectors[0] = r.Selectors
		copy(newSelectors[1:], sels)
		ctx["selectors"] = newSelectors
	}

	// Ensure function registry is available in context
	if evalCtx != nil {
		// For *Eval context, set FunctionRegistry if not already set
		if evalCtx.FunctionRegistry == nil && ruleset.FunctionRegistry != nil {
			if registry, ok := ruleset.FunctionRegistry.(*Registry); ok {
				evalCtx.FunctionRegistry = registry
			}
		}
	} else {
		// For map context, set functionRegistry if not already set
		if _, exists := ctx["functionRegistry"]; !exists && ruleset.FunctionRegistry != nil {
			ctx["functionRegistry"] = ruleset.FunctionRegistry
		}
	}

	// Evaluate imports
	if ruleset.Root || ruleset.AllowImports || !ruleset.StrictImports {
		err := ruleset.EvalImports(context)
		if err != nil {
			return nil, err
		}
	}

	// Run pre-eval visitors after imports are processed (plugins are loaded)
	// This should only run for the root ruleset to avoid multiple runs
	if ruleset.Root && pluginEvalCtx != nil && pluginEvalCtx.LazyPluginBridge != nil {
		if pluginEvalCtx.LazyPluginBridge.IsInitialized() {
			if bridge, err := pluginEvalCtx.LazyPluginBridge.GetBridge(); err == nil && bridge.HasPreEvalVisitors() {
				runPreEvalVisitorReplacementsOnRuleset(ruleset, bridge)
			}
		}
	}

	// Pre-warm plugin function cache after imports are processed
	// This batches all plugin function calls to reduce IPC overhead
	if ruleset.Root && pluginEvalCtx != nil && pluginEvalCtx.LazyPluginBridge != nil {
		if pluginEvalCtx.LazyPluginBridge.IsInitialized() {
			warmCount, warmErr := WarmPluginCacheFromLazyBridge(ruleset, pluginEvalCtx.LazyPluginBridge, pluginEvalCtx)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if warmErr != nil {
					fmt.Fprintf(os.Stderr, "[Ruleset.Eval] Cache pre-warming failed: %v\n", warmErr)
				} else if warmCount > 0 {
					fmt.Fprintf(os.Stderr, "[Ruleset.Eval] Pre-warmed %d plugin function cache entries\n", warmCount)
				}
			}
		}
	}

	// Evaluate rules that need to be evaluated first
	rsRules := ruleset.Rules
	for i, rule := range rsRules {
		if r, ok := rule.(interface{ EvalFirst() bool }); ok && r.EvalFirst() {
			// Handle MixinDefinition specifically (returns *MixinDefinition, not any)
			if eval, ok := rule.(interface{ Eval(any) (*MixinDefinition, error) }); ok {
				evaluated, err := eval.Eval(context)
				if err != nil {
					return nil, err
				}
				rsRules[i] = evaluated
			} else if eval, ok := rule.(interface{ Eval(any) (any, error) }); ok {
				evaluated, err := eval.Eval(context)
				if err != nil {
					return nil, err
				}
				rsRules[i] = evaluated
			}
		}
	}

	// Track media blocks
	mediaBlockCount := 0
	if evalCtx != nil && evalCtx.MediaBlocks != nil {
		mediaBlockCount = len(evalCtx.MediaBlocks)
	} else if mediaBlocks, ok := ctx["mediaBlocks"].([]any); ok {
		mediaBlockCount = len(mediaBlocks)
	}

	if os.Getenv("LESS_GO_TRACE") != "" {
		firstSelContent := "none"
		if len(selectors) > 0 {
			if s, ok := selectors[0].(*Selector); ok && len(s.Elements) > 0 {
				firstSelContent = fmt.Sprintf("%v", s.Elements[0].Value)
			}
		}
		fmt.Fprintf(os.Stderr, "[RULESET.Eval] START: firstSel=%s, mediaBlockCount=%d, Root=%v\n",
			firstSelContent, mediaBlockCount, r.Root)
	}

	// Evaluate mixin calls and variable calls - match JavaScript logic closely
	if rsRules != nil {
		i := 0
		for i < len(rsRules) {
			rule := rsRules[i]
			if r, ok := rule.(interface{ GetType() string }); ok {
				switch r.GetType() {
				case "MixinCall":
					if eval, ok := rule.(interface{ Eval(any) ([]any, error) }); ok {
						rules, err := eval.Eval(context)
						if err != nil {
							return nil, err
						}
						// Match JavaScript filter logic: !(ruleset.variable(r.name))
						filtered := make([]any, 0, len(rules))
						for _, r := range rules {
							if decl, ok := r.(*Declaration); ok && decl.variable {
								// Match JavaScript: return !(ruleset.variable(r.name))
								if nameStr, ok := decl.name.(string); ok {
									if ruleset.Variable(nameStr) == nil {
										filtered = append(filtered, r) // Include if variable doesn't exist
									}
									// Skip if variable already exists (don't pollute scope)
								} else {
									filtered = append(filtered, r)
								}
							} else {
								filtered = append(filtered, r)
							}
						}
						
						// rsRules.splice.apply(rsRules, [i, 1].concat(rules))
						newRules := make([]any, len(rsRules)+len(filtered)-1)
						copy(newRules, rsRules[:i])
						copy(newRules[i:], filtered)
						copy(newRules[i+len(filtered):], rsRules[i+1:])
						rsRules = newRules
						ruleset.Rules = rsRules
						i += len(filtered) - 1
						ruleset.ResetCache()
					}
				case "VariableCall":
					if eval, ok := rule.(interface{ Eval(any) (any, error) }); ok {
						evaluated, err := eval.Eval(context)
						if err != nil {
							return nil, err
						}
						// Handle the result - it could be a map with "rules" key or a Ruleset
						var evalRules []any
						if evalMap, ok := evaluated.(map[string]any); ok {
							if rules, hasRules := evalMap["rules"].([]any); hasRules {
								evalRules = rules
							}
						} else if rs, ok := evaluated.(*Ruleset); ok {
							evalRules = rs.Rules
						}

						if evalRules != nil {
							// Match JavaScript: filter out all variable declarations
							rules := make([]any, 0, len(evalRules))
							for _, r := range evalRules {
								if decl, ok := r.(*Declaration); ok && decl.variable {
									// do not pollute the scope at all
									continue
								}
								rules = append(rules, r)
							}

							// NOTE: We no longer add mediaBlocks to rules here.
							// MediaBlocks are properly propagated through context via MixinDefinition.EvalCall
							// and DetachedRuleset.CallEval, and output via evalTop's MultiMedia Ruleset.

							// rsRules.splice.apply(rsRules, [i, 1].concat(rules))
							newRules := make([]any, len(rsRules)+len(rules)-1)
							copy(newRules, rsRules[:i])
							copy(newRules[i:], rules)
							copy(newRules[i+len(rules):], rsRules[i+1:])
							rsRules = newRules
							ruleset.Rules = rsRules
							i += len(rules) - 1
							ruleset.ResetCache()
						}
					}
				}
			}
			i++
		}
	}

	// Evaluate everything else
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[Ruleset.Eval] Evaluating %d rules, r=%p\n", len(rsRules), r)
	}
	for i, rule := range rsRules {
		if r, ok := rule.(interface{ EvalFirst() bool }); ok && r.EvalFirst() {
			continue // Already evaluated
		}

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[Ruleset.Eval]   Rule %d type=%T, r=%p\n", i, rule, r)
		}

		// Try different Eval signatures
		switch evalRule := rule.(type) {
		case interface{ Eval(any) (*MixinDefinition, error) }:
			// Handle MixinDefinition
			evaluated, err := evalRule.Eval(context)
			if err != nil {
				return nil, err
			}
			rsRules[i] = evaluated
		case interface{ Eval(any) (any, error) }:
			// Handle generic Eval
			evaluated, err := evalRule.Eval(context)
			if err != nil {
				return nil, err
			}
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				if _, isMedia := rule.(*Media); isMedia {
					fmt.Fprintf(os.Stderr, "[RULESET.Eval] Media node evaluated to type=%T\n", evaluated)
					if rs, ok := evaluated.(*Ruleset); ok {
						fmt.Fprintf(os.Stderr, "[RULESET.Eval]   Ruleset MultiMedia=%v, Rules=%d\n", rs.MultiMedia, len(rs.Rules))
					}
				}
			}
			rsRules[i] = evaluated
		case interface{ Eval(any) any }:
			// Handle Eval without error return
			rsRules[i] = evalRule.Eval(context)
		}
	}

	// CRITICAL FIX: Store evaluated rules back to ruleset
	// Without this, all rule modifications during evaluation (including Media nodes from inline imports) are lost
	ruleset.Rules = rsRules

	// Reset cache after evaluating rules since variable values may have changed
	ruleset.ResetCache()

	// Handle parent selector folding like JavaScript version
	i := 0
	for i < len(rsRules) {
		rule := rsRules[i]
		// for rulesets, check if it is a css guard and can be removed
		if rs, ok := rule.(*Ruleset); ok && len(rs.Selectors) == 1 {
			// check if it can be folded in (e.g. & where)
			if rs.Selectors[0] != nil {
				if sel, ok := rs.Selectors[0].(*Selector); ok && sel.IsJustParentSelector() {
					// Remove the parent ruleset
					rsRules = r.removeRuleAtIndex(rsRules, i)
					ruleset.Rules = rsRules
					i--

					// Add the sub rules
					for _, subRule := range rs.Rules {
						if r.shouldIncludeSubRule(subRule, rs) {
							i++
							rsRules = r.insertRuleAtIndex(rsRules, i, subRule)
							ruleset.Rules = rsRules
						}
					}
				}
			}
		}
		i++
	}

	// Pop the stack
	if evalCtx != nil {
		// Pop from *Eval context
		if len(evalCtx.Frames) > 0 {
			evalCtx.Frames = evalCtx.Frames[1:]
		}
	} else {
		// Pop from map context
		if frames, ok := ctx["frames"].([]any); ok && len(frames) > 0 {
			ctx["frames"] = frames[1:]
		}
	}
	if selectors, ok := ctx["selectors"].([]any); ok && len(selectors) > 0 {
		ctx["selectors"] = selectors[1:]
	}

	// Handle media blocks - check both *Eval and map contexts
	var mediaBlocks []any
	if evalCtx != nil && evalCtx.MediaBlocks != nil {
		mediaBlocks = evalCtx.MediaBlocks
	} else if mb, ok := ctx["mediaBlocks"].([]any); ok {
		mediaBlocks = mb
	}

	if os.Getenv("LESS_GO_TRACE") != "" {
		mbLen := 0
		if mediaBlocks != nil {
			mbLen = len(mediaBlocks)
		}
		// Get first selector content for identification
		firstSelContent := "none"
		if len(selectors) > 0 {
			if s, ok := selectors[0].(*Selector); ok && len(s.Elements) > 0 {
				firstSelContent = fmt.Sprintf("%v", s.Elements[0].Value)
			}
		}
		fmt.Fprintf(os.Stderr, "[RULESET.Eval] BubbleSelectors section: mediaBlockCount=%d, len(mediaBlocks)=%d, len(selectors)=%d, firstSel=%s, Root=%v\n",
			mediaBlockCount, mbLen, len(selectors), firstSelContent, ruleset.Root)
	}

	if mediaBlocks != nil {
		// Skip BubbleSelectors if all selectors are MediaEmpty (placeholder selectors from inner media rulesets)
		// MediaEmpty selectors are just `&` placeholders created by Media constructor - they should not bubble
		allMediaEmpty := true
		for _, sel := range selectors {
			if s, ok := sel.(*Selector); ok && !s.MediaEmpty {
				allMediaEmpty = false
				break
			}
		}

		if !allMediaEmpty {
			for i := mediaBlockCount; i < len(mediaBlocks); i++ {
				if mb, ok := mediaBlocks[i].(interface{ BubbleSelectors(any) }); ok {
					if os.Getenv("LESS_GO_TRACE") != "" {
						fmt.Fprintf(os.Stderr, "[RULESET.Eval] Calling BubbleSelectors on mediaBlock %d (type=%T)\n", i, mb)
					}
					mb.BubbleSelectors(selectors)
				}
			}
		} else if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[RULESET.Eval] Skipping BubbleSelectors - all selectors are MediaEmpty\n")
		}
	}

	// CRITICAL: If this is the file-level root ruleset, append any bubbled mediaBlocks to Rules
	// This is how directives like @supports and @document bubble to the top level
	// Match JavaScript behavior: after evaluation, mediaBlocks at root get appended to rules
	// Only do this when mediaPath is empty (we're at the top level, not nested in a @media/@container)
	var mediaPath []any
	if evalCtx != nil {
		mediaPath = evalCtx.MediaPath
	} else {
		if mp, ok := ctx["mediaPath"].([]any); ok {
			mediaPath = mp
		}
	}

	if os.Getenv("LESS_GO_TRACE") != "" {
		fmt.Fprintf(os.Stderr, "[RULESET.Eval] Root check: Root=%v, len(mediaPath)=%d, len(mediaBlocks)=%d\n",
			ruleset.Root, len(mediaPath), func() int { if mediaBlocks != nil { return len(mediaBlocks) } else { return 0 } }())
	}

	if ruleset.Root && len(mediaPath) == 0 && mediaBlocks != nil && len(mediaBlocks) > 0 {
		if os.Getenv("LESS_GO_TRACE") != "" {
			fmt.Fprintf(os.Stderr, "[RULESET.Eval] Processing %d mediaBlocks for root Rules\n", len(mediaBlocks))
		}

		// Replace empty rulesets (placeholders from bubbling directives) with actual mediaBlocks
		// This maintains the original order of all directives
		mediaBlockIndex := 0
		newRules := make([]any, 0, len(ruleset.Rules))

		for _, rule := range ruleset.Rules {
			// Check if this is an empty placeholder ruleset from a bubbling directive
			if rs, ok := rule.(*Ruleset); ok {
				if len(rs.Selectors) == 0 && len(rs.Rules) == 0 && mediaBlockIndex < len(mediaBlocks) {
					// Replace the empty placeholder with the corresponding mediaBlock
					newRules = append(newRules, mediaBlocks[mediaBlockIndex])
					mediaBlockIndex++
					continue
				}
			}
			newRules = append(newRules, rule)
		}

		// If there are any remaining mediaBlocks (shouldn't happen in normal cases), append them
		for mediaBlockIndex < len(mediaBlocks) {
			newRules = append(newRules, mediaBlocks[mediaBlockIndex])
			mediaBlockIndex++
		}

		ruleset.Rules = newRules

		// Clear mediaBlocks from context (they've been consumed)
		if evalCtx != nil {
			evalCtx.MediaBlocks = nil
			evalCtx.MediaPath = nil
		} else {
			ctx["mediaBlocks"] = nil
			ctx["mediaPath"] = nil
		}
	}

	return ruleset, nil
}

// EvalImports evaluates import rules like JavaScript version
func (r *Ruleset) EvalImports(context any) error {
	rules := r.Rules
	var i int
	var importRules any

	if rules == nil {
		return nil
	}

	debug := os.Getenv("LESS_GO_DEBUG") == "1"
	if debug {
		fmt.Fprintf(os.Stderr, "[EvalImports] Processing %d rules\n", len(rules))
	}

	for i = 0; i < len(rules); i++ {
		if ruleType, ok := rules[i].(interface{ GetType() string }); ok && ruleType.GetType() == "Import" {
			if debug {
				fmt.Fprintf(os.Stderr, "[EvalImports] Found Import at index %d\n", i)
			}
			if eval, ok := rules[i].(interface{ Eval(any) (any, error) }); ok {
				evaluated, err := eval.Eval(context)
				if err != nil {
					return err
				}
				importRules = evaluated

				// Handle different return types like JavaScript version
				if importRulesSlice, ok := importRules.([]any); ok {
					// if (importRules && (importRules.length || importRules.length === 0))
					if len(importRulesSlice) > 0 || len(importRulesSlice) == 0 {
						// Replace import rule with its evaluated result
						newRules := make([]any, len(rules)+len(importRulesSlice)-1)
						copy(newRules, rules[:i])
						copy(newRules[i:], importRulesSlice)
						copy(newRules[i+len(importRulesSlice):], rules[i+1:])
						rules = newRules
						r.Rules = rules
						i += len(importRulesSlice) - 1
					} else {
						// rules.splice(i, 1, importRules);
						rules[i] = importRules
					}
				} else {
					// rules.splice(i, 1, importRules);
					rules[i] = importRules
				}
				r.ResetCache()
			}
		}
	}
	return nil
}

func (r *Ruleset) MakeImportant() any {
	var newRules []any
	if r.Rules != nil {
		newRules = make([]any, len(r.Rules))
		for i, rule := range r.Rules {
			if imp, ok := rule.(interface{ MakeImportant() any }); ok {
				newRules[i] = imp.MakeImportant()
			} else {
				newRules[i] = rule
			}
		}
	}

	return NewRuleset(r.Selectors, newRules, r.StrictImports, r.VisibilityInfo())
}

func (r *Ruleset) MatchArgs(args []any) bool {
	return len(args) == 0
}

func (r *Ruleset) MatchCondition(args []any, context any) bool {
	if len(r.Selectors) == 0 {
		return false
	}
	
	lastSelector := r.Selectors[len(r.Selectors)-1]
	
	// Check evaldCondition
	if sel, ok := lastSelector.(*Selector); ok {
		if !sel.EvaldCondition {
			return false
		}
		
		// Check condition
		if sel.Condition != nil {
			if eval, ok := sel.Condition.(interface{ Eval(any) (any, error) }); ok {
				// Create new context for evaluation like JavaScript version
				ctx := make(map[string]any)
				if c, ok := context.(map[string]any); ok {
					for k, v := range c {
						ctx[k] = v
					}
				}
				if frames, ok := ctx["frames"].([]any); ok {
					ctx["frames"] = frames
				}
				
				// IMPORTANT: Preserve the defaultFunc in the context
				// This is needed for the default() function in mixin guards
				if evalCtx, ok := context.(EvalContext); ok {
					if df := evalCtx.GetDefaultFunc(); df != nil {
						ctx["defaultFunc"] = df
					}
				} else if c, ok := context.(map[string]any); ok {
					if df, exists := c["defaultFunc"]; exists {
						ctx["defaultFunc"] = df
					}
				}
				
				result, err := eval.Eval(ctx)
				if err != nil {
					return false
				}
				
				// Check if result is falsy
				if isFalsy(result) {
					return false
				}
			}
		}
	}
	
	return true
}

// Helper function to check if a value is falsy (like JavaScript)
func isFalsy(v any) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case bool:
		return !val
	case int:
		return val == 0
	case float64:
		return val == 0
	case string:
		return val == ""
	default:
		return false
	}
}

// ruleBlocksVisibility checks if a rule blocks visibility and is not explicitly visible
// This is used to skip rules from reference imports during CSS generation
func ruleBlocksVisibility(rule any) bool {
	// Check if the rule has visibility blocking methods
	if visNode, ok := rule.(interface{ BlocksVisibility() bool; IsVisible() *bool }); ok {
		if visNode.BlocksVisibility() {
			nodeVisible := visNode.IsVisible()
			// If visibility is nil or false, check for visible paths from extends
			if nodeVisible == nil || !*nodeVisible {
				// For rulesets, check if they have visible paths (from extends)
				// This matches the logic in ToCSSVisitor.IsVisibleRuleset
				if rs, isRuleset := rule.(*Ruleset); isRuleset {
					// Check if ruleset has visible paths with visible selectors
					if rs.Paths != nil && len(rs.Paths) > 0 {
						for _, path := range rs.Paths {
							if len(path) > 0 {
								// Check if any selector in the path is explicitly visible
								for _, pathElem := range path {
									if sel, ok := pathElem.(*Selector); ok {
										if sel.Node != nil {
											selVis := sel.Node.IsVisible()
											if selVis != nil && *selVis {
												// Has a visible selector from extends - don't filter
												return false
											}
										}
									}
								}
							}
						}
					}
				}
				// No visible paths, filter it out
				return true
			}
		}
	}
	return false
}

// isExtendOnlyRuleset checks if a ruleset contains only extend declarations.
// Such rulesets don't output any CSS content and shouldn't add newlines.
func isExtendOnlyRuleset(rs *Ruleset) bool {
	if rs == nil || len(rs.Rules) == 0 {
		return false
	}
	for _, rule := range rs.Rules {
		if _, isExtend := rule.(*Extend); !isExtend {
			// Found a non-extend rule
			return false
		}
	}
	return true
}

// atRuleHasOnlySilentContent checks if an AtRule has only silent content (line comments).
// Such AtRules don't output any CSS and shouldn't add newlines.
func atRuleHasOnlySilentContent(a *AtRule) bool {
	if a == nil || a.Rules == nil {
		return false
	}
	// Check if it's a keyframes rule (which should always be output)
	if strings.Contains(a.Name, "keyframes") {
		return false
	}
	// Check if all rulesets in this AtRule contain only silent content
	for _, rule := range a.Rules {
		if ruleset, ok := rule.(*Ruleset); ok {
			if len(ruleset.Rules) > 0 {
				hasVisibleContent := false
				for _, r := range ruleset.Rules {
					if comment, isComment := r.(*Comment); isComment {
						if !comment.IsLineComment {
							hasVisibleContent = true
							break
						}
					} else {
						hasVisibleContent = true
						break
					}
				}
				if hasVisibleContent {
					return false
				}
			}
		} else {
			// Non-ruleset rule, might have visible content
			return false
		}
	}
	return true
}

// hasNoVisibleContent checks if a ruleset has no visible CSS content.
// This includes rulesets where all children have bubbled up to the parent,
// or rulesets that only contain variable declarations, comments, or other invisible content.
func hasNoVisibleContent(rs *Ruleset) bool {
	if rs == nil {
		return true
	}
	// If no rules at all, no visible content
	if len(rs.Rules) == 0 {
		return true
	}
	// Check if any rule produces visible output
	for _, rule := range rs.Rules {
		switch r := rule.(type) {
		case *Extend:
			// Extends don't produce output
			continue
		case *MixinDefinition:
			// Mixin definitions don't produce output
			continue
		case *Comment:
			// Silent comments (// style) don't produce output
			if r.IsLineComment {
				continue
			}
			// Block comments (/* */) do produce output
			return false
		case *Declaration:
			// Check if it's a variable declaration
			if r.GetVariable() {
				// Variable declaration - doesn't produce output
				continue
			}
			// Regular declaration produces output
			return false
		case *Ruleset:
			// Nested rulesets might have bubbled up, check if they still produce output here
			// If the nested ruleset has paths set, it will output via its own GenCSS
			// If it doesn't have visible content, it won't output
			if !hasNoVisibleContent(r) {
				return false
			}
		default:
			// Any other rule type might produce output
			return false
		}
	}
	return true
}

func (r *Ruleset) ResetCache() {
	r.variables = nil
	r.properties = nil
	// Use nil instead of make() - lookups is lazily initialized in Find() when needed
	// This avoids allocating a map that may never be used
	r.lookups = nil
}

func (r *Ruleset) Variables() map[string]any {
	if r.variables != nil {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[Variables] Returning cached %d variables\n", len(r.variables))
		}
		return r.variables
	}

	if r.Rules == nil {
		r.variables = make(map[string]any, 8)
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[Variables] Rules is nil, returning empty map\n")
		}
		return r.variables
	}

	// Use reduce-like pattern from JavaScript version
	r.variables = make(map[string]any, 8)
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[Variables] Building variables from %d rules\n", len(r.Rules))
	}
	for _, rule := range r.Rules {
		if decl, ok := rule.(*Declaration); ok && decl.variable {
			if name, ok := decl.name.(string); ok {
				r.variables[name] = decl
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[Variables] Found variable: %s\n", name)
				}
			}
		}
		// Handle import rules with variables like JavaScript version
		if ruleType, ok := rule.(interface{ GetType() string }); ok && ruleType.GetType() == "Import" {
			if importRule, ok := rule.(interface{
				GetRoot() interface{
					Variables() map[string]any
					Variable(string) any
				}
			}); ok {
				if root := importRule.GetRoot(); root != nil {
					vars := root.Variables()
					for name := range vars {
						if vars[name] != nil {
							r.variables[name] = root.Variable(name)
						}
					}
				}
			}
		}
	}

	return r.variables
}

func (r *Ruleset) Properties() map[string][]any {
	if r.properties != nil {
		return r.properties
	}

	if r.Rules == nil {
		r.properties = make(map[string][]any, 4)
		return r.properties
	}

	// Use reduce-like pattern from JavaScript version
	r.properties = make(map[string][]any, 4)
	for _, rule := range r.Rules {
		if decl, ok := rule.(*Declaration); ok && !decl.variable {
			var name string
			// Handle name like JavaScript: name.length === 1 && name[0] instanceof Keyword
			if nameSlice, ok := decl.name.([]any); ok {
				if len(nameSlice) == 1 {
					// Single element - extract value from it
					if kw, ok := nameSlice[0].(*Keyword); ok {
						name = kw.value
					} else {
						name = fmt.Sprintf("%v", nameSlice[0])
					}
				} else if len(nameSlice) > 1 {
					// Multiple elements - concatenate their values
					// This handles cases like compound property names
					parts := make([]string, 0, len(nameSlice))
					for _, elem := range nameSlice {
						if kw, ok := elem.(*Keyword); ok {
							parts = append(parts, kw.value)
						} else if str, ok := elem.(string); ok {
							parts = append(parts, str)
						} else if anon, ok := elem.(*Anonymous); ok {
							if val, ok := anon.Value.(string); ok {
								parts = append(parts, val)
							} else {
								parts = append(parts, fmt.Sprintf("%v", anon.Value))
							}
						} else {
							parts = append(parts, fmt.Sprintf("%v", elem))
						}
					}
					name = strings.Join(parts, "")
				}
			} else if nameStr, ok := decl.name.(string); ok {
				name = nameStr
			} else {
				name = fmt.Sprintf("%v", decl.name)
			}
			
			key := "$" + name
			// Properties don't overwrite as they can merge (from JavaScript comment)
			if _, exists := r.properties[key]; !exists {
				r.properties[key] = []any{decl}
			} else {
				r.properties[key] = append(r.properties[key], decl)
			}
		}
	}
	
	return r.properties
}

func (r *Ruleset) Variable(name string) map[string]any {
	// NOTE: We intentionally do NOT cache Variable() results because:
	// 1. The underlying variables map can change during evaluation
	// 2. Caching nil for missing variables breaks mixin lookups when variables are added later
	// 3. The Variables() map itself is already cached per ruleset

	vars := r.Variables()
	if decl, exists := vars[name]; exists {
		var result map[string]any
		if d, ok := decl.(*Declaration); ok {
			// Transform the declaration to parse Anonymous values into proper nodes
			transformed := r.transformDeclaration(d)

			// Return the expected format with value and important fields
			var value any
			if transformedDecl, ok := transformed.(*Declaration); ok {
				value = transformedDecl.Value
			} else {
				value = d.Value
			}

			result = map[string]any{
				"value": value,
			}

			// Check if the declaration has important flag
			if d.GetImportant() {
				// Store the actual important string value, not just a boolean
				result["important"] = d.important
			}
		} else {
			// Handle other types (like mock declarations in tests)
			result = map[string]any{
				"value": r.ParseValue(decl),
			}

			// Check if the declaration has important flag
			if declMap, ok := decl.(map[string]any); ok {
				if important, hasImportant := declMap["important"]; hasImportant {
					result["important"] = important
				}
			}
		}

		return result
	}

	return nil
}

func (r *Ruleset) Property(name string) []any {
	props := r.Properties()
	if decl, exists := props[name]; exists {
		// Transform declarations to parse Anonymous values into proper nodes (e.g., "10px" -> *Dimension)
		// This matches what Variable() does and is needed for namespace lookups in operations
		transformed := make([]any, len(decl))
		for i, d := range decl {
			transformed[i] = r.transformDeclaration(d)
		}
		return transformed
	}
	return nil
}

func (r *Ruleset) HasVariable(name string) bool {
	vars := r.Variables()
	_, exists := vars[name]
	return exists
}

// Matches JavaScript rules.variables
func (r *Ruleset) HasVariables() bool {
	return true // Rulesets always support variables
}

// Matches JavaScript rules.properties
func (r *Ruleset) HasProperties() bool {
	return true // Rulesets always support properties
}

func (r *Ruleset) LastDeclaration() any {
	if r.Rules == nil {
		return nil
	}
	
	// for (let i = this.rules.length; i > 0; i--) like JavaScript
	for i := len(r.Rules); i > 0; i-- {
		decl := r.Rules[i-1]
		if declaration, ok := decl.(*Declaration); ok {
			return r.ParseValue(declaration)
		}
	}
	return nil
}

func (r *Ruleset) ParseValue(toParse any) any {
	if toParse == nil {
		return nil
	}
	
	// Handle arrays
	if arr, ok := toParse.([]any); ok {
		nodes := make([]any, len(arr))
		for i, item := range arr {
			nodes[i] = r.transformDeclaration(item)
		}
		return nodes
	}
	
	return r.transformDeclaration(toParse)
}

func (r *Ruleset) transformDeclaration(decl any) any {
	if d, ok := decl.(*Declaration); ok {
		// Match JavaScript logic: if (decl.value instanceof Anonymous && !decl.parsed)
		if d.Value != nil && len(d.Value.Value) > 0 {
			// Check if not parsed (similar to JS !decl.parsed)
			parsed := false
			if d.Parsed != nil {
				if p, ok := d.Parsed.(bool); ok {
					parsed = p
				}
			}
			if anon, ok := d.Value.Value[0].(*Anonymous); ok && anon != nil && !parsed {
				// Check if needs parsing and ValueParseFunc is available
				if str, ok := anon.Value.(string); ok && str != "" && r.ValueParseFunc != nil {
					// Parse using ValueParseFunc (equivalent to JS parseNode call)
					result, err := r.ValueParseFunc(str, r.ParseContext, r.ParseImports, d.FileInfo(), anon.GetIndex())
					if err != nil {
						// If parsing fails, create a copy and mark as parsed to avoid infinite loops
						nodeCopy := NewNode()
						nodeCopy.CopyVisibilityInfo(d.Node.VisibilityInfo())
						nodeCopy.Parsed = true
						dCopy := &Declaration{
							Node:      nodeCopy,
							name:      d.name,
							Value:     d.Value,
							important: d.important,
							merge:     d.merge,
							inline:    d.inline,
							variable:  d.variable,
						}
						return dCopy
					} else if len(result) > 0 {
						
						// The parser returns Value>Expression>Dimension for "10px"
						// We should use the parsed Value directly, not wrap it again
						var valueCopy *Value
						if parsedValue, ok := result[0].(*Value); ok {
							valueCopy = parsedValue
						} else {
							// Fallback: wrap in Value if not already a Value
							valueCopy, _ = NewValue([]any{result[0]})
						}
						
						nodeCopy := NewNode()
						nodeCopy.CopyVisibilityInfo(d.Node.VisibilityInfo())
						nodeCopy.Parsed = true
						dCopy := &Declaration{
							Node:      nodeCopy,
							name:      d.name,
							Value:     valueCopy,
							important: d.important,
							merge:     d.merge,
							inline:    d.inline,
							variable:  d.variable,
						}
						if len(result) > 1 {
							// Handle important flag if present
							if important, ok := result[1].(string); ok {
								dCopy.important = important
							}
						}
						return dCopy
					}
				} else if str != "" {
					// Create a copy and mark as parsed even if no parser function available
					nodeCopy := NewNode()
					nodeCopy.CopyVisibilityInfo(d.Node.VisibilityInfo())
					nodeCopy.Parsed = true
					dCopy := &Declaration{
						Node:      nodeCopy,
						name:      d.name,
						Value:     d.Value,
						important: d.important,
						merge:     d.merge,
						inline:    d.inline,
						variable:  d.variable,
					}
					return dCopy
				}
			}
		}
	}
	return decl
}

func (r *Ruleset) Rulesets() []any {
	// NOTE: We intentionally do NOT cache Rulesets() results because Rules can be
	// modified during evaluation (e.g., mixin expansion, visitor transformations).
	// Caching would return stale data and break guard conditions and mixin lookups.
	if r.Rules == nil {
		return []any{}
	}

	// Pre-allocate with estimated capacity - typically a fraction of rules are rulesets
	filtered := make([]any, 0, len(r.Rules)/4+1)
	for _, rule := range r.Rules {
		if rs, ok := rule.(interface{ IsRuleset() bool }); ok && rs.IsRuleset() {
			filtered = append(filtered, rule)
		}
	}

	return filtered
}

func (r *Ruleset) PrependRule(rule any) {
	if r.Rules == nil {
		r.Rules = []any{rule}
	} else {
		newRules := make([]any, len(r.Rules)+1)
		newRules[0] = rule
		copy(newRules[1:], r.Rules)
		r.Rules = newRules
	}
	r.SetParent(rule, r.Node)
}

// Find finds rules matching a selector like JavaScript version
func (r *Ruleset) Find(selector any, self any, filter func(any) bool) []any {
	if self == nil {
		self = r
	}

	var key string
	if sel, ok := selector.(interface{ ToCSS() string }); ok {
		key = sel.ToCSS()
	} else if sel, ok := selector.(interface{ ToCSS(any) string }); ok {
		key = sel.ToCSS(nil)
	} else {
		key = fmt.Sprintf("%v", selector)
	}

	if cached, exists := r.lookups[key]; exists {
		return cached
	}

	// Pre-allocate rules with small initial capacity to avoid reallocation
	// Most Find calls return 0-2 results
	rules := make([]any, 0, 4)
	var match int
	var foundMixins []any

	// this.rulesets().forEach(function (rule) { ... }) pattern
	rulesets := r.Rulesets()

	for _, rule := range rulesets {
		if rule == self {
			continue
		}
		
		// Handle both *Ruleset and *MixinDefinition (which embeds *Ruleset)
		var rulesetSelectors []any
		switch r := rule.(type) {
		case *Ruleset:
			rulesetSelectors = r.Selectors
		case *MixinDefinition:
			rulesetSelectors = r.Selectors
		default:
			// Check if it has a Selectors field via interface
			if rs, ok := rule.(interface{ GetSelectors() []any }); ok {
				rulesetSelectors = rs.GetSelectors()
			}
		}
		
		if rulesetSelectors != nil {
			for j := 0; j < len(rulesetSelectors); j++ {
				if sel, ok := selector.(*Selector); ok {
					if ruleSelector, ok := rulesetSelectors[j].(*Selector); ok {
						match = sel.Match(ruleSelector)
						if match > 0 {
							if len(sel.Elements) > match {
								if filter == nil || filter(rule) {
									// Create new selector with remaining elements like JavaScript
									remainingElements := make([]*Element, len(sel.Elements)-match)
									copy(remainingElements, sel.Elements[match:])
									newSelector, err := NewSelector(remainingElements, nil, nil, sel.GetIndex(), sel.FileInfo(), nil)
									if err == nil {
										// Use Find method on the rule (works for both Ruleset and MixinDefinition)
										if finder, ok := rule.(interface{ Find(any, any, func(any) bool) []any }); ok {
											foundMixins = finder.Find(newSelector, self, filter)
										}
										for i := 0; i < len(foundMixins); i++ {
											if mixin, ok := foundMixins[i].(map[string]any); ok {
												if path, ok := mixin["path"].([]any); ok {
													// foundMixins[i].path.push(rule);
													newPath := make([]any, len(path)+1)
													copy(newPath, path)
													newPath[len(path)] = rule
													mixin["path"] = newPath
												}
											}
										}
										// Array.prototype.push.apply(rules, foundMixins);
										rules = append(rules, foundMixins...)
									}
								}
							} else {
								rules = append(rules, map[string]any{
									"rule": rule,
									"path": []any{},
								})
							}
							break
						}
					}
				} else {
					// Handle any selector with Match method using interface
					if selectorWithMatch, ok := selector.(interface{ Match(any) int }); ok {
						match = selectorWithMatch.Match(rulesetSelectors[j])
						if match > 0 {
							rules = append(rules, map[string]any{
								"rule": rule,
								"path": []any{},
							})
							break
						}
					}
				}
			}
		}
	}

	// Lazily initialize lookups map
	if r.lookups == nil {
		r.lookups = make(map[string][]any, 4)
	}
	r.lookups[key] = rules
	return rules
}

func (r *Ruleset) GenCSS(context any, output *CSSOutput) {
	// Debug: trace all div ruleset entries to GenCSS
	if os.Getenv("LESS_GO_DEBUG_VIS") == "1" && len(r.Selectors) > 0 {
		if sel, ok := r.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
			elemVal := sel.Elements[0].Value
			if str, ok := elemVal.(string); ok && str == "div" {
				fmt.Fprintf(os.Stderr, "[GenCSS ENTRY] div ruleset=%p, Node=%p, Paths=%v, Root=%v\n",
					r, r.Node, r.Paths != nil && len(r.Paths) > 0, r.Root)
				if r.Node != nil {
					fmt.Fprintf(os.Stderr, "[GenCSS ENTRY]   BlocksVisibility=%v\n", r.Node.BlocksVisibility())
				}
			}
		}
	}

	// Skip rulesets that are inside mixin definitions - they should only be output when the mixin is called
	if r.InsideMixinDefinition {
		return
	}

	// Debug output for MultiMedia rulesets
	if os.Getenv("LESS_GO_DEBUG") == "1" && r.MultiMedia {
		fmt.Fprintf(os.Stderr, "[RULESET.GenCSS] MultiMedia ruleset with %d rules\n", len(r.Rules))
		for i, rule := range r.Rules {
			fmt.Fprintf(os.Stderr, "[RULESET.GenCSS]   rule[%d]: type=%T\n", i, rule)
		}
	}

	// Don't skip rulesets with visibility blocks here - they may contain visible paths
	// from extends. The path filtering logic below will filter out invisible paths.
	// This matches JavaScript behavior which doesn't have an early return for visibility.

	ctx, ok := context.(map[string]any)
	if !ok {
		ctx = make(map[string]any)
	}

	// Check if this ruleset should be treated as top-level
	// When rulesets are extracted from parent rulesets and output separately,
	// they should be formatted as top-level rulesets even though Root=false
	isTopLevel := false
	if tl, ok := ctx["topLevel"].(bool); ok && tl {
		isTopLevel = true
	}

	// Set tab level
	tabLevel := 0
	if tl, ok := ctx["tabLevel"].(int); ok {
		tabLevel = tl
	}

	// Check if this is a media-empty ruleset (used by Media queries)
	// Media-empty rulesets should not increment tabLevel since they don't output braces
	// Note: After JoinSelectorVisitor runs, r.Paths might be an empty array instead of nil,
	// so we also check len(r.Paths) == 0
	isMediaEmpty := false
	if !r.Root && (r.Paths == nil || len(r.Paths) == 0) && len(r.Selectors) == 1 {
		if sel, ok := r.Selectors[0].(*Selector); ok && sel.MediaEmpty {
			isMediaEmpty = true
		}
	}


	// Check if this is a container ruleset (no selectors/paths)
	// Container rulesets inside at-rules (@supports, @document) don't output their own braces,
	// so they shouldn't increment tabLevel
	isContainer := false
	if !r.Root && (r.Paths == nil || len(r.Paths) == 0) && (r.Selectors == nil || len(r.Selectors) == 0) {
		isContainer = true
	}

	// For non-root, non-media-empty, non-container rulesets: increment tabLevel
	// But skip this for top-level rulesets (extracted rulesets that should be formatted at root level)
	if !r.Root && !isMediaEmpty && !isTopLevel && !isContainer {
		tabLevel++
		ctx["tabLevel"] = tabLevel
	}
	
	compress := false
	if c, ok := ctx["compress"].(bool); ok {
		compress = c
	}
	
	var tabRuleStr, tabSetStr string
	if compress {
		tabRuleStr = ""
		tabSetStr = ""
	} else {
		// For top-level rulesets, calculate tabRuleStr as if we had incremented tabLevel
		// This ensures declarations inside have correct indentation (2 spaces)
		effectiveTabLevel := tabLevel
		if !r.Root && isTopLevel {
			effectiveTabLevel = tabLevel + 1
		}

		// JavaScript: Array(tabLevel + 1).join('  ') produces (tabLevel) * 2 spaces
		// JavaScript: Array(tabLevel).join('  ') produces (tabLevel - 1) * 2 spaces (minimum 0)
		tabRuleStr = strings.Repeat("  ", effectiveTabLevel)
		if tabLevel > 0 {
			tabSetStr = strings.Repeat("  ", tabLevel-1)
		} else {
			tabSetStr = ""
		}
	}
	
	// Organize rules by type like JavaScript version
	var charsetRuleNodes []any
	// Pre-allocate ruleNodes with capacity of Rules length to avoid reallocation
	ruleNodes := make([]any, 0, len(r.Rules))

	var charsetNodeIndex int = 0
	var importNodeIndex int = 0

	if r.Rules != nil {
		for i, rule := range r.Rules {
			_ = i // suppress unused variable warning when debug is disabled
			// Skip silent comments entirely - they don't generate output
			// This prevents extra blank lines from being added after the last visible rule
			if comment, ok := rule.(*Comment); ok {
				isSilent := comment.IsSilent(ctx)
				if os.Getenv("LESS_GO_DEBUG_COMMENT") == "1" {
					fmt.Fprintf(os.Stderr, "[GenCSS] Comment: IsLineComment=%v, Value=%q, IsSilent=%v\n",
						comment.IsLineComment, comment.Value, isSilent)
				}
				if isSilent {
					continue // Skip silent comments
				}
				// Also skip comments that block visibility and are not explicitly visible
				// This handles comments from reference imports
				if comment.Node != nil && comment.Node.BlocksVisibility() {
					nodeVisible := comment.Node.IsVisible()
					if nodeVisible == nil || !*nodeVisible {
						continue // Skip comments from reference imports
					}
				}
				// Non-silent, visible comments are included
				if importNodeIndex == i {
					importNodeIndex++
				}
				ruleNodes = append(ruleNodes, rule)
			} else if _, ok := rule.(*Extend); ok {
				// Skip Extend rules entirely - they don't generate CSS output
				// Extend rules are processed during the extend visitor phase and should not appear in CSS
				// This prevents extra blank lines from being added
				continue
			} else if ruleBlocksVisibility(rule) {
				// Debug: log when we skip a div ruleset
				if os.Getenv("LESS_GO_DEBUG_VIS") == "1" {
					if rs, ok := rule.(*Ruleset); ok && len(rs.Selectors) > 0 {
						if sel, ok := rs.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
							elemVal := sel.Elements[0].Value
							if str, ok := elemVal.(string); ok && str == "div" {
								fmt.Fprintf(os.Stderr, "[GenCSS] SKIPPING div ruleset=%p, Node=%p\n", rs, rs.Node)
								if rs.Node != nil {
									fmt.Fprintf(os.Stderr, "[GenCSS] BlocksVisibility=%v, VisibilityBlocks=%v, IsVisible=%v\n",
										rs.Node.BlocksVisibility(), rs.Node.VisibilityBlocks, rs.Node.IsVisible())
								}
							}
						}
					}
				}
				// Skip rules that block visibility and are not explicitly visible
				// This handles rulesets from reference imports
				continue
			} else if charset, ok := rule.(interface{ IsCharset() bool }); ok && charset.IsCharset() {
				// Insert at charsetNodeIndex position
				// Use slices.Insert if available, or manual splice
				ruleNodes = append(ruleNodes, nil)
				copy(ruleNodes[charsetNodeIndex+1:], ruleNodes[charsetNodeIndex:])
				ruleNodes[charsetNodeIndex] = rule
				charsetNodeIndex++
				importNodeIndex++
			} else if ruleType, ok := rule.(interface{ GetType() string }); ok && ruleType.GetType() == "Import" {
				// Insert at importNodeIndex position
				ruleNodes = append(ruleNodes, nil)
				copy(ruleNodes[importNodeIndex+1:], ruleNodes[importNodeIndex:])
				ruleNodes[importNodeIndex] = rule
				importNodeIndex++
			} else {
				ruleNodes = append(ruleNodes, rule)
			}
		}
	}

	// ruleNodes = charsetRuleNodes.concat(ruleNodes);
	ruleNodes = append(charsetRuleNodes, ruleNodes...)

	if os.Getenv("LESS_GO_DEBUG") == "1" && r.MultiMedia {
		fmt.Fprintf(os.Stderr, "[RULESET.GenCSS] MultiMedia has %d ruleNodes after organizing\n", len(ruleNodes))
	}

	// Check if this ruleset contains only extends (no actual CSS output)
	// If so, we'll skip generating selectors/braces but still complete normally for proper spacing
	hasOnlyExtends := !r.Root && len(r.Rules) > 0 && len(ruleNodes) == 0


	// Track how many paths were actually output (for visibility filtering)
	outputCount := 0

	if !r.Root && !isMediaEmpty && !hasOnlyExtends {
		// Generate debug info
		if debugInfo := GetDebugInfo(ctx, r, tabSetStr); debugInfo != "" {
			output.Add(debugInfo, nil, nil)
			output.Add(tabSetStr, nil, nil)
		}

		// Generate selectors - prefer Paths if available, otherwise fall back to Selectors
		if r.Paths != nil && len(r.Paths) > 0 {
			// Use Paths (set by JoinSelectorVisitor)
			sep := ","
			if !compress {
				sep = ",\n" + tabSetStr
			}

			// Filter paths to only include visible selectors
			// This implements reference import functionality where selectors from reference imports
			// are hidden unless explicitly made visible (via extend or mixin call)
			for _, path := range r.Paths {
				if len(path) == 0 {
					continue
				}

				// Visibility filtering for reference imports:
				// Check if THIS RULESET blocks visibility (from reference import)
				// If it does, paths need at least one explicitly visible selector (from extend)
				// If it doesn't, all paths pass through
				rulesetBlocksVisibility := r.Node != nil && r.Node.BlocksVisibility()

				// Debug: log path visibility for div rulesets
				if os.Getenv("LESS_GO_DEBUG_VIS") == "1" && len(r.Selectors) > 0 {
					if sel, ok := r.Selectors[0].(*Selector); ok && len(sel.Elements) > 0 {
						elemVal := sel.Elements[0].Value
						if str, ok := elemVal.(string); ok && str == "div" {
							fmt.Fprintf(os.Stderr, "[GenCSS path check] div ruleset=%p, Node=%p, BlocksVisibility=%v\n",
								r, r.Node, rulesetBlocksVisibility)
						}
					}
				}

				pathIsVisible := true
				if rulesetBlocksVisibility {
					// Ruleset blocks visibility - need at least one explicitly visible selector
					pathHasVisibleSelector := false
					for _, pathElem := range path {
						if sel, ok := pathElem.(*Selector); ok {
							// Check if selector is explicitly visible (set by extend processing)
							if sel.Node != nil {
								nodeVis := sel.Node.IsVisible()
								if nodeVis != nil && *nodeVis && sel.EvaldCondition {
									pathHasVisibleSelector = true
									break
								}
							}
						}
					}
					pathIsVisible = pathHasVisibleSelector
				}

				if !pathIsVisible {
					continue
				}

				if outputCount > 0 {
					output.Add(sep, nil, nil)
				}
				outputCount++

				// Always set firstSelector to true for the first selector in a path
				// This ensures no extra space is added at the beginning of selectors
				ctx["firstSelector"] = true
				if gen, ok := path[0].(interface{ GenCSS(any, *CSSOutput) }); ok {
					gen.GenCSS(ctx, output)
				}

				ctx["firstSelector"] = false
				for j := 1; j < len(path); j++ {
					if gen, ok := path[j].(interface{ GenCSS(any, *CSSOutput) }); ok {
						gen.GenCSS(ctx, output)
					}
				}
			}
		} else if r.Selectors != nil && len(r.Selectors) > 0 {
			// Fallback: if Paths is nil, use Selectors directly
			// This handles cases where JoinSelectorVisitor hasn't run yet (e.g., mixin expansions)
			// Note: InsideMixinDefinition flag check at line 1360 prevents nested rulesets
			// inside mixin definitions from reaching this point

			sep := ","
			if !compress {
				sep = ",\n" + tabSetStr
			}

			for i, selector := range r.Selectors {
				if i > 0 {
					output.Add(sep, nil, nil)
				}

				ctx["firstSelector"] = true
				if gen, ok := selector.(interface{ GenCSS(any, *CSSOutput) }); ok {
					gen.GenCSS(ctx, output)
				}
				ctx["firstSelector"] = false
			}
		} else {
			// No paths and no selectors - skip selector output
			// This can happen for rulesets that only contain declarations
		}

		// Add opening brace (unless we skipped selector output or all paths were filtered)
		// For reference imports: if all paths were filtered out due to visibility, skip the entire ruleset
		if outputCount > 0 || (r.Selectors != nil && len(r.Selectors) > 0 && r.Paths == nil) {
			if compress {
				output.Add("{", nil, nil)
			} else {
				output.Add(" {\n", nil, nil)
			}
			output.Add(tabRuleStr, nil, nil)
		} else if r.Paths != nil && len(r.Paths) > 0 && outputCount == 0 {
			// All paths were filtered out - skip the entire ruleset
			return
		}
	}

	// Generate CSS for rules (skip if this ruleset contains only extends)
	// Note: Silent comments have been filtered out above, so we only process rules that generate output
	if !hasOnlyExtends {
	for i, rule := range ruleNodes {
		if i+1 == len(ruleNodes) {
			ctx["lastRule"] = true
		} else {
			// Explicitly set to false for non-last rules to override any parent context setting
			ctx["lastRule"] = false
		}

		currentLastRule := false
		if lr, ok := ctx["lastRule"].(bool); ok {
			currentLastRule = lr
		}

		// Check IsRulesetLike - handle both bool return and any return (AtRule returns any)
		isRulesetLikeForCtx := false
		if rl, ok := rule.(interface{ IsRulesetLike() any }); ok {
			result := rl.IsRulesetLike()
			if result != nil {
				if b, ok := result.(bool); ok {
					isRulesetLikeForCtx = b
				} else {
					// Non-nil, non-bool (like rules slice) means ruleset-like
					isRulesetLikeForCtx = true
				}
			}
		} else if rl, ok := rule.(interface{ IsRulesetLike() bool }); ok {
			isRulesetLikeForCtx = rl.IsRulesetLike()
		}
		if isRulesetLikeForCtx {
			ctx["lastRule"] = false
		}

		// For file-level root rulesets: mark child rulesets as top-level so they format correctly
		// This ensures extracted rulesets (that were moved to root's rules) are treated as top-level
		// Only do this for the actual file root (tabLevel == 0), not for container rulesets like @keyframes
		childContext := ctx
		if r.Root && tabLevel == 0 {
			if childRuleset, ok := rule.(*Ruleset); ok && !childRuleset.Root {
				// Create a new context for the child with topLevel flag
				// Don't modify tabLevel - let it stay as-is so declarations inside have correct indentation
				// Pre-allocate with capacity for all keys + 1 for topLevel
				childContext = make(map[string]any, len(ctx)+1)
				for k, v := range ctx {
					childContext[k] = v
				}
				childContext["topLevel"] = true
				// Note: We don't set tabLevel=0 because that would affect indentation of declarations
				// The topLevel flag will prevent incrementing tabLevel, which is what we want
			}
		}

		// Generate CSS for the rule
		if gen, ok := rule.(interface{ GenCSS(any, *CSSOutput) }); ok {
			gen.GenCSS(childContext, output)
		} else if val, ok := rule.(interface{ GetValue() any }); ok {
			output.Add(fmt.Sprintf("%v", val.GetValue()), nil, nil)
		}

		ctx["lastRule"] = currentLastRule

		// Add newline after rule if it's not the last rule
		if !currentLastRule {
			shouldAddNewline := false

			// Check if rule is visible (for declarations, etc.)
			// Also check if rule blocks visibility but is not explicitly visible (from reference imports)
			// IMPORTANT: Check for Node visibility (BlocksVisibility/IsVisible *bool) FIRST,
			// before checking simple IsVisible() bool, because Comment has both methods
			// and we want to respect the Node's visibility from reference imports
			if visNode, ok := rule.(interface{ BlocksVisibility() bool; IsVisible() *bool }); ok {
				blocksVis := visNode.BlocksVisibility()
				if blocksVis {
					// Node blocks visibility (from reference import) - only visible if explicitly marked
					nodeVisible := visNode.IsVisible()
					if nodeVisible != nil && *nodeVisible {
						shouldAddNewline = true
					}
					// else: invisible, don't add newline
				} else {
					// Node doesn't block visibility - check simple IsVisible
					if vis2, ok2 := rule.(interface{ IsVisible() bool }); ok2 && vis2.IsVisible() {
						shouldAddNewline = true
					}
				}
			} else if vis, ok := rule.(interface{ IsVisible() bool }); ok && vis.IsVisible() {
				// Node doesn't have BlocksVisibility/IsVisible *bool, and IsVisible() is true
				shouldAddNewline = true
			}

			// Special case: Add newlines for child rulesets inside container rulesets
			// (like @keyframes which creates a container ruleset with no selectors)
			// Check IsRulesetLike - handle both bool return and any return (AtRule returns any)
			isRulesetLike := false
			if rl, ok := rule.(interface{ IsRulesetLike() any }); ok {
				result := rl.IsRulesetLike()
				if result != nil {
					if b, ok := result.(bool); ok {
						isRulesetLike = b
					} else {
						// Non-nil, non-bool (like rules slice) means ruleset-like
						isRulesetLike = true
					}
				}
			} else if rl, ok := rule.(interface{ IsRulesetLike() bool }); ok {
				isRulesetLike = rl.IsRulesetLike()
			}
			if isRulesetLike {
				// Skip adding newlines for MixinDefinitions - they don't output anything
				if _, isMixinDef := rule.(*MixinDefinition); isMixinDef {
					// Don't add newline after mixin definitions
				} else if rs, isRuleset := rule.(*Ruleset); isRuleset && isExtendOnlyRuleset(rs) {
					// Don't add newline after extend-only rulesets - they don't output anything
				} else if rs, isRuleset := rule.(*Ruleset); isRuleset && hasNoVisibleContent(rs) {
					// Don't add newline after rulesets with no visible content (e.g., parent containers after children bubbled)
				} else if at, isAtRule := rule.(*AtRule); isAtRule && atRuleHasOnlySilentContent(at) {
					// Don't add newline after AtRules with only silent content (line comments)
				} else {
					// Only add newline if parent ruleset has no selectors (is a container)
					// and we're not at the file-level root (tabLevel > 0)
					isParentContainer := (r.Paths == nil || len(r.Paths) == 0) && (r.Selectors == nil || len(r.Selectors) == 0)
					if isParentContainer && tabLevel > 0 {
						shouldAddNewline = true
					}
					// Also add newline for top-level rulesets (children of root)
					if r.Root && tabLevel == 0 {
						shouldAddNewline = true
					}
				}
			}


			// Special case: Import nodes should always have a newline after them
			if ruleType, ok := rule.(interface{ GetType() string }); ok && ruleType.GetType() == "Import" {
				shouldAddNewline = true
			}

			// Special case: AtRules without rules (like @charset, @namespace) at root level need newlines
			if at, isAtRule := rule.(*AtRule); isAtRule && r.Root && tabLevel == 0 {
				// AtRules without rules produce output and need newlines
				if at.Rules == nil || len(at.Rules) == 0 {
					shouldAddNewline = true
				}
			}

			if shouldAddNewline && !compress {
				output.Add("\n"+tabRuleStr, nil, nil)
			}
		} else {
			ctx["lastRule"] = false
		}
	}
	}

	// Decrement tab level FIRST for correct newline logic
	// Do this for all non-root rulesets (except top-level, media-empty, and container), even if we skip output (for extend-only rulesets)
	// Skip for top-level because we didn't increment it
	// Skip for media-empty rulesets since they didn't increment tabLevel
	// Skip for container rulesets since they didn't increment tabLevel
	if !r.Root && !isTopLevel && !isMediaEmpty && !isContainer {
		tabLevel--
		ctx["tabLevel"] = tabLevel
	}

	// Add closing brace (skip if this ruleset contains only extends, all paths were filtered, or is a container)
	// Container rulesets don't output their own braces - they're transparent wrappers
	if !r.Root && !isMediaEmpty && !hasOnlyExtends && !isContainer && (outputCount > 0 || r.Paths == nil) {
		if compress {
			output.Add("}", nil, nil)
		} else {
			output.Add("\n"+tabSetStr+"}", nil, nil)
		}
	}
	
	// Add final newline for first root
	if !output.IsEmpty() && !compress && r.FirstRoot {
		output.Add("\n", nil, nil)
	}
}

func (r *Ruleset) JoinSelectors(paths *[][]any, context [][]any, selectors []any) {
	for _, selector := range selectors {
		r.JoinSelector(paths, context, selector)
	}
}

// JoinSelector joins a single selector with the current context
// This is a complex method that implements the JavaScript selector joining logic
func (r *Ruleset) JoinSelector(paths *[][]any, context [][]any, selector any) {
	sel, ok := selector.(*Selector)
	if !ok {
		return
	}

	// Debug: log selector elements
	if os.Getenv("LESS_GO_DEBUG_JOIN") == "1" {
		fmt.Fprintf(os.Stderr, "[JoinSelector] selector elements (%d): ", len(sel.Elements))
		for i, el := range sel.Elements {
			fmt.Fprintf(os.Stderr, "[%d]=%q ", i, el.Value)
		}
		fmt.Fprintf(os.Stderr, ", context len=%d, MediaEmpty=%v\n", len(context), sel.MediaEmpty)
	}

	// createSelector helper function
	createSelector := func(containedElement any, originalElement *Element) (*Selector, error) {
		element := NewElement(nil, containedElement, originalElement.IsVariable, originalElement.GetIndex(), originalElement.FileInfo(), nil)
		return NewSelector([]*Element{element}, nil, nil, 0, make(map[string]any), nil)
	}

	// addReplacementIntoPath helper function
	addReplacementIntoPath := func(beginningPath []any, addPath []any, replacedElement *Element, originalSelector *Selector) []any {
		var newSelectorPath []any
		var newJoinedSelector *Selector

		// Construct the joined selector - if & is the first thing this will be empty,
		// if not newJoinedSelector will be the last set of elements in the selector
		if len(beginningPath) > 0 {
			newSelectorPath = make([]any, len(beginningPath))
			copy(newSelectorPath, beginningPath)
			if lastSel, ok := newSelectorPath[len(newSelectorPath)-1].(*Selector); ok {
				newSelectorPath = newSelectorPath[:len(newSelectorPath)-1]
				// Create a copy of lastSel.Elements to avoid modifying the original
				lastSelElements := make([]*Element, len(lastSel.Elements))
				copy(lastSelElements, lastSel.Elements)
				newJoinedSelector, _ = originalSelector.CreateDerived(lastSelElements, nil, nil)
			}
		} else {
			newJoinedSelector, _ = originalSelector.CreateDerived([]*Element{}, nil, nil)
		}

		if len(addPath) > 0 {
			// /deep/ is a CSS4 selector - (removed, so should deprecate)
			// that is valid without anything in front of it
			// so if the & does not have a combinator that is "" or " " then
			// and there is a combinator on the parent, then grab that.
			combinator := replacedElement.Combinator

			if firstPathSel, ok := addPath[0].(*Selector); ok && len(firstPathSel.Elements) > 0 {
				parentEl := firstPathSel.Elements[0]
				if combinator.EmptyOrWhitespace && !parentEl.Combinator.EmptyOrWhitespace {
					combinator = parentEl.Combinator
				}
				// Join the elements so far with the first part of the parent
				// Debug: Print what we're doing
				if os.Getenv("LESS_GO_DEBUG_SELECTOR") == "1" {
					fmt.Fprintf(os.Stderr, "DEBUG addReplacementIntoPath: newJoinedSelector before=%v, parentEl=%v\n",
						elementSliceToString(newJoinedSelector.Elements), elementToString(parentEl))
				}
				newJoinedSelector.Elements = append(newJoinedSelector.Elements, NewElement(
					combinator,
					parentEl.Value,
					replacedElement.IsVariable,
					replacedElement.GetIndex(),
					replacedElement.FileInfo(),
					nil,
				))
				newJoinedSelector.Elements = append(newJoinedSelector.Elements, firstPathSel.Elements[1:]...)
				if os.Getenv("LESS_GO_DEBUG_SELECTOR") == "1" {
					fmt.Fprintf(os.Stderr, "DEBUG addReplacementIntoPath: newJoinedSelector after=%v\n",
						elementSliceToString(newJoinedSelector.Elements))
				}
			}
		}

		// Now add the joined selector - but only if it is not empty
		if len(newJoinedSelector.Elements) != 0 {
			newSelectorPath = append(newSelectorPath, newJoinedSelector)
		}

		// Put together the parent selectors after the join (e.g. the rest of the parent)
		if len(addPath) > 1 {
			restOfPath := addPath[1:]
			for _, pathItem := range restOfPath {
				if pathSel, ok := pathItem.(*Selector); ok {
					newDerived, _ := pathSel.CreateDerived(pathSel.Elements, []any{}, nil)
					newSelectorPath = append(newSelectorPath, newDerived)
				} else {
					newSelectorPath = append(newSelectorPath, pathItem)
				}
			}
		}
		return newSelectorPath
	}

	// addAllReplacementsIntoPath helper function  
	addAllReplacementsIntoPath := func(beginningPaths [][]any, addPaths []any, replacedElement *Element, originalSelector *Selector, result *[][]any) {
		for j := 0; j < len(beginningPaths); j++ {
			newSelectorPath := addReplacementIntoPath(beginningPaths[j], addPaths, replacedElement, originalSelector)
			*result = append(*result, newSelectorPath)
		}
	}

	// mergeElementsOnToSelectors helper function
	mergeElementsOnToSelectors := func(elements []*Element, selectors *[][]any) {
		if len(elements) == 0 {
			return
		}
		if len(*selectors) == 0 {
			newSel, _ := NewSelector(elements, nil, nil, 0, make(map[string]any), nil)
			*selectors = append(*selectors, []any{newSel})
			return
		}

		for idx, sel := range *selectors {
			// If the previous thing in sel is a parent this needs to join on to it
			if len(sel) > 0 {
				if lastSel, ok := sel[len(sel)-1].(*Selector); ok {
					newElements := append(lastSel.Elements, elements...)
					newDerived, _ := lastSel.CreateDerived(newElements, nil, nil)
					(*selectors)[idx][len(sel)-1] = newDerived
				}
			} else {
				newSel, _ := NewSelector(elements, nil, nil, 0, make(map[string]any), nil)
				(*selectors)[idx] = append((*selectors)[idx], newSel)
			}
		}
	}

	// Helper function to find nested selector
	findNestedSelector := func(element *Element) *Selector {
		if paren, ok := element.Value.(*Paren); ok {
			if nestedSel, ok := paren.Value.(*Selector); ok {
				return nestedSel
			}
		}
		return nil
	}

	// replaceParentSelector helper function
	var replaceParentSelector func(*[][]any, [][]any, *Selector) bool
	replaceParentSelector = func(paths *[][]any, context [][]any, inSelector *Selector) bool {
		hadParentSelector := false
		currentElements := []*Element{}
		newSelectors := [][]any{{}}

		for _, el := range inSelector.Elements {
			// Debug: log element value and type
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG replaceParentSelector] Element: Value=%#v, Type=%T, Combinator=%v\n",
					el.Value, el.Value, el.Combinator)
			}

			// Non parent reference elements just get added
			if valueStr, ok := el.Value.(string); !ok || valueStr != "&" {
				nestedSelector := findNestedSelector(el)
				if nestedSelector != nil {
					// Merge the current list of non parent selector elements
					// on to the current list of selectors to add
					mergeElementsOnToSelectors(currentElements, &newSelectors)

					nestedPaths := [][]any{}
					replacedNewSelectors := [][]any{}
					replaced := replaceParentSelector(&nestedPaths, context, nestedSelector)
					hadParentSelector = hadParentSelector || replaced

					// The nestedPaths array should have only one member - replaceParentSelector does not multiply selectors
					for k := 0; k < len(nestedPaths); k++ {
						// Reconstruct the selector from the path (which has & expanded)
						// nestedPaths[k] is a path containing selector(s) with expanded &
						var reconstructedSelector *Selector
						if len(nestedPaths[k]) == 1 {
							// Simple case: path contains one selector
							if sel, ok := nestedPaths[k][0].(*Selector); ok {
								reconstructedSelector = sel
							}
						} else if len(nestedPaths[k]) > 1 {
							// Complex case: path contains multiple selectors that need to be joined
							// Merge all elements from all selectors in the path
							var allElements []*Element
							for _, pathSel := range nestedPaths[k] {
								if sel, ok := pathSel.(*Selector); ok {
									allElements = append(allElements, sel.Elements...)
								}
							}
							if len(allElements) > 0 {
								reconstructedSelector, _ = NewSelector(allElements, nil, nil, 0, make(map[string]any), nil)
							}
						}

						if reconstructedSelector != nil {
							// Create a Paren containing the reconstructed selector with expanded &
							paren := NewParen(reconstructedSelector)
							// Create a new element with this Paren as the value
							newElement := NewElement(
								el.Combinator,
								paren,
								el.IsVariable,
								el.GetIndex(),
								el.FileInfo(),
								nil,
							)
							// Create a selector from this element
							if replacementSelector, err := createSelector(newElement, el); err == nil {
								addAllReplacementsIntoPath(newSelectors, []any{replacementSelector}, el, inSelector, &replacedNewSelectors)
							}
						}
					}
					newSelectors = replacedNewSelectors
					currentElements = []*Element{}
				} else {
					currentElements = append(currentElements, el)
				}
			} else {
				hadParentSelector = true
				// The new list of selectors to add
				selectorsMultiplied := [][]any{}

				// Merge the current list of non parent selector elements
				// on to the current list of selectors to add
				mergeElementsOnToSelectors(currentElements, &newSelectors)

				// Loop through our current selectors
				for j := 0; j < len(newSelectors); j++ {
					sel := newSelectors[j]
					// If we don't have any parent paths, the & might be in a mixin so that it can be used
					// whether there are parents or not
					if len(context) == 0 {
						// The combinator used on el should now be applied to the next element instead so that
						// it is not lost
						if len(sel) > 0 {
							if firstSel, ok := sel[0].(*Selector); ok {
								newElement := NewElement(el.Combinator, "", el.IsVariable, el.GetIndex(), el.FileInfo(), nil)
								firstSel.Elements = append(firstSel.Elements, newElement)
							}
						}
						selectorsMultiplied = append(selectorsMultiplied, sel)
					} else {
						// And the parent selectors
						for k := 0; k < len(context); k++ {
							// We need to put the current selectors
							// then join the last selector's elements on to the parents selectors
							newSelectorPath := addReplacementIntoPath(sel, context[k], el, inSelector)
							// Add that to our new set of selectors
							selectorsMultiplied = append(selectorsMultiplied, newSelectorPath)
						}
					}
				}

				// Our new selectors has been multiplied, so reset the state
				newSelectors = selectorsMultiplied
				currentElements = []*Element{}
			}
		}

		// If we have any elements left over (e.g. .a& .b == .b)
		// add them on to all the current selectors
		mergeElementsOnToSelectors(currentElements, &newSelectors)

		for idx := 0; idx < len(newSelectors); idx++ {
			length := len(newSelectors[idx])
			if length > 0 {
				*paths = append(*paths, newSelectors[idx])
				if lastSelector, ok := newSelectors[idx][length-1].(*Selector); ok {
					newDerived, _ := lastSelector.CreateDerived(lastSelector.Elements, inSelector.ExtendList, nil)
					newSelectors[idx][length-1] = newDerived
				}
			}
		}

		return hadParentSelector
	}

	// deriveSelector helper function
	deriveSelector := func(visibilityInfo map[string]any, deriveFrom *Selector) *Selector {
		newSelector, _ := deriveFrom.CreateDerived(deriveFrom.Elements, deriveFrom.ExtendList, &deriveFrom.EvaldCondition)
		newSelector.CopyVisibilityInfo(visibilityInfo)
		return newSelector
	}

	// Main joinSelector logic
	newPaths := [][]any{}
	hadParentSelector := replaceParentSelector(&newPaths, context, sel)

	if !hadParentSelector {
		if len(context) > 0 {
			newPaths = [][]any{}
			for idx := 0; idx < len(context); idx++ {
				concatenated := make([]any, len(context[idx])+1)
				
				// Map over context[idx] using deriveSelector  
				for j, ctxSel := range context[idx] {
					if ctxSelector, ok := ctxSel.(*Selector); ok {
						concatenated[j] = deriveSelector(sel.VisibilityInfo(), ctxSelector)
					} else {
						concatenated[j] = ctxSel
					}
				}
				concatenated[len(context[idx])] = sel
				newPaths = append(newPaths, concatenated)
			}
		} else {
			newPaths = [][]any{{sel}}
		}
	}

	for idx := 0; idx < len(newPaths); idx++ {
		*paths = append(*paths, newPaths[idx])
	}
}

func GetDebugInfo(context map[string]any, ruleset *Ruleset, separator string) string {
	if context == nil || ruleset == nil {
		return ""
	}
	
	// Check if dumpLineNumbers is enabled and not compressing
	dumpLineNumbers, hasDump := context["dumpLineNumbers"]
	compress, hasCompress := context["compress"]
	
	if !hasDump || (hasCompress && compress.(bool)) {
		return ""
	}
	
	// Create debug context from ruleset
	debugCtx := createDebugContextFromRuleset(context, ruleset)
	if debugCtx == nil {
		return ""
	}
	
	// Get line number and filename from debug context
	lineNumber := 0
	fileName := ""
	
	if ln, ok := debugCtx["lineNumber"].(int); ok {
		lineNumber = ln
	}
	if fn, ok := debugCtx["fileName"].(string); ok {
		fileName = fn
	}
	
	if lineNumber == 0 || fileName == "" {
		return ""
	}
	
	var result string
	switch dumpLineNumbers {
	case "comments":
		result = asComment(lineNumber, fileName)
	case "mediaquery":
		result = asMediaQuery(lineNumber, fileName)
	case "all":
		result = asComment(lineNumber, fileName)
		if separator != "" {
			result += separator
		}
		result += asMediaQuery(lineNumber, fileName)
	}
	
	return result
}

func asComment(lineNumber int, fileName string) string {
	return fmt.Sprintf("/* line %d, %s */", lineNumber, fileName)
}

func asMediaQuery(lineNumber int, fileName string) string {
	return fmt.Sprintf("@media -sass-debug-info{filename{font-family:file\\:\\/\\/%s}line{font-family:\\00003%d}}",
		strings.ReplaceAll(fileName, "/", "\\/"), lineNumber)
}

func createDebugContextFromRuleset(context map[string]any, ruleset *Ruleset) map[string]any {
	fileInfo := ruleset.FileInfo()
	if fileInfo == nil {
		return nil
	}
	
	var filename string
	var lineNumber int
	
	if fn, ok := fileInfo["filename"]; ok {
		if fnStr, ok := fn.(string); ok {
			filename = fnStr
		}
	}
	
	if ln, ok := fileInfo["lineNumber"]; ok {
		if lnInt, ok := ln.(int); ok {
			lineNumber = lnInt
		}
	}
	if lineNumber == 0 {
		lineNumber = 1 // Default line number
	}
	
	return map[string]any{
		"fileName":   filename,
		"lineNumber": lineNumber,
	}
}

// Helper methods for array manipulation and rule checking

func (r *Ruleset) removeRuleAtIndex(rules []any, index int) []any {
	newRules := make([]any, len(rules)-1)
	copy(newRules, rules[:index])
	copy(newRules[index:], rules[index+1:])
	return newRules
}

func (r *Ruleset) insertRuleAtIndex(rules []any, index int, rule any) []any {
	newRules := make([]any, len(rules)+1)
	copy(newRules, rules[:index])
	newRules[index] = rule
	copy(newRules[index+1:], rules[index:])
	return newRules
}

// Matches JavaScript logic
func (r *Ruleset) shouldIncludeSubRule(subRule any, parentRuleset *Ruleset) bool {
	// Copy visibility info like JavaScript version
	if subNode, ok := subRule.(interface{ CopyVisibilityInfo(map[string]any) }); ok {
		subNode.CopyVisibilityInfo(parentRuleset.VisibilityInfo())
	}
	
	// Check if it's a variable declaration (like JavaScript: !(subRule instanceof Declaration) || !subRule.variable)
	if r.isVariableDeclaration(subRule) {
		return false // Don't include variable declarations
	}
	
	return true // Include everything else
}

func (r *Ruleset) isVariableDeclaration(rule any) bool {
	// Handle real Declaration types
	if decl, ok := rule.(*Declaration); ok {
		return decl.variable
	}
	
	// Handle mock declarations using reflection (for tests)
	if node, ok := rule.(interface{ GetType() string }); ok && node.GetType() == "Declaration" {
		if v := reflect.ValueOf(rule); v.Kind() == reflect.Ptr && !v.IsNil() {
			if elem := v.Elem(); elem.Kind() == reflect.Struct {
				if field := elem.FieldByName("variable"); field.IsValid() && field.Kind() == reflect.Bool {
					return field.Bool()
				}
			}
		}
	}
	
	return false
}

// flattenArray flattens a nested array structure (equivalent to utils.flattenArray in JavaScript)
func flattenArray(arr []any) []any {
	var result []any
	for _, item := range arr {
		if slice, ok := item.([]any); ok {
			result = append(result, flattenArray(slice)...)
		} else {
			result = append(result, item)
		}
	}
	return result
}

// Used by ExtendFinderVisitor
func (r *Ruleset) SetAllExtends(extends []*Extend) {
	r.AllExtends = extends
}

// Used by ProcessExtendsVisitor
func (r *Ruleset) GetAllExtends() []*Extend {
	return r.AllExtends
}

// runPreEvalVisitorReplacementsOnRuleset runs pre-eval visitor replacements on the ruleset.
// This is called after imports are processed to transform the AST before evaluation.
func runPreEvalVisitorReplacementsOnRuleset(ruleset *Ruleset, bridge *NodeJSPluginBridge) {
	// Collect all Variable nodes from the ruleset
	variables := collectRulesetVariableNodes(ruleset)

	if len(variables) == 0 {
		return
	}

	// Build the variable info list for JavaScript
	varInfos := make([]VariableInfo, 0, len(variables))
	varMap := make(map[string]*variableLocationRuleset)
	for id, vloc := range variables {
		varInfos = append(varInfos, VariableInfo{
			ID:   id,
			Name: vloc.variable.GetName(),
		})
		varMap[id] = vloc
	}

	// Check which variables should be replaced
	replacements, err := bridge.CheckVariableReplacements(varInfos)
	if err != nil {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[runPreEvalVisitorReplacementsOnRuleset] Error checking replacements: %v\n", err)
		}
		return
	}

	if len(replacements) == 0 {
		return
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[runPreEvalVisitorReplacementsOnRuleset] Found %d replacements\n", len(replacements))
	}

	// Apply the replacements
	for id, replInfo := range replacements {
		vloc, ok := varMap[id]
		if !ok {
			continue
		}

		// Create the replacement node based on the type
		replNode := createNodeFromReplacementRuleset(replInfo)
		if replNode == nil {
			continue
		}

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[runPreEvalVisitorReplacementsOnRuleset] Replacing variable %s (id=%s) with %T\n",
				vloc.variable.GetName(), id, replNode)
		}

		// Apply the replacement to the parent
		applyReplacementRuleset(vloc.parent, vloc.index, replNode)
	}
}

// variableLocationRuleset tracks where a Variable node is in the AST
type variableLocationRuleset struct {
	variable *Variable
	parent   any
	index    int // Index within parent's slice
}

// collectRulesetVariableNodes walks the ruleset and collects all Variable nodes with their locations
func collectRulesetVariableNodes(ruleset *Ruleset) map[string]*variableLocationRuleset {
	result := make(map[string]*variableLocationRuleset)
	idCounter := 0

	var walk func(n any, parent any, index int)
	walk = func(n any, parent any, index int) {
		if n == nil {
			return
		}

		switch v := n.(type) {
		case *Variable:
			id := fmt.Sprintf("var_%d", idCounter)
			idCounter++
			result[id] = &variableLocationRuleset{
				variable: v,
				parent:   parent,
				index:    index,
			}
		case *Ruleset:
			for i, rule := range v.Rules {
				walk(rule, v, i)
			}
		case *Declaration:
			walk(v.Value, v, 0)
		case *Value:
			for i, val := range v.Value {
				walk(val, v, i)
			}
		case *Expression:
			for i, val := range v.Value {
				walk(val, v, i)
			}
		case *MixinDefinition:
			for i, rule := range v.Rules {
				walk(rule, v, i)
			}
		case *MixinCall:
			for i, arg := range v.Arguments {
				if argVal, ok := arg.(map[string]any); ok {
					if val, ok := argVal["value"]; ok {
						walk(val, v, i)
					}
				} else {
					walk(arg, v, i)
				}
			}
		case *DetachedRuleset:
			if v.ruleset != nil {
				walk(v.ruleset, v, 0)
			}
		case *AtRule:
			if v.Value != nil {
				walk(v.Value, v, 0)
			}
			if v.Rules != nil {
				for i, rule := range v.Rules {
					walk(rule, v, i)
				}
			}
		case *Selector:
			for i, elem := range v.Elements {
				walk(elem, v, i)
			}
		case *Element:
			if v.Value != nil {
				walk(v.Value, v, 0)
			}
		case *Paren:
			if v.Value != nil {
				walk(v.Value, v, 0)
			}
		case *Negative:
			if v.Value != nil {
				walk(v.Value, v, 0)
			}
		case *Operation:
			for i, op := range v.Operands {
				walk(op, v, i)
			}
		case *Call:
			for i, arg := range v.Args {
				walk(arg, v, i)
			}
		case *Condition:
			walk(v.Lvalue, v, 0)
			walk(v.Rvalue, v, 1)
		}
	}

	walk(ruleset, nil, 0)
	return result
}

// createNodeFromReplacementRuleset creates a Go AST node from JavaScript replacement info
func createNodeFromReplacementRuleset(info map[string]any) any {
	nodeType, ok := info["_type"].(string)
	if !ok {
		return nil
	}

	switch nodeType {
	case "Quoted":
		value, _ := info["value"].(string)
		quote, _ := info["quote"].(string)
		escaped, _ := info["escaped"].(bool)
		return NewQuoted(quote, value, escaped, 0, nil)
	case "Dimension":
		var val float64
		switch v := info["value"].(type) {
		case float64:
			val = v
		case int:
			val = float64(v)
		}
		unit, _ := info["unit"].(string)
		dim, err := NewDimension(val, unit)
		if err != nil {
			return nil
		}
		return dim
	case "Color":
		rgb, _ := info["rgb"].([]any)
		alpha := 1.0
		if a, ok := info["alpha"].(float64); ok {
			alpha = a
		}
		var r, g, b float64
		if len(rgb) >= 3 {
			r, _ = rgb[0].(float64)
			g, _ = rgb[1].(float64)
			b, _ = rgb[2].(float64)
		}
		return &Color{
			Node:  NewNode(),
			RGB:   []float64{r, g, b},
			Alpha: alpha,
		}
	case "Keyword":
		value, _ := info["value"].(string)
		return &Keyword{
			Node:  NewNode(),
			value: value,
		}
	case "Anonymous":
		value, _ := info["value"].(string)
		return NewAnonymous(value, 0, nil, false, false, nil)
	default:
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[createNodeFromReplacementRuleset] Unknown node type: %s\n", nodeType)
		}
		return nil
	}
}

// applyReplacementRuleset replaces a node at the given index in the parent
func applyReplacementRuleset(parent any, index int, replacement any) {
	switch p := parent.(type) {
	case *Value:
		if index >= 0 && index < len(p.Value) {
			p.Value[index] = replacement
		}
	case *Expression:
		if index >= 0 && index < len(p.Value) {
			p.Value[index] = replacement
		}
	case *Ruleset:
		if index >= 0 && index < len(p.Rules) {
			p.Rules[index] = replacement
		}
	case *Declaration:
		// Value is the first child
		if index == 0 {
			if val, ok := replacement.(*Value); ok {
				p.Value = val
			}
		}
	case *MixinDefinition:
		if index >= 0 && index < len(p.Rules) {
			p.Rules[index] = replacement
		}
	case *Call:
		if index >= 0 && index < len(p.Args) {
			p.Args[index] = replacement
		}
	case *Operation:
		if index >= 0 && index < len(p.Operands) {
			p.Operands[index] = replacement
		}
	case *Paren:
		if index == 0 {
			p.Value = replacement
		}
	case *Negative:
		if index == 0 {
			p.Value = replacement
		}
	case *AtRule:
		if index == 0 && p.Value != nil {
			// Replacing the value
		} else if p.Rules != nil && index >= 0 && index < len(p.Rules) {
			p.Rules[index] = replacement
		}
	default:
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[applyReplacementRuleset] Unknown parent type: %T\n", parent)
		}
	}
} 