package less_go

// MathHelperError represents an argument error
type MathHelperError struct {
	Type    string
	Message string
}

func (e *MathHelperError) Error() string {
	return e.Message
}

// MathHelper applies a mathematical function to a dimension value.
// If unit is nil, uses the dimension's unit; otherwise unifies and uses the provided unit.
func MathHelper(fn func(float64) float64, unit *Unit, n any) (*Dimension, error) {
	dim, ok := n.(*Dimension)
	if !ok {
		return nil, &MathHelperError{
			Type:    "Argument",
			Message: "argument must be a number",
		}
	}

	var resultUnit *Unit
	var workingDim *Dimension = dim

	if unit == nil {
		resultUnit = dim.Unit
	} else {
		workingDim = dim.Unify()
		resultUnit = unit
	}

	value := workingDim.Value
	result := fn(value)
	return NewDimension(result, resultUnit)
}