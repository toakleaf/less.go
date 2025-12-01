package less_go

type ImportItem struct {
	callback func(...any)
	args     []any
	isReady  bool
}

type ImportSequencer struct {
	imports          []*ImportItem
	variableImports  []func()
	onSequencerEmpty func()
	currentDepth     int
}

func NewImportSequencer(onSequencerEmpty func()) *ImportSequencer {
	return &ImportSequencer{
		imports:          make([]*ImportItem, 0),
		variableImports:  make([]func(), 0),
		onSequencerEmpty: onSequencerEmpty,
		currentDepth:     0,
	}
}

func (is *ImportSequencer) AddImport(callback func(...any)) func(...any) {
	importItem := &ImportItem{
		callback: callback,
		args:     nil,
		isReady:  false,
	}
	is.imports = append(is.imports, importItem)

	return func(args ...any) {
		importItem.args = args
		importItem.isReady = true
		is.TryRun()
	}
}

func (is *ImportSequencer) AddVariableImport(callback func()) {
	is.variableImports = append(is.variableImports, callback)
}

func (is *ImportSequencer) TryRun() {
	is.currentDepth++
	defer func() {
		is.currentDepth--
		if is.currentDepth == 0 && is.onSequencerEmpty != nil {
			is.onSequencerEmpty()
		}
	}()

	for {
		for len(is.imports) > 0 {
			importItem := is.imports[0]
			if !importItem.isReady {
				return
			}
			is.imports = is.imports[1:]
			importItem.callback(importItem.args...)
		}

		if len(is.variableImports) == 0 {
			break
		}

		variableImport := is.variableImports[0]
		is.variableImports = is.variableImports[1:]
		variableImport()
	}
}