package less_go

import (
	"math"
)

type BlendMode func(cb, cs float64) float64

func ColorBlend(mode BlendMode, color1, color2 *Color) *Color {
	if color1 == nil || color2 == nil {
		return nil
	}

	ab := color1.Alpha        // backdrop alpha
	as := color2.Alpha        // source alpha
	var ar float64            // result alpha
	var cr float64            // result channel
	r := make([]float64, 3)   // result RGB

	// Calculate result alpha: as + ab * (1 - as)
	ar = as + ab*(1-as)

	// Blend each color channel
	for i := 0; i < 3; i++ {
		// Normalize to 0-1 range
		cb := color1.RGB[i] / 255.0
		cs := color2.RGB[i] / 255.0
		
		// Apply blend mode
		cr = mode(cb, cs)
		
		// Apply alpha compositing if result alpha > 0
		if ar > 0 {
			cr = (as*cs + ab*(cb - as*(cb+cs-cr))) / ar
		}
		
		// Denormalize back to 0-255 range
		r[i] = cr * 255.0
	}

	return NewColor(r, ar, "")
}

func Multiply(cb, cs float64) float64 {
	return cb * cs
}

func Screen(cb, cs float64) float64 {
	return cb + cs - cb*cs
}

func Overlay(cb, cs float64) float64 {
	cb *= 2
	if cb <= 1 {
		return Multiply(cb, cs)
	}
	return Screen(cb-1, cs)
}

func Softlight(cb, cs float64) float64 {
	d := 1.0
	e := cb
	
	if cs > 0.5 {
		e = 1.0
		if cb > 0.25 {
			d = math.Sqrt(cb)
		} else {
			d = ((16*cb-12)*cb+4)*cb
		}
	}
	
	return cb - (1-2*cs)*e*(d-cb)
}

func Hardlight(cb, cs float64) float64 {
	return Overlay(cs, cb)
}

func Difference(cb, cs float64) float64 {
	return math.Abs(cb - cs)
}

func Exclusion(cb, cs float64) float64 {
	return cb + cs - 2*cb*cs
}

func Average(cb, cs float64) float64 {
	return (cb + cs) / 2
}

func Negation(cb, cs float64) float64 {
	return 1 - math.Abs(cb+cs-1)
}

func ColorBlendMultiply(color1, color2 *Color) *Color {
	return ColorBlend(Multiply, color1, color2)
}

func ColorBlendScreen(color1, color2 *Color) *Color {
	return ColorBlend(Screen, color1, color2)
}

func ColorBlendOverlay(color1, color2 *Color) *Color {
	return ColorBlend(Overlay, color1, color2)
}

func ColorBlendSoftlight(color1, color2 *Color) *Color {
	return ColorBlend(Softlight, color1, color2)
}

func ColorBlendHardlight(color1, color2 *Color) *Color {
	return ColorBlend(Hardlight, color1, color2)
}

func ColorBlendDifference(color1, color2 *Color) *Color {
	return ColorBlend(Difference, color1, color2)
}

func ColorBlendExclusion(color1, color2 *Color) *Color {
	return ColorBlend(Exclusion, color1, color2)
}

func ColorBlendAverage(color1, color2 *Color) *Color {
	return ColorBlend(Average, color1, color2)
}

func ColorBlendNegation(color1, color2 *Color) *Color {
	return ColorBlend(Negation, color1, color2)
}

func GetColorBlendingFunctions() map[string]any {
	return map[string]any{
		"multiply":   ColorBlendMultiply,
		"screen":     ColorBlendScreen,
		"overlay":    ColorBlendOverlay,
		"softlight":  ColorBlendSoftlight,
		"hardlight":  ColorBlendHardlight,
		"difference": ColorBlendDifference,
		"exclusion":  ColorBlendExclusion,
		"average":    ColorBlendAverage,
		"negation":   ColorBlendNegation,
	}
}

func wrapBlendFunc(fn func(*Color, *Color) *Color) func(any, any) any {
	return func(color1, color2 any) any {
		// Convert first color
		c1 := toColor(color1)
		if c1 == nil {
			return nil
		}

		// Convert second color
		c2 := toColor(color2)
		if c2 == nil {
			return nil
		}

		return fn(c1, c2)
	}
}

// wrappedColorBlendingFunctions holds the pre-computed wrapped color blending functions map.
// Initialized once at package init time for efficiency.
var wrappedColorBlendingFunctions map[string]interface{}

func init() {
	// Get the raw blend functions
	blendFuncs := GetColorBlendingFunctions()

	// Wrap each blend function to handle any -> *Color conversion
	wrappedColorBlendingFunctions = make(map[string]interface{})
	for name, fn := range blendFuncs {
		if blendFn, ok := fn.(func(*Color, *Color) *Color); ok {
			wrappedColorBlendingFunctions[name] = &ColorFunctionWrapper{
				name: name,
				fn:   wrapBlendFunc(blendFn),
			}
		}
	}
}

// GetWrappedColorBlendingFunctions returns color blending functions wrapped for registry.
// The map is pre-computed at init time and cached for efficiency.
func GetWrappedColorBlendingFunctions() map[string]interface{} {
	return wrappedColorBlendingFunctions
}