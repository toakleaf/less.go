package less_go

import (
	"testing"
)

func TestNodeArena_BasicAllocation(t *testing.T) {
	arena := NewNodeArena(DefaultArenaConfig())
	defer PutArena(arena)

	// Allocate various node types
	node := arena.AllocNode()
	if node == nil {
		t.Fatal("AllocNode returned nil")
	}

	ruleset := arena.AllocRuleset()
	if ruleset == nil {
		t.Fatal("AllocRuleset returned nil")
	}
	if ruleset.Selectors == nil {
		t.Fatal("Ruleset.Selectors should be initialized")
	}
	if ruleset.Rules == nil {
		t.Fatal("Ruleset.Rules should be initialized")
	}

	decl := arena.AllocDeclaration()
	if decl == nil {
		t.Fatal("AllocDeclaration returned nil")
	}

	selector := arena.AllocSelector()
	if selector == nil {
		t.Fatal("AllocSelector returned nil")
	}
	if selector.Elements == nil {
		t.Fatal("Selector.Elements should be initialized")
	}

	element := arena.AllocElement()
	if element == nil {
		t.Fatal("AllocElement returned nil")
	}

	expr := arena.AllocExpression()
	if expr == nil {
		t.Fatal("AllocExpression returned nil")
	}
	if expr.Value == nil {
		t.Fatal("Expression.Value should be initialized")
	}

	// Check stats
	stats := arena.Stats()
	if stats.NodeAllocs != 1 {
		t.Errorf("Expected 1 node alloc, got %d", stats.NodeAllocs)
	}
	if stats.RulesetAllocs != 1 {
		t.Errorf("Expected 1 ruleset alloc, got %d", stats.RulesetAllocs)
	}
	if stats.TotalAllocs != 6 {
		t.Errorf("Expected 6 total allocs, got %d", stats.TotalAllocs)
	}
}

func TestNodeArena_Reset(t *testing.T) {
	arena := NewNodeArena(DefaultArenaConfig())
	defer PutArena(arena)

	// Allocate some nodes
	for i := 0; i < 100; i++ {
		arena.AllocNode()
		arena.AllocRuleset()
	}

	stats := arena.Stats()
	if stats.NodeAllocs != 100 {
		t.Errorf("Expected 100 node allocs, got %d", stats.NodeAllocs)
	}

	// Reset the arena
	arena.Reset()

	stats = arena.Stats()
	if stats.NodeAllocs != 0 {
		t.Errorf("Expected 0 node allocs after reset, got %d", stats.NodeAllocs)
	}
	if stats.TotalAllocs != 0 {
		t.Errorf("Expected 0 total allocs after reset, got %d", stats.TotalAllocs)
	}
}

func TestNodeArena_Growth(t *testing.T) {
	// Create arena with small initial capacity
	config := &ArenaConfig{
		Nodes:        10,
		Rulesets:     10,
		Declarations: 10,
		Selectors:    10,
		Elements:     10,
		Expressions:  10,
		Values:       10,
		Dimensions:   10,
		Colors:       10,
		Keywords:     10,
		Anonymouses:  10,
		Quoteds:      10,
		Units:        10,
		Combinators:  10,
	}
	arena := NewNodeArena(config)
	defer PutArena(arena)

	// Allocate more than initial capacity
	for i := 0; i < 50; i++ {
		node := arena.AllocNode()
		if node == nil {
			t.Fatalf("AllocNode returned nil at iteration %d", i)
		}
	}

	stats := arena.Stats()
	if stats.NodeAllocs != 50 {
		t.Errorf("Expected 50 node allocs, got %d", stats.NodeAllocs)
	}
}

func TestArenaPool(t *testing.T) {
	// Get arena from pool
	arena1 := GetArena()
	if arena1 == nil {
		t.Fatal("GetArena returned nil")
	}

	// Allocate some nodes
	arena1.AllocNode()
	arena1.AllocRuleset()

	stats := arena1.Stats()
	if stats.TotalAllocs != 2 {
		t.Errorf("Expected 2 allocs, got %d", stats.TotalAllocs)
	}

	// Return to pool
	PutArena(arena1)

	// Get another arena - should be reset
	arena2 := GetArena()
	stats = arena2.Stats()
	if stats.TotalAllocs != 0 {
		t.Errorf("Expected 0 allocs after reuse from pool, got %d", stats.TotalAllocs)
	}
	PutArena(arena2)
}

