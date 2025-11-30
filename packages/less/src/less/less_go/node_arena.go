package less_go

// NodeArena provides zero-allocation node reuse for single-threaded compilation.
// Unlike sync.Pool, this has no mutex overhead since Less compilation is single-threaded.
//
// The arena pre-allocates slabs of nodes and tracks available slots via free lists.
// When a node is "allocated", we return a pre-existing object and reset it.
// The arena can be reset to reuse all nodes for the next compilation.
type NodeArena struct {
	// Slabs of pre-allocated nodes
	rulesets     []*Ruleset
	selectors    []*Selector
	elements     []*Element
	declarations []*Declaration
	expressions  []*Expression
	units        []*Unit
	nodes        []*Node

	// Current allocation index for each type
	rulesetIdx     int
	selectorIdx    int
	elementIdx     int
	declarationIdx int
	expressionIdx  int
	unitIdx        int
	nodeIdx        int
}

// Default slab sizes - tuned for typical LESS files
const (
	defaultRulesetSlabSize     = 256
	defaultSelectorSlabSize    = 512
	defaultElementSlabSize     = 1024
	defaultDeclarationSlabSize = 512
	defaultExpressionSlabSize  = 512
	defaultUnitSlabSize        = 256
	defaultNodeSlabSize        = 1024
)

// NewNodeArena creates a new arena with default slab sizes.
// The arena pre-allocates memory for faster node creation during parsing.
func NewNodeArena() *NodeArena {
	return NewNodeArenaWithSizes(
		defaultRulesetSlabSize,
		defaultSelectorSlabSize,
		defaultElementSlabSize,
		defaultDeclarationSlabSize,
		defaultExpressionSlabSize,
		defaultUnitSlabSize,
		defaultNodeSlabSize,
	)
}

// NewNodeArenaWithSizes creates an arena with custom slab sizes.
func NewNodeArenaWithSizes(rulesetSize, selectorSize, elementSize, declarationSize, expressionSize, unitSize, nodeSize int) *NodeArena {
	arena := &NodeArena{
		rulesets:     make([]*Ruleset, 0, rulesetSize),
		selectors:    make([]*Selector, 0, selectorSize),
		elements:     make([]*Element, 0, elementSize),
		declarations: make([]*Declaration, 0, declarationSize),
		expressions:  make([]*Expression, 0, expressionSize),
		units:        make([]*Unit, 0, unitSize),
		nodes:        make([]*Node, 0, nodeSize),
	}
	return arena
}

// Reset resets the arena for reuse, allowing all previously allocated nodes
// to be reused in the next compilation. This is O(1) - we just reset indices.
func (a *NodeArena) Reset() {
	a.rulesetIdx = 0
	a.selectorIdx = 0
	a.elementIdx = 0
	a.declarationIdx = 0
	a.expressionIdx = 0
	a.unitIdx = 0
	a.nodeIdx = 0
}

// GetRuleset returns a Ruleset from the arena, allocating a new one if needed.
// The returned Ruleset is reset and ready for use.
func (a *NodeArena) GetRuleset() *Ruleset {
	if a.rulesetIdx < len(a.rulesets) {
		r := a.rulesets[a.rulesetIdx]
		a.rulesetIdx++
		resetRuleset(r)
		return r
	}

	// Need to allocate a new Ruleset
	r := &Ruleset{
		Selectors: make([]any, 0, 8),
		Rules:     make([]any, 0, 16),
	}
	a.rulesets = append(a.rulesets, r)
	a.rulesetIdx++
	return r
}

// GetSelector returns a Selector from the arena.
func (a *NodeArena) GetSelector() *Selector {
	if a.selectorIdx < len(a.selectors) {
		s := a.selectors[a.selectorIdx]
		a.selectorIdx++
		resetSelector(s)
		return s
	}

	s := &Selector{
		Elements:   make([]*Element, 0, 4),
		ExtendList: make([]any, 0, 2),
	}
	a.selectors = append(a.selectors, s)
	a.selectorIdx++
	return s
}

// GetElement returns an Element from the arena.
func (a *NodeArena) GetElement() *Element {
	if a.elementIdx < len(a.elements) {
		e := a.elements[a.elementIdx]
		a.elementIdx++
		resetElement(e)
		return e
	}

	e := &Element{}
	a.elements = append(a.elements, e)
	a.elementIdx++
	return e
}

// GetDeclaration returns a Declaration from the arena.
func (a *NodeArena) GetDeclaration() *Declaration {
	if a.declarationIdx < len(a.declarations) {
		d := a.declarations[a.declarationIdx]
		a.declarationIdx++
		resetDeclaration(d)
		return d
	}

	d := &Declaration{}
	a.declarations = append(a.declarations, d)
	a.declarationIdx++
	return d
}

