package less_go

import (
	"sync"
)

// NodeArena provides arena-style allocation for AST nodes to reduce GC pressure.
// Instead of allocating individual nodes on the heap, nodes are allocated from
// pre-allocated slices. After compilation completes, the arena can be reset
// and reused for the next compilation.
//
// This reduces GC pressure by:
// 1. Allocating nodes in contiguous memory blocks
// 2. Reducing the number of individual allocations
// 3. Allowing bulk deallocation by resetting the arena
type NodeArena struct {
	mu sync.Mutex

	// Pre-allocated slices for common node types
	nodes        []Node
	rulesets     []Ruleset
	declarations []Declaration
	selectors    []Selector
	elements     []Element
	expressions  []Expression
	values       []Value
	dimensions   []Dimension
	colors       []Color
	keywords     []Keyword
	anonymouses  []Anonymous
	quoteds      []Quoted
	units        []Unit
	combinators  []Combinator

	// Current indices into each slice
	nodeIdx        int
	rulesetIdx     int
	declarationIdx int
	selectorIdx    int
	elementIdx     int
	expressionIdx  int
	valueIdx       int
	dimensionIdx   int
	colorIdx       int
	keywordIdx     int
	anonymousIdx   int
	quotedIdx      int
	unitIdx        int
	combinatorIdx  int

	// Initial capacities (used for growing)
	initialCapacity int
}

// ArenaConfig specifies the initial capacity for each node type in the arena
type ArenaConfig struct {
	Nodes        int
	Rulesets     int
	Declarations int
	Selectors    int
	Elements     int
	Expressions  int
	Values       int
	Dimensions   int
	Colors       int
	Keywords     int
	Anonymouses  int
	Quoteds      int
	Units        int
	Combinators  int
}

// DefaultArenaConfig returns a configuration sized for typical LESS compilations
func DefaultArenaConfig() *ArenaConfig {
	return &ArenaConfig{
		Nodes:        2000,
		Rulesets:     500,
		Declarations: 500,
		Selectors:    300,
		Elements:     600,
		Expressions:  400,
		Values:       400,
		Dimensions:   300,
		Colors:       200,
		Keywords:     300,
		Anonymouses:  200,
		Quoteds:      200,
		Units:        200,
		Combinators:  600,
	}
}

// LargeArenaConfig returns a configuration for large LESS files (like Bootstrap)
func LargeArenaConfig() *ArenaConfig {
	return &ArenaConfig{
		Nodes:        10000,
		Rulesets:     2000,
		Declarations: 2000,
		Selectors:    1500,
		Elements:     3000,
		Expressions:  2000,
		Values:       2000,
		Dimensions:   1500,
		Colors:       1000,
		Keywords:     1500,
		Anonymouses:  1000,
		Quoteds:      1000,
		Units:        1000,
		Combinators:  3000,
	}
}

// NewNodeArena creates a new arena with the given configuration
func NewNodeArena(config *ArenaConfig) *NodeArena {
	if config == nil {
		config = DefaultArenaConfig()
	}

	return &NodeArena{
		nodes:           make([]Node, config.Nodes),
		rulesets:        make([]Ruleset, config.Rulesets),
		declarations:    make([]Declaration, config.Declarations),
		selectors:       make([]Selector, config.Selectors),
		elements:        make([]Element, config.Elements),
		expressions:     make([]Expression, config.Expressions),
		values:          make([]Value, config.Values),
		dimensions:      make([]Dimension, config.Dimensions),
		colors:          make([]Color, config.Colors),
		keywords:        make([]Keyword, config.Keywords),
		anonymouses:     make([]Anonymous, config.Anonymouses),
		quoteds:         make([]Quoted, config.Quoteds),
		units:           make([]Unit, config.Units),
		combinators:     make([]Combinator, config.Combinators),
		initialCapacity: 1000, // Default grow size
	}
}

