package less_go

import (
	"fmt"
	"strings"
	"testing"
)

type TestCSSOutput struct {
	*CSSOutput
	str string
}

func NewTestCSSOutput() *TestCSSOutput {
	tco := &TestCSSOutput{
		str: "",
	}
	addFunc := func(chunk any, fileInfo any, index any) {
		if chunk != nil {
			tco.str += fmt.Sprintf("%v", chunk)
		}
	}
	tco.CSSOutput = &CSSOutput{
		Add: addFunc,
	}
	return tco
}

func TestQueryInParens(t *testing.T) {
	var context map[string]any
	var query *QueryInParens

	setup := func() {
		context = map[string]any{
			"frames":        []any{},
			"importantScope": []any{},
			"math":          0,
			"numPrecision":  8,
		}
	}

	t.Run("constructor", func(t *testing.T) {
		t.Run("should initialize with basic properties", func(t *testing.T) {
			setup()
			op := "and"
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			op2 := "or"
			r := NewAnonymous("right", 0, nil, false, false, nil)
			i := 1

			query = NewQueryInParens(op, l, m, op2, r, i)

			if query.op != "and" {
				t.Errorf("Expected op to be 'and', got '%s'", query.op)
			}
			if query.lvalue != l {
				t.Error("Expected lvalue to match input")
			}
			if query.mvalue != m {
				t.Error("Expected mvalue to match input")
			}
			if query.op2 != "or" {
				t.Errorf("Expected op2 to be 'or', got '%s'", query.op2)
			}
			if query.rvalue != r {
				t.Error("Expected rvalue to match input")
			}
			if query.Index != 1 {
				t.Errorf("Expected _index to be 1, got %d", query.Index)
			}
			if len(query.mvalues) != 0 {
				t.Error("Expected mvalues to be empty")
			}
		})

		t.Run("should handle missing op2 and rvalue", func(t *testing.T) {
			setup()
			op := "and"
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)

			query = NewQueryInParens(op, l, m, "", nil, 0)

			if query.op != "and" {
				t.Errorf("Expected op to be 'and', got '%s'", query.op)
			}
			if query.lvalue != l {
				t.Error("Expected lvalue to match input")
			}
			if query.mvalue != m {
				t.Error("Expected mvalue to match input")
			}
			if query.op2 != "" {
				t.Errorf("Expected op2 to be empty, got '%s'", query.op2)
			}
			if query.rvalue != nil {
				t.Error("Expected rvalue to be nil")
			}
		})

		t.Run("should handle index parameter", func(t *testing.T) {
			setup()
			op := "and"
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			i := 5

			query = NewQueryInParens(op, l, m, "", nil, i)

			if query.Index != 5 {
				t.Errorf("Expected _index to be 5, got %d", query.Index)
			}
		})

		t.Run("should trim whitespace from operators", func(t *testing.T) {
			setup()
			op := "  and  "
			op2 := "  or  "
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			r := NewAnonymous("right", 0, nil, false, false, nil)

			query = NewQueryInParens(op, l, m, op2, r, 0)

			if query.op != "and" {
				t.Errorf("Expected op to be 'and', got '%s'", query.op)
			}
			if query.op2 != "or" {
				t.Errorf("Expected op2 to be 'or', got '%s'", query.op2)
			}
		})

		t.Run("should handle null/undefined values for l and m", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)

			query1 := NewQueryInParens("and", nil, m, "", nil, 0)
			query2 := NewQueryInParens("and", l, nil, "", nil, 0)

			if query1.lvalue != nil {
				t.Error("Expected lvalue to be nil")
			}
			if query2.mvalue != nil {
				t.Error("Expected mvalue to be nil")
			}
		})

		t.Run("should handle empty string operators", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("", l, m, "", nil, 0)
			if query.op != "" {
				t.Errorf("Expected op to be empty, got '%s'", query.op)
			}
		})

		t.Run("should handle non-string operators", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)

			query = NewQueryInParens("123", l, m, "", nil, 0)
			if query.op != "123" {
				t.Errorf("Expected op to be '123', got '%s'", query.op)
			}

			query = NewQueryInParens("true", l, m, "", nil, 0)
			if query.op != "true" {
				t.Errorf("Expected op to be 'true', got '%s'", query.op)
			}

			query = NewQueryInParens("[object Object]", l, m, "", nil, 0)
			if query.op != "[object Object]" {
				t.Errorf("Expected op to be '[object Object]', got '%s'", query.op)
			}
		})
	})

	t.Run("eval", func(t *testing.T) {
		t.Run("should evaluate lvalue, mvalue, and rvalue", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			r := NewAnonymous("right", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "or", r, 0)

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}

			// Eval should return a NEW instance (important for mixin expansion)
			if result == query {
				t.Error("Expected result to be a different instance (new node)")
			}
			resultQuery := result.(*QueryInParens)
			if resultQuery.op != query.op {
				t.Error("Expected op to be preserved")
			}
			if resultQuery.op2 != query.op2 {
				t.Error("Expected op2 to be preserved")
			}
			if resultQuery.lvalue == nil {
				t.Error("Expected lvalue to be set")
			}
			if resultQuery.mvalue == nil {
				t.Error("Expected mvalue to be set")
			}
			if resultQuery.rvalue == nil {
				t.Error("Expected rvalue to be set")
			}
		})

		t.Run("should handle variable declarations in context", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			r := NewAnonymous("right", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "or", r, 0)

			value, _ := NewValue([]any{NewAnonymous("value", 0, nil, false, false, nil)})
			varDecl, _ := NewDeclaration(
				"@var",
				value,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			context["frames"] = []any{map[string]any{
				"type":  "Ruleset",
				"rules": []any{varDecl},
			}}

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}

			if result == query {
				t.Error("Expected result to be a different instance (new node)")
			}
			resultQuery := result.(*QueryInParens)
			if resultQuery.lvalue == nil {
				t.Error("Expected lvalue to be set")
			}
			if resultQuery.mvalue == nil {
				t.Error("Expected mvalue to be set")
			}
		})

		t.Run("should return new instance on each eval for mixin expansion", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			value, _ := NewValue([]any{NewAnonymous("value", 0, nil, false, false, nil)})
			varDecl, _ := NewDeclaration(
				"@var",
				value,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			context["frames"] = []any{map[string]any{
				"type":  "Ruleset",
				"rules": []any{varDecl},
			}}

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}
			result2, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Second eval failed: %v", err)
			}

			// Each evaluation must return a new instance (critical for mixin expansion)
			if result == result2 {
				t.Error("Expected each eval to return a different instance")
			}
			if result == query {
				t.Error("Expected result to be different from original")
			}
			if result2 == query {
				t.Error("Expected result2 to be different from original")
			}
		})

		t.Run("should handle empty context frames", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			context["frames"] = []any{}
			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}

			// Should return a new instance (important for mixin expansion)
			if result == query {
				t.Error("Expected result to be a different instance")
			}
			resultQuery := result.(*QueryInParens)
			// Verify values are set
			if resultQuery.lvalue == nil {
				t.Error("Expected lvalue to be set")
			}
			if resultQuery.mvalue == nil {
				t.Error("Expected mvalue to be set")
			}
		})

		t.Run("should handle multiple rulesets with variables", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			value1, _ := NewValue([]any{NewAnonymous("value1", 0, nil, false, false, nil)})
			value2, _ := NewValue([]any{NewAnonymous("value2", 0, nil, false, false, nil)})
			varDecl1, _ := NewDeclaration(
				"@var1",
				value1,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			varDecl2, _ := NewDeclaration(
				"@var2",
				value2,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			context["frames"] = []any{
				map[string]any{
					"type":  "Ruleset",
					"rules": []any{varDecl1},
				},
				map[string]any{
					"type":  "Ruleset",
					"rules": []any{varDecl2},
				},
			}

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}
			if result == query {
				t.Error("Expected result to be a different instance")
			}
		})

		t.Run("should handle non-Ruleset frame types", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			context["frames"] = []any{map[string]any{
				"type":  "OtherType",
				"rules": []any{},
			}}
			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}

			// Should return a new instance (important for mixin expansion)
			if result == query {
				t.Error("Expected result to be a different instance")
			}
			resultQuery := result.(*QueryInParens)
			// Verify values are set
			if resultQuery.lvalue == nil {
				t.Error("Expected lvalue to be set")
			}
		})

		t.Run("should handle nested variable declarations", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			value1, _ := NewValue([]any{NewAnonymous("value1", 0, nil, false, false, nil)})
			value2, _ := NewValue([]any{NewAnonymous("value2", 0, nil, false, false, nil)})
			varDecl1, _ := NewDeclaration(
				"@var1",
				value1,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			varDecl2, _ := NewDeclaration(
				"@var2",
				value2,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			context["frames"] = []any{
				map[string]any{
					"type":  "Ruleset",
					"rules": []any{varDecl1},
				},
				map[string]any{
					"type":  "Ruleset",
					"rules": []any{varDecl2},
				},
			}

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}
			if result == query {
				t.Error("Expected result to be a different instance")
			}
		})

		t.Run("should preserve operators and values", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}
			resultQuery := result.(*QueryInParens)
			if resultQuery.op != "and" {
				t.Errorf("Expected op to be 'and', got '%s'", resultQuery.op)
			}
		})

		t.Run("should handle op2 without rvalue", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "or", nil, 0)

			result, err := query.Eval(context)
			if err != nil {
				t.Fatalf("Eval failed: %v", err)
			}
			if result.(*QueryInParens).op2 != "or" {
				t.Errorf("Expected op2 to be 'or', got '%s'", result.(*QueryInParens).op2)
			}
			if result.(*QueryInParens).rvalue != nil {
				t.Error("Expected rvalue to be nil")
			}
		})
	})

	t.Run("genCSS", func(t *testing.T) {
		t.Run("should generate CSS for basic query", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			output := NewTestCSSOutput()

			query.GenCSS(context, output.CSSOutput)
			if output.str != "left and middle" {
				t.Errorf("Expected output to be 'left and middle', got '%s'", output.str)
			}
		})

		t.Run("should generate CSS for query with op2 and rvalue", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			r := NewAnonymous("right", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "or", r, 0)

			output := NewTestCSSOutput()

			query.GenCSS(context, output.CSSOutput)
			if output.str != "left and middle or right" {
				t.Errorf("Expected output to be 'left and middle or right', got '%s'", output.str)
			}
		})

		t.Run("should handle mvalues array", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			value, _ := NewValue([]any{NewAnonymous("value", 0, nil, false, false, nil)})
			varDecl, _ := NewDeclaration(
				"@var",
				value,
				"",
				false,
				0,
				nil,
				false,
				true,
			)
			context["frames"] = []any{map[string]any{
				"type":  "Ruleset",
				"rules": []any{varDecl},
			}}

			query.Eval(context)
			query.mvalue = NewAnonymous("new-middle", 0, nil, false, false, nil)
			query.mvalues = []any{query.mvalue}

			output := NewTestCSSOutput()

			query.GenCSS(context, output.CSSOutput)
			if output.str != "left and new-middle" {
				t.Errorf("Expected output to be 'left and new-middle', got '%s'", output.str)
			}
		})

		t.Run("should handle multiple mvalues", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			query.mvalues = []any{
				NewAnonymous("value1", 0, nil, false, false, nil),
				NewAnonymous("value2", 0, nil, false, false, nil),
				NewAnonymous("value3", 0, nil, false, false, nil),
			}

			output := NewTestCSSOutput()

			query.GenCSS(context, output.CSSOutput)
			if output.str != "left and value1" {
				t.Errorf("Expected output to be 'left and value1', got '%s'", output.str)
			}
		})

		t.Run("should handle undefined output", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic when output is nil")
				}
			}()
			query.GenCSS(context, nil)
		})

		t.Run("should handle empty mvalues array with undefined mvalue", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, nil, "", nil, 0)
			query.mvalues = []any{}

			output := NewTestCSSOutput()

			// Should output a placeholder comment instead of panicking
			query.GenCSS(context, output.CSSOutput)

			if !strings.Contains(output.str, "missing value") {
				t.Errorf("Expected missing value comment in output, got: %s", output.str)
			}
		})

		t.Run("should handle empty operators after trimming", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("   ", l, m, "   ", nil, 0)

			output := NewTestCSSOutput()

			query.GenCSS(context, output.CSSOutput)
			if output.str != "left  middle" {
				t.Errorf("Expected output to be 'left  middle', got '%s'", output.str)
			}
		})

		t.Run("should handle output.add throwing error", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			output := &CSSOutput{
				Add: func(chunk any, fileInfo any, index any) {
					panic("Test error")
				},
			}

			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic when output.add throws error")
				}
			}()
			query.GenCSS(context, output)
		})
	})

	t.Run("accept", func(t *testing.T) {
		t.Run("should visit all values", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			r := NewAnonymous("right", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "or", r, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					if anon, ok := node.(*Anonymous); ok {
						return NewAnonymous(anon.Value.(string)+"-visited", 0, nil, false, false, nil)
					}
					return node
				},
			}

			query.Accept(visitor)

			if query.lvalue.(*Anonymous).Value != "left-visited" {
				t.Errorf("Expected lvalue to be 'left-visited', got '%v'", query.lvalue.(*Anonymous).Value)
			}
			if query.mvalue.(*Anonymous).Value != "middle-visited" {
				t.Errorf("Expected mvalue to be 'middle-visited', got '%v'", query.mvalue.(*Anonymous).Value)
			}
			if query.rvalue.(*Anonymous).Value != "right-visited" {
				t.Errorf("Expected rvalue to be 'right-visited', got '%v'", query.rvalue.(*Anonymous).Value)
			}
		})

		t.Run("should handle missing rvalue", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					if anon, ok := node.(*Anonymous); ok {
						return NewAnonymous(anon.Value.(string)+"-visited", 0, nil, false, false, nil)
					}
					return node
				},
			}

			query.Accept(visitor)

			if query.lvalue.(*Anonymous).Value != "left-visited" {
				t.Errorf("Expected lvalue to be 'left-visited', got '%v'", query.lvalue.(*Anonymous).Value)
			}
			if query.mvalue.(*Anonymous).Value != "middle-visited" {
				t.Errorf("Expected mvalue to be 'middle-visited', got '%v'", query.mvalue.(*Anonymous).Value)
			}
			if query.rvalue != nil {
				t.Error("Expected rvalue to be nil")
			}
		})

		t.Run("should handle visitor returning null", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					return nil
				},
			}

			query.Accept(visitor)
		})

		t.Run("should handle visitor returning different node type", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					value, _ := NewValue([]any{})
					decl, _ := NewDeclaration("@var", value, "", false, 0, nil, false, true)
					return decl
				},
			}

			query.Accept(visitor)
		})

		t.Run("should handle visitor throwing error", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					panic("Test error")
				},
			}

			defer func() {
				if r := recover(); r == nil {
					t.Error("Expected panic when visitor throws error")
				}
			}()
			query.Accept(visitor)
		})

		t.Run("should handle visitor returning undefined", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					return nil
				},
			}

			query.Accept(visitor)
		})

		t.Run("should handle visitor modifying node in place", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					if anon, ok := node.(*Anonymous); ok {
						anon.Value = anon.Value.(string) + "-modified"
						return anon
					}
					return node
				},
			}

			query.Accept(visitor)
			if query.lvalue.(*Anonymous).Value != "left-modified" {
				t.Errorf("Expected lvalue to be 'left-modified', got '%v'", query.lvalue.(*Anonymous).Value)
			}
			if query.mvalue.(*Anonymous).Value != "middle-modified" {
				t.Errorf("Expected mvalue to be 'middle-modified', got '%v'", query.mvalue.(*Anonymous).Value)
			}
		})

		t.Run("should handle visitor returning non-Node object", func(t *testing.T) {
			setup()
			l := NewAnonymous("left", 0, nil, false, false, nil)
			m := NewAnonymous("middle", 0, nil, false, false, nil)
			query = NewQueryInParens("and", l, m, "", nil, 0)

			visitor := &TestVisitor{
				visit: func(node any) any {
					return map[string]any{"notANode": true}
				},
			}

			query.Accept(visitor)
		})
	})
}

type TestVisitor struct {
type TestVisitor struct {
	visit func(any) any
}

func (v *TestVisitor) Visit(node any) any {
	return v.visit(node)
} 