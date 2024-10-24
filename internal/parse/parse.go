// Precedence and associativity rules are based on
// the C programming language. For full spec, see [C Operator Precedence].
//
// For full language spec, see [The Lox Language]. Note that some tweaks and additions
// have been made in this specific implementation of the language such as
// the implementation of a c-like comma operator and conditional operator
// among others.
//
// [C Operator Precedence]: https://en.cppreference.com/w/c/language/operator_precedence
// [The Lox Language]: https://craftinginterpreters.com/the-lox-language.html
package parse

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/LucazFFz/lox/internal/ast"
	"github.com/LucazFFz/lox/internal/token"
)

type parser struct {
	tokens          []token.Token
	current         int
	parseErrOccured bool
	report          func(error)
}

func newParser(tokens []token.Token, report func(error)) *parser {
	return &parser{tokens, 0, false, report}
}

type ParseError struct {
	Message string
	Line    int
	Lexme   string
}

func (e ParseError) Error() string {
	if e.Lexme == "" {
		return fmt.Sprintf("[%d] error - %s \n", e.Line, e.Message)
	}

	return fmt.Sprintf("[%d] error at \"%s\" - %s \n", e.Line, e.Lexme, e.Message)
}

// Parse generates an abstract syntax tree (ast.Expr) based on the given tokens.
// The parser will use error productions and synchronize itself between
// statements where possible to provide best effort error reporting.
//
// Parameters:
//
//   - tokens: A list of tokens to be parsed.
//   - report: A callback function which is invoked when an error occur.
//
// NOTE: Report is invoked on both resolved and unresolved errors.
//
// Returns:
//
//   - ast.Expr: An abstract syntax tree.
//   - error: An error used to signalize that a parse error occured.
//
// NOTE: The returned error do not contain any information regarding
// the given parse errors, that information is passed to report.
func Parse(tokens []token.Token, report func(error)) ([]ast.Stmt, error) {
	parser := newParser(tokens, report)
	var stmts []ast.Stmt = make([]ast.Stmt, 0)

	for parser.peek().Type != token.EOF {
		stmt, err := declaration(parser)
		// parser.advance()
		if err == nil {
			stmts = append(stmts, stmt)
		}
	}

	if parser.parseErrOccured {
		return nil, errors.New("parse error occured")
	}

	return stmts, nil
}

func ParseExpression(tokens []token.Token, report func(error)) (ast.Expr, error) {
	parser := newParser(tokens, report)
	expr, err := expression(parser)
	if err != nil {
		return nil, err
	}

	if parser.parseErrOccured {
		return nil, errors.New("parse error occured")
	}

	return expr, nil
}

// program -> declaration* EOF;

// Production rules:
//   - declaration -> varDeclaration | statement;
func declaration(s *parser) (ast.Stmt, error) {
	if s.match(token.VAR) {
		s.advance()
		stmt, err := varDeclaration(s)
		if err != nil {
			// reset the parser state between declarations
			// to aviod cascading errors
			s.synchronize()
			return nil, err
		}
		return stmt, nil
	}

	return statement(s)
}

// Production rules:
//   - varDeclaration -> "var" IDENTIFIER ( "=" expression)? ";";
func varDeclaration(s *parser) (ast.Stmt, error) {
	var name token.Token
	err := s.consume(token.IDENTIFIER, "expected variable name")
	if err != nil {
		return nil, err
	}

	name = s.previous()
	var initializer ast.Expr = ast.Nothing{}
	if s.match(token.EQUAL) {
		s.advance()
		initializer, err = expression(s)
		if err != nil {
			return nil, err
		}
	}

	if err := s.consume(token.SEMICOLON, "expected ';' after variable declaration"); err != nil {
		return nil, err
	}

	return ast.Var{Name: name, Initializer: initializer}, nil
}

// Production rules:
//   - statement -> exprStmt | printStmt | blockStmt;
func statement(s *parser) (ast.Stmt, error) {
	if s.match(token.PRINT) {
		s.advance()
		return printStmt(s)
	}

	if s.match(token.LEFT_BRACE) {
		s.advance()
		return blockStmt(s)
	}

	return expressionStmt(s)
}

// Production rules:
//   - printStmt -> "print" expression ";";
func printStmt(s *parser) (ast.Stmt, error) {
	expr, err := expression(s)
	// expressions usually do not return errors but create
	// error productions
	if err != nil {
		return nil, err
	}

	if err := s.consume(token.SEMICOLON, "expected ';' after expression"); err != nil {
		return nil, err
	}

	return ast.Print{Expr: expr}, nil
}

