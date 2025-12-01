package less_go

import (
	"fmt"
	"strings"
)

type Attribute struct {
	*Node
	Key   any
	Op    string
	Value any
	Cif   string
}

func NewAttribute(key any, op string, value any, cif string) *Attribute {
	return &Attribute{
		Node:  NewNode(),
		Key:   key,
		Op:    op,
		Value: value,
		Cif:   cif,
	}
}

func (a *Attribute) Type() string {
	return "Attribute"
}

func (a *Attribute) GetType() string {
	return "Attribute"
}

func (a *Attribute) Eval(context any) (any, error) {
	var key any
	var value any

	if a.Key != nil {
		if evaluable, ok := a.Key.(interface{ Eval(any) (any, error) }); ok {
			var err error
			key, err = evaluable.Eval(context)
			if err != nil {
				return nil, err
			}
		} else if evaluable, ok := a.Key.(ParserEvaluable); ok {
			key = evaluable.Eval(context)
		} else {
			key = a.Key
		}
	}

	if a.Value != nil {
		if evaluable, ok := a.Value.(interface{ Eval(any) (any, error) }); ok {
			var err error
			value, err = evaluable.Eval(context)
			if err != nil {
				return nil, err
			}
		} else if evaluable, ok := a.Value.(ParserEvaluable); ok {
			value = evaluable.Eval(context)
		} else {
			value = a.Value
		}
	}

	return NewAttribute(key, a.Op, value, a.Cif), nil
}

func (a *Attribute) GenCSS(context any, output *CSSOutput) {
	output.Add(a.ToCSS(context), nil, nil)
}

func (a *Attribute) ToCSS(context any) string {
	if a.Key == nil {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteString("[")

	if cssable, ok := a.Key.(CSSable); ok {
		builder.WriteString(cssable.ToCSS(context))
	} else {
		builder.WriteString(fmt.Sprintf("%v", a.Key))
	}

	if a.Op != "" {
		builder.WriteString(a.Op)
		if a.Value != nil {
			if cssable, ok := a.Value.(CSSable); ok {
				builder.WriteString(cssable.ToCSS(context))
			} else {
				builder.WriteString(fmt.Sprintf("%v", a.Value))
			}
		}
	}

	if a.Cif != "" {
		builder.WriteString(" ")
		builder.WriteString(a.Cif)
	}

	builder.WriteString("]")
	return builder.String()
}

type ParserEvaluable interface {
	Eval(any) any
}

type CSSable interface {
	ToCSS(any) string
} 