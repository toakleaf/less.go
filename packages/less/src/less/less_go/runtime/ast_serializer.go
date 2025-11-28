package runtime

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
)

// NodeTypeID represents a unique identifier for each AST node type.
type NodeTypeID uint16

// Node type constants - these map to JavaScript type names.
const (
	TypeUnknown NodeTypeID = iota
	TypeAnonymous
	TypeAssignment
	TypeAtRule
	TypeAttribute
	TypeCall
	TypeColor
	TypeCombinator
	TypeComment
	TypeCondition
	TypeContainer
	TypeDeclaration
	TypeDetachedRuleset
	TypeDimension
	TypeElement
	TypeExpression
	TypeExtend
	TypeImport
	TypeJavaScript
	TypeKeyword
	TypeMedia
	TypeMixinCall
	TypeMixinDefinition
	TypeNamespaceValue
	TypeNegative
	TypeOperation
	TypeParen
	TypeProperty
	TypeQueryInParens
	TypeQuoted
	TypeRuleset
	TypeSelector
	TypeSelectorList
	TypeUnicodeDescriptor
	TypeUnit
	TypeURL
	TypeValue
	TypeVariable
	TypeVariableCall
	TypeNode // Base Node type
	// Add more as needed
)

// TypeNames maps type IDs to their string names (for JavaScript compatibility).
var TypeNames = map[NodeTypeID]string{
	TypeUnknown:          "Unknown",
	TypeAnonymous:        "Anonymous",
	TypeAssignment:       "Assignment",
	TypeAtRule:           "AtRule",
	TypeAttribute:        "Attribute",
	TypeCall:             "Call",
	TypeColor:            "Color",
	TypeCombinator:       "Combinator",
	TypeComment:          "Comment",
	TypeCondition:        "Condition",
	TypeContainer:        "Container",
	TypeDeclaration:      "Declaration",
	TypeDetachedRuleset:  "DetachedRuleset",
	TypeDimension:        "Dimension",
	TypeElement:          "Element",
	TypeExpression:       "Expression",
	TypeExtend:           "Extend",
	TypeImport:           "Import",
	TypeJavaScript:       "JavaScript",
	TypeKeyword:          "Keyword",
	TypeMedia:            "Media",
	TypeMixinCall:        "MixinCall",
	TypeMixinDefinition:  "MixinDefinition",
	TypeNamespaceValue:   "NamespaceValue",
	TypeNegative:         "Negative",
	TypeOperation:        "Operation",
	TypeParen:            "Paren",
	TypeProperty:         "Property",
	TypeQueryInParens:    "QueryInParens",
	TypeQuoted:           "Quoted",
	TypeRuleset:          "Ruleset",
	TypeSelector:         "Selector",
	TypeSelectorList:     "SelectorList",
	TypeUnicodeDescriptor: "UnicodeDescriptor",
	TypeUnit:             "Unit",
	TypeURL:              "URL",
	TypeValue:            "Value",
	TypeVariable:         "Variable",
	TypeVariableCall:     "VariableCall",
	TypeNode:             "Node",
}

// TypeNameToID maps string type names to their IDs.
var TypeNameToID = func() map[string]NodeTypeID {
	m := make(map[string]NodeTypeID)
	for id, name := range TypeNames {
		m[name] = id
	}
	return m
}()

// FlatNode represents a node in the flattened AST buffer.
// Each node is 24 bytes in the binary representation.
type FlatNode struct {
	TypeID      NodeTypeID // Index into type table (2 bytes)
	Flags       uint16     // Node flags (2 bytes)
	ChildIndex  uint32     // Index of first child in Nodes array (0 if none)
	NextIndex   uint32     // Index of next sibling (0 if none)
	ParentIndex uint32     // Index of parent node (0 if root)
	PropsOffset uint32     // Offset into properties buffer
	PropsLength uint32     // Length of properties in buffer
}

// FlatNode flags
const (
	FlagParens       uint16 = 1 << 0
	FlagParensInOp   uint16 = 1 << 1
	FlagVisible      uint16 = 1 << 2
	FlagInvisible    uint16 = 1 << 3
	FlagVisibleSet   uint16 = 1 << 4
	FlagHasFileInfo  uint16 = 1 << 5
	FlagHasIndex     uint16 = 1 << 6
)

