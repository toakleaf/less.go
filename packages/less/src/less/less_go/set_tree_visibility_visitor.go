package less_go

// VisibilityNode interface for nodes that support visibility operations
type VisibilityNode interface {
	BlocksVisibility() bool
	EnsureVisibility()
	EnsureInvisibility()
}

// AcceptableNode interface for nodes that accept visitors
type AcceptableNode interface {
	Accept(visitor any)
}

// SetTreeVisibilityVisitor implements the visitor pattern to set tree visibility
type SetTreeVisibilityVisitor struct {
	visible bool
}

// NewSetTreeVisibilityVisitor creates a new SetTreeVisibilityVisitor instance
func NewSetTreeVisibilityVisitor(visible any) *SetTreeVisibilityVisitor {
	return &SetTreeVisibilityVisitor{
		visible: isTruthyValue(visible),
	}
}

// isTruthyValue determines if a value is truthy (JavaScript-like behavior)
// This is evaluated once at construction time to avoid repeated checks
func isTruthyValue(value any) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case float32:
		return v != 0.0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	default:
		// Non-nil non-basic types are truthy
		return true
	}
}

// Run starts the visitor on the root node
func (v *SetTreeVisibilityVisitor) Run(root any) {
	v.Visit(root)
}

// VisitArray visits an array of nodes using type assertions
// Returns the nodes array to match the interface expected by Ruleset.Accept
func (v *SetTreeVisibilityVisitor) VisitArray(nodes []any) []any {
	for _, node := range nodes {
		v.Visit(node)
	}
	return nodes
}

// Visit visits a single node using type assertions for fast dispatch
func (v *SetTreeVisibilityVisitor) Visit(node any) any {
	if node == nil {
		return node
	}

	// Fast path: Check for []any slice (most common array type)
	if arr, ok := node.([]any); ok {
		v.VisitArray(arr)
		return node
	}

	// Check if node blocks visibility using interface assertion (no reflection)
	if visNode, ok := node.(interface{ BlocksVisibility() bool }); ok {
		if visNode.BlocksVisibility() {
			// Match JavaScript behavior: if node blocks visibility, return early without visiting children
			return node
		}
	}

	// Set visibility based on visitor's visible flag
	if v.visible {
		if visNode, ok := node.(interface{ EnsureVisibility() }); ok {
			visNode.EnsureVisibility()
		}
	} else {
		if visNode, ok := node.(interface{ EnsureInvisibility() }); ok {
			visNode.EnsureInvisibility()
		}
	}

	// Call accept method if it exists
	if acceptNode, ok := node.(interface{ Accept(any) }); ok {
		acceptNode.Accept(v)
	}

	return node
}