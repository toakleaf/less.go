package less_go

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

)

// Dimension represents a number with a unit
// It embeds Node and holds a numeric value and a unit.
type Dimension struct {
	*Node
	Value float64
	Unit  *Unit
}

// NewDimension creates a new Dimension instance.
// value can be a number (int, float64) or a numeric string. unit can be a string or *Unit. If unit is nil or empty string, an empty Unit is used.
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

// NewDimensionFrom is a helper constructor that creates a Dimension from a float64 value and *Unit without error handling.
func NewDimensionFrom(value float64, unit *Unit) *Dimension {
	d := &Dimension{
		Node:  NewNode(),
		Value: value,
		Unit:  unit,
	}
	d.SetParent(d.Unit, d.Node)
	return d
}

// GetType returns the node type
func (d *Dimension) GetType() string {
	return "Dimension"
}

// Accept accepts a visitor and updates the unit.
func (d *Dimension) Accept(visitor Visitor) {
	result := visitor.Visit(d.Unit)
	if u, ok := result.(*Unit); ok {
		d.Unit = u
	}
}

// Eval returns the dimension itself.
func (d *Dimension) Eval(context any) (any, error) {
	return d, nil
}

// ToColor converts the Dimension to a grayscale Color.
func (d *Dimension) ToColor() *Color {
	return NewColor([]float64{d.Value, d.Value, d.Value}, 1, "")
}

// GenCSS generates the CSS representation for the Dimension.
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
		// Instead of panicking, output an error comment in CSS
		output.Add(fmt.Sprintf("/* Error: Multiple units in dimension. Bad unit: %s */", d.Unit.ToString()), nil, nil)
		return
	}

	roundedValue := d.Fround(context, d.Value)
	strValue := fmt.Sprintf("%v", roundedValue)
	if roundedValue != 0 && math.Abs(roundedValue) < 0.000001 {
		// Mimic JavaScript's toFixed(20) and trim trailing zeros
		strValue = strings.TrimRight(fmt.Sprintf("%.20f", roundedValue), "0")
		strValue = strings.TrimRight(strValue, ".")
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

// ToCSS generates CSS string representation
func (d *Dimension) ToCSS(context any) string {
	var strs []string
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			strs = append(strs, fmt.Sprintf("%v", chunk))
		},
		IsEmpty: func() bool {
			return len(strs) == 0
		},
	}
	d.GenCSS(context, output)
	return strings.Join(strs, "")
}

// Operate performs an arithmetic operation between two Dimensions and returns a new Dimension.
func (d *Dimension) Operate(context any, op string, other *Dimension) *Dimension {
	value := d.OperateArithmetic(context, op, d.Value, other.Value)
	unit := d.Unit.Clone()

	if op == "+" || op == "-" {
		if len(unit.Numerator) == 0 && len(unit.Denominator) == 0 {
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
			if ctx, ok := SafeTypeAssertion[map[string]any](context); ok {
				if strict, exists := SafeMapAccess(ctx, "strictUnits"); exists {
					if strictVal, ok := SafeTypeAssertion[bool](strict); ok && strictVal {
						if otherConverted.Unit.ToString() != unit.ToString() {
							// Instead of panicking, return a dimension with error information
							// This maintains compatibility while avoiding panics
							return NewDimensionFrom(0, NewUnit(nil, nil, fmt.Sprintf("error-incompatible-units-%s-%s", unit.ToString(), otherConverted.Unit.ToString())))
						}
					}
				}
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

// OperateArithmetic wraps Node.Operate for performing arithmetic operations.
func (d *Dimension) OperateArithmetic(context any, op string, a, b float64) float64 {
	return d.Node.Operate(context, op, a, b)
}

// Compare compares the Dimension with another. It returns a pointer to int if comparable, or nil if not (simulating undefined in JS).
func (d *Dimension) Compare(other any) *int {
	o, ok := other.(*Dimension)
	if !ok || o == nil {
		return nil
	}
	var a, b *Dimension
	if d.Unit.IsEmpty() || o.Unit.IsEmpty() {
		a = d
		b = o
	} else {
		a = d.Unify()
		b = o.Unify()
		if a.Unit.Compare(b.Unit) != 0 {
			return nil
		}
	}
	cmp := NumericCompare(a.Value, b.Value)
	return &cmp
}

// Unify converts the Dimension to standard units.
func (d *Dimension) Unify() *Dimension {
	conv := map[string]any{ "length": "px", "duration": "s", "angle": "rad" }
	return d.ConvertTo(conv)
}

// ConvertTo converts the Dimension to specified units. 'conversions' may be a string or a map from group to target unit.
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