package less_go

import (
	"sync"
	"unsafe"
)

// unsafePointer converts a typed pointer to unsafe.Pointer for use in deduplication maps
func unsafePointer[T any](p *T) unsafe.Pointer {
	return unsafe.Pointer(p)
}

// nodePool is a pool for reusing Node objects.
var nodePool = sync.Pool{
	New: func() any {
		return &Node{}
	},
}

// rulesetPool is a pool for reusing Ruleset objects.
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

// elementPool is a pool for reusing Element objects.
var elementPool = sync.Pool{
	New: func() any {
		return &Element{}
	},
}

// unitPool is a pool for reusing Unit objects.
var unitPool = sync.Pool{
	New: func() any {
		return &Unit{
			Numerator:   make([]string, 0, 4),
			Denominator: make([]string, 0, 2),
		}
	},
}

func GetNodeFromPool() *Node {
	return nodePool.Get().(*Node)
}

func ReleaseNode(n *Node) {
	if n == nil {
		return
	}
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

func resetRuleset(r *Ruleset) {
	r.Node = nil
	for i := range r.Selectors {
		r.Selectors[i] = nil
	}
	r.Selectors = r.Selectors[:0]

	for i := range r.Rules {
		r.Rules[i] = nil
	}
	r.Rules = r.Rules[:0]
	r.StrictImports = false
	r.AllowRoot = false
	r.Root = false
	r.ExtendOnEveryPath = false
	r.FirstRoot = false
	r.AllowImports = false
	r.MultiMedia = false
	r.InsideMixinDefinition = false
	r.lookups = nil
	r.variables = nil
	r.properties = nil
	r.OriginalRuleset = nil
	r.FunctionRegistry = nil
	r.SelectorsParseFunc = nil
	r.ValueParseFunc = nil
	r.ParseContext = nil
	r.ParseImports = nil
	r.Parse = nil
	r.DebugInfo = nil
	r.Paths = nil
	r.AllExtends = nil
	r.LoadedPluginFunctions = nil
}

func GetRulesetFromPool() *Ruleset {
	r := rulesetPool.Get().(*Ruleset)
	resetRuleset(r)
	return r
}

func ReleaseRuleset(r *Ruleset) {
	if r == nil {
		return
	}
	resetRuleset(r)
	rulesetPool.Put(r)
}

func (r *Ruleset) Release() {
	ReleaseRuleset(r)
}

func resetExpression(e *Expression) {
	e.Node = nil
	for i := range e.Value {
		e.Value[i] = nil
	}
	e.Value = e.Value[:0]
	e.NoSpacing = false
	e.Parens = false
	e.ParensInOp = false
}

func GetExpressionFromPool() *Expression {
	e := expressionPool.Get().(*Expression)
	resetExpression(e)
	return e
}

func ReleaseExpression(e *Expression) {
	if e == nil {
		return
	}
	resetExpression(e)
	expressionPool.Put(e)
}

func (e *Expression) Release() {
	ReleaseExpression(e)
}

func resetSelector(s *Selector) {
	s.Node = nil
	for i := range s.Elements {
		s.Elements[i] = nil
	}
	s.Elements = s.Elements[:0]

	for i := range s.ExtendList {
		s.ExtendList[i] = nil
	}
	s.ExtendList = s.ExtendList[:0]
	s.Condition = nil
	s.EvaldCondition = false
	s.MixinElements_ = nil
	s.MediaEmpty = false
	s.ParseFunc = nil
	s.ParseContext = nil
	s.ParseImports = nil
}

func GetSelectorFromPool() *Selector {
	s := selectorPool.Get().(*Selector)
	resetSelector(s)
	return s
}

func ReleaseSelector(s *Selector) {
	if s == nil {
		return
	}
	resetSelector(s)
	selectorPool.Put(s)
}

func (s *Selector) Release() {
	ReleaseSelector(s)
}

func resetDeclaration(d *Declaration) {
	d.Node = nil
	d.name = nil
	d.Value = nil
	d.important = ""
	d.merge = nil
	d.inline = false
	d.variable = false
}

func GetDeclarationFromPool() *Declaration {
	d := declarationPool.Get().(*Declaration)
	resetDeclaration(d)
	return d
}

func ReleaseDeclaration(d *Declaration) {
	if d == nil {
		return
	}
	resetDeclaration(d)
	declarationPool.Put(d)
}

func (d *Declaration) Release() {
	ReleaseDeclaration(d)
}

func resetElement(e *Element) {
	e.Node = nil
	e.Combinator = nil
	e.Value = nil
	e.IsVariable = false
}

func GetElementFromPool() *Element {
	e := elementPool.Get().(*Element)
	resetElement(e)
	return e
}

func ReleaseElement(e *Element) {
	if e == nil {
		return
	}
	resetElement(e)
	elementPool.Put(e)
}

func (e *Element) Release() {
	ReleaseElement(e)
}

func resetUnit(u *Unit) {
	u.Node = nil
	// Clear slices but keep capacity
	for i := range u.Numerator {
		u.Numerator[i] = ""
	}
	u.Numerator = u.Numerator[:0]

	for i := range u.Denominator {
		u.Denominator[i] = ""
	}
	u.Denominator = u.Denominator[:0]
	u.BackupUnit = ""
}

func GetUnitFromPool() *Unit {
	u := unitPool.Get().(*Unit)
	resetUnit(u)
	return u
}

func ReleaseUnit(u *Unit) {
	if u == nil {
		return
	}
	resetUnit(u)
	unitPool.Put(u)
}

func (u *Unit) Release() {
	ReleaseUnit(u)
}

// Releasable is an interface for nodes that can be released back to pools
type Releasable interface {
	Release()
}

// ReleaseTree recursively releases all pooled nodes in an AST tree.
// This should be called after ToCSS is complete and the tree is no longer needed.
// It uses a seen map to avoid double-releasing shared nodes.
func ReleaseTree(root any) {
	if root == nil {
		return
	}
	seen := make(map[uintptr]bool)
	releaseTreeInternal(root, seen)
}

// releaseTreeInternal is the internal recursive implementation
func releaseTreeInternal(node any, seen map[uintptr]bool) {
	if node == nil {
		return
	}

	// Handle slices of nodes
	switch v := node.(type) {
	case []any:
		for _, item := range v {
			releaseTreeInternal(item, seen)
		}
		return
	case []*Selector:
		for _, item := range v {
			releaseTreeInternal(item, seen)
		}
		return
	case []*Element:
		for _, item := range v {
			releaseTreeInternal(item, seen)
		}
		return
	case [][]any:
		for _, path := range v {
			for _, item := range path {
				releaseTreeInternal(item, seen)
			}
		}
		return
	}

	// Get pointer for deduplication
	ptr := getNodePointer(node)
	if ptr == 0 {
		return
	}
	if seen[ptr] {
		return
	}
	seen[ptr] = true

	// Release children first, then the node itself
	switch n := node.(type) {
	case *Ruleset:
		if n == nil {
			return
		}
		// Release children
		for _, sel := range n.Selectors {
			releaseTreeInternal(sel, seen)
		}
		for _, rule := range n.Rules {
			releaseTreeInternal(rule, seen)
		}
		if n.Paths != nil {
			for _, path := range n.Paths {
				for _, item := range path {
					releaseTreeInternal(item, seen)
				}
			}
		}
		// Don't release OriginalRuleset - it may be shared
		// Release the ruleset's embedded Node
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
		// Release the ruleset itself
		ReleaseRuleset(n)

	case *Selector:
		if n == nil {
			return
		}
		for _, elem := range n.Elements {
			releaseTreeInternal(elem, seen)
		}
		for _, ext := range n.ExtendList {
			releaseTreeInternal(ext, seen)
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
		ReleaseSelector(n)

	case *Element:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
		ReleaseElement(n)

	case *Expression:
		if n == nil {
			return
		}
		for _, val := range n.Value {
			releaseTreeInternal(val, seen)
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
		ReleaseExpression(n)

	case *Declaration:
		if n == nil {
			return
		}
		releaseTreeInternal(n.name, seen)
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
		ReleaseDeclaration(n)

	case *Unit:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
		ReleaseUnit(n)

	case *Node:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		ReleaseNode(n)

	// Handle other node types that embed *Node but aren't pooled
	// We still want to release their embedded Node
	case *Value:
		if n == nil {
			return
		}
		for _, val := range n.Value {
			releaseTreeInternal(val, seen)
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Dimension:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Unit, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Color:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Quoted:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Keyword:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Anonymous:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Call:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Args, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *URL:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Operation:
		if n == nil {
			return
		}
		for _, operand := range n.Operands {
			releaseTreeInternal(operand, seen)
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Negative:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Paren:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Comment:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *AtRule:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		releaseTreeInternal(n.Rules, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *MixinDefinition:
		if n == nil {
			return
		}
		for _, param := range n.Params {
			releaseTreeInternal(param, seen)
		}
		releaseTreeInternal(n.Condition, seen)
		// MixinDefinition embeds *Ruleset, release the Rules
		for _, rule := range n.Rules {
			releaseTreeInternal(rule, seen)
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *MixinCall:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Selector, seen)
		releaseTreeInternal(n.Arguments, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Import:
		if n == nil {
			return
		}
		releaseTreeInternal(n.GetPath(), seen)
		// features field is private, skip releasing it
		releaseTreeInternal(n.GetRoot(), seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Media:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Features, seen)
		releaseTreeInternal(n.Rules, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Container:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Features, seen)
		releaseTreeInternal(n.Rules, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *NestableAtRulePrototype:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Features, seen)
		releaseTreeInternal(n.Rules, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Extend:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Selector, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *DetachedRuleset:
		if n == nil {
			return
		}
		releaseTreeInternal(n.GetRuleset(), seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Variable:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Property:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Combinator:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Condition:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Lvalue, seen)
		releaseTreeInternal(n.Rvalue, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Attribute:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Key, seen)
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *Assignment:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *UnicodeDescriptor:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *NamespaceValue:
		if n == nil {
			return
		}
		releaseTreeInternal(n.Value, seen)
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}

	case *VariableCall:
		if n == nil {
			return
		}
		if n.Node != nil {
			nodePtr := getNodePointer(n.Node)
			if nodePtr != 0 && !seen[nodePtr] {
				seen[nodePtr] = true
				ReleaseNode(n.Node)
			}
		}
	}
}

// getNodePointer returns a unique pointer value for deduplication
func getNodePointer(node any) uintptr {
	if node == nil {
		return 0
	}
	switch n := node.(type) {
	case *Ruleset:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Selector:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Element:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Expression:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Declaration:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Unit:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Node:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Value:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Dimension:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Color:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Quoted:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Keyword:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Anonymous:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Call:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *URL:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Operation:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Negative:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Paren:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Comment:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *AtRule:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *MixinDefinition:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *MixinCall:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Import:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Media:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Container:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *NestableAtRulePrototype:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Extend:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *DetachedRuleset:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Variable:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Property:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Combinator:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Condition:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Attribute:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *Assignment:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *UnicodeDescriptor:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *NamespaceValue:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	case *VariableCall:
		if n == nil {
			return 0
		}
		return uintptr(unsafePointer(n))
	default:
		return 0
	}
}
