package go_parser

import (
	"regexp"
)

// ParserInput represents the parser input state
type ParserInput struct {
	input                    string
	i                        int
	j                        int
	current                  string
	currentPos               int
	furthest                 int
	furthestPossibleErrorMessage string
	chunks                   []string
	saveStack                []saveState
	autoCommentAbsorb        bool
	commentStore             []inputComment
	finished                 bool
}

type saveState struct {
	current string
	i       int
	j       int
}

// inputComment represents a parsed comment in the input
type inputComment struct {
	index        int
	text         string
	isLineComment bool
}

const (
	charCodeSpace        = 32
	charCodeTab          = 9
	charCodeLF           = 10
	charCodeCR           = 13
	charCodePlus         = 43
	charCodeComma        = 44
	charCodeForwardSlash = 47
	charCode9            = 57
)

// NewParserInput creates a new ParserInput instance
func NewParserInput() *ParserInput {
	return &ParserInput{
		autoCommentAbsorb: true,
		commentStore:      make([]inputComment, 0),
		finished:          false,
	}
}

func (p *ParserInput) skipWhitespace(length int) bool {
	oldi := p.i
	oldj := p.j
	curr := oldi - p.currentPos             // Calculate current position within the chunk *before* advancing
	endIndex := oldi + len(p.current) - curr // Calculate end index relative to *start* of skipWhitespace
	mem := oldi + length                 // Target index after initial length advancement
	p.i = mem                            // Advance p.i by initial length

	inp := p.input
	var c byte
	var nextChar byte
	var comment inputComment

	// Loop from the new p.i to skip actual whitespace/comments
	for ; p.i < endIndex && p.i < len(inp); p.i++ {
		c = inp[p.i]

		if p.autoCommentAbsorb && c == charCodeForwardSlash {
			if p.i+1 < len(inp) {
				nextChar = inp[p.i+1]
				if nextChar == '/' { // Line comment
					comment = inputComment{index: p.i, isLineComment: true}
					nextNewLine := -1
					for k := p.i + 2; k < len(inp); k++ { // Use k
						if inp[k] == '\n' {
							nextNewLine = k
							break
						}
					}

					finalPos := 0
					if nextNewLine < 0 {
						finalPos = len(inp) // Comment goes to end of input
					} else {
						finalPos = nextNewLine // Comment ends at newline
					}

					comment.text = inp[comment.index:finalPos] // Slice text correctly
					p.commentStore = append(p.commentStore, comment)
					p.i = finalPos // Explicitly set p.i to the final position

					if p.i >= len(inp) {
						// When we reach EOF via a line comment, increment the index
						// one more time to match JS behavior
						p.i++
						break // Exit loop cleanly if we hit EOF
					}
					// If we didn't break, it means we stopped at a newline.
					// We need the outer loop to increment p.i PAST the newline.
					continue
				} else if nextChar == '*' { // Block comment
					nextStarSlash := -1
					for k := p.i + 2; k < len(inp)-1; k++ { // Use k
						if inp[k] == '*' && inp[k+1] == '/' {
							nextStarSlash = k
							break
						}
					}
					if nextStarSlash >= 0 {
						comment = inputComment{
							index:         p.i,
							text:          inp[p.i : nextStarSlash+2],
							isLineComment: false,
						}
						p.i = nextStarSlash + 1 // Set p.i to the '/' of '*/'
						p.commentStore = append(p.commentStore, comment)
						if p.i >= len(inp) { // Check if we are already at or past EOF
							break // Exit loop if we consumed the rest of input
						}
						continue // Outer loop increments p.i
					}
				}
			}
			// Not a comment, break loop normally if '/' is not whitespace
			if c != charCodeSpace && c != charCodeLF && c != charCodeTab && c != charCodeCR {
				break
			}
		}

		if c != charCodeSpace && c != charCodeLF && c != charCodeTab && c != charCodeCR {
			break // Break for non-whitespace characters
		}
	}

	// Calculate the slice start index based on the final position relative to the original currentPos
	sliceStart := p.i - oldi + curr
	if sliceStart < len(p.current) {
		p.current = p.current[sliceStart:]
	} else {
		p.current = ""
	}
	p.currentPos = p.i // Update currentPos to the final index

	if len(p.current) == 0 {
		if p.j < len(p.chunks)-1 {
			p.current = p.chunks[p.j+1]
			p.j++
			p.skipWhitespace(0) // skip space at the beginning of a new chunk
			return true // things changed
		}
		p.finished = true
	}

	return oldi != p.i || oldj != p.j
}

