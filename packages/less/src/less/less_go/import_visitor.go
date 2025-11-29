package less_go

import (
	"fmt"
	"os"
)

// ImportVisitor processes import nodes in the AST
type ImportVisitor struct {
	visitor              *Visitor
	importer             any // Will be called with Push method
	finish               func(error)
	context              *Eval
	importCount          int
	onceFileDetectionMap map[string]bool
	recursionDetector    map[string]bool
	sequencer            *ImportSequencer
	isReplacing          bool
	isFinished           bool
	error                error
}

// NewImportVisitor creates a new ImportVisitor with the given importer and finish callback
func NewImportVisitor(importer any, finish func(error)) *ImportVisitor {
	iv := &ImportVisitor{
		importer:             importer,
		finish:               finish,
		context:              NewEval(nil, make([]any, 0)),
		importCount:          0,
		onceFileDetectionMap: make(map[string]bool),
		recursionDetector:    make(map[string]bool),
		isReplacing:          false,
		isFinished:           false,
	}

	iv.visitor = NewVisitor(iv)
	iv.sequencer = NewImportSequencer(iv.onSequencerEmpty)

	return iv
}

// IsReplacing returns the replacing status of the visitor
// This implements the Implementation interface to avoid reflection fallback
func (iv *ImportVisitor) IsReplacing() bool {
	return iv.isReplacing
}

// VisitNode implements direct dispatch without reflection for better performance
func (iv *ImportVisitor) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {
	switch n := node.(type) {
	case *Import:
		iv.VisitImport(n, visitArgs)
		return n, true
	case *Declaration:
		iv.VisitDeclaration(n, visitArgs)
		return n, true
	case *AtRule:
		iv.VisitAtRule(n, visitArgs)
		return n, true
	case *MixinDefinition:
		iv.VisitMixinDefinition(n, visitArgs)
		return n, true
	case *Ruleset:
		iv.VisitRuleset(n, visitArgs)
		return n, true
	case *Media:
		iv.VisitMedia(n, visitArgs)
		return n, true
	default:
		return node, false
	}
}

// VisitNodeOut implements direct dispatch for visitOut methods
func (iv *ImportVisitor) VisitNodeOut(node any) bool {
	switch n := node.(type) {
	case *Declaration:
		iv.VisitDeclarationOut(n)
		return true
	case *AtRule:
		iv.VisitAtRuleOut(n)
		return true
	case *MixinDefinition:
		iv.VisitMixinDefinitionOut(n)
		return true
	case *Ruleset:
		iv.VisitRulesetOut(n)
		return true
	case *Media:
		iv.VisitMediaOut(n)
		return true
	default:
		return false
	}
}

// Run processes the root node
func (iv *ImportVisitor) Run(root any) {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG ImportVisitor.Run] Called - processing imports\n")
	}
	defer func() {
		iv.isFinished = true
		iv.sequencer.TryRun()
	}()

	// Handle panics and convert them to errors
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				iv.error = err
			} else {
				// Convert panic to LessError
				iv.error = NewLessError(ErrorDetails{
					Message: fmt.Sprintf("%v", r),
					Index:   0,
				}, nil, "")
			}
		}
	}()

	iv.visitor.Visit(root)
}

// onSequencerEmpty is called when the sequencer has no more work
func (iv *ImportVisitor) onSequencerEmpty() {
	if !iv.isFinished {
		return
	}
	iv.finish(iv.error)
}

