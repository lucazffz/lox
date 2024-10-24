// Code generated by tools/expr_gen.py. DO NOT EDIT.

package ast

import "github.com/LucazFFz/lox/internal/token"

type Stmt interface {
    EvaluateStmt
    PrettyPrint
}

type Expression struct {
    Expr Expr;
}

type Print struct {
    Expr Expr;
}

type Var struct {
    Name token.Token;
    Initializer Expr;
}

type Block struct {
    Statements[] Stmt
}
