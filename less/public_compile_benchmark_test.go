package less_go

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkPublicCompile covers the main exported API. BenchmarkLargeSuite is
// intentionally a lower-level compiler-core benchmark that reuses a Factory.
func BenchmarkPublicCompile(b *testing.B) {
	filename := filepath.Join("../testdata/less/_main/functions.less")
	input, err := os.ReadFile(filename)
	if err != nil {
		b.Fatal(err)
	}
	source := string(input)

	options := &CompileOptions{
		Filename:          filename,
		JavascriptEnabled: true,
	}

	for i := 0; i < 5; i++ {
		if _, err := Compile(source, options); err != nil {
			b.Fatal(err)
		}
	}

	b.SetBytes(int64(len(source)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Compile(source, options); err != nil {
			b.Fatal(err)
		}
	}
}
