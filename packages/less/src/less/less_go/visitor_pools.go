package less_go

import (
	"sync"
)

// Visitor pools for reusing visitor instances across compilations.
// The methodLookup map in each Visitor is built via reflection and is expensive.
// By pooling visitors, we avoid rebuilding these maps on every compilation.

// joinSelectorVisitorPool pools JoinSelectorVisitor instances
var joinSelectorVisitorPool = sync.Pool{
	New: func() any {
		jsv := &JoinSelectorVisitor{
			contexts: make([]*contextInfo, 0, 8),
		}
		jsv.visitor = NewVisitor(jsv)
		return jsv
	},
}

// GetJoinSelectorVisitor retrieves a JoinSelectorVisitor from the pool
func GetJoinSelectorVisitor() *JoinSelectorVisitor {
	v := joinSelectorVisitorPool.Get().(*JoinSelectorVisitor)
	v.Reset()
	return v
}

// ReleaseJoinSelectorVisitor returns a JoinSelectorVisitor to the pool
func ReleaseJoinSelectorVisitor(v *JoinSelectorVisitor) {
	if v == nil {
		return
	}
	joinSelectorVisitorPool.Put(v)
}

// extendFinderVisitorPool pools ExtendFinderVisitor instances
var extendFinderVisitorPool = sync.Pool{
	New: func() any {
		efv := &ExtendFinderVisitor{
			contexts:        make([]any, 0, 16),
			allExtendsStack: make([][]any, 0, 8),
		}
		efv.visitor = NewVisitor(efv)
		return efv
	},
}

// GetExtendFinderVisitor retrieves an ExtendFinderVisitor from the pool
func GetExtendFinderVisitor() *ExtendFinderVisitor {
	v := extendFinderVisitorPool.Get().(*ExtendFinderVisitor)
	v.Reset()
	return v
}

// ReleaseExtendFinderVisitor returns an ExtendFinderVisitor to the pool
func ReleaseExtendFinderVisitor(v *ExtendFinderVisitor) {
	if v == nil {
		return
	}
	extendFinderVisitorPool.Put(v)
}

// processExtendsVisitorPool pools ProcessExtendsVisitor instances
var processExtendsVisitorPool = sync.Pool{
	New: func() any {
		pev := &ProcessExtendsVisitor{
			extendIndices:    make(map[string]bool),
			allExtendsStack:  make([][]*Extend, 0, 8),
			mediaAtRuleStack: make([]any, 0, 8),
		}
		pev.visitor = NewVisitor(pev)
		return pev
	},
}

// GetProcessExtendsVisitor retrieves a ProcessExtendsVisitor from the pool
func GetProcessExtendsVisitor() *ProcessExtendsVisitor {
	v := processExtendsVisitorPool.Get().(*ProcessExtendsVisitor)
	v.Reset()
	return v
}

// ReleaseProcessExtendsVisitor returns a ProcessExtendsVisitor to the pool
func ReleaseProcessExtendsVisitor(v *ProcessExtendsVisitor) {
	if v == nil {
		return
	}
	processExtendsVisitorPool.Put(v)
}

// toCSSVisitorPool pools ToCSSVisitor instances
var toCSSVisitorPool = sync.Pool{
	New: func() any {
		v := &ToCSSVisitor{
			charset:     false,
			isReplacing: true,
		}
		v.utils = NewCSSVisitorUtils(nil)
		v.visitor = NewVisitor(v)
		return v
	},
}

// GetToCSSVisitor retrieves a ToCSSVisitor from the pool and configures it with the given context
func GetToCSSVisitor(context any) *ToCSSVisitor {
	v := toCSSVisitorPool.Get().(*ToCSSVisitor)
	v.Reset(context)
	return v
}

// ReleaseToCSSVisitor returns a ToCSSVisitor to the pool
func ReleaseToCSSVisitor(v *ToCSSVisitor) {
	if v == nil {
		return
	}
	toCSSVisitorPool.Put(v)
}

// setTreeVisibilityVisitorPool pools SetTreeVisibilityVisitor instances
var setTreeVisibilityVisitorPool = sync.Pool{
	New: func() any {
		return &SetTreeVisibilityVisitor{}
	},
}

// GetSetTreeVisibilityVisitor retrieves a SetTreeVisibilityVisitor from the pool
func GetSetTreeVisibilityVisitor(visible any) *SetTreeVisibilityVisitor {
	v := setTreeVisibilityVisitorPool.Get().(*SetTreeVisibilityVisitor)
	v.Reset(visible)
	return v
}

// ReleaseSetTreeVisibilityVisitor returns a SetTreeVisibilityVisitor to the pool
func ReleaseSetTreeVisibilityVisitor(v *SetTreeVisibilityVisitor) {
	if v == nil {
		return
	}
	setTreeVisibilityVisitorPool.Put(v)
}

// cssVisitorUtilsPool pools CSSVisitorUtils instances
var cssVisitorUtilsPool = sync.Pool{
	New: func() any {
		utils := &CSSVisitorUtils{}
		utils.visitor = NewVisitor(utils)
		return utils
	},
}

// GetCSSVisitorUtils retrieves a CSSVisitorUtils from the pool
func GetCSSVisitorUtils(context any) *CSSVisitorUtils {
	u := cssVisitorUtilsPool.Get().(*CSSVisitorUtils)
	u.Reset(context)
	return u
}

// ReleaseCSSVisitorUtils returns a CSSVisitorUtils to the pool
func ReleaseCSSVisitorUtils(u *CSSVisitorUtils) {
	if u == nil {
		return
	}
	cssVisitorUtilsPool.Put(u)
}
