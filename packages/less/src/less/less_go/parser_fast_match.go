package less_go

// Fast string matching functions to replace regex operations
// These are optimized alternatives to regexp.MustCompile patterns

// isWordChar returns true if c is a word character (\w) - [a-zA-Z0-9_]
func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// isWordOrHyphenChar returns true if c is [\w-]
func isWordOrHyphenChar(c byte) bool {
	return isWordChar(c) || c == '-'
}

// isDigit returns true if c is [0-9]
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isHexChar returns true if c is [0-9A-Fa-f]
func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

// isWhitespaceChar returns true if c is \s (space, tab, newline, carriage return)
func isWhitespaceChar(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// isLowerAlpha returns true if c is [a-z]
func isLowerAlpha(c byte) bool {
	return c >= 'a' && c <= 'z'
}

// matchKeyword matches ^\w+ and returns the matched string or ""
// Replaces: reKeyword = regexp.MustCompile(`^\w+`)
func matchKeyword(s string) string {
	if len(s) == 0 || !isWordChar(s[0]) {
		return ""
	}
	i := 1
	for i < len(s) && isWordChar(s[i]) {
		i++
	}
	return s[:i]
}

// matchWhitespace matches ^\s* and returns the length of the match
// Replaces: reWhitespace = regexp.MustCompile(`^\s*`)
func matchWhitespace(s string) int {
	i := 0
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return i
}

// matchNumber matches ^\d+ and returns the matched string or ""
// Replaces: reNumber = regexp.MustCompile(`^\d+`)
func matchNumber(s string) string {
	if len(s) == 0 || !isDigit(s[0]) {
		return ""
	}
	i := 1
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	return s[:i]
}

// matchCaseFlag matches ^[iIsS] and returns the matched character or ""
// Replaces: reCaseFlag = regexp.MustCompile(`^[iIsS]`)
func matchCaseFlag(s string) string {
	if len(s) == 0 {
		return ""
	}
	c := s[0]
	if c == 'i' || c == 'I' || c == 's' || c == 'S' {
		return s[:1]
	}
	return ""
}

// matchWordIdent matches ^[\w-]+ and returns the matched string or ""
// Replaces: reWordIdent = regexp.MustCompile(`^[\w-]+`)
func matchWordIdent(s string) string {
	if len(s) == 0 || !isWordOrHyphenChar(s[0]) {
		return ""
	}
	i := 1
	for i < len(s) && isWordOrHyphenChar(s[i]) {
		i++
	}
	return s[:i]
}

// matchImportant matches ^! *important and returns the matched string or ""
// Replaces: reImportant = regexp.MustCompile(`^! *important`)
func matchImportant(s string) string {
	if len(s) == 0 || s[0] != '!' {
		return ""
	}
	i := 1
	// Skip spaces (not all whitespace, just spaces)
	for i < len(s) && s[i] == ' ' {
		i++
	}
	// Check for "important"
	if i+9 <= len(s) && s[i:i+9] == "important" {
		return s[:i+9]
	}
	return ""
}

// matchAlphaOpacity matches ^opacity= and returns the matched string or ""
// Replaces: reAlphaOpacity = regexp.MustCompile(`^opacity=`)
func matchAlphaOpacity(s string) string {
	if len(s) >= 8 && s[:8] == "opacity=" {
		return "opacity="
	}
	return ""
}

// matchCommentStart matches ^\/[*/] (detects /* or //)
// Returns the matched length (0, 2)
// Replaces: reCommentStart = regexp.MustCompile(`^\/[*/]`)
func matchCommentStart(s string) bool {
	if len(s) >= 2 && s[0] == '/' && (s[1] == '*' || s[1] == '/') {
		return true
	}
	return false
}

// matchCloseParen matches ^\)
// Replaces: reCloseParen = regexp.MustCompile(`^\)`)
func matchCloseParen(s string) bool {
	return len(s) > 0 && s[0] == ')'
}

// matchMixinCall matches ^[.#]\( and returns the matched string or ""
// Replaces: reMixinCall = regexp.MustCompile(`^[.#]\(`)
func matchMixinCall(s string) string {
	if len(s) >= 2 && (s[0] == '.' || s[0] == '#') && s[1] == '(' {
		return s[:2]
	}
	return ""
}

// matchSelectorPrefixes matches ^[.#:] and returns the matched string or ""
// Replaces: reSelectorPrefixes = regexp.MustCompile(`^[.#:]`)
func matchSelectorPrefixes(s string) string {
	if len(s) > 0 && (s[0] == '.' || s[0] == '#' || s[0] == ':') {
		return s[:1]
	}
	return ""
}

// matchPluginDirective matches ^@plugin?\s+ and returns the matched string or ""
// Replaces: rePluginDirective = regexp.MustCompile(`^@plugin?\s+`)
func matchPluginDirective(s string) string {
	// Must start with @plugi
	if len(s) < 7 || s[0] != '@' || s[1] != 'p' || s[2] != 'l' || s[3] != 'u' || s[4] != 'g' || s[5] != 'i' {
		return ""
	}
	i := 6
	// Optional 'n'
	if i < len(s) && s[i] == 'n' {
		i++
	}
	// Must have at least one whitespace
	if i >= len(s) || !isWhitespaceChar(s[i]) {
		return ""
	}
	// Skip all whitespace
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return s[:i]
}

// matchPluginDirectiveStrict matches ^@plugin\s+ and returns the matched string or ""
// Replaces: rePluginDirectiveStrict = regexp.MustCompile(`^@plugin\s+`)
func matchPluginDirectiveStrict(s string) string {
	if len(s) < 8 || s[:7] != "@plugin" {
		return ""
	}
	i := 7
	// Must have at least one whitespace
	if i >= len(s) || !isWhitespaceChar(s[i]) {
		return ""
	}
	// Skip all whitespace
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return s[:i]
}

// matchImportDirective matches ^@import\s+ and returns the matched string or ""
// Replaces: reImportDirective = regexp.MustCompile(`^@import\s+`)
func matchImportDirective(s string) string {
	if len(s) < 8 || s[:7] != "@import" {
		return ""
	}
	i := 7
	// Must have at least one whitespace
	if i >= len(s) || !isWhitespaceChar(s[i]) {
		return ""
	}
	// Skip all whitespace
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return s[:i]
}

// matchVariableAt matches ^@@?[\w-]+ and returns []string{fullMatch} or nil
// Replaces: reVariableAt = regexp.MustCompile(`^@@?[\w-]+`)
func matchVariableAt(s string) []string {
	if len(s) == 0 || s[0] != '@' {
		return nil
	}
	i := 1
	// Optional second @
	if i < len(s) && s[i] == '@' {
		i++
	}
	// Must have at least one [\w-]
	if i >= len(s) || !isWordOrHyphenChar(s[i]) {
		return nil
	}
	// Match rest of [\w-]+
	for i < len(s) && isWordOrHyphenChar(s[i]) {
		i++
	}
	return []string{s[:i]}
}

// matchVariableDollar matches ^\$[\w-]+ and returns []string{fullMatch} or nil
// Replaces: reVariableDollar = regexp.MustCompile(`^\$[\w-]+`)
func matchVariableDollar(s string) []string {
	if len(s) < 2 || s[0] != '$' {
		return nil
	}
	i := 1
	// Must have at least one [\w-]
	if !isWordOrHyphenChar(s[i]) {
		return nil
	}
	// Match rest of [\w-]+
	for i < len(s) && isWordOrHyphenChar(s[i]) {
		i++
	}
	return []string{s[:i]}
}

// matchIdentifier matches ^[_A-Za-z-][_A-Za-z0-9-]+ and returns the matched string or ""
// Replaces: reIdentifier = regexp.MustCompile(`^[_A-Za-z-][_A-Za-z0-9-]+`)
func matchIdentifier(s string) string {
	if len(s) == 0 {
		return ""
	}
	c := s[0]
	// First char must be [_A-Za-z-]
	if c != '_' && c != '-' && !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
		return ""
	}
	// Must have at least one more character
	if len(s) < 2 {
		return ""
	}
	i := 1
	// Match rest of [_A-Za-z0-9-]+
	for i < len(s) {
		c = s[i]
		if c != '_' && c != '-' && !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			break
		}
		i++
	}
	// Need at least 2 chars total
	if i < 2 {
		return ""
	}
	return s[:i]
}