func TestCompilationContext(t *testing.T) {
	ctx := NewCompilationContext()
	if ctx.Arena == nil {
		t.Fatal("CompilationContext.Arena should not be nil")
	}

	// Use the arena
	ctx.Arena.AllocNode()
	ctx.Arena.AllocRuleset()

	stats := ctx.Arena.Stats()
	if stats.TotalAllocs != 2 {
		t.Errorf("Expected 2 allocs, got %d", stats.TotalAllocs)
	}

	// Close returns arena to pool
	ctx.Close()
	if ctx.Arena != nil {
		t.Error("Arena should be nil after Close")
	}
}

func TestArenaAwareAllocation(t *testing.T) {
	arena := NewNodeArena(DefaultArenaConfig())
	defer PutArena(arena)

	// Test arena-aware functions
	node := GetNodeFromArena(arena)
	if node == nil {
		t.Fatal("GetNodeFromArena returned nil")
	}

	ruleset := GetRulesetFromArena(arena)
	if ruleset == nil {
		t.Fatal("GetRulesetFromArena returned nil")
	}

	decl := GetDeclarationFromArena(arena)
	if decl == nil {
		t.Fatal("GetDeclarationFromArena returned nil")
	}

	selector := GetSelectorFromArena(arena)
	if selector == nil {
		t.Fatal("GetSelectorFromArena returned nil")
	}

	element := GetElementFromArena(arena)
	if element == nil {
		t.Fatal("GetElementFromArena returned nil")
	}

	expr := GetExpressionFromArena(arena)
	if expr == nil {
		t.Fatal("GetExpressionFromArena returned nil")
	}

	unit := GetUnitFromArena(arena)
	if unit == nil {
		t.Fatal("GetUnitFromArena returned nil")
	}

	stats := arena.Stats()
	if stats.TotalAllocs != 7 {
		t.Errorf("Expected 7 allocs, got %d", stats.TotalAllocs)
	}

	// Test with nil arena (should use pool)
	nodeFromPool := GetNodeFromArena(nil)
	if nodeFromPool == nil {
		t.Fatal("GetNodeFromArena(nil) returned nil")
	}
	ReleaseNode(nodeFromPool)
}

// Benchmark arena allocation vs pool allocation
func BenchmarkArenaAlloc_Node(b *testing.B) {
	arena := NewNodeArena(LargeArenaConfig())
	defer PutArena(arena)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.AllocNode()
		if i%1000 == 0 {
			arena.Reset()
		}
	}
}

func BenchmarkPoolAlloc_Node(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := GetNodeFromPool()
		ReleaseNode(n)
	}
}

func BenchmarkArenaAlloc_Ruleset(b *testing.B) {
	arena := NewNodeArena(LargeArenaConfig())
	defer PutArena(arena)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.AllocRuleset()
		if i%500 == 0 {
			arena.Reset()
		}
	}
}

func BenchmarkPoolAlloc_Ruleset(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := GetRulesetFromPool()
		ReleaseRuleset(r)
	}
}

func BenchmarkArenaAlloc_Mixed(b *testing.B) {
	arena := NewNodeArena(LargeArenaConfig())
	defer PutArena(arena)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arena.AllocNode()
		arena.AllocRuleset()
		arena.AllocDeclaration()
		arena.AllocSelector()
		arena.AllocElement()
		arena.AllocExpression()
		if i%200 == 0 {
			arena.Reset()
		}
	}
}

func BenchmarkPoolAlloc_Mixed(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n := GetNodeFromPool()
		r := GetRulesetFromPool()
		d := GetDeclarationFromPool()
		s := GetSelectorFromPool()
		e := GetElementFromPool()
		ex := GetExpressionFromPool()
		ReleaseNode(n)
		ReleaseRuleset(r)
		ReleaseDeclaration(d)
		ReleaseSelector(s)
		ReleaseElement(e)
		ReleaseExpression(ex)
	}
}

// BenchmarkCompilation benchmarks a simple compilation with arena allocation
func BenchmarkArena_SimpleCompilation(b *testing.B) {
	lessCode := `
	@color: #333;
	.container {
		color: @color;
		padding: 10px;
		.inner {
			margin: 5px;
		}
	}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Compile(lessCode, nil)
		if err != nil {
			b.Fatalf("Compilation failed: %v", err)
		}
	}
}
