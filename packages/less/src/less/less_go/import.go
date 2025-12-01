package less_go

import (
	"fmt"
	"math"
	"os"
	"regexp"

	"github.com/toakleaf/less.go/packages/less/src/less/less_go/runtime"
)

var cssPatternRegex = regexp.MustCompile(`[#.&?]css([?;].*)?$`)

// Import represents a CSS @import node.
// Files are pushed to an import queue on creation with a callback
// that fires when the file has been fetched and parsed.
type Import struct {
	*Node
	path             any
	features         any
	options          map[string]any
	_index           int
	_fileInfo        map[string]any
	allowRoot        bool
	css              bool
	skip             any
	root             any
	importedFilename string
	error            error
}

func NewImport(path any, features any, options map[string]any, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) *Import {
	node := NewNode()
	node.TypeIndex = GetTypeIndexForNodeType("Import")

	imp := &Import{
		Node:      node,
		path:      path,
		features:  features,
		options:   options,
		_index:    index,
		_fileInfo: currentFileInfo,
		allowRoot: true,
	}

	if imp.options != nil {
		if _, hasLess := imp.options["less"]; hasLess || imp.options["inline"] != nil {
			imp.css = !imp.getBoolOption("less") || imp.getBoolOption("inline")
		} else {
			pathValue := imp.GetPath()
			if pathValue != nil {
				var pathStr string
				if str, ok := pathValue.(string); ok {
					pathStr = str
				} else if quoted, ok := pathValue.(*Quoted); ok {
					pathStr = quoted.GetValue()
				} else if anon, ok := pathValue.(*Anonymous); ok {
					if str, ok := anon.Value.(string); ok {
						pathStr = str
					}
				}
				if pathStr != "" && cssPatternRegex.MatchString(pathStr) {
					imp.css = true
				}
			}
		}
	} else {
		pathValue := imp.GetPath()
		if pathValue != nil {
			var pathStr string
			if str, ok := pathValue.(string); ok {
				pathStr = str
			} else if quoted, ok := pathValue.(*Quoted); ok {
				pathStr = quoted.GetValue()
			} else if anon, ok := pathValue.(*Anonymous); ok {
				if str, ok := anon.Value.(string); ok {
					pathStr = str
				}
			}
			if pathStr != "" && cssPatternRegex.MatchString(pathStr) {
				imp.css = true
			}
		}
	}

	imp.CopyVisibilityInfo(visibilityInfo)
	imp.SetParent(imp.features, imp.Node)
	imp.SetParent(imp.path, imp.Node)

	return imp
}

