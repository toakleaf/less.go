package less_go

import "sync"

// stringInterner provides a thread-safe string interning mechanism.
// Interning reduces memory usage by ensuring that identical strings
// share the same underlying memory, and can improve comparison speed
// since interned strings can often be compared by pointer.
type stringInterner struct {
	mu      sync.RWMutex
	strings map[string]string
}

// defaultInterner is the global string interner instance.
// Pre-allocated with capacity for common CSS strings.
var defaultInterner = &stringInterner{
	strings: make(map[string]string, 1024),
}

// Intern returns a canonical version of the string.
// If the string has been interned before, the previously interned
// version is returned. This ensures that all occurrences of the
// same string value share the same memory.
//
// This function is thread-safe and uses a read-write lock for
// optimal concurrent read performance.
func Intern(s string) string {
	// Don't intern empty strings or very long strings
	// Long strings are unlikely to be repeated and waste memory in the intern table
	if len(s) == 0 || len(s) > 128 {
		return s
	}

	// Fast path: check if already interned (read lock only)
	defaultInterner.mu.RLock()
	if existing, ok := defaultInterner.strings[s]; ok {
		defaultInterner.mu.RUnlock()
		return existing
	}
	defaultInterner.mu.RUnlock()

	// Slow path: add to intern table (write lock required)
	defaultInterner.mu.Lock()
	defer defaultInterner.mu.Unlock()

	// Double-check after acquiring write lock to handle race conditions
	if existing, ok := defaultInterner.strings[s]; ok {
		return existing
	}

	// Store the string in the intern table
	defaultInterner.strings[s] = s
	return s
}

// InternBytes interns a string created from a byte slice.
// This is useful when parsing to avoid creating multiple string
// copies from the same byte slice content.
func InternBytes(b []byte) string {
	if len(b) == 0 || len(b) > 128 {
		return string(b)
	}
	return Intern(string(b))
}

// InternedCount returns the number of interned strings.
// Useful for debugging and monitoring memory usage.
func InternedCount() int {
	defaultInterner.mu.RLock()
	defer defaultInterner.mu.RUnlock()
	return len(defaultInterner.strings)
}

// ClearInternTable clears the intern table.
// This should only be used in testing or when you want to
// release all interned strings.
func ClearInternTable() {
	defaultInterner.mu.Lock()
	defer defaultInterner.mu.Unlock()
	defaultInterner.strings = make(map[string]string, 1024)
	// Re-intern common strings after clearing
	internCommonStrings()
}

