package runtime

import (
	"encoding/json"
	"testing"
)

// TestNode is a simple mock AST node for testing.
type TestNode struct {
	nodeType   string
	Value      string
	Number     float64
	Flag       bool
	Children   []*TestNode
	Child      *TestNode
	Parens     bool
	ParensInOp bool
}

func (t *TestNode) GetType() string {
	return t.nodeType
}

func TestFlatAST_AddString(t *testing.T) {
	flat := NewFlatAST()

	// Add strings
	idx1 := flat.AddString("hello")
	idx2 := flat.AddString("world")
	idx3 := flat.AddString("hello") // duplicate

	if idx1 != 0 {
		t.Errorf("first string index = %d, want 0", idx1)
	}
	if idx2 != 1 {
		t.Errorf("second string index = %d, want 1", idx2)
	}
	if idx3 != 0 {
		t.Errorf("duplicate string index = %d, want 0 (deduplicated)", idx3)
	}

	// Retrieve strings
	if s := flat.GetString(0); s != "hello" {
		t.Errorf("GetString(0) = %q, want %q", s, "hello")
	}
	if s := flat.GetString(1); s != "world" {
		t.Errorf("GetString(1) = %q, want %q", s, "world")
	}
}

func TestFlatAST_AddNode(t *testing.T) {
	flat := NewFlatAST()

	node1 := FlatNode{TypeID: TypeDimension, Flags: FlagParens}
	node2 := FlatNode{TypeID: TypeQuoted}

	idx1 := flat.AddNode(node1)
	idx2 := flat.AddNode(node2)

	if idx1 != 0 {
		t.Errorf("first node index = %d, want 0", idx1)
	}
	if idx2 != 1 {
		t.Errorf("second node index = %d, want 1", idx2)
	}
	if flat.NodeCount != 2 {
		t.Errorf("NodeCount = %d, want 2", flat.NodeCount)
	}
}

func TestFlatAST_Properties(t *testing.T) {
	flat := NewFlatAST()

	props := map[string]any{
		"value": "test",
		"num":   42.5,
		"flag":  true,
	}

	offset, length := flat.AddProperties(props)

	if offset != 0 {
		t.Errorf("props offset = %d, want 0", offset)
	}
	if length == 0 {
		t.Error("props length should not be 0")
	}

	// Retrieve properties
	retrieved := flat.GetProperties(offset, length)
	if retrieved == nil {
		t.Fatal("GetProperties returned nil")
	}

	if retrieved["value"] != "test" {
		t.Errorf("props['value'] = %v, want 'test'", retrieved["value"])
	}
	if retrieved["num"] != 42.5 {
		t.Errorf("props['num'] = %v, want 42.5", retrieved["num"])
	}
	if retrieved["flag"] != true {
		t.Errorf("props['flag'] = %v, want true", retrieved["flag"])
	}
}

