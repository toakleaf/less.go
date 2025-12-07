package less_go

import (
	"fmt"
	"os"
)

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

func (iv *ImportVisitor) IsReplacing() bool {
	return iv.isReplacing
}

func (iv *ImportVisitor) VisitNode(node any, visitArgs *VisitArgs) (any, bool) {
	switch n := node.(type) {
	case *AtRule:
		iv.VisitAtRule(n, visitArgs)
		return n, true
	case *Declaration:
		iv.VisitDeclaration(n, visitArgs)
		return n, true
	case *Import:
		iv.VisitImport(n, visitArgs)
		return n, true
	case *Media:
		iv.VisitMedia(n, visitArgs)
		return n, true
	case *MixinDefinition:
		iv.VisitMixinDefinition(n, visitArgs)
		return n, true
	case *Ruleset:
		iv.VisitRuleset(n, visitArgs)
		return n, true
	default:
		_ = n
		return node, true // Node type handled (no-op, avoids reflection)
	}
}

func (iv *ImportVisitor) VisitNodeOut(node any) bool {
	switch n := node.(type) {
	case *AtRule:
		iv.VisitAtRuleOut(n)
		return true
	case *Declaration:
		iv.VisitDeclarationOut(n)
		return true
	case *Media:
		iv.VisitMediaOut(n)
		return true
	case *MixinDefinition:
		iv.VisitMixinDefinitionOut(n)
		return true
	case *Ruleset:
		iv.VisitRulesetOut(n)
		return true
	default:
		_ = n
		return true // Node type handled (no-op, avoids reflection)
	}
}

func (iv *ImportVisitor) Run(root any) {
	if os.Getenv("LESS_GO_DEBUG") == "1" {
		fmt.Printf("[DEBUG ImportVisitor.Run] Called - processing imports\n")
	}
	defer func() {
		iv.isFinished = true
		iv.sequencer.TryRun()
	}()

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				iv.error = err
			} else {
				iv.error = NewLessError(ErrorDetails{
					Message: fmt.Sprintf("%v", r),
					Index:   0,
				}, nil, "")
			}
		}
	}()

	iv.visitor.Visit(root)
}

func (iv *ImportVisitor) onSequencerEmpty() {
	if !iv.isFinished {
		return
	}
	iv.finish(iv.error)
}

