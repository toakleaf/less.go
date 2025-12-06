package less_go

import "regexp"

// Pre-compiled regular expressions for parser performance
// These regexes are compiled once at package initialization instead of on every parse call.
// This provides a 5-10x performance improvement by eliminating repeated regex compilation overhead.

// Directive patterns
var (
	rePluginDirective       = regexp.MustCompile(`^@plugin?\s+`)
	rePluginDirectiveStrict = regexp.MustCompile(`^@plugin\s+`)
	reImportDirective       = regexp.MustCompile(`^@import\s+`)
)

// Variable patterns
var (
	reVariableAt           = regexp.MustCompile(`^@@?[\w-]+`)
	reVariableAtName       = regexp.MustCompile(`^(@[\w-]+)\s*:`)
	reVariableAtNameSimple = regexp.MustCompile(`^@[a-z-]+`)
	reVariableAtCurly      = regexp.MustCompile(`^@\{([\w-]+)\}`)
	reVariableDollar       = regexp.MustCompile(`^\$[\w-]+`)
	reVariableDollarCurly  = regexp.MustCompile(`^\$\{([\w-]+)\}`)
	reVariableCall         = regexp.MustCompile(`^(@[\w-]+)(\(\s*\))?`)
)

// Property and name patterns
var (
	rePropertyName       = regexp.MustCompile(`^([_a-zA-Z0-9-]+)\s*:`)
	rePropertyNameFull   = regexp.MustCompile(`^(\*?-?[_a-zA-Z0-9-]+)\s*:`)
	reIdentifier         = regexp.MustCompile(`^[_A-Za-z-][_A-Za-z0-9-]+`)
	reKeyword            = regexp.MustCompile(`^\w+`)
	reNamePrefix         = regexp.MustCompile(`^(?:[@$]{0,2})[_a-zA-Z0-9-]*`)
	reEntityName         = regexp.MustCompile(`^\[?(?:[\w-]|\\(?:[A-Fa-f0-9]{1,6} ?|[^A-Fa-f0-9]))+\]?`)
)

// Color patterns
var (
	// Matches hex colors and captures optional trailing character to detect invalid formats
	// Valid: #RGB, #RGBA, #RRGGBB, #RRGGBBAA
	// Invalid: #RRRR, #RRRRR (detected via trailing character in capture group 2)
	reColorHex = regexp.MustCompile(`^#([A-Fa-f0-9]{8}|[A-Fa-f0-9]{6}|[A-Fa-f0-9]{3,4})([\w.#\[])?`)
)

// Number and dimension patterns
var (
	reDimension       = regexp.MustCompile(`^([+-]?\d*\.?\d+)(%|[a-zA-Z_]+)?`)
	rePercentage      = regexp.MustCompile(`^(?:\d+\.\d+|\d+)%`)
	reNumber          = regexp.MustCompile(`^\d+`)
	rePercentSimple   = regexp.MustCompile(`^[0-9]+%`)
	reUnicodeRange    = regexp.MustCompile(`^U\+[0-9a-fA-F?]+(-[0-9a-fA-F?]+)?`)
)

// Selector patterns
var (
	reSelector          = regexp.MustCompile(`^[#.](?:[\w-]|\\(?:[A-Fa-f0-9]{1,6} ?|[^A-Fa-f0-9]))+`)
	reSelectorMixin     = regexp.MustCompile(`^([#.](?:[\w-]|\\(?:[A-Fa-f0-9]{1,6} ?|[^A-Fa-f0-9]))+)\s*\(`)
	reSelectorElement   = regexp.MustCompile(`^(?:[.#]?|:*)(?:[\w-]|@\{[\w-]+\}|[^\x00-\x9f]|\\(?:[A-Fa-f0-9]{1,6} ?|[^A-Fa-f0-9]))+`)
	reSelectorParens    = regexp.MustCompile(`^\([^&()@]+\)`)
	reSelectorPrefixes  = regexp.MustCompile(`^[.#:]`)
	reSelectorAttribute = regexp.MustCompile(`^(?:[_A-Za-z0-9-*]*\|)?(?:[_A-Za-z0-9-]|\\.)+`)
	reSelectorNamespace = regexp.MustCompile(`^(\*?)`)
	reSelectorNamePart  = regexp.MustCompile(`^((?:[\w-]+)|(?:[@$]\{[\w-]+\}))`)
)