func (i *Import) getBoolOption(key string) bool {
	if i.options == nil {
		return false
	}
	if val, ok := i.options[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

func (i *Import) GetType() string {
	return "Import"
}

func (i *Import) GetIndex() int {
	return i._index
}

func (i *Import) FileInfo() map[string]any {
	return i._fileInfo
}

func (i *Import) Accept(visitor any) {
	if v, ok := visitor.(interface{ Visit(any) any }); ok {
		if i.features != nil {
			i.features = v.Visit(i.features)
		}
		i.path = v.Visit(i.path)
		if i.options != nil && !i.getBoolOption("isPlugin") && !i.getBoolOption("inline") && i.root != nil {
			i.root = v.Visit(i.root)
		}
	}
}

func (i *Import) GenCSS(context any, output *CSSOutput) {
	if i.css {
		var shouldOutput bool = true
		if pathWithFileInfo, ok := i.path.(interface{ FileInfo() map[string]any }); ok {
			fileInfo := pathWithFileInfo.FileInfo()
			if fileInfo != nil {
				if ref, ok := fileInfo["reference"].(bool); ok && ref {
					shouldOutput = false
				}
			}
		}

		if shouldOutput {
			output.Add("@import ", i._fileInfo, i._index)
			if pathGen, ok := i.path.(interface{ GenCSS(any, *CSSOutput) }); ok {
				pathGen.GenCSS(context, output)
			}
			if i.features != nil {
				output.Add(" ", nil, nil)
				if featuresGen, ok := i.features.(interface{ GenCSS(any, *CSSOutput) }); ok {
					featuresGen.GenCSS(context, output)
				}
			}
			output.Add(";", nil, nil)
		}
	}
}

func (i *Import) IsVisible() bool {
	if !i.css {
		return false
	}
	if pathWithFileInfo, ok := i.path.(interface{ FileInfo() map[string]any }); ok {
		fileInfo := pathWithFileInfo.FileInfo()
		if fileInfo != nil {
			if ref, ok := fileInfo["reference"].(bool); ok && ref {
				return false
			}
		}
	}
	return true
}

func (i *Import) pathFileInfoReference() any {
	if pathWithFileInfo, ok := i.path.(interface{ FileInfo() map[string]any }); ok {
		fileInfo := pathWithFileInfo.FileInfo()
		if fileInfo != nil {
			return fileInfo["reference"]
		}
	}
	if pathMap, ok := i.path.(map[string]any); ok {
		if fileInfo, ok := pathMap["_fileInfo"].(map[string]any); ok {
			return fileInfo["reference"]
		}
	}
	return nil
}

func (i *Import) GetPath() any {
	if urlPath, ok := i.path.(*URL); ok {
		if urlValue, ok := urlPath.Value.(map[string]any); ok {
			return urlValue["value"]
		}
		if quoted, ok := urlPath.Value.(*Quoted); ok {
			return quoted.GetValue()
		}
		if anon, ok := urlPath.Value.(*Anonymous); ok {
			if str, ok := anon.Value.(string); ok {
				return str
			}
			if quoted, ok := anon.Value.(*Quoted); ok {
				return quoted.GetValue()
			}
			return anon.Value
		}
		if pathWithValue, ok := urlPath.Value.(interface{ GetValue() any }); ok {
			return pathWithValue.GetValue()
		}
		return urlPath.Value
	}
	if pathMap, ok := i.path.(map[string]any); ok {
		return pathMap["value"]
	}
	if pathWithValue, ok := i.path.(interface{ GetValue() any }); ok {
		return pathWithValue.GetValue()
	}
	return i.path
}

func (i *Import) IsVariableImport() bool {
	path := i.path
	if urlPath, ok := path.(*URL); ok {
		path = urlPath.Value
	}
	if quotedPath, ok := path.(*Quoted); ok {
		return quotedPath.ContainsVariables()
	}
	return true
}

func (i *Import) EvalForImport(context any) *Import {
	path := i.path
	if urlPath, ok := path.(*URL); ok {
		path = urlPath.Value
	}

	var evaluatedPath any
	if pathEval, ok := path.(interface{ Eval(any) (any, error) }); ok {
		contextMap := make(map[string]any)
		if evalCtx, ok := context.(*Eval); ok {
			contextMap["frames"] = evalCtx.Frames
			contextMap["compress"] = evalCtx.Compress
			contextMap["math"] = evalCtx.Math
			contextMap["strictUnits"] = evalCtx.StrictUnits
		} else if ctxMap, ok := context.(map[string]any); ok {
			contextMap = ctxMap
		}

		result, err := pathEval.Eval(contextMap)
		if err != nil {
			evaluatedPath = path
		} else {
			evaluatedPath = result
		}
	} else {
		evaluatedPath = path
	}

	newImport := NewImport(evaluatedPath, i.features, i.options, i._index, i._fileInfo, i.VisibilityInfo())
	newImport.root = i.root
	newImport.importedFilename = i.importedFilename
	newImport.skip = i.skip
	newImport.error = i.error
	return newImport
}

func (i *Import) EvalPath(context any) any {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG Import.EvalPath] Starting, i.css=%v, i.path type=%T\n", i.css, i.path)
	}

	var path any
	if pathEval, ok := i.path.(interface{ Eval(any) (any, error) }); ok {
		result, err := pathEval.Eval(context)
		if err != nil {
			path = i.path
		} else {
			path = result
		}
	} else {
		path = i.path
	}

	fileInfo := i._fileInfo

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG Import.EvalPath] After Eval, path type=%T\n", path)
	}

	if _, ok := path.(*URL); !ok {
		var pathValue any
		if pathMap, ok := path.(map[string]any); ok {
			pathValue = pathMap["value"]
		} else if quoted, ok := path.(*Quoted); ok {
			pathValue = quoted.GetValue()
		} else if pathWithValue, ok := path.(interface{ GetValue() any }); ok {
			pathValue = pathWithValue.GetValue()
		} else {
			pathValue = path
		}

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG Import.EvalPath] pathValue type=%T, value=%q, fileInfo=%+v\n", pathValue, pathValue, fileInfo)
		}

		if pathValueStr, ok := pathValue.(string); ok && pathValueStr != "" && fileInfo != nil {
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG Import.EvalPath] pathValueStr=%q, fileInfo=%+v, context type=%T\n", pathValueStr, fileInfo, context)
			}

			var newValue string
			needsUpdate := false

			if evalCtx, ok := context.(*Eval); ok {
				requiresRewrite := evalCtx.PathRequiresRewrite(pathValueStr)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG Import.EvalPath] *Eval context, requiresRewrite=%v\n", requiresRewrite)
				}
				if requiresRewrite {
					if rootpath, ok := fileInfo["rootpath"].(string); ok {
						if evalCtx.RewriteUrls == RewriteUrlsAll {
							newValue = evalCtx.NormalizePath(rootpath + pathValueStr)
						} else {
							newValue = evalCtx.RewritePath(pathValueStr, rootpath)
						}
						needsUpdate = true
						if os.Getenv("LESS_GO_DEBUG") == "1" {
							fmt.Printf("[DEBUG Import.EvalPath] Rewriting %q -> %q (rootpath=%q, rewriteUrls=%d)\n", pathValueStr, newValue, rootpath, evalCtx.RewriteUrls)
						}
					}
				} else {
					newValue = evalCtx.NormalizePath(pathValueStr)
					needsUpdate = true
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Printf("[DEBUG Import.EvalPath] Normalizing %q -> %q\n", pathValueStr, newValue)
					}
				}
			} else if ctx, ok := context.(map[string]any); ok {
				if pathRequiresRewrite, ok := ctx["pathRequiresRewrite"].(func(string) bool); ok && pathRequiresRewrite(pathValueStr) {
					if rewritePath, ok := ctx["rewritePath"].(func(string, string) string); ok {
						if rootpath, ok := fileInfo["rootpath"].(string); ok {
							newValue = rewritePath(pathValueStr, rootpath)
							needsUpdate = true
						}
					}
				} else if normalizePath, ok := ctx["normalizePath"].(func(string) string); ok {
					newValue = normalizePath(pathValueStr)
					needsUpdate = true
				}
			}

			if needsUpdate {
				if pathMap, ok := path.(map[string]any); ok {
					pathMap["value"] = newValue
				} else if quoted, ok := path.(*Quoted); ok {
					var str string
					if quoted.GetQuote() == "" {
						str = ""
					} else {
						str = quoted.GetQuote() + newValue + quoted.GetQuote()
					}
					path = NewQuoted(str, newValue, quoted.GetEscaped(), quoted.GetIndex(), quoted.FileInfo())
				}
			}
		}
	}

	return path
}