// matchPercentage matches ^(?:\d+\.\d+|\d+)% and returns the matched string or ""
// Replaces: rePercentage = regexp.MustCompile(`^(?:\d+\.\d+|\d+)%`)
func matchPercentage(s string) string {
	if len(s) == 0 || !isDigit(s[0]) {
		return ""
	}
	i := 1
	// Match digits
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	// Check for decimal part
	if i < len(s) && s[i] == '.' {
		if i+1 < len(s) && isDigit(s[i+1]) {
			i++ // skip '.'
			for i < len(s) && isDigit(s[i]) {
				i++
			}
		}
	}
	// Must end with %
	if i >= len(s) || s[i] != '%' {
		return ""
	}
	return s[:i+1]
}

// matchPercentSimple matches ^[0-9]+% and returns the matched string or ""
// Replaces: rePercentSimple = regexp.MustCompile(`^[0-9]+%`)
func matchPercentSimple(s string) string {
	if len(s) == 0 || !isDigit(s[0]) {
		return ""
	}
	i := 1
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	if i >= len(s) || s[i] != '%' {
		return ""
	}
	return s[:i+1]
}

// matchMediaAll matches ^(!?all) and returns []string{fullMatch, group1} or nil
// Replaces: reMediaAll = regexp.MustCompile(`^(!?all)`)
func matchMediaAll(s string) []string {
	if len(s) >= 4 && s[:4] == "!all" {
		return []string{"!all", "!all"}
	}
	if len(s) >= 3 && s[:3] == "all" {
		return []string{"all", "all"}
	}
	return nil
}

