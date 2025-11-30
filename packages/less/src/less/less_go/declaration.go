package less_go

import (
	"fmt"
	"os"
	"strings"
)

type Declaration struct {
	*Node
	name      any
	Value     *Value
	important string
	merge     any // Can be bool or string ('+' for comma merge)
	inline    bool
	variable  bool
}

// Uses sync.Pool to reuse Declaration objects.
func NewDeclaration(name any, value any, important any, merge any, index int, fileInfo map[string]any, inline bool, variable any) (*Declaration, error) {
	node := NewNode()
	node.TypeIndex = GetTypeIndexForNodeType("Declaration")

	d := GetDeclarationFromPool()
	d.Node = node
	// Intern property names when they are strings (most common case)
	if nameStr, ok := name.(string); ok {
		d.name = Intern(nameStr)
	} else {
		d.name = name
	}
	d.important = ""
	d.merge = merge
	d.inline = inline
	d.variable = false

	// Set index and fileInfo
	d.Index = index
	d.SetFileInfo(fileInfo)

	// Handle important flag
	if important != nil {
		if str, ok := important.(string); ok {
			// Match JavaScript: always prepend a space to the trimmed important string
			// JavaScript: this.important = important ? ` ${important.trim()}` : '';
			str = strings.TrimSpace(str)
			if str != "" {
				d.important = Intern(" " + str)
			}
		}
	}

	// Handle value - match JavaScript logic: (value instanceof Node) ? value : new Value([value ? new Anonymous(value) : null])
	if val, ok := value.(*Value); ok {
		d.Value = val
	} else {
		// Check if value is already a Node type (matches JavaScript: value instanceof Node)
		isNode := false
		switch value.(type) {
		case *Node, *Color, *Dimension, *Quoted, *Anonymous, *Keyword, *Expression, *Call, *Ruleset, *Declaration:
			isNode = true
		case interface{ GetType() string }:
			// Has GetType method, likely a node
			isNode = true
		}
		
		if isNode {
			// Value is already a Node, wrap it in Value([node])
			newValue, err := NewValue([]any{value})
			if err != nil {
				return nil, err
			}
			d.Value = newValue
		} else {
			// Value is not a Node, wrap in Value([Anonymous(value)])
			var anonymousValue any
			if value != nil {
				anonymousValue = NewAnonymous(value, 0, nil, false, false, nil)
			} else {
				anonymousValue = nil
			}
			newValue, err := NewValue([]any{anonymousValue})
			if err != nil {
				return nil, err
			}
			d.Value = newValue
		}
	}

	// Handle variable flag
	if variable != nil {
		if v, ok := variable.(bool); ok {
			d.variable = v
		}
	} else {
		// Check if name starts with '@'
		if str, ok := name.(string); ok {
			d.variable = len(str) > 0 && str[0] == '@'
		}
	}

	// Set allowRoot through the Node interface
	if n, ok := interface{}(d.Node).(interface{ SetAllowRoot(bool) }); ok {
		n.SetAllowRoot(true)
	}
	d.SetParent(d.Value, d.Node)

	return d, nil
}

func (d *Declaration) GetType() string {
	return "Declaration"
}

func (d *Declaration) GetTypeIndex() int {
	// Return from Node field if set, otherwise get from registry
	if d.Node != nil && d.Node.TypeIndex != 0 {
		return d.Node.TypeIndex
	}
	return GetTypeIndexForNodeType("Declaration")
}

func (d *Declaration) GetVariable() bool {
	return d.variable
}

func (d *Declaration) GetName() string {
	if nameStr, ok := d.name.(string); ok {
		return nameStr
	}
	return fmt.Sprintf("%v", d.name)
}

func (d *Declaration) GetMerge() any {
	return d.merge
}

func (d *Declaration) MergeType() string {
	switch m := d.merge.(type) {
	case string:
		return m
	case bool:
		if m {
			return "true" // space separated
		}
	}
	return ""
}

func (d *Declaration) GetImportant() bool {
	return d.important != ""
}

func (d *Declaration) GetValue() any {
	return d.Value
}

func (d *Declaration) SetValue(value any) {
	if v, ok := value.(*Value); ok {
		d.Value = v
	}
}

func (d *Declaration) SetImportant(important bool) {
	if important {
		// Match JavaScript: always prepend a space
		d.important = " !important"
	} else {
		d.important = ""
	}
}

