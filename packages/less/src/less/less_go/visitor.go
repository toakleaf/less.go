package less_go

import (
	"reflect"
	"strconv"
)

var _visitArgs = map[string]bool{"visitDeeper": true}
var _hasIndexed = false

// VisitArgs represents the visit arguments passed to visitor functions
type VisitArgs struct {
	VisitDeeper bool
}

// TreeRegistry represents the tree module containing all node types
type TreeRegistry struct {
	NodeTypes map[string]any
}

// Global tree registry - will be populated with actual node types
var treeRegistry = &TreeRegistry{
	NodeTypes: make(map[string]any),
}

type AcceptorNode interface {
	Accept(visitor any)
}

// NodeWithTypeIndex represents a node that has type and typeIndex
type NodeWithTypeIndex interface {
	GetTypeIndex() int
}

// NodeWithValue represents a node that has a value field (like MixinCall args)
type NodeWithValue interface {
	GetValue() any
}

// ArrayLikeNode represents nodes that behave like arrays (have splice method)
type ArrayLikeNode interface {
	Len() int
	Get(i int) any
}

// Implementation represents the visitor implementation interface
type Implementation interface {
	IsReplacing() bool
}

// noop function that returns the node unchanged
func _noop(node any) any {
	return node
}


// VisitFunc represents a visitor function
type VisitFunc func(node any, visitArgs *VisitArgs) any

// VisitOutFunc represents a visitor out function  
type VisitOutFunc func(node any)

// Visitor represents the visitor pattern implementation
type Visitor struct {
	implementation any
	visitInCache   map[int]VisitFunc
	visitOutCache  map[int]VisitOutFunc
	methodLookup   map[string]reflect.Value // Pre-built method lookup map
}

// NewVisitor creates a new visitor with the given implementation
func NewVisitor(implementation any) *Visitor {
	v := &Visitor{
		implementation: implementation,
		visitInCache:   make(map[int]VisitFunc),
		visitOutCache:  make(map[int]VisitOutFunc),
		methodLookup:   make(map[string]reflect.Value),
	}

	if !_hasIndexed {
		// Initialize tree with actual go_parser node types
		initializeTree()
		// Index the NodeTypes map directly since that's where our prototypes are
		ticker := 1
		for _, nodeProto := range treeRegistry.NodeTypes {
			if proto, ok := nodeProto.(*NodePrototype); ok {
				proto.SetTypeIndex(ticker)
				ticker++
			}
		}
		_hasIndexed = true
	}

	// Pre-build method lookup map to avoid MethodByName calls
	if implementation != nil {
		implValue := reflect.ValueOf(implementation)
		implType := implValue.Type()
		numMethods := implType.NumMethod()

		for i := 0; i < numMethods; i++ {
			method := implType.Method(i)
			// Store method by name for fast lookup
			v.methodLookup[method.Name] = implValue.Method(i)
		}
	}

	return v
}

// buildVisitFunctions creates visit functions using type switches instead of reflection
func (v *Visitor) buildVisitFunctions(nodeType string) (VisitFunc, VisitOutFunc) {
	// Type switch on implementation to avoid reflection
	switch impl := v.implementation.(type) {
	case *ToCSSVisitor:
		return v.buildToCSSVisitorFunctions(impl, nodeType)
	case *ImportVisitor:
		return v.buildImportVisitorFunctions(impl, nodeType)
	case *JoinSelectorVisitor:
		return v.buildJoinSelectorVisitorFunctions(impl, nodeType)
	case *ExtendFinderVisitor:
		return v.buildExtendFinderVisitorFunctions(impl, nodeType)
	case *ProcessExtendsVisitor:
		return v.buildProcessExtendsVisitorFunctions(impl, nodeType)
	case *SetTreeVisibilityVisitor:
		// SetTreeVisibilityVisitor doesn't use the standard Visit pattern
		// It has custom Visit/VisitArray methods
		return v.buildReflectionBasedFunctions(nodeType)
	default:
		// Fallback to reflection for unknown types (test mocks, etc.)
		return v.buildReflectionBasedFunctions(nodeType)
	}
}

