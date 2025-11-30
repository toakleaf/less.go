package less_go

import (
	"errors"
	"reflect"
	"testing"
)

type mockRenderContext struct {
	parseFunc func(string, map[string]any, func(error, any, any, map[string]any))
	options   map[string]any
}

func (mc *mockRenderContext) Parse(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
	if mc.parseFunc != nil {
		mc.parseFunc(input, options, callback)
	}
}

func (mc *mockRenderContext) GetOptions() map[string]any {
	return mc.options
}

type mockParseTree struct {
	toCSSFunc func(map[string]any) any
}

func (mpt *mockParseTree) ToCSS(options map[string]any) any {
	if mpt.toCSSFunc != nil {
		return mpt.toCSSFunc(options)
	}
	return map[string]any{"css": "body { color: red; }", "map": nil}
}

func createMockParseTreeConstructor(toCSSFunc func(map[string]any) any) func(any, any) any {
	return func(root any, imports any) any {
		return &mockParseTree{toCSSFunc: toCSSFunc}
	}
}

func TestRenderOptionsAsCallback(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)

	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			if input != ".class { color: blue; }" {
				t.Errorf("Expected input '.class { color: blue; }', got '%s'", input)
			}
			if sourceMap, ok := options["sourceMap"]; !ok || sourceMap != false {
				t.Errorf("Expected sourceMap: false in options")
			}
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}

	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackResult any
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
	}

	boundRender(".class { color: blue; }", callback)
	
	if callbackError != nil {
		t.Fatalf("Expected no error, got: %v", callbackError)
	}
	
	if callbackResult == nil {
		t.Fatal("Expected result, got nil")
	}

	if resultMap, ok := callbackResult.(map[string]any); ok {
		if css, ok := resultMap["css"]; !ok || css != "body { color: red; }" {
			t.Errorf("Expected CSS 'body { color: red; }', got '%v'", css)
		}
	}
}

func TestRenderOptionsAndCallback(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			if compress, ok := options["compress"]; !ok || compress != true {
				t.Errorf("Expected compress: true in merged options")
			}
			if sourceMap, ok := options["sourceMap"]; !ok || sourceMap != false {
				t.Errorf("Expected sourceMap: false from context options")
			}
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackResult any
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
	}
	
	options := map[string]any{"compress": true}

	boundRender(".class { color: blue; }", options, callback)
	
	if callbackError != nil {
		t.Fatalf("Expected no error, got: %v", callbackError)
	}
	
	if callbackResult == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestRenderParseErrors(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	parseError := errors.New("Parse error")
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(parseError, nil, nil, nil)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackResult any
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
	}
	
	boundRender(".class { color: blue; }", map[string]any{}, callback)
	
	if callbackError == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if callbackError != parseError {
		t.Errorf("Expected parse error, got different error: %v", callbackError)
	}
	
	if callbackResult != nil {
		t.Errorf("Expected nil result on error, got: %v", callbackResult)
	}
}

func TestRenderToCSSErrors(t *testing.T) {
	toCSSError := errors.New("toCSS error")
	mockEnvironment := map[string]any{}

	mockParseTreeConstructor := func(root any, imports any) any {
		return &mockParseTreeWithError{err: toCSSError}
	}
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackError error
	var callbackResult any
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
		t.Logf("Callback called with err=%v, output=%v", err, output)
	}
	
	boundRender(".class { color: blue; }", map[string]any{}, callback)
	
	if callbackError == nil {
		t.Fatalf("Expected error, got nil. Result was: %v", callbackResult)
	}
	
	if callbackError != toCSSError {
		t.Errorf("Expected toCSS error, got different error: %v", callbackError)
	}
}

type mockParseTreeWithError struct {
	err error
}

func (mpt *mockParseTreeWithError) ToCSS(options map[string]any) (any, error) {
	return nil, mpt.err
}

func TestRenderCreateParseTreeCorrectArguments(t *testing.T) {
	mockEnvironment := map[string]any{}
	var capturedRoot any
	var capturedImports any
	
	mockParseTreeConstructor := func(root any, imports any) any {
		capturedRoot = root
		capturedImports = imports
		return &mockParseTree{}
	}
	
	mockRoot := map[string]any{"type": "Root", "rules": []any{}}
	mockImports := []any{map[string]any{"path": "import.less"}}
	mockOptions := map[string]any{"sourceMap": true}
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, mockRoot, mockImports, mockOptions)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	callback := func(err error, output any) {}
	
	boundRender(".class { color: blue; }", map[string]any{}, callback)
	
	if !reflect.DeepEqual(capturedRoot, mockRoot) {
		t.Errorf("Expected root to be passed to ParseTree constructor")
	}
	
	if !reflect.DeepEqual(capturedImports, mockImports) {
		t.Errorf("Expected imports to be passed to ParseTree constructor")
	}
}

func TestRenderReturnPromise(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)

	result := boundRender(".class { color: blue; }", map[string]any{})
	
	if result == nil {
		t.Fatal("Expected promise, got nil")
	}
	
	promise, ok := result.(*RenderPromise)
	if !ok {
		t.Fatalf("Expected RenderPromise, got %T", result)
	}

	output, err := promise.Await()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if output == nil {
		t.Fatal("Expected output, got nil")
	}

	if resultMap, ok := output.(map[string]any); ok {
		if css, ok := resultMap["css"]; !ok || css != "body { color: red; }" {
			t.Errorf("Expected CSS 'body { color: red; }', got '%v'", css)
		}
	}
}

