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

// Reset resets the SetTreeVisibilityVisitor for reuse from the pool.
func (v *SetTreeVisibilityVisitor) Reset(visible any) {
	v.visible = isTruthyValue(visible)
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

// VisitArray visits an array of nodes using type assertions.
// This method matches the interface expected by Ruleset.Accept: VisitArray([]any) []any
func (v *SetTreeVisibilityVisitor) VisitArray(nodes []any) []any {
	for _, node := range nodes {
		v.Visit(node)
	}
	return nodes
}

// VisitArrayAny visits an array passed as any type.
// This is needed for compatibility with some Accept implementations that pass any instead of []any.
func (v *SetTreeVisibilityVisitor) visitArrayAny(nodes any) any {
	if arr, ok := nodes.([]any); ok {
		return v.VisitArray(arr)
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
	// Note: Some nodes (like Dimension, Unit) embed *Node which may be nil.
	// We need to check if the node has a valid embedded Node before calling BlocksVisibility.
	if nodeWithNode, ok := node.(interface{ GetNode() *Node }); ok {
		if n := nodeWithNode.GetNode(); n != nil {
			if n.BlocksVisibility() {
				// Match JavaScript behavior: if node blocks visibility, return early without visiting children
				return node
			}
		}
		// If n is nil, skip BlocksVisibility check - treat as not blocking
	} else if visNode, ok := node.(interface{ BlocksVisibility() bool }); ok {
		// Fallback for nodes that don't have GetNode() but do have BlocksVisibility()
		if visNode.BlocksVisibility() {
			return node
		}
	}

	// Set visibility based on visitor's visible flag
	// Same nil check pattern for EnsureVisibility/EnsureInvisibility
	if v.visible {
		if nodeWithNode, ok := node.(interface{ GetNode() *Node }); ok {
			if n := nodeWithNode.GetNode(); n != nil {
				n.EnsureVisibility()
			}
		} else if visNode, ok := node.(interface{ EnsureVisibility() }); ok {
			visNode.EnsureVisibility()
		}
	} else {
		if nodeWithNode, ok := node.(interface{ GetNode() *Node }); ok {
			if n := nodeWithNode.GetNode(); n != nil {
				n.EnsureInvisibility()
			}
		} else if visNode, ok := node.(interface{ EnsureInvisibility() }); ok {
			visNode.EnsureInvisibility()
		}
	}

	// Call accept method if it exists to visit children
	// Check for nil Node before calling Accept to avoid panic
	if nodeWithNode, ok := node.(interface{ GetNode() *Node }); ok {
		if n := nodeWithNode.GetNode(); n != nil {
			if acceptNode, ok := node.(interface{ Accept(any) }); ok {
				acceptNode.Accept(v)
			}
		}
	} else if acceptNode, ok := node.(interface{ Accept(any) }); ok {
		acceptNode.Accept(v)
	}

	return node
}