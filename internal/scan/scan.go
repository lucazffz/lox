package scan

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/LucazFFz/lox/internal/token"
	"strconv"
	"unicode"
)

type scanner struct {
	src            string
	tokenStart     int
	tokenEnd       int
	line           int
	keywords       map[string]token.TokenType
	tokens         []token.Token
	context        ScanContext
	report         func(error)
	scanErrOccured bool
}

func newScanner(source string, report func(error), context ScanContext) *scanner {
	keywords := map[string]token.TokenType{
		"class":  token.CLASS,
		"and":    token.AND,
		"else":   token.ELSE,
		"false":  token.FALSE,
		"for":    token.FOR,
		"fun":    token.FUN,
		"if":     token.IF,
		"nil":    token.NIL,
		"or":     token.OR,
		"print":  token.PRINT,
		"return": token.RETURN,
		"super":  token.SUPER,
		"this":   token.THIS,
		"true":   token.TRUE,
		"var":    token.VAR,
		"while":  token.WHILE,
        "break":  token.BREAK,
	}

	return &scanner{source, 0, 0, 1, keywords, []token.Token{}, context, report, false}
}

type ScanContext struct {
	IncludeComments   bool
	IncludeWhitespace bool
}

type ScanError struct {
	Message string
	Line    int
	Lexme   string
}

func (e ScanError) Error() string {
	return fmt.Sprintf("[%d] error at \"%s\" - %s \n", e.Line, e.Lexme, e.Message)
}

func Scan(source string, report func(error), context ScanContext) ([]token.Token, error) {
	s := newScanner(source, report, context)
	for !atEndOfFile(s) {
		s.tokenEnd = s.tokenStart
		scanToken(s)
	}

	s.tokens = append(s.tokens, token.NewToken(token.EOF, "", nil, s.line))

	return s.tokens, nil
}

func scanToken(s *scanner) {

	appendToken := func(s *scanner, typ token.TokenType) {
		lexme := getLexme(s, 0, 0)
		token := token.NewToken(typ, lexme, nil, s.line)
		s.tokens = append(s.tokens, token)
	}

	c := advance(s)
	switch c {
	case '(':
		appendToken(s, token.LEFT_PAREN)
	case ')':
		appendToken(s, token.RIGHT_PAREN)
	case '{':
		appendToken(s, token.LEFT_BRACE)
	case '}':
		appendToken(s, token.RIGHT_BRACE)
	case ',':
		appendToken(s, token.COMMA)
	case '.':
		appendToken(s, token.DOT)
	case '-':
		appendToken(s, token.MINUS)
	case ';':
		appendToken(s, token.SEMICOLON)
	case '+':
		appendToken(s, token.PLUS)
	case '*':
		appendToken(s, token.STAR)
	case ':':
		appendToken(s, token.COLON)
	case '?':
		appendToken(s, token.QUESTION)
	case '!':
		if match(s, '=') {
			appendToken(s, token.BANG_EQUAL)
			break
		}
		appendToken(s, token.BANG)
	case '=':
		if match(s, '=') {
			appendToken(s, token.EQUAL_EQUAL)
			break
		}
		appendToken(s, token.EQUAL)
	case '<':
		if match(s, '=') {
			appendToken(s, token.LESS_EQUAL)
			break
		}
		appendToken(s, token.LESS)
	case '>':
		if match(s, '=') {
			appendToken(s, token.GREATER_EQUAL)
			break
		}
		appendToken(s, token.GREATER)
	case '/':
		if peek(s) == '/' || peek(s) == '*' {
			lexme := handleComment(s)
			if s.context.IncludeComments {
				token := token.NewToken(token.COMMENT, lexme, nil, s.line)
				s.tokens = append(s.tokens, token)
			}
			break
		}

		token := token.NewToken(token.SLASH, getLexme(s, 0, 0), nil, s.line)
		s.tokens = append(s.tokens, token)
	case '\n':
		s.line++
		fallthrough
	case ' ', '\r', '\t':
		if s.context.IncludeWhitespace {
			token := token.NewToken(token.WHITESPACE, string(c), nil, s.line)
			s.tokens = append(s.tokens, token)
		}
	case '"':
		lexme, err := handleString(s)
		if err != nil {
			err := ScanError{Line: s.line, Lexme: lexme, Message: err.Error()}
			s.report(err)
			s.scanErrOccured = true
            s.tokens = append(s.tokens, token.NewToken(token.ERROR, lexme, nil, s.line))
			break
		}

		token := token.NewToken(token.STRING, lexme, []byte(lexme), s.line)
		s.tokens = append(s.tokens, token)
	default:
		if unicode.IsDigit(c) {
			number := handleNumber(s)
			buf := bytes.NewBuffer(make([]byte, 0, 8))
			if err := binary.Write(buf, binary.LittleEndian, number); err != nil {
				panic(err)
			}

			lexme := getLexme(s, 0, 0)
			token := token.NewToken(token.NUMBER, lexme, buf.Bytes(), s.line)
			s.tokens = append(s.tokens, token)
			break
		}

		if unicode.IsLetter(c) || c == '_' {
			typ, lexme := handleIdentifier(s)
			token := token.NewToken(typ, lexme, []byte(lexme), s.line)
			s.tokens = append(s.tokens, token)
			break
		}

		err := ScanError{Line: s.line, Lexme: getLexme(s, 0, 0), Message: "unexpected character '" + string(c) + "'"}
		s.tokens = append(s.tokens, token.NewToken(token.ERROR, getLexme(s, 0, 0), nil, s.line))
		s.scanErrOccured = true
		s.report(err)
	}
}

