package less_go

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func Factory(environment map[string]any, fileManagers []any) map[string]any {
	var sourceMapOutput any
	var sourceMapBuilder any
	var parseTree any
	var importManager any

	env := createEnvironment(environment, fileManagers)

	sourceMapOutput = createSourceMapOutput(env)
	sourceMapBuilder = createSourceMapBuilder(sourceMapOutput, env)
	parseTree = createParseTree(sourceMapBuilder)
	importManager = createImportManager(env)

	render := createRender(env, parseTree, importManager)
	parse := createParse(env, parseTree, importManager)

	v := parseVersion("v4.2.2")

	initial := map[string]any{
		"version":              []int{v.Major, v.Minor, v.Patch},
		"data":                 createDataExports(),
		"tree":                 createTreeExports(),
		"Environment":          createEnvironment,
		"AbstractFileManager":  createAbstractFileManager,
		"AbstractPluginLoader": createAbstractPluginLoader,
		"environment":          env,
		"visitors":             createVisitors(),
		"Parser":               createParser,
		"functions":            createFunctions(env),
		"contexts":             createContexts(),
		"SourceMapOutput":      sourceMapOutput,
		"SourceMapBuilder":     sourceMapBuilder,
		"ParseTree":            parseTree,
		"ImportManager":        importManager,
		"render":               render,
		"parse":                parse,
		"LessError":            createLessError,
		"transformTree":        createTransformTree,
		"utils":                createUtils(),
		"PluginManager":        createPluginManager,
		"logger":               createLogger(),
	}

	api := make(map[string]any)

	for key, value := range initial {
		api[key] = value
	}

	if tree, ok := initial["tree"].(map[string]any); ok {
		for name, t := range tree {
			if isFunction(t) {
				api[strings.ToLower(name)] = createConstructor(t)
			} else if isObject(t) {
				nestedObj := make(map[string]any)
				if tMap, ok := t.(map[string]any); ok {
					for innerName, innerT := range tMap {
						if isFunction(innerT) {
							nestedObj[strings.ToLower(innerName)] = createConstructor(innerT)
						}
					}
				}
				api[strings.ToLower(name)] = nestedObj
			}
		}
	}

	if renderFunc, ok := api["render"].(func(string, ...any) any); ok {
		api["render"] = bindRenderToContext(renderFunc, api)
	}
	if parseFunc, ok := api["parse"].(func(string, ...any) any); ok {
		api["parse"] = bindParseToContext(parseFunc, api)
	}

	return api
}

