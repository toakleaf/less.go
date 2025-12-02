package runtime

import (
	"testing"
)

// TestVisitorInfo tests VisitorInfo structure
func TestVisitorInfo(t *testing.T) {
	info := VisitorInfo{
		Index:            0,
		IsPreEvalVisitor: true,
		IsReplacing:      false,
	}

	if info.Index != 0 {
		t.Errorf("Index = %d, want 0", info.Index)
	}
	if !info.IsPreEvalVisitor {
		t.Error("IsPreEvalVisitor should be true")
	}
	if info.IsReplacing {
		t.Error("IsReplacing should be false")
	}
}

// TestVisitorResult tests VisitorResult structure
func TestVisitorResult(t *testing.T) {
	result := VisitorResult{
		Success:      true,
		VisitorCount: 2,
		Replacements: []VisitorReplacementSet{
			{
				VisitorIndex: 0,
				Replacements: []NodeReplacement{
					{
						ParentIndex: 1,
						ChildIndex:  0,
						Replacement: map[string]interface{}{"_type": "Dimension", "value": 10},
					},
				},
			},
		},
	}

	if !result.Success {
		t.Error("Success should be true")
	}
	if result.VisitorCount != 2 {
		t.Errorf("VisitorCount = %d, want 2", result.VisitorCount)
	}
	if len(result.Replacements) != 1 {
		t.Errorf("Replacements length = %d, want 1", len(result.Replacements))
	}
}

// TestNewJSVisitor tests JSVisitor creation
func TestNewJSVisitor(t *testing.T) {
	info := VisitorInfo{
		Index:            1,
		IsPreEvalVisitor: true,
		IsReplacing:      true,
	}

	visitor := NewJSVisitor(nil, info)

	if visitor.Index != 1 {
		t.Errorf("Index = %d, want 1", visitor.Index)
	}
	if !visitor.IsPreEvalVisitor {
		t.Error("IsPreEvalVisitor should be true")
	}
	if !visitor.IsReplacing {
		t.Error("IsReplacing should be true")
	}
}

// TestVisitorManager tests VisitorManager
func TestVisitorManager(t *testing.T) {
	vm := NewVisitorManager(nil)

	if vm == nil {
		t.Fatal("NewVisitorManager returned nil")
	}

	if len(vm.visitors) != 0 {
		t.Errorf("Initial visitors length = %d, want 0", len(vm.visitors))
	}
}

// TestVisitorManagerCategorization tests visitor categorization
func TestVisitorManagerCategorization(t *testing.T) {
	vm := NewVisitorManager(nil)

	// Add mock visitors directly
	preEval := &JSVisitor{Index: 0, IsPreEvalVisitor: true}
	postEval1 := &JSVisitor{Index: 1, IsPreEvalVisitor: false}
	postEval2 := &JSVisitor{Index: 2, IsPreEvalVisitor: false}

	vm.visitors = []*JSVisitor{preEval, postEval1, postEval2}

	preVisitors := vm.GetPreEvalVisitors()
	postVisitors := vm.GetPostEvalVisitors()

	if len(preVisitors) != 1 {
		t.Errorf("PreEval visitors = %d, want 1", len(preVisitors))
	}
	if len(postVisitors) != 2 {
		t.Errorf("PostEval visitors = %d, want 2", len(postVisitors))
	}
}

// TestParseReplacements tests replacement parsing
func TestParseReplacements(t *testing.T) {
	data := []interface{}{
		map[string]interface{}{
			"visitorIndex": float64(0),
			"replacements": []interface{}{
				map[string]interface{}{
					"parentIndex": float64(1),
					"childIndex":  float64(2),
					"replacement": map[string]interface{}{
						"_type": "Dimension",
						"value": float64(10),
					},
				},
			},
		},
	}

	result := parseReplacements(data)

	if len(result) != 1 {
		t.Fatalf("Result length = %d, want 1", len(result))
	}

	set := result[0]
	if set.VisitorIndex != 0 {
		t.Errorf("VisitorIndex = %d, want 0", set.VisitorIndex)
	}
	if len(set.Replacements) != 1 {
		t.Fatalf("Replacements length = %d, want 1", len(set.Replacements))
	}

	replacement := set.Replacements[0]
	if replacement.ParentIndex != 1 {
		t.Errorf("ParentIndex = %d, want 1", replacement.ParentIndex)
	}
	if replacement.ChildIndex != 2 {
		t.Errorf("ChildIndex = %d, want 2", replacement.ChildIndex)
	}
}

// TestParseVisitorResult tests result parsing
func TestParseVisitorResult(t *testing.T) {
	data := map[string]interface{}{
		"success":      true,
		"visitorCount": float64(3),
		"replacements": []interface{}{},
	}

	result, err := parseVisitorResult(data)
	if err != nil {
		t.Fatalf("parseVisitorResult failed: %v", err)
	}

	if !result.Success {
		t.Error("Success should be true")
	}
	if result.VisitorCount != 3 {
		t.Errorf("VisitorCount = %d, want 3", result.VisitorCount)
	}
}

// TestJSVisitorVisitWithoutRuntime tests error handling
func TestJSVisitorVisitWithoutRuntime(t *testing.T) {
	visitor := &JSVisitor{
		Index:            0,
		IsPreEvalVisitor: false,
		runtime:          nil,
	}

	_, err := visitor.Visit(nil)
	if err == nil {
		t.Error("Expected error when runtime is nil")
	}
}

// TestVisitorManagerRunPreEvalWithoutRuntime tests error handling
func TestVisitorManagerRunPreEvalWithoutRuntime(t *testing.T) {
	vm := NewVisitorManager(nil)

	_, err := vm.RunPreEvalVisitors(nil)
	if err == nil {
		t.Error("Expected error when runtime is nil")
	}
}

// TestVisitorManagerRunPostEvalWithoutRuntime tests error handling
func TestVisitorManagerRunPostEvalWithoutRuntime(t *testing.T) {
	vm := NewVisitorManager(nil)

	_, err := vm.RunPostEvalVisitors(nil)
	if err == nil {
		t.Error("Expected error when runtime is nil")
	}
}

// Benchmark tests

func BenchmarkParseReplacements(b *testing.B) {
	data := []interface{}{
		map[string]interface{}{
			"visitorIndex": float64(0),
			"replacements": []interface{}{
				map[string]interface{}{
					"parentIndex": float64(1),
					"childIndex":  float64(0),
					"replacement": map[string]interface{}{"_type": "Dimension"},
				},
				map[string]interface{}{
					"parentIndex": float64(2),
					"childIndex":  float64(1),
					"replacement": map[string]interface{}{"_type": "Color"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseReplacements(data)
	}
}

func BenchmarkParseVisitorResult(b *testing.B) {
	data := map[string]interface{}{
		"success":      true,
		"visitorCount": float64(5),
		"replacements": []interface{}{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseVisitorResult(data)
	}
}
