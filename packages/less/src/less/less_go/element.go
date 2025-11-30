package less_go

import (
	"fmt"
	"strings"
)

type Element struct {
	*Node
	Combinator  *Combinator
	Value       any
	IsVariable  bool
}

func NewElement(combinator any, value any, isVariable bool, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) *Element {
	var comb *Combinator
	switch c := combinator.(type) {
	case *Combinator:
		if c == nil {
			comb = NewCombinator("")
		} else {
			comb = c
		}
	case string:
		comb = NewCombinator(c)
	default:
		// Handle nil or unexpected types gracefully
		comb = NewCombinator("")
	}

	var val any
	switch v := value.(type) {
	case string:
		val = Intern(strings.TrimSpace(v))
	case nil:
		val = ""
	case byte:
		// Convert byte to string character (e.g., byte(38) -> "&")
		// Note: byte is an alias for uint8 in Go
		val = string(v)
	case bool:
		// JavaScript converts false to ""
		if !v {
			val = ""
		} else {
			val = v
		}
	case int:
		// JavaScript converts 0 to ""
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	case int64:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	case float32:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	case float64:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	default:
		// For objects and other types, preserve as is
		val = v
	}

	e := GetElementFromPool()
	e.Node = NewNode()
	e.Combinator = comb
	e.Value = val
	e.IsVariable = isVariable

	e.Index = index
	if currentFileInfo != nil {
		e.SetFileInfo(currentFileInfo)
	} else {
		e.SetFileInfo(make(map[string]any))
	}
	e.CopyVisibilityInfo(visibilityInfo)
	e.SetParent(comb, e.Node)

	return e
}

// NewElementWithArena creates an Element using arena allocation when available.
// This avoids sync.Pool mutex overhead for single-threaded compilation.
func NewElementWithArena(arena *NodeArena, combinator any, value any, isVariable bool, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) *Element {
	var comb *Combinator
	switch c := combinator.(type) {
	case *Combinator:
		if c == nil {
			comb = NewCombinator("")
		} else {
			comb = c
		}
	case string:
		comb = NewCombinator(c)
	default:
		comb = NewCombinator("")
	}

	var val any
	switch v := value.(type) {
	case string:
		val = Intern(strings.TrimSpace(v))
	case nil:
		val = ""
	case byte:
		val = string(v)
	case bool:
		if !v {
			val = ""
		} else {
			val = v
		}
	case int:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	case int64:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	case float32:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	case float64:
		if v == 0 {
			val = ""
		} else {
			val = v
		}
	default:
		val = v
	}

	e := GetElementFromArena(arena)
	e.Node = GetNodeFromArena(arena)
	e.Combinator = comb
	e.Value = val
	e.IsVariable = isVariable

	e.Index = index
	if currentFileInfo != nil {
		e.SetFileInfo(currentFileInfo)
	} else {
		e.SetFileInfo(make(map[string]any))
	}
	e.CopyVisibilityInfo(visibilityInfo)
	e.SetParent(comb, e.Node)

	return e
}

func (e *Element) Type() string {
	return "Element"
}

func (e *Element) GetType() string {
	return "Element"
}

func (e *Element) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		// Visit the combinator and handle its return value
		if e.Combinator != nil {
			if visitedComb := v.Visit(e.Combinator); visitedComb != nil {
				if comb, ok := visitedComb.(*Combinator); ok {
					e.Combinator = comb
				}
			}
		}

		// Visit the value only if it's an object (matching JavaScript's typeof value === 'object')
		// In JavaScript, objects include arrays, functions, and actual objects, but not primitives
		if e.Value != nil {
			switch e.Value.(type) {
			case string, bool, int, int64, float32, float64:
				// Don't visit primitive types
			default:
				// Visit objects (including structs, maps, slices, etc.)
				if visited := v.Visit(e.Value); visited != nil {
					e.Value = visited
				}
			}
		}
	}
}