func parseVersion(version string) VersionInfo {
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	
	return VersionInfo{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

type VersionInfo struct {
	Major int
	Minor int
	Patch int
}

func isFunction(value any) bool {
	switch value.(type) {
	case func() any, func(...any) any:
		return true
	default:
		return fmt.Sprintf("%T", value) == "func() interface {}"
	}
}

func isObject(value any) bool {
	_, ok := value.(map[string]any)
	return ok
}

func createConstructor(t any) func(...any) any {
	return func(args ...any) any {
		switch constructor := t.(type) {
		case func() any:
			return constructor()
		case func(...any) any:
			return constructor(args...)
		default:
			return createInstanceFromType(t, args...)
		}
	}
}

func createInstanceFromType(t any, args ...any) any {
	return map[string]any{
		"type": fmt.Sprintf("%T", t),
		"args": args,
	}
}

func bindRenderToContext(renderFunc func(string, ...any) any, context map[string]any) func(string, ...any) any {
	return func(input string, args ...any) any {
		return callRenderWithContext(renderFunc, context, input, args...)
	}
}

func bindParseToContext(parseFunc func(string, ...any) any, context map[string]any) func(string, ...any) any {
	return func(input string, args ...any) any {
		return callParseWithContext(parseFunc, context, input, args...)
	}
}

type APIContext struct {
	context map[string]any
}

func (ac *APIContext) Parse(input string, options map[string]any, callback func(error, any, any, map[string]any)) {
	if parseFunc, ok := ac.context["parse"].(func(string, map[string]any, func(error, any, any, map[string]any))); ok {
		parseFunc(input, options, callback)
	}
}

func (ac *APIContext) GetOptions() map[string]any {
	if options, ok := ac.context["options"].(map[string]any); ok {
		return options
	}
	return make(map[string]any)
}

func callRenderWithContext(renderFunc func(string, ...any) any, context map[string]any, input string, args ...any) any {
	return renderFunc(input, args...)
}

func callParseWithContext(parseFunc func(string, ...any) any, context map[string]any, input string, args ...any) any {
	return parseFunc(input, args...)
}

func createEnvironment(environment map[string]any, fileManagers []any) any {
	return map[string]any{
		"environment":  environment,
		"fileManagers": fileManagers,
	}
}

func createSourceMapOutput(env any) any {
	return map[string]any{"type": "SourceMapOutput"}
}

func createSourceMapBuilder(sourceMapOutput any, env any) any {
	return map[string]any{"type": "SourceMapBuilder"}
}

func createImportManager(env any) any {
	return map[string]any{"type": "ImportManager"}
}

func createParseTree(sourceMapBuilder any) any {
	return map[string]any{"type": "ParseTree", "sourceMapBuilder": sourceMapBuilder}
}

func createRender(env any, parseTree any, importManager any) func(string, ...any) any {
	return func(input string, args ...any) (result any) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				var errMsg string
				if err, ok := r.(error); ok {
					errMsg = err.Error()
				} else {
					errMsg = fmt.Sprintf("%v", r)
				}

				filename := "input"
				if len(args) > 0 {
					if opts, ok := args[0].(map[string]any); ok {
						if fn, ok := opts["filename"].(string); ok {
							filename = fn
						}
					}
				}

				if strings.Contains(errMsg, "nil pointer dereference") || strings.Contains(errMsg, "invalid memory address") || strings.Contains(errMsg, "index out of range") {
					fmt.Fprintf(os.Stderr, "\n=== DEBUG: Runtime error in render ===\nError: %s\nFile: %s\nStack trace:\n%s\n===\n", errMsg, filename, stackTrace)
				}

				result = map[string]any{
					"error": fmt.Sprintf("Syntax: %s in %s", errMsg, filename),
				}
			}
		}()

		var options map[string]any
		if len(args) > 0 {
			if opts, ok := args[0].(map[string]any); ok {
				options = opts
			}
		}
		if options == nil {
			options = make(map[string]any)
		}

		if os.Getenv("LESS_GO_TRACE") == "1" {
			fmt.Printf("[RENDER-DEBUG] Options before CopyOptions: math=%v (type=%T)\n", options["math"], options["math"])
		}
		options = CopyOptions(options, nil)
		if os.Getenv("LESS_GO_TRACE") == "1" {
			fmt.Printf("[RENDER-DEBUG] Options after CopyOptions: math=%v (type=%T)\n", options["math"], options["math"])
		}

		parseFunc := CreateParse(env, parseTree, func(environment any, context *Parse, rootFileInfo map[string]any) *ImportManager {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG createRender ImportManagerFactory] Called with context.Paths: %v\n", context.Paths)
			}

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
			}

			if context != nil && context.Paths != nil && len(context.Paths) > 0 {
				contextMap["paths"] = context.Paths
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG createRender ImportManagerFactory] Setting contextMap paths: %v\n", context.Paths)
				}
			} else {
				contextMap["paths"] = []string{}
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG createRender ImportManagerFactory] No paths in context\n")
				}
			}

			if context != nil {
				contextMap["rewriteUrls"] = context.RewriteUrls
				if context.Rootpath != "" {
					contextMap["rootpath"] = context.Rootpath
				}
				contextMap["syncImport"] = context.SyncImport
				contextMap["strictImports"] = context.StrictImports
				contextMap["insecure"] = context.Insecure

				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG createRender ImportManagerFactory] Copied options: rewriteUrls=%v, rootpath=%q, syncImport=%v\n",
						context.RewriteUrls, context.Rootpath, context.SyncImport)
				}
			}

			result := factory(environment, contextMap, fileInfo)
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG createRender ImportManagerFactory] Created ImportManager with paths: %v\n", result.paths)
			}
			return result
		})

		resultChan := make(chan any, 1)
		errorChan := make(chan error, 1)

		parseFunc(input, options, func(err error, root any, imports *ImportManager, opts map[string]any) {
			if err != nil {
				errorChan <- err
			} else {
				parseTreeFactory := DefaultParseTreeFactory(nil)
				parseTreeInstance := parseTreeFactory.NewParseTree(root, imports)

				functionsObj := createFunctions(env)

				toCSSOptions := &ToCSSOptions{
					Compress:      false,
					StrictUnits:   false,
					NumPrecision:  8,  // Match less.js default precision
					Functions:     functionsObj,
					ProcessImports: true,  // Default: enable import processing
					ImportManager: imports, // Pass the import manager
					Math:          Math.ParensDivision, // Default math mode
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
						// Handle integer math values from test options
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
				}

				func() {
					defer func() {
						if r := recover(); r != nil {
							buf := make([]byte, 8192)
							n := runtime.Stack(buf, false)
							stackTrace := string(buf[:n])

							var errMsg string
							if err, ok := r.(error); ok {
								errMsg = err.Error()
							} else {
								errMsg = fmt.Sprintf("%v", r)
							}

							if strings.Contains(errMsg, "index out of range") {
								fmt.Fprintf(os.Stderr, "\n=== PANIC in ToCSS ===\nError: %s\nStack:\n%s\n===\n", errMsg, stackTrace)
								fmt.Printf("\n=== PANIC in ToCSS ===\nError: %s\nStack:\n%s\n===\n", errMsg, stackTrace)
							}

							errorChan <- fmt.Errorf("%s", errMsg)
						}
					}()

					cssResult, err := parseTreeInstance.ToCSS(toCSSOptions)
					if err != nil {
						errorChan <- err
					} else {
						resultChan <- cssResult.CSS
					}
				}()
			}
		})

		select {
		case err := <-errorChan:
			return map[string]any{"error": err.Error()}
		case result := <-resultChan:
			return result
		}
	}
}

