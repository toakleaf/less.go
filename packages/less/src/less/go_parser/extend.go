package go_parser

import (
	"fmt"
)

// Extend represents an extend node in the Less AST.
// It corresponds to the JS Extend class in tree/extend.js.
type Extend struct {
    *Node
    Selector any
    Option string
    ObjectId int
    ParentIds []int
    AllowRoot bool
    AllowBefore bool
    AllowAfter bool
    SelfSelectors []*Selector
}

// extendNextID is a package-level counter for unique object IDs.
var extendNextID int = 0

// NewExtend creates a new Extend node with the given selector, option, index,
// file information, and visibility information.
func NewExtend(selector any, option string, index int, currentFileInfo map[string]any, visibilityInfo map[string]any) *Extend {
    e := &Extend{
        Node: NewNode(),
        Selector: selector,
        Option: option,
        ObjectId: extendNextID,
        ParentIds: []int{extendNextID},
        AllowRoot: true,
    }
    extendNextID++
    // Set index and file info
    e.Index = index
    if currentFileInfo != nil {
        e.SetFileInfo(currentFileInfo)
    }
    // Copy visibility info
    e.CopyVisibilityInfo(visibilityInfo)

    // Handle option flags
    switch option {
    case "!all", "all":
        e.AllowBefore = true
        e.AllowAfter = true
    default:
        e.AllowBefore = false
        e.AllowAfter = false
    }

    // Set parent relationship with selector, if it's a *Selector
    if sel, ok := selector.(*Selector); ok && sel != nil {
        e.SetParent(sel.Node, e.Node)
    }
    return e
}

// Type returns the node type for Extend.
func (e *Extend) Type() string {
    return "Extend"
}

// Accept calls the visitor on the selector.
// Panics if visitor is nil.
func (e *Extend) Accept(visitor Visitor) {
    if visitor == nil {
        panic("Extend.Accept: visitor is nil")
    }
    e.Selector = visitor.Visit(e.Selector)
}

// Eval evaluates the selector within the given context and returns a new Extend.
// It handles selector implementations returning (*Selector, error) or any, treating nil results as no selector.
func (e *Extend) Eval(context any) (*Extend, error) {
    // Default to the original selector
    var newSelector any = e.Selector
    // Handle selectors returning (*Selector, error)
    if evSel, ok := e.Selector.(interface{ Eval(any) (*Selector, error) }); ok {
        selResult, err := evSel.Eval(context)
        if err != nil {
            return nil, err
        }
        if selResult == nil {
            newSelector = nil
        } else {
            newSelector = selResult
        }
    } else if ev, ok := e.Selector.(interface{ Eval(any) any }); ok {
        newSelector = ev.Eval(context)
    }
    return NewExtend(newSelector, e.Option, e.GetIndex(), e.FileInfo(), e.VisibilityInfo()), nil
}

// Clone creates a copy of the Extend node.
// The context parameter is unused but present for compatibility with JS API.
func (e *Extend) Clone(context any) *Extend {
    return NewExtend(e.Selector, e.Option, e.GetIndex(), e.FileInfo(), e.VisibilityInfo())
}

// FindSelfSelectors concatenates the provided selectors into a single self selector.
// Panics on error creating the underlying Selector.
func (e *Extend) FindSelfSelectors(selectors []*Selector) {
    var selfElements []*Element
    for i, sel := range selectors {
        elements := sel.Elements
        if i > 0 && len(elements) > 0 && elements[0].Combinator.Value == "" {
            elements[0].Combinator.Value = " "
        }
        selfElements = append(selfElements, elements...)
    }
    newSel, err := NewSelector(selfElements, nil, nil, e.GetIndex(), e.FileInfo(), e.VisibilityInfo())
    if err != nil {
        panic(fmt.Sprintf("Extend.FindSelfSelectors: %v", err))
    }
    e.SelfSelectors = []*Selector{newSel}
    // Copy visibility info to the new self selector
    e.SelfSelectors[0].CopyVisibilityInfo(e.VisibilityInfo())
} 