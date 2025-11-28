package less_go

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// CompileResult represents the result of compiling LESS to CSS
type CompileResult struct {
	CSS     string   `json:"css"`
	Map     string   `json:"map,omitempty"`
	Imports []string `json:"imports,omitempty"`
}

// CompileOptions represents options for the Compile function
type CompileOptions struct {
	// Paths are additional include paths for @import resolution
	Paths []string

	// Filename is the name of the file being compiled (used for error messages and source maps)
	Filename string

	// Compress enables CSS minification
	Compress bool

	// StrictUnits controls unit checking for math operations
	StrictUnits bool

	// Math controls how math operations are evaluated
	// Use Math.Always, Math.ParensDivision, or Math.Parens
	Math MathType

	// RewriteUrls controls URL rewriting behavior
	RewriteUrls RewriteUrlsType

	// Rootpath is the base path for URL rewriting
	Rootpath string

	// UrlArgs is a query string to append to URLs
	UrlArgs string

	// EnableJavaScriptPlugins enables support for JavaScript plugins via Node.js
	// When true, the compiler will start a Node.js runtime to handle @plugin directives
	EnableJavaScriptPlugins bool

	// GlobalVars are variables to inject before compilation
	GlobalVars map[string]any

	// ModifyVars are variables to inject after compilation (override existing variables)
	ModifyVars map[string]any
}

// Compile compiles LESS source code to CSS.
// This is the main entry point for the less.go compiler.
//
// When EnableJavaScriptPlugins is true, the function will:
// - Create a lazy Node.js runtime that only starts if @plugin directives are encountered
// - Properly shut down the Node.js process when compilation completes
//
// Example usage:
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
// This reads the file and compiles it with appropriate options set for file-based compilation.
//
// Example usage:
//
//	result, err := less_go.CompileFile("styles.less", &less_go.CompileOptions{
//	    Compress: true,
//	    EnableJavaScriptPlugins: true,
//	})
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
