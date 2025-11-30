package less_go

import (
	"fmt"
	"strings"
)

type Anonymous struct {
	*Node
	Value      any
	Index      int
	FileInfo   map[string]any
	MapLines   bool
	RulesetLike bool
	AllowRoot  bool
}

func NewAnonymous(value any, index int, fileInfo map[string]any, mapLines bool, rulesetLike bool, visibilityInfo map[string]any) *Anonymous {
	node := NewNode()
	node.TypeIndex = GetTypeIndexForNodeType("Anonymous")

	anon := &Anonymous{
		Node:        node,
		Value:       value,
		Index:       index,
		FileInfo:    fileInfo,
		MapLines:    mapLines,
		RulesetLike: rulesetLike,
		AllowRoot:   true,
	}
	if visibilityInfo != nil {
		if blocks, ok := visibilityInfo["visibilityBlocks"].(int); ok {
			anon.VisibilityBlocks = &blocks
		}
		if visible, ok := visibilityInfo["nodeVisible"].(bool); ok {
			anon.NodeVisible = &visible
		}
	}
	return anon
}

func (a *Anonymous) GetType() string {
	return "Anonymous"
}

func (a *Anonymous) GetValue() any {
	return a.Value
}

func (a *Anonymous) GetTypeIndex() int {
	if a.Node != nil && a.Node.TypeIndex != 0 {
		return a.Node.TypeIndex
	}
	return GetTypeIndexForNodeType("Anonymous")
}

func (a *Anonymous) Eval(context any) (any, error) {
	visibilityInfo := map[string]any{}
	if a.VisibilityBlocks != nil {
		visibilityInfo["visibilityBlocks"] = *a.VisibilityBlocks
	}
	if a.NodeVisible != nil {
		visibilityInfo["nodeVisible"] = *a.NodeVisible
	}
	return NewAnonymous(a.Value, a.Index, a.FileInfo, a.MapLines, a.RulesetLike, visibilityInfo), nil
}

func (a *Anonymous) Compare(other any) any {
	if other == nil {
		return nil
	}

	if cssable, ok := other.(interface{ ToCSS(any) string }); ok {
		if a.ToCSS(nil) == cssable.ToCSS(nil) {
			return 0
		}
	}
	
	return nil
}

func (a *Anonymous) IsRulesetLike() bool {
	return a.RulesetLike
}

// Operate allows Anonymous values (like variables @z: 11) to participate in math expressions
func (a *Anonymous) Operate(context any, op string, other any) any {
	if a.Value != nil {
		if str, ok := a.Value.(string); ok {
			if dim, err := NewDimension(str, ""); err == nil {
				if otherDim, ok := other.(*Dimension); ok {
					return dim.Operate(context, op, otherDim)
				}
				if otherAnon, ok := other.(*Anonymous); ok {
					if otherStr, ok := otherAnon.Value.(string); ok {
						if otherDimension, err := NewDimension(otherStr, ""); err == nil {
							return dim.Operate(context, op, otherDimension)
						}
					}
				}
			}
		}
		if num, ok := a.Value.(float64); ok {
			dim, _ := NewDimension(num, "")
			if otherDim, ok := other.(*Dimension); ok {
				return dim.Operate(context, op, otherDim)
			}
		}
		if num, ok := a.Value.(int); ok {
			dim, _ := NewDimension(float64(num), "")
			if otherDim, ok := other.(*Dimension); ok {
				return dim.Operate(context, op, otherDim)
			}
		}
	}

	return NewOperation(op, []any{a, other}, false)
}

func (a *Anonymous) GenCSS(context any, output *CSSOutput) {
	visible := false
	if a.Value != nil {
		switch v := a.Value.(type) {
		case string:
			visible = v != ""
		case bool:
			visible = v
		case int, int64, float64:
			visible = true
		default:
			visible = true
		}
	}
	a.NodeVisible = &visible

	if *a.NodeVisible {
		if generator, ok := a.Value.(CSSGenerator); ok {
			generator.GenCSS(context, output)
		} else if a.Value != nil {
			output.Add(a.Value, a.FileInfo, a.Index)
		}
	}
}

func (a *Anonymous) ToCSS(context any) string {
	var chunks []string
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			if chunk != nil {
				chunks = append(chunks, fmt.Sprintf("%v", chunk))
			}
		},
		IsEmpty: func() bool {
			return len(chunks) == 0
		},
	}
	a.GenCSS(context, output)
	result := strings.Join(chunks, "")
	return result
}

func (a *Anonymous) IsVisible() bool {
	if a.Value != nil {
		switch v := a.Value.(type) {
		case string:
			return v != ""
		case bool:
			return v
		case int, int64, float64:
			return true
		default:
			return true
		}
	}
	return false
}

func (a *Anonymous) CopyVisibilityInfo(info map[string]any) {
	if info == nil {
		return
	}
	if blocks, ok := info["visibilityBlocks"].(int); ok {
		a.VisibilityBlocks = &blocks
	}
	if visible, ok := info["nodeVisible"].(bool); ok {
		a.NodeVisible = &visible
	}
} 