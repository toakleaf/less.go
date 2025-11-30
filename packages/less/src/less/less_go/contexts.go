package less_go

import (
	"fmt"
	"os"
	"strings"
)

type Parse struct {
	Paths           []string
	RewriteUrls     RewriteUrlsType
	Rootpath        string
	StrictImports   bool
	Insecure        bool
	DumpLineNumbers bool
	Compress        bool
	SyncImport      bool
	ChunkInput      bool
	Mime            string
	UseFileCache    bool
	ProcessImports  bool
	PluginManager   any
	Quiet         bool
}

func NewParse(options map[string]any) *Parse {
	p := &Parse{
		RewriteUrls: RewriteUrlsOff, // Default to OFF to match JavaScript default (false)
	}
	copyFromOriginal(options, p)
	if paths, ok := options["paths"].(string); ok {
		p.Paths = []string{paths}
	} else if paths, ok := options["paths"].([]string); ok {
		p.Paths = paths
	}
	return p
}

// ImportantScopeEntry represents a single entry in the !important scope stack.
// This replaces map[string]any with a typed struct to eliminate reflection overhead.
type ImportantScopeEntry struct {
	Important string // The " !important" suffix value when set
}

// Eval is the primary evaluation context for Less compilation.
// OPTIMIZATION: Uses typed fields instead of map[string]any to eliminate reflection overhead.
type Eval struct {
	Paths             []string
	Compress          bool
	Math              MathType
	StrictUnits       bool
	SourceMap         bool
	ImportMultiple    bool
	UrlArgs           string
	JavascriptEnabled bool
	PluginManager     any
	ImportantScope    []ImportantScopeEntry // Typed struct replaces []map[string]any
	RewriteUrls       RewriteUrlsType
	NumPrecision      int

	Frames       []any
	parserFrames []ParserFrame // Cached typed frames to avoid allocation in GetFrames()
	CalcStack    []bool
	ParensStack  []bool
	InCalc       bool
	MathOn       bool
	DefaultFunc  *DefaultFunc
	FunctionRegistry *Registry
	MediaBlocks  []any
	MediaPath    []any

	PluginBridge     *NodeJSPluginBridge
	LazyPluginBridge *LazyNodeJSPluginBridge // Lazy bridge for deferred initialization
}

func buildParserFramesCache(frames []any) []ParserFrame {
	if len(frames) == 0 {
		return nil
	}
	parserFrames := make([]ParserFrame, 0, len(frames))
	for _, frame := range frames {
		if parserFrame, ok := frame.(ParserFrame); ok {
			parserFrames = append(parserFrames, parserFrame)
		}
	}
	return parserFrames
}

func NewEval(options map[string]any, frames []any) *Eval {
	e := &Eval{
		Frames:         frames,
		parserFrames:   buildParserFramesCache(frames),
		MathOn:         true,
		ImportantScope: []ImportantScopeEntry{},
		NumPrecision:   0, // Default to 0 to preserve full JavaScript number precision
		RewriteUrls:    RewriteUrlsOff, // Default to OFF to match JavaScript default (false)
	}
	copyFromOriginal(options, e)
	if paths, ok := options["paths"].(string); ok {
		e.Paths = []string{paths}
	} else if paths, ok := options["paths"].([]string); ok {
		e.Paths = paths
	}
	return e
}

func NewEvalFromEval(parent *Eval, frames []any) *Eval {
	return &Eval{
		Paths:             parent.Paths,
		Compress:          parent.Compress,
		Math:              parent.Math,
		StrictUnits:       parent.StrictUnits,
		SourceMap:         parent.SourceMap,
		ImportMultiple:    parent.ImportMultiple,
		UrlArgs:           parent.UrlArgs,
		JavascriptEnabled: parent.JavascriptEnabled,
		PluginManager:     parent.PluginManager,
		ImportantScope:    parent.ImportantScope,
		RewriteUrls:       parent.RewriteUrls,
		NumPrecision:      parent.NumPrecision,
		Frames:            frames,
		parserFrames:      buildParserFramesCache(frames),
		CalcStack:         nil, // Fresh stacks for new context
		ParensStack:       nil,
		InCalc:            false,
		MathOn:            parent.MathOn,
		DefaultFunc:       parent.DefaultFunc,
		FunctionRegistry:  parent.FunctionRegistry,
		MediaBlocks:       parent.MediaBlocks,
		MediaPath:         parent.MediaPath,
		PluginBridge:      parent.PluginBridge,
		LazyPluginBridge:  parent.LazyPluginBridge,
	}
}

