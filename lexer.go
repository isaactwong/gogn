package gogn

import (
	"fmt"
	"strings"
)

type location struct {
	line uint
	col  uint
}

type keyword string

const (
	selectKeyword keyword = "select"
	fromKeyword   keyword = "from"
	whereKeyword  keyword = "where"
	asKeyword     keyword = "as"
	tableKeyword  keyword = "table"
	createKeyword keyword = "create"
	insertKeyword keyword = "insert"
	intoKeyword   keyword = "into"
	valuesKeyword keyword = "values"
	intKeyword    keyword = "int"
	textKeyword   keyword = "text"
)

func validKeywords() []string {
	keywords := []keyword{
		selectKeyword,
		fromKeyword,
		whereKeyword,
		asKeyword,
		tableKeyword,
		createKeyword,
		insertKeyword,
		intoKeyword,
		valuesKeyword,
		intKeyword,
		textKeyword,
	}

	var options []string
	for _, k := range keywords {
		options = append(options, string(k))
	}
	return options
}

type symbol string

const (
	semicolonSymbol  symbol = ";"
	asteriskSymbol   symbol = "*"
	commaSymbol      symbol = ","
	leftParenSymbol  symbol = "("
	rightParenSymbol symbol = ")"
	concatSymbol     symbol = "||"
	equalsSymbol     symbol = "="
)

func validSymbols() []string {
	symbols := []symbol{
		semicolonSymbol,
		asteriskSymbol,
		commaSymbol,
		leftParenSymbol,
		rightParenSymbol,
		concatSymbol,
		equalsSymbol,
	}

	var options []string
	for _, s := range symbols {
		options = append(options, string(s))
	}

	return options
}

type tokenKind uint

const (
	keywordKind    tokenKind = iota
	symbolKind     tokenKind = iota
	identifierKind tokenKind = iota
	stringKind     tokenKind = iota
	numericKind    tokenKind = iota
)

type token struct {
	value string
	kind  tokenKind
	loc   location
}

type cursor struct {
	pointer uint
	loc     location
}

func (t *token) equals(other *token) bool {
	return t.value == other.value && t.kind == other.kind
}

type lexer func(string, cursor) (*token, cursor, bool)

func lex(source string) ([]*token, error) {
	tokens := []*token{}
	cur := cursor{}
	//	lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumeric, lexIdentifier}

	for cur.pointer < uint(len(source)) {
		if token, newCursor, ok := lexForwardFromCursor(source, cur); ok {
			cur = newCursor

			// Omit nil tokens for valid but empty syntax like newlines.
			if token != nil {
				tokens = append(tokens, token)
			}
		} else {
			hint := ""
			if len(tokens) > 0 {
				hint = " after " + tokens[len(tokens)-1].value
			}
			return nil, fmt.Errorf("Unable to lex tokens%s at %d:%d", hint, cur.loc.line, cur.loc.col)
		}
	}

	// lex:
	// for cur.pointer < uint(len(source)) {
	// for _, l := range lexers {
	// if token, newCursor, ok := l(source, cur); ok {
	// cur = newCursor
	//
	// // Omit nil tokens for valid but empty syntax like newlines.
	// if token != nil {
	// tokens = append(tokens, token)
	// }
	// continue lex
	// }
	// }
	//
	// hint := ""
	// if len(tokens) > 0 {
	// hint = " after " + tokens[len(tokens)-1].value
	// }
	// return nil, fmt.Errorf("Unable to lex tokens%s at %d:%d", hint, cur.loc.line, cur.loc.col)
	// }

	return tokens, nil
}

func lexForwardFromCursor(source string, currentPosition cursor) (*token, cursor, bool) {
	lexers := []lexer{lexKeyword, lexSymbol, lexString, lexNumeric, lexIdentifier}

	for _, l := range lexers {
		if tok, newPosition, ok := l(source, currentPosition); ok {
			ok = true
			return tok, newPosition, ok
		}
	}
	return nil, currentPosition, false
}

func lexNumeric(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	periodFound := false
	expMarkerFound := false

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		cur.loc.col++

		isDigit := c >= '0' && c <= '9'
		isPeriod := c == '.'
		isExpMarker := c == 'e'

		// Number must start with a digit or a period.
		if cur.pointer == ic.pointer {
			if !isDigit && !isPeriod {
				return nil, ic, false
			}
			periodFound = isPeriod
			continue
		}

		if isPeriod {
			if periodFound {
				return nil, ic, false
			}

			periodFound = true
			continue
		}

		if isExpMarker {
			if expMarkerFound {
				return nil, ic, false
			}

			// No periods are allowed after expMarker
			periodFound = true
			expMarkerFound = true

			// expMarker must be followed by digits.
			if cur.pointer == uint(len(source)-1) {
				return nil, ic, false
			}

			cNext := source[cur.pointer+1]
			if cNext == '-' || cNext == '+' {
				cur.pointer++
				cur.loc.col++
			}
			continue
		}
		if !isDigit {
			break
		}
	}
	if cur.pointer == ic.pointer {
		return nil, ic, false
	}
	return &token{
			value: source[ic.pointer:cur.pointer],
			loc:   ic.loc,
			kind:  numericKind,
		},
		cur,
		true
}

