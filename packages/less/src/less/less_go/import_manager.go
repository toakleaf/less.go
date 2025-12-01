package less_go

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Promise interface {
	Then(onSuccess func(*LoadedFile), onError func(error))
}

type PromiseResult struct {
	File  *LoadedFile
	Error error
}

type ConcretePromise struct {
	resultChan chan PromiseResult
}

func (p *ConcretePromise) Then(onSuccess func(*LoadedFile), onError func(error)) {
	go func() {
		result := <-p.resultChan
		if result.Error != nil {
			onError(result.Error)
		} else {
			onSuccess(result.File)
		}
	}()
}

func NewPromise() *ConcretePromise {
	return &ConcretePromise{
		resultChan: make(chan PromiseResult, 1),
	}
}

func (p *ConcretePromise) Resolve(file *LoadedFile) {
	p.resultChan <- PromiseResult{File: file, Error: nil}
}

func (p *ConcretePromise) Reject(err error) {
	p.resultChan <- PromiseResult{File: nil, Error: err}
}

// FileInfo contains information about a file being processed
// Mirrors the JavaScript FileInfo structure:
//
//	'rewriteUrls' - option - whether to adjust URL's to be relative
//	'filename' - full resolved filename of current file
//	'rootpath' - path to append to normal URLs for this node
//	'currentDirectory' - path to the current file, absolute
//	'rootFilename' - filename of the base file
//	'entryPath' - absolute path to the entry file
//	'reference' - whether the file should not be output and only output parts that are referenced
type FileInfo struct {
	RewriteUrls      bool   `json:"rewriteUrls"`
	Filename         string `json:"filename"`
	Rootpath         string `json:"rootpath"`
	CurrentDirectory string `json:"currentDirectory"`
	RootFilename     string `json:"rootFilename"`
	EntryPath        string `json:"entryPath"`
	Reference        bool   `json:"reference"`
}

type ImportManager struct {
	less                 any
	rootFilename         string
	paths                []string
	contents             map[string]string
	contentsIgnoredChars map[string]string
	mime                 string
	error                error
	context              map[string]any
	queue                []string
	files                map[string]*FileCache
	environment          ImportManagerEnvironment
}

type FileCache struct {
	Root    any            `json:"root"`
	Options map[string]any `json:"options"`
}

type ImportOptions struct {
	Optional    bool           `json:"optional"`
	Inline      bool           `json:"inline"`
	IsPlugin    bool           `json:"isPlugin"`
	PluginArgs  map[string]any `json:"pluginArgs"`
	Reference   bool           `json:"reference"`
	Multiple    bool           `json:"multiple"`
}

// DeferredPluginInfo holds information for loading plugins at the correct scope depth
// during evaluation rather than globally during import.
type DeferredPluginInfo struct {
	Path             string
	CurrentDirectory string
	PluginArgs       map[string]any
	FullPath         string
}

type ImportManagerEnvironment interface {
	GetFileManager(path, currentDirectory string, context map[string]any, environment ImportManagerEnvironment) FileManager
}

type FileManager interface {
	LoadFileSync(path, currentDirectory string, context map[string]any, environment ImportManagerEnvironment) *LoadedFile
	LoadFile(path, currentDirectory string, context map[string]any, environment ImportManagerEnvironment, callback func(error, *LoadedFile)) any
	GetPath(filename string) string
	Join(path1, path2 string) string
	PathDiff(currentDirectory, entryPath string) string
	IsPathAbsolute(path string) bool
	AlwaysMakePathsAbsolute() bool
}

type LoadedFile struct {
	Filename string `json:"filename"`
	Contents string `json:"contents"`
	Message  string `json:"message,omitempty"`
}

type ParseCallback func(error, any, bool, string)

type ParserInterface interface {
	Parse(str string, callback func(*LessError, *Ruleset), additionalData *AdditionalData)
}

func NewImportManager(environment ImportManagerEnvironment) func(less any, context map[string]any, rootFileInfo *FileInfo) *ImportManager {
	return func(less any, context map[string]any, rootFileInfo *FileInfo) *ImportManager {
		var paths []string
		if pathsVal, exists := context["paths"]; exists {
			if pathsStr, ok := pathsVal.(string); ok {
				paths = []string{pathsStr}
			} else if pathsSlice, ok := pathsVal.([]string); ok {
				paths = pathsSlice
			} else if pathsAny, ok := pathsVal.([]any); ok {
				paths = make([]string, len(pathsAny))
				for i, p := range pathsAny {
					if pathStr, ok := p.(string); ok {
						paths[i] = pathStr
					}
				}
			}
		}
		if paths == nil {
			paths = []string{}
		}

		var mime string
		if mimeVal, exists := context["mime"]; exists {
			if mimeStr, ok := mimeVal.(string); ok {
				mime = mimeStr
			}
		}

		return &ImportManager{
			less:                 less,
			rootFilename:         rootFileInfo.Filename,
			paths:                paths,
			contents:             make(map[string]string),
			contentsIgnoredChars: make(map[string]string),
			mime:                 mime,
			error:                nil,
			context:              context,
			queue:                []string{},
			files:                make(map[string]*FileCache),
			environment:          environment,
		}
	}
}

