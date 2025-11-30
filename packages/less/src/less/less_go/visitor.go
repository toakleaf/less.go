package less_go

import (
	"reflect"
	"strconv"
)

var _hasIndexed = false

type VisitArgs struct {
	VisitDeeper bool
}

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

type Implementation interface {
	IsReplacing() bool
}

// DirectDispatchVisitor interface for direct dispatch without reflection
// Implementations can optionally implement this for better performance
type DirectDispatchVisitor interface {
	VisitNode(node any, visitArgs *VisitArgs) (result any, handled bool)
	VisitNodeOut(node any) bool
}

func _noop(node any) any {
	return node
}

type VisitFunc func(node any, visitArgs *VisitArgs) any

type VisitOutFunc func(node any)

type Visitor struct {
	implementation any
	visitInCache   map[int]VisitFunc
	visitOutCache  map[int]VisitOutFunc
	methodLookup   map[string]reflect.Value // Pre-built method lookup map
}

func NewVisitor(implementation any) *Visitor {
	v := &Visitor{
		implementation: implementation,
		visitInCache:   make(map[int]VisitFunc),
		visitOutCache:  make(map[int]VisitOutFunc),
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

	// Only build method lookup map if NOT using direct dispatch
	// This avoids reflection overhead for implementations that handle all node types
	if implementation != nil {
		if _, usesDirectDispatch := implementation.(DirectDispatchVisitor); !usesDirectDispatch {
			v.methodLookup = make(map[string]reflect.Value)
			implValue := reflect.ValueOf(implementation)
			implType := implValue.Type()
			numMethods := implType.NumMethod()

			for i := 0; i < numMethods; i++ {
				method := implType.Method(i)
				v.methodLookup[method.Name] = implValue.Method(i)
			}
		}
	}

	return v
}

func (v *Visitor) Visit(node any) any {
	if node == nil {
		return node
	}

	var nodeTypeIndex int
	var nodeType string

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

	visitArgs := &VisitArgs{VisitDeeper: true}

	// Fast path: Check if implementation supports direct dispatch (no reflection)
	if directDispatcher, ok := v.implementation.(DirectDispatchVisitor); ok {
		newNode, handled := directDispatcher.VisitNode(node, visitArgs)
		if handled {
			if v.isReplacing() {
				node = newNode
			}
		}
	} else {
		// Slow path: Use reflection-based dispatch (backward compatibility)
		var visitFunc VisitFunc
		var visitOutFunc VisitOutFunc

		if cachedFunc, exists := v.visitInCache[nodeTypeIndex]; exists {
			visitFunc = cachedFunc
			visitOutFunc = v.visitOutCache[nodeTypeIndex]
		} else {
			// Build function name like JS: `visit${node.type}`
			// Use string concatenation instead of fmt.Sprintf for performance
			fnName := "Visit" + nodeType

			// Use pre-built method lookup map instead of MethodByName
			visitMethod, visitMethodExists := v.methodLookup[fnName]
			visitOutMethod, visitOutMethodExists := v.methodLookup[fnName+"Out"]

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

			v.visitInCache[nodeTypeIndex] = visitFunc
			v.visitOutCache[nodeTypeIndex] = visitOutFunc
		}

		if visitFunc != nil {
			newNode := visitFunc(node, visitArgs)
			if v.isReplacing() {
				node = newNode
			}
		}
	}

	if visitArgs.VisitDeeper && node != nil {
		nodeVal := reflect.ValueOf(node)
		if nodeVal.Kind() == reflect.Ptr && !nodeVal.IsNil() {
			nodeVal = nodeVal.Elem()
		}

		if nodeVal.Kind() == reflect.Struct {
			lengthField := nodeVal.FieldByName("length")
			elementsField := nodeVal.FieldByName("Elements")

			if lengthField.IsValid() && lengthField.Kind() == reflect.Int && lengthField.Int() > 0 {
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
				if accepter, ok := node.(interface{ Accept(any) }); ok {
					accepter.Accept(v)
				}
			}
		} else if accepter, ok := node.(interface{ Accept(any) }); ok {
			accepter.Accept(v)
		}
	}

	if directDispatcher, ok := v.implementation.(DirectDispatchVisitor); ok {
		directDispatcher.VisitNodeOut(node)
	} else {
		// Slow path: reflection-based dispatch (visitOutFunc was cached in the else block above)
		if visitOutFunc, exists := v.visitOutCache[nodeTypeIndex]; exists && visitOutFunc != nil {
			visitOutFunc(node)
		}
	}

	return node
}

func (v *Visitor) VisitArray(nodes []any, nonReplacing ...bool) []any {
	if nodes == nil {
		return nodes
	}

	cnt := len(nodes)
	var isNonReplacing bool
	if len(nonReplacing) > 0 {
		isNonReplacing = nonReplacing[0]
	}

	if isNonReplacing || !v.isReplacing() {
		for i := 0; i < cnt; i++ {
			v.Visit(nodes[i])
		}
		return nodes
	}

	out := make([]any, 0)
	for i := 0; i < cnt; i++ {
		evaluated := v.Visit(nodes[i])
		if evaluated == nil {
			continue // Skip undefined results like JS
		}

		if v.isArrayLike(evaluated) {
			if arrayItems := v.convertToSlice(evaluated); len(arrayItems) > 0 {
				v.Flatten(arrayItems, &out)
			}
		} else {
			out = append(out, evaluated)
		}
	}
	return out
}

func (v *Visitor) Flatten(arr []any, out *[]any) []any {
	if out == nil {
		result := make([]any, 0)
		out = &result
	}

	for _, item := range arr {
		if item == nil {
			continue // Skip undefined items like JS
		}

		if v.isArrayLike(item) {
			if nestedItems := v.convertToSlice(item); len(nestedItems) > 0 {
				v.Flatten(nestedItems, out)
			}
		} else {
			*out = append(*out, item)
		}
	}

	return *out
}

func (v *Visitor) isReplacing() bool {
	if impl, ok := v.implementation.(Implementation); ok {
		return impl.IsReplacing()
	}

	// Use reflection to check for isReplacing property (like JS)
	implValue := reflect.ValueOf(v.implementation)
	if implValue.Kind() == reflect.Ptr {
		implValue = implValue.Elem()
	}

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

func (v *Visitor) hasSpliceMethod(obj any) bool {
	if obj == nil {
		return false
	}

	objValue := reflect.ValueOf(obj)
	spliceMethod := objValue.MethodByName("Splice")
	return spliceMethod.IsValid()
}

func (v *Visitor) isArrayLike(obj any) bool {
	if obj == nil {
		return false
	}

	objValue := reflect.ValueOf(obj)
	if objValue.Kind() == reflect.Slice {
		return true
	}

	// Also check for splice method (JS-style arrays)
	return v.hasSpliceMethod(obj)
}

func (v *Visitor) convertToSlice(obj any) []any {
	if obj == nil {
		return nil
	}

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
		TypeIndex: 0,
	}
}

// GetTypeIndexForNodeType returns the TypeIndex for a given node type string.
// Used by node constructors to set the TypeIndex field.
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