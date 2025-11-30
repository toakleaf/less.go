package less_go

// This file contains typed struct replacements for map[string]any to eliminate
// reflection overhead. These structs provide type-safe access to commonly used
// data structures in the Less compiler.

// NodeFileInfo contains information about the source file for a node.
// This replaces map[string]any{"filename": ..., "rootpath": ..., ...}
// which was causing reflection overhead in field access.
type NodeFileInfo struct {
	Filename         string
	Rootpath         string
	CurrentDirectory string
	RootFilename     string
	EntryPath        string
	Reference        bool
	RewriteUrls      bool
}

// NewNodeFileInfo creates a NodeFileInfo with optional initialization from a map.
// This provides backward compatibility during migration.
func NewNodeFileInfo() *NodeFileInfo {
	return &NodeFileInfo{}
}

// NewNodeFileInfoFromMap creates a NodeFileInfo from a map[string]any for backward compatibility.
func NewNodeFileInfoFromMap(m map[string]any) *NodeFileInfo {
	if m == nil {
		return nil
	}
	fi := &NodeFileInfo{}
	if v, ok := m["filename"].(string); ok {
		fi.Filename = v
	}
	if v, ok := m["rootpath"].(string); ok {
		fi.Rootpath = v
	}
	if v, ok := m["currentDirectory"].(string); ok {
		fi.CurrentDirectory = v
	}
	if v, ok := m["rootFilename"].(string); ok {
		fi.RootFilename = v
	}
	if v, ok := m["entryPath"].(string); ok {
		fi.EntryPath = v
	}
	if v, ok := m["reference"].(bool); ok {
		fi.Reference = v
	}
	if v, ok := m["rewriteUrls"].(bool); ok {
		fi.RewriteUrls = v
	}
	return fi
}

// ToMap converts NodeFileInfo back to map[string]any for backward compatibility.
func (fi *NodeFileInfo) ToMap() map[string]any {
	if fi == nil {
		return nil
	}
	m := make(map[string]any)
	if fi.Filename != "" {
		m["filename"] = fi.Filename
	}
	if fi.Rootpath != "" {
		m["rootpath"] = fi.Rootpath
	}
	if fi.CurrentDirectory != "" {
		m["currentDirectory"] = fi.CurrentDirectory
	}
	if fi.RootFilename != "" {
		m["rootFilename"] = fi.RootFilename
	}
	if fi.EntryPath != "" {
		m["entryPath"] = fi.EntryPath
	}
	if fi.Reference {
		m["reference"] = fi.Reference
	}
	if fi.RewriteUrls {
		m["rewriteUrls"] = fi.RewriteUrls
	}
	return m
}

// Clone creates a copy of NodeFileInfo.
func (fi *NodeFileInfo) Clone() *NodeFileInfo {
	if fi == nil {
		return nil
	}
	return &NodeFileInfo{
		Filename:         fi.Filename,
		Rootpath:         fi.Rootpath,
		CurrentDirectory: fi.CurrentDirectory,
		RootFilename:     fi.RootFilename,
		EntryPath:        fi.EntryPath,
		Reference:        fi.Reference,
		RewriteUrls:      fi.RewriteUrls,
	}
}

// NodeVisibilityInfo contains visibility information for a node.
// This replaces map[string]any{"visibilityBlocks": ..., "nodeVisible": ...}
type NodeVisibilityInfo struct {
	VisibilityBlocks *int
	NodeVisible      *bool
}

// NewNodeVisibilityInfo creates a new NodeVisibilityInfo.
func NewNodeVisibilityInfo() *NodeVisibilityInfo {
	return &NodeVisibilityInfo{}
}

// NewNodeVisibilityInfoFromMap creates a NodeVisibilityInfo from a map[string]any.
func NewNodeVisibilityInfoFromMap(m map[string]any) *NodeVisibilityInfo {
	if m == nil {
		return nil
	}
	vi := &NodeVisibilityInfo{}
	if v, ok := m["visibilityBlocks"].(*int); ok {
		vi.VisibilityBlocks = v
	} else if v, ok := m["visibilityBlocks"].(int); ok {
		vi.VisibilityBlocks = &v
	}
	if v, ok := m["nodeVisible"].(*bool); ok {
		vi.NodeVisible = v
	} else if v, ok := m["nodeVisible"].(bool); ok {
		vi.NodeVisible = &v
	}
	return vi
}

// ToMap converts NodeVisibilityInfo back to map[string]any for backward compatibility.
func (vi *NodeVisibilityInfo) ToMap() map[string]any {
	if vi == nil {
		return nil
	}
	return map[string]any{
		"visibilityBlocks": vi.VisibilityBlocks,
		"nodeVisible":      vi.NodeVisible,
	}
}

// Clone creates a copy of NodeVisibilityInfo.
func (vi *NodeVisibilityInfo) Clone() *NodeVisibilityInfo {
	if vi == nil {
		return nil
	}
	result := &NodeVisibilityInfo{}
	if vi.VisibilityBlocks != nil {
		v := *vi.VisibilityBlocks
		result.VisibilityBlocks = &v
	}
	if vi.NodeVisible != nil {
		v := *vi.NodeVisible
		result.NodeVisible = &v
	}
	return result
}