// FlatAST represents a complete flattened AST structure.
type FlatAST struct {
	// Header information
	Version     uint32 // Format version
	NodeCount   uint32 // Number of nodes
	RootIndex   uint32 // Index of root node

	// Tables
	Nodes       []FlatNode        // Array of flat nodes
	TypeTable   []string          // Node type names (for validation)
	StringTable []string          // String values
	PropBuffer  []byte            // Node-specific properties (JSON encoded)

	// String deduplication
	stringIndex map[string]uint32 // Maps strings to their index in StringTable
}

// NewFlatAST creates a new empty FlatAST.
func NewFlatAST() *FlatAST {
	return &FlatAST{
		Version:     1,
		Nodes:       make([]FlatNode, 0),
		TypeTable:   make([]string, 0),
		StringTable: make([]string, 0),
		PropBuffer:  make([]byte, 0),
		stringIndex: make(map[string]uint32),
	}
}

// AddString adds a string to the string table and returns its index.
// Duplicate strings are deduplicated.
func (f *FlatAST) AddString(s string) uint32 {
	if idx, ok := f.stringIndex[s]; ok {
		return idx
	}
	idx := uint32(len(f.StringTable))
	f.StringTable = append(f.StringTable, s)
	f.stringIndex[s] = idx
	return idx
}

// GetString retrieves a string from the string table by index.
func (f *FlatAST) GetString(idx uint32) string {
	if int(idx) >= len(f.StringTable) {
		return ""
	}
	return f.StringTable[idx]
}

// AddNode adds a node to the AST and returns its index.
func (f *FlatAST) AddNode(node FlatNode) uint32 {
	idx := uint32(len(f.Nodes))
	f.Nodes = append(f.Nodes, node)
	f.NodeCount = uint32(len(f.Nodes))
	return idx
}

// AddProperties adds properties to the buffer and returns offset and length.
func (f *FlatAST) AddProperties(props map[string]any) (uint32, uint32) {
	if props == nil || len(props) == 0 {
		return 0, 0
	}

	data, err := json.Marshal(props)
	if err != nil {
		return 0, 0
	}

	offset := uint32(len(f.PropBuffer))
	f.PropBuffer = append(f.PropBuffer, data...)
	return offset, uint32(len(data))
}

// GetProperties retrieves properties from the buffer.
func (f *FlatAST) GetProperties(offset, length uint32) map[string]any {
	if length == 0 || int(offset+length) > len(f.PropBuffer) {
		return nil
	}

	data := f.PropBuffer[offset : offset+length]
	var props map[string]any
	if err := json.Unmarshal(data, &props); err != nil {
		return nil
	}
	return props
}

// Binary format constants
const (
	FlatASTMagic   uint32 = 0x4C455353 // "LESS"
	FlatASTVersion uint32 = 1
	FlatNodeSize   int    = 24 // bytes per FlatNode
)

// ToBytes serializes the FlatAST to a byte buffer.
func (f *FlatAST) ToBytes() ([]byte, error) {
	// Calculate total size
	headerSize := 4 + 4 + 4 + 4 + 4 + 4 + 4 // magic, version, nodeCount, rootIndex, offsets
	nodesSize := len(f.Nodes) * FlatNodeSize

	// Calculate string table size (length-prefixed strings)
	stringTableSize := 4 // count
	for _, s := range f.StringTable {
		stringTableSize += 4 + len(s) // length + data
	}

	// Calculate type table size
	typeTableSize := 4 // count
	for _, t := range f.TypeTable {
		typeTableSize += 4 + len(t)
	}

	propBufferSize := 4 + len(f.PropBuffer) // length + data

	totalSize := headerSize + nodesSize + stringTableSize + typeTableSize + propBufferSize
	buf := make([]byte, totalSize)
	offset := 0

	// Write header
	binary.LittleEndian.PutUint32(buf[offset:], FlatASTMagic)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], f.Version)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], f.NodeCount)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], f.RootIndex)
	offset += 4

	// Write offsets (we'll fill these in after writing sections)
	nodesOffset := uint32(headerSize)
	stringTableOffset := nodesOffset + uint32(nodesSize)
	typeTableOffset := stringTableOffset + uint32(stringTableSize)
	propBufferOffset := typeTableOffset + uint32(typeTableSize)

	binary.LittleEndian.PutUint32(buf[offset:], nodesOffset)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], stringTableOffset)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:], typeTableOffset)
	offset += 4

	// Write nodes
	offset = int(nodesOffset)
	for _, node := range f.Nodes {
		binary.LittleEndian.PutUint16(buf[offset:], uint16(node.TypeID))
		offset += 2
		binary.LittleEndian.PutUint16(buf[offset:], node.Flags)
		offset += 2
		binary.LittleEndian.PutUint32(buf[offset:], node.ChildIndex)
		offset += 4
		binary.LittleEndian.PutUint32(buf[offset:], node.NextIndex)
		offset += 4
		binary.LittleEndian.PutUint32(buf[offset:], node.ParentIndex)
		offset += 4
		binary.LittleEndian.PutUint32(buf[offset:], node.PropsOffset)
		offset += 4
		binary.LittleEndian.PutUint32(buf[offset:], node.PropsLength)
		offset += 4
	}

	// Write string table
	offset = int(stringTableOffset)
	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(f.StringTable)))
	offset += 4
	for _, s := range f.StringTable {
		binary.LittleEndian.PutUint32(buf[offset:], uint32(len(s)))
		offset += 4
		copy(buf[offset:], s)
		offset += len(s)
	}

	// Write type table
	offset = int(typeTableOffset)
	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(f.TypeTable)))
	offset += 4
	for _, t := range f.TypeTable {
		binary.LittleEndian.PutUint32(buf[offset:], uint32(len(t)))
		offset += 4
		copy(buf[offset:], t)
		offset += len(t)
	}

	// Write prop buffer
	offset = int(propBufferOffset)
	binary.LittleEndian.PutUint32(buf[offset:], uint32(len(f.PropBuffer)))
	offset += 4
	copy(buf[offset:], f.PropBuffer)

	return buf, nil
}

