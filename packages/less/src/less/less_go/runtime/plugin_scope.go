package runtime

import (
	"sync"
)

// PluginScope represents a scope in the plugin hierarchy.
// It manages functions, visitors, and other plugin components that are
// scoped to a particular level in the LESS AST (e.g., file-level, ruleset-level, mixin-level).
//
// Plugin scoping follows these rules:
// - Global plugins (@plugin at file root) affect the entire file
// - Local plugins (@plugin inside rulesets) only affect that scope and children
// - Child scopes can shadow parent functions (local overrides global)
// - Visitors from parent scopes are inherited
type PluginScope struct {
	parent         *PluginScope
	plugins        []*Plugin
	functions      map[string]*JSFunctionDefinition
	visitors       []*JSVisitor
	preProcessors  []ProcessorWithPriority
	postProcessors []ProcessorWithPriority
	fileManagers   []any
	mu             sync.RWMutex
}

// ProcessorWithPriority wraps a processor with its priority for ordering.
type ProcessorWithPriority struct {
	Processor any
	Priority  int
}

// NewPluginScope creates a new plugin scope with an optional parent.
// If parent is nil, this is a root (global) scope.
func NewPluginScope(parent *PluginScope) *PluginScope {
	return &PluginScope{
		parent:         parent,
		plugins:        make([]*Plugin, 0),
		functions:      make(map[string]*JSFunctionDefinition),
		visitors:       make([]*JSVisitor, 0),
		preProcessors:  make([]ProcessorWithPriority, 0),
		postProcessors: make([]ProcessorWithPriority, 0),
		fileManagers:   make([]any, 0),
	}
}

// NewRootPluginScope creates a root (global) plugin scope.
func NewRootPluginScope() *PluginScope {
	return NewPluginScope(nil)
}

// AddPlugin registers a plugin in this scope.
// It extracts the plugin's functions and visitors and registers them locally.
func (ps *PluginScope) AddPlugin(plugin *Plugin, runtime *NodeJSRuntime) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.plugins = append(ps.plugins, plugin)

	// Register plugin's functions in this scope
	for _, name := range plugin.Functions {
		ps.functions[name] = NewJSFunctionDefinition(name, runtime)
	}

	// Note: Visitors are typically registered through the PluginManager
	// and need the runtime to create JSVisitor instances
}

// AddFunction registers a function in this scope.
// This allows a local plugin to shadow a function from a parent scope.
func (ps *PluginScope) AddFunction(name string, fn *JSFunctionDefinition) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.functions[name] = fn
}

// AddVisitor registers a visitor in this scope.
func (ps *PluginScope) AddVisitor(visitor *JSVisitor) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.visitors = append(ps.visitors, visitor)
}

// AddPreProcessor adds a pre-processor to this scope with the given priority.
func (ps *PluginScope) AddPreProcessor(processor any, priority int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.preProcessors = append(ps.preProcessors, ProcessorWithPriority{
		Processor: processor,
		Priority:  priority,
	})
}

// AddPostProcessor adds a post-processor to this scope with the given priority.
func (ps *PluginScope) AddPostProcessor(processor any, priority int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.postProcessors = append(ps.postProcessors, ProcessorWithPriority{
		Processor: processor,
		Priority:  priority,
	})
}

// AddFileManager adds a file manager to this scope.
func (ps *PluginScope) AddFileManager(manager any) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.fileManagers = append(ps.fileManagers, manager)
}

// LookupFunction looks up a function by name, searching from this scope up to the root.
// Local functions shadow parent functions with the same name.
// Returns the function and true if found, nil and false otherwise.
func (ps *PluginScope) LookupFunction(name string) (*JSFunctionDefinition, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Check local scope first
	if fn, ok := ps.functions[name]; ok {
		return fn, true
	}

	// Check parent scopes
	if ps.parent != nil {
		return ps.parent.LookupFunction(name)
	}

	return nil, false
}