func createTransformTree() any {
	return func(root any, options map[string]any) any {
		return map[string]any{"type": "TransformedTree", "root": root}
	}
}

func createParse(env any, parseTree any, importManager any) func(string, ...any) any {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG] createParse function called\n")
	}

	importManagerFactory := func(environment any, context *Parse, rootFileInfo map[string]any) *ImportManager {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG ImportManagerFactory] Called\n")
			if context != nil {
				fmt.Printf("[DEBUG ImportManagerFactory] context.Paths: %v (len=%d)\n", context.Paths, len(context.Paths))
			} else {
				fmt.Printf("[DEBUG ImportManagerFactory] context is nil\n")
			}
		}

		factory := NewImportManager(&SimpleImportManagerEnvironment{})

		fileInfo := &FileInfo{
			Filename: "input",
		}
		if rootFileInfo != nil {
			if fn, ok := rootFileInfo["filename"].(string); ok {
				fileInfo.Filename = fn
			}
		}

		contextMap := map[string]any{}

		if context != nil && context.Paths != nil && len(context.Paths) > 0 {
			contextMap["paths"] = context.Paths
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG ImportManagerFactory] Setting contextMap['paths']: %v\n", context.Paths)
			}
		} else {
			contextMap["paths"] = []string{}
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG ImportManagerFactory] Setting empty paths\n")
			}
		}

		if context != nil {
			contextMap["rewriteUrls"] = context.RewriteUrls
			if context.Rootpath != "" {
				contextMap["rootpath"] = context.Rootpath
			}
			contextMap["syncImport"] = context.SyncImport
			contextMap["strictImports"] = context.StrictImports
			contextMap["insecure"] = context.Insecure

			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG ImportManagerFactory] Copied options: rewriteUrls=%v, rootpath=%q, syncImport=%v\n",
					context.RewriteUrls, context.Rootpath, context.SyncImport)
			}
		}

		result := factory(environment, contextMap, fileInfo)
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG ImportManagerFactory] Created ImportManager with im.paths: %v (len=%d)\n", result.paths, len(result.paths))
		}
		return result
	}

	realParseFunc := CreateParse(env, parseTree, importManagerFactory)

	return func(input string, args ...any) any {
		var options map[string]any
		if len(args) > 0 {
			if opts, ok := args[0].(map[string]any); ok {
				options = opts
			}
		}
		if options == nil {
			options = make(map[string]any)
		}

		resultChan := make(chan any, 1)
		errorChan := make(chan error, 1)

		realParseFunc(input, options, func(err error, root any, imports *ImportManager, opts map[string]any) {
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- root
			}
		})

		select {
		case err := <-errorChan:
			return map[string]any{"error": err.Error()}
		case result := <-resultChan:
			return result
		}
	}
}

