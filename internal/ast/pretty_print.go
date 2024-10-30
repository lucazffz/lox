package ast

import (
	"fmt"
	"strconv"
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
	return t.Value.Print()
}

func (t Unary) Print() string {
	return parenthesize(t.Op.Lexme, t.Right)
}

func (t Ternary) Print() string {
	return parenthesize("ternary", t.Condition, t.Left, t.Right)
}

func (t Variable) Print() string {
	return parenthesize(t.Name.Lexme)
}

func (t Assign) Print() string {
	return fmt.Sprintf("(assign %s %s)", t.Name.Lexme, t.Value.Print())
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

// values
func (v Boolean) Print() string {
	if v {
		return "true"
	} else {
		return "false"
	}
}

func (v Number) Print() string {
	return strconv.FormatFloat(v.AsNumber(), 'f', -1, 64)
}

func (v Nil) Print() string {
	return "nil"
}

func (v Object) Print() string {
	return "object"
}

func (v String) Print() string {
	return v.AsString()
}

// statements
func (s Expression) Print() string {
	return parenthesize("expr", s.Expr)
}

func (s Print) Print() string {
	return parenthesize("print", s.Expr)
}

func (s Var) Print() string {
	return parenthesize("var", s.Initializer)
}

func (s If) Print() string {
	if s.ElseBranch != nil {
		return parenthesize("if", s.Condition, s.ThenBranch, s.ElseBranch)
	}
	return parenthesize("if", s.Condition, s.ThenBranch)
}

func (s Block) Print() string {
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