// buildToCSSVisitorFunctions creates functions for ToCSSVisitor
func (v *Visitor) buildToCSSVisitorFunctions(impl *ToCSSVisitor, nodeType string) (VisitFunc, VisitOutFunc) {
	switch nodeType {
	case "Declaration":
		return func(n any, args *VisitArgs) any { return impl.VisitDeclaration(n, args) }, nil
	case "MixinDefinition":
		return func(n any, args *VisitArgs) any { return impl.VisitMixinDefinition(n, args) }, nil
	case "Extend":
		return func(n any, args *VisitArgs) any { return impl.VisitExtend(n, args) }, nil
	case "Comment":
		return func(n any, args *VisitArgs) any { return impl.VisitComment(n, args) }, nil
	case "Media":
		return func(n any, args *VisitArgs) any { return impl.VisitMedia(n, args) }, nil
	case "Container":
		return func(n any, args *VisitArgs) any { return impl.VisitContainer(n, args) }, nil
	case "Import":
		return func(n any, args *VisitArgs) any { return impl.VisitImport(n, args) }, nil
	case "AtRule":
		return func(n any, args *VisitArgs) any { return impl.VisitAtRule(n, args) }, nil
	case "Anonymous":
		return func(n any, args *VisitArgs) any { return impl.VisitAnonymous(n, args) }, nil
	case "Ruleset":
		return func(n any, args *VisitArgs) any { return impl.VisitRuleset(n, args) }, nil
	default:
		return func(n any, args *VisitArgs) any { return _noop(n) }, nil
	}
}

// buildImportVisitorFunctions creates functions for ImportVisitor
func (v *Visitor) buildImportVisitorFunctions(impl *ImportVisitor, nodeType string) (VisitFunc, VisitOutFunc) {
	switch nodeType {
	case "Import":
		return func(n any, args *VisitArgs) any { impl.VisitImport(n, args); return n }, nil
	case "Media":
		return func(n any, args *VisitArgs) any { impl.VisitMedia(n, args); return n },
			func(n any) { impl.VisitMediaOut(n) }
	case "AtRule":
		return func(n any, args *VisitArgs) any { impl.VisitAtRule(n, args); return n },
			func(n any) { impl.VisitAtRuleOut(n) }
	case "Declaration":
		return func(n any, args *VisitArgs) any { impl.VisitDeclaration(n, args); return n },
			func(n any) { impl.VisitDeclarationOut(n) }
	case "MixinDefinition":
		return func(n any, args *VisitArgs) any { impl.VisitMixinDefinition(n, args); return n },
			func(n any) { impl.VisitMixinDefinitionOut(n) }
	case "Ruleset":
		return func(n any, args *VisitArgs) any { impl.VisitRuleset(n, args); return n },
			func(n any) { impl.VisitRulesetOut(n) }
	default:
		return func(n any, args *VisitArgs) any { return _noop(n) }, nil
	}
}

// buildJoinSelectorVisitorFunctions creates functions for JoinSelectorVisitor
func (v *Visitor) buildJoinSelectorVisitorFunctions(impl *JoinSelectorVisitor, nodeType string) (VisitFunc, VisitOutFunc) {
	switch nodeType {
	case "Ruleset":
		return func(n any, args *VisitArgs) any { return impl.VisitRuleset(n, args) },
			func(n any) { impl.VisitRulesetOut(n) }
	case "Media":
		return func(n any, args *VisitArgs) any { return impl.VisitMedia(n, args) }, nil
	case "Container":
		return func(n any, args *VisitArgs) any { return impl.VisitContainer(n, args) }, nil
	case "AtRule":
		return func(n any, args *VisitArgs) any { return impl.VisitAtRule(n, args) }, nil
	case "Declaration":
		return func(n any, args *VisitArgs) any { return impl.VisitDeclaration(n, args) }, nil
	case "MixinDefinition":
		return func(n any, args *VisitArgs) any { return impl.VisitMixinDefinition(n, args) }, nil
	default:
		return func(n any, args *VisitArgs) any { return _noop(n) }, nil
	}
}

