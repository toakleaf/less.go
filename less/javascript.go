package less_go

import (
	"fmt"
	"strings"
)

type JavaScript struct {
	*JsEvalNode
	escaped    bool
	expression string
}

func NewJavaScript(string string, escaped bool, index int, currentFileInfo map[string]any) *JavaScript {
	jsEvalNode := NewJsEvalNode()
	jsEvalNode.Index = index
	jsEvalNode.SetFileInfo(currentFileInfo)

	return &JavaScript{
		JsEvalNode: jsEvalNode,
		escaped:    escaped,
		expression: string,
	}
}

func (j *JavaScript) GetType() string {
	return "JavaScript"
}

func (j *JavaScript) Type() string {
	return "JavaScript"
}

func (j *JavaScript) GetIndex() int {
	return j.JsEvalNode.GetIndex()
}

func (j *JavaScript) FileInfo() map[string]any {
	return j.JsEvalNode.FileInfo()
}

func (j *JavaScript) Eval(context any) (any, error) {
	result, err := j.EvaluateJavaScript(j.expression, context)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case float64:
		if !parserIsNaN(v) {
			dim, err := NewDimension(v, nil)
			if err != nil {
				return nil, err
			}
			return dim, nil
		}
	case string:
		return NewQuoted(`"`+v+`"`, v, j.escaped, j.GetIndex(), j.FileInfo()), nil
	case *JSArrayResult:
		return NewAnonymous(v.Value, 0, nil, false, false, nil), nil
	case *JSEmptyResult:
		return NewAnonymous("", 0, nil, false, false, nil), nil
	case []any:
		var values []string
		for _, item := range v {
			values = append(values, fmt.Sprintf("%v", item))
		}
		return NewAnonymous(strings.Join(values, ", "), 0, nil, false, false, nil), nil
	}

	return NewAnonymous(result, 0, nil, false, false, nil), nil
}

func parserIsNaN(f float64) bool {
	return f != f
} 