func evalName(context any, name []any) (string, error) {
	value := ""
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			if chunk != nil {
				switch v := chunk.(type) {
				case string:
					value += v
				case *Keyword:
					value += v.value
				default:
					value += fmt.Sprintf("%v", v)
				}
			}
		},
	}

	// JavaScript always evaluates each name element first
	for _, n := range name {
		// Always evaluate first, matching JavaScript behavior
		var evaluated any = n
		if evaluator, ok := n.(Evaluator); ok {
			evald, err := evaluator.Eval(context)
			if err != nil {
				// Propagate errors like JavaScript does
				// This ensures undefined variables in property names are caught
				return "", err
			} else if evald != nil {
				evaluated = evald
			}
		}

		// Then generate CSS
		if generator, ok := evaluated.(CSSGenerator); ok {
			generator.GenCSS(context, output)
		} else {
			// Handle simple types
			switch v := evaluated.(type) {
			case *Keyword:
				output.Add(v.value, nil, nil)
			case *Anonymous:
				output.Add(v.Value, nil, nil)
			default:
				output.Add(v, nil, nil)
			}
		}
	}

	return value, nil
}

func (d *Declaration) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok && d.Value != nil {
		result := v.Visit(d.Value)
		if resultValue, ok := result.(*Value); ok {
			d.Value = resultValue
		}
	}
}

func (d *Declaration) Eval(context any) (any, error) {
	if context == nil {
		return nil, fmt.Errorf("context is required for Declaration.Eval")
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[Declaration.Eval] name=%v, value type=%T\n", d.name, d.Value)
	}

	mathBypass := false
	var prevMath any
	var name any = d.name
	variable := d.variable

	// Handle name evaluation
	if str, ok := name.(string); !ok {
		nameArr, ok := name.([]any)
		if ok && len(nameArr) == 1 {
			if keyword, ok := nameArr[0].(*Keyword); ok {
				name = keyword.value
			} else {
				evaluatedName, err := evalName(context, nameArr)
				if err != nil {
					return nil, err
				}
				name = evaluatedName
				// Only set variable=false for dynamic/interpolated names, not simple variable declarations
				variable = false
			}
		} else if ok {
			evaluatedName, err := evalName(context, nameArr)
			if err != nil {
				return nil, err
			}
			name = evaluatedName
			// Only set variable=false for dynamic/interpolated names
			variable = false
		}
		// Note: Don't set variable=false for simple name extractions (like Keyword to string)
		// The original d.variable value should be preserved in those cases
	} else {
		name = str
	}

	// Handle font and math context
	if name == "font" {
		if ctx, ok := context.(map[string]any); ok {
			if mathVal, ok := ctx["math"]; ok {
				if os.Getenv("LESS_GO_TRACE") == "1" {
					fmt.Printf("[DECL-DEBUG] Font declaration: current math=%v\n", mathVal)
				}
				if mathVal == Math.Always {
					mathBypass = true
					prevMath = mathVal
					ctx["math"] = Math.ParensDivision
					if os.Getenv("LESS_GO_TRACE") == "1" {
						fmt.Printf("[DECL-DEBUG] Changed math from Always to ParensDivision for font\n")
					}
				}
			}
		}
	}

	// Create important scope
	if evalCtx, ok := context.(*Eval); ok {
		// For *Eval context, use typed PushImportantScope
		evalCtx.PushImportantScope()
	} else if ctx, ok := context.(map[string]any); ok {
		// For map context, manage importantScope in the map
		if importantScope, ok := ctx["importantScope"].([]any); ok {
			ctx["importantScope"] = append(importantScope, map[string]any{})
		} else {
			// Initialize importantScope if it doesn't exist
			ctx["importantScope"] = []any{map[string]any{}}
		}
	}

	// Evaluate value
	var evaldValue any
	var err error
	evaldValue, err = d.Value.Eval(context)
	if err != nil {
		return nil, err
	}

	// Check for detached ruleset - match JavaScript: if (!this.variable && evaldValue.type === 'DetachedRuleset')
	if !variable {
		// Check if evaldValue is a DetachedRuleset type
		isDetachedRuleset := false
		if _, ok := evaldValue.(*DetachedRuleset); ok {
			isDetachedRuleset = true
		} else if node, ok := evaldValue.(interface{ GetType() string }); ok {
			if node.GetType() == "DetachedRuleset" {
				isDetachedRuleset = true
			}
		}

		if isDetachedRuleset {
			return nil, &LessError{
				Type:    "Syntax",
				Message: "Rulesets cannot be evaluated on a property.",
				Index:   d.Index,
				Filename: func() string {
					if d.FileInfo() != nil {
						if f, ok := d.FileInfo()["filename"].(string); ok {
							return f
						}
					}
					return ""
				}(),
			}
		}
	}

	// Handle important flag
	important := d.important
	if evalCtx, ok := context.(*Eval); ok {
		// For *Eval context, use typed PopImportantScope
		lastScope := evalCtx.PopImportantScope()
		// Check if we should use the important flag from the scope
		if important == "" && lastScope.Important != "" {
			important = lastScope.Important
		}
	} else if ctx, ok := context.(map[string]any); ok {
		// For map context, pop from importantScope in the map
		if importantScope, ok := ctx["importantScope"].([]any); ok && len(importantScope) > 0 {
			// Pop the scope
			lastScope := importantScope[len(importantScope)-1]
			ctx["importantScope"] = importantScope[:len(importantScope)-1]

			// Check if we should use the important flag from the scope
			if important == "" && lastScope != nil {
				if scope, ok := lastScope.(map[string]any); ok {
					if imp, ok := scope["important"].(string); ok && imp != "" {
						important = imp
					}
				}
			}
		}
	}

	// Restore math context
	if mathBypass {
		if ctx, ok := context.(map[string]any); ok {
			if os.Getenv("LESS_GO_TRACE") == "1" {
				fmt.Printf("[DECL-DEBUG] Restoring math to %v after font declaration\n", prevMath)
			}
			ctx["math"] = prevMath
		}
	}

	newDecl, err := NewDeclaration(name, evaldValue, important, d.merge, d.Index, d.FileInfo(), d.inline, variable)
	if err != nil {
		return nil, err
	}

	return newDecl, nil
}