func (im *ImportManager) Push(path string, tryAppendExtension bool, currentFileInfo *FileInfo, importOptions *ImportOptions, callback ParseCallback) {
	var pluginLoader PluginLoader
	if pluginManager, exists := im.context["pluginManager"]; exists {
		if pm, ok := pluginManager.(*PluginManager); ok && pm != nil && pm.Loader != nil {
			pluginLoader = pm.Loader
		} else if pm, ok := pluginManager.(map[string]any); ok {
			if loader, exists := pm["Loader"]; exists {
				if pl, ok := loader.(PluginLoader); ok {
					pluginLoader = pl
				}
			}
		}
	}

	im.queue = append(im.queue, path)

	fileParsedFunc := func(e error, root any, fullPath string) {
		for i, queuePath := range im.queue {
			if queuePath == path {
				im.queue = append(im.queue[:i], im.queue[i+1:]...)
				break
			}
		}

		importedEqualsRoot := fullPath == im.rootFilename

		if importOptions.Optional && e != nil {
			callback(nil, map[string]any{"rules": []any{}}, false, "")
			logger := NewLogger()
			logger.Info(fmt.Sprintf("The file %s was skipped because it was not found and the import was marked optional.", fullPath))
		} else {
			if _, exists := im.files[fullPath]; !exists && !importOptions.Inline {
				im.files[fullPath] = &FileCache{
					Root:    root,
					Options: map[string]any{},
				}
				if importOptions.Optional {
					im.files[fullPath].Options["optional"] = true
				}
				if importOptions.Inline {
					im.files[fullPath].Options["inline"] = true
				}
				if importOptions.IsPlugin {
					im.files[fullPath].Options["isPlugin"] = true
				}
				if importOptions.Reference {
					im.files[fullPath].Options["reference"] = true
				}
				if importOptions.Multiple {
					im.files[fullPath].Options["multiple"] = true
				}
				if importOptions.PluginArgs != nil {
					im.files[fullPath].Options["pluginArgs"] = importOptions.PluginArgs
				}
			}

			if e != nil && im.error == nil {
				im.error = e
			}

			callback(e, root, importedEqualsRoot, fullPath)
		}
	}

	newFileInfo := &FileInfo{
		RewriteUrls:  false,
		EntryPath:    currentFileInfo.EntryPath,
		Rootpath:     currentFileInfo.Rootpath,
		RootFilename: currentFileInfo.RootFilename,
	}

	if rewriteUrls, exists := im.context["rewriteUrls"]; exists {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG ImportManager.Push] context rewriteUrls value=%v (type=%T)\n", rewriteUrls, rewriteUrls)
		}
		if rwUrls, ok := rewriteUrls.(bool); ok {
			newFileInfo.RewriteUrls = rwUrls
		} else if rwType, ok := rewriteUrls.(RewriteUrlsType); ok {
			newFileInfo.RewriteUrls = rwType != RewriteUrlsOff
			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG ImportManager.Push] converted RewriteUrlsType %d to bool %v\n", rwType, newFileInfo.RewriteUrls)
			}
		}
	}

	fileManager := im.environment.GetFileManager(path, currentFileInfo.CurrentDirectory, im.context, im.environment)

	if fileManager == nil {
		fileParsedFunc(fmt.Errorf("Could not find a file-manager for %s", path), nil, "")
		return
	}

	loadFileCallback := func(loadedFile *LoadedFile) {
		if loadedFile == nil {
			fileParsedFunc(fmt.Errorf("LoadedFile is nil for path: %s", path), nil, "")
			return
		}
		if loadedFile.Message != "" {
			fileParsedFunc(fmt.Errorf("%s", loadedFile.Message), nil, "")
			return
		}

		resolvedFilename := loadedFile.Filename
		contents := strings.TrimPrefix(loadedFile.Contents, "\uFEFF")

		newFileInfo.CurrentDirectory = fileManager.GetPath(resolvedFilename)

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG ImportManager.loadFileCallback] newFileInfo.RewriteUrls=%v, currentDir=%q, entryPath=%q\n",
				newFileInfo.RewriteUrls, newFileInfo.CurrentDirectory, newFileInfo.EntryPath)
		}

		if newFileInfo.RewriteUrls {
			rootpath := ""
			if contextRootpath, exists := im.context["rootpath"]; exists {
				if rp, ok := contextRootpath.(string); ok {
					rootpath = rp
				}
			}

			pathDiff := fileManager.PathDiff(newFileInfo.CurrentDirectory, newFileInfo.EntryPath)
			newFileInfo.Rootpath = fileManager.Join(rootpath, pathDiff)

			if os.Getenv("LESS_GO_DEBUG") == "1" {
				fmt.Printf("[DEBUG ImportManager.loadFileCallback] Calculating rootpath: contextRootpath=%q, pathDiff=%q, result=%q\n",
					rootpath, pathDiff, newFileInfo.Rootpath)
			}

			if !fileManager.IsPathAbsolute(newFileInfo.Rootpath) && fileManager.AlwaysMakePathsAbsolute() {
				newFileInfo.Rootpath = fileManager.Join(newFileInfo.EntryPath, newFileInfo.Rootpath)
				if os.Getenv("LESS_GO_DEBUG") == "1" {
					fmt.Printf("[DEBUG ImportManager.loadFileCallback] Made absolute: %q\n", newFileInfo.Rootpath)
				}
			}
		}

		newFileInfo.Filename = resolvedFilename

		contextWithPaths := im.cloneContext()
		if len(im.paths) > 0 {
			contextWithPaths["paths"] = im.paths
		}
		newEnv := NewParse(contextWithPaths)
		newEnv.ProcessImports = false
		im.contents[resolvedFilename] = contents

		if currentFileInfo.Reference || importOptions.Reference {
			newFileInfo.Reference = true
		}

		if importOptions.IsPlugin {
			plugin := pluginLoader.EvalPlugin(contents, newEnv, im, importOptions.PluginArgs, newFileInfo)
			if lessErr, ok := plugin.(*LessError); ok {
				var err error
				if lessErr != nil {
					err = lessErr
				}
				fileParsedFunc(err, nil, resolvedFilename)
			} else {
				fileParsedFunc(nil, plugin, resolvedFilename)
			}
		} else if importOptions.Inline {
			fileParsedFunc(nil, contents, resolvedFilename)
		} else {
			// Check per-compilation cache first (for @import (multiple) handling)
			if cached, exists := im.files[resolvedFilename]; exists {
				cachedMultiple := false
				if mult, ok := cached.Options["multiple"].(bool); ok {
					cachedMultiple = mult
				}

				if !cachedMultiple && !importOptions.Multiple {
					fileParsedFunc(nil, cached.Root, resolvedFilename)
					return
				}
			}

			// Check global parse cache for unchanged files (incremental compilation optimization)
			// Skip cache for @import (multiple) which requires fresh parsing each time
			globalCache := GetGlobalParseCache()
			if !importOptions.Multiple {
				if cachedRoot, hit := globalCache.Get(resolvedFilename, contents); hit {
					if os.Getenv("LESS_GO_DEBUG") == "1" {
						fmt.Printf("[DEBUG ImportManager] Global cache HIT for %s\n", resolvedFilename)
					}
					fileParsedFunc(nil, cachedRoot, resolvedFilename)
					return
				}
			}

			parserContext := im.cloneContext()
			parserImports := map[string]any{
				"contents": im.contents,
			}
			parserFileInfo := map[string]any{
				"filename":         newFileInfo.Filename,
				"rootpath":         newFileInfo.Rootpath,
				"currentDirectory": newFileInfo.CurrentDirectory,
				"rootFilename":     newFileInfo.RootFilename,
				"entryPath":        newFileInfo.EntryPath,
				"reference":        newFileInfo.Reference,
			}

			if parserFactory, exists := im.context["parserFactory"]; exists {
				if pf, ok := parserFactory.(func(map[string]any, map[string]any, map[string]any, int) ParserInterface); ok {
					parser := pf(parserContext, parserImports, parserFileInfo, 0)
					parser.Parse(contents, func(e *LessError, root *Ruleset) {
						var err error
						if e != nil {
							err = e
						}
						// Store in global cache on successful parse
						if e == nil && root != nil {
							globalCache.Put(resolvedFilename, contents, root)
							if os.Getenv("LESS_GO_DEBUG") == "1" {
								fmt.Printf("[DEBUG ImportManager] Global cache STORE for %s\n", resolvedFilename)
							}
						}
						fileParsedFunc(err, root, resolvedFilename)
					}, nil)
				}
			} else {
				fileParsedFunc(fmt.Errorf("no parser factory available"), nil, resolvedFilename)
			}
		}
	}

	context := im.cloneContext()

	if len(im.paths) > 0 {
		context["paths"] = im.paths
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG ImportManager.Push] Setting context paths from im.paths: %v\n", im.paths)
		}
	} else {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Printf("[DEBUG ImportManager.Push] im.paths is empty\n")
		}
	}

	if tryAppendExtension {
		if importOptions.IsPlugin {
			context["ext"] = ".js"
		} else {
			context["ext"] = ".less"
		}
	}

	var loadedFile *LoadedFile
	var promise any

	if importOptions.IsPlugin {
		if pluginLoader == nil {
			fileParsedFunc(fmt.Errorf("plugin imports are not supported: no plugin loader available for '%s'", path), nil, path)
			return
		}

		// Defer plugin loading until evaluation to register at correct scope depth
		deferredInfo := &DeferredPluginInfo{
			Path:             path,
			CurrentDirectory: currentFileInfo.CurrentDirectory,
			PluginArgs:       importOptions.PluginArgs,
			FullPath:         path,
		}

		if fileManager != nil {
			deferredInfo.FullPath = fileManager.Join(currentFileInfo.CurrentDirectory, path)
		}

		fileParsedFunc(nil, deferredInfo, deferredInfo.FullPath)
		return
	}

	if syncImport, exists := context["syncImport"]; exists && syncImport.(bool) {
		loadedFile = fileManager.LoadFileSync(path, currentFileInfo.CurrentDirectory, context, im.environment)
	} else {
		promise = fileManager.LoadFile(path, currentFileInfo.CurrentDirectory, context, im.environment,
			func(err error, file *LoadedFile) {
				if err != nil {
					fileParsedFunc(err, nil, "")
				} else {
					loadFileCallback(file)
				}
			})
	}

	if loadedFile != nil {
		if loadedFile.Filename == "" {
			fileParsedFunc(fmt.Errorf("%s", loadedFile.Message), nil, "")
		} else {
			loadFileCallback(loadedFile)
		}
	} else if promise != nil {
		im.handlePromise(promise, loadFileCallback, fileParsedFunc)
	}
}

