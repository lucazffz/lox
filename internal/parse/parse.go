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
//   - declaration -> varDeclaration | funDeclaration | statement;
func declaration(s *parser) (ast.Stmt, error) {
	if s.match(token.VAR) {
		s.advance()
		stmt, err := varDeclaration(s)
		if err != nil {
			// reset the parser state between declarations
			// to avoid cascading errors
			s.synchronize()
			return nil, err
		}
		return stmt, nil
	}
	if s.match(token.FUN) {
		s.advance()
		stmt, err := function(s, "function")
		if err != nil {
			// reset the parser state between declarations
			// to avoid cascading errors
			s.synchronize()
			return nil, err
		}
		return stmt, nil
	}

	return statement(s)
}

// Production rules:
//   - funDeclaration -> "fun" IDENTIFIER "(" parameters? ")" blockStmt;
//   - parameters -> IDENTIFIER ("," IDENTIFIER)*;
func function(s *parser, kind string) (ast.Stmt, error) {
	if err := s.consume(token.IDENTIFIER, fmt.Sprintf("expected %s name", kind)); err != nil {
		return nil, err
	}

	name := s.previous()
	if err := s.consume(token.LEFT_PAREN, fmt.Sprintf("expected '(' after %s name", kind)); err != nil {
		return nil, err
	}

	var parameters []token.Token
	if !s.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				err := ParseError{
					Line:    s.peek().Line,
					Lexme:   s.peek().Lexme,
					Message: "cannot have more than 255 arguments"}
				return nil, err
			}
			if err := s.consume(token.IDENTIFIER, "expected parameter name"); err != nil {
				return nil, err
			}

			parameters = append(parameters, s.previous())

			if !s.match(token.COMMA) {
				break
			}

			s.advance()
		}
	}

	if err := s.consume(token.RIGHT_PAREN, "expected ')' after parameters"); err != nil {
		return nil, err
	}

	if err := s.consume(token.LEFT_BRACE, fmt.Sprintf("expected '{' before %s body", kind)); err != nil {
		return nil, err
	}

	block, err := blockStmt(s)
	if err != nil {
		return nil, err
	}

	// will never panic because blockStmt will always return a block
	body := block.(ast.BlockStmt).Statements
	return ast.FunctionStmt{Name: name, Parameters: parameters, Body: body}, nil
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
	var initializer ast.Expr = ast.NothingExpr{}
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

	return ast.VarStmt{Name: name, Initializer: initializer}, nil
}

// Production rules:
//   - statement -> exprStmt | printStmt | blockStmt |
//     ifStmt | whileStmt | forStmt | breakStmt | returnStmt;
func statement(s *parser) (ast.Stmt, error) {
	if s.match(token.IF) {
		s.advance()
		return ifStmt(s)
	}

	if s.match(token.WHILE) {
		s.advance()
		return whileStmt(s)
	}

	if s.match(token.FOR) {
		s.advance()
		return forStmt(s)
	}

	// Production rules:
	// - breakStmt -> "break" ";";
	if s.match(token.BREAK) {
		s.advance()
		if err := s.consume(token.SEMICOLON, "expected ';' after statement"); err != nil {
			return nil, err
		}
		return ast.BreakStmt{}, nil
	}

	// Production rules:
	// - returnStmt -> "return" expression? ";";
	if s.match(token.RETURN) {
		s.advance()
		var expr ast.Expr
		var err error
		if !s.check(token.SEMICOLON) {
			expr, err = expression(s)
			if err != nil {
				return nil, err
			}
		}

		if err := s.consume(token.SEMICOLON, "expected ';' after statement"); err != nil {
			return nil, err
		}

		return ast.ReturnStmt{Expr: expr}, nil
	}

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

	return ast.PrintStmt{Expr: expr}, nil
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

	return ast.BlockStmt{Statements: statements}, nil
}

// Production rules:
// - ifStmt -> "if" "(" expression ")" statement ("else" statement)?;
func ifStmt(s *parser) (ast.Stmt, error) {
	s.consume(token.LEFT_PAREN, "expected '(' after 'if'")
	condition, err := expression(s)
	if err != nil {
		return nil, err
	}
	s.consume(token.RIGHT_PAREN, "expected ')' after 'if'")
	thenBranch, err := statement(s)
	if err != nil {
		return nil, err
	}
	var elseBranch ast.Stmt = nil
	if s.match(token.ELSE) {
		s.advance()
		var err error = nil
		elseBranch, err = statement(s)
		if err != nil {
			return nil, err
		}
	}

	return ast.IfStmt{Condition: condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch}, nil
}