// VisitImport handles Import nodes - matches JavaScript visitImport
func (iv *ImportVisitor) VisitImport(importNode any, visitArgs *VisitArgs) {
	inlineCSS := false
	css := false

	// Handle Import struct directly
	if imp, ok := importNode.(*Import); ok {
		// Get inline option from Import struct
		if imp.options != nil {
			if inline, hasInline := imp.options["inline"].(bool); hasInline {
				inlineCSS = inline
			}
		}
		// Get css property from Import struct
		css = imp.css
	} else if node, ok := importNode.(map[string]any); ok {
		// Handle legacy map-based Import nodes
		if options, hasOptions := node["options"].(map[string]any); hasOptions {
			if inline, hasInline := options["inline"].(bool); hasInline {
				inlineCSS = inline
			}
		}

		// Check if css field is explicitly set
		if cssValue, hasCss := node["css"].(bool); hasCss {
			css = cssValue
		} else {
			// If css field is not set, check if the path ends with .css
			// This matches the JavaScript Import constructor logic (import.js lines 33-36)
			path := iv.getPath(importNode)
			if path != "" {
				// Check for CSS file extension using the same regex as JavaScript
				if cssPatternRegex.MatchString(path) {
					css = true
					// Set the css field on the node for consistency
					node["css"] = true
				}
			}
		}
	}

	if !css || inlineCSS {
		// Create context with copied frames - matches JavaScript
		// Pass parent context properties (including importMultiple) to child
		frames := CopyArray(iv.context.Frames)
		context := NewEvalFromEval(iv.context, frames)

		var importParent any
		if len(context.Frames) > 0 {
			importParent = context.Frames[0]
		}

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			parentType := fmt.Sprintf("%T", importParent)
			fmt.Fprintf(os.Stderr, "[ImportVisitor.VisitImport] Processing import=%p, css=%v, inline=%v, parent type=%s\n", importNode, css, inlineCSS, parentType)
		}

		// Skip processing VARIABLE imports inside mixin definitions during the initial visitor pass.
		// These imports will be processed when the mixin is actually called/evaluated,
		// at which point variables will be available for variable imports.
		// Non-variable imports (like inline imports with known paths) should still be processed.
		_, isMixinDef := importParent.(*MixinDefinition)
		isVarImport := iv.isVariableImport(importNode)

		if isMixinDef && isVarImport {
			visitArgs.VisitDeeper = false
			return
		}

		iv.importCount++

		if isVarImport {
			iv.sequencer.AddVariableImport(func() {
				iv.processImportNode(importNode, context, importParent)
			})
		} else {
			iv.processImportNode(importNode, context, importParent)
		}
	}
	visitArgs.VisitDeeper = false
}

// processImportNode processes an individual import node - matches JavaScript
func (iv *ImportVisitor) processImportNode(importNode any, context *Eval, importParent any) {
	var evaldImportNode any
	inlineCSS := false

	// Get inline option - handle both Import structs and maps
	if imp, ok := importNode.(*Import); ok {
		if imp.options != nil {
			if inline, hasInline := imp.options["inline"].(bool); hasInline {
				inlineCSS = inline
			}
		}
	} else if node, ok := importNode.(map[string]any); ok {
		if options, hasOptions := node["options"].(map[string]any); hasOptions {
			if inline, hasInline := options["inline"].(bool); hasInline {
				inlineCSS = inline
			}
		}
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Fprintf(os.Stderr, "[ImportVisitor.processImportNode] importNode=%p, inlineCSS=%v\n", importNode, inlineCSS)
	}

	// For variable imports, update context frames to include variables from
	// imports processed since this import was deferred (variable hoisting)
	// NOTE: This partially fixes basic interpolation but forward references
	// (using variables defined in later imports) remain unsupported.
	// See IMPORT_INTERPOLATION_INVESTIGATION.md for details on why full fix
	// causes regressions in error detection tests.
	if iv.isVariableImport(importNode) {
		context.Frames = CopyArray(iv.context.Frames)
	}


	// Try to evaluate the import node - matches JavaScript try/catch
	func() {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = NewLessError(ErrorDetails{
						Message: fmt.Sprintf("%v", r),
						Index:   iv.getIndex(importNode),
					}, nil, iv.getFilename(importNode))
				}

				// Set error on import node and mark as CSS
				iv.setProperty(importNode, "css", true)
				iv.setProperty(importNode, "error", err)
			}
		}()

		evaldImportNode = iv.evalForImport(importNode, context)
	}()

	// Check if evaldImportNode is CSS - must handle both Import structs and maps
	isCSS := false
	cssUndefined := false
	if evaldImp, ok := evaldImportNode.(*Import); ok {
		// Handle Import struct directly
		isCSS = evaldImp.css
		// For Import structs, css is always defined (defaults to false)
		cssUndefined = false
	} else {
		// Handle map-based nodes
		evaldCSS := iv.getProperty(evaldImportNode, "css")
		if evaldCSS == nil {
			cssUndefined = true
			isCSS = false
		} else {
			cssUndefined = false
			isCSS = evaldCSS.(bool)
		}
	}

	if evaldImportNode != nil && (!isCSS || inlineCSS) {
		// Set context.importMultiple if multiple option is true
		if iv.getOptionBool(evaldImportNode, "multiple", false) {
			context.ImportMultiple = true
		}

		// Try appending less extension if CSS status is undefined
		tryAppendLessExtension := cssUndefined

		// Replace import node in parent rules - matches JavaScript
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ImportVisitor.processImportNode] Replacing importNode=%p with evaldImportNode=%p in parent\n", importNode, evaldImportNode)
		}
		iv.replaceRuleInParent(importParent, importNode, evaldImportNode)

		onImported := func(args ...any) {
			iv.onImported(evaldImportNode, context, args...)
		}
		sequencedOnImported := iv.sequencer.AddImport(onImported)

		// Call importer.push - matches JavaScript
		iv.callImporterPush(
			evaldImportNode,
			tryAppendLessExtension,
			sequencedOnImported,
		)
	} else {
		iv.importCount--
		if iv.isFinished {
			iv.sequencer.TryRun()
		}
	}
}