func TestFlatAST_ToBytes_FromBytes(t *testing.T) {
	// Create a FlatAST with some data
	flat := NewFlatAST()
	flat.AddString("hello")
	flat.AddString("world")
	flat.TypeTable = append(flat.TypeTable, "TestType1", "TestType2")

	props := map[string]any{"key": "value"}
	offset, length := flat.AddProperties(props)

	flat.AddNode(FlatNode{
		TypeID:      TypeDimension,
		Flags:       FlagParens | FlagHasIndex,
		PropsOffset: offset,
		PropsLength: length,
	})
	flat.AddNode(FlatNode{
		TypeID:      TypeQuoted,
		ParentIndex: 0,
		ChildIndex:  0,
	})

	flat.RootIndex = 0

	// Serialize to bytes
	data, err := flat.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// Deserialize from bytes
	restored, err := FromBytes(data)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Verify data
	if restored.Version != flat.Version {
		t.Errorf("Version = %d, want %d", restored.Version, flat.Version)
	}
	if restored.NodeCount != flat.NodeCount {
		t.Errorf("NodeCount = %d, want %d", restored.NodeCount, flat.NodeCount)
	}
	if restored.RootIndex != flat.RootIndex {
		t.Errorf("RootIndex = %d, want %d", restored.RootIndex, flat.RootIndex)
	}

	// Verify string table
	if len(restored.StringTable) != len(flat.StringTable) {
		t.Errorf("StringTable length = %d, want %d", len(restored.StringTable), len(flat.StringTable))
	}
	for i, s := range flat.StringTable {
		if restored.StringTable[i] != s {
			t.Errorf("StringTable[%d] = %q, want %q", i, restored.StringTable[i], s)
		}
	}

	// Verify nodes
	if len(restored.Nodes) != len(flat.Nodes) {
		t.Errorf("Nodes length = %d, want %d", len(restored.Nodes), len(flat.Nodes))
	}
	for i, node := range flat.Nodes {
		rNode := restored.Nodes[i]
		if rNode.TypeID != node.TypeID {
			t.Errorf("Node[%d].TypeID = %d, want %d", i, rNode.TypeID, node.TypeID)
		}
		if rNode.Flags != node.Flags {
			t.Errorf("Node[%d].Flags = %d, want %d", i, rNode.Flags, node.Flags)
		}
	}

	// Verify properties can be retrieved
	restoredProps := restored.GetProperties(restored.Nodes[0].PropsOffset, restored.Nodes[0].PropsLength)
	if restoredProps == nil {
		t.Error("Could not retrieve properties from restored FlatAST")
	} else if restoredProps["key"] != "value" {
		t.Errorf("Restored props['key'] = %v, want 'value'", restoredProps["key"])
	}
}

func TestGetTypeID(t *testing.T) {
	tests := []struct {
		name     string
		node     any
		expected NodeTypeID
	}{
		{"Dimension", &TestNode{nodeType: "Dimension"}, TypeDimension},
		{"Quoted", &TestNode{nodeType: "Quoted"}, TypeQuoted},
		{"Color", &TestNode{nodeType: "Color"}, TypeColor},
		{"Ruleset", &TestNode{nodeType: "Ruleset"}, TypeRuleset},
		{"Unknown", &TestNode{nodeType: "NonExistent"}, TypeUnknown},
		{"Nil", nil, TypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTypeID(tt.node)
			if got != tt.expected {
				t.Errorf("GetTypeID() = %d (%s), want %d (%s)",
					got, TypeNames[got], tt.expected, TypeNames[tt.expected])
			}
		})
	}
}

func TestASTFlattener_SimpleNode(t *testing.T) {
	node := &TestNode{
		nodeType: "Dimension",
		Value:    "10px",
		Number:   10.0,
		Flag:     true,
	}

	flat, err := FlattenAST(node)
	if err != nil {
		t.Fatalf("FlattenAST failed: %v", err)
	}

	if flat.NodeCount != 1 {
		t.Errorf("NodeCount = %d, want 1", flat.NodeCount)
	}

	if flat.Nodes[0].TypeID != TypeDimension {
		t.Errorf("TypeID = %d, want %d (Dimension)", flat.Nodes[0].TypeID, TypeDimension)
	}
}

func TestASTFlattener_WithChildren(t *testing.T) {
	child1 := &TestNode{nodeType: "Keyword", Value: "auto"}
	child2 := &TestNode{nodeType: "Dimension", Number: 100}

	parent := &TestNode{
		nodeType: "Value",
		Children: []*TestNode{child1, child2},
	}

	flat, err := FlattenAST(parent)
	if err != nil {
		t.Fatalf("FlattenAST failed: %v", err)
	}

	// Should have 3 nodes: parent + 2 children
	if flat.NodeCount != 3 {
		t.Errorf("NodeCount = %d, want 3", flat.NodeCount)
	}

	// Verify parent has correct child index
	parentNode := flat.Nodes[flat.RootIndex]
	if parentNode.ChildIndex == 0 {
		t.Error("Parent should have children")
	}

	// Verify children are linked
	firstChild := flat.Nodes[parentNode.ChildIndex]
	if firstChild.NextIndex == 0 {
		t.Error("First child should have a sibling")
	}
}

