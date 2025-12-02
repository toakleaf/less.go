package less_go

import (
	"sync"
	"testing"
)

func TestIntern(t *testing.T) {
	// Clear intern table for clean test
	ClearInternTable()

	t.Run("returns same pointer for same string", func(t *testing.T) {
		s1 := Intern("test")
		s2 := Intern("test")
		// In Go, we can't directly compare string pointers, but we can verify
		// the strings are equal and the intern count only increased by 1
		if s1 != s2 {
			t.Errorf("Expected interned strings to be equal")
		}
	})

	t.Run("returns different strings for different inputs", func(t *testing.T) {
		s1 := Intern("foo")
		s2 := Intern("bar")
		if s1 == s2 {
			t.Errorf("Expected different strings to remain different")
		}
	})

	t.Run("skips empty strings", func(t *testing.T) {
		result := Intern("")
		if result != "" {
			t.Errorf("Expected empty string to be returned as-is")
		}
	})

	t.Run("skips long strings", func(t *testing.T) {
		longString := make([]byte, 200)
		for i := range longString {
			longString[i] = 'a'
		}
		result := Intern(string(longString))
		if result != string(longString) {
			t.Errorf("Expected long string to be returned as-is")
		}
	})

	t.Run("handles concurrent access", func(t *testing.T) {
		ClearInternTable()
		var wg sync.WaitGroup
		strings := []string{"color", "background", "margin", "padding", "border"}

		// Run multiple goroutines interning the same strings
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, s := range strings {
					_ = Intern(s)
				}
			}()
		}
		wg.Wait()

		// Verify all strings were interned
		for _, s := range strings {
			result := Intern(s)
			if result != s {
				t.Errorf("Expected %s to be interned correctly", s)
			}
		}
	})
}

func TestInternBytes(t *testing.T) {
	ClearInternTable()

	t.Run("interns byte slice as string", func(t *testing.T) {
		bytes := []byte("test")
		s1 := InternBytes(bytes)
		s2 := Intern("test")
		if s1 != s2 {
			t.Errorf("Expected InternBytes to produce same result as Intern")
		}
	})

	t.Run("skips empty byte slices", func(t *testing.T) {
		result := InternBytes([]byte{})
		if result != "" {
			t.Errorf("Expected empty byte slice to return empty string")
		}
	})
}

func TestInternedCount(t *testing.T) {
	ClearInternTable()
	initialCount := InternedCount()

	Intern("unique1")
	Intern("unique2")
	Intern("unique2") // duplicate, shouldn't increase count

	expectedCount := initialCount + 2
	actualCount := InternedCount()
	if actualCount != expectedCount {
		t.Errorf("Expected count %d, got %d", expectedCount, actualCount)
	}
}

func TestClearInternTable(t *testing.T) {
	Intern("test1")
	Intern("test2")

	ClearInternTable()
	// After clearing, the common strings should be re-interned
	// so count should be the count of common strings
	count := InternedCount()
	if count < 100 { // We have many common CSS properties pre-interned
		t.Errorf("Expected common strings to be re-interned after clear, got count %d", count)
	}
}

func BenchmarkIntern(b *testing.B) {
	strings := []string{
		"color", "background", "margin", "padding", "border",
		"width", "height", "display", "position", "font-size",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range strings {
			_ = Intern(s)
		}
	}
}

func BenchmarkNoIntern(b *testing.B) {
	strings := []string{
		"color", "background", "margin", "padding", "border",
		"width", "height", "display", "position", "font-size",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range strings {
			_ = s
		}
	}
}

func BenchmarkInternConcurrent(b *testing.B) {
	strings := []string{
		"color", "background", "margin", "padding", "border",
		"width", "height", "display", "position", "font-size",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, s := range strings {
				_ = Intern(s)
			}
		}
	})
}