// buildExtendFinderVisitorFunctions creates functions for ExtendFinderVisitor
func (v *Visitor) buildExtendFinderVisitorFunctions(impl *ExtendFinderVisitor, nodeType string) (VisitFunc, VisitOutFunc) {
	switch nodeType {
	case "Ruleset":
		return func(n any, args *VisitArgs) any { impl.VisitRuleset(n, args); return n },
			func(n any) { impl.VisitRulesetOut(n) }
	case "Media":
		return func(n any, args *VisitArgs) any { impl.VisitMedia(n, args); return n },
			func(n any) { impl.VisitMediaOut(n) }
	case "AtRule":
		return func(n any, args *VisitArgs) any { impl.VisitAtRule(n, args); return n },
			func(n any) { impl.VisitAtRuleOut(n) }
	case "Declaration":
		return func(n any, args *VisitArgs) any { impl.VisitDeclaration(n, args); return n }, nil
	case "MixinDefinition":
		return func(n any, args *VisitArgs) any { impl.VisitMixinDefinition(n, args); return n }, nil
	default:
		return func(n any, args *VisitArgs) any { return _noop(n) }, nil
	}
}

// buildProcessExtendsVisitorFunctions creates functions for ProcessExtendsVisitor
func (v *Visitor) buildProcessExtendsVisitorFunctions(impl *ProcessExtendsVisitor, nodeType string) (VisitFunc, VisitOutFunc) {
	switch nodeType {
	case "Ruleset":
		return func(n any, args *VisitArgs) any { impl.VisitRuleset(n, args); return n }, nil
	case "Media":
		return func(n any, args *VisitArgs) any { impl.VisitMedia(n, args); return n },
			func(n any) { impl.VisitMediaOut(n) }
	case "AtRule":
		return func(n any, args *VisitArgs) any { impl.VisitAtRule(n, args); return n },
			func(n any) { impl.VisitAtRuleOut(n) }
	case "Declaration":
		return func(n any, args *VisitArgs) any { impl.VisitDeclaration(n, args); return n }, nil
	case "MixinDefinition":
		return func(n any, args *VisitArgs) any { impl.VisitMixinDefinition(n, args); return n }, nil
	case "Selector":
		return func(n any, args *VisitArgs) any { impl.VisitSelector(n, args); return n }, nil
	default:
		return func(n any, args *VisitArgs) any { return _noop(n) }, nil
	}
}

// buildReflectionBasedFunctions creates functions using reflection (fallback for unknown types)
func (v *Visitor) buildReflectionBasedFunctions(nodeType string) (VisitFunc, VisitOutFunc) {
	// Build function name like JS: `visit${node.type}`
	fnName := "Visit" + nodeType

	// Use pre-built method lookup map instead of MethodByName
	visitMethod, visitMethodExists := v.methodLookup[fnName]
	visitOutMethod, visitOutMethodExists := v.methodLookup[fnName+"Out"]

	var visitFunc VisitFunc
	var visitOutFunc VisitOutFunc

	// Create visit function (use _noop if method doesn't exist)
	if visitMethodExists && visitMethod.IsValid() {
		visitFunc = func(n any, args *VisitArgs) any {
			results := visitMethod.Call([]reflect.Value{
				reflect.ValueOf(n),
				reflect.ValueOf(args),
			})
			if len(results) > 0 {
				return results[0].Interface()
			}
			return n
		}
	} else {
		visitFunc = func(n any, args *VisitArgs) any {
			return _noop(n)
		}
	}

	// Create visitOut function (_noop if method doesn't exist)
	if visitOutMethodExists && visitOutMethod.IsValid() {
		visitOutFunc = func(n any) {
			visitOutMethod.Call([]reflect.Value{reflect.ValueOf(n)})
		}
	} else {
		visitOutFunc = func(n any) {
			// _noop for visitOut
		}
	}

	return visitFunc, visitOutFunc
}