// matchImportOptions matches ^(less|css|multiple|once|inline|reference|optional)
// Returns []string{fullMatch, group1} or nil
// Replaces: reImportOptions = regexp.MustCompile(`^(less|css|multiple|once|inline|reference|optional)`)
func matchImportOptions(s string) []string {
	options := []string{"reference", "optional", "multiple", "inline", "less", "once", "css"}
	for _, opt := range options {
		if len(s) >= len(opt) && s[:len(opt)] == opt {
			return []string{opt, opt}
		}
	}
	return nil
}

// matchNegativeLookup matches ^-[@$(] and returns the matched string or ""
// Replaces: reNegativeLookup = regexp.MustCompile(`^-[@$(]`)
func matchNegativeLookup(s string) string {
	if len(s) >= 2 && s[0] == '-' && (s[1] == '@' || s[1] == '$' || s[1] == '(') {
		return s[:2]
	}
	return ""
}

// matchOperatorAttr matches ^[|~*$^]?= and returns the matched string or ""
// Replaces: reOperatorAttr = regexp.MustCompile(`^[|~*$^]?=`)
func matchOperatorAttr(s string) string {
	if len(s) == 0 {
		return ""
	}
	if s[0] == '=' {
		return "="
	}
	if len(s) >= 2 && s[1] == '=' {
		c := s[0]
		if c == '|' || c == '~' || c == '*' || c == '$' || c == '^' {
			return s[:2]
		}
	}
	return ""
}

// matchTermBrace matches ^[{;] and returns true if matched
// Replaces: reTermBrace = regexp.MustCompile(`^[{;]`)
func matchTermBrace(s string) bool {
	return len(s) > 0 && (s[0] == '{' || s[0] == ';')
}

// matchDataURI matches ^\s*data: and returns true if matched
// Replaces: reDataURI = regexp.MustCompile(`^\s*data:`)
func matchDataURI(s string) bool {
	i := 0
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return i+5 <= len(s) && s[i:i+5] == "data:"
}