// Production rules:
// - whileStmt -> "while" "(" expression ")" statement;
func whileStmt(s *parser) (ast.Stmt, error) {
	s.consume(token.LEFT_PAREN, "expected '(' after 'while'")
	condition, err := expression(s)
	if err != nil {
		return nil, err
	}

	s.consume(token.RIGHT_PAREN, "expected ')' after 'while'")
	body, err := statement(s)
	if err != nil {
		return nil, err
	}

	return ast.WhileStmt{Condition: condition, Body: body}, nil
}

// Production rules:
//   - forStmt -> "for" "(" ( varDecl | exprStmt | ";")
//     expression? ";"
//     expression? ")" statement;
func forStmt(s *parser) (ast.Stmt, error) {
	s.consume(token.LEFT_PAREN, "expected '(' after 'for'")

	var initializer ast.Stmt = nil
	var err error = nil
	if s.match(token.SEMICOLON) {
		s.advance()
		initializer = nil
	} else if s.match(token.VAR) {
		s.advance()
		initializer, err = varDeclaration(s)
		if err != nil {
			return nil, err
		}
	} else {
		initializer, err = expressionStmt(s)
		if err != nil {
			return nil, err
		}
	}

	var condition ast.Expr = nil
	if !s.check(token.SEMICOLON) {
		condition, err = expression(s)
		if err != nil {
			return nil, err
		}
	}
	s.consume(token.SEMICOLON, "expected ';' after loop condition")

	var incrementer ast.Expr = nil
	if !s.check(token.RIGHT_PAREN) {
		incrementer, err = expression(s)
		if err != nil {
			return nil, err
		}
	}
	s.consume(token.RIGHT_PAREN, "expected ')' after for clause")

	// create ast
	var body ast.Stmt = nil
	body, err = statement(s)
	if err != nil {
		return nil, err
	}

	if incrementer != nil {
		body = ast.BlockStmt{
			Statements: []ast.Stmt{
				body,
				ast.ExpressionStmt{Expr: incrementer}},
		}
	}

	if condition == nil {
		var value ast.LoxBoolean = true
		condition = ast.LiteralExpr{Value: value}
	}

	body = ast.WhileStmt{Condition: condition, Body: body}

	if initializer != nil {
		body = ast.BlockStmt{
			Statements: []ast.Stmt{initializer, body},
		}
	}

	return body, nil
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

	return ast.ExpressionStmt{Expr: expr}, nil
}

// Production rules:
//   - expression -> assignment;
//   - precedence: none
//   - associativity: none
func expression(s *parser) (ast.Expr, error) {
	return assignment(s)
}