// Production rules:
//   - blockStmt -> "{" declaration* "}";
func blockStmt(s *parser) (ast.Stmt, error) {
	var statements []ast.Stmt

	for !s.check(token.RIGHT_BRACE) && !s.atEndOfFile() {
		stmt, err := declaration(s)
		if err != nil {
			return nil, err
		}

		statements = append(statements, stmt)
	}

	if err := s.consume(token.RIGHT_BRACE, "expected '}' after block statement"); err != nil {
		return nil, err
	}

	return ast.Block{Statements: statements}, nil
}

// Production rules:
//   - expressionStmt -> expression ";";
func expressionStmt(s *parser) (ast.Stmt, error) {
	expr, err := expression(s)
	// expressions usually do not return errors but create
	// error productions
	if err != nil {
		return nil, err
	}

	if err := s.consume(token.SEMICOLON, "expected ';' after expression"); err != nil {
		return nil, err
	}

	return ast.Expression{Expr: expr}, nil
}

// Production rules:
//   - expression -> assignment;
//   - precedence: none
//   - ssociativity: none
func expression(s *parser) (ast.Expr, error) {
	return assignment(s)
}

// Production rules:
//   - assignment -> IDENTIFIER "=" assignment | conditional;
//   - precedence: 16
//   - associativity: right-to-left
func assignment(s *parser) (ast.Expr, error) {
	expr, err := comma(s)
	if err != nil {
		return nil, err
	}

	if s.match(token.EQUAL) {
		s.advance()
		value, err := assignment(s)
		if err != nil {
			return nil, err
		}

		if expr, ok := expr.(ast.Variable); ok {
			return ast.Assign{Name: expr.Name, Value: value}, nil
		}

		err = ParseError{
			Line:    s.previous().Line,
			Lexme:   s.previous().Lexme,
			Message: "invalid assignment target"}
		s.report(err)
		return nil, errors.New("")
	}

	return expr, nil
}

