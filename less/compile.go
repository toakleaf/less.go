package less_go

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type CompileResult struct {
	CSS     string   `json:"css"`
	Map     string   `json:"map,omitempty"`
	Imports []string `json:"imports,omitempty"`
}

// PluginSpec specifies a plugin to load before compilation
type PluginSpec struct {
	// Name is the plugin name or path
	// Can be: relative path, absolute path, or npm module name
	// NPM modules are resolved with "less-plugin-" prefix first
	Name string

	// Options is an optional string of options to pass to the plugin
	// Passed to the plugin's setOptions() method
	Options string
}

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

	// JavascriptEnabled enables inline JavaScript evaluation in LESS files
	// When true, `expression` syntax can be used for JavaScript expressions
	// Note: This also requires EnableJavaScriptPlugins to be true for the runtime
	JavascriptEnabled bool

	// Plugins specifies plugins to load before compilation
	// When plugins are specified, EnableJavaScriptPlugins is automatically enabled
	Plugins []PluginSpec

	// GlobalVars are variables to inject before compilation
	GlobalVars map[string]any

	// ModifyVars are variables to inject after compilation (override existing variables)
	ModifyVars map[string]any

	// SourceMap enables source map generation
	SourceMap bool

	// SourceMapOptions contains detailed source map configuration
	SourceMapOptions *SourceMapOptions
}

// SourceMapOptions contains source map generation settings
type SourceMapOptions struct {
	// SourceMapFilename is the name of the source map file
	SourceMapFilename string

	// SourceMapURL overrides the source map URL in the CSS output
	SourceMapURL string

	// SourceMapBasepath is the base path to remove from source paths
	SourceMapBasepath string

	// SourceMapRootpath is the root path to prepend to source paths
	SourceMapRootpath string

	// SourceMapOutputFilename is the output filename for path calculation
	SourceMapOutputFilename string

	// OutputSourceFiles embeds the source content in the source map
	OutputSourceFiles bool

	// SourceMapFileInline embeds the source map as a data URI in the CSS
	SourceMapFileInline bool

	// DisableSourcemapAnnotation disables adding the sourceMappingURL comment
	DisableSourcemapAnnotation bool
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

	// Auto-enable JavaScript plugins when plugins are specified
	if len(options.Plugins) > 0 {
		options.EnableJavaScriptPlugins = true
	}

	optionsMap := convertCompileOptionsToMap(options)

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

		// Preload plugins specified in options
		if len(options.Plugins) > 0 && lessContext.PluginBridge != nil {
			// Determine base directory for plugin resolution
			baseDir := "."
			if options.Filename != "" {
				baseDir = filepath.Dir(options.Filename)
			}
			// If we have include paths, use the first one as a fallback
			if len(options.Paths) > 0 && baseDir == "." {
				baseDir = options.Paths[0]
			}

			for _, plugin := range options.Plugins {
				context := make(map[string]any)
				if plugin.Options != "" {
					context["options"] = map[string]any{"_args": plugin.Options}
				}

				result := lessContext.PluginBridge.LoadPluginSync(plugin.Name, baseDir, context, nil, nil)
				if err, ok := result.(error); ok {
					return nil, fmt.Errorf("failed to load plugin %q: %w", plugin.Name, err)
				}
			}
		}
	} else {
		lessContext = NewLessContext(optionsMap)
	}

	lessContext.Functions = createFunctions(nil).(*DefaultFunctions)

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

	// Source map options
	if options.SourceMap || options.SourceMapOptions != nil {
		sourceMapOpts := make(map[string]any)
		sourceMapOpts["enabled"] = true

		if options.SourceMapOptions != nil {
			opts := options.SourceMapOptions
			if opts.SourceMapFilename != "" {
				sourceMapOpts["sourceMapFilename"] = opts.SourceMapFilename
			}
			if opts.SourceMapURL != "" {
				sourceMapOpts["sourceMapURL"] = opts.SourceMapURL
			}
			if opts.SourceMapBasepath != "" {
				sourceMapOpts["sourceMapBasepath"] = opts.SourceMapBasepath
			}
			if opts.SourceMapRootpath != "" {
				sourceMapOpts["sourceMapRootpath"] = opts.SourceMapRootpath
			}
			if opts.SourceMapOutputFilename != "" {
				sourceMapOpts["sourceMapOutputFilename"] = opts.SourceMapOutputFilename
			}
			sourceMapOpts["outputSourceFiles"] = opts.OutputSourceFiles
			sourceMapOpts["sourceMapFileInline"] = opts.SourceMapFileInline
			sourceMapOpts["disableSourcemapAnnotation"] = opts.DisableSourcemapAnnotation
		}

		result["sourceMap"] = sourceMapOpts
	}

	return result
}