// Production rules:
//   - assignment -> IDENTIFIER "=" (assignment | comma);
//   - precedence: 16
//   - associativity: right-to-left
func assignment(s *parser) (ast.Expr, error) {
	expr, err := conditional(s)
	if err != nil {
		return nil, err
	}

	if s.match(token.EQUAL) {
		s.advance()
		value, err := assignment(s)
		if err != nil {
			return nil, err
		}

		if expr, ok := expr.(ast.VariableExpr); ok {
			return ast.AssignExpr{Name: expr.Name, Value: value}, nil
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

// // Production rules:
// //   - comma -> conditional ("," conditional)*;
// //   - precedence: 15
// //   - associativity: left-to-right
// func comma(s *parser) (ast.Expr, error) {
// 	expr, err := conditional(s)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	for s.match(token.COMMA) {
// 		operator := s.previous()
// 		if right, err := conditional(s); err != nil {
// 			return nil, err
// 		} else {
// 			expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
// 		}
// 	}
//
// 	return expr, nil
// }

// Production rules:
//   - conditional -> logical_or "?" logical_or ":" (conditional | logical_or);
//   - precedence: 13
//   - associativity: right-to-left
func conditional(s *parser) (ast.Expr, error) {
	expr, err := logicalOr(s)
	if err != nil {
		return nil, err
	}

	if !s.match(token.QUESTION) {
		return expr, nil
	}

	s.advance()
	left, err := logicalOr(s)
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

	expr = ast.TernaryExpr{Condition: expr, Left: left, Right: right}
	return expr, nil
}

// Production rules:
// - logical_or -> logical_and ("or" logical_and)*;
// - precedence: 12
// associativity: left-to-right
func logicalOr(s *parser) (ast.Expr, error) {
	expr, err := logicalAnd(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.OR) {
		operator := s.peek()
		s.advance()
		right, err := logicalAnd(s)
		if err != nil {
			right = handleMissingExpression(s, s.previous().Lexme,
				"missing right-hand-side operand (logical_or)")
		}
		expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

// Production rules:
// - logical_and -> equality ("and" equality)*;
// - precedence: 11
// associativity: left-to-right
func logicalAnd(s *parser) (ast.Expr, error) {
	expr, err := equality(s)
	if err != nil {
		return nil, err
	}

	for s.match(token.AND) {
		operator := s.peek()
		s.advance()
		right, err := equality(s)
		if err != nil {
			right = handleMissingExpression(s, s.previous().Lexme,
				"missing right-hand-side operand (logical_and)")
		}
		expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

// Production rules:
//   - equality -> (nothing | comparison) (("!=" | "==") (nothing | comparison))*;
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
			expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
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

		expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

func handleMissingExpression(s *parser, lexme string, msg string) ast.Expr {
	s.parseErrOccured = true
	s.report(ParseError{Line: s.peek().Line, Lexme: lexme, Message: msg})
	return ast.NothingExpr{}
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

		expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
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

		expr = ast.BinaryExpr{Left: expr, Op: operator, Right: right}
	}

	return expr, nil
}

// Production rules:
//   - unary -> ("!" | "-") (unary | nothing) | call;
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

		return ast.UnaryExpr{Op: operator, Right: right}, nil
	}

	return call(s)
}

// Production rules:
//   - call -> primary ("(" arguments? ")")*;
//   - precedence: 1
//   - associativity: left-to-right
func call(s *parser) (ast.Expr, error) {
	expr, err := functionExpr(s)
	if err != nil {
		return nil, err
	}

	for {
		if !s.match(token.LEFT_PAREN) {
			return expr, nil
		}

		s.advance()
		arguments := []ast.Expr{}
		if !s.check(token.RIGHT_PAREN) {
			for {
				if len(arguments) >= 255 {
					err := ParseError{
						Line:    s.peek().Line,
						Lexme:   s.peek().Lexme,
						Message: "cannot have more than 255 arguments"}
					return nil, err
				}

				if e, err := expression(s); err != nil {
					return nil, err
				} else {
					arguments = append(arguments, e)
				}

				if !s.match(token.COMMA) {
					break
				}

				s.advance()
			}
		}

		if err := s.consume(
			token.RIGHT_PAREN,
			"expected ')' after arguments"); err != nil {
			return nil, err
		}
		paren := s.peek()
		expr = ast.CallStmt{Callee: expr, Paren: paren, Arguments: arguments}
	}
}

func functionExpr(s *parser) (ast.Expr, error) {
    if !s.match(token.FUN) {
        return primary(s)
    }

    s.advance()
        
	if err := s.consume(token.LEFT_PAREN, "expected '(' after function"); err != nil {
		return nil, err
	}

	var parameters []token.Token
	if !s.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				err := ParseError{
					Line:    s.peek().Line,
					Lexme:   s.peek().Lexme,
					Message: "cannot have more than 255 arguments"}
				return nil, err
			}
			if err := s.consume(token.IDENTIFIER, "expected parameter name"); err != nil {
				return nil, err
			}

			parameters = append(parameters, s.previous())

			if !s.match(token.COMMA) {
				break
			}

			s.advance()
		}
	}

	if err := s.consume(token.RIGHT_PAREN, "expected ')' after parameters"); err != nil {
		return nil, err
	}

	if err := s.consume(token.LEFT_BRACE, "expected '{' before %s body"); err != nil {
		return nil, err
	}

	block, err := blockStmt(s)
	if err != nil {
		return nil, err
	}

	// will never panic because blockStmt will always return a block
	body := block.(ast.BlockStmt).Statements
	return ast.FunctionExpr{Parameters: parameters, Body: body}, nil
}

// Production rules:
//   - primary -> NUMBER | STRING | IDENTIFIER | nothing | "true" | "false" | "nil" | "(" expression ")";
//   - precedence: 1
//   - associativity: none
func primary(s *parser) (ast.Expr, error) {
	switch s.peek().Type {
	case token.FALSE:
		s.advance()
		return ast.LiteralExpr{Value: ast.LoxBoolean(false)}, nil
	case token.TRUE:
		s.advance()
		return ast.LiteralExpr{Value: ast.LoxBoolean(true)}, nil
	case token.NIL:
		s.advance()
		return ast.LiteralExpr{Value: ast.LoxNil{}}, nil
	case token.NUMBER:
		s.advance()

		var num float64
		b := s.previous().Literal
		buf := bytes.NewReader(b)
		err := binary.Read(buf, binary.LittleEndian, &num)
		if err != nil {
			panic(err)
		}

		return ast.LiteralExpr{Value: ast.LoxNumber(num)}, nil
	case token.STRING:
		s.advance()
		value := s.previous().Literal
		return ast.LiteralExpr{Value: ast.LoxString(value)}, nil
	case token.LEFT_PAREN:
		s.advance()
		if expr, err := expression(s); err != nil {
			return nil, err
		} else {
			s.consume(token.RIGHT_PAREN, "expected ')' after expression (primary)")
			return ast.GroupingExpr{Expr: expr}, nil
		}
	case token.IDENTIFIER:
		s.advance()
		return ast.VariableExpr{Name: s.previous()}, nil
	case token.ERROR:
		s.parseErrOccured = true
		return ast.NothingExpr{}, nil
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