// onImported handles the result of an import operation - matches JavaScript
func (iv *ImportVisitor) onImported(importNode any, context *Eval, args ...any) {
	// Parse callback arguments
	var e error
	var root any
	var importedAtRoot bool
	var fullPath string

	if len(args) >= 1 && args[0] != nil {
		if err, ok := args[0].(error); ok {
			e = err
		}
	}
	if len(args) >= 2 {
		root = args[1]
	}
	if len(args) >= 3 {
		if iar, ok := args[2].(bool); ok {
			importedAtRoot = iar
		}
	}
	if len(args) >= 4 {
		if fp, ok := args[3].(string); ok {
			fullPath = fp
		}
	}

	if os.Getenv("LESS_GO_DEBUG") == "1" {
		inline := iv.getOptionBool(importNode, "inline", false)
		rootType := fmt.Sprintf("%T", root)
		fmt.Fprintf(os.Stderr, "[ImportVisitor.onImported] inline=%v, rootType=%s, fullPath=%q\n", inline, rootType, fullPath)
	}

	// Handle error - matches JavaScript
	if e != nil {
		if lessErr, ok := e.(*LessError); ok {
			if lessErr.Filename == "" {
				lessErr.Index = iv.getIndex(importNode)
				lessErr.Filename = iv.getFilename(importNode)
			}
		}
		iv.error = e
	}

	inlineCSS := iv.getOptionBool(importNode, "inline", false)
	isPlugin := iv.getOptionBool(importNode, "isPlugin", false)  
	isOptional := iv.getOptionBool(importNode, "optional", false)
	duplicateImport := importedAtRoot || iv.recursionDetector[fullPath]

	// Handle skip logic - matches JavaScript
	if !context.ImportMultiple {
		if duplicateImport {
			iv.setProperty(importNode, "skip", true)
		} else {
			// Set skip as function that checks onceFileDetectionMap
			iv.setProperty(importNode, "skip", func() bool {
				if iv.onceFileDetectionMap[fullPath] {
					return true
				}
				iv.onceFileDetectionMap[fullPath] = true
				return false
			})
		}
	}

	// Skip optional imports without fullPath
	if fullPath == "" && isOptional {
		iv.setProperty(importNode, "skip", true)
	}

	// Process root if provided - matches JavaScript
	if root != nil {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ImportVisitor.onImported] Setting root on importNode=%p type=%T\n", importNode, importNode)
		}
		iv.setProperty(importNode, "root", root)
		iv.setProperty(importNode, "importedFilename", fullPath)

		if !inlineCSS && !isPlugin && (context.ImportMultiple || !duplicateImport) {
			iv.recursionDetector[fullPath] = true

			// Save context - matches JavaScript
			oldContext := iv.context
			iv.context = context

			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						iv.error = err
					}
				}
			}()

			iv.visitor.Visit(root)

			// Add the imported file's root to the frames for variable hoisting
			// This allows variables from imported files to be available to subsequent imports
			// Only add if it's a ruleset and not already in frames
			if rootRuleset, ok := root.(*Ruleset); ok {
				// Check if this ruleset is already in the frames
				alreadyInFrames := false
				for _, frame := range oldContext.Frames {
					if frame == rootRuleset {
						alreadyInFrames = true
						break
					}
				}
				if !alreadyInFrames {
					// Add the imported file's ruleset to the frames
					oldContext.Frames = append(oldContext.Frames, rootRuleset)
				}
			}
			iv.context = oldContext
		}
	}

	iv.importCount--

	if iv.isFinished {
		iv.sequencer.TryRun()
	}
}

