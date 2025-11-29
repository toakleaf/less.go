package runtime

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// Binary format for prefetched variables buffer:
//
// Header:
//   [4 bytes] magic: 0x50524546 ("PREF")
//   [4 bytes] version: 1
//   [4 bytes] variable count
//
// For each variable:
//   [4 bytes] name length
//   [N bytes] name (UTF-8)
//   [1 byte]  important flag (0 or 1)
//   [1 byte]  type (0=null, 1=Dimension, 2=Color, 3=Quoted, 4=Keyword, 5=Expression)
//   [variable] value data (type-specific, see below)
//
// Type-specific value encodings:
//   Dimension (type=1):
//     [8 bytes] value (float64 LE)
//     [4 bytes] unit length
//     [N bytes] unit string (UTF-8)
//
//   Color (type=2):
//     [8 bytes] R (float64 LE)
//     [8 bytes] G (float64 LE)
//     [8 bytes] B (float64 LE)
//     [8 bytes] alpha (float64 LE)
//
//   Quoted (type=3):
//     [4 bytes] string length
//     [N bytes] string value (UTF-8)
//     [1 byte]  quote character
//     [1 byte]  escaped flag (0 or 1)
//
//   Keyword (type=4):
//     [4 bytes] string length
//     [N bytes] string value (UTF-8)
//
//   Expression (type=5):
//     [4 bytes] value count
//     For each value:
//       [1 byte]  type
//       [variable] value data

const (
	PrefetchMagic   uint32 = 0x50524546 // "PREF"
	PrefetchVersion uint32 = 1

	// Variable types
	VarTypeNull       byte = 0
	VarTypeDimension  byte = 1
	VarTypeColor      byte = 2
	VarTypeQuoted     byte = 3
	VarTypeKeyword    byte = 4
	VarTypeExpression byte = 5
	VarTypeAnonymous  byte = 6
	VarTypeVariable   byte = 7 // Variable reference (name only, needs lookup)
)

// BinaryVariableWriter writes variables in binary format to a buffer.
type BinaryVariableWriter struct {
	buf []byte
}

// NewBinaryVariableWriter creates a new binary variable writer.
func NewBinaryVariableWriter() *BinaryVariableWriter {
	return &BinaryVariableWriter{
		buf: make([]byte, 0, 4096),
	}
}

// Reset clears the buffer for reuse.
func (w *BinaryVariableWriter) Reset() {
	w.buf = w.buf[:0]
}

// WriteHeader writes the prefetch buffer header.
func (w *BinaryVariableWriter) WriteHeader(varCount uint32) {
	w.writeUint32(PrefetchMagic)
	w.writeUint32(PrefetchVersion)
	w.writeUint32(varCount)
}

// WriteVariable writes a single variable to the buffer.
// Returns the number of bytes written.
func (w *BinaryVariableWriter) WriteVariable(name string, decl any) (int, error) {
	startLen := len(w.buf)

	// Write variable name
	w.writeString(name)

	if decl == nil {
		// Write null variable
		w.buf = append(w.buf, 0) // not important
		w.buf = append(w.buf, VarTypeNull)
		return len(w.buf) - startLen, nil
	}

	// Get the value from the declaration
	valueProvider, ok := decl.(interface{ GetValue() any })
	if !ok {
		// Not a valid declaration - debug output
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[WriteVariable] %s: decl type %T does not have GetValue() any\n", name, decl)
		}
		w.buf = append(w.buf, 0) // not important
		w.buf = append(w.buf, VarTypeNull)
		return len(w.buf) - startLen, nil
	}

	value := valueProvider.GetValue()
	if value == nil {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[WriteVariable] %s: GetValue() returned nil\n", name)
		}
		w.buf = append(w.buf, 0) // not important
		w.buf = append(w.buf, VarTypeNull)
		return len(w.buf) - startLen, nil
	}

	// Get important flag
	important := byte(0)
	if ip, ok := decl.(interface{ GetImportant() bool }); ok && ip.GetImportant() {
		important = 1
	} else if ip, ok := decl.(interface{ GetImportant() string }); ok && ip.GetImportant() != "" {
		important = 1
	}
	w.buf = append(w.buf, important)

	// Write value based on type
	if err := w.writeValue(value); err != nil {
		// Reset to before this variable and write as null
		w.buf = w.buf[:startLen]
		w.writeString(name)
		w.buf = append(w.buf, important)
		w.buf = append(w.buf, VarTypeNull)
		return len(w.buf) - startLen, nil
	}

	return len(w.buf) - startLen, nil
}