func TestASTFlattener_NestedChildren(t *testing.T) {
	grandchild := &TestNode{nodeType: "Dimension", Number: 50}
	child := &TestNode{
		nodeType: "Expression",
		Child:    grandchild,
	}
	parent := &TestNode{
		nodeType: "Declaration",
		Child:    child,
	}

	flat, err := FlattenAST(parent)
	if err != nil {
		t.Fatalf("FlattenAST failed: %v", err)
	}

	// Should have 3 nodes
	if flat.NodeCount != 3 {
		t.Errorf("NodeCount = %d, want 3", flat.NodeCount)
	}

	// Verify the chain: Declaration -> Expression -> Dimension
	parentNode := flat.Nodes[flat.RootIndex]
	if parentNode.TypeID != TypeDeclaration {
		t.Errorf("Root type = %d, want Declaration (%d)", parentNode.TypeID, TypeDeclaration)
	}
}

func TestUnflattenAST_Simple(t *testing.T) {
	// Create a simple flat AST
	flat := NewFlatAST()
	flat.AddNode(FlatNode{
		TypeID: TypeDimension,
		Flags:  FlagParens,
	})
	flat.RootIndex = 0

	// Unflatten
	root, err := UnflattenAST(flat)
	if err != nil {
		t.Fatalf("UnflattenAST failed: %v", err)
	}

	if root.Type != "Dimension" {
		t.Errorf("Type = %q, want 'Dimension'", root.Type)
	}
	if !root.Parens {
		t.Error("Parens should be true")
	}
}

func TestUnflattenAST_WithChildren(t *testing.T) {
	// Create a flat AST with parent-child relationship
	flat := NewFlatAST()

	// Add parent first (index 0)
	flat.AddNode(FlatNode{
		TypeID:     TypeRuleset,
		ChildIndex: 1, // First child at index 1
	})

	// Add first child (index 1)
	flat.AddNode(FlatNode{
		TypeID:      TypeDeclaration,
		ParentIndex: 0,
		NextIndex:   2, // Next sibling at index 2
	})

	// Add second child (index 2)
	flat.AddNode(FlatNode{
		TypeID:      TypeDeclaration,
		ParentIndex: 0,
	})

	flat.RootIndex = 0

	// Unflatten
	root, err := UnflattenAST(flat)
	if err != nil {
		t.Fatalf("UnflattenAST failed: %v", err)
	}

	if root.Type != "Ruleset" {
		t.Errorf("Root type = %q, want 'Ruleset'", root.Type)
	}

	if len(root.Children) != 2 {
		t.Errorf("Children count = %d, want 2", len(root.Children))
	}

	for i, child := range root.Children {
		if child.Type != "Declaration" {
			t.Errorf("Child[%d] type = %q, want 'Declaration'", i, child.Type)
		}
		// Parent should be set (but pointer equality may differ due to reconstruction)
		if child.Parent == nil {
			t.Errorf("Child[%d] parent is nil", i)
		} else if child.Parent.Type != "Ruleset" {
			t.Errorf("Child[%d] parent type = %q, want 'Ruleset'", i, child.Parent.Type)
		}
	}
}

func TestRoundtrip_FlattenUnflatten(t *testing.T) {
	// Create original AST
	child1 := &TestNode{nodeType: "Keyword", Value: "inherit"}
	child2 := &TestNode{nodeType: "Dimension", Number: 100}

	parent := &TestNode{
		nodeType: "Value",
		Children: []*TestNode{child1, child2},
		Parens:   true,
	}

	// Flatten
	flat, err := FlattenAST(parent)
	if err != nil {
		t.Fatalf("FlattenAST failed: %v", err)
	}

	// Serialize to bytes
	data, err := flat.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// Deserialize from bytes
	restored, err := FromBytes(data)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Unflatten
	root, err := UnflattenAST(restored)
	if err != nil {
		t.Fatalf("UnflattenAST failed: %v", err)
	}

	// Verify structure
	if root.Type != "Value" {
		t.Errorf("Root type = %q, want 'Value'", root.Type)
	}
	if !root.Parens {
		t.Error("Root parens should be true")
	}
	if len(root.Children) != 2 {
		t.Errorf("Children count = %d, want 2", len(root.Children))
	}

	// Verify child types
	if root.Children[0].Type != "Keyword" {
		t.Errorf("First child type = %q, want 'Keyword'", root.Children[0].Type)
	}
	if root.Children[1].Type != "Dimension" {
		t.Errorf("Second child type = %q, want 'Dimension'", root.Children[1].Type)
	}
}

