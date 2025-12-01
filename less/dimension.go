package less_go

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Dimension struct {
	*Node
	Value float64
	Unit  *Unit
}

func NewDimension(value any, unit any) (*Dimension, error) {
	var v float64
	switch t := value.(type) {
	case float64:
		v = t
	case int:
		v = float64(t)
	case string:
		var err error
		v, err = strconv.ParseFloat(t, 64)
		if err != nil {
			return nil, errors.New("Dimension is not a number")
		}
	default:
		return nil, errors.New("Dimension is not a number")
	}
	if math.IsNaN(v) {
		return nil, errors.New("Dimension is not a number")
	}

	var u *Unit
	if unit == nil {
		u = NewUnit(nil, nil, "")
	} else {
		switch t := unit.(type) {
		case string:
			if t != "" {
				u = NewUnit([]string{t}, nil, t)
			} else {
				u = NewUnit(nil, nil, "")
			}
		case *Unit:
			u = t
		default:
			str := fmt.Sprintf("%v", t)
			if str != "" {
				u = NewUnit([]string{str}, nil, str)
			} else {
				u = NewUnit(nil, nil, "")
			}
		}
	}

	d := &Dimension{
		Node:  NewNode(),
		Value: v,
		Unit:  u,
	}
	d.SetParent(d.Unit, d.Node)
	return d, nil
}

// Returns nil if the value is NaN.
func NewDimensionFrom(value float64, unit *Unit) *Dimension {
	if math.IsNaN(value) {
		return nil
	}
	d := &Dimension{
		Node:  NewNode(),
		Value: value,
		Unit:  unit,
	}
	d.SetParent(d.Unit, d.Node)
	return d
}

func (d *Dimension) GetType() string {
	return "Dimension"
}

func (d *Dimension) GetValue() float64 {
	return d.Value
}

func (d *Dimension) GetUnit() any {
	return d.Unit
}

func (d *Dimension) Accept(visitor any) {
	// Handle both *Visitor and interface{ Visit(any) any }
	if v, ok := visitor.(*Visitor); ok {
		result := v.Visit(d.Unit)
		if u, ok := result.(*Unit); ok {
			d.Unit = u
		}
	} else if v, ok := visitor.(interface{ Visit(any) any }); ok {
		result := v.Visit(d.Unit)
		if u, ok := result.(*Unit); ok {
			d.Unit = u
		}
	}
}

func (d *Dimension) Eval(context any) (any, error) {
	return d, nil
}

func (d *Dimension) ToColor() *Color {
	return NewColor([]float64{d.Value, d.Value, d.Value}, 1, "")
}

func (d *Dimension) GenCSS(context any, output *CSSOutput) {
	var strictUnits bool
	var compress bool
	if ctx, ok := SafeTypeAssertion[map[string]any](context); ok {
		if val, exists := SafeMapAccess(ctx, "strictUnits"); exists {
			if strictVal, ok := SafeTypeAssertion[bool](val); ok {
				strictUnits = strictVal
			}
		}
		if comp, exists := SafeMapAccess(ctx, "compress"); exists {
			if compVal, ok := SafeTypeAssertion[bool](comp); ok {
				compress = compVal
			}
		}
	}
	if strictUnits && !d.Unit.IsSingular() {
		// Match JavaScript: throw error for multiple units in strict mode
		panic(&LessError{
			Type:    "Dimension",
			Message: fmt.Sprintf("Multiple units in dimension. Correct the units or use the unit function. Bad unit: %s", d.Unit.ToString()),
		})
	}

	roundedValue := d.Fround(context, d.Value)

	// Normalize -0 to 0 (handles cases like -0.0000000001 rounding to -0)
	if roundedValue == 0 || math.Abs(roundedValue) < 1e-10 {
		roundedValue = 0
	}

	// Match JavaScript: String(value) which doesn't use scientific notation
	// for reasonable-sized numbers and preserves full precision
	var strValue string
	if roundedValue == 0 {
		strValue = "0"
	} else if roundedValue != 0 && math.Abs(roundedValue) < 0.000001 {
		// Very small numbers: would be output as 1e-6 etc. in JavaScript
		// Mimic JavaScript's toFixed(20) and trim trailing zeros
		strValue = strings.TrimRight(fmt.Sprintf("%.20f", roundedValue), "0")
		strValue = strings.TrimRight(strValue, ".")
	} else {
		// For normal numbers, match JavaScript's String() behavior
		// JavaScript doesn't use scientific notation for reasonable numbers
		// Check if the number is close to an integer
		if math.Abs(roundedValue-math.Round(roundedValue)) < 1e-10 {
			// Integer or very close to it - format without decimal places
			strValue = fmt.Sprintf("%.0f", roundedValue)
		} else {
			// Has decimal places - use %f and trim trailing zeros
			// Use sufficient precision to match JavaScript
			strValue = fmt.Sprintf("%.10f", roundedValue)
			strValue = strings.TrimRight(strValue, "0")
			strValue = strings.TrimRight(strValue, ".")
		}
	}
	if compress {
		if roundedValue == 0 && d.Unit.IsLength() {
			output.Add(strValue, nil, nil)
			return
		}
		if roundedValue > 0 && roundedValue < 1 {
			strValue = strings.TrimPrefix(strValue, "0")
		}
	}
	output.Add(strValue, nil, nil)
	d.Unit.GenCSS(context, output)
}