func (e *Eval) EnterCalc() {
	if e.CalcStack == nil {
		e.CalcStack = make([]bool, 0)
	}
	e.CalcStack = append(e.CalcStack, true)
	e.InCalc = true
}

func (e *Eval) ExitCalc() {
	if len(e.CalcStack) > 0 {
		e.CalcStack = e.CalcStack[:len(e.CalcStack)-1]
		if len(e.CalcStack) == 0 {
			e.InCalc = false
		}
	}
}

func (e *Eval) InParenthesis() {
	debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
	if e.ParensStack == nil {
		e.ParensStack = make([]bool, 0)
	}
	e.ParensStack = append(e.ParensStack, true)
	if debugTrace {
		fmt.Printf("[TRACE] InParenthesis: stack len now %d\n", len(e.ParensStack))
	}
}

func (e *Eval) OutOfParenthesis() {
	debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
	if len(e.ParensStack) > 0 {
		e.ParensStack = e.ParensStack[:len(e.ParensStack)-1]
	}
	if debugTrace {
		fmt.Printf("[TRACE] OutOfParenthesis: stack len now %d\n", len(e.ParensStack))
	}
}

func (e *Eval) ToMap() map[string]any {
	return map[string]any{
		"paths":             e.Paths,
		"compress":          e.Compress,
		"math":              e.Math,
		"strictUnits":       e.StrictUnits,
		"sourceMap":         e.SourceMap,
		"importMultiple":    e.ImportMultiple,
		"urlArgs":           e.UrlArgs,
		"javascriptEnabled": e.JavascriptEnabled,
		"pluginManager":     e.PluginManager,
		"importantScope":    e.GetImportantScopeAny(), // Convert to []map[string]any
		"rewriteUrls":       e.RewriteUrls,
		"numPrecision":      e.NumPrecision,
		"mediaBlocks":       e.MediaBlocks,
		"mediaPath":         e.MediaPath,
		"inParenthesis": func() {
			e.InParenthesis()
		},
		"outOfParenthesis": func() {
			e.OutOfParenthesis()
		},
		"isMathOn": func(op string) bool {
			return e.IsMathOnWithOp(op)
		},
		"inCalc": e.InCalc,
	}
}

func (e *Eval) CopyEvalToMap(target map[string]any, includeMediaContext bool) {
	target["paths"] = e.Paths
	target["compress"] = e.Compress
	target["math"] = e.Math
	target["strictUnits"] = e.StrictUnits
	target["sourceMap"] = e.SourceMap
	target["importMultiple"] = e.ImportMultiple
	target["urlArgs"] = e.UrlArgs
	target["javascriptEnabled"] = e.JavascriptEnabled
	target["pluginManager"] = e.PluginManager
	// Convert ImportantScope to []map[string]any for backward compatibility
	target["importantScope"] = e.GetImportantScopeAny()
	target["rewriteUrls"] = e.RewriteUrls
	target["numPrecision"] = e.NumPrecision
	target["inCalc"] = e.InCalc
	target["mathOn"] = e.MathOn
	target["_evalContext"] = e
	target["inParenthesis"] = func() {
		e.InParenthesis()
	}
	target["outOfParenthesis"] = func() {
		e.OutOfParenthesis()
	}
	target["isMathOn"] = func(op string) bool {
		return e.IsMathOnWithOp(op)
	}

	if includeMediaContext {
		target["mediaBlocks"] = e.MediaBlocks
		target["mediaPath"] = e.MediaPath
	}

	if e.PluginBridge != nil {
		target["pluginBridge"] = e.PluginBridge
	} else if e.LazyPluginBridge != nil {
		target["pluginBridge"] = e.LazyPluginBridge
	}
}

func (e *Eval) IsMathOn() bool {
	return e.IsMathOnWithOp("")
}