// FromBytes deserializes a byte buffer to a FlatAST.
func FromBytes(data []byte) (*FlatAST, error) {
	if len(data) < 28 { // Minimum header size
		return nil, fmt.Errorf("buffer too small")
	}

	offset := 0

	// Read and verify magic
	magic := binary.LittleEndian.Uint32(data[offset:])
	if magic != FlatASTMagic {
		return nil, fmt.Errorf("invalid magic: expected %x, got %x", FlatASTMagic, magic)
	}
	offset += 4

	f := &FlatAST{
		stringIndex: make(map[string]uint32),
	}

	// Read header
	f.Version = binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	f.NodeCount = binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	f.RootIndex = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	nodesOffset := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	stringTableOffset := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	typeTableOffset := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	// Read nodes
	offset = int(nodesOffset)
	f.Nodes = make([]FlatNode, f.NodeCount)
	for i := uint32(0); i < f.NodeCount; i++ {
		f.Nodes[i].TypeID = NodeTypeID(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
		f.Nodes[i].Flags = binary.LittleEndian.Uint16(data[offset:])
		offset += 2
		f.Nodes[i].ChildIndex = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		f.Nodes[i].NextIndex = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		f.Nodes[i].ParentIndex = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		f.Nodes[i].PropsOffset = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		f.Nodes[i].PropsLength = binary.LittleEndian.Uint32(data[offset:])
		offset += 4
	}

	// Read string table
	offset = int(stringTableOffset)
	stringCount := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	f.StringTable = make([]string, stringCount)
	for i := uint32(0); i < stringCount; i++ {
		strLen := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		f.StringTable[i] = string(data[offset : offset+int(strLen)])
		f.stringIndex[f.StringTable[i]] = i
		offset += int(strLen)
	}

	// Read type table
	offset = int(typeTableOffset)
	typeCount := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	f.TypeTable = make([]string, typeCount)
	for i := uint32(0); i < typeCount; i++ {
		strLen := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		f.TypeTable[i] = string(data[offset : offset+int(strLen)])
		offset += int(strLen)
	}

	// Read prop buffer
	propLen := binary.LittleEndian.Uint32(data[offset:])
	offset += 4
	f.PropBuffer = make([]byte, propLen)
	copy(f.PropBuffer, data[offset:offset+int(propLen)])

	return f, nil
}

// ASTFlattener converts Go AST nodes to FlatAST format.
type ASTFlattener struct {
	flat *FlatAST
}

// NewASTFlattener creates a new AST flattener.
func NewASTFlattener() *ASTFlattener {
	return &ASTFlattener{
		flat: NewFlatAST(),
	}
}

// GetTypeID returns the type ID for a node based on its GetType() method.
func GetTypeID(node any) NodeTypeID {
	if node == nil {
		return TypeUnknown
	}

	// Try GetType() method first
	if typer, ok := node.(interface{ GetType() string }); ok {
		typeName := typer.GetType()
		if id, ok := TypeNameToID[typeName]; ok {
			return id
		}
	}

	// Fall back to Type() method
	if typer, ok := node.(interface{ Type() string }); ok {
		typeName := typer.Type()
		if id, ok := TypeNameToID[typeName]; ok {
			return id
		}
	}

	// Fall back to reflection-based type name
	t := reflect.TypeOf(node)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeName := t.Name()
	if id, ok := TypeNameToID[typeName]; ok {
		return id
	}

	return TypeUnknown
}

// FlattenNode flattens a single node and its children recursively.
// Returns the index of the flattened node.
func (af *ASTFlattener) FlattenNode(node any, parentIndex uint32) (uint32, error) {
	if node == nil {
		return 0, nil
	}

	typeID := GetTypeID(node)
	if typeID == TypeUnknown {
		return 0, fmt.Errorf("unknown node type: %T", node)
	}

	// Create the flat node
	flatNode := FlatNode{
		TypeID:      typeID,
		ParentIndex: parentIndex,
	}

	// Extract common Node properties if available
	if n, ok := getBaseNode(node); ok {
		if n.Parens {
			flatNode.Flags |= FlagParens
		}
		if n.ParensInOp {
			flatNode.Flags |= FlagParensInOp
		}
		if n.NodeVisible != nil {
			flatNode.Flags |= FlagVisibleSet
			if *n.NodeVisible {
				flatNode.Flags |= FlagVisible
			} else {
				flatNode.Flags |= FlagInvisible
			}
		}
		if n.Index != 0 {
			flatNode.Flags |= FlagHasIndex
		}
		if len(n.fileInfo) > 0 {
			flatNode.Flags |= FlagHasFileInfo
		}
	}

	// Extract type-specific properties
	props := af.extractProperties(node)
	if props != nil {
		flatNode.PropsOffset, flatNode.PropsLength = af.flat.AddProperties(props)
	}

	// Add the node first to get its index
	nodeIndex := af.flat.AddNode(flatNode)

	// Flatten children
	children := getChildren(node)
	var prevChildIndex uint32 = 0
	var firstChildIndex uint32 = 0

	for i, child := range children {
		if child == nil {
			continue
		}

		childIndex, err := af.FlattenNode(child, nodeIndex)
		if err != nil {
			return 0, err
		}

		if childIndex == 0 {
			continue
		}

		if i == 0 || firstChildIndex == 0 {
			firstChildIndex = childIndex
		}

		// Link siblings
		if prevChildIndex != 0 {
			af.flat.Nodes[prevChildIndex].NextIndex = childIndex
		}
		prevChildIndex = childIndex
	}

	// Update child index
	af.flat.Nodes[nodeIndex].ChildIndex = firstChildIndex

	return nodeIndex, nil
}

// Flatten flattens an entire AST starting from the root.
func (af *ASTFlattener) Flatten(root any) (*FlatAST, error) {
	rootIndex, err := af.FlattenNode(root, 0)
	if err != nil {
		return nil, err
	}

	af.flat.RootIndex = rootIndex
	return af.flat, nil
}

// FlattenAST is a convenience function to flatten an AST.
func FlattenAST(root any) (*FlatAST, error) {
	flattener := NewASTFlattener()
	return flattener.Flatten(root)
}

// extractProperties extracts type-specific properties from a node.
func (af *ASTFlattener) extractProperties(node any) map[string]any {
	props := make(map[string]any)

	// Use reflection to extract exported fields
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip embedded Node fields (handled separately)
		if field.Name == "Node" && field.Type.Kind() == reflect.Ptr {
			continue
		}

		// Skip nil values
		if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
			continue
		}

		// Handle specific field types
		switch fieldValue.Kind() {
		case reflect.String:
			if s := fieldValue.String(); s != "" {
				props[field.Name] = af.flat.AddString(s)
			}
		case reflect.Int, reflect.Int64:
			if n := fieldValue.Int(); n != 0 {
				props[field.Name] = n
			}
		case reflect.Float64:
			if f := fieldValue.Float(); f != 0 {
				props[field.Name] = f
			}
		case reflect.Bool:
			props[field.Name] = fieldValue.Bool()
		case reflect.Slice, reflect.Array:
			// Skip slice fields - children are handled separately
			continue
		case reflect.Ptr:
			// Skip pointer fields - usually child nodes
			continue
		case reflect.Interface:
			if !fieldValue.IsNil() {
				// Try to serialize as JSON
				if data, err := json.Marshal(fieldValue.Interface()); err == nil {
					props[field.Name] = string(data)
				}
			}
		}
	}

	if len(props) == 0 {
		return nil
	}
	return props
}