func compileWithContext(lessContext *LessContext, input string, options map[string]any) (*CompileResult, error) {
	env := createEnvironment(nil, nil)
	pluginManager := NewPluginManager(lessContext)

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
			"pluginManager": pluginManager,
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

	resultChan := make(chan *CompileResult, 1)
	errorChan := make(chan error, 1)

	mergedOptions := make(map[string]any)
	if lessContext.Options != nil {
		for k, v := range lessContext.Options {
			mergedOptions[k] = v
		}
	}
	for k, v := range options {
		mergedOptions[k] = v
	}

	mergedOptions["pluginManager"] = pluginManager

	if lessContext.PluginBridge != nil {
		mergedOptions["pluginBridge"] = lessContext.PluginBridge
	}

	parseFunc(input, mergedOptions, func(err error, root any, imports *ImportManager, opts map[string]any) {
		if err != nil {
			errorChan <- err
			return
		}

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

			// Create source map builder factory if source maps are enabled
			var sourceMapBuilderFactory any
			if sourceMapOpts := opts["sourceMap"]; sourceMapOpts != nil {
				sourceMapBuilderFactory = func(options any) *SourceMapBuilder {
					builderOpts := SourceMapBuilderOptions{}
					if optsMap, ok := options.(map[string]any); ok {
						if v, ok := optsMap["sourceMapFilename"].(string); ok {
							builderOpts.SourceMapFilename = v
						}
						if v, ok := optsMap["sourceMapURL"].(string); ok {
							builderOpts.SourceMapURL = v
						}
						if v, ok := optsMap["sourceMapOutputFilename"].(string); ok {
							builderOpts.SourceMapOutputFilename = v
						}
						if v, ok := optsMap["sourceMapBasepath"].(string); ok {
							builderOpts.SourceMapBasepath = v
						}
						if v, ok := optsMap["sourceMapRootpath"].(string); ok {
							builderOpts.SourceMapRootpath = v
						}
						if v, ok := optsMap["outputSourceFiles"].(bool); ok {
							builderOpts.OutputSourceFiles = v
						}
						if v, ok := optsMap["sourceMapFileInline"].(bool); ok {
							builderOpts.SourceMapFileInline = v
						}
						if v, ok := optsMap["disableSourcemapAnnotation"].(bool); ok {
							builderOpts.DisableSourcemapAnnotation = v
						}
					}
					return NewSourceMapBuilder(builderOpts)
				}
			}

			parseTreeFactory := DefaultParseTreeFactory(sourceMapBuilderFactory)
			parseTreeInstance := parseTreeFactory.NewParseTree(root, imports)

			functionsObj := createFunctions(env)

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
				if pluginBridge := opts["pluginBridge"]; pluginBridge != nil {
					toCSSOptions.PluginBridge = pluginBridge
					toCSSOptions.PluginManager = opts["pluginManager"]
				}
				if javascriptEnabled, ok := opts["javascriptEnabled"].(bool); ok {
					toCSSOptions.JavascriptEnabled = javascriptEnabled
				}
				// Pass source map options
				if sourceMapOpts := opts["sourceMap"]; sourceMapOpts != nil {
					toCSSOptions.SourceMap = sourceMapOpts
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

	select {
	case err := <-errorChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}
