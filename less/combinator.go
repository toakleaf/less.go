package less_go

var NoSpaceCombinators = map[string]bool{
	"":  true,
	" ": true,
	"|": true,
}

type Combinator struct {
	*Node
	Value            string
	EmptyOrWhitespace bool
}

func NewCombinator(value string) *Combinator {
	c := &Combinator{
		Node: NewNode(),
	}

	if value == " " {
		c.Value = " "
		c.EmptyOrWhitespace = true
	} else if value != "" {
		c.Value = Intern(trimWhitespace(value))
		c.EmptyOrWhitespace = c.Value == ""
	} else {
		c.Value = ""
		c.EmptyOrWhitespace = true
	}

	return c
}

func trimWhitespace(s string) string {
	if s == "" {
		return s
	}
	
	runes := []rune(s)
	start := 0
	end := len(runes) - 1

	for start <= end && isJSWhitespace(runes[start]) {
		start++
	}

	for end >= start && isJSWhitespace(runes[end]) {
		end--
	}
	
	if start > end {
		return ""
	}
	
	return string(runes[start : end+1])
}

func isJSWhitespace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', // ASCII whitespace
		0x00A0, // NO-BREAK SPACE
		0x1680, // OGHAM SPACE MARK
		0x180E, // MONGOLIAN VOWEL SEPARATOR
		0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200A, // Various spaces
		0x2028, // LINE SEPARATOR
		0x2029, // PARAGRAPH SEPARATOR
		0x202F, // NARROW NO-BREAK SPACE
		0x205F, // MEDIUM MATHEMATICAL SPACE
		0x3000, // IDEOGRAPHIC SPACE
		0xFEFF: // ZERO WIDTH NO-BREAK SPACE (BOM)
		return true
	default:
		return false
	}
}

func (c *Combinator) Type() string {
	return "Combinator"
}

func (c *Combinator) GetType() string {
	return "Combinator"
}

func (c *Combinator) GenCSS(context any, output *CSSOutput) {
	var spaceOrEmpty string
	if ctx, ok := context.(map[string]any); ok {
		if compress, ok := ctx["compress"].(bool); ok && compress {
			spaceOrEmpty = ""
		} else if NoSpaceCombinators[c.Value] {
			spaceOrEmpty = ""
		} else {
			spaceOrEmpty = " "
		}
	} else {
		spaceOrEmpty = " "
	}
	output.Add(spaceOrEmpty+c.Value+spaceOrEmpty, nil, nil)
} 