// BaseNode interface for nodes that embed *Node
type BaseNode interface {
	GetNode() *Node
}

// Node represents the base node structure (simplified for serialization)
type Node struct {
	Parent          *Node
	VisibilityBlocks *int
	NodeVisible     *bool
	RootNode        *Node
	Parsed          any
	Value           any
	Index           int
	fileInfo        map[string]any
	Parens          bool
	ParensInOp      bool
	TypeIndex       int
}

// getBaseNode extracts the embedded *Node from a node type.
func getBaseNode(node any) (*Node, bool) {
	// Try BaseNode interface first
	if bn, ok := node.(BaseNode); ok {
		if n := bn.GetNode(); n != nil {
			// Convert to our local Node type
			return &Node{
				Parens:      n.Parens,
				ParensInOp:  n.ParensInOp,
				NodeVisible: n.NodeVisible,
				Index:       n.Index,
			}, true
		}
	}

	// Use reflection as fallback
	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	n := &Node{}
	found := false

	// First try embedded Node field
	nodeField := v.FieldByName("Node")
	if nodeField.IsValid() && nodeField.Kind() == reflect.Ptr && !nodeField.IsNil() {
		// Try to get fields from the embedded node
		nodeVal := nodeField.Elem()

		if parens := nodeVal.FieldByName("Parens"); parens.IsValid() {
			n.Parens = parens.Bool()
			found = true
		}
		if parensInOp := nodeVal.FieldByName("ParensInOp"); parensInOp.IsValid() {
			n.ParensInOp = parensInOp.Bool()
			found = true
		}
		if index := nodeVal.FieldByName("Index"); index.IsValid() {
			n.Index = int(index.Int())
			found = true
		}
		if visible := nodeVal.FieldByName("NodeVisible"); visible.IsValid() && !visible.IsNil() {
			b := visible.Elem().Bool()
			n.NodeVisible = &b
			found = true
		}
	}

	// Also check for direct Parens/ParensInOp fields on the struct itself
	// (for test nodes and other node types that don't embed *Node)
	if parens := v.FieldByName("Parens"); parens.IsValid() && parens.Kind() == reflect.Bool {
		n.Parens = parens.Bool()
		found = true
	}
	if parensInOp := v.FieldByName("ParensInOp"); parensInOp.IsValid() && parensInOp.Kind() == reflect.Bool {
		n.ParensInOp = parensInOp.Bool()
		found = true
	}
	if index := v.FieldByName("Index"); index.IsValid() && index.Kind() == reflect.Int {
		n.Index = int(index.Int())
		found = true
	}

	if found {
		return n, true
	}

	return nil, false
}

