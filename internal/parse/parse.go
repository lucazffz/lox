package parse

import (
	"errors"

	"github.com/LucazFFz/lox/internal/ast"
	"github.com/LucazFFz/lox/internal/token"
)

// Precedence and associativity rules are based on
// the C programming language.

type parser struct {
	tokens  []token.Token
	current int
	report  func(int, string, string)
}

func newParser(tokens []token.Token, report func(int, string, string)) *parser {
	return &parser{tokens, 0, report}
}

func Parse(tokens []token.Token, report func(int, string, string)) (ast.Expr, error) {
	parser := newParser(tokens, report)
	expr, err := expression(parser)
	if err != nil {
		return nil, err
	}

	return expr, nil
}

// expression -> ternary;
// precedence: none
// associativity: none
func expression(s *parser) (ast.Expr, error) {
	return comma(s)
}

// comma -> ternary ("," ternary)*;
// precedence: 15
// associativity: left-to-right
func comma(s *parser) (ast.Expr, error) {
	expr, err := ternary(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.COMMA) {
		operator := s.previous()
		if right, err := ternary(s); err != nil {
			return nil, err
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, nil
}

// ternary -> equlity "?" equality ":" ternary | equality;
// precedence: 13
// associativity: right-to-left
func ternary(s *parser) (ast.Expr, error) {
	expr, err := equality(s)
	if err != nil {
		return nil, err
	}

	if !s.match(token.QUESTION) {
		return expr, nil
	}

	left, err := equality(s)
	if err != nil {
		return nil, err
	}

	if !s.match(token.COLON) {
		return nil, errors.New("expected ':' as part of ternary operator")
	}

	right, err := ternary(s)
	if err != nil {
		return nil, err
	}

	expr = ast.Ternary{Condition: expr, Left: left, Right: right}
	return expr, nil
}

// equality -> comparison (("!=" | "==") comparison)*;
// precedence: 7
// associativity: left-to-right
func equality(s *parser) (ast.Expr, error) {
	expr, err := comparison(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.EQUAL_EQUAL, token.BANG_EQUAL) {
		operator := s.previous()
		if right, err := comparison(s); err != nil {
			return nil, err
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, nil
}

// comparison -> term ((">" | ">=" | "<" | "<=") term)*;
// precedence: 6
// associativity: left-to-right
func comparison(s *parser) (ast.Expr, error) {
	expr, err := term(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := s.previous()
		if right, err := term(s); err != nil {
			return nil, err
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, nil
}

// term -> factor (("-" | "+") factor)*;
// precedence: 4
// associativity: left-to-right
func term(s *parser) (ast.Expr, error) {
	expr, err := factor(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.MINUS, token.PLUS) {
		operator := s.previous()
		if right, err := factor(s); err != nil {
			return nil, err
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, nil
}

// factor -> unary (("/" | "*") unary)*;
// precedence: 3
// associativity: left-to-right
func factor(s *parser) (ast.Expr, error) {
	expr, err := unary(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.SLASH, token.STAR) {
		operator := s.previous()
		if right, err := unary(s); err != nil {
			return nil, err
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, err
}

// unary -> ("!" | "-") unary | primary;
// precedence: 2
// associativity: right-to-left
func unary(s *parser) (ast.Expr, error) {
	if s.match(token.BANG, token.MINUS) {
		operator := s.previous()
		if right, err := unary(s); err != nil {
			return nil, err
		} else {
			return ast.Unary{Op: operator, Right: right}, nil
		}
	}

	return primary(s)
}

// primary -> NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")";
// precedence: 1
// associativity: none
func primary(s *parser) (ast.Expr, error) {
	if s.match(token.FALSE) {
		return ast.Literal{Value: "false"}, nil
	}
	if s.match(token.TRUE) {
		return ast.Literal{Value: "true"}, nil
	}
	if s.match(token.NIL) {
		return ast.Literal{Value: "nil"}, nil
	}
	if s.match(token.NUMBER, token.STRING) {
		return ast.Literal{Value: s.previous().Lexme}, nil
	}
	if s.match(token.LEFT_PAREN) {
		if expr, err := expression(s); err != nil {
			return nil, err
		} else {
			s.consume(token.RIGHT_PAREN, "expected ')' after expression")
			return ast.Grouping{Expr: expr}, nil
		}
	}

	return nil, errors.New("could not parse tokens")
}

func (s *parser) synchronize() {
	s.advance()

	for !s.atEndOfFile() {
		if s.previous().Type == token.SEMICOLON {
			return
		}

		switch s.peek().Type {
		case token.CLASS:
			return
		case token.FUN:
			return
		case token.VAR:
			return
		case token.FOR:
			return
		case token.IF:
			return
		case token.WHILE:
			return
		case token.PRINT:
			return
		case token.RETURN:
			return
		}

		s.advance()
	}
}

func (s *parser) consume(typ token.TokenType, msg string) {
	if s.check(typ) {
		s.advance()
	}

	s.report(s.peek().Line, s.peek().Lexme, msg)
}

func (s *parser) match(types ...token.TokenType) bool {
	for _, typ := range types {
		if s.check(typ) {
			s.advance()
			return true
		}
	}

	return false
}

func (s *parser) check(typ token.TokenType) bool {
	if s.atEndOfFile() {
		return false
	}
	return s.peek().Type == typ
}

func (s *parser) advance() token.Token {
	if !s.atEndOfFile() {
		s.current++
	}
	return s.previous()
}

func (s *parser) previous() token.Token {
	return s.tokens[s.current-1]
}

func (s *parser) peek() token.Token {
	return s.tokens[s.current]
}

func (s *parser) atEndOfFile() bool {
	return s.peek().Type == token.EOF
}