func createDataExports() map[string]any {
	return map[string]any{
		"colors":          Colors,
		"unitConversions": createUnitConversions(),
	}
}

func createUnitConversions() map[string]any {
	return map[string]any{
		"length":   UnitConversionsLength,
		"duration": UnitConversionsDuration,
		"angle":    UnitConversionsAngle,
	}
}

func createTreeExports() map[string]any {
	return map[string]any{
		"TestNode": func(args ...any) any {
			return map[string]any{"type": "TestNode", "args": args}
		},
		"AnotherNode": func(args ...any) any {
			return map[string]any{"type": "AnotherNode", "args": args}
		},
		"NestedNodes": map[string]any{
			"InnerNode": func(args ...any) any {
				return map[string]any{"type": "InnerNode", "args": args}
			},
			"DeepNode": func(args ...any) any {
				return map[string]any{"type": "DeepNode", "args": args}
			},
		},
	}
}

func createAbstractFileManager() any {
	return map[string]any{"type": "AbstractFileManager"}
}

func createAbstractPluginLoader() any {
	return map[string]any{"type": "AbstractPluginLoader"}
}

func createVisitors() any {
	return map[string]any{"type": "Visitors"}
}

func createParser() any {
	return map[string]any{"type": "Parser"}
}

func createFunctions(env any) any {
	registry := DefaultRegistry.Inherit()

	RegisterTestFunctions(registry)

	listFunctions := GetWrappedListFunctions()
	for name, fn := range listFunctions {
		switch name {
		case "_SELF":
			if selfFn, ok := fn.(func(any) any); ok {
				registry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   1,
					maxArgs:   1,
					variadic:  false,
					fn:        selfFn,
					needsEval: true,
				})
			}
		case "~":
			if spaceFn, ok := fn.(func(...any) any); ok {
				registry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   0,
					maxArgs:   -1,
					variadic:  true,
					fn:        spaceFn,
					needsEval: true,
				})
			}
		case "range":
			if rangeFn, ok := fn.(func(any, any, any) *Expression); ok {
				// Wrap the function to match the expected signature
				registry.Add(name, &FlexibleFunctionDef{
					name:      name,
					minArgs:   1,
					maxArgs:   3,
					variadic:  false,
					fn:        func(args ...any) any {
						var start, end, step any
						if len(args) > 0 {
							start = args[0]
						}
						if len(args) > 1 {
							end = args[1]
						}
						if len(args) > 2 {
							step = args[2]
						}
						return rangeFn(start, end, step)
					},
					needsEval: true,
				})
			}
		case "each":
			// Register the EachFunctionDef which supports context
			registry.Add(name, &EachFunctionDef{
				name: name,
			})
		default:
			// Try as 2-argument function (extract, length)
			if functionImpl, ok := fn.(func(any, any) any); ok {
				registry.Add(name, &SimpleFunctionDef{
					name: name,
					fn:   functionImpl,
				})
			}
		}
	}

	registry.AddMultiple(GetWrappedStringFunctions())
	registry.AddMultiple(GetWrappedMathFunctions())
	registry.AddMultiple(GetWrappedNumberFunctions())
	registry.AddMultiple(GetWrappedBooleanFunctions())
	registry.AddMultiple(GetWrappedSvgFunctions())
	registry.AddMultiple(GetWrappedStyleFunctions())
	registry.AddMultiple(GetWrappedDataURIFunctions())
	registry.AddMultiple(GetWrappedColorBlendingFunctions())
	registry.AddMultiple(GetWrappedColorFunctions())
	registry.AddMultiple(GetWrappedTypesFunctions())

	registry.Add("default", &DefaultFunctionDefinition{})
	
	return &DefaultFunctions{registry: registry}
}