// Operator patterns
var (
	reOperatorSpaced  = regexp.MustCompile(`^[-+]\s+`)
	reOperatorCompare = regexp.MustCompile(`^[0-9a-z-]*\s*([<>]=|<=|>=|[<>]|=)`)
	reOperatorAttr    = regexp.MustCompile(`^[|~*$^]?=`)
)

// Spacing detection patterns
var (
	// Matches identifier followed by whitespace then opening paren - used to detect spacing
	// before parentheses in media features (e.g., "and (" vs "layer(")
	reSpacingBeforeParen = regexp.MustCompile(`^[0-9a-z-]*\s+\(`)
)

// Function and call patterns
var (
	reFunctionURL  = regexp.MustCompile(`(?i)^url\(`)
	reFunctionName = regexp.MustCompile(`^([\w-]+|%|~|progid:[\w.]+)\(`)
	reCallValid    = regexp.MustCompile(`^[\w]+\(`)
	reURLContent   = regexp.MustCompile(`^(?:(?:\\[()'""])|[^()'""])+`)
)

// Keyword patterns
var (
	reImportant     = regexp.MustCompile(`^! *important`)
	reMediaAll      = regexp.MustCompile(`^(!?all)`)
	reImportOptions = regexp.MustCompile(`^(less|css|multiple|once|inline|reference|optional)`)
	reCaseFlag      = regexp.MustCompile(`^[iIsS]`)
	reWordIdent     = regexp.MustCompile(`^[\w-]+`)
)

// Combinator patterns
var (
	reCombinatorSlashed = regexp.MustCompile(`^\/[a-z]+\/`)
)

// Comment patterns
var (
	reCommentStart = regexp.MustCompile(`^\/[*/]`)
)

// Whitespace patterns
var (
	reWhitespace = regexp.MustCompile(`^\s*`)
)

// Mixin patterns
var (
	reMixinCall      = regexp.MustCompile(`^[.#]\(`)
	reMixinNamespace = regexp.MustCompile(`^[^{]*\}`)
)

// Guard patterns
var (
	reGuardCondition = regexp.MustCompile(`^,\s*(not\s*)?\(`)
)

// Special patterns
var (
	reCloseParen        = regexp.MustCompile(`^\)`)
	reNegativeLookup    = regexp.MustCompile(`^-[@$(]`)
	reAnonymous         = regexp.MustCompile(`^([^.#@$+/'"*` + "`" + `(;{}-]*);`)
	reJavaScript        = regexp.MustCompile("^[^`]*`")
	reAlphaOpacity      = regexp.MustCompile(`^opacity=`)
	rePluginArgs        = regexp.MustCompile(`^\s*([^);]+)\)\s*`)
	reSelectorCombinator = regexp.MustCompile(`^((?:\+_|\+)?)\s*:`)
)

// Terminator patterns (used in permissive parsing)
var (
	reTermSemicolon = regexp.MustCompile(`[;}]`)
	reTermBrace     = regexp.MustCompile(`^[{;]`)
)

// URL-specific patterns
var (
	reURLEscapeChars = regexp.MustCompile(`[()'"\s]`)
	reDataURI        = regexp.MustCompile(`^\s*data:`)
)

// File manager patterns
var (
	reFileExtension   = regexp.MustCompile(`(\.[a-z]*$)|([?;].*)$`)
	reAbsolutePath    = regexp.MustCompile(`(?i)^(?:[a-z-]+:|\/|\\|#)`)
	reURLParts        = regexp.MustCompile(`(?i)^((?:[a-z-]+:)?\/{2}(?:[^/?#]*\/)|([/\\]))?((?:[^/\\?#]*[/\\])*)([^/\\?#]*)([#?].*)?$`)
)

// String and misc patterns
var (
	reStringFormat    = regexp.MustCompile(`%[sdaSDA]`)
	reBase64Suffix    = regexp.MustCompile(`;base64$`)
	reVariableAtBrace = regexp.MustCompile(`@\{([\w-]+)\}`)
	reFilenameOnly    = regexp.MustCompile(`[^/\\]*$`)
)