// Reset clears the arena for reuse without deallocating the underlying memory.
// This should be called after each compilation completes.
func (a *NodeArena) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Reset indices - no need to zero the memory, just reuse it
	a.nodeIdx = 0
	a.rulesetIdx = 0
	a.declarationIdx = 0
	a.selectorIdx = 0
	a.elementIdx = 0
	a.expressionIdx = 0
	a.valueIdx = 0
	a.dimensionIdx = 0
	a.colorIdx = 0
	a.keywordIdx = 0
	a.anonymousIdx = 0
	a.quotedIdx = 0
	a.unitIdx = 0
	a.combinatorIdx = 0
}

// Stats returns allocation statistics for the arena
type ArenaStats struct {
	NodeAllocs        int
	RulesetAllocs     int
	DeclarationAllocs int
	SelectorAllocs    int
	ElementAllocs     int
	ExpressionAllocs  int
	ValueAllocs       int
	DimensionAllocs   int
	ColorAllocs       int
	KeywordAllocs     int
	AnonymousAllocs   int
	QuotedAllocs      int
	UnitAllocs        int
	CombinatorAllocs  int
	TotalAllocs       int
}

// Stats returns the current allocation statistics
func (a *NodeArena) Stats() ArenaStats {
	a.mu.Lock()
	defer a.mu.Unlock()

	return ArenaStats{
		NodeAllocs:        a.nodeIdx,
		RulesetAllocs:     a.rulesetIdx,
		DeclarationAllocs: a.declarationIdx,
		SelectorAllocs:    a.selectorIdx,
		ElementAllocs:     a.elementIdx,
		ExpressionAllocs:  a.expressionIdx,
		ValueAllocs:       a.valueIdx,
		DimensionAllocs:   a.dimensionIdx,
		ColorAllocs:       a.colorIdx,
		KeywordAllocs:     a.keywordIdx,
		AnonymousAllocs:   a.anonymousIdx,
		QuotedAllocs:      a.quotedIdx,
		UnitAllocs:        a.unitIdx,
		CombinatorAllocs:  a.combinatorIdx,
		TotalAllocs: a.nodeIdx + a.rulesetIdx + a.declarationIdx + a.selectorIdx +
			a.elementIdx + a.expressionIdx + a.valueIdx + a.dimensionIdx +
			a.colorIdx + a.keywordIdx + a.anonymousIdx + a.quotedIdx +
			a.unitIdx + a.combinatorIdx,
	}
}

// AllocNode allocates a Node from the arena
func (a *NodeArena) AllocNode() *Node {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.nodeIdx >= len(a.nodes) {
		// Grow the slice
		newNodes := make([]Node, len(a.nodes)+a.initialCapacity)
		a.nodes = append(a.nodes, newNodes...)
	}

	n := &a.nodes[a.nodeIdx]
	a.nodeIdx++

	// Initialize the node
	*n = Node{}
	return n
}

// AllocRuleset allocates a Ruleset from the arena
func (a *NodeArena) AllocRuleset() *Ruleset {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.rulesetIdx >= len(a.rulesets) {
		newRulesets := make([]Ruleset, len(a.rulesets)+a.initialCapacity/2)
		a.rulesets = append(a.rulesets, newRulesets...)
	}

	r := &a.rulesets[a.rulesetIdx]
	a.rulesetIdx++

	// Reuse existing slices if they have capacity, otherwise allocate
	// This reduces allocations when arena is reused after Reset()
	if cap(r.Selectors) >= 8 {
		r.Selectors = r.Selectors[:0]
	} else {
		r.Selectors = make([]any, 0, 8)
	}
	if cap(r.Rules) >= 16 {
		r.Rules = r.Rules[:0]
	} else {
		r.Rules = make([]any, 0, 16)
	}
	// Clear other fields
	r.Node = nil
	r.StrictImports = false
	r.AllowRoot = false
	r.lookups = nil
	r.variables = nil
	r.properties = nil
	r.OriginalRuleset = nil
	r.Root = false
	r.ExtendOnEveryPath = false
	r.Paths = nil
	r.FirstRoot = false
	r.AllowImports = false
	r.AllExtends = nil
	r.FunctionRegistry = nil
	r.SelectorsParseFunc = nil
	r.ValueParseFunc = nil
	r.ParseContext = nil
	r.ParseImports = nil
	r.Parse = nil
	r.DebugInfo = nil
	r.MultiMedia = false
	r.InsideMixinDefinition = false
	r.LoadedPluginFunctions = nil
	return r
}

