package tree

// NestableAtRulePrototype provides methods for nestable at-rules like @media and @supports
type NestableAtRulePrototype struct {
	Features interface{}   // The features/conditions part of the at-rule
	Rules    []interface{} // The rules inside the at-rule
	Type     string        // The type of at-rule (e.g., "Media", "Supports")
}

// NewNestableAtRulePrototype creates a new instance
func NewNestableAtRulePrototype() *NestableAtRulePrototype {
	return &NestableAtRulePrototype{}
}

// IsRulesetLike indicates this behaves like a ruleset
func (n *NestableAtRulePrototype) IsRulesetLike() bool {
	return true
}

// Accept visits the node with a visitor
func (n *NestableAtRulePrototype) Accept(visitor interface{}) {
	if v, ok := visitor.(interface{ Visit(interface{}) interface{} }); ok {
		if n.Features != nil {
			n.Features = v.Visit(n.Features)
		}
	}

	if v, ok := visitor.(interface {
		VisitArray([]interface{}) []interface{}
	}); ok {
		if n.Rules != nil {
			n.Rules = v.VisitArray(n.Rules)
		}
	}
}

// VisibilityInfo interface for getting visibility information
type VisibilityInfo interface {
	VisibilityInfo() interface{}
}

// Indexable interface for getting index information
type Indexable interface {
	GetIndex() int
}

// FileInfoProvider interface for getting file information
type FileInfoProvider interface {
	FileInfo() interface{}
}

// Parenter interface for setting parent relationships
type Parenter interface {
	SetParent(child interface{}, parent interface{})
}

// EvalTop evaluates the at-rule at the top level
func (n *NestableAtRulePrototype) EvalTop(context interface{}) (interface{}, error) {
	result := interface{}(n)

	// Extract context properties
	var mediaBlocks []interface{}

	if ctx, ok := context.(map[string]interface{}); ok {
		if mb, exists := ctx["mediaBlocks"]; exists {
			if mbSlice, ok := mb.([]interface{}); ok {
				mediaBlocks = mbSlice
			}
		}
	}

	// Render all dependent Media blocks
	if len(mediaBlocks) > 1 {
		// Create empty selectors
		var selectors []interface{}
		if creator, ok := result.(interface{ CreateEmptySelectors() []interface{} }); ok {
			selectors = creator.CreateEmptySelectors()
		}

		// Create new Ruleset - this would need actual Ruleset implementation
		if rulesetCreator, ok := context.(interface {
			NewRuleset([]interface{}, []interface{}) interface{}
		}); ok {
			result = rulesetCreator.NewRuleset(selectors, mediaBlocks)

			// Set multiMedia property if the result supports it
			if multiMediaSetter, ok := result.(interface{ SetMultiMedia(bool) }); ok {
				multiMediaSetter.SetMultiMedia(true)
			}

			// Copy visibility info
			if resultWithVis, ok := result.(interface{ CopyVisibilityInfo(interface{}) }); ok {
				if visProvider, ok := n.(VisibilityInfo); ok {
					resultWithVis.CopyVisibilityInfo(visProvider.VisibilityInfo())
				}
			}

			// Set parent relationship
			if parenter, ok := result.(Parenter); ok {
				parenter.SetParent(result, n)
			}
		}
	}

	// Clean up context
	if ctx, ok := context.(map[string]interface{}); ok {
		delete(ctx, "mediaBlocks")
		delete(ctx, "mediaPath")
	}

	return result, nil
}