// writeValue writes a value node in binary format.
func (w *BinaryVariableWriter) writeValue(value any) error {
	if value == nil {
		w.buf = append(w.buf, VarTypeNull)
		return nil
	}

	// Try to get type from the value
	typer, ok := value.(interface{ GetType() string })
	if !ok {
		// Not a typed node - check for primitive types
		switch v := value.(type) {
		case string:
			w.buf = append(w.buf, VarTypeKeyword)
			w.writeString(v)
			return nil
		case float64:
			w.buf = append(w.buf, VarTypeDimension)
			w.writeFloat64(v)
			w.writeString("") // no unit
			return nil
		case int:
			w.buf = append(w.buf, VarTypeDimension)
			w.writeFloat64(float64(v))
			w.writeString("") // no unit
			return nil
		default:
			return fmt.Errorf("unsupported primitive type: %T", value)
		}
	}

	nodeType := typer.GetType()
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[writeValue] nodeType=%s, value type=%T\n", nodeType, value)
	}
	switch nodeType {
	case "Value":
		// Value is a wrapper around an array of values
		// Unwrap it and write the inner value(s)
		if getter, ok := value.(interface{ GetValue() []any }); ok {
			arr := getter.GetValue()
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[writeValue] Value wrapper: %d inner values\n", len(arr))
				for i, v := range arr {
					if t, ok := v.(interface{ GetType() string }); ok {
						fmt.Printf("[writeValue]   inner[%d] type=%s\n", i, t.GetType())
					} else {
						fmt.Printf("[writeValue]   inner[%d] type=%T\n", i, v)
					}
				}
			}
			if len(arr) == 1 {
				// Single value - unwrap and write
				return w.writeValue(arr[0])
			} else if len(arr) > 1 {
				// Multiple values - write as Expression
				beforeLen := len(w.buf)
				w.buf = append(w.buf, VarTypeExpression)
				w.writeUint32(uint32(len(arr)))
				for i, v := range arr {
					subBefore := len(w.buf)
					if err := w.writeValue(v); err != nil {
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							fmt.Printf("[writeValue] Value inner[%d] failed: %v, writing null\n", i, err)
						}
						w.buf = append(w.buf, VarTypeNull)
					}
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Printf("[writeValue] Value inner[%d] wrote %d bytes (type=%d at offset %d)\n", i, len(w.buf)-subBefore, w.buf[subBefore], subBefore)
					}
				}
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[writeValue] Value multi total: %d bytes\n", len(w.buf)-beforeLen)
				}
				return nil
			}
		}
		return fmt.Errorf("Value type with no inner values")

	case "Dimension":
		w.buf = append(w.buf, VarTypeDimension)
		if err := w.writeDimension(value); err != nil {
			return err
		}

	case "Color":
		w.buf = append(w.buf, VarTypeColor)
		if err := w.writeColor(value); err != nil {
			return err
		}

	case "Quoted":
		w.buf = append(w.buf, VarTypeQuoted)
		if err := w.writeQuoted(value); err != nil {
			return err
		}

	case "Keyword":
		w.buf = append(w.buf, VarTypeKeyword)
		if err := w.writeKeyword(value); err != nil {
			return err
		}

	case "Anonymous":
		w.buf = append(w.buf, VarTypeAnonymous)
		if err := w.writeAnonymous(value); err != nil {
			return err
		}

	case "Expression":
		// Check if Expression has single value - if so, unwrap it
		if getter, ok := value.(interface{ GetValue() []any }); ok {
			arr := getter.GetValue()
			if len(arr) == 1 {
				// Unwrap single-value Expression
				return w.writeValue(arr[0])
			}
		}
		w.buf = append(w.buf, VarTypeExpression)
		if err := w.writeExpression(value); err != nil {
			return err
		}

	case "Variable":
		// Variable nodes are unevaluated references
		// Write the variable name so JavaScript can look it up
		w.buf = append(w.buf, VarTypeVariable)
		varName := ""
		if getter, ok := value.(interface{ GetName() string }); ok {
			varName = getter.GetName()
		}
		w.writeString(varName)

	case "Combinator", "Element", "Selector", "Ruleset", "MixinCall", "MixinDefinition", "DetachedRuleset":
		// Skip AST structural nodes that can't be serialized as values
		w.buf = append(w.buf, VarTypeNull)

	default:
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[writeValue] Unsupported node type: %s\n", nodeType)
		}
		return fmt.Errorf("unsupported node type: %s", nodeType)
	}

	return nil
}

// writeDimension writes a Dimension node.
func (w *BinaryVariableWriter) writeDimension(value any) error {
	// Get numeric value
	if getter, ok := value.(interface{ GetValue() float64 }); ok {
		w.writeFloat64(getter.GetValue())
	} else {
		w.writeFloat64(0)
	}

	// Get unit string
	unitStr := ""
	if unitGetter, ok := value.(interface{ GetUnit() any }); ok {
		unit := unitGetter.GetUnit()
		if unit != nil {
			if s, ok := unit.(fmt.Stringer); ok {
				unitStr = s.String()
			} else if s, ok := unit.(interface{ ToString() string }); ok {
				unitStr = s.ToString()
			}
		}
	}
	w.writeString(unitStr)

	return nil
}

