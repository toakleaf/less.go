package less_go

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Anonymous represents an anonymous node in the Less AST
type Anonymous struct {
	*Node
	Value      any
	Index      int
	FileInfo   map[string]any
	MapLines   bool
	RulesetLike bool
	AllowRoot  bool
}

// NewAnonymous creates a new Anonymous instance
func NewAnonymous(value any, index int, fileInfo map[string]any, mapLines bool, rulesetLike bool, visibilityInfo map[string]any) *Anonymous {
	anon := &Anonymous{
		Node:       NewNode(),
		Value:      value,
		Index:      index,
		FileInfo:   fileInfo,
		MapLines:   mapLines,
		RulesetLike: rulesetLike,
		AllowRoot:  true,
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

// Eval evaluates the anonymous value
// If the value is a numeric string, converts it to a Dimension
// Otherwise returns a new Anonymous with the same value
func (a *Anonymous) Eval(context any) (any, error) {
	// Try to convert numeric string values to Dimension objects
	// This enables variables like @base: 8 to participate in math operations
	if str, ok := a.Value.(string); ok {
		// Try to parse as a dimension (number with optional unit)
		if dim := tryParseDimension(str); dim != nil {
			return dim, nil
		}
		// Try to parse as a color
		if color := tryParseColor(str); color != nil {
			return color, nil
		}
	}

	// Match JavaScript: just returns a new Anonymous with the same value
	// Create a new Anonymous instance like JavaScript version
	visibilityInfo := map[string]any{}
	if a.VisibilityBlocks != nil {
		visibilityInfo["visibilityBlocks"] = *a.VisibilityBlocks
	}
	if a.NodeVisible != nil {
		visibilityInfo["nodeVisible"] = *a.NodeVisible
	}
	return NewAnonymous(a.Value, a.Index, a.FileInfo, a.MapLines, a.RulesetLike, visibilityInfo), nil
}

// tryParseDimension attempts to parse a string as a numeric dimension with optional unit
// Examples: "8" -> Dimension(8), "10px" -> Dimension(10, "px"), "1.5em" -> Dimension(1.5, "em")
func tryParseDimension(s string) *Dimension {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Pattern: optional sign, digits, optional decimal part, optional unit
	// Example: "-10.5px" or "8" or ".5em"
	// Using a regex to parse numeric + unit
	re := regexp.MustCompile(`^([-+]?\d*\.?\d+(?:[eE][-+]?\d+)?)\s*([a-zA-Z%]*)\s*$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil || len(matches) < 2 {
		return nil
	}

	numStr := matches[1]
	unitStr := matches[2]

	// Try to parse the numeric part
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil
	}

	// Create dimension with optional unit
	dim, err := NewDimension(num, unitStr)
	if err != nil {
		return nil
	}

	return dim
}

// tryParseColor attempts to parse a string as a color
// Examples: "#fff", "#ffffff", "red" (keyword)
func tryParseColor(s string) *Color {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Try hex color
	if strings.HasPrefix(s, "#") {
		// Simple hex color validation
		hexPart := s[1:]
		if len(hexPart) == 3 || len(hexPart) == 6 {
			// Try to create color from hex
			color := NewColor(s, 1.0, s)
			if color != nil {
				return color
			}
		}
	}

	// Could add keyword color detection here if needed
	// For now, we'll just handle hex colors

	return nil
}

// Compare compares two nodes
func (a *Anonymous) Compare(other any) any {
	if other == nil {
		return nil
	}

	// Check if other has a ToCSS method (like JavaScript version checks other.toCSS)
	// JavaScript: return other.toCSS && this.toCSS() === other.toCSS() ? 0 : undefined;
	if cssable, ok := other.(interface{ ToCSS(any) string }); ok {
		if a.ToCSS(nil) == cssable.ToCSS(nil) {
			return 0
		}
	}
	
	return nil
}

// IsRulesetLike returns whether the node is ruleset-like
func (a *Anonymous) IsRulesetLike() bool {
	return a.RulesetLike
}

// Operate performs mathematical operations on Anonymous values
// This allows variables like @z: 11; to participate in mathematical expressions
func (a *Anonymous) Operate(context any, op string, other any) any {
	// Try to convert this Anonymous to a Dimension for mathematical operations
	if a.Value != nil {
		if str, ok := a.Value.(string); ok {
			// Try to parse as a number
			if dim, err := NewDimension(str, ""); err == nil {
				// Successfully parsed as dimension, delegate to dimension's operate method
				if otherDim, ok := other.(*Dimension); ok {
					return dim.Operate(context, op, otherDim)
				}
				// If other is also Anonymous, try to convert it too
				if otherAnon, ok := other.(*Anonymous); ok {
					if otherStr, ok := otherAnon.Value.(string); ok {
						if otherDimension, err := NewDimension(otherStr, ""); err == nil {
							return dim.Operate(context, op, otherDimension)
						}
					}
				}
			}
		}
		// If this anonymous represents a number value
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
	
	// If we can't convert to dimension, return a new operation node
	return NewOperation(op, []any{a, other}, false)
}

// GenCSS generates CSS representation
func (a *Anonymous) GenCSS(context any, output *CSSOutput) {
	// Set visibility based on value's truthiness like JavaScript version
	// In JS: this.nodeVisible = Boolean(this.value);
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
		// Check if the value implements CSSGenerator
		if generator, ok := a.Value.(CSSGenerator); ok {
			generator.GenCSS(context, output)
		} else if a.Value != nil {
			// For simple values like strings, add directly
			// JavaScript passes mapLines as 4th param, but Go's Add only takes 3 params
			output.Add(a.Value, a.FileInfo, a.Index)
		}
	}
}

// ToCSS generates CSS string representation
func (a *Anonymous) ToCSS(context any) string {
	// Use GenCSS internally
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

// IsVisible returns whether the node is visible for spacing purposes
// This is used by Ruleset.GenCSS to determine if a newline should be added after this node
func (a *Anonymous) IsVisible() bool {
	// Anonymous nodes are visible if they have a non-empty value
	// This matches the logic in GenCSS where nodeVisible is set based on value truthiness
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

// CopyVisibilityInfo copies visibility information from another node
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