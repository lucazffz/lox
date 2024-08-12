package ast

import "github.com/LucazFFz/lox/internal/token"

type Expr interface {
	PrettyPrint
}

type Binary struct {
	Left  Expr
	Op    token.Token
	Right Expr
}

type Grouping struct {
	Expr Expr
}

type Literal struct {
	Value string
}

type Unary struct {
	Op    token.Token
	Right Expr
}
