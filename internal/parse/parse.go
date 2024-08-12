package parse

import (
	"github.com/LucazFFz/lox/internal/ast"
	"github.com/LucazFFz/lox/internal/token"
)

type parser struct {
	tokens  []token.Token
	current int
	report  func(int, string, string)
}

func newParser(tokens []token.Token, report func(int, string, string)) *parser {
	return &parser{tokens, 0, report}
}

func expression(s *parser) ast.Expr {
	return equality(s)
}

func equality(s *parser) ast.Expr {
	expr := comparison(s)

	for s.match(token.EQUAL_EQUAL, token.BANG_EQUAL) {
		operator := s.previous()
		right := comparison(s)
		expr = ast.Binary{expr, operator, right}
	}

	return expr
}

func comparison(s *parser) ast.Expr {
	expr := term(s)

	for s.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := s.previous()
		right := term(s)
		expr = ast.Binary{expr, operator, right}
	}

	return expr
}

func term(s *parser) ast.Expr {
	expr := factor(s)

	for s.match(token.MINUS, token.PLUS) {
		operator := s.previous()
		right := factor(s)
		expr = ast.Binary{expr, operator, right}
	}

	return expr
}

func factor(s *parser) ast.Expr {
	expr := unary(s)

	for s.match(token.SLASH, token.STAR) {
		operator := s.previous()
		right := unary(s)
		expr = ast.Binary{expr, operator, right}
	}

	return expr
}

func unary(s *parser) ast.Expr {
	if s.match(token.BANG, token.MINUS) {
		operator := s.previous()
		right := unary(s)
		return ast.Unary{operator, right}
	}

	return primary(s)
}

func primary(s *parser) ast.Expr {
	if s.match(token.FALSE) {
		return ast.Literal{"false"}
	}
	if s.match(token.TRUE) {
		return ast.Literal{"true"}
	}
	if s.match(token.NIL) {
		return ast.Literal{"nil"}
	}
	if s.match(token.NUMBER, token.STRING) {
		return ast.Literal{s.previous().Lexme}
	}
	if s.match(token.LEFT_PAREN) {
		expr := expression(s)
		s.consume(token.RIGHT_PAREN, "Expected ')' after expression")
		return ast.Grouping{expr}
	}

	return nil
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