// EvalNested evaluates the at-rule in a nested context
func (n *NestableAtRulePrototype) EvalNested(context interface{}) (interface{}, error) {
	var mediaPath []interface{}
	var mediaBlocks []interface{}

	// Extract context properties
	if ctx, ok := context.(map[string]interface{}); ok {
		if mp, exists := ctx["mediaPath"]; exists {
			if mpSlice, ok := mp.([]interface{}); ok {
				mediaPath = mpSlice
			}
		}
		if mb, exists := ctx["mediaBlocks"]; exists {
			if mbSlice, ok := mb.([]interface{}); ok {
				mediaBlocks = mbSlice
			}
		}
	}

	// Create path with current rule
	path := append(mediaPath, n)

	// Extract the media-query conditions separated with `,` (OR)
	for i := 0; i < len(path); i++ {
		pathItem := path[i]

		// Check if type matches
		if typeProvider, ok := pathItem.(interface{ GetType() string }); ok {
			if typeProvider.GetType() != n.Type {
				// Remove from mediaBlocks and return self
				if len(mediaBlocks) > i {
					copy(mediaBlocks[i:], mediaBlocks[i+1:])
					mediaBlocks = mediaBlocks[:len(mediaBlocks)-1]
				}
				return n, nil
			}
		}

		// Get features value
		var value interface{}
		if featuresProvider, ok := pathItem.(interface{ GetFeatures() interface{} }); ok {
			features := featuresProvider.GetFeatures()

			// Check if features is a Value type with value property
			if valueProvider, ok := features.(interface{ GetValue() interface{} }); ok {
				value = valueProvider.GetValue()
			} else {
				value = features
			}
		}

		// Convert to array if needed
		if valueSlice, ok := value.([]interface{}); ok {
			path[i] = valueSlice
		} else {
			path[i] = []interface{}{value}
		}
	}

	// Trace all permutations to generate the resulting media-query
	permutations := n.permute(path)

	// Create expressions from permutations
	var expressions []interface{}
	for _, permPath := range permutations {
		if permSlice, ok := permPath.([]interface{}); ok {
			// Convert each fragment to CSS or Anonymous
			var pathFragments []interface{}
			for _, fragment := range permSlice {
				if _, ok := fragment.(interface{ ToCSS() string }); ok {
					pathFragments = append(pathFragments, fragment)
				} else {
					// Create Anonymous node - would need actual Anonymous implementation
					if anonCreator, ok := context.(interface{ NewAnonymous(interface{}) interface{} }); ok {
						pathFragments = append(pathFragments, anonCreator.NewAnonymous(fragment))
					}
				}
			}

			// Insert 'and' between fragments
			var finalFragments []interface{}
			for i, frag := range pathFragments {
				if i > 0 {
					// Insert 'and' - would need actual Anonymous implementation
					if anonCreator, ok := context.(interface{ NewAnonymous(string) interface{} }); ok {
						finalFragments = append(finalFragments, anonCreator.NewAnonymous("and"))
					}
				}
				finalFragments = append(finalFragments, frag)
			}

			// Create Expression - would need actual Expression implementation
			if exprCreator, ok := context.(interface {
				NewExpression([]interface{}) interface{}
			}); ok {
				expressions = append(expressions, exprCreator.NewExpression(finalFragments))
			}
		}
	}

	// Create Value with expressions
	if valueCreator, ok := context.(interface {
		NewValue([]interface{}) interface{}
	}); ok {
		n.Features = valueCreator.NewValue(expressions)

		// Set parent relationship
		if parenter, ok := interface{}(n).(Parenter); ok {
			parenter.SetParent(n.Features, n)
		}
	}

	// Return fake tree-node that doesn't output anything
	if rulesetCreator, ok := context.(interface {
		NewRuleset([]interface{}, []interface{}) interface{}
	}); ok {
		return rulesetCreator.NewRuleset([]interface{}{}, []interface{}{}), nil
	}

	return n, nil
}

// permute generates all permutations of the given array
func (n *NestableAtRulePrototype) permute(arr []interface{}) []interface{} {
	if len(arr) == 0 {
		return []interface{}{}
	}

	if len(arr) == 1 {
		if firstSlice, ok := arr[0].([]interface{}); ok {
			return firstSlice
		}
		return arr
	}

	var result []interface{}

	// Get first element
	var first []interface{}
	if firstSlice, ok := arr[0].([]interface{}); ok {
		first = firstSlice
	} else {
		first = []interface{}{arr[0]}
	}

	// Get rest recursively
	rest := n.permute(arr[1:])

	for _, restItem := range rest {
		for _, firstItem := range first {
			var combined []interface{}
			combined = append(combined, firstItem)

			if restItemSlice, ok := restItem.([]interface{}); ok {
				combined = append(combined, restItemSlice...)
			} else {
				combined = append(combined, restItem)
			}

			result = append(result, combined)
		}
	}

	return result
}

// BubbleSelectors bubbles selectors up in the hierarchy
func (n *NestableAtRulePrototype) BubbleSelectors(selectors interface{}) error {
	if selectors == nil {
		return nil
	}

	// Copy selectors
	var copiedSelectors interface{}
	if copier, ok := selectors.(interface{ Copy() interface{} }); ok {
		copiedSelectors = copier.Copy()
	} else {
		copiedSelectors = selectors
	}

	// Create new Ruleset with copied selectors and first rule
	if len(n.Rules) > 0 {
		if rulesetCreator, ok := interface{}(n).(interface {
			NewRuleset(interface{}, []interface{}) interface{}
		}); ok {
			newRuleset := rulesetCreator.NewRuleset(copiedSelectors, []interface{}{n.Rules[0]})
			n.Rules = []interface{}{newRuleset}

			// Set parent relationship
			if parenter, ok := interface{}(n).(Parenter); ok {
				parenter.SetParent(n.Rules, n)
			}
		}
	}

	return nil
}

// Utility interfaces that would be implemented by the actual node types

// FeatureProvider interface for getting features
type FeatureProvider interface {
	GetFeatures() interface{}
}

// TypeProvider interface for getting node type
type TypeProvider interface {
	GetType() string
}

// ValueProvider interface for getting value
type ValueProvider interface {
	GetValue() interface{}
}

// CSSProvider interface for converting to CSS
type CSSProvider interface {
	ToCSS() string
}

// NodeCreator interface for creating various node types
type NodeCreator interface {
	NewRuleset(selectors interface{}, rules []interface{}) interface{}
	NewValue(expressions []interface{}) interface{}
	NewExpression(fragments []interface{}) interface{}
	NewAnonymous(value interface{}) interface{}
}