func TestRenderPromiseResolveWithResult(t *testing.T) {
	expectedResult := map[string]any{"css": "compiled css", "map": "source map"}
	mockEnvironment := map[string]any{}
	
	mockParseTreeConstructor := createMockParseTreeConstructor(func(options map[string]any) any {
		return expectedResult
	})
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	promise := boundRender(".class { color: blue; }", map[string]any{}).(*RenderPromise)
	
	result, err := promise.Await()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("Expected specific result, got different result: %v", result)
	}
}

func TestRenderPromiseRejectWithParseError(t *testing.T) {
	parseError := errors.New("Parse error")
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(parseError, nil, nil, nil)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	promise := boundRender(".class { color: blue; }", map[string]any{}).(*RenderPromise)
	
	result, err := promise.Await()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if err != parseError {
		t.Errorf("Expected parse error, got: %v", err)
	}
	
	if result != nil {
		t.Errorf("Expected nil result on error, got: %v", result)
	}
}

func TestRenderPromiseRejectWithToCSSError(t *testing.T) {
	toCSSError := errors.New("toCSS error")
	mockEnvironment := map[string]any{}
	
	mockParseTreeConstructor := func(root any, imports any) any {
		return &mockParseTreeWithError{err: toCSSError}
	}
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	promise := boundRender(".class { color: blue; }", map[string]any{}).(*RenderPromise)
	
	result, err := promise.Await()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	if err != toCSSError {
		t.Errorf("Expected toCSS error, got: %v", err)
	}
	
	if result != nil {
		t.Errorf("Expected nil result on error, got: %v", result)
	}
}

func TestRenderHandleOptionsWithoutCallback(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			if compress, ok := options["compress"]; !ok || compress != true {
				t.Errorf("Expected compress: true in options")
			}
			if sourceMap, ok := options["sourceMap"]; !ok || sourceMap != true {
				t.Errorf("Expected sourceMap: true in options")
			}
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	options := map[string]any{"compress": true, "sourceMap": true}
	promise := boundRender(".class { color: blue; }", options).(*RenderPromise)
	
	_, err := promise.Await()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRenderMaintainContextThroughPromise(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	customContext := &mockRenderContext{
		options: map[string]any{"custom": "option"},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			if custom, ok := options["custom"]; !ok || custom != "option" {
				t.Errorf("Expected custom option to be preserved")
			}
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, customContext, mockEnvironment, mockParseTreeConstructor)
	
	promise := boundRender(".class { color: blue; }").(*RenderPromise)
	
	_, err := promise.Await()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRenderHandleNullOptions(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackResult any
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
	}

	boundRender(".class { color: blue; }", nil, callback)
	
	if callbackError != nil {
		t.Fatalf("Expected no error, got: %v", callbackError)
	}
	
	if callbackResult == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestRenderHandleUndefinedOptions(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackResult any
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
	}

	var undefinedOptions interface{} = nil
	boundRender(".class { color: blue; }", undefinedOptions, callback)
	
	if callbackError != nil {
		t.Fatalf("Expected no error, got: %v", callbackError)
	}
	
	if callbackResult == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestRenderHandleEmptyInput(t *testing.T) {
	mockEnvironment := map[string]any{}
	mockParseTreeConstructor := createMockParseTreeConstructor(nil)
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			if input != "" {
				t.Errorf("Expected empty input, got: %s", input)
			}
			callback(nil, map[string]any{"type": "Root"}, []any{}, options)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackResult any
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
		callbackResult = output
	}
	
	boundRender("", callback)
	
	if callbackError != nil {
		t.Fatalf("Expected no error, got: %v", callbackError)
	}
	
	if callbackResult == nil {
		t.Fatal("Expected result, got nil")
	}
}

func TestRenderPassThroughOptionsFromParseCallback(t *testing.T) {
	mockEnvironment := map[string]any{}
	modifiedOptions := map[string]any{"sourceMap": true, "modified": true}
	
	var toCSSReceivedOptions map[string]any
	mockParseTreeConstructor := createMockParseTreeConstructor(func(options map[string]any) any {
		toCSSReceivedOptions = options
		return map[string]any{"css": "body { color: red; }", "map": nil}
	})
	
	mockContext := &mockRenderContext{
		options: map[string]any{"sourceMap": false},
		parseFunc: func(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
			callback(nil, map[string]any{"type": "Root"}, []any{}, modifiedOptions)
		},
	}
	
	renderFunc := CreateRender(mockEnvironment, mockParseTreeConstructor)
	boundRender := Bind(renderFunc, mockContext, mockEnvironment, mockParseTreeConstructor)
	
	var callbackError error
	callback := func(err error, output any) {
		callbackError = err
	}
	
	boundRender(".class { color: blue; }", map[string]any{}, callback)
	
	if callbackError != nil {
		t.Fatalf("Expected no error, got: %v", callbackError)
	}

	if toCSSReceivedOptions == nil {
		t.Fatal("Expected toCSS to receive options")
	}
	
	if modified, ok := toCSSReceivedOptions["modified"]; !ok || modified != true {
		t.Errorf("Expected modified options to be passed to toCSS")
	}
}