// Visit visits a node using the visitor pattern
func (v *Visitor) Visit(node any) any {
	if node == nil {
		return node
	}

	var nodeTypeIndex int
	var nodeType string

	// Try to get type from the node first
	if nodeWithType, ok := node.(NodeWithType); ok {
		nodeType = nodeWithType.GetType()
	} else {
		// Fallback: get type from reflection (struct name)
		nodeVal := reflect.ValueOf(node)
		if nodeVal.Kind() == reflect.Ptr && !nodeVal.IsNil() {
			nodeType = nodeVal.Elem().Type().Name()
		} else if nodeVal.Kind() == reflect.Struct {
			nodeType = nodeVal.Type().Name()
		}
	}

	// Try to get typeIndex from the node - this mirrors JS behavior
	if nodeWithTypeIndex, ok := node.(NodeWithTypeIndex); ok {
		nodeTypeIndex = nodeWithTypeIndex.GetTypeIndex()
	}

	if nodeTypeIndex == 0 {
		// MixinCall args aren't a node type? - exact JS comment
		if nodeWithValue, ok := node.(NodeWithValue); ok {
			if value := nodeWithValue.GetValue(); value != nil {
				if valueWithTypeIndex, ok := value.(NodeWithTypeIndex); ok && valueWithTypeIndex.GetTypeIndex() != 0 {
					v.Visit(value)
				}
			}
		}
		return node
	}

	var visitFunc VisitFunc
	var visitOutFunc VisitOutFunc
	visitArgs := &VisitArgs{VisitDeeper: true}

	// Check cache first
	if cachedFunc, exists := v.visitInCache[nodeTypeIndex]; exists {
		visitFunc = cachedFunc
		visitOutFunc = v.visitOutCache[nodeTypeIndex]
	} else {
		// Build visit functions without reflection for known visitor types
		// This is a more comprehensive optimization than the previous approach which only
		// optimized 4 hot-path node types. We now optimize all visitor types and all their methods.
		visitFunc, visitOutFunc = v.buildVisitFunctions(nodeType)

		// Cache the functions
		v.visitInCache[nodeTypeIndex] = visitFunc
		v.visitOutCache[nodeTypeIndex] = visitOutFunc
	}

	// Call visit function (if not _noop)
	if visitFunc != nil {
		newNode := visitFunc(node, visitArgs)
		if v.isReplacing() {
			node = newNode
		}
	}

	// Visit deeper if requested and node exists
	if visitArgs.VisitDeeper && node != nil {
		// Check if node has length property (array-like)
		nodeVal := reflect.ValueOf(node)
		if nodeVal.Kind() == reflect.Ptr && !nodeVal.IsNil() {
			nodeVal = nodeVal.Elem()
		}

		// Check for array-like behavior: has length property and numeric indexing
		if nodeVal.Kind() == reflect.Struct {
			lengthField := nodeVal.FieldByName("length")
			elementsField := nodeVal.FieldByName("Elements")
			
			if lengthField.IsValid() && lengthField.Kind() == reflect.Int && lengthField.Int() > 0 {
				// Array-like node processing
				length := int(lengthField.Int())
				
				// First try Elements field (Go-style array-like nodes)
				if elementsField.IsValid() && elementsField.Kind() == reflect.Slice {
					elementsSlice := elementsField
					for i := 0; i < length && i < elementsSlice.Len(); i++ {
						item := elementsSlice.Index(i).Interface()
						if accepter, ok := item.(interface{ Accept(any) }); ok {
							accepter.Accept(v)
						}
					}
				} else {
					// Fallback: try to get element at index i (like node[i] in JS)
					for i := 0; i < length; i++ {
						indexField := nodeVal.FieldByName(strconv.Itoa(i))
						if indexField.IsValid() && indexField.CanInterface() {
							item := indexField.Interface()
							if accepter, ok := item.(interface{ Accept(any) }); ok {
								accepter.Accept(v)
							}
						}
					}
				}
			} else {
				// Regular node - call accept if available
				if accepter, ok := node.(interface{ Accept(any) }); ok {
					accepter.Accept(v)
				}
			}
		} else if accepter, ok := node.(interface{ Accept(any) }); ok {
			accepter.Accept(v)
		}
	}

	// Call visitOut function (if not _noop)
	if visitOutFunc != nil {
		visitOutFunc(node)
	}

	return node
}

