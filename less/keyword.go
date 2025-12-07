package less_go

import (
	"fmt"
	"strings"
)

type Keyword struct {
	*Node
	value string
	type_ string
}

func NewKeyword(value string) *Keyword {
	internedValue := Intern(value)
	k := &Keyword{
		Node:  NewNode(),
		value: internedValue,
		type_: "Keyword",
	}
	k.Value = internedValue
	return k
}

func (k *Keyword) Type() string {
	return k.type_
}

func (k *Keyword) GetType() string {
	return "Keyword"
}

func (k *Keyword) GetValue() string {
	return k.value
}

func (k *Keyword) GenCSS(context any, output *CSSOutput) {
	if k.value == "%" {
		panic(map[string]string{
			"type":    "Syntax",
			"message": "Invalid % without number",
		})
	}
	output.Add(k.value, nil, nil)
}

func (k *Keyword) ToCSS(context any) string {
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
	k.GenCSS(context, output)
	return builder.String()
}

func (k *Keyword) Eval(context any) (any, error) {
	return k, nil
}

var (
	KeywordTrue  = NewKeyword("true")
	KeywordFalse = NewKeyword("false")
) 