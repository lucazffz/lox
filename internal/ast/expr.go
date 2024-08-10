package ast

import "github.com/LucazFFz/lox/internal/token"

type Expr interface {
	PrettyPrint
}

type Binary struct {
	left  Expr
	op    token.Token
	right Expr
}

type Grouping struct {
	expr Expr
}

type Literal struct {
	value string
}

type Unary struct {
	op    token.Token
	right Expr
}