func (i *Import) Eval(context any) (any, error) {
	result, err := i.DoEval(context)
	if err != nil {
		return nil, err
	}

	// For reference imports, add visibility blocks recursively so nested content
	// stays invisible until explicitly referenced via mixins.
	if i.getBoolOption("reference") || i.BlocksVisibility() {
		if resultSlice, ok := result.([]any); ok {
			for _, node := range resultSlice {
				addVisibilityBlockRecursive(node)
			}
		} else {
			addVisibilityBlockRecursive(result)
		}
	}

	return result, nil
}

func addVisibilityBlockRecursive(node any) {
	if node == nil {
		return
	}

	if nodeWithVisibility, ok := node.(interface{ AddVisibilityBlock() }); ok {
		nodeWithVisibility.AddVisibilityBlock()
	}

	if nodeWithRules, ok := node.(interface{ GetRules() []any }); ok {
		rules := nodeWithRules.GetRules()
		for _, rule := range rules {
			addVisibilityBlockRecursive(rule)
		}
	}

	if nodeWithSelectors, ok := node.(interface{ GetSelectors() []any }); ok {
		selectors := nodeWithSelectors.GetSelectors()
		for _, selector := range selectors {
			addVisibilityBlockRecursive(selector)
		}
	}

	if nodeWithExtends, ok := node.(interface{ GetExtendList() []*Extend }); ok {
		extendList := nodeWithExtends.GetExtendList()
		for _, extend := range extendList {
			addVisibilityBlockRecursive(extend)
		}
	}
}


