package less_go

import (
	"sync"
)

// Node pools for frequently allocated node types to reduce GC pressure.
// Memory profiling shows significant allocations in NewNode and Ruleset.Eval
// during Bootstrap4 compilation (27MB in NewNode, 26MB in Ruleset.Eval).

// nodePool is a pool for reusing Node objects.
var nodePool = sync.Pool{
	New: func() any {
		return &Node{}
	},
}

// rulesetPool is a pool for reusing Ruleset objects.
// Rulesets are the most frequently allocated node type (16.5MB in profiling).
var rulesetPool = sync.Pool{
	New: func() any {
		return &Ruleset{
			Selectors: make([]any, 0, 8),
			Rules:     make([]any, 0, 16),
		}
	},
}

// expressionPool is a pool for reusing Expression objects.
var expressionPool = sync.Pool{
	New: func() any {
		return &Expression{
			Value: make([]any, 0, 4),
		}
	},
}

// selectorPool is a pool for reusing Selector objects.
var selectorPool = sync.Pool{
	New: func() any {
		return &Selector{
			Elements:   make([]*Element, 0, 4),
			ExtendList: make([]any, 0, 2),
		}
	},
}

// declarationPool is a pool for reusing Declaration objects.
var declarationPool = sync.Pool{
	New: func() any {
		return &Declaration{}
	},
}

// GetNodeFromPool retrieves a Node from the pool.
// Call ReleaseNode when the Node is no longer needed.
func GetNodeFromPool() *Node {
	return nodePool.Get().(*Node)
}

// ReleaseNode returns a Node to the pool for reuse.
// The node should not be used after calling this function.
func ReleaseNode(n *Node) {
	if n == nil {
		return
	}
	// Clear all fields to prevent memory leaks
	n.Parent = nil
	n.VisibilityBlocks = nil
	n.NodeVisible = nil
	n.RootNode = nil
	n.Parsed = nil
	n.Value = nil
	n.Index = 0
	n.fileInfo = nil
	n.Parens = false
	n.ParensInOp = false
	n.TypeIndex = 0
	nodePool.Put(n)
}

// resetRuleset clears all fields of a Ruleset for reuse.
func resetRuleset(r *Ruleset) {
	// Reset embedded Node pointer (will be set by caller)
	r.Node = nil

	// Clear slices but keep underlying arrays for reuse
	for i := range r.Selectors {
		r.Selectors[i] = nil
	}
	r.Selectors = r.Selectors[:0]

	for i := range r.Rules {
		r.Rules[i] = nil
	}
	r.Rules = r.Rules[:0]

	// Clear boolean fields
	r.StrictImports = false
	r.AllowRoot = false
	r.Root = false
	r.ExtendOnEveryPath = false
	r.FirstRoot = false
	r.AllowImports = false
	r.MultiMedia = false
	r.InsideMixinDefinition = false

	// Clear cache maps
	r.lookups = nil
	r.variables = nil
	r.properties = nil
	r.rulesets = nil

	// Clear reference fields
	r.OriginalRuleset = nil
	r.FunctionRegistry = nil
	r.SelectorsParseFunc = nil
	r.ValueParseFunc = nil
	r.ParseContext = nil
	r.ParseImports = nil
	r.Parse = nil
	r.DebugInfo = nil

	// Clear Paths
	r.Paths = nil

	// Clear AllExtends
	r.AllExtends = nil

	// Clear LoadedPluginFunctions
	r.LoadedPluginFunctions = nil
}

// GetRulesetFromPool retrieves a Ruleset from the pool.
// The caller must set the Node field and other required fields.
func GetRulesetFromPool() *Ruleset {
	r := rulesetPool.Get().(*Ruleset)
	resetRuleset(r)
	return r
}

// ReleaseRuleset returns a Ruleset to the pool for reuse.
// The ruleset should not be used after calling this function.
func ReleaseRuleset(r *Ruleset) {
	if r == nil {
		return
	}
	resetRuleset(r)
	rulesetPool.Put(r)
}

// Release returns the Ruleset to the pool for reuse.
// This is a method version for convenience.
func (r *Ruleset) Release() {
	ReleaseRuleset(r)
}

// resetExpression clears all fields of an Expression for reuse.
func resetExpression(e *Expression) {
	// Reset embedded Node pointer (will be set by caller)
	e.Node = nil

	// Clear slice but keep underlying array for reuse
	for i := range e.Value {
		e.Value[i] = nil
	}
	e.Value = e.Value[:0]

	// Reset boolean fields
	e.NoSpacing = false
	e.Parens = false
	e.ParensInOp = false
}

// GetExpressionFromPool retrieves an Expression from the pool.
// The caller must set the Node field and other required fields.
func GetExpressionFromPool() *Expression {
	e := expressionPool.Get().(*Expression)
	resetExpression(e)
	return e
}

// ReleaseExpression returns an Expression to the pool for reuse.
// The expression should not be used after calling this function.
func ReleaseExpression(e *Expression) {
	if e == nil {
		return
	}
	resetExpression(e)
	expressionPool.Put(e)
}

// Release returns the Expression to the pool for reuse.
// This is a method version for convenience.
func (e *Expression) Release() {
	ReleaseExpression(e)
}

// resetSelector clears all fields of a Selector for reuse.
func resetSelector(s *Selector) {
	// Reset embedded Node pointer (will be set by caller)
	s.Node = nil

	// Clear slices but keep underlying arrays for reuse
	for i := range s.Elements {
		s.Elements[i] = nil
	}
	s.Elements = s.Elements[:0]

	for i := range s.ExtendList {
		s.ExtendList[i] = nil
	}
	s.ExtendList = s.ExtendList[:0]

	// Reset other fields
	s.Condition = nil
	s.EvaldCondition = false
	s.MixinElements_ = nil
	s.MediaEmpty = false
	s.ParseFunc = nil
	s.ParseContext = nil
	s.ParseImports = nil
}

// GetSelectorFromPool retrieves a Selector from the pool.
// The caller must set the Node field and other required fields.
func GetSelectorFromPool() *Selector {
	s := selectorPool.Get().(*Selector)
	resetSelector(s)
	return s
}

// ReleaseSelector returns a Selector to the pool for reuse.
// The selector should not be used after calling this function.
func ReleaseSelector(s *Selector) {
	if s == nil {
		return
	}
	resetSelector(s)
	selectorPool.Put(s)
}

// Release returns the Selector to the pool for reuse.
// This is a method version for convenience.
func (s *Selector) Release() {
	ReleaseSelector(s)
}

// resetDeclaration clears all fields of a Declaration for reuse.
func resetDeclaration(d *Declaration) {
	// Reset embedded Node pointer (will be set by caller)
	d.Node = nil

	// Reset all fields
	d.name = nil
	d.Value = nil
	d.important = ""
	d.merge = nil
	d.inline = false
	d.variable = false
}

// GetDeclarationFromPool retrieves a Declaration from the pool.
// The caller must set the Node field and other required fields.
func GetDeclarationFromPool() *Declaration {
	d := declarationPool.Get().(*Declaration)
	resetDeclaration(d)
	return d
}

// ReleaseDeclaration returns a Declaration to the pool for reuse.
// The declaration should not be used after calling this function.
func ReleaseDeclaration(d *Declaration) {
	if d == nil {
		return
	}
	resetDeclaration(d)
	declarationPool.Put(d)
}

// Release returns the Declaration to the pool for reuse.
// This is a method version for convenience.
func (d *Declaration) Release() {
	ReleaseDeclaration(d)
}
