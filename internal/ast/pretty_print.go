package ast

import (
	"strings"
)

type PrettyPrint interface {
	Print() string
}

func (t Binary) Print() string {
	return parenthesize(t.op.Lexme, t.left, t.right)
}

func (t Grouping) Print() string {
	return parenthesize("group", t.expr)
}

func (t Literal) Print() string {
	if t.value == "" {
		return "nil"
	}
	return t.value
}

func (t Unary) Print() string {
	return parenthesize(t.op.Lexme, t.right)
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
