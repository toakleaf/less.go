package less_go

import (
	"fmt"
	"os"
)

// VariableCall represents a variable call node in the Less AST
type VariableCall struct {
	*Node
	variable  string
	_index    int
	_fileInfo map[string]any
	allowRoot bool
}

func NewVariableCall(variable string, index int, currentFileInfo map[string]any) *VariableCall {
	return &VariableCall{
		Node:      NewNode(),
		variable:  variable,
		_index:    index,
		_fileInfo: currentFileInfo,
		allowRoot: true,
	}
}

func (vc *VariableCall) Type() string {
	return "VariableCall"
}

func (vc *VariableCall) GetType() string {
	return "VariableCall"
}

func (vc *VariableCall) GetIndex() int {
	return vc._index
}

func (vc *VariableCall) FileInfo() map[string]any {
	return vc._fileInfo
}

// Eval evaluates the variable call - match JavaScript implementation
func (vc *VariableCall) Eval(context any) (result any, err error) {
	// Debug: trace incoming context for all variable calls
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		switch ctx := context.(type) {
		case *Eval:
			fmt.Fprintf(os.Stderr, "[VariableCall.Eval] %s context: *Eval, mediaPath len=%d, mediaBlocks len=%d\n", vc.variable, len(ctx.MediaPath), len(ctx.MediaBlocks))
		case map[string]any:
			mp, _ := ctx["mediaPath"].([]any)
			mb, _ := ctx["mediaBlocks"].([]any)
			fmt.Fprintf(os.Stderr, "[VariableCall.Eval] %s context: map, mediaPath len=%d, mediaBlocks len=%d\n", vc.variable, len(mp), len(mb))
		default:
			fmt.Fprintf(os.Stderr, "[VariableCall.Eval] %s context: unknown type %T\n", vc.variable, context)
		}
	}

	// Use defer/recover to catch panics and convert them to errors
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			if lessErr, ok := r.(*LessError); ok {
				err = lessErr
			} else if lessErr, ok := r.(LessError); ok {
				err = &lessErr
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	// Match JavaScript: let detachedRuleset = new Variable(this.variable, this.getIndex(), this.fileInfo()).eval(context);
	variable := NewVariable(vc.variable, vc.GetIndex(), vc.FileInfo())
	// In JavaScript, if eval throws, execution stops. In Go, we need to check the error.
	detachedRuleset, varErr := variable.Eval(context)
	if varErr != nil {
		return nil, varErr
	}

	// Debug: check what we got
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] %s evaluated to type: %T, value: %+v\n", vc.variable, detachedRuleset, detachedRuleset)
		if dr, ok := detachedRuleset.(*DetachedRuleset); ok {
			fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] %s DetachedRuleset frames: nil=%v, len=%d\n", vc.variable, dr.frames == nil, len(dr.frames))
		}
	}

	errorMsg := fmt.Sprintf("Could not evaluate variable call %s", vc.variable)

	// Handle MixinCall - when a variable contains a mixin call, we need to evaluate it
	// This handles cases like: @alias: .mixin(); @alias();
	if mixinCall, ok := detachedRuleset.(*MixinCall); ok {
		rules, err := mixinCall.Eval(context)
		if err != nil {
			return nil, err
		}
		// Return the rules in the format expected by ruleset.go
		return map[string]any{"rules": rules}, nil
	}

	// Match JavaScript: if (!detachedRuleset.ruleset)
	var hasRuleset bool
	if dr, ok := detachedRuleset.(*DetachedRuleset); ok && dr.ruleset != nil {
		hasRuleset = true
	} else if dr, ok := detachedRuleset.(interface{ GetRuleset() any }); ok && dr.GetRuleset() != nil {
		// Also check for objects with GetRuleset method
		hasRuleset = true
	}

	if !hasRuleset {
		var rules any

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] No ruleset, checking for rules in type %T\n", detachedRuleset)
		}

		// Match JavaScript conditions in order
		if rulesObj, ok := detachedRuleset.(interface{ GetRules() []any }); ok && rulesObj.GetRules() != nil {
			// if (detachedRuleset.rules) - with GetRules() method
			rules = detachedRuleset
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] Found GetRules() method\n")
			}
		} else if mapObj, ok := detachedRuleset.(map[string]any); ok {
			// if (detachedRuleset.rules) - plain map with "rules" key
			if rulesVal, hasRules := mapObj["rules"]; hasRules && rulesVal != nil {
				rules = detachedRuleset
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] Found map with rules key\n")
				}
			} else if valArr, hasValue := mapObj["value"].([]any); hasValue {
				// else if (Array.isArray(detachedRuleset.value))
				rules = NewRuleset([]any{}, valArr, false, nil)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] Found map with value array\n")
				}
			}
		} else if arr, ok := detachedRuleset.([]any); ok {
			// else if (Array.isArray(detachedRuleset))
			rules = NewRuleset([]any{}, arr, false, nil)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] Found array, creating ruleset\n")
			}
		} else if valObj, ok := detachedRuleset.(interface{ GetValue() any }); ok {
			if arr, ok := valObj.GetValue().([]any); ok {
				// else if (Array.isArray(detachedRuleset.value))
				rules = NewRuleset([]any{}, arr, false, nil)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] Found GetValue() returning array\n")
				}
			}
		}

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] rules = %v (type %T)\n", rules, rules)
		}

		if rules == nil {
			// Match JavaScript: throw error (panic in Go)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Fprintf(os.Stderr, "[DEBUG VariableCall] Throwing error: %s\n", errorMsg)
			}
			panic(NewLessError(ErrorDetails{Message: errorMsg}, nil, ""))
		}
		
		// Match JavaScript: detachedRuleset = new DetachedRuleset(rules);
		if rulesetNode, ok := rules.(*Ruleset); ok {
			// Wrap Ruleset in a Node
			node := NewNode()
			node.Value = rulesetNode
			detachedRuleset = NewDetachedRuleset(node, nil)
		} else if node, ok := rules.(*Node); ok {
			detachedRuleset = NewDetachedRuleset(node, nil)
		} else {
			// Wrap in a Node if needed
			node := NewNode()
			node.Value = rules
			detachedRuleset = NewDetachedRuleset(node, nil)
		}
	}
	
	// Match JavaScript: if (detachedRuleset.ruleset) { return detachedRuleset.callEval(context); }
	if dr, ok := detachedRuleset.(*DetachedRuleset); ok && dr.ruleset != nil {
		// For VariableCall, we need to return the evaluated result in the format expected by ruleset.go
		evalResult := dr.CallEval(context)
		// The ruleset.go expects a map with "rules" key for VariableCall
		if rs, ok := evalResult.(*Ruleset); ok {
			return map[string]any{"rules": rs.Rules}, nil
		}
		return evalResult, nil
	} else if dr, ok := detachedRuleset.(interface{ CallEval(any) any }); ok {
		// Also check for objects with CallEval method (like mocks)
		evalResult := dr.CallEval(context)
		// The ruleset.go expects a map with "rules" key for VariableCall
		if rs, ok := evalResult.(*Ruleset); ok {
			return map[string]any{"rules": rs.Rules}, nil
		}
		return evalResult, nil
	}

	// Match JavaScript: throw error (panic in Go)
	panic(NewLessError(ErrorDetails{Message: errorMsg}, nil, ""))
} 