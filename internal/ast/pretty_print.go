package ast

import (
	"strings"
)

type PrettyPrint interface {
	Print() string
}

func (t Binary) Print() string {
	return parenthesize(t.Op.Lexme, t.Left, t.Right)
}

func (t Grouping) Print() string {
	return parenthesize("group", t.Expr)
}

func (t Literal) Print() string {
	if t.Value == "" {
		return "nil"
	}
	return t.Value
}

func (t Unary) Print() string {
	return parenthesize(t.Op.Lexme, t.Right)
}

func (t Ternary) Print() string {
	return parenthesize("ternary", t.Condition, t.Left, t.Right)
}

func (t Nothing) Print() string {
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