func (e *Eval) IsMathOnWithOp(op string) bool {
	debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
	if debugTrace {
		fmt.Printf("[TRACE] IsMathOnWithOp(%s): MathOn=%v, Math=%d, ParensStack len=%d\n", op, e.MathOn, e.Math, len(e.ParensStack))
	}
	if !e.MathOn {
		return false
	}
	if op == "/" && e.Math != MathAlways && (len(e.ParensStack) == 0) {
		if debugTrace {
			fmt.Printf("[TRACE] IsMathOnWithOp(%s): division without parens, returning false\n", op)
		}
		return false
	}
	if e.Math > MathParensDivision {
		result := len(e.ParensStack) > 0
		if debugTrace {
			fmt.Printf("[TRACE] IsMathOnWithOp(%s): PARENS mode, stack len=%d, returning %v\n", op, len(e.ParensStack), result)
		}
		return result
	}
	if debugTrace {
		fmt.Printf("[TRACE] IsMathOnWithOp(%s): returning true (default)\n", op)
	}
	return true
}

func (e *Eval) SetMathOn(mathOn bool) {
	e.MathOn = mathOn
}

func (e *Eval) IsInCalc() bool {
	return e.InCalc
}

func (e *Eval) GetFrames() []ParserFrame {
	return buildParserFramesCache(e.Frames)
}

func (e *Eval) GetImportantScope() []map[string]bool {
	// Convert from []ImportantScopeEntry to []map[string]bool for interface compatibility
	result := make([]map[string]bool, len(e.ImportantScope))
	for i, scope := range e.ImportantScope {
		scopeBool := make(map[string]bool, 1)
		if scope.Important != "" {
			scopeBool["important"] = true
		}
		result[i] = scopeBool
	}
	return result
}

func (e *Eval) GetDefaultFunc() *DefaultFunc {
	return e.DefaultFunc
}

func (e *Eval) PathRequiresRewrite(path string) bool {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG PathRequiresRewrite] path=%s, e.RewriteUrls=%d (Off=0, Local=1, All=2)\n", path, e.RewriteUrls)
	}
	if e.RewriteUrls == RewriteUrlsLocal {
		return isPathLocalRelative(path)
	}
	return isPathRelative(path)
}

func (e *Eval) RewritePath(path, rootpath string) string {
	if rootpath == "" {
		rootpath = ""
	}
	combined := rootpath + path
	newPath := e.NormalizePath(combined)

	// If a path was explicit relative and the rootpath was not an absolute path
	// we must ensure that the new path is also explicit relative.
	if isPathLocalRelative(path) &&
		isPathRelative(rootpath) &&
		!isPathLocalRelative(newPath) {
		newPath = "./" + newPath
	}

	return newPath
}

func (e *Eval) RewritePathForImport(path, rootpath string) string {
	if rootpath == "" {
		rootpath = ""
	}
	combined := rootpath + path
	return e.NormalizePath(combined)
}

func (e *Eval) NormalizePath(path string) string {
	segments := strings.Split(path, "/")
	pathSegments := make([]string, 0)

	for _, segment := range segments {
		switch segment {
		case ".":
			continue
		case "..":
			if len(pathSegments) == 0 || pathSegments[len(pathSegments)-1] == ".." {
				pathSegments = append(pathSegments, segment)
			} else {
				pathSegments = pathSegments[:len(pathSegments)-1]
			}
		default:
			pathSegments = append(pathSegments, segment)
		}
	}

	return strings.Join(pathSegments, "/")
}

func isPathRelative(path string) bool {
	if path == "" {
		return true
	}

	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "#") {
		return false
	}

	lowerPath := strings.ToLower(path)
	colonIndex := strings.Index(lowerPath, ":")
	if colonIndex > 0 {
		scheme := lowerPath[:colonIndex]
		for _, ch := range scheme {
			if !((ch >= 'a' && ch <= 'z') || ch == '-') {
				return true
			}
		}
		return false
	}

	return true
}

func isPathLocalRelative(path string) bool {
	return strings.HasPrefix(path, ".")
}

