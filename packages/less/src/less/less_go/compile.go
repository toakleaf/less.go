package less_go

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// CompileResult represents the result of compiling LESS to CSS.
// It contains the generated CSS along with optional source map and import information.
type CompileResult struct {
	// CSS contains the compiled CSS output.
	CSS string `json:"css"`

	// Map contains the source map in JSON format, if source maps were enabled.
	// Empty string if source maps were not generated.
	Map string `json:"map,omitempty"`

	// Imports lists all files that were imported during compilation.
	// This can be used for dependency tracking and watch mode implementations.
	Imports []string `json:"imports,omitempty"`
}

// CompileOptions configures the behavior of the LESS compiler.
// All fields are optional; a nil options pointer or zero-value struct uses sensible defaults.
//
// Example:
//
//	options := &CompileOptions{
//	    Filename:    "styles.less",
//	    Compress:    true,
//	    Math:        Math.ParensDivision,
//	    StrictUnits: true,
//	}
type CompileOptions struct {
	// Paths specifies additional directories to search when resolving @import statements.
	// These paths are searched in order after the directory of the importing file.
	Paths []string

	// Filename is the name of the input file being compiled.
	// Used for error messages, source maps, and resolving relative @import paths.
	// When compiling from a string, this provides context for the compilation.
	Filename string

	// Compress enables CSS minification in the output.
	// When true, whitespace is minimized and comments are removed.
	// Default: false (output is formatted for readability).
	Compress bool

	// StrictUnits enforces strict unit checking for mathematical operations.
	// When true, operations like "1px + 1em" will produce an error.
	// When false, incompatible units are allowed and the first unit is used.
	// Default: false.
	StrictUnits bool

	// Math controls how mathematical expressions are evaluated.
	// Possible values:
	//   - Math.Always: Evaluate all math expressions (Less.js v1.x behavior)
	//   - Math.ParensDivision: Only evaluate division in parentheses (default, Less.js v3.x+)
	//   - Math.Parens: Only evaluate math expressions in parentheses
	// Default: Math.ParensDivision (0 value maps to MathAlways, set explicitly for clarity).
	Math MathType

	// RewriteUrls controls how URLs in the output are rewritten relative to the entry file.
	// Possible values:
	//   - RewriteUrls.Off: No URL rewriting (default)
	//   - RewriteUrls.Local: Rewrite relative URLs only
	//   - RewriteUrls.All: Rewrite all URLs
	// Default: RewriteUrls.Off.
	RewriteUrls RewriteUrlsType

	// Rootpath specifies a path to prepend to all URLs in the output.
	// Useful when the CSS will be served from a different location than the source files.
	Rootpath string

	// UrlArgs specifies a query string to append to all URLs in the output.
	// Example: "v=1.0" would transform url("image.png") to url("image.png?v=1.0").
	UrlArgs string

	// EnableJavaScriptPlugins enables support for JavaScript plugins via Node.js.
	// When true, the compiler starts a Node.js runtime to handle @plugin directives
	// and JavaScript expressions in LESS files. The runtime is lazily initialized
	// only when needed and automatically cleaned up after compilation.
	// Default: false.
	EnableJavaScriptPlugins bool

	// JavascriptEnabled enables inline JavaScript evaluation in LESS files.
	// When true, backtick-delimited expressions like `1 + 1` are evaluated.
	// Note: This requires EnableJavaScriptPlugins to be true to provide the JS runtime.
	// Default: false.
	JavascriptEnabled bool

	// GlobalVars specifies variables to inject before compilation.
	// These variables are available throughout the LESS source and can be
	// overridden by variables defined in the source files.
	// Keys should not include the @ prefix.
	// Example: map[string]any{"baseColor": "#ff0000", "spacing": "10px"}
	GlobalVars map[string]any

	// ModifyVars specifies variables to inject after compilation.
	// These variables override any variables defined in the source files,
	// making them useful for theming and customization.
	// Keys should not include the @ prefix.
	// Example: map[string]any{"primaryColor": "#0000ff"}
	ModifyVars map[string]any
}