// VisitArray visits an array of nodes
func (v *Visitor) VisitArray(nodes []any, nonReplacing ...bool) []any {
	if nodes == nil {
		return nodes
	}

	cnt := len(nodes)
	var isNonReplacing bool
	if len(nonReplacing) > 0 {
		isNonReplacing = nonReplacing[0]
	}
	
	// Non-replacing mode
	if isNonReplacing || !v.isReplacing() {
		for i := 0; i < cnt; i++ {
			v.Visit(nodes[i])
		}
		return nodes
	}

	// Replacing mode
	out := make([]any, 0)
	for i := 0; i < cnt; i++ {
		evaluated := v.Visit(nodes[i])
		if evaluated == nil {
			continue // Skip undefined results like JS
		}
		
		// Check if result is array-like (Go slice or has splice method)
		if v.isArrayLike(evaluated) {
			// It's array-like, flatten it
			if arrayItems := v.convertToSlice(evaluated); len(arrayItems) > 0 {
				v.Flatten(arrayItems, &out)
			}
		} else {
			// Regular item, add to output
			out = append(out, evaluated)
		}
	}
	return out
}

// Flatten flattens nested arrays into a single array
func (v *Visitor) Flatten(arr []any, out *[]any) []any {
	if out == nil {
		result := make([]any, 0)
		out = &result
	}

	for _, item := range arr {
		if item == nil {
			continue // Skip undefined items like JS
		}
		
		// Check if item is array-like (Go slice or has splice method)
		if v.isArrayLike(item) {
			// Recursively flatten nested arrays
			if nestedItems := v.convertToSlice(item); len(nestedItems) > 0 {
				v.Flatten(nestedItems, out)
			}
		} else {
			// Regular item, add to output
			*out = append(*out, item)
		}
	}

	return *out
}

// isReplacing checks if the implementation is replacing
func (v *Visitor) isReplacing() bool {
	if impl, ok := v.implementation.(Implementation); ok {
		return impl.IsReplacing()
	}
	
	// Use reflection to check for isReplacing property (like JS)
	implValue := reflect.ValueOf(v.implementation)
	if implValue.Kind() == reflect.Ptr {
		implValue = implValue.Elem()
	}
	
	// Check for isReplacing field first (direct property access like JS)
	if implValue.Kind() == reflect.Struct {
		isReplacingField := implValue.FieldByName("isReplacing")
		if isReplacingField.IsValid() && isReplacingField.Kind() == reflect.Bool {
			return isReplacingField.Bool()
		}
		// Also check IsReplacing field (Go naming convention)
		isReplacingField = implValue.FieldByName("IsReplacing")
		if isReplacingField.IsValid() && isReplacingField.Kind() == reflect.Bool {
			return isReplacingField.Bool()
		}
	}
	
	// Fallback to method call
	method := reflect.ValueOf(v.implementation).MethodByName("IsReplacing")
	if method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			if boolResult, ok := results[0].Interface().(bool); ok {
				return boolResult
			}
		}
	}
	
	return false
}

// hasSpliceMethod checks if an object has a splice method (array-like detection)
func (v *Visitor) hasSpliceMethod(obj any) bool {
	if obj == nil {
		return false
	}
	
	// Check for splice method using reflection
	objValue := reflect.ValueOf(obj)
	spliceMethod := objValue.MethodByName("Splice")
	return spliceMethod.IsValid()
}

// isArrayLike checks if an object is array-like (Go slice or has splice method)
func (v *Visitor) isArrayLike(obj any) bool {
	if obj == nil {
		return false
	}
	
	// Check if it's a Go slice
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Slice {
		return true
	}
	
	// Also check for splice method (JS-style arrays)
	return v.hasSpliceMethod(obj)
}

// convertToSlice converts an array-like object to a Go slice
func (v *Visitor) convertToSlice(obj any) []any {
	if obj == nil {
		return nil
	}
	
	// Try ArrayLikeNode interface first
	if arrayLike, ok := obj.(ArrayLikeNode); ok {
		length := arrayLike.Len()
		result := make([]any, length)
		for i := 0; i < length; i++ {
			result[i] = arrayLike.Get(i)
		}
		return result
	}
	
	// Fallback: try to access like JS array (length property + numeric indexing)
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Ptr && !objValue.IsNil() {
		objValue = objValue.Elem()
	}
	
	if objValue.Kind() == reflect.Struct {
		lengthField := objValue.FieldByName("length")
		if lengthField.IsValid() && lengthField.Kind() == reflect.Int {
			length := int(lengthField.Int())
			result := make([]any, 0, length)
			for i := 0; i < length; i++ {
				indexField := objValue.FieldByName(strconv.Itoa(i))
				if indexField.IsValid() && indexField.CanInterface() {
					result = append(result, indexField.Interface())
				}
			}
			return result
		}
	}
	
	// If it's already a slice, return it
	if objValue.Kind() == reflect.Slice {
		result := make([]any, objValue.Len())
		for i := 0; i < objValue.Len(); i++ {
			result[i] = objValue.Index(i).Interface()
		}
		return result
	}
	
	return nil
}

