package less_go

import (
	"sort"
	"strings"
)

type Unit struct {
	*Node
	Numerator   []string
	Denominator []string
	BackupUnit  string
}

func NewUnit(numerator []string, denominator []string, backupUnit string) *Unit {
	u := GetUnitFromPool()
	u.Node = NewNode()

	if numerator != nil {
		numLen := len(numerator)
		// Reuse pooled slice capacity if sufficient, otherwise allocate
		if cap(u.Numerator) >= numLen {
			u.Numerator = u.Numerator[:numLen]
		} else {
			u.Numerator = make([]string, numLen)
		}
		copy(u.Numerator, numerator)
		sort.Strings(u.Numerator)
	}

	if denominator != nil {
		denLen := len(denominator)
		// Reuse pooled slice capacity if sufficient, otherwise allocate
		if cap(u.Denominator) >= denLen {
			u.Denominator = u.Denominator[:denLen]
		} else {
			u.Denominator = make([]string, denLen)
		}
		copy(u.Denominator, denominator)
		sort.Strings(u.Denominator)
	}

	if backupUnit != "" {
		u.BackupUnit = backupUnit
	} else if len(numerator) > 0 {
		u.BackupUnit = numerator[0]
	}

	return u
}

func (u *Unit) Type() string {
	return "Unit"
}

func (u *Unit) Clone() *Unit {
	newNum := make([]string, len(u.Numerator))
	copy(newNum, u.Numerator)
	newDen := make([]string, len(u.Denominator))
	copy(newDen, u.Denominator)

	return NewUnit(newNum, newDen, u.BackupUnit)
}

func (u *Unit) GenCSS(context any, output *CSSOutput) {
	strictUnits := false
	if ctx, ok := context.(map[string]any); ok {
		if strict, ok := ctx["strictUnits"].(bool); ok {
			strictUnits = strict
		}
	}

	if len(u.Numerator) == 1 {
		output.Add(u.Numerator[0], nil, nil)
	} else if !strictUnits && u.BackupUnit != "" {
		output.Add(u.BackupUnit, nil, nil)
	} else if !strictUnits && len(u.Denominator) > 0 {
		output.Add(u.Denominator[0], nil, nil)
	}
}

func (u *Unit) ToString() string {
	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.WriteString(strings.Join(u.Numerator, "*"))
	for i := 0; i < len(u.Denominator); i++ {
		builder.WriteString("/")
		builder.WriteString(u.Denominator[i])
	}
	return builder.String()
}

func (u *Unit) Compare(other *Unit) int {
	if other == nil {
		return 999 // undefined equivalent in JavaScript
	}

	if u.Is(other.ToString()) {
		return 0
	}

	return 999 // undefined equivalent in JavaScript
}

func (u *Unit) Is(unitString string) bool {
	return strings.EqualFold(u.ToString(), unitString)
}

func (u *Unit) IsLength() bool {
	// Match JavaScript: RegExp('^(px|em|ex|ch|rem|in|cm|mm|pc|pt|ex|vw|vh|vmin|vmax)$', 'gi').test(this.toCSS())
	// Note: JavaScript has 'ex' twice in the regex
	css := u.ToCSS(nil)
	lengthUnits := []string{
		"px", "em", "ex", "ch", "rem", "in", "cm", "mm", "pc", "pt",
		"vw", "vh", "vmin", "vmax",
	}
	
	for _, lengthUnit := range lengthUnits {
		if strings.EqualFold(css, lengthUnit) {
			return true
		}
	}
	return false
}

func (u *Unit) ToCSS(context any) string {
	var chunks []string
	output := &CSSOutput{
		Add: func(chunk any, fileInfo any, index any) {
			if chunk != nil {
				chunks = append(chunks, chunk.(string))
			}
		},
		IsEmpty: func() bool {
			return len(chunks) == 0
		},
	}

	u.GenCSS(context, output)
	return strings.Join(chunks, "")
}

func (u *Unit) IsEmpty() bool {
	return len(u.Numerator) == 0 && len(u.Denominator) == 0
}

func (u *Unit) IsSingular() bool {
	return len(u.Numerator) <= 1 && len(u.Denominator) == 0
}

func (u *Unit) Map(callback func(string, bool) string) {
	// Transform values in-place to avoid allocations
	for i := range u.Numerator {
		u.Numerator[i] = callback(u.Numerator[i], false)
	}

	for i := range u.Denominator {
		u.Denominator[i] = callback(u.Denominator[i], true)
	}

	sort.Strings(u.Numerator)
	sort.Strings(u.Denominator)
}

func (u *Unit) UsedUnits() map[string]string {
	result := make(map[string]string)

	conversions := map[string]map[string]float64{
		"length":   UnitConversionsLength,
		"duration": UnitConversionsDuration,
		"angle":    UnitConversionsAngle,
	}

	for groupName, group := range conversions {
		u.Map(func(atomicUnit string, isDenominator bool) string {
			if _, exists := group[atomicUnit]; exists {
				if _, hasResult := result[groupName]; !hasResult {
					result[groupName] = atomicUnit
				}
			}
			return atomicUnit
		})
	}

	return result
}

func (u *Unit) Cancel() {
	counter := make(map[string]int)

	for _, unit := range u.Numerator {
		counter[unit]++
	}

	for _, unit := range u.Denominator {
		counter[unit]--
	}

	// Reuse existing slice capacity instead of allocating new slices
	u.Numerator = u.Numerator[:0]
	u.Denominator = u.Denominator[:0]

	for unit, count := range counter {
		if count > 0 {
			for i := 0; i < count; i++ {
				u.Numerator = append(u.Numerator, unit)
			}
		} else if count < 0 {
			for i := 0; i < -count; i++ {
				u.Denominator = append(u.Denominator, unit)
			}
		}
	}

	sort.Strings(u.Numerator)
	sort.Strings(u.Denominator)
} 