func TestGenericNode_ToJSON(t *testing.T) {
	child := &GenericNode{
		Type: "Dimension",
		Properties: map[string]any{
			"value": 10.0,
		},
	}

	root := &GenericNode{
		Type:     "Value",
		Children: []*GenericNode{child},
		Parens:   true,
	}
	child.Parent = root

	jsonMap := root.ToJSON()

	if jsonMap["type"] != "Value" {
		t.Errorf("type = %v, want 'Value'", jsonMap["type"])
	}
	if jsonMap["parens"] != true {
		t.Errorf("parens = %v, want true", jsonMap["parens"])
	}

	children, ok := jsonMap["children"].([]map[string]any)
	if !ok {
		t.Fatalf("children is not []map[string]any")
	}
	if len(children) != 1 {
		t.Errorf("children length = %d, want 1", len(children))
	}
	if children[0]["type"] != "Dimension" {
		t.Errorf("child type = %v, want 'Dimension'", children[0]["type"])
	}

	// Verify JSON serialization
	data, err := json.Marshal(jsonMap)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("JSON output should not be empty")
	}
}

func TestTypeNames(t *testing.T) {
	// Verify all TypeNames have corresponding TypeNameToID entries
	for id, name := range TypeNames {
		if TypeNameToID[name] != id {
			t.Errorf("TypeNameToID[%q] = %d, want %d", name, TypeNameToID[name], id)
		}
	}

	// Verify bidirectional mapping
	for name, id := range TypeNameToID {
		if TypeNames[id] != name {
			t.Errorf("TypeNames[%d] = %q, want %q", id, TypeNames[id], name)
		}
	}
}

// Benchmark tests

func BenchmarkFlattenAST_SimpleNode(b *testing.B) {
	node := &TestNode{
		nodeType: "Dimension",
		Value:    "10px",
		Number:   10.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FlattenAST(node)
	}
}

func BenchmarkFlattenAST_WithChildren(b *testing.B) {
	children := make([]*TestNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = &TestNode{
			nodeType: "Dimension",
			Number:   float64(i),
		}
	}
	parent := &TestNode{
		nodeType: "Value",
		Children: children,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FlattenAST(parent)
	}
}

func BenchmarkToBytes(b *testing.B) {
	flat := NewFlatAST()
	for i := 0; i < 100; i++ {
		flat.AddString("test string " + string(rune(i)))
		flat.AddNode(FlatNode{
			TypeID: TypeDimension,
			Flags:  FlagParens,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = flat.ToBytes()
	}
}

func BenchmarkFromBytes(b *testing.B) {
	flat := NewFlatAST()
	for i := 0; i < 100; i++ {
		flat.AddString("test string " + string(rune(i)))
		flat.AddNode(FlatNode{
			TypeID: TypeDimension,
			Flags:  FlagParens,
		})
	}
	data, _ := flat.ToBytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FromBytes(data)
	}
}

func BenchmarkRoundtrip(b *testing.B) {
	children := make([]*TestNode, 10)
	for i := 0; i < 10; i++ {
		children[i] = &TestNode{
			nodeType: "Dimension",
			Number:   float64(i),
		}
	}
	parent := &TestNode{
		nodeType: "Value",
		Children: children,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flat, _ := FlattenAST(parent)
		data, _ := flat.ToBytes()
		restored, _ := FromBytes(data)
		_, _ = UnflattenAST(restored)
	}
}