func copyFromOriginal(original map[string]any, destination any) {
	if original == nil {
		return
	}

	switch d := destination.(type) {
	case *Parse:
		if paths, ok := original["paths"].([]string); ok {
			d.Paths = paths
		} else if path, ok := original["paths"].(string); ok {
			d.Paths = []string{path}
		}
		if rewriteUrls, ok := original["rewriteUrls"].(RewriteUrlsType); ok {
			d.RewriteUrls = rewriteUrls
		} else if rewriteUrlsStr, ok := original["rewriteUrls"].(string); ok {
			// Convert string values to RewriteUrlsType enum
			switch rewriteUrlsStr {
			case "all":
				d.RewriteUrls = RewriteUrlsAll
			case "local":
				d.RewriteUrls = RewriteUrlsLocal
			case "off", "false":
				d.RewriteUrls = RewriteUrlsOff
			}
		}
		if rootpath, ok := original["rootpath"].(string); ok {
			d.Rootpath = rootpath
		}
		if strictImports, ok := original["strictImports"].(bool); ok {
			d.StrictImports = strictImports
		}
		if insecure, ok := original["insecure"].(bool); ok {
			d.Insecure = insecure
		}
		if dumpLineNumbers, ok := original["dumpLineNumbers"].(bool); ok {
			d.DumpLineNumbers = dumpLineNumbers
		}
		if compress, ok := original["compress"].(bool); ok {
			d.Compress = compress
		}
		if syncImport, ok := original["syncImport"].(bool); ok {
			d.SyncImport = syncImport
		}
		if chunkInput, ok := original["chunkInput"].(bool); ok {
			d.ChunkInput = chunkInput
		}
		if mime, ok := original["mime"].(string); ok {
			d.Mime = mime
		}
		if useFileCache, ok := original["useFileCache"].(bool); ok {
			d.UseFileCache = useFileCache
		}
		if processImports, ok := original["processImports"].(bool); ok {
			d.ProcessImports = processImports
		}
		if pluginManager, ok := original["pluginManager"]; ok {
			d.PluginManager = pluginManager
		}
		if quiet, ok := original["quiet"].(bool); ok {
			d.Quiet = quiet
		}
	case *Eval:
		if paths, ok := original["paths"].([]string); ok {
			d.Paths = paths
		} else if path, ok := original["paths"].(string); ok {
			d.Paths = []string{path}
		}
		if compress, ok := original["compress"].(bool); ok {
			d.Compress = compress
		}
		debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
		if math, ok := original["math"].(MathType); ok {
			d.Math = math
			if mathOn, hasMathOn := original["mathOn"].(bool); hasMathOn {
				d.MathOn = mathOn
			} else {
				d.MathOn = true
			}
			if debugTrace {
				fmt.Printf("[TRACE] copyFromOriginal: MathType case, Math=%d, MathOn=%v\n", d.Math, d.MathOn)
			}
		} else if mathStr, ok := original["math"].(string); ok {
			switch mathStr {
			case "always":
				d.Math = MathAlways
			case "parens-division":
				d.Math = MathParensDivision
			case "parens", "strict":
				d.Math = MathParens
			}
			if mathOn, hasMathOn := original["mathOn"].(bool); hasMathOn {
				d.MathOn = mathOn
			} else {
				d.MathOn = true
			}
			debugTrace := os.Getenv("LESS_GO_TRACE") == "1"
			if debugTrace {
				fmt.Printf("[TRACE] copyFromOriginal: set Math=%d, MathOn=%v\n", d.Math, d.MathOn)
			}
		} else if mathInt, ok := original["math"].(int); ok {
			d.Math = MathType(mathInt)
			if mathOn, hasMathOn := original["mathOn"].(bool); hasMathOn {
				d.MathOn = mathOn
			} else {
				d.MathOn = true
			}
		} else if mathOn, ok := original["mathOn"].(bool); ok {
			d.MathOn = mathOn
		}
		if strictUnits, ok := original["strictUnits"].(bool); ok {
			d.StrictUnits = strictUnits
		}
		if sourceMap, ok := original["sourceMap"].(bool); ok {
			d.SourceMap = sourceMap
		}
		if importMultiple, ok := original["importMultiple"].(bool); ok {
			d.ImportMultiple = importMultiple
		}
		if urlArgs, ok := original["urlArgs"].(string); ok {
			d.UrlArgs = urlArgs
		}
		if javascriptEnabled, ok := original["javascriptEnabled"].(bool); ok {
			d.JavascriptEnabled = javascriptEnabled
		}
		if pluginManager, ok := original["pluginManager"]; ok {
			d.PluginManager = pluginManager
		}
		// Handle ImportantScope - convert from various formats to []ImportantScopeEntry
		if importantScope, ok := original["importantScope"].([]ImportantScopeEntry); ok {
			d.ImportantScope = importantScope
		} else if importantScope, ok := original["importantScope"].([]map[string]any); ok {
			d.ImportantScope = make([]ImportantScopeEntry, len(importantScope))
			for i, scope := range importantScope {
				if imp, ok := scope["important"].(string); ok {
					d.ImportantScope[i].Important = imp
				}
			}
		} else if importantScope, ok := original["importantScope"].([]any); ok {
			d.ImportantScope = make([]ImportantScopeEntry, len(importantScope))
			for i, scope := range importantScope {
				if scopeMap, ok := scope.(map[string]any); ok {
					if imp, ok := scopeMap["important"].(string); ok {
						d.ImportantScope[i].Important = imp
					}
				}
			}
		}
		if rewriteUrls, ok := original["rewriteUrls"].(RewriteUrlsType); ok {
			d.RewriteUrls = rewriteUrls
		} else if rewriteUrlsStr, ok := original["rewriteUrls"].(string); ok {
			// Convert string values to RewriteUrlsType enum
			switch rewriteUrlsStr {
			case "all":
				d.RewriteUrls = RewriteUrlsAll
			case "local":
				d.RewriteUrls = RewriteUrlsLocal
			case "off", "false":
				d.RewriteUrls = RewriteUrlsOff
			}
		}
		if numPrecision, ok := original["numPrecision"].(int); ok {
			d.NumPrecision = numPrecision
		}
		if pluginBridge, ok := original["pluginBridge"].(*NodeJSPluginBridge); ok {
			d.PluginBridge = pluginBridge
		} else if lazyBridge, ok := original["pluginBridge"].(*LazyNodeJSPluginBridge); ok {
			d.LazyPluginBridge = lazyBridge
		}
	}
}

