package less_go

import (
	"fmt"
	"strings"
)

// emptyToCSSContext is a shared read-only context for ToCSS calls that don't need context.
// This avoids allocating a new map on every ToCSS call.
// IMPORTANT: This map must NEVER be modified. ToCSS implementations should only read from it.
var emptyToCSSContext = map[string]any{}

// isVarNameChar checks if a character is valid in a variable/property name: [a-zA-Z0-9_-]
func isVarNameChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_' || c == '-'
}

// containsBracedVar checks if a string contains @{varname} pattern (for ContainsVariables)
func containsBracedVar(s string) bool {
	for i := 0; i < len(s)-2; i++ {
		if s[i] == '@' && s[i+1] == '{' {
			// Look for closing } with valid var name between
			for j := i + 2; j < len(s); j++ {
				if s[j] == '}' {
					if j > i+2 { // Must have at least one char in name
						return true
					}
					break
				}
				if !isVarNameChar(s[j]) {
					break
				}
			}
		}
	}
	return false
}

// interpolateBracedVars replaces @{varname} patterns with their values
// Returns the new string and true if any replacements were made
func interpolateBracedVars(s string, evalVar func(name string) (string, error)) (string, bool, error) {
	var result strings.Builder
	result.Grow(len(s))

	changed := false
	i := 0
	for i < len(s) {
		// Look for @{ pattern
		if i+2 < len(s) && s[i] == '@' && s[i+1] == '{' {
			// Find end of variable name
			end := -1
			for j := i + 2; j < len(s); j++ {
				if s[j] == '}' {
					end = j
					break
				}
				if !isVarNameChar(s[j]) {
					break
				}
			}

			if end != -1 && end > i+2 { // Valid variable pattern found
				varName := s[i+2 : end]
				replacement, err := evalVar(varName)
				if err != nil {
					return s, false, err
				}
				result.WriteString(replacement)
				i = end + 1
				changed = true
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String(), changed, nil
}

// interpolateBareVars replaces @varname patterns with their values (for escaped content)
// Returns the new string and true if any replacements were made
func interpolateBareVars(s string, evalVar func(name string) (string, error)) (string, bool, error) {
	var result strings.Builder
	result.Grow(len(s))

	changed := false
	i := 0
	for i < len(s) {
		// Look for @ followed by a valid var name char (but not @{ which is handled separately)
		if i+1 < len(s) && s[i] == '@' && s[i+1] != '{' && isVarNameChar(s[i+1]) {
			// Find end of variable name
			end := i + 1
			for end < len(s) && isVarNameChar(s[end]) {
				end++
			}

			varName := s[i+1 : end]
			replacement, err := evalVar(varName)
			if err != nil {
				return s, false, err
			}
			result.WriteString(replacement)
			i = end
			changed = true
			continue
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String(), changed, nil
}

// interpolateBracedProps replaces ${propname} patterns with their values
// Returns the new string and true if any replacements were made
func interpolateBracedProps(s string, evalProp func(name string) (string, error)) (string, bool, error) {
	var result strings.Builder
	result.Grow(len(s))

	changed := false
	i := 0
	for i < len(s) {
		// Look for ${ pattern
		if i+2 < len(s) && s[i] == '$' && s[i+1] == '{' {
			// Find end of property name
			end := -1
			for j := i + 2; j < len(s); j++ {
				if s[j] == '}' {
					end = j
					break
				}
				if !isVarNameChar(s[j]) {
					break
				}
			}

			if end != -1 && end > i+2 { // Valid property pattern found
				propName := s[i+2 : end]
				replacement, err := evalProp(propName)
				if err != nil {
					return s, false, err
				}
				result.WriteString(replacement)
				i = end + 1
				changed = true
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String(), changed, nil
}

// Quoted represents a quoted string in the Less AST
type Quoted struct {
	*Node
	escaped    bool
	value      string
	quote      string
	_index     int
	_fileInfo  map[string]any
	allowRoot  bool
}

// NewQuoted creates a new Quoted instance
func NewQuoted(str string, content string, escaped bool, index int, currentFileInfo map[string]any) *Quoted {
	// In JS, escaped defaults to true only when undefined
	// But when explicitly set to false, it should remain false
	if content == "" {
		content = ""
	}
	
	// Initialize fileInfo if nil
	if currentFileInfo == nil {
		currentFileInfo = make(map[string]any)
	}
	
	// Handle empty quote string safely
	var quote string
	if char, ok := SafeStringIndex(str, 0); ok {
		quote = Intern(string(char))
	} else {
		quote = ""
	}

	return &Quoted{
		Node:          NewNode(),
		escaped:       escaped,
		value:         content,
		quote:         quote,
		_index:        index,
		_fileInfo:     currentFileInfo,
		allowRoot:     escaped,
	}
}

// Type returns the node type
func (q *Quoted) Type() string {
	return "Quoted"
}

// GetType returns the node type
func (q *Quoted) GetType() string {
	return "Quoted"
}

// GetIndex returns the node's index
func (q *Quoted) GetIndex() int {
	return q._index
}

// GetValue returns the raw string value of the quoted string
func (q *Quoted) GetValue() string {
	return q.value
}

// GetQuote returns the quote character used
func (q *Quoted) GetQuote() string {
	return q.quote
}

// GetEscaped returns whether the quoted string is escaped
func (q *Quoted) GetEscaped() bool {
	return q.escaped
}

// FileInfo returns the node's file information
func (q *Quoted) FileInfo() map[string]any {
	return q._fileInfo
}

// GenCSS generates CSS representation
func (q *Quoted) GenCSS(context any, output *CSSOutput) {
	if !q.escaped {
		output.Add(q.quote, q.FileInfo(), q.GetIndex())
	}
	output.Add(q.value, nil, nil)
	if !q.escaped {
		output.Add(q.quote, nil, nil)
	}
}

// ToCSS generates CSS string representation
func (q *Quoted) ToCSS(context any) string {
	var strs []string
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			// Optimize: use type switch to avoid fmt.Sprintf allocation for common types
			switch v := chunk.(type) {
			case string:
				strs = append(strs, v)
			case fmt.Stringer:
				strs = append(strs, v.String())
			default:
				strs = append(strs, fmt.Sprintf("%v", chunk))
			}
		},
		IsEmpty: func() bool {
			return len(strs) == 0
		},
	}
	q.GenCSS(context, output)
	return strings.Join(strs, "")
}

// ContainsVariables checks if the quoted string contains variable interpolations
func (q *Quoted) ContainsVariables() bool {
	return containsBracedVar(q.value)
}

// Eval evaluates the quoted string, replacing variables and properties
func (q *Quoted) Eval(context any) (any, error) {
	value := q.value

	// Get frames from context safely - handle both EvalContext and map[string]any
	var frames []ParserFrame
	if evalCtx, ok := context.(interface{ GetFrames() []ParserFrame }); ok {
		frames = evalCtx.GetFrames()
	} else if ctx, ok := context.(map[string]any); ok {
		// Extract frames from map context
		if framesAny, exists := ctx["frames"]; exists {
			if frameSlice, ok := framesAny.([]any); ok {
				frames = make([]ParserFrame, 0, len(frameSlice))
				for _, f := range frameSlice {
					if frame, ok := f.(ParserFrame); ok {
						frames = append(frames, frame)
					}
				}
			}
		}
	}
	if frames == nil {
		frames = make([]ParserFrame, 0) // Provide empty frames if none available
	}

	// variableReplacement handles @{name} or @name syntax
	variableReplacement := func(name string) (string, error) {
		// First try direct frame access for the test case
		for _, frame := range frames {
			if varResult := frame.Variable("@" + name); varResult != nil {
				// Extract value and return map to pool immediately
				val, hasValue := varResult["value"]
				PutVariableResultMap(varResult)

				if hasValue {
					var result string
					if quoted, ok := val.(*Quoted); ok {
						result = quoted.value
					} else if anon, ok := val.(*Anonymous); ok {
						// Handle Anonymous objects by getting their value
						if str, ok := anon.Value.(string); ok {
							result = str
						} else if cssable, ok := anon.Value.(interface{ ToCSS(any) string }); ok {
							result = cssable.ToCSS(emptyToCSSContext)
						} else {
							result = fmt.Sprintf("%v", anon.Value)
						}
					} else if value, ok := val.(*Value); ok {
						// Handle Value objects by evaluating them
						evaluated, err := value.Eval(context)
						if err != nil {
							result = fmt.Sprintf("%v", val)
						} else if anon, ok := evaluated.(*Anonymous); ok {
							// Handle Anonymous results from Value eval
							if str, ok := anon.Value.(string); ok {
								result = str
							} else if cssable, ok := anon.Value.(interface{ ToCSS(any) string }); ok {
								result = cssable.ToCSS(emptyToCSSContext)
							} else {
								result = fmt.Sprintf("%v", anon.Value)
							}
						} else if quoted, ok := evaluated.(*Quoted); ok {
							// Handle Quoted results from Value eval
							result = quoted.value
						} else if expr, ok := evaluated.(*Expression); ok {
							// Handle Expression results from Value eval
							// Use ToCSS to get the full expression (e.g., "Arial, Verdana, San-Serif")
							result = expr.ToCSS(emptyToCSSContext)
						} else if evalValue, ok := evaluated.(*Value); ok {
							// Handle Value results (e.g., comma-separated lists like "Arial, Verdana, San-Serif")
							// Use ToCSS to get the full value list
							result = evalValue.ToCSS(emptyToCSSContext)
						} else if cssable, ok := evaluated.(interface{ ToCSS(any) string }); ok {
							result = cssable.ToCSS(emptyToCSSContext)
						} else {
							result = fmt.Sprintf("%v", evaluated)
						}
					} else if cssable, ok := val.(interface{ ToCSS(any) string }); ok {
						result = cssable.ToCSS(emptyToCSSContext)
					} else {
						result = fmt.Sprintf("%v", val)
					}
					return result, nil
				}
			}
		}

		// Fall back to Variable eval if frames don't have it
		v := NewVariable("@"+name, q.GetIndex(), q.FileInfo())
		result, err := v.Eval(context)
		if err != nil {
			return "", fmt.Errorf("variable @%s is undefined", name)
		}

		if quoted, ok := result.(*Quoted); ok {
			return quoted.value, nil
		}

		if cssable, ok := result.(interface{ ToCSS(any) string }); ok {
			return cssable.ToCSS(emptyToCSSContext), nil
		}

		return fmt.Sprintf("%v", result), nil
	}

	// propertyReplacement handles ${name} syntax
	propertyReplacement := func(name string) (string, error) {
		// Use Property eval to get the evaluated value
		p := NewProperty("$"+name, q.GetIndex(), q.FileInfo())
		result, err := p.Eval(context)
		if err != nil {
			return "", fmt.Errorf("property $%s is undefined", name)
		}

		// Extract the string value from the result
		if quoted, ok := result.(*Quoted); ok {
			return quoted.value, nil
		}

		if cssable, ok := result.(interface{ ToCSS(any) string }); ok {
			return cssable.ToCSS(emptyToCSSContext), nil
		}

		return fmt.Sprintf("%v", result), nil
	}

	// Process variable and property replacements using hand-written parsers
	// Variable interpolation strategy:
	// 1. Always try @{variable} syntax first (standard LESS syntax for all quoted strings)
	// 2. For escaped content (permissive-parsed like custom CSS properties),
	//    also try bare @variable syntax after @{variable} processing
	var err error
	var changed bool

	// First pass: always try @{variable} syntax (iteratively for nested interpolation)
	for {
		value, changed, err = interpolateBracedVars(value, variableReplacement)
		if err != nil {
			return nil, err
		}
		if !changed {
			break
		}
	}

	// Second pass: for escaped content, also try bare @variable syntax
	// This handles permissive-parsed content like `--custom: @var`
	if q.escaped {
		for {
			value, changed, err = interpolateBareVars(value, variableReplacement)
			if err != nil {
				return nil, err
			}
			if !changed {
				break
			}
		}
	}

	// Third pass: property interpolation ${name}
	for {
		value, changed, err = interpolateBracedProps(value, propertyReplacement)
		if err != nil {
			return nil, err
		}
		if !changed {
			break
		}
	}

	// Match JavaScript behavior: first parameter should be quote + value + quote
	return NewQuoted(q.quote+value+q.quote, value, q.escaped, q.GetIndex(), q.FileInfo()), nil
}

// Compare compares two quoted strings
func (q *Quoted) Compare(other any) *int {
	// Match JavaScript: if (other.type === 'Quoted' && !this.escaped && !other.escaped)
	if otherQuoted, ok := other.(*Quoted); ok && !q.escaped && !otherQuoted.escaped {
		// Match JavaScript: return Node.numericCompare(this.value, other.value);
		result := NumericCompareStrings(q.value, otherQuoted.value)
		return &result
	}
	
	// Match JavaScript: return other.toCSS && this.toCSS() === other.toCSS() ? 0 : undefined;
	if otherCSSable, ok := other.(interface{ ToCSS(any) string }); ok {
		if q.ToCSS(nil) == otherCSSable.ToCSS(nil) {
			result := 0
			return &result
		}
	}
	
	return nil // undefined in JavaScript
} 