func lexCharacterDelimited(source string, ic cursor, delimiter byte) (*token, cursor, bool) {
	cur := ic
	if len(source[cur.pointer:]) == 0 {
		return nil, ic, false
	}

	if source[cur.pointer] != delimiter {
		return nil, ic, false
	}

	cur.loc.col++
	cur.pointer++

	var value []byte

	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c := source[cur.pointer]
		if c == delimiter {
			// SQL escapes are via double characters, not backslash
			if cur.pointer+1 >= uint(len(source)) || source[cur.pointer+1] != delimiter {
				cur.pointer++
				cur.loc.col++
				return &token{
						value: string(value),
						loc:   ic.loc,
						kind:  stringKind,
					},
					cur,
					true
			} else {
				value = append(value, delimiter)
				cur.pointer++
				cur.loc.col++
			}
		}
		value = append(value, c)
		cur.loc.col++
	}
	return nil, ic, false
}

func lexString(source string, ic cursor) (*token, cursor, bool) {
	return lexCharacterDelimited(source, ic, '\'')
}

func lexSymbol(source string, ic cursor) (*token, cursor, bool) {
	c := source[ic.pointer]
	cur := ic

	// Will get overwritten later if not an ignored syntax
	cur.pointer++
	cur.loc.col++

	switch c {
	// Syntax that should be thrown away
	case '\n':
		cur.loc.line++
		cur.loc.col = 0
		fallthrough
	case '\t':
		fallthrough
	case ' ':
		return nil, cur, true
	}

	// Use `ic`, not `cur`
	match := longestMatch(source, ic, validSymbols())
	// Unknown character
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
			value: match,
			loc:   ic.loc,
			kind:  symbolKind,
		},
		cur,
		true
}

func lexKeyword(source string, ic cursor) (*token, cursor, bool) {
	cur := ic
	match := longestMatch(source, ic, validKeywords())
	if match == "" {
		return nil, ic, false
	}

	cur.pointer = ic.pointer + uint(len(match))
	cur.loc.col = ic.loc.col + uint(len(match))

	return &token{
		value: match,
		kind:  keywordKind,
		loc:   ic.loc,
	}, cur, true
}

func longestMatch(source string, ic cursor, options []string) string {
	var (
		value    []byte
		skipList []int
		match    string
	)

	cur := ic

	for cur.pointer < uint(len(source)) {
		value = append(value, strings.ToLower(string(source[cur.pointer]))[0])
		cur.pointer++

	match:
		for i, option := range options {
			for _, skip := range skipList {
				if i == skip {
					continue match
				}
			}

			// Deal with cases like INT and INTO
			if option == string(value) {
				skipList = append(skipList, i)
				if len(option) > len(match) {
					match = option
				}
				continue
			}

			sharesPrefix := string(value) == option[:cur.pointer-ic.pointer]
			tooLong := len(value) > len(option)
			if tooLong || !sharesPrefix {
				skipList = append(skipList, i)
			}
		}
		if len(skipList) == len(options) {
			break
		}
	}

	return match
}

func lexIdentifier(source string, ic cursor) (*token, cursor, bool) {
	if token, newCursor, ok := lexCharacterDelimited(source, ic, '"'); ok {
		return token, newCursor, true
	}

	cur := ic

	c := source[cur.pointer]
	isAlpha := isAlphabetical(c)
	if !isAlpha {
		return nil, ic, false
	}
	cur.pointer++
	cur.loc.col++

	value := []byte{c}
	for ; cur.pointer < uint(len(source)); cur.pointer++ {
		c = source[cur.pointer]
		if isAlphabetical(c) || isNumeric(c) || c == '$' || c == '_' {
			value = append(value, c)
			cur.loc.col++
			continue
		}
		break
	}

	if len(value) == 0 {
		return nil, ic, false
	}

	return &token{
		// Unquoted identifiers are case-insensitive
		value: strings.ToLower(string(value)),
		loc:   ic.loc,
		kind:  identifierKind,
	}, cur, true
}

func isAlphabetical(c byte) bool {
	isAlpha := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	return isAlpha
}

func isNumeric(c byte) bool {
	isNum := c >= '0' && c <= '9'
	return isNum
}