func handleComment(s *scanner) string {
	if match(s, '/') {
		for peek(s) != 0 && !atEndOfFile(s) {
			advance(s)
		}
		return getLexme(s, 2, -1)
	}

	if match(s, '*') {
		for peek(s) != '*' && peekNext(s) != '/' {
			if c := peek(s); c == '\n' {
				s.line++
			} else if c == rune(0) {
				return getLexme(s, 2, -1)

			}
			advance(s)
		}
		advance(s)
		advance(s)
		return getLexme(s, 2, -2)
	}

	return ""
}

func handleString(s *scanner) (string, error) {
	for peek(s) != '"' && !atEndOfFile(s) {
		if peek(s) == '\n' {
			s.line++
		}
		advance(s)
	}

	if atEndOfFile(s) {
		return "", errors.New("unterminated string")
	}

	advance(s)
	return getLexme(s, 1, -1), nil
}

func handleNumber(s *scanner) float64 {
	for unicode.IsDigit(peek(s)) {
		advance(s)
	}

	if peek(s) == '.' && unicode.IsDigit(peekNext(s)) {
		advance(s)
		for unicode.IsDigit(peek(s)) {
			advance(s)
		}
	}
	num, _ := strconv.ParseFloat(getLexme(s, 0, 0), 64)
	return num
}

func handleIdentifier(s *scanner) (token.TokenType, string) {
	for unicode.IsDigit(peek(s)) || unicode.IsLetter(peek(s)) || peek(s) == '_' {
		advance(s)
	}

	lexme := getLexme(s, 0, 0)
	typ, ok := s.keywords[lexme]

	if !ok {
		typ = token.IDENTIFIER
	}

	return typ, lexme
}

func getLexme(s *scanner, startOffset int, endOffset int) string {
	if s.tokenEnd+startOffset < 0 ||
		s.tokenStart+endOffset > len(s.src) ||
		s.tokenEnd+startOffset > s.tokenStart+endOffset {
		return ""
	}

	return s.src[s.tokenEnd+startOffset : s.tokenStart+endOffset]
}

func atEndOfFile(s *scanner) bool {
	return s.tokenStart >= len(s.src)
}

func match(s *scanner, expected rune) bool {
	if atEndOfFile(s) {
		return false
	}
	if peek(s) != expected {
		return false
	}
	advance(s)
	return true
}

func advance(s *scanner) rune {
	s.tokenStart++
	return rune(s.src[s.tokenStart-1])
}

func peek(s *scanner) rune {
	if atEndOfFile(s) {
		return rune(0)
	}
	return rune(s.src[s.tokenStart])
}

func peekNext(s *scanner) rune {
	if s.tokenStart+1 >= len(s.src) {
		return rune(0)
	}
	return rune(s.src[s.tokenStart+1])
}
