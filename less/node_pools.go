package less_go

import (
	"sync"
)

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

// dimensionPool is a pool for reusing Dimension objects.
var dimensionPool = sync.Pool{
	New: func() any {
		return &Dimension{}
	},
}

// keywordPool is a pool for reusing Keyword objects.
var keywordPool = sync.Pool{
	New: func() any {
		return &Keyword{}
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

func resetDimension(d *Dimension) {
	d.Node = nil
	d.Value = 0
	d.Unit = nil
}

func GetDimensionFromPool() *Dimension {
	d := dimensionPool.Get().(*Dimension)
	resetDimension(d)
	return d
}

func ReleaseDimension(d *Dimension) {
	if d == nil {
		return
	}
	resetDimension(d)
	dimensionPool.Put(d)
}

func (d *Dimension) Release() {
	ReleaseDimension(d)
}

func resetKeyword(k *Keyword) {
	if k.Node != nil {
		k.Node.Value = nil // Reset the inherited Value field from Node
	}
	k.Node = nil
	k.value = ""
	k.type_ = ""
}

func GetKeywordFromPool() *Keyword {
	k := keywordPool.Get().(*Keyword)
	resetKeyword(k)
	return k
}

func ReleaseKeyword(k *Keyword) {
	if k == nil {
		return
	}
	resetKeyword(k)
	keywordPool.Put(k)
}

func (k *Keyword) Release() {
	ReleaseKeyword(k)
}

// Static units for common cases to avoid allocations in hot paths.
// These are immutable and should never be modified or released to pool.
// Only includes units actually used in hot paths (math functions).
var (
	staticEmptyUnit *Unit
	staticRadUnit   *Unit
)

func initStaticUnits() {
	// Empty unit (used by sin, cos, tan)
	staticEmptyUnit = &Unit{
		Node:        NewNode(),
		Numerator:   []string{},
		Denominator: []string{},
		BackupUnit:  "",
	}

	// Rad unit (used by asin, acos, atan)
	staticRadUnit = &Unit{
		Node:        NewNode(),
		Numerator:   []string{"rad"},
		Denominator: []string{},
		BackupUnit:  "rad",
	}
}

// GetEmptyUnit returns a static empty unit. Do not modify or release.
func GetEmptyUnit() *Unit {
	return staticEmptyUnit
}

// GetRadUnit returns a static rad unit. Do not modify or release.
func GetRadUnit() *Unit {
	return staticRadUnit
}

func init() {
	initStaticUnits()
}