// getChildren extracts child nodes from a parent node.
func getChildren(node any) []any {
	var children []any

	v := reflect.ValueOf(node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip the embedded Node field
		if field.Name == "Node" {
			continue
		}

		// Handle slices of nodes
		if fieldValue.Kind() == reflect.Slice {
			for j := 0; j < fieldValue.Len(); j++ {
				elem := fieldValue.Index(j)
				if elem.Kind() == reflect.Interface {
					elem = elem.Elem()
				}
				if isASTNode(elem) {
					children = append(children, elem.Interface())
				}
			}
		}

		// Handle pointer fields that are nodes
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			if isASTNode(fieldValue) {
				children = append(children, fieldValue.Interface())
			}
		}

		// Handle interface fields that might contain nodes
		if fieldValue.Kind() == reflect.Interface && !fieldValue.IsNil() {
			elem := fieldValue.Elem()
			if isASTNode(elem) {
				children = append(children, fieldValue.Interface())
			}
		}
	}

	return children
}

// isASTNode checks if a value is an AST node (has GetType or Type method).
func isASTNode(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}

	// Check for GetType() string method
	method := v.MethodByName("GetType")
	if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 1 {
		if method.Type().Out(0).Kind() == reflect.String {
			return true
		}
	}

	// Check for Type() string method
	method = v.MethodByName("Type")
	if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 1 {
		if method.Type().Out(0).Kind() == reflect.String {
			return true
		}
	}

	return false
}

// GenericNode represents a reconstructed AST node from flat format.
// This is used when we don't have the original Go type constructors.
type GenericNode struct {
	Type       string
	Properties map[string]any
	Children   []*GenericNode
	Parent     *GenericNode
	Index      int
	Parens     bool
	ParensInOp bool
	Visible    *bool
	FileInfo   map[string]any
}

// GetType returns the node type name.
func (g *GenericNode) GetType() string {
	return g.Type
}