func (d *Declaration) GenCSS(context any, output *CSSOutput) {
	// Check visibility - skip if node blocks visibility and is not explicitly visible
	// Also skip if this is a variable declaration (variables are not output as CSS)
	if d.variable {
		return
	}
	if d.Node != nil && d.Node.BlocksVisibility() {
		nodeVisible := d.Node.IsVisible()
		if nodeVisible == nil || !*nodeVisible {
			// Node blocks visibility and is not explicitly visible, skip output
			return
		}
	}

	compress := false
	if ctx, ok := context.(map[string]any); ok {
		if c, ok := ctx["compress"].(bool); ok {
			compress = c
		}
	}

	// Format name as string for CSS output
	nameStr := ""
	switch n := d.name.(type) {
	case string:
		nameStr = n
	case *Keyword:
		nameStr = n.value
	case *Anonymous:
		nameStr = fmt.Sprintf("%v", n.Value)
	case []any:
		evaluatedName, err := evalName(context, n)
		if err != nil {
			// Panic to be caught by defer/recover above
			panic(err)
		}
		nameStr = evaluatedName
	default:
		nameStr = fmt.Sprintf("%v", n)
	}

	// Add name
	if compress {
		output.Add(nameStr, d.FileInfo(), d.GetIndex())
		output.Add(":", d.FileInfo(), d.GetIndex())
	} else {
		output.Add(nameStr, d.FileInfo(), d.GetIndex())
		output.Add(": ", d.FileInfo(), d.GetIndex())
	}

	// Add value with error handling to match JavaScript
	// Use defer/recover to catch panics and convert to proper errors
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error with proper index and filename information
			var errMsg string
			switch e := r.(type) {
			case error:
				errMsg = e.Error()
			default:
				errMsg = fmt.Sprintf("%v", e)
			}
			
			// Create an error with index and filename similar to JavaScript
			filename := ""
			if d.FileInfo() != nil {
				if f, ok := d.FileInfo()["filename"].(string); ok {
					filename = f
				}
			}
			
			// Re-panic with enhanced error message
			panic(fmt.Errorf("%s (index: %d, filename: %s)", errMsg, d.GetIndex(), filename))
		}
	}()

	// For inline declarations (e.g., in media features), ensure Variables are fully evaluated
	// This handles cases where variables are nested in declarations within parens
	if d.inline {
		// Try to evaluate the value to resolve any nested variables
		if evaluated, err := d.Value.Eval(context); err == nil && evaluated != nil {
			// Use the evaluated value instead
			if gen, ok := evaluated.(interface{ GenCSS(any, *CSSOutput) }); ok {
				gen.GenCSS(context, output)
			} else {
				output.Add(fmt.Sprintf("%v", evaluated), d.FileInfo(), d.GetIndex())
			}
		} else {
			// Fall back to normal GenCSS if evaluation fails
			d.Value.GenCSS(context, output)
		}
	} else {
		// Normal (non-inline) declarations use the standard GenCSS
		d.Value.GenCSS(context, output)
	}

	// Add important and semicolon
	if d.important != "" {
		output.Add(d.important, d.FileInfo(), d.GetIndex())
	}

	if !d.inline && !(compress && isLastRule(context)) {
		output.Add(";", d.FileInfo(), d.GetIndex())
	} else {
		output.Add("", d.FileInfo(), d.GetIndex())
	}
}


func isLastRule(context any) bool {
	if ctx, ok := context.(map[string]any); ok {
		if lastRule, ok := ctx["lastRule"].(bool); ok {
			return lastRule
		}
	}
	return false
}

func (d *Declaration) IsVisible() bool {
	return true
}

func (d *Declaration) MakeImportant() any {
	// If already important, preserve the existing important value
	importantValue := "!important"
	if d.important != "" {
		importantValue = strings.TrimSpace(d.important)
	}
	newDecl, _ := NewDeclaration(d.name, d.Value, importantValue, d.merge, d.GetIndex(), d.FileInfo(), d.inline, d.variable)
	return newDecl
}

func (d *Declaration) ToCSS(context any) string {
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
	d.GenCSS(context, output)
	return strings.Join(strs, "")
} 