// Compile compiles LESS source code to CSS.
// This is the main entry point for the less.go compiler.
//
// Parameters:
//   - input: The LESS source code to compile as a string.
//   - options: Configuration options for the compilation. Pass nil to use defaults.
//
// Returns:
//   - *CompileResult: Contains the compiled CSS, optional source map, and list of imported files.
//   - error: Non-nil if compilation fails due to syntax errors, undefined variables, etc.
//
// When options.EnableJavaScriptPlugins is true, the function will:
//   - Create a lazy Node.js runtime that only starts if @plugin directives are encountered
//   - Properly shut down the Node.js process when compilation completes
//
// Example:
//
//	result, err := less_go.Compile(lessSource, &less_go.CompileOptions{
//	    Filename: "styles.less",
//	    Compress: true,
//	    EnableJavaScriptPlugins: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.CSS)
func Compile(input string, options *CompileOptions) (*CompileResult, error) {
	if options == nil {
		options = &CompileOptions{}
	}

	// Convert CompileOptions to map for internal functions
	optionsMap := convertCompileOptionsToMap(options)

	// Create context - use plugin-enabled context if JavaScript plugins are enabled
	var lessContext *LessContext
	var cleanup func() error

	if options.EnableJavaScriptPlugins {
		lessContext, cleanup = NewLessContextWithPlugins(optionsMap)
		defer func() {
			if cleanup != nil {
				if err := cleanup(); err != nil {
					// Log but don't fail - compilation already completed
					fmt.Fprintf(os.Stderr, "[less.go] Warning: failed to close plugin bridge: %v\n", err)
				}
			}
		}()
	} else {
		lessContext = NewLessContext(optionsMap)
	}

	// Set up functions
	lessContext.Functions = createFunctions(nil).(*DefaultFunctions)

	// Perform compilation with error handling
	result, err := compileWithContext(lessContext, input, optionsMap)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CompileFile compiles a LESS file to CSS.
// This is a convenience function that reads the file and compiles it with
// the Filename option automatically set for proper @import resolution.
//
// Parameters:
//   - filename: Path to the LESS file to compile. Can be absolute or relative.
//   - options: Configuration options for the compilation. Pass nil to use defaults.
//     The Filename field will be set to the provided filename if not already set.
//
// Returns:
//   - *CompileResult: Contains the compiled CSS, optional source map, and list of imported files.
//   - error: Non-nil if the file cannot be read or compilation fails.
//
// Example:
//
//	result, err := less_go.CompileFile("styles.less", &less_go.CompileOptions{
//	    Compress: true,
//	    EnableJavaScriptPlugins: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("styles.css", []byte(result.CSS), 0644)
func CompileFile(filename string, options *CompileOptions) (*CompileResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	if options == nil {
		options = &CompileOptions{}
	}
	options.Filename = filename

	return Compile(string(content), options)
}

// convertCompileOptionsToMap converts CompileOptions to a map for internal use
func convertCompileOptionsToMap(options *CompileOptions) map[string]any {
	result := make(map[string]any)

	if len(options.Paths) > 0 {
		result["paths"] = options.Paths
	}
	if options.Filename != "" {
		result["filename"] = options.Filename
	}
	if options.Compress {
		result["compress"] = true
	}
	if options.StrictUnits {
		result["strictUnits"] = true
	}
	if options.Math != 0 {
		result["math"] = options.Math
	}
	if options.RewriteUrls != 0 {
		result["rewriteUrls"] = options.RewriteUrls
	}
	if options.Rootpath != "" {
		result["rootpath"] = options.Rootpath
	}
	if options.UrlArgs != "" {
		result["urlArgs"] = options.UrlArgs
	}
	if options.GlobalVars != nil {
		result["globalVars"] = options.GlobalVars
	}
	if options.ModifyVars != nil {
		result["modifyVars"] = options.ModifyVars
	}
	// When EnableJavaScriptPlugins is true, also enable JavascriptEnabled
	// This ensures inline JavaScript expressions can be evaluated
	if options.JavascriptEnabled || options.EnableJavaScriptPlugins {
		result["javascriptEnabled"] = true
	}

	return result
}

// compileWithContext performs the actual compilation using the given context
func compileWithContext(lessContext *LessContext, input string, options map[string]any) (*CompileResult, error) {
	// Create environment and parseTree
	env := createEnvironment(nil, nil)

	// Create plugin manager early so it can be passed to import manager
	pluginManager := NewPluginManager(lessContext)

	// Create parse function
	parseFunc := CreateParse(env, nil, func(environment any, context *Parse, rootFileInfo map[string]any) *ImportManager {
		factory := NewImportManager(&SimpleImportManagerEnvironment{})

		fileInfo := &FileInfo{
			Filename: "input",
		}
		if rootFileInfo != nil {
			if fn, ok := rootFileInfo["filename"].(string); ok {
				fileInfo.Filename = fn
			}
		}

		contextMap := map[string]any{
			"parserFactory": func(parserContext map[string]any, parserImports map[string]any, parserFileInfo map[string]any, currentIndex int) ParserInterface {
				return NewParser(parserContext, parserImports, parserFileInfo, currentIndex)
			},
			"pluginManager": pluginManager, // Pass plugin manager to import manager
		}

		if context != nil && context.Paths != nil && len(context.Paths) > 0 {
			contextMap["paths"] = context.Paths
		} else {
			contextMap["paths"] = []string{}
		}

		if context != nil {
			contextMap["rewriteUrls"] = context.RewriteUrls
			if context.Rootpath != "" {
				contextMap["rootpath"] = context.Rootpath
			}
			contextMap["syncImport"] = context.SyncImport
			contextMap["strictImports"] = context.StrictImports
			contextMap["insecure"] = context.Insecure
		}

		return factory(environment, contextMap, fileInfo)
	})

	// Create channels for result
	resultChan := make(chan *CompileResult, 1)
	errorChan := make(chan error, 1)

	// Merge context options with provided options
	mergedOptions := make(map[string]any)
	if lessContext.Options != nil {
		for k, v := range lessContext.Options {
			mergedOptions[k] = v
		}
	}
	for k, v := range options {
		mergedOptions[k] = v
	}

	// Pass the plugin manager to the parse function
	mergedOptions["pluginManager"] = pluginManager

	// Pass the plugin bridge through options for use in TransformTree/Eval
	if lessContext.PluginBridge != nil {
		mergedOptions["pluginBridge"] = lessContext.PluginBridge
	}

	// Call parse with callback
	parseFunc(input, mergedOptions, func(err error, root any, imports *ImportManager, opts map[string]any) {
		if err != nil {
			errorChan <- err
			return
		}

		// Compile with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stackTrace := string(buf[:n])

					var errMsg string
					if e, ok := r.(error); ok {
						errMsg = e.Error()
					} else {
						errMsg = fmt.Sprintf("%v", r)
					}

					if strings.Contains(errMsg, "index out of range") || strings.Contains(errMsg, "nil pointer") {
						fmt.Fprintf(os.Stderr, "\n=== PANIC in compile ===\nError: %s\nStack:\n%s\n===\n", errMsg, stackTrace)
					}

					errorChan <- fmt.Errorf("compilation failed: %s", errMsg)
				}
			}()

			// Create ParseTree and call ToCSS
			parseTreeFactory := DefaultParseTreeFactory(nil)
			parseTreeInstance := parseTreeFactory.NewParseTree(root, imports)

			// Get functions
			functionsObj := createFunctions(env)

			// Convert options
			toCSSOptions := &ToCSSOptions{
				Compress:       false,
				StrictUnits:    false,
				NumPrecision:   8,
				Functions:      functionsObj,
				ProcessImports: true,
				ImportManager:  imports,
				Math:           Math.ParensDivision,
			}
			if opts != nil {
				if compress, ok := opts["compress"].(bool); ok {
					toCSSOptions.Compress = compress
				}
				if strictUnits, ok := opts["strictUnits"].(bool); ok {
					toCSSOptions.StrictUnits = strictUnits
				}
				if rewriteUrls, ok := opts["rewriteUrls"]; ok {
					toCSSOptions.RewriteUrls = rewriteUrls
				}
				if rootpath, ok := opts["rootpath"].(string); ok {
					toCSSOptions.Rootpath = rootpath
				}
				if math, ok := opts["math"].(MathType); ok {
					toCSSOptions.Math = math
				} else if mathInt, ok := opts["math"].(int); ok {
					toCSSOptions.Math = MathType(mathInt)
				}
				if paths, ok := opts["paths"].([]string); ok {
					toCSSOptions.Paths = paths
				}
				if urlArgs, ok := opts["urlArgs"].(string); ok {
					toCSSOptions.UrlArgs = urlArgs
				}
				if processImports, ok := opts["processImports"].(bool); ok {
					toCSSOptions.ProcessImports = processImports
				}
				// Pass the plugin bridge and plugin manager if present
				if pluginBridge := opts["pluginBridge"]; pluginBridge != nil {
					toCSSOptions.PluginBridge = pluginBridge
					toCSSOptions.PluginManager = opts["pluginManager"]
				}
				// Pass javascriptEnabled option for inline JavaScript evaluation
				if javascriptEnabled, ok := opts["javascriptEnabled"].(bool); ok {
					toCSSOptions.JavascriptEnabled = javascriptEnabled
				}
			}

			cssResult, err := parseTreeInstance.ToCSS(toCSSOptions)
			if err != nil {
				errorChan <- err
				return
			}

			resultChan <- &CompileResult{
				CSS:     cssResult.CSS,
				Map:     cssResult.Map,
				Imports: cssResult.Imports,
			}
		}()
	})

	// Wait for result
	select {
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}