func (p *ParserInput) Save() {
	p.currentPos = p.i
	p.saveStack = append(p.saveStack, saveState{
		current: p.current,
		i:       p.i,
		j:       p.j,
	})
}

func (p *ParserInput) Restore(possibleErrorMessage string) {
	if p.i > p.furthest || (p.i == p.furthest && possibleErrorMessage != "" && p.furthestPossibleErrorMessage == "") {
		p.furthest = p.i
		p.furthestPossibleErrorMessage = possibleErrorMessage
	}
	state := p.saveStack[len(p.saveStack)-1]
	p.saveStack = p.saveStack[:len(p.saveStack)-1]
	p.current = state.current
	p.currentPos = p.i
	p.i = state.i
	p.j = state.j
}

func (p *ParserInput) Forget() {
	p.saveStack = p.saveStack[:len(p.saveStack)-1]
}

func (p *ParserInput) IsWhitespace(offset int) bool {
	pos := p.i + offset
	if pos < 0 || pos >= len(p.input) {
		return false
	}
	code := p.input[pos]
	return code == charCodeSpace || code == charCodeCR || code == charCodeTab || code == charCodeLF
}

func (p *ParserInput) Re(tok *regexp.Regexp) any {
	if p.i > p.currentPos {
		p.current = p.current[p.i-p.currentPos:]
		p.currentPos = p.i
	}

	match := tok.FindStringSubmatch(p.current)
	if match == nil {
		return nil
	}

	matchLen := len(match[0])
	p.i += matchLen
	p.skipWhitespace(0)
	if len(match) == 1 {
		return match[0]
	}
	return match
}

func (p *ParserInput) Char(tok byte) any {
	if p.i >= len(p.input) || p.input[p.i] != tok {
		return nil
	}
	p.skipWhitespace(1) // Advance by 1 and skip subsequent whitespace/comments
	return tok
}

func (p *ParserInput) PeekChar(tok byte) any {
	if p.i >= len(p.input) || p.input[p.i] != tok {
		return nil
	}
	return tok
}

func (p *ParserInput) Str(tok string) any {
	tokLength := len(tok)
	if p.i+tokLength > len(p.input) {
		return nil
	}

	for i := 0; i < tokLength; i++ {
		if p.input[p.i+i] != tok[i] {
			return nil
		}
	}

	p.i += tokLength
	p.skipWhitespace(0)
	return tok
}

func (p *ParserInput) Quoted(loc int) any {
	pos := loc
	if pos < 0 {
		pos = p.i
	}
	if pos >= len(p.input) {
		return nil
	}

	startChar := p.input[pos]
	if startChar != '\'' && startChar != '"' {
		return nil
	}

	length := len(p.input)
	currentPosition := pos

	for i := 1; i+currentPosition < length; i++ {
		nextChar := p.input[i+currentPosition]
		switch nextChar {
		case '\\':
			i++
			continue
		case '\r', '\n':
			// ignore newline in quoted string
		case startChar:
			str := p.input[currentPosition : i+currentPosition+1]
			if loc < 0 {
				p.i = currentPosition + i + 1
				p.skipWhitespace(0)
				return str
			}
			return []any{startChar, str}
		}
	}
	return nil
}