// internCommonStrings pre-interns common CSS property names,
// pseudo-classes, and other frequently used strings.
func internCommonStrings() {
	// Common CSS property names
	properties := []string{
		// Layout
		"display", "position", "top", "right", "bottom", "left",
		"float", "clear", "z-index", "overflow", "overflow-x", "overflow-y",
		"visibility", "clip", "clip-path",

		// Flexbox
		"flex", "flex-grow", "flex-shrink", "flex-basis", "flex-direction",
		"flex-wrap", "flex-flow", "justify-content", "align-items",
		"align-content", "align-self", "order",

		// Grid
		"grid", "grid-template", "grid-template-columns", "grid-template-rows",
		"grid-template-areas", "grid-gap", "grid-row-gap", "grid-column-gap",
		"grid-auto-columns", "grid-auto-rows", "grid-auto-flow",
		"grid-column", "grid-row", "gap", "row-gap", "column-gap",

		// Box model
		"width", "height", "min-width", "max-width", "min-height", "max-height",
		"margin", "margin-top", "margin-right", "margin-bottom", "margin-left",
		"padding", "padding-top", "padding-right", "padding-bottom", "padding-left",
		"border", "border-width", "border-style", "border-color",
		"border-top", "border-right", "border-bottom", "border-left",
		"border-top-width", "border-right-width", "border-bottom-width", "border-left-width",
		"border-top-style", "border-right-style", "border-bottom-style", "border-left-style",
		"border-top-color", "border-right-color", "border-bottom-color", "border-left-color",
		"border-radius", "border-top-left-radius", "border-top-right-radius",
		"border-bottom-left-radius", "border-bottom-right-radius",
		"box-sizing", "box-shadow", "outline", "outline-width", "outline-style", "outline-color",

		// Typography
		"font", "font-family", "font-size", "font-style", "font-weight",
		"font-variant", "font-stretch", "line-height", "letter-spacing",
		"word-spacing", "text-align", "text-decoration", "text-transform",
		"text-indent", "text-shadow", "white-space", "word-break", "word-wrap",
		"text-overflow", "vertical-align",

		// Colors and backgrounds
		"color", "background", "background-color", "background-image",
		"background-repeat", "background-position", "background-size",
		"background-attachment", "background-origin", "background-clip",
		"opacity", "filter",

		// Transforms and animations
		"transform", "transform-origin", "transform-style",
		"perspective", "perspective-origin",
		"transition", "transition-property", "transition-duration",
		"transition-timing-function", "transition-delay",
		"animation", "animation-name", "animation-duration",
		"animation-timing-function", "animation-delay", "animation-iteration-count",
		"animation-direction", "animation-fill-mode", "animation-play-state",

		// Lists
		"list-style", "list-style-type", "list-style-position", "list-style-image",

		// Tables
		"table-layout", "border-collapse", "border-spacing",
		"caption-side", "empty-cells",

		// UI
		"cursor", "pointer-events", "user-select", "resize",

		// Misc
		"content", "quotes", "counter-reset", "counter-increment",
		"will-change", "appearance", "object-fit", "object-position",
	}

	// Common pseudo-classes and pseudo-elements
	pseudos := []string{
		":hover", ":active", ":focus", ":visited", ":link",
		":first-child", ":last-child", ":nth-child", ":nth-of-type",
		":first-of-type", ":last-of-type", ":only-child", ":only-of-type",
		":empty", ":not", ":checked", ":disabled", ":enabled",
		":before", ":after", "::before", "::after",
		"::first-line", "::first-letter", "::selection", "::placeholder",
	}

	// Common element names
	elements := []string{
		"html", "body", "head", "title", "meta", "link", "style", "script",
		"div", "span", "p", "a", "img", "br", "hr",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "li", "dl", "dt", "dd",
		"table", "tr", "td", "th", "thead", "tbody", "tfoot", "caption",
		"form", "input", "button", "select", "option", "textarea", "label",
		"fieldset", "legend",
		"header", "footer", "nav", "main", "section", "article", "aside",
		"figure", "figcaption", "details", "summary",
		"audio", "video", "canvas", "svg", "iframe",
	}

	// Common units and values
	units := []string{
		"px", "em", "rem", "%", "vh", "vw", "vmin", "vmax",
		"pt", "pc", "in", "cm", "mm", "ex", "ch",
		"deg", "rad", "grad", "turn",
		"s", "ms",
		"auto", "none", "inherit", "initial", "unset",
		"block", "inline", "inline-block", "flex", "inline-flex", "grid", "inline-grid",
		"absolute", "relative", "fixed", "sticky", "static",
		"hidden", "visible", "scroll",
		"solid", "dashed", "dotted", "double", "groove", "ridge", "inset", "outset",
		"normal", "bold", "bolder", "lighter",
		"center", "left", "right", "top", "bottom",
		"transparent", "currentColor",
		"!important",
	}

	// Common LESS-specific strings
	lessStrings := []string{
		"@import", "@media", "@keyframes", "@font-face", "@supports",
		"@charset", "@namespace", "@page", "@viewport",
		"when", "and", "not", "or",
		"true", "false", "null",
		"rgb", "rgba", "hsl", "hsla", "url", "calc", "var",
	}

	// Intern all common strings
	for _, s := range properties {
		defaultInterner.strings[s] = s
	}
	for _, s := range pseudos {
		defaultInterner.strings[s] = s
	}
	for _, s := range elements {
		defaultInterner.strings[s] = s
	}
	for _, s := range units {
		defaultInterner.strings[s] = s
	}
	for _, s := range lessStrings {
		defaultInterner.strings[s] = s
	}
}

// init pre-interns common CSS strings at startup
func init() {
	internCommonStrings()
}