func (e *Eval) EnterPluginScope() any {
	if e.PluginBridge != nil {
		return e.PluginBridge.EnterScope()
	}
	// Also check LazyPluginBridge if PluginBridge is nil
	if e.LazyPluginBridge != nil {
		return e.LazyPluginBridge.EnterScope()
	}
	return nil
}

func (e *Eval) ExitPluginScope() {
	if e.PluginBridge != nil {
		e.PluginBridge.ExitScope()
		return
	}
	// Also check LazyPluginBridge if PluginBridge is nil
	if e.LazyPluginBridge != nil {
		e.LazyPluginBridge.ExitScope()
	}
}

func (e *Eval) LookupPluginFunction(name string) (any, bool) {
	if e.PluginBridge != nil {
		return e.PluginBridge.LookupFunction(name)
	}
	// Also check LazyPluginBridge if PluginBridge is nil
	if e.LazyPluginBridge != nil {
		return e.LazyPluginBridge.LookupFunction(name)
	}
	return nil, false
}

func (e *Eval) HasPluginFunction(name string) bool {
	if e.PluginBridge != nil {
		return e.PluginBridge.HasFunction(name)
	}
	// Also check LazyPluginBridge if PluginBridge is nil
	if e.LazyPluginBridge != nil {
		return e.LazyPluginBridge.HasFunction(name)
	}
	return false
}

func (e *Eval) CallPluginFunction(name string, args ...any) (any, error) {
	if e.PluginBridge != nil {
		return e.PluginBridge.CallFunctionWithContext(name, e, args...)
	}
	// Also check LazyPluginBridge if PluginBridge is nil
	if e.LazyPluginBridge != nil {
		return e.LazyPluginBridge.CallFunctionWithContext(name, e, args...)
	}
	return nil, fmt.Errorf("no plugin bridge available")
}

func (e *Eval) GetFramesAny() []any {
	return e.Frames
}

