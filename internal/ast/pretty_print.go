package ast

import (
	"fmt"
	"strconv"
	"strings"
)

type PrettyPrint interface {
	Print() string
}

func (t BinaryExpr) Print() string {
	return parenthesize(t.Op.Lexme, t.Left, t.Right)
}

func (t GroupingExpr) Print() string {
	return parenthesize("group", t.Expr)
}

func (t LiteralExpr) Print() string {
	return t.Value.Print()
}

func (t UnaryExpr) Print() string {
	return parenthesize(t.Op.Lexme, t.Right)
}

func (t TernaryExpr) Print() string {
	return parenthesize("ternary", t.Condition, t.Left, t.Right)
}

func (t VariableExpr) Print() string {
	return parenthesize(t.Name.Lexme)
}

func (t AssignExpr) Print() string {
	return fmt.Sprintf("(assign %s %s)", t.Name.Lexme, t.Value.Print())
}


func (t NothingExpr) Print() string {
	return parenthesize("Nothing")
}

func parenthesize(name string, exprs ...PrettyPrint) string {
	var builder = strings.Builder{}
	builder.WriteString("(")
	builder.WriteString(name)

	for _, expr := range exprs {
		builder.WriteString(" ")
		builder.WriteString(expr.Print())
	}

	builder.WriteString(")")
	return builder.String()
}

// values
func (v LoxBoolean) Print() string {
	if v {
		return "true"
	} else {
		return "false"
	}
}

func (v LoxNumber) Print() string {
	return strconv.FormatFloat(AsNumber(v), 'f', -1, 64)
}

func (v LoxNil) Print() string {
	return "nil"
}

func (v LoxObject) Print() string {
	return "object"
}

func (v LoxString) Print() string {
	return AsString(v)
}

// statements
func (s ExpressionStmt) Print() string {
	return parenthesize("expr", s.Expr)
}

func (s PrintStmt) Print() string {
	return parenthesize("print", s.Expr)
}

func (s VarStmt) Print() string {
	return parenthesize("var", s.Initializer)
}

func (s IfStmt) Print() string {
	if s.ElseBranch != nil {
		return parenthesize("if", s.Condition, s.ThenBranch, s.ElseBranch)
	}
	return parenthesize("if", s.Condition, s.ThenBranch)
}

func (s WhileStmt) Print() string {
	return parenthesize("while", s.Condition, s.Body)
}

func (s BlockStmt) Print() string {
	// cannot do parenthesize("block", s.Statements...)
	// because go will not convert from Stmt[] to PrettyPrint[]
	// because it generally does not do implicit conversions with time
	// complexity > O(1) apparently
	args := make([]PrettyPrint, len(s.Statements))
	for i := range s.Statements {
		args[i] = s.Statements[i]
	}
	return parenthesize("block", args...)
}

func (s BreakStmt) Print() string {
	return parenthesize("break")
}

func (s ReturnStmt) Print() string {
    return parenthesize("return", s.Expr)
}

func (t FunctionStmt) Print() string {
	return parenthesize("function")
}


func (t CallStmt) Print() string {
	// args := make([]PrettyPrint, len(t.Arguments)+1)
	// args[0] = t.Callee
	// for i := range args {
	// 	args[i+1] = t.Arguments[i]
	// }
	// return parenthesize("call", args...)
	//
    return ""
}