func (im *ImportManager) cloneContext() map[string]any {
	return Clone(im.context)
}

func (im *ImportManager) handlePromise(promise any, loadFileCallback func(*LoadedFile), fileParsedFunc func(error, any, string)) {
	switch p := promise.(type) {
	case *ConcretePromise:
		p.Then(loadFileCallback, func(err error) {
			fileParsedFunc(err, nil, "")
		})
	case Promise:
		p.Then(loadFileCallback, func(err error) {
			fileParsedFunc(err, nil, "")
		})
	case <-chan *LoadedFile:
		go func() {
			result := <-p
			if result != nil {
				loadFileCallback(result)
			} else {
				fileParsedFunc(fmt.Errorf("channel closed unexpectedly"), nil, "")
			}
		}()
	case <-chan PromiseResult:
		go func() {
			result := <-p
			if result.Error != nil {
				fileParsedFunc(result.Error, nil, "")
			} else {
				loadFileCallback(result.File)
			}
		}()
	case func() (*LoadedFile, error):
		go func() {
			file, err := p()
			if err != nil {
				fileParsedFunc(err, nil, "")
			} else {
				loadFileCallback(file)
			}
		}()
	default:
		if im.tryReflectionPromise(promise, loadFileCallback, fileParsedFunc) {
			return
		}
		fileParsedFunc(fmt.Errorf("unsupported promise type: %T", promise), nil, "")
	}
}