// Helper methods for property access - matches JavaScript direct property access

func (iv *ImportVisitor) getProperty(node any, prop string) any {
	if n, ok := node.(map[string]any); ok {
		return n[prop]
	} else if imp, ok := node.(*Import); ok {
		// Handle Import struct properties
		switch prop {
		case "root":
			return imp.root
		case "importedFilename":
			return imp.importedFilename
		case "skip":
			return imp.skip
		}
	}
	return nil
}

func (iv *ImportVisitor) setProperty(node any, prop string, value any) {
	if n, ok := node.(map[string]any); ok {
		n[prop] = value
	} else if imp, ok := node.(*Import); ok {
		// Handle Import struct properties
		switch prop {
		case "root":
			imp.root = value
		case "importedFilename":
			if filename, ok := value.(string); ok {
				imp.importedFilename = filename
			}
		case "skip":
			imp.skip = value
		}
	}
}

func (iv *ImportVisitor) getOptionBool(node any, option string, defaultValue bool) bool {
	if n, ok := node.(map[string]any); ok {
		if options, hasOptions := n["options"].(map[string]any); hasOptions {
			if val, hasVal := options[option].(bool); hasVal {
				return val
			}
		}
	} else if imp, ok := node.(*Import); ok {
		if imp.options != nil {
			if val, hasVal := imp.options[option].(bool); hasVal {
				return val
			}
		}
	}
	return defaultValue
}

func (iv *ImportVisitor) getIndex(node any) int {
	if n, ok := node.(map[string]any); ok {
		if idx, hasIdx := n["index"].(int); hasIdx {
			return idx
		}
	}
	return 0
}

func (iv *ImportVisitor) getFilename(node any) string {
	if n, ok := node.(map[string]any); ok {
		if fileInfo, hasFileInfo := n["fileInfo"].(map[string]any); hasFileInfo {
			if filename, hasFilename := fileInfo["filename"].(string); hasFilename {
				return filename
			}
		}
	}
	return ""
}

func (iv *ImportVisitor) isVariableImport(node any) bool {
	// Check if node is an Import struct
	if imp, ok := node.(*Import); ok {
		return imp.IsVariableImport()
	}

	// Call isVariableImport method if it exists (for map-based nodes)
	if n, ok := node.(map[string]any); ok {
		if method, hasMethod := n["isVariableImport"]; hasMethod {
			if fn, ok := method.(func() bool); ok {
				return fn()
			}
		}
	}
	return false
}

func (iv *ImportVisitor) evalForImport(node any, context *Eval) any {
	// Handle Import struct directly
	if imp, ok := node.(*Import); ok {
		return imp.EvalForImport(context)
	}

	// Call evalForImport method if it exists (legacy map-based nodes)
	if n, ok := node.(map[string]any); ok {
		if method, hasMethod := n["evalForImport"]; hasMethod {
			if fn, ok := method.(func(*Eval) any); ok {
				return fn(context)
			}
		}
	}
	return node
}

