package less_go

// MathType controls how mathematical expressions are evaluated in LESS.
// This affects when division and other operations are performed versus
// when they are left as literal CSS output.
type MathType int

const (
	// MathAlways evaluates all mathematical expressions.
	// Division is always performed, e.g., "1/2" becomes "0.5".
	// This was the default behavior in Less.js v1.x.
	MathAlways MathType = iota

	// MathParensDivision only performs division inside parentheses.
	// Other math operations (addition, subtraction, multiplication) are always performed.
	// This is the default behavior and matches Less.js v3.x+ defaults.
	// Example: "1/2" stays as "1/2", but "(1/2)" becomes "0.5".
	MathParensDivision

	// MathParens only performs math operations inside parentheses.
	// All operations require parentheses to be evaluated.
	// Example: "1 + 2" stays as "1 + 2", but "(1 + 2)" becomes "3".
	MathParens
)

// RewriteUrlsType controls how URLs in the compiled CSS are rewritten
// relative to the output file or entry file.
type RewriteUrlsType int

const (
	// RewriteUrlsOff disables URL rewriting. URLs are output as-is.
	RewriteUrlsOff RewriteUrlsType = iota

	// RewriteUrlsLocal rewrites only relative URLs (not starting with /).
	// URLs are adjusted relative to the entry LESS file.
	RewriteUrlsLocal

	// RewriteUrlsAll rewrites all URLs (both relative and absolute paths).
	// URLs are adjusted relative to the entry LESS file.
	RewriteUrlsAll
)

// Math provides named constants for math mode options.
// Use these values when setting CompileOptions.Math.
//
// Example:
//
//	options := &CompileOptions{
//	    Math: Math.ParensDivision,
//	}
var Math = struct{ Always, ParensDivision, Parens MathType }{MathAlways, MathParensDivision, MathParens}

// RewriteUrls provides named constants for URL rewriting options.
// Use these values when setting CompileOptions.RewriteUrls.
//
// Example:
//
//	options := &CompileOptions{
//	    RewriteUrls: RewriteUrls.Local,
//	}
var RewriteUrls = struct{ Off, Local, All RewriteUrlsType }{RewriteUrlsOff, RewriteUrlsLocal, RewriteUrlsAll} 