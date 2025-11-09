package less_go

import (
	"fmt"
)

// SelectorList represents a list of selectors separated by commas
// This is used for parenthesized selector lists like :is(.a, .b, .c)
// where the list contains Selector nodes and Anonymous comma nodes
type SelectorList struct {
	*Node
	Selectors []any // Contains Selector nodes and Anonymous nodes (for commas)
}

// NewSelectorList creates a new SelectorList instance
func NewSelectorList(selectors []any) *SelectorList {
	return &SelectorList{
		Node:      NewNode(),
		Selectors: selectors,
	}
}

// Type returns the type of the node
func (sl *SelectorList) Type() string {
	return "SelectorList"
}

// GetType returns the type of the node for visitor pattern consistency
func (sl *SelectorList) GetType() string {
	return "SelectorList"
}

// GenCSS generates CSS representation
func (sl *SelectorList) GenCSS(context any, output *CSSOutput) {
	for i, item := range sl.Selectors {
		if item != nil {
			if gen, ok := item.(interface{ GenCSS(any, *CSSOutput) }); ok {
				gen.GenCSS(context, output)
			} else if str, ok := item.(string); ok {
				output.Add(str, nil, nil)
			}
		}
		// Add space after comma for readability
		if i < len(sl.Selectors)-1 {
			if _, ok := item.(*Anonymous); ok {
				output.Add(" ", nil, nil)
			}
		}
	}
}

// ToCSS generates a CSS string representation
func (sl *SelectorList) ToCSS(context any) string {
	strs := []string{}
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			if strChunk, ok := chunk.(string); ok {
				strs = append(strs, strChunk)
			} else {
				strs = append(strs, fmt.Sprintf("%v", chunk))
			}
		},
		IsEmpty: func() bool {
			return len(strs) == 0
		},
	}
	sl.GenCSS(context, output)

	result := ""
	for _, s := range strs {
		result += s
	}
	return result
}

// Eval evaluates the selector list and returns a new list with evaluated selectors
func (sl *SelectorList) Eval(context any) any {
	evaluatedSelectors := make([]any, len(sl.Selectors))
	for i, item := range sl.Selectors {
		if item == nil {
			evaluatedSelectors[i] = nil
			continue
		}

		// Try to evaluate if the item has an Eval method
		if evalItem, ok := item.(interface{ Eval(any) any }); ok {
			evaluatedSelectors[i] = evalItem.Eval(context)
		} else if evalItem, ok := item.(interface{ Eval(any) (any, error) }); ok {
			result, _ := evalItem.Eval(context)
			if result != nil {
				evaluatedSelectors[i] = result
			} else {
				evaluatedSelectors[i] = item
			}
		} else {
			evaluatedSelectors[i] = item
		}
	}

	return NewSelectorList(evaluatedSelectors)
}