func (im *ImportManager) tryReflectionPromise(promise any, loadFileCallback func(*LoadedFile), fileParsedFunc func(error, any, string)) bool {
	v := reflect.ValueOf(promise)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	thenMethod := v.MethodByName("Then")
	if thenMethod.IsValid() && thenMethod.Type().NumIn() == 2 {
		successCallback := reflect.ValueOf(loadFileCallback)
		errorCallback := reflect.ValueOf(func(err error) {
			fileParsedFunc(err, nil, "")
		})
		
		go func() {
			thenMethod.Call([]reflect.Value{successCallback, errorCallback})
		}()
		return true
	}

	if v.Kind() == reflect.Chan {
		go func() {
			chosen, recv, recvOK := reflect.Select([]reflect.SelectCase{
				{Dir: reflect.SelectRecv, Chan: v},
			})

			if chosen == 0 && recvOK {
				if file, ok := recv.Interface().(*LoadedFile); ok {
					loadFileCallback(file)
				} else if result, ok := recv.Interface().(PromiseResult); ok {
					if result.Error != nil {
						fileParsedFunc(result.Error, nil, "")
					} else {
						loadFileCallback(result.File)
					}
				} else {
					fileParsedFunc(fmt.Errorf("unexpected channel value type: %T", recv.Interface()), nil, "")
				}
			} else {
				fileParsedFunc(fmt.Errorf("channel closed unexpectedly"), nil, "")
			}
		}()
		return true
	}
	
	return false
}

func (im *ImportManager) Contents() map[string]string {
	return im.contents
}

func (im *ImportManager) RootFilename() string {
	return im.rootFilename
}

func (im *ImportManager) Files() map[string]*FileCache {
	return im.files
}