func (iv *ImportVisitor) replaceRuleInParent(parent any, oldRule any, newRule any) {
	if p, ok := parent.(map[string]any); ok {
		if rules, hasRules := p["rules"].([]any); hasRules {
			for i, rule := range rules {
				if rule == oldRule {
					rules[i] = newRule
					break
				}
			}
		}
	} else if ruleset, ok := parent.(*Ruleset); ok {
		// Handle *Ruleset parent
		if ruleset.Rules != nil {
			for i, rule := range ruleset.Rules {
				if rule == oldRule {
					ruleset.Rules[i] = newRule
					break
				}
			}
		}
	} else if mixin, ok := parent.(*MixinDefinition); ok {
		// Handle *MixinDefinition parent
		if mixin.Rules != nil {
			for i, rule := range mixin.Rules {
				if rule == oldRule {
					mixin.Rules[i] = newRule
					break
				}
			}
		}
	}
}

func (iv *ImportVisitor) callImporterPush(importNode any, tryAppendLessExtension bool, callback func(...any)) {
	// Handle ImportManager struct directly
	if importManager, ok := iv.importer.(*ImportManager); ok {
		path := iv.getPath(importNode)
		fileInfo := iv.getFileInfo(importNode)
		options := iv.getOptions(importNode)

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			inline := false
			if opts := iv.getOptions(importNode); opts != nil {
				if inl, ok := opts["inline"].(bool); ok {
					inline = inl
				}
			}
			currentDir := ""
			if fileInfo != nil {
				if cd, ok := fileInfo["currentDirectory"].(string); ok {
					currentDir = cd
				}
			}
			fmt.Fprintf(os.Stderr, "[ImportVisitor.callImporterPush] path=%q, inline=%v, currentDir=%q\n", path, inline, currentDir)
		}
		
		// Convert fileInfo map to FileInfo struct
		currentFileInfo := &FileInfo{}
		if fileInfo != nil {
			if cd, ok := fileInfo["currentDirectory"].(string); ok {
				currentFileInfo.CurrentDirectory = cd
			}
			if ep, ok := fileInfo["entryPath"].(string); ok {
				currentFileInfo.EntryPath = ep
			}
			if fn, ok := fileInfo["filename"].(string); ok {
				currentFileInfo.Filename = fn
			}
			if rp, ok := fileInfo["rootpath"].(string); ok {
				currentFileInfo.Rootpath = rp
			}
			if rfn, ok := fileInfo["rootFilename"].(string); ok {
				currentFileInfo.RootFilename = rfn
			}
			if ref, ok := fileInfo["reference"].(bool); ok {
				currentFileInfo.Reference = ref
			}
		}
		
		// Convert options map to ImportOptions struct
		importOptions := &ImportOptions{}
		if options != nil {
			if opt, ok := options["optional"].(bool); ok {
				importOptions.Optional = opt
			}
			if inline, ok := options["inline"].(bool); ok {
				importOptions.Inline = inline
			}
			if ref, ok := options["reference"].(bool); ok {
				importOptions.Reference = ref
			}
			if mult, ok := options["multiple"].(bool); ok {
				importOptions.Multiple = mult
			}
			if plugin, ok := options["isPlugin"].(bool); ok {
				importOptions.IsPlugin = plugin
			}
			// Handle pluginArgs - can be a string or map
			if args, ok := options["pluginArgs"].(map[string]any); ok {
				importOptions.PluginArgs = args
			} else if argsStr, ok := options["pluginArgs"].(string); ok && argsStr != "" {
				// Convert string args to the expected format
				// The string value is passed directly as the options to setOptions()
				importOptions.PluginArgs = map[string]any{"_args": argsStr}
			}
		}
		
		// Create a callback that matches the ImportManager.Push signature
		pushCallback := func(err error, root any, importedEqualsRoot bool, fullPath string) {
			if err != nil {
				callback(err, nil, false, "")
			} else {
				callback(nil, root, importedEqualsRoot, fullPath)
			}
		}
		
		importManager.Push(path, tryAppendLessExtension, currentFileInfo, importOptions, pushCallback)
		return
	}
	
	// Fallback: handle map-based importer (legacy)
	if imp, ok := iv.importer.(map[string]any); ok {
		if pushMethod, hasPush := imp["push"]; hasPush {
			if fn, ok := pushMethod.(func(string, bool, map[string]any, map[string]any, func(...any))); ok {
				path := iv.getPath(importNode)
				fileInfo := iv.getFileInfo(importNode)
				options := iv.getOptions(importNode)
				fn(path, tryAppendLessExtension, fileInfo, options, callback)
			}
		}
	}
}

