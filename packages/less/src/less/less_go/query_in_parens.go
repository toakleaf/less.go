package less_go

import "strings"

// QueryInParens represents a query in parentheses node in the Less AST
type QueryInParens struct {
	*Node
	op      string
	lvalue  any
	mvalue  any
	op2     string
	rvalue  any
	mvalues []any
}

// NewQueryInParens creates a new QueryInParens instance
func NewQueryInParens(op string, l any, m any, op2 string, r any, i int) *QueryInParens {
	q := &QueryInParens{
		Node:    NewNode(),
		op:      strings.TrimSpace(op),
		lvalue:  l,
		mvalue:  m,
		op2:     strings.TrimSpace(op2),
		rvalue:  r,
		mvalues: make([]any, 0),
	}
	q.Index = i
	return q
}

// Type returns the type of the node
func (q *QueryInParens) Type() string {
	return "QueryInParens"
}

// GetType returns the type of the node
func (q *QueryInParens) GetType() string {
	return "QueryInParens"
}

// Accept visits the node with a visitor
func (q *QueryInParens) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		q.lvalue = v.Visit(q.lvalue)
		q.mvalue = v.Visit(q.mvalue)
		if q.rvalue != nil {
			q.rvalue = v.Visit(q.rvalue)
		}
	}
}

// Eval evaluates the query
// IMPORTANT: This method returns a NEW QueryInParens instance with evaluated values
// rather than mutating the original. This is critical for mixin expansion where
// the same QueryInParens node may be evaluated multiple times with different contexts.
func (q *QueryInParens) Eval(context any) (any, error) {
	// Create a new QueryInParens to avoid mutating the original
	result := &QueryInParens{
		Node:    NewNode(),
		op:      q.op,
		op2:     q.op2,
		mvalues: make([]any, 0),
	}
	result.Index = q.Index

	// Evaluate lvalue
	lvalue := q.lvalue
	if evaluator, ok := lvalue.(interface{ Eval(any) (any, error) }); ok {
		var err error
		lvalue, err = evaluator.Eval(context)
		if err != nil {
			return nil, err
		}
	} else if evaluator, ok := lvalue.(interface{ Eval(any) any }); ok {
		lvalue = evaluator.Eval(context)
	}
	result.lvalue = lvalue

	// Evaluate mvalue (the middle/comparison value)
	mvalue := q.mvalue
	if evaluator, ok := mvalue.(interface{ Eval(any) (any, error) }); ok {
		var err error
		mvalue, err = evaluator.Eval(context)
		if err != nil {
			return nil, err
		}
	} else if evaluator, ok := mvalue.(interface{ Eval(any) any }); ok {
		mvalue = evaluator.Eval(context)
	}
	result.mvalue = mvalue

	// Evaluate rvalue if present
	if q.rvalue != nil {
		rvalue := q.rvalue
		if evaluator, ok := rvalue.(interface{ Eval(any) (any, error) }); ok {
			var err error
			rvalue, err = evaluator.Eval(context)
			if err != nil {
				return nil, err
			}
		} else if evaluator, ok := rvalue.(interface{ Eval(any) any }); ok {
			rvalue = evaluator.Eval(context)
		}
		result.rvalue = rvalue
	}

	return result, nil
}

// GenCSS generates CSS representation
func (q *QueryInParens) GenCSS(context any, output *CSSOutput) {
	if q.lvalue != nil {
		if anon, ok := q.lvalue.(*Anonymous); ok {
			output.Add(anon.Value, nil, nil)
		} else if generator, ok := q.lvalue.(CSSGenerator); ok {
			generator.GenCSS(context, output)
		} else {
			output.Add(q.lvalue, nil, nil)
		}
	}

	output.Add(" "+q.op+" ", nil, nil)

	if len(q.mvalues) > 0 {
		if val, ok := SafeSliceIndex(q.mvalues, 0); ok {
			q.mvalue = val
			q.mvalues = q.mvalues[1:]
		}
	}

	if !SafeNilCheck(q.mvalue) {
		if anon, ok := SafeTypeAssertion[*Anonymous](q.mvalue); ok {
			output.Add(anon.Value, nil, nil)
		} else if generator, ok := SafeTypeAssertion[CSSGenerator](q.mvalue); ok {
			SafeGenCSS(generator, context, output)
		} else {
			output.Add(q.mvalue, nil, nil)
		}
	} else if len(q.mvalues) == 0 {
		// Instead of panicking, output a placeholder or empty value
		// This maintains CSS generation while avoiding panics
		output.Add("/* missing value */", nil, nil)
	}

	if q.rvalue != nil {
		output.Add(" "+q.op2+" ", nil, nil)
		if anon, ok := q.rvalue.(*Anonymous); ok {
			output.Add(anon.Value, nil, nil)
		} else if generator, ok := q.rvalue.(CSSGenerator); ok {
			generator.GenCSS(context, output)
		} else {
			output.Add(q.rvalue, nil, nil)
		}
	}
} 