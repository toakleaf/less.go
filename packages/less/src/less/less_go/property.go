package less_go

import (
	"fmt"
)

// Property represents a property node in the Less AST
type Property struct {
	*Node
	name       string
	evaluating bool
}

// NewProperty creates a new Property instance
func NewProperty(name string, index int, fileInfo map[string]any) *Property {
	p := &Property{
		Node: NewNode(),
		name: Intern(name),
	}
	p.Index = index
	p.SetFileInfo(fileInfo)
	return p
}

// Type returns the type of the node
func (p *Property) Type() string {
	return "Property"
}

// GetType returns the type of the node
func (p *Property) GetType() string {
	return "Property"
}

// GetName returns the property name
func (p *Property) GetName() string {
	return p.name
}

// Eval evaluates the property in the given context
func (p *Property) Eval(context any) (any, error) {
	if p.evaluating {
		return nil, fmt.Errorf("Name: recursive property reference for %s (index: %d, filename: %s)",
			p.name, p.GetIndex(), p.FileInfo()["filename"])
	}

	p.evaluating = true
	defer func() { p.evaluating = false }()

	var frames []ParserFrame
	if evalCtx, ok := context.(interface{ GetFrames() []ParserFrame }); ok {
		frames = evalCtx.GetFrames()
	} else if ctx, ok := context.(map[string]any); ok {
		if framesAny, exists := ctx["frames"]; exists {
			if frameSlice, ok := framesAny.([]any); ok {
				frames = make([]ParserFrame, 0, len(frameSlice))
				for _, f := range frameSlice {
					if frame, ok := f.(ParserFrame); ok {
						frames = append(frames, frame)
					} else if frameMap, ok := f.(map[string]any); ok {
						// Handle test frames that are map[string]any with property function
						if propFunc, ok := frameMap["property"].(func(string) []any); ok {
							frames = append(frames, &mapFrame{
								propertyFunc: propFunc,
							})
						}
					}
				}
			}
		}
	}

	property := p.find(frames, func(frame ParserFrame) any {
		vArr := frame.Property(p.name)
		if vArr == nil || len(vArr) == 0 {
			return nil
		}

		for i, v := range vArr {
			if decl, ok := v.(*Declaration); ok {
				vArr[i] = decl
			} else if declMap, ok := v.(map[string]any); ok {
				merge, _ := declMap["merge"].(bool)
				index, _ := declMap["index"].(int)
				currentFileInfo, _ := declMap["currentFileInfo"].(map[string]any)
				inline, _ := declMap["inline"].(bool)
				
				newDecl, _ := NewDeclaration(
					declMap["name"],
					declMap["value"],
					declMap["important"],
					merge,
					index,
					currentFileInfo,
					inline,
					declMap["variable"],
				)
				vArr[i] = newDecl
			}
		}

		if len(vArr) > 1 {
			vArr = p.mergeRules(vArr)
		}

		lastItem := vArr[len(vArr)-1]
		var value any
		var important string

		switch decl := lastItem.(type) {
		case *Declaration:
			value = decl.Value
			important = decl.important
		case interface{ Value() any }:
			value = decl.Value()
		default:
			value = lastItem
		}

		if important != "" {
			if ctx, ok := context.(map[string]any); ok {
				if importantScope, ok := ctx["importantScope"].([]any); ok && len(importantScope) > 0 {
					if lastScope, ok := importantScope[len(importantScope)-1].(map[string]any); ok {
						lastScope["important"] = important
					}
				}
			}
		}

		if valueNode, ok := value.(Node); ok {
			result := valueNode.Eval(context)
			return result
		}
		if evalable, ok := value.(interface{ Eval(any) (any, error) }); ok {
			result, err := evalable.Eval(context)
			if err != nil {
				return nil
			}
			return result
		}

		return value
	})

	if property != nil {
		p.evaluating = false
		return property, nil
	}
	return nil, fmt.Errorf("Name: property '%s' is undefined (index: %d, filename: %s)",
		p.name, p.GetIndex(), p.FileInfo()["filename"])
}

// find searches through frames for the first element that satisfies the predicate
func (p *Property) find(frames []ParserFrame, predicate func(ParserFrame) any) any {
	for _, frame := range frames {
		if result := predicate(frame); result != nil {
			return result
		}
	}
	return nil
}

// Find searches through an array for the first element that satisfies the predicate
func (p *Property) Find(arr []any, predicate func(any) any) any {
	for _, item := range arr {
		if result := predicate(item); result != nil {
			return result
		}
	}
	return nil
}

// mapFrame adapter to implement ParserFrame for test frames that are map[string]any
type mapFrame struct {
	propertyFunc func(string) []any
}

func (m *mapFrame) Variable(name string) map[string]any {
	// Property frames don't have variables
	return nil
}

func (m *mapFrame) Property(name string) []any {
	if m.propertyFunc != nil {
		return m.propertyFunc(name)
	}
	return nil
}

func (p *Property) mergeRules(rules []any) []any {
	if rules == nil || len(rules) == 0 {
		return rules
	}

	groups := make(map[string]*[]int)
	groupsArr := []*[]int{}
	for i := 0; i < len(rules); i++ {
		if decl, ok := rules[i].(*Declaration); ok && decl.GetMerge() != nil && decl.GetMerge() != false {
			key := decl.GetName()
			if _, exists := groups[key]; !exists {
				group := []int{}
				groups[key] = &group
				groupsArr = append(groupsArr, &group)
			}
			// Append to the group
			*groups[key] = append(*groups[key], i)
		}
	}
	for _, indicesPtr := range groupsArr {
		indices := *indicesPtr
		if len(indices) > 1 {
			result := rules[indices[0]].(*Declaration)
			space := []any{}
			comma := []any{}
			for _, idx := range indices {
				decl := rules[idx].(*Declaration)
				if decl.MergeType() == "+" && len(space) > 0 {
					if expr, err := NewExpression(space, false); err == nil {
						comma = append(comma, expr)
					}
					space = []any{}
				}
				space = append(space, decl.Value)
				if decl.GetImportant() {
					result.important = decl.important
				}
			}
			if len(space) > 0 {
				if expr, err := NewExpression(space, false); err == nil {
					comma = append(comma, expr)
				}
			}
			if mergedValue, err := NewValue(comma); err == nil {
				result.Value = mergedValue
			}
		}
	}
	if len(groupsArr) > 0 {
		keep := make(map[int]bool)
		for i := range rules {
			keep[i] = true
		}
		for _, indicesPtr := range groupsArr {
			indices := *indicesPtr
			if len(indices) > 1 {
				for i := 1; i < len(indices); i++ {
					keep[indices[i]] = false
				}
			}
		}
		newRules := make([]any, 0, len(rules))
		for i, rule := range rules {
			if keep[i] {
				newRules = append(newRules, rule)
			}
		}
		return newRules
	}

	return rules
} 