func (p *ParserInput) ParseUntil(tok any) any {
	var returnVal any
	var inComment bool
	var blockDepth int
	blockStack := make([]byte, 0)
	parseGroups := make([]any, 0)
	length := len(p.input)
	lastPos := p.i
	i := p.i
	loop := true
	var testChar func(byte) bool

	switch t := tok.(type) {
	case string:
		testChar = func(char byte) bool { return char == t[0] }
	case *regexp.Regexp:
		testChar = func(char byte) bool { return t.MatchString(string(char)) }
	default:
		return nil
	}

	for loop {
		if i >= length {
			loop = false
			break
		}

		nextChar := p.input[i]
		if blockDepth == 0 && testChar(nextChar) {
			returnVal = p.input[lastPos:i]
			if returnVal != "" {
				parseGroups = append(parseGroups, returnVal)
			} else {
				parseGroups = append(parseGroups, " ")
			}
			returnVal = parseGroups
			p.i = i
			p.skipWhitespace(0)
			loop = false
		} else {
			if inComment {
				if nextChar == '*' && i+1 < length && p.input[i+1] == '/' {
					i++
					blockDepth--
					inComment = false
				}
				i++
				continue
			}
			switch nextChar {
			case '\\':
				i++
				if i < length {
					parseGroups = append(parseGroups, p.input[lastPos:i+1])
					lastPos = i + 1
				}
			case '/':
				if i+1 < length && p.input[i+1] == '*' {
					i++
					inComment = true
					blockDepth++
				}
			case '\'', '"':
				quoteResult := p.Quoted(i)
				if quoteResult != nil {
					if quoteSlice, ok := quoteResult.([]any); ok {
						parseGroups = append(parseGroups, p.input[lastPos:i], quoteSlice[1])
						i += len(quoteSlice[1].(string)) - 1
						lastPos = i + 1
					} else {
						p.i = i
						p.skipWhitespace(0)
						returnVal = nextChar
						loop = false
					}
				} else {
					p.i = i
					p.skipWhitespace(0)
					returnVal = nextChar
					loop = false
				}
			case '{':
				blockStack = append(blockStack, '}')
				blockDepth++
			case '(':
				blockStack = append(blockStack, ')')
				blockDepth++
			case '[':
				blockStack = append(blockStack, ']')
				blockDepth++
			case '}', ')', ']':
				if len(blockStack) > 0 {
					expected := blockStack[len(blockStack)-1]
					blockStack = blockStack[:len(blockStack)-1]
					if nextChar == expected {
						blockDepth--
					} else {
						p.i = i
						p.skipWhitespace(0)
						returnVal = string(expected)
						loop = false
					}
				}
			}
			i++
		}
	}

	if returnVal != nil {
		return returnVal
	}
	return nil
}

func (p *ParserInput) Peek(tok any) bool {
	switch t := tok.(type) {
	case string:
		if p.i+len(t) > len(p.input) {
			return false
		}
		for i := 0; i < len(t); i++ {
			if p.input[p.i+i] != t[i] {
				return false
			}
		}
		return true
	case *regexp.Regexp:
		return t.MatchString(p.current)
	default:
		return false
	}
}

func (p *ParserInput) CurrentChar() byte {
	if p.i >= len(p.input) {
		return 0
	}
	return p.input[p.i]
}

func (p *ParserInput) PrevChar() byte {
	if p.i <= 0 {
		return 0
	}
	return p.input[p.i-1]
}

func (p *ParserInput) GetInput() string {
	return p.input
}

func (p *ParserInput) PeekNotNumeric() bool {
	if p.i >= len(p.input) {
		return true
	}
	c := p.input[p.i]
	return (c > charCode9 || c < charCodePlus) || c == charCodeForwardSlash || c == charCodeComma
}

func (p *ParserInput) Start(str string, chunkInput bool, failFunction func(string, int)) {
	p.input = str
	p.i = 0
	p.j = 0
	p.currentPos = 0
	p.furthest = 0
	p.furthestPossibleErrorMessage = ""
	p.commentStore = make([]inputComment, 0)
	p.finished = false

	if chunkInput {
		p.chunks = Chunker(str, failFunction)
	} else {
		p.chunks = []string{str}
	}

	p.current = p.chunks[0]
	p.skipWhitespace(0)
}

type EndState struct {
	IsFinished              bool
	Furthest                int
	FurthestPossibleErrorMessage string
	FurthestReachedEnd      bool
	FurthestChar            byte
}

func (p *ParserInput) End() EndState {
	var message string
	isFinished := p.i >= len(p.input)

	if p.i < p.furthest {
		message = p.furthestPossibleErrorMessage
		p.i = p.furthest
	}

	var furthestChar byte
	if p.i < len(p.input) {
		furthestChar = p.input[p.i]
	}

	return EndState{
		IsFinished:              isFinished,
		Furthest:                p.i,
		FurthestPossibleErrorMessage: message,
		FurthestReachedEnd:      p.i >= len(p.input)-1,
		FurthestChar:            furthestChar,
	}
} 