package gogn

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestTokenLexNumeric(t *testing.T) {
	tests := []struct {
		isValidNumber bool
		number        string
	}{
		{isValidNumber: true, number: "105"},
		{isValidNumber: true, number: "123."},
		{isValidNumber: true, number: "123.145"},
		{isValidNumber: true, number: "1e5"},
		{isValidNumber: true, number: "1.e21"},
		{isValidNumber: true, number: "1.1e2"},
		{isValidNumber: true, number: "1.1e-2"},
		{isValidNumber: true, number: "1.1e+2"},
		{isValidNumber: true, number: "1e-1"},
		{isValidNumber: true, number: ".1"},
		{isValidNumber: true, number: "0.105"},
		{isValidNumber: true, number: "1.105"},
		{isValidNumber: false, number: "e4"},
		{isValidNumber: false, number: "1.."},
		{isValidNumber: false, number: "1ee4"},
		{isValidNumber: false, number: " 1"},
	}

	for _, test := range tests {
		tok, _, ok := lexNumeric(test.number, cursor{})
		assert.Equal(t, test.isValidNumber, ok, test.number)
		if ok {
			assert.Equal(t, strings.TrimSpace(test.number), tok.value, test.number)
		}
	}
}
func TestTokenLexIdentifier(t *testing.T) {
	tests := []struct {
		isValidIdentifier bool
		input             string
		value             string
	}{
		{
			isValidIdentifier: true,
			input:             "a",
			value:             "a",
		},
		{
			isValidIdentifier: true,
			input:             "abc",
			value:             "abc",
		},
		{
			isValidIdentifier: true,
			input:             "abc ",
			value:             "abc",
		},
		{
			isValidIdentifier: true,
			input:             `" abc "`,
			value:             ` abc `,
		},
		{
			isValidIdentifier: true,
			input:             "a9$",
			value:             "a9$",
		},
		{
			isValidIdentifier: true,
			input:             "userName",
			value:             "username",
		},
		{
			isValidIdentifier: true,
			input:             `"userName"`,
			value:             "userName",
		},
		{
			isValidIdentifier: false,
			input:             `"`,
		},
		{
			isValidIdentifier: false,
			input:             "_sadsfa",
		},
		{
			isValidIdentifier: false,
			input:             "9sadsfa",
		},
		{
			isValidIdentifier: false,
			input:             " abc",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexIdentifier(test.input, cursor{})
		assert.Equal(t, test.isValidIdentifier, ok, test.input)
		if ok {
			assert.Equal(t, test.value, tok.value, test.input)
		}
	}
}

func TestTokenLexKeyword(t *testing.T) {
	tests := []struct {
		isValidKeyword bool
		value          string
	}{
		{
			isValidKeyword: true,
			value:          "select ",
		},
		{
			isValidKeyword: true,
			value:          "from",
		},
		{
			isValidKeyword: true,
			value:          "as",
		},
		{
			isValidKeyword: true,
			value:          "SELECT",
		},
		{
			isValidKeyword: true,
			value:          "into",
		},
		{
			isValidKeyword: false,
			value:          " into",
		},
		{
			isValidKeyword: false,
			value:          "flubbrety",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexKeyword(test.value, cursor{})
		assert.Equal(t, test.isValidKeyword, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, strings.ToLower(test.value), tok.value, test.value)
		}
	}
}

func TestTokenLexString(t *testing.T) {
	tests := []struct {
		isValidString bool
		value         string
	}{
		{isValidString: true, value: "'abc'"},
		{isValidString: true, value: "'ab c'"},
		{isValidString: true, value: "'a b'"},
		{isValidString: true, value: "'b'"},
		{isValidString: true, value: "'a '' b'"},
		{isValidString: false, value: "a"},
		{isValidString: false, value: "'"},
		{isValidString: false, value: ""},
		{isValidString: false, value: " 'foo'"},
	}

	for _, test := range tests {
		tok, _, ok := lexString(test.value, cursor{})
		assert.Equal(t, test.isValidString, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, test.value[1:len(test.value)-1], tok.value, test.value)
		}
	}
}

func TestTokenLexSymbol(t *testing.T) {
	tests := []struct {
		isValidSymbol bool
		value         string
	}{
		{
			isValidSymbol: true,
			value:         "= ",
		},
		{
			isValidSymbol: true,
			value:         "||",
		},
	}

	for _, test := range tests {
		tok, _, ok := lexSymbol(test.value, cursor{})
		assert.Equal(t, test.isValidSymbol, ok, test.value)
		if ok {
			test.value = strings.TrimSpace(test.value)
			assert.Equal(t, test.value, tok.value, test.value)
		}
	}
}

func TestLex(t *testing.T) {
	tests := []struct {
		input  string
		tokens []token
		err    error
	}{
		{
			input: "select a",
			tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "a",
					kind:  identifierKind,
				},
			},
			err: nil,
		},
		{
			input: "select 'test'",
			tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "test",
					kind:  stringKind,
				},
			},
			err: nil,
		},

		{
			input: "CREATE TABLE u (id INT, name TEXT)",
			tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(createKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: string(tableKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 13, line: 0},
					value: "u",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 15, line: 0},
					value: "(",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 16, line: 0},
					value: "id",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 19, line: 0},
					value: "int",
					kind:  keywordKind,
				},
				{
					loc:   location{col: 22, line: 0},
					value: ",",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 24, line: 0},
					value: "name",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 29, line: 0},
					value: "text",
					kind:  keywordKind,
				},
				{
					loc:   location{col: 33, line: 0},
					value: ")",
					kind:  symbolKind,
				},
			},
			err: nil,
		},
		{
			input: "SELECT id FROM users;",
			tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "id",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 10, line: 0},
					value: string(fromKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 15, line: 0},
					value: "users",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 20, line: 0},
					value: ";",
					kind:  symbolKind,
				},
			},
			err: nil,
		},
		{
			input: "select 'foo' || 'bar';",
			tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(selectKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: "foo",
					kind:  stringKind,
				},
				{
					loc:   location{col: 13, line: 0},
					value: string(concatSymbol),
					kind:  symbolKind,
				},
				{
					loc:   location{col: 16, line: 0},
					value: "bar",
					kind:  stringKind,
				},
				{
					loc:   location{col: 21, line: 0},
					value: string(semicolonSymbol),
					kind:  symbolKind,
				},
			},
			err: nil,
		},
		{
			input: "insert into users values (105, 233)",
			tokens: []token{
				{
					loc:   location{col: 0, line: 0},
					value: string(insertKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 7, line: 0},
					value: string(intoKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 12, line: 0},
					value: "users",
					kind:  identifierKind,
				},
				{
					loc:   location{col: 18, line: 0},
					value: string(valuesKeyword),
					kind:  keywordKind,
				},
				{
					loc:   location{col: 25, line: 0},
					value: "(",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 26, line: 0},
					value: "105",
					kind:  numericKind,
				},
				{
					loc:   location{col: 30, line: 0},
					value: ",",
					kind:  symbolKind,
				},
				{
					loc:   location{col: 32, line: 0},
					value: "233",
					kind:  numericKind,
				},
				{
					loc:   location{col: 36, line: 0},
					value: ")",
					kind:  symbolKind,
				},
			},
			err: nil,
		},
	}

	for _, test := range tests {
		tokens, err := lex(test.input)
		assert.Equal(t, test.err, err, test.input)
		assert.Equal(t, len(test.tokens), len(tokens), test.input)

		for i, tok := range tokens {
			assert.Equal(t, &test.tokens[i], tok, test.input)
		}
	}
}