// NewMixinEvalContext creates a new *Eval context for mixin evaluation.
// OPTIMIZATION: Directly creates *Eval instead of map[string]any, avoiding reflection.
// The new context shares the parent's configuration but has new frames and fresh media context.
func (e *Eval) NewMixinEvalContext(frames []any) *Eval {
	return &Eval{
		Paths:             e.Paths,
		Compress:          e.Compress,
		Math:              e.Math,
		StrictUnits:       e.StrictUnits,
		SourceMap:         e.SourceMap,
		ImportMultiple:    e.ImportMultiple,
		UrlArgs:           e.UrlArgs,
		JavascriptEnabled: e.JavascriptEnabled,
		PluginManager:     e.PluginManager,
		ImportantScope:    e.ImportantScope, // Share the same scope stack
		RewriteUrls:       e.RewriteUrls,
		NumPrecision:      e.NumPrecision,
		Frames:            frames,
		parserFrames:      buildParserFramesCache(frames),
		CalcStack:         nil, // Fresh stacks
		ParensStack:       nil,
		InCalc:            e.InCalc,
		MathOn:            e.MathOn,
		DefaultFunc:       e.DefaultFunc,
		FunctionRegistry:  e.FunctionRegistry,
		// NOTE: MediaBlocks and MediaPath are NOT copied - matching JavaScript behavior
		// where mixin body evaluation gets a fresh media context
		PluginBridge:     e.PluginBridge,
		LazyPluginBridge: e.LazyPluginBridge,
	}
}

// CopyWithFrames creates a shallow copy of the Eval context with new frames.
// OPTIMIZATION: This is more efficient than CopyEvalToMap for internal use.
func (e *Eval) CopyWithFrames(frames []any) *Eval {
	return &Eval{
		Paths:             e.Paths,
		Compress:          e.Compress,
		Math:              e.Math,
		StrictUnits:       e.StrictUnits,
		SourceMap:         e.SourceMap,
		ImportMultiple:    e.ImportMultiple,
		UrlArgs:           e.UrlArgs,
		JavascriptEnabled: e.JavascriptEnabled,
		PluginManager:     e.PluginManager,
		ImportantScope:    e.ImportantScope,
		RewriteUrls:       e.RewriteUrls,
		NumPrecision:      e.NumPrecision,
		Frames:            frames,
		parserFrames:      buildParserFramesCache(frames),
		CalcStack:         e.CalcStack,
		ParensStack:       e.ParensStack,
		InCalc:            e.InCalc,
		MathOn:            e.MathOn,
		DefaultFunc:       e.DefaultFunc,
		FunctionRegistry:  e.FunctionRegistry,
		MediaBlocks:       e.MediaBlocks,
		MediaPath:         e.MediaPath,
		PluginBridge:      e.PluginBridge,
		LazyPluginBridge:  e.LazyPluginBridge,
	}
}

// GetImportantScopeAny converts ImportantScope to []map[string]any for backward compatibility.
// OPTIMIZATION: This is only called when converting to map contexts; direct struct access is preferred.
func (e *Eval) GetImportantScopeAny() []map[string]any {
	result := make([]map[string]any, len(e.ImportantScope))
	for i, scope := range e.ImportantScope {
		m := make(map[string]any, 1)
		if scope.Important != "" {
			m["important"] = scope.Important
		}
		result[i] = m
	}
	return result
}

// PushImportantScope adds a new empty scope entry to the important scope stack.
// OPTIMIZATION: Direct struct manipulation instead of map allocation.
func (e *Eval) PushImportantScope() {
	e.ImportantScope = append(e.ImportantScope, ImportantScopeEntry{})
}

// PopImportantScope removes the top scope entry from the important scope stack.
func (e *Eval) PopImportantScope() ImportantScopeEntry {
	if len(e.ImportantScope) == 0 {
		return ImportantScopeEntry{}
	}
	last := e.ImportantScope[len(e.ImportantScope)-1]
	e.ImportantScope = e.ImportantScope[:len(e.ImportantScope)-1]
	return last
}

// SetImportantInCurrentScope sets the important value in the current scope.
func (e *Eval) SetImportantInCurrentScope(important string) {
	if len(e.ImportantScope) > 0 {
		e.ImportantScope[len(e.ImportantScope)-1].Important = important
	}
}

// GetImportantFromCurrentScope gets the important value from the current scope.
func (e *Eval) GetImportantFromCurrentScope() string {
	if len(e.ImportantScope) > 0 {
		return e.ImportantScope[len(e.ImportantScope)-1].Important
	}
	return ""
}