// matchFunctionURL matches (?i)^url\( (case-insensitive) and returns the matched string or ""
// Replaces: reFunctionURL = regexp.MustCompile(`(?i)^url\(`)
func matchFunctionURL(s string) string {
	if len(s) < 4 {
		return ""
	}
	// Case-insensitive check for "url("
	if (s[0] == 'u' || s[0] == 'U') &&
		(s[1] == 'r' || s[1] == 'R') &&
		(s[2] == 'l' || s[2] == 'L') &&
		s[3] == '(' {
		return s[:4]
	}
	return ""
}

// matchVariableAtNameSimple matches ^@[a-z-]+ and returns []string{fullMatch} or nil
// Replaces: reVariableAtNameSimple = regexp.MustCompile(`^@[a-z-]+`)
func matchVariableAtNameSimple(s string) []string {
	if len(s) < 2 || s[0] != '@' {
		return nil
	}
	i := 1
	// Must have at least one [a-z-]
	c := s[i]
	if !(c >= 'a' && c <= 'z') && c != '-' {
		return nil
	}
	// Match rest of [a-z-]+
	for i < len(s) {
		c = s[i]
		if !(c >= 'a' && c <= 'z') && c != '-' {
			break
		}
		i++
	}
	return []string{s[:i]}
}

// matchCombinatorSlashed matches ^\/[a-z]+\/ and returns the matched string or ""
// Replaces: reCombinatorSlashed = regexp.MustCompile(`^\/[a-z]+\/`)
func matchCombinatorSlashed(s string) string {
	if len(s) < 3 || s[0] != '/' {
		return ""
	}
	i := 1
	// Must have at least one [a-z]
	if !isLowerAlpha(s[i]) {
		return ""
	}
	for i < len(s) && isLowerAlpha(s[i]) {
		i++
	}
	// Must end with /
	if i >= len(s) || s[i] != '/' {
		return ""
	}
	return s[:i+1]
}

// matchOperatorSpaced matches ^[-+]\s+ and returns []string{fullMatch} or nil
// Replaces: reOperatorSpaced = regexp.MustCompile(`^[-+]\s+`)
func matchOperatorSpaced(s string) []string {
	if len(s) < 2 {
		return nil
	}
	if s[0] != '-' && s[0] != '+' {
		return nil
	}
	if !isWhitespaceChar(s[1]) {
		return nil
	}
	i := 2
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return []string{s[:i]}
}

// matchWhitespaceStr matches ^\s* and returns the matched whitespace string
// Used as a direct replacement for reWhitespace
func matchWhitespaceStr(s string) string {
	i := 0
	for i < len(s) && isWhitespaceChar(s[i]) {
		i++
	}
	return s[:i]
}

// matchVariableAtCurly matches ^@\{([\w-]+)\} and returns []string{fullMatch, group1} or nil
// Replaces: reVariableAtCurly = regexp.MustCompile(`^@\{([\w-]+)\}`)
func matchVariableAtCurly(s string) []string {
	if len(s) < 4 || s[0] != '@' || s[1] != '{' {
		return nil
	}
	i := 2
	start := i
	// Must have at least one [\w-]
	if i >= len(s) || !isWordOrHyphenChar(s[i]) {
		return nil
	}
	for i < len(s) && isWordOrHyphenChar(s[i]) {
		i++
	}
	// Must end with }
	if i >= len(s) || s[i] != '}' {
		return nil
	}
	return []string{s[:i+1], s[start:i]}
}

// matchVariableDollarCurly matches ^\$\{([\w-]+)\} and returns []string{fullMatch, group1} or nil
// Replaces: reVariableDollarCurly = regexp.MustCompile(`^\$\{([\w-]+)\}`)
func matchVariableDollarCurly(s string) []string {
	if len(s) < 4 || s[0] != '$' || s[1] != '{' {
		return nil
	}
	i := 2
	start := i
	// Must have at least one [\w-]
	if i >= len(s) || !isWordOrHyphenChar(s[i]) {
		return nil
	}
	for i < len(s) && isWordOrHyphenChar(s[i]) {
		i++
	}
	// Must end with }
	if i >= len(s) || s[i] != '}' {
		return nil
	}
	return []string{s[:i+1], s[start:i]}
}
