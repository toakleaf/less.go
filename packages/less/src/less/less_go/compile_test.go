package less_go

import (
	"strings"
	"testing"
)

func TestCompile_BasicInput(t *testing.T) {
	input := `.test { color: red; }`
	result, err := Compile(input, nil)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !strings.Contains(result.CSS, ".test") {
		t.Errorf("CSS should contain '.test': %s", result.CSS)
	}

	if !strings.Contains(result.CSS, "color") {
		t.Errorf("CSS should contain 'color': %s", result.CSS)
	}
}

func TestCompile_WithVariables(t *testing.T) {
	input := `@color: blue; .test { color: @color; }`
	result, err := Compile(input, nil)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !strings.Contains(result.CSS, "blue") {
		t.Errorf("CSS should contain 'blue': %s", result.CSS)
	}
}

func TestCompile_WithNesting(t *testing.T) {
	input := `.parent { .child { color: red; } }`
	result, err := Compile(input, nil)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !strings.Contains(result.CSS, ".parent .child") {
		t.Errorf("CSS should contain '.parent .child': %s", result.CSS)
	}
}

func TestCompile_WithMixin(t *testing.T) {
	input := `.mixin() { color: red; } .test { .mixin(); }`
	result, err := Compile(input, nil)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !strings.Contains(result.CSS, ".test") {
		t.Errorf("CSS should contain '.test': %s", result.CSS)
	}
}

func TestCompile_WithOptions(t *testing.T) {
	input := `.test { color: red; }`
	result, err := Compile(input, &CompileOptions{
		Filename: "test.less",
	})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}
}

func TestCompile_WithPluginSupport(t *testing.T) {
	// This test verifies that enabling plugins doesn't break compilation
	// when no plugins are actually used
	input := `.test { color: red; }`
	result, err := Compile(input, &CompileOptions{
		EnableJavaScriptPlugins: true,
	})
	if err != nil {
		t.Fatalf("Compile with plugin support failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if !strings.Contains(result.CSS, ".test") {
		t.Errorf("CSS should contain '.test': %s", result.CSS)
	}
}

func TestCompileOptions_Conversion(t *testing.T) {
	options := &CompileOptions{
		Paths:       []string{"/path/one", "/path/two"},
		Filename:    "test.less",
		Compress:    true,
		StrictUnits: true,
		Math:        MathParens,
		RewriteUrls: RewriteUrlsAll,
		Rootpath:    "/root",
		UrlArgs:     "v=1",
	}

	converted := convertCompileOptionsToMap(options)

	if paths, ok := converted["paths"].([]string); !ok || len(paths) != 2 {
		t.Error("Paths not converted correctly")
	}

	if converted["filename"] != "test.less" {
		t.Error("Filename not converted correctly")
	}

	if converted["compress"] != true {
		t.Error("Compress not converted correctly")
	}

	if converted["strictUnits"] != true {
		t.Error("StrictUnits not converted correctly")
	}

	if converted["math"] != MathParens {
		t.Error("Math not converted correctly")
	}

	if converted["rewriteUrls"] != RewriteUrlsAll {
		t.Error("RewriteUrls not converted correctly")
	}

	if converted["rootpath"] != "/root" {
		t.Error("Rootpath not converted correctly")
	}

	if converted["urlArgs"] != "v=1" {
		t.Error("UrlArgs not converted correctly")
	}
}

func TestCompileOptions_EmptyConversion(t *testing.T) {
	options := &CompileOptions{}
	converted := convertCompileOptionsToMap(options)

	// Empty options should result in empty map for most fields
	if len(converted) != 0 {
		t.Errorf("Empty options should produce empty map, got %d entries", len(converted))
	}
}