// AllocDeclaration allocates a Declaration from the arena
func (a *NodeArena) AllocDeclaration() *Declaration {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.declarationIdx >= len(a.declarations) {
		newDeclarations := make([]Declaration, len(a.declarations)+a.initialCapacity/2)
		a.declarations = append(a.declarations, newDeclarations...)
	}

	d := &a.declarations[a.declarationIdx]
	a.declarationIdx++

	*d = Declaration{}
	return d
}

// AllocSelector allocates a Selector from the arena
func (a *NodeArena) AllocSelector() *Selector {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.selectorIdx >= len(a.selectors) {
		newSelectors := make([]Selector, len(a.selectors)+a.initialCapacity/2)
		a.selectors = append(a.selectors, newSelectors...)
	}

	s := &a.selectors[a.selectorIdx]
	a.selectorIdx++

	// Reuse existing slices if they have capacity
	if cap(s.Elements) >= 4 {
		s.Elements = s.Elements[:0]
	} else {
		s.Elements = make([]*Element, 0, 4)
	}
	if cap(s.ExtendList) >= 2 {
		s.ExtendList = s.ExtendList[:0]
	} else {
		s.ExtendList = make([]any, 0, 2)
	}
	// Clear other fields
	s.Node = nil
	s.Condition = nil
	s.EvaldCondition = false
	s.MixinElements_ = nil
	s.MediaEmpty = false
	s.ParseFunc = nil
	s.ParseContext = nil
	s.ParseImports = nil
	return s
}

// AllocElement allocates an Element from the arena
func (a *NodeArena) AllocElement() *Element {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.elementIdx >= len(a.elements) {
		newElements := make([]Element, len(a.elements)+a.initialCapacity)
		a.elements = append(a.elements, newElements...)
	}

	e := &a.elements[a.elementIdx]
	a.elementIdx++

	*e = Element{}
	return e
}

// AllocExpression allocates an Expression from the arena
func (a *NodeArena) AllocExpression() *Expression {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.expressionIdx >= len(a.expressions) {
		newExpressions := make([]Expression, len(a.expressions)+a.initialCapacity/2)
		a.expressions = append(a.expressions, newExpressions...)
	}

	e := &a.expressions[a.expressionIdx]
	a.expressionIdx++

	// Reuse existing slices if they have capacity
	if cap(e.Value) >= 4 {
		e.Value = e.Value[:0]
	} else {
		e.Value = make([]any, 0, 4)
	}
	// Clear other fields
	e.Node = nil
	e.NoSpacing = false
	e.Parens = false
	e.ParensInOp = false
	return e
}

// AllocValue allocates a Value from the arena
func (a *NodeArena) AllocValue() *Value {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.valueIdx >= len(a.values) {
		newValues := make([]Value, len(a.values)+a.initialCapacity/2)
		a.values = append(a.values, newValues...)
	}

	v := &a.values[a.valueIdx]
	a.valueIdx++

	// Reuse existing slices if they have capacity
	if cap(v.Value) >= 4 {
		v.Value = v.Value[:0]
	} else {
		v.Value = make([]any, 0, 4)
	}
	// Clear other fields
	v.Node = nil
	return v
}

// AllocDimension allocates a Dimension from the arena
func (a *NodeArena) AllocDimension() *Dimension {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.dimensionIdx >= len(a.dimensions) {
		newDimensions := make([]Dimension, len(a.dimensions)+a.initialCapacity/2)
		a.dimensions = append(a.dimensions, newDimensions...)
	}

	d := &a.dimensions[a.dimensionIdx]
	a.dimensionIdx++

	*d = Dimension{}
	return d
}