// GetLocalFunction returns a function only if it's defined in this scope (not inherited).
func (ps *PluginScope) GetLocalFunction(name string) (*JSFunctionDefinition, bool) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	fn, ok := ps.functions[name]
	return fn, ok
}

// GetAllFunctions returns all functions visible from this scope (including inherited).
func (ps *PluginScope) GetAllFunctions() map[string]*JSFunctionDefinition {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make(map[string]*JSFunctionDefinition)

	// First collect parent functions (if any)
	if ps.parent != nil {
		parentFuncs := ps.parent.GetAllFunctions()
		for name, fn := range parentFuncs {
			result[name] = fn
		}
	}

	// Then add local functions (overriding parent if same name)
	for name, fn := range ps.functions {
		result[name] = fn
	}

	return result
}

// GetVisitors returns all visitors from this scope and all parent scopes.
// Parent visitors are included because they should still run on child content.
func (ps *PluginScope) GetVisitors() []*JSVisitor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// Collect visitors from this scope
	visitors := make([]*JSVisitor, len(ps.visitors))
	copy(visitors, ps.visitors)

	// Add parent visitors
	if ps.parent != nil {
		visitors = append(visitors, ps.parent.GetVisitors()...)
	}

	return visitors
}

// GetLocalVisitors returns only visitors defined in this scope (not inherited).
func (ps *PluginScope) GetLocalVisitors() []*JSVisitor {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	result := make([]*JSVisitor, len(ps.visitors))
	copy(result, ps.visitors)
	return result
}

// GetPreEvalVisitors returns all pre-evaluation visitors from this scope and parents.
func (ps *PluginScope) GetPreEvalVisitors() []*JSVisitor {
	result := make([]*JSVisitor, 0)
	for _, v := range ps.GetVisitors() {
		if v.IsPreEvalVisitor {
			result = append(result, v)
		}
	}
	return result
}

// GetPostEvalVisitors returns all post-evaluation visitors from this scope and parents.
func (ps *PluginScope) GetPostEvalVisitors() []*JSVisitor {
	result := make([]*JSVisitor, 0)
	for _, v := range ps.GetVisitors() {
		if !v.IsPreEvalVisitor {
			result = append(result, v)
		}
	}
	return result
}

// GetPreProcessors returns all pre-processors from this scope and parents,
// sorted by priority.
func (ps *PluginScope) GetPreProcessors() []any {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var all []ProcessorWithPriority

	// Collect from parent first
	if ps.parent != nil {
		parentProcessors := ps.parent.getPreProcessorsWithPriority()
		all = append(all, parentProcessors...)
	}

	// Add local processors
	all = append(all, ps.preProcessors...)

	// Sort by priority (lower priority runs first)
	sortProcessorsByPriority(all)

	// Extract just the processors
	result := make([]any, len(all))
	for i, p := range all {
		result[i] = p.Processor
	}
	return result
}

// getPreProcessorsWithPriority returns pre-processors with their priority info.
func (ps *PluginScope) getPreProcessorsWithPriority() []ProcessorWithPriority {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]ProcessorWithPriority, len(ps.preProcessors))
	copy(result, ps.preProcessors)
	return result
}

// GetPostProcessors returns all post-processors from this scope and parents,
// sorted by priority.
func (ps *PluginScope) GetPostProcessors() []any {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var all []ProcessorWithPriority

	// Collect from parent first
	if ps.parent != nil {
		parentProcessors := ps.parent.getPostProcessorsWithPriority()
		all = append(all, parentProcessors...)
	}

	// Add local processors
	all = append(all, ps.postProcessors...)

	// Sort by priority
	sortProcessorsByPriority(all)

	// Extract just the processors
	result := make([]any, len(all))
	for i, p := range all {
		result[i] = p.Processor
	}
	return result
}

// getPostProcessorsWithPriority returns post-processors with their priority info.
func (ps *PluginScope) getPostProcessorsWithPriority() []ProcessorWithPriority {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]ProcessorWithPriority, len(ps.postProcessors))
	copy(result, ps.postProcessors)
	return result
}

