package less_go

import (
	"fmt"
	"strings"
)

// Paren represents a parenthesized value in the Less AST
type Paren struct {
	*Node
	Value     any  // This will store the node value
	NoSpacing bool // If true, no space should be added before this paren in output
}

// NewParen creates a new Paren instance with the provided node as value
func NewParen(node any) *Paren {
	return &Paren{
		Node:  NewNode(),
		Value: node,
	}
}

// NewParenWithSpacing creates a new Paren instance with explicit spacing control
func NewParenWithSpacing(node any, noSpacing bool) *Paren {
	return &Paren{
		Node:      NewNode(),
		Value:     node,
		NoSpacing: noSpacing,
	}
}

// Type returns the type of the node
func (p *Paren) Type() string {
	return "Paren"
}

// GetType returns the type of the node for visitor pattern consistency
func (p *Paren) GetType() string {
	return "Paren"
}

func (p *Paren) GenCSS(context any, output *CSSOutput) {
	output.Add("(", nil, nil)
	if valueWithGenCSS, ok := p.Value.(interface{ GenCSS(any, *CSSOutput) }); ok {
		valueWithGenCSS.GenCSS(context, output)
	}

	output.Add(")", nil, nil)
}

func (p *Paren) Eval(context any) any {
	var evaluatedValue any = p.Value
	if valueWithEval, ok := p.Value.(interface{ Eval(any) any }); ok {
		evaluatedValue = valueWithEval.Eval(context)
	} else if valueWithEval, ok := p.Value.(interface{ Eval(any) (any, error) }); ok {
		result, _ := valueWithEval.Eval(context)
		if result != nil {
			evaluatedValue = result
		}
	}

	// Preserve NoSpacing flag through evaluation
	return NewParenWithSpacing(evaluatedValue, p.NoSpacing)
}

func (p *Paren) ToCSS(context any) string {
	var builder strings.Builder
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			if chunk == nil {
				return
			}
			switch v := chunk.(type) {
			case string:
				builder.WriteString(v)
			case fmt.Stringer:
				builder.WriteString(v.String())
			default:
				fmt.Fprintf(&builder, "%v", chunk)
			}
		},
		IsEmpty: func() bool {
			return builder.Len() == 0
		},
	}
	p.GenCSS(context, output)
	return builder.String()
} 