// ImportantScopeEntry represents a single entry in the important scope stack.
// This replaces map[string]any{"important": ...}
type ImportantScopeEntry struct {
	Important string // Can be "" (not set), "!important", or other string values
}

// NewImportantScopeEntry creates a new ImportantScopeEntry.
func NewImportantScopeEntry() *ImportantScopeEntry {
	return &ImportantScopeEntry{}
}

// ToMap converts ImportantScopeEntry back to map[string]any for backward compatibility.
func (e *ImportantScopeEntry) ToMap() map[string]any {
	if e == nil {
		return nil
	}
	if e.Important != "" {
		return map[string]any{"important": e.Important}
	}
	return map[string]any{}
}

// VariableValue represents a variable lookup result.
// This replaces map[string]any{"value": ..., "important": ...}
type VariableValue struct {
	Value     any
	Important any // Can be bool or string
}

// NewVariableValue creates a new VariableValue.
func NewVariableValue(value any) *VariableValue {
	return &VariableValue{Value: value}
}

// NewVariableValueWithImportant creates a new VariableValue with important flag.
func NewVariableValueWithImportant(value any, important any) *VariableValue {
	return &VariableValue{Value: value, Important: important}
}

// ToMap converts VariableValue back to map[string]any for backward compatibility.
func (v *VariableValue) ToMap() map[string]any {
	if v == nil {
		return nil
	}
	m := map[string]any{"value": v.Value}
	if v.Important != nil {
		m["important"] = v.Important
	}
	return m
}

// MixinArg represents a mixin argument.
// This replaces map[string]any{"name": ..., "value": ..., "expand": ...}
type MixinArg struct {
	Name   string
	Value  any
	Expand bool
}

// NewMixinArg creates a new MixinArg.
func NewMixinArg(name string, value any, expand bool) *MixinArg {
	return &MixinArg{Name: name, Value: value, Expand: expand}
}

// ToMap converts MixinArg back to map[string]any for backward compatibility.
func (a *MixinArg) ToMap() map[string]any {
	if a == nil {
		return nil
	}
	m := map[string]any{}
	if a.Name != "" {
		m["name"] = a.Name
	}
	if a.Value != nil {
		m["value"] = a.Value
	}
	if a.Expand {
		m["expand"] = a.Expand
	}
	return m
}

// CSSContext represents the context used during CSS generation.
// This replaces map[string]any{"compress": ..., "tabLevel": ..., ...}
type CSSContext struct {
	Compress     bool
	TabLevel     int
	FirstSelector bool
	LastSelector  bool
}

// NewCSSContext creates a new CSSContext.
func NewCSSContext() *CSSContext {
	return &CSSContext{}
}

// NewCSSContextFromMap creates a CSSContext from a map[string]any.
func NewCSSContextFromMap(m map[string]any) *CSSContext {
	if m == nil {
		return nil
	}
	ctx := &CSSContext{}
	if v, ok := m["compress"].(bool); ok {
		ctx.Compress = v
	}
	if v, ok := m["tabLevel"].(int); ok {
		ctx.TabLevel = v
	}
	if v, ok := m["firstSelector"].(bool); ok {
		ctx.FirstSelector = v
	}
	if v, ok := m["lastSelector"].(bool); ok {
		ctx.LastSelector = v
	}
	return ctx
}

// ToMap converts CSSContext back to map[string]any for backward compatibility.
func (c *CSSContext) ToMap() map[string]any {
	if c == nil {
		return nil
	}
	return map[string]any{
		"compress":      c.Compress,
		"tabLevel":      c.TabLevel,
		"firstSelector": c.FirstSelector,
		"lastSelector":  c.LastSelector,
	}
}

// Helper functions for type conversion during migration

// FileInfoToMap converts either *NodeFileInfo or map[string]any to map[string]any.
func FileInfoToMap(fi any) map[string]any {
	if fi == nil {
		return nil
	}
	if m, ok := fi.(map[string]any); ok {
		return m
	}
	if nfi, ok := fi.(*NodeFileInfo); ok {
		return nfi.ToMap()
	}
	return nil
}

// VisibilityInfoToMap converts either *NodeVisibilityInfo or map[string]any to map[string]any.
func VisibilityInfoToMap(vi any) map[string]any {
	if vi == nil {
		return nil
	}
	if m, ok := vi.(map[string]any); ok {
		return m
	}
	if nvi, ok := vi.(*NodeVisibilityInfo); ok {
		return nvi.ToMap()
	}
	return nil
}

// GetFilename extracts filename from either *NodeFileInfo or map[string]any.
func GetFilename(fi any) string {
	if fi == nil {
		return ""
	}
	if nfi, ok := fi.(*NodeFileInfo); ok {
		return nfi.Filename
	}
	if m, ok := fi.(map[string]any); ok {
		if v, ok := m["filename"].(string); ok {
			return v
		}
	}
	return ""
}

// GetRootpath extracts rootpath from either *NodeFileInfo or map[string]any.
func GetRootpath(fi any) string {
	if fi == nil {
		return ""
	}
	if nfi, ok := fi.(*NodeFileInfo); ok {
		return nfi.Rootpath
	}
	if m, ok := fi.(map[string]any); ok {
		if v, ok := m["rootpath"].(string); ok {
			return v
		}
	}
	return ""
}