// GetFileManagers returns all file managers from this scope and parents.
func (ps *PluginScope) GetFileManagers() []any {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]any, len(ps.fileManagers))
	copy(result, ps.fileManagers)

	if ps.parent != nil {
		result = append(result, ps.parent.GetFileManagers()...)
	}

	return result
}

// GetPlugins returns all plugins registered in this scope.
func (ps *PluginScope) GetPlugins() []*Plugin {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	result := make([]*Plugin, len(ps.plugins))
	copy(result, ps.plugins)
	return result
}

// Parent returns the parent scope, or nil if this is the root.
func (ps *PluginScope) Parent() *PluginScope {
	return ps.parent
}

// IsRoot returns true if this is a root (global) scope.
func (ps *PluginScope) IsRoot() bool {
	return ps.parent == nil
}

// CreateChild creates a new child scope with this scope as its parent.
// This is used when entering a new scoping boundary (e.g., a ruleset with @plugin).
func (ps *PluginScope) CreateChild() *PluginScope {
	return NewPluginScope(ps)
}

// sortProcessorsByPriority sorts processors by priority (lower first).
// This is a simple insertion sort suitable for small lists.
func sortProcessorsByPriority(processors []ProcessorWithPriority) {
	for i := 1; i < len(processors); i++ {
		key := processors[i]
		j := i - 1
		for j >= 0 && processors[j].Priority > key.Priority {
			processors[j+1] = processors[j]
			j--
		}
		processors[j+1] = key
	}
}

// ScopedPluginManager wraps a PluginScope and provides the interface expected
// by the less_go PluginManager. This allows PluginScope to be used in the
// existing transform_tree.go visitor loop.
type ScopedPluginManager struct {
	scope    *PluginScope
	runtime  *NodeJSRuntime
	iterator int
}

// NewScopedPluginManager creates a new scoped plugin manager.
func NewScopedPluginManager(scope *PluginScope, runtime *NodeJSRuntime) *ScopedPluginManager {
	return &ScopedPluginManager{
		scope:    scope,
		runtime:  runtime,
		iterator: -1,
	}
}

// Visitor returns a visitor iterator for transform_tree.go compatibility.
func (spm *ScopedPluginManager) Visitor() *ScopedVisitorIterator {
	return &ScopedVisitorIterator{
		manager:  spm,
		visitors: spm.scope.GetVisitors(),
	}
}

// GetVisitors returns all visitors from the scope.
func (spm *ScopedPluginManager) GetVisitors() []any {
	jsVisitors := spm.scope.GetVisitors()
	result := make([]any, len(jsVisitors))
	for i, v := range jsVisitors {
		result[i] = v
	}
	return result
}

// GetPreProcessors returns all pre-processors from the scope.
func (spm *ScopedPluginManager) GetPreProcessors() []any {
	return spm.scope.GetPreProcessors()
}

// GetPostProcessors returns all post-processors from the scope.
func (spm *ScopedPluginManager) GetPostProcessors() []any {
	return spm.scope.GetPostProcessors()
}

// GetFileManagers returns all file managers from the scope.
func (spm *ScopedPluginManager) GetFileManagers() []any {
	return spm.scope.GetFileManagers()
}

// ScopedVisitorIterator provides visitor iteration for transform_tree.go.
type ScopedVisitorIterator struct {
	manager  *ScopedPluginManager
	visitors []*JSVisitor
}

// First resets the iterator to the beginning.
func (vi *ScopedVisitorIterator) First() {
	vi.manager.iterator = -1
}

// Get returns the next visitor in the iteration.
func (vi *ScopedVisitorIterator) Get() any {
	vi.manager.iterator++
	if vi.manager.iterator >= len(vi.visitors) {
		return nil
	}
	return vi.visitors[vi.manager.iterator]
}
