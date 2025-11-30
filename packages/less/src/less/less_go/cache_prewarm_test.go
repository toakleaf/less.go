package less_go

import (
	"testing"
)

func TestPluginCallCollector_BasicCollection(t *testing.T) {
	// Create some Call nodes that would be plugin functions
	pluginCall1 := NewCall("map-get", []any{
		NewKeyword("colors"),
		NewKeyword("primary"),
	}, 0, nil)

	pluginCall2 := NewCall("color-yiq", []any{
		NewColor([]float64{255, 255, 255}, 1.0, "#fff"),
	}, 0, nil)

	// Create a non-plugin call
	builtinCall := NewCall("darken", []any{
		NewColor([]float64{255, 0, 0}, 1.0, "red"),
		NewDimensionFrom(10.0, &Unit{Numerator: []string{"%"}}),
	}, 0, nil)

	// Create a simple ruleset with these calls
	ruleset := NewRuleset(nil, []any{
		pluginCall1,
		pluginCall2,
		builtinCall,
	}, false, nil)

	// Create collector with only plugin function names
	collector := NewPluginCallCollector([]string{"map-get", "color-yiq", "breakpoint-next"})

	// Collect calls
	calls := collector.Collect(ruleset)

	// Should have collected 2 plugin calls (not the darken call)
	if len(calls) != 2 {
		t.Errorf("Expected 2 plugin calls, got %d", len(calls))
	}

	// Check that we got the right functions
	foundMapGet := false
	foundColorYiq := false
	for _, call := range calls {
		if call.FunctionName == "map-get" {
			foundMapGet = true
		}
		if call.FunctionName == "color-yiq" {
			foundColorYiq = true
		}
	}

	if !foundMapGet {
		t.Error("Expected to find map-get call")
	}
	if !foundColorYiq {
		t.Error("Expected to find color-yiq call")
	}
}

func TestPluginCallCollector_Deduplication(t *testing.T) {
	// Create duplicate calls with same arguments
	call1 := NewCall("map-get", []any{
		NewKeyword("colors"),
		NewKeyword("primary"),
	}, 0, nil)

	call2 := NewCall("map-get", []any{
		NewKeyword("colors"),
		NewKeyword("primary"),
	}, 10, nil) // Different index, same args

	call3 := NewCall("map-get", []any{
		NewKeyword("colors"),
		NewKeyword("secondary"), // Different arg
	}, 20, nil)

	ruleset := NewRuleset(nil, []any{call1, call2, call3}, false, nil)

	collector := NewPluginCallCollector([]string{"map-get"})
	calls := collector.Collect(ruleset)

	// Should deduplicate - only 2 unique calls
	if len(calls) != 2 {
		t.Errorf("Expected 2 unique calls (deduplicated), got %d", len(calls))
	}
}

func TestPluginCallCollector_NestedCalls(t *testing.T) {
	// Create a nested structure with plugin calls at different levels
	innerCall := NewCall("color-yiq", []any{
		NewColor([]float64{0, 0, 0}, 1.0, "#000"),
	}, 0, nil)

	innerRuleset := NewRuleset(nil, []any{innerCall}, false, nil)

	outerCall := NewCall("map-get", []any{
		NewKeyword("theme-colors"),
		NewKeyword("dark"),
	}, 0, nil)

	outerRuleset := NewRuleset(nil, []any{outerCall, innerRuleset}, false, nil)

	collector := NewPluginCallCollector([]string{"map-get", "color-yiq"})
	calls := collector.Collect(outerRuleset)

	// Should find both nested calls
	if len(calls) != 2 {
		t.Errorf("Expected 2 calls from nested structure, got %d", len(calls))
	}
}

func TestPluginCallCollector_SkipsVariableArgs(t *testing.T) {
	// Create a call with a variable argument - can't be cached
	callWithVar := NewCall("map-get", []any{
		NewVariable("@colors", 0, nil),
		NewKeyword("primary"),
	}, 0, nil)

	// Create a call with literal arguments - can be cached
	callWithLiterals := NewCall("map-get", []any{
		NewKeyword("colors"),
		NewKeyword("primary"),
	}, 0, nil)

	ruleset := NewRuleset(nil, []any{callWithVar, callWithLiterals}, false, nil)

	collector := NewPluginCallCollector([]string{"map-get"})
	calls := collector.Collect(ruleset)

	// Should only have the literal call (variable call can't be serialized)
	if len(calls) != 1 {
		t.Errorf("Expected 1 cacheable call, got %d", len(calls))
	}

	if len(calls) > 0 && calls[0].Args[0].(*Keyword).Value != "colors" {
		t.Error("Expected the literal call to be collected")
	}
}

func TestPluginCallCollector_CacheKeyFormat(t *testing.T) {
	call := NewCall("map-get", []any{
		NewKeyword("colors"),
		NewKeyword("primary"),
	}, 0, nil)

	ruleset := NewRuleset(nil, []any{call}, false, nil)

	collector := NewPluginCallCollector([]string{"map-get"})
	calls := collector.Collect(ruleset)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	// Check cache key format: "funcName:arg1|arg2"
	expectedKey := "map-get:colors|primary"
	if calls[0].CacheKey != expectedKey {
		t.Errorf("Expected cache key %q, got %q", expectedKey, calls[0].CacheKey)
	}
}

func TestPluginCallCollector_EmptyPluginList(t *testing.T) {
	call := NewCall("map-get", []any{NewKeyword("colors")}, 0, nil)
	ruleset := NewRuleset(nil, []any{call}, false, nil)

	// No plugin functions registered
	collector := NewPluginCallCollector([]string{})
	calls := collector.Collect(ruleset)

	if len(calls) != 0 {
		t.Errorf("Expected 0 calls with empty plugin list, got %d", len(calls))
	}
}