func (iv *ImportVisitor) VisitImport(importNode any, visitArgs *VisitArgs) {
	inlineCSS := false
	css := false

	if imp, ok := importNode.(*Import); ok {
		if imp.options != nil {
			if inline, hasInline := imp.options["inline"].(bool); hasInline {
				inlineCSS = inline
			}
		}
		css = imp.css
	} else if node, ok := importNode.(map[string]any); ok {
		if options, hasOptions := node["options"].(map[string]any); hasOptions {
			if inline, hasInline := options["inline"].(bool); hasInline {
				inlineCSS = inline
			}
		}

		if cssValue, hasCss := node["css"].(bool); hasCss {
			css = cssValue
		} else {
			path := iv.getPath(importNode)
			if path != "" {
				if cssPatternRegex.MatchString(path) {
					css = true
					node["css"] = true
				}
			}
		}
	}

	if !css || inlineCSS {
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

		// Skip variable imports inside mixin definitions - they're processed when mixin is called
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

func (iv *ImportVisitor) processImportNode(importNode any, context *Eval, importParent any) {
	var evaldImportNode any
	inlineCSS := false

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

	// NOTE: We intentionally do NOT overwrite context.Frames here.
	// For variable imports, the context passed in already has the correct frames
	// from when the import was visited in VisitImport(). Variable imports are
	// deferred via the sequencer and processed after the AST visit completes,
	// at which point iv.context.Frames is empty. Overwriting context.Frames
	// with iv.context.Frames would lose the variable definitions needed for
	// interpolation (e.g., @import "@{prefix}-@{suffix}.less" would fail to
	// resolve @prefix and @suffix).

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

	isCSS := false
	cssUndefined := false
	if evaldImp, ok := evaldImportNode.(*Import); ok {
		isCSS = evaldImp.css
		cssUndefined = false
	} else {
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
		if iv.getOptionBool(evaldImportNode, "multiple", false) {
			context.ImportMultiple = true
		}

		tryAppendLessExtension := cssUndefined

		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ImportVisitor.processImportNode] Replacing importNode=%p with evaldImportNode=%p in parent\n", importNode, evaldImportNode)
		}
		iv.replaceRuleInParent(importParent, importNode, evaldImportNode)

		onImported := func(args ...any) {
			iv.onImported(evaldImportNode, context, args...)
		}
		sequencedOnImported := iv.sequencer.AddImport(onImported)

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

func (iv *ImportVisitor) onImported(importNode any, context *Eval, args ...any) {
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

	if !context.ImportMultiple {
		if duplicateImport {
			iv.setProperty(importNode, "skip", true)
		} else {
			iv.setProperty(importNode, "skip", func() bool {
				if iv.onceFileDetectionMap[fullPath] {
					return true
				}
				iv.onceFileDetectionMap[fullPath] = true
				return false
			})
		}
	}

	if fullPath == "" && isOptional {
		iv.setProperty(importNode, "skip", true)
	}

	if root != nil {
		if os.Getenv("LESS_GO_DEBUG") == "1" {
			fmt.Fprintf(os.Stderr, "[ImportVisitor.onImported] Setting root on importNode=%p type=%T\n", importNode, importNode)
		}
		iv.setProperty(importNode, "root", root)
		iv.setProperty(importNode, "importedFilename", fullPath)

		if !inlineCSS && !isPlugin && (context.ImportMultiple || !duplicateImport) {
			iv.recursionDetector[fullPath] = true

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

			// Add imported file's root to frames for variable hoisting
			if rootRuleset, ok := root.(*Ruleset); ok {
				alreadyInFrames := false
				for _, frame := range oldContext.Frames {
					if frame == rootRuleset {
						alreadyInFrames = true
						break
					}
				}
				if !alreadyInFrames {
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

func (iv *ImportVisitor) getProperty(node any, prop string) any {
	if n, ok := node.(map[string]any); ok {
		return n[prop]
	} else if imp, ok := node.(*Import); ok {
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
	if imp, ok := node.(*Import); ok {
		return imp.IsVariableImport()
	}

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
	if imp, ok := node.(*Import); ok {
		return imp.EvalForImport(context)
	}

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
		if ruleset.Rules != nil {
			for i, rule := range ruleset.Rules {
				if rule == oldRule {
					ruleset.Rules[i] = newRule
					break
				}
			}
		}
	} else if mixin, ok := parent.(*MixinDefinition); ok {
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
			if args, ok := options["pluginArgs"].(map[string]any); ok {
				importOptions.PluginArgs = args
			} else if argsStr, ok := options["pluginArgs"].(string); ok && argsStr != "" {
				importOptions.PluginArgs = map[string]any{"_args": argsStr}
			}
		}

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
	if imp, ok := node.(*Import); ok {
		path := imp.GetPath()
		if pathStr, ok := path.(string); ok {
			return pathStr
		}
		if quoted, ok := path.(*Quoted); ok {
			return quoted.GetValue()
		}
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
	if imp, ok := node.(*Import); ok {
		fileInfo := make(map[string]any)
		if imp._fileInfo != nil {
			for k, v := range imp._fileInfo {
				fileInfo[k] = v
			}
		}
		if _, hasCD := fileInfo["currentDirectory"]; !hasCD && imp._fileInfo != nil {
			if cd, ok := imp._fileInfo["currentDirectory"]; ok {
				fileInfo["currentDirectory"] = cd
			}
		}
		return fileInfo
	}

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
	if imp, ok := node.(*Import); ok {
		return imp.options
	}

	if n, ok := node.(map[string]any); ok {
		if options, hasOptions := n["options"].(map[string]any); hasOptions {
			return options
		}
	}
	return nil
}

func (iv *ImportVisitor) VisitDeclaration(declNode any, visitArgs *VisitArgs) {
	if iv.isDetachedRuleset(declNode) {
		iv.context.Frames = append([]any{declNode}, iv.context.Frames...)
	} else if iv.declarationContainsDetachedRuleset(declNode) {
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
	if n, ok := mediaNode.(map[string]any); ok {
		if rules, hasRules := n["rules"].([]any); hasRules && len(rules) > 0 {
			iv.context.Frames = append([]any{rules[0]}, iv.context.Frames...)
		}
	} else if media, ok := mediaNode.(interface{ GetRules() []any }); ok {
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

func (iv *ImportVisitor) declarationContainsDetachedRuleset(node any) bool {
	if decl, ok := node.(*Declaration); ok {
		if decl.Value != nil && len(decl.Value.Value) > 0 {
			for _, v := range decl.Value.Value {
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
		return false
	}

	if n, ok := node.(map[string]any); ok {
		if nodeType, hasType := n["type"].(string); hasType && nodeType == "Declaration" {
			if value, hasValue := n["value"]; hasValue {
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