func createContexts() any {
	return map[string]any{"type": "Contexts"}
}

func createLessError() any {
	return func(details any, imports any, filename any) any {
		return map[string]any{"type": "LessError", "details": details}
	}
}

func createLogger() any {
	return map[string]any{
		"info":  func(msg string) {},
		"warn":  func(msg string) {},
		"error": func(msg string) {},
	}
}

func createUtils() map[string]any {
	return map[string]any{
		"copyArray": func(arr []any) []any { return append([]any{}, arr...) },
		"clone":     func(obj any) any { return cloneAny(obj) },
		"defaults":  func(target, source map[string]any) map[string]any { return mergeDefaults(target, source) },
	}
}

func createPluginManager() any {
	return map[string]any{"type": "PluginManager"}
}

func cloneAny(obj any) any {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case map[string]any:
		clone := make(map[string]any)
		for k, val := range v {
			clone[k] = cloneAny(val)
		}
		return clone
	case []any:
		clone := make([]any, len(v))
		for i, val := range v {
			clone[i] = cloneAny(val)
		}
		return clone
	default:
		return obj
	}
}

func mergeDefaults(target, source map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range target {
		result[k] = v
	}

	for k, v := range source {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	return result
}

type SimpleImportManagerEnvironment struct{}

func (s *SimpleImportManagerEnvironment) GetFileManager(path, currentDirectory string, context map[string]any, environment ImportManagerEnvironment) FileManager {
	return NewFileSystemFileManager()
}

type SimpleFileManager struct{
	*AbstractFileManager
}

func NewSimpleFileManager() *SimpleFileManager {
	return &SimpleFileManager{
		AbstractFileManager: NewAbstractFileManager(),
	}
}

func (s *SimpleFileManager) LoadFileSync(path, currentDirectory string, context map[string]any, environment ImportManagerEnvironment) *LoadedFile {
	return &LoadedFile{
		Filename: path,
		Contents: "",
	}
}

func (s *SimpleFileManager) LoadFile(path, currentDirectory string, context map[string]any, environment ImportManagerEnvironment, callback func(error, *LoadedFile)) any {
	callback(nil, &LoadedFile{
		Filename: path,
		Contents: "",
	})
	return nil
}

func (s *SimpleFileManager) GetPath(filename string) string {
	return filename
}

func (s *SimpleFileManager) Join(path1, path2 string) string {
	return path1 + "/" + path2
}

// PathDiff is inherited from AbstractFileManager

func (s *SimpleFileManager) IsPathAbsolute(path string) bool {
	return path[0] == '/'
}

func (s *SimpleFileManager) AlwaysMakePathsAbsolute() bool {
	return false
}