// writeColor writes a Color node.
// Format: [R:f64][G:f64][B:f64][Alpha:f64][ValueLen:u32][Value:string]
// The Value field preserves the original color form (e.g., "#fff" vs "#ffffff")
func (w *BinaryVariableWriter) writeColor(value any) error {
	// Get RGB values
	rgb := []float64{0, 0, 0}
	if rgbGetter, ok := value.(interface{ GetRGB() []float64 }); ok {
		rgb = rgbGetter.GetRGB()
	}

	// Write R, G, B
	for i := 0; i < 3; i++ {
		if i < len(rgb) {
			w.writeFloat64(rgb[i])
		} else {
			w.writeFloat64(0)
		}
	}

	// Get alpha
	alpha := 1.0
	if alphaGetter, ok := value.(interface{ GetAlpha() float64 }); ok {
		alpha = alphaGetter.GetAlpha()
	}
	w.writeFloat64(alpha)

	// Get original value string (preserves form like "#fff" vs "#ffffff")
	originalValue := ""
	if valueGetter, ok := value.(interface{ GetColorValue() string }); ok {
		originalValue = valueGetter.GetColorValue()
	}
	w.writeString(originalValue)

	return nil
}

// writeQuoted writes a Quoted node.
func (w *BinaryVariableWriter) writeQuoted(value any) error {
	// Get string value
	strVal := ""
	if getter, ok := value.(interface{ GetValue() string }); ok {
		strVal = getter.GetValue()
	}
	w.writeString(strVal)

	// Get quote character
	quote := '"'
	if getter, ok := value.(interface{ GetQuote() string }); ok {
		q := getter.GetQuote()
		if len(q) > 0 {
			quote = rune(q[0])
		}
	}
	w.buf = append(w.buf, byte(quote))

	// Get escaped flag
	escaped := byte(0)
	if getter, ok := value.(interface{ GetEscaped() bool }); ok && getter.GetEscaped() {
		escaped = 1
	}
	w.buf = append(w.buf, escaped)

	return nil
}

// writeKeyword writes a Keyword node.
func (w *BinaryVariableWriter) writeKeyword(value any) error {
	strVal := ""
	if getter, ok := value.(interface{ GetValue() string }); ok {
		strVal = getter.GetValue()
	}
	w.writeString(strVal)
	return nil
}

// writeAnonymous writes an Anonymous node.
func (w *BinaryVariableWriter) writeAnonymous(value any) error {
	// Anonymous can have any value type
	if getter, ok := value.(interface{ GetValue() any }); ok {
		v := getter.GetValue()
		switch val := v.(type) {
		case string:
			w.writeString(val)
		case float64:
			w.writeString(fmt.Sprintf("%g", val))
		case int:
			w.writeString(fmt.Sprintf("%d", val))
		default:
			w.writeString(fmt.Sprintf("%v", v))
		}
	} else {
		w.writeString("")
	}
	return nil
}

// writeExpression writes an Expression node.
func (w *BinaryVariableWriter) writeExpression(value any) error {
	// Get array of values
	var values []any
	if getter, ok := value.(interface{ GetValue() []any }); ok {
		values = getter.GetValue()
	}

	// Write value count
	w.writeUint32(uint32(len(values)))

	// Write each value recursively
	for _, v := range values {
		if err := w.writeValue(v); err != nil {
			// Write null for unsupported values
			w.buf = append(w.buf, VarTypeNull)
		}
	}

	return nil
}

// Bytes returns the written buffer.
func (w *BinaryVariableWriter) Bytes() []byte {
	return w.buf
}

// Helper functions for writing primitive types

func (w *BinaryVariableWriter) writeUint32(v uint32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	w.buf = append(w.buf, b...)
}

func (w *BinaryVariableWriter) writeFloat64(v float64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, math.Float64bits(v))
	w.buf = append(w.buf, b...)
}

func (w *BinaryVariableWriter) writeString(s string) {
	w.writeUint32(uint32(len(s)))
	w.buf = append(w.buf, s...)
}

// WritePrefetchedVariables writes a map of prefetched variables to binary format.
// Returns the binary buffer ready to be written to shared memory.
func WritePrefetchedVariables(variables map[string]any) []byte {
	w := NewBinaryVariableWriter()

	// Count valid variables first
	validVars := make([]struct {
		name string
		decl any
	}, 0, len(variables))

	for name, decl := range variables {
		if decl != nil {
			validVars = append(validVars, struct {
				name string
				decl any
			}{name, decl})
		}
	}

	// Write header
	w.WriteHeader(uint32(len(validVars)))

	// Write each variable
	for _, v := range validVars {
		w.WriteVariable(v.name, v.decl)
	}

	return w.Bytes()
}