// initializeTree initializes the tree registry with node type prototypes
// This matches the JS behavior of indexing node constructors by their prototype.type
func initializeTree() {
	// For now, create placeholder prototypes that match the expected node types
	// This will work with the dynamic method resolution pattern
	treeRegistry.NodeTypes = map[string]any{
		"Node":               createNodePrototype("Node"),
		"Color":              createNodePrototype("Color"),
		"AtRule":             createNodePrototype("AtRule"),
		"DetachedRuleset":    createNodePrototype("DetachedRuleset"),
		"Operation":          createNodePrototype("Operation"),
		"Dimension":          createNodePrototype("Dimension"),
		"Unit":               createNodePrototype("Unit"),
		"Keyword":            createNodePrototype("Keyword"),
		"Variable":           createNodePrototype("Variable"),
		"Property":           createNodePrototype("Property"),
		"Ruleset":            createNodePrototype("Ruleset"),
		"Element":            createNodePrototype("Element"),
		"Attribute":          createNodePrototype("Attribute"),
		"Combinator":         createNodePrototype("Combinator"),
		"Selector":           createNodePrototype("Selector"),
		"Quoted":             createNodePrototype("Quoted"),
		"Expression":         createNodePrototype("Expression"),
		"Declaration":        createNodePrototype("Declaration"),
		"Call":               createNodePrototype("Call"),
		"URL":                createNodePrototype("URL"),
		"Import":             createNodePrototype("Import"),
		"Comment":            createNodePrototype("Comment"),
		"Anonymous":          createNodePrototype("Anonymous"),
		"Value":              createNodePrototype("Value"),
		"JavaScript":         createNodePrototype("JavaScript"),
		"Assignment":         createNodePrototype("Assignment"),
		"Condition":          createNodePrototype("Condition"),
		"QueryInParens":      createNodePrototype("QueryInParens"),
		"Paren":              createNodePrototype("Paren"),
		"Media":              createNodePrototype("Media"),
		"Container":          createNodePrototype("Container"),
		"UnicodeDescriptor":  createNodePrototype("UnicodeDescriptor"),
		"Negative":           createNodePrototype("Negative"),
		"Extend":             createNodePrototype("Extend"),
		"VariableCall":       createNodePrototype("VariableCall"),
		"NamespaceValue":     createNodePrototype("NamespaceValue"),
		// Mixin types
		"MixinCall":       createNodePrototype("MixinCall"),
		"MixinDefinition": createNodePrototype("MixinDefinition"),
	}
}

// createNodePrototype creates a prototype object for a node type
type NodePrototype struct {
	Type      string
	TypeIndex int
}

func (np *NodePrototype) SetTypeIndex(index int) {
	np.TypeIndex = index
}

func (np *NodePrototype) GetPrototype() any {
	return np
}

func createNodePrototype(nodeType string) *NodePrototype {
	return &NodePrototype{
		Type:      nodeType,
		TypeIndex: 0, // Will be set by indexNodeTypes
	}
}

// GetTypeIndexForNodeType returns the TypeIndex for a given node type string
// This is used by node constructors to set the TypeIndex field
func GetTypeIndexForNodeType(nodeType string) int {
	if !_hasIndexed {
		// Initialize if needed - this ensures prototypes are indexed
		NewVisitor(nil)
	}

	if nodeProto, ok := treeRegistry.NodeTypes[nodeType]; ok {
		if proto, ok := nodeProto.(*NodePrototype); ok {
			return proto.TypeIndex
		}
	}
	return 0
}