// ASTUnflattener reconstructs AST nodes from FlatAST format.
type ASTUnflattener struct {
	flat  *FlatAST
	nodes []*GenericNode
}

// NewASTUnflattener creates a new AST unflattener.
func NewASTUnflattener(flat *FlatAST) *ASTUnflattener {
	return &ASTUnflattener{
		flat:  flat,
		nodes: make([]*GenericNode, len(flat.Nodes)),
	}
}

// Unflatten reconstructs the AST from the flat representation.
// Returns the root GenericNode.
func (au *ASTUnflattener) Unflatten() (*GenericNode, error) {
	if au.flat.NodeCount == 0 {
		return nil, fmt.Errorf("empty AST")
	}

	// First pass: create all nodes
	for i, flatNode := range au.flat.Nodes {
		node := &GenericNode{
			Type:       TypeNames[flatNode.TypeID],
			Properties: au.flat.GetProperties(flatNode.PropsOffset, flatNode.PropsLength),
			Children:   make([]*GenericNode, 0),
		}

		// Extract flags
		if flatNode.Flags&FlagParens != 0 {
			node.Parens = true
		}
		if flatNode.Flags&FlagParensInOp != 0 {
			node.ParensInOp = true
		}
		if flatNode.Flags&FlagVisibleSet != 0 {
			visible := flatNode.Flags&FlagVisible != 0
			node.Visible = &visible
		}

		au.nodes[i] = node
	}

	// Second pass: link parent/child relationships
	for i, flatNode := range au.flat.Nodes {
		node := au.nodes[i]

		// Link parent - note: ParentIndex=0 means parent is at index 0 (which could be root)
		// Only link if not root itself and parent index is valid
		if uint32(i) != au.flat.RootIndex && int(flatNode.ParentIndex) < len(au.nodes) {
			node.Parent = au.nodes[flatNode.ParentIndex]
		}

		// Collect children by following the sibling chain
		if flatNode.ChildIndex != 0 {
			childIdx := flatNode.ChildIndex
			for childIdx != 0 && int(childIdx) < len(au.nodes) {
				child := au.nodes[childIdx]
				node.Children = append(node.Children, child)
				childIdx = au.flat.Nodes[childIdx].NextIndex
			}
		}
	}

	// Return root node
	if int(au.flat.RootIndex) >= len(au.nodes) {
		return nil, fmt.Errorf("invalid root index: %d", au.flat.RootIndex)
	}

	return au.nodes[au.flat.RootIndex], nil
}

// UnflattenAST is a convenience function to unflatten an AST.
func UnflattenAST(flat *FlatAST) (*GenericNode, error) {
	unflattener := NewASTUnflattener(flat)
	return unflattener.Unflatten()
}

// ResolveString resolves a string index to its value.
func (g *GenericNode) ResolveString(flat *FlatAST, propName string) string {
	if g.Properties == nil {
		return ""
	}
	if idx, ok := g.Properties[propName]; ok {
		switch v := idx.(type) {
		case float64:
			return flat.GetString(uint32(v))
		case int:
			return flat.GetString(uint32(v))
		case uint32:
			return flat.GetString(v)
		case string:
			return v
		}
	}
	return ""
}

// GetFloat64 retrieves a float64 property.
func (g *GenericNode) GetFloat64(propName string) (float64, bool) {
	if g.Properties == nil {
		return 0, false
	}
	if v, ok := g.Properties[propName]; ok {
		switch n := v.(type) {
		case float64:
			return n, true
		case int:
			return float64(n), true
		case int64:
			return float64(n), true
		}
	}
	return 0, false
}

// GetBool retrieves a boolean property.
func (g *GenericNode) GetBool(propName string) (bool, bool) {
	if g.Properties == nil {
		return false, false
	}
	if v, ok := g.Properties[propName]; ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// ToJSON converts the GenericNode tree to a JSON-serializable map.
func (g *GenericNode) ToJSON() map[string]any {
	result := map[string]any{
		"type": g.Type,
	}

	if g.Properties != nil && len(g.Properties) > 0 {
		result["properties"] = g.Properties
	}

	if len(g.Children) > 0 {
		children := make([]map[string]any, len(g.Children))
		for i, child := range g.Children {
			children[i] = child.ToJSON()
		}
		result["children"] = children
	}

	if g.Parens {
		result["parens"] = true
	}
	if g.ParensInOp {
		result["parensInOp"] = true
	}
	if g.Visible != nil {
		result["visible"] = *g.Visible
	}

	return result
}