// Production rules:
//   - comma -> conditional ("," conditional)*;
//   - precedence: 15
//   - associativity: left-to-right
func comma(s *parser) (ast.Expr, error) {
	expr, err := conditional(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.COMMA) {
		operator := s.previous()
		if right, err := conditional(s); err != nil {
			return nil, err
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, nil
}

// Production rules:
//   - conditional -> equlity "?" equality ":" conditional | equality;
//   - precedence: 13
//   - associativity: right-to-left
func conditional(s *parser) (ast.Expr, error) {
	expr, err := equality(s)
	if err != nil {
		return nil, err
	}

	if !s.match(token.QUESTION) {
		return expr, nil
	}

	s.advance()
	left, err := equality(s)
	if err != nil {
		return nil, err
	}

	if !s.match(token.COLON) {
		err := ParseError{
			Line:    s.peek().Line,
			Lexme:   s.peek().Lexme,
			Message: "expected ':' as part of conditional operator (conditional)"}
		s.report(err)
		return nil, errors.New("")
	}

	s.advance()
	right, err := conditional(s)
	if err != nil {
		return nil, err
	}

	expr = ast.Ternary{Condition: expr, Left: left, Right: right}
	return expr, nil
}

// Production rules:
//   - quality -> (nothing | comparison) (("!=" | "==") (nothing | comparison))*;
//   - precedence: 7
//   - associativity: left-to-right
func equality(s *parser) (ast.Expr, error) {
	expr, err := comparison(s)
	if err != nil {
		if s.match(token.EQUAL_EQUAL, token.BANG_EQUAL) {
			expr = handleMissingExpression(s, s.peek().Lexme,
				"missing left-hand-side operand (equality)")
		} else {
			return nil, err
		}
	}

	for s.match(token.EQUAL_EQUAL, token.BANG_EQUAL) {
		operator := s.peek()
		s.advance()
		if right, err := comparison(s); err != nil {
			expr = handleMissingExpression(s, s.peek().Lexme,
				"missing left-hand-side operand (equality)")
		} else {
			expr = ast.Binary{Left: expr, Op: operator, Right: right}
		}
	}

	return expr, nil
}

// Production rules:
//   - comparison -> (nothing | term) ((">" | ">=" | "<" | "<=") (nothing | term))*;
//   - precedence: 6
//   - associativity: left-to-right
func comparison(s *parser) (ast.Expr, error) {
	expr, err := term(s)
	if err != nil {
		if s.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
			expr = handleMissingExpression(s, s.peek().Lexme,
				"missing left-hand-side operand (comparison)")
		} else {
			return nil, err
		}
	}

	for s.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := s.peek()
		s.advance()
		right, err := term(s)
		if err != nil {
			right = handleMissingExpression(s, s.previous().Lexme,
				"missing right-hand-side operand (comparison)")
		}

		expr = ast.Binary{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

func handleMissingExpression(s *parser, lexme string, msg string) ast.Expr {
	s.parseErrOccured = true
	s.report(ParseError{Line: s.peek().Line, Lexme: lexme, Message: msg})
	return ast.Nothing{}
}

// Production rules:
//   - term -> (nothing | factor) (("-" | "+") (nothing | factor))*;
//   - precedence: 4
//   - associativity: left-to-right
func term(s *parser) (ast.Expr, error) {
	expr, err := factor(s)
	if err != nil {
		if s.match(token.MINUS, token.PLUS) {
			expr = handleMissingExpression(s, s.peek().Lexme,
				"missing left-hand-side operand (term)")
		} else {
			return nil, err
		}
	}

	for s.match(token.MINUS, token.PLUS) {
		operator := s.peek()
		s.advance()
		right, err := factor(s)
		if err != nil {
			right = handleMissingExpression(s, s.previous().Lexme,
				"missing right-hand-side operand (term)")
		}

		expr = ast.Binary{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

// Production rules:
//   - factor -> (unary | nothing) (("/" | "*") (unary | nothing))*;
//   - precedence: 3
//   - associativity: left-to-right
func factor(s *parser) (ast.Expr, error) {
	expr, err := unary(s)
	if err != nil {
		if s.match(token.SLASH, token.STAR) {
			expr = handleMissingExpression(s, s.peek().Lexme,
				"missing left-hand-side operand (factor)")
		} else {
			return nil, err
		}
	}

	for s.match(token.SLASH, token.STAR) {
		operator := s.peek()
		s.advance()
		right, err := unary(s)
		if err != nil {
			right = handleMissingExpression(s, s.previous().Lexme,
				"missing right-hand-side operand (factor)")
		}

		expr = ast.Binary{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

// Production rules:
//   - unary -> ("!" | "-") (unary | nothing) | primary;
//   - precedence: 2
//   - associativity: right-to-left
func unary(s *parser) (ast.Expr, error) {
	if s.match(token.BANG, token.MINUS) {
		operator := s.peek()
		s.advance()
		right, err := unary(s)
		if err != nil {
			right = handleMissingExpression(s, s.previous().Lexme,
				"missing operand (unary)")
		}

		return ast.Unary{Op: operator, Right: right}, nil
	}

	return primary(s)
}

// Production rules:
//   - primary -> NUMBER | STRING | nothing | "true" | "false" | "nil" | "(" expression ")";
//   - precedence: 1
//   - associativity: none
func primary(s *parser) (ast.Expr, error) {
	switch s.peek().Type {
	case token.FALSE:
		s.advance()
		return ast.Literal{Value: ast.Boolean(false)}, nil
	case token.TRUE:
		s.advance()
		return ast.Literal{Value: ast.Boolean(true)}, nil
	case token.NIL:
		s.advance()
		return ast.Literal{Value: ast.Nil{}}, nil
	case token.NUMBER:
		s.advance()

		var num float64
		b := s.previous().Literal
		buf := bytes.NewReader(b)
		err := binary.Read(buf, binary.LittleEndian, &num)
		if err != nil {
			panic(err)
		}

		return ast.Literal{Value: ast.Number(num)}, nil
	case token.STRING:
		s.advance()
		value := s.previous().Literal
		return ast.Literal{Value: ast.String(value)}, nil
	case token.LEFT_PAREN:
		s.advance()
		if expr, err := expression(s); err != nil {
			return nil, err
		} else {
			s.consume(token.RIGHT_PAREN, "expected ')' after expression (primary)")
			return ast.Grouping{Expr: expr}, nil
		}
	case token.IDENTIFIER:
		s.advance()
		return ast.Variable{Name: s.previous()}, nil
	case token.ERROR:
		s.parseErrOccured = true
		return ast.Nothing{}, nil
	default:
		err := ParseError{
			Line:    s.peek().Line,
			Lexme:   s.peek().Lexme,
			Message: "unexpected token"}
		s.report(err)
		return nil, errors.New("")
	}
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

func (s *parser) consume(typ token.TokenType, msg string) error {
	if s.check(typ) {
		s.advance()
		return nil
	}

	err := ParseError{
		Line:    s.peek().Line,
		Lexme:   s.peek().Lexme,
		Message: msg}
	s.parseErrOccured = true
	s.report(err)
	return errors.New("")
}

func (s *parser) match(types ...token.TokenType) bool {
	for _, typ := range types {
		if s.check(typ) {
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