// AllocColor allocates a Color from the arena
func (a *NodeArena) AllocColor() *Color {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.colorIdx >= len(a.colors) {
		newColors := make([]Color, len(a.colors)+a.initialCapacity/2)
		a.colors = append(a.colors, newColors...)
	}

	c := &a.colors[a.colorIdx]
	a.colorIdx++

	*c = Color{}
	return c
}

// AllocKeyword allocates a Keyword from the arena
func (a *NodeArena) AllocKeyword() *Keyword {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.keywordIdx >= len(a.keywords) {
		newKeywords := make([]Keyword, len(a.keywords)+a.initialCapacity/2)
		a.keywords = append(a.keywords, newKeywords...)
	}

	k := &a.keywords[a.keywordIdx]
	a.keywordIdx++

	*k = Keyword{}
	return k
}

// AllocAnonymous allocates an Anonymous from the arena
func (a *NodeArena) AllocAnonymous() *Anonymous {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.anonymousIdx >= len(a.anonymouses) {
		newAnonymouses := make([]Anonymous, len(a.anonymouses)+a.initialCapacity/2)
		a.anonymouses = append(a.anonymouses, newAnonymouses...)
	}

	anon := &a.anonymouses[a.anonymousIdx]
	a.anonymousIdx++

	*anon = Anonymous{}
	return anon
}

// AllocQuoted allocates a Quoted from the arena
func (a *NodeArena) AllocQuoted() *Quoted {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.quotedIdx >= len(a.quoteds) {
		newQuoteds := make([]Quoted, len(a.quoteds)+a.initialCapacity/2)
		a.quoteds = append(a.quoteds, newQuoteds...)
	}

	q := &a.quoteds[a.quotedIdx]
	a.quotedIdx++

	*q = Quoted{}
	return q
}

// AllocUnit allocates a Unit from the arena
func (a *NodeArena) AllocUnit() *Unit {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.unitIdx >= len(a.units) {
		newUnits := make([]Unit, len(a.units)+a.initialCapacity/2)
		a.units = append(a.units, newUnits...)
	}

	u := &a.units[a.unitIdx]
	a.unitIdx++

	// Reuse existing slices if they have capacity
	if cap(u.Numerator) >= 4 {
		u.Numerator = u.Numerator[:0]
	} else {
		u.Numerator = make([]string, 0, 4)
	}
	if cap(u.Denominator) >= 2 {
		u.Denominator = u.Denominator[:0]
	} else {
		u.Denominator = make([]string, 0, 2)
	}
	// Clear other fields
	u.Node = nil
	u.BackupUnit = ""
	return u
}

// AllocCombinator allocates a Combinator from the arena
func (a *NodeArena) AllocCombinator() *Combinator {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.combinatorIdx >= len(a.combinators) {
		newCombinators := make([]Combinator, len(a.combinators)+a.initialCapacity)
		a.combinators = append(a.combinators, newCombinators...)
	}

	c := &a.combinators[a.combinatorIdx]
	a.combinatorIdx++

	*c = Combinator{}
	return c
}

// Global arena pool for reusing arenas between compilations
var arenaPool = sync.Pool{
	New: func() any {
		return NewNodeArena(DefaultArenaConfig())
	},
}

// GetArena gets an arena from the pool
func GetArena() *NodeArena {
	return arenaPool.Get().(*NodeArena)
}

// PutArena returns an arena to the pool after resetting it
func PutArena(a *NodeArena) {
	if a == nil {
		return
	}
	a.Reset()
	arenaPool.Put(a)
}

// CompilationContext wraps arena allocation for a single compilation
type CompilationContext struct {
	Arena *NodeArena
}

// NewCompilationContext creates a new compilation context with an arena
func NewCompilationContext() *CompilationContext {
	return &CompilationContext{
		Arena: GetArena(),
	}
}

// Close returns the arena to the pool
func (cc *CompilationContext) Close() {
	if cc.Arena != nil {
		PutArena(cc.Arena)
		cc.Arena = nil
	}
}