func (iv *ImportVisitor) getPath(node any) string {
	// Handle Import struct directly
	if imp, ok := node.(*Import); ok {
		path := imp.GetPath()
		if pathStr, ok := path.(string); ok {
			return pathStr
		}
		// Handle Quoted path - use direct type assertion since we know it's *Quoted
		if quoted, ok := path.(*Quoted); ok {
			value := quoted.GetValue()
			return value
		}
		// Handle Anonymous path (inline imports can have Anonymous nodes)
		// Anonymous.Value can be a string or a Quoted object
		if anon, ok := path.(*Anonymous); ok {
			if str, ok := anon.Value.(string); ok {
				return str
			}
			if quoted, ok := anon.Value.(*Quoted); ok {
				return quoted.GetValue()
			}
		}
		return ""
	}

	// Fallback: handle map-based node (legacy)
	if n, ok := node.(map[string]any); ok {
		if method, hasMethod := n["getPath"]; hasMethod {
			if fn, ok := method.(func() string); ok {
				return fn()
			}
		}
	}
	return ""
}

func (iv *ImportVisitor) getFileInfo(node any) map[string]any {
	// Handle Import struct directly
	if imp, ok := node.(*Import); ok {
		// Create fileInfo map from Import struct fields
		fileInfo := make(map[string]any)
		if imp._fileInfo != nil {
			// Copy from the _fileInfo field
			for k, v := range imp._fileInfo {
				fileInfo[k] = v
			}
		}
		// Add current directory if not present (GetFileInfo method doesn't exist, use _fileInfo)
		if _, hasCD := fileInfo["currentDirectory"]; !hasCD && imp._fileInfo != nil {
			if cd, ok := imp._fileInfo["currentDirectory"]; ok {
				fileInfo["currentDirectory"] = cd
			}
		}
		return fileInfo
	}
	
	// Fallback: handle map-based node (legacy)
	if n, ok := node.(map[string]any); ok {
		if method, hasMethod := n["fileInfo"]; hasMethod {
			if fn, ok := method.(func() map[string]any); ok {
				return fn()
			}
		}
	}
	return nil
}

func (iv *ImportVisitor) getOptions(node any) map[string]any {
	// Handle Import struct directly
	if imp, ok := node.(*Import); ok {
		return imp.options
	}
	
	// Fallback: handle map-based node (legacy)
	if n, ok := node.(map[string]any); ok {
		if options, hasOptions := n["options"].(map[string]any); hasOptions {
			return options
		}
	}
	return nil
}

// Frame management methods - matches JavaScript prototype methods

func (iv *ImportVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) {
	if iv.isDetachedRuleset(declNode) {
		iv.context.Frames = append([]any{declNode}, iv.context.Frames...)
	} else if iv.declarationContainsDetachedRuleset(declNode) {
		// If the Declaration's value contains a DetachedRuleset, we need to visit deeper
		// to process any @plugin imports inside the DetachedRuleset body.
		// Don't set VisitDeeper = false, allowing the visitor to process children.
		iv.context.Frames = append([]any{declNode}, iv.context.Frames...)
	} else {
		visitArgs.VisitDeeper = false
	}
}

func (iv *ImportVisitor) VisitDeclarationOut(declNode any) {
	if (iv.isDetachedRuleset(declNode) || iv.declarationContainsDetachedRuleset(declNode)) && len(iv.context.Frames) > 0 {
		iv.context.Frames = iv.context.Frames[1:]
	}
}

func (iv *ImportVisitor) VisitAtRule(atRuleNode any, visitArgs *VisitArgs) {
	iv.context.Frames = append([]any{atRuleNode}, iv.context.Frames...)
}

