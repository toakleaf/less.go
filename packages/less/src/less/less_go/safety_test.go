package less_go

import (
	"testing"
)

func TestSafeStringIndex(t *testing.T) {
	str := "hello"
	if char, ok := SafeStringIndex(str, 0); !ok || char != 'h' {
		t.Errorf("Expected 'h', got %c", char)
	}

	if char, ok := SafeStringIndex(str, 10); ok {
		t.Errorf("Expected false for out of bounds access, got true with char %c", char)
	}

	if char, ok := SafeStringIndex(str, -1); ok {
		t.Errorf("Expected false for negative index, got true with char %c", char)
	}

	if char, ok := SafeStringIndex("", 0); ok {
		t.Errorf("Expected false for empty string access, got true with char %c", char)
	}
}

func TestSafeStringSlice(t *testing.T) {
	str := "hello world"

	if slice, ok := SafeStringSlice(str, 0, 5); !ok || slice != "hello" {
		t.Errorf("Expected 'hello', got '%s'", slice)
	}

	if slice, ok := SafeStringSlice(str, 0, 100); ok {
		t.Errorf("Expected false for out of bounds slice, got true with slice '%s'", slice)
	}

	if slice, ok := SafeStringSlice(str, 5, 2); ok {
		t.Errorf("Expected false for invalid slice, got true with slice '%s'", slice)
	}
}

func TestSafeSliceIndex(t *testing.T) {
	slice := []any{"a", "b", "c"}

	if val, ok := SafeSliceIndex(slice, 1); !ok || val != "b" {
		t.Errorf("Expected 'b', got %v", val)
	}

	if val, ok := SafeSliceIndex(slice, 10); ok {
		t.Errorf("Expected false for out of bounds access, got true with val %v", val)
	}

	if val, ok := SafeSliceIndex(nil, 0); ok {
		t.Errorf("Expected false for nil slice access, got true with val %v", val)
	}
}

func TestSafeTypeAssertion(t *testing.T) {
	var val any = "hello"

	if str, ok := SafeTypeAssertion[string](val); !ok || str != "hello" {
		t.Errorf("Expected 'hello', got '%s'", str)
	}

	if num, ok := SafeTypeAssertion[int](val); ok {
		t.Errorf("Expected false for failed assertion, got true with num %d", num)
	}

	if str, ok := SafeTypeAssertion[string](nil); ok {
		t.Errorf("Expected false for nil assertion, got true with str '%s'", str)
	}
}

func TestSafeNilCheck(t *testing.T) {
	if !SafeNilCheck(nil) {
		t.Error("Expected true for nil value")
	}

	str := "hello"
	if SafeNilCheck(str) {
		t.Error("Expected false for non-nil value")
	}

	var ptr *string
	if !SafeNilCheck(ptr) {
		t.Error("Expected true for nil pointer")
	}

	str2 := "world"
	ptr2 := &str2
	if SafeNilCheck(ptr2) {
		t.Error("Expected false for non-nil pointer")
	}
}

func TestRecoverableOperation(t *testing.T) {
	result, err := RecoverableOperation(func() int {
		return 42
	})
	if err != nil || result != 42 {
		t.Errorf("Expected 42 with no error, got %d with error %v", result, err)
	}

	result2, err2 := RecoverableOperation(func() string {
		panic("test panic")
	})
	if err2 == nil {
		t.Error("Expected error from panicking operation")
	}
	if result2 != "" {
		t.Errorf("Expected empty string result from panicking operation, got %s", result2)
	}
}

func TestSafeToCSS(t *testing.T) {
	result := SafeToCSS(nil, nil)
	if result != "" {
		t.Errorf("Expected empty string for nil value, got '%s'", result)
	}

	result2 := SafeToCSS("hello", nil)
	if result2 != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result2)
	}
}

func TestParserInputSafety(t *testing.T) {
	input := NewParserInput()
	input.Start("test", false, nil)

	char := input.CurrentChar()
	_ = char

	prev := input.PrevChar()
	_ = prev

	isWhite := input.IsWhitespace(100)
	_ = isWhite

	input.i = 1000
	result := input.PeekChar('x')
	if result != nil {
		t.Error("Expected nil for out of bounds PeekChar")
	}
}

func TestChunkerSafety(t *testing.T) {
	testInputs := []string{
		"",
		"a",
		"\\",
		"/*",
		"'unclosed string",
		"test /* comment */ more",
	}

	for _, input := range testInputs {
		t.Run("Input: "+input, func(t *testing.T) {
			chunks := Chunker(input, func(msg string, pos int) {})
			_ = chunks
		})
	}
} 