func (e *Element) Eval(context any) (any, error) {
	var evaluatedValue any = e.Value
	wasInterpolated := false

	// Match JavaScript logic: this.value.eval ? this.value.eval(context) : this.value
	if e.Value != nil {
		if evalValue, ok := e.Value.(interface{ Eval(any) (any, error) }); ok {
			evaluated, err := evalValue.Eval(context)
			if err != nil {
				return nil, err
			}
			evaluatedValue = evaluated
		} else if evalValue, ok := e.Value.(interface{ Eval(any) any }); ok {
			evaluatedValue = evalValue.Eval(context)
		} else if strValue, ok := e.Value.(string); ok {
			// Check if string contains variable interpolation
			if strings.Contains(strValue, "@{") {
				// Create a Quoted node to handle interpolation
				quoted := NewQuoted("", strValue, true, e.GetIndex(), e.FileInfo())
				evaluated, err := quoted.Eval(context)
				if err != nil {
					return nil, err
				}
				// Extract the string value from the evaluated Quoted node
				if quotedResult, ok := evaluated.(*Quoted); ok {
					evaluatedValue = quotedResult.value
				} else {
					evaluatedValue = evaluated
				}
				// Mark that interpolation occurred so we can set IsVariable flag
				wasInterpolated = true
			}
		}
	}

	// Unwrap Paren nodes that contain simple values (fixes double parentheses in selectors like :nth-child(@{num}))
	// This recursively unwraps nested Paren nodes until we reach a non-Paren value
	for {
		paren, ok := evaluatedValue.(*Paren)
		if !ok || paren == nil {
			break
		}

		innerValue := paren.Value
		if innerValue == nil {
			break
		}

		// Check if the inner value is another Paren (for recursive unwrapping)
		if _, isNestedParen := innerValue.(*Paren); isNestedParen {
			evaluatedValue = innerValue
			continue
		}

		// Check if it's a simple value type or a Quoted string - these don't need explicit Paren wrapping in selectors
		switch v := innerValue.(type) {
		case *Dimension, *Keyword, *Anonymous, string, int, float64:
			evaluatedValue = innerValue
		case *Quoted:
			// For Quoted nodes, unwrap to the string value if it's escaped (unquoted)
			if v.GetEscaped() {
				evaluatedValue = v.GetValue()
			} else {
				evaluatedValue = innerValue
			}
		default:
			// For complex values (selectors, expressions, etc.), keep the Paren
			// This preserves important grouping in cases like :not(.foo&)
			goto endUnwrap
		}
		break
	}
	endUnwrap:

	// Handle potential nil Node
	index := 0
	if e.Node != nil {
		index = e.GetIndex()
	}

	fileInfo := make(map[string]any)
	if e.Node != nil {
		fileInfo = e.FileInfo()
	}

	visibilityInfo := make(map[string]any)
	if e.Node != nil {
		visibilityInfo = e.VisibilityInfo()
	}

	// Set IsVariable to true if interpolation occurred, otherwise use original value
	// This allows the ruleset evaluation to detect interpolated selectors and re-parse them
	isVariable := e.IsVariable || wasInterpolated

	newElement := NewElement(
		e.Combinator,
		evaluatedValue,
		isVariable,
		index,
		fileInfo,
		visibilityInfo,
	)

	return newElement, nil
}

func (e *Element) Clone() *Element {
	return NewElement(
		e.Combinator,
		e.Value,
		e.IsVariable,
		e.GetIndex(),
		e.FileInfo(),
		e.VisibilityInfo(),
	)
}

func (e *Element) GenCSS(context any, output *CSSOutput) {
	if output == nil {
		return
	}
	output.Add(e.ToCSS(context), e.FileInfo(), e.GetIndex())
}

func (e *Element) ToCSS(context any) string {
	// Match JavaScript logic: context = context || {}
	ctx := make(map[string]any)
	if c, ok := context.(map[string]any); ok {
		ctx = c
	}

	var valueCSS string
	value := e.Value
	firstSelector := false
	if fs, exists := ctx["firstSelector"].(bool); exists {
		firstSelector = fs
	}

	// If value is a Paren, set firstSelector to true
	if _, ok := value.(*Paren); ok {
		// selector in parens should not be affected by outer selector
		// flags (breaks only interpolated selectors - see #1973)
		ctx["firstSelector"] = true
	}

	// Convert value to CSS
	if value == nil {
		valueCSS = ""
	} else if cssValue, ok := value.(interface{ ToCSS(any) string }); ok {
		valueCSS = cssValue.ToCSS(ctx)
	} else if strValue, ok := value.(string); ok {
		valueCSS = strValue
	} else {
		valueCSS = fmt.Sprintf("%v", value)
	}

	// Restore firstSelector
	ctx["firstSelector"] = firstSelector

	// Handle empty value with & combinator
	if valueCSS == "" && e.Combinator != nil && len(e.Combinator.Value) > 0 && e.Combinator.Value[0] == '&' {
		return ""
	}

	// Get combinator CSS using the same pattern as JavaScript Node.toCSS
	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	if e.Combinator != nil {
		output := &CSSOutput{
			Add: func(chunk any, fileInfo any, index any) {
				if chunk != nil {
					if strChunk, ok := chunk.(string); ok {
						builder.WriteString(strChunk)
					} else {
						builder.WriteString(fmt.Sprintf("%v", chunk))
					}
				}
			},
			IsEmpty: func() bool {
				return builder.Len() == 0
			},
		}
		e.Combinator.GenCSS(ctx, output)
	}
	builder.WriteString(valueCSS)

	return builder.String()
} 