func (iv *ImportVisitor) VisitAtRuleOut(atRuleNode any) {
	if len(iv.context.Frames) > 0 {
		iv.context.Frames = iv.context.Frames[1:]
	}
}

func (iv *ImportVisitor) VisitMixinDefinition(mixinDefinitionNode any, visitArgs *VisitArgs) {
	iv.context.Frames = append([]any{mixinDefinitionNode}, iv.context.Frames...)
}

func (iv *ImportVisitor) VisitMixinDefinitionOut(mixinDefinitionNode any) {
	if len(iv.context.Frames) > 0 {
		iv.context.Frames = iv.context.Frames[1:]
	}
}

func (iv *ImportVisitor) VisitRuleset(rulesetNode any, visitArgs *VisitArgs) {
	iv.context.Frames = append([]any{rulesetNode}, iv.context.Frames...)
}

func (iv *ImportVisitor) VisitRulesetOut(rulesetNode any) {
	if len(iv.context.Frames) > 0 {
		iv.context.Frames = iv.context.Frames[1:]
	}
}

func (iv *ImportVisitor) VisitMedia(mediaNode any, visitArgs *VisitArgs) {
	// Add mediaNode.rules[0] to frames - matches JavaScript behavior
	if n, ok := mediaNode.(map[string]any); ok {
		if rules, hasRules := n["rules"].([]any); hasRules && len(rules) > 0 {
			iv.context.Frames = append([]any{rules[0]}, iv.context.Frames...)
		}
	} else if media, ok := mediaNode.(interface{ GetRules() []any }); ok {
		// Handle Media nodes that implement GetRules interface
		rules := media.GetRules()
		if len(rules) > 0 {
			iv.context.Frames = append([]any{rules[0]}, iv.context.Frames...)
		}
	}
}

func (iv *ImportVisitor) VisitMediaOut(mediaNode any) {
	if len(iv.context.Frames) > 0 {
		iv.context.Frames = iv.context.Frames[1:]
	}
}

func (iv *ImportVisitor) isDetachedRuleset(node any) bool {
	if n, ok := node.(map[string]any); ok {
		if nodeType, hasType := n["type"].(string); hasType {
			return nodeType == "DetachedRuleset"
		}
	}
	return false
}

// declarationContainsDetachedRuleset checks if a Declaration node's value contains a DetachedRuleset.
// This is needed to ensure @plugin imports inside DetachedRulesets are processed by the ImportVisitor.
func (iv *ImportVisitor) declarationContainsDetachedRuleset(node any) bool {
	// Check for *Declaration type
	if decl, ok := node.(*Declaration); ok {
		if decl.Value != nil && len(decl.Value.Value) > 0 {
			for _, v := range decl.Value.Value {
				// Check if the value is a DetachedRuleset
				if _, ok := v.(*DetachedRuleset); ok {
					return true
				}
				// Also check for map representation
				if m, ok := v.(map[string]any); ok {
					if nodeType, hasType := m["type"].(string); hasType && nodeType == "DetachedRuleset" {
						return true
					}
				}
			}
		}
		return false
	}

	// Check for map representation of Declaration
	if n, ok := node.(map[string]any); ok {
		if nodeType, hasType := n["type"].(string); hasType && nodeType == "Declaration" {
			if value, hasValue := n["value"]; hasValue {
				// Check if value is a Value with DetachedRuleset
				if valueObj, ok := value.(*Value); ok && len(valueObj.Value) > 0 {
					for _, v := range valueObj.Value {
						if _, ok := v.(*DetachedRuleset); ok {
							return true
						}
						if m, ok := v.(map[string]any); ok {
							if nodeType, hasType := m["type"].(string); hasType && nodeType == "DetachedRuleset" {
								return true
							}
						}
					}
				}
				// Check if value is directly a DetachedRuleset
				if _, ok := value.(*DetachedRuleset); ok {
					return true
				}
				if valueMap, ok := value.(map[string]any); ok {
					if vType, hasType := valueMap["type"].(string); hasType && vType == "DetachedRuleset" {
						return true
					}
				}
			}
		}
	}
	return false
}