func (d *Dimension) ToCSS(context any) string {
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

func (d *Dimension) Operate(context any, op string, other *Dimension) *Dimension {
	value := d.OperateArithmetic(context, op, d.Value, other.Value)
	unit := d.Unit.Clone()

	if op == "+" || op == "-" {
		if len(unit.Numerator) == 0 && len(unit.Denominator) == 0 {
			unit.Release() // Release empty clone before replacing
			unit = other.Unit.Clone()
			if d.Unit.BackupUnit != "" {
				unit.BackupUnit = d.Unit.BackupUnit
			}
		} else if len(other.Unit.Numerator) == 0 && len(unit.Denominator) == 0 {
			// do nothing - use the first operand's unit
			// value is already calculated above
		} else {
			// Convert other to this dimension's units before operating
			usedUnits := d.Unit.UsedUnits()
			// Convert map[string]string to map[string]any for ConvertTo
			conversionMap := make(map[string]any)
			for k, v := range usedUnits {
				conversionMap[k] = v
			}
			otherConverted := other.ConvertTo(conversionMap)

			// Check strictUnits setting from context
			var strictUnits bool
			// First try Eval context (used during evaluation)
			if evalCtx, ok := context.(*Eval); ok {
				strictUnits = evalCtx.StrictUnits
			} else if ctx, ok := SafeTypeAssertion[map[string]any](context); ok {
				// Fallback to map-based context for compatibility (used during CSS generation)
				if strict, exists := SafeMapAccess(ctx, "strictUnits"); exists {
					if strictVal, ok := SafeTypeAssertion[bool](strict); ok {
						strictUnits = strictVal
					}
				}
			}

			if strictUnits && otherConverted.Unit.ToString() != unit.ToString() {
				// Match JavaScript: throw error for incompatible units in strict mode
				panic(&LessError{
					Type:    "Operation",
					Message: fmt.Sprintf("Incompatible units. Change the units or use the unit function. Bad units: '%s' and '%s'.", unit.ToString(), otherConverted.Unit.ToString()),
				})
			}
			// Recalculate value with converted units
			value = d.OperateArithmetic(context, op, d.Value, otherConverted.Value)
		}
	} else if op == "*" {
		unit.Numerator = append(unit.Numerator, other.Unit.Numerator...)
		unit.Denominator = append(unit.Denominator, other.Unit.Denominator...)
		sort.Strings(unit.Numerator)
		sort.Strings(unit.Denominator)
		unit.Cancel()
	} else if op == "/" {
		unit.Numerator = append(unit.Numerator, other.Unit.Denominator...)
		unit.Denominator = append(unit.Denominator, other.Unit.Numerator...)
		sort.Strings(unit.Numerator)
		sort.Strings(unit.Denominator)
		unit.Cancel()
	}
	return NewDimensionFrom(value, unit)
}

func (d *Dimension) OperateArithmetic(context any, op string, a, b float64) float64 {
	return d.Node.Operate(context, op, a, b)
}

func (d *Dimension) Compare(other any) *int {
	o, ok := other.(*Dimension)
	if !ok || o == nil {
		return nil
	}
	var a, b *Dimension
	var unifiedA, unifiedB *Unit // Track temp units for release
	if d.Unit.IsEmpty() || o.Unit.IsEmpty() {
		a = d
		b = o
	} else {
		a = d.Unify()
		b = o.Unify()
		unifiedA = a.Unit
		unifiedB = b.Unit
		if a.Unit.Compare(b.Unit) != 0 {
			unifiedA.Release()
			unifiedB.Release()
			return nil
		}
	}
	cmp := NumericCompare(a.Value, b.Value)
	if unifiedA != nil {
		unifiedA.Release()
	}
	if unifiedB != nil {
		unifiedB.Release()
	}
	return &cmp
}

func (d *Dimension) Unify() *Dimension {
	conv := map[string]any{ "length": "px", "duration": "s", "angle": "rad" }
	return d.ConvertTo(conv)
}

func (d *Dimension) ConvertTo(conversions any) *Dimension {
	value := d.Value
	unit := d.Unit.Clone()
	var convMap map[string]string

	switch t := conversions.(type) {
	case string:
		convMap = make(map[string]string)
		if _, ok := UnitConversionsLength[t]; ok {
			convMap["length"] = t
		}
		if _, ok := UnitConversionsDuration[t]; ok {
			convMap["duration"] = t
		}
		if _, ok := UnitConversionsAngle[t]; ok {
			convMap["angle"] = t
		}
	case map[string]any:
		convMap = make(map[string]string)
		for key, val := range t {
			if s, ok := val.(string); ok {
				convMap[key] = s
			}
		}
	default:
		return d
	}

	for groupName, targetUnit := range convMap {
		var group map[string]float64
		switch groupName {
		case "length":
			group = UnitConversionsLength
		case "duration":
			group = UnitConversionsDuration
		case "angle":
			group = UnitConversionsAngle
		default:
			continue
		}
		unit.Map(func(atomicUnit string, denominator bool) string {
			if factor, exists := group[atomicUnit]; exists {
				if targetFactor, ok := group[targetUnit]; ok {
					if denominator {
						value = value / (factor / targetFactor)
					} else {
						value = value * (factor / targetFactor)
					}
					return targetUnit
				}
			}
			return atomicUnit
		})
	}

	unit.Cancel()
	return NewDimensionFrom(value, unit)
} 