// GetExpression returns an Expression from the arena.
func (a *NodeArena) GetExpression() *Expression {
	if a.expressionIdx < len(a.expressions) {
		e := a.expressions[a.expressionIdx]
		a.expressionIdx++
		resetExpression(e)
		return e
	}

	e := &Expression{
		Value: make([]any, 0, 4),
	}
	a.expressions = append(a.expressions, e)
	a.expressionIdx++
	return e
}

// GetUnit returns a Unit from the arena.
func (a *NodeArena) GetUnit() *Unit {
	if a.unitIdx < len(a.units) {
		u := a.units[a.unitIdx]
		a.unitIdx++
		resetUnit(u)
		return u
	}

	u := &Unit{
		Numerator:   make([]string, 0, 4),
		Denominator: make([]string, 0, 2),
	}
	a.units = append(a.units, u)
	a.unitIdx++
	return u
}

// GetNode returns a Node from the arena.
func (a *NodeArena) GetNode() *Node {
	if a.nodeIdx < len(a.nodes) {
		n := a.nodes[a.nodeIdx]
		a.nodeIdx++
		// Reset node fields
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
		return n
	}

	n := &Node{}
	a.nodes = append(a.nodes, n)
	a.nodeIdx++
	return n
}

// Stats returns allocation statistics for debugging/profiling.
func (a *NodeArena) Stats() ArenaStats {
	return ArenaStats{
		RulesetsAllocated:     len(a.rulesets),
		RulesetsUsed:          a.rulesetIdx,
		SelectorsAllocated:    len(a.selectors),
		SelectorsUsed:         a.selectorIdx,
		ElementsAllocated:     len(a.elements),
		ElementsUsed:          a.elementIdx,
		DeclarationsAllocated: len(a.declarations),
		DeclarationsUsed:      a.declarationIdx,
		ExpressionsAllocated:  len(a.expressions),
		ExpressionsUsed:       a.expressionIdx,
		UnitsAllocated:        len(a.units),
		UnitsUsed:             a.unitIdx,
		NodesAllocated:        len(a.nodes),
		NodesUsed:             a.nodeIdx,
	}
}

// ArenaStats contains allocation statistics.
type ArenaStats struct {
	RulesetsAllocated     int
	RulesetsUsed          int
	SelectorsAllocated    int
	SelectorsUsed         int
	ElementsAllocated     int
	ElementsUsed          int
	DeclarationsAllocated int
	DeclarationsUsed      int
	ExpressionsAllocated  int
	ExpressionsUsed       int
	UnitsAllocated        int
	UnitsUsed             int
	NodesAllocated        int
	NodesUsed             int
}

// Arena-aware node constructors that fall back to pool when arena is nil

// GetRulesetFromArena returns a Ruleset from arena if available, otherwise from pool.
func GetRulesetFromArena(arena *NodeArena) *Ruleset {
	if arena != nil {
		return arena.GetRuleset()
	}
	return GetRulesetFromPool()
}

// GetSelectorFromArena returns a Selector from arena if available, otherwise from pool.
func GetSelectorFromArena(arena *NodeArena) *Selector {
	if arena != nil {
		return arena.GetSelector()
	}
	return GetSelectorFromPool()
}

// GetElementFromArena returns an Element from arena if available, otherwise from pool.
func GetElementFromArena(arena *NodeArena) *Element {
	if arena != nil {
		return arena.GetElement()
	}
	return GetElementFromPool()
}

// GetDeclarationFromArena returns a Declaration from arena if available, otherwise from pool.
func GetDeclarationFromArena(arena *NodeArena) *Declaration {
	if arena != nil {
		return arena.GetDeclaration()
	}
	return GetDeclarationFromPool()
}

// GetExpressionFromArena returns an Expression from arena if available, otherwise from pool.
func GetExpressionFromArena(arena *NodeArena) *Expression {
	if arena != nil {
		return arena.GetExpression()
	}
	return GetExpressionFromPool()
}

// GetUnitFromArena returns a Unit from arena if available, otherwise from pool.
func GetUnitFromArena(arena *NodeArena) *Unit {
	if arena != nil {
		return arena.GetUnit()
	}
	return GetUnitFromPool()
}

// GetNodeFromArena returns a Node from arena if available, otherwise from pool.
func GetNodeFromArena(arena *NodeArena) *Node {
	if arena != nil {
		return arena.GetNode()
	}
	return GetNodeFromPool()
}