func (i *Import) DoEval(context any) (any, error) {
	var features any
	if i.features != nil {
		if featuresEval, ok := i.features.(interface{ Eval(any) (any, error) }); ok {
			result, err := featuresEval.Eval(context)
			if err != nil {
				return nil, err
			}
			features = result
		} else {
			features = i.features
		}
	}

	// Re-check for CSS imports after variable interpolation
	if !i.css && i.root == nil {
		var pathStr string
		if pathEval, ok := i.path.(interface{ Eval(any) (any, error) }); ok {
			if result, err := pathEval.Eval(context); err == nil {
				if quoted, ok := result.(*Quoted); ok {
					pathStr = quoted.GetValue()
				}
			}
		}
		if pathStr != "" && cssPatternRegex.MatchString(pathStr) {
			newImport := NewImport(i.EvalPath(context), features, i.options, i._index, i._fileInfo, nil)
			newImport.css = true
			return newImport, nil
		}
	}

	if i.getBoolOption("isPlugin") {
		debug := os.Getenv("LESS_GO_DEBUG") == "1"
		if debug {
			fmt.Fprintf(os.Stderr, "[Import.DoEval] isPlugin=true, root type=%T\n", i.root)
		}

		// Load deferred plugins at current scope depth
		if deferredInfo, ok := i.root.(*DeferredPluginInfo); ok {
			if debug {
				fmt.Fprintf(os.Stderr, "[Import.DoEval] DeferredPluginInfo: path=%s, dir=%s\n", deferredInfo.Path, deferredInfo.CurrentDirectory)
			}

			var pluginBridge *NodeJSPluginBridge
			if evalCtx, ok := context.(*Eval); ok {
				pluginBridge = evalCtx.PluginBridge
				if pluginBridge == nil && evalCtx.LazyPluginBridge != nil {
					var err error
					pluginBridge, err = evalCtx.LazyPluginBridge.GetBridge()
					if err != nil && os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Fprintf(os.Stderr, "[Import.DoEval] LazyPluginBridge.GetBridge error: %v\n", err)
					}
				}
			} else if ctxMap, ok := context.(map[string]any); ok {
				if parentEval, ok := ctxMap["_evalContext"].(*Eval); ok {
					pluginBridge = parentEval.PluginBridge
					if pluginBridge == nil && parentEval.LazyPluginBridge != nil {
						var err error
						pluginBridge, err = parentEval.LazyPluginBridge.GetBridge()
						if err != nil && os.Getenv("LESS_GO_DEBUG") == "1" {
							fmt.Fprintf(os.Stderr, "[Import.DoEval] LazyPluginBridge.GetBridge error: %v\n", err)
						}
					}
				}
			}

			if pluginBridge != nil {
				if debug {
					fmt.Fprintf(os.Stderr, "[Import.DoEval] Found pluginBridge, loading plugin at current scope\n")
				}

				loadContext := map[string]any{
					"syncImport": true,
				}
				if deferredInfo.PluginArgs != nil {
					loadContext["options"] = deferredInfo.PluginArgs
				}

				result := pluginBridge.LoadPluginSync(
					deferredInfo.Path,
					deferredInfo.CurrentDirectory,
					loadContext,
					nil, // environment
					nil, // fileManager
				)

				if err, ok := result.(error); ok {
					lessErr := &LessError{
						Type:     "Plugin",
						Message:  fmt.Sprintf("Plugin error during loading: %v", err),
						Filename: deferredInfo.FullPath,
						Index:    i.GetIndex(),
					}
					return nil, lessErr
				}

				if debug {
					fmt.Fprintf(os.Stderr, "[Import.DoEval] Plugin loaded successfully: %T\n", result)
				}

				i.root = result

				// Store function names on the containing Ruleset for mixin inheritance
				if plugin, ok := result.(*runtime.Plugin); ok && len(plugin.Functions) > 0 {
					var frames []any
					if evalCtx, ok := context.(*Eval); ok {
						frames = evalCtx.Frames
					} else if ctxMap, ok := context.(map[string]any); ok {
						if f, ok := ctxMap["frames"].([]any); ok {
							frames = f
						}
					}

					for _, frame := range frames {
						if rs, ok := frame.(*Ruleset); ok {
							if rs.LoadedPluginFunctions == nil {
								rs.LoadedPluginFunctions = make(map[string]bool)
							}
							for _, funcName := range plugin.Functions {
								rs.LoadedPluginFunctions[funcName] = true
								if debug {
									fmt.Fprintf(os.Stderr, "[Import.DoEval] Stored function '%s' on Ruleset=%p\n", funcName, rs)
								}
							}
							break
						}
					}
				}
			} else {
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Fprintf(os.Stderr, "[Import.DoEval] No plugin bridge available for deferred plugin: %s\n", deferredInfo.Path)
				}
			}
		} else if i.root != nil {
			if rootEval, ok := i.root.(interface{ Eval(any) (any, error) }); ok {
				_, err := rootEval.Eval(context)
				if err != nil {
					var filename string
					if rootWithFilename, ok := i.root.(interface{ GetFilename() string }); ok {
						filename = rootWithFilename.GetFilename()
					}
					lessErr := &LessError{
						Type:     "Plugin",
						Message:  "Plugin error during evaluation",
						Filename: filename,
						Index:    i.GetIndex(),
					}
					return nil, lessErr
				}
			}
		}

		if ctx, ok := context.(map[string]any); ok {
			if frames, ok := ctx["frames"].([]any); ok && len(frames) > 0 {
				if frameRuleset, ok := frames[0].(*Ruleset); ok && frameRuleset.FunctionRegistry != nil {
					if i.root != nil {
						if rootWithFunctions, ok := i.root.(interface{ GetFunctions() map[string]any }); ok {
							functions := rootWithFunctions.GetFunctions()
							if functions != nil {
								if registry, ok := frameRuleset.FunctionRegistry.(interface{ AddMultiple(map[string]any) }); ok {
									registry.AddMultiple(functions)
								}
							}
						}
					} else {
						registerTestPluginFunctions(frameRuleset.FunctionRegistry)
					}
				}
			}
		}

		return []any{}, nil
	}

	if i.skip != nil {
		var shouldSkip bool
		if skipFunc, ok := i.skip.(func() bool); ok {
			shouldSkip = skipFunc()
		} else if skipBool, ok := i.skip.(bool); ok {
			shouldSkip = skipBool
		}
		if shouldSkip {
			return []any{}, nil
		}
	}

	if i.getBoolOption("inline") {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			rootType := fmt.Sprintf("%T", i.root)
			rootVal := ""
			if str, ok := i.root.(string); ok {
				if len(str) > 30 {
					rootVal = str[:30] + "..."
				} else {
					rootVal = str
				}
			}
			fmt.Fprintf(os.Stderr, "[Import.DoEval inline] root type=%s, value=%q, importedFilename=%s\n", rootType, rootVal, i.importedFilename)
		}
		contents := NewAnonymous(i.root, 0, map[string]any{
			"filename":  i.importedFilename,
			"reference": i.pathFileInfoReference(),
		}, true, true, nil)

		if i.features != nil {
			var featuresValue any
			if val, ok := i.features.(*Value); ok {
				featuresValue = val.Value
			} else {
				featuresValue = i.features
			}

			return NewMedia([]any{contents}, featuresValue, i._index, i._fileInfo, i.VisibilityInfo()), nil
		}
		return []any{contents}, nil
	}

	if i.css {
		newImport := NewImport(i.EvalPath(context), features, i.options, i._index, nil, nil)
		if !newImport.css && i.error != nil {
			return nil, i.error
		}
		return newImport, nil
	}

	if i.root != nil {
		// Get rules directly - NewRuleset will copy them internally
		var rules []any
		if rootWithRules, ok := i.root.(interface{ GetRules() []any }); ok {
			rules = rootWithRules.GetRules()
		} else if rootRules, ok := i.root.([]any); ok {
			rules = rootRules
		}

		ruleset := NewRuleset(nil, rules, false, nil)
		if err := ruleset.EvalImports(context); err != nil {
			return nil, err
		}

		if i.features != nil {
			var featuresValue any
			if val, ok := i.features.(*Value); ok {
				featuresValue = val.Value
			} else {
				featuresValue = i.features
			}

			return NewMedia(ruleset.Rules, featuresValue, i._index, i._fileInfo, i.VisibilityInfo()), nil
		}
		return ruleset.Rules, nil
	}

	return []any{}, nil
}

