package token

import (
	"fmt"
)

//go:generate stringer -type=TokenType
type TokenType uint8

type Token struct {
	Type    TokenType
	Lexme   string
	Literal []byte
	Line    int
}

func NewToken(token TokenType, lexme string, literal []byte, line int) Token {
	return Token{token, lexme, literal, line}
}

func (t Token) String() string {
	return fmt.Sprintf(`[%v] "%s" (%d)`, t.Type, t.Lexme, t.Line)
}

const (
	WHITESPACE TokenType = iota
	COMMENT
	EOF

	// Single-character tokens.
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	PLUS
	MINUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL
	COLON
	QUESTION

	// Literals
	IDENTIFIER
	STRING
	NUMBER

	// Keywords
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE
)