func TestPluginCallCollector_NilRoot(t *testing.T) {
	collector := NewPluginCallCollector([]string{"map-get"})
	calls := collector.Collect(nil)

	if len(calls) != 0 {
		t.Errorf("Expected 0 calls for nil root, got %d", len(calls))
	}
}

func TestPluginCallCollector_DimensionArgs(t *testing.T) {
	call := NewCall("some-plugin", []any{
		NewDimensionFrom(100.0, &Unit{Numerator: []string{"px"}}),
		NewDimensionFrom(50.0, &Unit{Numerator: []string{"%"}}),
	}, 0, nil)

	ruleset := NewRuleset(nil, []any{call}, false, nil)

	collector := NewPluginCallCollector([]string{"some-plugin"})
	calls := collector.Collect(ruleset)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	// Check that dimension args are properly serialized
	expectedKey := "some-plugin:100px|50%"
	if calls[0].CacheKey != expectedKey {
		t.Errorf("Expected cache key %q, got %q", expectedKey, calls[0].CacheKey)
	}
}

func TestPluginCallCollector_ColorArgs(t *testing.T) {
	call := NewCall("color-plugin", []any{
		NewColor([]float64{255, 128, 64}, 0.5, ""),
	}, 0, nil)

	ruleset := NewRuleset(nil, []any{call}, false, nil)

	collector := NewPluginCallCollector([]string{"color-plugin"})
	calls := collector.Collect(ruleset)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	// Check that color args are properly serialized
	expectedKey := "color-plugin:rgba(255,128,64,0.5)"
	if calls[0].CacheKey != expectedKey {
		t.Errorf("Expected cache key %q, got %q", expectedKey, calls[0].CacheKey)
	}
}

func TestPluginCallCollector_QuotedArgs(t *testing.T) {
	call := NewCall("string-plugin", []any{
		NewQuoted("\"", "hello world", false, 0, nil),
		NewQuoted("'", "test", false, 0, nil),
	}, 0, nil)

	ruleset := NewRuleset(nil, []any{call}, false, nil)

	collector := NewPluginCallCollector([]string{"string-plugin"})
	calls := collector.Collect(ruleset)

	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	// Check that quoted args are properly serialized
	expectedKey := `string-plugin:"hello world"|'test'`
	if calls[0].CacheKey != expectedKey {
		t.Errorf("Expected cache key %q, got %q", expectedKey, calls[0].CacheKey)
	}
}

func TestSerializeArgsForBatch(t *testing.T) {
	args := []any{
		NewDimensionFrom(100.0, &Unit{Numerator: []string{"px"}}),
		NewColor([]float64{255, 0, 0}, 1.0, "red"),
		NewKeyword("solid"),
		NewQuoted("\"", "hello", false, 0, nil),
	}

	serialized, err := serializeArgsForBatch(args, nil)
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	if len(serialized) != 4 {
		t.Errorf("Expected 4 serialized args, got %d", len(serialized))
	}

	// Check dimension serialization
	dim, ok := serialized[0].(map[string]any)
	if !ok {
		t.Errorf("Expected dimension to be map, got %T", serialized[0])
	} else {
		if dim["_type"] != "Dimension" {
			t.Errorf("Expected _type Dimension, got %v", dim["_type"])
		}
		if dim["value"] != 100.0 {
			t.Errorf("Expected value 100, got %v", dim["value"])
		}
		if dim["unit"] != "px" {
			t.Errorf("Expected unit px, got %v", dim["unit"])
		}
	}

	// Check color serialization
	color, ok := serialized[1].(map[string]any)
	if !ok {
		t.Errorf("Expected color to be map, got %T", serialized[1])
	} else {
		if color["_type"] != "Color" {
			t.Errorf("Expected _type Color, got %v", color["_type"])
		}
		if color["alpha"] != 1.0 {
			t.Errorf("Expected alpha 1.0, got %v", color["alpha"])
		}
	}

	// Check keyword serialization
	keyword, ok := serialized[2].(map[string]any)
	if !ok {
		t.Errorf("Expected keyword to be map, got %T", serialized[2])
	} else {
		if keyword["_type"] != "Keyword" {
			t.Errorf("Expected _type Keyword, got %v", keyword["_type"])
		}
		if keyword["value"] != "solid" {
			t.Errorf("Expected value solid, got %v", keyword["value"])
		}
	}

	// Check quoted serialization
	quoted, ok := serialized[3].(map[string]any)
	if !ok {
		t.Errorf("Expected quoted to be map, got %T", serialized[3])
	} else {
		if quoted["_type"] != "Quoted" {
			t.Errorf("Expected _type Quoted, got %v", quoted["_type"])
		}
		if quoted["value"] != "hello" {
			t.Errorf("Expected value hello, got %v", quoted["value"])
		}
	}
}

func TestWarmPluginCache_NilBridge(t *testing.T) {
	ruleset := NewRuleset(nil, []any{}, false, nil)

	count, err := WarmPluginCache(ruleset, nil, nil)
	if err != nil {
		t.Errorf("Expected no error for nil bridge, got: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 warmed entries for nil bridge, got %d", count)
	}
}

func TestWarmPluginCacheFromLazyBridge_Uninitialized(t *testing.T) {
	ruleset := NewRuleset(nil, []any{}, false, nil)

	// Create an uninitialized lazy bridge
	lazyBridge := NewLazyNodeJSPluginBridge()

	count, err := WarmPluginCacheFromLazyBridge(ruleset, lazyBridge, nil)
	if err != nil {
		t.Errorf("Expected no error for uninitialized bridge, got: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 warmed entries for uninitialized bridge, got %d", count)
	}
}