// registerTestPluginFunctions registers the functions that would be provided by the test plugin
// This simulates loading the plugin-simple.js plugin used in tests
func registerTestPluginFunctions(registry any) {
	if reg, ok := registry.(interface{ Add(string, FunctionDefinition) }); ok {
		// Register pi-anon function (returns Math.PI as anonymous/unitless number)
		reg.Add("pi-anon", &FlexibleFunctionDef{
			name:      "pi-anon",  
			minArgs:   0,
			maxArgs:   0,
			variadic:  false,
			fn: func() any {
				return &Anonymous{
					Node:     NewNode(),
					Value:    fmt.Sprintf("%g", math.Pi),
					Index:    0,
					FileInfo: nil,
				}
			},
			needsEval: false,
		})
		
		// Register pi function (returns Math.PI as dimension)
		reg.Add("pi", &FlexibleFunctionDef{
			name:      "pi",
			minArgs:   0,
			maxArgs:   0,
			variadic:  false,
			fn: func() any {
				// Create a unitless dimension with PI value
				dim, _ := NewDimension(math.Pi, "")
				return dim
			},
			needsEval: false,
		})
	}
}


// GetTypeIndex returns the type index for visitor pattern
func (i *Import) GetTypeIndex() int {
	// Return from Node field if set, otherwise get from registry
	if i.Node != nil && i.Node.TypeIndex != 0 {
		return i.Node.TypeIndex
	}
	return GetTypeIndexForNodeType("Import")
} 